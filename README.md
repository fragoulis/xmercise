# Companies Service

Microservice for managing companies.

## Prerequisites

- Docker and Docker Compose
- Go, for running `go install` and later the service
- `make` (optional, it is just helpful)
- `goose` migration CLI
- `golangci-lint` linter

Install goose if needed:

```sh
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Install golangci-lint if needed:

```sh
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Configure environment

Copy the example environment file:

```sh
cp .env.example .env
```

The default values start a local PostgreSQL database and build this database URL:

```sh
postgres://root:password@localhost:5432/exercise?sslmode=disable
```

Change `.env` if you need different local credentials or port.

## Start the database

Start PostgreSQL with Docker Compose:

```sh
make deps
```

This runs PostgreSQL in the foreground. In another terminal, continue with migrations.

If you prefer detached mode:

```sh
docker compose up -d
```

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

## Useful commands

Open a PostgreSQL shell:

```sh
make db-shell
```

Create a new migration:

```sh
make migrate-create name=create_example_table
```

Stop local dependencies:

```sh
docker compose down
```

Remove local database data too:

```sh
docker compose down -v
```
