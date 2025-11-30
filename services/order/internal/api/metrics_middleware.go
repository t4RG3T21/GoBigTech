package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/t4RG3T21/GoBigTech/services/order/internal/service"
)

var promMetrics *service.PrometheusMetrics

// SetPrometheusMetrics устанавливает Prometheus метрики для middleware
func SetPrometheusMetrics(metrics *service.PrometheusMetrics) {
	promMetrics = metrics
}

// MetricsMiddleware добавляет метрики для HTTP запросов
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создаем ResponseWriter для перехвата статуса
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		// Записываем метрику HTTP запроса в Prometheus
		if promMetrics != nil {
			promMetrics.RecordHTTPRequest(r.Method, r.URL.Path, strconv.Itoa(rw.statusCode))
		}

		_ = time.Since(start) // Для будущего расширения
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
