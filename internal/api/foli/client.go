package foli

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

// Client represents a Föli API client
type Client struct {
	httpClient *resty.Client
	config     config.APIEndpointConfig
	logger     *zap.Logger
}

// NewClient creates a new Föli API client
func NewClient(cfg config.APIEndpointConfig, logger *zap.Logger) *Client {
	return &Client{
		httpClient: api.NewHTTPClient(cfg.Timeout, cfg.RetryAttempts, logger),
		config:     cfg,
		logger:     logger,
	}
}

// FetchTransit fetches transit data for the configured line
func (c *Client) FetchTransit(ctx context.Context) ([]DepartureInfo, error) {
	startTime := time.Now()
	c.logger.Info("fetching transit data",
		zap.String("line", c.config.Line),
	)

	// Föli SIRI Stop Monitoring endpoint
	// Documentation: https://data.foli.fi/
	// We default to stop 300 (Puutori) if no stop is configured, as siri/sm requires a stop code
	url := fmt.Sprintf("%s/siri/sm/300", c.config.BaseURL)

	var responseData []byte
	var fetchErr error

	// Execute request with retry logic
	err := api.RetryWithBackoff(ctx, func() error {
		resp, err := c.httpClient.R().
			SetContext(ctx).
			SetHeader("Accept", "application/json").
			SetHeader("Accept-Encoding", "gzip").
			Get(url)

		if err != nil {
			fetchErr = err
			return err
		}

		if resp.StatusCode() != 200 {
			fetchErr = domain.NewAPIError("Föli", "fetch transit", resp.StatusCode(),
				fmt.Errorf("unexpected status code: %d", resp.StatusCode()))
			return fetchErr
		}

		responseData = resp.Body()
		return nil
	}, c.logger, "Föli")

	if err != nil {
		c.logger.Error("failed to fetch transit data",
			zap.Error(err),
			zap.Duration("elapsed", time.Since(startTime)),
		)
		if fetchErr != nil {
			return nil, fetchErr
		}
		return nil, domain.NewAPIError("Föli", "fetch transit", 0, err)
	}

	// Parse the response
	departures, err := ParseSIRIResponse(responseData, c.config.Line)
	if err != nil {
		c.logger.Error("failed to parse transit response",
			zap.Error(err),
		)
		return nil, domain.NewAPIError("Föli", "parse transit", 0, err)
	}

	// Limit to next 5 departures
	maxDepartures := 5
	if len(departures) > maxDepartures {
		departures = departures[:maxDepartures]
	}

	c.logger.Info("successfully fetched transit data",
		zap.Int("departure_count", len(departures)),
		zap.Duration("elapsed", time.Since(startTime)),
	)

	return departures, nil
}
