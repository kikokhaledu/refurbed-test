GO ?= go
COMPOSE ?= docker compose
BACKEND_DIR ?= backend
GOLANGCI_LINT_VERSION ?= v1.64.8
GOLANGCI_LINT ?= $(GO) run github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

ifeq ($(OS),Windows_NT)
NULL := NUL
else
NULL := /dev/null
endif

.PHONY: help up down reset seed logs test test-backend test-race fmt vet lint compose-check pre-push clean

help:
	@echo Available targets:
	@echo   make up            - Build and start backend + Vue frontend with Docker Compose
	@echo   make down          - Stop Compose services
	@echo   make reset         - Recreate full stack from scratch
	@echo   make seed          - Validate seed data files
	@echo   make logs          - Tail backend + frontend logs
	@echo   make test          - Run backend tests
	@echo   make test-race     - Run backend race tests (requires CGO + gcc)
	@echo   make fmt           - Format backend Go code
	@echo   make vet           - Run go vet on backend
	@echo   make lint          - Run golangci-lint on backend
	@echo   make compose-check - Validate docker-compose.yml
	@echo   make pre-push      - Run formatting, vet, lint, tests, and compose validation
	@echo   make clean         - Stop Compose services and remove volumes

up: compose-check
	$(COMPOSE) up --build -d backend frontend

down:
	$(COMPOSE) down --remove-orphans

reset: down seed up

seed:
	@test -f $(BACKEND_DIR)/data/metadata.json
	@test -f $(BACKEND_DIR)/data/details.json
	@echo "Seed data is present in $(BACKEND_DIR)/data."

logs:
	$(COMPOSE) logs -f backend frontend

test: test-backend

test-backend:
	cd $(BACKEND_DIR) && $(GO) test ./...

test-race:
	cd $(BACKEND_DIR) && CGO_ENABLED=1 $(GO) test ./... -race

fmt:
	cd $(BACKEND_DIR) && $(GO) fmt ./...

vet:
	cd $(BACKEND_DIR) && $(GO) vet ./...

lint:
	cd $(BACKEND_DIR) && $(GOLANGCI_LINT) run ./...

compose-check:
	$(COMPOSE) config > $(NULL)

pre-push: fmt vet lint test compose-check
	@echo "Pre-push checks passed."

clean:
	$(COMPOSE) down --remove-orphans --volumes
