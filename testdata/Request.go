package testdata

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/labstack/echo/v4"
)

// Request holds an API request parameters
type Request struct {
	Method  string
	Path    string
	Body    interface{}
	Headers map[string]string
}

// Response wraps http.Response and adds a body string field
type Response struct {
	http.Response
	BodyString string
	StatusCode int
}

// Request performs a request to the API using httptest with the given echo instance
func (s *Suite) Request(e *echo.Echo, r *Request, responseBody ...interface{}) *Response {
	if e == nil {
		s.t.Fatal("Echo instance is required. Call SetupAPI() first.")
	}

	jsonBody := &bytes.Buffer{}
	if r.Body != nil {
		err := json.NewEncoder(jsonBody).Encode(r.Body)
		if err != nil {
			s.t.Fatalf("Failed to encode request body: %v", err)
			return nil
		}
	}

	req := httptest.NewRequest(r.Method, r.Path, jsonBody)
	
	// Set headers
	for header, value := range r.Headers {
		req.Header.Set(header, value)
	}

	// Set content-type if body is provided and no content-type header is set
	if r.Body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", echo.MIMEApplicationJSON)
	}

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Create response
	resp := &Response{
		StatusCode: rec.Code,
		BodyString: rec.Body.String(),
	}

	// Copy response headers
	resp.Header = rec.Header()

	// Unmarshal response body if provided
	if len(responseBody) > 0 && len(rec.Body.Bytes()) > 0 {
		data := rec.Body.Bytes()
		
		// Handle potential JSON string wrapping
		var modifiedData string
		if isJSONString(string(data)) && !isJSON(string(data)) {
			if err := json.Unmarshal(data, &modifiedData); err != nil {
				s.t.Fatalf("Failed to unmarshal JSON string: %v", err)
			}
		} else {
			modifiedData = string(data)
		}

		if err := json.Unmarshal([]byte(modifiedData), responseBody[0]); err != nil {
			s.t.Fatalf("Failed to unmarshal response body: %v", err)
		}
	}

	return resp
}

func isJSONString(s string) bool {
	var js string
	return json.Unmarshal([]byte(s), &js) == nil
}

func isJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}
