-- +goose Up
CREATE TABLE rolls (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id     uuid NOT NULL REFERENCES campaigns (id) ON DELETE CASCADE,
    play_session_id uuid,
    user_id         uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    character_id    uuid REFERENCES characters (id) ON DELETE SET NULL,
    visibility      text NOT NULL CHECK (visibility IN ('public', 'gm')),
    kind            text NOT NULL CHECK (kind IN ('assimilacao', 'generic')),
    spec            jsonb NOT NULL,
    dice            jsonb NOT NULL,
    result          jsonb,
    status          text NOT NULL CHECK (status IN ('pending', 'final')),
    seed            bigint NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX rolls_campaign_idx ON rolls (campaign_id, created_at DESC);

-- +goose Down
DROP TABLE rolls;
