package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// HTTPMetrics содержит общие Prometheus метрики для HTTP запросов
type HTTPMetrics struct {
	// RequestsTotal - счетчик HTTP запросов
	RequestsTotal *prometheus.CounterVec

	// RequestDuration - гистограмма длительности HTTP запросов
	RequestDuration *prometheus.HistogramVec
}

// NewHTTPMetrics создает новый экземпляр HTTP метрик для указанного сервиса
// serviceName используется для префикса метрик (например, "order" -> "order_http_requests_total")
func NewHTTPMetrics(serviceName string) *HTTPMetrics {
	return &HTTPMetrics{
		RequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: serviceName + "_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    serviceName + "_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"method", "endpoint", "status"},
		),
	}
}

// RecordRequest записывает метрики HTTP запроса
func (m *HTTPMetrics) RecordRequest(method, endpoint, status string, durationSeconds float64) {
	if m == nil {
		return
	}
	m.RequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	m.RequestDuration.WithLabelValues(method, endpoint, status).Observe(durationSeconds)
}
