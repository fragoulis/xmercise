SHELL := /bin/sh
.DEFAULT_GOAL := help

include .env

.PHONY: generate
generate: ## Generate HTTP port from the OpenAPI contract.
	go generate ./api

.PHONY: deps
deps: ## Start local dependencies.
	docker compose up

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
