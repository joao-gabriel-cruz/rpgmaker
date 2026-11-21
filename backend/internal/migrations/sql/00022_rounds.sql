-- +goose Up
CREATE TABLE campaign_rounds (
    campaign_id uuid PRIMARY KEY REFERENCES campaigns (id) ON DELETE CASCADE,
    active      boolean NOT NULL DEFAULT false,
    round       int NOT NULL DEFAULT 0,
    phase       text NOT NULL DEFAULT 'preparation'
        CHECK (phase IN ('preparation', 'infected', 'assimilator', 'end')),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE campaign_rounds;
