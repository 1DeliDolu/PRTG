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

// PRTGAPI defines the interface for API operations.
type PRTGAPI interface {
	GetGroups() (*PrtgGroupListResponse, error)
	GetDevices() (*PrtgDevicesListResponse, error)
	GetSensors() (*PrtgSensorsListResponse, error)
	// Additional methods like GetTextData, GetPropertyData, etc. can be declared here.
}



/* ##################################### query ##################################################################### */

func (d *Datasource) query(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	var qm queryModel

	if err := json.Unmarshal(query.JSON, &qm); err != nil {
		return backend.DataResponse{
			Frames: []*data.Frame{
				data.NewFrame(fmt.Sprintf("error_%s", query.RefID)),
			},
		}
	}

	baseFrameName := fmt.Sprintf("query_%s_%s", query.RefID, qm.QueryType)

	switch qm.QueryType {
	case "metrics":
		return d.handleMetricsQuery(qm, query.TimeRange, baseFrameName)
	case "text", "raw":
		return d.handlePropertyQuery(qm, qm.Property, qm.FilterProperty, baseFrameName)
	default:
		return backend.DataResponse{
			Frames: []*data.Frame{
				data.NewFrame(fmt.Sprintf("%s_unknown", baseFrameName)),
			},
		}
	}
}

/* ################################################ handleMetricsQuery #########################################################*/
func (d *Datasource) handleMetricsQuery(qm queryModel, timeRange backend.TimeRange, baseFrameName string) backend.DataResponse {
	var response backend.DataResponse

	historicalData, err := d.api.GetHistoricalData(qm.SensorId, timeRange.From.UTC(), timeRange.To.UTC())
	if err != nil {
		return backend.DataResponse{
			Frames: []*data.Frame{
				data.NewFrame(fmt.Sprintf("%s_error", baseFrameName)),
			},
		}
	}

	var channels []string
	if len(qm.Channels) > 0 {
		channels = qm.Channels
	} else if qm.Channel != "" {
		channels = []string{qm.Channel}
	}

	for _, channelName := range channels {
		if channelName == "" {
			continue
		}

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

		frameName := fmt.Sprintf("%s_%s", baseFrameName, channelName)

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

		frame := data.NewFrame(frameName,
			data.NewField("Time", nil, timesM),
			data.NewField("Value", nil, valuesM).SetConfig(&data.FieldConfig{
				DisplayName: displayName,
			}),
		)

		response.Frames = append(response.Frames, frame)
	}

	if len(response.Frames) == 0 {
		response.Frames = append(response.Frames, data.NewFrame(fmt.Sprintf("%s_empty", baseFrameName)))
	}

	return response
}

/* ################################################ handlePropertyQuery #########################################################*/
func (d *Datasource) handlePropertyQuery(qm queryModel, property, filterProperty string, baseFrameName string) backend.DataResponse {

	if qm.Property == "" || filterProperty == "" {
		return backend.DataResponse{
			Frames: []*data.Frame{
				data.NewFrame(fmt.Sprintf("%s_missing_properties", baseFrameName)),
			},
		}
	}

	if qm.QueryType == "raw" && !strings.HasSuffix(filterProperty, "_raw") {
		filterProperty += "_raw"
	}

	valuesRT := make([]interface{}, 0)
	timesRT := make([]time.Time, 0)

	if !d.isValidPropertyType(qm.Property) {
		return backend.DataResponse{
			Frames: []*data.Frame{
				data.NewFrame(fmt.Sprintf("%s %s_invalid_property", property, baseFrameName)),
			},
		}
	}

	switch qm.Property {
	case "group":
		groups, err := d.api.GetGroups()
		if err != nil {
			return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("API request failed: %v", err))
		}
		for _, g := range groups.Groups {
			if g.Group == qm.Group {
				timestamp, _, err := parsePRTGDateTime(g.Datetime)
				if err != nil {
					backend.Logger.Warn("Date parsing failed", "datetime", g.Datetime, "error", err)
					continue
				}

				var value interface{}
				switch filterProperty {
				case "active":
					value = g.Active
				case "active_raw":
					value = g.ActiveRAW
				case "message":
					value = cleanMessageHTML(g.Message)
				case "message_raw":
					value = g.MessageRAW
				case "priority":
					value = g.Priority
				case "priority_raw":
					value = g.PriorityRAW
				case "status":
					value = g.Status
				case "status_raw":
					value = g.StatusRAW
				case "tags":
					value = g.Tags
				case "tags_raw":
					value = g.TagsRAW
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
		devices, err := d.api.GetDevices(qm.Group) // Pass the group parameter
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
		sensors, err := d.api.GetSensors(qm.Device) // Pass the device parameter
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

/* ####################################### createPropertyFrame ################################################## */
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

/* ###################################### cleanMessageHTML ############################################################ */
func cleanMessageHTML(message string) string {
	message = strings.ReplaceAll(message, `<div class="status">`, "")
	message = strings.ReplaceAll(message, `<div class="moreicon">`, "")
	message = strings.ReplaceAll(message, "</div>", "")
	return strings.TrimSpace(message)
}

/* ######################################## isValidPropertyType ########################################################## */
func (d *Datasource) isValidPropertyType(propertyType string) bool {
	validProperties := []string{
		"group", "device", "sensor",
		"status", "status_raw",
		"message", "message_raw",
		"active", "active_raw",
		"priority", "priority_raw",
		"tags", "tags_raw",
	}

	propertyType = strings.ToLower(propertyType)
	for _, valid := range validProperties {
		if propertyType == valid {
			return true
		}
	}
	return false
}
