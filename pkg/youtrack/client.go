package youtrack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// RESTLogger is the interface for logging REST calls
type RESTLogger interface {
	LogRESTCall(method, path string, duration time.Duration)
	LogRESTError(method, path string, body interface{}, statusCode int, errMsg string)
}

type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     RESTLogger
}

// SetLogger sets the REST logger for the client
func (c *Client) SetLogger(logger RESTLogger) {
	c.logger = logger
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) doRequest(ctx *YouTrackContext, method, path string, query url.Values, body interface{}) (*http.Response, error) {
	start := time.Now()

	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	u.Path = path
	if query != nil {
		u.RawQuery = query.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx.Context(), method, u.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+ctx.APIKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Log successful REST call
	if c.logger != nil {
		c.logger.LogRESTCall(method, path, duration)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		errMsg := string(bodyBytes)

		// Log REST error
		if c.logger != nil {
			c.logger.LogRESTError(method, path, body, resp.StatusCode, errMsg)
		}

		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    errMsg,
		}
	}

	return resp, nil
}

func (c *Client) Get(ctx *YouTrackContext, path string, query url.Values) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodGet, path, query, nil)
}

func (c *Client) Post(ctx *YouTrackContext, path string, body interface{}) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodPost, path, nil, body)
}

func (c *Client) PostWithQuery(ctx *YouTrackContext, path string, query url.Values, body interface{}) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodPost, path, query, body)
}

func (c *Client) Put(ctx *YouTrackContext, path string, body interface{}) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodPut, path, nil, body)
}

func (c *Client) Delete(ctx *YouTrackContext, path string) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}
