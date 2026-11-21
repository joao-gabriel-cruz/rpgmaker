-- +goose Up
ALTER TABLE scenes DROP CONSTRAINT scenes_width_check, DROP CONSTRAINT scenes_depth_check;
ALTER TABLE scenes ADD CONSTRAINT scenes_width_check CHECK (width BETWEEN 16 AND 128),
    ADD CONSTRAINT scenes_depth_check CHECK (depth BETWEEN 16 AND 128);

-- +goose Down
ALTER TABLE scenes DROP CONSTRAINT scenes_width_check, DROP CONSTRAINT scenes_depth_check;
ALTER TABLE scenes ADD CONSTRAINT scenes_width_check CHECK (width BETWEEN 16 AND 64),
    ADD CONSTRAINT scenes_depth_check CHECK (depth BETWEEN 16 AND 64);
