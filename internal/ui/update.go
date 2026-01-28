package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"
)

// Update handles Bubble Tea messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "r":
			// Cooldown to prevent API spamming
			if time.Since(m.lastRefresh) < m.refreshCooldown {
				m.logger.Info("refresh requested too soon, ignoring",
					zap.Duration("since_last", time.Since(m.lastRefresh)),
				)
				m.showCooldownMsg = true
				return m, clearCooldownCmd()
			}

			m.logger.Info("refresh requested")
			m.lastRefresh = time.Now()
			m.weatherLoading = true
			m.transitLoading = true
			m.roadLoading = true
			m.weatherErr = nil
			m.transitErr = nil
			m.roadErr = nil
			return m, m.fetchAllCmd()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case weatherFetchStartMsg:
		m.weatherLoading = true
		m.weatherErr = nil
		return m, nil

	case weatherFetchSuccessMsg:
		m.weatherLoading = false
		m.weatherData = msg.weather
		m.weatherErr = nil
		m.lastUpdate = time.Now()
		m.logger.Info("weather data updated successfully",
			zap.Float64("temperature", msg.weather.Temperature),
		)
		return m, nil

	case weatherFetchErrorMsg:
		m.weatherLoading = false
		m.weatherErr = msg.err
		m.logger.Error("weather fetch error",
			zap.Error(msg.err),
		)
		return m, nil

	case transitFetchStartMsg:
		m.transitLoading = true
		m.transitErr = nil
		return m, nil

	case transitFetchSuccessMsg:
		m.transitLoading = false
		m.transitData = msg.transit
		m.transitErr = nil
		m.lastUpdate = time.Now()
		m.logger.Info("transit data updated successfully",
			zap.Int("departure_count", len(msg.transit.Departures)),
		)
		return m, nil

	case transitFetchErrorMsg:
		m.transitLoading = false
		m.transitErr = msg.err
		m.logger.Error("transit fetch error",
			zap.Error(msg.err),
		)
		return m, nil

	case roadFetchStartMsg:
		m.roadLoading = true
		m.roadErr = nil
		return m, nil

	case roadFetchSuccessMsg:
		m.roadLoading = false
		m.roadData = msg.roadConditions
		m.roadErr = nil
		m.lastUpdate = time.Now()
		m.logger.Info("road conditions data updated successfully",
			zap.Int("segment_count", len(msg.roadConditions.Segments)),
		)
		return m, nil

	case roadFetchErrorMsg:
		m.roadLoading = false
		m.roadErr = msg.err
		m.logger.Error("road conditions fetch error",
			zap.Error(msg.err),
		)
		return m, nil

	case clearCooldownMsg:
		m.showCooldownMsg = false
		return m, nil
	}

	return m, nil
}
