package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"promviz/internal/backend"
	"promviz/internal/backend/influxdb"
	"promviz/internal/backend/prom"
	"promviz/internal/config"
)

func TestCreateBackendPrometheus(t *testing.T) {
	cfg := &config.Config{
		Backend: "prometheus",
		Prometheus: prom.Config{
			URL: "http://localhost:9090",
		},
		Queries: []backend.Query{
			{Name: "Test", Expr: "test_metric"},
		},
	}

	backend, err := createBackend(cfg)
	if err != nil {
		t.Fatalf("createBackend should not return error, got %v", err)
	}

	if backend == nil {
		t.Fatal("createBackend should not return nil")
	}

	if backend.Name() != "prometheus" {
		t.Errorf("Expected backend name 'prometheus', got '%s'", backend.Name())
	}
}

func TestCreateBackendPrometheusDefault(t *testing.T) {
	cfg := &config.Config{
		Backend: "", // Empty should default to prometheus
		Prometheus: prom.Config{
			URL: "http://localhost:9090",
		},
		Queries: []backend.Query{
			{Name: "Test", Expr: "test_metric"},
		},
	}

	backend, err := createBackend(cfg)
	if err != nil {
		t.Fatalf("createBackend should not return error, got %v", err)
	}

	if backend == nil {
		t.Fatal("createBackend should not return nil")
	}

	if backend.Name() != "prometheus" {
		t.Errorf("Expected backend name 'prometheus', got '%s'", backend.Name())
	}
}

func TestCreateBackendInfluxDB(t *testing.T) {
	cfg := &config.Config{
		Backend: "influxdb",
		InfluxDB: influxdb.Config{
			URL:    "http://localhost:8086",
			Token:  "test-token",
			Org:    "test-org",
			Bucket: "test-bucket",
		},
		Queries: []backend.Query{
			{Name: "Test", Expr: "test_expr"},
		},
	}

	backend, err := createBackend(cfg)
	if err != nil {
		t.Fatalf("createBackend should not return error, got %v", err)
	}

	if backend == nil {
		t.Fatal("createBackend should not return nil")
	}

	if backend.Name() != "influxdb" {
		t.Errorf("Expected backend name 'influxdb', got '%s'", backend.Name())
	}
}

func TestCreateBackendUnsupported(t *testing.T) {
	cfg := &config.Config{
		Backend: "unsupported",
		Queries: []backend.Query{
			{Name: "Test", Expr: "test_expr"},
		},
	}

	backend, err := createBackend(cfg)
	if err == nil {
		t.Error("createBackend should return error for unsupported backend")
	}

	if backend != nil {
		t.Error("createBackend should return nil backend on error")
	}

	if !strings.Contains(err.Error(), "unsupported backend: unsupported") {
		t.Errorf("Error should mention unsupported backend, got: %v", err)
	}
}

func TestNewAppSuccess(t *testing.T) {
	// Create temporary config file
	configContent := `prometheus:
  url: "http://localhost:9090"

queries:
  - name: Test Query
    expr: test_metric
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	// Note: This will fail at Connect() stage since we don't have a real Prometheus server
	// But we can test that the config loading and backend creation works
	_, err = New(configPath)

	// We expect a connection error, not a config or backend creation error
	if err == nil {
		t.Error("Expected connection error since no Prometheus server is running")
	}

	if !strings.Contains(err.Error(), "failed to connect to Prometheus") {
		t.Errorf("Expected Prometheus connection error, got: %v", err)
	}
}

func TestNewAppConfigError(t *testing.T) {
	// Test with non-existent config file
	_, err := New("nonexistent.yaml")
	if err == nil {
		t.Error("New should return error for non-existent config file")
	}

	if !strings.Contains(err.Error(), "failed to load config") {
		t.Errorf("Error should mention failed to load config, got: %v", err)
	}
}

func TestNewAppInvalidConfig(t *testing.T) {
	// Create temporary config file with invalid YAML
	configContent := `invalid yaml [[[`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	_, err = New(configPath)
	if err == nil {
		t.Error("New should return error for invalid YAML")
	}

	if !strings.Contains(err.Error(), "failed to load config") {
		t.Errorf("Error should mention failed to load config, got: %v", err)
	}
}

func TestNewAppMissingPrometheusURL(t *testing.T) {
	// Create temporary config file with missing Prometheus URL
	configContent := `prometheus:
  # url is missing

queries:
  - name: Test Query
    expr: test_metric
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	_, err = New(configPath)
	if err == nil {
		t.Error("New should return error for missing Prometheus URL")
	}

	if !strings.Contains(err.Error(), "failed to load config") {
		t.Errorf("Error should mention failed to load config, got: %v", err)
	}
}

func TestNewAppInfluxDBConfig(t *testing.T) {
	// Create temporary config file for InfluxDB
	configContent := `backend: influxdb
influxdb:
  url: "http://localhost:8086"
  token: "test-token"
  org: "test-org"
  bucket: "test-bucket"

queries:
  - name: Test Query
    expr: 'r._measurement == "cpu"'
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	// Note: This will fail at Connect() stage since we don't have a real InfluxDB server
	_, err = New(configPath)

	// We expect a connection error, not a config or backend creation error
	if err == nil {
		t.Error("Expected connection error since no InfluxDB server is running")
	}

	if !strings.Contains(err.Error(), "failed to connect to InfluxDB") {
		t.Errorf("Expected InfluxDB connection error, got: %v", err)
	}
}

func TestNewAppUnsupportedBackend(t *testing.T) {
	// Create temporary config file with unsupported backend
	configContent := `backend: unsupported
queries:
  - name: Test Query
    expr: test_metric
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	_, err = New(configPath)
	if err == nil {
		t.Error("New should return error for unsupported backend")
	}

	if !strings.Contains(err.Error(), "failed to load config") {
		t.Errorf("Error should mention failed to load config, got: %v", err)
	}
}

// Mock tests would require more complex setup with test servers
// For now, we focus on the configuration and backend creation logic
// Integration tests with actual servers would be in a separate test suite
