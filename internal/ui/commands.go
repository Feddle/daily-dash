package ui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"
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

		m.logger.Debug("starting transit fetch", zap.String("stop", m.foliStartStop))
		transit, err := m.coordinator.FetchTransit(ctx, m.foliStartStop)

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

// fetchStopsCmd fetches all stops asynchronously
func (m Model) fetchStopsCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		m.logger.Debug("starting stops fetch")
		stops, err := m.coordinator.FetchStops(ctx)

		if err != nil {
			m.logger.Error("stops fetch failed", zap.Error(err))
			return stopsFetchErrorMsg{err: err}
		}

		m.logger.Debug("stops fetch successful", zap.Int("count", len(stops)))
		return stopsFetchSuccessMsg{stops: stops}
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

// clearCooldownCmd clears the cooldown message after a delay
func clearCooldownCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return clearCooldownMsg{}
	})
}

// tickCmd returns a command that triggers every minute to update the clock
func tickCmd() tea.Cmd {
	return tea.Tick(time.Minute, func(t time.Time) tea.Msg {
		return tickMsg{time: t}
	})
}
