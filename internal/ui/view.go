package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/feddle/daily-dash/internal/ui/components"
)

var (
	footerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)
)

// View renders the current state of the UI
func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	// Render header
	header := components.RenderHeader(m.lastUpdate)

	// Render weather panel
	weatherPanel := components.RenderWeather(m.weatherData, m.weatherLoading, m.weatherErr)

	// Render transit panel
	transitPanel := components.RenderTransit(m.transitData, m.transitLoading, m.transitErr)

	// Render road panel
	roadPanel := components.RenderRoad(m.roadData, m.roadLoading, m.roadErr)

	// Build main content area - arrange weather and transit horizontally
	topRow := lipgloss.JoinHorizontal(lipgloss.Top,
		weatherPanel,
		"  ", // Spacing between panels
		transitPanel,
	)

	// Build bottom row with road conditions
	bottomRow := roadPanel

	// Combine rows vertically
	content := lipgloss.JoinVertical(lipgloss.Left,
		topRow,
		"",
		bottomRow,
	)

	// Build footer
	footer := footerStyle.Render("Press 'r' to refresh | Press 'q' to quit")

	// Combine all elements
	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		content,
		footer,
	)
}
