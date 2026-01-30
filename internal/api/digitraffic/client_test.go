package digitraffic

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/feddle/daily-dash/internal/config"
	"github.com/feddle/daily-dash/internal/domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestClient_FetchRoadConditions(t *testing.T) {
	// Mock server returning Metadata
	mockMetadata := `{
		"features": [
			{
				"properties": {
					"id": "1",
					"description": "Vt 1: Helsinki - Turku"
				}
			},
			{
				"properties": {
					"id": "2",
					"description": "Vt 4: Helsinki - Oulu"
				}
			}
		]
	}`

	// Mock server returning Forecasts
	mockForecasts := `{
		"dataUpdatedTime": "2023-10-27T10:00:00Z",
		"forecastSections": [
			{
				"id": "1",
				"forecasts": [
					{
						"time": "2023-10-27T10:00:00Z",
						"type": "OBSERVATION",
						"overallRoadCondition": "NORMAL_CONDITION",
						"roadTemperature": 5.5,
						"forecastConditionReason": {
							"roadCondition": "DRY"
						}
					}
				]
			},
			{
				"id": "2",
				"forecasts": [
					{
						"time": "2023-10-27T10:00:00Z",
						"type": "FORECAST",
						"overallRoadCondition": "SLIPPERY_CONDITION",
						"roadTemperature": -1.2,
						"forecastConditionReason": {
							"roadCondition": "SNOW"
						}
					}
				]
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/weather/v1/forecast-sections-simple" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintln(w, mockMetadata)
			return
		}
		if r.URL.Path == "/api/weather/v1/forecast-sections-simple/forecasts" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintln(w, mockForecasts)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	logger := zap.NewNop()
	cfg := config.APIEndpointConfig{
		BaseURL:       server.URL,
		Timeout:       5 * time.Second,
		RetryAttempts: 1,
	}

	client := NewClient(cfg, logger)

	// Test case 1: Fetch all (empty region)
	conditions, err := client.FetchRoadConditions(context.Background(), "")
	assert.NoError(t, err)
	assert.Len(t, conditions, 2)

	// Verify mappings
	// Order is not guaranteed in map iteration? Actually slice append order depends on range over forecastSections.
	// Since json parsing order is usually preserved for arrays, let's check content.

	var c1, c2 RoadConditionData
	for _, c := range conditions {
		switch c.Route {
		case "Vt 1: Helsinki - Turku":
			c1 = c
		case "Vt 4: Helsinki - Oulu":
			c2 = c
		}
	}

	assert.Equal(t, "Vt 1: Helsinki - Turku", c1.Route)
	assert.Equal(t, "Normal (Dry)", c1.Condition)
	assert.Equal(t, 5.5, c1.Temperature)

	assert.Equal(t, "Vt 4: Helsinki - Oulu", c2.Route)
	assert.Equal(t, "Slippery (Snow)", c2.Condition)
	assert.Equal(t, -1.2, c2.Temperature)

	// Test case 2: Filtering
	conditions, err = client.FetchRoadConditions(context.Background(), "Oulu")
	assert.NoError(t, err)
	assert.Len(t, conditions, 1)
	assert.Equal(t, "Vt 4: Helsinki - Oulu", conditions[0].Route)
}

func TestDetermineRoadCondition(t *testing.T) {
	tests := []struct {
		input    string
		expected domain.RoadCondition
	}{
		{"Normal (Dry)", domain.Normal},
		{"Slippery (Snow)", domain.Slippery},
		{"Very Slippery (Ice)", domain.Difficult},
		{"Unknown", domain.Unknown},
		{"Normal Condition", domain.Normal},
		{"Normal Condition (Wet)", domain.Normal},
	}

	for _, tt := range tests {
		result := DetermineRoadCondition(tt.input)
		assert.Equal(t, tt.expected, result, "Input: %s", tt.input)
	}
}
