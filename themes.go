package main

import (
	"os"
	"sort"

	goghthemes "github.com/willyv3/gogh-themes"
)

// Theme provides all colors for the application.
type Theme struct {
	// Base colors
	Background string
	Foreground string
	Subtle     string

	// Primary ANSI colors (0-7)
	Black   string
	Red     string
	Green   string
	Yellow  string
	Blue    string
	Magenta string
	Cyan    string
	White   string

	// Bright ANSI colors (8-15)
	BrightBlack   string
	BrightRed     string
	BrightGreen   string
	BrightYellow  string
	BrightBlue    string
	BrightMagenta string
	BrightCyan    string
	BrightWhite   string

	// Semantic aliases for convenience
	Purple string // Alias for Magenta
	Gray   string // Alias for White
	Dark   string // Alias for Black

	// Contribution graph intensity levels (0-4)
	ContribNone   string
	ContribLow    string
	ContribMed    string
	ContribHigh   string
	ContribHigher string
}

// themes registry - all themes from gogh-themes package
var themes = make(map[string]Theme)

// CurrentTheme is the active theme
var CurrentTheme Theme

// currentThemeName tracks the current theme name for cycling
var currentThemeName string

// themeOrder defines the order for cycling through themes
var themeOrder []string

// InitTheme initializes the theme system
func InitTheme() {
	// Load all themes from gogh-themes package
	loadAllThemes()

	// Build sorted theme order
	buildThemeOrder()

	// Set initial theme
	themeName := os.Getenv("GITTUI_THEME")
	if themeName == "" {
		themeName = "Dracula" // Default theme
	}

	theme, exists := themes[themeName]
	if !exists {
		// Fall back to first available theme
		if len(themeOrder) > 0 {
			themeName = themeOrder[0]
			theme = themes[themeName]
		}
	}

	CurrentTheme = theme
	currentThemeName = themeName
}

// loadAllThemes loads all themes from gogh-themes package
func loadAllThemes() {
	allGoghThemes := goghthemes.All()

	for name, goghTheme := range allGoghThemes {
		// Convert gogh theme to our Theme struct with full 16-color support
		theme := Theme{
			Background: goghTheme.Background,
			Foreground: goghTheme.Foreground,
			Subtle:     generateShade(goghTheme.Background, 1.3), // 30% brighter

			// Primary ANSI colors (0-7)
			Black:   goghTheme.Black,
			Red:     goghTheme.Red,
			Green:   goghTheme.Green,
			Yellow:  goghTheme.Yellow,
			Blue:    goghTheme.Blue,
			Magenta: goghTheme.Magenta,
			Cyan:    goghTheme.Cyan,
			White:   goghTheme.White,

			// Bright ANSI colors (8-15)
			BrightBlack:   goghTheme.BrightBlack,
			BrightRed:     goghTheme.BrightRed,
			BrightGreen:   goghTheme.BrightGreen,
			BrightYellow:  goghTheme.BrightYellow,
			BrightBlue:    goghTheme.BrightBlue,
			BrightMagenta: goghTheme.BrightMagenta,
			BrightCyan:    goghTheme.BrightCyan,
			BrightWhite:   goghTheme.BrightWhite,

			// Semantic aliases
			Purple: goghTheme.Magenta,
			Gray:   goghTheme.White,
			Dark:   goghTheme.Black,

			// Generate contribution graph gradient (using blue as base)
			ContribNone:   goghTheme.Background,
			ContribLow:    generateShade(goghTheme.Blue, 0.3),
			ContribMed:    generateShade(goghTheme.Blue, 0.5),
			ContribHigh:   generateShade(goghTheme.Blue, 0.7),
			ContribHigher: goghTheme.Blue,
		}

		themes[name] = theme
	}
}

// buildThemeOrder creates alphabetically sorted theme cycling order
func buildThemeOrder() {
	themeOrder = make([]string, 0, len(themes))
	for name := range themes {
		themeOrder = append(themeOrder, name)
	}
	sort.Strings(themeOrder)
}

// NextTheme cycles to the next theme in the rotation
func NextTheme() string {
	// Find current theme index
	currentIndex := 0
	for i, name := range themeOrder {
		if name == currentThemeName {
			currentIndex = i
			break
		}
	}

	// Cycle to next theme (wrap around)
	nextIndex := (currentIndex + 1) % len(themeOrder)
	nextThemeName := themeOrder[nextIndex]

	// Apply new theme
	CurrentTheme = themes[nextThemeName]
	currentThemeName = nextThemeName

	return nextThemeName
}

// GetCurrentThemeName returns the name of the active theme
func GetCurrentThemeName() string {
	return currentThemeName
}

// GetThemeCount returns the total number of available themes
func GetThemeCount() int {
	return len(themes)
}

// generateShade generates a lighter or darker shade of a hex color
// factor < 1.0 = darker, factor > 1.0 = brighter
func generateShade(hexColor string, factor float64) string {
	// Remove # prefix if present
	if len(hexColor) > 0 && hexColor[0] == '#' {
		hexColor = hexColor[1:]
	}

	// Parse hex to RGB
	var r, g, b int64
	if len(hexColor) == 6 {
		r, _ = parseHex(hexColor[0:2])
		g, _ = parseHex(hexColor[2:4])
		b, _ = parseHex(hexColor[4:6])
	} else {
		return "#000000"
	}

	// Apply factor
	r = int64(float64(r) * factor)
	g = int64(float64(g) * factor)
	b = int64(float64(b) * factor)

	// Clamp to 0-255
	if r > 255 {
		r = 255
	}
	if g > 255 {
		g = 255
	}
	if b > 255 {
		b = 255
	}

	return formatHex(r, g, b)
}

// parseHex parses a 2-character hex string to int64
func parseHex(s string) (int64, error) {
	var result int64
	for i := 0; i < len(s); i++ {
		result *= 16
		c := s[i]
		if c >= '0' && c <= '9' {
			result += int64(c - '0')
		} else if c >= 'a' && c <= 'f' {
			result += int64(c - 'a' + 10)
		} else if c >= 'A' && c <= 'F' {
			result += int64(c - 'A' + 10)
		}
	}
	return result, nil
}

// formatHex formats RGB values to hex string
func formatHex(r, g, b int64) string {
	return "#" + toHex(r) + toHex(g) + toHex(b)
}

// toHex converts a number to 2-digit hex
func toHex(n int64) string {
	const hexDigits = "0123456789ABCDEF"
	return string([]byte{hexDigits[n/16], hexDigits[n%16]})
}
