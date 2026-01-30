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

// FetchRoadConditions fetches road conditions for a specific region.
// If region is empty, it uses the configured default region.
// If region is "*", it fetches all regions (no filter).
// FetchRoadConditions fetches road conditions for a specific region using Forecast Sections API.
// If region is empty, it uses the configured default region.
// If region is "*", it fetches all regions (no filter).
func (c *Client) FetchRoadConditions(ctx context.Context, region string) ([]RoadConditionData, error) {
	startTime := time.Now()

	// Handle region defaulting
	filterRegion := region
	switch region {
	case "":
		filterRegion = c.config.Region
	case "*":
		filterRegion = "" // No filter
	}

	c.logger.Info("fetching road conditions (forecast sections)",
		zap.String("region", filterRegion),
	)

	// Step 1: Fetch Metadata (Road names)
	// Cache-busting or long-term caching could be done, but for now we fetch it.
	// In a real production app, cache this for 24h.
	metadata, err := c.fetchForecastSectionsMetadata(ctx)
	if err != nil {
		c.logger.Error("failed to fetch forecast metadata", zap.Error(err))
		return nil, err
	}

	// Step 2: Fetch Forecasts (Conditions)
	forecasts, err := c.fetchForecastSectionsForecasts(ctx)
	if err != nil {
		c.logger.Error("failed to fetch forecast data", zap.Error(err))
		return nil, err
	}

	// Step 3: Parse and Combine
	conditions, err := ParseForecastResponse(forecasts, metadata, filterRegion)
	if err != nil {
		c.logger.Error("failed to parse road conditions response", zap.Error(err))
		return nil, domain.NewAPIError("Digitraffic", "parse road conditions", 0, err)
	}

	c.logger.Info("successfully fetched road conditions",
		zap.Int("segment_count", len(conditions)),
		zap.Duration("elapsed", time.Since(startTime)),
	)

	return conditions, nil
}

func (c *Client) fetchForecastSectionsMetadata(ctx context.Context) ([]byte, error) {
	// https://tie.digitraffic.fi/api/weather/v1/forecast-sections-simple
	// Use BaseURL from config (default: https://tie.digitraffic.fi)
	url := fmt.Sprintf("%s/api/weather/v1/forecast-sections-simple", c.config.BaseURL)
	return c.fetchWithRetry(ctx, url)
}

func (c *Client) fetchForecastSectionsForecasts(ctx context.Context) ([]byte, error) {
	// https://tie.digitraffic.fi/api/weather/v1/forecast-sections-simple/forecasts
	url := fmt.Sprintf("%s/api/weather/v1/forecast-sections-simple/forecasts", c.config.BaseURL)
	return c.fetchWithRetry(ctx, url)
}

func (c *Client) fetchWithRetry(ctx context.Context, url string) ([]byte, error) {
	var responseData []byte
	var fetchErr error

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
			fetchErr = domain.NewAPIError("Digitraffic", "fetch url", resp.StatusCode(),
				fmt.Errorf("unexpected status code: %d from %s", resp.StatusCode(), url))
			return fetchErr
		}

		responseData = resp.Body()
		return nil
	}, c.logger, "Digitraffic")

	if err != nil {
		if fetchErr != nil {
			return nil, fetchErr
		}
		return nil, domain.NewAPIError("Digitraffic", "fetch retry failed", 0, err)
	}

	return responseData, nil
}
