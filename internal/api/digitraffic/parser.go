package digitraffic

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/feddle/daily-dash/internal/domain"
)

// ParseWeatherStationResponse parses the Digitraffic JSON response
func ParseWeatherStationResponse(jsonData []byte, regionFilter string) ([]RoadConditionData, error) {
	var response WeatherStationResponse
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	var conditions []RoadConditionData

	for _, station := range response.Stations {
		// Filter by region if specified
		if regionFilter != "" {
			if !strings.Contains(strings.ToLower(station.Properties.Name), strings.ToLower(regionFilter)) {
				continue
			}
		}

		// Extract temperature and road condition
		var temperature float64
		var roadCondition string
		hasTemperature := false

		for _, sensor := range station.Properties.SensorValues {
			switch sensor.Name {
			case "ILMA", "AIR_TEMPERATURE_1":
				temperature = sensor.Value
				hasTemperature = true
			case "TIE_1", "ROAD_SURFACE_CONDITION_1":
				// Road surface condition codes
				// 1 = Dry, 2 = Moist, 3 = Wet, 4 = Slippery, etc.
				roadCondition = interpretRoadCondition(sensor.Value)
			}
		}

		// Only add if we have useful data
		if hasTemperature || roadCondition != "" {
			route := "Unknown"
			if station.Properties.RoadAddress != nil {
				route = fmt.Sprintf("Route %d", station.Properties.RoadAddress.Road)
			}

			if roadCondition == "" {
				roadCondition = "Normal"
			}

			data := RoadConditionData{
				Route:       route,
				Temperature: temperature,
				Condition:   roadCondition,
				Location:    station.Properties.Name,
				UpdatedTime: station.Properties.DataUpdatedTime,
			}

			conditions = append(conditions, data)
		}
	}

	return conditions, nil
}

// interpretRoadCondition interprets the road surface condition code
func interpretRoadCondition(code float64) string {
	// Digitraffic road surface condition codes
	switch int(code) {
	case 0:
		return "Unknown"
	case 1:
		return "Dry"
	case 2:
		return "Moist"
	case 3:
		return "Wet"
	case 4:
		return "Slippery"
	case 5:
		return "Frosty"
	case 6:
		return "Snowy"
	case 7:
		return "Icy"
	case 8:
		return "Slushy"
	default:
		return "Normal"
	}
}

// DetermineRoadCondition maps condition string to domain RoadCondition enum
func DetermineRoadCondition(conditionStr string, temperature float64) domain.RoadCondition {
	lower := strings.ToLower(conditionStr)

	// Difficult conditions
	if strings.Contains(lower, "icy") || strings.Contains(lower, "slippery") {
		return domain.Difficult
	}

	// Slippery conditions
	if strings.Contains(lower, "frosty") || strings.Contains(lower, "snowy") ||
		strings.Contains(lower, "slushy") || strings.Contains(lower, "wet") {
		return domain.Slippery
	}

	// Temperature-based assessment
	if temperature < 0 && (strings.Contains(lower, "moist") || strings.Contains(lower, "wet")) {
		return domain.Slippery
	}

	// Normal conditions
	return domain.Normal
}
