-- +goose Up
CREATE TABLE campaigns (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name       text NOT NULL,
    system_id  text NOT NULL,
    settings   jsonb NOT NULL DEFAULT '{}',
    owner_id   uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE campaign_members (
    campaign_id uuid NOT NULL REFERENCES campaigns (id) ON DELETE CASCADE,
    user_id     uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role        text NOT NULL CHECK (role IN ('gm', 'player')),
    joined_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (campaign_id, user_id)
);
CREATE INDEX campaign_members_user_idx ON campaign_members (user_id);

CREATE TABLE invites (
    token       text PRIMARY KEY,
    campaign_id uuid NOT NULL REFERENCES campaigns (id) ON DELETE CASCADE,
    role        text NOT NULL CHECK (role IN ('gm', 'player')),
    expires_at  timestamptz NOT NULL,
    used_by     uuid REFERENCES users (id) ON DELETE SET NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX invites_campaign_idx ON invites (campaign_id);

-- +goose Down
DROP TABLE invites;
DROP TABLE campaign_members;
DROP TABLE campaigns;
