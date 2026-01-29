package domain

import (
	"testing"
	"time"
)

func TestNormalizeWeather_TimezoneConversion(t *testing.T) {
	// 12:00 UTC
	utcTime := "2024-01-28T12:00:00Z"

	weather := NormalizeWeather(0, 0, 0, "Test", utcTime)

	// Check if the location is Local
	if weather.Timestamp.Location() != time.Local {
		t.Errorf("Expected timestamp location to be Local, got %v", weather.Timestamp.Location())
	}

	// Verify the instant is the same
	parsedUTC, _ := time.Parse(time.RFC3339, utcTime)
	if !weather.Timestamp.Equal(parsedUTC) {
		t.Errorf("Expected timestamp to represent the same instant as %v, got %v", parsedUTC, weather.Timestamp)
	}
}
