package foli

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/feddle/daily-dash/internal/api"
	"github.com/feddle/daily-dash/internal/domain"
	"go.uber.org/zap"
)

// FetchStops fetches all stops from the Föli GTFS API
func (c *Client) FetchStops(ctx context.Context) ([]GTFSStop, error) {
	startTime := time.Now()
	c.logger.Info("fetching all stops")

	url := fmt.Sprintf("%s/gtfs/stops", c.config.BaseURL)

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
			fetchErr = domain.NewAPIError("Föli", "fetch stops", resp.StatusCode(),
				fmt.Errorf("unexpected status code: %d", resp.StatusCode()))
			return fetchErr
		}

		body := resp.Body()

		// Check if this is metadata with a "go" link
		var meta struct {
			Go string `json:"go"`
		}
		if err := json.Unmarshal(body, &meta); err == nil && meta.Go != "" {
			// Found a link, fetch the actual data
			targetURL := meta.Go
			if len(targetURL) > 2 && targetURL[:2] == "//" {
				targetURL = "http:" + targetURL
			}
			c.logger.Debug("following metadata link", zap.String("url", targetURL))
			
			resp2, err2 := c.httpClient.R().
				SetContext(ctx).
				SetHeader("Accept", "application/json").
				SetHeader("Accept-Encoding", "gzip").
				Get(targetURL)
			
			if err2 == nil && resp2.StatusCode() == 200 {
				responseData = resp2.Body()
				return nil
			} else if err2 != nil {
				fetchErr = err2
				return err2
			} else {
				fetchErr = fmt.Errorf("metadata link failed: %d", resp2.StatusCode())
				return fetchErr
			}
		}

		responseData = body
		return nil
	}, c.logger, "Föli")

	if err != nil {
		c.logger.Error("failed to fetch stops",
			zap.Error(err),
			zap.Duration("elapsed", time.Since(startTime)),
		)
		if fetchErr != nil {
			return nil, fetchErr
		}
		return nil, domain.NewAPIError("Föli", "fetch stops", 0, err)
	}

	// Parse map of stops
	var stopsMap map[string]GTFSStop
	if err := json.Unmarshal(responseData, &stopsMap); err != nil {
		snippet := string(responseData)
		if len(snippet) > 100 {
			snippet = snippet[:100]
		}
		c.logger.Error("failed to parse stops response", zap.Error(err), zap.String("snippet", snippet))
		return nil, fmt.Errorf("failed to parse stops: %w", err)
	}

	var stops []GTFSStop
	for id, stop := range stopsMap {
		if stop.ID == "" {
			stop.ID = id
		}
		if stop.Code == "" {
			stop.Code = stop.ID
		}
		stops = append(stops, stop)
	}

	// Sort by name
	sort.Slice(stops, func(i, j int) bool {
		return stops[i].Name < stops[j].Name
	})

	c.logger.Info("successfully fetched stops",
		zap.Int("count", len(stops)),
		zap.Duration("elapsed", time.Since(startTime)),
	)

	return stops, nil
}
