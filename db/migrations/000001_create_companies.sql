-- +goose Up
CREATE TABLE IF NOT EXISTS companies (
    id uuid PRIMARY KEY,
    name text NOT NULL,
    description text NULL,
    employees_count integer NOT NULL,
    registered boolean NOT NULL,
    type text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS companies_name_lower_unique_idx
    ON companies (lower(name));

-- +goose Down
DROP INDEX IF EXISTS companies_name_lower_unique_idx;
DROP TABLE IF EXISTS companies;
