package tracing

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware добавляет трассировку для HTTP запросов
// serviceName используется для создания tracer (например, "order-service-http")
func TracingMiddleware(serviceName string) func(http.Handler) http.Handler {
	tracer := otel.Tracer(serviceName)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Извлекаем контекст трассировки из заголовков HTTP
			propagator := otel.GetTextMapPropagator()
			ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			// Создаем span для HTTP запроса
			attributes := []attribute.KeyValue{
				semconv.HTTPMethod(r.Method),
				semconv.HTTPRoute(r.URL.Path),
				semconv.HTTPURL(r.URL.String()),
				attribute.String("http.user_agent", r.UserAgent()),
				attribute.String("http.host", r.Host),
			}

			// Добавляем схему
			scheme := r.URL.Scheme
			if scheme == "" {
				scheme = "http"
			}
			attributes = append(attributes, semconv.HTTPScheme(scheme))

			ctx, span := tracer.Start(
				ctx,
				"http.request",
				trace.WithAttributes(attributes...),
				trace.WithSpanKind(trace.SpanKindServer),
			)
			defer span.End()

			// Создаем ResponseWriter для перехвата статуса
			rw := &tracedResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Выполняем следующий обработчик с контекстом трассировки
			next.ServeHTTP(rw, r.WithContext(ctx))

			// Добавляем атрибуты после обработки запроса
			span.SetAttributes(
				semconv.HTTPStatusCode(rw.statusCode),
				attribute.Bool("http.error", rw.statusCode >= 400),
			)

			// Устанавливаем статус в span
			if rw.statusCode >= 500 {
				span.RecordError(nil) // Можно добавить реальную ошибку, если есть
				span.SetStatus(codes.Error, "HTTP 5xx error")
			} else if rw.statusCode >= 400 {
				span.SetStatus(codes.Error, "HTTP 4xx error")
			} else {
				span.SetStatus(codes.Ok, "")
			}
		})
	}
}

type tracedResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *tracedResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
