package plugin

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

func NewApi(baseURL, apiKey string, cacheTime, requestTimeout time.Duration) *Api {
	return &Api{
		baseURL:   baseURL,
		apiKey:    apiKey,
		timeout:   requestTimeout,
		cacheTime: cacheTime,
		cache:     make(map[string]cacheItem),
	}
}

// ClearCache clears all cached data
func (a *Api) ClearCache() {
	a.cacheMu.Lock()
	defer a.cacheMu.Unlock()
	a.cache = make(map[string]cacheItem)
}
// buildApiUrl erstellt eine standardisierte PRTG-API-URL mit übergebenen Parametern.
func (a *Api) buildApiUrl(method string, params map[string]string) (string, error) {
	baseUrl := fmt.Sprintf("%s/api/%s", a.baseURL, method)
	u, err := url.Parse(baseUrl)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	q := url.Values{}
	q.Set("apitoken", a.apiKey)

	for key, value := range params {
		q.Set(key, value)
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

// SetTimeout aktualisiert das Timeout für API-Anfragen.
func (a *Api) SetTimeout(timeout time.Duration) {
	if timeout > 0 {
		if timeout < 20*time.Second {
			timeout = 20 * time.Second // Minimum 20 seconds
		}
		a.timeout = timeout
	}
}

// baseExecuteRequest führt die HTTP-Anfrage durch und liefert den Response-Body.
func (a *Api) baseExecuteRequest(endpoint string, params map[string]string) ([]byte, error) {
	apiUrl, err := a.buildApiUrl(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	client := &http.Client{
		Timeout: a.timeout,
		Transport: &http.Transport{
			// Achtung: InsecureSkipVerify sollte in Produktionsumgebungen überprüft werden!
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		log.DefaultLogger.Error("Access denied: please verify API token and permissions")
		return nil, fmt.Errorf("access denied: please verify API token and permissions")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return body, nil
}

/* ====================================== STATUS HANDLER ======================================== */
func (a *Api) GetStatusList() (*PrtgStatusListResponse, error) {
	body, err := a.baseExecuteRequest("status.json", nil)
	if err != nil {
		return nil, err
	}

	var response PrtgStatusListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &response, nil
}

/* ====================================== GROUP HANDLER ========================================= */
func (a *Api) GetGroups() (*PrtgGroupListResponse, error) {
	params := map[string]string{
		"content": "groups",
		"columns": "active,channel,datetime,device,group,message,objid,priority,sensor,status,tags",
		"count":   "50000",
		"output":  "json", // Explicitly request JSON output
	}

	body, err := a.baseExecuteRequest("table.json", params)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("empty response from PRTG API")
	}

	// Log raw response for debugging
	log.DefaultLogger.Debug("Raw PRTG response",
		"endpoint", "groups",
		"responseSize", len(body),
		"response", string(body),
	)

	var response PrtgGroupListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w, body: %s", err, string(body))
	}

	// Validate response
	if response.Groups == nil {
		return nil, fmt.Errorf("invalid response structure: groups array is nil")
	}

	return &response, nil
}

/* ====================================== DEVICE HANDLER ======================================== */
func (a *Api) GetDevices(group string) (*PrtgDevicesListResponse, error) {
	if group == "" {
		return nil, fmt.Errorf("group parameter is required")
	}

	params := map[string]string{
		"content":      "devices",
		"columns":      "active,channel,datetime,device,group,message,objid,priority,sensor,status,tags",
		"count":        "50000",
		"filter_group": group,
	}

	body, err := a.baseExecuteRequest("table.json", params)
	if err != nil {
		return nil, err
	}

	var response PrtgDevicesListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

/* ====================================== SENSOR HANDLER ======================================== */
func (a *Api) GetSensors(device string) (*PrtgSensorsListResponse, error) {
	if device == "" {
		return nil, fmt.Errorf("device parameter is required")
	}

	params := map[string]string{
		"content":       "sensors",
		"columns":       "active,channel,datetime,device,group,message,objid,priority,sensor,status,tags",
		"count":         "50000",
		"filter_device": device,
	}

	body, err := a.baseExecuteRequest("table.json", params)
	if err != nil {
		return nil, err
	}

	var response PrtgSensorsListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

/* ====================================== CHANNEL HANDLER ======================================= */
func (a *Api) GetChannels(objid string) (*PrtgChannelValueStruct, error) {
	params := map[string]string{
		"content":    "values",
		"id":         objid,
		"columns":    "value_,datetime",
		"usecaption": "true",
	}

	body, err := a.baseExecuteRequest("historicdata.json", params)
	if err != nil {
		return nil, err
	}

	var response PrtgChannelValueStruct
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetHistoricalData ruft historische Daten für den angegebenen Sensor und Zeitraum ab.
func (a *Api) GetHistoricalData(sensorID string, startDate, endDate time.Time) (*PrtgHistoricalDataResponse, error) {
	// Input validation
	if sensorID == "" {
		return nil, fmt.Errorf("invalid query: missing sensor ID")
	}

	// Convert to PRTG's timezone (Europe/Berlin)
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		loc = time.Local
	}

	// Add a small buffer to ensure we don't miss data at day boundaries
	localStartDate := startDate.In(loc).Add(-1 * time.Hour)
	localEndDate := endDate.In(loc).Add(1 * time.Hour)

	// Format dates for PRTG API
	const format = "2006-01-02-15-04-05"
	sdate := localStartDate.Format(format)
	edate := localEndDate.Format(format)

	// Calculate adjusted time range
	hours := localEndDate.Sub(localStartDate).Hours()

	// Rest of the averaging logic...
	var avg string
	switch {
	case hours <= 12:
		avg = "0"
	case hours <= 24:
		avg = "120"
	case hours <= 48:
		avg = "300"
	case hours <= 96:
		avg = "600"
	case hours <= 168:
		avg = "900"
	case hours <= 336:
		avg = "1800"
	case hours <= 720:
		avg = "3600"
	case hours <= 1440:
		avg = "7200"
	case hours <= 2880:
		avg = "14400"
	case hours <= 4320:
		avg = "28800"
	case hours <= 10080:
		avg = "43200"
	case hours <= 20160:
		avg = "57600"	
	case hours <= 43200:
		avg = "86400"
	default:
		avg = "172800" // 2 days
	
	}

	params := map[string]string{
		"id":         sensorID,
		"columns":    "datetime,value_",
		"avg":        avg,
		"sdate":      sdate,
		"edate":      edate,
		"count":      "50000",
		"usecaption": "1",
	}

	log.DefaultLogger.Debug("Requesting historical data",
		"sensorID", sensorID,
		"startDate", sdate,
		"endDate", edate,
		"avg", avg,
	)

	// Use cacheTime for response caching
	cacheKey := fmt.Sprintf("hist_%s_%s_%s", sensorID, startDate.Format(time.RFC3339), endDate.Format(time.RFC3339))

	// Check cache
	a.cacheMu.RLock()
	if cached, exists := a.cache[cacheKey]; exists && time.Now().Before(cached.expiry) {
		a.cacheMu.RUnlock()
		var response PrtgHistoricalDataResponse
		if err := json.Unmarshal(cached.data, &response); err == nil {
			return &response, nil
		}
	}
	a.cacheMu.RUnlock()

	// Make API request
	body, err := a.baseExecuteRequest("historicdata.json", params)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch historical data: %w", err)
	}

	// Parse response
	var response PrtgHistoricalDataResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Validate response
	if len(response.HistData) == 0 {
		log.DefaultLogger.Debug("No data found for the given time range",
			"sensorID", sensorID,
			"startDate", startDate,
			"endDate", endDate,
		)
		return &response, nil // Return empty response instead of error
	}

	// Cache the response
	if len(response.HistData) > 0 {
		a.cacheMu.Lock()
		if data, err := json.Marshal(response); err == nil {
			a.cache[cacheKey] = cacheItem{
				data:   data,
				expiry: time.Now().Add(a.cacheTime),
			}
		}
		a.cacheMu.Unlock()
	}

	return &response, nil
}

/* ====================================== MANUAL METHOD HANDLER ================================= */
func (a *Api) ExecuteManualMethod(method string, objectId string) (*PrtgManualMethodResponse, error) {
	params := map[string]string{}

	if objectId != "" {
		params["id"] = objectId
	}

	body, err := a.baseExecuteRequest(method, params)
	if err != nil {
		return nil, fmt.Errorf("manual API request failed: %w", err)
	}

	var rawData map[string]interface{}
	if err := json.Unmarshal(body, &rawData); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var keyValues []KeyValue
	flattenJSON("", rawData, &keyValues)

	return &PrtgManualMethodResponse{
		Manuel:    rawData,
		KeyValues: keyValues,
	}, nil
}

/* ====================================== FLATTEN JSON ======================================== */
func flattenJSON(prefix string, data interface{}, result *[]KeyValue) {
	switch v := data.(type) {
	case map[string]interface{}:
		for k, val := range v {
			key := k
			if prefix != "" {
				key = prefix + "." + k
			}
			switch child := val.(type) {
			case map[string]interface{}:
				flattenJSON(key, child, result)
			case []interface{}:
				for i, item := range child {
					arrayKey := fmt.Sprintf("%s[%d]", key, i)
					flattenJSON(arrayKey, item, result)
				}
			default:
				*result = append(*result, KeyValue{
					Key:   key,
					Value: val,
				})
			}
		}
	default:
		if prefix != "" {
			*result = append(*result, KeyValue{
				Key:   prefix,
				Value: v,
			})
		}
	}
}

/* ====================================== ANNOTATION HANDLER ====================================== */
func (a *Api) GetAnnotationData(query *AnnotationQuery) (*AnnotationResponse, error) {
	// Get time range
	fromTime := time.Unix(0, query.From*int64(time.Millisecond))
	toTime := time.Unix(0, query.To*int64(time.Millisecond))

	histData, err := a.GetHistoricalData(query.SensorID, fromTime, toTime)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch historical data for annotations: %w", err)
	}

	annotations := make([]Annotation, 0)
	for i, data := range histData.HistData {
		t, _, err := parsePRTGDateTime(data.Datetime)
		if err != nil {
			continue
		}

		// Create unique ID using datasource UID format
		uid := fmt.Sprintf("uid:%s_%d", query.SensorID, i)

		annotation := Annotation{
			ID:      uid,
			Time:    t.UnixMilli(),
			TimeEnd: t.UnixMilli(),
			Title:   fmt.Sprintf("Sensor: %s", query.SensorID),
			Tags:    []string{"prtg", fmt.Sprintf("sensor:%s", query.SensorID)},
			Type:    "annotation",
			Data:    data.Value,
		}

		annotations = append(annotations, annotation)
	}

	// Apply limit after all filtering
	if query.Limit > 0 && int64(len(annotations)) > query.Limit {
		annotations = annotations[:query.Limit]
	}

	return &AnnotationResponse{
		Annotations: annotations,
		Total:       len(annotations),
	}, nil
}

// Add GetCacheTime method to implement PRTGAPI interface
func (a *Api) GetCacheTime() time.Duration {
	return a.cacheTime
}