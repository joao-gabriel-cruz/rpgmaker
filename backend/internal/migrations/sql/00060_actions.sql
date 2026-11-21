-- +goose Up
-- Campaign overrides of the default action catalog and custom actions; data holds the whole action.
CREATE TABLE campaign_actions (
    campaign_id uuid NOT NULL REFERENCES campaigns (id) ON DELETE CASCADE,
    id          text NOT NULL,
    data        jsonb NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (campaign_id, id)
);

-- +goose Down
DROP TABLE campaign_actions;
