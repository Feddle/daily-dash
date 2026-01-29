package ui

import (
	"time"

	"github.com/feddle/daily-dash/internal/api/foli"
	"github.com/feddle/daily-dash/internal/domain"
)

// Message types for Bubble Tea

// Weather messages
type weatherFetchStartMsg struct{}

type weatherFetchSuccessMsg struct {
	weather *domain.Weather
}

type weatherFetchErrorMsg struct {
	err error
}

// Transit messages
type transitFetchStartMsg struct{}

type transitFetchSuccessMsg struct {
	transit *domain.Transit
}

type transitFetchErrorMsg struct {
	err error
}

// Road conditions messages
type roadFetchStartMsg struct{}

type roadFetchSuccessMsg struct {
	roadConditions *domain.RoadConditions
}

type roadFetchErrorMsg struct {
	err error
}

// Stops messages

type stopsFetchSuccessMsg struct {
	stops []foli.GTFSStop
}

type stopsFetchErrorMsg struct {
	err error
}

// General messages
type clearCooldownMsg struct{}

type tickMsg struct {
	time time.Time
}
