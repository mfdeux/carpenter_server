package main

import (
	"crypto/tls"
	"encoding/gob"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

// UserAgent is a type
type UserAgent struct {
	String string `json:"string"`
	Weight int    `json:"weight"`
	CanUse func(*http.Request) bool
}

// Proxy is a type
type Proxy struct {
	URL    url.URL `json:"url"`
	Weight int     `json:"weight"`
	CanUse func(*http.Request) bool
}

// HTTPClient is a client used to make HTTP requests
type HTTPClient struct {
	client    *http.Client
	Headers   map[string]string
	Cookies   map[string]string
	Proxy     string
	CookieJar CookieJar
	Timeout   time.Duration
	WG        sync.WaitGroup
}

// DownloadUserAgents is a func
func (c *HTTPClient) DownloadUserAgents() {
	resp, err := c.client.Get("url")
}

// LoadUserAgents is a func
func (c *HTTPClient) LoadUserAgents(filename string, appendOnly bool) error {
	decodeFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer decodeFile.Close()

	// Create a decoder
	decoder := gob.NewDecoder(decodeFile)

	// Place to decode into
	userAgents := []UserAgent{}

	// Decode -- We need to pass a pointer otherwise accounts2 isn't modified
	decoder.Decode(&userAgents)
	if appendOnly {
		c.UserAgents = append(c.UserAgents, userAgents...)
	} else {
		c.UserAgents = userAgents
	}

	return nil
}

// Sleep generates a normally distributed random delay with given mean and stdDev
// and blocks for that duration.
func Sleep(mean time.Duration, stdDev time.Duration) {
	fMean := float64(mean)
	fStdDev := float64(stdDev)
	delay := time.Duration(math.Max(1, rand.NormFloat64()*fStdDev+fMean))
	time.Sleep(delay)
}

// DownloadProxies is a func
func (c *HTTPClient) DownloadProxies() {
	url := "http://pubproxy.com/api/proxy?limit=50&format=json&type=socks5"
	// 	data: [
	// {
	// ipPort: "89.236.17.106:3128",
	// ip: "89.236.17.106",
	// port: "3128",
	// country: "SE",
	// last_checked: "2018-05-28 20:38:44",
	// proxy_level: "elite",
	// type: "socks5",
	// speed: "9",
	// support: {
	// https: 1,
	// get: 1,
	// post: 1,
	// cookies: 1,
	// referer: 1,
	// user_agent: 1,
	// google: 0
	// }
	// },
	resp, err := c.client.Get("url")
}

// LoadProxies is a func
func (c *HTTPClient) LoadProxies(filename string, appendOnly bool) error {
	decodeFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer decodeFile.Close()

	// Create a decoder
	decoder := gob.NewDecoder(decodeFile)

	// Place to decode into
	proxies := []Proxy{}

	// Decode -- We need to pass a pointer otherwise accounts2 isn't modified
	decoder.Decode(&proxies)
	if appendOnly {
		c.Proxies = append(c.Proxies, proxies...)
	} else {
		c.Proxies = proxies
	}

	return nil
}

// func (c *HTTPClient) Get(url string, params map[string]string, headers map[string]string, toText bool, toJSON bool, toGQ bool) {
// 	req, err := http.NewRequest("GET", facebookPageURL, nil)
// 	if err != nil {
// 		log.Errorf("Unable to make GET request for user: %s", pageID)
// 		return
// 	}
// 	req.Header.Set("User-Agent", "Chrome 55.0x")
// 	// Set headers
// 	for hName, hValue := range c.Headers {
// 		req.Header.Set(hName, hValue)
// 	}
// 	// Set cookies
// 	for cName, cValue := range c.Cookies {
// 		req.AddCookie(&http.Cookie{
// 			Name:  cName,
// 			Value: cValue,
// 		})
// 	}

// 	resp, err := c.client.Do(req)
// 	if err != nil {
// 		log.Errorf("Unable to fetch Facebook posts for page: %s", pageID)
// 		return
// 	}
// 	defer resp.Body.Close()

// 	doc, err := goquery.NewDocumentFromReader(resp.Body)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	var body []byte
// 	if resp.StatusCode == http.StatusOK {
// 		body, err = ioutil.ReadAll(resp.Body)
// 		if err != nil {
// 			log.Errorf("Unable to read response body for page: %s", pageID)
// 			return
// 		}
// 	} else {
// 		log.Errorf("Unable to fetch Facebook posts for page: %s as received bad HTTP status code", pageID)
// 		return
// 	}
// 	return string(body), nil
// }

// func (c *HTTPClient) GetAll(url map[string]string) {

// }

// CookieJar is a type
type CookieJar struct {
	sync.Mutex
	cookies map[string][]*http.Cookie
}

// SetCookies is a func
func (jar *CookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.Lock()
	if _, ok := jar.cookies[u.Host]; ok {
		for _, c := range cookies {
			jar.cookies[u.Host] = append(jar.cookies[u.Host], c)
		}
	} else {
		jar.cookies[u.Host] = cookies
	}
	jar.Unlock()
}

// Cookies is a func
func (jar *CookieJar) Cookies(u *url.URL) []*http.Cookie {
	return jar.cookies[u.Host]
}

// LoadCookiesFromFile is a func
func (jar *CookieJar) LoadCookiesFromFile(filename string) error {
	decodeFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer decodeFile.Close()

	// Create a decoder
	decoder := gob.NewDecoder(decodeFile)

	// Place to decode into
	cookies := make(map[string][]*http.Cookie)

	// Decode -- We need to pass a pointer otherwise accounts2 isn't modified
	decoder.Decode(&cookies)
	jar.cookies = cookies
	return nil
}

// SaveCookiesToFile is a func
func (jar *CookieJar) SaveCookiesToFile(filename string) error {
	encodeFile, err := os.Create(filename)
	if err != nil {
		return err
	}

	// Since this is a binary format large parts of it will be unreadable
	encoder := gob.NewEncoder(encodeFile)

	// Write to the file
	if err := encoder.Encode(jar.cookies); err != nil {
		return err
	}
	encodeFile.Close()
	return nil
}

// NewCookieJar is a func
func NewCookieJar() *CookieJar {
	jar := new(CookieJar)
	jar.cookies = make(map[string][]*http.Cookie)
	return jar
}

// NewHTTPClient is a func
func NewHTTPClient() *http.Client {
	proxyURL, err := url.Parse("http://proxyIp:proxyPort")

	netClient := &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			Proxy: http.ProxyURL(proxyURL),
		},
	}
	return netClient
}

// baseUrl, err := url.Parse("http://google.com/search")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	params := url.Values{}
// 	params.Add("pass%word", "key%20word")

// 	baseUrl.RawQuery = params.Encode()
// 	fmt.Println(baseUrl)
