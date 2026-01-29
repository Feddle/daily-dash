package digitraffic

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/feddle/daily-dash/internal/domain"
)

// ParseWeatherStationResponse parses the Digitraffic GeoJSON response
func ParseWeatherStationResponse(jsonData []byte, regionFilter string) ([]RoadConditionData, error) {
	var response WeatherStationResponse
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	var conditions []RoadConditionData

	for _, feature := range response.Features {
		// Filter by region if specified
		// Properties.Name usually contains city e.g. "kt40_Turku_Kärsämäki"
		if regionFilter != "" {
			if !strings.Contains(strings.ToLower(feature.Properties.Name), strings.ToLower(regionFilter)) {
				continue
			}
		}

		// Since we don't have sensor values in this metadata endpoint, we use the station state
		state := feature.Properties.State

		// Map state to condition
		// States: OK, FAULT, DOUBT, CANCELLED, null
		condition := "Normal"
		if state == "OK" {
			condition = "Normal"
		} else if state != "" {
			condition = state // e.g. FAULT
		}

		// Create data entry
		data := RoadConditionData{
			Route:       feature.Properties.Name,
			Temperature: 0, // Not available in metadata
			Condition:   condition,
			Location:    feature.Properties.Name,
			UpdatedTime: feature.Properties.DataUpdatedTime,
		}

		conditions = append(conditions, data)
	}

	return conditions, nil
}

// DetermineRoadCondition maps condition string to domain RoadCondition enum
func DetermineRoadCondition(conditionStr string, temperature float64) domain.RoadCondition {
	lower := strings.ToLower(conditionStr)

	// If using metadata states
	if strings.Contains(lower, "fault") || strings.Contains(lower, "doubt") {
		return domain.Difficult // Or generic warning
	}

	// Normal defaults
	return domain.Normal
}
