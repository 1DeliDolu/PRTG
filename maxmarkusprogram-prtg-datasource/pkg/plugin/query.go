package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

/* =================================== QUERY HANDLER ========================================== */
// Modified to allow for mocking in tests
func (d *Datasource) query(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	// Parse cache time from query JSON
	var qm struct {
		CacheTime int64 `json:"cacheTime"`
	}
	if err := json.Unmarshal(query.JSON, &qm); err != nil {
		qm.CacheTime = 6000 // Default to 6 seconds if not specified
	}

	// Generate cache key including time range
	cacheKey := QueryCacheKey{
		RefID:     query.RefID,
		QueryType: query.QueryType,
		TimeRange: fmt.Sprintf("%v-%v", query.TimeRange.From.Unix(), query.TimeRange.To.Unix()),
	}

	// Check cache with proper expiration
	d.cacheMutex.RLock()
	if cached, exists := d.queryCache[cacheKey.String()]; exists &&
		time.Now().Before(cached.ValidUntil) {
		d.cacheMutex.RUnlock()
		return cached.Response
	}
	d.cacheMutex.RUnlock()

	// Execute query
	response := d.executeQuery(ctx, pCtx, query)

	// Cache successful responses
	if response.Error == nil {
		d.cacheMutex.Lock()
		d.queryCache[cacheKey.String()] = &QueryCacheEntry{
			Response:   response,
			ValidUntil: time.Now().Add(time.Duration(qm.CacheTime) * time.Millisecond),
		}
		d.cacheMutex.Unlock()
	}

	return response
}

// Add this helper method
func (d *Datasource) executeQuery(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	// Start tracing
	ctx, span := d.tracer.StartSpan(ctx, "query")
	defer span.End()
	backend.Logger.Info("PluginContext", "pCtx", pCtx)

	// Start timing and initial logging
	start := time.Now()
	d.logger.Info("Starting query execution",
		"queryType", query.QueryType,
		"refID", query.RefID,
		"timeRange", fmt.Sprintf("%v to %v", query.TimeRange.From, query.TimeRange.To),
	)

	// Parse query model
	var qm queryModel
	if err := json.Unmarshal(query.JSON, &qm); err != nil {
		d.logger.Error("Query parsing failed",
			"error", err,
			"raw_query", string(query.JSON),
		)
		d.metrics.IncError("query_parse_error")
		recordError(span, err, "Failed to parse query")
		return backend.ErrDataResponse(backend.StatusBadRequest, "failed to parse query")
	}
	// Generate stable cache key that includes time range and refId
	cacheKey := QueryCacheKey{
		RefID:      query.RefID,
		QueryType:  query.QueryType,
		SensorID:   qm.SensorId,
		Channel:    strings.Join(qm.ChannelArray, ","),
		TimeRange:  fmt.Sprintf("%v-%v", query.TimeRange.From.Unix(), query.TimeRange.To.Unix()),
		Property:   qm.Property,
		Parameters: fmt.Sprintf("%s_%s_%s", qm.Group, qm.Device, qm.Sensor), // Add unique identifiers
	}

	// Get cache duration from API
	cacheTime := d.api.GetCacheTime()

	// Calculate dynamic cache duration based on time range
	timeRange := query.TimeRange.To.Sub(query.TimeRange.From)
	var cacheDuration time.Duration

	switch {
	case timeRange <= time.Hour:
		cacheDuration = 6 * time.Second
	case timeRange <= 24*time.Hour:
		cacheDuration = 30 * time.Second
	default:
		cacheDuration = cacheTime
	}

	// Use String() method to convert cacheKey to string
	cacheKeyStr := cacheKey.String()

	// Check cache with proper expiration
	d.cacheMutex.RLock()
	if entry, exists := d.queryCache[cacheKeyStr]; exists && time.Now().Before(entry.ValidUntil) {
		d.cacheMutex.RUnlock()
		return entry.Response
	}
	d.cacheMutex.RUnlock()

	// Add query attributes to span
	addQueryAttributes(span, qm)

	// Defer metrics and logging
	defer func() {
		duration := time.Since(start).Seconds()
		d.metrics.ObserveQueryDuration(qm.QueryType, duration)
		d.logger.Info("Query completed",
			"duration", duration,
			"queryType", qm.QueryType,
			"refID", query.RefID,
		)
	}()

	// Execute query based on type
	var response backend.DataResponse
	switch qm.QueryType {
	case "metrics":
		if qm.Channel == "" && len(qm.ChannelArray) == 0 {
			d.logger.Error("Channel selection required for metrics query")
			d.metrics.IncError("missing_channel")
			return backend.ErrDataResponse(backend.StatusBadRequest, "channel selection required")
		}
		response = d.handleMetricsQuery(ctx, qm, query.TimeRange, fmt.Sprintf("metrics_%s", query.RefID))

		// Cache metrics queries for shorter duration to maintain stability
		if response.Error == nil {
			d.cacheMutex.Lock()
			d.queryCache[cacheKey.String()] = &QueryCacheEntry{
				Response:   response,
				ValidUntil: time.Now().Add(25 * time.Second), // Cache for 5 seconds
				Updating:   false,
			}
			d.cacheMutex.Unlock()
		}

	case "manual":
		d.logger.Debug("Executing manual query",
			"method", qm.ManualMethod,
			"objectId", qm.ManualObjectId,
		)
		response = d.handleManualQuery(qm, query.TimeRange, fmt.Sprintf("manual_%s", query.RefID))

	case "text", "raw":
		response = d.handlePropertyQuery(ctx, qm, qm.Property, qm.FilterProperty, fmt.Sprintf("property_%s", query.RefID))

	default:
		d.logger.Warn("Unknown query type",
			"type", qm.QueryType,
			"refID", query.RefID,
		)
		d.metrics.IncError("unknown_query_type")
		return backend.DataResponse{
			Frames: []*data.Frame{
				data.NewFrame(fmt.Sprintf("unknown_%s", query.RefID)),
			},
		}
	}

	// Cache response with proper duration
	if response.Error == nil {
		d.cacheMutex.Lock()
		d.queryCache[cacheKeyStr] = &QueryCacheEntry{
			Response:   response,
			ValidUntil: time.Now().Add(cacheDuration),
			Updating:   false,
		}
		d.cacheMutex.Unlock()

		d.logger.Debug("Cached response",
			"key", cacheKeyStr,
			"duration", cacheDuration,
		)
	}

	// Record any errors in the response
	if response.Error != nil {
		d.logger.Error("Query execution failed",
			"error", response.Error,
			"queryType", qm.QueryType,
			"refID", query.RefID,
		)
		d.metrics.IncError("query_execution")
		recordError(span, response.Error, "Query execution failed")
	}

	return response
}

/* =================================== METRICS HANDLER ======================================== */
func (d *Datasource) handleMetricsQuery(ctx context.Context, qm queryModel, timeRange backend.TimeRange, baseFrameName string) backend.DataResponse {
	_, span := d.tracer.StartSpan(ctx, "handleMetricsQuery")
	defer span.End()

	queryStart := time.Now()
	d.logger.Debug("Fetching historical data",
		"sensorId", qm.SensorId,
		"timeRange", fmt.Sprintf("%v to %v", timeRange.From, timeRange.To),
		"channels", qm.ChannelArray,
	)

	// Initialize response
	response := backend.DataResponse{
		Frames: make([]*data.Frame, 0),
	}

	// Fetch historical data once for all channels
	historicalData, err := d.api.GetHistoricalData(qm.SensorId, timeRange.From.UTC(), timeRange.To.UTC())
	if err != nil {
		d.logger.Error("Failed to fetch historical data",
			"error", err,
			"sensorId", qm.SensorId,
		)
		d.metrics.IncError("historical_data_fetch")
		recordError(span, err, "Failed to fetch historical data")
		return backend.ErrDataResponse(backend.StatusInternal, "failed to fetch data")
	}

	// Check if we have channels to process
	if len(qm.ChannelArray) == 0 && qm.Channel == "" {
		d.logger.Error("No channels specified")
		d.metrics.IncError("missing_channel")
		return backend.ErrDataResponse(backend.StatusBadRequest, "channel selection required")
	}
	// Use ChannelArray if available, otherwise fall back to single channel
	channels := qm.ChannelArray
	if len(channels) == 0 && qm.Channel != "" {
		channels = []string{qm.Channel}
	}

	// If multiple channels are selected, create a single frame with multiple series
	if len(channels) > 1 {
		// Create a single frame with time field and multiple value fields
		timesM := make([]time.Time, 0)
		channelData := make(map[string][]float64)
		
		// Initialize channel data maps
		for _, channelName := range channels {
			channelData[channelName] = make([]float64, 0)
		}

		if historicalData != nil && len(historicalData.HistData) > 0 {
			for _, item := range historicalData.HistData {
				parsedTime, _, err := parsePRTGDateTime(item.Datetime)
				if err != nil {
					continue
				}

				// Check if we have data for any of the requested channels
				hasData := false
				tempValues := make(map[string]float64)
				
				for _, channelName := range channels {
					if val, exists := item.Value[channelName]; exists {
						var floatVal float64
						switch v := val.(type) {
						case float64:
							floatVal = v
						case string:
							if parsed, err := strconv.ParseFloat(v, 64); err == nil {
								floatVal = parsed
							} else {
								continue
							}
						default:
							continue
						}
						tempValues[channelName] = floatVal
						hasData = true
					}
				}

				if hasData {
					timesM = append(timesM, parsedTime)
					// Add values for each channel (use 0 or NaN for missing values)
					for _, channelName := range channels {
						if val, exists := tempValues[channelName]; exists {
							channelData[channelName] = append(channelData[channelName], val)
						} else {
							// Use NaN for missing values to maintain alignment
							channelData[channelName] = append(channelData[channelName], 0)
						}
					}
				}
			}
		}

		// Create frame with time field
		fields := []*data.Field{
			data.NewField("Time", nil, timesM),
		}
		// Add a field for each channel
		for _, channelName := range channels {
			displayName := channelName
			if qm.IncludeGroupName && qm.Group != "" {
				displayName = fmt.Sprintf("%s - %s", qm.Group, displayName)
			}
			if qm.IncludeDeviceName && qm.Device != "" {
				displayName = fmt.Sprintf("%s - %s", qm.Device, displayName)
			}
			if qm.IncludeSensorName && qm.Sensor != "" {
				displayName = fmt.Sprintf("%s - %s", qm.Sensor, displayName)
			}

			field := data.NewField(channelName, nil, channelData[channelName]).SetConfig(&data.FieldConfig{
				DisplayName: displayName,
				Custom: map[string]interface{}{
					"refId":     baseFrameName,
					"channel":   channelName,
					"queryType": "multi-channel",
				},
			})
			fields = append(fields, field)
		}
		// Create single frame with all channels
		frame := data.NewFrame(fmt.Sprintf("%s_multi", baseFrameName), fields...)
		frame.Meta = &data.FrameMeta{
			Type: data.FrameTypeTimeSeriesMulti,
			Custom: map[string]interface{}{
				"from":       timeRange.From.UnixMilli(),
				"to":         timeRange.To.UnixMilli(),
				"channels":   channels,
				"stable":     true,
				"duration":   timeRange.To.Sub(timeRange.From).String(),
				"timezone":   "UTC",
				"queryType":  "multi-channel",
				"refId":      baseFrameName, // Keep refId stable
			},
		}

		response.Frames = append(response.Frames, frame)
	} else {
		// Single channel - use existing logic
		channelName := channels[0]
		timesM := make([]time.Time, 0)
		valuesM := make([]float64, 0)

		if historicalData != nil && len(historicalData.HistData) > 0 {
			for _, item := range historicalData.HistData {
				parsedTime, _, err := parsePRTGDateTime(item.Datetime)
				if err != nil {
					continue
				}

				if val, exists := item.Value[channelName]; exists {
					var floatVal float64
					switch v := val.(type) {
					case float64:
						floatVal = v
					case string:
						if parsed, err := strconv.ParseFloat(v, 64); err == nil {
							floatVal = parsed
						} else {
							continue
						}
					default:
						continue
					}

					timesM = append(timesM, parsedTime)
					valuesM = append(valuesM, floatVal)
				}
			}
		}

		// Build display name with optional prefixes
		displayName := channelName
		if qm.IncludeGroupName && qm.Group != "" {
			displayName = fmt.Sprintf("%s - %s", qm.Group, displayName)
		}
		if qm.IncludeDeviceName && qm.Device != "" {
			displayName = fmt.Sprintf("%s - %s", qm.Device, displayName)
		}
		if qm.IncludeSensorName && qm.Sensor != "" {
			displayName = fmt.Sprintf("%s - %s", qm.Sensor, displayName)
		}		// Create frame for single channel
		frame := data.NewFrame(fmt.Sprintf("%s_single", baseFrameName),
			data.NewField("Time", nil, timesM),
			data.NewField("Value", nil, valuesM).SetConfig(&data.FieldConfig{
				DisplayName: displayName,
				Custom: map[string]interface{}{
					"refId":     baseFrameName,
					"channel":   channelName,
					"queryType": "single-channel",
				},
			}),
		)

		frame.Meta = &data.FrameMeta{
			Type: data.FrameTypeTimeSeriesMulti,
			Custom: map[string]interface{}{
				"from":       timeRange.From.UnixMilli(),
				"to":         timeRange.To.UnixMilli(),
				"channel":    channelName,
				"stable":     true,
				"duration":   timeRange.To.Sub(timeRange.From).String(),
				"timezone":   "UTC",
				"queryType":  "single-channel",
				"refId":      baseFrameName, // Keep refId stable
			},
		}

		response.Frames = append(response.Frames, frame)
	}

	// If no frames were created, add an empty frame
	if len(response.Frames) == 0 {
		response.Frames = append(response.Frames, data.NewFrame(fmt.Sprintf("%s_empty", baseFrameName)))
	}

	duration := time.Since(queryStart)
	d.metrics.ObserveAPILatency("historical_data", duration.Seconds())

	return response
}

/* =================================== MANUAL QUERY HANDLER =================================== */
func (d *Datasource) handleManualQuery(qm queryModel, timeRange backend.TimeRange, frameBaseName string) backend.DataResponse {
	d.logger.Debug("Processing manual query",
		"method", qm.ManualMethod,
		"objectId", qm.ManualObjectId,
		"timeRange", fmt.Sprintf("%v to %v", timeRange.From, timeRange.To),
	)

	if qm.ManualMethod == "" {
		d.logger.Error("Manual method is required")
		d.metrics.IncError("missing_manual_method")
		return backend.ErrDataResponse(backend.StatusBadRequest, "manual method is required")
	}

	response, err := d.api.ExecuteManualMethod(qm.ManualMethod, qm.ManualObjectId)
	if err != nil {
		d.logger.Error("Manual query failed",
			"error", err,
			"method", qm.ManualMethod,
		)
		d.metrics.IncError("manual_query_failed")
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("API request failed: %v", err))
	}

	keys := make([]string, len(response.KeyValues))
	values := make([]string, len(response.KeyValues))

	for i, kv := range response.KeyValues {
		keys[i] = kv.Key
		switch v := kv.Value.(type) {
		case string:
			values[i] = v
		case float64:
			values[i] = strconv.FormatFloat(v, 'f', -1, 64)
		case bool:
			values[i] = strconv.FormatBool(v)
		case nil:
			values[i] = "null"
		default:
			values[i] = fmt.Sprintf("%v", v)
		}
	}

	frame := data.NewFrame(frameBaseName,
		data.NewField("Key", nil, keys).SetConfig(&data.FieldConfig{
			DisplayName: "Property",
		}),
		data.NewField("Value", nil, values).SetConfig(&data.FieldConfig{
			DisplayName: "Value",
		}),
	).SetMeta(&data.FrameMeta{
		Type:   data.FrameTypeTimeSeriesWide,
		Custom: response.Manuel,
	})

	return backend.DataResponse{
		Frames: []*data.Frame{frame},
	}
}

/* =================================== PROPERTY HANDLER ======================================= */
func (d *Datasource) handlePropertyQuery(ctx context.Context, qm queryModel, property, filterProperty string, baseFrameName string) backend.DataResponse {
	ctx, span := d.tracer.StartSpan(ctx, "handlePropertyQuery")
	backend.Logger.Info("Context", "ctx", ctx)
	defer span.End()

	d.logger.Debug("Processing property query",
		"property", property,
		"filterProperty", filterProperty,
	)

	// Raw mod kontrolü
	isRawMode := qm.QueryType == "raw"
	if isRawMode && !strings.HasSuffix(filterProperty, "_raw") {
		filterProperty += "_raw"
		d.logger.Debug("Converting to raw property",
			"original", property,
			"rawProperty", filterProperty,
		)
	}

	var timesRT []time.Time
	var valuesRT []interface{}

	switch property {
	case "group":
		groups, err := d.api.GetGroups()
		if err != nil {
			return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("API request failed: %v", err))
		}
		for _, g := range groups.Groups {
			if g.Group == qm.Group {
				timestamp, _, err := parsePRTGDateTime(g.Datetime)
				if err != nil {
					continue
				}

				var value interface{}
				switch filterProperty {
				case "active", "active_raw":
					value = selectRawOrFormatted(isRawMode, g.ActiveRAW, g.Active)
				case "message", "message_raw":
					value = selectRawOrFormatted(isRawMode, g.MessageRAW, cleanMessageHTML(g.Message))
				case "priority", "priority_raw":
					value = selectRawOrFormatted(isRawMode, g.PriorityRAW, g.Priority)
				case "status", "status_raw":
					value = selectRawOrFormatted(isRawMode, g.StatusRAW, g.Status)
				case "tags", "tags_raw":
					value = selectRawOrFormatted(isRawMode, g.TagsRAW, g.Tags)
				}

				if value != nil {
					timesRT = append(timesRT, timestamp.UTC())
					valuesRT = append(valuesRT, value)
				}
			}
		}
	case "device":
		if qm.Group == "" {
			return backend.ErrDataResponse(backend.StatusBadRequest, "group parameter is required for device query")
		}
		devices, err := d.api.GetDevices(qm.Group)
		if err != nil {
			return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("API request failed: %v", err))
		}
		for _, dev := range devices.Devices {
			if dev.Device == qm.Device {
				timestamp, _, err := parsePRTGDateTime(dev.Datetime)
				if err != nil {
					continue
				}

				var value interface{}
				switch filterProperty {
				case "active":
					value = dev.Active
				case "active_raw":
					value = dev.ActiveRAW
				case "message":
					value = cleanMessageHTML(dev.Message)
				case "message_raw":
					value = dev.MessageRAW
				case "priority":
					value = dev.Priority
				case "priority_raw":
					value = dev.PriorityRAW
				case "status":
					value = dev.Status
				case "status_raw":
					value = dev.StatusRAW
				case "tags":
					value = dev.Tags
				case "tags_raw":
					value = dev.TagsRAW
				}

				if value != nil {
					timesRT = append(timesRT, timestamp)
					valuesRT = append(valuesRT, value)
				}
			}
		}

	case "sensor":
		if qm.Device == "" {
			return backend.ErrDataResponse(backend.StatusBadRequest, "device parameter is required for sensor query")
		}
		sensors, err := d.api.GetSensors(qm.Device)
		if err != nil {
			return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("API request failed: %v", err))
		}

		for _, s := range sensors.Sensors {
			if s.Sensor == qm.Sensor {
				timestamp, _, err := parsePRTGDateTime(s.Datetime)
				if err != nil {
					continue
				}

				var value interface{}
				switch filterProperty {
				case "status", "status_raw":
					if filterProperty == "status_raw" {
						value = float64(s.StatusRAW)
					} else {
						value = s.Status
					}
				case "active", "active_raw":
					if filterProperty == "active_raw" {
						value = float64(s.ActiveRAW)
					} else {
						value = s.Active
					}
				case "priority", "priority_raw":
					if filterProperty == "priority_raw" {
						value = float64(s.PriorityRAW)
					} else {
						value = s.Priority
					}
				case "message", "message_raw":
					if filterProperty == "message_raw" {
						value = s.MessageRAW
					} else {
						value = cleanMessageHTML(s.Message)
					}
				case "tags", "tags_raw":
					if filterProperty == "tags_raw" {
						value = s.TagsRAW
					} else {
						value = s.Tags
					}
				}

				if value != nil {
					timesRT = []time.Time{timestamp}
					valuesRT = []interface{}{value}
					break
				}
			}
		}
	}

	frameName := fmt.Sprintf("%s_%s_%s", baseFrameName, qm.Property, filterProperty)
	frame := createPropertyFrame(timesRT, valuesRT, frameName, qm.Property, filterProperty)

	return backend.DataResponse{
		Frames: []*data.Frame{frame},
	}
}

/* =================================== FRAME CREATOR ========================================== */
func createPropertyFrame(times []time.Time, values []interface{}, frameName, property, filterProperty string) *data.Frame {
	if len(times) == 0 || len(values) == 0 {
		return data.NewFrame(frameName + "_empty")
	}

	timeField := data.NewField("Time", nil, times)
	var valueField *data.Field

	switch values[0].(type) {
	case float64, int:
		floatVals := make([]float64, len(values))
		for i, v := range values {
			switch tv := v.(type) {
			case float64:
				floatVals[i] = tv
			case int:
				floatVals[i] = float64(tv)
			}
		}
		valueField = data.NewField("Value", nil, floatVals)
	case string:
		strVals := make([]string, len(values))
		for i, v := range values {
			strVals[i] = v.(string)
		}
		valueField = data.NewField("Value", nil, strVals)
	default:
		strVals := make([]string, len(values))
		for i, v := range values {
			strVals[i] = fmt.Sprintf("%v", v)
		}
		valueField = data.NewField("Value", nil, strVals)
	}

	displayName := fmt.Sprintf("%s - (%s)", property, filterProperty)
	valueField.Config = &data.FieldConfig{
		DisplayName: displayName,
	}

	return data.NewFrame(frameName, timeField, valueField)
}

/* ###############################################  GetPropertyValue ################################################################*/
func (d *Datasource) GetPropertyValue(property string, item interface{}) string {
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	isRawRequest := strings.HasSuffix(property, "_raw")
	baseProperty := strings.TrimSuffix(property, "_raw")
	fieldName := cases.Title(language.English).String(baseProperty)

	if isRawRequest {
		fieldName += "_raw"
	}

	field := v.FieldByName(fieldName)
	if !field.IsValid() {

		alternatives := []string{
			baseProperty,
			baseProperty + "_raw",
			strings.ToLower(fieldName),
			strings.ToUpper(fieldName),
			baseProperty + "_RAW",
		}

		for _, alt := range alternatives {
			if f := v.FieldByName(alt); f.IsValid() {
				field = f
				break
			}
		}
	}

	if !field.IsValid() {
		return "Unknown"
	}

	val := field.Interface()
	switch v := val.(type) {
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		if isRawRequest {
			if v {
				return "1"
			}
			return "0"
		}
		return strconv.FormatBool(v)
	case string:
		if !isRawRequest && baseProperty == "message" {
			return cleanMessageHTML(v)
		}
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

/* ====================================== CLEAN MESSAGES HTML ====================================== */
func cleanMessageHTML(message string) string {
	message = strings.ReplaceAll(message, `<div class="status">`, "")
	message = strings.ReplaceAll(message, `<div class="moreicon">`, "")
	message = strings.ReplaceAll(message, "</div>", "")
	return strings.TrimSpace(message)
}

// Helper function to select between raw and formatted values
func selectRawOrFormatted(isRaw bool, rawValue, formattedValue interface{}) interface{} {
	if isRaw {
		return rawValue
	}
	return formattedValue
}