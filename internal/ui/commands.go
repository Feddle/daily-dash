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

		m.logger.Debug("starting transit fetch", zap.String("stop", m.foliStartStop), zap.String("dest", m.foliEndStop))
		transit, err := m.coordinator.FetchTransit(ctx, m.foliStartStop, m.foliStartStopName, m.foliEndStop, m.foliEndStopName)

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

		m.logger.Debug("starting road conditions fetch", zap.String("region", m.selectedRoadRegion))
		roadConditions, err := m.coordinator.FetchRoadConditions(ctx, m.selectedRoadRegion)

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

// fetchRoadStationsCmd fetches all road stations (regions)
func (m Model) fetchRoadStationsCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Fetch all conditions (empty region) to get list of available stations
		m.logger.Debug("fetching road stations list")
		conditions, err := m.coordinator.FetchRoadConditions(ctx, "*")

		if err != nil {
			return roadStationsFetchErrorMsg{err: err}
		}

		var stations []string
		seen := make(map[string]bool)

		if conditions != nil {
			for _, cond := range conditions.Segments {
				// With Forecast API, cond.Route is a descriptive string (e.g. "Vt 1: Helsinki - Turku")
				// We can just use that directly as the selectable item.
				// OR we can try to extract the city/region name if we want to mimic previous behavior.
				// But "Vt 1: Helsinki - Turku" is a much better selectable item than just "Turku".
				// So we use the full description.

				name := cond.Route
				if !seen[name] {
					stations = append(stations, name)
					seen[name] = true
				}
			}
		}

		return roadStationsFetchSuccessMsg{stations: stations}
	}
}

// tickCmd returns a command that triggers every minute to update the clock
func tickCmd() tea.Cmd {
	return tea.Tick(time.Minute, func(t time.Time) tea.Msg {
		return tickMsg{time: t}
	})
}

// updateWeatherLocationCmd updates the location and triggers a fetch
func (m Model) updateWeatherLocationCmd(location string) tea.Cmd {
	return func() tea.Msg {
		m.coordinator.SetWeatherLocation(location)
		// We trigger the fetch immediately, but we can reuse the existing fetchWeatherCmd
		// However, we need to return a Msg that update.go can handle if we want to chain things,
		// or we can just return the fetch command directly.
		// Since SetWeatherLocation is synchronous and thread-safe, we can just return the fetch cmd.
		return m.fetchWeatherCmd()()
	}
}

// fetchWeatherStationsCmd fetches weather stations asynchronously
func (m Model) fetchWeatherStationsCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		m.logger.Debug("starting weather stations fetch")
		stations, err := m.coordinator.FetchWeatherStations(ctx)

		if err != nil {
			m.logger.Error("weather stations fetch failed", zap.Error(err))
			return weatherStationsFetchErrorMsg{err: err}
		}

		m.logger.Debug("weather stations fetch successful", zap.Int("count", len(stations)))
		return weatherStationsFetchSuccessMsg{stations: stations}
	}
}
