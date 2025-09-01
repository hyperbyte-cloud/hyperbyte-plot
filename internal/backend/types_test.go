package backend

import (
	"context"
	"testing"
	"time"
)

// TestDataPoint tests the DataPoint struct
func TestDataPoint(t *testing.T) {
	timestamp := time.Now()
	value := 42.5

	result := DataPoint{
		Timestamp: timestamp,
		Value:     value,
	}

	if result.Timestamp != timestamp {
		t.Errorf("Expected timestamp %v, got %v", timestamp, result.Timestamp)
	}

	if result.Value != value {
		t.Errorf("Expected value %f, got %f", value, result.Value)
	}
}

// TestTimeSeriesResult tests the TimeSeriesResult struct
func TestTimeSeriesResult(t *testing.T) {
	points := []DataPoint{
		{Timestamp: time.Now(), Value: 1.0},
		{Timestamp: time.Now().Add(time.Minute), Value: 2.0},
	}

	result := TimeSeriesResult{Points: points}

	if len(result.Points) != 2 {
		t.Errorf("Expected 2 points, got %d", len(result.Points))
	}

	if result.Points[0].Value != 1.0 {
		t.Errorf("Expected first point value 1.0, got %f", result.Points[0].Value)
	}

	if result.Points[1].Value != 2.0 {
		t.Errorf("Expected second point value 2.0, got %f", result.Points[1].Value)
	}
}

// TestQuery tests the Query struct
func TestQuery(t *testing.T) {
	query := Query{
		Name: "Test Query",
		Expr: "test_expression",
	}

	if query.Name != "Test Query" {
		t.Errorf("Expected name 'Test Query', got '%s'", query.Name)
	}

	if query.Expr != "test_expression" {
		t.Errorf("Expected expr 'test_expression', got '%s'", query.Expr)
	}
}

// MockBackend implements Backend interface for testing
type MockBackend struct {
	connectFunc         func(ctx context.Context) error
	queryTimeSeriesFunc func(ctx context.Context, expr string) (*TimeSeriesResult, error)
	closeFunc           func() error
	nameFunc            func() string
}

func (m *MockBackend) Connect(ctx context.Context) error {
	if m.connectFunc != nil {
		return m.connectFunc(ctx)
	}
	return nil
}

func (m *MockBackend) QueryTimeSeries(ctx context.Context, expr string) (*TimeSeriesResult, error) {
	if m.queryTimeSeriesFunc != nil {
		return m.queryTimeSeriesFunc(ctx, expr)
	}
	// Return default time series with one data point
	return &TimeSeriesResult{
		Points: []DataPoint{{Timestamp: time.Now(), Value: 0}},
	}, nil
}

func (m *MockBackend) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func (m *MockBackend) Name() string {
	if m.nameFunc != nil {
		return m.nameFunc()
	}
	return "mock"
}

// TestBackendInterface tests that MockBackend implements Backend interface
func TestBackendInterface(t *testing.T) {
	var backend Backend = &MockBackend{}

	ctx := context.Background()

	// Test Connect
	err := backend.Connect(ctx)
	if err != nil {
		t.Errorf("Connect should not return error, got %v", err)
	}

	// Test QueryTimeSeries
	timeSeries, err := backend.QueryTimeSeries(ctx, "test")
	if err != nil {
		t.Errorf("QueryTimeSeries should not return error, got %v", err)
	}
	if timeSeries == nil {
		t.Error("QueryTimeSeries should not return nil")
	} else if len(timeSeries.Points) == 0 {
		t.Error("QueryTimeSeries should return at least one data point")
	} else if timeSeries.Points[0].Value != 0 {
		t.Errorf("Expected default value 0, got %f", timeSeries.Points[0].Value)
	}

	// Test Name
	name := backend.Name()
	if name != "mock" {
		t.Errorf("Expected name 'mock', got '%s'", name)
	}

	// Test Close
	err = backend.Close()
	if err != nil {
		t.Errorf("Close should not return error, got %v", err)
	}
}
