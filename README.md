# hyperbyte-plot

A modular Go CLI tool that visualizes metrics from various data sources with live-updating ASCII graphs in a terminal UI.

## Features

- 📊 Live ASCII graphs of metrics from various data sources
- 🔄 Real-time updates (configurable interval, default 5s)
- 📱 Multi-panel TUI with keyboard navigation
- 📝 YAML configuration for queries
- ⚡ Fast and lightweight terminal interface
- 🏗️ Modular architecture for easy extension to new data sources
- 🔌 Pluggable backend system (supports Prometheus, InfluxDB v2, and InfluxDB v1)

## Installation

```bash
# Clone and build
git clone <your-repo-url>
cd hyperbyte-hyperbyte-plot
go build -o hyperbyte-plot main.go
```

## Usage

```bash
# Run with default config file (queries.yaml)
./hyperbyte-plot

# Run with custom config file
./hyperbyte-plot --config /path/to/config.yaml
```

## Configuration

hyperbyte-plot supports multiple backend data sources through YAML configuration.

### Prometheus Configuration

```yaml
# backend: prometheus  # Optional, defaults to prometheus
prometheus:
  url: "http://localhost:9090"

queries:
  - name: CPU Usage
    expr: rate(node_cpu_seconds_total{mode="user"}[5m])
  - name: Memory Usage
    expr: node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes
  - name: Disk IO
    expr: rate(node_disk_reads_completed_total[5m])
```

### InfluxDB Configuration

```yaml
backend: influxdb
influxdb:
  url: "http://localhost:8086"
  token: "your-influxdb-token"
  org: "your-organization"
  bucket: "metrics"

queries:
  - name: CPU Usage
    expr: 'r._measurement == "cpu" and r._field == "usage_percent"'
  - name: Memory Usage
    expr: 'r._measurement == "mem" and r._field == "used_percent"'
  - name: Network IO
    expr: 'r._measurement == "net" and r._field == "bytes_recv"'
```

### Advanced InfluxDB Queries

For complex Flux queries, you can use multi-line expressions:

```yaml
backend: influxdb
influxdb:
  url: "http://localhost:8086"
  token: "your-token"
  org: "your-org"
  bucket: "telegraf"

queries:
  - name: Avg CPU
    expr: |
      from(bucket: "telegraf")
      |> range(start: -5m)
      |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_active")
      |> mean()
```

### InfluxDB v1 Configuration  

For InfluxDB v1 with InfluxQL queries:

```yaml
backend: influxdb1
influxdb1:
  url: "http://localhost:8086"
  username: "admin"
  password: "password"
  database: "telegraf"

queries:
  - name: CPU Usage
    expr: 'SELECT mean("usage_idle") FROM "cpu" WHERE time >= now() - 5m'
  - name: Memory Usage
    expr: 'SELECT mean("used_percent") FROM "mem" WHERE time >= now() - 5m'
  - name: Network IO
    expr: 'SELECT derivative(mean("bytes_recv"), 1s) FROM "net" WHERE time >= now() - 5m GROUP BY time(30s) ORDER BY time DESC LIMIT 1'
```

## Keyboard Controls

- `q` or `Q` - Quit the application
- `Tab` / `↓` / `→` - Move to next panel
- `Shift+Tab` / `↑` / `←` - Move to previous panel

## Dependencies

- [tview](https://github.com/rivo/tview) - Terminal UI framework
- [asciigraph](https://github.com/guptarohit/asciigraph) - ASCII graph plotting
- [prometheus/client_golang](https://github.com/prometheus/client_golang) - Prometheus API client
- [influxdb-client-go](https://github.com/influxdata/influxdb-client-go) - InfluxDB v2 API client
- [yaml.v2](https://gopkg.in/yaml.v2) - YAML configuration parsing

## Requirements

- Go 1.19 or later
- Access to a data source:
  - Prometheus server (for Prometheus backend)
  - InfluxDB v2.x server (for InfluxDB backend)
  - InfluxDB v1.x server (for InfluxDB v1 backend)
- Terminal with color support (recommended)

## Error Handling

The tool provides meaningful error messages for:
- Invalid configuration files
- Prometheus connection failures
- Query execution errors
- Network timeouts

## Example Output

```
┌─ CPU Usage ────────────────────────────────────────────────────────┐
│Current: 0.23                                                       │
│Last Updated: 14:32:15                                              │
│                                                                    │
│    0.30 ┤                                          ╭─╮             │
│    0.28 ┤                                      ╭───╯ ╰─╮           │
│    0.25 ┤                                  ╭───╯       ╰─╮         │
│    0.23 ┤                              ╭───╯             ╰─╮       │
│    0.20 ┤                          ╭───╯                   ╰─╮     │
│    0.18 ┤                      ╭───╯                         ╰─╮   │
│    0.15 ┤                  ╭───╯                               ╰─╮ │
│    0.13 ┤              ╭───╯                                     ╰ │
└────────────────────────────────────────────────────────────────────┘
```

## Architecture

hyperbyte-plot uses a modular architecture that makes it easy to add support for new data sources:

- **`internal/app`** - Application orchestration and lifecycle management
- **`internal/backend`** - Pluggable backend interface with implementations:
  - `prom/` - Prometheus backend
  - `influxdb/` - InfluxDB v2 backend  
  - `influxdb1/` - InfluxDB v1 backend
  - `mock/` - Example mock backend for testing
- **`internal/config`** - Configuration management and validation
- **`internal/ui`** - Terminal user interface components

### Adding New Data Sources

To add support for a new data source (e.g., InfluxDB, TimescaleDB):

1. Create a new package in `internal/backend/`
2. Implement the `Backend` interface
3. Update the configuration structure
4. Add backend selection logic

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed instructions and examples.

## Extending hyperbyte-plot

The modular design allows for easy extension:

- **New Backends**: Add TimescaleDB, PostgreSQL, Elasticsearch, etc.
- **UI Enhancements**: Different visualization modes, export options
- **Configuration**: Multiple config formats, environment variables
- **Metrics**: Custom aggregations, alerting, thresholds
