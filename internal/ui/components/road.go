package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/feddle/daily-dash/internal/domain"
)

var (
	// Road panel styles
	roadBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39")).
			Padding(1, 2).
			Width(87) // Match width of top row (35 + 2 + 50)

	roadTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39"))

	roadNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")) // Green

	roadSlipperyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("226")) // Yellow

	roadDifficultStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")) // Red

	roadUnknownStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")) // Grey
)

// RenderRoad renders the road conditions panel
func RenderRoad(roadConditions *domain.RoadConditions, loading bool, err error) string {
	var content string

	if loading {
		content = weatherLoadingStyle.Render("⟳ Loading road data...")
	} else if err != nil {
		content = weatherErrorStyle.Render("⚠ Error: Data fetch failed")
	} else if roadConditions == nil || len(roadConditions.Segments) == 0 {
		content = weatherLabelStyle.Render("No road data available")
	} else {
		// Build road conditions list
		lines := []string{}

		count := 0
		maxCount := 5
		for _, segment := range roadConditions.Segments {
			if count >= maxCount {
				break
			}
			count++
			var statusIcon string
			var rowStyle lipgloss.Style

			switch segment.Condition {
			case domain.Normal:
				statusIcon = "✓"
				rowStyle = roadNormalStyle
			case domain.Slippery:
				statusIcon = "⚠"
				rowStyle = roadSlipperyStyle
			case domain.Difficult:
				statusIcon = "✗"
				rowStyle = roadDifficultStyle
			case domain.Unknown:
				statusIcon = "?"
				rowStyle = roadUnknownStyle
			}

			// Format route and condition
			line := fmt.Sprintf("%s %-25s: %-18s Road: %5.1f°C  Air: %5.1f°C",
				statusIcon,
				segment.Route,
				segment.Description,
				segment.Temperature,
				segment.AirTemperature,
			)

			lines = append(lines, rowStyle.Render(line))
		}

		// Add legend
		if len(lines) > 0 {
			lines = append(lines, "")
			legend := fmt.Sprintf("%s Normal  %s Slippery  %s Difficult  %s Unknown",
				roadNormalStyle.Render("✓"),
				roadSlipperyStyle.Render("⚠"),
				roadDifficultStyle.Render("✗"),
				roadUnknownStyle.Render("?"),
			)
			lines = append(lines, weatherLabelStyle.Render(legend))
		}

		content = lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

	title := roadTitleStyle.Render("Road Conditions")

	panel := lipgloss.JoinVertical(lipgloss.Left, title, "", content)

	return roadBorderStyle.Render(panel)
}
