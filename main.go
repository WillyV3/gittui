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
	pullRequests  bool
}

// isLoading returns true if any data is still loading
func (ls loadingState) isLoading() bool {
	return ls.profile || ls.contributions || ls.languages || ls.repositories || ls.activities || ls.avatar || ls.pullRequests
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
	pullRequests    []PullRequest
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
type pullRequestsMsg []PullRequest
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
		pullRequests:  true,
	}

	return tea.Batch(
		m.spinner.Tick,
		fetchProfile(m.client, m.username, includePrivate),
		fetchContributions(m.client, m.username),
		fetchLanguages(m.client, m.username, includePrivate),
		fetchRepositories(m.client, m.username, includePrivate),
		fetchActivities(m.client, m.username, includePrivate),
		fetchPullRequests(m.client, m.username, includePrivate),
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
			// Cycle through themes (github-dark -> dracula -> nord -> github-dark)
			NextTheme()
			InitStyles() // Reinitialize styles with new theme colors
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Viewport will be sized dynamically in View() based on actual available space
		if !m.ready {
			m.viewport = viewport.New(m.width, 10) // Initial size, will be updated
			m.ready = true
		} else {
			m.viewport.Width = m.width
			// Height will be set in renderActivity() based on actual available space
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

	case pullRequestsMsg:
		m.pullRequests = msg
		m.loading.pullRequests = false

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

	// Header section with padding
	header := m.renderHeader(availableWidth)
	if header != "" {
		headerPadded := lipgloss.NewStyle().
			PaddingBottom(vPadding).
			Render(header)
		headerHeight := lipgloss.Height(headerPadded)
		sections = append(sections, headerPadded)
		availableHeight -= headerHeight
	}

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

	// Measure ASCII username and status bar to calculate activity space dynamically
	asciiUsername := m.renderASCIIUsername(availableWidth)
	asciiHeight := lipgloss.Height(asciiUsername)

	statusBar := m.renderStatusBar(availableWidth)
	statusHeight := lipgloss.Height(statusBar)

	// Multi-column row (PRs, Activity, etc.) - scalable for future columns
	if availableHeight > 10 {
		rowHeight := availableHeight - asciiHeight - statusHeight

		// Render multi-column row
		columnsRow := m.renderColumnsRow(availableWidth, rowHeight)
		if columnsRow != "" {
			sections = append(sections, columnsRow)
		}
	}

	// ASCII username above status bar
	sections = append(sections, asciiUsername)

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

// renderHeader renders profile information
func (m Model) renderHeader(width int) string {
	if m.profile == nil {
		return ""
	}

	// Calculate internal padding based on available space
	hMargin := 1
	if width > 100 {
		hMargin = 2
	}

	contentWidth := width - (hMargin * 2)

	// Profile title
	title := titleStyle.Render(fmt.Sprintf("%s (@%s)", m.profile.Name, m.profile.Login))

	// Bio and location
	var info []string
	if m.profile.Bio != "" {
		bio := lipgloss.NewStyle().Width(contentWidth).Render(m.profile.Bio)
		info = append(info, baseStyle.Render(bio))
	}
	var meta []string
	if m.profile.Location != "" {
		meta = append(meta, labelStyle.Render("Location: ")+baseStyle.Render(m.profile.Location))
	}
	if m.profile.Company != "" {
		meta = append(meta, labelStyle.Render("Company: ")+baseStyle.Render(m.profile.Company))
	}
	if len(meta) > 0 {
		info = append(info, strings.Join(meta, " | "))
	}

	// Stats - use actual repo count from fetched repos (includes orgs and private)
	repoCount := m.repoCount
	if repoCount == 0 {
		repoCount = m.profile.PublicRepos
	}

	stats := []string{
		fmt.Sprintf("%s %d", labelStyle.Render("Repos:"), repoCount),
		fmt.Sprintf("%s %d", labelStyle.Render("Gists:"), m.profile.PublicGists),
		fmt.Sprintf("%s %d", labelStyle.Render("Followers:"), m.profile.Followers),
		fmt.Sprintf("%s %d", labelStyle.Render("Following:"), m.profile.Following),
	}
	statsLine := baseStyle.Render(strings.Join(stats, " | "))
	info = append(info, statsLine)

	// Member since
	memberSince := labelStyle.Render("Member since: ") +
		baseStyle.Render(m.profile.CreatedAt.Format("January 2006"))
	info = append(info, memberSince)

	content := lipgloss.JoinVertical(lipgloss.Left, title, strings.Join(info, "\n"))

	// Add horizontal padding
	return lipgloss.NewStyle().
		PaddingLeft(hMargin).
		PaddingRight(hMargin).
		Render(content)
}

// renderASCIIUsername renders the username in ASCII art with braille avatar
func (m Model) renderASCIIUsername(width int) string {
	if m.profile == nil {
		return ""
	}

	// Calculate internal padding
	hMargin := 1
	if width > 100 {
		hMargin = 2
	}

	var components []string

	// Add braille avatar if available (stacked on top)
	if m.avatarImage != nil {
		// Render with current theme colors
		renderer := NewColorizedBrailleRenderer(CurrentTheme)
		avatarBraille := renderer.RenderColorized(m.avatarImage)
		components = append(components, avatarBraille)
	}

	// Add ASCII username below avatar (uppercase for consistent ASCII art rendering)
	// Truncate long usernames to fit terminal width (each char ~4 cols + space)
	username := m.profile.Login
	maxChars := (width - (hMargin * 2)) / 5 // ~5 terminal cols per ASCII char
	if len(username) > maxChars-1 { // -1 for @ symbol
		username = username[:maxChars-1]
	}
	asciiName := RenderASCII("@" + strings.ToUpper(username))
	asciiStyled := accentStyle.Render(asciiName)
	components = append(components, asciiStyled)

	// Stack vertically with avatar on top
	content := lipgloss.JoinVertical(lipgloss.Left, components...)

	return lipgloss.NewStyle().
		PaddingLeft(hMargin).
		PaddingRight(hMargin).
		Render(content)
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

// renderStatsRow renders languages, streaks, and top repos in 3 columns
func (m Model) renderStatsRow(width int) string {
	// Calculate horizontal padding based on available space
	hPadding := 2 // Horizontal padding between columns
	if width < 80 {
		hPadding = 1 // Less padding on narrow terminals
	}

	// Divide width into 3 columns with padding between
	availableWidth := width - (hPadding * 2) // 2 gaps between 3 columns
	colWidth := availableWidth / 3

	// Render each column with padding
	languagesContent := m.renderLanguages(colWidth)
	languages := lipgloss.NewStyle().
		PaddingRight(hPadding).
		Render(languagesContent)

	streaksContent := m.renderStreaks(colWidth)
	streaks := lipgloss.NewStyle().
		PaddingRight(hPadding).
		Render(streaksContent)

	repos := m.renderTopRepos(colWidth)

	return lipgloss.JoinHorizontal(lipgloss.Top, languages, streaks, repos)
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
	pushRate := CalculatePushStats(m.activities, m.pushGranularity)

	var lines []string
	lines = append(lines, title, "")

	// Show 4 stats
	lines = append(lines, labelStyle.Render("Total Contributions"))
	lines = append(lines, accentStyle.Render(fmt.Sprintf("%d", stats.Total)))
	lines = append(lines, "")

	lines = append(lines, labelStyle.Render("Current Streak"))
	lines = append(lines, accentStyle.Render(fmt.Sprintf("%d days", currentStreak)))
	lines = append(lines, "")

	lines = append(lines, labelStyle.Render("Longest Streak"))
	lines = append(lines, accentStyle.Render(fmt.Sprintf("%d days", longestStreak)))
	lines = append(lines, "")

	// Push frequency stat with configurable granularity
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

// renderPullRequests renders the PR dashboard
func (m Model) renderPullRequests(width, height int) string {
	hMargin := 1
	if width > 100 {
		hMargin = 2
	}

	// Determine if we're showing open or closed PRs
	var statusLabel string
	if len(m.pullRequests) > 0 {
		if m.pullRequests[0].State == "open" {
			statusLabel = fmt.Sprintf("Pull Requests (%d open)", len(m.pullRequests))
		} else {
			statusLabel = "Pull Requests (recent closed)"
		}
	} else {
		statusLabel = "Pull Requests"
	}

	title := titleStyle.Render(statusLabel)

	if len(m.pullRequests) == 0 {
		content := lipgloss.JoinVertical(lipgloss.Left, title, "", labelStyle.Render("No pull requests"))
		return lipgloss.NewStyle().
			PaddingLeft(hMargin).
			PaddingRight(hMargin).
			Render(content)
	}

	var lines []string
	lines = append(lines, title, "")

	// Show up to 5 PRs
	displayPRs := m.pullRequests
	if len(displayPRs) > 5 {
		displayPRs = displayPRs[:5]
	}

	for _, pr := range displayPRs {
		// Status indicator
		var statusIcon string
		var statusColor lipgloss.Style

		// Closed/Merged PRs have priority over review status
		if pr.State == "closed" {
			if pr.IsMerged() {
				statusIcon = "✓"
				statusColor = lipgloss.NewStyle().Foreground(lipgloss.Color(CurrentTheme.Purple)) // Merged
			} else {
				statusIcon = "✗"
				statusColor = lipgloss.NewStyle().Foreground(lipgloss.Color(CurrentTheme.Red)) // Closed without merge
			}
		} else if pr.Draft {
			statusIcon = "◐"
			statusColor = lipgloss.NewStyle().Foreground(lipgloss.Color(CurrentTheme.Dark)) // Draft
		} else {
			// Open PR - check review status
			switch pr.ReviewDecision {
			case "APPROVED":
				statusIcon = "✓"
				statusColor = lipgloss.NewStyle().Foreground(lipgloss.Color(CurrentTheme.Green)) // Approved
			case "CHANGES_REQUESTED":
				statusIcon = "⚠"
				statusColor = lipgloss.NewStyle().Foreground(lipgloss.Color(CurrentTheme.Red)) // Changes requested
			case "REVIEW_REQUIRED":
				statusIcon = "⏳"
				statusColor = lipgloss.NewStyle().Foreground(lipgloss.Color(CurrentTheme.Yellow)) // Pending review
			default:
				statusIcon = "●"
				statusColor = lipgloss.NewStyle().Foreground(lipgloss.Color(CurrentTheme.Gray)) // No reviews
			}
		}

		// PR title (truncate if too long)
		prTitle := pr.Title
		maxTitleWidth := width - hMargin*2 - 20
		if len(prTitle) > maxTitleWidth {
			prTitle = prTitle[:maxTitleWidth-3] + "..."
		}

		// PR line: [icon] title
		prLine := statusColor.Render(statusIcon) + " " + baseStyle.Render(prTitle)
		lines = append(lines, prLine)

		// Details line: repo • reviews
		details := labelStyle.Render(pr.Repo.FullName)
		if pr.ApprovedCount > 0 {
			details += accentStyle.Render(fmt.Sprintf("  ✓ %d approved", pr.ApprovedCount))
		}
		if pr.ChangesCount > 0 {
			details += statusColor.Render(fmt.Sprintf("  ⚠ %d changes", pr.ChangesCount))
		}
		if pr.ApprovedCount == 0 && pr.ChangesCount == 0 && pr.CommentCount > 0 {
			details += labelStyle.Render(fmt.Sprintf("  💬 %d comments", pr.CommentCount))
		}
		if pr.ReviewDecision == "" {
			details += subtleStyle.Render("  ⏳ awaiting review")
		}

		lines = append(lines, "  "+details)
		lines = append(lines, "") // Spacing
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return lipgloss.NewStyle().
		PaddingLeft(hMargin).
		PaddingRight(hMargin).
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
		{render: m.renderPullRequests, name: "prs"},
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

// renderActivity renders recent activity timeline
func (m Model) renderActivity(width, height int) string {
	// Calculate internal padding
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

	// Update viewport height to match actual available space
	// Title + blank line = 2 lines
	viewportHeight := height - 2
	if viewportHeight < 5 {
		viewportHeight = 5
	}
	m.viewport.Height = viewportHeight

	content := lipgloss.JoinVertical(lipgloss.Left, title, "", m.viewport.View())
	return lipgloss.NewStyle().
		PaddingLeft(hMargin).
		PaddingRight(hMargin).
		Render(content)
}

// renderActivityList creates the activity list content
func (m Model) renderActivityList() string {
	if len(m.activities) == 0 {
		return labelStyle.Render("No recent activity")
	}

	// Calculate max items that can fit in viewport
	// Title takes 2 lines, each activity is 1 line
	maxItems := m.viewport.Height - 2

	// Enforce bounds: minimum 5, maximum 10
	if maxItems < 5 {
		maxItems = 5
	}
	if maxItems > 10 {
		maxItems = 10
	}

	// Only show what fits
	displayActivities := m.activities
	if len(displayActivities) > maxItems {
		displayActivities = displayActivities[:maxItems]
	}

	var items []string
	for _, activity := range displayActivities {
		// Format timestamp
		timeAgo := formatTimeAgo(activity.Timestamp)

		// Privacy indicator
		privacy := ""
		if !activity.Public {
			privacy = subtleStyle.Render("[private] ")
		}

		// Format line
		line := fmt.Sprintf("%s %s %s%s %s",
			subtleStyle.Render(timeAgo),
			labelStyle.Render(activity.Type),
			privacy,
			baseStyle.Render(activity.Repo),
			subtleStyle.Render(activity.Action),
		)
		items = append(items, line)
	}

	return strings.Join(items, "\n")
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
	parts = append(parts, keyStyle.Render("t")+descStyle.Render(": theme ")+
		valueStyle.Render(fmt.Sprintf("[%s]", themeName))+
		descStyle.Render(fmt.Sprintf(" (%d available)", themeCount)))

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
func CalculatePushStats(activities []Activity, granularity PushGranularity) float64 {
	if len(activities) == 0 {
		return 0.0
	}

	// Count push events
	pushCount := 0
	var oldestPush, newestPush time.Time

	for _, activity := range activities {
		if activity.Type == "PushEvent" {
			pushCount++
			if oldestPush.IsZero() || activity.Timestamp.Before(oldestPush) {
				oldestPush = activity.Timestamp
			}
			if newestPush.IsZero() || activity.Timestamp.After(newestPush) {
				newestPush = activity.Timestamp
			}
		}
	}

	if pushCount == 0 {
		return 0.0
	}

	// Calculate time span
	timeSpan := newestPush.Sub(oldestPush)
	if timeSpan == 0 {
		// If all pushes are at the same time, default to 1 unit of the granularity
		switch granularity {
		case PushPerHour:
			timeSpan = time.Hour
		case PushPerDay:
			timeSpan = 24 * time.Hour
		case PushPerWeek:
			timeSpan = 7 * 24 * time.Hour
		case PushPerMonth:
			timeSpan = 30 * 24 * time.Hour
		}
	}

	// Calculate rate based on granularity
	// Always calculate based on actual time span and scale to desired unit
	var rate float64
	hours := timeSpan.Hours()
	if hours == 0 {
		hours = 1 // Avoid division by zero
	}

	switch granularity {
	case PushPerHour:
		rate = float64(pushCount) / hours
	case PushPerDay:
		days := hours / 24
		rate = float64(pushCount) / days
	case PushPerWeek:
		weeks := hours / (24 * 7)
		rate = float64(pushCount) / weeks
	case PushPerMonth:
		months := hours / (24 * 30)
		rate = float64(pushCount) / months
	default:
		rate = float64(pushCount)
	}

	return rate
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

func fetchPullRequests(client *GitHubClient, username string, publicOnly bool) tea.Cmd {
	return func() tea.Msg {
		prs, err := client.FetchPullRequests(username, publicOnly)
		if err != nil {
			return errMsg(err)
		}
		return pullRequestsMsg(prs)
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
