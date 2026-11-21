package scenes

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

// Store persists scenes in Postgres.
type Store struct{ pool *pgxpool.Pool }

// NewStore wraps a pgx pool.
func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

const sceneCols = "id, campaign_id, name, width, depth, grid, objects, draft_grid, draft_objects, is_active, settings, created_at, updated_at"

func scanScene(row pgx.Row) (Scene, error) {
	var s Scene
	var data string
	var objects, draftObjects []byte
	var draftGrid *string
	err := row.Scan(&s.ID, &s.CampaignID, &s.Name, &s.Width, &s.Depth, &data, &objects, &draftGrid, &draftObjects, &s.IsActive,
		&s.Settings, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Scene{}, ErrNotFound
	}
	if err != nil {
		return Scene{}, err
	}
	s.Grid = maps.Serialized{Width: s.Width, Depth: s.Depth, Data: data}
	s.Objects = []maps.Object{}
	if err := json.Unmarshal(objects, &s.Objects); err != nil {
		return Scene{}, err
	}
	if draftGrid != nil {
		s.Draft = &Draft{Grid: maps.Serialized{Width: s.Width, Depth: s.Depth, Data: *draftGrid}, Objects: []maps.Object{}}
		return s, json.Unmarshal(draftObjects, &s.Draft.Objects)
	}
	return s, nil
}

// Create inserts an inactive scene.
func (s *Store) Create(ctx context.Context, campaignID, name string, w, d int, grid string) (Scene, error) {
	return scanScene(s.pool.QueryRow(ctx,
		`INSERT INTO scenes (campaign_id, name, width, depth, grid) VALUES ($1,$2,$3,$4,$5) RETURNING `+sceneCols,
		campaignID, name, w, d, grid))
}

// List returns the campaign's scenes oldest first, without grid data.
func (s *Store) List(ctx context.Context, campaignID string) ([]Summary, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, name, width, depth, is_active FROM scenes WHERE campaign_id = $1 ORDER BY created_at, id`, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Summary{}
	for rows.Next() {
		var m Summary
		if err := rows.Scan(&m.ID, &m.Name, &m.Width, &m.Depth, &m.IsActive); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// Get loads one scene of the campaign.
func (s *Store) Get(ctx context.Context, campaignID, id string) (Scene, error) {
	return scanScene(s.pool.QueryRow(ctx,
		`SELECT `+sceneCols+` FROM scenes WHERE id = $1 AND campaign_id = $2`, id, campaignID))
}

func affected(tag interface{ RowsAffected() int64 }, err error) error {
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Rename changes the scene name.
func (s *Store) Rename(ctx context.Context, campaignID, id, name string) error {
	return affected(s.pool.Exec(ctx,
		`UPDATE scenes SET name = $3, updated_at = now() WHERE id = $1 AND campaign_id = $2`, id, campaignID, name))
}

// Delete removes the scene.
func (s *Store) Delete(ctx context.Context, campaignID, id string) error {
	return affected(s.pool.Exec(ctx, `DELETE FROM scenes WHERE id = $1 AND campaign_id = $2`, id, campaignID))
}

// Activate makes the scene the only active one of its campaign.
func (s *Store) Activate(ctx context.Context, campaignID, id string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx,
		`UPDATE scenes SET is_active = false WHERE campaign_id = $1 AND is_active AND id <> $2`, campaignID, id); err != nil {
		return err
	}
	if err := affected(tx.Exec(ctx,
		`UPDATE scenes SET is_active = true, updated_at = now() WHERE id = $1 AND campaign_id = $2`, id, campaignID)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// SaveGrid stores the serialized grid.
func (s *Store) SaveGrid(ctx context.Context, id, grid string) error {
	return affected(s.pool.Exec(ctx, `UPDATE scenes SET grid = $2, updated_at = now() WHERE id = $1`, id, grid))
}

// SaveObjects stores the object list.
func (s *Store) SaveObjects(ctx context.Context, id string, objects []maps.Object) error {
	raw, err := json.Marshal(objects)
	if err != nil {
		return err
	}
	return affected(s.pool.Exec(ctx, `UPDATE scenes SET objects = $2, updated_at = now() WHERE id = $1`, id, raw))
}
