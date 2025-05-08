package plugin

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewApi(t *testing.T) {
	baseURL := "http://prtg.example.com"
	apiKey := "testkey12345"
	cacheTime := 5 * time.Minute
	timeout := 30 * time.Second

	api := NewApi(baseURL, apiKey, cacheTime, timeout)

	assert.Equal(t, baseURL, api.baseURL)
	assert.Equal(t, apiKey, api.apiKey)
	assert.Equal(t, timeout, api.timeout)
	assert.Equal(t, cacheTime, api.cacheTime)
	assert.NotNil(t, api.cache)
}

func TestBuildApiUrl(t *testing.T) {
	api := NewApi("https://prtg.example.com", "apikey123", time.Minute, time.Second*30)

	tests := []struct {
		name          string
		method        string
		params        map[string]string
		expectedURL   string
		expectedError bool
	}{
		{
			name:        "Basic URL",
			method:      "table.json",
			params:      nil,
			expectedURL: "https://prtg.example.com/api/table.json?apitoken=apikey123",
		},
		{
			name:   "URL with parameters",
			method: "historicdata.json",
			params: map[string]string{
				"id":     "1234",
				"count":  "100",
				"output": "json",
			},
			expectedURL: "https://prtg.example.com/api/historicdata.json?apitoken=apikey123&count=100&id=1234&output=json",
		},
		{
			name:          "Invalid URL",
			method:        "data.json",
			params:        nil,
			expectedError: true,
			expectedURL:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Invalid URL" {
				api.baseURL = "http://[::1]:namedport" // Invalid URL
			}

			url, err := api.buildApiUrl(tt.method, tt.params)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURL, url)
			}
		})
	}
}

func TestSetTimeout(t *testing.T) {
	api := NewApi("https://prtg.example.com", "apikey123", time.Minute, time.Second*30)

	tests := []struct {
		name          string
		timeout       time.Duration
		expectedValue time.Duration
	}{
		{
			name:          "Set longer timeout",
			timeout:       60 * time.Second,
			expectedValue: 60 * time.Second,
		},
		{
			name:          "Set shorter timeout gets minimum",
			timeout:       5 * time.Second,
			expectedValue: 10 * time.Second, // Minimum is 10 seconds
		},
		{
			name:          "Negative timeout does nothing",
			timeout:       -5 * time.Second,
			expectedValue: 10 * time.Second, // Keeps previous value
		},
		{
			name:          "Zero timeout does nothing",
			timeout:       0,
			expectedValue: 10 * time.Second, // Keeps previous value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api.SetTimeout(tt.timeout)
			assert.Equal(t, tt.expectedValue, api.timeout)
		})
	}
}

func TestBaseExecuteRequest(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		endpoint       string
		params         map[string]string
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:         "Successful request",
			statusCode:   http.StatusOK,
			responseBody: `{"status": "ok"}`,
			endpoint:     "table.json",
			params:       map[string]string{"content": "groups"},
			expectError:  false,
		},
		{
			name:           "Forbidden status",
			statusCode:     http.StatusForbidden,
			responseBody:   `{"error": "Access denied"}`,
			endpoint:       "table.json",
			params:         map[string]string{"content": "groups"},
			expectError:    true,
			expectedErrMsg: "access denied",
		},
		{
			name:           "Server error",
			statusCode:     http.StatusInternalServerError,
			responseBody:   `{"error": "Internal server error"}`,
			endpoint:       "table.json",
			params:         map[string]string{"content": "groups"},
			expectError:    true,
			expectedErrMsg: "unexpected status code: 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				fmt.Fprintln(w, tt.responseBody)
			}))
			defer server.Close()

			api := NewApi(server.URL, "testkey", time.Minute, 30*time.Second)
			body, err := api.baseExecuteRequest(tt.endpoint, tt.params)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				require.NoError(t, err)
				assert.Contains(t, string(body), "status")
			}
		})
	}
}

func TestGetCacheTime(t *testing.T) {
	expectedCacheTime := 5 * time.Minute
	api := NewApi("https://prtg.example.com", "apikey123", expectedCacheTime, 30*time.Second)
	
	actualCacheTime := api.GetCacheTime()
	assert.Equal(t, expectedCacheTime, actualCacheTime)
}
