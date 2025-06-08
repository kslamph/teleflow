// Package teleflow provides a Telegram bot framework for building conversational flows.
// This file contains the imageHandler component which processes ImageSpec specifications
// into formats suitable for sending through the Telegram Bot API.
package teleflow

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// processedImage contains processed image data ready for sending to Telegram.
// The structure supports multiple image sources and formats:
//   - data: Raw image bytes for in-memory images or decoded base64
//   - isBase64: Flag indicating if the source was base64 encoded
//   - filePath: File path or URL for file-based or remote images
//
// The Telegram Bot API supports different ways to send images, and this structure
// provides the flexibility to handle all supported methods.
type processedImage struct {
	// data contains raw image bytes for images loaded into memory
	data []byte
	// isBase64 indicates whether the original source was base64 encoded
	isBase64 bool
	// filePath contains the file path or URL for file-based or remote images
	filePath string
}

// imageHandler processes ImageSpec specifications into processedImage instances
// ready for transmission to Telegram. It supports multiple image sources:
//   - Static strings: File paths, URLs, or base64 data URIs
//   - Dynamic functions: Functions that return image specifications based on context
//
// The handler automatically detects image types and formats, validates file sizes
// and extensions, and prepares images in the format most suitable for the
// Telegram Bot API.
type imageHandler struct{}

// newImageHandler creates and initializes a new image handler.
// The handler is stateless and ready to process ImageSpec instances immediately.
//
// Returns a configured imageHandler instance.
func newImageHandler() *imageHandler {
	return &imageHandler{}
}

// processImage processes an ImageSpec and returns a processedImage ready for sending
// to Telegram, or nil if no image is specified.
//
// Supported ImageSpec Types:
//   - string: Static image reference (file path, URL, or base64 data URI)
//   - func(*Context) string: Dynamic function that returns image reference based on context
//   - []byte: Raw image data bytes (e.g., from dynamically generated images like QR codes)
//   - func(*Context) []byte: Dynamic function that returns raw image bytes based on context
//
// Image Source Detection:
//   - Base64 data URIs: Detected by "data:image/" prefix, decoded to bytes
//   - Local files: Detected by file existence check, validated and loaded
//   - URLs/Remote: Everything else treated as URL for Telegram to fetch
//   - Raw bytes: Direct byte data processed for transmission
//
// Validation and Processing:
//   - File size limits: Enforces Telegram's 50MB limit for local files and raw bytes
//   - Format validation: Checks file extensions for supported image formats
//   - Error handling: Returns descriptive errors for invalid or inaccessible images
//
// Parameters:
//   - imageSpec: The image specification to process (string, []byte, or function)
//   - ctx: Context for dynamic image functions (unused for static types)
//
// Returns:
//   - *processedImage: Processed image ready for Telegram API, or nil if no image
//   - error: Any error that occurred during processing or validation
func (ih *imageHandler) processImage(imageSpec ImageSpec, ctx *Context) (*processedImage, error) {
	if imageSpec == nil {
		return nil, nil // No image specified
	}

	switch img := imageSpec.(type) {
	case string:
		// Static image path/URL/base64 - process directly
		return ih.processStaticImage(img)

	case func(*Context) string:
		// Dynamic image function - evaluate with context first
		imagePath := img(ctx)
		if imagePath == "" {
			return nil, nil // Function returned empty, no image to process
		}
		return ih.processStaticImage(imagePath)

	case []byte:
		// Raw image bytes - process directly
		return ih.processRawBytes(img)

	case func(*Context) []byte:
		// Dynamic raw bytes function - evaluate with context first
		imageBytes := img(ctx)
		if len(imageBytes) == 0 {
			return nil, nil // Function returned empty, no image to process
		}
		return ih.processRawBytes(imageBytes)

	default:
		return nil, fmt.Errorf("unsupported image type: %T (expected string, []byte, func(*Context) string, or func(*Context) []byte)", img)
	}
}

// processStaticImage processes a static image string by detecting its type
// and routing it to the appropriate specialized processing method.
//
// Image Type Detection:
//   - Base64 data URIs: Identified by "data:image/" prefix
//   - Local files: Identified by successful file existence check
//   - URLs/Remote: Everything else (assumed to be URLs or external references)
//
// Processing is delegated to specialized methods that handle validation,
// format checking, and preparation for the specific image source type.
//
// Parameters:
//   - imageStr: Static image string (file path, URL, or base64 data URI)
//
// Returns processed image ready for Telegram API or error if processing fails.
func (ih *imageHandler) processStaticImage(imageStr string) (*processedImage, error) {
	// Check if it's base64 encoded data URI
	if strings.HasPrefix(imageStr, "data:image/") {
		return ih.processBase64Image(imageStr)
	}

	// Check if it's a valid local file path
	if _, err := os.Stat(imageStr); err == nil {
		return ih.processFileImage(imageStr)
	}

	// Assume it's a URL or external reference
	return ih.processURLImage(imageStr)
}

// processBase64Image processes base64 encoded image data URIs by extracting
// and decoding the base64 data into raw bytes suitable for Telegram transmission.
//
// Expected Format: "data:image/type;base64,encodedData"
//   - Validates the data URI format has exactly two comma-separated parts
//   - Extracts the base64 encoded data (after the comma)
//   - Decodes the base64 string into raw image bytes
//   - Marks the result as base64-sourced for tracking purposes
//
// Base64 images are particularly useful for:
//   - Generated images from in-memory processing
//   - Images received from APIs or external services
//   - Embedded images in templates or configurations
//
// Parameters:
//   - base64Str: Complete base64 data URI string
//
// Returns processedImage with decoded bytes or error if format/decoding fails.
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

// processFileImage processes a local file path by validating the file and loading
// its contents into memory for transmission to Telegram.
//
// Validation Steps:
//   - Confirms file exists and is accessible
//   - Enforces Telegram's 50MB file size limit
//   - Validates file extension against supported image formats
//   - Loads file contents into memory for sending
//
// Supported Image Formats:
//   - JPEG (.jpg, .jpeg): Most common format, good compression
//   - PNG (.png): Lossless compression, supports transparency
//   - GIF (.gif): Animated images and simple graphics
//   - BMP (.bmp): Uncompressed bitmap format
//   - WebP (.webp): Modern format with excellent compression
//
// The method loads the entire file into memory, which is suitable for most
// use cases but may not be optimal for very large images. Consider implementing
// streaming for production systems with large image processing requirements.
//
// Parameters:
//   - filePath: Path to the local image file to process
//
// Returns processedImage with file data loaded or error if validation/loading fails.
func (ih *imageHandler) processFileImage(filePath string) (*processedImage, error) {
	// Validate file exists and get file info
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("image file not found: %s", filePath)
	}

	// Check file size against Telegram's 50MB limit
	if info.Size() > 50*1024*1024 { // 50MB
		return nil, fmt.Errorf("image file too large: %d bytes (max 50MB)", info.Size())
	}

	// Validate file extension against supported formats
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

	// Read file data into memory
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

// processURLImage processes a URL or external image reference by storing the URL
// for Telegram to fetch directly. This method implements a pass-through strategy
// where the image URL is provided to Telegram's API for remote fetching.
//
// URL Processing Strategy:
//   - No local validation or downloading is performed
//   - URL is passed directly to Telegram Bot API
//   - Telegram handles fetching, validation, and caching
//   - Reduces local bandwidth and storage requirements
//   - Relies on Telegram's robust image processing infrastructure
//
// Advantages of URL Pass-Through:
//   - Minimal local resource usage
//   - Leverages Telegram's CDN and caching
//   - Handles dynamic URLs and redirects automatically
//   - No local storage or cleanup required
//
// Considerations:
//   - URL must be accessible to Telegram's servers
//   - No local validation of image format or size
//   - Network issues may cause delivery failures
//   - Production systems may want additional URL validation
//
// Parameters:
//   - imageURL: URL or external reference to the image
//
// Returns processedImage configured for URL-based delivery.
func (ih *imageHandler) processURLImage(imageURL string) (*processedImage, error) {
	// Store URL for Telegram to fetch directly - no local processing needed
	return &processedImage{
		filePath: imageURL,
		isBase64: false,
	}, nil
}

// processRawBytes processes raw image byte data by validating size limits
// and preparing the bytes for transmission to Telegram.
//
// Raw Bytes Processing:
//   - Validates byte array is not empty
//   - Enforces Telegram's 50MB file size limit
//   - No format validation (assumes caller provides valid image data)
//   - Stores bytes directly for transmission
//
// This method is particularly useful for:
//   - Dynamically generated images (QR codes, charts, etc.)
//   - Images processed or modified in memory
//   - Images received from external APIs as byte arrays
//   - Custom image generation workflows
//
// Size Validation:
//   - Checks against Telegram's 50MB limit for photo uploads
//   - Returns descriptive error if size exceeds limit
//   - Empty byte arrays are treated as no image
//
// Format Considerations:
//   - No automatic format detection or validation
//   - Caller is responsible for providing valid image format
//   - Supported formats depend on Telegram's capabilities (JPEG, PNG, GIF, WebP, etc.)
//   - Invalid formats will be rejected by Telegram API during send
//
// Parameters:
//   - imageBytes: Raw image data as byte array
//
// Returns processedImage with raw bytes ready for transmission or error if validation fails.
func (ih *imageHandler) processRawBytes(imageBytes []byte) (*processedImage, error) {
	// Validate byte array is not empty
	if len(imageBytes) == 0 {
		return nil, nil // Empty bytes, no image to process
	}

	// Check size against Telegram's 50MB limit
	if len(imageBytes) > 50*1024*1024 { // 50MB
		return nil, fmt.Errorf("image data too large: %d bytes (max 50MB)", len(imageBytes))
	}

	return &processedImage{
		data:     imageBytes,
		isBase64: false,
	}, nil
}
