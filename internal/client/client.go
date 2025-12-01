package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Client wraps HTTP client for registry API calls
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
	Verbose    bool
}

// NewClient creates a new API client
func NewClient(baseURL, token string, timeout time.Duration, verbose bool) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		Verbose: verbose,
	}
}

// doRequest executes an HTTP request with authentication
func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := c.BaseURL + path
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	// Add Basic Auth if token is provided
	if c.Token != "" {
		req.Header.Set("Authorization", "Basic "+c.Token)
	}

	// Execute request
	if c.Verbose {
		fmt.Fprintf(os.Stderr, "[DEBUG] %s %s\n", method, url)
	}

	return c.HTTPClient.Do(req)
}

// Get executes a GET request
func (c *Client) Get(path string) (*http.Response, error) {
	return c.doRequest("GET", path, nil)
}

// Post executes a POST request
func (c *Client) Post(path string, body interface{}) (*http.Response, error) {
	return c.doRequest("POST", path, body)
}

// Put executes a PUT request
func (c *Client) Put(path string, body interface{}) (*http.Response, error) {
	return c.doRequest("PUT", path, body)
}

// Delete executes a DELETE request
func (c *Client) Delete(path string) (*http.Response, error) {
	return c.doRequest("DELETE", path, nil)
}
