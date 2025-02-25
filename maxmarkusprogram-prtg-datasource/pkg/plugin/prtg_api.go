package plugin

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
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

/* ====================================== URL BUILDER ====================================== */
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

/* =================================== REQUEST EXECUTOR ====================================== */
func (a *Api) baseExecuteRequest(endpoint string, params map[string]string) ([]byte, error) {
	ctx := context.Background()
	var responseBody []byte
	var responseErr error

	err := wrapAPICall(ctx, endpoint, "GET", params, func() error {
		startTime := time.Now()

		// Track API request
		defer func() {
			duration := time.Since(startTime).Seconds()
			observeAPIRequestDuration(endpoint, duration)
		}()

		// Check cache
		apiUrl, err := a.buildApiUrl(endpoint, params)
		if err != nil {
			incrementErrors("url_build")
			return fmt.Errorf("failed to build URL: %w", err)
		}

		if a.cacheTime > 0 {
			a.cacheMu.RLock()
			if item, ok := a.cache[apiUrl]; ok && time.Now().Before(item.expiry) {
				a.cacheMu.RUnlock()
				incrementCacheMetric(true, endpoint)
				responseBody = item.data
				return nil
			}
			a.cacheMu.RUnlock()
			incrementCacheMetric(false, endpoint)
		}

		client := &http.Client{
			Timeout: a.timeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}

		req, err := http.NewRequest("GET", apiUrl, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusForbidden {
			log.DefaultLogger.Error("Access denied: please verify API token and permissions")
			return fmt.Errorf("access denied: please verify API token and permissions")
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		if a.cacheTime > 0 {
			a.cacheMu.Lock()
			a.cache[apiUrl] = cacheItem{
				data:   body,
				expiry: time.Now().Add(a.cacheTime),
			}
			a.cacheMu.Unlock()
		}

		if resp.StatusCode == http.StatusOK {
			incrementAPIRequests(endpoint, "success")
		} else {
			incrementAPIRequests(endpoint, "error")
			incrementErrors("http_" + strconv.Itoa(resp.StatusCode))
		}

		responseBody = body
		responseErr = err
		return err
	})

	if err != nil {
		return nil, err
	}

	return responseBody, responseErr
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

/* ====================================== HISTORY HANDLER ======================================= */
func (a *Api) GetHistoricalData(sensorID string, startDate, endDate time.Time) (*PrtgHistoricalDataResponse, error) {
	if sensorID == "" {
		return nil, fmt.Errorf("invalid query: missing sensor ID")
	}

	startTime := startDate.Add(time.Hour)
	endTime := endDate.Add(time.Hour)

	const format = "2006-01-02-15-04-05"
	sdate := startTime.Format(format)
	edate := endTime.Format(format)

	hours := endTime.Sub(startTime).Hours()

	if hours <= 0 {
		return nil, fmt.Errorf("invalid time range: start date %v must be before end date %v", startTime, endTime)
	}

	var avg string
	switch {
	case hours <= 24:
		avg = "0"
	case hours <= 168:
		avg = "300"
	case hours <= 744:
		avg = "3600"
	default:
		avg = "86400"
	}

	backend.Logger.Debug(fmt.Sprintf("Average: %v, Total Hours: %v, Start Date (ISO): %v, End Date (ISO): %v",
		avg,
		hours,
		startTime.Format(time.RFC3339),
		endTime.Format(time.RFC3339)))

	params := map[string]string{
		"id":         sensorID,
		"columns":    "datetime,value_",
		"sdate":      sdate,
		"edate":      edate,
		"count":      "50000",
		"avg":        avg,
		"pctshow":    "false",
		"pctmode":    "false",
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

/* ====================================== MANUAL METHOD HANDLER ================================= */
func (a *Api) ExecuteManualMethod(method string, objectId string) (*ManualResponse, error) {
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

	return &ManualResponse{
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
