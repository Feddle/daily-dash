package main

import (
	"fmt"
	"os"

	"github.com/feddle/daily-dash/internal/cache"
	"github.com/feddle/daily-dash/internal/config"
	"github.com/feddle/daily-dash/internal/coordinator"
	"github.com/feddle/daily-dash/internal/logger"
	"github.com/feddle/daily-dash/internal/ui"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Force log output to file to avoid interfering with TUI
	if cfg.Logging.Output == "stdout" || cfg.Logging.Output == "stderr" {
		cfg.Logging.Output = "daily-dash.log"
	}

	// Initialize logger
	log, err := logger.New(cfg.Logging.Level, cfg.Logging.Format, cfg.Logging.Output)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer func() { _ = log.Sync() }()

	log.Info("starting daily-dash",
		zap.String("location", cfg.API.FMI.Location),
		zap.String("transit_line", cfg.API.Foli.Line),
	)

	// Initialize cache
	var dataCache cache.Cache
	if cfg.Cache.Enabled {
		dataCache, err = cache.NewMemoryCache(cfg.Cache.MaxSizeMB, log)
		if err != nil {
			return fmt.Errorf("failed to initialize cache: %w", err)
		}
		defer dataCache.Close()
		log.Info("cache initialized",
			zap.Int64("max_size_mb", cfg.Cache.MaxSizeMB),
		)
	} else {
		log.Info("cache disabled")
	}

	// Initialize coordinator
	coord := coordinator.New(cfg, log, dataCache)

	// Run the TUI application
	if err := ui.Run(coord, log, cfg); err != nil {
		return fmt.Errorf("failed to run UI: %w", err)
	}

	log.Info("daily-dash stopped")
	return nil
}
