package foli

import (
	"context"
	"encoding/json"
	"testing"
	"time"

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
	httpmock.RegisterResponder("GET", "https://data.foli.fi/gtfs/", resp)

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

func TestClient_FetchTransit_TimetableRollover(t *testing.T) {
	logger := zap.NewNop()
	cfg := config.APIEndpointConfig{
		BaseURL: "https://data.foli.fi",
	}
	client := NewClient(cfg, logger)

	// Mock HTTP transport
	httpmock.ActivateNonDefault(client.httpClient.GetClient())
	defer httpmock.DeactivateAndReset()

	// 1. Mock SIRI Response (Stop Monitoring)
	// Stop Code: "100"
	// Time: 2026-01-02 00:30:00 UTC (1767313800)
	aimedTime := int64(1767313800)
	siriResp := SIRIJSONResponse{
		Status: "OK",
		Result: []VehicleDeparture{
			{
				LineRef:            "1",
				DestinationDisplay: "Market Square",
				AimedDepartureTime: aimedTime,
				ExpectedDepartureTime: aimedTime, // On time
				RecordedAtTime:     aimedTime - 60,
				TripRef:            "trip1",
			},
		},
	}
	siriBody, _ := json.Marshal(siriResp)
	httpmock.RegisterResponder("GET", "https://data.foli.fi/siri/sm/100",
		httpmock.NewBytesResponder(200, siriBody))

	// 2. Mock GTFS Dataset Info
	httpmock.RegisterResponder("GET", "https://data.foli.fi/gtfs/",
		httpmock.NewJsonResponderOrPanic(200, map[string]string{"latest": "dataset1"}))

	// 3. Mock GTFS Trip Stops
	// Trip "trip1" has stops: Start (100) at 24:30:00, End (200) at 24:45:00
	// 24:30:00 means 00:30 next day relative to service start.
	tripStops := []GTFSStopTime{
		{StopID: "100", StopSequence: 10, ArrivalTime: "24:30:00"},
		{StopID: "200", StopSequence: 15, ArrivalTime: "24:45:00"},
	}
	httpmock.RegisterResponder("GET", "https://data.foli.fi/gtfs/v0/dataset1/stop_times/trip/trip1",
		httpmock.NewJsonResponderOrPanic(200, tripStops))

	ctx := context.Background()
	// Fetch transit for stop "100" to destination "200"
	departures, err := client.FetchTransit(ctx, "100", "200", "")

	assert.NoError(t, err)
	assert.Len(t, departures, 1)
	if len(departures) > 0 {
		// Expected Arrival: 00:30 + 15m = 00:45 on Jan 2.
		// 2026-01-02 00:45:00 UTC (assuming aimedTime was UTC)
		// We use time.Unix(aimedTime, 0) which is local.
		expectedTime := time.Unix(aimedTime, 0).Add(15 * time.Minute)
		
		actualTime := departures[0].DestinationArrival

		assert.WithinDuration(t, expectedTime, actualTime, 1*time.Second, 
			"Expected destination arrival to be 15m after start. Expected: %v, Got: %v", expectedTime, actualTime)
	}
}
