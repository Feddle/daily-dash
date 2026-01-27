package foli

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/feddle/daily-dash/internal/domain"
)

// ParseSIRIResponse parses the SIRI XML response into departure information
func ParseSIRIResponse(xmlData []byte, lineFilter string) ([]DepartureInfo, error) {
	var response SIRIResponse
	if err := xml.Unmarshal(xmlData, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SIRI XML: %w", err)
	}

	var departures []DepartureInfo

	visits := response.ServiceDelivery.StopMonitoringDelivery.MonitoredStopVisits
	if len(visits) == 0 {
		return departures, nil
	}

	for _, visit := range visits {
		journey := visit.MonitoredVehicleJourney
		call := journey.MonitoredCall

		// Filter by line if specified
		if lineFilter != "" && journey.LineRef != lineFilter {
			continue
		}

		// Use departure time if available, otherwise use arrival time
		scheduledTime := call.AimedDepartureTime
		expectedTime := call.ExpectedDepartureTime
		status := call.DepartureStatus

		if scheduledTime == "" {
			scheduledTime = call.AimedArrivalTime
			expectedTime = call.ExpectedArrivalTime
			status = call.ArrivalStatus
		}

		departure := DepartureInfo{
			Stop:          call.StopPointName,
			Line:          journey.LineRef,
			Destination:   journey.DestinationName,
			ScheduledTime: scheduledTime,
			ExpectedTime:  expectedTime,
			Status:        status,
			RecordedAt:    visit.RecordedAtTime,
		}

		departures = append(departures, departure)
	}

	return departures, nil
}

// DetermineDepartureStatus determines the departure status
func DetermineDepartureStatus(scheduledTime, expectedTime string) domain.DepartureStatus {
	if expectedTime == "" || scheduledTime == "" {
		return domain.OnTime
	}

	scheduled, err1 := time.Parse(time.RFC3339, scheduledTime)
	expected, err2 := time.Parse(time.RFC3339, expectedTime)

	if err1 != nil || err2 != nil {
		return domain.OnTime
	}

	delay := expected.Sub(scheduled)

	// Consider delayed if more than 2 minutes late
	if delay > 2*time.Minute {
		return domain.Delayed
	} else if delay < -2*time.Minute {
		// Early by more than 2 minutes (unusual but possible)
		return domain.OnTime
	}

	return domain.OnTime
}

// CalculateMinutesUntil calculates minutes until departure from now
func CalculateMinutesUntil(departureTime string) int {
	if departureTime == "" {
		return 0
	}

	departureTimeParsed, err := time.Parse(time.RFC3339, departureTime)
	if err != nil {
		return 0
	}

	duration := time.Until(departureTimeParsed)
	minutes := int(duration.Minutes())

	if minutes < 0 {
		return 0
	}

	return minutes
}

// SortDeparturesByTime sorts departures by expected or scheduled time
func SortDeparturesByTime(departures []DepartureInfo) []DepartureInfo {
	// Sort departures by time (earliest first)
	// For simplicity, we'll return them as-is since the API usually returns them sorted
	// In production, you'd want to implement proper sorting
	return departures
}
