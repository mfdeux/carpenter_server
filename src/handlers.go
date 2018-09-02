package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

// ActionRequest is a type
type ActionRequest struct {
	ID             uuid.UUID         `json:"id"`
	Method         string            `json:"method"`
	URL            string            `json:"url"`
	InheritHeaders bool              `json:"inherit_headers"`
	InheritCookies bool              `json:"inherit_cookies"`
	Headers        map[string]string `json:"headers"`
	Cookies        map[string]string `json:"cookies"`
	Proxy          string            `json:"proxy"`
	Body           string            `json:"body"`
	Log            bool              `json:"log"`
	LogBody        bool              `json:"log_body"`
	Tags           []string          `json:"tags"`
	Request        *http.Request
}

func (ar *ActionRequest) toHTTPRequest() *HTTPRequest {
	request := newBaseHTTPRequest(ar.Proxy)
	inheritedHeaders := make(map[string]string)
	inheritedCookies := make(map[string]string)
	if ar.InheritHeaders == true {
		requestHeaders := ar.Request.Header
		for key, values := range requestHeaders {
			var joinedValues string
			for _, value := range values {
				joinedValues += fmt.Sprintf("%s", value)
			}
			inheritedHeaders[key] = joinedValues
		}
	}
	if ar.InheritCookies == true {
		for _, cookie := range ar.Request.Cookies() {
			inheritedCookies[cookie.Name] = cookie.Value
		}
	}
	request.ID = ar.ID
	request.Method = ar.Method
	request.URL = ar.URL
	request.Body = ar.Body
	request.Headers = mergeMaps(inheritedHeaders, ar.Headers)
	request.Cookies = mergeMaps(inheritedCookies, ar.Cookies)
	return request
}

// ActionResponse is a type
type ActionResponse struct {
	ID         uuid.UUID         `json:"id"`
	RequestID  uuid.UUID         `json:"request_id"`
	Method     string            `json:"method"`
	URL        string            `json:"url"`
	StatusCode int               `json:"status_code"`
	Date       time.Time         `json:"date"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

func handleActionRequest(c *gin.Context) {
	var actionRequest ActionRequest
	c.BindJSON(&actionRequest)
	actionRequest.ID = uuid.NewV4()
	actionRequest.Request = c.Request
	httpRequest := actionRequest.toHTTPRequest()
	actionResponse := httpRequest.toActionResponse()
	c.JSON(200, actionResponse)
}

func handleStatsRequest(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]int{"last_hour": getIntervalSums(5, time.Duration(1)*time.Hour), "last_24h": getIntervalSums(5, time.Duration(24)*time.Hour), "last_7d": getIntervalSums(5, time.Duration(7*24)*time.Hour), "last_30d": getIntervalSums(5, time.Duration(30*24)*time.Hour)})
	return
}

func handleHealthCheck(c *gin.Context) {
	c.String(http.StatusOK, "")
	return
}
