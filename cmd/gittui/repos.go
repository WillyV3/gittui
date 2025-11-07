package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ReposComponent displays top repositories
type ReposComponent struct {
	repos   []Repository
	loading bool
}

// NewReposComponent creates a new repos component
func NewReposComponent() ReposComponent {
	return ReposComponent{
		loading: true,
	}
}

// SetRepos updates the repository data
func (r *ReposComponent) SetRepos(repos []Repository) {
	r.repos = repos
	r.loading = false
}

// Update handles messages
func (r ReposComponent) Update(msg tea.Msg) (ReposComponent, tea.Cmd) {
	return r, nil
}

// View renders the repos component
func (r ReposComponent) View() string {
	title := titleStyle.Render("Top Repositories")

	if r.loading {
		return lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			labelStyle.Render("Loading..."))
	}

	if len(r.repos) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			labelStyle.Render("No repositories"))
	}

	var lines []string
	lines = append(lines, title, "")

	// Limit to top 3 repos
	displayRepos := r.repos
	if len(displayRepos) > 3 {
		displayRepos = displayRepos[:3]
	}

	for _, repo := range displayRepos {
		// Repo name with org if applicable
		repoName := repo.Name
		if repo.Owner.Type == "Organization" {
			repoName = fmt.Sprintf("%s (%s)", repo.Name, repo.Owner.Login)
		}
		lines = append(lines, labelStyle.Render(repoName))

		// Stars and language
		stats := fmt.Sprintf("⭐ %d", repo.Stars)
		if repo.Language != "" {
			stats += fmt.Sprintf(" • %s", repo.Language)
		}
		lines = append(lines, accentStyle.Render(stats))
		lines = append(lines, "")
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
