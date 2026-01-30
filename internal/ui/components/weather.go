package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/feddle/daily-dash/internal/domain"
)

var (
	// Weather panel styles
	weatherBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("39")).
				Padding(1, 2).
				Width(35)

	weatherTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("39"))

	weatherLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("246"))

	weatherValueStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("255"))

	weatherErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196"))

	weatherLoadingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("226"))
)

// RenderWeather renders the weather panel
func RenderWeather(weather *domain.Weather, loading bool, err error) string {
	var content string

	if loading {
		content = weatherLoadingStyle.Render("⟳ Loading weather data...")
	} else if err != nil {
		content = weatherErrorStyle.Render("⚠ Error: Data fetch failed")
	} else if weather == nil {
		content = weatherLabelStyle.Render("No weather data available")
	} else {
		// Temperature icon
		tempIcon := getTemperatureIcon(weather.Temperature)

		// Build weather display
		lines := []string{
			fmt.Sprintf("%s %s: %s",
				tempIcon,
				weatherLabelStyle.Render("Temperature"),
				weatherValueStyle.Render(fmt.Sprintf("%.1f°C", weather.Temperature)),
			),
			fmt.Sprintf("💧 %s: %s",
				weatherLabelStyle.Render("Humidity"),
				weatherValueStyle.Render(fmt.Sprintf("%.0f%%", weather.Humidity)),
			),
			fmt.Sprintf("💨 %s: %s",
				weatherLabelStyle.Render("Wind"),
				weatherValueStyle.Render(fmt.Sprintf("%.1f m/s", weather.WindSpeed)),
			),
			fmt.Sprintf("%s %s: %s",
				getConditionsIcon(weather.Conditions),
				weatherLabelStyle.Render("Conditions"),
				weatherValueStyle.Render(weather.Conditions),
			),
			"",
			weatherLabelStyle.Render(fmt.Sprintf("Updated: %s",
				weather.Timestamp.Format("15:04:05"))),
		}

		content = lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

	locationName := "Unknown"
	if weather != nil && weather.Location != "" {
		locationName = weather.Location
	}

	title := weatherTitleStyle.Render(fmt.Sprintf("Weather (%s)", locationName))

	panel := lipgloss.JoinVertical(lipgloss.Left, title, "", content)

	return weatherBorderStyle.Render(panel)
}

// getTemperatureIcon returns an icon based on temperature
func getTemperatureIcon(temp float64) string {
	if temp < -10 {
		return "❄️"
	} else if temp < 0 {
		return "🌡️"
	} else if temp < 15 {
		return "🌡️"
	} else if temp < 25 {
		return "☀️"
	} else {
		return "🔥"
	}
}

// getConditionsIcon returns an icon based on weather conditions
func getConditionsIcon(conditions string) string {
	switch conditions {
	case "Clear":
		return "☀️"
	case "Partly Cloudy":
		return "⛅"
	case "Cloudy":
		return "☁️"
	case "Rainy":
		return "🌧️"
	case "Snowy":
		return "❄️"
	case "Foggy":
		return "🌫️"
	case "Windy", "Breezy":
		return "💨"
	default:
		return "🌡️"
	}
}
