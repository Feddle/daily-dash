package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/feddle/daily-dash/internal/api/fmi"
	"github.com/feddle/daily-dash/internal/api/foli"
	"go.uber.org/zap"
)

// Update handles Bubble Tea messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle selection states
	// Handle selection states
	if m.selectingLocation {
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.width = msg.Width
			m.height = msg.Height
			m.weatherList.SetSize(msg.Width, msg.Height-2)

		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.selectingLocation = false
				return m, nil
			case "enter":
				selectedItem := m.weatherList.SelectedItem()
				if selectedItem != nil {
					if wi, ok := selectedItem.(weatherItem); ok {
						m.selectingLocation = false
						m.weatherLoading = true
						return m, m.updateWeatherLocationCmd(wi.id)
					}
					if i, ok := selectedItem.(item); ok {
						m.selectingLocation = false
						m.weatherLoading = true
						return m, m.updateWeatherLocationCmd(string(i))
					}
				}
				return m, nil
			default:
				// Auto-enable filtering
				if m.weatherList.FilterState() != list.Filtering {
					if msg.Type == tea.KeyRunes {
						m.weatherList, _ = m.weatherList.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
						if msg.String() == "/" {
							return m, nil
						}
					}
				}
			}
		}
		var cmd tea.Cmd
		m.weatherList, cmd = m.weatherList.Update(msg)
		return m, cmd
	}

	if m.selectingStart || m.selectingEnd || m.selectingRoad {
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.width = msg.Width
			m.height = msg.Height
			if m.selectingRoad {
				m.roadStationList.SetSize(msg.Width, msg.Height-2)
			} else {
				m.stopList.SetSize(msg.Width, msg.Height-2)
			}

		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.selectingStart = false
				m.selectingEnd = false
				m.selectingRoad = false
				return m, nil

			case "enter":
				// Handle Road Selection
				if m.selectingRoad {
					selectedItem := m.roadStationList.SelectedItem()
					if selectedItem != nil {
						if stationItem, ok := selectedItem.(item); ok {
							m.selectedRoadRegion = string(stationItem)
							m.selectingRoad = false
							m.roadLoading = true
							m.roadErr = nil
							return m, m.fetchRoadCmd()
						}
					}
					return m, nil
				}

				// Handle Transit Stop Selection
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

			default:
				// Auto-enable filtering if the user types a character (and isn't already filtering)
				// For Road List
				if m.selectingRoad && m.roadStationList.FilterState() != list.Filtering {
					if msg.Type == tea.KeyRunes {
						m.roadStationList, _ = m.roadStationList.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
						if msg.String() == "/" {
							return m, nil
						}
					}
				} else if !m.selectingRoad && m.stopList.FilterState() != list.Filtering {
					// For Transit List
					if msg.Type == tea.KeyRunes {
						m.stopList, _ = m.stopList.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
						if msg.String() == "/" {
							return m, nil
						}
					}
				}
			}
		}

		var cmd tea.Cmd
		if m.selectingRoad {
			m.roadStationList, cmd = m.roadStationList.Update(msg)
		} else {
			m.stopList, cmd = m.stopList.Update(msg)
		}
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

		case "R": // Shift+r
			if !m.loadingRoadStations {
				// Check if we already have stations
				if len(m.roadStations) > 0 {
					m.selectingRoad = true
					m.selectingStart = false // Ensure other modes are off
					m.selectingEnd = false

					m.roadStationList.Title = "Select Road Region"
					m.roadStationList.ResetFilter()
					m.roadStationList.SetItems(stringToItems(m.roadStations))
					m.roadStationList.SetSize(m.width, m.height-2)
					return m, nil
				}
				// Fetch stations
				m.loadingRoadStations = true
				return m, m.fetchRoadStationsCmd()
			}

		case "S": // Shift+s
			if !m.loadingStops {
				// Check if we already have stops
				if len(m.stops) > 0 {
					m.selectingStart = true
					m.selectingRoad = false
					m.selectingEnd = false

					m.stopList.Title = "Select Start Stop"
					m.stopList.ResetFilter()
					m.stopList.SetItems(stopsToListItems(m.stops))
					m.stopList.SetSize(m.width, m.height-2)
					return m, nil
				}
				// Fetch stops
				m.loadingStops = true
				return m, m.fetchStopsCmd()
			}

		case "W": // Shift+w for Weather Location
			if !m.loadingStations {
				// If stations already cached, show immediately
				if len(m.weatherStations) > 0 {
					m.selectingLocation = true
					m.selectingStart = false
					m.selectingEnd = false
					m.selectingRoad = false
					m.weatherList.ResetFilter()
					m.weatherList.SetItems(weatherStationsToItems(m.weatherStations))
					m.weatherList.SetSize(m.width, m.height-2)
					return m, nil
				}
				// First time: fetch stations
				m.loadingStations = true
				return m, m.fetchWeatherStationsCmd()
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
		m.stopList = list.New(items, list.NewDefaultDelegate(), m.width, m.height-2)
		m.stopList.Title = "Select Start Stop"
		m.stopList.SetShowHelp(false)
		m.stopList.SetShowStatusBar(false)
		m.stopList.SetShowPagination(false)

		// Disable keys that we don't want to show but keep Filter enabled for simulation
		m.stopList.KeyMap.ShowFullHelp = key.NewBinding() // Disable '?'
		m.stopList.KeyMap.Quit = key.NewBinding()         // Disable 'q' in list

		return m, nil

	case stopsFetchErrorMsg:
		m.loadingStops = false
		m.logger.Error("failed to load stops", zap.Error(msg.err))
		// Maybe show a flash message? For now just log.
		return m, nil

	case roadStationsFetchSuccessMsg:
		m.loadingRoadStations = false
		m.roadStations = msg.stations
		m.selectingRoad = true
		m.selectingStart = false
		m.selectingEnd = false

		// Initialize list
		items := stringToItems(m.roadStations)
		m.roadStationList = list.New(items, list.NewDefaultDelegate(), m.width, m.height-2)
		m.roadStationList.Title = "Select Road Region"
		m.roadStationList.SetShowHelp(false)
		m.roadStationList.SetShowStatusBar(false)
		m.roadStationList.SetShowPagination(false)

		// Disable keys
		m.roadStationList.KeyMap.ShowFullHelp = key.NewBinding()
		m.roadStationList.KeyMap.Quit = key.NewBinding()

		return m, nil

	case roadStationsFetchErrorMsg:
		m.loadingRoadStations = false
		m.logger.Error("failed to load road stations", zap.Error(msg.err))
		return m, nil

	case weatherStationsFetchSuccessMsg:
		m.loadingStations = false
		m.weatherStations = msg.stations
		m.selectingLocation = true
		m.selectingStart = false
		m.selectingEnd = false
		m.selectingRoad = false

		// Initialize list
		items := weatherStationsToItems(m.weatherStations)
		m.weatherList = list.New(items, list.NewDefaultDelegate(), m.width, m.height-2)
		m.weatherList.Title = "Select Weather Location"
		m.weatherList.SetShowHelp(false)
		m.weatherList.SetShowStatusBar(false)
		m.weatherList.SetShowPagination(false)
		m.weatherList.KeyMap.ShowFullHelp = key.NewBinding()
		m.weatherList.KeyMap.Quit = key.NewBinding()
		return m, nil

	case weatherStationsFetchErrorMsg:
		m.loadingStations = false
		m.logger.Error("failed to load weather stations, using fallback", zap.Error(msg.err))

		// Use fallback stations
		m.weatherStations = fmi.FallbackStations
		m.selectingLocation = true
		m.selectingStart = false
		m.selectingEnd = false
		m.selectingRoad = false

		// Initialize list with fallback
		items := weatherStationsToItems(m.weatherStations)
		m.weatherList = list.New(items, list.NewDefaultDelegate(), m.width, m.height-2)
		m.weatherList.Title = "Select Weather Location (Fallback)"
		m.weatherList.SetShowHelp(false)
		m.weatherList.SetShowStatusBar(false)
		m.weatherList.SetShowPagination(false)
		m.weatherList.KeyMap.ShowFullHelp = key.NewBinding()
		m.weatherList.KeyMap.Quit = key.NewBinding()
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

// Simple string item for list
type item string

func (i item) FilterValue() string { return string(i) }
func (i item) Title() string       { return string(i) }
func (i item) Description() string { return "" }

func stringToItems(strings []string) []list.Item {
	items := make([]list.Item, len(strings))
	for i, s := range strings {
		items[i] = item(s)
	}
	return items
}

type weatherItem struct {
	name string
	id   string
}

func (i weatherItem) FilterValue() string { return i.name }
func (i weatherItem) Title() string       { return i.name }
func (i weatherItem) Description() string { return i.id }

func weatherStationsToItems(stations []fmi.WeatherStation) []list.Item {
	items := make([]list.Item, len(stations))
	for i, s := range stations {
		items[i] = weatherItem{name: s.Name, id: s.FMISID}
	}
	return items
}
