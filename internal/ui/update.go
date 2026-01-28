package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/feddle/daily-dash/internal/api/foli"
	"go.uber.org/zap"
)

// Update handles Bubble Tea messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle stop selection state
	if m.selectingStart || m.selectingEnd {
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.width = msg.Width
			m.height = msg.Height
			m.stopList.SetSize(msg.Width, msg.Height)

		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.selectingStart = false
				m.selectingEnd = false
				return m, nil

			case "enter":
				selectedItem := m.stopList.SelectedItem()
				if selectedItem != nil {
					if stop, ok := selectedItem.(foli.GTFSStop); ok {
						if m.selectingStart {
							m.foliStartStop = stop.Code
							m.foliStartStopName = stop.Name
							m.selectingStart = false
							m.selectingEnd = true
							m.stopList.Title = "Select End Stop"
							m.stopList.ResetFilter()
							return m, nil
						} else if m.selectingEnd {
							m.foliEndStop = stop.ID
							m.foliEndStopName = stop.Name
							m.selectingEnd = false
							return m, m.fetchTransitCmd()
						}
					}
				}
			}
		}

		var cmd tea.Cmd
		m.stopList, cmd = m.stopList.Update(msg)
		return m, cmd
	}

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

		case "f":
			if !m.loadingStops {
				// Check if we already have stops
				if len(m.stops) > 0 {
					m.selectingStart = true
					m.stopList.Title = "Select Start Stop"
					m.stopList.SetItems(stopsToListItems(m.stops))
					m.stopList.SetSize(m.width, m.height)
					return m, nil
				}
				// Fetch stops
				m.loadingStops = true
				return m, m.fetchStopsCmd()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.stopList.SetSize(msg.Width, msg.Height)
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

	case stopsFetchSuccessMsg:
		m.loadingStops = false
		m.stops = msg.stops
		m.selectingStart = true
		
		// Initialize list
		items := stopsToListItems(m.stops)
		m.stopList = list.New(items, list.NewDefaultDelegate(), m.width, m.height)
		m.stopList.Title = "Select Start Stop"
		
		return m, nil

	case stopsFetchErrorMsg:
		m.loadingStops = false
		m.logger.Error("failed to load stops", zap.Error(msg.err))
		// Maybe show a flash message? For now just log.
		return m, nil

	case clearCooldownMsg:
		m.showCooldownMsg = false
		return m, nil

	case tickMsg:
		return m, tickCmd()
	}

	return m, nil
}

func stopsToListItems(stops []foli.GTFSStop) []list.Item {
	items := make([]list.Item, len(stops))
	for i, stop := range stops {
		items[i] = stop
	}
	return items
}
