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

// MutateFunc edits the locked refuge in place; returning an error aborts without writing.
type MutateFunc func(r *Refuge) error

// Store persists refuges in Postgres.
type Store struct{ pool *pgxpool.Pool }

// NewStore wraps a pgx pool.
func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

const cols = `campaign_id, name, description, notes, state, items, updated_at`

func scan(row pgx.Row) (Refuge, error) {
	var r Refuge
	var state, items []byte
	if err := row.Scan(&r.CampaignID, &r.Name, &r.Description, &r.Notes, &state, &items, &r.UpdatedAt); err != nil {
		return Refuge{}, err
	}
	r.Items = []json.RawMessage{}
	if err := json.Unmarshal(items, &r.Items); err != nil {
		return Refuge{}, err
	}
	return r, decodeState(state, &r)
}

type queryer interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func ensure(ctx context.Context, q queryer, campaignID, lock string) (Refuge, error) {
	state, _ := encodeState(Refuge{Editable: Editable{Sheet: DefaultSheet()}, CrisisSettings: DefaultCrisisSettings()})
	if err := q.QueryRow(ctx, `INSERT INTO refuges (campaign_id, state) VALUES ($1, $2) ON CONFLICT DO NOTHING RETURNING 1`,
		campaignID, state).Scan(new(int)); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return Refuge{}, err
	}
	return scan(q.QueryRow(ctx, `SELECT `+cols+` FROM refuges WHERE campaign_id = $1`+lock, campaignID))
}

// Get returns the campaign's refuge, creating an empty one on demand.
func (s *Store) Get(ctx context.Context, campaignID string) (Refuge, error) {
	return ensure(ctx, s.pool, campaignID, "")
}

// Mutate runs fn over the row-locked refuge and writes the result atomically.
func (s *Store) Mutate(ctx context.Context, campaignID string, fn MutateFunc) (Refuge, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Refuge{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	r, err := ensure(ctx, tx, campaignID, " FOR UPDATE")
	if err != nil {
		return Refuge{}, err
	}
	if err := fn(&r); err != nil {
		return Refuge{}, err
	}
	out, err := write(ctx, tx, r)
	if err != nil {
		return Refuge{}, err
	}
	return out, tx.Commit(ctx)
}

func write(ctx context.Context, tx pgx.Tx, r Refuge) (Refuge, error) {
	state, err := encodeState(r)
	if err != nil {
		return Refuge{}, err
	}
	items, _ := json.Marshal(emptyList(r.Items))
	return scan(tx.QueryRow(ctx, `UPDATE refuges SET name=$2, description=$3, notes=$4, state=$5, items=$6, updated_at=now()
		WHERE campaign_id=$1 RETURNING `+cols, r.CampaignID, r.Name, r.Description, r.Notes, state, items))
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
	if _, err := tx.Exec(ctx, `UPDATE characters SET sheet=$3, updated_at=now() WHERE campaign_id=$1 AND id=$2`,
		campaignID, charID, []byte(sheet)); err != nil {
		return Refuge{}, err
	}
	r.Items = items
	out, err := write(ctx, tx, r)
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
