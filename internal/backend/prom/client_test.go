package prom

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestConfigGetURL(t *testing.T) {
	config := &Config{
		URL: "http://prometheus:9090",
	}

	url := config.GetURL()
	expected := "http://prometheus:9090"

	if url != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, url)
	}
}

func TestNewClient(t *testing.T) {
	config := &Config{
		URL: "http://localhost:9090",
	}

	client, err := NewClient(config)

	if err != nil {
		t.Fatalf("NewClient should not return error, got %v", err)
	}

	if client == nil {
		t.Fatal("NewClient should not return nil")
	}

	if client.config.URL != config.URL {
		t.Errorf("Expected config URL %s, got %s", config.URL, client.config.URL)
	}

	if client.client == nil {
		t.Error("Prometheus client should be initialized")
	}

	if client.api == nil {
		t.Error("Prometheus API should be initialized")
	}
}

func TestNewClientInvalidURL(t *testing.T) {
	config := &Config{
		URL: "://invalid-url",
	}

	client, err := NewClient(config)

	if err == nil {
		t.Error("NewClient should return error for invalid URL")
	}

	if client != nil {
		t.Error("NewClient should return nil client on error")
	}
}

func TestClientName(t *testing.T) {
	config := &Config{URL: "http://localhost:9090"}
	client, err := NewClient(config)

	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	name := client.Name()
	expected := "prometheus"

	if name != expected {
		t.Errorf("Expected name '%s', got '%s'", expected, name)
	}
}

func TestClientClose(t *testing.T) {
	config := &Config{URL: "http://localhost:9090"}
	client, err := NewClient(config)

	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	err = client.Close()
	if err != nil {
		t.Errorf("Close should not return error, got %v", err)
	}
}

// Mock Prometheus server for testing
func createMockPrometheusServer(response string, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		w.Write([]byte(response))
	}))
}

func TestClientConnect(t *testing.T) {
	// Mock successful response for label names endpoint
	mockResponse := `{
		"status": "success",
		"data": ["__name__", "job", "instance"]
	}`

	server := createMockPrometheusServer(mockResponse, http.StatusOK)
	defer server.Close()

	config := &Config{URL: server.URL}
	client, err := NewClient(config)

	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		t.Errorf("Connect should not return error, got %v", err)
	}
}

func TestClientConnectFailure(t *testing.T) {
	// Use non-existent server
	config := &Config{URL: "http://localhost:1"}
	client, err := NewClient(config)

	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err == nil {
		t.Error("Connect should return error for non-existent server")
	}

	if !strings.Contains(err.Error(), "failed to connect to Prometheus") {
		t.Errorf("Error should mention Prometheus connection failure, got: %v", err)
	}
}

func TestClientQueryMatrix(t *testing.T) {
	// Mock successful matrix response (range query)
	mockResponse := `{
		"status": "success",
		"data": {
			"resultType": "matrix",
			"result": [
				{
					"metric": {"__name__": "cpu_usage"},
					"values": [
						[1609459200, "42.5"],
						[1609459260, "43.0"],
						[1609459320, "41.8"]
					]
				}
			]
		}
	}`

	server := createMockPrometheusServer(mockResponse, http.StatusOK)
	defer server.Close()

	config := &Config{URL: server.URL}
	client, err := NewClient(config)

	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx := context.Background()
	timeSeries, err := client.QueryTimeSeries(ctx, "cpu_usage")

	if err != nil {
		t.Fatalf("QueryTimeSeries should not return error, got %v", err)
	}

	if timeSeries == nil || len(timeSeries.Points) == 0 {
		t.Fatal("QueryTimeSeries should return time series data")
	}

	// Should have 3 data points from the mock response
	if len(timeSeries.Points) != 3 {
		t.Errorf("Expected 3 data points, got %d", len(timeSeries.Points))
	}

	// Check first value
	expected := 42.5
	value := timeSeries.Points[0].Value
	if value != expected {
		t.Errorf("Expected first value %f, got %f", expected, value)
	}

	// Check second value
	expectedSecond := 43.0
	secondValue := timeSeries.Points[1].Value
	if secondValue != expectedSecond {
		t.Errorf("Expected second value %f, got %f", expectedSecond, secondValue)
	}
}

func TestClientQueryMatrix2(t *testing.T) {
	// Mock successful matrix response with single data point
	mockResponse := `{
		"status": "success",
		"data": {
			"resultType": "matrix",
			"result": [
				{
					"metric": {"__name__": "cpu_usage"},
					"values": [
						[1609459200, "85.3"]
					]
				}
			]
		}
	}`

	server := createMockPrometheusServer(mockResponse, http.StatusOK)
	defer server.Close()

	config := &Config{URL: server.URL}
	client, err := NewClient(config)

	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx := context.Background()
	timeSeries, err := client.QueryTimeSeries(ctx, "scalar(cpu_usage)")

	if err != nil {
		t.Fatalf("QueryTimeSeries should not return error, got %v", err)
	}

	if timeSeries == nil || len(timeSeries.Points) == 0 {
		t.Fatal("QueryTimeSeries should return time series data")
	}

	// Should have 1 data point from the mock response
	if len(timeSeries.Points) != 1 {
		t.Errorf("Expected 1 data point, got %d", len(timeSeries.Points))
	}

	expected := 85.3
	value := timeSeries.Points[0].Value
	if value != expected {
		t.Errorf("Expected value %f, got %f", expected, value)
	}
}

func TestClientQueryEmptyMatrix(t *testing.T) {
	// Mock empty matrix response
	mockResponse := `{
		"status": "success",
		"data": {
			"resultType": "matrix",
			"result": []
		}
	}`

	server := createMockPrometheusServer(mockResponse, http.StatusOK)
	defer server.Close()

	config := &Config{URL: server.URL}
	client, err := NewClient(config)

	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx := context.Background()
	timeSeries, err := client.QueryTimeSeries(ctx, "nonexistent_metric")

	if err != nil {
		t.Errorf("QueryTimeSeries should not return error for empty result, got %v", err)
	}

	if timeSeries == nil {
		t.Error("QueryTimeSeries should return valid result even if empty")
	}

	if len(timeSeries.Points) != 0 {
		t.Errorf("Expected empty points for no data, got %d points", len(timeSeries.Points))
	}
}

func TestClientQueryError(t *testing.T) {
	// Mock error response
	mockResponse := `{
		"status": "error",
		"errorType": "bad_data",
		"error": "invalid query"
	}`

	server := createMockPrometheusServer(mockResponse, http.StatusBadRequest)
	defer server.Close()

	config := &Config{URL: server.URL}
	client, err := NewClient(config)

	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx := context.Background()
	_, err = client.QueryTimeSeries(ctx, "invalid{query")

	if err == nil {
		t.Error("Query should return error for invalid query")
	}

	if !strings.Contains(err.Error(), "query failed") {
		t.Errorf("Error should mention query failure, got: %v", err)
	}
}

func TestClientQueryUnsupportedType(t *testing.T) {
	// Mock unsupported result type
	mockResponse := `{
		"status": "success",
		"data": {
			"resultType": "string",
			"result": "some string value"
		}
	}`

	server := createMockPrometheusServer(mockResponse, http.StatusOK)
	defer server.Close()

	config := &Config{URL: server.URL}
	client, err := NewClient(config)

	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx := context.Background()
	_, err = client.QueryTimeSeries(ctx, "rate(cpu_usage[5m])")

	if err == nil {
		t.Error("Query should return error for unsupported result type")
	}

	if !strings.Contains(err.Error(), "query failed") {
		t.Errorf("Error should mention query failure, got: %v", err)
	}
}
