package main

import (
	"fmt"
	"image"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// loadingState tracks which data is currently being fetched
type loadingState struct {
	profile       bool
	contributions bool
	languages     bool
	repositories  bool
	activities    bool
	avatar        bool
}

// isLoading returns true if any data is still loading
func (ls loadingState) isLoading() bool {
	return ls.profile || ls.contributions || ls.languages || ls.repositories || ls.activities || ls.avatar
}

// PushGranularity represents the time period for push stats
type PushGranularity string

const (
	PushPerHour  PushGranularity = "hour"
	PushPerDay   PushGranularity = "day"
	PushPerWeek  PushGranularity = "week"
	PushPerMonth PushGranularity = "month"
)

// Model represents the application state following Elm architecture
type Model struct {
	username        string
	isOwnProfile    bool // Determined once at startup - viewing authenticated user's profile
	publicOnly      bool // Toggle with 'P' key
	pushGranularity PushGranularity
	client          *GitHubClient
	profile         *ProfileData
	contributions   []Contribution
	languages       []LanguageStats
	repoCount       int
	repositories    []Repository
	activities      []Activity
	avatarImage     image.Image
	graph           *Graph
	viewport        viewport.Model
	spinner         spinner.Model
	loading         loadingState
	err             error
	ready           bool
	width           int
	height          int
}

// Messages for async data fetching
type profileMsg *ProfileData
type contributionsMsg []Contribution
type languagesMsg struct {
	languages []LanguageStats
	repoCount int
}
type repositoriesMsg []Repository
type activitiesMsg []Activity
type avatarMsg image.Image
type errMsg error

// Init initializes the model and kicks off data fetching
func (m Model) Init() tea.Cmd {
	includePrivate := m.isOwnProfile && !m.publicOnly

	// Mark all data as loading
	m.loading = loadingState{
		profile:       true,
		contributions: true,
		languages:     true,
		repositories:  true,
		activities:    true,
	}

	return tea.Batch(
		m.spinner.Tick,
		fetchProfile(m.client, m.username, includePrivate),
		fetchContributions(m.client, m.username),
		fetchLanguages(m.client, m.username, includePrivate),
		fetchRepositories(m.client, m.username, includePrivate),
		fetchActivities(m.client, m.username, includePrivate),
	)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Ignore all keys except quit while loading
		if m.loading.isLoading() && msg.String() != "q" && msg.String() != "ctrl+c" {
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			// Refresh all data
			m.loading = loadingState{
				profile:       true,
				contributions: true,
				languages:     true,
				repositories:  true,
				activities:    true,
			}
			includePrivate := m.isOwnProfile && !m.publicOnly
			return m, tea.Batch(
				fetchProfile(m.client, m.username, includePrivate),
				fetchContributions(m.client, m.username),
				fetchLanguages(m.client, m.username, includePrivate),
				fetchRepositories(m.client, m.username, includePrivate),
				fetchActivities(m.client, m.username, includePrivate),
			)
		case "p", "P":
			// Toggle public/private view (only affects own profile)
			if !m.isOwnProfile {
				return m, nil
			}
			m.publicOnly = !m.publicOnly
			m.loading = loadingState{
				profile:      true,
				languages:    true,
				repositories: true,
				activities:   true,
			}
			includePrivate := !m.publicOnly
			return m, tea.Batch(
				fetchProfile(m.client, m.username, includePrivate),
				fetchLanguages(m.client, m.username, includePrivate),
				fetchRepositories(m.client, m.username, includePrivate),
				fetchActivities(m.client, m.username, includePrivate),
			)
		case "g", "G":
			// Cycle through push granularity (hour -> day -> week -> month -> hour)
			switch m.pushGranularity {
			case PushPerHour:
				m.pushGranularity = PushPerDay
			case PushPerDay:
				m.pushGranularity = PushPerWeek
			case PushPerWeek:
				m.pushGranularity = PushPerMonth
			case PushPerMonth:
				m.pushGranularity = PushPerHour
			default:
				m.pushGranularity = PushPerDay
			}
			return m, nil
		case "t", "T":
			// Cycle through themes
			NextTheme()
			InitStyles()
			m.viewport.Style = lipgloss.NewStyle().
				Foreground(lipgloss.Color(CurrentTheme.Foreground))
			m.viewport.SetContent(m.renderActivityList())
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			m.viewport = viewport.New(m.width, m.calculateActivityViewportHeight())
			m.viewport.Style = lipgloss.NewStyle().
				Foreground(lipgloss.Color(CurrentTheme.Foreground))
			m.ready = true
		} else {
			m.viewport.Width = m.width
			m.viewport.Height = m.calculateActivityViewportHeight()
		}

	case profileMsg:
		m.profile = msg
		m.loading.profile = false
		// Fetch avatar braille art after profile is loaded
		if msg != nil && msg.AvatarURL != "" {
			return m, fetchAvatar(msg.AvatarURL)
		}

	case contributionsMsg:
		m.contributions = msg
		m.graph = NewGraph(msg)
		m.loading.contributions = false

	case languagesMsg:
		m.languages = msg.languages
		m.repoCount = msg.repoCount
		m.loading.languages = false

	case repositoriesMsg:
		m.repositories = msg
		m.loading.repositories = false

	case activitiesMsg:
		m.activities = msg
		m.viewport.SetContent(m.renderActivityList())
		m.viewport.GotoTop()
		m.loading.activities = false

	case avatarMsg:
		m.avatarImage = msg
		m.loading.avatar = false

	case errMsg:
		m.err = msg
		// Clear all loading flags on error
		m.loading = loadingState{}
		return m, nil

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	// Update viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the TUI
func (m Model) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	if m.loading.isLoading() || m.profile == nil {
		return m.renderLoading()
	}

	// Calculate available space
	// Reserve 1 line for status bar, and lines for spacing between sections
	availableHeight := m.height - 1
	availableWidth := m.width

	// Calculate padding based on available space
	vPadding := 1 // Vertical padding between sections
	if m.height < 40 {
		vPadding = 0 // No padding on small terminals
	}

	var sections []string

	// Braille logo at top (right aligned with margin, contrasting background)
	logo := ` ⢀⡀ ⠄ ⣰⡀ ⣰⡀ ⡀⢀ ⠄
 ⣑⡺ ⠇ ⠘⠤ ⠘⠤ ⠣⠼ ⠇`
	styledLogo := lipgloss.NewStyle().
		Width(m.width).
		Background(lipgloss.Color(CurrentTheme.Subtle)).
		Align(lipgloss.Right).
		PaddingTop(1).
		PaddingRight(5).
		Render(accentStyle.Render(logo))
	sections = append(sections, styledLogo)

	// Contribution graph with padding
	graph := m.renderGraphSection(availableWidth)
	if graph != "" {
		graphPadded := lipgloss.NewStyle().
			PaddingBottom(vPadding).
			Render(graph)
		graphHeight := lipgloss.Height(graphPadded)
		sections = append(sections, graphPadded)
		availableHeight -= graphHeight
	}

	// Stats row (languages + streaks side by side) with padding
	// Constrain to graph width for consistency (graph max is 108 chars)
	statsWidth := availableWidth
	const maxGraphWidth = 4 + (52 * 2) // dayLabelWidth + (weeksToDisplay * cellWidth) from contrib_graph.go
	if statsWidth > maxGraphWidth {
		statsWidth = maxGraphWidth
	}
	statsRow := m.renderStatsRow(statsWidth)
	if statsRow != "" {
		statsPadded := lipgloss.NewStyle().
			PaddingBottom(vPadding).
			Render(statsRow)
		statsHeight := lipgloss.Height(statsPadded)
		sections = append(sections, statsPadded)
		availableHeight -= statsHeight
	}

	// Measure bottom section (ASCII art + profile info side by side) and status bar
	bottomSection := m.renderBottomSection(availableWidth)
	bottomHeight := lipgloss.Height(bottomSection)

	statusBar := m.renderStatusBar(availableWidth)
	statusHeight := lipgloss.Height(statusBar)

	// Multi-column row (PRs, Activity, etc.) - scalable for future columns
	if availableHeight > 10 {
		rowHeight := availableHeight - bottomHeight - statusHeight

		// Render multi-column row
		columnsRow := m.renderColumnsRow(availableWidth, rowHeight)
		if columnsRow != "" {
			sections = append(sections, columnsRow)
		}
	}

	// Bottom section: ASCII art/avatar on left, profile info on right
	sections = append(sections, bottomSection)

	// Status bar
	sections = append(sections, statusBar)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderLoading shows loading state with details
func (m Model) renderLoading() string {
	var loading []string

	if m.loading.profile {
		loading = append(loading, "profile")
	}
	if m.loading.contributions {
		loading = append(loading, "contributions")
	}
	if m.loading.languages {
		loading = append(loading, "languages")
	}
	if m.loading.repositories {
		loading = append(loading, "repositories")
	}
	if m.loading.activities {
		loading = append(loading, "activity")
	}

	status := "Initializing..."
	if len(loading) > 0 {
		status = fmt.Sprintf("Loading %s...", strings.Join(loading, ", "))
	}

	msg := fmt.Sprintf("%s %s", m.spinner.View(), status)

	// Center the loading message
	loadingBox := lipgloss.NewStyle().
		Width(80).
		Align(lipgloss.Center).
		Padding(2).
		Render(msg)

	return loadingStyle.Render(loadingBox)
}

// renderBottomSection renders ASCII art/avatar on left, profile info on right
func (m Model) renderBottomSection(width int) string {
	if m.profile == nil {
		return ""
	}

	// Calculate internal padding
	hMargin := 1
	if width > 100 {
		hMargin = 2
	}

	// Profile info box (top layer, on the right)
	profileInfo := m.renderProfileInfo()
	profileWithBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(CurrentTheme.Blue)).
		Padding(1, 2).
		Render(profileInfo)

	// Left side: Avatar and ASCII art (bottom layer)
	var leftComponents []string

	// Add braille avatar if available
	if m.avatarImage != nil {
		renderer := NewColorizedBrailleRenderer(CurrentTheme)
		avatarBraille := renderer.RenderColorized(m.avatarImage)
		leftComponents = append(leftComponents, avatarBraille)
	}

	// Add ASCII username below avatar
	username := m.profile.Login
	maxChars := 20 // Reasonable limit for ASCII art
	if len(username) > maxChars-1 {
		username = username[:maxChars-1]
	}
	asciiName := RenderASCII("@" + strings.ToUpper(username))
	asciiStyled := accentStyle.Render(asciiName)
	leftComponents = append(leftComponents, asciiStyled)

	leftContent := lipgloss.JoinVertical(lipgloss.Left, leftComponents...)

	// Calculate spacing between left content and profile
	spacing := 1 // Minimal spacing to bring profile very close

	// Split into lines and insert profile box
	leftLines := strings.Split(leftContent, "\n")
	profileLines := strings.Split(profileWithBorder, "\n")

	var finalLines []string
	for i := 0; i < len(leftLines); i++ {
		line := leftLines[i]
		if i < len(profileLines) {
			// Calculate position based on actual line width (allows overlap with ASCII art)
			lineWidth := lipgloss.Width(line)
			profileX := lineWidth + spacing

			// Ensure minimum spacing from current line
			if profileX < spacing {
				profileX = spacing
			}

			// Add profile line to the right of this specific line
			paddingNeeded := profileX - lineWidth
			if paddingNeeded > 0 {
				line = line + strings.Repeat(" ", paddingNeeded) + profileLines[i]
			} else {
				// If line is too wide, just add spacing
				line = line + strings.Repeat(" ", spacing) + profileLines[i]
			}
		}
		finalLines = append(finalLines, line)
	}

	// Add remaining profile lines if profile is taller (just add spacing at start)
	for i := len(leftLines); i < len(profileLines); i++ {
		line := strings.Repeat(" ", spacing) + profileLines[i]
		finalLines = append(finalLines, line)
	}

	combined := strings.Join(finalLines, "\n")

	return lipgloss.NewStyle().
		PaddingLeft(hMargin).
		PaddingRight(hMargin).
		Render(combined)
}

// wrapText wraps text to specified width, breaking on word boundaries
func wrapText(text string, width int) []string {
	if len(text) <= width {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)

	currentLine := ""
	for _, word := range words {
		// If adding this word would exceed width
		if len(currentLine)+len(word)+1 > width {
			// If current line is empty and word alone is too long, truncate it
			if currentLine == "" {
				lines = append(lines, word[:width-3]+"...")
				continue
			}
			// Save current line and start new one
			lines = append(lines, currentLine)
			currentLine = word
		} else {
			// Add word to current line
			if currentLine == "" {
				currentLine = word
			} else {
				currentLine += " " + word
			}
		}
	}

	// Add last line
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// renderProfileInfo renders profile information (moved from header)
func (m Model) renderProfileInfo() string {
	if m.profile == nil {
		return ""
	}

	// Set max width for profile box content
	maxWidth := 40

	var info []string

	// Name and username - wrap if needed
	nameText := fmt.Sprintf("%s (@%s)", m.profile.Name, m.profile.Login)
	title := titleStyle.Render(nameText)
	info = append(info, title)
	info = append(info, "")

	// Bio - wrap text to multiple lines
	if m.profile.Bio != "" {
		bioLines := wrapText(m.profile.Bio, maxWidth)
		for _, line := range bioLines {
			info = append(info, baseStyle.Render(line))
		}
		info = append(info, "")
	}

	// Location and company - truncate if needed
	if m.profile.Location != "" {
		loc := m.profile.Location
		if len(loc) > 20 {
			loc = loc[:17] + "..."
		}
		info = append(info, labelStyle.Render("Loc: ")+baseStyle.Render(loc))
	}
	if m.profile.Company != "" {
		comp := m.profile.Company
		if len(comp) > 20 {
			comp = comp[:17] + "..."
		}
		info = append(info, labelStyle.Render("Co: ")+baseStyle.Render(comp))
	}
	if m.profile.Location != "" || m.profile.Company != "" {
		info = append(info, "")
	}

	// Stats - compact format
	repoCount := m.repoCount
	if repoCount == 0 {
		repoCount = m.profile.PublicRepos
	}
	info = append(info, fmt.Sprintf("%s %d | %s %d",
		labelStyle.Render("Repos:"), repoCount,
		labelStyle.Render("Gists:"), m.profile.PublicGists))
	info = append(info, fmt.Sprintf("%s %d | %s %d",
		labelStyle.Render("Followers:"), m.profile.Followers,
		labelStyle.Render("Following:"), m.profile.Following))
	info = append(info, "")

	// Member since
	memberSince := labelStyle.Render("Member: ") +
		baseStyle.Render(m.profile.CreatedAt.Format("Jan 2006"))
	info = append(info, memberSince)

	return lipgloss.JoinVertical(lipgloss.Left, info...)
}

// renderGraphSection renders the contribution graph
func (m Model) renderGraphSection(width int) string {
	// Calculate internal padding
	hMargin := 1
	if width > 100 {
		hMargin = 2
	}

	if m.graph == nil {
		return lipgloss.NewStyle().
			PaddingLeft(hMargin).
			Render(labelStyle.Render("Loading contributions..."))
	}

	graph := m.graph.Render()
	return lipgloss.NewStyle().
		PaddingLeft(hMargin).
		PaddingRight(hMargin).
		Render(graph)
}

// renderStatsRow renders languages, streaks, activity metrics, and top repos in 4 columns
func (m Model) renderStatsRow(width int) string {
	// Calculate horizontal padding based on available space
	hPadding := 2 // Horizontal padding between columns
	if width < 80 {
		hPadding = 1 // Less padding on narrow terminals
	}

	// Divide width into 4 columns with padding between
	availableWidth := width - (hPadding * 3) // 3 gaps between 4 columns
	colWidth := availableWidth / 4

	// Render each column with padding
	languagesContent := m.renderLanguages(colWidth)
	languages := lipgloss.NewStyle().
		PaddingRight(hPadding).
		Render(languagesContent)

	streaksContent := m.renderStreaks(colWidth)
	streaks := lipgloss.NewStyle().
		PaddingRight(hPadding).
		Render(streaksContent)

	activityMetricsContent := m.renderActivityMetrics(colWidth)
	activityMetrics := lipgloss.NewStyle().
		PaddingRight(hPadding).
		Render(activityMetricsContent)

	repos := m.renderTopRepos(colWidth)

	return lipgloss.JoinHorizontal(lipgloss.Top, languages, streaks, activityMetrics, repos)
}

// renderLanguages renders top programming languages with bar charts
func (m Model) renderLanguages(width int) string {
	// Calculate internal padding
	hMargin := 1
	contentWidth := width - (hMargin * 2)

	title := titleStyle.Render("Top Languages")

	if len(m.languages) == 0 {
		content := lipgloss.JoinVertical(lipgloss.Left, title, "", labelStyle.Render("No language data"))
		return lipgloss.NewStyle().
			PaddingLeft(hMargin).
			Render(content)
	}

	var bars []string
	bars = append(bars, title, "")

	// Bar width is proportional to available space
	maxBarWidth := contentWidth - 25 // Reserve space for labels
	if maxBarWidth < 10 {
		maxBarWidth = 10
	}

	// Limit to top 3 languages for consistency
	displayLangs := m.languages
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

	content := lipgloss.JoinVertical(lipgloss.Left, bars...)
	return lipgloss.NewStyle().
		PaddingLeft(hMargin).
		Render(content)
}

// renderStreaks renders contribution streak statistics
func (m Model) renderStreaks(width int) string {
	// Calculate internal padding
	hMargin := 1

	title := titleStyle.Render("Contribution Stats")

	if len(m.contributions) == 0 {
		content := lipgloss.JoinVertical(lipgloss.Left, title, "", labelStyle.Render("No contribution data"))
		return lipgloss.NewStyle().
			PaddingLeft(hMargin).
			Render(content)
	}

	stats := CalculateStats(m.contributions)
	currentStreak := calculateCurrentStreak(m.contributions)
	longestStreak := calculateLongestStreak(m.contributions)

	var lines []string
	lines = append(lines, title, "")

	// Show 3 stats (push rate moved to Activity Metrics column)
	lines = append(lines, labelStyle.Render("Total Contributions"))
	lines = append(lines, accentStyle.Render(fmt.Sprintf("%d", stats.Total)))
	lines = append(lines, "")

	lines = append(lines, labelStyle.Render("Current Streak"))
	lines = append(lines, accentStyle.Render(fmt.Sprintf("%d days", currentStreak)))
	lines = append(lines, "")

	lines = append(lines, labelStyle.Render("Longest Streak"))
	lines = append(lines, accentStyle.Render(fmt.Sprintf("%d days", longestStreak)))
	lines = append(lines, "")

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return lipgloss.NewStyle().
		PaddingLeft(hMargin).
		Render(content)
}

// renderTopRepos renders top repositories by stars
func (m Model) renderTopRepos(width int) string {
	// Calculate internal padding
	hMargin := 1

	title := titleStyle.Render("Top Repositories")

	if len(m.repositories) == 0 {
		content := lipgloss.JoinVertical(lipgloss.Left, title, "", labelStyle.Render("No repositories"))
		return lipgloss.NewStyle().
			PaddingLeft(hMargin).
			Render(content)
	}

	var lines []string
	lines = append(lines, title, "")

	// Limit to top 3 repos for consistency
	displayRepos := m.repositories
	if len(displayRepos) > 3 {
		displayRepos = displayRepos[:3]
	}

	for _, repo := range displayRepos {
		// Repo name (grey like stat labels)
		repoName := repo.Name
		if len(repoName) > width-hMargin-5 {
			repoName = repoName[:width-hMargin-8] + "..."
		}
		lines = append(lines, labelStyle.Render(repoName))

		// Stars and language (green accent like stat values)
		stats := fmt.Sprintf("⭐ %d", repo.Stars)
		if repo.Language != "" {
			stats += fmt.Sprintf(" • %s", repo.Language)
		}
		lines = append(lines, accentStyle.Render(stats))

		// Add spacing between repos
		lines = append(lines, "")
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return lipgloss.NewStyle().
		PaddingLeft(hMargin).
		Render(content)
}

// renderActivityMetrics renders activity-based metrics (push rate, peak hour, etc.)
func (m Model) renderActivityMetrics(width int) string {
	// Calculate internal padding
	hMargin := 1

	title := titleStyle.Render("Activity Metrics")

	if len(m.activities) == 0 {
		content := lipgloss.JoinVertical(lipgloss.Left, title, "", labelStyle.Render("No activity data"))
		return lipgloss.NewStyle().
			PaddingLeft(hMargin).
			Render(content)
	}

	// Calculate stats
	pushRate := CalculatePushStats(m.activities, m.pushGranularity)
	peakHour, _ := CalculatePeakCodingHour(m.activities, ThisWeek)

	var lines []string
	lines = append(lines, title, "")

	// Push rate with granularity label
	var granularityLabel string
	switch m.pushGranularity {
	case PushPerHour:
		granularityLabel = "Pushes/Hour"
	case PushPerDay:
		granularityLabel = "Pushes/Day"
	case PushPerWeek:
		granularityLabel = "Pushes/Week"
	case PushPerMonth:
		granularityLabel = "Pushes/Month"
	}
	lines = append(lines, labelStyle.Render(granularityLabel))
	lines = append(lines, accentStyle.Render(fmt.Sprintf("%.2f", pushRate)))
	lines = append(lines, "")

	// Peak coding hour (always this week)
	lines = append(lines, labelStyle.Render("Peak Hour (This Week)"))
	lines = append(lines, accentStyle.Render(peakHour))
	lines = append(lines, "")

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return lipgloss.NewStyle().
		PaddingLeft(hMargin).
		Render(content)
}

// renderColumnsRow renders a multi-column layout (DRY and scalable)
// Add new columns by adding to the columns slice
func (m Model) renderColumnsRow(width, height int) string {
	// Define columns to render (easily add more here)
	type column struct {
		render func(int, int) string
		name   string
	}

	columns := []column{
		{render: m.renderActivity, name: "activity"},
		// Future: add more columns here
		// {render: m.renderIssues, name: "issues"},
		// {render: m.renderNotifications, name: "notifications"},
	}

	numCols := len(columns)
	if numCols == 0 {
		return ""
	}

	// Calculate responsive horizontal padding between columns
	hPadding := 2
	if width < 80 {
		hPadding = 1
	} else if width > 150 {
		hPadding = 3
	}

	// Calculate column width: (total width - total padding) / number of columns
	totalPadding := hPadding * (numCols - 1)
	colWidth := (width - totalPadding) / numCols

	// Ensure minimum column width
	minColWidth := 40
	if colWidth < minColWidth {
		colWidth = minColWidth
	}

	// Render each column
	var renderedColumns []string
	for i, col := range columns {
		// Render column content
		content := col.render(colWidth, height)

		// Add padding to all columns except the last one
		if i < numCols-1 {
			content = lipgloss.NewStyle().
				PaddingRight(hPadding).
				Render(content)
		}

		renderedColumns = append(renderedColumns, content)
	}

	// Join columns horizontally, aligned at top
	return lipgloss.JoinHorizontal(lipgloss.Top, renderedColumns...)
}

// renderActivity renders recent activity viewport
func (m Model) renderActivity(width, height int) string {
	hMargin := 1
	if width > 100 {
		hMargin = 2
	}

	title := titleStyle.Render(fmt.Sprintf("Recent Activity (%d)", len(m.activities)))

	if !m.ready {
		content := lipgloss.JoinVertical(lipgloss.Left, title, "", labelStyle.Render("Loading..."))
		return lipgloss.NewStyle().
			PaddingLeft(hMargin).
			PaddingRight(hMargin).
			Render(content)
	}

	// Render table header (frozen, outside viewport)
	timeWidth := 10
	eventWidth := 20
	repoWidth := 30
	actionWidth := 40

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.Blue)).
		Bold(true)

	borderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.Gray))

	divider := borderStyle.Render("│")

	// Build header with consistent border styling
	header := fmt.Sprintf("%s %s %s %s %s %s %s",
		headerStyle.Render(fmt.Sprintf("%-*s", timeWidth, "Time")),
		divider,
		headerStyle.Render(fmt.Sprintf("%-*s", eventWidth, "Event")),
		divider,
		headerStyle.Render(fmt.Sprintf("%-*s", repoWidth, "Repository")),
		divider,
		headerStyle.Render(fmt.Sprintf("%-*s", actionWidth, "Action")),
	)

	separator := borderStyle.Render(strings.Repeat("─", timeWidth) + "─┼─" +
		strings.Repeat("─", eventWidth) + "─┼─" +
		strings.Repeat("─", repoWidth) + "─┼─" +
		strings.Repeat("─", actionWidth))

	// Stack: title, header, separator, viewport content
	content := lipgloss.JoinVertical(lipgloss.Left, title, "", header, separator, m.viewport.View())
	return lipgloss.NewStyle().
		PaddingLeft(hMargin).
		PaddingRight(hMargin).
		Render(content)
}

// renderActivityList creates the activity list content with table styling
func (m Model) renderActivityList() string {
	if len(m.activities) == 0 {
		return labelStyle.Render("No recent activity")
	}

	// Column widths (must match renderActivity header)
	timeWidth := 10
	eventWidth := 20
	repoWidth := 30
	actionWidth := 40

	var rows []string

	// Styles for colorizing columns
	timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(CurrentTheme.Cyan))
	eventStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(CurrentTheme.Yellow))
	repoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(CurrentTheme.Green))
	actionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(CurrentTheme.Foreground))
	dividerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(CurrentTheme.Gray))

	// Data rows
	for _, activity := range m.activities {
		timeAgo := formatTimeAgo(activity.Timestamp)
		privacy := ""
		if !activity.Public {
			privacy = "[private] "
		}

		// Truncate to fit columns
		eventType := activity.Type
		if len(eventType) > eventWidth {
			eventType = eventType[:eventWidth-1] + "…"
		}

		repo := privacy + activity.Repo
		if len(repo) > repoWidth {
			repo = repo[:repoWidth-1] + "…"
		}

		action := activity.Action
		if len(action) > actionWidth {
			action = action[:actionWidth-1] + "…"
		}

		// Build row with colorized columns
		line := fmt.Sprintf("%s %s %s %s %s %s %s",
			timeStyle.Render(fmt.Sprintf("%-*s", timeWidth, timeAgo)),
			dividerStyle.Render("│"),
			eventStyle.Render(fmt.Sprintf("%-*s", eventWidth, eventType)),
			dividerStyle.Render("│"),
			repoStyle.Render(fmt.Sprintf("%-*s", repoWidth, repo)),
			dividerStyle.Render("│"),
			actionStyle.Render(fmt.Sprintf("%-*s", actionWidth, action)),
		)

		rows = append(rows, line)
	}

	return strings.Join(rows, "\n")
}


// renderStatusBar renders the bottom status bar with keybindings
func (m Model) renderStatusBar(width int) string {
	// Styles for different parts of status bar
	// Use high-contrast colors against Subtle background
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.Blue)).
		Background(lipgloss.Color(CurrentTheme.Subtle)).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.Foreground)). // Use Foreground for better contrast
		Background(lipgloss.Color(CurrentTheme.Subtle))

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.Green)).
		Background(lipgloss.Color(CurrentTheme.Subtle)).
		Bold(true)

	sepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.Gray)). // Use Gray instead of Dark for better visibility
		Background(lipgloss.Color(CurrentTheme.Subtle))

	// Build status bar with styled components
	var parts []string

	// q: quit
	parts = append(parts, keyStyle.Render("q")+descStyle.Render(": quit"))

	// r: refresh
	parts = append(parts, keyStyle.Render("r")+descStyle.Render(": refresh"))

	// g: cycle push stats
	parts = append(parts, keyStyle.Render("g")+descStyle.Render(": cycle push stats"))

	// t: theme [name] (count)
	themeName := GetCurrentThemeName()
	themeCount := GetThemeCount()
	// Truncate long theme names to prevent wrapping
	if len(themeName) > 20 {
		themeName = themeName[:17] + "..."
	}
	parts = append(parts, keyStyle.Render("t")+descStyle.Render(": theme ")+
		valueStyle.Render(fmt.Sprintf("[%s]", themeName))+
		descStyle.Render(fmt.Sprintf(" (%d)", themeCount)))

	// p: toggle view [mode] (only for own profile)
	if m.isOwnProfile {
		viewMode := "ALL"
		if m.publicOnly {
			viewMode = "PUBLIC"
		}
		parts = append(parts, keyStyle.Render("p")+descStyle.Render(": toggle view ")+
			valueStyle.Render(fmt.Sprintf("[%s]", viewMode)))
	}

	// ↑↓: scroll activity
	parts = append(parts, keyStyle.Render("↑↓")+descStyle.Render(": scroll activity"))

	// Join with separator
	separator := sepStyle.Render(" | ")
	help := strings.Join(parts, separator)

	return lipgloss.NewStyle().
		Width(width).
		Foreground(lipgloss.Color(CurrentTheme.Gray)). // Set default foreground for any gaps
		Background(lipgloss.Color(CurrentTheme.Subtle)).
		Render(help)
}

// Helper functions

// calculateActivityViewportHeight calculates the appropriate viewport height
// based on available terminal space, accounting for all other UI sections
func (m Model) calculateActivityViewportHeight() int {
	if m.height < 20 {
		return 5 // Minimum height for small terminals
	}

	// Estimate space taken by other sections:
	// - Header: ~6 lines
	// - Graph: ~10 lines
	// - Stats row: ~12 lines
	// - ASCII username: ~4 lines
	// - Status bar: 1 line
	// - Activity title + spacing: 2 lines
	// - Vertical padding: ~3 lines
	estimatedOtherContent := 38

	availableHeight := m.height - estimatedOtherContent

	// Enforce reasonable bounds
	if availableHeight < 5 {
		availableHeight = 5
	}
	if availableHeight > 15 {
		availableHeight = 15
	}

	return availableHeight
}

// calculateCurrentStreak calculates the current contribution streak
func calculateCurrentStreak(contributions []Contribution) int {
	if len(contributions) == 0 {
		return 0
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	streak := 0
	expectedDate := today

	// Start from the most recent contribution and work backwards
	for i := len(contributions) - 1; i >= 0; i-- {
		contrib := contributions[i]
		contribDate := time.Date(contrib.Date.Year(), contrib.Date.Month(), contrib.Date.Day(), 0, 0, 0, 0, time.UTC)

		// Check if this is the date we're expecting
		if contribDate.Equal(expectedDate) {
			if contrib.Count > 0 {
				streak++
				expectedDate = expectedDate.AddDate(0, 0, -1) // Move back one day
			} else {
				// Hit a day with 0 contributions
				break
			}
		} else if contribDate.Before(expectedDate) {
			// We've skipped a day - check if it's just today missing
			if streak == 0 && expectedDate.Equal(today) {
				// Today has no contributions yet, start from yesterday
				expectedDate = today.AddDate(0, 0, -1)
				// Re-check this contribution date
				if contribDate.Equal(expectedDate) && contrib.Count > 0 {
					streak++
					expectedDate = expectedDate.AddDate(0, 0, -1)
				} else {
					break
				}
			} else {
				// Streak is broken
				break
			}
		}
	}

	return streak
}

// calculateLongestStreak calculates the longest contribution streak
func calculateLongestStreak(contributions []Contribution) int {
	if len(contributions) == 0 {
		return 0
	}

	longest := 0
	current := 0
	var prevDate time.Time

	for _, contrib := range contributions {
		contribDate := time.Date(contrib.Date.Year(), contrib.Date.Month(), contrib.Date.Day(), 0, 0, 0, 0, time.UTC)

		// Check if this is consecutive from previous date
		if current > 0 {
			expectedDate := prevDate.AddDate(0, 0, 1)
			if !contribDate.Equal(expectedDate) {
				// Gap found, reset streak
				current = 0
			}
		}

		if contrib.Count > 0 {
			current++
			if current > longest {
				longest = current
			}
			prevDate = contribDate
		} else {
			current = 0
		}
	}

	return longest
}

// CalculatePushStats calculates push frequency based on granularity
// Deprecated: Use CalculatePushRate from stats.go instead
func CalculatePushStats(activities []Activity, granularity PushGranularity) float64 {
	return CalculatePushRate(activities, granularity)
}

// formatTimeAgo formats a timestamp as "X ago"
func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		mins := int(duration.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		return fmt.Sprintf("%dh ago", hours)
	case duration < 7*24*time.Hour:
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	default:
		return t.Format("Jan 2")
	}
}

// Async command functions

func fetchProfile(client *GitHubClient, username string, publicOnly bool) tea.Cmd {
	return func() tea.Msg {
		profile, err := client.FetchProfile(username, publicOnly)
		if err != nil {
			return errMsg(err)
		}
		return profileMsg(profile)
	}
}

func fetchContributions(client *GitHubClient, username string) tea.Cmd {
	return func() tea.Msg {
		contributions, err := client.FetchContributions(username)
		if err != nil {
			return errMsg(err)
		}
		return contributionsMsg(contributions)
	}
}

func fetchLanguages(client *GitHubClient, username string, publicOnly bool) tea.Cmd {
	return func() tea.Msg {
		languages, repoCount, err := client.FetchLanguages(username, publicOnly)
		if err != nil {
			return errMsg(err)
		}
		return languagesMsg{
			languages: languages,
			repoCount: repoCount,
		}
	}
}

func fetchRepositories(client *GitHubClient, username string, publicOnly bool) tea.Cmd {
	return func() tea.Msg {
		repositories, err := client.FetchTopRepositories(username, publicOnly)
		if err != nil {
			return errMsg(err)
		}
		return repositoriesMsg(repositories)
	}
}

func fetchActivities(client *GitHubClient, username string, publicOnly bool) tea.Cmd {
	return func() tea.Msg {
		activities, err := client.FetchRecentActivity(username, publicOnly)
		if err != nil {
			return errMsg(err)
		}
		return activitiesMsg(activities)
	}
}

func fetchAvatar(avatarURL string) tea.Cmd {
	return func() tea.Msg {
		// Fetch avatar image (80x80 pixels for larger display)
		img, err := FetchAvatarImage(avatarURL, 80)
		if err != nil {
			// Don't fail the whole app if avatar fails, just return nil
			return avatarMsg(nil)
		}
		return avatarMsg(img)
	}
}

func main() {
	// Recover from panics to restore terminal state
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Fatal error: %v\n", r)
			os.Exit(1)
		}
	}()

	// Initialize theme system (must be first!)
	InitTheme()
	InitStyles()

	// Get username from args or use authenticated user
	username := ""
	if len(os.Args) > 1 {
		username = os.Args[1]
	} else {
		// Get authenticated user from gh CLI
		username = getAuthenticatedUser()
	}

	if username == "" {
		fmt.Println("Usage: gittui [username]")
		fmt.Println("Or run 'gh auth login' to use your authenticated profile")
		os.Exit(1)
	}

	// Create GitHub client
	client, err := NewGitHubClient()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Determine if viewing own profile (check once at startup)
	isOwnProfile := false
	if authUser, err := client.FetchAuthenticatedUser(); err == nil {
		isOwnProfile = (authUser.Login == username)
	}

	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = loadingStyle

	// Create initial model
	m := Model{
		username:        username,
		isOwnProfile:    isOwnProfile,
		publicOnly:      false,          // Default to showing all (private included) for own profile
		pushGranularity: PushPerDay,     // Default to pushes per day
		client:          client,
		loading: loadingState{
			profile:       true,
			contributions: true,
			languages:     true,
			activities:    true,
		},
		spinner: s,
	}

	// Run the program
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// getAuthenticatedUser gets the authenticated username from gh CLI
func getAuthenticatedUser() string {
	client, err := NewGitHubClient()
	if err != nil {
		return ""
	}

	// Fetch authenticated user
	profile, err := client.FetchAuthenticatedUser()
	if err != nil {
		return ""
	}

	return profile.Login
}
