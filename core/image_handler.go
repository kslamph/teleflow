package teleflow

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// processedImage contains processed image data ready for sending
type processedImage struct {
	data     []byte
	isBase64 bool
	filePath string
}

// imageHandler processes ImageSpec into ProcessedImage
type imageHandler struct{}

// newImageHandler creates a new image handler
func newImageHandler() *imageHandler {
	return &imageHandler{}
}

// processImage processes an ImageSpec and returns ProcessedImage or nil if no image
func (ih *imageHandler) processImage(imageSpec ImageSpec, ctx *Context) (*processedImage, error) {
	if imageSpec == nil {
		return nil, nil // No image specified
	}

	switch img := imageSpec.(type) {
	case string:
		// Static image path/URL/base64
		return ih.processStaticImage(img)

	case func(*Context) string:
		// Dynamic image function
		imagePath := img(ctx)
		if imagePath == "" {
			return nil, nil // Function returned empty, no image
		}
		return ih.processStaticImage(imagePath)

	default:
		return nil, fmt.Errorf("unsupported image type: %T (expected string or func(*Context) string)", img)
	}
}

// processStaticImage processes a static image string (path, URL, or base64)
func (ih *imageHandler) processStaticImage(imageStr string) (*processedImage, error) {
	// Check if it's base64 encoded
	if strings.HasPrefix(imageStr, "data:image/") {
		return ih.processBase64Image(imageStr)
	}

	// Check if it's a valid file path
	if _, err := os.Stat(imageStr); err == nil {
		return ih.processFileImage(imageStr)
	}

	// Assume it's a URL or external reference
	return ih.processURLImage(imageStr)
}

// processBase64Image processes base64 encoded image data
func (ih *imageHandler) processBase64Image(base64Str string) (*processedImage, error) {
	// Extract base64 data (skip data:image/type;base64, prefix)
	parts := strings.Split(base64Str, ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid base64 image format: expected 'data:image/type;base64,data'")
	}

	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 image: %w", err)
	}

	return &processedImage{
		data:     data,
		isBase64: true,
	}, nil
}

// processFileImage processes a local file path
func (ih *imageHandler) processFileImage(filePath string) (*processedImage, error) {
	// Validate file exists
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("image file not found: %s", filePath)
	}

	// Check file size (Telegram limit is 50MB)
	if info.Size() > 50*1024*1024 { // 50MB
		return nil, fmt.Errorf("image file too large: %d bytes (max 50MB)", info.Size())
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	validExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}
	isValid := false
	for _, validExt := range validExts {
		if ext == validExt {
			isValid = true
			break
		}
	}

	if !isValid {
		return nil, fmt.Errorf("unsupported image format: %s (supported: %v)", ext, validExts)
	}

	// Read file data
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read image file: %w", err)
	}

	return &processedImage{
		data:     data,
		isBase64: false,
		filePath: filePath,
	}, nil
}

// processURLImage processes a URL or external image reference
func (ih *imageHandler) processURLImage(imageURL string) (*processedImage, error) {
	// For URLs, we'll store the path and let Telegram handle it
	// This is a simplified implementation - in production you might want to
	// validate URLs, download and cache images, etc.

	return &processedImage{
		filePath: imageURL,
		isBase64: false,
	}, nil
}