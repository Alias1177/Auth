# Используем multi-stage build

# Этап 1: Сборка бинарника Go
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Копируем файлы модулей и скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код проекта
COPY . .

# Билдим Go-приложение
RUN cd cmd/services && CGO_ENABLED=0 go build -ldflags "-s -w" -o auth-app

# Этап 2: Запуск приложения в минимальном контейнере
FROM alpine:latest

WORKDIR /app

# Копируем только готовый бинарник из builder-стадии
COPY --from=builder /app/cmd/services/auth-app ./auth-app

# Копируем .env файл (по необходимости)
COPY .env .env

# Указываем открываемый порт приложения (порт, на котором слушает Go-сервер)
EXPOSE 8080

# Запускаем готовый бинарник
ENTRYPOINT ["./auth-app"]