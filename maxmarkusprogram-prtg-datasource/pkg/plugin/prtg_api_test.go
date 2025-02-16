package plugin

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"
)

// ✅ Running API tests with Mock HTTP Server
func setupMockServer(responseBody string, statusCode int) (*httptest.Server, *Api) {
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

// ✅ Tests if the API URL is built correctly
func TestBuildApiUrl(t *testing.T) {
	api := NewApi("http://localhost", "test-api-key", 10*time.Second, 10*time.Second)

	params := map[string]string{"id": "123"}
	apiUrl, err := api.buildApiUrl("getSensor", params)
	if err != nil {
		t.Fatalf("Failed to build API URL: %v", err)
	}

	parsedUrl, _ := url.Parse(apiUrl)
	if parsedUrl.Query().Get("apitoken") != "test-api-key" {
		t.Errorf("API key missing in URL")
	}
	if parsedUrl.Query().Get("id") != "123" {
		t.Errorf("Expected id=123, got %v", parsedUrl.Query().Get("id"))
	}
}

// ✅ Tests for API response handling
func TestGetStatusList(t *testing.T) {
	server, api := setupMockServer(`{"prtgversion": "21.2.68.1492"}`, http.StatusOK)
	defer server.Close()

	status, err := api.GetStatusList()
	if err != nil {
		t.Fatalf("GetStatusList() failed: %v", err)
	}
	if status.PrtgVersion != "21.2.68.1492" {
		t.Errorf("Expected status '21.2.68.1492', got: %v", status.PrtgVersion)
	}

}

// ✅ Test fetching groups
func TestGetGroups(t *testing.T) {
	mockResponse := `{"groups": [{"group": "Network Devices"}]}`
	server, api := setupMockServer(mockResponse, http.StatusOK)
	defer server.Close()

	groups, err := api.GetGroups()
	if err != nil {
		t.Fatalf("GetGroups() failed: %v", err)
	}
	if len(groups.Groups) == 0 {
		t.Errorf("Expected at least 1 group, got 0")
	}
}

// ✅ Test fetching devices
func TestGetDevices(t *testing.T) {
	mockResponse := `{"devices": [{"device": "Router"}]}`
	server, api := setupMockServer(mockResponse, http.StatusOK)
	defer server.Close()

	devices, err := api.GetDevices()
	if err != nil {
		t.Fatalf("GetDevices() failed: %v", err)
	}
	if len(devices.Devices) == 0 {
		t.Errorf("Expected at least 1 device, got 0")
	}

}

// ✅ Test fetching sensors
func TestGetSensors(t *testing.T) {
	mockResponse := `{"sensors": [{"sensor": "CPU Load"}]}`
	server, api := setupMockServer(mockResponse, http.StatusOK)
	defer server.Close()

	sensors, err := api.GetSensors()
	if err != nil {
		t.Fatalf("GetSensors() failed: %v", err)
	}
	if len(sensors.Sensors) == 0 {
		t.Errorf("Expected at least 1 sensor, got 0")
	}
}

// ✅ Tarihsel veri çekme testi
func TestGetHistoricalData(t *testing.T) {
	mockResponse := `{"histdata": [{"datetime": "2025-02-15T12:00:00Z", "value": 78.9}]}`
	server, api := setupMockServer(mockResponse, http.StatusOK)
	defer server.Close()

	startDate := time.Now().Add(-24 * time.Hour).UnixMilli()
	endDate := time.Now().UnixMilli()

	histData, err := api.GetHistoricalData("1234", startDate, endDate)
	if err != nil {
		t.Fatalf("GetHistoricalData() failed: %v", err)
	}
	if len(histData.HistData) == 0 {
		t.Errorf("Expected at least 1 historical data point, got 0")
	}
}

// ✅ API Hata Durumlarını Test Etme
func TestApiErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		expectErr  bool
	}{
		{"Forbidden Access", http.StatusForbidden, true},
		{"Not Found", http.StatusNotFound, true},
		{"Server Error", http.StatusInternalServerError, true},
		{"OK", http.StatusOK, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, api := setupMockServer(`{"status": "ok"}`, tt.statusCode)
			defer server.Close()

			_, err := api.GetStatusList()
			if tt.expectErr && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// ✅ Yardımcı Fonksiyon: JSON Fixture Yükleme
func loadFixture(filePath string) string {
	data, err := os.ReadFile("testdata" + filePath)
	if err != nil {
		return `{"error": "Could not load fixture"}`
	}
	return string(data)
}
