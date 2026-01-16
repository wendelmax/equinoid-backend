package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics contém todas as métricas do Prometheus
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPActiveRequests  prometheus.Gauge

	// Database metrics
	DBConnectionsTotal  prometheus.Gauge
	DBConnectionsActive prometheus.Gauge
	DBConnectionsIdle   prometheus.Gauge
	DBQueryDuration     *prometheus.HistogramVec
	DBQueriesTotal      *prometheus.CounterVec

	// Cache metrics
	CacheOperationsTotal *prometheus.CounterVec
	CacheHitRatio        *prometheus.GaugeVec
	CacheDuration        *prometheus.HistogramVec

	// Business metrics
	EquinosTotal      prometheus.Gauge
	UsuariosTotal     prometheus.Gauge
	EventosTotal      *prometheus.CounterVec
	CoberturasTotal   *prometheus.CounterVec
	GestacaosTotal    *prometheus.CounterVec
	ValorizacaoTotal  *prometheus.CounterVec
	PostsSociaisTotal *prometheus.CounterVec
	OfertasTotal      *prometheus.CounterVec

	// Authentication metrics
	AuthAttemptsTotal    *prometheus.CounterVec
	ActiveSessions       prometheus.Gauge
	TokenGenerationTotal *prometheus.CounterVec

	// System metrics
	ApplicationStartTime prometheus.Gauge
	MemoryUsage          prometheus.Gauge
	GoroutinesActive     prometheus.Gauge
	CPUUsage             prometheus.Gauge

	// Error metrics
	ErrorsTotal         *prometheus.CounterVec
	PanicsTotal         prometheus.Counter
	SecurityEventsTotal *prometheus.CounterVec

	// Performance metrics
	ResponseTime       *prometheus.HistogramVec
	ThroughputRPS      *prometheus.GaugeVec
	LatencyPercentiles *prometheus.SummaryVec

	// External API metrics
	ExternalAPICallsTotal  *prometheus.CounterVec
	ExternalAPIDuration    *prometheus.HistogramVec
	ExternalAPIErrorsTotal *prometheus.CounterVec

	// Queue metrics (for async operations)
	QueueSize     *prometheus.GaugeVec
	QueueDuration *prometheus.HistogramVec
}

// NewMetrics cria uma nova instância de métricas
func NewMetrics() *Metrics {
	return &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),

		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),

		HTTPActiveRequests: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_active_requests",
				Help: "Number of active HTTP requests",
			},
		),

		// Database metrics
		DBConnectionsTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_connections_total",
				Help: "Total number of database connections",
			},
		),

		DBConnectionsActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_connections_active",
				Help: "Number of active database connections",
			},
		),

		DBConnectionsIdle: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_connections_idle",
				Help: "Number of idle database connections",
			},
		),

		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_query_duration_seconds",
				Help:    "Duration of database queries",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 2.0, 5.0},
			},
			[]string{"operation", "table"},
		),

		DBQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "db_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"operation", "table", "status"},
		),

		// Cache metrics
		CacheOperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_operations_total",
				Help: "Total number of cache operations",
			},
			[]string{"operation", "status"},
		),

		CacheHitRatio: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "cache_hit_ratio",
				Help: "Cache hit ratio",
			},
			[]string{"cache_type"},
		),

		CacheDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "cache_operation_duration_seconds",
				Help:    "Duration of cache operations",
				Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1},
			},
			[]string{"operation"},
		),

		// Business metrics
		EquinosTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "equinos_total",
				Help: "Total number of registered horses",
			},
		),

		UsuariosTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "usuarios_total",
				Help: "Total number of registered users",
			},
		),

		EventosTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "eventos_total",
				Help: "Total number of events",
			},
			[]string{"tipo_evento", "status"},
		),

		CoberturasTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "coberturas_total",
				Help: "Total number of breeding covers",
			},
			[]string{"tipo_cobertura", "status"},
		),

		GestacaosTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gestacoes_total",
				Help: "Total number of pregnancies",
			},
			[]string{"status"},
		),

		ValorizacaoTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "valorizacao_total",
				Help: "Total number of valorization records",
			},
			[]string{"categoria", "status"},
		),

		PostsSociaisTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "posts_sociais_total",
				Help: "Total number of social posts",
			},
			[]string{"tipo_conteudo", "status"},
		),

		OfertasTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ofertas_total",
				Help: "Total number of offers",
			},
			[]string{"tipo_oferta", "status"},
		),

		// Authentication metrics
		AuthAttemptsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_attempts_total",
				Help: "Total number of authentication attempts",
			},
			[]string{"method", "status"},
		),

		ActiveSessions: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "active_sessions",
				Help: "Number of active user sessions",
			},
		),

		TokenGenerationTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "token_generation_total",
				Help: "Total number of tokens generated",
			},
			[]string{"token_type", "status"},
		),

		// System metrics
		ApplicationStartTime: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "application_start_time_seconds",
				Help: "Unix timestamp when the application started",
			},
		),

		MemoryUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "memory_usage_bytes",
				Help: "Current memory usage in bytes",
			},
		),

		GoroutinesActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "goroutines_active",
				Help: "Number of active goroutines",
			},
		),

		CPUUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "cpu_usage_percent",
				Help: "Current CPU usage percentage",
			},
		),

		// Error metrics
		ErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "errors_total",
				Help: "Total number of errors",
			},
			[]string{"component", "error_type"},
		),

		PanicsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "panics_total",
				Help: "Total number of panics",
			},
		),

		SecurityEventsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "security_events_total",
				Help: "Total number of security events",
			},
			[]string{"event_type", "severity"},
		),

		// Performance metrics
		ResponseTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "response_time_seconds",
				Help:    "Response time of requests",
				Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
			},
			[]string{"endpoint", "method"},
		),

		ThroughputRPS: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "throughput_requests_per_second",
				Help: "Throughput in requests per second",
			},
			[]string{"endpoint"},
		),

		LatencyPercentiles: promauto.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:       "latency_percentiles_seconds",
				Help:       "Latency percentiles",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.95: 0.005, 0.99: 0.001},
			},
			[]string{"endpoint", "method"},
		),

		// External API metrics
		ExternalAPICallsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "external_api_calls_total",
				Help: "Total number of external API calls",
			},
			[]string{"service", "endpoint", "status"},
		),

		ExternalAPIDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "external_api_duration_seconds",
				Help:    "Duration of external API calls",
				Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0, 60.0},
			},
			[]string{"service", "endpoint"},
		),

		ExternalAPIErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "external_api_errors_total",
				Help: "Total number of external API errors",
			},
			[]string{"service", "endpoint", "error_type"},
		),

		// Queue metrics
		QueueSize: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "queue_size",
				Help: "Current queue size",
			},
			[]string{"queue_name", "status"},
		),

		QueueDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "queue_processing_duration_seconds",
				Help:    "Duration of queue processing",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"queue_name", "operation"},
		),
	}
}

// RecordHTTPRequest registra uma requisição HTTP
func (m *Metrics) RecordHTTPRequest(method, endpoint string, statusCode int, duration time.Duration) {
	status := strconv.Itoa(statusCode)

	m.HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
	m.ResponseTime.WithLabelValues(endpoint, method).Observe(duration.Seconds())
	m.LatencyPercentiles.WithLabelValues(endpoint, method).Observe(duration.Seconds())
}

// RecordDBQuery registra uma consulta ao banco de dados
func (m *Metrics) RecordDBQuery(operation, table, status string, duration time.Duration) {
	m.DBQueriesTotal.WithLabelValues(operation, table, status).Inc()
	m.DBQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordCacheOperation registra uma operação de cache
func (m *Metrics) RecordCacheOperation(operation, status string, duration time.Duration) {
	m.CacheOperationsTotal.WithLabelValues(operation, status).Inc()
	m.CacheDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordAuthAttempt registra uma tentativa de autenticação
func (m *Metrics) RecordAuthAttempt(method, status string) {
	m.AuthAttemptsTotal.WithLabelValues(method, status).Inc()
}

// RecordError registra um erro
func (m *Metrics) RecordError(component, errorType string) {
	m.ErrorsTotal.WithLabelValues(component, errorType).Inc()
}

// RecordSecurityEvent registra um evento de segurança
func (m *Metrics) RecordSecurityEvent(eventType, severity string) {
	m.SecurityEventsTotal.WithLabelValues(eventType, severity).Inc()
}

// RecordBusinessEvent registra um evento de negócio
func (m *Metrics) RecordBusinessEvent(eventType string, labels ...string) {
	switch eventType {
	case "evento":
		if len(labels) >= 2 {
			m.EventosTotal.WithLabelValues(labels[0], labels[1]).Inc()
		}
	case "cobertura":
		if len(labels) >= 2 {
			m.CoberturasTotal.WithLabelValues(labels[0], labels[1]).Inc()
		}
	case "gestacao":
		if len(labels) >= 1 {
			m.GestacaosTotal.WithLabelValues(labels[0]).Inc()
		}
	case "valorizacao":
		if len(labels) >= 2 {
			m.ValorizacaoTotal.WithLabelValues(labels[0], labels[1]).Inc()
		}
	case "post_social":
		if len(labels) >= 2 {
			m.PostsSociaisTotal.WithLabelValues(labels[0], labels[1]).Inc()
		}
	case "oferta":
		if len(labels) >= 2 {
			m.OfertasTotal.WithLabelValues(labels[0], labels[1]).Inc()
		}
	}
}

// RecordExternalAPICall registra uma chamada à API externa
func (m *Metrics) RecordExternalAPICall(service, endpoint, status string, duration time.Duration) {
	m.ExternalAPICallsTotal.WithLabelValues(service, endpoint, status).Inc()
	m.ExternalAPIDuration.WithLabelValues(service, endpoint).Observe(duration.Seconds())
}

// UpdateSystemMetrics atualiza métricas do sistema
func (m *Metrics) UpdateSystemMetrics(memoryUsage float64, goroutines int, cpuUsage float64) {
	m.MemoryUsage.Set(memoryUsage)
	m.GoroutinesActive.Set(float64(goroutines))
	m.CPUUsage.Set(cpuUsage)
}

// UpdateBusinessCounts atualiza contadores de negócio
func (m *Metrics) UpdateBusinessCounts(equinos, usuarios int) {
	m.EquinosTotal.Set(float64(equinos))
	m.UsuariosTotal.Set(float64(usuarios))
}

// SetApplicationStartTime define o timestamp de início da aplicação
func (m *Metrics) SetApplicationStartTime() {
	m.ApplicationStartTime.Set(float64(time.Now().Unix()))
}

// IncrementActiveRequests incrementa requisições ativas
func (m *Metrics) IncrementActiveRequests() {
	m.HTTPActiveRequests.Inc()
}

// DecrementActiveRequests decrementa requisições ativas
func (m *Metrics) DecrementActiveRequests() {
	m.HTTPActiveRequests.Dec()
}

// UpdateActiveSessions atualiza o número de sessões ativas
func (m *Metrics) UpdateActiveSessions(count int) {
	m.ActiveSessions.Set(float64(count))
}

// RecordPanic registra um panic
func (m *Metrics) RecordPanic() {
	m.PanicsTotal.Inc()
}
