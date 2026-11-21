package ambience

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store persists scene ambience.
type Store interface {
	Get(ctx context.Context, sceneID string) (Ambience, error)
	Save(ctx context.Context, a Ambience) error
}

// PgStore is the Postgres Store.
type PgStore struct{ pool *pgxpool.Pool }

// NewStore wraps a pgx pool.
func NewStore(pool *pgxpool.Pool) *PgStore { return &PgStore{pool: pool} }

// Get returns the stored ambience or the default one.
func (s *PgStore) Get(ctx context.Context, sceneID string) (Ambience, error) {
	a := Default(sceneID)
	err := s.pool.QueryRow(ctx, `SELECT time_of_day, weather FROM scene_ambience WHERE scene_id = $1`,
		sceneID).Scan(&a.TimeOfDay, &a.Weather)
	if errors.Is(err, pgx.ErrNoRows) {
		return Default(sceneID), nil
	}
	return a, err
}

// Save upserts the ambience.
func (s *PgStore) Save(ctx context.Context, a Ambience) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO scene_ambience (scene_id, time_of_day, weather) VALUES ($1,$2,$3)
		ON CONFLICT (scene_id) DO UPDATE SET time_of_day = $2, weather = $3, updated_at = now()`,
		a.SceneID, a.TimeOfDay, a.Weather)
	return err
}
