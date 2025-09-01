package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"promviz/internal/backend"
	"promviz/internal/backend/influxdb"
	"promviz/internal/backend/prom"
)

// TestPrometheusServer creates a mock Prometheus server for testing
func TestPrometheusServer() (*httptest.Server, *prom.Config) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/labels" {
			// Mock labels endpoint for connection testing
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"status": "success",
				"data": ["__name__", "job", "instance"]
			}`))
		} else if r.URL.Path == "/api/v1/query" {
			// Mock query endpoint
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {"__name__": "test_metric"},
							"value": [1609459200, "42.5"]
						}
					]
				}
			}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	config := &prom.Config{
		URL: server.URL,
	}

	return server, config
}

// TestInfluxDBServer creates a mock InfluxDB server for testing
func TestInfluxDBServer() (*httptest.Server, *influxdb.Config) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/query" {
			// Mock query endpoint
			w.Header().Set("Content-Type", "application/csv")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,0,2023-01-01T00:00:00Z,2023-01-01T01:00:00Z,2023-01-01T00:30:00Z,42.5,usage_user,cpu,cpu-total,server1
`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	config := &influxdb.Config{
		URL:    server.URL,
		Token:  "test-token",
		Org:    "test-org",
		Bucket: "test-bucket",
	}

	return server, config
}

// BenchmarkBackend provides a simple benchmark for backend implementations
func BenchmarkBackend(b *testing.B, backend backend.Backend, query string) {
	ctx := context.Background()

	// Test connection first
	if err := backend.Connect(ctx); err != nil {
		b.Fatalf("Failed to connect to backend: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := backend.QueryTimeSeries(ctx, query)
		if err != nil {
			b.Fatalf("QueryTimeSeries failed: %v", err)
		}
	}
}

// TestBackendInterface provides a generic test for any backend implementation
func TestBackendInterface(t *testing.T, backend backend.Backend, testQuery string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test Name method
	name := backend.Name()
	if name == "" {
		t.Error("Backend name should not be empty")
	}

	// Test Connect method
	err := backend.Connect(ctx)
	if err != nil {
		t.Errorf("Connect should not return error: %v", err)
	}

	// Test QueryTimeSeries method
	timeSeries, err := backend.QueryTimeSeries(ctx, testQuery)
	if err != nil {
		t.Errorf("QueryTimeSeries should not return error: %v", err)
	}

	if timeSeries == nil {
		t.Error("QueryTimeSeries should not return nil")
		return
	}

	if len(timeSeries.Points) == 0 {
		t.Error("QueryTimeSeries should return at least one data point")
		return
	}

	// Check that the latest value is non-negative
	latestValue := timeSeries.Points[len(timeSeries.Points)-1].Value
	if latestValue < 0 {
		t.Errorf("QueryTimeSeries should return non-negative values, got %f", latestValue)
	}

	// Test Close method
	err = backend.Close()
	if err != nil {
		t.Errorf("Close should not return error: %v", err)
	}
}

// MockErrorBackend implements Backend but returns errors for testing
type MockErrorBackend struct {
	connectError error
	queryError   error
	closeError   error
}

func (m *MockErrorBackend) Connect(ctx context.Context) error {
	return m.connectError
}

func (m *MockErrorBackend) Query(ctx context.Context, expr string) (float64, error) {
	return 0, m.queryError
}

func (m *MockErrorBackend) Close() error {
	return m.closeError
}

func (m *MockErrorBackend) Name() string {
	return "mock-error"
}

// NewMockErrorBackend creates a backend that returns specific errors
func NewMockErrorBackend(connectErr, queryErr, closeErr error) *MockErrorBackend {
	return &MockErrorBackend{
		connectError: connectErr,
		queryError:   queryErr,
		closeError:   closeErr,
	}
}
