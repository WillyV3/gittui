package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LanguagesComponent displays top programming languages
type LanguagesComponent struct {
	languages []LanguageStats
	repoCount int
	loading   bool
}

// NewLanguagesComponent creates a new languages component
func NewLanguagesComponent() LanguagesComponent {
	return LanguagesComponent{
		loading: true,
	}
}

// SetLanguages updates the language data
func (l *LanguagesComponent) SetLanguages(languages []LanguageStats, repoCount int) {
	l.languages = languages
	l.repoCount = repoCount
	l.loading = false
}

// Update handles messages
func (l LanguagesComponent) Update(msg tea.Msg) (LanguagesComponent, tea.Cmd) {
	return l, nil
}

// View renders the languages component
func (l LanguagesComponent) View() string {
	title := titleStyle.Render("Top Languages")

	if l.loading {
		return lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			labelStyle.Render("Loading..."))
	}

	if len(l.languages) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			labelStyle.Render("No language data"))
	}

	var bars []string
	bars = append(bars, title, "")

	// Bar width - using fixed width for consistency
	maxBarWidth := 25

	// Limit to top 3 languages for consistency
	displayLangs := l.languages
	if len(displayLangs) > 3 {
		displayLangs = displayLangs[:3]
	}

	for _, lang := range displayLangs {
		label := fmt.Sprintf("%-12s %5.1f%%", lang.Name, lang.Percentage*100)
		labelLine := baseStyle.Render(label)

		// barStyle sets the width, just render a space to fill it
		bar := barStyle(lang.Percentage, maxBarWidth, lang.Color).Render(" ")

		bars = append(bars, labelLine)
		bars = append(bars, bar)
		bars = append(bars, "") // Spacing between items
	}

	return lipgloss.JoinVertical(lipgloss.Left, bars...)
}
