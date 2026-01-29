package foli

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
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

// FetchTransit fetches transit data for the configured line and optional stop and destination stop
func (c *Client) FetchTransit(ctx context.Context, stopCode, destStopID, line string) ([]DepartureInfo, error) {
	startTime := time.Now()
	c.logger.Info("fetching transit data",
		zap.String("line", line),
	)

	// Föli SIRI Stop Monitoring endpoint
	// Documentation: https://data.foli.fi/
	// Use provided stop, configured stop, or default
	if stopCode == "" {
		stopCode = c.config.Stop
	}
	if stopCode == "" {
		stopCode = "300"
	}
	url := fmt.Sprintf("%s/siri/sm/%s", c.config.BaseURL, stopCode)

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
			err := domain.NewAPIError("Föli", "fetch transit", resp.StatusCode(),
				fmt.Errorf("unexpected status code: %d", resp.StatusCode()))
			if resp.StatusCode() >= 400 && resp.StatusCode() < 500 {
				return backoff.Permanent(err)
			}
			fetchErr = err
			return err
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
	stopName := c.config.StopName
	if stopName == "" {
		stopName = fmt.Sprintf("Stop %s", stopCode)
	}
	departures, err := ParseSIRIResponse(responseData, line, stopName)
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

	// Calculate destination arrival times if destination stop is provided
	if destStopID != "" {
		dataset, err := c.FetchDatasetInfo(ctx)
		if err != nil {
			c.logger.Error("failed to fetch dataset info", zap.Error(err))
		} else if dataset != nil && dataset.Latest != "" {
			for i := range departures {
				if departures[i].TripID == "" {
					continue
				}

				stops, err := c.FetchTripStops(ctx, dataset.Latest, departures[i].TripID)

				// Retry if error OR empty result
				if err != nil || len(stops) == 0 {
					// Try stripping prefix from TripRef if it contains "__"
					// SIRI TripRef format: modification__tripId
					// Sometimes GTFS uses only tripId
					if len(departures[i].TripID) > 5 {
						parts := strings.Split(departures[i].TripID, "__")
						if len(parts) == 2 {
							stopsRetry, errRetry := c.FetchTripStops(ctx, dataset.Latest, parts[1])
							if errRetry == nil && len(stopsRetry) > 0 {
								stops = stopsRetry
								err = nil
							}
						}
					}

					if err != nil {
						c.logger.Debug("failed to fetch trip stops", zap.String("trip", departures[i].TripID), zap.Error(err))
						departures[i].DebugInfo = fmt.Sprintf("Err: %v", err)
						continue
					}
				}

				departures[i].DebugInfo = fmt.Sprintf("Stops: %d", len(stops))

				for _, s := range stops {
					if idsMatch(s.StopID, destStopID) {
						// Found destination stop
						departures[i].DestinationStop = destStopID
						// c.logger.Debug("Destination stop found", zap.String("stop", destStopID), zap.String("arrival", s.ArrivalTime))

						// c.logger.Info("Destination stop found", zap.String("stop", destStopID), zap.String("arrival", s.ArrivalTime))

						// Parse scheduled times
						// GTFS time is HH:MM:SS, potentially > 24:00:00
						// We need to combine it with the valid date of the trip.
						// Simplification: Check real-time delay at start stop and apply to destination scheduled time.

						// 1. Calculate delay at start stop
						var scheduledStart, expectedStart time.Time
						var err1, err2 error

						if departures[i].ScheduledTime != "" {
							scheduledStart, err1 = time.Parse(time.RFC3339, departures[i].ScheduledTime)
						} else {
							// Fallback if missing
							scheduledStart = time.Now()
						}

						if departures[i].ExpectedTime != "" {
							expectedStart, err2 = time.Parse(time.RFC3339, departures[i].ExpectedTime)
						} else {
							expectedStart = scheduledStart
						}

						delay := time.Duration(0)
						if err1 == nil && err2 == nil {
							delay = expectedStart.Sub(scheduledStart)
						} else {
							// If parsing failed, assume no delay? Or use provided delay field from SIRI if available?
							// For now, assume 0.
						}

						// 2. Parse destination scheduled arrival time
						// Format is HH:MM:SS
						if len(s.ArrivalTime) >= 8 {
							// We need the date part. Use scheduledStart date.
							// Note: GTFS times can be > 24h, meaning next day relative to trip start date.

							// Correct approach:
							// Get YYYY-MM-DD from scheduledStart
							// Ensure we have a valid date
							if scheduledStart.IsZero() {
								scheduledStart = time.Now()
							}

							baseDate := time.Date(scheduledStart.Year(), scheduledStart.Month(), scheduledStart.Day(), 0, 0, 0, 0, scheduledStart.Location())

							// Parse duration from string "HH:MM:SS"
							// Since time.ParseDuration doesn't support "HH:MM:SS" directly and GTFS can be "25:00:00", we need custom parsing.
							var h, m, sec int
							fmt.Sscanf(s.ArrivalTime, "%d:%d:%d", &h, &m, &sec)

							destScheduledTime := baseDate.Add(time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(sec)*time.Second)

							// Apply delay
							destExpectedTime := destScheduledTime.Add(delay)

							departures[i].DestinationArrival = destExpectedTime
						}
						break
					}
				}
			}
		}
	}

	c.logger.Info("successfully fetched transit data",
		zap.Int("departure_count", len(departures)),
		zap.Duration("elapsed", time.Since(startTime)),
	)

	return departures, nil
}

// FetchDatasetInfo fetches the current GTFS dataset information
func (c *Client) FetchDatasetInfo(ctx context.Context) (*GTFSDatasetInfo, error) {
	url := fmt.Sprintf("%s/gtfs", c.config.BaseURL)
	var datasetInfo GTFSDatasetInfo

	err := api.RetryWithBackoff(ctx, func() error {
		resp, err := c.httpClient.R().
			SetContext(ctx).
			SetResult(&datasetInfo).
			Get(url)

		if err != nil {
			return err
		}
		if resp.StatusCode() != 200 {
			err := fmt.Errorf("unexpected status code: %d", resp.StatusCode())
			if resp.StatusCode() >= 400 && resp.StatusCode() < 500 {
				return backoff.Permanent(err)
			}
			return err
		}
		return nil
	}, c.logger, "Föli GTFS Info")

	if err != nil {
		return nil, err
	}
	return &datasetInfo, nil
}

// FetchTripStops fetches the stop times for a specific trip from the GTFS API
func (c *Client) FetchTripStops(ctx context.Context, dataset, tripID string) ([]GTFSStopTime, error) {
	// The trip ID in valid SIRI ref often has extra parts, we need to extract the core ID if needed.
	// However, initial testing shows the ID from SIRI `__tripref` might work directly or need slight adjustment.
	// Based on curl test: `00020969__1021041100` worked.

	url := fmt.Sprintf("%s/gtfs/v0/%s/stop_times/trip/%s", c.config.BaseURL, dataset, tripID)
	var stopTimes []GTFSStopTime

	err := api.RetryWithBackoff(ctx, func() error {
		resp, err := c.httpClient.R().
			SetContext(ctx).
			SetResult(&stopTimes).
			Get(url)

		if err != nil {
			return err
		}
		if resp.StatusCode() != 200 {
			err := fmt.Errorf("unexpected status code: %d", resp.StatusCode())
			if resp.StatusCode() >= 400 && resp.StatusCode() < 500 {
				return backoff.Permanent(err)
			}
			return err
		}
		return nil
	}, c.logger, "Föli GTFS Trip")

	if err != nil {
		return nil, err
	}

	// Sort by stop sequence to be safe
	sort.Slice(stopTimes, func(i, j int) bool {
		return stopTimes[i].StopSequence < stopTimes[j].StopSequence
	})

	return stopTimes, nil
}

// idsMatch checks if two stop IDs represent the same stop, handling "T" prefixes and leading zeros
func idsMatch(id1, id2 string) bool {
	if id1 == id2 {
		return true
	}

	// Remove "T" prefix (common in Föli GTFS vs SIRI) and leading zeros
	clean1 := strings.TrimLeft(strings.TrimPrefix(id1, "T"), "0")
	clean2 := strings.TrimLeft(strings.TrimPrefix(id2, "T"), "0")

	return clean1 != "" && clean1 == clean2
}
