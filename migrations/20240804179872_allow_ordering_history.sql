-- +goose Up
CREATE TABLE allow_ordering_audit (
    id SERIAL PRIMARY KEY,
    start_period TIMESTAMPTZ NOT NULL DEFAULT now() UNIQUE,
    end_period TIMESTAMPTZ UNIQUE
);

-- +goose Down
DROP TABLE allow_ordering_audit;