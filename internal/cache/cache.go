package cache

import "time"

// Cache defines the interface for caching operations
type Cache interface {
	// Get retrieves a value from the cache
	Get(key string) (interface{}, bool)

	// Set stores a value in the cache with a TTL
	Set(key string, value interface{}, ttl time.Duration) bool

	// Delete removes a value from the cache
	Delete(key string)

	// Clear removes all values from the cache
	Clear()

	// Close closes the cache and releases resources
	Close()
}

// CacheKey constants for different data types
const (
	WeatherCacheKey         = "weather"
	TransitCacheKey         = "transit"
	RoadConditionsCacheKey  = "road_conditions"
	WeatherStationsCacheKey = "weather_stations"
)
