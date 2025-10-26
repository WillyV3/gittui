package main

import (
	"fmt"
	"strings"
	"time"
)

// TimeWindow represents different time ranges for statistics
type TimeWindow int

const (
	ThisWeek TimeWindow = iota
	ThisMonth
	ThisYear
)

// String returns the display name for a time window
func (tw TimeWindow) String() string {
	switch tw {
	case ThisWeek:
		return "This Week"
	case ThisMonth:
		return "This Month"
	case ThisYear:
		return "This Year"
	default:
		return "Unknown"
	}
}

// ActivityStats holds calculated statistics from activity data
type ActivityStats struct {
	PushRate      float64
	PRRate        float64
	ReviewRate    float64
	PeakCodingHour string
	HourDistribution map[int]int // hour -> activity count
}

// CalculatePushRate calculates push frequency based on 7-day rolling average
// Daily rate: pushes in last 7 days / 7
// Weekly rate: daily rate × 7
// Monthly rate: daily rate × 30
func CalculatePushRate(activities []Activity, granularity PushGranularity) float64 {
	now := time.Now()
	sevenDaysAgo := now.AddDate(0, 0, -7)

	// Count push events in the last 7 days
	pushCount := 0
	for _, activity := range activities {
		if activity.Type == "PushEvent" && activity.Timestamp.After(sevenDaysAgo) {
			pushCount++
		}
	}

	if pushCount == 0 {
		return 0.0
	}

	// Calculate daily average from last 7 days
	dailyAverage := float64(pushCount) / 7.0

	// Project based on granularity
	var rate float64
	switch granularity {
	case PushPerHour:
		// Daily average / 24 hours
		rate = dailyAverage / 24.0
	case PushPerDay:
		// Already calculated
		rate = dailyAverage
	case PushPerWeek:
		// Daily average × 7 days
		rate = dailyAverage * 7.0
	case PushPerMonth:
		// Daily average × 30 days
		rate = dailyAverage * 30.0
	default:
		rate = dailyAverage
	}

	return rate
}

// CalculatePeakCodingHour analyzes activity timestamps to find the most active hour
func CalculatePeakCodingHour(activities []Activity, window TimeWindow) (string, map[int]int) {
	now := time.Now()
	var cutoff time.Time

	// Determine time window cutoff
	switch window {
	case ThisWeek:
		// Last 7 days
		cutoff = now.AddDate(0, 0, -7)
	case ThisMonth:
		// Last 30 days
		cutoff = now.AddDate(0, 0, -30)
	case ThisYear:
		// Last 365 days
		cutoff = now.AddDate(0, 0, -365)
	}

	// Count activities by hour (0-23)
	hourCounts := make(map[int]int)

	for _, activity := range activities {
		// Filter by time window
		if activity.Timestamp.Before(cutoff) {
			continue
		}

		hour := activity.Timestamp.Hour()
		hourCounts[hour]++
	}

	// Find peak hour (iterate in sorted order for deterministic results)
	maxCount := 0
	peakHour := 0

	// Sort hours to ensure consistent results when counts are equal
	var hours []int
	for hour := range hourCounts {
		hours = append(hours, hour)
	}

	// Sort hours so we pick the earliest hour if there's a tie
	for i := 0; i < len(hours); i++ {
		for j := i + 1; j < len(hours); j++ {
			if hours[i] > hours[j] {
				hours[i], hours[j] = hours[j], hours[i]
			}
		}
	}

	for _, hour := range hours {
		count := hourCounts[hour]
		if count > maxCount {
			maxCount = count
			peakHour = hour
		}
	}

	// Format peak hour as time range
	var peakString string
	if maxCount == 0 {
		peakString = "No data"
	} else {
		// Convert to 12-hour format with AM/PM
		startHour := peakHour
		endHour := (peakHour + 1) % 24

		startPeriod := "AM"
		endPeriod := "AM"

		displayStart := startHour
		if startHour >= 12 {
			startPeriod = "PM"
			if startHour > 12 {
				displayStart = startHour - 12
			}
		}
		if displayStart == 0 {
			displayStart = 12
		}

		displayEnd := endHour
		if endHour >= 12 {
			endPeriod = "PM"
			if endHour > 12 {
				displayEnd = endHour - 12
			}
		}
		if displayEnd == 0 {
			displayEnd = 12
		}

		peakString = formatHourRange(displayStart, startPeriod, displayEnd, endPeriod)
	}

	return peakString, hourCounts
}

// formatHourRange formats hour range nicely
func formatHourRange(start int, startPeriod string, end int, endPeriod string) string {
	// Convert to lowercase for consistency
	startPeriod = strings.ToLower(startPeriod)
	endPeriod = strings.ToLower(endPeriod)

	if startPeriod == endPeriod {
		// Same period: "2-3pm"
		return fmt.Sprintf("%d-%d%s", start, end, endPeriod)
	}
	// Different periods: "11am-12pm"
	return fmt.Sprintf("%d%s-%d%s", start, startPeriod, end, endPeriod)
}

// formatTime formats hour number as string
func formatTime(hour int) string {
	return fmt.Sprintf("%d", hour)
}

// CalculateActivityStats calculates all activity statistics
func CalculateActivityStats(activities []Activity, pushGranularity PushGranularity, timeWindow TimeWindow) ActivityStats {
	stats := ActivityStats{}

	// Calculate push rate
	stats.PushRate = CalculatePushRate(activities, pushGranularity)

	// Calculate peak coding hour
	peakHour, distribution := CalculatePeakCodingHour(activities, timeWindow)
	stats.PeakCodingHour = peakHour
	stats.HourDistribution = distribution

	// TODO: Add PR rate, review rate, etc.

	return stats
}
