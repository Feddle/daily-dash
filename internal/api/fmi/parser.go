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

	observations := make(map[string]float64)
	var lastTime string

	for _, member := range response.Members {
		points := member.PointTimeSeriesObservation.Result.MeasurementTimeseries.Points
		if len(points) == 0 {
			continue
		}

		// Get the most recent point
		latestPoint := points[len(points)-1]
		// Update lastTime to the most recent timestamp found
		if latestPoint.TVP.Time > lastTime {
			lastTime = latestPoint.TVP.Time
		}
		value := latestPoint.TVP.Value

		// Identify parameter from ObservedProperty
		// Href example: http://xml.fmi.fi/schema/wfs/2.0/Query/StoredQuery/fmi::observations::weather::timevaluepair#t2m
		href := member.PointTimeSeriesObservation.ObservedProperty.Href
		param := parseParamFromHref(href)
		
		if param != "" {
			observations[param] = value
		}
	}

	return observations, lastTime, nil
}

func parseParamFromHref(href string) string {
	parts := strings.Split(href, "#")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

// ExtractWeatherData extracts specific weather parameters from observations
func ExtractWeatherData(observations map[string]float64, timestamp string) *ObservationData {
	data := &ObservationData{
		Time: timestamp,
	}

	// Map parameters based on FMI parameter names
	// t2m: Temperature (deg C)
	// rh: Relative Humidity (%)
	// ws_10min: Wind Speed (m/s)
	
	if temp, ok := observations["t2m"]; ok {
		data.Temperature = temp
	}
	if humidity, ok := observations["rh"]; ok {
		data.Humidity = humidity
	}
	if windSpeed, ok := observations["ws_10min"]; ok {
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


