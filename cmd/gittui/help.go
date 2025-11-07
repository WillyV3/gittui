package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpModal is a modal overlay for viewing keybindings and help
type HelpModal struct {
	width   int
	height  int
	visible bool
	isOwnProfile bool
	hasMultipleAvatarSources bool
}

// NewHelpModal creates a new help modal
func NewHelpModal() HelpModal {
	return HelpModal{
		visible: false,
	}
}

// Update handles messages for the help modal
func (m HelpModal) Update(msg tea.Msg) (HelpModal, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc", "q", "?", "h":
			m.visible = false
			return m, nil
		}
	}

	if windowMsg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = windowMsg.Width
		m.height = windowMsg.Height
	}

	return m, nil
}

// SetContext updates context for help display
func (m *HelpModal) SetContext(isOwnProfile bool, hasMultipleAvatarSources bool) {
	m.isOwnProfile = isOwnProfile
	m.hasMultipleAvatarSources = hasMultipleAvatarSources
}

// View renders the help modal
func (m HelpModal) View() string {
	if !m.visible {
		return ""
	}

	// Build help content
	var sections []string

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(CurrentTheme.Blue)).
		Render("Keybindings & Help")
	sections = append(sections, title, "")

	// Section style
	sectionTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(CurrentTheme.Green))

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.Blue)).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.Foreground))

	// General
	sections = append(sections, sectionTitleStyle.Render("General"))
	sections = append(sections,
		keyStyle.Render("  q")+" "+descStyle.Render("Quit application"),
		keyStyle.Render("  r")+" "+descStyle.Render("Refresh all data"),
		keyStyle.Render("  ?/h")+" "+descStyle.Render("Toggle this help"),
		"")

	// Theme
	sections = append(sections, sectionTitleStyle.Render("Theme"))
	sections = append(sections,
		keyStyle.Render("  t")+" "+descStyle.Render("Cycle through themes (361 available)"),
		keyStyle.Render("  T")+" "+descStyle.Render("Save current theme as default"),
		"")

	// Avatar controls
	sections = append(sections, sectionTitleStyle.Render("Avatar"))
	sections = append(sections,
		keyStyle.Render("  -")+" "+descStyle.Render("Decrease avatar size (min: 30)"),
		keyStyle.Render("  =")+" "+descStyle.Render("Increase avatar size (max: 240)"),
		keyStyle.Render("  b")+" "+descStyle.Render("Toggle avatar renderer (mosaic/ascii)"))

	if m.isOwnProfile && m.hasMultipleAvatarSources {
		sections = append(sections,
			keyStyle.Render("  v")+" "+descStyle.Render("Cycle avatar source (personal/orgs)"))
	}
	sections = append(sections, "")

	// Activity
	sections = append(sections, sectionTitleStyle.Render("Activity"))
	sections = append(sections,
		keyStyle.Render("  a")+" "+descStyle.Render("Toggle activity modal"),
		keyStyle.Render("  g")+" "+descStyle.Render("Cycle push stats granularity (hour/day/week/month)"),
		"")

	// Activity modal controls
	sections = append(sections, sectionTitleStyle.Render("Activity Modal (when open)"))
	sections = append(sections,
		keyStyle.Render("  ↑/↓, j/k")+" "+descStyle.Render("Scroll line by line"),
		keyStyle.Render("  u/d")+" "+descStyle.Render("Scroll half page"),
		keyStyle.Render("  pgup/pgdn")+" "+descStyle.Render("Scroll full page"),
		keyStyle.Render("  esc/q")+" "+descStyle.Render("Close modal"),
		"")

	// View toggle (only for own profile)
	if m.isOwnProfile {
		sections = append(sections, sectionTitleStyle.Render("Privacy"))
		sections = append(sections,
			keyStyle.Render("  p")+" "+descStyle.Render("Toggle public/private view"),
			"")
	}

	// Configuration
	sections = append(sections, sectionTitleStyle.Render("Configuration"))
	sections = append(sections,
		descStyle.Render("  Config file: ~/.local/gittui/config.json"),
		descStyle.Render("  Run setup: make config"),
		"")

	content := strings.Join(sections, "\n")

	// Modal box
	modalWidth := 70
	modalBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(CurrentTheme.Blue)).
		Padding(2, 4).
		Width(modalWidth).
		Render(content)

	// Center the modal
	boxWidth := lipgloss.Width(modalBox)
	boxHeight := lipgloss.Height(modalBox)

	leftPadding := (m.width - boxWidth) / 2
	if leftPadding < 0 {
		leftPadding = 0
	}

	topPadding := (m.height - boxHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	centered := lipgloss.NewStyle().
		PaddingLeft(leftPadding).
		PaddingTop(topPadding).
		Render(modalBox)

	return centered
}

// Show makes the modal visible
func (m *HelpModal) Show() {
	m.visible = true
}

// Hide hides the modal
func (m *HelpModal) Hide() {
	m.visible = false
}

// IsVisible returns whether the modal is currently visible
func (m HelpModal) IsVisible() bool {
	return m.visible
}
