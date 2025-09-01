//go:build integration
// +build integration

package main

import (
	"testing"

	"promviz/internal/backend/influxdb"
	"promviz/internal/backend/prom"
)

func TestPrometheusIntegration(t *testing.T) {
	server, config := TestPrometheusServer()
	defer server.Close()

	client, err := prom.NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create Prometheus client: %v", err)
	}
	defer client.Close()

	TestBackendInterface(t, client, "test_metric")
}

func TestInfluxDBIntegration(t *testing.T) {
	server, config := TestInfluxDBServer()
	defer server.Close()

	client, err := influxdb.NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create InfluxDB client: %v", err)
	}
	defer client.Close()

	TestBackendInterface(t, client, `r._measurement == "cpu"`)
}

func BenchmarkPrometheusQuery(b *testing.B) {
	server, config := TestPrometheusServer()
	defer server.Close()

	client, err := prom.NewClient(config)
	if err != nil {
		b.Fatalf("Failed to create Prometheus client: %v", err)
	}
	defer client.Close()

	BenchmarkBackend(b, client, "test_metric")
}

func BenchmarkInfluxDBQuery(b *testing.B) {
	server, config := TestInfluxDBServer()
	defer server.Close()

	client, err := influxdb.NewClient(config)
	if err != nil {
		b.Fatalf("Failed to create InfluxDB client: %v", err)
	}
	defer client.Close()

	BenchmarkBackend(b, client, `r._measurement == "cpu"`)
}
