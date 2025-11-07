package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/cli/go-gh/v2/pkg/auth"
	ghAPI "github.com/cli/go-gh/v2/pkg/api"
)

const (
	githubAPIURL     = "https://api.github.com"
	githubGraphQLURL = "https://api.github.com/graphql"
)

// GitHubClient handles all GitHub API interactions
type GitHubClient struct {
	httpClient *http.Client
}

// NewGitHubClient creates a new GitHub API client using official go-gh library
// This is the IDIOMATIC and SECURE way - it handles:
// - Environment variables (GITHUB_TOKEN, GH_TOKEN)
// - gh CLI authentication
// - Proper token storage and security
// - No manual exec.Command calls
// Falls back to unauthenticated client if no auth is available (public data only)
func NewGitHubClient() (*GitHubClient, error) {
	// Try to get authenticated HTTP client first
	opts := &ghAPI.ClientOptions{
		Host:    "github.com",
		Timeout: 10 * time.Second,
	}

	httpClient, err := ghAPI.NewHTTPClient(*opts)
	if err != nil {
		// Fall back to unauthenticated client (public data only)
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	// Check if we have authentication
	token, _ := auth.TokenForHost("github.com")
	if token == "" {
		// No authentication - use basic HTTP client
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	return &GitHubClient{
		httpClient: httpClient,
	}, nil
}

// Repository represents a GitHub repository
type Repository struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Stars       int    `json:"stargazers_count"`
	Forks       int    `json:"forks_count"`
	Private     bool   `json:"private"`
	Owner       struct {
		Login string `json:"login"`
		Type  string `json:"type"` // "User" or "Organization"
	} `json:"owner"`
}

// ProfileData contains user profile information
type ProfileData struct {
	Login       string
	Name        string
	Bio         string
	Location    string
	Company     string
	AvatarURL   string
	PublicRepos int
	PublicGists int
	Followers   int
	Following   int
	CreatedAt   time.Time
}

// LanguageStats represents language usage statistics
type LanguageStats struct {
	Name       string
	Percentage float64
	Color      string
}

// Activity represents a recent activity item
type Activity struct {
	Type      string
	Repo      string
	Action    string
	Timestamp time.Time
	Public    bool
}


// FetchOrgAvatars fetches avatar URLs for given organizations
func (c *GitHubClient) FetchOrgAvatars(orgNames []string) (map[string]string, error) {
	avatars := make(map[string]string)

	for _, orgName := range orgNames {
		url := fmt.Sprintf("%s/orgs/%s", githubAPIURL, orgName)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue // Skip this org on error
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			continue // Skip this org on error
		}

		if resp.StatusCode == 200 {
			var org struct {
				AvatarURL string `json:"avatar_url"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&org); err == nil {
				avatars[orgName] = org.AvatarURL
			}
		}
		resp.Body.Close()
	}

	return avatars, nil
}

// GetAuthenticatedUser returns the login of the authenticated user
// Returns empty string if not authenticated
func (c *GitHubClient) GetAuthenticatedUser() (string, error) {
	req, err := http.NewRequest("GET", githubAPIURL+"/user", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		// Not authenticated
		return "", nil
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to get authenticated user: status %d", resp.StatusCode)
	}

	var user struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", err
	}

	return user.Login, nil
}

// FetchProfile fetches user profile data
// includePrivate: if true, uses /user endpoint to get private counts (only works for authenticated user)
func (c *GitHubClient) FetchProfile(username string, includePrivate bool) (*ProfileData, error) {
	url := fmt.Sprintf("%s/users/%s", githubAPIURL, username)
	if includePrivate {
		url = fmt.Sprintf("%s/user", githubAPIURL)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// Authorization header automatically added by go-gh HTTPClient
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var data struct {
		Login             string    `json:"login"`
		Name              string    `json:"name"`
		Bio               string    `json:"bio"`
		Location          string    `json:"location"`
		Company           string    `json:"company"`
		AvatarURL         string    `json:"avatar_url"`
		PublicRepos       int       `json:"public_repos"`
		TotalPrivateRepos int       `json:"total_private_repos"`
		PublicGists       int       `json:"public_gists"`
		Followers         int       `json:"followers"`
		Following         int       `json:"following"`
		CreatedAt         time.Time `json:"created_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &ProfileData{
		Login:       data.Login,
		Name:        data.Name,
		Bio:         data.Bio,
		Location:    data.Location,
		Company:     data.Company,
		AvatarURL:   data.AvatarURL,
		PublicRepos: data.PublicRepos,
		PublicGists: data.PublicGists,
		Followers:   data.Followers,
		Following:   data.Following,
		CreatedAt:   data.CreatedAt,
	}, nil
}

// FetchAuthenticatedUser fetches the authenticated user's profile
func (c *GitHubClient) FetchAuthenticatedUser() (*ProfileData, error) {
	url := fmt.Sprintf("%s/user", githubAPIURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// Authorization header automatically added by go-gh HTTPClient
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var data struct {
		Login       string    `json:"login"`
		Name        string    `json:"name"`
		Bio         string    `json:"bio"`
		Location    string    `json:"location"`
		Company     string    `json:"company"`
		AvatarURL   string    `json:"avatar_url"`
		PublicRepos int       `json:"public_repos"`
		PublicGists int       `json:"public_gists"`
		Followers   int       `json:"followers"`
		Following   int       `json:"following"`
		CreatedAt   time.Time `json:"created_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	profile := &ProfileData{
		Login:       data.Login,
		Name:        data.Name,
		Bio:         data.Bio,
		Location:    data.Location,
		Company:     data.Company,
		AvatarURL:   data.AvatarURL,
		PublicRepos: data.PublicRepos,
		PublicGists: data.PublicGists,
		Followers:   data.Followers,
		Following:   data.Following,
		CreatedAt:   data.CreatedAt,
	}

	return profile, nil
}

// FetchContributions fetches contribution calendar data using GraphQL
func (c *GitHubClient) FetchContributions(username string) ([]Contribution, error) {
	query := `
	query($username: String!) {
		user(login: $username) {
			contributionsCollection {
				contributionCalendar {
					weeks {
						contributionDays {
							contributionCount
							date
						}
					}
				}
			}
		}
	}`

	variables := map[string]interface{}{
		"username": username,
	}

	reqBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", githubGraphQLURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	// Authorization header automatically added by go-gh HTTPClient
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			User struct {
				ContributionsCollection struct {
					ContributionCalendar struct {
						Weeks []struct {
							ContributionDays []struct {
								ContributionCount int    `json:"contributionCount"`
								Date              string `json:"date"`
							} `json:"contributionDays"`
						} `json:"weeks"`
					} `json:"contributionCalendar"`
				} `json:"contributionsCollection"`
			} `json:"user"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var contributions []Contribution
	for _, week := range result.Data.User.ContributionsCollection.ContributionCalendar.Weeks {
		for _, day := range week.ContributionDays {
			date, err := time.Parse("2006-01-02", day.Date)
			if err != nil {
				continue
			}
			contributions = append(contributions, Contribution{
				Date:  date,
				Count: day.ContributionCount,
			})
		}
	}

	return contributions, nil
}

// FetchLanguages fetches top language statistics and total repo count
// includePrivate: if true, uses /user/repos with affiliation to get all repos including org repos
func (c *GitHubClient) FetchLanguages(username string, includePrivate bool) ([]LanguageStats, int, error) {
	// Load config to get org filtering preferences
	config, err := LoadConfig()
	if err != nil {
		// Fall back to defaults if config can't be loaded
		config = DefaultConfig()
	}

	var baseURL string

	if includePrivate {
		baseURL = fmt.Sprintf("%s/user/repos?per_page=100&affiliation=owner,collaborator,organization_member", githubAPIURL)
	} else {
		baseURL = fmt.Sprintf("%s/users/%s/repos?per_page=100", githubAPIURL, username)
	}

	// Fetch all repos with pagination - need owner info for filtering
	var allRepos []struct {
		Name     string `json:"name"`
		Language string `json:"language"`
		Fork     bool   `json:"fork"`
		Private  bool   `json:"private"`
		Owner    struct {
			Login string `json:"login"`
			Type  string `json:"type"` // "User" or "Organization"
		} `json:"owner"`
	}

	page := 1
	for {
		url := fmt.Sprintf("%s&page=%d", baseURL, page)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, 0, err
		}
		// Authorization header automatically added by go-gh HTTPClient
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, 0, err
		}

		var repos []struct {
			Name     string `json:"name"`
			Language string `json:"language"`
			Fork     bool   `json:"fork"`
			Private  bool   `json:"private"`
			Owner    struct {
				Login string `json:"login"`
				Type  string `json:"type"`
			} `json:"owner"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
			resp.Body.Close()
			return nil, 0, err
		}
		resp.Body.Close()

		if len(repos) == 0 {
			break
		}

		allRepos = append(allRepos, repos...)

		// If we got less than 100, we're done
		if len(repos) < 100 {
			break
		}

		page++
	}

	// Filter repos based on config preferences
	var filteredRepos []struct {
		Name     string `json:"name"`
		Language string `json:"language"`
		Fork     bool   `json:"fork"`
		Private  bool   `json:"private"`
		Owner    struct {
			Login string `json:"login"`
			Type  string `json:"type"`
		} `json:"owner"`
	}

	// Build allowed orgs map for O(1) lookup
	allowedOrgsMap := make(map[string]bool)
	for _, org := range config.AllowedOrgs {
		allowedOrgsMap[org] = true
	}

	for _, repo := range allRepos {
		// Public repos are always included
		if !repo.Private {
			filteredRepos = append(filteredRepos, repo)
			continue
		}

		// Private repos - apply filtering
		isPersonal := repo.Owner.Type == "User" && repo.Owner.Login == username
		isAllowedOrg := repo.Owner.Type == "Organization" && allowedOrgsMap[repo.Owner.Login]

		if (isPersonal && config.IncludePersonalPrivate) || isAllowedOrg {
			filteredRepos = append(filteredRepos, repo)
		}
	}

	// Total repo count (after filtering)
	totalRepoCount := len(filteredRepos)

	// Count languages (excluding forks)
	langCount := make(map[string]int)
	total := 0
	for _, repo := range filteredRepos {
		if !repo.Fork && repo.Language != "" {
			langCount[repo.Language]++
			total++
		}
	}

	// Convert to sorted slice
	var languages []LanguageStats
	for lang, count := range langCount {
		languages = append(languages, LanguageStats{
			Name:       lang,
			Percentage: float64(count) / float64(total),
			Color:      getLanguageColor(lang),
		})
	}

	// Sort by percentage (descending)
	for i := 0; i < len(languages); i++ {
		for j := i + 1; j < len(languages); j++ {
			if languages[j].Percentage > languages[i].Percentage {
				languages[i], languages[j] = languages[j], languages[i]
			}
		}
	}

	// Return top 5
	if len(languages) > 5 {
		languages = languages[:5]
	}

	return languages, totalRepoCount, nil
}

// FetchTopRepositories fetches user's top repositories sorted by stars
func (c *GitHubClient) FetchTopRepositories(username string, includePrivate bool) ([]Repository, error) {
	// Load config to get org filtering preferences
	config, err := LoadConfig()
	if err != nil {
		// Fall back to defaults if config can't be loaded
		config = DefaultConfig()
	}

	var baseURL string

	if includePrivate {
		baseURL = fmt.Sprintf("%s/user/repos?per_page=100&affiliation=owner,collaborator,organization_member", githubAPIURL)
	} else {
		baseURL = fmt.Sprintf("%s/users/%s/repos?per_page=100", githubAPIURL, username)
	}

	// Fetch all repos with pagination
	var allRepos []Repository
	page := 1

	for {
		url := fmt.Sprintf("%s&page=%d", baseURL, page)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		// Authorization header automatically added by go-gh HTTPClient
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		var repos []Repository
		if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		allRepos = append(allRepos, repos...)

		// If we got less than 100 repos, we've reached the last page
		if len(repos) < 100 {
			break
		}
		page++
	}

	// Filter repos based on config preferences
	var filteredRepos []Repository

	// Build allowed orgs map for O(1) lookup
	allowedOrgsMap := make(map[string]bool)
	for _, org := range config.AllowedOrgs {
		allowedOrgsMap[org] = true
	}

	for _, repo := range allRepos {
		// Public repos are always included
		if !repo.Private {
			filteredRepos = append(filteredRepos, repo)
			continue
		}

		// Private repos - apply filtering
		isPersonal := repo.Owner.Type == "User" && repo.Owner.Login == username
		isAllowedOrg := repo.Owner.Type == "Organization" && allowedOrgsMap[repo.Owner.Login]

		if (isPersonal && config.IncludePersonalPrivate) || isAllowedOrg {
			filteredRepos = append(filteredRepos, repo)
		}
	}

	// Sort by stars (stargazers_count) - GitHub API doesn't support server-side sorting by stars
	sort.Slice(filteredRepos, func(i, j int) bool {
		return filteredRepos[i].Stars > filteredRepos[j].Stars
	})

	// Return top 5 repos
	if len(filteredRepos) > 5 {
		filteredRepos = filteredRepos[:5]
	}

	return filteredRepos, nil
}

// FetchRecentActivity fetches recent user activity with pagination
// includePrivate: if true, uses /users/{username}/events to get private events (only works for authenticated user)
func (c *GitHubClient) FetchRecentActivity(username string, includePrivate bool) ([]Activity, error) {
	var baseURL string

	if includePrivate {
		baseURL = fmt.Sprintf("%s/users/%s/events?per_page=100", githubAPIURL, username)
	} else {
		baseURL = fmt.Sprintf("%s/users/%s/events/public?per_page=100", githubAPIURL, username)
	}

	var allEvents []struct {
		Type      string `json:"type"`
		CreatedAt string `json:"created_at"`
		Public    bool   `json:"public"`
		Repo      struct {
			Name string `json:"name"`
		} `json:"repo"`
		Payload map[string]interface{} `json:"payload"`
	}

	// Fetch up to 3 pages (300 events max)
	for page := 1; page <= 3; page++ {
		url := fmt.Sprintf("%s&page=%d", baseURL, page)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		var events []struct {
			Type      string `json:"type"`
			CreatedAt string `json:"created_at"`
			Public    bool   `json:"public"`
			Repo      struct {
				Name string `json:"name"`
			} `json:"repo"`
			Payload map[string]interface{} `json:"payload"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		if len(events) == 0 {
			break
		}

		allEvents = append(allEvents, events...)

		// If we got less than 100, we're done
		if len(events) < 100 {
			break
		}
	}

	var activities []Activity
	for _, event := range allEvents {
		timestamp, _ := time.Parse(time.RFC3339, event.CreatedAt)
		action := getActionDescription(event.Type, event.Payload)

		activities = append(activities, Activity{
			Type:      event.Type,
			Repo:      event.Repo.Name,
			Action:    action,
			Timestamp: timestamp,
			Public:    event.Public,
		})
	}

	return activities, nil
}

// getActionDescription creates human-readable action description
func getActionDescription(eventType string, payload map[string]interface{}) string {
	switch eventType {
	case "PushEvent":
		if commits, ok := payload["commits"].([]interface{}); ok {
			return fmt.Sprintf("Pushed %d commit(s)", len(commits))
		}
		return "Pushed commits"
	case "CreateEvent":
		if refType, ok := payload["ref_type"].(string); ok {
			return fmt.Sprintf("Created %s", refType)
		}
		return "Created repository"
	case "PullRequestEvent":
		if action, ok := payload["action"].(string); ok {
			return fmt.Sprintf("Pull request %s", action)
		}
		return "Pull request activity"
	case "IssuesEvent":
		if action, ok := payload["action"].(string); ok {
			return fmt.Sprintf("Issue %s", action)
		}
		return "Issue activity"
	case "WatchEvent":
		return "Starred repository"
	case "ForkEvent":
		return "Forked repository"
	default:
		return eventType
	}
}

// getLanguageColor returns GitHub's language colors
func getLanguageColor(lang string) string {
	colors := map[string]string{
		"Go":         "#00ADD8",
		"JavaScript": "#f1e05a",
		"TypeScript": "#3178c6",
		"Python":     "#3572A5",
		"Rust":       "#dea584",
		"Java":       "#b07219",
		"C":          "#555555",
		"C++":        "#f34b7d",
		"Ruby":       "#701516",
		"PHP":        "#4F5D95",
		"Swift":      "#ffac45",
		"Kotlin":     "#A97BFF",
		"Shell":      "#89e051",
		"HTML":       "#e34c26",
		"CSS":        "#563d7c",
	}
	if color, ok := colors[lang]; ok {
		return color
	}
	return "#858585" // Default gray
}
