package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Async command functions for fetching data

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
	return fetchAvatarWithWidth(avatarURL, 60)
}

func fetchAvatarSmall(avatarURL string) tea.Cmd {
	return fetchAvatarWithWidth(avatarURL, 30)
}

func fetchAvatarWithWidth(avatarURL string, width int) tea.Cmd {
	return func() tea.Msg {
		// Fetch avatar ASCII art at specified width
		ascii, err := FetchAvatarASCII(avatarURL, width)
		if err != nil {
			// Don't fail the whole app if avatar fails
			return avatarMsg("")
		}
		return avatarMsg(ascii)
	}
}

func fetchOrgAvatars(client *GitHubClient, orgNames []string) tea.Cmd {
	return func() tea.Msg {
		avatars, err := client.FetchOrgAvatars(orgNames)
		if err != nil {
			// Return empty map on error
			return orgAvatarsMsg(make(map[string]string))
		}
		return orgAvatarsMsg(avatars)
	}
}
