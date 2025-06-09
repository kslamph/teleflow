package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// GenerateImage generates a simple image with text
func GenerateImage(text string, width int, height int) ([]byte, error) {
	// Create a new RGBA image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with gradient background
	bgColor := color.RGBA{R: 70, G: 130, B: 180, A: 255} // Steel blue
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bgColor}, image.Point{}, draw.Src)

	// Add text if provided
	if text != "" {
		// Wrap text for better display
		lines := wrapText(text, 50)                             // Wrap at 50 characters
		textColor := color.RGBA{R: 255, G: 255, B: 255, A: 255} // White

		// Calculate starting Y position to center text vertically
		lineHeight := 20
		totalHeight := len(lines) * lineHeight
		startY := (height - totalHeight) / 2

		for i, line := range lines {
			// Calculate X position to center text horizontally (rough approximation)
			charWidth := 7 // Approximate character width for basicfont.Face7x13
			textWidth := len(line) * charWidth
			startX := (width - textWidth) / 2
			if startX < 10 {
				startX = 10
			}

			point := fixed.Point26_6{
				X: fixed.Int26_6(startX * 64),
				Y: fixed.Int26_6((startY + (i+1)*lineHeight) * 64),
			}

			d := &font.Drawer{
				Dst:  img,
				Src:  image.NewUniform(textColor),
				Face: basicfont.Face7x13,
				Dot:  point,
			}
			d.DrawString(line)
		}
	}

	// Encode the image to PNG
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	return buf.Bytes(), nil
}

// wrapText wraps text at the specified width
func wrapText(text string, width int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	var lines []string
	var currentLine string

	for _, word := range words {
		// If adding this word would exceed width, start a new line
		if len(currentLine)+len(word)+1 > width && currentLine != "" {
			lines = append(lines, currentLine)
			currentLine = word
		} else {
			if currentLine == "" {
				currentLine = word
			} else {
				currentLine += " " + word
			}
		}
	}

	// Add the last line
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// GenerateMapImage generates a mock map image for location confirmation
func GenerateMapImage(address string, width int, height int) ([]byte, error) {
	// Create a new RGBA image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with map-like background (green-ish)
	bgColor := color.RGBA{R: 144, G: 238, B: 144, A: 255} // Light green
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bgColor}, image.Point{}, draw.Src)

	// Add some "roads" (simple lines)
	roadColor := color.RGBA{R: 169, G: 169, B: 169, A: 255} // Gray

	// Horizontal roads
	for y := height / 4; y < height; y += height / 3 {
		for x := 0; x < width; x++ {
			if y < height && x < width {
				img.Set(x, y, roadColor)
				if y+1 < height {
					img.Set(x, y+1, roadColor)
				}
			}
		}
	}

	// Vertical roads
	for x := width / 4; x < width; x += width / 3 {
		for y := 0; y < height; y++ {
			if x < width && y < height {
				img.Set(x, y, roadColor)
				if x+1 < width {
					img.Set(x+1, y, roadColor)
				}
			}
		}
	}

	// Add a marker (red dot) in the center
	markerColor := color.RGBA{R: 255, G: 0, B: 0, A: 255} // Red
	centerX, centerY := width/2, height/2

	// Draw a simple marker (5x5 square)
	for dx := -2; dx <= 2; dx++ {
		for dy := -2; dy <= 2; dy++ {
			x, y := centerX+dx, centerY+dy
			if x >= 0 && x < width && y >= 0 && y < height {
				img.Set(x, y, markerColor)
			}
		}
	}

	// Add text
	mapText := fmt.Sprintf("Map for: %s", address)
	lines := wrapText(mapText, 40)
	textColor := color.RGBA{R: 0, G: 0, B: 0, A: 255} // Black

	for i, line := range lines {
		point := fixed.Point26_6{
			X: fixed.Int26_6(10 * 64),
			Y: fixed.Int26_6((20 + i*15) * 64),
		}

		d := &font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(textColor),
			Face: basicfont.Face7x13,
			Dot:  point,
		}
		d.DrawString(line)
	}

	// Encode the image to PNG
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return nil, fmt.Errorf("failed to encode map image: %w", err)
	}

	return buf.Bytes(), nil
}

// GeneratePromoImage generates a promotional image for products
func GeneratePromoImage(category string, width int, height int) ([]byte, error) {
	// Create a new RGBA image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with promotional background (orange-ish)
	bgColor := color.RGBA{R: 255, G: 140, B: 0, A: 255} // Dark orange
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bgColor}, image.Point{}, draw.Src)

	// Add promotional text
	promoText := fmt.Sprintf("🔥 LATEST %s! 🔥", strings.ToUpper(category))
	lines := wrapText(promoText, 30)
	textColor := color.RGBA{R: 255, G: 255, B: 255, A: 255} // White

	// Center the text
	lineHeight := 25
	totalHeight := len(lines) * lineHeight
	startY := (height - totalHeight) / 2

	for i, line := range lines {
		charWidth := 7
		textWidth := len(line) * charWidth
		startX := (width - textWidth) / 2
		if startX < 10 {
			startX = 10
		}

		point := fixed.Point26_6{
			X: fixed.Int26_6(startX * 64),
			Y: fixed.Int26_6((startY + (i+1)*lineHeight) * 64),
		}

		d := &font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(textColor),
			Face: basicfont.Face7x13,
			Dot:  point,
		}
		d.DrawString(line)
	}

	// Encode the image to PNG
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return nil, fmt.Errorf("failed to encode promo image: %w", err)
	}

	return buf.Bytes(), nil
}
