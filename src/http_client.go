package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
)

func mergeMaps(ms ...map[string]string) map[string]string {
	res := make(map[string]string)
	for _, m := range ms {
		for k, v := range m {
			res[k] = v
		}
	}
	return res
}

// HTTPRequest is a client used to make HTTP requests
type HTTPRequest struct {
	ID      uuid.UUID
	client  *http.Client
	Method  string
	URL     string
	Body    string
	Headers map[string]string
	Cookies map[string]string
	Proxy   string
}

func newBaseHTTPRequest(proxyURL string) *HTTPRequest {
	client := NewHTTPClient(proxyURL)
	return &HTTPRequest{
		client: client,
		Proxy:  proxyURL,
	}
}

// Execute the HTTP request
func (r *HTTPRequest) Execute() (*http.Response, []byte) {
	req, _ := http.NewRequest(r.Method, r.URL, bytes.NewBuffer([]byte(r.Body)))
	for key, value := range r.Headers {
		req.Header.Add(key, value)
	}
	for key, value := range r.Cookies {
		cookie := http.Cookie{Name: key, Value: value}
		req.AddCookie(&cookie)
	}
	resp, err := r.client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("hi")
	}

	return resp, ret
}

// Execute the HTTP request
func (r *HTTPRequest) toActionResponse() *ActionResponse {
	resp, body := r.Execute()
	respHeaders := make(map[string]string)
	for key, values := range resp.Header {
		var joinedValues string
		for _, value := range values {
			joinedValues += fmt.Sprintf("%s", value)
		}
		respHeaders[key] = joinedValues
	}
	actionResponse := &ActionResponse{
		ID:         uuid.NewV4(),
		RequestID:  r.ID,
		Method:     r.Method,
		URL:        r.URL,
		StatusCode: resp.StatusCode,
		Date:       time.Now(),
		Headers:    respHeaders,
		Body:       string(body),
	}
	return actionResponse
}

// CookieJar is a type
type CookieJar struct {
	sync.Mutex
	cookies map[string][]*http.Cookie
}

// NewHTTPClient is a func
func NewHTTPClient(proxyURL string) *http.Client {
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if proxyURL != "" {
		parsedProxyURL, _ := url.Parse(proxyURL)
		transport.Proxy = http.ProxyURL(parsedProxyURL)
	}

	netClient := &http.Client{
		Transport: transport,
	}
	return netClient
}
