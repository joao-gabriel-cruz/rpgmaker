-- +goose Up
CREATE TABLE drawings (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    scene_id   uuid NOT NULL REFERENCES scenes (id) ON DELETE CASCADE,
    user_id    uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    color      text NOT NULL,
    width      int NOT NULL,
    points     jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX drawings_scene_idx ON drawings (scene_id, created_at);

-- +goose Down
DROP TABLE drawings;
