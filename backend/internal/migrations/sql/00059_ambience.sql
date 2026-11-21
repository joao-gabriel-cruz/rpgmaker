-- +goose Up
CREATE TABLE scene_ambience (
    scene_id    uuid PRIMARY KEY REFERENCES scenes (id) ON DELETE CASCADE,
    time_of_day text NOT NULL DEFAULT 'day' CHECK (time_of_day IN ('dawn', 'day', 'dusk', 'night')),
    weather     text NOT NULL DEFAULT 'clear' CHECK (weather IN ('clear', 'rain', 'storm', 'fog', 'snow', 'ash')),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE scene_ambience;
