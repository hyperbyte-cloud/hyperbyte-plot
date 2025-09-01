package mock

import (
	"context"
	"math/rand"
	"time"

	"promviz/internal/backend"
)

// Config holds mock backend configuration
type Config struct {
	Seed int64 `yaml:"seed"`
}

// GetURL returns a mock URL for demonstration
func (c *Config) GetURL() string {
	return "mock://localhost"
}

// Client is a mock backend for testing/demonstration
type Client struct {
	config *Config
	rand   *rand.Rand
}

// NewClient creates a new mock backend client
func NewClient(config *Config) *Client {
	seed := config.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	return &Client{
		config: config,
		rand:   rand.New(rand.NewSource(seed)),
	}
}

// Connect simulates establishing a connection (always succeeds)
func (c *Client) Connect(ctx context.Context) error {
	// Simulate connection delay
	time.Sleep(100 * time.Millisecond)
	return nil
}

// QueryTimeSeries simulates executing a query and returns time series data
func (c *Client) QueryTimeSeries(ctx context.Context, expr string) (*backend.TimeSeriesResult, error) {
	// Simulate query processing time
	time.Sleep(time.Duration(c.rand.Intn(50)) * time.Millisecond)

	// Generate 5 minutes of data with 1-minute intervals
	var points []backend.DataPoint
	now := time.Now()

	for i := 4; i >= 0; i-- {
		timestamp := now.Add(-time.Duration(i) * time.Minute)

		// Generate value based on the query expression
		var baseValue float64
		switch expr {
		case "cpu_usage":
			baseValue = 50 + c.rand.Float64()*30 // 50-80% range
		case "memory_usage":
			baseValue = 4000 + c.rand.Float64()*2000 // 4000-6000 MB range
		case "disk_usage":
			baseValue = 20 + c.rand.Float64()*40 // 20-60% range
		case "network_bytes":
			baseValue = 1000 + c.rand.Float64()*5000 // 1000-6000 bytes range
		default:
			baseValue = c.rand.Float64() * 1000
		}

		points = append(points, backend.DataPoint{
			Timestamp: timestamp,
			Value:     baseValue,
		})
	}

	return &backend.TimeSeriesResult{Points: points}, nil
}

// Close closes the mock connection (no-op)
func (c *Client) Close() error {
	return nil
}

// Name returns the backend type name
func (c *Client) Name() string {
	return "mock"
}
