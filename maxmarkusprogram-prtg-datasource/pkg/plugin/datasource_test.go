package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMetrics is a test implementation of Metrics
type TestMetrics struct {
	*Metrics
	registry prometheus.Registerer
}

var (
	metricsFactory = NewMetrics // Store the original metrics factory function
)

func TestNewDatasource(t *testing.T) {
	tests := []struct {
		name     string
		settings backend.DataSourceInstanceSettings
		wantErr  bool
	}{
		{
			name: "Valid configuration",
			settings: backend.DataSourceInstanceSettings{
				JSONData: []byte(`{
					"path": "test.example.com",
					"cacheTime": 60
				}`),
				DecryptedSecureJSONData: map[string]string{
					"apiToken": "test-token",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := prometheus.NewRegistry()
			ctx := context.Background()

			testMetrics := &TestMetrics{
				Metrics:  NewMetrics(registry),
				registry: registry,
			}

			originalFactory := metricsFactory
			metricsFactory = func(reg prometheus.Registerer) *Metrics {
				return testMetrics.Metrics
			}
			defer func() {
				metricsFactory = originalFactory
			}()

			ds, err := NewDatasource(ctx, tt.settings)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, ds)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, ds)

			datasource, ok := ds.(*Datasource)
			require.True(t, ok)

			assert.NotNil(t, datasource.logger)
			assert.NotNil(t, datasource.tracer)
			assert.NotNil(t, datasource.metrics)
			assert.NotNil(t, datasource.api)

			var jsonData map[string]interface{}
			err = json.Unmarshal(tt.settings.JSONData, &jsonData)
			require.NoError(t, err)
			expectedBaseURL := "https://" + jsonData["path"].(string)
			assert.Equal(t, expectedBaseURL, datasource.baseURL)

			if cacheTime, ok := jsonData["cacheTime"].(float64); ok && cacheTime <= 0 {
				assert.Equal(t, 30*time.Second, datasource.api.GetCacheTime())
			}
		})
	}
}

func TestDispose(t *testing.T) {
	ds := &Datasource{}
	ds.Dispose() // Should not panic or error
}

func TestQueryData(t *testing.T) {
	tests := []struct {
		name          string
		queries       []backend.DataQuery
		expectedResps map[string]bool // map of refIDs to expected presence
	}{
		{
			name: "Single query",
			queries: []backend.DataQuery{
				{
					RefID: "A",
				},
			},
			expectedResps: map[string]bool{
				"A": true,
			},
		},
		{
			name: "Multiple queries",
			queries: []backend.DataQuery{
				{
					RefID: "A",
				},
				{
					RefID: "B",
				},
			},
			expectedResps: map[string]bool{
				"A": true,
				"B": true,
			},
		},
		{
			name:          "No queries",
			queries:       []backend.DataQuery{},
			expectedResps: map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			ds := &Datasource{
				logger:  NewLogger(),
				tracer:  NewTracer(NewLogger()),
				metrics: NewMetrics(prometheus.NewRegistry()),
			}

			req := &backend.QueryDataRequest{
				Queries: tt.queries,
			}

			resp, err := ds.QueryData(ctx, req)

			require.NoError(t, err)
			require.NotNil(t, resp)

			// Verify response contains expected refIDs
			for refID, expected := range tt.expectedResps {
				_, exists := resp.Responses[refID]
				assert.Equal(t, expected, exists, "Response presence mismatch for refID %s", refID)
			}
		})
	}
}

func TestParsePRTGDateTime(t *testing.T) {
	tests := []struct {
		name         string
		datetime     string
		wantTime     string
		wantUnixTime string
		wantErr      bool
	}{
		{
			name:         "Valid date with standard format",
			datetime:     "02.01.2023 15:04:05",
			wantTime:     "2023-01-02 14:04:05 +0000 UTC", // -1h for Berlin->UTC
			wantUnixTime: "1672667045",
			wantErr:      false,
		},
		{
			name:         "Valid date with RFC3339 format",
			datetime:     "2023-01-02T15:04:05+01:00",
			wantTime:     "2023-01-02 14:04:05 +0000 UTC",
			wantUnixTime: "1672667045",
			wantErr:      false,
		},
		{
			name:         "Date range with valid end date",
			datetime:     "01.01.2023 10:00:00 - 02.01.2023 15:04:05",
			wantTime:     "2023-01-02 14:04:05 +0000 UTC",
			wantUnixTime: "1672667045",
			wantErr:      false,
		},
		{
			name:         "Invalid date format",
			datetime:     "invalid-date",
			wantTime:     "",
			wantUnixTime: "",
			wantErr:      true,
		},
		{
			name:         "Empty string",
			datetime:     "",
			wantTime:     "",
			wantUnixTime: "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTime, gotUnixTime, err := parsePRTGDateTime(tt.datetime)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, gotUnixTime)
				assert.True(t, gotTime.IsZero())
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantTime, gotTime.String())
			assert.Equal(t, tt.wantUnixTime, gotUnixTime)
		})
	}
}

func TestCheckHealth(t *testing.T) {
	tests := []struct {
		name           string
		settings       *backend.DataSourceInstanceSettings
		mockStatus     *PrtgStatusListResponse
		mockErr        error
		expectedStatus backend.HealthStatus
		expectedMsg    string
	}{
		{
			name: "Successful health check",
			settings: &backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"path": "prtg.example.com"}`),
				DecryptedSecureJSONData: map[string]string{
					"apiKey": "test-api-key",
				},
			},
			mockStatus: &PrtgStatusListResponse{
				Version: "23.1.84.1375",
			},
			expectedStatus: backend.HealthStatusOk,
			expectedMsg:    "Data source is working. PRTG Version: 23.1.84.1375",
		},
		{
			name: "Missing API key",
			settings: &backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"path": "prtg.example.com"}`),
				DecryptedSecureJSONData: map[string]string{
					"apiKey": "",
				},
			},
			expectedStatus: backend.HealthStatusError,
			expectedMsg:    "API key is required but not configured",
		},
		{
			name: "Connection error",
			settings: &backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"path": "prtg.example.com"}`),
				DecryptedSecureJSONData: map[string]string{
					"apiKey": "test-api-key",
				},
			},
			mockErr:        fmt.Errorf("connection failed"),
			expectedStatus: backend.HealthStatusError,
			expectedMsg:    "PRTG connection failed: connection failed",
		},
		{
			name: "Invalid response",
			settings: &backend.DataSourceInstanceSettings{
				JSONData: []byte(`{"path": "prtg.example.com"}`),
				DecryptedSecureJSONData: map[string]string{
					"apiKey": "test-api-key",
				},
			},
			mockStatus:     &PrtgStatusListResponse{Version: ""},
			expectedStatus: backend.HealthStatusError,
			expectedMsg:    "Invalid response from PRTG server",
		},
		{
			name: "Invalid settings",
			settings: &backend.DataSourceInstanceSettings{
				JSONData: []byte(`invalid json`),
			},
			expectedStatus: backend.HealthStatusError,
			expectedMsg:    "Configuration error: invalid character 'i' looking for beginning of value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			registry := prometheus.NewRegistry()

			mockApi := &MockApi{
				statusResponse: tt.mockStatus,
				err:            tt.mockErr,
			}

			ds := &Datasource{
				api:     mockApi,
				logger:  NewLogger(),
				tracer:  NewTracer(NewLogger()),
				metrics: NewMetrics(registry),
				baseURL: "https://prtg.example.com",
			}

			req := &backend.CheckHealthRequest{
				PluginContext: backend.PluginContext{
					DataSourceInstanceSettings: tt.settings,
				},
			}

			result, err := ds.CheckHealth(ctx, req)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.expectedStatus, result.Status)
			assert.Equal(t, tt.expectedMsg, result.Message)
		})
	}
}

// MockApi implements the API interface for testing
type MockApi struct {
	statusResponse *PrtgStatusListResponse
	groups         *PrtgGroupListResponse
	devices        *PrtgDevicesListResponse
	sensors        *PrtgSensorsListResponse
	histData       *PrtgHistoricalDataResponse
	err            error
	timeout        time.Duration
	cacheTime      time.Duration
	manualResponse *PrtgManualMethodResponse
}

// Add required interface methods
func (m *MockApi) GetCacheTime() time.Duration {
	return m.cacheTime
}

func (m *MockApi) SetTimeout(timeout time.Duration) {
	m.timeout = timeout
}

func (m *MockApi) GetStatusList() (*PrtgStatusListResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.statusResponse, nil
}

func (m *MockApi) GetGroups() (*PrtgGroupListResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.groups, nil
}

func (m *MockApi) GetDevices(group string) (*PrtgDevicesListResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.devices, nil
}

func (m *MockApi) GetSensors(device string) (*PrtgSensorsListResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.sensors, nil
}

func (m *MockApi) GetChannels(objid string) (*PrtgChannelValueStruct, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &PrtgChannelValueStruct{}, nil
}

func (m *MockApi) GetHistoricalData(sensorID string, startDate, endDate time.Time) (*PrtgHistoricalDataResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.histData, nil
}

func (m *MockApi) ExecuteManualMethod(method string, objectId string) (*PrtgManualMethodResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.manualResponse, nil
}

func (m *MockApi) GetAnnotationData(query *AnnotationQuery) (*AnnotationResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &AnnotationResponse{}, nil
}
