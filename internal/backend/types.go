package backend

import (
	"context"
	"time"
)

// DataPoint represents a single metric data point
type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// TimeSeriesResult represents a time series of metric data points
type TimeSeriesResult struct {
	Points []DataPoint `json:"points"`
}

// Query represents a named query configuration
type Query struct {
	Name string `yaml:"name"`
	Expr string `yaml:"expr"`
}

// Backend defines the interface for metric data sources
type Backend interface {
	// Connect establishes connection to the backend
	Connect(ctx context.Context) error

	// QueryTimeSeries executes a query and returns time series data
	QueryTimeSeries(ctx context.Context, expr string) (*TimeSeriesResult, error)

	// Close closes the connection to the backend
	Close() error

	// Name returns the backend type name
	Name() string
}

// Config represents backend-specific configuration
type Config interface {
	GetURL() string
}
