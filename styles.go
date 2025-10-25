package main

import "github.com/charmbracelet/lipgloss"

// Styles are initialized after theme is loaded
// All styles dynamically use CurrentTheme for colors

// GetBaseStyle returns the base text style with theme foreground color
func GetBaseStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.Foreground))
}

// GetTitleStyle returns the title style with theme blue color
func GetTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(CurrentTheme.Blue))
}

// GetLabelStyle returns the label style with theme gray color
func GetLabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.Gray))
}

// GetAccentStyle returns the accent style with theme green color
func GetAccentStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.Green)).
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
