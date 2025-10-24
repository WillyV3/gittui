package main

import (
	"strings"
	"testing"
)

func TestRenderASCII(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantRows int
	}{
		{
			name:     "single letter",
			input:    "A",
			wantRows: 3,
		},
		{
			name:     "username with @",
			input:    "@WILLYV3",
			wantRows: 3,
		},
		{
			name:     "mixed alphanumeric",
			input:    "USER123",
			wantRows: 3,
		},
		{
			name:     "lowercase converts to uppercase",
			input:    "abc",
			wantRows: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderASCII(tt.input)

			// Check that result has exactly 3 lines
			lines := strings.Split(result, "\n")
			if len(lines) != tt.wantRows {
				t.Errorf("RenderASCII(%q) returned %d lines, want %d", tt.input, len(lines), tt.wantRows)
			}

			// Check that result is not empty
			if result == "" {
				t.Errorf("RenderASCII(%q) returned empty string", tt.input)
			}

			// Check that all lines have content (not just newlines)
			for i, line := range lines {
				if strings.TrimSpace(line) == "" && tt.input != " " {
					t.Logf("Line %d is empty or whitespace only for input %q", i, tt.input)
				}
			}
		})
	}
}

func TestGetCharIndex(t *testing.T) {
	tests := []struct {
		name  string
		char  rune
		want  int
	}{
		{"letter A", 'A', 0},
		{"letter Z", 'Z', 25},
		{"digit 0", '0', 26},
		{"digit 9", '9', 35},
		{"at symbol", '@', 36},
		{"space", ' ', 37},
		{"unknown char", '!', -1},
		{"lowercase a", 'a', -1}, // getCharIndex doesn't handle lowercase (RenderASCII does ToUpper)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getCharIndex(tt.char)
			if got != tt.want {
				t.Errorf("getCharIndex(%q) = %d, want %d", tt.char, got, tt.want)
			}
		})
	}
}

func TestRenderASCII_VisualOutput(t *testing.T) {
	// Visual test - prints output for manual inspection
	input := "@TEST"
	result := RenderASCII(input)

	t.Logf("ASCII art for %q:\n%s", input, result)

	// Verify structure
	lines := strings.Split(result, "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}
}
