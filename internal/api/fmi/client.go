package fmi

import (
	"context"
	"fmt"
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

// FetchWeather fetches current weather data for the configured location
func (c *Client) FetchWeather(ctx context.Context) (*ObservationData, error) {
	startTime := time.Now()
	c.logger.Info("fetching weather data",
		zap.String("location", c.config.Location),
	)

	// Build FMI WFS query
	// FMI API documentation: https://www.ilmatieteenlaitos.fi/tallennetut-kyselyt
	queryParams := map[string]string{
		"service":        "WFS",
		"version":        "2.0.0",
		"request":        "getFeature",
		"storedquery_id": "fmi::observations::weather::timevaluepair",
		"place":          c.config.Location,
		"parameters":     "t2m,rh,ws_10min",
	}

	var responseData []byte
	var fetchErr error

	// Execute request with retry logic
	err := api.RetryWithBackoff(ctx, func() error {
		resp, err := c.httpClient.R().
			SetContext(ctx).
			SetQueryParams(queryParams).
			Get(c.config.BaseURL)

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
