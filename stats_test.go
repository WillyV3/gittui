package main

import (
	"testing"
	"time"
)

func TestFormatTime(t *testing.T) {
	tests := []struct {
		hour     int
		expected string
	}{
		{1, "1"},
		{2, "2"},
		{9, "9"},
		{10, "10"},
		{11, "11"},
		{12, "12"},
	}

	for _, tt := range tests {
		result := formatTime(tt.hour)
		if result != tt.expected {
			t.Errorf("formatTime(%d) = %s, want %s", tt.hour, result, tt.expected)
		}
	}
}

func TestFormatHourRange(t *testing.T) {
	tests := []struct {
		start       int
		startPeriod string
		end         int
		endPeriod   string
		expected    string
	}{
		{2, "PM", 3, "PM", "2-3PM"},
		{11, "AM", 12, "PM", "11AM-12PM"},
		{1, "AM", 2, "AM", "1-2AM"},
	}

	for _, tt := range tests {
		result := formatHourRange(tt.start, tt.startPeriod, tt.end, tt.endPeriod)
		if result != tt.expected {
			t.Errorf("formatHourRange(%d%s, %d%s) = %s, want %s",
				tt.start, tt.startPeriod, tt.end, tt.endPeriod, result, tt.expected)
		}
	}
}

func TestCalculatePeakCodingHour(t *testing.T) {
	// Create test activities with known timestamps
	now := time.Now()
	activities := []Activity{
		{Type: "PushEvent", Timestamp: now.Add(-1 * time.Hour)},        // Hour: now-1
		{Type: "PushEvent", Timestamp: now.Add(-2 * time.Hour)},        // Hour: now-2
		{Type: "PushEvent", Timestamp: now.Add(-2 * time.Hour)},        // Hour: now-2 (duplicate)
		{Type: "IssueEvent", Timestamp: now.Add(-3 * time.Hour)},       // Hour: now-3
		{Type: "PushEvent", Timestamp: now.Add(-50 * time.Hour)},       // Outside ThisWeek window
	}

	peakHour, distribution := CalculatePeakCodingHour(activities, ThisWeek)

	// Should have counted 4 activities within the week
	totalCount := 0
	for _, count := range distribution {
		totalCount += count
	}

	if totalCount != 4 {
		t.Errorf("Expected 4 activities within week, got %d", totalCount)
	}

	// Peak should be the hour with 2 activities (now-2)
	expectedHour := now.Add(-2 * time.Hour).Hour()
	if distribution[expectedHour] != 2 {
		t.Errorf("Expected 2 activities at hour %d, got %d", expectedHour, distribution[expectedHour])
	}

	// Peak string should not be "No data"
	if peakHour == "No data" {
		t.Errorf("Expected peak hour string, got 'No data'")
	}

	t.Logf("Peak hour: %s", peakHour)
	t.Logf("Distribution: %v", distribution)
}
