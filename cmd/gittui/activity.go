package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ActivityModal is a modal overlay for viewing activity in detail
type ActivityModal struct {
	viewport   viewport.Model
	activities []Activity
	width      int
	height     int
	visible    bool
}

// NewActivityModal creates a new activity modal
func NewActivityModal() ActivityModal {
	vp := viewport.New(80, 20)
	vp.MouseWheelEnabled = true
	vp.HighPerformanceRendering = false // Set to false for smoother rendering
	vp.KeyMap = viewport.DefaultKeyMap() // Use default vim-style keys
	return ActivityModal{
		viewport: vp,
		visible:  false,
	}
}

// Update handles messages for the activity modal
func (m ActivityModal) Update(msg tea.Msg) (ActivityModal, tea.Cmd) {
	var cmd tea.Cmd

	// Handle our own keys first
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc", "q":
			m.visible = false
			return m, nil
		}
	}

	// Handle window resize
	if windowMsg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = windowMsg.Width
		m.height = windowMsg.Height
		// Modal takes 80% of screen
		modalWidth := int(float64(windowMsg.Width) * 0.8)
		modalHeight := int(float64(windowMsg.Height) * 0.8)
		m.viewport.Width = modalWidth - 4  // Account for border + padding
		m.viewport.Height = modalHeight - 6 // Account for border + padding + title
	}

	// Pass all messages to viewport for scrolling (only once)
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// SetActivities updates the activities data
func (m *ActivityModal) SetActivities(activities []Activity) {
	m.activities = activities
	// Set viewport content immediately when activities are loaded
	content := m.renderActivityList(m.activities)
	m.viewport.SetContent(content)
}

// View renders the activity modal
func (m ActivityModal) View() string {
	if !m.visible {
		return ""
	}

	// Title with count
	activityCount := len(m.activities)
	titleText := fmt.Sprintf("Recent Activity (%d)", activityCount)
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(CurrentTheme.Blue)).
		Render(titleText)

	// Help text
	helpText := "ESC/Q: close • ↑/↓/j/k: scroll • u/d: half page • pgup/pgdn: full page"
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.Subtle)).
		Render(helpText)

	// Modal content
	modalContent := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		m.viewport.View(),
		"",
		help,
	)

	// Modal box with border
	modalBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(CurrentTheme.Blue)).
		Padding(1, 2).
		Width(m.viewport.Width + 4).
		Render(modalContent)

	// Center the modal
	modalWidth := lipgloss.Width(modalBox)
	modalHeight := lipgloss.Height(modalBox)

	leftPadding := (m.width - modalWidth) / 2
	if leftPadding < 0 {
		leftPadding = 0
	}

	topPadding := (m.height - modalHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	// Add padding to center
	centered := lipgloss.NewStyle().
		PaddingLeft(leftPadding).
		PaddingTop(topPadding).
		Render(modalBox)

	return centered
}

// Show makes the modal visible
func (m *ActivityModal) Show() {
	m.visible = true
	// Reset viewport to top when opening modal
	m.viewport.GotoTop()
}

// Hide hides the modal
func (m *ActivityModal) Hide() {
	m.visible = false
}

// IsVisible returns whether the modal is currently visible
func (m ActivityModal) IsVisible() bool {
	return m.visible
}

// renderActivityList renders the activity list for the modal
func (m ActivityModal) renderActivityList(activities []Activity) string {
	if len(activities) == 0 {
		return labelStyle.Render("No recent activity")
	}

	var lines []string
	for i, activity := range activities {
		// Icon based on type
		var icon string
		var color lipgloss.Color
		switch activity.Type {
		case "PushEvent":
			icon = "↑"
			color = lipgloss.Color(CurrentTheme.Green)
		case "PullRequestEvent":
			icon = "⇄"
			color = lipgloss.Color(CurrentTheme.Blue)
		case "IssuesEvent":
			icon = "●"
			color = lipgloss.Color(CurrentTheme.Yellow)
		case "CreateEvent":
			icon = "+"
			color = lipgloss.Color(CurrentTheme.Cyan)
		case "WatchEvent":
			icon = "★"
			color = lipgloss.Color(CurrentTheme.Magenta)
		default:
			icon = "·"
			color = lipgloss.Color(CurrentTheme.Foreground)
		}

		iconStyled := lipgloss.NewStyle().
			Foreground(color).
			Render(icon)

		// Repo name
		privacy := ""
		if !activity.Public {
			privacy = "[private] "
		}
		repoStyled := lipgloss.NewStyle().
			Foreground(lipgloss.Color(CurrentTheme.Blue)).
			Render(privacy + activity.Repo)

		// Action
		actionStyled := lipgloss.NewStyle().
			Foreground(lipgloss.Color(CurrentTheme.Foreground)).
			Render(activity.Action)

		// Time
		timeAgo := formatTimeAgo(activity.Timestamp)
		timeStyled := lipgloss.NewStyle().
			Foreground(lipgloss.Color(CurrentTheme.Subtle)).
			Render(timeAgo)

		line := fmt.Sprintf("%s %s %s %s",
			iconStyled,
			repoStyled,
			actionStyled,
			timeStyled,
		)

		lines = append(lines, line)

		// Add separator between items (except last)
		if i < len(activities)-1 {
			separator := lipgloss.NewStyle().
				Foreground(lipgloss.Color(CurrentTheme.Subtle)).
				Render(strings.Repeat("─", m.viewport.Width))
			lines = append(lines, separator)
		}
	}

	return strings.Join(lines, "\n")
}
