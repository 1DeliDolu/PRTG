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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/1DeliDolu/PRTG/maxmarkusprogram/prtg/pkg/models"
)

// Logger interface defines the logging methods required by the datasource

var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
	_ backend.CallResourceHandler   = (*Datasource)(nil)
)

/*  ################################# NewDatasource #################################################### */
func NewDatasource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	config, err := models.LoadPluginSettings(settings)
	if err != nil {
		return nil, err
	}

	baseURL := fmt.Sprintf("https://%s", config.Path)
	cacheTime := config.CacheTime
	if cacheTime <= 0 {
		cacheTime = 30 * time.Second
	}

	logger := NewLogger()
	tracer := NewTracer(logger)
	metrics := NewMetrics(prometheus.DefaultRegisterer)

	return &Datasource{
		baseURL: baseURL,
		api:     NewApi(baseURL, config.Secrets.ApiKey, cacheTime, 10*time.Second),
		logger:  logger,
		tracer:  tracer,
		metrics: metrics,
	}, nil
}


/*  ########################################### Dispose ################################################### */
func (d *Datasource) Dispose() {
}

/*  ########################################### QueryData ################################################### */
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	ctx, span := d.tracer.StartSpan(ctx, "QueryData")
	defer span.End()

	response := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		start := time.Now()
		d.logger.Debug("Processing query", "refID", q.RefID)

		res := d.query(ctx, req.PluginContext, q)
		response.Responses[q.RefID] = res

		duration := time.Since(start).Seconds()
		d.metrics.ObserveQueryDuration("query", duration)
	}

	return response, nil
}

/* ######################################## parsePRTGDateTime ##############################################################  */
func parsePRTGDateTime(datetime string) (time.Time, string, error) {
	// If the datetime contains a range (indicated by " - "), take the end time
	if strings.Contains(datetime, " - ") {
		parts := strings.Split(datetime, " - ")
		if len(parts) == 2 {
			datetime = strings.TrimSpace(parts[1])
		}
	}

	backend.Logger.Debug(fmt.Sprintf("Parsing PRTG datetime: %s", datetime))

	// PRTG sends times in local timezone, so we need to handle both formats
	layouts := []string{
		"02.01.2006 15:04:05",
		time.RFC3339,
	}

	loc, err := time.LoadLocation("Europe/Berlin") // PRTG server's timezone
	if err != nil {
		loc = time.Local // Fallback to system local time if timezone not found
	}

	var parseErr error
	for _, layout := range layouts {
		parsedTime, err := time.ParseInLocation(layout, datetime, loc)
		if err == nil {
			// Convert to UTC for consistency
			utcTime := parsedTime.UTC()
			unixTime := utcTime.Unix()
			return utcTime, strconv.FormatInt(unixTime, 10), nil
		}
		parseErr = err
	}

	backend.Logger.Error("Date parsing failed for all formats",
		"datetime", datetime,
		"error", parseErr)
	return time.Time{}, "", fmt.Errorf("failed to parse time '%s': %w", datetime, parseErr)
}

/* ######################################## CheckHealth ##############################################################  */
func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	ctx, span := d.tracer.StartSpan(ctx, "CheckHealth")
	backend.Logger.Debug("CheckHealth", "ctx", ctx)
	defer span.End()

	d.logger.Debug("Starting health check")

	res := &backend.CheckHealthResult{}

	// Load and validate settings
	config, err := models.LoadPluginSettings(*req.PluginContext.DataSourceInstanceSettings)
	if err != nil {
		d.logger.Error("Failed to load plugin settings", "error", err)
		res.Status = backend.HealthStatusError
		res.Message = fmt.Sprintf("Configuration error: %v", err)
		d.metrics.IncError("config_load")
		return res, nil
	}

	// Validate API key
	if config.Secrets.ApiKey == "" {
		d.logger.Error("API key is missing in configuration")
		res.Status = backend.HealthStatusError
		res.Message = "API key is required but not configured"
		d.metrics.IncError("missing_api_key")
		return res, nil
	}

	// Check PRTG server connectivity
	d.logger.Debug("Checking PRTG server connection", "url", d.baseURL)
	status, err := d.api.GetStatusList()
	if err != nil {
		d.logger.Error("Failed to connect to PRTG server",
			"error", err,
			"url", d.baseURL,
		)
		res.Status = backend.HealthStatusError
		res.Message = fmt.Sprintf("PRTG connection failed: %v", err)
		d.metrics.IncError("connection_failed")
		return res, nil
	}

	// Validate PRTG response
	if status == nil || status.Version == "" {
		d.logger.Error("Invalid response from PRTG server")
		res.Status = backend.HealthStatusError
		res.Message = "Invalid response from PRTG server"
		d.metrics.IncError("invalid_response")
		return res, nil
	}

	// Success case
	d.logger.Info("Health check successful",
		"prtgVersion", status.Version,
		"url", d.baseURL,
	)
	res.Status = backend.HealthStatusOk
	res.Message = fmt.Sprintf("Data source is working. PRTG Version: %s", status.Version)
	d.metrics.IncAPIRequest("health")

	return res, nil
}

/* ######################################## CallResource ##############################################################  */
func (d *Datasource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	ctx, span := d.tracer.StartSpan(ctx, "CallResource") // Now properly using ctx

	backend.Logger.Debug("CallResource", "ctx", ctx)

	defer span.End()

	start := time.Now()
	defer func() {
		duration := time.Since(start)
		d.metrics.ObserveAPILatency(req.Path, duration.Seconds())
		d.logger.Info("Resource call completed",
			"path", req.Path,
			"duration", duration,
		)
	}()

	d.logger.Debug("Resource call started",
		"path", req.Path,
		"method", req.Method,
	)

	pathParts := strings.Split(req.Path, "/")
	var err error

	switch pathParts[0] {
	case "groups":
		err = d.handleGetGroups(sender)
	case "devices":
		if len(pathParts) < 2 {
			d.logger.Error("Missing group parameter")
			errorResponse := map[string]string{"error": "group parameter is required"}
			errorJSON, _ := json.Marshal(errorResponse)
			return sender.Send(&backend.CallResourceResponse{
				Status:  http.StatusBadRequest,
				Headers: map[string][]string{"Content-Type": {"application/json"}},
				Body:    errorJSON,
			})
		}
		err = d.handleGetDevices(sender, pathParts[1])
	case "sensors":
		if len(pathParts) < 2 {
			errorResponse := map[string]string{"error": "device parameter is required"}
			errorJSON, _ := json.Marshal(errorResponse)
			return sender.Send(&backend.CallResourceResponse{
				Status:  http.StatusBadRequest,
				Headers: map[string][]string{"Content-Type": {"application/json"}},
				Body:    errorJSON,
			})
		}
		device := pathParts[1]
		err = d.handleGetSensors(sender, device)

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
		err = d.handleGetChannel(sender, pathParts[1])
	default:
		d.logger.Warn("Unknown resource path", "path", req.Path)
		return sender.Send(&backend.CallResourceResponse{Status: http.StatusNotFound})
	}

	return err
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

/* ######################################### handleGetDevices ############################################################*/
func (d *Datasource) handleGetDevices(sender backend.CallResourceResponseSender, group string) error {
	if group == "" {
		errorResponse := map[string]string{"error": "missing group parameter"}
		errorJSON, _ := json.Marshal(errorResponse)
		return sender.Send(&backend.CallResourceResponse{
			Status:  http.StatusBadRequest,
			Headers: map[string][]string{"Content-Type": {"application/json"}},
			Body:    errorJSON,
		})
	}

	devices, err := d.api.GetDevices(group)
	if err != nil {
		errorResponse := map[string]string{"error": err.Error()}
		errorJSON, _ := json.Marshal(errorResponse)
		return sender.Send(&backend.CallResourceResponse{
			Status:  http.StatusInternalServerError,
			Headers: map[string][]string{"Content-Type": {"application/json"}},
			Body:    errorJSON,
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

/* ######################################### handleGetSensors ############################################################*/
func (d *Datasource) handleGetSensors(sender backend.CallResourceResponseSender, device string) error {
	if device == "" {
		errorResponse := map[string]string{"error": "missing device parameter"}
		errorJSON, _ := json.Marshal(errorResponse)
		return sender.Send(&backend.CallResourceResponse{
			Status:  http.StatusBadRequest,
			Headers: map[string][]string{"Content-Type": {"application/json"}},
			Body:    errorJSON,
		})
	}

	sensors, err := d.api.GetSensors(device)
	if err != nil {
		errorResponse := map[string]string{"error": err.Error()}
		errorJSON, _ := json.Marshal(errorResponse)
		return sender.Send(&backend.CallResourceResponse{
			Status:  http.StatusInternalServerError,
			Headers: map[string][]string{"Content-Type": {"application/json"}},
			Body:    errorJSON,
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

/*  ########################################  handleGetChannel ########################################  */
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
