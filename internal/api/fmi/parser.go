package fmi

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/feddle/daily-dash/internal/domain"
)

// ParseWeatherResponse parses the FMI XML response into ObservationData
func ParseWeatherResponse(xmlData []byte) (map[string]float64, string, error) {
	var response WFSResponse
	if err := xml.Unmarshal(xmlData, &response); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	if len(response.Members) == 0 {
		return nil, "", domain.ErrInvalidResponse
	}

	// Extract observations: FMI returns multiple members, one for each parameter
	// We need to collect temperature, humidity, wind speed, etc.
	observations := make(map[string]float64)
	var lastTime string

	for _, member := range response.Members {
		points := member.PointTimeSeriesObservation.Result.MeasurementTimeseries.Points
		if len(points) == 0 {
			continue
		}

		// Get the most recent point
		latestPoint := points[len(points)-1]
		lastTime = latestPoint.TVP.Time
		value := latestPoint.TVP.Value

		// FMI returns separate members for each parameter
		// We'll use a simple heuristic: collect all values
		// In a real implementation, you'd parse the parameter name from the observation metadata

		// For now, we'll assume the order: temperature, humidity, wind speed
		// This is a simplified parser - a production version would parse parameter names
		observations[fmt.Sprintf("param_%d", len(observations))] = value
	}

	return observations, lastTime, nil
}

// ExtractWeatherData extracts specific weather parameters from observations
func ExtractWeatherData(observations map[string]float64, timestamp string) *ObservationData {
	// This is a simplified extraction
	// In production, you'd map parameter IDs to specific measurements
	data := &ObservationData{
		Time: timestamp,
	}

	// Map parameters (this is simplified - real FMI data has parameter IDs)
	if temp, ok := observations["param_0"]; ok {
		data.Temperature = temp
	}
	if humidity, ok := observations["param_1"]; ok {
		data.Humidity = humidity
	}
	if windSpeed, ok := observations["param_2"]; ok {
		data.WindSpeed = windSpeed
	}

	return data
}

// DetermineConditions determines weather conditions based on observations
func DetermineConditions(temperature, humidity float64) string {
	// Simple heuristic for conditions
	if temperature < 0 {
		if humidity > 85 {
			return "Snowy"
		}
		return "Cold"
	} else if temperature > 25 {
		if humidity < 30 {
			return "Hot & Dry"
		}
		return "Hot"
	} else if humidity > 85 {
		return "Rainy"
	} else if humidity > 70 {
		return "Cloudy"
	}

	// Check for moderate conditions
	sunnyHumidity := 50.0
	if humidity < sunnyHumidity {
		return "Clear"
	}

	return "Partly Cloudy"
}

// ParseTime parses FMI timestamp format
func ParseTime(timeStr string) string {
	// FMI uses ISO 8601 format
	// Example: 2024-01-27T15:00:00Z
	parts := strings.Split(timeStr, "T")
	if len(parts) >= 2 {
		return strings.TrimSuffix(parts[1], "Z")
	}
	return timeStr
}
