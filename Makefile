SHELL := /bin/sh
.DEFAULT_GOAL := help

-include .env

export COMPANIES_HTTP_ADDR
export COMPANIES_DATABASE_URL
export COMPANIES_JWT_SECRET
export COMPANIES_AUTH_ENABLED
export COMPANIES_LOG_FORMAT
export COMPANIES_LOG_LEVEL
export COMPANIES_JWT_ISSUER
export COMPANIES_JWT_AUDIENCE

.PHONY: generate
generate:
	go generate ./api

.PHONY: build
build: ## Build the Docker image.
	docker build --tag exercise .

.PHONY: deps
deps:
	docker compose up

.PHONY: run
run:
	go run ./cmd/exercise

.PHONY: dev
dev: ## Run the service with live reload.
	air

.PHONY: db-shell
db-shell:
	docker compose exec postgres sh -lc 'psql -U "$$POSTGRES_USER" -d "$$POSTGRES_DB"'

.PHONY: migrate-up
migrate-up:
	goose -dir db/migrations postgres "$(COMPANIES_DATABASE_URL)" up

.PHONY: migrate-down
migrate-down:
	goose -dir db/migrations postgres "$(COMPANIES_DATABASE_URL)" down

.PHONY: migrate-status
migrate-status:
	goose -dir db/migrations postgres "$(COMPANIES_DATABASE_URL)" status

.PHONY: migrate-create
migrate-create:
	@test -n "$(name)" || (echo "name is required. Usage: make migrate-create name=create_table" && exit 1)
	goose -dir db/migrations create "$(name)" sql

.PHONY: clean-up
clean-up:
	docker compose down -v

.PHONY: help
help: ## Show this help.
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make <target>\n\nTargets:\n"} /^[a-zA-Z0-9_.-]+:.*##/ {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
