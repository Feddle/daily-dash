package domain

import (
	"time"
)

// NormalizeWeather converts FMI observation data to domain Weather model
func NormalizeWeather(temperature, humidity, windSpeed float64, location, timestamp string) *Weather {
	// Parse timestamp
	parsedTime, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		// If parsing fails, use current time
		parsedTime = time.Now()
	}
	parsedTime = parsedTime.Local()

	// Determine conditions based on weather parameters
	conditions := determineConditions(temperature, humidity, windSpeed)

	return &Weather{
		Location:    location,
		Temperature: temperature,
		Humidity:    humidity,
		WindSpeed:   windSpeed,
		Conditions:  conditions,
		Timestamp:   parsedTime,
	}
}

// determineConditions determines weather conditions based on observations
func determineConditions(temperature, humidity, windSpeed float64) string {
	// Simple heuristics for weather conditions
	if temperature < -10 {
		return "Very Cold"
	} else if temperature < 0 {
		if humidity > 85 {
			return "Snowy"
		}
		return "Cold"
	} else if temperature > 30 {
		return "Very Hot"
	} else if temperature > 25 {
		if humidity < 30 {
			return "Hot & Dry"
		}
		return "Hot"
	}

	// Check precipitation
	if humidity > 90 && windSpeed < 3 {
		return "Foggy"
	} else if humidity > 85 {
		if temperature < 2 {
			return "Snowy"
		}
		return "Rainy"
	} else if humidity > 70 {
		return "Cloudy"
	}

	// Wind conditions
	if windSpeed > 15 {
		return "Windy"
	} else if windSpeed > 10 {
		return "Breezy"
	}

	// Clear/partly cloudy
	if humidity < 50 {
		return "Clear"
	}

	return "Partly Cloudy"
}

// NormalizeTransit converts Föli departure data to domain Transit model
func NormalizeTransit(line string, departureInfos []struct {
	Stop          string
	ScheduledTime string
	ExpectedTime  string
	Status        string
}) *Transit {
	var departures []Departure

	for _, info := range departureInfos {
		// Parse times
		scheduledTime, err1 := time.Parse(time.RFC3339, info.ScheduledTime)
		expectedTime, err2 := time.Parse(time.RFC3339, info.ExpectedTime)

		// If parsing fails, skip this departure
		if err1 != nil {
			continue
		}

		// Use expected time if available, otherwise use scheduled time
		if err2 != nil {
			expectedTime = scheduledTime
		}

		// Calculate minutes until departure
		minutesUntil := int(time.Until(expectedTime).Minutes())
		if minutesUntil < 0 {
			minutesUntil = 0
		}

		// Determine status based on time difference
		status := OnTime
		delay := expectedTime.Sub(scheduledTime)
		if delay > 2*time.Minute {
			status = Delayed
		}

		// Check if explicitly cancelled
		if info.Status == "cancelled" || info.Status == "canceled" {
			status = Cancelled
		}

		departure := Departure{
			Stop:          info.Stop,
			ScheduledTime: scheduledTime,
			ExpectedTime:  expectedTime,
			Status:        status,
			MinutesUntil:  minutesUntil,
		}

		departures = append(departures, departure)
	}

	return &Transit{
		Line:       line,
		Departures: departures,
		Timestamp:  time.Now(),
	}
}

// NormalizeRoadConditions converts Digitraffic road data to domain RoadConditions model
func NormalizeRoadConditions(region string, conditionData []struct {
	Route       string
	Temperature float64
	Condition   string
	Location    string
}) *RoadConditions {
	var segments []RoadSegment

	for _, data := range conditionData {
		// Map condition string to domain enum
		condition := mapStringToRoadCondition(data.Condition, data.Temperature)

		segment := RoadSegment{
			Route:       data.Route,
			Condition:   condition,
			Temperature: data.Temperature,
			Description: data.Condition,
		}

		segments = append(segments, segment)
	}

	return &RoadConditions{
		Region:    region,
		Segments:  segments,
		Timestamp: time.Now(),
	}
}

// mapStringToRoadCondition maps condition string to domain RoadCondition
func mapStringToRoadCondition(conditionStr string, temperature float64) RoadCondition {
	lower := conditionStr

	// Difficult conditions
	if lower == "Icy" || lower == "Slippery" {
		return Difficult
	}

	// Slippery conditions
	if lower == "Frosty" || lower == "Snowy" || lower == "Slushy" || lower == "Wet" {
		return Slippery
	}

	// Temperature-based assessment
	if temperature < 0 && (lower == "Moist" || lower == "Wet") {
		return Slippery
	}

	// Normal conditions
	return Normal
}
