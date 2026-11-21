package items

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository is the persistence port.
type Repository interface {
	ListItems(ctx context.Context, campaignID string) ([]Item, error)
	GetItem(ctx context.Context, campaignID, id string) (Item, error)
	CreateItem(ctx context.Context, campaignID string, it Item) (Item, error)
	UpdateItem(ctx context.Context, campaignID string, it Item) (Item, error)
	DeleteItem(ctx context.Context, campaignID, id string) error
	InsertLoot(ctx context.Context, l Loot) (Loot, error)
	ListLoot(ctx context.Context, campaignID, sceneID string) ([]Loot, error)
	GetLoot(ctx context.Context, campaignID, id string) (Loot, error)
	SaveLoot(ctx context.Context, l Loot) error
	DeleteLoot(ctx context.Context, campaignID, id string) error
}

// Store persists items and loot in Postgres.
type Store struct{ pool *pgxpool.Pool }

// NewStore wraps a pgx pool.
func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func scanItem(row pgx.Row) (Item, error) {
	var id string
	var raw []byte
	if err := row.Scan(&id, &raw); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Item{}, ErrNotFound
		}
		return Item{}, err
	}
	var it Item
	if err := json.Unmarshal(raw, &it); err != nil {
		return Item{}, err
	}
	it.ID = id
	if it.Traits == nil {
		it.Traits = []string{}
	}
	return it, nil
}

func encode(it Item) []byte {
	it.ID = ""
	raw, _ := json.Marshal(it)
	return raw
}

// ListItems returns the campaign compendium, oldest first.
func (s *Store) ListItems(ctx context.Context, campaignID string) ([]Item, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, data FROM items WHERE campaign_id = $1 ORDER BY created_at, id`, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Item{}
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// GetItem loads one item of the campaign.
func (s *Store) GetItem(ctx context.Context, campaignID, id string) (Item, error) {
	return scanItem(s.pool.QueryRow(ctx, `SELECT id, data FROM items WHERE campaign_id = $1 AND id = $2`, campaignID, id))
}

// CreateItem inserts an item.
func (s *Store) CreateItem(ctx context.Context, campaignID string, it Item) (Item, error) {
	return scanItem(s.pool.QueryRow(ctx, `INSERT INTO items (campaign_id, data) VALUES ($1, $2) RETURNING id, data`,
		campaignID, encode(it)))
}

// UpdateItem replaces an item's data.
func (s *Store) UpdateItem(ctx context.Context, campaignID string, it Item) (Item, error) {
	return scanItem(s.pool.QueryRow(ctx, `UPDATE items SET data = $3 WHERE campaign_id = $1 AND id = $2 RETURNING id, data`,
		campaignID, it.ID, encode(it)))
}

// DeleteItem removes an item; loot and sheet copies keep their snapshot.
func (s *Store) DeleteItem(ctx context.Context, campaignID, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM items WHERE campaign_id = $1 AND id = $2`, campaignID, id)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}
