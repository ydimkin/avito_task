.PHONY: help build test e2e lint up down restart logs logs-core logs-parent clean

help:
	@echo "Avito Kitchen Service — команды:"
	@echo ""
	@echo "  make build           -- сборка сервисов в бинарники в директории ./bin"
	@echo "  make test            -- запуск модульных тестов с детектором гонок"
	@echo "  make e2e             -- запуск сквозного E2E тестирования (bash сценарии)"
	@echo "  make lint            -- запуск линтера"
	@echo "  make up              -- запуск проекта в Docker Compose"
	@echo "  make down            -- остановка проекта в Docker Compose"
	@echo "  make restart         -- перезапуск проекта в Docker Compose"
	@echo "  make logs            -- просмотр всех логов Docker Compose"
	@echo "  make logs-core       -- просмотр логов core-service"
	@echo "  make logs-parent     -- просмотр логов parent-service"
	@echo "  make clean           -- очистка директории ./bin/"
	@echo ""

build:
	mkdir -p bin
	cd core-service && go build -o ../bin/core-service ./cmd/app/main.go
	cd parent-service && go build -o ../bin/parent-service ./cmd/app/main.go

test:
	go test -v -race ./core-service/... ./parent-service/...

e2e:
	./scripts/e2e_test.sh

lint:
	golangci-lint run ./core-service/... ./parent-service/...

up:
	docker compose up --build -d

down:
	docker compose down -v

restart: down up

logs:
	docker compose logs -f

logs-core:
	docker compose logs -f core-service

logs-parent:
	docker compose logs -f parent-service

clean:
	rm -rf bin/
