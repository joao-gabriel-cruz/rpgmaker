package tokens

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store persists tokens in Postgres.
type Store struct{ pool *pgxpool.Pool }

// NewStore wraps a pgx pool.
func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

const selectCols = `SELECT t.id, t.scene_id, t.character_id, t.x, t.z, t.props, c.name, c.portrait_url, c.owner_id,
	c.sheet->'appearance', CASE WHEN jsonb_typeof(c.sheet#>'{tugOfWar,assimilationLevel}') = 'number' THEN LEAST(GREATEST((c.sheet#>>'{tugOfWar,assimilationLevel}')::numeric, 0), 10)::int END,
	jsonb_strip_nulls(jsonb_build_object('conditions', c.sheet->'conditions', 'inventory', c.sheet->'inventory',
		'health', c.sheet->'health', 'tugOfWar', c.sheet->'tugOfWar', 'instincts', c.sheet->'instincts'))
	FROM tokens t JOIN characters c ON c.id = t.character_id JOIN scenes s ON s.id = t.scene_id`

func scan(row pgx.Row) (Token, error) {
	var t Token
	err := row.Scan(&t.ID, &t.SceneID, &t.CharacterID, &t.X, &t.Z, &t.Props, &t.Name, &t.PortraitURL, &t.OwnerID, &t.Appearance, &t.AssimilationLevel, &t.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return Token{}, ErrNotFound
	}
	return t, err
}

// Place inserts a token; one per character per scene.
func (s *Store) Place(ctx context.Context, sceneID, characterID string, x, z int) (Token, error) {
	var id string
	err := s.pool.QueryRow(ctx,
		`INSERT INTO tokens (scene_id, character_id, x, z) VALUES ($1,$2,$3,$4) RETURNING id`,
		sceneID, characterID, x, z).Scan(&id)
	var pe *pgconn.PgError
	if errors.As(err, &pe) && pe.Code == "23505" {
		return Token{}, ErrExists
	}
	if err != nil {
		return Token{}, err
	}
	return scan(s.pool.QueryRow(ctx, selectCols+` WHERE t.id = $1`, id))
}

// Get loads a token of the campaign.
func (s *Store) Get(ctx context.Context, campaignID, id string) (Token, error) {
	return scan(s.pool.QueryRow(ctx, selectCols+` WHERE t.id = $1 AND s.campaign_id = $2`, id, campaignID))
}

// List returns the tokens of a scene.
func (s *Store) List(ctx context.Context, campaignID, sceneID string) ([]Token, error) {
	rows, err := s.pool.Query(ctx,
		selectCols+` WHERE t.scene_id = $1 AND s.campaign_id = $2 ORDER BY t.created_at, t.id`, sceneID, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Token{}
	for rows.Next() {
		t, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// Move stores the new cell.
func (s *Store) Move(ctx context.Context, id string, x, z int) error {
	return affected(s.pool.Exec(ctx, `UPDATE tokens SET x = $2, z = $3, updated_at = now() WHERE id = $1`, id, x, z))
}

// Remove deletes the token.
func (s *Store) Remove(ctx context.Context, id string) error {
	return affected(s.pool.Exec(ctx, `DELETE FROM tokens WHERE id = $1`, id))
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

// Character loads the character of the campaign (any kind).
func (s *Store) Character(ctx context.Context, campaignID, id string) (Char, error) {
	var c Char
	err := s.pool.QueryRow(ctx, `SELECT id, name, owner_id FROM characters WHERE id = $1 AND campaign_id = $2`,
		id, campaignID).Scan(&c.ID, &c.Name, &c.OwnerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Char{}, ErrNotFound
	}
	return c, err
}
