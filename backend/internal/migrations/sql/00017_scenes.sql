-- +goose Up
CREATE TABLE scenes (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id uuid NOT NULL REFERENCES campaigns (id) ON DELETE CASCADE,
    name        text NOT NULL,
    width       int NOT NULL CHECK (width BETWEEN 16 AND 64),
    depth       int NOT NULL CHECK (depth BETWEEN 16 AND 64),
    grid        text NOT NULL,
    objects     jsonb NOT NULL DEFAULT '[]',
    is_active   boolean NOT NULL DEFAULT false,
    settings    jsonb NOT NULL DEFAULT '{}',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX scenes_campaign_idx ON scenes (campaign_id, created_at);
CREATE UNIQUE INDEX scenes_one_active_idx ON scenes (campaign_id) WHERE is_active;

-- +goose Down
DROP TABLE scenes;
