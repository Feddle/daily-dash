package fmi

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

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
	// Try parsing as URL first
	if u, err := url.Parse(href); err == nil {
		if val := u.Query().Get("param"); val != "" {
			return val
		}
	}

	// Fallback to fragment
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

// ParseStationsResponse parses the FMI stations XML response into WeatherStation slice
func ParseStationsResponse(xmlData []byte) ([]WeatherStation, error) {
	var response StationsWFSResponse
	if err := xml.Unmarshal(xmlData, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	if len(response.Members) == 0 {
		return nil, domain.ErrInvalidResponse
	}

	now := time.Now()
	stations := make([]WeatherStation, 0, len(response.Members))

	for _, member := range response.Members {
		facility := member.Facility

		// Extract FMISID from InspireID
		fmisid := facility.InspireID.Identifier.LocalID
		if fmisid == "" {
			continue // Skip stations without FMISID
		}

		// Parse latitude and longitude from RepresentativePoint.Point.Pos
		lat, lon, err := parsePosition(facility.RepresentativePoint.Point.Pos)
		if err != nil {
			// Log warning but continue - we can still use the station without coordinates
			continue
		}

		// Check if station is active
		active := isStationActive(facility.Period.Period.ActivityTime.TimePeriod, now)
		if !active {
			continue // Skip inactive stations
		}

		station := WeatherStation{
			Name:      facility.Name,
			FMISID:    fmisid,
			Latitude:  lat,
			Longitude: lon,
			Active:    active,
		}

		stations = append(stations, station)
	}

	if len(stations) == 0 {
		return nil, fmt.Errorf("no active stations found in response")
	}

	return stations, nil
}

// parsePosition parses the "lat lon" format from Point.Pos
func parsePosition(pos string) (lat, lon float64, err error) {
	parts := strings.Fields(pos)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid position format: %s", pos)
	}

	lat, err = strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid latitude: %s", parts[0])
	}

	lon, err = strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid longitude: %s", parts[1])
	}

	return lat, lon, nil
}

// isStationActive checks if a station is currently active
func isStationActive(period TimePeriod, now time.Time) bool {
	// If end date is empty, station is active
	if period.End == "" {
		return true
	}

	// Parse end date and check if it's in the future
	endTime, err := time.Parse(time.RFC3339, period.End)
	if err != nil {
		// If we can't parse the end date, assume active
		return true
	}

	return endTime.After(now)
}
