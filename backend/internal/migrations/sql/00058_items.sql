-- +goose Up
CREATE TABLE items (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id uuid NOT NULL REFERENCES campaigns (id) ON DELETE CASCADE,
    data        jsonb NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX items_campaign ON items (campaign_id, created_at);

CREATE TABLE loot (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    scene_id   uuid NOT NULL REFERENCES scenes (id) ON DELETE CASCADE,
    x          int NOT NULL,
    z          int NOT NULL,
    object_id  text,
    item_id    uuid REFERENCES items (id) ON DELETE SET NULL,
    item       jsonb NOT NULL,
    qty        int NOT NULL DEFAULT 1 CHECK (qty >= 1),
    difficulty int NOT NULL DEFAULT 0 CHECK (difficulty >= 0),
    hidden     boolean NOT NULL DEFAULT true,
    found_by   uuid REFERENCES characters (id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX loot_scene ON loot (scene_id);

-- Suggestions created by using an item are not tied to a scene.
ALTER TABLE effect_suggestions ALTER COLUMN scene_id DROP NOT NULL;

-- +goose Down
ALTER TABLE effect_suggestions ALTER COLUMN scene_id SET NOT NULL;
DROP TABLE loot;
DROP TABLE items;
