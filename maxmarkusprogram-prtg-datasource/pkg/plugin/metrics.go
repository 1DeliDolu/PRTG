package plugin

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

/* =================================== METRICS DEFINITIONS ===================================== */

var (
	// Active connections metric - now used with logging
	activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "grafana_plugin",
			Name:      "prtg_active_connections",
			Help:      "Current number of active connections.",
		},
	)

	// API Request metrics
	apiRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "grafana_plugin",
			Name:      "prtg_api_requests_total",
			Help:      "Total number of PRTG API requests by method and status.",
		},
		[]string{"method", "status"},
	)

	// API Response time metrics
	apiRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "grafana_plugin",
			Name:      "prtg_api_request_duration_seconds",
			Help:      "Duration of PRTG API requests in seconds.",
			Buckets:   prometheus.ExponentialBuckets(0.1, 2, 10), // 0.1s to ~51.2s
		},
		[]string{"method"},
	)

	// Cache metrics
	cacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "grafana_plugin",
			Name:      "prtg_cache_hits_total",
			Help:      "Total number of cache hits by request type.",
		},
		[]string{"type"},
	)

	cacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "grafana_plugin",
			Name:      "prtg_cache_misses_total",
			Help:      "Total number of cache misses by request type.",
		},
		[]string{"type"},
	)

	// Error metrics
	errorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "grafana_plugin",
			Name:      "prtg_errors_total",
			Help:      "Total number of errors by type.",
		},
		[]string{"error_type"},
	)
)

/* =================================== METRIC HELPERS ======================================== */

// validateLabel ensures label values are safe and within bounds
func validateLabel(label string, maxLength int) string {
	if len(label) > maxLength {
		return label[:maxLength]
	}
	return label
}

// incrementAPIRequests increments the API request counter
func incrementAPIRequests(method, status string) {
	method = validateLabel(method, 50)
	status = validateLabel(status, 20)
	apiRequestsTotal.WithLabelValues(method, status).Inc()
}

// observeAPIRequestDuration records the duration of an API request
func observeAPIRequestDuration(method string, duration float64) {
	method = validateLabel(method, 50)
	apiRequestDuration.WithLabelValues(method).Observe(duration)
}

// incrementCacheMetric increments cache hit/miss counters
func incrementCacheMetric(hit bool, requestType string) {
	requestType = validateLabel(requestType, 30)
	if hit {
		cacheHitsTotal.WithLabelValues(requestType).Inc()
	} else {
		cacheMissesTotal.WithLabelValues(requestType).Inc()
	}
}

// incrementErrors increments the error counter
func incrementErrors(errorType string) {
	errorType = validateLabel(errorType, 50)
	errorsTotal.WithLabelValues(errorType).Inc()
}

/* =================================== METRICS STRUCT ======================================== */
type Metrics struct {
	apiRequests   *prometheus.CounterVec
	apiLatency    *prometheus.HistogramVec
	queryDuration *prometheus.HistogramVec
	cacheHits     *prometheus.CounterVec
	errorCounter  *prometheus.CounterVec
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		apiRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "prtg_api_requests_total",
				Help: "Total number of API requests made to PRTG",
			},
			[]string{"endpoint"},
		),
		apiLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "prtg_api_request_duration_seconds",
				Help: "Duration of API requests to PRTG",
			},
			[]string{"endpoint"},
		),
		queryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "prtg_query_duration_seconds",
				Help: "Duration of PRTG queries",
			},
			[]string{"query_type"},
		),
		cacheHits: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "prtg_cache_hits_total",
				Help: "Total number of cache hits",
			},
			[]string{"type"},
		),
		errorCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "prtg_errors_total",
				Help: "Total number of errors",
			},
			[]string{"type"},
		),
	}

	reg.MustRegister(m.apiRequests)
	reg.MustRegister(m.apiLatency)
	reg.MustRegister(m.queryDuration)
	reg.MustRegister(m.cacheHits)
	reg.MustRegister(m.errorCounter)

	return m
}

func (m *Metrics) IncAPIRequest(endpoint string) {
	m.apiRequests.WithLabelValues(endpoint).Inc()
}

func (m *Metrics) ObserveAPILatency(endpoint string, duration float64) {
	m.apiLatency.WithLabelValues(endpoint).Observe(duration)
}

func (m *Metrics) ObserveQueryDuration(queryType string, duration float64) {
	m.queryDuration.WithLabelValues(queryType).Observe(duration)
}

func (m *Metrics) IncCacheHit(type_ string) {
	m.cacheHits.WithLabelValues(type_).Inc()
}

func (m *Metrics) IncError(type_ string) {
	m.errorCounter.WithLabelValues(type_).Inc()
}

// Add this method to the Metrics struct
func (m *Metrics) UpdateActiveConnections(count float64, logger PrtgLogger) {
	activeConnections.Set(count)
	logger.Debug("Updated active connections metric",
		"count", count,
		"metric", "prtg_active_connections",
	)
}
