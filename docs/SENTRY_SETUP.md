# Настройка Sentry для мониторинга ошибок

## Обзор

Sentry интегрирован в проект для мониторинга ошибок, производительности и пользовательского опыта. Система автоматически отслеживает HTTP запросы, ошибки базы данных, и другие критические события.

## Конфигурация

### Переменные окружения

Добавьте следующие переменные в ваш `.env` файл:

```env
# Sentry Configuration
SENTRY_DSN=https://your-sentry-dsn@sentry.io/project-id
SENTRY_ENVIRONMENT=development
SENTRY_DEBUG=false
SENTRY_TRACES_SAMPLE_RATE=1.0
SENTRY_ENABLE_TRACING=true
```

### Параметры конфигурации

- `SENTRY_DSN` - DSN (Data Source Name) вашего Sentry проекта
- `SENTRY_ENVIRONMENT` - окружение (development, staging, production)
- `SENTRY_DEBUG` - включить отладочный режим Sentry
- `SENTRY_TRACES_SAMPLE_RATE` - частота сэмплирования трейсов (0.0 - 1.0)
- `SENTRY_ENABLE_TRACING` - включить трейсинг производительности

## Функциональность

### Автоматическое отслеживание

1. **HTTP запросы** - все HTTP запросы автоматически отслеживаются middleware
2. **Ошибки** - все ошибки отправляются в Sentry с контекстом
3. **Производительность** - трейсинг времени выполнения запросов
4. **Пользователи** - информация о пользователях добавляется автоматически

### Ручное отслеживание

#### Отправка ошибок

```go
import "github.com/Alias1177/Auth/pkg/sentry"

// Отправка ошибки с контекстом запроса
sentry.CaptureError(ctx, err, req)

// Отправка сообщения
sentry.CaptureMessageWithContext(ctx, "Custom message", req)

// Отправка предупреждения
sentry.CaptureWarning(ctx, "Warning message", req)
```

#### Добавление информации о пользователе

```go
// Добавить информацию о пользователе
sentry.AddUserInfo(ctx, userID, email)

// Добавить теги
sentry.AddTag(ctx, "feature", "auth")

// Добавить контекст
sentry.AddContext(ctx, "user_data", map[string]interface{}{
    "role": "admin",
    "permissions": []string{"read", "write"},
})
```

#### Трейсинг производительности

```go
import "github.com/Alias1177/Auth/pkg/sentry"

// Создать span для трейсинга
span := sentry.StartSpan(ctx, "database.query")
defer span.Finish()

// Добавить теги к span
span.SetTag("db.table", "users")
span.SetTag("db.operation", "select")
```

## Интеграция в обработчики

### Пример использования в обработчике

```go
func (h *Handler) SomeHandler(w http.ResponseWriter, r *http.Request) {
    // Обработка запроса
    user, err := h.service.GetUser(r.Context(), userID)
    if err != nil {
        // Отправляем ошибку в Sentry
        sentry.CaptureError(r.Context(), err, r)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // Добавляем информацию о пользователе
    sentry.AddUserInfo(r.Context(), user.ID, user.Email)

    // Успешная обработка
    // ...
}
```

## Мониторинг в Sentry

### Основные метрики

1. **Error Rate** - частота ошибок
2. **Response Time** - время ответа
3. **Throughput** - пропускная способность
4. **User Experience** - опыт пользователей

### Алерты

Настройте алерты в Sentry для:
- Высокой частоты ошибок
- Медленных запросов
- Критических ошибок
- Проблем с производительностью

### Дашборды

Создайте дашборды для отслеживания:
- Общей производительности приложения
- Ошибок по типам
- Географического распределения пользователей
- Популярных браузеров и устройств

## Лучшие практики

### 1. Контекст ошибок

Всегда добавляйте контекст к ошибкам:

```go
sentry.AddContext(ctx, "request_data", map[string]interface{}{
    "user_id": userID,
    "action": "login",
    "ip": r.RemoteAddr,
})
```

### 2. Фильтрация ошибок

Не отправляйте в Sentry:
- Ожидаемые ошибки (например, неправильный пароль)
- Ошибки валидации
- 404 ошибки

### 3. Производительность

- Используйте асинхронную отправку событий
- Не блокируйте основной поток
- Ограничивайте размер данных

### 4. Безопасность

- Не отправляйте чувствительные данные (пароли, токены)
- Фильтруйте персональные данные
- Используйте маскирование для конфиденциальной информации

## Отладка

### Включение отладки

```env
SENTRY_DEBUG=true
```

### Проверка подключения

```bash
# Проверьте логи приложения
docker logs auth-app | grep -i sentry
```

### Тестирование

```bash
# Отправьте тестовое событие
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"wrong"}'
```

## Troubleshooting

### Проблемы подключения

1. Проверьте правильность DSN
2. Убедитесь, что интернет-соединение доступно
3. Проверьте настройки файрвола

### Высокое потребление ресурсов

1. Уменьшите `SENTRY_TRACES_SAMPLE_RATE`
2. Отключите `SENTRY_ENABLE_TRACING` в development
3. Настройте фильтрацию событий

### Отсутствие событий

1. Проверьте логи Sentry
2. Убедитесь, что DSN корректный
3. Проверьте настройки проекта в Sentry 