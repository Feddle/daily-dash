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

	// Limit variables
	maxDepartures := 5
	var validDepartures []DepartureInfo

	// Calculate destination arrival times and filter if destination stop is provided
	if destStopID != "" {
		dataset, err := c.FetchDatasetInfo(ctx)
		if err != nil {
			c.logger.Error("failed to fetch dataset info", zap.Error(err))
		} else if dataset != nil && dataset.Latest != "" {
			checks := 0
			limitChecks := 15

			for i := range departures {
				if len(validDepartures) >= maxDepartures {
					break
				}
				if checks >= limitChecks {
					break
				}

				if departures[i].TripID == "" {
					continue
				}

				checks++
				stops, err := c.FetchTripStops(ctx, dataset.Latest, departures[i].TripID)

				// Retry if error OR empty result
				if err != nil || len(stops) == 0 {
					// Try stripping prefix from TripRef if it contains "__"
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

				// Find start stop sequence and time
				var startSequence = -1
				var startGTFSTimeStr string
				for _, s := range stops {
					if idsMatch(s.StopID, stopCode) {
						startSequence = s.StopSequence
						startGTFSTimeStr = s.ArrivalTime
						break
					}
				}

				departures[i].DebugInfo = fmt.Sprintf("Stops: %d, StartSeq: %d", len(stops), startSequence)

				tripConnects := false
				for _, s := range stops {
					if idsMatch(s.StopID, destStopID) {
						// Ensure destination is AFTER start stop
						if startSequence != -1 && s.StopSequence <= startSequence {
							continue
						}

						// Found destination stop
						tripConnects = true
						departures[i].DestinationStop = destStopID

						// Parse scheduled times
						var scheduledStart, expectedStart time.Time

						if departures[i].ScheduledTime != "" {
							scheduledStart, _ = time.Parse(time.RFC3339, departures[i].ScheduledTime)
						} else {
							scheduledStart = time.Now()
						}

						if departures[i].ExpectedTime != "" {
							expectedStart, _ = time.Parse(time.RFC3339, departures[i].ExpectedTime)
						} else {
							expectedStart = scheduledStart
						}

						// Calculate travel time based on GTFS difference
						// This avoids issues with service day rollovers (e.g. 25:00:00)
						if startGTFSTimeStr != "" && s.ArrivalTime != "" {
							var h1, m1, s1 int
							var h2, m2, s2 int
							_, err1 := fmt.Sscanf(startGTFSTimeStr, "%d:%d:%d", &h1, &m1, &s1)
							_, err2 := fmt.Sscanf(s.ArrivalTime, "%d:%d:%d", &h2, &m2, &s2)

							if err1 == nil && err2 == nil {
								startDur := time.Duration(h1)*time.Hour + time.Duration(m1)*time.Minute + time.Duration(s1)*time.Second
								destDur := time.Duration(h2)*time.Hour + time.Duration(m2)*time.Minute + time.Duration(s2)*time.Second

								// Calculate relative travel time
								travelTime := destDur - startDur

								// Apply to expected start time
								departures[i].DestinationArrival = expectedStart.Add(travelTime)
							}
						}
						break
					}
				}

				if tripConnects {
					validDepartures = append(validDepartures, departures[i])
				}
			}
		}
		// If dataset fail or other issues, validDepartures might be empty.
		// Use what we found.
		departures = validDepartures

	} else {
		// No destination selection, just take top N
		if len(departures) > maxDepartures {
			departures = departures[:maxDepartures]
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
	url := fmt.Sprintf("%s/gtfs/", c.config.BaseURL)
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
