package foli

import (
	"context"
	"testing"

	"github.com/feddle/daily-dash/internal/config"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestClient_FetchDatasetInfo(t *testing.T) {
	logger := zap.NewNop()
	cfg := config.APIEndpointConfig{
		BaseURL: "https://data.foli.fi",
	}
	client := NewClient(cfg, logger)

	// Mock HTTP transport
	httpmock.ActivateNonDefault(client.httpClient.GetClient())
	defer httpmock.DeactivateAndReset()

	// Mock response
	resp, _ := httpmock.NewJsonResponder(200, map[string]string{"latest": "20260128-124320"})
	httpmock.RegisterResponder("GET", "https://data.foli.fi/gtfs", resp)

	ctx := context.Background()
	info, err := client.FetchDatasetInfo(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "20260128-124320", info.Latest)
}

func TestClient_FetchTripStops(t *testing.T) {
	logger := zap.NewNop()
	cfg := config.APIEndpointConfig{
		BaseURL: "https://data.foli.fi",
	}
	client := NewClient(cfg, logger)

	// Mock HTTP transport
	httpmock.ActivateNonDefault(client.httpClient.GetClient())
	defer httpmock.DeactivateAndReset()

	// Mock response
	tripID := "00020969__1021041100"
	dataset := "20260128-124320"
	url := "https://data.foli.fi/gtfs/v0/" + dataset + "/stop_times/trip/" + tripID
	
	tripResp, _ := httpmock.NewJsonResponder(200, []map[string]interface{}{
		{"arrival_time": "15:05:00", "departure_time": "15:05:00", "stop_id": "1943", "stop_sequence": 0},
		{"arrival_time": "15:06:10", "departure_time": "15:06:10", "stop_id": "472", "stop_sequence": 1},
	})
	httpmock.RegisterResponder("GET", url, tripResp)

	ctx := context.Background()
	stops, err := client.FetchTripStops(ctx, dataset, tripID)

	require.NoError(t, err)
	require.Len(t, stops, 2)
	assert.Equal(t, "1943", stops[0].StopID)
	assert.Equal(t, "15:05:00", stops[0].ArrivalTime)
}
