package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func generateSimpleImage(width, height int, bgColor color.Color, text string, textColor color.Color) ([]byte, error) {
	// 1. Create a new RGBA image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 2. Fill the background
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bgColor}, image.Point{}, draw.Src)

	// 3. Add text (optional)
	if text != "" {
		// Position for the text (adjust as needed)
		// Point is where the dot (baseline start) of the text will be.
		point := fixed.Point26_6{
			X: fixed.Int26_6(10 * 64), // 10 pixels from left
			Y: fixed.Int26_6(30 * 64), // 30 pixels from top (baseline)
		}

		d := &font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(textColor),
			Face: basicfont.Face7x13, // A simple built-in font
			Dot:  point,
		}
		d.DrawString(text)
	}

	// 4. Encode the image to PNG (or JPEG, etc.) into a buffer
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	// For JPEG:
	// err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 75}) // 0-100 quality
	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	return buf.Bytes(), nil
}
