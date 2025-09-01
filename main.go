package main

import (
	"flag"
	"fmt"
	"os"

	"promviz/internal/app"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "queries.yaml", "Path to configuration file")
	flag.Parse()

	// Check if config file exists
	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Configuration file '%s' does not exist.\n", *configPath)
		fmt.Fprintf(os.Stderr, "Please create a configuration file or specify a different path with --config.\n\n")
		fmt.Fprintf(os.Stderr, "Example configurations:\n\n")
		fmt.Fprintf(os.Stderr, "Prometheus:\n")
		fmt.Fprintf(os.Stderr, `prometheus:
  url: "http://localhost:9090"

queries:
  - name: CPU Usage
    expr: rate(node_cpu_seconds_total{mode="user"}[5m])
  - name: Memory Usage  
    expr: node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes

`)
		fmt.Fprintf(os.Stderr, "InfluxDB v2 (Flux):\n")
		fmt.Fprintf(os.Stderr, `backend: influxdb
influxdb:
  url: "http://localhost:8086"
  token: "your-token"
  org: "your-org"
  bucket: "metrics"

queries:
  - name: CPU Usage
    expr: 'r._measurement == "cpu" and r._field == "usage_percent"'

`)
		fmt.Fprintf(os.Stderr, "InfluxDB v1 (InfluxQL):\n")
		fmt.Fprintf(os.Stderr, `backend: influxdb1
influxdb1:
  url: "http://localhost:8086"
  username: "admin"
  password: "password"
  database: "telegraf"

queries:
  - name: CPU Usage
    expr: 'SELECT mean("usage_idle") FROM "cpu" WHERE time >= now() - 5m'
`)
		os.Exit(1)
	}

	// Create and start the application
	application, err := app.New(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Handle graceful shutdown
	if err := application.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Application error: %v\n", err)
		os.Exit(1)
	}
}
