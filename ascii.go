package main

import (
	_ "embed"
	"strings"
)

//go:embed ascii.txt
var asciiFont string

// ASCII font map: A-Z, 0-9, @, space (3 lines each)
var asciiLines []string

func init() {
	asciiLines = strings.Split(strings.TrimSpace(asciiFont), "\n")
}

// RenderASCII converts text to ASCII art (3 lines tall)
// Supports: A-Z (case insensitive), 0-9, @, space
func RenderASCII(text string) string {
	text = strings.ToUpper(text)

	// Build 3 lines of output
	line1 := []string{}
	line2 := []string{}
	line3 := []string{}

	for _, char := range text {
		index := getCharIndex(char)
		if index == -1 {
			// Unknown character, use space
			line1 = append(line1, "   ")
			line2 = append(line2, "   ")
			line3 = append(line3, "   ")
			continue
		}

		// Each character is 3 lines starting at index*3
		startLine := index * 3
		line1 = append(line1, asciiLines[startLine])
		line2 = append(line2, asciiLines[startLine+1])
		line3 = append(line3, asciiLines[startLine+2])
	}

	return strings.Join(line1, " ") + "\n" +
		strings.Join(line2, " ") + "\n" +
		strings.Join(line3, " ")
}

// getCharIndex returns the index for a character in the ASCII font
// Order: A-Z (0-25), 0-9 (26-35), @ (36), space (37)
func getCharIndex(char rune) int {
	switch {
	case char >= 'A' && char <= 'Z':
		return int(char - 'A')
	case char >= '0' && char <= '9':
		return 26 + int(char-'0')
	case char == '@':
		return 36
	case char == ' ':
		return 37
	default:
		return -1 // Unknown character
	}
}
