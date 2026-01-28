package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/feddle/daily-dash/internal/config"
	"github.com/feddle/daily-dash/internal/coordinator"
	"go.uber.org/zap"
)

// Init initializes the Bubble Tea model
func (m Model) Init() tea.Cmd {
	// Trigger initial data fetch
	m.weatherLoading = true
	m.transitLoading = true
	m.roadLoading = true
	return m.fetchAllCmd()
}

// Run starts the Bubble Tea application
func Run(coord *coordinator.Coordinator, logger *zap.Logger, cfg *config.Config) error {
	model := NewModel(coord, logger, cfg)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
