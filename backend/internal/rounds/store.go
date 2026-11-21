package rounds

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store persists round state per campaign.
type Store interface {
	Get(ctx context.Context, campaignID string) (State, error)
	Save(ctx context.Context, campaignID string, s State) error
}

// PgStore is the Postgres Store.
type PgStore struct{ pool *pgxpool.Pool }

// NewStore wraps a pgx pool.
func NewStore(pool *pgxpool.Pool) *PgStore { return &PgStore{pool: pool} }

// Get returns the stored state, or Idle when the campaign never had one.
func (s *PgStore) Get(ctx context.Context, campaignID string) (State, error) {
	st := Idle()
	err := s.pool.QueryRow(ctx, `SELECT active, round, phase FROM campaign_rounds WHERE campaign_id = $1`,
		campaignID).Scan(&st.Active, &st.Round, &st.Phase)
	if errors.Is(err, pgx.ErrNoRows) {
		return Idle(), nil
	}
	return st, err
}

// Save upserts the state.
func (s *PgStore) Save(ctx context.Context, campaignID string, st State) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO campaign_rounds (campaign_id, active, round, phase)
		VALUES ($1,$2,$3,$4) ON CONFLICT (campaign_id) DO UPDATE
		SET active = $2, round = $3, phase = $4, updated_at = now()`,
		campaignID, st.Active, st.Round, st.Phase)
	return err
}
