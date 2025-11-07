package main

import (
	"fmt"
	"time"
)

// PushGranularity represents how push rate statistics are calculated
type PushGranularity string

const (
	PushPerHour  PushGranularity = "hour"
	PushPerDay   PushGranularity = "day"
	PushPerWeek  PushGranularity = "week"
	PushPerMonth PushGranularity = "month"
)

// loadingState tracks which data is currently loading
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

// Message types for async data fetching
type profileMsg *ProfileData
type contributionsMsg []Contribution
type languagesMsg struct {
	languages []LanguageStats
	repoCount int
}
type repositoriesMsg []Repository
type activitiesMsg []Activity
type avatarMsg string
type orgAvatarsMsg map[string]string // Map of org name -> avatar URL
type errMsg error

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

// formatTimeAgo formats a time as a human-readable "ago" string
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
