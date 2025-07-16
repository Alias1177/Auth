# Тестирование Google OAuth через Insomnia

## Предварительная настройка

1. Убедитесь, что в `.env` файле настроены Google OAuth credentials:
```env
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
```

2. В Google Cloud Console настройте разрешенные redirect URIs:
   - `http://localhost:8080/auth/google/callback` (для локальной разработки)
   - `https://your-domain.com/auth/google/callback` (для продакшена)

## Тестирование через Insomnia

### 1. Инициация OAuth потока

**Запрос:** `GET /auth/google`

**URL:** `http://localhost:8080/auth/google`

**Описание:** Этот запрос инициирует OAuth поток и перенаправляет пользователя на страницу авторизации Google.

**Ожидаемый результат:** 
- HTTP 302 Redirect на `https://accounts.google.com/oauth/authorize?...`
- В браузере откроется страница авторизации Google

**Примечание:** Этот запрос лучше тестировать в браузере, так как Insomnia не может обработать redirect.

### 2. Тестирование в браузере (рекомендуемый способ)

1. Откройте браузер и перейдите по адресу: `http://localhost:8080/auth/google`
2. Войдите в свой Google аккаунт
3. После успешной авторизации вы будете перенаправлены на callback URL
4. В ответе вы получите JSON с токенами:

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 123,
    "email": "user@gmail.com",
    "username": "username"
  }
}
```

### 3. Тестирование с полученными токенами

После получения токенов, вы можете использовать их для тестирования защищенных эндпоинтов:

#### Получение информации о пользователе

**Запрос:** `GET /user/me`

**Headers:**
```
Authorization: Bearer YOUR_ACCESS_TOKEN
Content-Type: application/json
```

**Ожидаемый результат:**
```json
{
  "id": 123,
  "email": "user@gmail.com",
  "username": "username",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### Обновление токена

**Запрос:** `POST /refresh-token`

**Headers:**
```
Content-Type: application/json
```

**Body:**
```json
{
  "refresh_token": "YOUR_REFRESH_TOKEN"
}
```

**Ожидаемый результат:**
```json
{
  "access_token": "NEW_ACCESS_TOKEN",
  "refresh_token": "NEW_REFRESH_TOKEN"
}
```

### 4. Тестирование logout

**Запрос:** `GET /logout/google`

**URL:** `http://localhost:8080/logout/google`

**Описание:** Выход из OAuth сессии

**Ожидаемый результат:** HTTP 307 Temporary Redirect на `/`

## Альтернативный способ тестирования через Insomnia

### Создание коллекции в Insomnia

1. Создайте новую коллекцию "OAuth Testing"
2. Добавьте переменные окружения:
   - `base_url`: `http://localhost:8080`
   - `access_token`: (будет заполнено после OAuth)
   - `refresh_token`: (будет заполнено после OAuth)

### Запросы для коллекции

#### 1. OAuth Initiation
- **Method:** GET
- **URL:** `{{base_url}}/auth/google`
- **Note:** Этот запрос покажет redirect URL, который нужно открыть в браузере

#### 2. Get User Info (после получения токена)
- **Method:** GET
- **URL:** `{{base_url}}/user/me`
- **Headers:**
  ```
  Authorization: Bearer {{access_token}}
  ```

#### 3. Refresh Token
- **Method:** POST
- **URL:** `{{base_url}}/refresh-token`
- **Headers:**
  ```
  Content-Type: application/json
  ```
- **Body:**
  ```json
  {
    "refresh_token": "{{refresh_token}}"
  }
  ```

## Отладка проблем

### 1. Проверка логов
Следите за логами сервера для диагностики проблем:
```bash
# В терминале где запущен сервер
tail -f logs/app.log
```

### 2. Проверка конфигурации
Убедитесь, что OAuth правильно инициализирован:
```bash
# Проверьте что сервер запускается без ошибок
go run cmd/app/main.go
```

### 3. Проверка Google OAuth настроек
- Client ID и Client Secret корректны
- Redirect URI добавлен в Google Cloud Console
- Приложение не в режиме тестирования (или email добавлен в тестовые пользователи)

## Примеры ошибок и решений

### "OAuth authentication failed"
- Проверьте Google OAuth credentials
- Убедитесь, что redirect URI настроен правильно

### "Failed to create user"
- Проверьте подключение к базе данных
- Убедитесь, что таблица users существует

### "Failed to generate token"
- Проверьте JWT_SECRET в .env файле
- Убедитесь, что TokenManager правильно инициализирован 