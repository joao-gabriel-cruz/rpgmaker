package effects

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store is the persistence port.
type Store interface {
	Overrides(ctx context.Context, campaignID string) (map[string][]Rule, error)
	SaveRules(ctx context.Context, campaignID, kind, key string, rules []Rule) error
	DeleteRules(ctx context.Context, campaignID, kind, key string) error
	Insert(ctx context.Context, s Suggestion) (Suggestion, error)
	Pending(ctx context.Context, campaignID string) ([]Suggestion, error)
	Resolve(ctx context.Context, campaignID, id, status string) (Suggestion, error)
	Reopen(ctx context.Context, campaignID, id string) error
}

// PgStore is the Postgres Store.
type PgStore struct{ pool *pgxpool.Pool }

// NewStore wraps a pgx pool.
func NewStore(pool *pgxpool.Pool) *PgStore { return &PgStore{pool: pool} }

// Overrides returns the campaign's customized rules keyed by "kind:key".
func (s *PgStore) Overrides(ctx context.Context, campaignID string) (map[string][]Rule, error) {
	rows, err := s.pool.Query(ctx, `SELECT kind, key, rules FROM effect_rules WHERE campaign_id = $1`, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string][]Rule{}
	for rows.Next() {
		var kind, key string
		var raw []byte
		if err := rows.Scan(&kind, &key, &raw); err != nil {
			return nil, err
		}
		var rules []Rule
		if err := json.Unmarshal(raw, &rules); err != nil {
			return nil, err
		}
		out[id(kind, key)] = rules
	}
	return out, rows.Err()
}

// SaveRules upserts the override of one terrain or prefab.
func (s *PgStore) SaveRules(ctx context.Context, campaignID, kind, key string, rules []Rule) error {
	raw, err := json.Marshal(rules)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `INSERT INTO effect_rules (campaign_id, kind, key, rules) VALUES ($1,$2,$3,$4)
		ON CONFLICT (campaign_id, kind, key) DO UPDATE SET rules = $4`, campaignID, kind, key, raw)
	return err
}

// DeleteRules removes the override so the default applies again.
func (s *PgStore) DeleteRules(ctx context.Context, campaignID, kind, key string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM effect_rules WHERE campaign_id = $1 AND kind = $2 AND key = $3`,
		campaignID, kind, key)
	return err
}

const suggestionCols = `id, campaign_id, character_id, character_name, COALESCE(scene_id::text, ''), x, z, source, trigger, effect, amount, message, created_at`

func scan(row pgx.Row) (Suggestion, error) {
	var s Suggestion
	err := row.Scan(&s.ID, &s.CampaignID, &s.CharacterID, &s.CharacterName, &s.SceneID, &s.X, &s.Z,
		&s.Source, &s.Trigger, &s.Effect, &s.Amount, &s.Message, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Suggestion{}, ErrNotFound
	}
	return s, err
}

// Insert stores a new pending suggestion.
func (s *PgStore) Insert(ctx context.Context, in Suggestion) (Suggestion, error) {
	return scan(s.pool.QueryRow(ctx, `INSERT INTO effect_suggestions
		(campaign_id, character_id, character_name, scene_id, x, z, source, trigger, effect, amount, message)
		VALUES ($1,$2,$3,NULLIF($4,'')::uuid,$5,$6,$7,$8,$9,$10,$11) RETURNING `+suggestionCols,
		in.CampaignID, in.CharacterID, in.CharacterName, in.SceneID, in.X, in.Z, in.Source, in.Trigger,
		in.Effect, in.Amount, in.Message))
}

// Pending lists the open suggestions, oldest first.
func (s *PgStore) Pending(ctx context.Context, campaignID string) ([]Suggestion, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+suggestionCols+` FROM effect_suggestions
		WHERE campaign_id = $1 AND status = 'pending' ORDER BY created_at, id`, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Suggestion{}
	for rows.Next() {
		sg, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sg)
	}
	return out, rows.Err()
}

// Resolve atomically claims a pending suggestion, so concurrent confirmations apply it once.
func (s *PgStore) Resolve(ctx context.Context, campaignID, id, status string) (Suggestion, error) {
	return scan(s.pool.QueryRow(ctx, `UPDATE effect_suggestions SET status = $3
		WHERE id = $1 AND campaign_id = $2 AND status = 'pending' RETURNING `+suggestionCols, id, campaignID, status))
}

// Reopen puts a claimed suggestion back to pending after a failed application.
func (s *PgStore) Reopen(ctx context.Context, campaignID, id string) error {
	_, err := s.pool.Exec(ctx, `UPDATE effect_suggestions SET status = 'pending' WHERE id = $1 AND campaign_id = $2`,
		id, campaignID)
	return err
}
