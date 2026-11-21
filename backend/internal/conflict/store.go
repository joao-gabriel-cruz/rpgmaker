package conflict

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store persists conflicts in Postgres.
type Store struct{ pool *pgxpool.Pool }

// NewStore wraps a pgx pool.
func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

const cols = `id, campaign_id, scene_id, title, doc, updated_at`

func scan(row pgx.Row) (*Conflict, error) {
	var c Conflict
	var raw []byte
	err := row.Scan(&c.ID, &c.CampaignID, &c.SceneID, &c.Title, &raw, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &c.Doc); err != nil {
		return nil, err
	}
	return &c, nil
}

// Active returns the campaign's active conflict, or nil.
func (s *Store) Active(ctx context.Context, campaignID string) (*Conflict, error) {
	return scan(s.pool.QueryRow(ctx, `SELECT `+cols+` FROM conflicts WHERE campaign_id = $1 AND active`, campaignID))
}

// Save inserts a new active conflict (empty ID) or updates the existing one.
func (s *Store) Save(ctx context.Context, c Conflict) (Conflict, error) {
	doc, err := json.Marshal(c.Doc)
	if err != nil {
		return Conflict{}, err
	}
	var row pgx.Row
	if c.ID == "" {
		row = s.pool.QueryRow(ctx, `INSERT INTO conflicts (campaign_id, scene_id, title, doc) VALUES ($1,$2,$3,$4) RETURNING `+cols,
			c.CampaignID, c.SceneID, c.Title, doc)
	} else {
		row = s.pool.QueryRow(ctx, `UPDATE conflicts SET scene_id=$2, title=$3, doc=$4, updated_at=now() WHERE id=$1 RETURNING `+cols,
			c.ID, c.SceneID, c.Title, doc)
	}
	saved, err := scan(row)
	if err != nil || saved == nil {
		return Conflict{}, errors.Join(err, errors.New("conflict not saved"))
	}
	return *saved, nil
}

// Close deactivates the campaign's active conflict.
func (s *Store) Close(ctx context.Context, campaignID string) error {
	_, err := s.pool.Exec(ctx, `UPDATE conflicts SET active = false, updated_at = now() WHERE campaign_id = $1 AND active`, campaignID)
	return err
}

// Assign records a roll as spent; the primary key makes a second attempt fail.
func (s *Store) Assign(ctx context.Context, conflictID, rollID, objectiveID string, successes int) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO conflict_assignments (roll_id, conflict_id, objective_id, successes) VALUES ($1,$2,$3,$4)`,
		rollID, conflictID, objectiveID, successes)
	var pg *pgconn.PgError
	if errors.As(err, &pg) && pg.Code == "23505" {
		return ErrAlreadyAssigned
	}
	return err
}
