# PromViz Architecture

This document describes the modular architecture of PromViz and how to extend it with new data sources.

## Package Structure

```
promviz/
├── main.go                           # Entry point
├── internal/
│   ├── app/
│   │   └── app.go                   # Application orchestration
│   ├── backend/
│   │   ├── types.go                 # Backend interface and common types
│   │   ├── prom/
│   │   │   └── client.go           # Prometheus implementation
│   │   ├── influxdb/
│   │   │   └── client.go           # InfluxDB v2 implementation
│   │   ├── influxdb1/
│   │   │   └── client.go           # InfluxDB v1 implementation
│   │   └── mock/
│   │       └── client.go           # Mock implementation (example)
│   ├── config/
│   │   └── config.go               # Configuration management
│   └── ui/
│       └── ui.go                   # Terminal user interface
├── queries.yaml                    # Configuration file
└── go.mod                          # Dependencies
```

## Architecture Overview

### 1. Main Entry Point (`main.go`)
- Minimal entry point that handles CLI flags
- Creates and starts the application
- Handles graceful shutdown

### 2. Application Layer (`internal/app`)
- Orchestrates all components
- Manages the application lifecycle
- Coordinates between backend and UI
- Handles periodic metric updates

### 3. Backend Layer (`internal/backend`)
- **Interface Definition**: `Backend` interface in `types.go`
- **Implementations**: Each database/source has its own package
  - `prom/`: Prometheus implementation
  - `influxdb/`: InfluxDB v2 implementation
  - `influxdb1/`: InfluxDB v1 implementation  
  - `mock/`: Example mock implementation
- **Extensible**: Easy to add new data sources

### 4. Configuration Layer (`internal/config`)
- YAML configuration parsing and validation
- Backend-specific configuration management
- Centralized configuration access

### 5. UI Layer (`internal/ui`)
- Terminal user interface using `tview`
- ASCII graph rendering with `asciigraph`
- Keyboard navigation and event handling
- Query history management

## Backend Interface

All backends must implement the `Backend` interface:

```go
type Backend interface {
    Connect(ctx context.Context) error
    Query(ctx context.Context, expr string) (float64, error)
    Close() error
    Name() string
}
```

## Adding New Data Sources

To add a new data source (e.g., InfluxDB, TimescaleDB), follow these steps:

### 1. Create a New Backend Package

```bash
mkdir -p internal/backend/influxdb
```

### 2. Implement the Backend Interface

Create `internal/backend/influxdb/client.go`:

```go
package influxdb

import (
    "context"
    "fmt"
)

// Config holds InfluxDB-specific configuration
type Config struct {
    URL    string `yaml:"url"`
    Token  string `yaml:"token"`
    Org    string `yaml:"org"`
    Bucket string `yaml:"bucket"`
}

func (c *Config) GetURL() string {
    return c.URL
}

// Client wraps the InfluxDB client
type Client struct {
    config *Config
    // Add InfluxDB client fields here
}

func NewClient(config *Config) (*Client, error) {
    // Initialize InfluxDB client
    return &Client{config: config}, nil
}

func (c *Client) Connect(ctx context.Context) error {
    // Test connection to InfluxDB
    return nil
}

func (c *Client) Query(ctx context.Context, expr string) (float64, error) {
    // Execute Flux query and return result
    return 0.0, nil
}

func (c *Client) Close() error {
    // Close InfluxDB connection
    return nil
}

func (c *Client) Name() string {
    return "influxdb"
}
```

### 3. Update Configuration

Add InfluxDB config to your configuration struct in `internal/config/config.go`:

```go
type Config struct {
    Prometheus prom.Config     `yaml:"prometheus,omitempty"`
    InfluxDB   influxdb.Config `yaml:"influxdb,omitempty"`
    Queries    []backend.Query `yaml:"queries"`
    Backend    string          `yaml:"backend"` // "prometheus", "influxdb", etc.
}
```

### 4. Update Backend Factory

Modify `createBackend()` in `internal/app/app.go`:

```go
func createBackend(cfg *config.Config) (backend.Backend, error) {
    switch cfg.Backend {
    case "prometheus", "":
        return prom.NewClient(cfg.GetPrometheusConfig())
    case "influxdb":
        return influxdb.NewClient(cfg.GetInfluxDBConfig())
    case "influxdb1":
        return influxdb1.NewClient(cfg.GetInfluxDB1Config())
    case "timescaledb":
        return timescaledb.NewClient(cfg.GetTimescaleDBConfig())
    default:
        return nil, fmt.Errorf("unsupported backend: %s", cfg.Backend)
    }
}
```

### 5. Update Configuration File

Example `queries.yaml` for InfluxDB:

```yaml
backend: influxdb
influxdb:
  url: "http://localhost:8086"
  token: "your-token-here"
  org: "your-org"
  bucket: "metrics"

queries:
  - name: CPU Usage
    expr: 'r._measurement == "cpu" and r._field == "usage_percent"'
  - name: Memory Usage
    expr: 'r._measurement == "mem" and r._field == "used_percent"'
```

## InfluxDB Support

PromViz now includes full InfluxDB v2 support with the following features:

### Configuration

```yaml
backend: influxdb
influxdb:
  url: "http://localhost:8086"
  token: "your-influxdb-token"
  org: "your-organization"
  bucket: "telegraf"

queries:
  - name: Simple Filter
    expr: 'r._measurement == "cpu" and r._field == "usage_idle"'
  
  - name: Complex Flux Query
    expr: |
      from(bucket: "telegraf")
      |> range(start: -5m)
      |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_active")
      |> mean()
```

### Features

- **Simple Expressions**: Use simplified filters for basic queries
- **Full Flux Queries**: Support for complex multi-line Flux queries
- **Automatic Query Wrapping**: Simple expressions are automatically wrapped with bucket context
- **Error Handling**: Comprehensive connection and query error reporting
- **Authentication**: Token-based authentication with organization/bucket scoping

## InfluxDB v1 Support

PromViz also supports InfluxDB v1 with InfluxQL queries:

### Configuration

```yaml
backend: influxdb1
influxdb1:
  url: "http://localhost:8086"
  username: "admin"
  password: "password"
  database: "telegraf"

queries:
  - name: Simple Field Query
    expr: 'usage_idle'
    
  - name: Full InfluxQL Query
    expr: |
      SELECT mean("usage_idle") 
      FROM "cpu" 
      WHERE time >= now() - 5m 
      GROUP BY time(1m) 
      ORDER BY time DESC 
      LIMIT 1
```

### Features

- **InfluxQL Support**: Full support for InfluxDB v1's SQL-like query language
- **HTTP API**: Direct HTTP implementation without external dependencies
- **Authentication**: Username/password authentication
- **Simple Expressions**: Basic field queries are auto-wrapped in SELECT statements
- **Complex Queries**: Full InfluxQL syntax support for advanced use cases
- **Error Handling**: Comprehensive error reporting for connection and query issues

## Benefits of This Architecture

1. **Separation of Concerns**: Each layer has a single responsibility
2. **Testability**: Each component can be tested independently
3. **Extensibility**: Easy to add new data sources without modifying existing code
4. **Maintainability**: Clear boundaries between components
5. **Reusability**: UI and app logic can work with any backend implementation

## Design Patterns Used

- **Strategy Pattern**: Backend interface allows swapping data sources
- **Factory Pattern**: Backend creation based on configuration
- **Observer Pattern**: UI updates based on data changes
- **Facade Pattern**: App package provides simple interface to complex subsystems

This architecture ensures that PromViz can grow to support multiple data sources while maintaining clean, maintainable code.
