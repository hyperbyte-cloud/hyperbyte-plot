package influxdb1

import (
	"context"
	"strings"
	"testing"
)

func TestConfigGetURL(t *testing.T) {
	config := &Config{
		URL:      "http://influxdb:8086",
		Username: "admin",
		Password: "password",
		Database: "telegraf",
	}

	url := config.GetURL()
	expected := "http://influxdb:8086"

	if url != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, url)
	}
}

func TestNewClient(t *testing.T) {
	config := &Config{
		URL:      "http://localhost:8086",
		Username: "admin",
		Password: "password",
		Database: "telegraf",
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
		t.Error("InfluxDB v1 client should be initialized")
	}
}

func TestNewClientMissingURL(t *testing.T) {
	config := &Config{
		Username: "admin",
		Password: "password",
		Database: "telegraf",
	}

	client, err := NewClient(config)

	if err == nil {
		t.Error("NewClient should return error for missing URL")
	}

	if client != nil {
		t.Error("NewClient should return nil client on error")
	}

	if !strings.Contains(err.Error(), "URL is required") {
		t.Errorf("Error should mention missing URL, got: %v", err)
	}
}

func TestNewClientMissingDatabase(t *testing.T) {
	config := &Config{
		URL:      "http://localhost:8086",
		Username: "admin",
		Password: "password",
	}

	client, err := NewClient(config)

	if err == nil {
		t.Error("NewClient should return error for missing database")
	}

	if client != nil {
		t.Error("NewClient should return nil client on error")
	}

	if !strings.Contains(err.Error(), "database is required") {
		t.Errorf("Error should mention missing database, got: %v", err)
	}
}

func TestClientName(t *testing.T) {
	config := &Config{
		URL:      "http://localhost:8086",
		Username: "admin",
		Password: "password",
		Database: "telegraf",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	name := client.Name()
	expected := "influxdb1"

	if name != expected {
		t.Errorf("Expected name '%s', got '%s'", expected, name)
	}
}

func TestClientClose(t *testing.T) {
	config := &Config{
		URL:      "http://localhost:8086",
		Username: "admin",
		Password: "password",
		Database: "telegraf",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	err = client.Close()
	if err != nil {
		t.Errorf("Close should not return error, got %v", err)
	}
}

func TestClientConnectFailure(t *testing.T) {
	// Use non-existent server
	config := &Config{
		URL:      "http://localhost:1",
		Username: "admin",
		Password: "password",
		Database: "telegraf",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx := context.Background()
	err = client.Connect(ctx)
	if err == nil {
		t.Error("Connect should return error for non-existent server")
	}

	if !strings.Contains(err.Error(), "failed to connect to InfluxDB v1") {
		t.Errorf("Error should mention InfluxDB v1 connection failure, got: %v", err)
	}
}

func TestClientQueryFailure(t *testing.T) {
	// Use non-existent server
	config := &Config{
		URL:      "http://localhost:1",
		Username: "admin",
		Password: "password",
		Database: "telegraf",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx := context.Background()
	_, err = client.QueryTimeSeries(ctx, "SELECT mean(usage_idle) FROM cpu")

	if err == nil {
		t.Error("QueryTimeSeries should return error for non-existent server")
	}

	if !strings.Contains(err.Error(), "query failed") {
		t.Errorf("Error should mention query failure, got: %v", err)
	}
}

func TestGetDefaultMeasurement(t *testing.T) {
	config := &Config{
		URL:      "http://localhost:8086",
		Username: "admin",
		Password: "password",
		Database: "telegraf",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	tests := []struct {
		expr     string
		expected string
	}{
		{"cpu_usage", "cpu"},
		{"memory_usage", "mem"},
		{"disk_free", "disk"},
		{"network_bytes", "net"},
		{"unknown_metric", "metrics"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := client.getDefaultMeasurement(tt.expr)
			if result != tt.expected {
				t.Errorf("Expected measurement '%s' for expr '%s', got '%s'", tt.expected, tt.expr, result)
			}
		})
	}
}

func TestConvertToFloat64(t *testing.T) {
	config := &Config{
		URL:      "http://localhost:8086",
		Username: "admin",
		Password: "password",
		Database: "telegraf",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	tests := []struct {
		name     string
		value    interface{}
		expected float64
		hasError bool
	}{
		{"float64", float64(42.5), 42.5, false},
		{"int64", int64(42), 42.0, false},
		{"int", int(42), 42.0, false},
		{"string number", "42.5", 42.5, false},
		{"string invalid", "invalid", 0, true},
		{"nil", nil, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.convertToFloat64(tt.value)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %f, got %f", tt.expected, result)
				}
			}
		})
	}
}
