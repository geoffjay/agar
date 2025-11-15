package tools

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// DownloadTool implements file download functionality
type DownloadTool struct {
	client *http.Client
}

// DownloadParams defines the parameters for the Download tool
type DownloadParams struct {
	URL          string `json:"url"`
	OutputPath   string `json:"output_path"`
	Resume       bool   `json:"resume,omitempty"`
	Checksum     string `json:"checksum,omitempty"`      // Expected checksum
	ChecksumType string `json:"checksum_type,omitempty"` // "md5", "sha256"
	Timeout      int    `json:"timeout,omitempty"`       // seconds
	MaxRetries   int    `json:"max_retries,omitempty"`
	ChunkSize    int64  `json:"chunk_size,omitempty"`    // bytes
}

// DownloadResult represents the result of a download operation
type DownloadResult struct {
	Path         string `json:"path"`
	Size         int64  `json:"size"`
	Checksum     string `json:"checksum"`
	ChecksumType string `json:"checksum_type"`
	Duration     int64  `json:"duration_ms"`
	Resumed      bool   `json:"resumed,omitempty"`
	BytesResumed int64  `json:"bytes_resumed,omitempty"`
}

// NewDownloadTool creates a new Download tool instance
func NewDownloadTool() *DownloadTool {
	return &DownloadTool{
		client: &http.Client{
			Timeout: 300 * time.Second, // 5 minutes default
		},
	}
}

// Name returns the tool's name
func (t *DownloadTool) Name() string {
	return "download"
}

// Description returns the tool's description
func (t *DownloadTool) Description() string {
	return "Download files from URLs with resume support, integrity verification, and progress tracking"
}

// Schema returns the JSON schema for the tool's parameters
func (t *DownloadTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "URL to download from",
			},
			"output_path": map[string]interface{}{
				"type":        "string",
				"description": "Path where the file should be saved",
			},
			"resume": map[string]interface{}{
				"type":        "boolean",
				"description": "Resume interrupted download if possible",
			},
			"checksum": map[string]interface{}{
				"type":        "string",
				"description": "Expected checksum for integrity verification",
			},
			"checksum_type": map[string]interface{}{
				"type":        "string",
				"description": "Checksum algorithm: md5 or sha256",
				"enum":        []string{"md5", "sha256"},
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Timeout in seconds (default: 300, max: 3600)",
				"minimum":     1,
				"maximum":     3600,
			},
			"max_retries": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of retries (default: 3, max: 10)",
				"minimum":     0,
				"maximum":     10,
			},
		},
		"required": []string{"url", "output_path"},
	}
}

// Validate checks if the parameters are valid
func (t *DownloadTool) Validate(params json.RawMessage) error {
	var p DownloadParams
	if err := json.Unmarshal(params, &p); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	if p.URL == "" {
		return fmt.Errorf("url is required")
	}

	if p.OutputPath == "" {
		return fmt.Errorf("output_path is required")
	}

	if p.ChecksumType != "" && p.ChecksumType != "md5" && p.ChecksumType != "sha256" {
		return fmt.Errorf("checksum_type must be 'md5' or 'sha256'")
	}

	if p.Timeout < 0 || p.Timeout > 3600 {
		return fmt.Errorf("timeout must be between 1 and 3600 seconds")
	}

	if p.MaxRetries < 0 || p.MaxRetries > 10 {
		return fmt.Errorf("max_retries must be between 0 and 10")
	}

	return nil
}

// Execute runs the tool with the given parameters
func (t *DownloadTool) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var p DownloadParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Default values
	if p.Timeout == 0 {
		p.Timeout = 300
	}
	if p.MaxRetries == 0 {
		p.MaxRetries = 3
	}
	if p.ChunkSize == 0 {
		p.ChunkSize = 1024 * 1024 // 1MB chunks
	}

	// Update client timeout
	t.client.Timeout = time.Duration(p.Timeout) * time.Second

	// Ensure output directory exists
	dir := filepath.Dir(p.OutputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Execute with retries
	maxAttempts := p.MaxRetries + 1
	var lastErr error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		result, err := t.downloadFile(ctx, p)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Don't retry on context errors
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

// downloadFile performs the actual download
func (t *DownloadTool) downloadFile(ctx context.Context, p DownloadParams) (*DownloadResult, error) {
	startTime := time.Now()

	// Check if file exists for resume
	var resumeFrom int64
	var resumed bool
	if p.Resume {
		if info, err := os.Stat(p.OutputPath); err == nil {
			resumeFrom = info.Size()
			resumed = true
		}
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", p.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set range header for resume
	if resumeFrom > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", resumeFrom))
	}

	// Execute request
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	// Verify resume was accepted
	if resumeFrom > 0 && resp.StatusCode != http.StatusPartialContent {
		// Server doesn't support resume, start from beginning
		resumeFrom = 0
		resumed = false
	}

	// Open output file
	var file *os.File
	if resumeFrom > 0 {
		file, err = os.OpenFile(p.OutputPath, os.O_APPEND|os.O_WRONLY, 0644)
	} else {
		file, err = os.Create(p.OutputPath)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open output file: %w", err)
	}
	defer file.Close()

	// Create hasher if checksum verification is needed
	var hasher hash.Hash
	if p.ChecksumType != "" {
		switch p.ChecksumType {
		case "md5":
			hasher = md5.New()
		case "sha256":
			hasher = sha256.New()
		}

		// If resuming, we can't verify the full checksum
		if resumed {
			hasher = nil
		}
	}

	// Download with streaming
	var written int64
	if hasher != nil {
		// Write to both file and hasher
		multiWriter := io.MultiWriter(file, hasher)
		written, err = io.Copy(multiWriter, resp.Body)
	} else {
		written, err = io.Copy(file, resp.Body)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	duration := time.Since(startTime)

	// Calculate final size
	totalSize := resumeFrom + written

	// Verify checksum if provided
	var calculatedChecksum string
	if hasher != nil {
		calculatedChecksum = hex.EncodeToString(hasher.Sum(nil))
		if p.Checksum != "" && calculatedChecksum != p.Checksum {
			// Remove the downloaded file on checksum mismatch
			os.Remove(p.OutputPath)
			return nil, fmt.Errorf("checksum mismatch: expected %s, got %s", p.Checksum, calculatedChecksum)
		}
	}

	result := &DownloadResult{
		Path:         p.OutputPath,
		Size:         totalSize,
		Checksum:     calculatedChecksum,
		ChecksumType: p.ChecksumType,
		Duration:     duration.Milliseconds(),
		Resumed:      resumed,
	}

	if resumed {
		result.BytesResumed = resumeFrom
	}

	return result, nil
}
