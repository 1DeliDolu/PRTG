package plugin

import (
	"context"
	"net/http"

	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// ✅ Datasource oluşturma test
func TestNewDatasource(t *testing.T) {
	settings := backend.DataSourceInstanceSettings{
		JSONData: []byte(`{"Path":"prtg.example.com"}`),
		DecryptedSecureJSONData: map[string]string{
			"ApiKey": "test-api-key",
		},
	}

	ds, err := NewDatasource(context.Background(), settings)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	if ds.(*Datasource).api.apiKey != "test-api-key" {
		t.Errorf("Expected API key to be 'test-api-key', got %v", ds.(*Datasource).api.apiKey)
	}
}

// ✅ QueryData test
func TestQueryData(t *testing.T) {
	server, api := setupMockServer(`{"sensors": [{"sensor": "CPU Load"}]}`, http.StatusOK)
	defer server.Close()

	ds := &Datasource{api: api}

	req := &backend.QueryDataRequest{
		Queries: []backend.DataQuery{
			{RefID: "A", JSON: []byte(`{"queryType":"sensors"}`)},
		},
	}

	resp, err := ds.QueryData(context.Background(), req)
	if err != nil {
		t.Fatalf("QueryData failed: %v", err)
	}

	if len(resp.Responses) == 0 {
		t.Errorf("Expected at least 1 response, got 0")
	}
}

// ✅ CheckHealth test
func TestCheckHealth(t *testing.T) {
	server, api := setupMockServer(`{"prtgversion": "21.2.68.1492"}`, http.StatusOK)
	defer server.Close()

	ds := &Datasource{api: api}

	res, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	if err != nil {
		t.Fatalf("CheckHealth failed: %v", err)
	}

	if res.Status != backend.HealthStatusOk {
		t.Errorf("Expected HealthStatusOk, got %v", res.Status)
	}
}

// ✅ CallResource test: Grupları çekme
func TestCallResourceGroups(t *testing.T) {
	server, api := setupMockServer(`{"groups": [{"group": "Network Devices"}]}`, http.StatusOK)
	defer server.Close()

	ds := &Datasource{api: api}
	req := &backend.CallResourceRequest{Path: "groups"}

	respSender := &mockResourceResponseSender{}
	err := ds.CallResource(context.Background(), req, respSender)
	if err != nil {
		t.Fatalf("CallResource failed: %v", err)
	}

	if respSender.status != http.StatusOK {
		t.Errorf("Expected status 200, got %v", respSender.status)
	}
}

// ✅ CallResource test: Cihazları çekme
func TestCallResourceDevices(t *testing.T) {
	server, api := setupMockServer(`{"devices": [{"device": "Router"}]}`, http.StatusOK)
	defer server.Close()

	ds := &Datasource{api: api}
	req := &backend.CallResourceRequest{Path: "devices"}

	respSender := &mockResourceResponseSender{}
	err := ds.CallResource(context.Background(), req, respSender)
	if err != nil {
		t.Fatalf("CallResource failed: %v", err)
	}

	if respSender.status != http.StatusOK {
		t.Errorf("Expected status 200, got %v", respSender.status)
	}
}

// ✅ CallResource test: Sensörleri çekme
func TestCallResourceSensors(t *testing.T) {
	server, api := setupMockServer(`{"sensors": [{"sensor": "CPU Load"}]}`, http.StatusOK)
	defer server.Close()

	ds := &Datasource{api: api}
	req := &backend.CallResourceRequest{Path: "sensors"}

	respSender := &mockResourceResponseSender{}
	err := ds.CallResource(context.Background(), req, respSender)
	if err != nil {
		t.Fatalf("CallResource failed: %v", err)
	}

	if respSender.status != http.StatusOK {
		t.Errorf("Expected status 200, got %v", respSender.status)
	}
}

// ✅ CallResource test: Kanal verilerini çekme
func TestCallResourceChannels(t *testing.T) {
	server, api := setupMockServer(`{"values": [{"value": 45.6, "datetime": "2025-02-15T12:00:00Z"}]}`, http.StatusOK)
	defer server.Close()

	ds := &Datasource{api: api}
	req := &backend.CallResourceRequest{Path: "channels/1234"}

	respSender := &mockResourceResponseSender{}
	err := ds.CallResource(context.Background(), req, respSender)
	if err != nil {
		t.Fatalf("CallResource failed: %v", err)
	}

	if respSender.status != http.StatusOK {
		t.Errorf("Expected status 200, got %v", respSender.status)
	}
}

// ✅ Hata testleri: CallResource yanlış path
func TestCallResource_InvalidPath(t *testing.T) {
	ds := &Datasource{}
	req := &backend.CallResourceRequest{Path: "invalidpath"}

	respSender := &mockResourceResponseSender{}
	err := ds.CallResource(context.Background(), req, respSender)
	if err != nil {
		t.Fatalf("CallResource failed: %v", err)
	}

	if respSender.status != http.StatusNotFound {
		t.Errorf("Expected status 404, got %v", respSender.status)
	}
}

// ✅ CallResource test: Eksik objid
func TestCallResource_MissingObjID(t *testing.T) {
	ds := &Datasource{}
	req := &backend.CallResourceRequest{Path: "channels"}

	respSender := &mockResourceResponseSender{}
	err := ds.CallResource(context.Background(), req, respSender)
	if err != nil {
		t.Fatalf("CallResource failed: %v", err)
	}

	if respSender.status != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %v", respSender.status)
	}
}

// ✅ Mock response sender
type mockResourceResponseSender struct {
	status int
	body   []byte
}

func (m *mockResourceResponseSender) Send(resp *backend.CallResourceResponse) error {
	m.status = resp.Status
	m.body = resp.Body
	return nil
}
