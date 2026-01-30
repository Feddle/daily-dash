package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/feddle/daily-dash/internal/api/fmi"
	"github.com/feddle/daily-dash/internal/api/foli"
	"github.com/feddle/daily-dash/internal/config"
	"github.com/feddle/daily-dash/internal/coordinator"
	"github.com/feddle/daily-dash/internal/domain"
	"go.uber.org/zap"
)

// Model represents the Bubble Tea application state
type Model struct {
	coordinator     *coordinator.Coordinator
	logger          *zap.Logger
	width           int
	height          int
	refreshCooldown time.Duration

	// Data
	weatherData *domain.Weather
	transitData *domain.Transit
	roadData    *domain.RoadConditions

	// Loading states
	weatherLoading bool
	transitLoading bool
	roadLoading    bool

	// Errors
	weatherErr error
	transitErr error
	roadErr    error

	// General state
	lastUpdate      time.Time
	lastRefresh     time.Time
	showCooldownMsg bool
	quitting        bool

	// Föli Stop Selection
	selectingStart    bool
	selectingEnd      bool
	loadingStops      bool
	stops             []foli.GTFSStop
	stopList          list.Model
	foliStartStop     string
	foliStartStopName string
	foliEndStop       string
	foliEndStopName   string

	// Road Region Selection
	selectingRoad       bool
	loadingRoadStations bool
	roadStations        []string
	roadStationList     list.Model
	selectedRoadRegion  string

	// Weather Location Selection
	selectingLocation bool
	loadingStations   bool
	weatherStations   []fmi.WeatherStation
	weatherList       list.Model
}

// NewModel creates a new UI model
func NewModel(coord *coordinator.Coordinator, logger *zap.Logger, cfg *config.Config) Model {
	return Model{
		coordinator:         coord,
		logger:              logger,
		width:               80,
		height:              24,
		refreshCooldown:     cfg.App.RefreshCooldown,
		weatherData:         nil,
		transitData:         nil,
		roadData:            nil,
		weatherLoading:      false,
		transitLoading:      false,
		roadLoading:         false,
		weatherErr:          nil,
		transitErr:          nil,
		roadErr:             nil,
		lastUpdate:          time.Time{},
		lastRefresh:         time.Time{},
		showCooldownMsg:     false,
		quitting:            false,
		selectingStart:      false,
		selectingEnd:        false,
		loadingStops:        false,
		stops:               nil,
		stopList:            list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
		foliStartStop:       "", // Will use default if empty
		foliStartStopName:   "",
		foliEndStop:         "",
		foliEndStopName:     "",
		selectingRoad:       false,
		loadingRoadStations: false,
		roadStations:        nil,
		roadStationList:     list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
		selectedRoadRegion:  "", // Will use config default if empty
		selectingLocation:   false,
		loadingStations:     false,
		weatherStations:     nil,
		weatherList:         list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
	}
}
