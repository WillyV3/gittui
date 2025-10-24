package main

import (
	"fmt"
	"os"
)

func main() {
	// Parse pattern from command line
	pattern := "random"
	if len(os.Args) > 1 {
		pattern = os.Args[1]
	}

	// Validate pattern
	validPatterns := []string{"random", "active", "streaky", "gradient", "sparse", "consistent", "custom"}
	if !isValidPattern(pattern, validPatterns) {
		printUsage(validPatterns)
		os.Exit(1)
	}

	// Generate test data
	contributions := generateContributions(pattern)

	// Create and render graph
	graph := NewGraph(contributions)
	output := graph.Render()

	// Clear screen and display
	fmt.Print("\033[H\033[2J")
	fmt.Println(output)

	// Show statistics
	stats := CalculateStats(contributions)
	fmt.Printf("\n%s\n", stats.String())

	// Show usage hint
	fmt.Printf("\nPattern: %s\n", pattern)
	printUsageHint()
}

// generateContributions creates test data based on the pattern.
func generateContributions(pattern string) []Contribution {
	if pattern == "custom" {
		return GenerateCustomData()
	}
	return GenerateTestData(pattern)
}

// isValidPattern checks if the pattern is in the valid list.
func isValidPattern(pattern string, valid []string) bool {
	for _, v := range valid {
		if pattern == v {
			return true
		}
	}
	return false
}

// printUsage shows all available patterns.
func printUsage(patterns []string) {
	fmt.Printf("Unknown pattern. Available patterns:\n")
	for _, p := range patterns {
		fmt.Printf("  %s\n", p)
	}
}

// printUsageHint shows how to try other patterns.
func printUsageHint() {
	fmt.Println("\nTry different patterns:")
	fmt.Println("  go run . random")
	fmt.Println("  go run . active")
	fmt.Println("  go run . streaky")
	fmt.Println("  go run . gradient")
	fmt.Println("  go run . sparse")
	fmt.Println("  go run . consistent")
	fmt.Println("  go run . custom")
}
