package characters

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store persists characters in Postgres.
type Store struct{ pool *pgxpool.Pool }

// NewStore wraps a pgx pool.
func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

const selectCols = `SELECT c.id, c.campaign_id, c.owner_id, COALESCE(u.display_name, ''), c.kind, c.name,
	c.portrait_url, c.sheet, c.updated_at FROM characters c LEFT JOIN users u ON u.id = c.owner_id`

func scan(row pgx.Row) (Character, error) {
	var c Character
	err := row.Scan(&c.ID, &c.CampaignID, &c.OwnerID, &c.OwnerName, &c.Kind, &c.Name, &c.PortraitURL, &c.Sheet, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Character{}, ErrNotFound
	}
	return c, err
}

// Create inserts a character and returns it with the owner name resolved.
func (s *Store) Create(ctx context.Context, c Character) (Character, error) {
	var id string
	err := s.pool.QueryRow(ctx, `INSERT INTO characters (campaign_id, owner_id, kind, name, portrait_url, sheet)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		c.CampaignID, c.OwnerID, c.Kind, c.Name, c.PortraitURL, []byte(c.Sheet)).Scan(&id)
	if err != nil {
		return Character{}, err
	}
	return s.Get(ctx, c.CampaignID, id)
}

// Get loads a character of the campaign, or ErrNotFound.
func (s *Store) Get(ctx context.Context, campaignID, id string) (Character, error) {
	return scan(s.pool.QueryRow(ctx, selectCols+` WHERE c.campaign_id = $1 AND c.id = $2`, campaignID, id))
}

// List returns the campaign's characters; kind "" means every kind.
func (s *Store) List(ctx context.Context, campaignID, kind string) ([]Character, error) {
	rows, err := s.pool.Query(ctx, selectCols+` WHERE c.campaign_id = $1 AND ($2 = '' OR c.kind = $2)
		ORDER BY c.name, c.id`, campaignID, kind)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Character{}
	for rows.Next() {
		c, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// UpdateSheet replaces the sheet and the denormalized name and portrait.
func (s *Store) UpdateSheet(ctx context.Context, campaignID, id string, c Character) (Character, error) {
	tag, err := s.pool.Exec(ctx, `UPDATE characters SET name = $3, portrait_url = $4, sheet = $5, updated_at = now()
		WHERE campaign_id = $1 AND id = $2`, campaignID, id, c.Name, c.PortraitURL, []byte(c.Sheet))
	if err != nil {
		return Character{}, err
	}
	if tag.RowsAffected() == 0 {
		return Character{}, ErrNotFound
	}
	return s.Get(ctx, campaignID, id)
}

// Delete removes a character; ErrNotFound if absent.
func (s *Store) Delete(ctx context.Context, campaignID, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM characters WHERE campaign_id = $1 AND id = $2`, campaignID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
