package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strconv"
	"strings"

	"github.com/kevin-cantwell/dotmatrix"
)

// ColorizedBrailleRenderer handles theme-aware avatar colorization
type ColorizedBrailleRenderer struct {
	theme      Theme
	themeCache []color.RGBA // Parsed theme colors for fast lookup
}

// ContrastFilter increases image contrast before braille conversion
type ContrastFilter struct {
	Factor float64 // 1.0 = no change, >1.0 = more contrast
}

// Filter applies contrast adjustment to the image
func (f *ContrastFilter) Filter(img image.Image) image.Image {
	bounds := img.Bounds()
	adjusted := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			oldColor := img.At(x, y)
			r, g, b, a := oldColor.RGBA()

			// Convert to 0-255 range
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)

			// Apply contrast adjustment (pivot around middle gray 128)
			r8 = f.adjustContrast(r8)
			g8 = f.adjustContrast(g8)
			b8 = f.adjustContrast(b8)

			adjusted.Set(x, y, color.RGBA{R: r8, G: g8, B: b8, A: uint8(a >> 8)})
		}
	}

	return adjusted
}

// adjustContrast applies contrast adjustment to a single channel
func (f *ContrastFilter) adjustContrast(val uint8) uint8 {
	// Pivot around middle gray (128)
	fval := float64(val)
	adjusted := ((fval - 128.0) * f.Factor) + 128.0

	// Clamp to 0-255
	if adjusted < 0 {
		return 0
	}
	if adjusted > 255 {
		return 255
	}
	return uint8(adjusted)
}

// SharpnessFilter enhances edges using unsharp mask technique
type SharpnessFilter struct {
	Amount float64 // 1.0 = no change, >1.0 = more sharpening
}

// Filter applies edge enhancement to the image
func (f *SharpnessFilter) Filter(img image.Image) image.Image {
	bounds := img.Bounds()
	sharpened := image.NewRGBA(bounds)

	// Create a blurred version for edge detection
	blurred := f.gaussianBlur(img)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			origR, origG, origB, origA := img.At(x, y).RGBA()
			blurR, blurG, blurB, _ := blurred.At(x, y).RGBA()

			// Convert to 0-255
			or, og, ob := uint8(origR>>8), uint8(origG>>8), uint8(origB>>8)
			br, bg, bb := uint8(blurR>>8), uint8(blurG>>8), uint8(blurB>>8)

			// Unsharp mask: original + (original - blurred) * amount
			sr := f.sharpen(or, br)
			sg := f.sharpen(og, bg)
			sb := f.sharpen(ob, bb)

			sharpened.Set(x, y, color.RGBA{R: sr, G: sg, B: sb, A: uint8(origA >> 8)})
		}
	}

	return sharpened
}

// sharpen applies unsharp mask to a single channel
func (f *SharpnessFilter) sharpen(orig, blur uint8) uint8 {
	edge := float64(orig) - float64(blur)
	result := float64(orig) + (edge * (f.Amount - 1.0))

	if result < 0 {
		return 0
	}
	if result > 255 {
		return 255
	}
	return uint8(result)
}

// gaussianBlur applies a simple 3x3 Gaussian blur
func (f *SharpnessFilter) gaussianBlur(img image.Image) image.Image {
	bounds := img.Bounds()
	blurred := image.NewRGBA(bounds)

	// Simple 3x3 Gaussian kernel (normalized)
	kernel := [3][3]float64{
		{1.0 / 16, 2.0 / 16, 1.0 / 16},
		{2.0 / 16, 4.0 / 16, 2.0 / 16},
		{1.0 / 16, 2.0 / 16, 1.0 / 16},
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var rSum, gSum, bSum float64
			var aVal uint32

			// Convolve with kernel
			for ky := -1; ky <= 1; ky++ {
				for kx := -1; kx <= 1; kx++ {
					px := x + kx
					py := y + ky

					// Clamp to bounds
					if px < bounds.Min.X {
						px = bounds.Min.X
					}
					if px >= bounds.Max.X {
						px = bounds.Max.X - 1
					}
					if py < bounds.Min.Y {
						py = bounds.Min.Y
					}
					if py >= bounds.Max.Y {
						py = bounds.Max.Y - 1
					}

					r, g, b, a := img.At(px, py).RGBA()
					weight := kernel[ky+1][kx+1]

					rSum += float64(r>>8) * weight
					gSum += float64(g>>8) * weight
					bSum += float64(b>>8) * weight
					aVal = a
				}
			}

			blurred.Set(x, y, color.RGBA{
				R: uint8(rSum),
				G: uint8(gSum),
				B: uint8(bSum),
				A: uint8(aVal >> 8),
			})
		}
	}

	return blurred
}

// GammaFilter adjusts midtones using gamma correction
type GammaFilter struct {
	Gamma float64 // 1.0 = no change, <1.0 = darker, >1.0 = brighter midtones
}

// Filter applies gamma correction to the image
func (f *GammaFilter) Filter(img image.Image) image.Image {
	bounds := img.Bounds()
	adjusted := image.NewRGBA(bounds)

	// Precompute gamma lookup table for performance
	var gammaLUT [256]uint8
	invGamma := 1.0 / f.Gamma
	for i := 0; i < 256; i++ {
		normalized := float64(i) / 255.0
		corrected := math.Pow(normalized, invGamma)
		gammaLUT[i] = uint8(corrected * 255.0)
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()

			// Apply gamma using lookup table
			r8 := gammaLUT[uint8(r>>8)]
			g8 := gammaLUT[uint8(g>>8)]
			b8 := gammaLUT[uint8(b>>8)]

			adjusted.Set(x, y, color.RGBA{R: r8, G: g8, B: b8, A: uint8(a >> 8)})
		}
	}

	return adjusted
}

// ChainFilter applies multiple filters in sequence
type ChainFilter struct {
	Filters []dotmatrix.Filter
}

// Filter applies all filters in order
func (f *ChainFilter) Filter(img image.Image) image.Image {
	result := img
	for _, filter := range f.Filters {
		result = filter.Filter(result)
	}
	return result
}

// NewColorizedBrailleRenderer creates a renderer with the current theme
func NewColorizedBrailleRenderer(theme Theme) *ColorizedBrailleRenderer {
	return &ColorizedBrailleRenderer{
		theme:      theme,
		themeCache: parseThemeColors(theme),
	}
}

// parseThemeColors converts theme hex colors to RGBA for distance calculations
// Uses full 16-color ANSI palette for rich avatar colorization
func parseThemeColors(theme Theme) []color.RGBA {
	hexColors := []string{
		// Primary ANSI colors (0-7)
		theme.Black,
		theme.Red,
		theme.Green,
		theme.Yellow,
		theme.Blue,
		theme.Magenta,
		theme.Cyan,
		theme.White,

		// Bright ANSI colors (8-15)
		theme.BrightBlack,
		theme.BrightRed,
		theme.BrightGreen,
		theme.BrightYellow,
		theme.BrightBlue,
		theme.BrightMagenta,
		theme.BrightCyan,
		theme.BrightWhite,

		// Special colors for edge cases
		theme.Foreground,
		theme.Background,
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
// Uses dotmatrix for braille generation with theme-aware colorization
func (r *ColorizedBrailleRenderer) RenderColorized(img image.Image) string {
	bounds := img.Bounds()

	// Step 1: Apply filter chain and generate braille with dotmatrix
	var buf bytes.Buffer

	// Optimal combo for avatars: Sharpen edges -> Boost contrast -> Adjust midtones
	// This preserves detail while making features pop for braille conversion
	config := &dotmatrix.Config{
		Filter: &ChainFilter{
			Filters: []dotmatrix.Filter{
				&SharpnessFilter{Amount: 10.0}, // Extreme edge enhancement for maximum detail
				&GammaFilter{Gamma: 0.1},       // Darken significantly for better contrast
				&ContrastFilter{Factor: 0.8},   // Reduced contrast to preserve tonal range
			},
		},
		Drawer: draw.FloydSteinberg, // Floyd-Steinberg dithering
	}

	printer := dotmatrix.NewPrinter(&buf, config)
	err := printer.Print(img)
	if err != nil {
		return "" // Return empty on error
	}

	brailleOutput := buf.String()
	lines := strings.Split(strings.TrimSuffix(brailleOutput, "\n"), "\n")

	// Step 2: Colorize each braille character based on original image
	var colorized strings.Builder

	for lineIdx, line := range lines {
		charIdx := 0
		for _, char := range line {
			// Calculate position in original image (2 pixels wide, 4 pixels tall per braille)
			x := charIdx * 2
			y := lineIdx * 4

			// Skip if out of bounds
			if x >= bounds.Dx() || y >= bounds.Dy() {
				colorized.WriteRune(char)
				charIdx++
				continue
			}

			// Get dominant color from original image at this block
			dominantColor := r.getDominantColor(img, x, y)
			themeColor := r.mapToThemeColor(dominantColor)

			// Write colorized braille character
			if char == ' ' || char == '⠀' {
				colorized.WriteRune(' ')
			} else {
				colorized.WriteString(fmt.Sprintf("\033[38;2;%d;%d;%dm%c\033[0m",
					themeColor.R, themeColor.G, themeColor.B, char))
			}

			charIdx++
		}
		colorized.WriteString("\n")
	}

	return colorized.String()
}

// getDominantColor finds the average non-background color in a 2x4 block
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

