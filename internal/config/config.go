package config

import "time"

// Config represents the application configuration
type Config struct {
	App     AppConfig     `mapstructure:"app"`
	API     APIConfig     `mapstructure:"api"`
	Cache   CacheConfig   `mapstructure:"cache"`
	Logging LoggingConfig `mapstructure:"logging"`
	UI      UIConfig      `mapstructure:"ui"`
}

// AppConfig contains general application settings
type AppConfig struct {
	RefreshInterval time.Duration `mapstructure:"refresh_interval"`
	RefreshCooldown time.Duration `mapstructure:"refresh_cooldown"`
	Timeout         time.Duration `mapstructure:"timeout"`
}

// APIConfig contains settings for all external APIs
type APIConfig struct {
	FMI         APIEndpointConfig `mapstructure:"fmi"`
	Foli        APIEndpointConfig `mapstructure:"foli"`
	Digitraffic APIEndpointConfig `mapstructure:"digitraffic"`
}

// APIEndpointConfig contains configuration for a single API endpoint
type APIEndpointConfig struct {
	BaseURL       string        `mapstructure:"base_url"`
	Location      string        `mapstructure:"location,omitempty"`
	Line          string        `mapstructure:"line,omitempty"`
	Region        string        `mapstructure:"region,omitempty"`
	Timeout       time.Duration `mapstructure:"timeout"`
	RetryAttempts int           `mapstructure:"retry_attempts"`
}

// CacheConfig contains cache settings
type CacheConfig struct {
	Enabled   bool           `mapstructure:"enabled"`
	MaxSizeMB int64          `mapstructure:"max_size_mb"`
	TTL       CacheTTLConfig `mapstructure:"ttl"`
}

// CacheTTLConfig contains TTL settings for different data types
type CacheTTLConfig struct {
	Weather time.Duration `mapstructure:"weather"`
	Transit time.Duration `mapstructure:"transit"`
	Road    time.Duration `mapstructure:"road"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// UIConfig contains UI-related settings
type UIConfig struct {
	RefreshKey string `mapstructure:"refresh_key"`
	QuitKey    string `mapstructure:"quit_key"`
}
