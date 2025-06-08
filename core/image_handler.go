package teleflow

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type processedImage struct {
	data []byte

	isBase64 bool

	filePath string
}

type imageHandler struct{}

func newImageHandler() *imageHandler {
	return &imageHandler{}
}

func (ih *imageHandler) processImage(imageSpec ImageSpec, ctx *Context) (*processedImage, error) {
	if imageSpec == nil {
		return nil, nil
	}

	switch img := imageSpec.(type) {
	case string:

		return ih.processStaticImage(img)

	case func(*Context) string:

		imagePath := img(ctx)
		if imagePath == "" {
			return nil, nil
		}
		return ih.processStaticImage(imagePath)

	case []byte:

		return ih.processRawBytes(img)

	case func(*Context) []byte:

		imageBytes := img(ctx)
		if len(imageBytes) == 0 {
			return nil, nil
		}
		return ih.processRawBytes(imageBytes)

	default:
		return nil, fmt.Errorf("unsupported image type: %T (expected string, []byte, func(*Context) string, or func(*Context) []byte)", img)
	}
}

func (ih *imageHandler) processStaticImage(imageStr string) (*processedImage, error) {

	if strings.HasPrefix(imageStr, "data:image/") {
		return ih.processBase64Image(imageStr)
	}

	if _, err := os.Stat(imageStr); err == nil {
		return ih.processFileImage(imageStr)
	}

	return ih.processURLImage(imageStr)
}

func (ih *imageHandler) processBase64Image(base64Str string) (*processedImage, error) {

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

func (ih *imageHandler) processFileImage(filePath string) (*processedImage, error) {

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("image file not found: %s", filePath)
	}

	if info.Size() > 50*1024*1024 {
		return nil, fmt.Errorf("image file too large: %d bytes (max 50MB)", info.Size())
	}

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

func (ih *imageHandler) processURLImage(imageURL string) (*processedImage, error) {

	return &processedImage{
		filePath: imageURL,
		isBase64: false,
	}, nil
}

func (ih *imageHandler) processRawBytes(imageBytes []byte) (*processedImage, error) {

	if len(imageBytes) == 0 {
		return nil, nil
	}

	if len(imageBytes) > 50*1024*1024 {
		return nil, fmt.Errorf("image data too large: %d bytes (max 50MB)", len(imageBytes))
	}

	return &processedImage{
		data:     imageBytes,
		isBase64: false,
	}, nil
}
