package mock

import (
	"context"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	config := &Config{
		Seed: 12345,
	}

	client := NewClient(config)

	if client == nil {
		t.Fatal("NewClient should not return nil")
	}

	if client.config.Seed != 12345 {
		t.Errorf("Expected seed 12345, got %d", client.config.Seed)
	}

	if client.rand == nil {
		t.Error("Random generator should be initialized")
	}
}

func TestNewClientWithDefaultSeed(t *testing.T) {
	config := &Config{
		Seed: 0, // Should use current time
	}

	client := NewClient(config)

	if client == nil {
		t.Fatal("NewClient should not return nil")
	}

	// The client should work even if seed was 0 - we don't need to verify the internal seed value
	// Just verify the random generator works
	if client.rand == nil {
		t.Error("Random generator should be initialized")
	}
}

func TestConfigGetURL(t *testing.T) {
	config := &Config{
		Seed: 123,
	}

	url := config.GetURL()
	expected := "mock://localhost"

	if url != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, url)
	}
}

func TestClientConnect(t *testing.T) {
	config := &Config{Seed: 12345}
	client := NewClient(config)

	ctx := context.Background()
	err := client.Connect(ctx)

	if err != nil {
		t.Errorf("Connect should not return error, got %v", err)
	}
}

func TestClientQuery(t *testing.T) {
	config := &Config{Seed: 12345}
	client := NewClient(config)

	ctx := context.Background()

	// Test different query types
	tests := []struct {
		name  string
		query string
	}{
		{"CPU Query", "cpu_usage"},
		{"Memory Query", "memory_usage"},
		{"Generic Query", "some_other_metric"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeSeries, err := client.QueryTimeSeries(ctx, tt.query)
			if err != nil {
				t.Errorf("QueryTimeSeries should not return error, got %v", err)
			}

			if timeSeries == nil || len(timeSeries.Points) == 0 {
				t.Error("QueryTimeSeries should return at least one data point")
				return
			}

			// Check all values are within expected ranges
			for _, point := range timeSeries.Points {
				value := point.Value
				switch tt.query {
				case "cpu_usage":
					if value < 50 || value > 80 {
						t.Errorf("CPU usage should be 50-80, got %f", value)
					}
				case "memory_usage":
					if value < 4000 || value > 6000 {
						t.Errorf("Memory usage should be 4000-6000, got %f", value)
					}
				case "disk_usage":
					if value < 20 || value > 60 {
						t.Errorf("Disk usage should be 20-60, got %f", value)
					}
				case "network_bytes":
					if value < 1000 || value > 6000 {
						t.Errorf("Network bytes should be 1000-6000, got %f", value)
					}
				default:
					if value < 0 || value > 1000 {
						t.Errorf("Generic metric should be 0-1000, got %f", value)
					}
				}
			}
		})
	}
}

func TestClientQueryDeterministic(t *testing.T) {
	// Test that same seed produces same results
	config := &Config{Seed: 12345}
	client1 := NewClient(config)
	client2 := NewClient(config)

	ctx := context.Background()

	timeSeries1, err1 := client1.QueryTimeSeries(ctx, "cpu_usage")
	timeSeries2, err2 := client2.QueryTimeSeries(ctx, "cpu_usage")

	if err1 != nil || err2 != nil {
		t.Fatalf("QueryTimeSeries should not return errors, got %v, %v", err1, err2)
	}

	if timeSeries1 == nil || timeSeries2 == nil {
		t.Fatal("QueryTimeSeries should not return nil")
	}

	if len(timeSeries1.Points) != len(timeSeries2.Points) {
		t.Errorf("Same seed should produce same number of points, got %d and %d",
			len(timeSeries1.Points), len(timeSeries2.Points))
	}

	// Compare all data points
	for i := 0; i < len(timeSeries1.Points) && i < len(timeSeries2.Points); i++ {
		if timeSeries1.Points[i].Value != timeSeries2.Points[i].Value {
			t.Errorf("Same seed should produce same values at point %d, got %f and %f",
				i, timeSeries1.Points[i].Value, timeSeries2.Points[i].Value)
		}
	}
}

func TestClientClose(t *testing.T) {
	config := &Config{Seed: 12345}
	client := NewClient(config)

	err := client.Close()
	if err != nil {
		t.Errorf("Close should not return error, got %v", err)
	}
}

func TestClientName(t *testing.T) {
	config := &Config{Seed: 12345}
	client := NewClient(config)

	name := client.Name()
	expected := "mock"

	if name != expected {
		t.Errorf("Expected name '%s', got '%s'", expected, name)
	}
}

func TestQueryPerformance(t *testing.T) {
	config := &Config{Seed: 12345}
	client := NewClient(config)

	ctx := context.Background()

	start := time.Now()
	_, err := client.QueryTimeSeries(ctx, "cpu_usage")
	duration := time.Since(start)

	if err != nil {
		t.Errorf("QueryTimeSeries should not return error, got %v", err)
	}

	// Query should complete within reasonable time (allow for some variance)
	if duration > 200*time.Millisecond {
		t.Errorf("Query took too long: %v", duration)
	}
}
