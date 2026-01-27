package ui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// fetchWeatherCmd fetches weather data asynchronously
func (m Model) fetchWeatherCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		m.logger.Debug("starting weather fetch")
		weather, err := m.coordinator.FetchWeather(ctx)

		if err != nil {
			m.logger.Error("weather fetch failed")
			return weatherFetchErrorMsg{err: err}
		}

		m.logger.Debug("weather fetch successful")
		return weatherFetchSuccessMsg{weather: weather}
	}
}

// fetchTransitCmd fetches transit data asynchronously
func (m Model) fetchTransitCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		m.logger.Debug("starting transit fetch")
		transit, err := m.coordinator.FetchTransit(ctx)

		if err != nil {
			m.logger.Error("transit fetch failed")
			return transitFetchErrorMsg{err: err}
		}

		m.logger.Debug("transit fetch successful")
		return transitFetchSuccessMsg{transit: transit}
	}
}

// fetchRoadCmd fetches road conditions data asynchronously
func (m Model) fetchRoadCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		m.logger.Debug("starting road conditions fetch")
		roadConditions, err := m.coordinator.FetchRoadConditions(ctx)

		if err != nil {
			m.logger.Error("road conditions fetch failed")
			return roadFetchErrorMsg{err: err}
		}

		m.logger.Debug("road conditions fetch successful")
		return roadFetchSuccessMsg{roadConditions: roadConditions}
	}
}

// fetchAllCmd fetches all data (weather, transit, and road) asynchronously
func (m Model) fetchAllCmd() tea.Cmd {
	return tea.Batch(
		m.fetchWeatherCmd(),
		m.fetchTransitCmd(),
		m.fetchRoadCmd(),
	)
}
