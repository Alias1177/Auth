# Используем multi-stage build

# Этап 1: Сборка бинарника Go
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Устанавливаем git для go mod download
RUN apk add --no-cache git

# Копируем файлы модулей и скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем только необходимые исходные файлы для сборки
COPY cmd/service ./cmd/service
COPY internal ./internal
COPY pkg ./pkg
COPY db ./db

# Устанавливаем переменные окружения для ускорения сборки
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# Билдим Go-приложение с оптимизациями
RUN cd cmd/service && \
    go build \
    -ldflags "-s -w" \
    -o auth-app

# Этап 2: Запуск приложения в минимальном контейнере
FROM alpine:latest

# Устанавливаем ca-certificates для HTTPS запросов
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Копируем только готовый бинарник из builder-стадии
COPY --from=builder /app/cmd/service/auth-app ./auth-app

# Копируем миграции
COPY db/migrations ./db/migrations

# Создаем пользователя для безопасности
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Меняем владельца файлов
RUN chown -R appuser:appgroup /app

# Переключаемся на непривилегированного пользователя
USER appuser

# Указываем открываемый порт приложения
EXPOSE 8080

# Запускаем готовый бинарник
ENTRYPOINT ["./auth-app"]