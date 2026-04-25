package feishu

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

// MediaHandler handles image and file operations.
type MediaHandler struct {
	client    *Client
	config    *Config
}

// NewMediaHandler creates a new media handler.
func NewMediaHandler(client *Client, cfg *Config) *MediaHandler {
	return &MediaHandler{
		client: client,
		config: cfg,
	}
}

// UploadImage uploads an image and returns the image key.
func (mh *MediaHandler) UploadImage(ctx context.Context, data []byte, filename string) (string, error) {
	if !mh.config.EnableImages {
		return "", ErrImagesDisabled
	}

	// Check file size
	if len(data) > 10*1024*1024 { // 10MB limit
		return "", ErrImageTooLarge
	}

	// Upload to Feishu
	imageKey, err := mh.client.UploadImage(ctx, data)
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}

	return imageKey, nil
}

// UploadFile uploads a file and returns the file key.
func (mh *MediaHandler) UploadFile(ctx context.Context, data []byte, filename string) (string, error) {
	// Check file size
	if len(data) > 50*1024*1024 { // 50MB limit
		return "", ErrFileTooLarge
	}

	// Upload to Feishu
	fileKey, err := mh.client.UploadFile(ctx, data, filename)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return fileKey, nil
}

// DownloadImage downloads an image from a URL.
func (mh *MediaHandler) DownloadImage(ctx context.Context, url string) ([]byte, string, error) {
	// Validate URL
	if err := mh.validateURL(url); err != nil {
		return nil, "", err
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Read content
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	// Extract filename
	filename := mh.extractFilename(resp)

	return data, filename, nil
}

// ProcessImage processes an image for Agent use.
func (mh *MediaHandler) ProcessImage(ctx context.Context, attachment *ImageAttachment) (*ProcessedMedia, error) {
	if !mh.config.EnableImages {
		return nil, ErrImagesDisabled
	}

	// Download image if URL is provided
	var data []byte
	var err error

	if attachment.URL != "" {
		data, _, err = mh.DownloadImage(ctx, attachment.URL)
		if err != nil {
			return nil, err
		}
	} else if attachment.ImageKey != "" {
		// Image key is already uploaded to Feishu
		// In production, you'd download from Feishu
		// For now, we'll use the key directly
		return &ProcessedMedia{
			Type:     MediaTypeImage,
			Key:      attachment.ImageKey,
			URL:      attachment.URL,
			Width:    attachment.Width,
			Height:   attachment.Height,
			AltText:  attachment.AltText,
		}, nil
	}

	// Upload the downloaded image
	imageKey, err := mh.UploadImage(ctx, data, "")
	if err != nil {
		return nil, err
	}

	return &ProcessedMedia{
		Type:     MediaTypeImage,
		Key:      imageKey,
		URL:      attachment.URL,
		Width:    attachment.Width,
		Height:   attachment.Height,
		AltText:  attachment.AltText,
	}, nil
}

// ProcessFile processes a file for Agent use.
func (mh *MediaHandler) ProcessFile(ctx context.Context, attachment *FileAttachment) (*ProcessedMedia, error) {
	var data []byte
	var filename string
	var err error

	if attachment.URL != "" {
		data, filename, err = mh.DownloadImage(ctx, attachment.URL)
		if err != nil {
			return nil, err
		}
	} else {
		filename = attachment.Name
	}

	// Upload to Feishu
	fileKey, err := mh.UploadFile(ctx, data, filename)
	if err != nil {
		return nil, err
	}

	return &ProcessedMedia{
		Type:   MediaTypeFile,
		Key:    fileKey,
		URL:    attachment.URL,
		Name:   filename,
		Size:   int64(len(data)),
		MimeType: attachment.Type,
	}, nil
}

// ProcessedMedia represents processed media ready for Agent use.
type ProcessedMedia struct {
	Type     MediaType
	Key      string
	URL      string
	Name     string
	Size     int64
	MimeType string
	Width    int
	Height   int
	AltText  string
}

// validateURL validates a URL is safe to access.
func (mh *MediaHandler) validateURL(url string) error {
	// Prevent SSRF by checking scheme
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return ErrInvalidURLScheme
	}

	// Block localhost and private networks
	host := strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")
	if idx := strings.Index(host, "/"); idx != -1 {
		host = host[:idx]
	}

	// Check for localhost
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return ErrInvalidHost
	}

	// Check for private networks
	if strings.HasPrefix(host, "192.168.") ||
		strings.HasPrefix(host, "10.") ||
		strings.HasPrefix(host, "172.") {
		return ErrInvalidHost
	}

	return nil
}

// extractFilename extracts filename from HTTP response.
func (mh *MediaHandler) extractFilename(resp *http.Response) string {
	// Try Content-Disposition header
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		// Parse filename*=utf-8''filename or filename="filename"
		parts := strings.Split(cd, ";")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "filename=") {
				filename := strings.Trim(part[9:], `"`)
				return filepath.Base(filename)
			}
		}
	}

	// Try URL path
	url := resp.Request.URL.String()
	if idx := strings.LastIndex(url, "/"); idx != -1 {
		return filepath.Base(url[idx+1:])
	}

	return "file"
}

// SupportedImageTypes returns supported image MIME types.
func SupportedImageTypes() []string {
	return []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
		"image/bmp",
	}
}

// SupportedFileTypes returns supported file MIME types.
func SupportedFileTypes() []string {
	return []string{
		"application/pdf",
		"text/plain",
		"text/csv",
		"application/json",
		"application/xml",
	}
}

// IsImageType checks if MIME type is an image.
func IsImageType(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/")
}

// IsFileSupported checks if file type is supported.
func IsFileSupported(mimeType string) bool {
	supported := SupportedFileTypes()
	for _, t := range supported {
		if strings.EqualFold(mimeType, t) {
			return true
		}
	}
	return false
}

// Media errors
var (
	ErrImagesDisabled  = fmt.Errorf("image processing is disabled")
	ErrFileTooLarge   = fmt.Errorf("file too large")
	ErrImageTooLarge  = fmt.Errorf("image too large")
	ErrInvalidURLScheme = fmt.Errorf("invalid URL scheme")
	ErrInvalidHost    = fmt.Errorf("invalid host (SSRF protection)")
)