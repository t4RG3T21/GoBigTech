package httpmetrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/t4RG3T21/GoBigTech/platform/metrics"
)

// MetricsMiddleware добавляет метрики для HTTP запросов
// Использует переданные HTTP метрики для записи данных
func MetricsMiddleware(httpMetrics *metrics.HTTPMetrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Создаем ResponseWriter для перехвата статуса
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r)

			// Записываем метрику HTTP запроса
			duration := time.Since(start).Seconds()
			if httpMetrics != nil {
				httpMetrics.RecordRequest(
					r.Method,
					r.URL.Path,
					strconv.Itoa(rw.statusCode),
					duration,
				)
			}
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
