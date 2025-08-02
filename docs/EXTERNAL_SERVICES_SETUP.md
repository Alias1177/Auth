# Настройка внешних подключений к сервисам

## Обзор

Auth Service настроен для работы с внешними сервисами на хосте `31.97.76.108`:
- **Kafka**: `31.97.76.108:9092`
- **Notification Service**: `http://31.97.76.108:8080`

## Конфигурация

### Переменные окружения

```bash
# Внешние сервисы
KAFKA_BROKER_ADDRESS=31.97.76.108:9092
KAFKA_EMAIL_TOPIC=notifications
NOTIFICATION_SERVICE_URL=http://31.97.76.108:8080
```

### Docker Compose

В `docker-compose.yaml` настроены переменные окружения для подключения к внешним сервисам:

```yaml
environment:
  - KAFKA_BROKER_ADDRESS=31.97.76.108:9092
  - KAFKA_EMAIL_TOPIC=notifications
  - NOTIFICATION_SERVICE_URL=http://31.97.76.108:8080
```

## Архитектура подключений

```
┌─────────────────┐    Kafka    ┌─────────────────────┐
│   Auth Service  │ ──────────► │ External Kafka      │
│   (Local)       │             │ (31.97.76.108:9092) │
│                 │             │                     │
│ - User Auth     │             │ - Message Broker    │
│ - Password Reset│             │ - Topic: notifications│
│ - Registration  │             │                     │
└─────────────────┘             └─────────────────────┘
         │                                │
         │ HTTP API                       │ HTTP API
         ▼                                ▼
┌─────────────────┐             ┌─────────────────┐
│   Frontend      │             │ Notification    │
│                 │             │ Service         │
│ - Login Form    │             │ (31.97.76.108:8080)│
│ - Reset Form    │             │ - Email Sending │
└─────────────────┘             └─────────────────┘
```

## Надежность подключений

### Kafka Producer

- **Retry логика**: 3 попытки при ошибках подключения
- **Timeout**: 10 секунд на чтение/запись
- **Batch processing**: Оптимизирован для внешних подключений

### Notification Client

- **Retry логика**: 3 попытки при HTTP ошибках
- **Timeout**: 30 секунд на HTTP запросы
- **Exponential backoff**: Увеличение задержки между попытками

## Тестирование подключений

### Автоматический тест

```bash
# Запуск скрипта тестирования
./scripts/test-connections.sh
```

### Ручное тестирование

```bash
# Тест Kafka подключения
nc -z -w5 31.97.76.108 9092

# Тест Notification Service
curl -s --connect-timeout 5 http://31.97.76.108:8080/health
```

## Troubleshooting

### Проблемы с Kafka

1. **Connection refused**
   ```bash
   # Проверьте доступность порта
   nc -z -w5 31.97.76.108 9092
   ```

2. **Authentication failed**
   - Проверьте настройки SASL/SSL в Kafka
   - Убедитесь, что топик `notifications` существует

3. **Network timeout**
   - Проверьте сетевое подключение
   - Увеличьте timeout в конфигурации

### Проблемы с Notification Service

1. **Connection timeout**
   ```bash
   # Проверьте доступность сервиса
   curl -v http://31.97.76.108:8080/health
   ```

2. **HTTP 404/500 errors**
   - Проверьте логи Notification Service
   - Убедитесь, что endpoint `/api/validate` существует

3. **DNS resolution issues**
   ```bash
   # Проверьте DNS резолвинг
   nslookup 31.97.76.108
   ```

## Мониторинг

### Логи

```bash
# Логи Auth Service
docker-compose logs -f auth-service

# Фильтрация по Kafka
docker-compose logs auth-service | grep -i kafka

# Фильтрация по Notification
docker-compose logs auth-service | grep -i notification
```

### Метрики

- **Kafka Producer**: Количество отправленных сообщений
- **Notification Client**: Время ответа и количество ошибок
- **Network**: Latency и packet loss

## Безопасность

### Рекомендации

1. **Network Security**
   - Используйте VPN для подключения к внешним сервисам
   - Настройте firewall rules
   - Используйте SSL/TLS для HTTP подключений

2. **Authentication**
   - Настройте SASL для Kafka
   - Используйте API keys для Notification Service
   - Регулярно ротируйте секреты

3. **Monitoring**
   - Настройте алерты на недоступность сервисов
   - Мониторьте latency и error rates
   - Логируйте все внешние подключения

## Производительность

### Оптимизации

1. **Connection Pooling**
   - HTTP клиент использует connection pooling
   - Kafka producer оптимизирован для batch processing

2. **Caching**
   - Redis кэш для пользовательских данных
   - Кэширование результатов валидации кодов

3. **Async Processing**
   - Kafka producer работает асинхронно
   - HTTP запросы с timeout для предотвращения блокировки

## Обновление конфигурации

### Изменение адресов сервисов

1. Обновите переменные окружения в `docker-compose.yaml`
2. Перезапустите контейнеры:
   ```bash
   docker-compose down
   docker-compose up -d
   ```

3. Проверьте подключения:
   ```bash
   ./scripts/test-connections.sh
   ```

### Добавление новых сервисов

1. Добавьте конфигурацию в `internal/config/config.go`
2. Создайте клиент в `pkg/`
3. Обновите container initialization
4. Добавьте тесты подключения 