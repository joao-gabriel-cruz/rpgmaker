-- +goose Up
-- gen_random_uuid() for the primary keys of upcoming tables.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- +goose Down
DROP EXTENSION IF EXISTS pgcrypto;
