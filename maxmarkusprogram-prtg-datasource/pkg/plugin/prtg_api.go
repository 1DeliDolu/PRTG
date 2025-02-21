package plugin

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
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

/*  ################################################# buildApiUrl ##################################################### */
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

func (a *Api) SetTimeout(timeout time.Duration) {
	if timeout > 0 {
		a.timeout = timeout
	}
}

/* #############################################  baseExecuteRequest #####################################################*/
func (a *Api) baseExecuteRequest(endpoint string, params map[string]string) ([]byte, error) {
	apiUrl, err := a.buildApiUrl(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	if a.cacheTime > 0 {
		a.cacheMu.RLock()
		if item, ok := a.cache[apiUrl]; ok && time.Now().Before(item.expiry) {
			a.cacheMu.RUnlock()
			return item.data, nil
		}
		a.cacheMu.RUnlock()
	}

	client := &http.Client{
		Timeout: a.timeout,
		Transport: &http.Transport{

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

	if resp.StatusCode == http.StatusForbidden {
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

	if a.cacheTime > 0 {
		a.cacheMu.Lock()
		a.cache[apiUrl] = cacheItem{
			data:   body,
			expiry: time.Now().Add(a.cacheTime),
		}
		a.cacheMu.Unlock()
	}

	return body, nil
}

/*  ########################################## GetStausList ################################################## */
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

/*  ########################################## GetGroups ################################################## */
func (a *Api) GetGroups() (*PrtgGroupListResponse, error) {
	params := map[string]string{
		"content": "groups",
		"columns": "active,channel,datetime,device,group,message,objid,priority,sensor,status,tags",
		"count":   "50000",
	}

	body, err := a.baseExecuteRequest("table.json", params)
	if err != nil {
		return nil, err
	}

	var response PrtgGroupListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

/*  ########################################## GetDevices ################################################## */
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

/*  ########################################## GetSensors ################################################## */
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

/*  ########################################## GetChannels ################################################## */
func (a *Api) GetChannels(objid string) (*PrtgChannelValueStruct, error) {
	params := map[string]string{
		"content":    "values",
		"id":         objid,
		"columns":    "value_,datetime",
		"usecaption": "true",
		"count":      "50000",
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

<<<<<<< HEAD
/*  ########################################## GetHistoricalData ################################################## */
func (a *Api) GetHistoricalData(sensorID string, startDate, endDate time.Time) (*PrtgHistoricalDataResponse, error) {
=======
// GetHistoricalData ruft historische Daten für den angegebenen Sensor und Zeitraum ab.
func (a *Api) GetHistoricalData(sensorID string, startDate, endDate time.Time) (*PrtgHistoricalDataResponse, error) {
	// Input validation
>>>>>>> b7ec34b15515724822d7961b43e74d64b1be22b5
	if sensorID == "" {
		return nil, fmt.Errorf("invalid query: missing sensor ID")
	}

<<<<<<< HEAD
	// Zaman aralığını 1 saat geriye al (önceki -1 yerine +1 yapıyoruz)
	startTime := startDate.Add(time.Hour)
	endTime := endDate.Add(time.Hour)

=======
	// Convert timestamps to local time
	startTime := startDate
	endTime := endDate

	// Format dates in local time
>>>>>>> b7ec34b15515724822d7961b43e74d64b1be22b5
	const format = "2006-01-02-15-04-05"
	sdate := startTime.Format(format)
	edate := endTime.Format(format)

<<<<<<< HEAD
=======
	backend.Logger.Debug("Fetching historical data", "sensorID", sensorID, "startDate", sdate, "endDate", edate)

	// Calculate hours and validate time range
>>>>>>> b7ec34b15515724822d7961b43e74d64b1be22b5
	hours := endTime.Sub(startTime).Hours()

	if hours <= 0 {
		return nil, fmt.Errorf("invalid time range: start date %v must be before end date %v", startTime, endTime)
	}

	var avg string
	switch {
	case hours <= 24:
		avg = "0"
	case hours <= 168: // 1 haftaya kadar
		avg = "300" // 5 dakikalık ortalama
	case hours <= 744: // 1 aya kadar
		avg = "3600" // Saatlik ortalama
	default: // 1 aydan fazla
		avg = "86400" // Günlük ortalama
	}

<<<<<<< HEAD
	// Debug log için ISO formatında tarih göster
	backend.Logger.Debug(fmt.Sprintf("Average: %v, Total Hours: %v, Start Date (ISO): %v, End Date (ISO): %v",
		avg,
		hours,
		startTime.Format(time.RFC3339),
		endTime.Format(time.RFC3339)))

=======
	// Set up API request parameters
>>>>>>> b7ec34b15515724822d7961b43e74d64b1be22b5
	params := map[string]string{
		"id":      sensorID,
		"columns": "datetime,value_",
		"sdate":   sdate,
		"edate":   edate,
		"count":   "50000",
		"avg":     avg,
		/* "pctshow":    "false",
		"pctmode":    "false", */
		"usecaption": "1",
	}

	body, err := a.baseExecuteRequest("historicdata.json", params)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch historical data: %w", err)
	}

	var response PrtgHistoricalDataResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.HistData) == 0 {
		return nil, fmt.Errorf("no data found for the given time range")
	}

	return &response, nil
}
