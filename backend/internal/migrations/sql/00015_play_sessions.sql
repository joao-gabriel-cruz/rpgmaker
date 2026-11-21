-- +goose Up
CREATE TABLE play_sessions (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id uuid NOT NULL REFERENCES campaigns (id) ON DELETE CASCADE,
    started_at  timestamptz NOT NULL DEFAULT now(),
    ended_at    timestamptz
);
CREATE UNIQUE INDEX play_sessions_active_idx ON play_sessions (campaign_id) WHERE ended_at IS NULL;

CREATE TABLE roll_difficulties (
    campaign_id  uuid NOT NULL REFERENCES campaigns (id) ON DELETE CASCADE,
    character_id uuid NOT NULL REFERENCES characters (id) ON DELETE CASCADE,
    d12          int NOT NULL CHECK (d12 BETWEEN 1 AND 3),
    PRIMARY KEY (campaign_id, character_id)
);

-- +goose Down
DROP TABLE roll_difficulties;
DROP TABLE play_sessions;
