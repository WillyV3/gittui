// Package main provides a GitHub-style contribution graph renderer using Bubble Tea and Lipgloss.
// This is a standalone component that can be imported and used in any terminal application.
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Constants define the graph dimensions and styling.
const (
	weeksToDisplay     = 52
	daysPerWeek        = 7
	cellWidth          = 2 // Character width per cell (block + space)
	dayLabelWidth      = 4
	minTerminalWidth   = dayLabelWidth + (weeksToDisplay * cellWidth)
	colorLevelZero     = "#161b22" // No contributions
	colorLevelOne      = "#0e4429" // 1-3 contributions
	colorLevelTwo      = "#006d32" // 4-6 contributions
	colorLevelThree    = "#26a641" // 7-9 contributions
	colorLevelFour     = "#39d353" // 10+ contributions
	colorTitle         = "#58a6ff"
	colorLabel         = "#7d8590"
	colorBorder        = "#58a6ff"
	colorWarning       = "#484f58"
	blockChar          = "▄" // Lower half-height block for compact squares
)

// Contribution represents a single day's contribution count.
type Contribution struct {
	Date  time.Time
	Count int
}

// Graph represents a contribution graph with all necessary data.
type Graph struct {
	contributions []Contribution
	title         string
	showLegend    bool
}

// NewGraph creates a new contribution graph from a slice of contributions.
func NewGraph(contributions []Contribution) *Graph {
	return &Graph{
		contributions: contributions,
		title:         "Contribution Activity",
		showLegend:    true,
	}
}

// Render generates the complete contribution graph as a string.
// Returns the full graph if terminal is wide enough, otherwise shows a width warning.
func (g *Graph) Render() string {
	grid := g.buildGrid()

	var output strings.Builder

	output.WriteString(g.renderTitle() + "\n\n")
	output.WriteString(g.renderMonthLabels() + "\n")
	output.WriteString(g.renderGrid(grid))

	if g.showLegend {
		output.WriteString("\n" + g.renderLegend() + "\n")
	}

	return output.String()
}

// RenderResponsive renders the graph or a width warning based on terminal width.
func (g *Graph) RenderResponsive(terminalWidth int) string {
	if terminalWidth < minTerminalWidth {
		return renderWidthWarning(terminalWidth, minTerminalWidth)
	}
	return g.Render()
}

// buildGrid converts the linear contribution data into a 2D week-by-day grid.
// Grid layout: [day of week][week number] where Sunday = 0.
func (g *Graph) buildGrid() [daysPerWeek][weeksToDisplay]int {
	var grid [daysPerWeek][weeksToDisplay]int

	if len(g.contributions) == 0 {
		return grid
	}

	// Find the Sunday before the first contribution
	startDate := g.contributions[0].Date
	for startDate.Weekday() != time.Sunday {
		startDate = startDate.AddDate(0, 0, -1)
	}

	// Fill grid with contribution counts
	for _, contrib := range g.contributions {
		daysSinceStart := int(contrib.Date.Sub(startDate).Hours() / 24)
		week := daysSinceStart / 7
		dayOfWeek := int(contrib.Date.Weekday())

		if week >= 0 && week < weeksToDisplay && dayOfWeek >= 0 && dayOfWeek < daysPerWeek {
			grid[dayOfWeek][week] = contrib.Count
		}
	}

	return grid
}

// renderTitle returns the styled graph title.
func (g *Graph) renderTitle() string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colorTitle)).
		Render(g.title)
}

// renderMonthLabels generates the month name row across the top of the graph.
func (g *Graph) renderMonthLabels() string {
	if len(g.contributions) == 0 {
		return ""
	}

	months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun",
	                   "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

	// Find the Sunday before the first contribution
	startDate := g.contributions[0].Date
	for startDate.Weekday() != time.Sunday {
		startDate = startDate.AddDate(0, 0, -1)
	}

	// Build label row as character array for precise positioning
	totalWidth := weeksToDisplay * cellWidth
	labelChars := make([]rune, totalWidth)
	for i := range labelChars {
		labelChars[i] = ' '
	}

	// Place month labels at week positions where month changes
	currentMonth := -1
	for week := 0; week < weeksToDisplay; week++ {
		weekStart := startDate.AddDate(0, 0, week*7)
		month := int(weekStart.Month()) - 1

		if month != currentMonth {
			pos := week * cellWidth
			monthName := months[month]
			for i, ch := range monthName {
				if pos+i < len(labelChars) {
					labelChars[pos+i] = ch
				}
			}
			currentMonth = month
		}
	}

	styled := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorLabel)).
		Render(string(labelChars))

	return strings.Repeat(" ", dayLabelWidth) + styled
}

// renderGrid generates all 7 day rows with their contribution cells.
func (g *Graph) renderGrid(grid [daysPerWeek][weeksToDisplay]int) string {
	var rows []string

	dayLabels := []string{"", "Mon", "", "Wed", "", "Fri", ""}

	for day := 0; day < daysPerWeek; day++ {
		var row strings.Builder

		// Add day label
		label := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorLabel)).
			Width(dayLabelWidth).
			Render(dayLabels[day])
		row.WriteString(label)

		// Add cells for each week
		for week := 0; week < weeksToDisplay; week++ {
			count := grid[day][week]
			row.WriteString(renderCell(count))
		}

		rows = append(rows, row.String())
	}

	return strings.Join(rows, "\n")
}

// renderCell creates a single contribution cell with color based on count.
func renderCell(count int) string {
	level := getContributionLevel(count)
	color := getColorForLevel(level)

	block := lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Render(blockChar)

	return block + " " // Block + space for horizontal separation
}

// getContributionLevel maps a contribution count to an intensity level (0-4).
func getContributionLevel(count int) int {
	switch {
	case count == 0:
		return 0
	case count <= 3:
		return 1
	case count <= 6:
		return 2
	case count <= 9:
		return 3
	default:
		return 4
	}
}

// getColorForLevel returns the GitHub green color for a given intensity level.
func getColorForLevel(level int) string {
	colors := []string{
		colorLevelZero,
		colorLevelOne,
		colorLevelTwo,
		colorLevelThree,
		colorLevelFour,
	}

	if level >= 0 && level < len(colors) {
		return colors[level]
	}
	return colorLevelZero
}

// renderLegend creates the "Less -> More" color scale indicator.
func (g *Graph) renderLegend() string {
	var parts []string

	parts = append(parts, lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorLabel)).
		Render("Less"))

	for level := 0; level < 5; level++ {
		color := getColorForLevel(level)
		cell := lipgloss.NewStyle().
			Foreground(lipgloss.Color(color)).
			Render(blockChar + " ")
		parts = append(parts, cell)
	}

	parts = append(parts, lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorLabel)).
		Render("More"))

	return strings.Repeat(" ", dayLabelWidth) + strings.Join(parts, "")
}

// renderWidthWarning displays a helpful message when the terminal is too narrow.
// The message adapts to the available width.
func renderWidthWarning(currentWidth, minWidth int) string {
	maxBoxWidth := currentWidth - 4 // Leave margin for box border

	message := "Increase terminal width to view contribution grid"
	detail := fmt.Sprintf("Need %d columns, have %d", minWidth, currentWidth)

	// Adapt message to available width
	if maxBoxWidth < len(message) {
		if maxBoxWidth < 30 {
			message = "Terminal too narrow"
			detail = ""
		} else if maxBoxWidth < 50 {
			message = "Increase width for graph"
			detail = ""
		}
	}

	// Build styled box
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colorBorder)).
		Padding(1, 2).
		Width(maxBoxWidth).
		Align(lipgloss.Center)

	content := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorLabel)).
		Bold(true).
		Render(message)

	if detail != "" {
		detailStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorWarning)).
			Render(detail)
		content = lipgloss.JoinVertical(lipgloss.Center, content, "", detailStyle)
	}

	return style.Render(content)
}

// Stats calculates summary statistics for a set of contributions.
type Stats struct {
	Total       int
	ActiveDays  int
	TotalDays   int
	AverageDay  int
	MaxDay      int
}

// CalculateStats computes statistics from a slice of contributions.
func CalculateStats(contributions []Contribution) Stats {
	stats := Stats{
		TotalDays: len(contributions),
	}

	for _, contrib := range contributions {
		stats.Total += contrib.Count
		if contrib.Count > 0 {
			stats.ActiveDays++
		}
		if contrib.Count > stats.MaxDay {
			stats.MaxDay = contrib.Count
		}
	}

	if stats.TotalDays > 0 {
		stats.AverageDay = stats.Total / stats.TotalDays
	}

	return stats
}

// FormatStats returns a human-readable string of contribution statistics.
func (s Stats) String() string {
	return fmt.Sprintf(
		"Total: %d contributions | Active: %d/%d days | Avg: %d/day | Max: %d/day",
		s.Total, s.ActiveDays, s.TotalDays, s.AverageDay, s.MaxDay,
	)
}
