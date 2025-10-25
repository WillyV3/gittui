package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strconv"
	"strings"
)

// ColorizedBrailleRenderer handles theme-aware avatar colorization
type ColorizedBrailleRenderer struct {
	theme      Theme
	themeCache []color.RGBA // Parsed theme colors for fast lookup
}

// NewColorizedBrailleRenderer creates a renderer with the current theme
func NewColorizedBrailleRenderer(theme Theme) *ColorizedBrailleRenderer {
	return &ColorizedBrailleRenderer{
		theme:      theme,
		themeCache: parseThemeColors(theme),
	}
}

// parseThemeColors converts theme hex colors to RGBA for distance calculations
func parseThemeColors(theme Theme) []color.RGBA {
	hexColors := []string{
		theme.Blue,
		theme.Green,
		theme.Red,
		theme.Yellow,
		theme.Purple,
		theme.Foreground,
		theme.ContribHigher,
		theme.ContribHigh,
		theme.ContribMed,
		theme.ContribLow,
	}

	var colors []color.RGBA
	for _, hex := range hexColors {
		if c := hexToRGBA(hex); c != nil {
			colors = append(colors, *c)
		}
	}
	return colors
}

// hexToRGBA converts hex color string to RGBA
func hexToRGBA(hex string) *color.RGBA {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return nil
	}

	r, err1 := strconv.ParseInt(hex[0:2], 16, 64)
	g, err2 := strconv.ParseInt(hex[2:4], 16, 64)
	b, err3 := strconv.ParseInt(hex[4:6], 16, 64)

	if err1 != nil || err2 != nil || err3 != nil {
		return nil
	}

	return &color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: 255,
	}
}

// RenderColorized generates colorized braille art from an avatar image
func (r *ColorizedBrailleRenderer) RenderColorized(img image.Image) string {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var output strings.Builder

	// Braille patterns: 2 pixels wide, 4 pixels tall
	for y := 0; y < height; y += 4 {
		for x := 0; x < width; x += 2 {
			// Get dominant color for this braille block
			dominantColor := r.getDominantColor(img, x, y)

			// Map to nearest theme color
			themeColor := r.mapToThemeColor(dominantColor)

			// Get braille pattern (from dotmatrix logic)
			brailleChar := r.getBraillePattern(img, x, y)

			// Skip empty braille
			if brailleChar == '⠀' { // Empty braille
				output.WriteString(" ")
				continue
			}

			// Write with ANSI 24-bit color
			output.WriteString(fmt.Sprintf("\033[38;2;%d;%d;%dm%c\033[0m",
				themeColor.R, themeColor.G, themeColor.B, brailleChar))
		}
		output.WriteString("\n")
	}

	return output.String()
}

// getDominantColor finds the most common non-background color in a 2x4 block
func (r *ColorizedBrailleRenderer) getDominantColor(img image.Image, startX, startY int) color.RGBA {
	bounds := img.Bounds()

	var rSum, gSum, bSum int
	var count int

	// Sample 2x4 block
	for y := startY; y < startY+4 && y < bounds.Max.Y; y++ {
		for x := startX; x < startX+2 && x < bounds.Max.X; x++ {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()

			// Skip transparent/very dark pixels (likely background)
			if a < 32768 || (r < 8192 && g < 8192 && b < 8192) {
				continue
			}

			rSum += int(r >> 8)
			gSum += int(g >> 8)
			bSum += int(b >> 8)
			count++
		}
	}

	// Average color
	if count == 0 {
		// Fallback to theme foreground
		return *hexToRGBA(r.theme.Foreground)
	}

	return color.RGBA{
		R: uint8(rSum / count),
		G: uint8(gSum / count),
		B: uint8(bSum / count),
		A: 255,
	}
}

// mapToThemeColor finds the closest theme color using Euclidean distance
func (r *ColorizedBrailleRenderer) mapToThemeColor(target color.RGBA) color.RGBA {
	if len(r.themeCache) == 0 {
		return target // Fallback
	}

	closest := r.themeCache[0]
	minDistance := colorDistance(target, closest)

	for _, themeColor := range r.themeCache[1:] {
		dist := colorDistance(target, themeColor)
		if dist < minDistance {
			minDistance = dist
			closest = themeColor
		}
	}

	return closest
}

// colorDistance calculates Euclidean distance in RGB space
func colorDistance(c1, c2 color.RGBA) float64 {
	dr := float64(c1.R) - float64(c2.R)
	dg := float64(c1.G) - float64(c2.G)
	db := float64(c1.B) - float64(c2.B)
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

// getBraillePattern converts a 2x4 pixel block to braille character
// Adapted from dotmatrix library logic
func (r *ColorizedBrailleRenderer) getBraillePattern(img image.Image, startX, startY int) rune {
	bounds := img.Bounds()

	// Braille dot positions (Unicode offset)
	// 0x2800 is the base braille character ⠀
	// Dots are numbered:
	// 1 4
	// 2 5
	// 3 6
	// 7 8

	var pattern int
	dotMap := []int{0x01, 0x08, 0x02, 0x10, 0x04, 0x20, 0x40, 0x80}

	index := 0
	for x := startX; x < startX+2 && x < bounds.Max.X; x++ {
		for y := startY; y < startY+4 && y < bounds.Max.Y; y++ {
			c := img.At(x, y)
			_, _, _, a := c.RGBA()

			// If pixel is visible (not transparent), set the dot
			if a > 32768 {
				pattern |= dotMap[index]
			}
			index++
		}
	}

	return rune(0x2800 + pattern)
}
