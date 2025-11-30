package service

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PrometheusMetrics содержит Prometheus метрики для Order Service
type PrometheusMetrics struct {
	// OrdersCreated - счетчик созданных заказов
	OrdersCreated prometheus.Counter

	// OrderRevenue - счетчик выручки
	OrderRevenue prometheus.Counter

	// OrderProcessingDuration - гистограмма времени обработки заказов
	OrderProcessingDuration prometheus.Histogram

	// OrdersFailed - счетчик неудачных заказов по причинам
	OrdersFailed *prometheus.CounterVec

	// HTTPRequestsTotal - счетчик HTTP запросов
	HTTPRequestsTotal *prometheus.CounterVec
}

// NewPrometheusMetrics создает новый экземпляр Prometheus метрик
// Метрики регистрируются в глобальном реестре prometheus через promauto
func NewPrometheusMetrics() *PrometheusMetrics {
	return &PrometheusMetrics{
		OrdersCreated: promauto.NewCounter(prometheus.CounterOpts{
			Name: "order_orders_created_total",
			Help: "The total number of created orders",
		}),

		OrderRevenue: promauto.NewCounter(prometheus.CounterOpts{
			Name: "order_revenue_total",
			Help: "Total revenue from all orders",
		}),

		OrderProcessingDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "order_processing_duration_seconds",
			Help:    "Time taken to process orders",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10},
		}),

		OrdersFailed: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "order_orders_failed_total",
			Help: "The total number of failed orders",
		}, []string{"reason"}),

		HTTPRequestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "order_http_requests_total",
			Help: "Total HTTP requests to order service",
		}, []string{"method", "endpoint", "status"}),
	}
}

// RecordOrderSuccess записывает метрики успешного заказа
func (m *PrometheusMetrics) RecordOrderSuccess(amount float64, durationSeconds float64) {
	if m == nil {
		return
	}
	m.OrdersCreated.Inc()
	m.OrderRevenue.Add(amount)
	m.OrderProcessingDuration.Observe(durationSeconds)
}

// RecordOrderFailure записывает метрики неудачного заказа
func (m *PrometheusMetrics) RecordOrderFailure(reason string) {
	if m == nil {
		return
	}
	m.OrdersFailed.WithLabelValues(reason).Inc()
}

// RecordHTTPRequest записывает метрики HTTP запроса
func (m *PrometheusMetrics) RecordHTTPRequest(method, endpoint, status string) {
	if m == nil {
		return
	}
	m.HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
}
