package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// YAMLTheme represents the structure of theme YAML files
// These come from terminal color schemes with 16 ANSI colors
type YAMLTheme struct {
	Name       string `yaml:"name"`
	Author     string `yaml:"author"`
	Variant    string `yaml:"variant"` // dark or light
	Background string `yaml:"background"`
	Foreground string `yaml:"foreground"`
	Cursor     string `yaml:"cursor"`

	// 16 ANSI colors (color_01 through color_16)
	Color01 string `yaml:"color_01"` // Black
	Color02 string `yaml:"color_02"` // Red
	Color03 string `yaml:"color_03"` // Green
	Color04 string `yaml:"color_04"` // Yellow
	Color05 string `yaml:"color_05"` // Blue
	Color06 string `yaml:"color_06"` // Magenta
	Color07 string `yaml:"color_07"` // Cyan
	Color08 string `yaml:"color_08"` // White
	Color09 string `yaml:"color_09"` // Bright Black
	Color10 string `yaml:"color_10"` // Bright Red
	Color11 string `yaml:"color_11"` // Bright Green
	Color12 string `yaml:"color_12"` // Bright Yellow
	Color13 string `yaml:"color_13"` // Bright Blue
	Color14 string `yaml:"color_14"` // Bright Magenta
	Color15 string `yaml:"color_15"` // Bright Cyan
	Color16 string `yaml:"color_16"` // Bright White
}

// ConvertToTheme converts a YAML theme to our internal Theme structure
// Maps ANSI terminal colors to semantic UI colors with proper contrast
func (yt *YAMLTheme) ConvertToTheme() Theme {
	return Theme{
		// Base colors
		Background: yt.Background,
		Foreground: yt.Foreground,
		Subtle:     generateShade(yt.Background, 1.3), // Slightly lighter than background for subtle areas

		// Semantic UI colors - map from bright ANSI colors
		Blue:   yt.Color13, // Bright Blue - titles, links
		Green:  yt.Color11, // Bright Green - success, approved
		Red:    yt.Color10, // Bright Red - errors, danger
		Yellow: yt.Color12, // Bright Yellow - warnings, pending
		Purple: yt.Color14, // Bright Magenta - special status
		Gray:   yt.Color08, // White (dimmer) - labels, muted text (good contrast with Subtle)
		Dark:   yt.Color01, // Black - darker muted

		// Contribution graph - auto-generate gradient from primary accent
		ContribNone:   yt.Background,
		ContribLow:    generateShade(yt.Color13, 0.3), // Use Blue as base, 30% brightness
		ContribMed:    generateShade(yt.Color13, 0.5), // 50% brightness
		ContribHigh:   generateShade(yt.Color13, 0.7), // 70% brightness
		ContribHigher: yt.Color13,                     // Full brightness
	}
}

// generateShade adjusts the brightness of a color
// factor < 1.0 darkens, factor > 1.0 brightens, factor = 1.0 returns original
func generateShade(hexColor string, factor float64) string {
	// Remove # prefix if present
	hex := strings.TrimPrefix(hexColor, "#")

	// Parse RGB components
	if len(hex) != 6 {
		return hexColor // Return original if invalid
	}

	r, _ := strconv.ParseInt(hex[0:2], 16, 64)
	g, _ := strconv.ParseInt(hex[2:4], 16, 64)
	b, _ := strconv.ParseInt(hex[4:6], 16, 64)

	// Apply brightness factor
	r = int64(float64(r) * factor)
	g = int64(float64(g) * factor)
	b = int64(float64(b) * factor)

	// Clamp to valid range (0-255)
	if r < 0 { r = 0 }
	if r > 255 { r = 255 }
	if g < 0 { g = 0 }
	if g > 255 { g = 255 }
	if b < 0 { b = 0 }
	if b > 255 { b = 255 }

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// LoadThemeFromYAML loads a single YAML theme file
func LoadThemeFromYAML(filePath string) (*YAMLTheme, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read theme file: %w", err)
	}

	var yamlTheme YAMLTheme
	if err := yaml.Unmarshal(data, &yamlTheme); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &yamlTheme, nil
}

// LoadAllThemes loads all YAML themes from a directory
// Returns a map of theme-name -> Theme
func LoadAllThemes(themesDir string) (map[string]Theme, error) {
	themeMap := make(map[string]Theme)

	// Find all .yml files
	files, err := filepath.Glob(filepath.Join(themesDir, "*.yml"))
	if err != nil {
		return nil, fmt.Errorf("failed to list theme files: %w", err)
	}

	// Load each theme
	for _, file := range files {
		yamlTheme, err := LoadThemeFromYAML(file)
		if err != nil {
			// Skip invalid themes, don't fail entire load
			continue
		}

		// Convert to our Theme format
		theme := yamlTheme.ConvertToTheme()

		// Use lowercase name with spaces replaced by hyphens for consistency
		themeName := strings.ToLower(strings.ReplaceAll(yamlTheme.Name, " ", "-"))
		themeMap[themeName] = theme
	}

	return themeMap, nil
}

// GetThemesByVariant filters themes by dark/light variant
func GetThemesByVariant(allThemes map[string]Theme, themesDir string, variant string) []string {
	var filtered []string

	files, _ := filepath.Glob(filepath.Join(themesDir, "*.yml"))
	for _, file := range files {
		yamlTheme, err := LoadThemeFromYAML(file)
		if err != nil {
			continue
		}

		if yamlTheme.Variant == variant {
			themeName := strings.ToLower(strings.ReplaceAll(yamlTheme.Name, " ", "-"))
			filtered = append(filtered, themeName)
		}
	}

	return filtered
}
