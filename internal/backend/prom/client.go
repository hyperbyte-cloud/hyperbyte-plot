package prom

import (
	"context"
	"fmt"
	"log"
	"time"

	"promviz/internal/backend"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// Config holds Prometheus-specific configuration
type Config struct {
	URL string `yaml:"url"`
}

// GetURL returns the Prometheus server URL
func (c *Config) GetURL() string {
	return c.URL
}

// Client wraps the Prometheus API client
type Client struct {
	client api.Client
	api    v1.API
	config *Config
}

// NewClient creates a new Prometheus backend client
func NewClient(config *Config) (*Client, error) {
	client, err := api.NewClient(api.Config{
		Address: config.URL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus client: %w", err)
	}

	return &Client{
		client: client,
		api:    v1.NewAPI(client),
		config: config,
	}, nil
}

// Connect establishes connection to Prometheus and tests connectivity
func (c *Client) Connect(ctx context.Context) error {
	// Test connection by trying to fetch label names
	_, _, err := c.api.LabelNames(ctx, nil, time.Now().Add(-time.Minute), time.Now())
	if err != nil {
		return fmt.Errorf("failed to connect to Prometheus at %s: %w", c.config.URL, err)
	}
	return nil
}

// QueryTimeSeries executes a PromQL range query and returns time series data
func (c *Client) QueryTimeSeries(ctx context.Context, expr string) (*backend.TimeSeriesResult, error) {
	// Query for the last 5 minutes with 1-minute step
	end := time.Now()
	start := end.Add(-5 * time.Minute)
	step := time.Minute

	result, warnings, err := c.api.QueryRange(ctx, expr, v1.Range{
		Start: start,
		End:   end,
		Step:  step,
	})
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	if len(warnings) > 0 {
		log.Printf("Warnings: %v", warnings)
	}

	switch result.Type() {
	case model.ValMatrix:
		matrix := result.(model.Matrix)
		var points []backend.DataPoint

		for _, sampleStream := range matrix {
			for _, sample := range sampleStream.Values {
				points = append(points, backend.DataPoint{
					Timestamp: sample.Timestamp.Time(),
					Value:     float64(sample.Value),
				})
			}
		}

		return &backend.TimeSeriesResult{Points: points}, nil
	default:
		return nil, fmt.Errorf("unsupported result type for range query: %v", result.Type())
	}
}

// Close closes the connection (no-op for Prometheus client)
func (c *Client) Close() error {
	// Prometheus client doesn't require explicit closing
	return nil
}

// Name returns the backend type name
func (c *Client) Name() string {
	return "prometheus"
}
