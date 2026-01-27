# Daily-Dash

A TUI (Terminal User Interface) dashboard for Turku, Finland, displaying live weather, public transit schedules, and road conditions.

## Features

- **Weather Data**: Real-time weather information from Finnish Meteorological Institute (FMI)
- **Transit Schedule**: Live Föli Line 1 departure times with delay information
- **Road Conditions**: Current road conditions in the Turku region from Digitraffic
- **Manual Refresh**: Press 'r' to update all data concurrently
- **Graceful Error Handling**: Partial failures don't crash the app
- **In-Memory Caching**: Reduces API load with configurable TTLs

## Installation

### Prerequisites

- Go 1.23 or later
- Internet connection (required for live data)

### Install from Source

```bash
git clone https://github.com/feddle/daily-dash.git
cd daily-dash
make build
```

Or install directly:

```bash
go install github.com/feddle/daily-dash/cmd/daily-dash@latest
```

## Usage

Run the application:

```bash
./bin/daily-dash
```

Or with make:

```bash
make run
```

### Keyboard Controls

- `r` - Refresh all data
- `q` or `Ctrl+C` - Quit the application

## Configuration

The application can be configured using a YAML file or environment variables.

### Config File Locations

Daily-Dash looks for `config.yaml` in the following locations (in order):

1. `./configs/config.yaml` (project directory)
2. `./config.yaml` (current directory)
3. `~/.config/daily-dash/config.yaml` (user config directory)

### Configuration Options

```yaml
app:
  refresh_interval: 5m  # Future auto-refresh feature
  timeout: 30s

api:
  fmi:
    base_url: "https://opendata.fmi.fi/wfs"
    location: "Turku"
    timeout: 15s
    retry_attempts: 3

  foli:
    base_url: "https://data.foli.fi"
    line: "1"
    timeout: 10s
    retry_attempts: 3

  digitraffic:
    base_url: "https://tie.digitraffic.fi/api/v1"
    region: "Turku"
    timeout: 15s
    retry_attempts: 3

cache:
  enabled: true
  max_size_mb: 10
  ttl:
    weather: 10m
    transit: 2m
    road: 15m

logging:
  level: "info"  # debug, info, warn, error
  format: "console"  # console, json
  output: "stdout"  # stdout, stderr, or file path
```

### Environment Variables

All configuration options can be set via environment variables with the `DAILY_DASH_` prefix:

```bash
export DAILY_DASH_API_FMI_LOCATION=Helsinki
export DAILY_DASH_LOGGING_LEVEL=debug
export DAILY_DASH_CACHE_ENABLED=false
```

Environment variables use underscores instead of dots: `api.fmi.location` becomes `DAILY_DASH_API_FMI_LOCATION`.

## Development

### Project Structure

```
daily-dash/
├── cmd/daily-dash/       # Application entry point
├── internal/
│   ├── api/             # API clients (FMI, Föli, Digitraffic)
│   ├── cache/           # Caching layer
│   ├── config/          # Configuration loading
│   ├── coordinator/     # Concurrent data fetching
│   ├── domain/          # Business logic and models
│   ├── logger/          # Logging setup
│   └── ui/              # Bubble Tea TUI
├── configs/             # Default configuration
├── Makefile             # Build automation
└── README.md            # This file
```

### Building

```bash
make build
```

### Running Tests

```bash
make test
```

With coverage report:

```bash
make test-coverage
```

### Code Formatting

```bash
make fmt
```

### Linting

```bash
make lint
```

## Data Sources

- **Weather**: [Finnish Meteorological Institute (FMI)](https://www.ilmatieteenlaitos.fi/) - Open Data (CC BY 4.0)
- **Transit**: [Föli Public Transport](https://www.foli.fi/) - Real-time SIRI/GTFS data
- **Road Conditions**: [Digitraffic](https://www.digitraffic.fi/) - Finnish Transport Infrastructure Agency

## Technology Stack

- **TUI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Styling**: [Lipgloss](https://github.com/charmbracelet/lipgloss)
- **HTTP Client**: [Resty](https://github.com/go-resty/resty)
- **Caching**: [Ristretto](https://github.com/dgraph-io/ristretto)
- **Configuration**: [Viper](https://github.com/spf13/viper)
- **Logging**: [Zap](https://github.com/uber-go/zap)
- **Retry Logic**: [Backoff](https://github.com/cenkalti/backoff)

## Roadmap

- [x] Phase 1: Project scaffold
- [ ] Phase 2: Weather panel
- [ ] Phase 3: Transit panel
- [ ] Phase 4: Road conditions panel
- [ ] Phase 5: Concurrent fetching with error handling
- [ ] Phase 6: Polish and finalize

Future enhancements:
- Auto-refresh at configurable intervals
- Multiple transit lines
- Multi-day weather forecast
- Severe weather alerts
- Desktop notifications

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## Credits

Built with [Claude Code](https://claude.ai/claude-code)
