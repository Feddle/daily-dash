package coordinator

import (
	"context"
	"fmt"
	"strings"
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

// FetchTransit fetches transit data with caching for a specific stop (empty for default) and optional destination stop
// stopName and destStopName are optional overrides for display purposes
func (c *Coordinator) FetchTransit(ctx context.Context, stopCode, stopName, destStopCode, destStopName string) (*domain.Transit, error) {
	c.logger.Debug("fetching transit with caching")

	// Check cache first
	cacheKey := fmt.Sprintf("%s:%s:%s", cache.TransitCacheKey, stopCode, destStopCode)
	if c.config.Cache.Enabled {
		if cached, found := c.cache.Get(cacheKey); found {
			c.logger.Debug("returning cached transit data")
			if transit, ok := cached.(*domain.Transit); ok {
				return transit, nil
			}
		}
	}

	// Determine line to filter by
	// If custom stops are used (stopCode provided and different from config), assume we want ALL lines
	// Unless the user explicitly asked to filter? UI doesn't support that yet.
	// For now: if stops are overridden, clear line filter so we can see ANY connection.
	lineFilter := c.config.API.Foli.Line
	if stopCode != "" && stopCode != c.config.API.Foli.Stop {
		lineFilter = ""
	}

	// Cache miss or disabled - fetch from API
	c.logger.Debug("cache miss, fetching from Föli API")
	departures, err := c.foliClient.FetchTransit(ctx, stopCode, destStopCode, lineFilter)
	if err != nil {
		return nil, err
	}

	// Convert to domain format
	var departureData []struct {
		Stop            string
		ScheduledTime   string
		ExpectedTime    string
		Status          string
		DestinationStop string
		ArrivalTime     time.Time
		DebugInfo       string
	}

	for _, dep := range departures {
		currentStop := dep.Stop
		if stopName != "" {
			currentStop = stopName
		}

		departureData = append(departureData, struct {
			Stop            string
			ScheduledTime   string
			ExpectedTime    string
			Status          string
			DestinationStop string
			ArrivalTime     time.Time
			DebugInfo       string
		}{
			Stop:            currentStop,
			ScheduledTime:   dep.ScheduledTime,
			ExpectedTime:    dep.ExpectedTime,
			Status:          dep.Status,
			DestinationStop: dep.DestinationStop,
			ArrivalTime:     dep.DestinationArrival,
			DebugInfo:       dep.DebugInfo,
		})
	}

	// Determine stop name from departures or config
	finalStopName := c.config.API.Foli.StopName
	if stopName != "" {
		finalStopName = stopName
	} else if len(departures) > 0 {
		finalStopName = departures[0].Stop
	}
	if finalStopName == "" {
		finalStopName = stopCode
	}

	// Normalize to domain model
	transit := domain.NormalizeTransit(lineFilter, finalStopName, departureData)

	// Override destination names if provided
	// Override destination names if provided, and ensure column visibility
	if destStopName != "" {
		for i := range transit.Departures {
			// Always set the destination name so the UI shows the columns
			// If the client didn't find a connection, ArrivalTime will be zero
			transit.Departures[i].DestinationStop = destStopName
		}

		// Check if any direct connection was found
		foundConnection := false
		for _, dep := range transit.Departures {
			if !dep.ArrivalTime.IsZero() {
				foundConnection = true
				break
			}
		}

		if !foundConnection && len(transit.Departures) > 0 {
			transit.Warning = fmt.Sprintf("No connection to %s (ID: %s)", destStopName, destStopCode)
		}
	}

	// Store in cache
	if c.config.Cache.Enabled {
		c.cache.Set(cacheKey, transit, c.config.Cache.TTL.Transit)
		c.logger.Debug("transit data cached",
			zap.Duration("ttl", c.config.Cache.TTL.Transit),
		)
	}

	return transit, nil
}

// FetchStops fetches all stops from Föli API
func (c *Coordinator) FetchStops(ctx context.Context) ([]foli.GTFSStop, error) {
	return c.foliClient.FetchStops(ctx)
}

// FetchRoadConditions fetches road conditions with caching for a specific region.
// If region is empty, configuration default is used.
func (c *Coordinator) FetchRoadConditions(ctx context.Context, region string) (*domain.RoadConditions, error) {
	requestedRegion := region
	if requestedRegion == "" {
		requestedRegion = c.config.API.Digitraffic.Region
	}

	c.logger.Debug("fetching road conditions with caching", zap.String("region", requestedRegion))

	allKey := fmt.Sprintf("%s:*", cache.RoadConditionsCacheKey)
	specificKey := fmt.Sprintf("%s:%s", cache.RoadConditionsCacheKey, requestedRegion)

	// Helper to filter segments from a global RoadConditions object
	filterSegments := func(fullData *domain.RoadConditions, filter string) *domain.RoadConditions {
		if filter == "*" || filter == "" {
			return fullData
		}

		var filtered []domain.RoadSegment
		lowerFilter := strings.ToLower(filter)

		for _, seg := range fullData.Segments {
			if strings.Contains(strings.ToLower(seg.Route), lowerFilter) ||
				strings.Contains(strings.ToLower(seg.Description), lowerFilter) {
				filtered = append(filtered, seg)
			}
		}

		return &domain.RoadConditions{
			Region:    filter,
			Segments:  filtered,
			Timestamp: fullData.Timestamp,
		}
	}

	if c.config.Cache.Enabled {
		// 1. Check if we have ALL regions cached
		if cached, found := c.cache.Get(allKey); found {
			if allData, ok := cached.(*domain.RoadConditions); ok {
				c.logger.Debug("returning road conditions from global cache", zap.String("region", requestedRegion))
				return filterSegments(allData, requestedRegion), nil
			}
		}

		// 2. Check if we have this SPECIFIC region cached
		if cached, found := c.cache.Get(specificKey); found {
			if specificData, ok := cached.(*domain.RoadConditions); ok {
				c.logger.Debug("returning road conditions from specific cache", zap.String("region", requestedRegion))
				return specificData, nil
			}
		}
	}

	// Cache miss - fetch EVERYTHING from API to populate the global cache
	c.logger.Debug("cache miss, fetching all road conditions from Digitraffic API")
	conditions, err := c.digitrafficClient.FetchRoadConditions(ctx, "*")
	if err != nil {
		return nil, err
	}

	// Convert to domain format
	var conditionData []struct {
		Route          string
		Temperature    float64
		AirTemperature float64
		Condition      string
		Location       string
	}

	for _, cond := range conditions {
		conditionData = append(conditionData, struct {
			Route          string
			Temperature    float64
			AirTemperature float64
			Condition      string
			Location       string
		}{
			Route:          cond.Route,
			Temperature:    cond.Temperature,
			AirTemperature: cond.AirTemperature,
			Condition:      cond.Condition,
			Location:       cond.Location,
		})
	}

	// Normalize everything
	allConditions := domain.NormalizeRoadConditions("All Finland", conditionData)

	// Store in cache
	if c.config.Cache.Enabled {
		c.cache.Set(allKey, allConditions, c.config.Cache.TTL.Road)
		c.logger.Debug("global road conditions data cached")
	}

	// Filter and return the requested region
	return filterSegments(allConditions, requestedRegion), nil
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
		// Use default configured stop
		transit, err := c.FetchTransit(gCtx, "", "", "", "")
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
		roadConditions, err := c.FetchRoadConditions(gCtx, "")
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
