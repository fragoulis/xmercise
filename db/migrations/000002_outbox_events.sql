-- +goose Up
CREATE TABLE IF NOT EXISTS outbox_events (
    id uuid PRIMARY KEY,
    event_type text NOT NULL,
    aggregate_type text NOT NULL,
    aggregate_id uuid NOT NULL,
    occurred_at timestamptz NOT NULL,
    payload jsonb NOT NULL,
    created_at timestamptz NOT NULL,
    published_at timestamptz NULL,
    attempts integer NOT NULL,
    last_error text NULL
);

CREATE INDEX IF NOT EXISTS outbox_events_unpublished_idx
    ON outbox_events (created_at, id)
    WHERE published_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS outbox_events_unpublished_idx;
DROP TABLE IF EXISTS outbox_events;
