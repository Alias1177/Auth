package middleware

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Глобальный экземпляр middleware для доступа из разных частей приложения
var (
	globalMetricsMiddleware *MetricsMiddleware
	once                    sync.Once
)

// MetricsMiddleware содержит счетчики для HTTP-кодов и время ответа
type MetricsMiddleware struct {
	// Существующие метрики
	httpCodes      *prometheus.CounterVec
	responseTime   *prometheus.HistogramVec
	activeRequests *prometheus.GaugeVec

	// Новые метрики
	responseTimeSummary *prometheus.SummaryVec // Для среднего времени ответа
	responseTimeP95     *prometheus.SummaryVec // Для p95 времени ответа

	mutex     sync.RWMutex
	codeCount map[string]int

	// Карта для хранения путей запросов
	pathCache    map[*http.Request]string
	pathCacheMux sync.RWMutex
}

// NewMetricsMiddleware создает новый экземпляр middleware для сбора метрик
func NewMetricsMiddleware(serviceName string) *MetricsMiddleware {
	// Инициализация глобального синглтона
	once.Do(func() {
		// Метрика для подсчёта HTTP кодов ответа по эндпоинтам
		httpCodes := promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: serviceName + "_http_response_codes_total_v2",
				Help: "Количество HTTP-ответов по кодам",
			},
			[]string{"code", "method", "path"},
		)

		// Гистограмма для времени ответа
		responseTime := promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    serviceName + "_http_response_time_seconds_v2",
				Help:    "Время ответа HTTP-запросов",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"code", "method", "path"},
		)

		// Остальные метрики...
		responseTimeSummary := promauto.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:       serviceName + "_http_response_time_avg_seconds_v2",
				Help:       "Среднее время ответа HTTP-запросов",
				MaxAge:     10 * time.Minute,
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			},
			[]string{"method", "path"},
		)

		responseTimeP95 := promauto.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:       serviceName + "_http_response_time_p95_seconds_v2",
				Help:       "95-й процентиль времени ответа HTTP-запросов",
				MaxAge:     10 * time.Minute,
				Objectives: map[float64]float64{0.95: 0.01},
			},
			[]string{"method", "path"},
		)

		activeRequests := promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: serviceName + "_http_active_requests_v2",
				Help: "Количество активных HTTP-запросов",
			},
			[]string{"method", "path"},
		)

		globalMetricsMiddleware = &MetricsMiddleware{
			httpCodes:           httpCodes,
			responseTime:        responseTime,
			responseTimeSummary: responseTimeSummary,
			responseTimeP95:     responseTimeP95,
			activeRequests:      activeRequests,
			codeCount:           make(map[string]int),
			mutex:               sync.RWMutex{},
			pathCache:           make(map[*http.Request]string),
			pathCacheMux:        sync.RWMutex{},
		}
	})

	return globalMetricsMiddleware
}

// GetMetricsMiddleware возвращает глобальный экземпляр middleware
func GetMetricsMiddleware() *MetricsMiddleware {
	return globalMetricsMiddleware
}

// RecordPathForRequest устанавливает путь для конкретного запроса
func (m *MetricsMiddleware) RecordPathForRequest(r *http.Request, path string) {
	m.pathCacheMux.Lock()
	defer m.pathCacheMux.Unlock()

	// Сохраняем путь для этого запроса
	m.pathCache[r] = path

	// Выводим отладочную информацию
	fmt.Printf("УСТАНОВЛЕН ПУТЬ ДЛЯ ЗАПРОСА: %s -> %s\n", r.URL.Path, path)
}

// getPathForRequest получает установленный путь для запроса
func (m *MetricsMiddleware) getPathForRequest(r *http.Request) string {
	m.pathCacheMux.RLock()
	defer m.pathCacheMux.RUnlock()

	if path, ok := m.pathCache[r]; ok {
		return path
	}
	return ""
}

// Middleware создает middleware для Chi роутера
func (m *MetricsMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Пытаемся получить предустановленный путь из кэша
		path := m.getPathForRequest(r)
		fmt.Printf("Пытаемся получить путь из кэша: %s\n", path)

		// Если путь не установлен в кэше, используем другие методы
		if path == "" {
			// Явно установленный путь в контексте
			pathFromCtx, ok := r.Context().Value(PathKey).(string)
			if ok && pathFromCtx != "" {
				path = pathFromCtx
				fmt.Printf("Получили путь из контекста: %s\n", path)
			} else {
				// Пробуем получить из Chi
				routeCtx := chi.RouteContext(r.Context())
				if routeCtx != nil && routeCtx.RoutePattern() != "" {
					path = routeCtx.RoutePattern()
					fmt.Printf("Получили путь из Chi: %s\n", path)
				} else {
					// Используем фактический путь запроса
					path = r.URL.Path
					fmt.Printf("Используем путь URL: %s\n", path)
				}
			}
		}

		// ВАЖНО: используем жестко заданное значение для отладки
		path = "EXPLICIT_PATH_DEBUG"
		fmt.Printf("ФИНАЛЬНЫЙ ПУТЬ: %s\n", path)

		// Увеличиваем счетчик активных запросов
		m.activeRequests.WithLabelValues(r.Method, path).Inc()

		// Создаем обертку для записи кода ответа
		wrapper := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // По умолчанию 200 OK
		}

		// Вызываем следующий обработчик
		next.ServeHTTP(wrapper, r)

		// После обработки запроса убираем из кэша
		if m.pathCache != nil {
			m.pathCacheMux.Lock()
			delete(m.pathCache, r)
			m.pathCacheMux.Unlock()
		}

		// Уменьшаем счетчик активных запросов
		m.activeRequests.WithLabelValues(r.Method, path).Dec()

		// Получаем код ответа
		statusCode := wrapper.statusCode
		statusCodeStr := strconv.Itoa(statusCode)

		// Группируем коды по сотням (2xx, 4xx, 5xx)
		statusGroup := statusCodeStr[0:1] + "xx"

		// Увеличиваем счетчик для конкретного кода
		m.httpCodes.WithLabelValues(statusCodeStr, r.Method, path).Inc()

		// Увеличиваем счетчик для группы кодов (2xx, 4xx, 5xx)
		m.httpCodes.WithLabelValues(statusGroup, r.Method, path).Inc()

		// Измеряем время ответа
		duration := time.Since(start).Seconds()

		// Записываем в гистограмму
		m.responseTime.WithLabelValues(statusCodeStr, r.Method, path).Observe(duration)

		// Записываем в summary для среднего времени
		m.responseTimeSummary.WithLabelValues(r.Method, path).Observe(duration)

		// Записываем в summary для p95
		m.responseTimeP95.WithLabelValues(r.Method, path).Observe(duration)

		// Сохраняем статистику в локальной карте для быстрого доступа
		m.mutex.Lock()
		m.codeCount[fmt.Sprintf("%s-%s-%s", statusCodeStr, r.Method, path)]++
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
