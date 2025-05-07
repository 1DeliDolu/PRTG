package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/1DeliDolu/PRTG/maxmarkusprogram/prtg/pkg/models"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/experimental/concurrent"
	"github.com/prometheus/client_golang/prometheus"
)

// Logger interface defines the logging methods required by the datasource

var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
	_ backend.CallResourceHandler   = (*Datasource)(nil)
	_ backend.StreamHandler         = (*Datasource)(nil) // Add streaming support
)

// Add queue and mutex at package level
var (
	requestQueue = make([]*ResourceRequest, 0)
	queueLock    sync.Mutex
)

// Add new type for request handling
type ResourceRequest struct {
	Request *backend.CallResourceRequest
	Sender  backend.CallResourceResponseSender
}

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

	ds := &Datasource{
		baseURL: baseURL,
		api:     NewApi(baseURL, config.Secrets.ApiKey, cacheTime, 10*time.Second),
		logger:  logger,
		tracer:  tracer,
		metrics: metrics,
	}

	// Initialize query type multiplexer
	queryTypeMux := datasource.NewQueryTypeMux()
	queryTypeMux.HandleFunc("metrics", ds.handleMetricsQueryType)
	queryTypeMux.HandleFunc("manual", ds.handleManualQueryType)
	queryTypeMux.HandleFunc("text", ds.handlePropertyQueryType)
	queryTypeMux.HandleFunc("raw", ds.handlePropertyQueryType)
	queryTypeMux.HandleFunc("", ds.handleQueryFallback)

	ds.mux = queryTypeMux
	return ds, nil
}

/*  ########################################### Dispose ################################################### */
func (d *Datasource) Dispose() {
}

/*  ########################################### QueryData ################################################### */
func (d *Datasource) handleSingleQueryData(ctx context.Context, q concurrent.Query) backend.DataResponse {
	// Start tracing span for single query
	ctx, span := d.tracer.StartSpan(ctx, "handleSingleQueryData")
	defer span.End()

	start := time.Now()
	d.logger.Debug("Processing concurrent query",
		"refId", q.DataQuery.RefID, // Update: Using the correct field name
	)

	// Execute the query using existing query logic
	res := d.query(ctx, q.PluginContext, q.DataQuery) // Update: Using the correct field names

	// Record metrics and logging
	duration := time.Since(start)
	d.metrics.ObserveQueryDuration("single_query", duration.Seconds())
	d.logger.Debug("Concurrent query processed",
		"refId", q.DataQuery.RefID, // Update: Using the correct field name
		"duration", duration,
		"status", res.Status,
		"hasError", res.Error != nil,
	)

	return res
}

/*  ########################################### MaxConcurrentQueries ################################################### */

const (
	MaxConcurrentQueries = 25
)

func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// Check if number of queries exceeds the limit
	if len(req.Queries) > MaxConcurrentQueries {
		return &backend.QueryDataResponse{
			Responses: map[string]backend.DataResponse{
				req.Queries[0].RefID: {
					Error: fmt.Errorf("number of concurrent queries (%d) exceeds maximum limit (%d)",
						len(req.Queries), MaxConcurrentQueries),
					Status: backend.StatusTooManyRequests,
				},
			},
		}, nil
	}

	// Create a wait group to manage concurrent queries
	var wg sync.WaitGroup
	responses := make(map[string]backend.DataResponse)
	responseLock := sync.Mutex{}

	// Process each query concurrently
	for _, q := range req.Queries {
		wg.Add(1)
		go func(query backend.DataQuery) {
			defer wg.Done()

			// Get response for the query
			response := d.handleSingleQueryData(ctx, concurrent.Query{
				DataQuery:     query,
				PluginContext: req.PluginContext,
			})

			// Safely store the response
			responseLock.Lock()
			responses[query.RefID] = response
			responseLock.Unlock()
		}(q)
	}

	// Wait for all queries to complete
	wg.Wait()

	return &backend.QueryDataResponse{
		Responses: responses,
	}, nil
}

// Add these new methods to handle different query types
func (d *Datasource) handleMetricsQueryType(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	ctx, span := d.tracer.StartSpan(ctx, "handleMetricsQueryType")
	defer span.End()

	response := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		response.Responses[q.RefID] = d.handleSingleQueryData(ctx, concurrent.Query{
			DataQuery:     q,
			PluginContext: req.PluginContext,
		})
	}

	return response, nil
}

func (d *Datasource) handleManualQueryType(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	_, span := d.tracer.StartSpan(ctx, "handleManualQueryType")
	defer span.End()

	response := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		// Parse the query model
		var qm queryModel
		if err := json.Unmarshal(q.JSON, &qm); err != nil {
			response.Responses[q.RefID] = backend.ErrDataResponse(backend.StatusBadRequest, "failed to parse query")
			continue
		}

		// Call the existing manual query handler
		response.Responses[q.RefID] = d.handleManualQuery(qm, q.TimeRange, fmt.Sprintf("manual_%s", q.RefID))
	}

	return response, nil
}

func (d *Datasource) handlePropertyQueryType(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	ctx, span := d.tracer.StartSpan(ctx, "handlePropertyQueryType")
	defer span.End()

	response := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		// Parse the query model
		var qm queryModel
		if err := json.Unmarshal(q.JSON, &qm); err != nil {
			response.Responses[q.RefID] = backend.ErrDataResponse(backend.StatusBadRequest, "failed to parse query")
			continue
		}

		// Call the existing property query handler
		response.Responses[q.RefID] = d.handlePropertyQuery(
			ctx,
			qm,
			qm.Property,
			qm.FilterProperty,
			fmt.Sprintf("property_%s", q.RefID),
		)
	}

	return response, nil
}

func (d *Datasource) handleQueryFallback(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	d.logger.Warn("Query type not supported", "queries", len(req.Queries))
	return backend.NewQueryDataResponse(), nil
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
		"02.01.2006 15:04:05", //Case when PRTG send time in this format or
		"2006-01-02 15:04:05", //case when PRTG send in this format
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

	// Queue the incoming request
	queueLock.Lock()
	requestQueue = append(requestQueue, &ResourceRequest{
		Request: req,
		Sender:  sender,
	})
	queueLock.Unlock()

	// Process queued requests
	return d.processQueuedRequests()
}

func (d *Datasource) processQueuedRequests() error {
	queueLock.Lock()
	defer queueLock.Unlock()

	if len(requestQueue) == 0 {
		return nil
	}

	// Define processing order
	orderedPaths := []string{"groups", "devices", "sensors", "channels"}
	var lastError error

	// Process requests in order
	for _, pathType := range orderedPaths {
		for i := 0; i < len(requestQueue); i++ {
			req := requestQueue[i]
			pathParts := strings.Split(req.Request.Path, "/")

			if pathParts[0] != pathType {
				continue
			}

			// Process request based on type
			var err error
			switch pathType {
			case "groups":
				err = d.handleGetGroups(req.Sender)
			case "devices":
				if len(pathParts) < 2 {
					err = sendErrorResponse(req.Sender, "group parameter is required", http.StatusBadRequest)
				} else {
					err = d.handleGetDevices(req.Sender, pathParts[1])
				}
			case "sensors":
				if len(pathParts) < 2 {
					err = sendErrorResponse(req.Sender, "device parameter is required", http.StatusBadRequest)
				} else {
					err = d.handleGetSensors(req.Sender, pathParts[1])
				}
			case "channels":
				if len(pathParts) < 2 {
					err = sendErrorResponse(req.Sender, "sensor parameter is required", http.StatusBadRequest)
				} else {
					err = d.handleGetChannel(req.Sender, pathParts[1])
				}
			}

			if err != nil {
				lastError = err
				d.logger.Error("Error processing request",
					"path", req.Request.Path,
					"error", err)
			}

			// Remove processed request from queue
			requestQueue = append(requestQueue[:i], requestQueue[i+1:]...)
			i-- // Adjust index after removal
		}
	}

	return lastError
}

func sendErrorResponse(sender backend.CallResourceResponseSender, message string, statusCode int) error {
	errorResponse := map[string]string{"error": message}
	errorJSON, _ := json.Marshal(errorResponse)
	return sender.Send(&backend.CallResourceResponse{
		Status:  statusCode,
		Headers: map[string][]string{"Content-Type": {"application/json"}},
		Body:    errorJSON,
	})
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
func (d *Datasource) handleGetChannel(sender backend.CallResourceResponseSender, sensorId string) error {
	if sensorId == "" {
		errorResponse := map[string]string{"error": "missing objid parameter"}
		errorJSON, _ := json.Marshal(errorResponse)
		return sender.Send(&backend.CallResourceResponse{
			Status:  http.StatusBadRequest,
			Headers: map[string][]string{"Content-Type": {"application/json"}},
			Body:    errorJSON,
		})
	}
	channels, err := d.api.GetChannels(sensorId)
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
