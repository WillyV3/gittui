package main

import "github.com/charmbracelet/lipgloss"

// GitHub color palette (additional colors for main app)
const (
	colorBackground = "#0d1117"
	colorForeground = "#c9d1d9"
	colorAccent     = "#39d353"
	colorSubtle     = "#484f58"
)

// Base text styles - no dimensions, pure styling
var (
	baseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorForeground))

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#58a6ff"))

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7d8590"))

	accentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorAccent)).
			Bold(true)

	subtleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorSubtle))

	// Status bar style
	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7d8590")).
			Background(lipgloss.Color(colorSubtle))
)

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

// Error style
var errorStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#f85149")).
	Bold(true)

// Loading style
var loadingStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#7d8590")).
	Bold(true)
