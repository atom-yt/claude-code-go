package paste

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectImage_FileURIScheme(t *testing.T) {
	// Create a temporary test image
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "test.png")

	// Create a minimal PNG file
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, // IHDR chunk length
		0x49, 0x48, 0x44, 0x52, // IHDR chunk type
		0x00, 0x00, 0x00, 0x01, // width: 1
		0x00, 0x00, 0x00, 0x01, // height: 1
		0x08, 0x02, 0x00, 0x00, 0x00, // bit depth, color type, compression, filter, interlace
		0x90, 0x77, 0x53, 0xDE, // CRC
		0x00, 0x00, 0x00, 0x0C, // IDAT chunk length
		0x49, 0x44, 0x54, 0x41, 0x54, // IDAT chunk type
		0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00, 0x00, 0x03, 0x00, 0x01, 0x00, // IDAT data
		0x18, 0xDD, 0x8D, 0xB4, // CRC
		0x00, 0x00, 0x00, 0x00, // IEND chunk length
		0x49, 0x45, 0x4E, 0x44, // IEND chunk type
		0xAE, 0x42, 0x60, 0x82, // CRC
	}
	if err := os.WriteFile(imgPath, pngData, 0644); err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}

	tests := []struct {
		name string
		input string
		want bool
	}{
		{"file:// URI", "file://" + imgPath, true},
		{"file:// with localhost", "file://localhost" + imgPath, true},
		{"direct path", imgPath, true},
		{"non-existent file", "/path/to/nonexistent.png", false},
		{"text input", "hello world", false},
		{"non-image extension", "/path/to/file.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := DetectImage(tt.input)
			if tt.want && img == nil {
				t.Errorf("DetectImage() returned nil, want non-nil")
			}
			if !tt.want && img != nil {
				t.Errorf("DetectImage() returned %v, want nil", img)
			}
			if img != nil {
				if img.Filename != "test.png" {
					t.Errorf("Image.Filename = %q, want test.png", img.Filename)
				}
				if img.MIME != "image/png" {
					t.Errorf("Image.MIME = %q, want image/png", img.MIME)
				}
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name string
		b    int64
		want string
	}{
		{"bytes", 100, "100 B"},
		{"kilobytes", 1536, "1.5 KB"},
		{"megabytes", 1572864, "1.5 MB"},
		{"gigabytes", 1610612736, "1.5 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatBytes(tt.b); got != tt.want {
				t.Errorf("FormatBytes() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsImagePath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	pngFile := filepath.Join(tmpDir, "test.png")
	jpgFile := filepath.Join(tmpDir, "test.jpg")
	txtFile := filepath.Join(tmpDir, "test.txt")
	dirFile := filepath.Join(tmpDir, "testdir")

	os.WriteFile(pngFile, []byte{}, 0644)
	os.WriteFile(jpgFile, []byte{}, 0644)
	os.WriteFile(txtFile, []byte{}, 0644)
	os.Mkdir(dirFile, 0755)

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"png file", pngFile, true},
		{"jpg file", jpgFile, true},
		{"txt file", txtFile, false},
		{"directory", dirFile, false},
		{"non-existent", "/nonexistent.png", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// isImagePath is not exported, so we test via DetectImage
			img := DetectImage(tt.path)
			if tt.want && img == nil {
				t.Errorf("DetectImage(%q) = nil, want non-nil", tt.path)
			}
			if !tt.want && img != nil {
				t.Errorf("DetectImage(%q) = %v, want nil", tt.path, img)
			}
		})
	}
}

func TestImage_FormatAsAttachment(t *testing.T) {
	img := &Image{
		Filename: "test.png",
		Data:     []byte{0x89, 0x50, 0x4E, 0x47},
		MIME:     "image/png",
		Size:     4,
	}

	got := img.FormatAsAttachment()
	if got == "" {
		t.Error("FormatAsAttachment() returned empty string")
	}

	wantPrefix := "[Image: test.png,"
	if !contains(got, wantPrefix) {
		t.Errorf("FormatAsAttachment() = %q, want to contain %q", got, wantPrefix)
	}
}

func TestImage_FormatAsMarkdown(t *testing.T) {
	tmpDir := t.TempDir()

	img := &Image{
		Filename: "test.png",
		Data:     []byte{0x89, 0x50, 0x4E, 0x47},
		MIME:     "image/png",
		Size:     4,
	}

	got := img.FormatAsMarkdown(tmpDir)
	if got == "" {
		t.Error("FormatAsMarkdown() returned empty string")
	}

	// Check if image was saved
	dstPath := filepath.Join(tmpDir, "test.png")
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Errorf("FormatAsMarkdown() did not save image to %s", dstPath)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}