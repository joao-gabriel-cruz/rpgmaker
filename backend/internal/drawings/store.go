package drawings

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store persists drawings in Postgres.
type Store struct{ pool *pgxpool.Pool }

// NewStore wraps a pgx pool.
func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

const selectCols = `SELECT d.id, d.scene_id, d.user_id, d.color, d.width, d.points, d.created_at
	FROM drawings d JOIN scenes s ON s.id = d.scene_id`

func scan(row pgx.Row) (Drawing, error) {
	var d Drawing
	var raw []byte
	err := row.Scan(&d.ID, &d.SceneID, &d.UserID, &d.Color, &d.Width, &raw, &d.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Drawing{}, ErrNotFound
	}
	if err != nil {
		return Drawing{}, err
	}
	return d, json.Unmarshal(raw, &d.Points)
}

// Add inserts a stroke.
func (s *Store) Add(ctx context.Context, d Drawing) (Drawing, error) {
	raw, err := json.Marshal(d.Points)
	if err != nil {
		return Drawing{}, err
	}
	return scan(s.pool.QueryRow(ctx, `WITH ins AS (
		INSERT INTO drawings (scene_id, user_id, color, width, points) VALUES ($1,$2,$3,$4,$5) RETURNING *)
		SELECT ins.id, ins.scene_id, ins.user_id, ins.color, ins.width, ins.points, ins.created_at
		FROM ins JOIN scenes s ON s.id = ins.scene_id`, d.SceneID, d.UserID, d.Color, d.Width, raw))
}

// Get loads a drawing of the campaign.
func (s *Store) Get(ctx context.Context, campaignID, id string) (Drawing, error) {
	return scan(s.pool.QueryRow(ctx, selectCols+` WHERE d.id = $1 AND s.campaign_id = $2`, id, campaignID))
}

// List returns the strokes of a scene, oldest first.
func (s *Store) List(ctx context.Context, campaignID, sceneID string) ([]Drawing, error) {
	rows, err := s.pool.Query(ctx,
		selectCols+` WHERE d.scene_id = $1 AND s.campaign_id = $2 ORDER BY d.created_at, d.id`, sceneID, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Drawing{}
	for rows.Next() {
		d, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// Remove deletes one stroke.
func (s *Store) Remove(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM drawings WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Clear deletes every stroke of a scene.
func (s *Store) Clear(ctx context.Context, sceneID string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM drawings WHERE scene_id = $1`, sceneID)
	return err
}
