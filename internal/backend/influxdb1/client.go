package influxdb1

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"promviz/internal/backend"

	client "github.com/influxdata/influxdb/client/v2"
)

// Config holds InfluxDB v1-specific configuration
type Config struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	UseHTTPS bool   `yaml:"use_https,omitempty"`
}

// GetURL returns the InfluxDB v1 server URL
func (c *Config) GetURL() string {
	return c.URL
}

// Client wraps the InfluxDB v1 client
type Client struct {
	client client.Client
	config *Config
}

// NewClient creates a new InfluxDB v1 backend client
func NewClient(config *Config) (*Client, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("InfluxDB v1 URL is required")
	}
	if config.Database == "" {
		return nil, fmt.Errorf("InfluxDB v1 database is required")
	}

	// Create InfluxDB v1 client configuration
	conf := client.HTTPConfig{
		Addr:     config.URL,
		Username: config.Username,
		Password: config.Password,
		Timeout:  time.Duration(30) * time.Second,
	}

	// Create client
	influxClient, err := client.NewHTTPClient(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create InfluxDB v1 client: %w", err)
	}

	return &Client{
		client: influxClient,
		config: config,
	}, nil
}

// Connect establishes connection to InfluxDB v1 and tests connectivity
func (c *Client) Connect(ctx context.Context) error {
	// Test connection by running a simple SHOW DATABASES query
	query := client.Query{
		Command:  "SHOW DATABASES",
		Database: "",
	}

	response, err := c.client.Query(query)
	if err != nil {
		return fmt.Errorf("failed to connect to InfluxDB v1 at %s: %w", c.config.URL, err)
	}

	if response.Error() != nil {
		return fmt.Errorf("InfluxDB v1 query error: %w", response.Error())
	}

	return nil
}

// QueryTimeSeries executes an InfluxQL query and returns time series data
func (c *Client) QueryTimeSeries(ctx context.Context, expr string) (*backend.TimeSeriesResult, error) {
	// Build the InfluxQL query - default to 5 minutes of data with 1-minute intervals
	var queryStr string
	if strings.Contains(strings.ToUpper(expr), "SELECT") {
		// Full InfluxQL query provided
		queryStr = expr
	} else {
		// Simple expression - wrap in SELECT statement with time series aggregation
		measurement := c.getDefaultMeasurement(expr)
		queryStr = fmt.Sprintf("SELECT mean(\"%s\") FROM \"%s\" WHERE time >= now() - 5m GROUP BY time(1m) fill(0) ORDER BY time DESC", expr, measurement)
	}

	query := client.Query{
		Command:  queryStr,
		Database: c.config.Database,
	}

	response, err := c.client.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	if response.Error() != nil {
		return nil, fmt.Errorf("InfluxDB v1 query error: %w", response.Error())
	}

	// Process the response
	if len(response.Results) == 0 {
		return &backend.TimeSeriesResult{Points: []backend.DataPoint{}}, nil
	}

	result := response.Results[0]
	if len(result.Series) == 0 {
		return &backend.TimeSeriesResult{Points: []backend.DataPoint{}}, nil
	}

	series := result.Series[0]
	if len(series.Values) == 0 {
		return &backend.TimeSeriesResult{Points: []backend.DataPoint{}}, nil
	}

	// Convert to time series data points
	var points []backend.DataPoint
	for _, values := range series.Values {
		if len(values) < 2 {
			continue
		}

		// Parse timestamp (first column)
		timestampStr, ok := values[0].(string)
		if !ok {
			continue
		}
		timestamp, err := time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			continue
		}

		// Parse value (second column)
		if values[1] == nil {
			// Skip null values or use 0 for fill(0)
			points = append(points, backend.DataPoint{
				Timestamp: timestamp,
				Value:     0,
			})
			continue
		}

		value, err := c.convertToFloat64(values[1])
		if err != nil {
			continue
		}

		points = append(points, backend.DataPoint{
			Timestamp: timestamp,
			Value:     value,
		})
	}

	return &backend.TimeSeriesResult{Points: points}, nil
}

// convertToFloat64 converts various types to float64
func (c *Client) convertToFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case int64:
		return float64(v), nil
	case int:
		return float64(v), nil
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, nil
		}
		return 0, fmt.Errorf("cannot convert string value to float: %s", v)
	case json.Number:
		if f, err := v.Float64(); err == nil {
			return f, nil
		}
		return 0, fmt.Errorf("cannot convert json.Number to float: %s", v)
	default:
		return 0, fmt.Errorf("unsupported value type: %T", v)
	}
}

// getDefaultMeasurement tries to extract measurement name from expression
func (c *Client) getDefaultMeasurement(expr string) string {
	// For simple field expressions, we need to specify a measurement
	// This is a simple heuristic - in practice, users should provide full queries
	if strings.Contains(expr, "cpu") {
		return "cpu"
	}
	if strings.Contains(expr, "memory") || strings.Contains(expr, "mem") {
		return "mem"
	}
	if strings.Contains(expr, "disk") {
		return "disk"
	}
	if strings.Contains(expr, "net") {
		return "net"
	}
	// Default fallback
	return "metrics"
}

// Close closes the connection to InfluxDB v1
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// Name returns the backend type name
func (c *Client) Name() string {
	return "influxdb1"
}
