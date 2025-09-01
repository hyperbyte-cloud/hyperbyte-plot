package influxdb

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"promviz/internal/backend"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

// Config holds InfluxDB-specific configuration
type Config struct {
	URL    string `yaml:"url"`
	Token  string `yaml:"token"`
	Org    string `yaml:"org"`
	Bucket string `yaml:"bucket"`
}

// GetURL returns the InfluxDB server URL
func (c *Config) GetURL() string {
	return c.URL
}

// Client wraps the InfluxDB client
type Client struct {
	client   influxdb2.Client
	queryAPI api.QueryAPI
	config   *Config
}

// NewClient creates a new InfluxDB backend client
func NewClient(config *Config) (*Client, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("InfluxDB URL is required")
	}
	if config.Token == "" {
		return nil, fmt.Errorf("InfluxDB token is required")
	}
	if config.Org == "" {
		return nil, fmt.Errorf("InfluxDB organization is required")
	}
	if config.Bucket == "" {
		return nil, fmt.Errorf("InfluxDB bucket is required")
	}

	// Create InfluxDB client
	client := influxdb2.NewClient(config.URL, config.Token)
	queryAPI := client.QueryAPI(config.Org)

	return &Client{
		client:   client,
		queryAPI: queryAPI,
		config:   config,
	}, nil
}

// Connect establishes connection to InfluxDB and tests connectivity
func (c *Client) Connect(ctx context.Context) error {
	// Test connection by running a simple query
	query := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: -1m)
		|> limit(n: 1)
	`, c.config.Bucket)

	result, err := c.queryAPI.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to connect to InfluxDB at %s: %w", c.config.URL, err)
	}

	// Close the result to free resources
	if result != nil {
		result.Close()
	}

	return nil
}

// QueryTimeSeries executes a Flux query and returns time series data
func (c *Client) QueryTimeSeries(ctx context.Context, expr string) (*backend.TimeSeriesResult, error) {
	// If the expression doesn't contain bucket reference, wrap it with bucket info
	query := expr
	if !strings.Contains(query, "from(bucket:") {
		query = fmt.Sprintf(`
			from(bucket: "%s")
			|> range(start: -5m)
			|> filter(fn: (r) => %s)
			|> aggregateWindow(every: 1m, fn: mean, createEmpty: true)
			|> fill(value: 0.0)
			|> sort(columns: ["_time"], desc: true)
		`, c.config.Bucket, expr)
	}

	result, err := c.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer result.Close()

	var points []backend.DataPoint

	// Process the result
	for result.Next() {
		record := result.Record()
		timestamp := record.Time()

		if record.Value() != nil {
			var value float64
			switch v := record.Value().(type) {
			case float64:
				value = v
			case int64:
				value = float64(v)
			case string:
				if f, err := strconv.ParseFloat(v, 64); err == nil {
					value = f
				} else {
					continue
				}
			default:
				continue
			}

			points = append(points, backend.DataPoint{
				Timestamp: timestamp,
				Value:     value,
			})
		}
	}

	if result.Err() != nil {
		return nil, fmt.Errorf("error reading query result: %w", result.Err())
	}

	return &backend.TimeSeriesResult{Points: points}, nil
}

// Close closes the connection to InfluxDB
func (c *Client) Close() error {
	if c.client != nil {
		c.client.Close()
	}
	return nil
}

// Name returns the backend type name
func (c *Client) Name() string {
	return "influxdb"
}
