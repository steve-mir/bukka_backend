package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HTTPRequest struct holds information for making HTTP requests.
type HTTPRequest struct {
	BaseURL string
	Headers map[string]string
}

// NewHTTPRequest creates a new HTTPRequest instance with the given base URL and headers.
func NewHTTPRequest(baseURL string, headers map[string]string) *HTTPRequest {
	return &HTTPRequest{
		BaseURL: baseURL,
		Headers: headers,
	}
}

// Get makes a GET request to the specified endpoint.
func (r *HTTPRequest) Get(endpoint string) ([]byte, int, error) {
	url := fmt.Sprintf("%s/%s", r.BaseURL, endpoint)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, err
	}

	// Set headers
	for key, value := range r.Headers {
		req.Header.Set(key, value)
	}

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return body, resp.StatusCode, nil
}

// func (r *HTTPRequest) Get(endpoint string) ([]byte, error) {
// 	url := fmt.Sprintf("%s/%s", r.BaseURL, endpoint)
// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Set headers
// 	for key, value := range r.Headers {
// 		req.Header.Set(key, value)
// 	}

// 	// Perform the request
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	// Read the response body
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return body, nil
// }

// Post makes a POST request to the specified endpoint with the given payload.
func (r *HTTPRequest) Post(endpoint string, payload interface{}) ([]byte, int, error) {
	url := fmt.Sprintf("%s/%s", r.BaseURL, endpoint)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, 0, err
	}

	// Set headers
	for key, value := range r.Headers {
		req.Header.Set(key, value)
	}

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return body, resp.StatusCode, nil
}
