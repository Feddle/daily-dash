package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Load reads configuration from file and environment variables
func Load() (*Config, error) {
	v := viper.New()

	// Set config name and paths
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Look for config in multiple locations
	v.AddConfigPath("./configs")
	v.AddConfigPath(".")

	// Also check user config directory
	if home, err := os.UserHomeDir(); err == nil {
		v.AddConfigPath(filepath.Join(home, ".config", "daily-dash"))
	}

	// Enable environment variable support
	v.SetEnvPrefix("DAILY_DASH")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Set defaults
	setDefaults(v)

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found; use defaults
	}

	// Unmarshal config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate config
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default values for configuration
func setDefaults(v *viper.Viper) {
	// App defaults
	v.SetDefault("app.refresh_interval", "5m")
	v.SetDefault("app.timeout", "30s")

	// FMI API defaults
	v.SetDefault("api.fmi.base_url", "https://opendata.fmi.fi/wfs")
	v.SetDefault("api.fmi.location", "Turku")
	v.SetDefault("api.fmi.timeout", "15s")
	v.SetDefault("api.fmi.retry_attempts", 3)

	// Föli API defaults
	v.SetDefault("api.foli.base_url", "https://data.foli.fi")
	v.SetDefault("api.foli.line", "1")
	v.SetDefault("api.foli.timeout", "10s")
	v.SetDefault("api.foli.retry_attempts", 3)

	// Digitraffic API defaults
	v.SetDefault("api.digitraffic.base_url", "https://tie.digitraffic.fi/api/v1")
	v.SetDefault("api.digitraffic.region", "Turku")
	v.SetDefault("api.digitraffic.timeout", "15s")
	v.SetDefault("api.digitraffic.retry_attempts", 3)

	// Cache defaults
	v.SetDefault("cache.enabled", true)
	v.SetDefault("cache.max_size_mb", 10)
	v.SetDefault("cache.ttl.weather", "10m")
	v.SetDefault("cache.ttl.transit", "2m")
	v.SetDefault("cache.ttl.road", "15m")
	v.SetDefault("cache.ttl.stations", "24h")

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "console")
	v.SetDefault("logging.output", "stdout")

	// UI defaults
	v.SetDefault("ui.refresh_key", "r")
	v.SetDefault("ui.quit_key", "q")
}

// validate validates the configuration
func validate(cfg *Config) error {
	if cfg.App.Timeout <= 0 {
		return fmt.Errorf("app.timeout must be positive")
	}

	if cfg.API.FMI.BaseURL == "" {
		return fmt.Errorf("api.fmi.base_url is required")
	}

	if cfg.API.Foli.BaseURL == "" {
		return fmt.Errorf("api.foli.base_url is required")
	}

	if cfg.API.Digitraffic.BaseURL == "" {
		return fmt.Errorf("api.digitraffic.base_url is required")
	}

	if cfg.Cache.MaxSizeMB < 0 {
		return fmt.Errorf("cache.max_size_mb must be non-negative")
	}

	return nil
}
