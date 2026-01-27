package components

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			MarginBottom(1)

	headerSubtitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))
)

// RenderHeader renders the header with title and last update time
func RenderHeader(lastUpdate time.Time) string {
	title := headerStyle.Render("Daily Dash - Turku Dashboard")

	var subtitle string
	if !lastUpdate.IsZero() {
		subtitle = headerSubtitleStyle.Render(
			fmt.Sprintf("Last updated: %s", lastUpdate.Format("15:04:05")),
		)
	} else {
		subtitle = headerSubtitleStyle.Render("Press 'r' to refresh")
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, subtitle)
}
