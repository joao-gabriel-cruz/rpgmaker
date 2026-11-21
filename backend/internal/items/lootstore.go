package items

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

const lootCols = `l.id, l.scene_id, l.x, l.z, l.object_id, l.item_id, l.item, l.qty, l.difficulty, l.hidden, l.found_by`

func scanLoot(row pgx.Row) (Loot, error) {
	var l Loot
	err := row.Scan(&l.ID, &l.SceneID, &l.X, &l.Z, &l.ObjectID, &l.ItemID, &l.Item, &l.Qty, &l.Difficulty, &l.Hidden, &l.FoundBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return Loot{}, ErrNotFound
	}
	return l, err
}

// InsertLoot stores new loot; the scene must belong to the campaign checked by the caller.
func (s *Store) InsertLoot(ctx context.Context, l Loot) (Loot, error) {
	return scanLoot(s.pool.QueryRow(ctx, `WITH l AS (INSERT INTO loot (scene_id, x, z, object_id, item_id, item, qty, difficulty, hidden)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING *) SELECT `+lootCols+` FROM l`,
		l.SceneID, l.X, l.Z, l.ObjectID, l.ItemID, []byte(l.Item), l.Qty, l.Difficulty, l.Hidden))
}

// ListLoot returns the scene's loot (every row; callers filter by role).
func (s *Store) ListLoot(ctx context.Context, campaignID, sceneID string) ([]Loot, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+lootCols+` FROM loot l JOIN scenes sc ON sc.id = l.scene_id
		WHERE sc.campaign_id = $1 AND l.scene_id = $2 ORDER BY l.created_at, l.id`, campaignID, sceneID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Loot{}
	for rows.Next() {
		l, err := scanLoot(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// GetLoot loads one loot of the campaign.
func (s *Store) GetLoot(ctx context.Context, campaignID, id string) (Loot, error) {
	return scanLoot(s.pool.QueryRow(ctx, `SELECT `+lootCols+` FROM loot l JOIN scenes sc ON sc.id = l.scene_id
		WHERE sc.campaign_id = $1 AND l.id = $2`, campaignID, id))
}

// SaveLoot updates qty, visibility and finder.
func (s *Store) SaveLoot(ctx context.Context, l Loot) error {
	_, err := s.pool.Exec(ctx, `UPDATE loot SET qty = $2, hidden = $3, found_by = $4 WHERE id = $1`, l.ID, l.Qty, l.Hidden, l.FoundBy)
	return err
}

// DeleteLoot removes a loot of the campaign.
func (s *Store) DeleteLoot(ctx context.Context, campaignID, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM loot l USING scenes sc WHERE sc.id = l.scene_id AND sc.campaign_id = $1 AND l.id = $2`,
		campaignID, id)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}
