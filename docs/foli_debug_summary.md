# Föli Dashboard Debugging Summary

## Issue
Users are experiencing a "No connection found" warning when selecting valid start and end stops (e.g., **Satamakatu** -> **Kauppahalli**) that are known to be connected by Föli lines (e.g., Line 1). The arrival time column remains empty/dash ("-").

## findings

### 1. Data Mismatch
There appears to be a disconnect between the real-time SIRI data (Vehicle Monitoring/Stop Monitoring) and the static GTFS schedule data.
- **Trip IDs**: SIRI returns IDs like `modification__tripId` (e.g., `00020969__1021041100`), while GTFS often uses just `tripId` (`1021041100`).
- **Stop IDs**: Föli stop codes (e.g., "14") are sometimes represented differently in GTFS stop lists (e.g., "T14", "0014").

### 2. Implemented Fixes
We have implemented several robustness improvements to address these data issues:
- **Loose ID Matching**: The matching logic now ignores "T" prefixes and leading zeros when comparing the selected End Stop ID with the stops listed in the trip's schedule.
- **Trip ID Fallback**: If fetching the trip schedule fails with the raw SIRI ID, the system now automatically attempts to strip the prefix (splitting by `__`) and retries fetching the schedule.
- **Line Filter Disabling**: When custom stops are selected, the dashboard automatically disables the default "Line 1" filter, allowing it to find connections on *any* line that serves the start stop.
- **Stop List Details**: The stop selection UI now displays Stop IDs explicitly to help users verify they are selecting the correct stop (and direction).

### 3. Current Status
Despite these fixes, the specific combination of Satamakatu -> Kauppahalli is still reporting "No connection".
- The debug info indicates that we are successfully fetching trip schedules (`Stops: N`), but the destination Stop ID is simply not found in the list of stops for that trip.
- This strongly suggests a **directionality issue**: The user might be selecting a "Start Stop" that serves the *opposite* direction, or the specific bus trip captured by real-time data has already passed the destination or is on a variant route that skips it.

## Future Recommendations
1.  **Direction Filtering**: Update the UI to explicitly show the direction of stops (e.g., "Towards Market Square" vs "Towards Port") to prevent selecting the wrong side of the street.
2.  **Stop Sequence Checking**: Implement logic to verify that the Destination Stop comes *after* the Start Stop in the trip sequence (currently we just check for existence in the trip).
3.  **Expanded Debugging**: Add a dedicated debug view or log dump to inspect the full list of stops returned for a trip, allowing us to see exactly what the API is returning vs what we are looking for.
