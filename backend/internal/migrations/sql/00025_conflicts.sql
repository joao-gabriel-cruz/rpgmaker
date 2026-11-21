-- +goose Up
ALTER TABLE rolls DROP CONSTRAINT rolls_kind_check;
ALTER TABLE rolls ADD CONSTRAINT rolls_kind_check CHECK (kind IN ('assimilacao', 'generic', 'threat'));

CREATE TABLE conflicts (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id uuid NOT NULL REFERENCES campaigns (id) ON DELETE CASCADE,
    scene_id    uuid,
    title       text NOT NULL,
    doc         jsonb NOT NULL,
    active      boolean NOT NULL DEFAULT true,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX conflicts_one_active_idx ON conflicts (campaign_id) WHERE active;

-- The roll id as primary key makes "assign a roll once" atomic.
CREATE TABLE conflict_assignments (
    roll_id      uuid PRIMARY KEY REFERENCES rolls (id) ON DELETE CASCADE,
    conflict_id  uuid NOT NULL REFERENCES conflicts (id) ON DELETE CASCADE,
    objective_id text NOT NULL,
    successes    int NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE conflict_assignments;
DROP TABLE conflicts;
DELETE FROM rolls WHERE kind = 'threat';
ALTER TABLE rolls DROP CONSTRAINT rolls_kind_check;
ALTER TABLE rolls ADD CONSTRAINT rolls_kind_check CHECK (kind IN ('assimilacao', 'generic'));
