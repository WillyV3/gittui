package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProfileComponent handles the profile display (avatar + info)
type ProfileComponent struct {
	profile    *ProfileData
	avatarASCII string
	loading    bool
}

// NewProfileComponent creates a new profile component
func NewProfileComponent() ProfileComponent {
	return ProfileComponent{
		loading: true,
	}
}

// SetProfile updates the profile data
func (p *ProfileComponent) SetProfile(profile *ProfileData) {
	p.profile = profile
	p.loading = false
}

// SetAvatar updates the avatar ASCII art
func (p *ProfileComponent) SetAvatar(avatar string) {
	p.avatarASCII = avatar
}

// Update handles messages for the profile component
func (p ProfileComponent) Update(msg tea.Msg) (ProfileComponent, tea.Cmd) {
	// Profile component is mostly passive, updates come from parent
	return p, nil
}

// View renders the profile component
func (p ProfileComponent) View() string {
	if p.loading || p.profile == nil {
		return labelStyle.Render("Loading profile...")
	}

	// Avatar on left, info on right
	avatar := p.avatarASCII
	if avatar == "" {
		avatar = labelStyle.Render("Loading avatar...")
	}

	// Profile info
	var infoLines []string

	// Name and login
	nameStr := p.profile.Name
	if nameStr == "" {
		nameStr = p.profile.Login
	}
	infoLines = append(infoLines,
		lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{
				Light: CurrentTheme.Blue,
				Dark:  CurrentTheme.BrightBlue,
			}).
			Render(nameStr))

	if p.profile.Name != "" {
		infoLines = append(infoLines,
			lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{
					Light: CurrentTheme.Gray,
					Dark:  CurrentTheme.BrightBlack,
				}).
				Render("@"+p.profile.Login))
	}

	infoLines = append(infoLines, "")

	// Bio
	if p.profile.Bio != "" {
		infoLines = append(infoLines, baseStyle.Render(p.profile.Bio))
		infoLines = append(infoLines, "")
	}

	// Stats - use global styles for consistency with other components
	infoLines = append(infoLines,
		labelStyle.Render("Followers: ")+
			accentStyle.Render(fmt.Sprintf("%d", p.profile.Followers))+
			labelStyle.Render("  Following: ")+
			accentStyle.Render(fmt.Sprintf("%d", p.profile.Following)))

	infoLines = append(infoLines,
		labelStyle.Render("Public Repos: ")+
			accentStyle.Render(fmt.Sprintf("%d", p.profile.PublicRepos)))

	infoLines = append(infoLines,
		labelStyle.Render("Public Gists: ")+
			accentStyle.Render(fmt.Sprintf("%d", p.profile.PublicGists)))

	// Company, Location
	if p.profile.Company != "" {
		infoLines = append(infoLines, "")
		infoLines = append(infoLines,
			baseStyle.Render("🏢 "+p.profile.Company))
	}

	if p.profile.Location != "" {
		infoLines = append(infoLines,
			baseStyle.Render("📍 "+p.profile.Location))
	}

	profileInfo := lipgloss.JoinVertical(lipgloss.Left, infoLines...)

	// Box the profile info - use titleStyle colors for consistency
	profileBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(titleStyle.GetForeground()).
		Padding(1, 1).
		Render(profileInfo)

	// Join avatar and profile box horizontally with spacing
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		avatar,
		lipgloss.NewStyle().Width(2).Render(""),
		profileBox,
	)
}

// ViewCompact renders a compact profile view with small avatar for small screens
func (p ProfileComponent) ViewCompact() string {
	if p.loading || p.profile == nil {
		return labelStyle.Render("Loading profile...")
	}

	// Show small avatar if available
	avatar := p.avatarASCII
	if avatar == "" {
		avatar = labelStyle.Render("Loading...")
	}

	// Profile info
	var infoLines []string

	// Name and login
	nameStr := p.profile.Name
	if nameStr == "" {
		nameStr = p.profile.Login
	}
	infoLines = append(infoLines,
		lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{
				Light: CurrentTheme.Blue,
				Dark:  CurrentTheme.BrightBlue,
			}).
			Render(nameStr))

	if p.profile.Name != "" {
		infoLines = append(infoLines,
			lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{
					Light: CurrentTheme.Gray,
					Dark:  CurrentTheme.BrightBlack,
				}).
				Render("@"+p.profile.Login))
	}

	infoLines = append(infoLines, "")

	// Bio (if present)
	if p.profile.Bio != "" {
		infoLines = append(infoLines, baseStyle.Render(p.profile.Bio))
		infoLines = append(infoLines, "")
	}

	// Compact stats - single line
	infoLines = append(infoLines,
		labelStyle.Render("Followers: ")+
			accentStyle.Render(fmt.Sprintf("%d", p.profile.Followers))+
			labelStyle.Render("  Following: ")+
			accentStyle.Render(fmt.Sprintf("%d", p.profile.Following)))

	infoLines = append(infoLines,
		labelStyle.Render("Repos: ")+
			accentStyle.Render(fmt.Sprintf("%d", p.profile.PublicRepos))+
			labelStyle.Render("  Gists: ")+
			accentStyle.Render(fmt.Sprintf("%d", p.profile.PublicGists)))

	// Location and company on one line if both present
	if p.profile.Location != "" || p.profile.Company != "" {
		infoLines = append(infoLines, "")
		var extraInfo []string
		if p.profile.Location != "" {
			extraInfo = append(extraInfo, "📍 "+p.profile.Location)
		}
		if p.profile.Company != "" {
			extraInfo = append(extraInfo, "🏢 "+p.profile.Company)
		}
		infoLines = append(infoLines, baseStyle.Render(strings.Join(extraInfo, "  ")))
	}

	profileInfo := lipgloss.JoinVertical(lipgloss.Left, infoLines...)

	// Box the profile info
	profileBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(titleStyle.GetForeground()).
		Padding(1, 1).
		Render(profileInfo)

	// Stack avatar on top of profile box vertically
	return lipgloss.JoinVertical(
		lipgloss.Left,
		avatar,
		"",
		profileBox,
	)
}
