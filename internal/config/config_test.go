package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"promviz/internal/backend"
	"promviz/internal/backend/influxdb"
	"promviz/internal/backend/influxdb1"
	"promviz/internal/backend/prom"
)

func TestLoadConfigPrometheus(t *testing.T) {
	// Create temporary config file
	configContent := `prometheus:
  url: "http://localhost:9090"

queries:
  - name: CPU Usage
    expr: rate(cpu_usage[5m])
  - name: Memory Usage
    expr: memory_used / memory_total
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig should not return error, got %v", err)
	}

	// Test defaults
	if config.Backend != "prometheus" {
		t.Errorf("Expected backend 'prometheus', got '%s'", config.Backend)
	}

	// Test Prometheus config
	if config.Prometheus.URL != "http://localhost:9090" {
		t.Errorf("Expected Prometheus URL 'http://localhost:9090', got '%s'", config.Prometheus.URL)
	}

	// Test queries
	if len(config.Queries) != 2 {
		t.Errorf("Expected 2 queries, got %d", len(config.Queries))
	}

	expectedQueries := []backend.Query{
		{Name: "CPU Usage", Expr: "rate(cpu_usage[5m])"},
		{Name: "Memory Usage", Expr: "memory_used / memory_total"},
	}

	for i, expected := range expectedQueries {
		if config.Queries[i].Name != expected.Name {
			t.Errorf("Expected query name '%s', got '%s'", expected.Name, config.Queries[i].Name)
		}
		if config.Queries[i].Expr != expected.Expr {
			t.Errorf("Expected query expr '%s', got '%s'", expected.Expr, config.Queries[i].Expr)
		}
	}
}

func TestLoadConfigInfluxDB(t *testing.T) {
	// Create temporary config file
	configContent := `backend: influxdb
influxdb:
  url: "http://localhost:8086"
  token: "test-token"
  org: "test-org"
  bucket: "test-bucket"

queries:
  - name: CPU Usage
    expr: 'r._measurement == "cpu"'
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig should not return error, got %v", err)
	}

	// Test backend
	if config.Backend != "influxdb" {
		t.Errorf("Expected backend 'influxdb', got '%s'", config.Backend)
	}

	// Test InfluxDB config
	if config.InfluxDB.URL != "http://localhost:8086" {
		t.Errorf("Expected InfluxDB URL 'http://localhost:8086', got '%s'", config.InfluxDB.URL)
	}
	if config.InfluxDB.Token != "test-token" {
		t.Errorf("Expected InfluxDB token 'test-token', got '%s'", config.InfluxDB.Token)
	}
	if config.InfluxDB.Org != "test-org" {
		t.Errorf("Expected InfluxDB org 'test-org', got '%s'", config.InfluxDB.Org)
	}
	if config.InfluxDB.Bucket != "test-bucket" {
		t.Errorf("Expected InfluxDB bucket 'test-bucket', got '%s'", config.InfluxDB.Bucket)
	}

	// Test queries
	if len(config.Queries) != 1 {
		t.Errorf("Expected 1 query, got %d", len(config.Queries))
	}

	if config.Queries[0].Name != "CPU Usage" {
		t.Errorf("Expected query name 'CPU Usage', got '%s'", config.Queries[0].Name)
	}
	if config.Queries[0].Expr != `r._measurement == "cpu"` {
		t.Errorf("Expected query expr 'r._measurement == \"cpu\"', got '%s'", config.Queries[0].Expr)
	}
}

func TestLoadConfigInfluxDB1(t *testing.T) {
	// Create temporary config file
	configContent := `backend: influxdb1
influxdb1:
  url: "http://localhost:8086"
  username: "admin"
  password: "password"
  database: "telegraf"

queries:
  - name: CPU Usage
    expr: 'SELECT mean("usage_idle") FROM "cpu" WHERE time >= now() - 5m'
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig should not return error, got %v", err)
	}

	// Test backend
	if config.Backend != "influxdb1" {
		t.Errorf("Expected backend 'influxdb1', got '%s'", config.Backend)
	}

	// Test InfluxDB v1 config
	if config.InfluxDB1.URL != "http://localhost:8086" {
		t.Errorf("Expected InfluxDB v1 URL 'http://localhost:8086', got '%s'", config.InfluxDB1.URL)
	}
	if config.InfluxDB1.Username != "admin" {
		t.Errorf("Expected InfluxDB v1 username 'admin', got '%s'", config.InfluxDB1.Username)
	}
	if config.InfluxDB1.Password != "password" {
		t.Errorf("Expected InfluxDB v1 password 'password', got '%s'", config.InfluxDB1.Password)
	}
	if config.InfluxDB1.Database != "telegraf" {
		t.Errorf("Expected InfluxDB v1 database 'telegraf', got '%s'", config.InfluxDB1.Database)
	}

	// Test queries
	if len(config.Queries) != 1 {
		t.Errorf("Expected 1 query, got %d", len(config.Queries))
	}

	if config.Queries[0].Name != "CPU Usage" {
		t.Errorf("Expected query name 'CPU Usage', got '%s'", config.Queries[0].Name)
	}
	if config.Queries[0].Expr != `SELECT mean("usage_idle") FROM "cpu" WHERE time >= now() - 5m` {
		t.Errorf("Expected query expr 'SELECT mean(\"usage_idle\") FROM \"cpu\" WHERE time >= now() - 5m', got '%s'", config.Queries[0].Expr)
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := LoadConfig("nonexistent.yaml")
	if err == nil {
		t.Error("LoadConfig should return error for nonexistent file")
	}

	if !strings.Contains(err.Error(), "failed to read config file") {
		t.Errorf("Error should mention failed to read config file, got: %v", err)
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	// Create temporary config file with invalid YAML
	configContent := `prometheus:
  url: "http://localhost:9090"
  invalid yaml syntax [
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	_, err = LoadConfig(configPath)
	if err == nil {
		t.Error("LoadConfig should return error for invalid YAML")
	}

	if !strings.Contains(err.Error(), "failed to parse YAML") {
		t.Errorf("Error should mention failed to parse YAML, got: %v", err)
	}
}

func TestValidatePrometheusConfig(t *testing.T) {
	config := &Config{
		Prometheus: prom.Config{URL: "http://localhost:9090"},
		Queries: []backend.Query{
			{Name: "Test", Expr: "test_metric"},
		},
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("Validate should not return error for valid Prometheus config, got %v", err)
	}

	// Should default to prometheus backend
	if config.Backend != "prometheus" {
		t.Errorf("Backend should default to 'prometheus', got '%s'", config.Backend)
	}
}

func TestValidatePrometheusMissingURL(t *testing.T) {
	config := &Config{
		Backend: "prometheus",
		Queries: []backend.Query{
			{Name: "Test", Expr: "test_metric"},
		},
	}

	err := config.Validate()
	if err == nil {
		t.Error("Validate should return error for missing Prometheus URL")
	}

	if !strings.Contains(err.Error(), "prometheus.url is required") {
		t.Errorf("Error should mention missing Prometheus URL, got: %v", err)
	}
}

func TestValidateInfluxDBConfig(t *testing.T) {
	config := &Config{
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

	err := config.Validate()
	if err != nil {
		t.Errorf("Validate should not return error for valid InfluxDB config, got %v", err)
	}
}

func TestValidateInfluxDB1Config(t *testing.T) {
	config := &Config{
		Backend: "influxdb1",
		InfluxDB1: influxdb1.Config{
			URL:      "http://localhost:8086",
			Username: "admin",
			Password: "password",
			Database: "telegraf",
		},
		Queries: []backend.Query{
			{Name: "Test", Expr: "test_expr"},
		},
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("Validate should not return error for valid InfluxDB v1 config, got %v", err)
	}
}

func TestValidateInfluxDBMissingFields(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		errorMsg string
	}{
		{
			name: "Missing URL",
			config: &Config{
				Backend: "influxdb",
				InfluxDB: influxdb.Config{
					Token:  "test-token",
					Org:    "test-org",
					Bucket: "test-bucket",
				},
				Queries: []backend.Query{{Name: "Test", Expr: "test"}},
			},
			errorMsg: "influxdb.url is required",
		},
		{
			name: "Missing Token",
			config: &Config{
				Backend: "influxdb",
				InfluxDB: influxdb.Config{
					URL:    "http://localhost:8086",
					Org:    "test-org",
					Bucket: "test-bucket",
				},
				Queries: []backend.Query{{Name: "Test", Expr: "test"}},
			},
			errorMsg: "influxdb.token is required",
		},
		{
			name: "Missing Org",
			config: &Config{
				Backend: "influxdb",
				InfluxDB: influxdb.Config{
					URL:    "http://localhost:8086",
					Token:  "test-token",
					Bucket: "test-bucket",
				},
				Queries: []backend.Query{{Name: "Test", Expr: "test"}},
			},
			errorMsg: "influxdb.org is required",
		},
		{
			name: "Missing Bucket",
			config: &Config{
				Backend: "influxdb",
				InfluxDB: influxdb.Config{
					URL:   "http://localhost:8086",
					Token: "test-token",
					Org:   "test-org",
				},
				Queries: []backend.Query{{Name: "Test", Expr: "test"}},
			},
			errorMsg: "influxdb.bucket is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if err == nil {
				t.Error("Validate should return error")
			}

			if !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("Error should contain '%s', got: %v", tt.errorMsg, err)
			}
		})
	}
}

func TestValidateUnsupportedBackend(t *testing.T) {
	config := &Config{
		Backend: "unsupported",
		Queries: []backend.Query{
			{Name: "Test", Expr: "test_metric"},
		},
	}

	err := config.Validate()
	if err == nil {
		t.Error("Validate should return error for unsupported backend")
	}

	if !strings.Contains(err.Error(), "unsupported backend: unsupported") {
		t.Errorf("Error should mention unsupported backend, got: %v", err)
	}
}

func TestValidateInfluxDB1MissingFields(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		errorMsg string
	}{
		{
			name: "Missing URL",
			config: &Config{
				Backend: "influxdb1",
				InfluxDB1: influxdb1.Config{
					Username: "admin",
					Password: "password",
					Database: "telegraf",
				},
				Queries: []backend.Query{{Name: "Test", Expr: "test"}},
			},
			errorMsg: "influxdb1.url is required",
		},
		{
			name: "Missing Database",
			config: &Config{
				Backend: "influxdb1",
				InfluxDB1: influxdb1.Config{
					URL:      "http://localhost:8086",
					Username: "admin",
					Password: "password",
				},
				Queries: []backend.Query{{Name: "Test", Expr: "test"}},
			},
			errorMsg: "influxdb1.database is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if err == nil {
				t.Error("Validate should return error")
			}

			if !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("Error should contain '%s', got: %v", tt.errorMsg, err)
			}
		})
	}
}

func TestValidateMissingQueries(t *testing.T) {
	config := &Config{
		Prometheus: prom.Config{URL: "http://localhost:9090"},
		Queries:    []backend.Query{},
	}

	err := config.Validate()
	if err == nil {
		t.Error("Validate should return error for missing queries")
	}

	if !strings.Contains(err.Error(), "at least one query is required") {
		t.Errorf("Error should mention missing queries, got: %v", err)
	}
}

func TestValidateInvalidQueries(t *testing.T) {
	tests := []struct {
		name     string
		queries  []backend.Query
		errorMsg string
	}{
		{
			name: "Missing query name",
			queries: []backend.Query{
				{Name: "", Expr: "test_metric"},
			},
			errorMsg: "query 0: name is required",
		},
		{
			name: "Missing query expression",
			queries: []backend.Query{
				{Name: "Test", Expr: ""},
			},
			errorMsg: "query 0: expr is required",
		},
		{
			name: "Multiple invalid queries",
			queries: []backend.Query{
				{Name: "Valid", Expr: "test_metric"},
				{Name: "", Expr: "test_metric2"},
			},
			errorMsg: "query 1: name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Prometheus: prom.Config{URL: "http://localhost:9090"},
				Queries:    tt.queries,
			}

			err := config.Validate()
			if err == nil {
				t.Error("Validate should return error for invalid queries")
			}

			if !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("Error should contain '%s', got: %v", tt.errorMsg, err)
			}
		})
	}
}

func TestGetPrometheusConfig(t *testing.T) {
	config := &Config{
		Prometheus: prom.Config{URL: "http://localhost:9090"},
	}

	promConfig := config.GetPrometheusConfig()
	if promConfig.URL != "http://localhost:9090" {
		t.Errorf("Expected Prometheus URL 'http://localhost:9090', got '%s'", promConfig.URL)
	}
}

func TestGetInfluxDBConfig(t *testing.T) {
	config := &Config{
		InfluxDB: influxdb.Config{
			URL:    "http://localhost:8086",
			Token:  "test-token",
			Org:    "test-org",
			Bucket: "test-bucket",
		},
	}

	influxConfig := config.GetInfluxDBConfig()
	if influxConfig.URL != "http://localhost:8086" {
		t.Errorf("Expected InfluxDB URL 'http://localhost:8086', got '%s'", influxConfig.URL)
	}
	if influxConfig.Token != "test-token" {
		t.Errorf("Expected InfluxDB token 'test-token', got '%s'", influxConfig.Token)
	}
}

func TestGetInfluxDB1Config(t *testing.T) {
	config := &Config{
		InfluxDB1: influxdb1.Config{
			URL:      "http://localhost:8086",
			Username: "admin",
			Password: "password",
			Database: "telegraf",
		},
	}

	influx1Config := config.GetInfluxDB1Config()
	if influx1Config.URL != "http://localhost:8086" {
		t.Errorf("Expected InfluxDB v1 URL 'http://localhost:8086', got '%s'", influx1Config.URL)
	}
	if influx1Config.Username != "admin" {
		t.Errorf("Expected InfluxDB v1 username 'admin', got '%s'", influx1Config.Username)
	}
	if influx1Config.Password != "password" {
		t.Errorf("Expected InfluxDB v1 password 'password', got '%s'", influx1Config.Password)
	}
	if influx1Config.Database != "telegraf" {
		t.Errorf("Expected InfluxDB v1 database 'telegraf', got '%s'", influx1Config.Database)
	}
}
