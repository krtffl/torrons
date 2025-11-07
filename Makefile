REVISION := $(shell git rev-parse --short=8 HEAD || echo unknown)
VERSION ?= $(shell git describe --tags --always)
BUILT := $(shell date +%Y-%m-%dT%H:%M:%S)
BRANCH := $(shell git show-ref | grep "$(REVISION)" | grep -v HEAD | awk '{print $$2}' | sed 's|refs/remotes/origin/||' | sed 's|refs/heads/||' | sort | head -n 1)
MODULE_NAME := $(shell cat go.mod | head -n 1 | cut -d " " -f2- | cut -d "/" -f2-)
BASE_DIR := $(PWD)

GO_LDFLAGS ?= -X github.com/krtffl/torrons/internal/version.Version=$(VERSION) \
			-X github.com/krtffl/torrons/internal/version.Branch=$(BRANCH) \
			-X github.com/krtffl/torrons/internal/version.Revision=$(REVISION) \
			-X github.com/krtffl/torrons/internal/version.Built=$(BUILT) \
			-s \
			-w

BUILD_DIR=build
DIST_DIR=out

# Load environment variables from .env if it exists
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Database connection string
DB_URL := postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)

.PHONY: run build dist dist-arm64 clean migrate migrate-up migrate-down migrate-create migrate-version help

run:
	go run cmd/server/main.go

build: clean
	CGO_ENABLED=0 go build -a -ldflags "$(GO_LDFLAGS)" -o="$(BUILD_DIR)/server" ./cmd/server

dist: clean
	CGO_ENABLED=0 go build -a -ldflags "$(GO_LDFLAGS)" -o="$(DIST_DIR)/server" ./cmd/server
	@mkdir $(DIST_DIR)/config && cp ./config/config.default.yaml $(DIST_DIR)/config

dist-arm64: clean
	CGO_ENABLED=0 CC=aarch64-linux-gnu-gcc GOOS=linux GOARCH=arm64 go build -a -ldflags "$(GO_LDFLAGS)" -o="$(DIST_DIR)/server" ./cmd/server
	@mkdir $(DIST_DIR)/config && cp ./config/config.default.yaml $(DIST_DIR)/config

clean:
	@rm -rf ./${BUILD_DIR} ./${DIST_DIR}

# Migration commands
migrate: migrate-up

migrate-up:
	@echo "Running migrations..."
	@migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	@echo "Rolling back last migration..."
	@migrate -path migrations -database "$(DB_URL)" down 1

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

migrate-version:
	@migrate -path migrations -database "$(DB_URL)" version

migrate-force:
	@read -p "Enter version to force: " version; \
	migrate -path migrations -database "$(DB_URL)" force $$version

# Help command
help:
	@echo "Available commands:"
	@echo "  make run              - Run the application"
	@echo "  make build            - Build the application"
	@echo "  make dist             - Build production binary"
	@echo "  make migrate          - Run all pending migrations (alias for migrate-up)"
	@echo "  make migrate-up       - Run all pending migrations"
	@echo "  make migrate-down     - Rollback last migration"
	@echo "  make migrate-create   - Create a new migration file"
	@echo "  make migrate-version  - Show current migration version"
	@echo "  make migrate-force    - Force migration to specific version (use with caution)"
	@echo "  make clean            - Clean build artifacts"
	@echo "  make help             - Show this help message"