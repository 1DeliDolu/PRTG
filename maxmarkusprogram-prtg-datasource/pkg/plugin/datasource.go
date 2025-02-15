package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/maxmarkusprogram/prtg/pkg/models"
)

// Ensure Datasource implements required Grafana SDK interfaces
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
	_ backend.CallResourceHandler   = (*Datasource)(nil)
)

// NewDatasource creates a new datasource instance by extracting data from plugin settings.
func NewDatasource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	config, err := models.LoadPluginSettings(settings)
	if err != nil {
		return nil, err
	}
	baseURL := fmt.Sprintf("https://%s", config.Path)
	backend.Logger.Info("Base URL", "url", baseURL)

	// If cache time is not defined, default to 30 seconds
	cacheTime := config.CacheTime
	if cacheTime <= 0 {
		cacheTime = 30 * time.Second
	}

	return &Datasource{
		baseURL: baseURL,
		api:     NewApi(baseURL, config.Secrets.ApiKey, cacheTime, 10*time.Second),
	}, nil
}

// Dispose is called when the datasource settings are changed.
func (d *Datasource) Dispose() {
}

// QueryData processes incoming queries and returns the results.
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()

	// Call the query method for each query.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)
		response.Responses[q.RefID] = res
	}

	return response, nil
}

// parsePRTGDateTime parses PRTG datetime strings in various formats.
func parsePRTGDateTime(datetime string) (time.Time, string, error) {
	// Try different known PRTG date formats
	layouts := []string{
		"02.01.2006 15:04:05",
		time.RFC3339,
	}
	var parseErr error
	for _, layout := range layouts {
		parsedTime, err := time.Parse(layout, datetime)
		if err == nil {
			unixTime := parsedTime.Unix()
			return parsedTime, strconv.FormatInt(unixTime, 10), nil
		}
		parseErr = err
	}
	backend.Logger.Error("Date parsing failed for all formats",
		"datetime", datetime,
		"error", parseErr)
	return time.Time{}, "", fmt.Errorf("failed to parse time '%s': %w", datetime, parseErr)
}

// Annotations returns annotations for a time series.
func (d *Datasource) AnnotationsQueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()

	for _, query := range req.Queries {
		var qm queryModel
		if err := json.Unmarshal(query.JSON, &qm); err != nil {
			return nil, err
		}
		fromTime := query.TimeRange.From.UnixMilli()
		toTime := query.TimeRange.To.UnixMilli()
		for _, q := range req.Queries {
			// Example: Fetch annotation events from PRTG API (Modify as per actual API)
			annotations, err := d.api.GetAnnotations(fromTime, toTime, qm.ObjectId) // Replace with actual API call
			if err != nil {
				backend.Logger.Error("Failed to fetch annotations", "error", err)
				response.Responses[q.RefID] = backend.DataResponse{
					Error: err,
				}
				continue
			}
			// Convert API response to Grafana annotation format
			frame := data.NewFrame("annotations",
				data.NewField("time", nil, []time.Time{}),
				data.NewField("text", nil, []string{}),
				data.NewField("tags", nil, [][]string{}),
			)

			for _, annotation := range annotations {
				timestamp := annotation.Time

				frame.AppendRow(timestamp, annotation.Text, annotation.Tags)
			}

			response.Responses[q.RefID] = backend.DataResponse{
				Frames: []*data.Frame{frame},
			}
		}
	}
	return response, nil
}

// CheckHealth checks the plugin configuration.
func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	res := &backend.CheckHealthResult{}

	// Load configuration
	config, err := models.LoadPluginSettings(*req.PluginContext.DataSourceInstanceSettings)
	if err != nil {
		res.Status = backend.HealthStatusError
		res.Message = "Unable to load settings"
		return res, nil
	}

	// Check API key
	if config.Secrets.ApiKey == "" {
		res.Status = backend.HealthStatusError
		res.Message = "API key is missing"
		return res, nil
	}

	// Get PRTG status including version
	status, err := d.api.GetStatusList()
	if err != nil {
		res.Status = backend.HealthStatusError
		res.Message = fmt.Sprintf("Failed to get PRTG status: %v", err)
		return res, nil
	}

	// Return success with version information
	res.Status = backend.HealthStatusOk
	res.Message = fmt.Sprintf("Data source is working. PRTG Version: %s", status.Version)
	return res, nil
}

// CallResource routes requests to the appropriate handlers based on the URL path.
func (d *Datasource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	pathParts := strings.Split(req.Path, "/")
	switch pathParts[0] {
	case "groups":
		return d.handleGetGroups(sender)
	case "devices":
		return d.handleGetDevices(sender)
	case "sensors":
		return d.handleGetSensors(sender)
	case "channels":
		if len(pathParts) < 2 {
			errorResponse := map[string]string{"error": "missing objid parameter"}
			errorJSON, _ := json.Marshal(errorResponse)
			return sender.Send(&backend.CallResourceResponse{
				Status:  http.StatusBadRequest,
				Headers: map[string][]string{"Content-Type": {"application/json"}},
				Body:    errorJSON,
			})
		}
		return d.handleGetChannel(sender, pathParts[1])
	default:
		return sender.Send(&backend.CallResourceResponse{Status: http.StatusNotFound})
	}
}

func (d *Datasource) handleGetGroups(sender backend.CallResourceResponseSender) error {
	groups, err := d.api.GetGroups()
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte(err.Error()),
		})
	}
	body, err := json.Marshal(groups)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte(fmt.Sprintf("error marshaling groups: %v", err)),
		})
	}
	return sender.Send(&backend.CallResourceResponse{
		Status:  http.StatusOK,
		Headers: map[string][]string{"Content-Type": {"application/json"}},
		Body:    body,
	})
}

func (d *Datasource) handleGetDevices(sender backend.CallResourceResponseSender) error {
	devices, err := d.api.GetDevices()
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte(err.Error()),
		})
	}
	body, err := json.Marshal(devices)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte(fmt.Sprintf("error marshaling devices: %v", err)),
		})
	}
	return sender.Send(&backend.CallResourceResponse{
		Status:  http.StatusOK,
		Headers: map[string][]string{"Content-Type": {"application/json"}},
		Body:    body,
	})
}

func (d *Datasource) handleGetSensors(sender backend.CallResourceResponseSender) error {
	sensors, err := d.api.GetSensors()
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte(err.Error()),
		})
	}
	body, err := json.Marshal(sensors)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte(fmt.Sprintf("error marshaling sensors: %v", err)),
		})
	}
	return sender.Send(&backend.CallResourceResponse{
		Status:  http.StatusOK,
		Headers: map[string][]string{"Content-Type": {"application/json"}},
		Body:    body,
	})
}

func (d *Datasource) handleGetChannel(sender backend.CallResourceResponseSender, objid string) error {
	if objid == "" {
		errorResponse := map[string]string{"error": "missing objid parameter"}
		errorJSON, _ := json.Marshal(errorResponse)
		return sender.Send(&backend.CallResourceResponse{
			Status:  http.StatusBadRequest,
			Headers: map[string][]string{"Content-Type": {"application/json"}},
			Body:    errorJSON,
		})
	}
	channels, err := d.api.GetChannels(objid)
	if err != nil {
		errorResponse := map[string]string{"error": err.Error()}
		errorJSON, _ := json.Marshal(errorResponse)
		return sender.Send(&backend.CallResourceResponse{
			Status:  http.StatusInternalServerError,
			Headers: map[string][]string{"Content-Type": {"application/json"}},
			Body:    errorJSON,
		})
	}
	body, err := json.Marshal(channels)
	if err != nil {
		errorResponse := map[string]string{"error": fmt.Sprintf("error marshaling channels: %v", err)}
		errorJSON, _ := json.Marshal(errorResponse)
		return sender.Send(&backend.CallResourceResponse{
			Status:  http.StatusInternalServerError,
			Headers: map[string][]string{"Content-Type": {"application/json"}},
			Body:    errorJSON,
		})
	}
	return sender.Send(&backend.CallResourceResponse{
		Status:  http.StatusOK,
		Headers: map[string][]string{"Content-Type": {"application/json"}},
		Body:    body,
	})
}
