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

Install `golangci-lint` if needed:

```sh
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Configure environment

Copy the example environment file:

```sh
cp .env.example .env
```

Set `COMPANIES_LOG_FORMAT` to `text` or `json`. It defaults to `text`. Set `COMPANIES_LOG_LEVEL` to `DEBUG`,
`INFO`, `WARN`, or `ERROR`. It defaults to `INFO`.

`config.yaml` sets the log level to `debug` for local development. Run the service with it:

```sh
make run CONFIG=config.yaml
```

`COMPANIES_LOG_LEVEL` overrides the file value:

```sh
make run CONFIG=config.yaml COMPANIES_LOG_LEVEL=info
```

## Generate the HTTP port

The OpenAPI contract is `api/openapi.yaml`. Regenerate the strict Chi HTTP port after changing it:

```sh
make generate
```

## Start the database

Start PostgreSQL with Docker Compose:

```sh
make deps
```

This starts PostgreSQL in the foreground. In another terminal, continue with migrations.

## Run migrations

Apply all database migrations:

```sh
make migrate-up
```

Check migration status:

```sh
make migrate-status
```

Rollback the latest migration:

```sh
make migrate-down
```

## Run the service

Start the API after applying migrations:

```sh
make run
```

The server listens on `COMPANIES_HTTP_ADDR`. Requests require an HS256 bearer token with `sub`, `exp`, and `iat` claims, signed with `COMPANIES_JWT_SECRET`.

## Exercise the API

With the service running, provide a valid bearer token and run the curl smoke test:

```sh
TOKEN=<jwt> ./scripts/test-api.sh
```

Set `BASE_URL` to target another address. The script exercises one successful request and one validation failure for every endpoint.

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
