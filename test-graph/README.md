# Contribution Graph Test Harness

Standalone test environment for perfecting the contribution graph rendering.

## Quick Start

```bash
# Install dependencies
go mod tidy

# Run with random pattern
go run .

# Try different patterns
go run . active
go run . streaky
go run . gradient
go run . sparse
go run . consistent
go run . custom
```

## Patterns

- **random**: Random contributions (0-14 per day)
- **active**: High activity user (5-24 per day)
- **streaky**: Active weekdays, quiet weekends
- **gradient**: Increasing activity over the year
- **sparse**: Mostly empty with occasional bursts
- **consistent**: Steady 5-9 contributions per day
- **custom**: Wave pattern for visual testing

## Architecture

- `graph.go`: Core rendering logic using half-height blocks (▀)
- `testdata.go`: Test data generators
- `main.go`: CLI entry point

## Testing Strategy

1. Test with different patterns to verify rendering
2. Check color levels (0-4) display correctly
3. Verify squares look actually square
4. Test edge cases (empty weeks, high counts)

## Testing Responsive Width

The graph requires 108 columns minimum. To test the responsive message:

```bash
# Resize terminal to narrow width, then run:
go run . streaky

# Or test in responsive mode (handles resize):
go run . --interactive
```

## Once Perfect

Copy `graph.go` logic into main gittui app and replace test data with real GitHub API responses.
