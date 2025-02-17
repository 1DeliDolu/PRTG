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

// query processes a single query. If QueryType is "metrics", it creates a time series,
// otherwise property-based queries are handled by handlePropertyQuery.
func (d *Datasource) query(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	_ = ctx  // ! Unused parameter: ctx is intentionally not used.
	_ = pCtx // ! Unused parameter: pCtx is intentionally not used.

	var response backend.DataResponse
	var qm queryModel

	if err := json.Unmarshal(query.JSON, &qm); err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("JSON unmarshal error: %v", err))
	}

	switch qm.QueryType {
	case "metrics":
		// Metrics handling code
		fromTime := query.TimeRange.From.UnixMilli()
		toTime := query.TimeRange.To.UnixMilli()

		historicalData, err := d.api.GetHistoricalData(qm.ObjectId, fromTime, toTime)
		if err != nil {
			backend.Logger.Error("API request failed", "error", err)
			return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("API request failed: %v", err))
		}

		// Assumption: historicalData.Treesize contains the value from the JSON ("treesize")
		times := make([]time.Time, 0)
		values := make([]float64, 0)

		for _, item := range historicalData.HistData {
			parsedTime, _, err := parsePRTGDateTime(item.Datetime)
			if err != nil {
				backend.Logger.Warn("Date parsing failed", "datetime", item.Datetime, "error", err)
				continue
			}
			if val, ok := item.Value[qm.Channel]; ok {
				switch v := val.(type) {
				case float64:
					values = append(values, v)
				case string:
					if floatVal, err := strconv.ParseFloat(v, 64); err == nil {
						values = append(values, floatVal)
					} else {
						backend.Logger.Warn("Cannot convert value to float64", "value", v, "error", err)
						continue
					}
				default:
					backend.Logger.Warn("Unexpected value type", "type", fmt.Sprintf("%T", v), "value", v)
					continue
				}
				times = append(times, parsedTime)
			} else {
				backend.Logger.Warn("Channel not found in item.Value, using default value", "channel", qm.Channel)
				times = append(times, parsedTime)
				values = append(values, 0.0)
			}
		}

		var parts []string
		if qm.IncludeGroupName && qm.Group != "" {
			parts = append(parts, qm.Group)
		}
		if qm.IncludeDeviceName && qm.Device != "" {
			parts = append(parts, qm.Device)
		}
		if qm.IncludeSensorName && qm.Sensor != "" {
			parts = append(parts, qm.Sensor)
		}
		parts = append(parts, qm.Channel)
		displayName := strings.Join(parts, " - ")

		frame := data.NewFrame("response",
			data.NewField("Time", nil, times),
			data.NewField("Value", nil, values).SetConfig(&data.FieldConfig{
				DisplayName: displayName,
			}),
		)

		response.Frames = append(response.Frames, frame)

	case "text":
		// Handle text mode by using the non-raw property
		return d.handlePropertyQuery(qm, qm.FilterProperty)

	case "raw":
		// Handle raw mode by appending "_raw" to the filter property
		rawProperty := qm.FilterProperty
		if !strings.HasSuffix(rawProperty, "_raw") {
			rawProperty += "_raw"
		}
		return d.handlePropertyQuery(qm, rawProperty)

	default:
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("Unknown query type: %s", qm.QueryType))
	}

	return response
}

// handlePropertyQuery processes a property query based on the queryModel (qm)
// and a filter property.
func (d *Datasource) handlePropertyQuery(qm queryModel, filterProperty string) backend.DataResponse {
	var response backend.DataResponse
	var times []time.Time
	var values []interface{}

	if !d.isValidPropertyType(qm.Property) {
		return backend.ErrDataResponse(backend.StatusBadRequest, "Invalid property type")
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

				// Retrieve the property value based on filterProperty
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
					times = append(times, timestamp)
					values = append(values, value)
				}
			}
		}

	case "device":
		// Similar structure for devices
		devices, err := d.api.GetDevices()
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
					times = append(times, timestamp)
					values = append(values, value)
				}
			}
		}

	case "sensor":
		sensors, err := d.api.GetSensors()
		if err != nil {
			return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("API request failed: %v", err))
		}
		for _, s := range sensors.Sensors {
			if s.Sensor == qm.Sensor {
				timestamp, _, err := parsePRTGDateTime(s.Datetime)
				if err != nil {
					backend.Logger.Error("Failed to parse sensor datetime",
						"sensor", s.Sensor,
						"datetime", s.Datetime,
						"error", err)
					continue
				}

				// Retrieve the value based on filterProperty
				var value interface{}
				switch filterProperty {
				case "status", "status_raw":
					if filterProperty == "status_raw" {
						value = float64(s.StatusRAW) // Convert to float64 for consistent graphing
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
					times = append(times, timestamp)
					values = append(values, value)
				}
			}
		}
	}

	// Create a frame with proper field configuration
	if len(times) > 0 && len(values) > 0 {
		timeField := data.NewField("Time", nil, times)

		// Determine the type of values and create an appropriate field
		var valueField *data.Field
		if len(values) > 0 {
			switch values[0].(type) {
			case float64, int:
				// Convert all values to float64
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
				// Keep string values as they are
				strVals := make([]string, len(values))
				for i, v := range values {
					strVals[i] = v.(string)
				}
				valueField = data.NewField("Value", nil, strVals)
			default:
				// Convert other types to strings
				strVals := make([]string, len(values))
				for i, v := range values {
					strVals[i] = fmt.Sprintf("%v", v)
				}
				valueField = data.NewField("Value", nil, strVals)
			}
		}

		// Set display name
		displayName := fmt.Sprintf("%s - %s (%s)", qm.Property, qm.Sensor, filterProperty)
		valueField.Config = &data.FieldConfig{
			DisplayName: displayName,
		}

		frame := data.NewFrame("response",
			timeField,
			valueField,
		)

		response.Frames = append(response.Frames, frame)
	}

	return response
}

// GetPropertyValue retrieves the property value from an item using reflection.
func (d *Datasource) GetPropertyValue(property string, item interface{}) string {
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Handle raw property requests
	isRawRequest := strings.HasSuffix(property, "_raw")
	baseProperty := strings.TrimSuffix(property, "_raw")
	fieldName := cases.Title(language.English).String(baseProperty)
	// First letter to uppercase for matching struct field names

	// Append proper suffix based on request type
	if isRawRequest {
		fieldName += "_raw" // Use lowercase suffix as in JSON
	}

	// Retrieve the field using the exact field name from JSON
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		// Try alternative field names if the first attempt fails
		alternatives := []string{
			baseProperty,               // try lowercase
			baseProperty + "_raw",      // try lowercase with raw
			strings.ToLower(fieldName), // try all lowercase
			strings.ToUpper(fieldName), // try all uppercase
			baseProperty + "_RAW",      // try uppercase RAW
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

	// Convert the value to a string based on the field's type
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

// cleanMessageHTML is a helper function that removes HTML from a message.
func cleanMessageHTML(message string) string {
	message = strings.ReplaceAll(message, `<div class="status">`, "")
	message = strings.ReplaceAll(message, `<div class="moreicon">`, "")
	message = strings.ReplaceAll(message, "</div>", "")
	return strings.TrimSpace(message)
}

// isValidPropertyType checks if the given property type and name are valid.
func (d *Datasource) isValidPropertyType(propertyType string) bool {
	validProperties := []string{
		"group", "device", "sensor", // object types
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
