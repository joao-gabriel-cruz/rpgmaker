-- +goose Up
CREATE TABLE mcp_links (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id  uuid NOT NULL REFERENCES campaigns (id) ON DELETE CASCADE,
    user_id      uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    secret_hash  text NOT NULL UNIQUE,
    created_at   timestamptz NOT NULL DEFAULT now(),
    last_used_at timestamptz,
    revoked_at   timestamptz
);
CREATE UNIQUE INDEX mcp_links_one_active ON mcp_links (campaign_id, user_id) WHERE revoked_at IS NULL;

-- +goose Down
DROP TABLE mcp_links;
