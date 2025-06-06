package teleflow

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ImageType represents the type of image for Telegram
type ImageType int

const (
	ImageTypePhoto ImageType = iota
	ImageTypeDocument
)

// ProcessedImage contains processed image data ready for sending
type ProcessedImage struct {
	Data     []byte
	Type     ImageType
	IsBase64 bool
	FilePath string
}

// ImageHandler processes ImageSpec into ProcessedImage
type ImageHandler struct{}

// NewImageHandler creates a new image handler
func NewImageHandler() *ImageHandler {
	return &ImageHandler{}
}

// ProcessImage processes an ImageSpec and returns ProcessedImage or nil if no image
func (ih *ImageHandler) ProcessImage(imageSpec ImageSpec, ctx *Context) (*ProcessedImage, error) {
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
func (ih *ImageHandler) processStaticImage(imageStr string) (*ProcessedImage, error) {
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
func (ih *ImageHandler) processBase64Image(base64Str string) (*ProcessedImage, error) {
	// Extract base64 data (skip data:image/type;base64, prefix)
	parts := strings.Split(base64Str, ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid base64 image format: expected 'data:image/type;base64,data'")
	}

	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 image: %w", err)
	}

	return &ProcessedImage{
		Data:     data,
		Type:     ImageTypePhoto,
		IsBase64: true,
	}, nil
}

// processFileImage processes a local file path
func (ih *ImageHandler) processFileImage(filePath string) (*ProcessedImage, error) {
	// Validate file exists
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("image file not found: %s", filePath)
	}

	// Check file size (Telegram limit is 50MB)
	if info.Size() > 50*1024*1024 {
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

	return &ProcessedImage{
		Data:     data,
		Type:     ImageTypePhoto,
		IsBase64: false,
		FilePath: filePath,
	}, nil
}

// processURLImage processes a URL or external image reference
func (ih *ImageHandler) processURLImage(imageURL string) (*ProcessedImage, error) {
	// For URLs, we'll store the path and let Telegram handle it
	// This is a simplified implementation - in production you might want to
	// validate URLs, download and cache images, etc.

	return &ProcessedImage{
		FilePath: imageURL,
		Type:     ImageTypePhoto,
		IsBase64: false,
	}, nil
}
