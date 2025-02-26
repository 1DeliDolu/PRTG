package plugin

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

)

func TestApi(t *testing.T) {
	t.Run("NewApi initialization", func(t *testing.T) {
		baseURL := "http://test.com"
		apiKey := "testkey123"
		cacheTime := 5 * time.Minute
		timeout := 30 * time.Second

		api := NewApi(baseURL, apiKey, cacheTime, timeout)

		assert.NotNil(t, api)
		assert.Equal(t, cacheTime, api.GetCacheTime())
	})

	t.Run("GetCacheTime returns correct duration", func(t *testing.T) {
		expectedCache := 10 * time.Minute
		api := NewApi("http://test.com", "key", expectedCache, time.Second)

		actualCache := api.GetCacheTime()

		assert.Equal(t, expectedCache, actualCache)
	})

	t.Run("NewApi with zero values", func(t *testing.T) {
		api := NewApi("", "", 0, 0)

		assert.NotNil(t, api)
		assert.Equal(t, time.Duration(0), api.GetCacheTime())
	})

	/* ====================================== URL BUILDER ====================================== */
	t.Run("buildApiUrl builds correct URL", func(t *testing.T) {
		api := NewApi("http://test.com", "testkey", time.Minute, time.Second)
		params := map[string]string{
			"param1": "value1",
			"param2": "value2",
		}

		url, err := api.buildApiUrl("testmethod", params)

		assert.NoError(t, err)
		assert.Contains(t, url, "http://test.com/api/testmethod")
		assert.Contains(t, url, "apitoken=testkey")
		assert.Contains(t, url, "param1=value1")
		assert.Contains(t, url, "param2=value2")
	})

	t.Run("buildApiUrl with invalid URL returns error", func(t *testing.T) {
		api := NewApi("://invalid", "testkey", time.Minute, time.Second)

		url, err := api.buildApiUrl("method", nil)

		assert.Error(t, err)
		assert.Empty(t, url)
	})

	t.Run("SetTimeout updates timeout value", func(t *testing.T) {
		api := NewApi("http://test.com", "key", time.Minute, time.Second)
		newTimeout := 5 * time.Second

		api.SetTimeout(newTimeout)

		assert.Equal(t, newTimeout, api.timeout)
	})

	t.Run("SetTimeout with zero value keeps original timeout", func(t *testing.T) {
		originalTimeout := 2 * time.Second
		api := NewApi("http://test.com", "key", time.Minute, originalTimeout)

		api.SetTimeout(0)

		assert.Equal(t, originalTimeout, api.timeout)
	})

	/* =================================== REQUEST EXECUTOR ====================================== */
	t.Run("baseExecuteRequest with invalid endpoint returns error", func(t *testing.T) {
		api := NewApi("http://test.com", "testkey", time.Minute, time.Second)
		
		response, err := api.baseExecuteRequest("", nil)
		
		assert.Error(t, err)
		assert.Nil(t, response)
	})

	t.Run("baseExecuteRequest uses cache when available", func(t *testing.T) {
		api := NewApi("http://test.com", "testkey", time.Minute, time.Second)
		testData := []byte("cached data")
		testURL := "http://test.com/api/test?apitoken=testkey"
		
		api.cache[testURL] = cacheItem{
			data: testData,
			expiry: time.Now().Add(time.Minute),
		}
		
		response, err := api.baseExecuteRequest("test", nil)
		
		assert.NoError(t, err)
		assert.Equal(t, testData, response)
	})

	t.Run("baseExecuteRequest bypasses cache when expired", func(t *testing.T) {
		api := NewApi("http://test.com", "testkey", time.Minute, time.Second)
		testData := []byte("cached data")
		testURL := "http://test.com/api/test?apitoken=testkey"
		
		api.cache[testURL] = cacheItem{
			data: testData,
			expiry: time.Now().Add(-time.Minute), // Expired cache
		}
		
		response, err := api.baseExecuteRequest("test", nil)
		
		assert.Error(t, err) // Should fail because no real HTTP server
		assert.Nil(t, response)
	})

	t.Run("baseExecuteRequest with invalid parameters returns error", func(t *testing.T) {
		api := NewApi("://invalid", "testkey", time.Minute, time.Second)
		
		response, err := api.baseExecuteRequest("test", nil)
		
		assert.Error(t, err)
		assert.Nil(t, response)
	})

/* ====================================== STATUS HANDLER ====================================== */
t.Run("GetStatusList returns error on request failure", func(t *testing.T) {
	api := NewApi("http://test.com", "testkey", time.Minute, time.Second)

	response, err := api.GetStatusList()

	assert.Error(t, err)
	assert.Nil(t, response)
})

t.Run("GetStatusList returns error on invalid JSON response", func(t *testing.T) {
	api := NewApi("http://test.com", "testkey", time.Minute, time.Second)
	invalidJSON := []byte(`{"invalid json`)
	testURL := "http://test.com/api/status.json?apitoken=testkey"
	
	api.cache[testURL] = cacheItem{
		data: invalidJSON,
		expiry: time.Now().Add(time.Minute),
	}

	response, err := api.GetStatusList()

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to parse response")
})

t.Run("GetStatusList successfully parses valid response", func(t *testing.T) {
	api := NewApi("http://test.com", "testkey", time.Minute, time.Second)
	validJSON := []byte(`{"prtg-version": "1.2.3", "status": "success"}`)
	testURL := "http://test.com/api/status.json?apitoken=testkey"
	
	api.cache[testURL] = cacheItem{
		data: validJSON,
		expiry: time.Now().Add(time.Minute),
	}

	response, err := api.GetStatusList()

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "1.2.3", response.PrtgVersion)
	assert.Equal(t, "success", response.Version)
})





/* ====================================== GROUP HANDLER ======================================== */
t.Run("GetGroups returns error on request failure", func(t *testing.T) {
	api := NewApi("http://test.com", "testkey", time.Minute, time.Second)

	response, err := api.GetGroups()

	assert.Error(t, err)
	assert.Nil(t, response)
})

t.Run("GetGroups returns error on invalid JSON response", func(t *testing.T) {
	api := NewApi("http://test.com", "testkey", time.Minute, time.Second)
	invalidJSON := []byte(`{"invalid json`)
	testURL := "http://test.com/api/table.json?apitoken=testkey&columns=active%2Cchannel%2Cdatetime%2Cdevice%2Cgroup%2Cmessage%2Cobjid%2Cpriority%2Csensor%2Cstatus%2Ctags&content=groups&count=50000"
	
	api.cache[testURL] = cacheItem{
		data: invalidJSON,
		expiry: time.Now().Add(time.Minute),
	}

	response, err := api.GetGroups()

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to parse response")
})

t.Run("GetGroups successfully parses valid response", func(t *testing.T) {
	api := NewApi("http://test.com", "testkey", time.Minute, time.Second)
	validJSON := []byte(`{"prtg-version":"1.2.3","treesize":123,"groups":[{"objid":1,"group":"Test Group"}]}`)
	testURL := "http://test.com/api/table.json?apitoken=testkey&columns=active%2Cchannel%2Cdatetime%2Cdevice%2Cgroup%2Cmessage%2Cobjid%2Cpriority%2Csensor%2Cstatus%2Ctags&content=groups&count=50000"
	
	api.cache[testURL] = cacheItem{
		data: validJSON,
		expiry: time.Now().Add(time.Minute),
	}

	response, err := api.GetGroups()

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "1.2.3", response.PrtgVersion)
	assert.Equal(t, 123, response.TreeSize)
	assert.Len(t, response.Groups, 1)
	assert.Equal(t, 1, response.Groups[0].ObjectId)
	assert.Equal(t, "Test Group", response.Groups[0].Group)
})


/* ====================================== DEVICE HANDLER ======================================== */
t.Run("GetDevices returns error on empty group", func(t *testing.T) {
	api := NewApi("http://test.com", "testkey", time.Minute, time.Second)

	response, err := api.GetDevices("")

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "group parameter is required")
})

t.Run("GetDevices returns error on request failure", func(t *testing.T) {
	api := NewApi("http://test.com", "testkey", time.Minute, time.Second)

	response, err := api.GetDevices("TestGroup")

	assert.Error(t, err)
	assert.Nil(t, response)
})

t.Run("GetDevices returns error on invalid JSON response", func(t *testing.T) {
	api := NewApi("http://test.com", "testkey", time.Minute, time.Second)
	invalidJSON := []byte(`{"invalid json`)
	testURL := "http://test.com/api/table.json?apitoken=testkey&columns=active%2Cchannel%2Cdatetime%2Cdevice%2Cgroup%2Cmessage%2Cobjid%2Cpriority%2Csensor%2Cstatus%2Ctags&content=devices&count=50000&filter_group=TestGroup"
	
	api.cache[testURL] = cacheItem{
		data: invalidJSON,
		expiry: time.Now().Add(time.Minute),
	}

	response, err := api.GetDevices("TestGroup")

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to parse response")
})

t.Run("GetDevices successfully parses valid response", func(t *testing.T) {
	api := NewApi("http://test.com", "testkey", time.Minute, time.Second)
	validJSON := []byte(`{"prtg-version":"1.2.3","treesize":123,"devices":[{"objid":1,"device":"Test Device","group":"Test Group"}]}`)
	testURL := "http://test.com/api/table.json?apitoken=testkey&columns=active%2Cchannel%2Cdatetime%2Cdevice%2Cgroup%2Cmessage%2Cobjid%2Cpriority%2Csensor%2Cstatus%2Ctags&content=devices&count=50000&filter_group=TestGroup"
	
	api.cache[testURL] = cacheItem{
		data: validJSON,
		expiry: time.Now().Add(time.Minute),
	}

	response, err := api.GetDevices("TestGroup")

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "1.2.3", response.PrtgVersion)
	assert.Equal(t, 123, response.TreeSize)
	assert.Len(t, response.Devices, 1)
	assert.Equal(t, 1, response.Devices[0].ObjectId)
	assert.Equal(t, "Test Device", response.Devices[0].Device)
	assert.Equal(t, "Test Group", response.Devices[0].Group)
})

/* ====================================== SENSOR HANDLER ======================================== */
t.Run("GetSensors returns error on empty device", func(t *testing.T) {
	api := NewApi("http://test.com", "testkey", time.Minute, time.Second)

	response, err := api.GetSensors("")

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "device parameter is required")
})

t.Run("GetSensors successfully parses valid response", func(t *testing.T) {
	api := NewApi("http://test.com", "testkey", time.Minute, time.Second)
	validJSON := []byte(`{"prtg-version":"1.2.3","treesize":123,"sensors":[{"objid":1,"sensor":"Test Sensor","device":"Test Device"}]}`)
	testURL := "http://test.com/api/table.json?apitoken=testkey&columns=active%2Cchannel%2Cdatetime%2Cdevice%2Cgroup%2Cmessage%2Cobjid%2Cpriority%2Csensor%2Cstatus%2Ctags&content=sensors&count=50000&filter_device=TestDevice"
	
	api.cache[testURL] = cacheItem{
		data: validJSON,
		expiry: time.Now().Add(time.Minute),
	}

	response, err := api.GetSensors("TestDevice")

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "1.2.3", response.PrtgVersion)
	assert.Equal(t, 123, response.TreeSize)
	assert.Len(t, response.Sensors, 1)
	assert.Equal(t, 1, response.Sensors[0].ObjectId)
	assert.Equal(t, "Test Sensor", response.Sensors[0].Sensor)
})

/* ====================================== CHANNEL HANDLER ======================================== */
t.Run("GetChannels successfully parses valid response", func(t *testing.T) {
	api := NewApi("http://test.com", "testkey", time.Minute, time.Second)
	validJSON := []byte(`{"prtg-version":"1.2.3","histdata":[{"datetime":"2023-01-01","value_":123.45}]}`)
	testURL := "http://test.com/api/historicdata.json?apitoken=testkey&columns=value_%2Cdatetime&content=values&count=50000&id=123&usecaption=true"
	
	api.cache[testURL] = cacheItem{
		data: validJSON,
		expiry: time.Now().Add(time.Minute),
	}

	response, err := api.GetChannels("123")

	assert.NoError(t, err)
	assert.NotNil(t, response)

})

/* ====================================== HISTORY HANDLER ======================================== */
t.Run("GetHistoricalData returns error on empty sensor ID", func(t *testing.T) {
	api := NewApi("http://test.com", "testkey", time.Minute, time.Second)
	startDate := time.Now().Add(-24 * time.Hour)
	endDate := time.Now()

	response, err := api.GetHistoricalData("", startDate, endDate)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "missing sensor ID")
})

t.Run("GetHistoricalData returns error on invalid time range", func(t *testing.T) {
	api := NewApi("http://test.com", "testkey", time.Minute, time.Second)
	startDate := time.Now()
	endDate := time.Now().Add(-24 * time.Hour)

	response, err := api.GetHistoricalData("123", startDate, endDate)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid time range")
})

t.Run("GetHistoricalData successfully parses valid response", func(t *testing.T) {
	api := NewApi("http://test.com", "testkey", time.Minute, time.Second)
	validJSON := []byte(`{"prtg-version":"1.2.3","histdata":[{"datetime":"2023-01-01","value_":123.45}]}`)
	testURL := "http://test.com/api/historicdata.json?apitoken=testkey&avg=0&columns=datetime%2Cvalue_&count=50000&edate=2023-01-02-01-00-00&id=123&pctmode=false&pctshow=false&sdate=2023-01-01-01-00-00&usecaption=1"
	
	api.cache[testURL] = cacheItem{
		data: validJSON,
		expiry: time.Now().Add(time.Minute),
	}

	startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)
	response, err := api.GetHistoricalData("123", startDate, endDate)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "1.2.3", response.PrtgVersion)
	assert.Len(t, response.HistData, 1)
})

/* ====================================== MANUAL METHOD HANDLER ================================= */
t.Run("ExecuteManualMethod successfully parses response", func(t *testing.T) {
	api := NewApi("http://test.com", "testkey", time.Minute, time.Second)
	validJSON := []byte(`{"result":"success","data":{"key1":"value1","nested":{"key2":"value2"}}}`)
	testURL := "http://test.com/api/testmethod?apitoken=testkey&id=123"
	
	api.cache[testURL] = cacheItem{
		data: validJSON,
		expiry: time.Now().Add(time.Minute),
	}

	response, err := api.ExecuteManualMethod("testmethod", "123")

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotNil(t, response.Manuel)
	assert.NotEmpty(t, response.KeyValues)
})

t.Run("FlattenJSON correctly flattens nested structure", func(t *testing.T) {
	var result []KeyValue
	testData := map[string]interface{}{
		"key1": "value1",
		"nested": map[string]interface{}{
			"key2": "value2",
		},
		"array": []interface{}{
			"item1",
			map[string]interface{}{
				"key3": "value3",
			},
		},
	}

	flattenJSON("", testData, &result)

	assert.NotEmpty(t, result)
	assert.Contains(t, result, KeyValue{Key: "key1", Value: "value1"})
	assert.Contains(t, result, KeyValue{Key: "nested.key2", Value: "value2"})
	assert.Contains(t, result, KeyValue{Key: "array[0]", Value: "item1"})
	assert.Contains(t, result, KeyValue{Key: "array[1].key3", Value: "value3"})
})


}

