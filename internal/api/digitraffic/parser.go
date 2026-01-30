package digitraffic

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/feddle/daily-dash/internal/domain"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ParseForecastResponse parses the Digitraffic Forecast Response and joins it with Metadata
func ParseForecastResponse(forecastData []byte, metadataData []byte, regionFilter string) ([]RoadConditionData, error) {
	// 1. Parse Metadata
	var metadataResp ForecastMetadataResponse
	if err := json.Unmarshal(metadataData, &metadataResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata JSON: %w", err)
	}

	// Create lookup map: ID -> Description (Name)
	// Example Description: "Vt 1: Helsinki - Turku"
	idToName := make(map[string]string)
	for _, f := range metadataResp.Features {
		idToName[f.Properties.ID] = f.Properties.Description
	}

	// 2. Parse Forecasts
	var forecastResp ForecastResponse
	if err := json.Unmarshal(forecastData, &forecastResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal forecast JSON: %w", err)
	}

	var conditions []RoadConditionData

	for _, section := range forecastResp.ForecastSections {
		// Use ID to look up name
		name, exists := idToName[section.ID]
		if !exists {
			name = section.ID // Fallback
		}

		// Filter by region if specified
		// Filter by region if specified
		if regionFilter != "" && !strings.Contains(strings.ToLower(name), strings.ToLower(regionFilter)) {
			// Debug logging (temporary)
			// fmt.Printf("Skipping %s (filter: %s)\n", name, regionFilter)
			continue
		}

		// Find the most relevant forecast (first OBSERVATION or 0h FORECAST)
		var bestForecast *Forecast
		for i := range section.Forecasts {
			f := &section.Forecasts[i]
			// We prefer OBSERVATION, or the current time forecast
			// The API typically returns them in order.
			// Let's take the first one as it's the most current.
			bestForecast = f
			break
		}

		if bestForecast == nil {
			continue // No data for this section
		}

		// Create data entry
		// Use ConditionToDisplay to get the formatted string
		displayCondition := ConditionToDisplay(bestForecast.OverallRoadCondition, bestForecast.ForecastConditionReason.RoadCondition)

		data := RoadConditionData{
			Route:          name,
			Temperature:    bestForecast.RoadTemperature, // Surface temperature
			AirTemperature: bestForecast.Temperature,     // Air temperature
			Condition:      displayCondition,
			Location:       name,
			UpdatedTime:    bestForecast.Time,
		}

		conditions = append(conditions, data)
	}

	return conditions, nil
}

// ConditionToDisplay maps raw API values to cleaner display strings
func ConditionToDisplay(overall string, reason string) string {
	// overall: NORMAL_CONDITION, SLIPPERY_CONDITION, VERY_SLIPPERY_CONDITION
	// reason: DRY, WET, SNOW, ICE, etc.

	overall = strings.ReplaceAll(overall, "_CONDITION", "")
	caser := cases.Title(language.English)
	overall = caser.String(strings.ToLower(overall))

	if reason != "" {
		reason = caser.String(strings.ToLower(reason))
		return fmt.Sprintf("%s (%s)", overall, reason)
	}
	return overall
}

// DetermineRoadCondition maps condition string to domain RoadCondition enum
// This is used by the Normalizer later.
func DetermineRoadCondition(conditionStr string) domain.RoadCondition {
	lower := strings.ToLower(conditionStr)

	if strings.Contains(lower, "very slippery") || strings.Contains(lower, "difficult") {
		return domain.Difficult
	}
	if strings.Contains(lower, "slippery") || strings.Contains(lower, "ice") || strings.Contains(lower, "snow") {
		return domain.Slippery
	}

	if strings.Contains(lower, "unknown") {
		return domain.Unknown
	}

	return domain.Normal
}
