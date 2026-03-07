# ================================================
#  Messaging API - Makefile
# ================================================

APP_NAME     := messaging-api
BINARY_NAME  := api
MAIN_PKG     := ./cmd/api
MIGRATE_PKG  := ./cmd/migrate

GO           := go
DOCKER       := docker
DOCKER_COMPOSE := docker compose

# Цвета для вывода (работает в большинстве терминалов)
NO_COLOR    := \033[0m
OK_COLOR    := \033[32;01m
WARN_COLOR  := \033[33;01m
ERROR_COLOR := \033[31;01m
INFO_COLOR  := \033[36;01m

# ================================================
# Основные команды
# ================================================

.PHONY: all build run dev test lint fmt clean help

all: fmt lint test build  ## Всё сразу: форматирование → линтинг → тесты → сборка

build:                  ## Собрать бинарник
	@echo "$(INFO_COLOR)Сборка бинарника...$(NO_COLOR)"
	$(GO) build -o $(BINARY_NAME) $(MAIN_PKG)

run: build              ## Собрать и запустить локально
	@echo "$(OK_COLOR)Запуск приложения...$(NO_COLOR)"
	./$(BINARY_NAME)

dev:                    ## Запуск с hot-reload (требуется air)
	@echo "$(INFO_COLOR)Запуск в режиме разработки с air...$(NO_COLOR)"
	@if command -v air > /dev/null; then \
		air -c .air.toml; \
	else \
		echo "$(WARN_COLOR)air не найден. Установите: go install github.com/air-verse/air@latest$(NO_COLOR)"; \
		exit 1; \
	fi

test:                   ## Запустить все тесты
	@echo "$(INFO_COLOR)Запуск тестов...$(NO_COLOR)"
	$(GO) test ./... -count=1 -race -coverprofile=coverage.out
	$(GO) tool cover -func=coverage.out

lint: vet golangci      ## Запустить все проверки (vet + golangci-lint)

vet:                    ## go vet
	$(GO) vet ./...

golangci:               ## golangci-lint (если установлен)
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "$(WARN_COLOR)golangci-lint не найден. Установите: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NO_COLOR)"; \
	fi

fmt:                    ## go fmt + goimports (если установлен)
	$(GO) fmt ./...
	@if command -v goimports > /dev/null; then \
		goimports -w -l .; \
	fi

clean:                  ## Очистить сгенерированные файлы
	rm -f $(BINARY_NAME) coverage.out

# ================================================
# Docker & Compose
# ================================================

.PHONY: docker-build up down logs ps restart

docker-build:           ## Собрать Docker-образ api
	@echo "$(INFO_COLOR)Сборка Docker-образа $(APP_NAME):local ...$(NO_COLOR)"
	$(DOCKER) build -t $(APP_NAME):local .

up:                     ## Поднять весь стек (docker compose up)
	@echo "$(OK_COLOR)Запуск docker compose...$(NO_COLOR)"
	$(DOCKER_COMPOSE) up --build

upd:                    ## Поднять в фоне (detached)
	$(DOCKER_COMPOSE) up -d --build

down:                   ## Остановить и удалить контейнеры + volumes
	@echo "$(WARN_COLOR)Остановка и очистка...$(NO_COLOR)"
	$(DOCKER_COMPOSE) down -v

logs:                   ## Показать логи api
	$(DOCKER_COMPOSE) logs -f api

ps:                     ## Показать состояние контейнеров
	$(DOCKER_COMPOSE) ps

restart: down up        ## Перезапустить весь стек

# ================================================
# Миграции и вспомогательное
# ================================================

migrate:                ## Запустить миграции
	@echo "$(INFO_COLOR)Применение миграций...$(NO_COLOR)"
	$(GO) run $(MIGRATE_PKG)

migrate-create name=?:  ## Создать новую миграцию (make migrate-create name=create_users_table)
	@if [ -z "$(name)" ]; then \
		echo "$(ERROR_COLOR)Укажите имя миграции: make migrate-create name=имя_миграции$(NO_COLOR)"; \
		exit 1; \
	fi
	@echo "$(INFO_COLOR)Создание миграции: $(name)$(NO_COLOR)"
	# Если используешь goose:
	# goose -dir migrations create $(name) sql
	# Или просто шаблон:
	mkdir -p migrations && touch migrations/$(shell date +%Y%m%d%H%M%S)_$(name).up.sql migrations/$(shell date +%Y%m%d%H%M%S)_$(name).down.sql

# ================================================
# Справка
# ================================================

help:                   ## Показать эту справку
	@echo "$(OK_COLOR)Доступные команды:$(NO_COLOR)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(INFO_COLOR)%-18s$(NO_COLOR) %s\n", $$1, $$2}'
