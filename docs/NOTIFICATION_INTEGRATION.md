# Интеграция Auth Service с Notification Service

## Обзор

Auth Service теперь интегрирован с Notification Service для обработки email уведомлений. Вместо локальной генерации кодов и отправки email, Auth Service отправляет запросы в Kafka, которые обрабатывает Notification Service.

## Архитектура

```
┌─────────────────┐    Kafka    ┌─────────────────────┐
│   Auth Service  │ ──────────► │ Notification Service│
│   (Port 8081)   │             │   (Port 8080)       │
│                 │             │                     │
│ - User Auth     │             │ - Email Sending     │
│ - Password Reset│             │ - Code Generation   │
│ - Registration  │             │ - Code Validation   │
└─────────────────┘             └─────────────────────┘
         │                                │
         │ HTTP API                       │ SMTP
         ▼                                ▼
┌─────────────────┐             ┌─────────────────┐
│   Frontend      │             │   Email Server  │
│                 │             │                 │
│ - Login Form    │             │ - Gmail/SMTP    │
│ - Reset Form    │             │ - HTML Templates│
└─────────────────┘             └─────────────────┘
```

## Изменения в Auth Service

### 1. Обновленный Kafka Producer

- Добавлены структурированные сообщения для password reset и registration
- Поддержка JSON сериализации сообщений
- Улучшенное логирование

### 2. Новый Notification Client

- HTTP клиент для валидации кодов
- Таймауты и обработка ошибок
- Структурированные запросы/ответы

### 3. Обновленный Password Reset Service

- Убрана локальная генерация кодов
- Убрано локальное хранение в Redis
- Интеграция с Kafka для отправки запросов
- Валидация кодов через HTTP API

## Конфигурация

### Переменные окружения

```bash
# Kafka Configuration
KAFKA_BROKER_ADDRESS=kafka:9092
KAFKA_EMAIL_TOPIC=notifications

# Notification Service Configuration
NOTIFICATION_SERVICE_URL=http://notification-service:8080

# Email Configuration for Notification Service
MAIL=your-email@gmail.com
SECRET=your-app-password
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
```

### Docker Compose

Обновлен `docker-compose.yaml`:
- Добавлен Notification Service
- Изменен порт Auth Service на 8081
- Добавлены зависимости между сервисами
- Обновлена конфигурация Prometheus

## API Endpoints

### Password Reset Flow

1. **Запрос сброса пароля**
   ```bash
   POST http://localhost:8081/api/forgot-password
   Content-Type: application/json
   
   {
     "email": "user@example.com"
   }
   ```

2. **Подтверждение сброса пароля**
   ```bash
   POST http://localhost:8081/api/reset-password
   Content-Type: application/json
   
   {
     "email": "user@example.com",
     "code": "123456",
     "password": "newpassword123"
   }
   ```

### Валидация кода (внутренний API)

```bash
POST http://notification-service:8080/api/validate
Content-Type: application/json

{
  "email": "user@example.com",
  "code": "123456"
}
```

## Запуск

### 1. Подготовка окружения

```bash
# Скопируйте example.env в .env и настройте переменные
cp example.env .env

# Настройте email параметры в .env
MAIL=your-email@gmail.com
SECRET=your-app-password
```

### 2. Запуск сервисов

```bash
# Запуск всех сервисов
docker-compose up -d

# Проверка статуса
docker-compose ps
```

### 3. Проверка логов

```bash
# Логи Auth Service
docker-compose logs -f auth-app

# Логи Notification Service
docker-compose logs -f notification-service

# Логи Kafka
docker-compose logs -f kafka
```

## Тестирование

### 1. Тест восстановления пароля

```bash
# 1. Отправьте запрос на восстановление пароля
curl -X POST http://localhost:8081/api/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com"}'

# 2. Проверьте логи Notification Service
docker-compose logs notification-service

# 3. Проверьте, что email отправлен
# (проверьте почту user@example.com)

# 4. Валидируйте код
curl -X POST http://localhost:8080/api/validate \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "code": "123456"}'
```

### 2. Тест регистрации

```bash
# Отправьте запрос на регистрацию
curl -X POST http://localhost:8081/api/register \
  -H "Content-Type: application/json" \
  -d '{"email": "newuser@example.com", "username": "newuser", "password": "password123"}'

# Проверьте логи
docker-compose logs notification-service
```

## Мониторинг

### Prometheus

- Auth Service: `http://localhost:9091/targets` (job: auth_service)
- Notification Service: `http://localhost:9091/targets` (job: notification_service)

### Grafana

- URL: `http://localhost:3000`
- Login: `admin`
- Password: `ochenslozhniyparol`

## Troubleshooting

### 1. Kafka Connection Issues

```bash
# Проверьте статус Kafka
docker-compose logs kafka

# Проверьте подключение к Kafka
docker-compose exec auth-app nc -z kafka 9092
```

### 2. Notification Service Issues

```bash
# Проверьте логи Notification Service
docker-compose logs notification-service

# Проверьте доступность сервиса
curl http://localhost:8080/health
```

### 3. Email Issues

- Проверьте настройки SMTP в .env
- Убедитесь, что MAIL и SECRET правильно настроены
- Проверьте логи Notification Service на ошибки SMTP

## Безопасность

### Рекомендации

1. **Email Configuration**
   - Используйте App Passwords для Gmail
   - Не храните пароли в коде
   - Используйте переменные окружения

2. **Kafka Security**
   - В продакшене используйте SASL/SSL
   - Настройте ACL для топиков
   - Мониторьте доступ к Kafka

3. **API Security**
   - Используйте HTTPS в продакшене
   - Добавьте rate limiting
   - Логируйте все запросы

## Производительность

### Оптимизации

1. **Kafka Producer**
   - Batch processing для множественных запросов
   - Retry механизм для failed messages
   - Circuit breaker для Notification Service

2. **HTTP Client**
   - Connection pooling
   - Timeout configuration
   - Retry logic

3. **Monitoring**
   - Метрики для Kafka producer
   - Latency monitoring
   - Error rate tracking 