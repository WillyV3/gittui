package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

// GraphComponent displays the contribution graph
type GraphComponent struct {
	graph   *Graph
	width   int
	loading bool
}

// NewGraphComponent creates a new graph component
func NewGraphComponent() GraphComponent {
	return GraphComponent{
		loading: true,
	}
}

// SetContributions updates the contribution data
func (g *GraphComponent) SetContributions(contributions []Contribution) {
	g.graph = NewGraph(contributions)
	g.loading = false
}

// SetWidth updates the width for responsive rendering
func (g *GraphComponent) SetWidth(width int) {
	g.width = width
}

// Update handles messages
func (g GraphComponent) Update(msg tea.Msg) (GraphComponent, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		g.width = msg.Width
	}
	return g, nil
}

// View renders the graph component
func (g GraphComponent) View() string {
	if g.loading || g.graph == nil {
		return labelStyle.Render("Loading contributions...")
	}

	return g.graph.Render()
}
