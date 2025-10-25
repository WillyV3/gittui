package main

import (
	"os"
	"path/filepath"
	"sort"
)

// Theme provides all colors for the application.
// Colors can be ANSI, ANSI256, or Hex format.
type Theme struct {
	// Base colors
	Background string
	Foreground string
	Subtle     string

	// Accent colors - semantic naming for UI elements
	Blue   string // Titles, links, borders
	Green  string // Success, approved states
	Red    string // Errors, danger, rejections
	Yellow string // Warnings, pending states
	Purple string // Special status (merged PRs)
	Gray   string // Muted text, labels
	Dark   string // Darker muted text

	// Contribution graph intensity levels (0-4)
	ContribNone   string // No contributions
	ContribLow    string // 1-3 contributions
	ContribMed    string // 4-6 contributions
	ContribHigh   string // 7-9 contributions
	ContribHigher string // 10+ contributions
}

// themes registry - all available themes (built-in + loaded from files)
var themes = map[string]Theme{
	"github-dark": githubDarkTheme(),
	"dracula":     draculaTheme(),
	"nord":        nordTheme(),
}

// CurrentTheme is the active theme (set at startup)
var CurrentTheme Theme

// currentThemeName tracks the current theme name for cycling
var currentThemeName string

// themeOrder defines the order for cycling through themes
// Will be populated with all available theme names, sorted
var themeOrder []string

// builtInThemes are the hand-crafted themes (always available)
var builtInThemes = []string{"github-dark", "dracula", "nord"}

// InitTheme initializes the theme system
// 1. Loads all themes from YAML files
// 2. Sets current theme from env var or default
func InitTheme() {
	// Load all themes from the themes directory
	loadExternalThemes()

	// Build theme order (built-in first, then alphabetically)
	buildThemeOrder()

	// Set initial theme
	themeName := os.Getenv("GITTUI_THEME")
	if themeName == "" {
		themeName = "github-dark" // Default theme
	}

	theme, exists := themes[themeName]
	if !exists {
		// Invalid theme name, fall back to default
		themeName = "github-dark"
		theme = themes["github-dark"]
	}

	CurrentTheme = theme
	currentThemeName = themeName
}

// loadExternalThemes loads all YAML theme files from the themes directory
func loadExternalThemes() {
	// Get the directory where the executable is located
	exePath, err := os.Executable()
	if err != nil {
		return // Silently fail - built-in themes still work
	}
	exeDir := filepath.Dir(exePath)
	themesDir := filepath.Join(exeDir, "themes")

	// Try alternate location (development mode - relative to source)
	if _, err := os.Stat(themesDir); os.IsNotExist(err) {
		themesDir = "themes"
	}

	// Load all themes
	loadedThemes, err := LoadAllThemes(themesDir)
	if err != nil {
		return // Silently fail - built-in themes still work
	}

	// Add loaded themes to registry (don't overwrite built-in)
	for name, theme := range loadedThemes {
		if _, isBuiltIn := themes[name]; !isBuiltIn {
			themes[name] = theme
		}
	}
}

// buildThemeOrder creates the cycling order: built-in themes first, then all others alphabetically
func buildThemeOrder() {
	// Start with built-in themes
	themeOrder = make([]string, len(builtInThemes))
	copy(themeOrder, builtInThemes)

	// Add all other themes alphabetically
	var otherThemes []string
	for name := range themes {
		isBuiltIn := false
		for _, builtIn := range builtInThemes {
			if name == builtIn {
				isBuiltIn = true
				break
			}
		}
		if !isBuiltIn {
			otherThemes = append(otherThemes, name)
		}
	}

	// Sort other themes alphabetically
	sort.Strings(otherThemes)

	// Append to theme order
	themeOrder = append(themeOrder, otherThemes...)
}

// NextTheme cycles to the next theme in the rotation
// Returns the name of the new theme
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

// GetAvailableThemes returns all theme names in cycling order
func GetAvailableThemes() []string {
	return themeOrder
}

// GetThemeCount returns the total number of available themes
func GetThemeCount() int {
	return len(themes)
}

// IsBuiltInTheme returns true if the theme is a hand-crafted built-in theme
func IsBuiltInTheme(name string) bool {
	for _, builtIn := range builtInThemes {
		if name == builtIn {
			return true
		}
	}
	return false
}

// githubDarkTheme - GitHub's official dark theme
// Based on GitHub Primer design system
func githubDarkTheme() Theme {
	return Theme{
		// Base
		Background: "#0d1117",
		Foreground: "#c9d1d9",
		Subtle:     "#484f58",

		// Accents
		Blue:   "#58a6ff",
		Green:  "#3fb950",
		Red:    "#f85149",
		Yellow: "#d29922",
		Purple: "#8957e5",
		Gray:   "#7d8590",
		Dark:   "#6e7681",

		// Contribution graph (GitHub green scale)
		ContribNone:   "#161b22",
		ContribLow:    "#0e4429",
		ContribMed:    "#006d32",
		ContribHigh:   "#26a641",
		ContribHigher: "#39d353",
	}
}

// draculaTheme - Dracula color scheme
// Popular dark theme with vibrant purple/pink accents
// https://draculatheme.com/contribute
func draculaTheme() Theme {
	return Theme{
		// Base
		Background: "#282a36",
		Foreground: "#f8f8f2",
		Subtle:     "#44475a",

		// Accents
		Blue:   "#8be9fd", // Cyan
		Green:  "#50fa7b",
		Red:    "#ff5555",
		Yellow: "#f1fa8c",
		Purple: "#bd93f9",
		Gray:   "#6272a4",
		Dark:   "#44475a",

		// Contribution graph (Dracula PURPLE scale - signature color)
		ContribNone:   "#282a36", // Background
		ContribLow:    "#44355b", // Dark purple
		ContribMed:    "#6d4a9e", // Medium purple
		ContribHigh:   "#9d6fc9", // Bright purple
		ContribHigher: "#bd93f9", // Brightest purple (matches accent)
	}
}

// nordTheme - Nord color scheme
// Arctic, north-bluish color palette
// https://www.nordtheme.com/docs/colors-and-palettes
func nordTheme() Theme {
	return Theme{
		// Base
		Background: "#2e3440",
		Foreground: "#eceff4",
		Subtle:     "#4c566a",

		// Accents
		Blue:   "#88c0d0", // Nord frost
		Green:  "#a3be8c", // Nord aurora green
		Red:    "#bf616a", // Nord aurora red
		Yellow: "#ebcb8b", // Nord aurora yellow
		Purple: "#b48ead", // Nord aurora purple
		Gray:   "#616e88",
		Dark:   "#4c566a",

		// Contribution graph (Nord BLUE/FROST scale - signature color)
		ContribNone:   "#2e3440", // Background
		ContribLow:    "#46586a", // Dark frost
		ContribMed:    "#5e7a94", // Medium frost
		ContribHigh:   "#81a1c1", // Bright frost (Nord frost)
		ContribHigher: "#88c0d0", // Brightest frost (matches accent)
	}
}
