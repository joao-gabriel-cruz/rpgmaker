-- +goose Up
ALTER TABLE scenes ADD COLUMN draft_grid text, ADD COLUMN draft_objects jsonb;

CREATE TABLE scene_snapshots (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    scene_id   uuid NOT NULL REFERENCES scenes (id) ON DELETE CASCADE,
    name       text NOT NULL,
    grid       text NOT NULL,
    objects    jsonb NOT NULL DEFAULT '[]',
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX scene_snapshots_scene_idx ON scene_snapshots (scene_id, created_at);

-- +goose Down
DROP TABLE scene_snapshots;
ALTER TABLE scenes DROP COLUMN draft_grid, DROP COLUMN draft_objects;
