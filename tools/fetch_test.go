package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchTool_Name(t *testing.T) {
	tool := NewFetchTool()
	if tool.Name() != "fetch" {
		t.Errorf("Expected name 'fetch', got '%s'", tool.Name())
	}
}

func TestFetchTool_Validate(t *testing.T) {
	tool := NewFetchTool()

	tests := []struct {
		name    string
		params  string
		wantErr bool
	}{
		{
			name:    "valid url",
			params:  `{"url": "https://example.com"}`,
			wantErr: false,
		},
		{
			name:    "valid with method",
			params:  `{"url": "https://example.com", "method": "POST"}`,
			wantErr: false,
		},
		{
			name:    "valid with headers",
			params:  `{"url": "https://example.com", "headers": {"Accept": "application/json"}}`,
			wantErr: false,
		},
		{
			name:    "missing url",
			params:  `{}`,
			wantErr: true,
		},
		{
			name:    "invalid url scheme",
			params:  `{"url": "ftp://example.com"}`,
			wantErr: true,
		},
		{
			name:    "invalid method",
			params:  `{"url": "https://example.com", "method": "INVALID"}`,
			wantErr: true,
		},
		{
			name:    "timeout too high",
			params:  `{"url": "https://example.com", "timeout": 500}`,
			wantErr: true,
		},
		{
			name:    "max_retries too high",
			params:  `{"url": "https://example.com", "max_retries": 10}`,
			wantErr: true,
		},
		{
			name:    "valid basic auth",
			params:  `{"url": "https://example.com", "auth": {"type": "basic", "user": "test", "pass": "pass"}}`,
			wantErr: false,
		},
		{
			name:    "invalid basic auth missing user",
			params:  `{"url": "https://example.com", "auth": {"type": "basic", "pass": "pass"}}`,
			wantErr: true,
		},
		{
			name:    "valid bearer auth",
			params:  `{"url": "https://example.com", "auth": {"type": "bearer", "token": "abc123"}}`,
			wantErr: false,
		},
		{
			name:    "invalid bearer auth missing token",
			params:  `{"url": "https://example.com", "auth": {"type": "bearer"}}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tool.Validate(json.RawMessage(tt.params))
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFetchTool_Execute_GET(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	tool := NewFetchTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"url": server.URL,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	fetchResult, ok := result.(*FetchResult)
	if !ok {
		t.Fatal("Result is not a FetchResult")
	}

	if fetchResult.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", fetchResult.StatusCode)
	}

	if fetchResult.Content != "Hello, World!" {
		t.Errorf("Expected content 'Hello, World!', got '%s'", fetchResult.Content)
	}

	if fetchResult.ContentType != "text/plain" {
		t.Errorf("Expected content type 'text/plain', got '%s'", fetchResult.ContentType)
	}
}

func TestFetchTool_Execute_POST(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		if string(body) != "test data" {
			t.Errorf("Expected body 'test data', got '%s'", string(body))
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Created"))
	}))
	defer server.Close()

	tool := NewFetchTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"url":    server.URL,
		"method": "POST",
		"body":   "test data",
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	fetchResult := result.(*FetchResult)

	if fetchResult.StatusCode != 201 {
		t.Errorf("Expected status code 201, got %d", fetchResult.StatusCode)
	}

	if fetchResult.Content != "Created" {
		t.Errorf("Expected content 'Created', got '%s'", fetchResult.Content)
	}
}

func TestFetchTool_Execute_WithHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom-Header") != "custom-value" {
			t.Errorf("Expected custom header, got '%s'", r.Header.Get("X-Custom-Header"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	tool := NewFetchTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"url": server.URL,
		"headers": map[string]string{
			"X-Custom-Header": "custom-value",
		},
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	fetchResult := result.(*FetchResult)
	if fetchResult.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", fetchResult.StatusCode)
	}
}

func TestFetchTool_Execute_BasicAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if user != "testuser" || pass != "testpass" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Authenticated"))
	}))
	defer server.Close()

	tool := NewFetchTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"url": server.URL,
		"auth": map[string]string{
			"type": "basic",
			"user": "testuser",
			"pass": "testpass",
		},
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	fetchResult := result.(*FetchResult)

	if fetchResult.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", fetchResult.StatusCode)
	}

	if fetchResult.Content != "Authenticated" {
		t.Errorf("Expected content 'Authenticated', got '%s'", fetchResult.Content)
	}
}

func TestFetchTool_Execute_BearerAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token-123" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Authenticated"))
	}))
	defer server.Close()

	tool := NewFetchTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"url": server.URL,
		"auth": map[string]string{
			"type":  "bearer",
			"token": "test-token-123",
		},
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	fetchResult := result.(*FetchResult)

	if fetchResult.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", fetchResult.StatusCode)
	}
}

func TestFetchTool_Execute_APIKeyAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey != "my-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Authenticated"))
	}))
	defer server.Close()

	tool := NewFetchTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"url": server.URL,
		"auth": map[string]string{
			"type":   "apikey",
			"apikey": "my-api-key",
		},
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	fetchResult := result.(*FetchResult)

	if fetchResult.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", fetchResult.StatusCode)
	}
}

func TestFetchTool_Execute_404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	tool := NewFetchTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"url": server.URL,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err == nil {
		t.Error("Expected error for 404 response")
	}

	// Should still return result with error
	fetchResult := result.(*FetchResult)
	if fetchResult.StatusCode != 404 {
		t.Errorf("Expected status code 404, got %d", fetchResult.StatusCode)
	}
}

func TestFetchTool_Execute_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tool := NewFetchTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"url":     server.URL,
		"timeout": 1, // 1 second timeout
	}
	paramsJSON, _ := json.Marshal(params)

	_, err := tool.Execute(ctx, paramsJSON)
	if err == nil {
		t.Error("Expected timeout error")
	}
}

func TestFetchTool_Execute_ResponseHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Response", "test-value")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	tool := NewFetchTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"url": server.URL,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	fetchResult := result.(*FetchResult)

	if fetchResult.Headers["X-Custom-Response"] != "test-value" {
		t.Errorf("Expected header 'X-Custom-Response' to be 'test-value', got '%s'",
			fetchResult.Headers["X-Custom-Response"])
	}

	if fetchResult.ContentType != "application/json" {
		t.Errorf("Expected content type 'application/json', got '%s'", fetchResult.ContentType)
	}
}

func TestFetchTool_Execute_JSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success", "code": 200}`))
	}))
	defer server.Close()

	tool := NewFetchTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"url":    server.URL,
		"format": "json",
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	fetchResult := result.(*FetchResult)

	// Verify we can parse the JSON content
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(fetchResult.Content), &jsonData); err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	if jsonData["message"] != "success" {
		t.Errorf("Expected message 'success', got '%v'", jsonData["message"])
	}
}

func TestFetchTool_Execute_Duration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	tool := NewFetchTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"url": server.URL,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	fetchResult := result.(*FetchResult)

	if fetchResult.Duration <= 0 {
		t.Error("Expected duration to be greater than 0")
	}

	if fetchResult.Duration < 100 {
		t.Errorf("Expected duration >= 100ms, got %dms", fetchResult.Duration)
	}
}

func TestFetchTool_Execute_Redirect(t *testing.T) {
	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Final destination"))
	}))
	defer redirectServer.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, redirectServer.URL, http.StatusMovedPermanently)
	}))
	defer server.Close()

	tool := NewFetchTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"url": server.URL,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	fetchResult := result.(*FetchResult)

	if fetchResult.Content != "Final destination" {
		t.Errorf("Expected content 'Final destination', got '%s'", fetchResult.Content)
	}

	if fetchResult.RedirectURL == "" {
		t.Error("Expected redirect URL to be set")
	}
}

func TestFetchTool_Execute_Size(t *testing.T) {
	content := "This is a test response with some content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	}))
	defer server.Close()

	tool := NewFetchTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"url": server.URL,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	fetchResult := result.(*FetchResult)

	expectedSize := int64(len(content))
	if fetchResult.Size != expectedSize {
		t.Errorf("Expected size %d, got %d", expectedSize, fetchResult.Size)
	}
}
