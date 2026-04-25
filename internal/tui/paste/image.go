package paste

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Image represents a pasted image.
type Image struct {
	Filename string
	Data     []byte
	MIME     string
	Size     int64
}

// DetectImage checks if input is an image paste.
// Supports file:// URIs and direct file paths with image extensions.
func DetectImage(input string) *Image {
	// Check for file:// URI
	input = strings.TrimSpace(input)
	if strings.HasPrefix(input, "file://") {
		uri := strings.TrimPrefix(input, "file://")
		// Remove hostname if present (file://localhost/path)
		if parts := strings.SplitN(uri, "/", 2); len(parts) == 2 {
			uri = "/" + parts[1]
		}
		return loadImage(uri)
	}

	// Check for direct file path with image extension
	if isImagePath(input) {
		return loadImage(input)
	}

	return nil
}

// isImagePath checks if path looks like an image file.
func isImagePath(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp", ".tiff":
		// Also check if file exists
		info, err := os.Stat(path)
		return err == nil && !info.IsDir()
	}
	return false
}

// loadImage reads an image file and returns Image struct.
func loadImage(path string) *Image {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	// Detect MIME type based on extension
	ext := strings.ToLower(filepath.Ext(path))
	mime := "image/png"
	switch ext {
	case ".jpg", ".jpeg":
		mime = "image/jpeg"
	case ".gif":
		mime = "image/gif"
	case ".webp":
		mime = "image/webp"
	case ".bmp":
		mime = "image/bmp"
	case ".tiff":
		mime = "image/tiff"
	}

	// Get file info for size
	info, _ := os.Stat(path)

	return &Image{
		Filename: filepath.Base(path),
		Data:     data,
		MIME:     mime,
		Size:     info.Size(),
	}
}

// FormatAsAttachment formats an image as an attachment message.
func (img *Image) FormatAsAttachment() string {
	sizeStr := FormatBytes(img.Size)
	return fmt.Sprintf("[Image: %s, %s, base64: %d chars]",
		img.Filename, sizeStr, len(base64.StdEncoding.EncodeToString(img.Data)))
}

// FormatAsMarkdown formats an image as a markdown image reference.
func (img *Image) FormatAsMarkdown(dir string) string {
	// Save image to session directory for markdown reference
	dstPath := filepath.Join(dir, img.Filename)
	if err := os.WriteFile(dstPath, img.Data, 0644); err != nil {
		// Fall back to inline message if save fails
		return img.FormatAsAttachment()
	}
	return fmt.Sprintf("![%s](%s)", img.Filename, img.Filename)
}

// FormatBytes formats byte size for display.
func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}