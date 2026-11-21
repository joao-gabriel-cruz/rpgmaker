// Package sessions owns the play session (start/end) of a campaign.
package sessions

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNoneActive means there is no running session.
var ErrNoneActive = errors.New("no active session")

// ErrAlreadyActive means a session is already running.
var ErrAlreadyActive = errors.New("session already active")

// Session is one play session.
type Session struct {
	ID        string
	StartedAt time.Time
}

// Store persists play sessions.
type Store interface {
	Active(ctx context.Context, campaignID string) (Session, error)
	Start(ctx context.Context, campaignID string) (Session, error)
	End(ctx context.Context, campaignID string) error
}

// PgStore is the Postgres Store.
type PgStore struct{ pool *pgxpool.Pool }

// NewStore wraps a pgx pool.
func NewStore(pool *pgxpool.Pool) *PgStore { return &PgStore{pool: pool} }

// Active returns the running session, or ErrNoneActive.
func (s *PgStore) Active(ctx context.Context, campaignID string) (Session, error) {
	var ses Session
	err := s.pool.QueryRow(ctx, `SELECT id, started_at FROM play_sessions
		WHERE campaign_id = $1 AND ended_at IS NULL`, campaignID).Scan(&ses.ID, &ses.StartedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrNoneActive
	}
	return ses, err
}

// Start opens a session; ErrAlreadyActive when one is running.
func (s *PgStore) Start(ctx context.Context, campaignID string) (Session, error) {
	var ses Session
	err := s.pool.QueryRow(ctx, `INSERT INTO play_sessions (campaign_id) VALUES ($1)
		ON CONFLICT DO NOTHING RETURNING id, started_at`, campaignID).Scan(&ses.ID, &ses.StartedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrAlreadyActive
	}
	return ses, err
}

// End closes the running session; ErrNoneActive when there is none.
func (s *PgStore) End(ctx context.Context, campaignID string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE play_sessions SET ended_at = now()
		WHERE campaign_id = $1 AND ended_at IS NULL`, campaignID)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNoneActive
	}
	return err
}
