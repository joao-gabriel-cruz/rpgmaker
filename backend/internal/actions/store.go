package actions

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store is the persistence port; it only holds what differs from the defaults.
type Store interface {
	List(ctx context.Context, campaignID string) ([]Action, error)
	Save(ctx context.Context, campaignID string, a Action) error
	Delete(ctx context.Context, campaignID, id string) error
}

// PgStore is the Postgres Store.
type PgStore struct{ pool *pgxpool.Pool }

// NewStore wraps a pgx pool.
func NewStore(pool *pgxpool.Pool) *PgStore { return &PgStore{pool: pool} }

// List returns the stored actions oldest first.
func (s *PgStore) List(ctx context.Context, campaignID string) ([]Action, error) {
	rows, err := s.pool.Query(ctx, `SELECT data FROM campaign_actions WHERE campaign_id = $1 ORDER BY created_at, id`, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Action{}
	for rows.Next() {
		var raw []byte
		var a Action
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &a); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// Save upserts one action.
func (s *PgStore) Save(ctx context.Context, campaignID string, a Action) error {
	raw, err := json.Marshal(a)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `INSERT INTO campaign_actions (campaign_id, id, data) VALUES ($1,$2,$3)
		ON CONFLICT (campaign_id, id) DO UPDATE SET data = $3`, campaignID, a.ID, raw)
	return err
}

// Delete removes a stored action; ErrNotFound when there was none.
func (s *PgStore) Delete(ctx context.Context, campaignID, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM campaign_actions WHERE campaign_id = $1 AND id = $2`, campaignID, id)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}
