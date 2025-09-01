package ui

import (
	"fmt"
	"testing"
	"time"

	"promviz/internal/backend"
)

func TestQueryHistory(t *testing.T) {
	history := &QueryHistory{
		Name:       "Test Query",
		TimeSeries: &backend.TimeSeriesResult{Points: []backend.DataPoint{}},
		LastError:  nil,
	}

	// Test initial state
	if history.Name != "Test Query" {
		t.Errorf("Expected name 'Test Query', got '%s'", history.Name)
	}

	if history.TimeSeries == nil {
		t.Error("TimeSeries should be initialized")
	}

	if len(history.TimeSeries.Points) != 0 {
		t.Errorf("Expected empty points, got %d", len(history.TimeSeries.Points))
	}

	if history.LastError != nil {
		t.Errorf("Expected no error, got %v", history.LastError)
	}
}

func TestQueryHistoryWithData(t *testing.T) {
	points := []backend.DataPoint{
		{Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), Value: 10.5},
		{Timestamp: time.Date(2023, 1, 1, 12, 1, 0, 0, time.UTC), Value: 15.2},
		{Timestamp: time.Date(2023, 1, 1, 12, 2, 0, 0, time.UTC), Value: 20.8},
	}

	history := &QueryHistory{
		Name:       "Test Query",
		TimeSeries: &backend.TimeSeriesResult{Points: points},
		LastError:  nil,
	}

	// Verify data
	if len(history.TimeSeries.Points) != 3 {
		t.Errorf("Expected 3 points, got %d", len(history.TimeSeries.Points))
	}

	expectedValues := []float64{10.5, 15.2, 20.8}
	for i, point := range history.TimeSeries.Points {
		if point.Value != expectedValues[i] {
			t.Errorf("Expected value %f at index %d, got %f", expectedValues[i], i, point.Value)
		}
	}
}

func TestNewTUI(t *testing.T) {
	queries := []backend.Query{
		{Name: "Query 1", Expr: "metric1"},
		{Name: "Query 2", Expr: "metric2"},
		{Name: "Query 3", Expr: "metric3"},
	}

	quitCalled := false
	onQuit := func() {
		quitCalled = true
	}

	tui := NewTUI(queries, onQuit)

	if tui == nil {
		t.Fatal("NewTUI should not return nil")
	}

	if tui.app == nil {
		t.Error("TUI app should be initialized")
	}

	if tui.flex == nil {
		t.Error("TUI flex container should be initialized")
	}

	if tui.scrollView == nil {
		t.Error("TUI scroll view should be initialized")
	}

	if len(tui.panels) != len(queries) {
		t.Errorf("Expected %d panels, got %d", len(queries), len(tui.panels))
	}

	if len(tui.histories) != len(queries) {
		t.Errorf("Expected %d histories, got %d", len(queries), len(tui.histories))
	}

	// Test that histories are properly initialized
	for i, query := range queries {
		if tui.histories[i].Name != query.Name {
			t.Errorf("Expected history name '%s', got '%s'", query.Name, tui.histories[i].Name)
		}

		if tui.histories[i].TimeSeries == nil {
			t.Error("TimeSeries should be initialized")
		}

		if len(tui.histories[i].TimeSeries.Points) != 0 {
			t.Errorf("Expected empty points, got %d", len(tui.histories[i].TimeSeries.Points))
		}

		if tui.histories[i].LastError != nil {
			t.Errorf("Expected no initial error, got %v", tui.histories[i].LastError)
		}
	}

	// Test quit function is stored
	if tui.onQuit == nil {
		t.Error("Quit function should be stored")
	}

	// Test quit function works
	tui.onQuit()
	if !quitCalled {
		t.Error("Quit function should be called")
	}
}

func TestTUIFocus(t *testing.T) {
	queries := []backend.Query{
		{Name: "Query 1", Expr: "metric1"},
		{Name: "Query 2", Expr: "metric2"},
		{Name: "Query 3", Expr: "metric3"},
	}

	tui := NewTUI(queries, nil)

	// Test initial focus
	if tui.focusIndex != 0 {
		t.Errorf("Expected initial focus index 0, got %d", tui.focusIndex)
	}

	// Test focus next
	tui.focusNext()
	if tui.focusIndex != 1 {
		t.Errorf("Expected focus index 1 after focusNext, got %d", tui.focusIndex)
	}

	tui.focusNext()
	if tui.focusIndex != 2 {
		t.Errorf("Expected focus index 2 after second focusNext, got %d", tui.focusIndex)
	}

	// Test wraparound
	tui.focusNext()
	if tui.focusIndex != 0 {
		t.Errorf("Expected focus index 0 after wraparound, got %d", tui.focusIndex)
	}

	// Test focus previous
	tui.focusPrev()
	if tui.focusIndex != 2 {
		t.Errorf("Expected focus index 2 after focusPrev from 0, got %d", tui.focusIndex)
	}

	tui.focusPrev()
	if tui.focusIndex != 1 {
		t.Errorf("Expected focus index 1 after second focusPrev, got %d", tui.focusIndex)
	}
}

func TestTUIFocusEmptyPanels(t *testing.T) {
	tui := NewTUI([]backend.Query{}, nil)

	// Test that focus methods don't panic with empty panels
	initialIndex := tui.focusIndex
	tui.focusNext()
	if tui.focusIndex != initialIndex {
		t.Errorf("Focus index should not change with empty panels")
	}

	tui.focusPrev()
	if tui.focusIndex != initialIndex {
		t.Errorf("Focus index should not change with empty panels")
	}
}

func TestUpdateTimeSeries(t *testing.T) {
	queries := []backend.Query{
		{Name: "Query 1", Expr: "metric1"},
		{Name: "Query 2", Expr: "metric2"},
	}

	tui := NewTUI(queries, nil)

	// Test valid update
	timeSeries := &backend.TimeSeriesResult{
		Points: []backend.DataPoint{
			{Timestamp: time.Now(), Value: 42.5},
			{Timestamp: time.Now().Add(time.Minute), Value: 45.0},
		},
	}

	// This should not panic
	tui.UpdateTimeSeries(0, timeSeries, nil)

	// Check that the time series was updated
	if tui.histories[0].TimeSeries == nil {
		t.Error("TimeSeries should not be nil after update")
	}

	if len(tui.histories[0].TimeSeries.Points) != 2 {
		t.Errorf("Expected 2 points in time series, got %d", len(tui.histories[0].TimeSeries.Points))
	}

	if tui.histories[0].TimeSeries.Points[0].Value != 42.5 {
		t.Errorf("Expected first value 42.5, got %f", tui.histories[0].TimeSeries.Points[0].Value)
	}

	if tui.histories[0].TimeSeries.Points[1].Value != 45.0 {
		t.Errorf("Expected second value 45.0, got %f", tui.histories[0].TimeSeries.Points[1].Value)
	}

	// Test error update
	testError := fmt.Errorf("test error")
	tui.UpdateTimeSeries(1, nil, testError)

	// Should store the error
	if tui.histories[1].LastError == nil {
		t.Error("Expected error to be stored")
	}

	if tui.histories[1].LastError.Error() != "test error" {
		t.Errorf("Expected error 'test error', got '%v'", tui.histories[1].LastError)
	}

	// Test invalid index (should not panic)
	tui.UpdateTimeSeries(-1, timeSeries, nil)
	tui.UpdateTimeSeries(10, timeSeries, nil)
}

func TestUpdateMetricCompatibility(t *testing.T) {
	queries := []backend.Query{
		{Name: "Query 1", Expr: "metric1"},
	}

	tui := NewTUI(queries, nil)

	// Test deprecated UpdateMetric method for backward compatibility
	dataPoint := backend.DataPoint{
		Timestamp: time.Now(),
		Value:     42.5,
	}

	// This should not panic
	tui.UpdateMetric(0, dataPoint, nil)

	// Check that the data point was converted to time series
	if tui.histories[0].TimeSeries == nil {
		t.Error("TimeSeries should not be nil after update")
	}

	if len(tui.histories[0].TimeSeries.Points) != 1 {
		t.Errorf("Expected 1 point in time series, got %d", len(tui.histories[0].TimeSeries.Points))
	}

	if tui.histories[0].TimeSeries.Points[0].Value != 42.5 {
		t.Errorf("Expected value 42.5, got %f", tui.histories[0].TimeSeries.Points[0].Value)
	}
}

func TestUpdateTimeSeriesWithEmptyData(t *testing.T) {
	queries := []backend.Query{
		{Name: "Query 1", Expr: "metric1"},
	}

	tui := NewTUI(queries, nil)

	// Test update with empty time series
	emptyTimeSeries := &backend.TimeSeriesResult{Points: []backend.DataPoint{}}
	tui.UpdateTimeSeries(0, emptyTimeSeries, nil)

	// Should handle empty data gracefully
	if tui.histories[0].TimeSeries == nil {
		t.Error("TimeSeries should not be nil")
	}

	if len(tui.histories[0].TimeSeries.Points) != 0 {
		t.Errorf("Expected 0 points, got %d", len(tui.histories[0].TimeSeries.Points))
	}
}
