#!/bin/bash

# Имя Docker-образа с инструментом миграции
MIGRATE_IMAGE="auth-app"

# Сеть Docker, в которой находятся PostgreSQL и Redis
DOCKER_NETWORK="bridge"

# Строка подключения к PostgreSQL
DATABASE_DSN="${DATABASE_DSN:-postgres://admin:secret@postgres:5432/mydb?sslmode=disable}"

# Адрес Redis
REDIS_ADDR="${REDIS_ADDR:-redis:6379}"

# Путь к директории с миграциями (относительно корня проекта)
MIGRATIONS_PATH="./db/migrations"

# ----------------------------------------------------------------------
# Проверки
# ----------------------------------------------------------------------

# Проверяем, установлен ли Docker
if ! command -v docker &> /dev/null
then
    echo "❌ Docker не установлен. Пожалуйста, установите Docker и попробуйте снова."
    exit 1
fi

# Проверяем, существует ли директория с миграциями
if ! [ -d "$MIGRATIONS_PATH" ]; then
    echo "❌ Директория с миграциями '$MIGRATIONS_PATH' не найдена."
    exit 1
fi

# ----------------------------------------------------------------------
# Функция для логирования
# ----------------------------------------------------------------------

log() {
    local level="$1"
    shift
    local message="$@"
    local timestamp=$(date "+%Y-%m-%d %H:%M:%S")

    if [ -n "$LOG_FILE" ]; then
        echo "$timestamp [$level] $message" >> "$LOG_FILE"
    fi
    echo "$timestamp [$level] $message"
}

# ----------------------------------------------------------------------
# Запуск контейнера для отката
# ----------------------------------------------------------------------

log "INFO" "🚀 Запуск отката миграций..."

docker run --rm -it \
    --network="$DOCKER_NETWORK" \
    -e DATABASE_DSN="$DATABASE_DSN" \
    -e REDIS_ADDR="$REDIS_ADDR" \
    -v "$(pwd)/$MIGRATIONS_PATH:/app/db/migrations" \
    "$MIGRATE_IMAGE" \
    /app/migrate -down -postgres -redis

# ----------------------------------------------------------------------
# Обработка результата
# ----------------------------------------------------------------------

# Получаем код возврата последней команды
RESULT=$?

if [ "$RESULT" -ne 0 ]; then
    log "ERROR" "❌ Ошибка при откате миграций. Код возврата: $RESULT"
    exit 1
else
    log "INFO" "✅ Откат миграций успешно завершен."
fi

log "INFO" "✅ Скрипт завершен."
exit 0