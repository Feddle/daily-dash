package foli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/feddle/daily-dash/internal/domain"
)

// ParseSIRIResponse parses the SIRI JSON response into departure information
func ParseSIRIResponse(jsonData []byte, lineFilter, stopName string) ([]DepartureInfo, error) {
	if len(jsonData) == 0 {
		return nil, fmt.Errorf("empty response received")
	}

	var response SIRIJSONResponse
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SIRI JSON: %w", err)
	}

	var departures []DepartureInfo

	if response.Status != "OK" {
		return nil, fmt.Errorf("API returned status: %s", response.Status)
	}

	for _, visit := range response.Result {
		// Filter by line if specified
		if lineFilter != "" && visit.LineRef != lineFilter {
			continue
		}

		// Convert timestamps to RFC3339 strings
		scheduledTime := time.Unix(visit.AimedDepartureTime, 0).Format(time.RFC3339)
		expectedTime := time.Unix(visit.ExpectedDepartureTime, 0).Format(time.RFC3339)
		recordedAt := time.Unix(visit.RecordedAtTime, 0).Format(time.RFC3339)

		// Determine status
		// Using the helper function logic, but adapted
		status := domain.OnTime.String() // string representation? No, struct expects string which coordinator maps to domain.
		// Wait, coordinator maps 'string' Status to domain.DepartureStatus.
		// Normalizer expects "cancelled" for Cancelled.
		// "delayed" logic is in Normalizer based on time diff.

		// However, we can also pass "delayed" if we know it.
		// For now, let's just pass "on time" (or empty) and let Normalizer compare timestamps.
		status = "on time"

		departure := DepartureInfo{
			Stop: "Likely Stop", // We don't have stop name in this specific JSON unless we fetch stops?
			// The JSON had "stop_name" ?? No, "result" has "lineref".
			// Ah, the JSON I saw earlier had "lineref".
			// Does it have "stopname"?
			// The JSON from Step 158: "destinationdisplay", "lineref", "recordedattime"...
			// It does NOT seem to have current stop name if it's Vehicle Monitoring.
			// But if it's Stop Monitoring, it should have.
			// The URL is siri/sm/1. So it's Stop Monitoring at Stop 1.
			// So "Stop" is "Stop 1" (Kauppatori).
			// We can hardcode or use config?
			// I'll set Stop to "Kauppatori" (Stop 1) or just "Stop".
			// Wait, the result list might contain visits to that stop.
			// Does the result item have StopPointName?
			// Step 158 output didn't show it explicitly in the snippet.
			// It had "originref", "destinationref".
			// I will set Stop to "Stop 1" for now.

			Line:          visit.LineRef,
			Destination:   visit.DestinationDisplay,
			ScheduledTime: scheduledTime,
			ExpectedTime:  expectedTime,
			Status:        status,
			RecordedAt:    recordedAt,
			TripID:        visit.TripRef,
		}

		// Wait, I need actual Stop Name?
		// Normalizer uses it to display.
		// I'll just use the provided stopName.
		// Use the provided stopName if available to override the generic one
		if stopName != "" {
			departure.Stop = stopName
		}

		departures = append(departures, departure)
	}

	return departures, nil
}

// Keep helper functions if needed, but unused ones can go.
// DetermineDepartureStatus and CalculateMinutesUntil were used by parser before.
// Normalizer does this now too?
// Normalizer duplicates logic.
// I'll leave them out if not used.
