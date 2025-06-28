# Makefile для управления миграциями баз данных

# Конфигурационные переменные
MIGRATIONS_PATH = ./db/migrations

# Команда помощи
.PHONY: help
help:
	@echo "Использование:"
	@echo "  make migrate-up     - Запустить все ожидающие миграции (PostgreSQL)"
	@echo "  make migrate-down   - Откатить последнюю миграцию (PostgreSQL)"
	@echo "  make migrate-create name=название_миграции - Создать новую миграцию"
	@echo "  make docker-migrate-up   - Запустить миграции через docker-compose"
	@echo "  make docker-migrate-down - Откатить миграции через docker-compose"

# Запустить все ожидающие миграции (PostgreSQL)
.PHONY: migrate-up
migrate-up:
	@echo "Применение миграций PostgreSQL..."
	go run cmd/migration/main.go -up
	@echo "Миграции PostgreSQL успешно применены."

# Откатить последнюю миграцию (PostgreSQL)
.PHONY: migrate-down
migrate-down:
	@echo "Откат миграции PostgreSQL..."
	go run cmd/migration/main.go -down
	@echo "Откат PostgreSQL успешно завершен."

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

# Миграции через docker-compose
.PHONY: docker-migrate-up
docker-migrate-up:
	docker-compose run --rm service make migrate-up

.PHONY: docker-migrate-down
docker-migrate-down:
	docker-compose run --rm service make migrate-down