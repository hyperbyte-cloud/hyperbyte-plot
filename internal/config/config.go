package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"promviz/internal/backend"
	"promviz/internal/backend/influxdb"
	"promviz/internal/backend/influxdb1"
	"promviz/internal/backend/mock"
	"promviz/internal/backend/prom"
)

// Config represents the complete application configuration
type Config struct {
	Backend    string           `yaml:"backend"` // "prometheus", "influxdb", "influxdb1", "mock", etc.
	Prometheus prom.Config      `yaml:"prometheus,omitempty"`
	InfluxDB   influxdb.Config  `yaml:"influxdb,omitempty"`
	InfluxDB1  influxdb1.Config `yaml:"influxdb1,omitempty"`
	Mock       mock.Config      `yaml:"mock,omitempty"`
	Queries    []backend.Query  `yaml:"queries"`
}

// LoadConfig loads and validates configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Default to Prometheus if no backend specified
	if c.Backend == "" {
		c.Backend = "prometheus"
	}

	// Validate backend-specific configuration
	switch c.Backend {
	case "prometheus":
		if c.Prometheus.URL == "" {
			return fmt.Errorf("prometheus.url is required")
		}
	case "influxdb":
		if c.InfluxDB.URL == "" {
			return fmt.Errorf("influxdb.url is required")
		}
		if c.InfluxDB.Token == "" {
			return fmt.Errorf("influxdb.token is required")
		}
		if c.InfluxDB.Org == "" {
			return fmt.Errorf("influxdb.org is required")
		}
		if c.InfluxDB.Bucket == "" {
			return fmt.Errorf("influxdb.bucket is required")
		}
	case "influxdb1":
		if c.InfluxDB1.URL == "" {
			return fmt.Errorf("influxdb1.url is required")
		}
		if c.InfluxDB1.Database == "" {
			return fmt.Errorf("influxdb1.database is required")
		}
	case "mock":
		// Mock backend has no required configuration
	default:
		return fmt.Errorf("unsupported backend: %s (supported: prometheus, influxdb, influxdb1, mock)", c.Backend)
	}

	if len(c.Queries) == 0 {
		return fmt.Errorf("at least one query is required")
	}

	for i, query := range c.Queries {
		if query.Name == "" {
			return fmt.Errorf("query %d: name is required", i)
		}
		if query.Expr == "" {
			return fmt.Errorf("query %d: expr is required", i)
		}
	}

	return nil
}

// GetPrometheusConfig returns the Prometheus configuration
func (c *Config) GetPrometheusConfig() *prom.Config {
	return &c.Prometheus
}

// GetInfluxDBConfig returns the InfluxDB configuration
func (c *Config) GetInfluxDBConfig() *influxdb.Config {
	return &c.InfluxDB
}

// GetInfluxDB1Config returns the InfluxDB v1 configuration
func (c *Config) GetInfluxDB1Config() *influxdb1.Config {
	return &c.InfluxDB1
}

// GetMockConfig returns the mock configuration
func (c *Config) GetMockConfig() *mock.Config {
	return &c.Mock
}
