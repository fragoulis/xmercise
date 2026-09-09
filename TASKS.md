# Tasks

[x] Define stack
[x] Define database schema
[x] Setup docker compose (postgres)
[x] Setup database migrations
[x] Define and implement domain models (company)
[x] Define and implement the db adapter
[x] Define and implement domain services (create, update, delete, findOne)
[x] Define openapi spec, automate generation and generate http port (CreateCompany, UpdateCompany, DeleteCompany, GetCompany)
[x] Define and implement the http port
[x] Create the main entrypoint
[x] Expose company type as a plain string through the HTTP and application layers
[x] Standardize Go formatting
[x] Rename internal/infrastructure to internal/adapter
[x] Keep company validation in the application service
[x] Return the domain company description as an optional pointer
[x] Introduce a production Dockerfile
[x] Make log level configurable
[x] Make log format configurable
[x] Add curl API smoke test script
[x] Validate HTTP company IDs in the application layer and return JSON errors
[x] Add GitHub Actions lint and test CI
[x] Add local debug logging config file
[x] Build Docker image locally and in CI
[x] Add main-only publish-artifacts CI placeholder
