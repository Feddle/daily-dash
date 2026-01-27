package cache

import (
	"time"

	"github.com/dgraph-io/ristretto"
	"go.uber.org/zap"
)

// MemoryCache is an in-memory cache implementation using Ristretto
type MemoryCache struct {
	cache  *ristretto.Cache
	logger *zap.Logger
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(maxSizeMB int64, logger *zap.Logger) (Cache, error) {
	// Convert MB to bytes for Ristretto
	maxCost := maxSizeMB * 1024 * 1024

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // Number of keys to track frequency
		MaxCost:     maxCost, // Maximum cache size
		BufferItems: 64,      // Number of keys per Get buffer
		Metrics:     false,   // Disable metrics for simplicity
	})
	if err != nil {
		return nil, err
	}

	return &MemoryCache{
		cache:  cache,
		logger: logger,
	}, nil
}

// Get retrieves a value from the cache
func (m *MemoryCache) Get(key string) (interface{}, bool) {
	value, found := m.cache.Get(key)
	if found {
		m.logger.Debug("cache hit",
			zap.String("key", key),
		)
	} else {
		m.logger.Debug("cache miss",
			zap.String("key", key),
		)
	}
	return value, found
}

// Set stores a value in the cache with a TTL
func (m *MemoryCache) Set(key string, value interface{}, ttl time.Duration) bool {
	// Estimate cost as 1 for simplicity
	// In production, you might want to estimate actual memory size
	cost := int64(1)

	success := m.cache.SetWithTTL(key, value, cost, ttl)
	if success {
		m.logger.Debug("cache set",
			zap.String("key", key),
			zap.Duration("ttl", ttl),
		)
	} else {
		m.logger.Warn("cache set failed",
			zap.String("key", key),
		)
	}

	// Wait for value to be processed
	m.cache.Wait()

	return success
}

// Delete removes a value from the cache
func (m *MemoryCache) Delete(key string) {
	m.cache.Del(key)
	m.logger.Debug("cache delete",
		zap.String("key", key),
	)
}

// Clear removes all values from the cache
func (m *MemoryCache) Clear() {
	m.cache.Clear()
	m.logger.Debug("cache cleared")
}

// Close closes the cache and releases resources
func (m *MemoryCache) Close() {
	m.cache.Close()
	m.logger.Debug("cache closed")
}
