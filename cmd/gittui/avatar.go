package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"

	"github.com/TheZoraiz/ascii-image-converter/aic_package"
	"github.com/charmbracelet/x/mosaic"
)

// FetchAvatarASCII fetches a GitHub avatar and converts it to colored ASCII art
// using the configured renderer (mosaic or ascii-image-converter)
func FetchAvatarASCII(avatarURL string, width int) (string, error) {
	config, _ := LoadConfig()
	renderer := config.AvatarRenderer
	if renderer == "" || renderer == "mosaic" {
		return fetchAvatarMosaic(avatarURL, width)
	}
	return fetchAvatarAsciiConverter(avatarURL, width)
}

// fetchAvatarMosaic renders avatar using charmbracelet/mosaic
func fetchAvatarMosaic(avatarURL string, width int) (string, error) {
	// 1. Fetch avatar image
	resp, err := http.Get(avatarURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch avatar: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch avatar: status %d", resp.StatusCode)
	}

	// 2. Decode image directly from response body
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// 3. Convert to terminal art using charmbracelet mosaic with hard-coded settings
	// Height is proportional to width with 2:1 ratio (terminal chars are taller than wide)
	height := width / 2

	// Create mosaic renderer with hard-coded default settings
	renderer := mosaic.New().
		Width(width).
		Height(height).
		Symbol(mosaic.Half).
		Dither(true).
		Threshold(128)

	rendered := renderer.Render(img)

	return rendered, nil
}

// fetchAvatarAsciiConverter renders avatar using ascii-image-converter
func fetchAvatarAsciiConverter(avatarURL string, width int) (string, error) {
	// Configure flags for ascii-image-converter
	// Use same 2:1 ratio as mosaic for consistent sizing
	height := width / 2

	flags := aic_package.DefaultFlags()
	flags.Dimensions = []int{width, height}  // Set both dimensions to match mosaic
	flags.Colored = true                     // Use colors from original image
	flags.Complex = true                     // Use more characters for better detail
	flags.Negative = false                   // Don't invert colors
	flags.Braille = false                    // Use ASCII characters, not braille

	// Convert the image URL to ASCII art
	asciiArt, err := aic_package.Convert(avatarURL, flags)
	if err != nil {
		return "", fmt.Errorf("failed to convert avatar: %w", err)
	}

	return asciiArt, nil
}
