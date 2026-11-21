package mcp

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgLinkRepo is the Postgres LinkRepo.
type PgLinkRepo struct{ pool *pgxpool.Pool }

// NewLinkRepo wraps a pgx pool.
func NewLinkRepo(pool *pgxpool.Pool) *PgLinkRepo { return &PgLinkRepo{pool: pool} }

const linkCols = `id, campaign_id, user_id, created_at, last_used_at`

func scanLink(row pgx.Row, notFound error) (Link, error) {
	var l Link
	err := row.Scan(&l.ID, &l.CampaignID, &l.UserID, &l.CreatedAt, &l.LastUsedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Link{}, notFound
	}
	return l, err
}

// Replace revokes the active link and inserts the new one atomically.
func (s *PgLinkRepo) Replace(ctx context.Context, campaignID, userID, hash string) (Link, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Link{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `UPDATE mcp_links SET revoked_at = now()
		WHERE campaign_id = $1 AND user_id = $2 AND revoked_at IS NULL`, campaignID, userID); err != nil {
		return Link{}, err
	}
	l, err := scanLink(tx.QueryRow(ctx, `INSERT INTO mcp_links (campaign_id, user_id, secret_hash)
		VALUES ($1,$2,$3) RETURNING `+linkCols, campaignID, userID, hash), ErrNoLink)
	if err != nil {
		return Link{}, err
	}
	return l, tx.Commit(ctx)
}

// Active returns the non-revoked link or ErrNoLink.
func (s *PgLinkRepo) Active(ctx context.Context, campaignID, userID string) (Link, error) {
	return scanLink(s.pool.QueryRow(ctx, `SELECT `+linkCols+` FROM mcp_links
		WHERE campaign_id = $1 AND user_id = $2 AND revoked_at IS NULL`, campaignID, userID), ErrNoLink)
}

// ByHash returns the non-revoked link with that hash or ErrInvalidLink.
func (s *PgLinkRepo) ByHash(ctx context.Context, hash string) (Link, error) {
	return scanLink(s.pool.QueryRow(ctx, `SELECT `+linkCols+` FROM mcp_links
		WHERE secret_hash = $1 AND revoked_at IS NULL`, hash), ErrInvalidLink)
}

// Revoke marks the active link revoked; no-op without one.
func (s *PgLinkRepo) Revoke(ctx context.Context, campaignID, userID string) error {
	_, err := s.pool.Exec(ctx, `UPDATE mcp_links SET revoked_at = now()
		WHERE campaign_id = $1 AND user_id = $2 AND revoked_at IS NULL`, campaignID, userID)
	return err
}

// Touch records a use of the link.
func (s *PgLinkRepo) Touch(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `UPDATE mcp_links SET last_used_at = now() WHERE id = $1`, id)
	return err
}
