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
				Padding(1, 2)

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
		content = weatherErrorStyle.Render("⚠ Error: Data fetch failed")
	} else if transit == nil || len(transit.Departures) == 0 {
		content = weatherLabelStyle.Render("No transit data available")
	} else {
		// Check if we have destination data
		hasDestination := false
		for _, dep := range transit.Departures {
			if dep.DestinationStop != "" {
				hasDestination = true
				break
			}
		}

		// Width adjustment
		width := 50
		if hasDestination {
			width = 80
		}

		// Build departure table
		var header string
		var separator string

		if hasDestination {
			header = transitHeaderStyle.Render(fmt.Sprintf("%-15s %5s %6s  %-15s %5s", "Start", "Time", "ETA", "End", "Arr"))
			separator = transitHeaderStyle.Render(strings.Repeat("─", 75))
		} else {
			header = transitHeaderStyle.Render(fmt.Sprintf("%-20s %8s %6s", "Stop", "Time", "ETA"))
			separator = transitHeaderStyle.Render(strings.Repeat("─", 45))
		}

		lines := []string{header, separator}

		for _, dep := range transit.Departures {
			// Format stop name (truncate if too long)
			stop := dep.Stop
			maxLen := 17
			if hasDestination {
				maxLen = 15
			}
			if len(stop) > maxLen {
				stop = stop[:maxLen-3] + "..."
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

			var row string
			if hasDestination {
				destStop := "Unknown"
				// We don't have destination stop name in Departure, only ID?
				// Wait, domain Departure has DestinationStop string.
				// In client.go I set it to destStopID.
				// For now I'll just show what I have.
				if dep.DestinationStop != "" {
					destStop = dep.DestinationStop
				}
				if len(destStop) > maxLen {
					destStop = destStop[:maxLen-3] + "..."
				}

				destTimeStr := "-"
				if !dep.ArrivalTime.IsZero() {
					destTimeStr = dep.ArrivalTime.Format("15:04")
				}

				row = fmt.Sprintf("%-15s %5s %6s  %-15s %5s", stop, timeStr, etaStr, destStop, destTimeStr)
			} else {
				row = fmt.Sprintf("%-20s %8s %6s", stop, timeStr, etaStr)
			}

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

		// Set width on border style for this render
		// Set width on border style for this render

		title := transitTitleStyle.Render(fmt.Sprintf("Föli Line %s", getLineNumber(transit)))
		var panel string

		if transit.Warning != "" {
			warningMsg := transit.Warning
			if len(transit.Departures) > 0 && transit.Departures[0].DebugInfo != "" {
				warningMsg += fmt.Sprintf("\n%s", transit.Departures[0].DebugInfo)
			}
			warning := transitCancelledStyle.Render(fmt.Sprintf("\n⚠ %s", warningMsg))
			panel = lipgloss.JoinVertical(lipgloss.Left, title, "", content, warning)
		} else {
			panel = lipgloss.JoinVertical(lipgloss.Left, title, "", content)
		}

		return transitBorderStyle.Width(width).Render(panel)
	}

	// Default return for error/loading case
	return transitBorderStyle.Width(50).Render(content)
}

// getLineNumber returns the line number from transit data
func getLineNumber(transit *domain.Transit) string {
	if transit == nil {
		return "?"
	}
	if transit.Line == "" {
		return "All"
	}
	return transit.Line
}
