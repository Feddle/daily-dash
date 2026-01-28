package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/feddle/daily-dash/internal/domain"
)

var (
	// Transit panel styles
	transitBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("39")).
				Padding(1, 2).
				Width(50)

	transitTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("39"))

	transitHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("246"))


	transitOnTimeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("82")) // Green

	transitDelayedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("226")) // Yellow

	transitCancelledStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")) // Red
)

// RenderTransit renders the transit panel
func RenderTransit(transit *domain.Transit, loading bool, err error) string {
	var content string

	if loading {
		content = weatherLoadingStyle.Render("⟳ Loading transit data...")
	} else if err != nil {
		content = weatherErrorStyle.Render(fmt.Sprintf("⚠ Error: %v", err))
	} else if transit == nil || len(transit.Departures) == 0 {
		content = weatherLabelStyle.Render("No transit data available")
	} else {
		// Build departure table
		lines := []string{
			transitHeaderStyle.Render(fmt.Sprintf("%-20s %8s %6s", "Stop", "Time", "ETA")),
			transitHeaderStyle.Render(strings.Repeat("─", 45)),
		}

		for _, dep := range transit.Departures {
			// Format stop name (truncate if too long)
			stop := dep.Stop
			if len(stop) > 20 {
				stop = stop[:17] + "..."
			}

			// Format time
			timeStr := dep.ExpectedTime.Format("15:04")

			// Format minutes until
			var etaStr string
			var statusIcon string
			var rowStyle lipgloss.Style

			switch dep.Status {
			case domain.OnTime:
				statusIcon = "●"
				rowStyle = transitOnTimeStyle
			case domain.Delayed:
				statusIcon = "◐"
				rowStyle = transitDelayedStyle
			case domain.Cancelled:
				statusIcon = "✗"
				rowStyle = transitCancelledStyle
			}

			switch dep.MinutesUntil {
			case 0:
				etaStr = "Now"
			case 1:
				etaStr = "1 min"
			default:
				etaStr = fmt.Sprintf("%d min", dep.MinutesUntil)
			}

			row := fmt.Sprintf("%-20s %8s %6s", stop, timeStr, etaStr)
			lines = append(lines, rowStyle.Render(fmt.Sprintf("%s %s", statusIcon, row)))
		}

		// Add legend
		lines = append(lines, "")
		legend := fmt.Sprintf("%s On Time  %s Delayed  %s Cancelled",
			transitOnTimeStyle.Render("●"),
			transitDelayedStyle.Render("◐"),
			transitCancelledStyle.Render("✗"),
		)
		lines = append(lines, transitHeaderStyle.Render(legend))

		content = lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

	title := transitTitleStyle.Render(fmt.Sprintf("Föli Line %s", getLineNumber(transit)))

	panel := lipgloss.JoinVertical(lipgloss.Left, title, "", content)

	return transitBorderStyle.Render(panel)
}

// getLineNumber returns the line number from transit data
func getLineNumber(transit *domain.Transit) string {
	if transit == nil {
		return "?"
	}
	if transit.Line == "" {
		return "1"
	}
	return transit.Line
}
