-- +goose Up
CREATE TABLE tokens (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    scene_id     uuid NOT NULL REFERENCES scenes (id) ON DELETE CASCADE,
    character_id uuid NOT NULL REFERENCES characters (id) ON DELETE CASCADE,
    x            int NOT NULL,
    z            int NOT NULL,
    props        jsonb NOT NULL DEFAULT '{}',
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    UNIQUE (scene_id, character_id)
);

-- +goose Down
DROP TABLE tokens;
