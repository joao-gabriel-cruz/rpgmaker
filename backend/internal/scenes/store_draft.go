package scenes

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

// SaveDraft stores the draft grid and objects; nil clears the draft.
func (s *Store) SaveDraft(ctx context.Context, id string, d *Draft) error {
	if d == nil {
		return affected(s.pool.Exec(ctx, `UPDATE scenes SET draft_grid = NULL, draft_objects = NULL WHERE id = $1`, id))
	}
	raw, err := json.Marshal(d.Objects)
	if err != nil {
		return err
	}
	return affected(s.pool.Exec(ctx, `UPDATE scenes SET draft_grid = $2, draft_objects = $3 WHERE id = $1`, id, d.Grid.Data, raw))
}

// Replace overwrites the real state and drops the draft in one statement.
func (s *Store) Replace(ctx context.Context, id, grid string, objects []maps.Object) error {
	raw, err := json.Marshal(objects)
	if err != nil {
		return err
	}
	return affected(s.pool.Exec(ctx,
		`UPDATE scenes SET grid = $2, objects = $3, draft_grid = NULL, draft_objects = NULL, updated_at = now() WHERE id = $1`,
		id, grid, raw))
}

// CreateSnapshot saves a named state of the scene.
func (s *Store) CreateSnapshot(ctx context.Context, sceneID, name, grid string, objects []maps.Object) (Snapshot, error) {
	raw, err := json.Marshal(objects)
	if err != nil {
		return Snapshot{}, err
	}
	var sn Snapshot
	err = s.pool.QueryRow(ctx,
		`INSERT INTO scene_snapshots (scene_id, name, grid, objects) VALUES ($1,$2,$3,$4) RETURNING id, scene_id, name, created_at`,
		sceneID, name, grid, raw).Scan(&sn.ID, &sn.SceneID, &sn.Name, &sn.CreatedAt)
	return sn, err
}

// ListSnapshots returns the scene's snapshots, newest first, without payload.
func (s *Store) ListSnapshots(ctx context.Context, sceneID string) ([]Snapshot, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, scene_id, name, created_at FROM scene_snapshots WHERE scene_id = $1 ORDER BY created_at DESC, id`, sceneID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Snapshot{}
	for rows.Next() {
		var sn Snapshot
		if err := rows.Scan(&sn.ID, &sn.SceneID, &sn.Name, &sn.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, sn)
	}
	return out, rows.Err()
}

// GetSnapshot loads one snapshot with its payload.
func (s *Store) GetSnapshot(ctx context.Context, sceneID, id string) (SnapshotData, error) {
	var d SnapshotData
	var raw []byte
	err := s.pool.QueryRow(ctx,
		`SELECT id, scene_id, name, created_at, grid, objects FROM scene_snapshots WHERE id = $1 AND scene_id = $2`, id, sceneID).
		Scan(&d.ID, &d.SceneID, &d.Name, &d.CreatedAt, &d.Grid, &raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return SnapshotData{}, ErrNotFound
	}
	if err != nil {
		return SnapshotData{}, err
	}
	d.Objects = []maps.Object{}
	return d, json.Unmarshal(raw, &d.Objects)
}

// DeleteSnapshot removes a snapshot.
func (s *Store) DeleteSnapshot(ctx context.Context, sceneID, id string) error {
	return affected(s.pool.Exec(ctx, `DELETE FROM scene_snapshots WHERE id = $1 AND scene_id = $2`, id, sceneID))
}
