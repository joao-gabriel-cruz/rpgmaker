-- +goose Up
CREATE TABLE effect_rules (
    campaign_id uuid NOT NULL REFERENCES campaigns (id) ON DELETE CASCADE,
    kind        text NOT NULL CHECK (kind IN ('terrain', 'prefab')),
    key         text NOT NULL,
    rules       jsonb NOT NULL DEFAULT '[]',
    PRIMARY KEY (campaign_id, kind, key)
);

CREATE TABLE effect_suggestions (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id    uuid NOT NULL REFERENCES campaigns (id) ON DELETE CASCADE,
    character_id   uuid NOT NULL REFERENCES characters (id) ON DELETE CASCADE,
    character_name text NOT NULL,
    scene_id       uuid NOT NULL REFERENCES scenes (id) ON DELETE CASCADE,
    x              int NOT NULL,
    z              int NOT NULL,
    source         text NOT NULL,
    trigger        text NOT NULL,
    effect         text NOT NULL,
    amount         int NOT NULL DEFAULT 0,
    message        text NOT NULL DEFAULT '',
    status         text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'applied', 'dismissed')),
    created_at     timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX effect_suggestions_pending ON effect_suggestions (campaign_id, created_at) WHERE status = 'pending';

-- +goose Down
DROP TABLE effect_suggestions;
DROP TABLE effect_rules;
