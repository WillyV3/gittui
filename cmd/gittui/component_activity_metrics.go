package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StreaksComponent displays contribution streaks
type StreaksComponent struct {
	contributions []Contribution
	loading       bool
}

// NewStreaksComponent creates a new streaks component
func NewStreaksComponent() StreaksComponent {
	return StreaksComponent{
		loading: true,
	}
}

// SetContributions updates the contribution data
func (s *StreaksComponent) SetContributions(contributions []Contribution) {
	s.contributions = contributions
	s.loading = false
}

// Update handles messages
func (s StreaksComponent) Update(msg tea.Msg) (StreaksComponent, tea.Cmd) {
	return s, nil
}

// View renders the streaks component
func (s StreaksComponent) View() string {
	title := titleStyle.Render("Streaks")

	if s.loading {
		return lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			labelStyle.Render("Loading..."))
	}

	// Calculate streaks
	currentStreak := calculateCurrentStreak(s.contributions)
	longestStreak := calculateLongestStreak(s.contributions)

	var lines []string
	lines = append(lines, title, "")

	// Current streak
	lines = append(lines, labelStyle.Render("Current Streak"))
	lines = append(lines, accentStyle.Render(fmt.Sprintf("%d days", currentStreak)))
	lines = append(lines, "")

	// Longest streak
	lines = append(lines, labelStyle.Render("Longest Streak"))
	lines = append(lines, accentStyle.Render(fmt.Sprintf("%d days", longestStreak)))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// ActivityMetricsComponent displays activity metrics
type ActivityMetricsComponent struct {
	activities      []Activity
	pushGranularity PushGranularity
	loading         bool
}

// NewActivityMetricsComponent creates a new activity metrics component
func NewActivityMetricsComponent() ActivityMetricsComponent {
	return ActivityMetricsComponent{
		pushGranularity: PushPerDay,
		loading:         true,
	}
}

// SetActivities updates the activity data
func (a *ActivityMetricsComponent) SetActivities(activities []Activity) {
	a.activities = activities
	a.loading = false
}

// SetPushGranularity updates the push granularity setting
func (a *ActivityMetricsComponent) SetPushGranularity(granularity PushGranularity) {
	a.pushGranularity = granularity
}

// Update handles messages
func (a ActivityMetricsComponent) Update(msg tea.Msg) (ActivityMetricsComponent, tea.Cmd) {
	return a, nil
}

// View renders the activity metrics component
func (a ActivityMetricsComponent) View() string {
	title := titleStyle.Render("Activity Metrics")

	if a.loading {
		return lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			labelStyle.Render("Loading..."))
	}

	if len(a.activities) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			labelStyle.Render("No activity data"))
	}

	// Calculate stats
	pushRate := CalculatePushRate(a.activities, a.pushGranularity)
	peakHour, _ := CalculatePeakCodingHour(a.activities, ThisWeek)

	var lines []string
	lines = append(lines, title, "")

	// Push rate with granularity label and calculation hint
	var granularityLabel, hint string
	switch a.pushGranularity {
	case PushPerHour:
		granularityLabel = "Pushes/Hour"
		hint = "(7-day avg ÷ 24)"
	case PushPerDay:
		granularityLabel = "Pushes/Day"
		hint = "(7-day avg)"
	case PushPerWeek:
		granularityLabel = "Pushes/Week"
		hint = "(7-day avg × 7)"
	case PushPerMonth:
		granularityLabel = "Pushes/Month"
		hint = "(7-day avg × 30)"
	}

	lines = append(lines, labelStyle.Render(granularityLabel))
	lines = append(lines, accentStyle.Render(fmt.Sprintf("%.2f", pushRate)))
	lines = append(lines, dimStyle.Render(hint))
	lines = append(lines, "")

	// Peak coding hour
	lines = append(lines, labelStyle.Render("Peak Hour (This Week)"))
	lines = append(lines, accentStyle.Render(peakHour))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
