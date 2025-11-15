package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// FetchTool implements HTTP/HTTPS request functionality
type FetchTool struct {
	client *http.Client
}

// FetchParams defines the parameters for the Fetch tool
type FetchParams struct {
	URL        string            `json:"url"`
	Method     string            `json:"method,omitempty"`      // GET, POST, PUT, DELETE, PATCH
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
	Format     string            `json:"format,omitempty"`      // "text", "json", "html", "xml"
	Timeout    int               `json:"timeout,omitempty"`     // seconds
	MaxRetries int               `json:"max_retries,omitempty"`
	Auth       *AuthConfig       `json:"auth,omitempty"`
}

// AuthConfig defines authentication configuration
type AuthConfig struct {
	Type   string `json:"type"`   // "basic", "bearer", "apikey"
	User   string `json:"user,omitempty"`
	Pass   string `json:"pass,omitempty"`
	Token  string `json:"token,omitempty"`
	APIKey string `json:"apikey,omitempty"`
	Header string `json:"header,omitempty"` // Header name for API key
}

// FetchResult represents the result of a fetch operation
type FetchResult struct {
	StatusCode  int               `json:"status_code"`
	Headers     map[string]string `json:"headers"`
	Content     string            `json:"content"`
	ContentType string            `json:"content_type"`
	Size        int64             `json:"size"`
	Duration    int64             `json:"duration_ms"`
	RedirectURL string            `json:"redirect_url,omitempty"`
}

// NewFetchTool creates a new Fetch tool instance
func NewFetchTool() *FetchTool {
	return &FetchTool{
		client: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("stopped after 10 redirects")
				}
				return nil
			},
		},
	}
}

// Name returns the tool's name
func (t *FetchTool) Name() string {
	return "fetch"
}

// Description returns the tool's description
func (t *FetchTool) Description() string {
	return "Fetch content from web resources with support for authentication, custom headers, and various HTTP methods"
}

// Schema returns the JSON schema for the tool's parameters
func (t *FetchTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "URL to fetch",
			},
			"method": map[string]interface{}{
				"type":        "string",
				"description": "HTTP method: GET, POST, PUT, DELETE, PATCH",
				"enum":        []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
			},
			"headers": map[string]interface{}{
				"type":        "object",
				"description": "Custom headers as key-value pairs",
			},
			"body": map[string]interface{}{
				"type":        "string",
				"description": "Request body content",
			},
			"format": map[string]interface{}{
				"type":        "string",
				"description": "Expected response format: text, json, html, xml",
				"enum":        []string{"text", "json", "html", "xml"},
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Timeout in seconds (default: 30, max: 300)",
				"minimum":     1,
				"maximum":     300,
			},
			"max_retries": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of retries on failure (default: 0)",
				"minimum":     0,
				"maximum":     5,
			},
		},
		"required": []string{"url"},
	}
}

// Validate checks if the parameters are valid
func (t *FetchTool) Validate(params json.RawMessage) error {
	var p FetchParams
	if err := json.Unmarshal(params, &p); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	if p.URL == "" {
		return fmt.Errorf("url is required")
	}

	if !strings.HasPrefix(p.URL, "http://") && !strings.HasPrefix(p.URL, "https://") {
		return fmt.Errorf("url must start with http:// or https://")
	}

	if p.Method != "" {
		validMethods := map[string]bool{"GET": true, "POST": true, "PUT": true, "DELETE": true, "PATCH": true}
		if !validMethods[strings.ToUpper(p.Method)] {
			return fmt.Errorf("invalid method: %s", p.Method)
		}
	}

	if p.Timeout < 0 || p.Timeout > 300 {
		return fmt.Errorf("timeout must be between 1 and 300 seconds")
	}

	if p.MaxRetries < 0 || p.MaxRetries > 5 {
		return fmt.Errorf("max_retries must be between 0 and 5")
	}

	if p.Auth != nil {
		if err := t.validateAuth(p.Auth); err != nil {
			return err
		}
	}

	return nil
}

// validateAuth validates authentication configuration
func (t *FetchTool) validateAuth(auth *AuthConfig) error {
	validTypes := map[string]bool{"basic": true, "bearer": true, "apikey": true}
	if !validTypes[auth.Type] {
		return fmt.Errorf("invalid auth type: %s", auth.Type)
	}

	switch auth.Type {
	case "basic":
		if auth.User == "" || auth.Pass == "" {
			return fmt.Errorf("basic auth requires user and pass")
		}
	case "bearer":
		if auth.Token == "" {
			return fmt.Errorf("bearer auth requires token")
		}
	case "apikey":
		if auth.APIKey == "" {
			return fmt.Errorf("apikey auth requires apikey")
		}
		if auth.Header == "" {
			auth.Header = "X-API-Key" // Default header
		}
	}

	return nil
}

// Execute runs the tool with the given parameters
func (t *FetchTool) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var p FetchParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Default values
	if p.Method == "" {
		p.Method = "GET"
	}
	if p.Timeout == 0 {
		p.Timeout = 30
	}

	// Update client timeout
	t.client.Timeout = time.Duration(p.Timeout) * time.Second

	// Execute with retries
	maxAttempts := p.MaxRetries + 1
	var lastErr error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		result, err := t.executeRequest(ctx, p)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Don't retry on client errors (4xx) or context errors
		if result != nil && result.StatusCode >= 400 && result.StatusCode < 500 {
			return result, err
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Wait before retry (exponential backoff)
		if attempt < maxAttempts-1 {
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			time.Sleep(backoff)
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", maxAttempts, lastErr)
}

// executeRequest performs a single HTTP request
func (t *FetchTool) executeRequest(ctx context.Context, p FetchParams) (*FetchResult, error) {
	startTime := time.Now()

	// Create request
	var bodyReader io.Reader
	if p.Body != "" {
		bodyReader = strings.NewReader(p.Body)
	}

	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(p.Method), p.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range p.Headers {
		req.Header.Set(key, value)
	}

	// Set authentication
	if p.Auth != nil {
		t.setAuthentication(req, p.Auth)
	}

	// Execute request
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	duration := time.Since(startTime)

	// Build result
	result := &FetchResult{
		StatusCode:  resp.StatusCode,
		Headers:     make(map[string]string),
		Content:     string(body),
		ContentType: resp.Header.Get("Content-Type"),
		Size:        int64(len(body)),
		Duration:    duration.Milliseconds(),
	}

	// Copy headers
	for key, values := range resp.Header {
		if len(values) > 0 {
			result.Headers[key] = values[0]
		}
	}

	// Check for redirects
	if resp.Request.URL.String() != p.URL {
		result.RedirectURL = resp.Request.URL.String()
	}

	// Check for errors based on status code
	if resp.StatusCode >= 400 {
		return result, fmt.Errorf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return result, nil
}

// setAuthentication sets authentication headers on the request
func (t *FetchTool) setAuthentication(req *http.Request, auth *AuthConfig) {
	switch auth.Type {
	case "basic":
		req.SetBasicAuth(auth.User, auth.Pass)
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+auth.Token)
	case "apikey":
		headerName := auth.Header
		if headerName == "" {
			headerName = "X-API-Key"
		}
		req.Header.Set(headerName, auth.APIKey)
	}
}
