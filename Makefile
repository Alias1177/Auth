# Makefile для управления миграциями баз данных

# Конфигурационные переменные
MIGRATIONS_PATH = ./db/migrations

# Команда помощи
.PHONY: help
help:
	@echo "Использование:"
	@echo "  make migrate-up     - Запустить все ожидающие миграции"
	@echo "  make migrate-down   - Откатить последнюю миграцию"
	@echo "  make migrate-pg-up  - Запустить только миграции PostgreSQL"
	@echo "  make migrate-pg-down - Откатить только миграции PostgreSQL"
	@echo "  make migrate-redis-up - Запустить только миграции Redis"
	@echo "  make migrate-redis-down - Откатить только миграции Redis"
	@echo "  make migrate-create name=название_миграции - Создать новую миграцию"

# Запустить все ожидающие миграции
.PHONY: migrate-up
migrate-up:
	@echo "Применение всех миграций..."
	go run cmd/migration/main.go -up
	@echo "Миграции успешно применены."

# Откатить последнюю миграцию
.PHONY: migrate-down
migrate-down:
	@echo "Откат всех последних миграций..."
	go run cmd/migration/main.go -down
	@echo "Откат успешно завершен."

# Запустить только миграции PostgreSQL
.PHONY: migrate-pg-up
migrate-pg-up:
	@echo "Применение миграций PostgreSQL..."
	go run cmd/migration/main.go -up -postgres
	@echo "Миграции PostgreSQL успешно применены."

# Откатить только миграции PostgreSQL
.PHONY: migrate-pg-down
migrate-pg-down:
	@echo "Откат миграций PostgreSQL..."
	go run cmd/migration/main.go -down -postgres
	@echo "Откат PostgreSQL успешно завершен."

# Запустить только миграции Redis
.PHONY: migrate-redis-up
migrate-redis-up:
	@echo "Применение миграций Redis..."
	go run cmd/migration/main.go -up -redis
	@echo "Миграции Redis успешно применены."

# Откатить только миграции Redis
.PHONY: migrate-redis-down
migrate-redis-down:
	@echo "Откат миграций Redis..."
	go run cmd/migration/main.go -down -redis
	@echo "Откат Redis успешно завершен."

# Создать новую миграцию PostgreSQL
.PHONY: migrate-create
migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Ошибка: Укажите название миграции с помощью 'name=название_миграции'"; \
		exit 1; \
	fi
	@echo "Создание новой миграции '$(name)'..."
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)
	@echo "Файлы миграции созданы."

# Сборка утилиты миграции
.PHONY: migrate-build
migrate-build:
	@echo "Сборка утилиты миграции..."
	go build -o bin/migrate cmd/migration/main.go
	@echo "Утилита собрана в bin/migrate"