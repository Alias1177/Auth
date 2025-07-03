package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsMiddleware содержит счетчики для HTTP-кодов и время ответа
type MetricsMiddleware struct {
	httpCodes      *prometheus.CounterVec
	responseTime   *prometheus.HistogramVec
	activeRequests *prometheus.GaugeVec
	mutex          sync.RWMutex
	codeCount      map[string]int
}

// NewMetricsMiddleware создает новый экземпляр middleware для сбора метрик
func NewMetricsMiddleware(serviceName string) *MetricsMiddleware {
	httpCodes := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: serviceName + "_http_response_codes_total",
			Help: "Количество HTTP-ответов по кодам",
		},
		[]string{"code", "method", "path"},
	)

	responseTime := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    serviceName + "_http_response_time_seconds",
			Help:    "Время ответа HTTP-запросов",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"code", "method", "path"},
	)

	activeRequests := promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: serviceName + "_http_active_requests",
			Help: "Количество активных HTTP-запросов",
		},
		[]string{"method", "path"},
	)

	return &MetricsMiddleware{
		httpCodes:      httpCodes,
		responseTime:   responseTime,
		activeRequests: activeRequests,
		codeCount:      make(map[string]int),
		mutex:          sync.RWMutex{},
	}
}

// Middleware создает middleware для Chi роутера
func (m *MetricsMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Увеличиваем счетчик активных запросов
		path := chi.RouteContext(r.Context()).RoutePattern()
		if path == "" {
			path = "unknown"
		}
		m.activeRequests.WithLabelValues(r.Method, path).Inc()

		// Создаем обертку для записи кода ответа
		wrapper := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // По умолчанию 200 OK
		}

		// Вызываем следующий обработчик
		next.ServeHTTP(wrapper, r)

		// Уменьшаем счетчик активных запросов
		m.activeRequests.WithLabelValues(r.Method, path).Dec()

		// Получаем код ответа
		statusCode := wrapper.statusCode

		// Преобразуем код в строку и группируем по сотням (2xx, 4xx, 5xx)
		statusCodeStr := strconv.Itoa(statusCode)
		statusGroup := statusCodeStr[0:1] + "xx"

		// Увеличиваем счетчик для конкретного кода
		m.httpCodes.WithLabelValues(statusCodeStr, r.Method, path).Inc()

		// Увеличиваем счетчик для группы кодов (2xx, 4xx, 5xx)
		m.httpCodes.WithLabelValues(statusGroup, r.Method, path).Inc()

		// Записываем время ответа
		duration := time.Since(start).Seconds()
		m.responseTime.WithLabelValues(statusCodeStr, r.Method, path).Observe(duration)

		// Сохраняем статистику в локальной карте для быстрого доступа
		m.mutex.Lock()
		m.codeCount[statusCodeStr]++
		m.codeCount[statusGroup]++
		m.mutex.Unlock()
	})
}

// PrintStats выводит текущую статистику по кодам
func (m *MetricsMiddleware) PrintStats() map[string]int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Создаем копию карты для потокобезопасности
	stats := make(map[string]int)
	for code, count := range m.codeCount {
		stats[code] = count
	}

	return stats
}

// MetricsHandler возвращает HTTP-хендлер для Prometheus метрик
func (m *MetricsMiddleware) MetricsHandler() http.Handler {
	return promhttp.Handler()
}

// responseWriterWrapper для перехвата кода ответа
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader перехватывает код ответа
func (rww *responseWriterWrapper) WriteHeader(statusCode int) {
	rww.statusCode = statusCode
	rww.ResponseWriter.WriteHeader(statusCode)
}
