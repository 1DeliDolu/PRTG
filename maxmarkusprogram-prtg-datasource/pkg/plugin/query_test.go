package plugin

import (
	"context"

	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// ✅ Mock API sunucusu oluştur
func setupMockAPI(responseBody string, statusCode int) (*httptest.Server, *Api) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		fmt.Fprint(w, responseBody)
	})

	server := httptest.NewServer(mux)
	api := NewApi(server.URL, "test-api-key", 10*time.Second, 10*time.Second)
	return server, api
}

// ✅ QueryData test: Metric sorgusu
func TestQueryData_Metrics(t *testing.T) {
	mockResponse := `{"histdata": [{"datetime": "2025-02-15T12:00:00Z", "value": 78.9}]}`
	server, api := setupMockAPI(mockResponse, http.StatusOK)
	defer server.Close()

	ds := &Datasource{api: api}
	query := backend.DataQuery{
		RefID: "A",
		JSON:  []byte(`{"queryType":"metrics","objectId":"1234","channel":"value"}`),
		TimeRange: backend.TimeRange{
			From: time.Now().Add(-24 * time.Hour),
			To:   time.Now(),
		},
	}

	resp := ds.query(context.Background(), backend.PluginContext{}, query)
	if len(resp.Frames) == 0 {
		t.Errorf("Expected at least 1 frame, got 0")
	}
}

// ✅ QueryData test: Group property sorgusu
func TestQueryData_Groups(t *testing.T) {
	mockResponse := `{"groups": [{"group": "Network Devices", "datetime": "2025-02-15T12:00:00Z", "status": "Up"}]}`
	server, api := setupMockAPI(mockResponse, http.StatusOK)
	defer server.Close()

	ds := &Datasource{api: api}
	query := backend.DataQuery{
		RefID: "B",
		JSON:  []byte(`{"queryType":"text","property":"group","group":"Network Devices","filterProperty":"status"}`),
	}

	resp := ds.query(context.Background(), backend.PluginContext{}, query)
	if len(resp.Frames) == 0 {
		t.Errorf("Expected at least 1 frame, got 0")
	}
}

// ✅ QueryData test: Device property sorgusu
func TestQueryData_Devices(t *testing.T) {
	mockResponse := `{"devices": [{"device": "Router", "datetime": "2025-02-15T12:00:00Z", "status": "Warning"}]}`
	server, api := setupMockAPI(mockResponse, http.StatusOK)
	defer server.Close()

	ds := &Datasource{api: api}
	query := backend.DataQuery{
		RefID: "C",
		JSON:  []byte(`{"queryType":"text","property":"device","device":"Router","filterProperty":"status"}`),
	}

	resp := ds.query(context.Background(), backend.PluginContext{}, query)
	if len(resp.Frames) == 0 {
		t.Errorf("Expected at least 1 frame, got 0")
	}
}

// ✅ QueryData test: Sensor property sorgusu
func TestQueryData_Sensors(t *testing.T) {
	mockResponse := `{"sensors": [{"sensor": "CPU Load", "datetime": "2025-02-15T12:00:00Z", "status": "Critical"}]}`
	server, api := setupMockAPI(mockResponse, http.StatusOK)
	defer server.Close()

	ds := &Datasource{api: api}
	query := backend.DataQuery{
		RefID: "D",
		JSON:  []byte(`{"queryType":"text","property":"sensor","sensor":"CPU Load","filterProperty":"status"}`),
	}

	resp := ds.query(context.Background(), backend.PluginContext{}, query)
	if len(resp.Frames) == 0 {
		t.Errorf("Expected at least 1 frame, got 0")
	}
}

// ✅ QueryData test: Hatalı JSON
func TestQueryData_InvalidJSON(t *testing.T) {
	ds := &Datasource{}
	query := backend.DataQuery{
		RefID: "E",
		JSON:  []byte(`{"queryType":`), // Bozuk JSON
	}

	resp := ds.query(context.Background(), backend.PluginContext{}, query)
	if len(resp.Frames) != 0 {
		t.Errorf("Expected 0 frames, got %d", len(resp.Frames))
	}
}

// ✅ QueryData test: Geçersiz QueryType
func TestQueryData_InvalidQueryType(t *testing.T) {
	ds := &Datasource{}
	query := backend.DataQuery{
		RefID: "F",
		JSON:  []byte(`{"queryType":"invalidType"}`),
	}

	resp := ds.query(context.Background(), backend.PluginContext{}, query)
	if len(resp.Frames) != 0 {
		t.Errorf("Expected 0 frames, got %d", len(resp.Frames))
	}
}

// ✅ isValidPropertyType test
func TestIsValidPropertyType(t *testing.T) {
	ds := &Datasource{}
	valid := []string{"group", "device", "sensor", "status", "message", "priority", "tags"}
	invalid := []string{"invalid", "unknown", "badproperty"}

	for _, prop := range valid {
		if !ds.isValidPropertyType(prop) {
			t.Errorf("Expected valid property for %s", prop)
		}
	}

	for _, prop := range invalid {
		if ds.isValidPropertyType(prop) {
			t.Errorf("Expected invalid property for %s", prop)
		}
	}
}

// ✅ GetPropertyValue test
func TestGetPropertyValue(t *testing.T) {
	ds := &Datasource{}
	testObj := struct {
		Status     string  `json:"status"`
		Priority   int     `json:"priority"`
		Percentage float64 `json:"percentage"`
		RawValue   bool    `json:"raw_value"`
	}{
		Status:     "OK",
		Priority:   5,
		Percentage: 87.4,
		RawValue:   true,
	}

	tests := []struct {
		property string
		expected string
	}{
		{"status", "OK"},
		{"priority", "5"},
		{"percentage", "87.4"},
		{"raw_value", "true"},
		{"invalid_property", "Unknown"},
	}

	for _, tt := range tests {
		result := ds.GetPropertyValue(tt.property, testObj)
		if result != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, result)
		}
	}
}

// ✅ cleanMessageHTML test
func TestCleanMessageHTML(t *testing.T) {
	message := `<div class="status">OK</div><div class="moreicon"></div>`
	expected := "OK"
	result := cleanMessageHTML(message)
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}
