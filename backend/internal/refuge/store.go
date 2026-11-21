package refuge

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MoveFunc computes the new sheet and refuge items from the locked current state.
type MoveFunc func(c CharState, items []json.RawMessage) (json.RawMessage, []json.RawMessage, error)

// Store persists refuges in Postgres.
type Store struct{ pool *pgxpool.Pool }

// NewStore wraps a pgx pool.
func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

const cols = `campaign_id, name, description, defense, morale, mobility, belligerence, water, food, wood, notes, items, updated_at`

func scan(row pgx.Row) (Refuge, error) {
	var r Refuge
	var items []byte
	err := row.Scan(&r.CampaignID, &r.Name, &r.Description, &r.Defense, &r.Morale, &r.Mobility,
		&r.Belligerence, &r.Water, &r.Food, &r.Wood, &r.Notes, &items, &r.UpdatedAt)
	if err != nil {
		return Refuge{}, err
	}
	r.Items = []json.RawMessage{}
	return r, json.Unmarshal(items, &r.Items)
}

type queryer interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func ensure(ctx context.Context, q queryer, campaignID, lock string) (Refuge, error) {
	if err := q.QueryRow(ctx, `INSERT INTO refuges (campaign_id) VALUES ($1) ON CONFLICT DO NOTHING RETURNING 1`,
		campaignID).Scan(new(int)); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return Refuge{}, err
	}
	return scan(q.QueryRow(ctx, `SELECT `+cols+` FROM refuges WHERE campaign_id = $1`+lock, campaignID))
}

// Get returns the campaign's refuge, creating an empty one on demand.
func (s *Store) Get(ctx context.Context, campaignID string) (Refuge, error) {
	return ensure(ctx, s.pool, campaignID, "")
}

// Update replaces the editable fields.
func (s *Store) Update(ctx context.Context, campaignID string, d Details) (Refuge, error) {
	if _, err := s.Get(ctx, campaignID); err != nil {
		return Refuge{}, err
	}
	return scan(s.pool.QueryRow(ctx, `UPDATE refuges SET name=$2, description=$3, defense=$4, morale=$5, mobility=$6,
		belligerence=$7, water=$8, food=$9, wood=$10, notes=$11, updated_at=now() WHERE campaign_id=$1 RETURNING `+cols,
		campaignID, d.Name, d.Description, d.Defense, d.Morale, d.Mobility, d.Belligerence, d.Water, d.Food, d.Wood, d.Notes))
}

// Move runs fn over the locked refuge and character rows and writes both atomically.
func (s *Store) Move(ctx context.Context, campaignID, charID string, fn MoveFunc) (Refuge, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Refuge{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	r, err := ensure(ctx, tx, campaignID, " FOR UPDATE")
	if err != nil {
		return Refuge{}, err
	}
	c, err := lockCharacter(ctx, tx, campaignID, charID)
	if err != nil {
		return Refuge{}, err
	}
	sheet, items, err := fn(c, r.Items)
	if err != nil {
		return Refuge{}, err
	}
	out, err := writeMove(ctx, tx, campaignID, charID, sheet, items)
	if err != nil {
		return Refuge{}, err
	}
	return out, tx.Commit(ctx)
}

func lockCharacter(ctx context.Context, tx pgx.Tx, campaignID, charID string) (CharState, error) {
	var c CharState
	err := tx.QueryRow(ctx, `SELECT owner_id, kind, sheet FROM characters WHERE campaign_id=$1 AND id=$2 FOR UPDATE`,
		campaignID, charID).Scan(&c.OwnerID, &c.Kind, &c.Sheet)
	if errors.Is(err, pgx.ErrNoRows) {
		return c, ErrCharNotFound
	}
	return c, err
}

func writeMove(ctx context.Context, tx pgx.Tx, campaignID, charID string, sheet json.RawMessage, items []json.RawMessage) (Refuge, error) {
	if _, err := tx.Exec(ctx, `UPDATE characters SET sheet=$3, updated_at=now() WHERE campaign_id=$1 AND id=$2`,
		campaignID, charID, []byte(sheet)); err != nil {
		return Refuge{}, err
	}
	raw, _ := json.Marshal(items)
	return scan(tx.QueryRow(ctx, `UPDATE refuges SET items=$2, updated_at=now() WHERE campaign_id=$1 RETURNING `+cols,
		campaignID, raw))
}

// AddItem appends an item unless the refuge is full.
func (s *Store) AddItem(ctx context.Context, campaignID string, item json.RawMessage) (Refuge, error) {
	if _, err := s.Get(ctx, campaignID); err != nil {
		return Refuge{}, err
	}
	r, err := scan(s.pool.QueryRow(ctx, `UPDATE refuges SET items = items || jsonb_build_array($2::jsonb), updated_at=now()
		WHERE campaign_id=$1 AND jsonb_array_length(items) < $3 RETURNING `+cols, campaignID, []byte(item), maxItems))
	if errors.Is(err, pgx.ErrNoRows) {
		return Refuge{}, errRefugeMax
	}
	return r, err
}

var resourceCols = map[string]string{"water": "water", "food": "food", "wood": "wood"}

// AddResource increments one resource column from a fixed whitelist.
func (s *Store) AddResource(ctx context.Context, campaignID, resource string, amount int) (Refuge, error) {
	col, ok := resourceCols[resource]
	if !ok {
		return Refuge{}, ErrInvalid
	}
	if _, err := s.Get(ctx, campaignID); err != nil {
		return Refuge{}, err
	}
	return scan(s.pool.QueryRow(ctx, `UPDATE refuges SET `+col+` = `+col+` + $2, updated_at=now() WHERE campaign_id=$1 RETURNING `+cols,
		campaignID, amount))
}
