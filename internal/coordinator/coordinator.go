package coordinator

import (
	"context"
	"sync"
	"time"

	"github.com/feddle/daily-dash/internal/api/digitraffic"
	"github.com/feddle/daily-dash/internal/api/fmi"
	"github.com/feddle/daily-dash/internal/api/foli"
	"github.com/feddle/daily-dash/internal/cache"
	"github.com/feddle/daily-dash/internal/config"
	"github.com/feddle/daily-dash/internal/domain"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Coordinator orchestrates data fetching from multiple APIs
type Coordinator struct {
	fmiClient         *fmi.Client
	foliClient        *foli.Client
	digitrafficClient *digitraffic.Client
	cache             cache.Cache
	config            *config.Config
	logger            *zap.Logger
}

// New creates a new Coordinator
func New(cfg *config.Config, logger *zap.Logger, cache cache.Cache) *Coordinator {
	return &Coordinator{
		fmiClient:         fmi.NewClient(cfg.API.FMI, logger),
		foliClient:        foli.NewClient(cfg.API.Foli, logger),
		digitrafficClient: digitraffic.NewClient(cfg.API.Digitraffic, logger),
		cache:             cache,
		config:            cfg,
		logger:            logger,
	}
}

// FetchWeather fetches weather data with caching
func (c *Coordinator) FetchWeather(ctx context.Context) (*domain.Weather, error) {
	c.logger.Debug("fetching weather with caching")

	// Check cache first
	if c.config.Cache.Enabled {
		if cached, found := c.cache.Get(cache.WeatherCacheKey); found {
			c.logger.Debug("returning cached weather data")
			if weather, ok := cached.(*domain.Weather); ok {
				return weather, nil
			}
		}
	}

	// Cache miss or disabled - fetch from API
	c.logger.Debug("cache miss, fetching from FMI API")
	data, err := c.fmiClient.FetchWeather(ctx)
	if err != nil {
		return nil, err
	}

	// Normalize to domain model
	weather := domain.NormalizeWeather(
		data.Temperature,
		data.Humidity,
		data.WindSpeed,
		c.config.API.FMI.Location,
		data.Time,
	)

	// Store in cache
	if c.config.Cache.Enabled {
		c.cache.Set(cache.WeatherCacheKey, weather, c.config.Cache.TTL.Weather)
		c.logger.Debug("weather data cached",
			zap.Duration("ttl", c.config.Cache.TTL.Weather),
		)
	}

	return weather, nil
}

// FetchTransit fetches transit data with caching
func (c *Coordinator) FetchTransit(ctx context.Context) (*domain.Transit, error) {
	c.logger.Debug("fetching transit with caching")

	// Check cache first
	if c.config.Cache.Enabled {
		if cached, found := c.cache.Get(cache.TransitCacheKey); found {
			c.logger.Debug("returning cached transit data")
			if transit, ok := cached.(*domain.Transit); ok {
				return transit, nil
			}
		}
	}

	// Cache miss or disabled - fetch from API
	c.logger.Debug("cache miss, fetching from Föli API")
	departures, err := c.foliClient.FetchTransit(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to domain format
	var departureData []struct {
		Stop          string
		ScheduledTime string
		ExpectedTime  string
		Status        string
	}

	for _, dep := range departures {
		departureData = append(departureData, struct {
			Stop          string
			ScheduledTime string
			ExpectedTime  string
			Status        string
		}{
			Stop:          dep.Stop,
			ScheduledTime: dep.ScheduledTime,
			ExpectedTime:  dep.ExpectedTime,
			Status:        dep.Status,
		})
	}

	// Normalize to domain model
	transit := domain.NormalizeTransit(c.config.API.Foli.Line, departureData)

	// Store in cache
	if c.config.Cache.Enabled {
		c.cache.Set(cache.TransitCacheKey, transit, c.config.Cache.TTL.Transit)
		c.logger.Debug("transit data cached",
			zap.Duration("ttl", c.config.Cache.TTL.Transit),
		)
	}

	return transit, nil
}

// FetchRoadConditions fetches road conditions with caching
func (c *Coordinator) FetchRoadConditions(ctx context.Context) (*domain.RoadConditions, error) {
	c.logger.Debug("fetching road conditions with caching")

	// Check cache first
	if c.config.Cache.Enabled {
		if cached, found := c.cache.Get(cache.RoadConditionsCacheKey); found {
			c.logger.Debug("returning cached road conditions data")
			if roadConditions, ok := cached.(*domain.RoadConditions); ok {
				return roadConditions, nil
			}
		}
	}

	// Cache miss or disabled - fetch from API
	c.logger.Debug("cache miss, fetching from Digitraffic API")
	conditions, err := c.digitrafficClient.FetchRoadConditions(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to domain format
	var conditionData []struct {
		Route       string
		Temperature float64
		Condition   string
		Location    string
	}

	for _, cond := range conditions {
		conditionData = append(conditionData, struct {
			Route       string
			Temperature float64
			Condition   string
			Location    string
		}{
			Route:       cond.Route,
			Temperature: cond.Temperature,
			Condition:   cond.Condition,
			Location:    cond.Location,
		})
	}

	// Normalize to domain model
	roadConditions := domain.NormalizeRoadConditions(c.config.API.Digitraffic.Region, conditionData)

	// Store in cache
	if c.config.Cache.Enabled {
		c.cache.Set(cache.RoadConditionsCacheKey, roadConditions, c.config.Cache.TTL.Road)
		c.logger.Debug("road conditions data cached",
			zap.Duration("ttl", c.config.Cache.TTL.Road),
		)
	}

	return roadConditions, nil
}

// FetchAll fetches all data concurrently using fan-out/fan-in pattern
func (c *Coordinator) FetchAll(ctx context.Context) (*domain.DashboardData, error) {
	c.logger.Info("fetching all dashboard data concurrently")
	startTime := time.Now()

	// Create data structure with mutex for thread-safe access
	data := &domain.DashboardData{
		Timestamp: time.Now(),
	}
	var mu sync.Mutex

	// Use errgroup for concurrent fetching with context cancellation
	g, gCtx := errgroup.WithContext(ctx)

	// Fetch weather concurrently
	g.Go(func() error {
		fetchStart := time.Now()
		weather, err := c.FetchWeather(gCtx)
		fetchDuration := time.Since(fetchStart)

		mu.Lock()
		defer mu.Unlock()

		if err != nil {
			c.logger.Error("failed to fetch weather",
				zap.Error(err),
				zap.Duration("duration", fetchDuration),
			)
			// Don't return error, allow partial success
			return nil
		}

		data.Weather = weather
		c.logger.Debug("weather fetch completed",
			zap.Duration("duration", fetchDuration),
		)
		return nil
	})

	// Fetch transit concurrently
	g.Go(func() error {
		fetchStart := time.Now()
		transit, err := c.FetchTransit(gCtx)
		fetchDuration := time.Since(fetchStart)

		mu.Lock()
		defer mu.Unlock()

		if err != nil {
			c.logger.Error("failed to fetch transit",
				zap.Error(err),
				zap.Duration("duration", fetchDuration),
			)
			// Don't return error, allow partial success
			return nil
		}

		data.Transit = transit
		c.logger.Debug("transit fetch completed",
			zap.Duration("duration", fetchDuration),
		)
		return nil
	})

	// Fetch road conditions concurrently
	g.Go(func() error {
		fetchStart := time.Now()
		roadConditions, err := c.FetchRoadConditions(gCtx)
		fetchDuration := time.Since(fetchStart)

		mu.Lock()
		defer mu.Unlock()

		if err != nil {
			c.logger.Error("failed to fetch road conditions",
				zap.Error(err),
				zap.Duration("duration", fetchDuration),
			)
			// Don't return error, allow partial success
			return nil
		}

		data.RoadConditions = roadConditions
		c.logger.Debug("road conditions fetch completed",
			zap.Duration("duration", fetchDuration),
		)
		return nil
	})

	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil {
		c.logger.Error("error during concurrent fetch", zap.Error(err))
		// Even if there's an error, we still return partial data
	}

	totalDuration := time.Since(startTime)
	successCount := 0
	if data.Weather != nil {
		successCount++
	}
	if data.Transit != nil {
		successCount++
	}
	if data.RoadConditions != nil {
		successCount++
	}

	c.logger.Info("dashboard data fetch complete",
		zap.Duration("total_duration", totalDuration),
		zap.Int("successful_fetches", successCount),
		zap.Int("total_fetches", 3),
	)

	return data, nil
}
