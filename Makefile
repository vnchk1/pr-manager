.PHONY: build run test migrate-up migrate-down clean docker-up docker-down lint health-check fmt

# Переменные
BINARY_NAME=pr-manager
MIGRATIONS_DIR=./migrations

# Сборка приложения
build:
	go build -o $(BINARY_NAME) ./cmd/api

# Запуск приложения
run:
	go run ./cmd/api

# Запуск тестов
test:
	go test ./... -v

# Запуск линтеров
lint:
	golangci-lint run

# Подъем миграций
migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "user=postgres password=password dbname=pr_manager sslmode=disable" up

# Сброс миграций
migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "user=postgres password=password dbname=pr_manager sslmode=disable" down

# Запуск через Docker
docker-up:
	docker-compose up --build

# Остановка Docker
docker-down:
	docker-compose down

# Очистка
clean:
	powershell -Command "if (Test-Path '$(BINARY_NAME)') { Remove-Item '$(BINARY_NAME)' }"
	docker-compose down -v

# Форматирование кода
fmt:
	go fmt ./...

# Health check
health-check:
	curl "http://localhost:8080/health"