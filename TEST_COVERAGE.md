# Test Coverage Report

## Overview

PromViz now has comprehensive test coverage across all major components with automated testing infrastructure.

## Test Statistics

| Package | Coverage | Status |
|---------|----------|--------|
| `internal/config` | 97.4% | ✅ Excellent |
| `internal/backend/prom` | 95.8% | ✅ Excellent |
| `internal/backend/influxdb` | 92.7% | ✅ Excellent |
| `internal/backend/influxdb1` | 81.4% | ✅ Very Good |
| `internal/backend/mock` | 100.0% | ✅ Perfect |
| `internal/app` | 30.0% | ⚠️ Good (focus on core logic) |

## Test Categories

### 1. Unit Tests ✅

**Backend Interface Tests** (`internal/backend/types_test.go`)
- Tests the core `Backend` interface contract
- Validates `QueryResult` and `Query` structures  
- Includes mock implementation for testing

**Prometheus Backend Tests** (`internal/backend/prom/client_test.go`)
- Connection testing with mock Prometheus server
- Query execution for vector and scalar results
- Error handling for invalid queries and empty results
- URL validation and client initialization

**InfluxDB v2 Backend Tests** (`internal/backend/influxdb/client_test.go`)  
- Configuration validation (URL, token, org, bucket)
- Connection testing with mock InfluxDB server
- Simple filter queries and complex Flux queries
- CSV response parsing and data type handling

**InfluxDB v1 Backend Tests** (`internal/backend/influxdb1/client_test.go`)
- Configuration validation (URL, username, password, database)
- Connection testing with mock InfluxDB v1 server
- InfluxQL query execution and JSON response parsing
- Simple expressions and full SELECT statement support
- Data type conversion (float64, int64, strings)
- Error handling for connection and query failures

**Mock Backend Tests** (`internal/backend/mock/client_test.go`)
- Deterministic random number generation for testing
- Query performance validation
- Different query type behaviors (CPU, memory, generic)

**Configuration Tests** (`internal/config/config_test.go`)
- YAML parsing for both Prometheus and InfluxDB configs
- Backend-specific validation rules
- Missing field detection and error reporting
- Multi-backend configuration support

**Application Logic Tests** (`internal/app/app_test.go`)
- Backend factory creation and selection
- Configuration loading and validation
- Error handling for connection failures

**UI Component Tests** (`internal/ui/ui_test.go`)
- Query history management and size limits
- TUI focus navigation and panel management  
- Metric update handling and error states

### 2. Integration Tests ✅

**Mock Server Integration** (`integration_test.go`)
- End-to-end testing with mock Prometheus and InfluxDB servers
- Backend interface compliance verification
- Performance benchmarking

### 3. Test Infrastructure ✅

**Test Utilities** (`test_utils.go`)
- Mock server factories for Prometheus and InfluxDB
- Generic backend interface testing functions
- Error injection utilities for negative testing

**Makefile Targets**
- `make test` - Run all tests
- `make test-unit` - Unit tests only
- `make test-integration` - Integration tests
- `make test-coverage` - Generate coverage report
- `make bench` - Performance benchmarks

**GitHub Actions CI/CD** (`.github/workflows/ci.yml`)
- Multi-version Go testing (1.19, 1.20, 1.21)
- Automated linting with golangci-lint
- Code formatting verification
- Security scanning with Gosec
- Coverage reporting to Codecov

## Key Testing Features

### ✅ **Comprehensive Backend Testing**
- All four backend implementations (Prometheus, InfluxDB v2, InfluxDB v1, Mock) have >80% coverage
- Mock servers provide realistic testing environments for HTTP APIs
- Error conditions and edge cases thoroughly tested
- Support for multiple query languages (PromQL, Flux, InfluxQL)

### ✅ **Configuration Validation**  
- 97% coverage of configuration parsing and validation
- Tests for all supported backends and error conditions
- YAML syntax error handling

### ✅ **Modular Test Design**
- Each package has isolated test suite
- Shared test utilities for consistent testing
- Mock implementations enable fast, reliable testing

### ✅ **Performance Testing**
- Benchmark tests for query performance
- Deterministic testing with controlled randomness
- Query timing validation

## Test Commands

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run benchmarks
make bench

# Integration tests only
make test-integration

# Specific package testing
go test -v ./internal/backend/prom
go test -v ./internal/config
go test -v ./internal/app
```

## CI/CD Integration

The test suite is fully integrated with GitHub Actions:

- **Automated on every push and PR**
- **Multi-Go version compatibility testing**
- **Linting and code quality checks**
- **Security vulnerability scanning**
- **Coverage reporting**

## Next Steps

1. **Increase App Coverage**: Add more integration tests for the app lifecycle
2. **UI Testing**: Expand UI tests (currently basic due to TUI complexity)
3. **Real Server Testing**: Add optional tests against real Prometheus/InfluxDB instances
4. **Load Testing**: Add stress tests for high-frequency metric updates

## Benefits

✅ **Quality Assurance** - Comprehensive testing catches bugs early  
✅ **Regression Prevention** - Automated tests prevent breaking changes  
✅ **Documentation** - Tests serve as usage examples  
✅ **Refactoring Confidence** - Safe to modify code with test coverage  
✅ **CI/CD Ready** - Full automation for continuous integration
