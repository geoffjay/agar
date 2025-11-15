package tools

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDownloadTool_Name(t *testing.T) {
	tool := NewDownloadTool()
	if tool.Name() != "download" {
		t.Errorf("Expected name 'download', got '%s'", tool.Name())
	}
}

func TestDownloadTool_Validate(t *testing.T) {
	tool := NewDownloadTool()

	tests := []struct {
		name    string
		params  string
		wantErr bool
	}{
		{
			name:    "valid params",
			params:  `{"url": "https://example.com/file.zip", "output_path": "/tmp/file.zip"}`,
			wantErr: false,
		},
		{
			name:    "valid with checksum",
			params:  `{"url": "https://example.com/file.zip", "output_path": "/tmp/file.zip", "checksum": "abc123", "checksum_type": "sha256"}`,
			wantErr: false,
		},
		{
			name:    "missing url",
			params:  `{"output_path": "/tmp/file.zip"}`,
			wantErr: true,
		},
		{
			name:    "missing output_path",
			params:  `{"url": "https://example.com/file.zip"}`,
			wantErr: true,
		},
		{
			name:    "invalid checksum_type",
			params:  `{"url": "https://example.com/file.zip", "output_path": "/tmp/file.zip", "checksum_type": "invalid"}`,
			wantErr: true,
		},
		{
			name:    "timeout too high",
			params:  `{"url": "https://example.com/file.zip", "output_path": "/tmp/file.zip", "timeout": 5000}`,
			wantErr: true,
		},
		{
			name:    "max_retries too high",
			params:  `{"url": "https://example.com/file.zip", "output_path": "/tmp/file.zip", "max_retries": 20}`,
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

func TestDownloadTool_Execute_SimpleDownload(t *testing.T) {
	content := []byte("Test file content for download")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "downloaded.txt")

	tool := NewDownloadTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"url":         server.URL,
		"output_path": outputPath,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	downloadResult, ok := result.(*DownloadResult)
	if !ok {
		t.Fatal("Result is not a DownloadResult")
	}

	if downloadResult.Path != outputPath {
		t.Errorf("Expected path '%s', got '%s'", outputPath, downloadResult.Path)
	}

	if downloadResult.Size != int64(len(content)) {
		t.Errorf("Expected size %d, got %d", len(content), downloadResult.Size)
	}

	// Verify file was created
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(data) != string(content) {
		t.Errorf("Content mismatch.\nExpected: %s\nGot: %s", string(content), string(data))
	}
}

func TestDownloadTool_Execute_WithMD5(t *testing.T) {
	content := []byte("Test content")

	// Calculate MD5
	hasher := md5.New()
	hasher.Write(content)
	expectedChecksum := hex.EncodeToString(hasher.Sum(nil))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "downloaded.txt")

	tool := NewDownloadTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"url":           server.URL,
		"output_path":   outputPath,
		"checksum":      expectedChecksum,
		"checksum_type": "md5",
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	downloadResult := result.(*DownloadResult)

	if downloadResult.Checksum != expectedChecksum {
		t.Errorf("Expected checksum '%s', got '%s'", expectedChecksum, downloadResult.Checksum)
	}

	if downloadResult.ChecksumType != "md5" {
		t.Errorf("Expected checksum type 'md5', got '%s'", downloadResult.ChecksumType)
	}
}

func TestDownloadTool_Execute_WithSHA256(t *testing.T) {
	content := []byte("Test content for SHA256")

	// Calculate SHA256
	hasher := sha256.New()
	hasher.Write(content)
	expectedChecksum := hex.EncodeToString(hasher.Sum(nil))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "downloaded.txt")

	tool := NewDownloadTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"url":           server.URL,
		"output_path":   outputPath,
		"checksum":      expectedChecksum,
		"checksum_type": "sha256",
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	downloadResult := result.(*DownloadResult)

	if downloadResult.Checksum != expectedChecksum {
		t.Errorf("Checksum mismatch.\nExpected: %s\nGot: %s", expectedChecksum, downloadResult.Checksum)
	}
}

func TestDownloadTool_Execute_ChecksumMismatch(t *testing.T) {
	content := []byte("Test content")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "downloaded.txt")

	tool := NewDownloadTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"url":           server.URL,
		"output_path":   outputPath,
		"checksum":      "wrongchecksum",
		"checksum_type": "md5",
	}
	paramsJSON, _ := json.Marshal(params)

	_, err := tool.Execute(ctx, paramsJSON)
	if err == nil {
		t.Error("Expected error for checksum mismatch")
	}

	// Verify file was removed on checksum failure
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Error("Expected file to be removed after checksum mismatch")
	}
}

func TestDownloadTool_Execute_Resume(t *testing.T) {
	fullContent := []byte("This is the full content of the file")
	partialSize := int64(10)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rangeHeader := r.Header.Get("Range")

		if rangeHeader != "" {
			// Resume request
			w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", partialSize, len(fullContent)-1, len(fullContent)))
			w.WriteHeader(http.StatusPartialContent)
			w.Write(fullContent[partialSize:])
		} else {
			// Full download
			w.WriteHeader(http.StatusOK)
			w.Write(fullContent)
		}
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "downloaded.txt")

	// Create partial file
	err := os.WriteFile(outputPath, fullContent[:partialSize], 0644)
	if err != nil {
		t.Fatalf("Failed to create partial file: %v", err)
	}

	tool := NewDownloadTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"url":         server.URL,
		"output_path": outputPath,
		"resume":      true,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	downloadResult := result.(*DownloadResult)

	if !downloadResult.Resumed {
		t.Error("Expected Resumed to be true")
	}

	if downloadResult.BytesResumed != partialSize {
		t.Errorf("Expected bytes resumed %d, got %d", partialSize, downloadResult.BytesResumed)
	}

	if downloadResult.Size != int64(len(fullContent)) {
		t.Errorf("Expected total size %d, got %d", len(fullContent), downloadResult.Size)
	}

	// Verify file content
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(data) != string(fullContent) {
		t.Errorf("Content mismatch.\nExpected: %s\nGot: %s", string(fullContent), string(data))
	}
}

func TestDownloadTool_Execute_CreateDirectory(t *testing.T) {
	content := []byte("Test content")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "nested", "dir", "file.txt")

	tool := NewDownloadTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"url":         server.URL,
		"output_path": outputPath,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	downloadResult := result.(*DownloadResult)

	if downloadResult.Path != outputPath {
		t.Errorf("Expected path '%s', got '%s'", outputPath, downloadResult.Path)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("File was not created")
	}
}

func TestDownloadTool_Execute_Duration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "file.txt")

	tool := NewDownloadTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"url":         server.URL,
		"output_path": outputPath,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := tool.Execute(ctx, paramsJSON)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	downloadResult := result.(*DownloadResult)

	if downloadResult.Duration <= 0 {
		t.Error("Expected duration to be greater than 0")
	}

	if downloadResult.Duration < 100 {
		t.Errorf("Expected duration >= 100ms, got %dms", downloadResult.Duration)
	}
}

func TestDownloadTool_Execute_404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "file.txt")

	tool := NewDownloadTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"url":         server.URL,
		"output_path": outputPath,
	}
	paramsJSON, _ := json.Marshal(params)

	_, err := tool.Execute(ctx, paramsJSON)
	if err == nil {
		t.Error("Expected error for 404 response")
	}
}
