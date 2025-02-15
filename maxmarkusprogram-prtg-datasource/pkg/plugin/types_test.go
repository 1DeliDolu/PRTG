package plugin

import (
	"encoding/json"
	"testing"
)

// ✅ PrtgGroupListResponse JSON Unmarshal Testi
func TestPrtgGroupListResponse_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"prtg-version": "23.1.1.1",
		"treesize": 1000,
		"groups": [
			{
				"group": "Network Devices",
				"active": true,
				"status": "Up",
				"objid": 12345
			}
		]
	}`

	var response PrtgGroupListResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if response.PrtgVersion != "23.1.1.1" {
		t.Errorf("Expected version '23.1.1.1', got %s", response.PrtgVersion)
	}
	if len(response.Groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(response.Groups))
	}
	if response.Groups[0].Group != "Network Devices" {
		t.Errorf("Expected group name 'Network Devices', got %s", response.Groups[0].Group)
	}
}

// ✅ PrtgDevicesListResponse JSON Unmarshal Testi
func TestPrtgDevicesListResponse_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"prtg-version": "23.1.1.1",
		"treesize": 1000,
		"devices": [
			{
				"device": "Main Router",
				"active": true,
				"status": "Warning",
				"objid": 67890
			}
		]
	}`

	var response PrtgDevicesListResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if response.PrtgVersion != "23.1.1.1" {
		t.Errorf("Expected version '23.1.1.1', got %s", response.PrtgVersion)
	}
	if len(response.Devices) != 1 {
		t.Errorf("Expected 1 device, got %d", len(response.Devices))
	}
	if response.Devices[0].Device != "Main Router" {
		t.Errorf("Expected device name 'Main Router', got %s", response.Devices[0].Device)
	}
}

// ✅ PrtgSensorsListResponse JSON Unmarshal Testi
func TestPrtgSensorsListResponse_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"prtg-version": "23.1.1.1",
		"treesize": 1000,
		"sensors": [
			{
				"sensor": "CPU Load",
				"active": true,
				"status": "Critical",
				"objid": 34567
			}
		]
	}`

	var response PrtgSensorsListResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if response.PrtgVersion != "23.1.1.1" {
		t.Errorf("Expected version '23.1.1.1', got %s", response.PrtgVersion)
	}
	if len(response.Sensors) != 1 {
		t.Errorf("Expected 1 sensor, got %d", len(response.Sensors))
	}
	if response.Sensors[0].Sensor != "CPU Load" {
		t.Errorf("Expected sensor name 'CPU Load', got %s", response.Sensors[0].Sensor)
	}
}

// ✅ PrtgChannelValueStruct JSON Unmarshal Testi
func TestPrtgChannelValueStruct_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"datetime": "2025-02-15T12:00:00Z",
		"value1": 78.9,
		"value2": "Critical"
	}`

	var response PrtgChannelValueStruct
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if len(response) != 3 {
		t.Errorf("Expected 3 keys in map, got %d", len(response))
	}

	if response["value1"] != 78.9 {
		t.Errorf("Expected value1 to be 78.9, got %v", response["value1"])
	}

	if response["value2"] != "Critical" {
		t.Errorf("Expected value2 to be 'Critical', got %v", response["value2"])
	}
}

// ✅ PrtgHistoricalDataResponse JSON Unmarshal Testi
func TestPrtgHistoricalDataResponse_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"prtg-version": "23.1.1.1",
		"treesize": 1000,
		"histdata": [
			{
				"datetime": "2025-02-15T12:00:00Z",
				"cpu_load": 65.5
			}
		]
	}`

	var response PrtgHistoricalDataResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if response.PrtgVersion != "23.1.1.1" {
		t.Errorf("Expected version '23.1.1.1', got %s", response.PrtgVersion)
	}
	if len(response.HistData) != 1 {
		t.Errorf("Expected 1 historical data point, got %d", len(response.HistData))
	}
	if response.HistData[0].Datetime != "2025-02-15T12:00:00Z" {
		t.Errorf("Expected datetime '2025-02-15T12:00:00Z', got %s", response.HistData[0].Datetime)
	}

	if response.HistData[0].Value["cpu_load"] != 65.5 {
		t.Errorf("Expected cpu_load to be 65.5, got %v", response.HistData[0].Value["cpu_load"])
	}
}