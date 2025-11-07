package main

import "github.com/charmbracelet/lipgloss"

// Styles are initialized after theme is loaded
// All styles dynamically use CurrentTheme for colors

// GetBaseStyle returns the base text style with adaptive foreground for good contrast
func GetBaseStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{
			Light: CurrentTheme.Foreground,
			Dark:  CurrentTheme.BrightWhite,
		})
}

// GetTitleStyle returns the title style with adaptive blue for good contrast
func GetTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{
			Light: CurrentTheme.Blue,
			Dark:  CurrentTheme.BrightBlue,
		})
}

// GetLabelStyle returns the label style with theme gray color
func GetLabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.Gray))
}

// GetAccentStyle returns the accent style with adaptive green for good contrast
func GetAccentStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{
			Light: CurrentTheme.Green,
			Dark:  CurrentTheme.BrightGreen,
		}).
		Bold(true)
}

// GetSubtleStyle returns the subtle style with theme subtle color
func GetSubtleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.Subtle))
}

// GetStatusBarStyle returns the status bar style
func GetStatusBarStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.Gray)).
		Background(lipgloss.Color(CurrentTheme.Subtle))
}

// GetErrorStyle returns the error style with theme red color
func GetErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.Red)).
		Bold(true)
}

// GetLoadingStyle returns the loading style with theme gray color
func GetLoadingStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.Gray)).
		Bold(true)
}

// GetDimStyle returns the dim style for hints and metadata
func GetDimStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.BrightBlack))
}

// Legacy global styles for backward compatibility
// These are now functions that return themed styles
var (
	baseStyle      = GetBaseStyle()
	titleStyle     = GetTitleStyle()
	labelStyle     = GetLabelStyle()
	accentStyle    = GetAccentStyle()
	subtleStyle    = GetSubtleStyle()
	statusBarStyle = GetStatusBarStyle()
	errorStyle     = GetErrorStyle()
	loadingStyle   = GetLoadingStyle()
	dimStyle       = GetDimStyle()
)

// InitStyles must be called after InitTheme() to set up global styles
func InitStyles() {
	baseStyle = GetBaseStyle()
	titleStyle = GetTitleStyle()
	labelStyle = GetLabelStyle()
	accentStyle = GetAccentStyle()
	subtleStyle = GetSubtleStyle()
	statusBarStyle = GetStatusBarStyle()
	errorStyle = GetErrorStyle()
	loadingStyle = GetLoadingStyle()
	dimStyle = GetDimStyle()
}

// Bar chart style for language breakdown
func barStyle(percentage float64, maxWidth int, color string) lipgloss.Style {
	width := int(percentage * float64(maxWidth))
	if width < 1 {
		width = 1
	}
	return lipgloss.NewStyle().
		Background(lipgloss.Color(color)).
		Width(width)
}
