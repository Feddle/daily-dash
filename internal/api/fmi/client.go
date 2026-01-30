package fmi

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/feddle/daily-dash/internal/api"
	"github.com/feddle/daily-dash/internal/config"
	"github.com/feddle/daily-dash/internal/domain"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// Client represents an FMI API client
type Client struct {
	httpClient *resty.Client
	config     config.APIEndpointConfig
	mu         sync.RWMutex
	logger     *zap.Logger
}

// NewClient creates a new FMI API client
func NewClient(cfg config.APIEndpointConfig, logger *zap.Logger) *Client {
	return &Client{
		httpClient: api.NewHTTPClient(cfg.Timeout, cfg.RetryAttempts, logger),
		config:     cfg,
		logger:     logger,
	}
}

// SetLocation updates the weather location
func (c *Client) SetLocation(location string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.Location = location
}

// FetchWeather fetches current weather data for the configured location
func (c *Client) FetchWeather(ctx context.Context) (*ObservationData, error) {
	startTime := time.Now()

	c.mu.RLock()
	location := c.config.Location
	baseURL := c.config.BaseURL
	c.mu.RUnlock()

	c.logger.Info("fetching weather data",
		zap.String("location", location),
	)

	// Build FMI WFS query
	// FMI API documentation: https://www.ilmatieteenlaitos.fi/tallennetut-kyselyt
	queryParams := map[string]string{
		"service":        "WFS",
		"version":        "2.0.0",
		"request":        "getFeature",
		"storedquery_id": "fmi::observations::weather::timevaluepair",
		"parameters":     "t2m,rh,ws_10min",
	}

	// Use fmisid if location is numeric, otherwise use place
	isNumeric := true
	if location == "" {
		isNumeric = false
	} else {
		for _, r := range location {
			if r < '0' || r > '9' {
				isNumeric = false
				break
			}
		}
	}

	if isNumeric {
		queryParams["fmisid"] = location
	} else {
		queryParams["place"] = location
	}

	var responseData []byte
	var fetchErr error

	// Execute request with retry logic
	err := api.RetryWithBackoff(ctx, func() error {
		resp, err := c.httpClient.R().
			SetContext(ctx).
			SetQueryParams(queryParams).
			Get(baseURL)

		if err != nil {
			fetchErr = err
			return err
		}

		if resp.StatusCode() != 200 {
			fetchErr = domain.NewAPIError("FMI", "fetch weather", resp.StatusCode(),
				fmt.Errorf("unexpected status code: %d", resp.StatusCode()))
			return fetchErr
		}

		responseData = resp.Body()
		return nil
	}, c.logger, "FMI")

	if err != nil {
		c.logger.Error("failed to fetch weather data",
			zap.Error(err),
			zap.Duration("elapsed", time.Since(startTime)),
		)
		if fetchErr != nil {
			return nil, fetchErr
		}
		return nil, domain.NewAPIError("FMI", "fetch weather", 0, err)
	}

	// Parse the response
	observations, timestamp, err := ParseWeatherResponse(responseData)
	if err != nil {
		c.logger.Error("failed to parse weather response",
			zap.Error(err),
		)
		return nil, domain.NewAPIError("FMI", "parse weather", 0, err)
	}

	data := ExtractWeatherData(observations, timestamp)

	c.logger.Info("successfully fetched weather data",
		zap.Float64("temperature", data.Temperature),
		zap.Duration("elapsed", time.Since(startTime)),
	)

	return data, nil
}

// FetchStations fetches the complete list of weather stations from FMI API
func (c *Client) FetchStations(ctx context.Context) ([]WeatherStation, error) {
	startTime := time.Now()

	c.mu.RLock()
	baseURL := c.config.BaseURL
	c.mu.RUnlock()

	c.logger.Info("fetching weather stations from FMI")

	// Build FMI WFS query for stations
	// Using fmi::ef::stations stored query with networkid 121 (FMI observation network)
	queryParams := map[string]string{
		"service":        "WFS",
		"version":        "2.0.0",
		"request":        "getFeature",
		"storedquery_id": "fmi::ef::stations",
		"networkid":      "121", // FMI observation network - filters out marine/aviation stations
	}

	var responseData []byte
	var fetchErr error

	// Execute request with retry logic
	err := api.RetryWithBackoff(ctx, func() error {
		resp, err := c.httpClient.R().
			SetContext(ctx).
			SetQueryParams(queryParams).
			Get(baseURL)

		if err != nil {
			fetchErr = err
			return err
		}

		if resp.StatusCode() != 200 {
			fetchErr = domain.NewAPIError("FMI", "fetch stations", resp.StatusCode(),
				fmt.Errorf("unexpected status code: %d", resp.StatusCode()))
			return fetchErr
		}

		responseData = resp.Body()
		return nil
	}, c.logger, "FMI")

	if err != nil {
		c.logger.Error("failed to fetch weather stations",
			zap.Error(err),
			zap.Duration("elapsed", time.Since(startTime)),
		)
		if fetchErr != nil {
			return nil, fetchErr
		}
		return nil, domain.NewAPIError("FMI", "fetch stations", 0, err)
	}

	// Parse the response
	stations, err := ParseStationsResponse(responseData)
	if err != nil {
		c.logger.Error("failed to parse stations response",
			zap.Error(err),
		)
		return nil, domain.NewAPIError("FMI", "parse stations", 0, err)
	}

	c.logger.Info("successfully fetched weather stations",
		zap.Int("station_count", len(stations)),
		zap.Duration("elapsed", time.Since(startTime)),
	)

	return stations, nil
}
