package vision

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store persists fog in Postgres.
type Store struct{ pool *pgxpool.Pool }

// NewStore wraps a pgx pool.
func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

// Fog loads the scene's fog or the defaults.
func (s *Store) Fog(ctx context.Context, sceneID string) (Fog, error) {
	f := Fog{}
	err := s.pool.QueryRow(ctx, `SELECT mode, vision_radius, revealed FROM fog WHERE scene_id = $1`, sceneID).
		Scan(&f.Mode, &f.Radius, &f.Revealed)
	if errors.Is(err, pgx.ErrNoRows) {
		return Fog{Mode: ModeOff, Radius: DefaultRadius}, nil
	}
	return f, err
}

// SaveConfig upserts mode and radius.
func (s *Store) SaveConfig(ctx context.Context, sceneID, mode string, radius int) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO fog (scene_id, mode, vision_radius) VALUES ($1,$2,$3)
		ON CONFLICT (scene_id) DO UPDATE SET mode = $2, vision_radius = $3, updated_at = now()`, sceneID, mode, radius)
	return err
}

// SaveRevealed upserts the manual reveal bitset.
func (s *Store) SaveRevealed(ctx context.Context, sceneID, revealed string) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO fog (scene_id, revealed) VALUES ($1,$2)
		ON CONFLICT (scene_id) DO UPDATE SET revealed = $2, updated_at = now()`, sceneID, revealed)
	return err
}

// Positions lists the tokens of a scene with their characters' owners.
func (s *Store) Positions(ctx context.Context, sceneID string) ([]Pos, error) {
	rows, err := s.pool.Query(ctx, `SELECT t.id, t.x, t.z, c.owner_id FROM tokens t
		JOIN characters c ON c.id = t.character_id WHERE t.scene_id = $1`, sceneID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Pos{}
	for rows.Next() {
		var p Pos
		if err := rows.Scan(&p.ID, &p.X, &p.Z, &p.OwnerID); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
