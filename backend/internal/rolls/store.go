package rolls

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store persists rolls in Postgres.
type Store struct{ pool *pgxpool.Pool }

// NewStore wraps a pgx pool.
func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

const selectCols = `SELECT r.id, r.campaign_id, r.user_id, u.display_name, r.character_id, COALESCE(c.name, ''),
	r.visibility, r.kind, r.spec, r.dice, r.result, r.status, r.seed, r.created_at
	FROM rolls r JOIN users u ON u.id = r.user_id LEFT JOIN characters c ON c.id = r.character_id`

func scan(row pgx.Row) (Record, error) {
	var r Record
	err := row.Scan(&r.ID, &r.CampaignID, &r.UserID, &r.UserName, &r.CharacterID, &r.CharacterName,
		&r.Visibility, &r.Kind, &r.Spec, &r.Dice, &r.Result, &r.Status, &r.Seed, &r.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Record{}, ErrNotFound
	}
	return r, err
}

// Insert stores a roll and returns it with names resolved.
func (s *Store) Insert(ctx context.Context, r Record) (Record, error) {
	var id string
	err := s.pool.QueryRow(ctx, `INSERT INTO rolls (campaign_id, user_id, character_id, visibility, kind, spec, dice, result, status, seed)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`,
		r.CampaignID, r.UserID, r.CharacterID, r.Visibility, r.Kind, []byte(r.Spec), []byte(r.Dice), nullable(r.Result), r.Status, r.Seed).Scan(&id)
	if err != nil {
		return Record{}, err
	}
	return s.Get(ctx, r.CampaignID, id)
}

func nullable(b json.RawMessage) any {
	if len(b) == 0 {
		return nil
	}
	return []byte(b)
}

// Get loads a roll of the campaign, or ErrNotFound.
func (s *Store) Get(ctx context.Context, campaignID, id string) (Record, error) {
	if !uuidRe.MatchString(id) {
		return Record{}, ErrNotFound
	}
	return scan(s.pool.QueryRow(ctx, selectCols+` WHERE r.campaign_id = $1 AND r.id = $2`, campaignID, id))
}

// Finalize stores the result of a pending roll; ErrAlreadyFinal when it was not pending.
func (s *Store) Finalize(ctx context.Context, campaignID, id string, result json.RawMessage) error {
	tag, err := s.pool.Exec(ctx, `UPDATE rolls SET result = $3, status = 'final'
		WHERE campaign_id = $1 AND id = $2 AND status = 'pending'`, campaignID, id, []byte(result))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrAlreadyFinal
	}
	return nil
}

// Recent returns the newest rolls first; gm-only rolls are included only on request.
func (s *Store) Recent(ctx context.Context, campaignID string, limit int, includeGM bool) ([]Record, error) {
	rows, err := s.pool.Query(ctx, selectCols+` WHERE r.campaign_id = $1 AND ($3 OR r.visibility = 'public')
		ORDER BY r.created_at DESC, r.id LIMIT $2`, campaignID, limit, includeGM)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Record{}
	for rows.Next() {
		r, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
