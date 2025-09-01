package influxdb

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConfigGetURL(t *testing.T) {
	config := &Config{
		URL:    "http://influxdb:8086",
		Token:  "test-token",
		Org:    "test-org",
		Bucket: "test-bucket",
	}

	url := config.GetURL()
	expected := "http://influxdb:8086"

	if url != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, url)
	}
}

func TestNewClient(t *testing.T) {
	config := &Config{
		URL:    "http://localhost:8086",
		Token:  "test-token",
		Org:    "test-org",
		Bucket: "test-bucket",
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
		t.Error("InfluxDB client should be initialized")
	}

	if client.queryAPI == nil {
		t.Error("InfluxDB query API should be initialized")
	}
}

func TestNewClientMissingURL(t *testing.T) {
	config := &Config{
		Token:  "test-token",
		Org:    "test-org",
		Bucket: "test-bucket",
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

func TestNewClientMissingToken(t *testing.T) {
	config := &Config{
		URL:    "http://localhost:8086",
		Org:    "test-org",
		Bucket: "test-bucket",
	}

	client, err := NewClient(config)

	if err == nil {
		t.Error("NewClient should return error for missing token")
	}

	if client != nil {
		t.Error("NewClient should return nil client on error")
	}

	if !strings.Contains(err.Error(), "token is required") {
		t.Errorf("Error should mention missing token, got: %v", err)
	}
}

func TestNewClientMissingOrg(t *testing.T) {
	config := &Config{
		URL:    "http://localhost:8086",
		Token:  "test-token",
		Bucket: "test-bucket",
	}

	client, err := NewClient(config)

	if err == nil {
		t.Error("NewClient should return error for missing organization")
	}

	if client != nil {
		t.Error("NewClient should return nil client on error")
	}

	if !strings.Contains(err.Error(), "organization is required") {
		t.Errorf("Error should mention missing organization, got: %v", err)
	}
}

func TestNewClientMissingBucket(t *testing.T) {
	config := &Config{
		URL:   "http://localhost:8086",
		Token: "test-token",
		Org:   "test-org",
	}

	client, err := NewClient(config)

	if err == nil {
		t.Error("NewClient should return error for missing bucket")
	}

	if client != nil {
		t.Error("NewClient should return nil client on error")
	}

	if !strings.Contains(err.Error(), "bucket is required") {
		t.Errorf("Error should mention missing bucket, got: %v", err)
	}
}

func TestClientName(t *testing.T) {
	config := &Config{
		URL:    "http://localhost:8086",
		Token:  "test-token",
		Org:    "test-org",
		Bucket: "test-bucket",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	name := client.Name()
	expected := "influxdb"

	if name != expected {
		t.Errorf("Expected name '%s', got '%s'", expected, name)
	}
}

func TestClientClose(t *testing.T) {
	config := &Config{
		URL:    "http://localhost:8086",
		Token:  "test-token",
		Org:    "test-org",
		Bucket: "test-bucket",
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

// Mock InfluxDB server for testing
func createMockInfluxDBServer(response string, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Different responses based on endpoint
		if strings.Contains(r.URL.Path, "/api/v2/query") {
			w.Header().Set("Content-Type", "application/csv")
			w.WriteHeader(statusCode)
			w.Write([]byte(response))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestClientConnect(t *testing.T) {
	// Mock successful CSV response for connection test
	mockResponse := `#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,0,2023-01-01T00:00:00Z,2023-01-01T01:00:00Z,2023-01-01T00:30:00Z,42.5,usage_user,cpu,cpu-total,server1
`

	server := createMockInfluxDBServer(mockResponse, http.StatusOK)
	defer server.Close()

	config := &Config{
		URL:    server.URL,
		Token:  "test-token",
		Org:    "test-org",
		Bucket: "test-bucket",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx := context.Background()
	err = client.Connect(ctx)
	if err != nil {
		t.Errorf("Connect should not return error, got %v", err)
	}
}

func TestClientConnectFailure(t *testing.T) {
	// Use non-existent server
	config := &Config{
		URL:    "http://localhost:1",
		Token:  "test-token",
		Org:    "test-org",
		Bucket: "test-bucket",
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

	if !strings.Contains(err.Error(), "failed to connect to InfluxDB") {
		t.Errorf("Error should mention InfluxDB connection failure, got: %v", err)
	}
}

func TestClientQuerySimpleFilter(t *testing.T) {
	// Mock successful CSV response
	mockResponse := `#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,0,2023-01-01T00:00:00Z,2023-01-01T01:00:00Z,2023-01-01T00:30:00Z,75.8,usage_user,cpu,cpu-total,server1
`

	server := createMockInfluxDBServer(mockResponse, http.StatusOK)
	defer server.Close()

	config := &Config{
		URL:    server.URL,
		Token:  "test-token",
		Org:    "test-org",
		Bucket: "test-bucket",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx := context.Background()
	timeSeries, err := client.QueryTimeSeries(ctx, `r._measurement == "cpu" and r._field == "usage_user"`)

	if err != nil {
		t.Fatalf("Query should not return error, got %v", err)
	}

	expected := 75.8
	if timeSeries == nil || len(timeSeries.Points) == 0 {
		t.Fatal("QueryTimeSeries should return time series data")
	}
	value := timeSeries.Points[0].Value
	if value != expected {
		t.Errorf("Expected value %f, got %f", expected, value)
	}
}

func TestClientQueryFullFlux(t *testing.T) {
	// Mock successful CSV response
	mockResponse := `#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,0,2023-01-01T00:00:00Z,2023-01-01T01:00:00Z,2023-01-01T00:30:00Z,42.1,usage_user,cpu,cpu-total,server1
`

	server := createMockInfluxDBServer(mockResponse, http.StatusOK)
	defer server.Close()

	config := &Config{
		URL:    server.URL,
		Token:  "test-token",
		Org:    "test-org",
		Bucket: "test-bucket",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx := context.Background()
	fullQuery := fmt.Sprintf(`from(bucket: "%s") |> range(start: -5m) |> filter(fn: (r) => r._measurement == "cpu") |> last()`, config.Bucket)
	timeSeries, err := client.QueryTimeSeries(ctx, fullQuery)

	if err != nil {
		t.Fatalf("Query should not return error, got %v", err)
	}

	expected := 42.1
	if timeSeries == nil || len(timeSeries.Points) == 0 {
		t.Fatal("QueryTimeSeries should return time series data")
	}
	value := timeSeries.Points[0].Value
	if value != expected {
		t.Errorf("Expected value %f, got %f", expected, value)
	}
}

func TestClientQueryNoData(t *testing.T) {
	// Mock empty CSV response
	mockResponse := `#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
`

	server := createMockInfluxDBServer(mockResponse, http.StatusOK)
	defer server.Close()

	config := &Config{
		URL:    server.URL,
		Token:  "test-token",
		Org:    "test-org",
		Bucket: "test-bucket",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx := context.Background()
	timeSeries, err := client.QueryTimeSeries(ctx, `r._measurement == "nonexistent"`)

	if err != nil {
		t.Errorf("QueryTimeSeries should not return error for no data, got %v", err)
	}

	if timeSeries == nil {
		t.Error("QueryTimeSeries should return valid result even if empty")
	}

	if len(timeSeries.Points) != 0 {
		t.Errorf("Expected empty points for no data, got %d points", len(timeSeries.Points))
	}
}

func TestClientQueryIntegerValue(t *testing.T) {
	// Mock CSV response with integer value
	mockResponse := `#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,0,2023-01-01T00:00:00Z,2023-01-01T01:00:00Z,2023-01-01T00:30:00Z,100,count,requests,total,server1
`

	server := createMockInfluxDBServer(mockResponse, http.StatusOK)
	defer server.Close()

	config := &Config{
		URL:    server.URL,
		Token:  "test-token",
		Org:    "test-org",
		Bucket: "test-bucket",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx := context.Background()
	timeSeries, err := client.QueryTimeSeries(ctx, `r._measurement == "requests"`)

	if err != nil {
		t.Fatalf("Query should not return error, got %v", err)
	}

	expected := float64(100)
	if timeSeries == nil || len(timeSeries.Points) == 0 {
		t.Fatal("QueryTimeSeries should return time series data")
	}
	value := timeSeries.Points[0].Value
	if value != expected {
		t.Errorf("Expected value %f, got %f", expected, value)
	}
}

func TestClientQueryError(t *testing.T) {
	// Mock error response
	mockResponse := `{"code":"invalid","message":"compilation failed: error at @1:8-1:9: undefined identifier \"invalid\""}`

	server := createMockInfluxDBServer(mockResponse, http.StatusBadRequest)
	defer server.Close()

	config := &Config{
		URL:    server.URL,
		Token:  "test-token",
		Org:    "test-org",
		Bucket: "test-bucket",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx := context.Background()
	_, err = client.QueryTimeSeries(ctx, "invalid flux query")

	if err == nil {
		t.Error("Query should return error for invalid query")
	}

	if !strings.Contains(err.Error(), "query failed") {
		t.Errorf("Error should mention query failure, got: %v", err)
	}
}
