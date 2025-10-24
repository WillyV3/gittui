package main

import (
	"math/rand"
	"time"
)

// GenerateTestData creates a year of contribution data for testing.
// Pattern determines the distribution: random, active, streaky, gradient, sparse, consistent.
func GenerateTestData(pattern string) []Contribution {
	now := time.Now()
	oneYearAgo := now.AddDate(-1, 0, 0)

	var days []Contribution

	for d := oneYearAgo; d.Before(now); d = d.AddDate(0, 0, 1) {
		count := 0

		switch pattern {
		case "random":
			count = rand.Intn(15) // 0-14 contributions
		case "active":
			count = rand.Intn(20) + 5 // 5-24 contributions
		case "streaky":
			// High activity on weekdays, low on weekends
			if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
				count = rand.Intn(3)
			} else {
				count = rand.Intn(15) + 5
			}
		case "gradient":
			// Increasing activity over the year
			daysSinceStart := int(d.Sub(oneYearAgo).Hours() / 24)
			count = (daysSinceStart * 20) / 365
		case "sparse":
			// Mostly empty with occasional bursts
			if rand.Float32() < 0.3 {
				count = rand.Intn(10)
			}
		case "consistent":
			// Steady contributions every day
			count = 5 + rand.Intn(5) // 5-9 per day
		default:
			count = rand.Intn(10)
		}

		days = append(days, Contribution{
			Date:  d,
			Count: count,
		})
	}

	return days
}

// GenerateCustomData for manual testing
func GenerateCustomData() []Contribution {
	now := time.Now()
	oneYearAgo := now.AddDate(-1, 0, 0)

	var days []Contribution

	// Create interesting patterns
	for d := oneYearAgo; d.Before(now); d = d.AddDate(0, 0, 1) {
		count := 0

		week := int(d.Sub(oneYearAgo).Hours() / 24 / 7)

		// Create waves pattern
		if week%4 == 0 {
			count = 15 // high
		} else if week%4 == 1 {
			count = 10 // medium-high
		} else if week%4 == 2 {
			count = 5 // medium
		} else {
			count = 2 // low
		}

		days = append(days, Contribution{
			Date:  d,
			Count: count,
		})
	}

	return days
}
