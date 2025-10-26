// +build ignore

package main

import (
	"fmt"
	"time"
)

func main() {
	// Create test activities
	now := time.Now()

	activities := []Activity{
		{Type: "PushEvent", Timestamp: time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)}, // 2:30 PM
		{Type: "PushEvent", Timestamp: time.Date(2025, 1, 15, 14, 45, 0, 0, time.UTC)}, // 2:45 PM
		{Type: "PushEvent", Timestamp: time.Date(2025, 1, 15, 14, 50, 0, 0, time.UTC)}, // 2:50 PM
		{Type: "IssueEvent", Timestamp: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)},  // 10 AM
		{Type: "PushEvent", Timestamp: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)}, // 10:30 AM
		{Type: "PushEvent", Timestamp: time.Date(2025, 1, 14, 9, 0, 0, 0, time.UTC)},   // 9 AM (yesterday)
	}

	fmt.Println("Current time:", now)
	fmt.Println("\nTest Activities:")
	for i, a := range activities {
		fmt.Printf("%d. %s at %s (hour: %d)\n", i+1, a.Type, a.Timestamp.Format("Mon 3:04 PM"), a.Timestamp.Hour())
	}

	// Test with different windows
	windows := []TimeWindow{ThisWeek, ThisMonth, ThisYear}

	for _, window := range windows {
		fmt.Printf("\n--- %s ---\n", window.String())
		peakHour, distribution := CalculatePeakCodingHour(activities, window)

		fmt.Printf("Peak Hour: %s\n", peakHour)
		fmt.Printf("Distribution: %v\n", distribution)

		total := 0
		for _, count := range distribution {
			total += count
		}
		fmt.Printf("Total activities counted: %d\n", total)
	}
}
