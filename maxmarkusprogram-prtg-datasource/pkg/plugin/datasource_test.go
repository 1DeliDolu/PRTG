package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
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

// TODO - OK
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

// TODO - OK
func TestDispose(t *testing.T) {
	ds := &Datasource{}
	ds.Dispose() // Should not panic or error
}


// TODO - OK
func TestQueryData(t *testing.T) {
	tests := []struct {
		name          string
		queries       []backend.DataQuery
		expectedResps map[string]backend.DataResponse
		wantErr       bool
	}{
		{
			name: "Single query",
			queries: []backend.DataQuery{
				{
					RefID: "A",
				},
			},
			expectedResps: map[string]backend.DataResponse{
				"A": {},
			},
			wantErr: false,
		},
		{
			name: "Multiple queries within limit",
			queries: []backend.DataQuery{
				{
					RefID: "A",
				},
				{
					RefID: "B",
				},
			},
			expectedResps: map[string]backend.DataResponse{
				"A": {},
				"B": {},
			},
			wantErr: false,
		},
		{
			name: "Too many queries",
			queries: func() []backend.DataQuery {
				queries := make([]backend.DataQuery, MaxConcurrentQueries+1)
				for i := 0; i < MaxConcurrentQueries+1; i++ {
					queries[i] = backend.DataQuery{RefID: fmt.Sprintf("%c", 'A'+i)}
				}
				return queries
			}(),
			expectedResps: map[string]backend.DataResponse{
				"A": {
					Error:  fmt.Errorf("query limit exceeded: %d/%d", MaxConcurrentQueries+1, MaxConcurrentQueries),
					Status: backend.StatusTooManyRequests,
				},
			},
			wantErr: false,
		},
		{
			name:          "No queries",
			queries:       []backend.DataQuery{},
			expectedResps: map[string]backend.DataResponse{},
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Create a mock query multiplexer
			mockMux := &mockQueryDataHandler{}
			if len(tt.queries) > 0 && len(tt.queries) <= MaxConcurrentQueries {
				mockResp := &backend.QueryDataResponse{
					Responses: make(map[string]backend.DataResponse),
				}
				for _, q := range tt.queries {
					mockResp.Responses[q.RefID] = backend.DataResponse{}
				}
				mockMux.response = mockResp
			}

			ds := &Datasource{
				logger:     NewLogger(),
				tracer:     NewTracer(NewLogger()),
				metrics:    NewMetrics(prometheus.NewRegistry()),
				mux:        mockMux,
				queryCache: make(map[string]*QueryCacheEntry),
				cacheMutex: sync.RWMutex{},
				cacheTime:  time.Minute,
			}

			req := &backend.QueryDataRequest{
				Queries: tt.queries,
			}

			resp, err := ds.QueryData(ctx, req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// For query limit exceeded case
			if len(tt.queries) > MaxConcurrentQueries {
				require.Len(t, resp.Responses, 1)
				response := resp.Responses[tt.queries[0].RefID]
				assert.Equal(t, backend.StatusTooManyRequests, response.Status)
				assert.Contains(t, response.Error.Error(), "query limit exceeded")
				return
			}

			// Verify response contains expected refIDs
			for refID, expectedResp := range tt.expectedResps {
				actualResp, exists := resp.Responses[refID]
				assert.True(t, exists, "Response should exist for refID %s", refID)
				if expectedResp.Error != nil {
					assert.Equal(t, expectedResp.Status, actualResp.Status)
					assert.Contains(t, actualResp.Error.Error(), expectedResp.Error.Error())
				}
			}

			// Test cache functionality
			if len(tt.queries) > 0 && len(tt.queries) <= MaxConcurrentQueries {
				// Second request should use cache
				mockMux.called = false
				secondResp, err := ds.QueryData(ctx, req)
				require.NoError(t, err)
				assert.False(t, mockMux.called, "Second request should use cache")
				assert.NotNil(t, secondResp)
			}
		})
	}
}

// Mock implementation of QueryDataHandler
type mockQueryDataHandler struct {
	response *backend.QueryDataResponse
	err      error
	called   bool
}

func (m *mockQueryDataHandler) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	m.called = true
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}


// TODO - OK
func TestCheckHealth(t *testing.T) {
	tests := []struct {
		name            string
		mockStatus      *PrtgStatusListResponse
		mockErr         error
		expectedStatus  backend.HealthStatus
		expectedMsg     string
		expectedDetails map[string]interface{}
	}{
		{
			name: "Successful health check",
			mockStatus: &PrtgStatusListResponse{
				Version:   "23.1.84.1375",
				TotalSens: 250,
			},
			expectedStatus: backend.HealthStatusOk,
			expectedMsg:    "Data source is working. PRTG Version: 23.1.84.1375",
			expectedDetails: map[string]interface{}{
				"version":      "23.1.84.1375",
				"totalSensors": float64(250),
			},
		},
		{
			name:           "Connection error",
			mockErr:        fmt.Errorf("connection failed"),
			expectedStatus: backend.HealthStatusError,
			expectedMsg:    "PRTG API error: connection failed",
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
			}

			result, err := ds.CheckHealth(ctx, &backend.CheckHealthRequest{})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.expectedStatus, result.Status)
			assert.Equal(t, tt.expectedMsg, result.Message)

			if tt.expectedDetails != nil {
				var details map[string]interface{}
				err := json.Unmarshal(result.JSONDetails, &details)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedDetails, details)
			}
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
