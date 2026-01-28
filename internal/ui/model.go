package ui

import (
	"time"

	"github.com/feddle/daily-dash/internal/config"
	"github.com/feddle/daily-dash/internal/coordinator"
	"github.com/feddle/daily-dash/internal/domain"
	"go.uber.org/zap"
)

// Model represents the Bubble Tea application state
type Model struct {
	coordinator *coordinator.Coordinator
	logger      *zap.Logger
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
	lastUpdate  time.Time
	lastRefresh     time.Time
	showCooldownMsg bool
	quitting        bool
}

// NewModel creates a new UI model
func NewModel(coord *coordinator.Coordinator, logger *zap.Logger, cfg *config.Config) Model {
	return Model{
		coordinator:     coord,
		logger:          logger,
		width:           80,
		height:          24,
		refreshCooldown: cfg.App.RefreshCooldown,
		weatherData:     nil,
		transitData:     nil,
		roadData:        nil,
		weatherLoading:  false,
		transitLoading:  false,
		roadLoading:     false,
		weatherErr:      nil,
		transitErr:      nil,
		roadErr:         nil,
		lastUpdate:      time.Time{},
		lastRefresh:     time.Time{},
		showCooldownMsg: false,
		quitting:        false,
	}
}
