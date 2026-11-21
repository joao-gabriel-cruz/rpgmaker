-- +goose Up
CREATE TABLE characters (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id  uuid NOT NULL REFERENCES campaigns (id) ON DELETE CASCADE,
    owner_id     uuid REFERENCES users (id) ON DELETE SET NULL,
    kind         text NOT NULL CHECK (kind IN ('pc', 'npc')),
    name         text NOT NULL,
    portrait_url text,
    sheet        jsonb NOT NULL DEFAULT '{}',
    updated_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX characters_campaign_idx ON characters (campaign_id);

-- +goose Down
DROP TABLE characters;
