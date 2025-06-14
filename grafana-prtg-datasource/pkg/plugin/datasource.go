package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/1DeliDolu/grafana-plugins/grafana-prtg-datasource/pkg/models"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/experimental/concurrent"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
	_ backend.CallResourceHandler   = (*Datasource)(nil)
	_ backend.StreamHandler         = (*Datasource)(nil)
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

	// Get cache time from settings with default
	var cacheTime time.Duration = 60 * time.Second // default 60 seconds
	if config.CacheTime > 0 {
		cacheTime = config.CacheTime * time.Second
	}

	baseURL := fmt.Sprintf("https://%s", config.Path)
	logger := NewLogger()
	tracer := NewTracer(logger)
	metrics := NewMetrics(prometheus.DefaultRegisterer)

	// Use apitoken parameter name to match PRTG API requirements
	ds := &Datasource{
		baseURL:    baseURL,
		api:        NewApi(baseURL, config.Secrets.ApiKey, cacheTime, 10*time.Second),
		logger:     logger,
		tracer:     tracer,
		metrics:    metrics,
		queryCache: make(map[string]*QueryCacheEntry), // Updated initialization
		cacheMutex: sync.RWMutex{},
		cacheTime:  cacheTime,
		streamManager: &streamManager{
			streams:          make(map[string]*activeStream),
			activeStreams:    make(map[string]map[string]*activeStream), // Map of panel -> streams
			defaultCacheTime: cacheTime,                                 // Add default cache time to stream manager
		},
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
	// Clear caches on disposal
	d.cacheMutex.Lock()
	d.queryCache = make(map[string]*QueryCacheEntry)
	d.cacheMutex.Unlock()

	// Clear API cache if available
	if apiImpl, ok := d.api.(*Api); ok {
		apiImpl.ClearCache()
	}
}

// ClearAllCaches clears all cached data (query cache and API cache)
func (d *Datasource) ClearAllCaches() {
	d.cacheMutex.Lock()
	d.queryCache = make(map[string]*QueryCacheEntry)
	d.cacheMutex.Unlock()

	// Clear API cache if available
	if apiImpl, ok := d.api.(*Api); ok {
		apiImpl.ClearCache()
	}
	
	d.logger.Debug("All caches cleared")
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
	MaxConcurrentQueries = 10
)

// generateCacheKey creates a unique string key for caching query results
func generateCacheKey(req *backend.QueryDataRequest) string {
	var keyBuilder strings.Builder
	for _, q := range req.Queries {
		keyBuilder.WriteString(fmt.Sprintf("%s:%d:%d:%s;",
			q.RefID,
			q.TimeRange.From.UnixNano(),
			q.TimeRange.To.UnixNano(),
			string(q.JSON)))
	}
	return keyBuilder.String()
}

func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// Handle the case of empty queries
	if len(req.Queries) == 0 {
		return &backend.QueryDataResponse{
			Responses: make(map[string]backend.DataResponse),
		}, nil
	}

	// Check maximum concurrent query limit
	if len(req.Queries) > MaxConcurrentQueries {
		return &backend.QueryDataResponse{
			Responses: map[string]backend.DataResponse{
				req.Queries[0].RefID: {
					Error:  fmt.Errorf("query limit exceeded: %d/%d", len(req.Queries), MaxConcurrentQueries),
					Status: backend.StatusTooManyRequests,
				},
			},
		}, nil
	}

	// Generate a stable cache key
	cacheKey := generateCacheKey(req)
	d.cacheMutex.RLock()
	if cached, exists := d.queryCache[cacheKey]; exists && time.Now().Before(cached.ValidUntil) {
		d.cacheMutex.RUnlock()
		response := backend.NewQueryDataResponse()
		response.Responses[req.Queries[0].RefID] = cached.Response
		return response, nil
	}
	d.cacheMutex.RUnlock()

	// Handle query through multiplexer
	response, err := d.mux.QueryData(ctx, req)
	if err != nil {
		return nil, err
	}

	// Cache the result
	d.cacheMutex.Lock()
	d.queryCache[cacheKey] = &QueryCacheEntry{
		Response:   response.Responses[req.Queries[0].RefID],
		ValidUntil: time.Now().Add(d.cacheTime),
		Updating:   false,
	}
	d.cacheMutex.Unlock()

	return response, nil
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

/* ######################################## CheckHealth ##############################################################  */
func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	// Clear any cached results to ensure fresh health check with new configuration
	d.ClearAllCaches()

	_, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	d.logger.Debug("Starting health check")

	status, err := d.api.GetStatusList()
	if err != nil {
		d.logger.Error("PRTG health check failed", "error", err)
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("PRTG API error: %s", err.Error()),
		}, nil
	}

	detailsJSON, _ := json.Marshal(map[string]interface{}{
		"version":      status.Version,
		"totalSensors": status.TotalSens,
	})

	return &backend.CheckHealthResult{
		Status:      backend.HealthStatusOk,
		Message:     fmt.Sprintf("Data source is working. PRTG Version: %s", status.Version),
		JSONDetails: detailsJSON, // Fixed: Changed JsonDetails to JSONDetails
	}, nil
}

/* ######################################## StreamHandler ##############################################################  */

const MaxQueueSize = 100

func (d *Datasource) processQueuedRequests() error {
	queueLock.Lock()
	defer queueLock.Unlock()

	if len(requestQueue) == 0 {
		return nil
	}

	if len(requestQueue) > MaxQueueSize {
		d.logger.Warn("Request queue overflow, dropping old requests")
		requestQueue = requestQueue[len(requestQueue)-MaxQueueSize:]
	}

	var lastError error
	for _, req := range requestQueue {
		err := d.processRequest(req)
		if err != nil {
			d.logger.Error("Failed to process request", "error", err)
			lastError = err
		}
	}

	requestQueue = requestQueue[:0]
	return lastError
}

func (d *Datasource) processRequest(req *ResourceRequest) error {
	path := req.Request.Path
	d.logger.Debug("Processing request", "path", path)

	switch {
	case strings.HasPrefix(path, "groups"):
		return d.handleGetGroups(req.Sender)

	case strings.HasPrefix(path, "devices/"):
		pathParts := strings.Split(path, "/")
		if len(pathParts) < 2 {
			return sendErrorResponse(req.Sender, "group parameter is required", http.StatusBadRequest)
		}
		return d.handleGetDevices(req.Sender, pathParts[1])

	case strings.HasPrefix(path, "sensors/"):
		pathParts := strings.Split(path, "/")
		if len(pathParts) < 2 {
			return sendErrorResponse(req.Sender, "device parameter is required", http.StatusBadRequest)
		}
		return d.handleGetSensors(req.Sender, pathParts[1])

	case strings.HasPrefix(path, "channels/"):
		pathParts := strings.Split(path, "/")
		if len(pathParts) < 2 {
			return sendErrorResponse(req.Sender, "sensor parameter is required", http.StatusBadRequest)
		}
		return d.handleGetChannel(req.Sender, pathParts[1])

	default:
		return sendErrorResponse(req.Sender, "invalid API endpoint", http.StatusNotFound)
	}
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

// Add helper method to track active streams by panel
func (d *Datasource) trackStream(panelId string, streamId string, stream *activeStream) {
	d.streamManager.mu.Lock()
	defer d.streamManager.mu.Unlock()

	// Initialize map for this panel if needed
	if _, exists := d.streamManager.activeStreams[panelId]; !exists {
		d.streamManager.activeStreams[panelId] = make(map[string]*activeStream)
	}

	// Add stream to both maps for quick lookup
	d.streamManager.streams[streamId] = stream
	d.streamManager.activeStreams[panelId][streamId] = stream

	d.logger.Debug("Stream tracked",
		"panelId", panelId,
		"streamId", streamId,
		"totalStreams", len(d.streamManager.streams),
		"panelStreams", len(d.streamManager.activeStreams[panelId]))
}

// Add helper method to get all streams for a panel
func (d *Datasource) getStreamsByPanel(panelId string) []*activeStream {
	d.streamManager.mu.RLock()
	defer d.streamManager.mu.RUnlock()

	result := make([]*activeStream, 0, 5) // Preallocate with common size
	if panelStreams, exists := d.streamManager.activeStreams[panelId]; exists {
		for _, stream := range panelStreams {
			result = append(result, stream)
		}
	}
	return result
}
