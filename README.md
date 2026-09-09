# Companies Service

Microservice for managing companies.

## Prerequisites

- Docker and Docker Compose
- Go, for running `go install` and later the service
- `make` (optional, it is just helpful)
- `goose` migration CLI
- `golangci-lint` linter

Install `goose` if needed:

```sh
go install github.com/pressly/goose/v3/cmd/goose@latest
```

### Optional

Install `golangci-lint` if needed:

```sh
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Configure environment

You can configure the application via environment variables and configuration file.

The precedence goes:

 1. COMPANIES_* environment variables
 2. Values in config.yaml
 3. Built-in defaults

Optionally, you can opy the example environment file:

```sh
cp .env.example .env
```

This is helpful if using environment variables.

## Generate the HTTP port

The OpenAPI contract is defined at `api/openapi.yaml`.

To generate the http port, run:

```sh
make generate
```

## Start the database

Start PostgreSQL in the foreground with Docker Compose:

```sh
make deps
```

In another terminal, continue with migrations.

## Run migrations

Apply all database migrations:

```sh
make migrate-up
```

## Run the service

Start the API after applying migrations:

```sh
make run
```

same as `go run ./cmd/exercise`.

or to use the config with your overrides:

```sh
go run ./cmd/exercise --config config.yaml
```

## Exercise the API

With the service and the database running, provide a valid bearer token and run the curl smoke test:

```sh
TOKEN=<jwt> ./scripts/test-api.sh
```

## Useful commands

Open a PostgreSQL shell:

```sh
make db-shell
```

Create a new migration:

```sh
make migrate-create name=create_example_table
```

## Clean up

Remove local database data:

```sh
make clean-up
```
