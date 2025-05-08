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

	// Metrics have been moved to the Metrics struct below
	// Keeping only activeConnections as it's used in UpdateActiveConnections
)



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
