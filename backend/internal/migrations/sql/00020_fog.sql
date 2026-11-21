-- +goose Up
CREATE TABLE fog (
    scene_id      uuid PRIMARY KEY REFERENCES scenes (id) ON DELETE CASCADE,
    mode          text NOT NULL DEFAULT 'off' CHECK (mode IN ('off', 'manual', 'dynamic')),
    vision_radius int NOT NULL DEFAULT 12,
    revealed      text NOT NULL DEFAULT '',
    updated_at    timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE fog;
