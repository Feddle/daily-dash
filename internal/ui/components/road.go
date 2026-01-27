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
			Width(35)

	roadTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39"))

	roadNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")) // Green

	roadSlipperyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("226")) // Yellow

	roadDifficultStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")) // Red
)

// RenderRoad renders the road conditions panel
func RenderRoad(roadConditions *domain.RoadConditions, loading bool, err error) string {
	var content string

	if loading {
		content = weatherLoadingStyle.Render("⟳ Loading road data...")
	} else if err != nil {
		content = weatherErrorStyle.Render(fmt.Sprintf("⚠ Error: %v", err))
	} else if roadConditions == nil || len(roadConditions.Segments) == 0 {
		content = weatherLabelStyle.Render("No road data available")
	} else {
		// Build road conditions list
		lines := []string{}

		for _, segment := range roadConditions.Segments {
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
			}

			// Format route and condition
			line := fmt.Sprintf("%s %s: %s (%.1f°C)",
				statusIcon,
				segment.Route,
				segment.Description,
				segment.Temperature,
			)

			lines = append(lines, rowStyle.Render(line))
		}

		// Add legend
		if len(lines) > 0 {
			lines = append(lines, "")
			legend := fmt.Sprintf("%s Normal  %s Slippery  %s Difficult",
				roadNormalStyle.Render("✓"),
				roadSlipperyStyle.Render("⚠"),
				roadDifficultStyle.Render("✗"),
			)
			lines = append(lines, weatherLabelStyle.Render(legend))
		}

		content = lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

	title := roadTitleStyle.Render(fmt.Sprintf("Road Conditions (%s)", getRegionName(roadConditions)))

	panel := lipgloss.JoinVertical(lipgloss.Left, title, "", content)

	return roadBorderStyle.Render(panel)
}

// getRegionName returns the region name from road conditions data
func getRegionName(roadConditions *domain.RoadConditions) string {
	if roadConditions == nil {
		return "?"
	}
	if roadConditions.Region == "" {
		return "Turku"
	}
	return roadConditions.Region
}
