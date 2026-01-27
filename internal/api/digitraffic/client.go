package digitraffic

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

// Client represents a Digitraffic API client
type Client struct {
	httpClient *resty.Client
	config     config.APIEndpointConfig
	logger     *zap.Logger
}

// NewClient creates a new Digitraffic API client
func NewClient(cfg config.APIEndpointConfig, logger *zap.Logger) *Client {
	return &Client{
		httpClient: api.NewHTTPClient(cfg.Timeout, cfg.RetryAttempts, logger),
		config:     cfg,
		logger:     logger,
	}
}

// FetchRoadConditions fetches road conditions for the configured region
func (c *Client) FetchRoadConditions(ctx context.Context) ([]RoadConditionData, error) {
	startTime := time.Now()
	c.logger.Info("fetching road conditions",
		zap.String("region", c.config.Region),
	)

	// Digitraffic weather station API
	// Documentation: https://www.digitraffic.fi/en/road-traffic/
	url := fmt.Sprintf("%s/weathercam-stations/data", c.config.BaseURL)

	// Alternatively, use weather data API
	// More reliable for road conditions
	url = "https://tie.digitraffic.fi/api/weathercam/v1/stations"

	var responseData []byte
	var fetchErr error

	// Execute request with retry logic
	err := api.RetryWithBackoff(ctx, func() error {
		resp, err := c.httpClient.R().
			SetContext(ctx).
			SetHeader("Accept", "application/json").
			Get(url)

		if err != nil {
			fetchErr = err
			return err
		}

		if resp.StatusCode() != 200 {
			fetchErr = domain.NewAPIError("Digitraffic", "fetch road conditions", resp.StatusCode(),
				fmt.Errorf("unexpected status code: %d", resp.StatusCode()))
			return fetchErr
		}

		responseData = resp.Body()
		return nil
	}, c.logger, "Digitraffic")

	if err != nil {
		c.logger.Error("failed to fetch road conditions",
			zap.Error(err),
			zap.Duration("elapsed", time.Since(startTime)),
		)
		if fetchErr != nil {
			return nil, fetchErr
		}
		return nil, domain.NewAPIError("Digitraffic", "fetch road conditions", 0, err)
	}

	// Parse the response
	conditions, err := ParseWeatherStationResponse(responseData, c.config.Region)
	if err != nil {
		c.logger.Error("failed to parse road conditions response",
			zap.Error(err),
		)
		return nil, domain.NewAPIError("Digitraffic", "parse road conditions", 0, err)
	}

	// Limit to top 3 road segments
	maxSegments := 3
	if len(conditions) > maxSegments {
		conditions = conditions[:maxSegments]
	}

	c.logger.Info("successfully fetched road conditions",
		zap.Int("segment_count", len(conditions)),
		zap.Duration("elapsed", time.Since(startTime)),
	)

	return conditions, nil
}
