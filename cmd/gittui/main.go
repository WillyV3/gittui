package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// clearNotificationMsg is sent after notification timeout
type clearNotificationMsg time.Time

// clearNotificationAfter returns a command that clears notification after delay
func clearNotificationAfter(delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(t time.Time) tea.Msg {
		return clearNotificationMsg(t)
	})
}

// Model is the main application model using composable components
type Model struct {
	username        string
	isOwnProfile    bool
	publicOnly      bool
	pushGranularity PushGranularity
	client          *GitHubClient

	// Composable components
	profile         ProfileComponent
	graph           GraphComponent
	languages       LanguagesComponent
	repos           ReposComponent
	streaks         StreaksComponent
	activityMetrics ActivityMetricsComponent
	activityModal   ActivityModal
	helpModal       HelpModal
	// avatarTuner    AvatarTunerModal // Dev tool only - not shipped

	// Avatar management
	avatarSources     []string          // Available avatar sources: ["personal", "org1", "org2"]
	avatarURLs        map[string]string // Map of source name -> avatar URL
	personalAvatarURL string            // User's personal avatar URL

	// UI state
	spinner      spinner.Model
	loading      loadingState
	err          error
	width        int
	height       int
	notification string // Temporary notification message
}

// NewModel creates a new model with components
func NewModel(username string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(CurrentTheme.Blue))

	client, err := NewGitHubClient()
	if err != nil {
		// Return a model with error state
		return Model{
			err: err,
		}
	}

	// Determine if viewing own profile by checking authenticated user
	// go-gh library handles auth automatically (GITHUB_TOKEN, GH_TOKEN, or gh CLI)
	isOwnProfile := false
	authUser, err := client.GetAuthenticatedUser()
	if err == nil && authUser != "" {
		isOwnProfile = (authUser == username)
	}

	// Initialize avatar sources (personal + configured orgs)
	config, _ := LoadConfig()
	avatarSources := []string{"personal"}
	avatarSources = append(avatarSources, config.AllowedOrgs...)
	avatarURLs := make(map[string]string)

	return Model{
		username:        username,
		isOwnProfile:    isOwnProfile,
		publicOnly:      false,
		pushGranularity: PushPerDay,
		client:          client,

		// Initialize components
		profile:         NewProfileComponent(),
		graph:           NewGraphComponent(),
		languages:       NewLanguagesComponent(),
		repos:           NewReposComponent(),
		streaks:         NewStreaksComponent(),
		activityMetrics: NewActivityMetricsComponent(),
		activityModal:   NewActivityModal(),
		helpModal:       NewHelpModal(),
		// avatarTuner:     NewAvatarTunerModal(), // Dev tool only

		// Avatar management
		avatarSources: avatarSources,
		avatarURLs:    avatarURLs,

		spinner: s,
		loading: loadingState{
			profile:       true,
			contributions: true,
			languages:     true,
			repositories:  true,
			activities:    true,
			avatar:        true,
		},
	}
}

// Init initializes the model and starts data fetching
func (m Model) Init() tea.Cmd {
	includePrivate := m.isOwnProfile && !m.publicOnly
	cmds := []tea.Cmd{
		m.spinner.Tick,
		fetchProfile(m.client, m.username, includePrivate),
		fetchContributions(m.client, m.username),
		fetchLanguages(m.client, m.username, includePrivate),
		fetchRepositories(m.client, m.username, includePrivate),
		fetchActivities(m.client, m.username, includePrivate),
	}

	// Fetch org avatars ONLY if viewing own profile
	if m.isOwnProfile {
		config, _ := LoadConfig()
		if len(config.AllowedOrgs) > 0 {
			cmds = append(cmds, fetchOrgAvatars(m.client, config.AllowedOrgs))
		}
	}

	return tea.Batch(cmds...)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If help modal is visible, let it handle the key events first
		if m.helpModal.IsVisible() {
			m.helpModal, cmd = m.helpModal.Update(msg)
			return m, cmd
		}

		// If activity modal is visible, let it handle the key events
		if m.activityModal.IsVisible() {
			m.activityModal, cmd = m.activityModal.Update(msg)
			return m, cmd
		}

		// Ignore all keys except quit and help while loading
		if m.loading.isLoading() && msg.String() != "q" && msg.String() != "ctrl+c" && msg.String() != "?" && msg.String() != "h" {
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "?", "h":
			// Toggle help modal
			if m.helpModal.IsVisible() {
				m.helpModal.Hide()
			} else {
				// Update context before showing
				m.helpModal.SetContext(m.isOwnProfile, len(m.avatarSources) > 1)
				m.helpModal.Show()
			}
			return m, nil
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
			// Cycle through push granularity
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
			m.activityMetrics.SetPushGranularity(m.pushGranularity)
			return m, nil
		case "t":
			// Cycle through themes
			NextTheme()
			InitStyles()
			return m, nil
		case "T":
			// Save current theme as default
			config, err := LoadConfig()
			if err != nil {
				config = DefaultConfig()
			}
			themeName := GetCurrentThemeName()
			config.Theme = themeName
			if err := SaveConfig(config); err != nil {
				m.notification = fmt.Sprintf("❌ Failed to save theme: %v", err)
			} else {
				m.notification = fmt.Sprintf("✓ Theme '%s' saved as default", themeName)
			}
			return m, clearNotificationAfter(2 * time.Second)
		case "-":
			// Decrease avatar size (only in normal mode)
			if m.width >= 100 {
				config, err := LoadConfig()
				if err != nil {
					config = DefaultConfig()
				}
				newSize := config.AvatarWidth - 10
				if newSize < 30 {
					newSize = 30
				}
				config.AvatarWidth = newSize
				if err := SaveConfig(config); err != nil {
					m.notification = fmt.Sprintf("❌ Failed to save: %v", err)
				} else {
					m.notification = fmt.Sprintf("✓ Avatar: %d (press 'r' to apply)", newSize)
				}
				return m, clearNotificationAfter(2 * time.Second)
			}
			return m, nil
		case "=", "+":
			// Increase avatar size (only in normal mode)
			if m.width >= 100 {
				config, err := LoadConfig()
				if err != nil {
					config = DefaultConfig()
				}
				newSize := config.AvatarWidth + 10
				if newSize > 240 {
					newSize = 240
				}
				config.AvatarWidth = newSize
				if err := SaveConfig(config); err != nil {
					m.notification = fmt.Sprintf("❌ Failed to save: %v", err)
				} else {
					m.notification = fmt.Sprintf("✓ Avatar: %d (press 'r' to apply)", newSize)
				}
				return m, clearNotificationAfter(2 * time.Second)
			}
			return m, nil
		case "v", "V":
			// Cycle through avatar sources (only in normal mode and own profile)
			if m.width >= 100 && m.isOwnProfile && len(m.avatarSources) > 1 {
				config, err := LoadConfig()
				if err != nil {
					config = DefaultConfig()
				}

				// Find current source index
				currentIdx := 0
				for i, source := range m.avatarSources {
					if source == config.AvatarSource {
						currentIdx = i
						break
					}
				}

				// Move to next source
				nextIdx := (currentIdx + 1) % len(m.avatarSources)
				nextSource := m.avatarSources[nextIdx]

				// Save to config
				config.AvatarSource = nextSource
				if err := SaveConfig(config); err != nil {
					m.notification = fmt.Sprintf("❌ Failed to save: %v", err)
					return m, clearNotificationAfter(2 * time.Second)
				}

				// Get the avatar URL for this source
				avatarURL, exists := m.avatarURLs[nextSource]
				if !exists {
					// Fallback to personal if URL not found
					avatarURL = m.personalAvatarURL
				}

				// Reload avatar with current width
				avatarWidth := config.AvatarWidth
				if avatarWidth < 30 || avatarWidth > 240 {
					avatarWidth = 60
				}

				// Display notification
				displayName := nextSource
				if nextSource == "personal" {
					displayName = "Personal"
				}
				m.notification = fmt.Sprintf("✓ Avatar: %s", displayName)

				return m, tea.Batch(
					fetchAvatarWithWidth(avatarURL, avatarWidth),
					clearNotificationAfter(2*time.Second),
				)
			}
			return m, nil
		case "b", "B":
			// Toggle avatar renderer (only in normal mode)
			if m.width >= 100 {
				config, err := LoadConfig()
				if err != nil {
					config = DefaultConfig()
				}

				// Toggle between mosaic and ascii
				var newRenderer string
				if config.AvatarRenderer == "mosaic" || config.AvatarRenderer == "" {
					newRenderer = "ascii"
				} else {
					newRenderer = "mosaic"
				}

				// Save to config
				config.AvatarRenderer = newRenderer
				if err := SaveConfig(config); err != nil {
					m.notification = fmt.Sprintf("❌ Failed to save: %v", err)
					return m, clearNotificationAfter(2 * time.Second)
				}

				// Get current avatar URL
				avatarURL := m.personalAvatarURL
				if config.AvatarSource != "" && config.AvatarSource != "personal" {
					if orgURL, exists := m.avatarURLs[config.AvatarSource]; exists {
						avatarURL = orgURL
					}
				}

				// Reload avatar with current width
				avatarWidth := config.AvatarWidth
				if avatarWidth < 30 || avatarWidth > 240 {
					avatarWidth = 60
				}

				// Display notification
				m.notification = fmt.Sprintf("✓ Renderer: %s", newRenderer)

				return m, tea.Batch(
					fetchAvatarWithWidth(avatarURL, avatarWidth),
					clearNotificationAfter(2*time.Second),
				)
			}
			return m, nil
		case "a", "A":
			// Toggle activity modal
			if m.activityModal.IsVisible() {
				m.activityModal.Hide()
			} else {
				m.activityModal.Show()
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update components
		m.graph, cmd = m.graph.Update(msg)
		cmds = append(cmds, cmd)

		// Update modal sizes
		m.activityModal, cmd = m.activityModal.Update(msg)
		cmds = append(cmds, cmd)

		m.helpModal, cmd = m.helpModal.Update(msg)
		cmds = append(cmds, cmd)

		return m, tea.Batch(cmds...)

	case profileMsg:
		m.profile.SetProfile(msg)
		m.loading.profile = false
		// Store personal avatar URL and fetch avatar after profile is loaded
		if msg != nil && msg.AvatarURL != "" {
			m.personalAvatarURL = msg.AvatarURL
			m.avatarURLs["personal"] = msg.AvatarURL

			// Determine which avatar to load
			avatarURL := msg.AvatarURL // Default to their avatar

			// ONLY apply avatar source config if viewing own profile
			if m.isOwnProfile {
				config, _ := LoadConfig()
				// Check if a different avatar source is configured
				if config.AvatarSource != "" && config.AvatarSource != "personal" {
					// Check if we have that org avatar URL
					if orgURL, exists := m.avatarURLs[config.AvatarSource]; exists {
						avatarURL = orgURL
					}
				}
			}

			if m.width < 100 {
				// Small screen mode: always use 30 width
				return m, fetchAvatarSmall(avatarURL)
			}
			// Normal mode: use configured width
			config, _ := LoadConfig()
			avatarWidth := config.AvatarWidth
			if avatarWidth < 30 || avatarWidth > 240 {
				avatarWidth = 60 // Fallback to default
			}
			return m, fetchAvatarWithWidth(avatarURL, avatarWidth)
		}

	case contributionsMsg:
		m.graph.SetContributions(msg)
		m.streaks.SetContributions(msg)
		m.loading.contributions = false

	case languagesMsg:
		m.languages.SetLanguages(msg.languages, msg.repoCount)
		m.loading.languages = false

	case repositoriesMsg:
		m.repos.SetRepos(msg)
		m.loading.repositories = false

	case activitiesMsg:
		m.activityMetrics.SetActivities(msg)
		m.activityModal.SetActivities(msg)
		m.loading.activities = false

	case avatarMsg:
		m.profile.SetAvatar(string(msg))
		m.loading.avatar = false

	case orgAvatarsMsg:
		// Store org avatar URLs in the map
		for orgName, avatarURL := range msg {
			m.avatarURLs[orgName] = avatarURL
		}

		// ONLY reload avatar with org source if viewing own profile
		if m.isOwnProfile {
			config, _ := LoadConfig()
			if config.AvatarSource != "" && config.AvatarSource != "personal" {
				if orgURL, exists := m.avatarURLs[config.AvatarSource]; exists {
					// Get avatar width
					avatarWidth := config.AvatarWidth
					if avatarWidth < 30 || avatarWidth > 240 {
						avatarWidth = 60
					}

					// Reload avatar with the org URL
					if m.width < 100 {
						return m, fetchAvatarSmall(orgURL)
					}
					return m, fetchAvatarWithWidth(orgURL, avatarWidth)
				}
			}
		}

	case errMsg:
		m.err = msg
		// Clear all loading flags on error
		m.loading = loadingState{}

	case clearNotificationMsg:
		m.notification = ""

	default:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the complete UI using composable components
func (m Model) View() string {
	if m.err != nil {
		return errorStyle.Render("Error: " + m.err.Error())
	}

	if m.loading.isLoading() {
		return m.renderLoading()
	}

	// If help modal is visible, show it (highest priority)
	if m.helpModal.IsVisible() {
		return m.helpModal.View()
	}

	// If activity modal is visible, show it centered
	if m.activityModal.IsVisible() {
		return m.activityModal.View()
	}

	// Check for small screen mode (width < 100)
	if m.width < 100 {
		return m.renderSmallScreen()
	}

	// ASCII Username Header
	var usernameHeader string
	if m.profile.profile != nil {
		asciiUsername := RenderASCII("@" + m.profile.profile.Login)
		usernameHeader = lipgloss.NewStyle().
			Width(m.width).
			Background(lipgloss.Color(CurrentTheme.Subtle)).
			Align(lipgloss.Right).
			PaddingTop(1).
			PaddingRight(5).
			Render(accentStyle.Render(asciiUsername))
	}

	// Build layout using lipgloss composition - NO manual width calculations!
	sections := []string{}
	if usernameHeader != "" {
		sections = append(sections, usernameHeader)
	}
	sections = append(sections, m.graph.View(), "") // spacer

	// Stats row: Languages | Streaks | Activity Metrics | Repos
	statsRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.languages.View(),
		"  ", // spacer
		m.streaks.View(),
		"  ", // spacer
		m.activityMetrics.View(),
		"  ", // spacer
		m.repos.View(),
	)
	sections = append(sections, statsRow)
	sections = append(sections, "") // spacer

	// Profile section
	sections = append(sections, m.profile.View())

	// Main content
	mainContent := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Status bar at bottom
	statusBar := m.renderStatusBar()

	// Calculate available height for content (total - status bar height)
	statusBarHeight := lipgloss.Height(statusBar)
	contentHeight := m.height - statusBarHeight

	// Render main content in available space
	contentArea := lipgloss.NewStyle().
		Width(m.width).
		Height(contentHeight).
		Render(mainContent)

	// Join content and status bar
	fullView := lipgloss.JoinVertical(
		lipgloss.Left,
		contentArea,
		statusBar,
	)

	return fullView
}

// renderSmallScreen renders a compact view for narrow terminals
func (m Model) renderSmallScreen() string {
	sections := []string{}

	// Title
	if m.profile.profile != nil {
		title := titleStyle.Render("@" + m.profile.profile.Login)
		sections = append(sections, title, "")
	}

	// Profile section (avatar + basic info) - compact mode
	sections = append(sections, m.profile.ViewCompact(), "")

	// Streaks
	sections = append(sections, m.streaks.View(), "")

	// Activity Metrics
	sections = append(sections, m.activityMetrics.View(), "")

	// Main content - no status bar in small screen mode
	mainContent := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Render full height (no status bar)
	contentArea := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render(mainContent)

	return contentArea
}

// renderLoading shows loading state with spinner
func (m Model) renderLoading() string {
	return lipgloss.NewStyle().
		Padding(2).
		Render(m.spinner.View() + " Loading profile data...")
}

// renderStatusBar renders the bottom status bar
func (m Model) renderStatusBar() string {
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.Blue)).
		Background(lipgloss.Color(CurrentTheme.Subtle)).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.Foreground)).
		Background(lipgloss.Color(CurrentTheme.Subtle))

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.Green)).
		Background(lipgloss.Color(CurrentTheme.Subtle)).
		Bold(true)

	sepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.Gray)).
		Background(lipgloss.Color(CurrentTheme.Subtle))

	var parts []string

	// Show notification OR minimal hints
	if m.notification != "" {
		// Show notification
		parts = append(parts, lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{
				Light: CurrentTheme.Green,
				Dark:  CurrentTheme.BrightGreen,
			}).
			Background(lipgloss.Color(CurrentTheme.Subtle)).
			Bold(true).
			Render(m.notification))
	} else {
		// Minimal hints
		parts = append(parts, keyStyle.Render("?")+descStyle.Render(": help"))
		parts = append(parts, keyStyle.Render("q")+descStyle.Render(": quit"))
		parts = append(parts, keyStyle.Render("r")+descStyle.Render(": refresh"))
		parts = append(parts, keyStyle.Render("t")+descStyle.Render(": theme"))

		// Current theme name
		themeName := GetCurrentThemeName()
		parts = append(parts, valueStyle.Render("["+themeName+"]"))
	}

	separator := sepStyle.Render(" | ")
	help := strings.Join(parts, separator)

	return lipgloss.NewStyle().
		Width(m.width).
		Background(lipgloss.Color(CurrentTheme.Subtle)).
		Render(help)
}

func main() {
	// Initialize theme system first (loads all 361 themes from gogh-themes)
	InitTheme()

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load config: %v\n", err)
		config = DefaultConfig()
	}

	// Set theme from config
	if err := SetThemeByName(config.Theme); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to set theme '%s': %v\n", config.Theme, err)
		SetThemeByName("Dracula") // Fallback to default
	}
	InitStyles()

	// Get username from: 1) CLI arg, 2) env var, 3) config file
	username := ""
	if len(os.Args) > 1 {
		username = os.Args[1]
	} else if envUser := os.Getenv("GITHUB_USERNAME"); envUser != "" {
		username = envUser
	} else if config.Username != "" {
		username = config.Username
	}

	if username == "" {
		println("Usage: gittui <github-username>")
		println("Or set GITHUB_USERNAME environment variable")
		println("Or run 'make config' to save username to config")
		println()
		println("Configuration:")
		println("  Config file: ~/.local/gittui/config.json")
		println("  Available themes:", GetThemeCount())
		println("  Current theme:", config.Theme)
		println()
		println("To change theme, edit ~/.local/gittui/config.json")
		println("Example themes: Dracula, Nord, Gruvbox-Dark, Catppuccin-Mocha, Tokyo-Night")
		println("List all themes: https://github.com/willyv3/gogh-themes")
		os.Exit(1)
	}

	m := NewModel(username)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		println("Error:", err.Error())
		os.Exit(1)
	}
}
