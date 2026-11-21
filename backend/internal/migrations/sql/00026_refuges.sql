-- +goose Up
CREATE TABLE refuges (
    campaign_id  uuid PRIMARY KEY REFERENCES campaigns (id) ON DELETE CASCADE,
    name         text NOT NULL DEFAULT 'Refúgio',
    description  text NOT NULL DEFAULT '',
    defense      int NOT NULL DEFAULT 0 CHECK (defense BETWEEN 0 AND 10),
    morale       int NOT NULL DEFAULT 0 CHECK (morale BETWEEN 0 AND 10),
    mobility     int NOT NULL DEFAULT 0 CHECK (mobility BETWEEN 0 AND 10),
    belligerence int NOT NULL DEFAULT 0 CHECK (belligerence BETWEEN 0 AND 10),
    water        int NOT NULL DEFAULT 0 CHECK (water >= 0),
    food         int NOT NULL DEFAULT 0 CHECK (food >= 0),
    wood         int NOT NULL DEFAULT 0 CHECK (wood >= 0),
    notes        text NOT NULL DEFAULT '',
    items        jsonb NOT NULL DEFAULT '[]',
    updated_at   timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE refuges;
