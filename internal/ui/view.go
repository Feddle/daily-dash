package ui

import (
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/feddle/daily-dash/internal/ui/components"
)

var (
	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(2)

	notificationStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("11")). // Yellow
				Italic(true)
)

// View renders the current state of the UI
func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	if m.selectingStart || m.selectingEnd {
		return m.stopList.View()
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
	timeStr := time.Now().Format("15:04")
	footer := footerStyle.Render(timeStr + " | Press 'r' to refresh | Press 'f' to select stop | Press 'q' to quit")

	var notification string
	if m.showCooldownMsg {
		notification = notificationStyle.Render("Please wait a few seconds before refreshing again...")
	}

	// Combine all elements
	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		content,
		"",
		notification,
		footer,
	)
}
