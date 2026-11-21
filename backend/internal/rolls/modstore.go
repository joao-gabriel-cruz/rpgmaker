package rolls

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// Usage reports which once-per-session modifiers the character already spent in the session.
func (s *Store) Usage(ctx context.Context, sessionID, characterID string) (Usage, error) {
	var u Usage
	err := s.pool.QueryRow(ctx, `SELECT
		COALESCE(bool_or((spec->'request'->>'useOccupation')::boolean), false),
		COALESCE(bool_or((spec->'request'->>'useMarkingEvent')::boolean), false)
		FROM rolls WHERE play_session_id = $1 AND character_id = $2 AND kind = 'assimilacao'`,
		sessionID, characterID).Scan(&u.OccupationUsed, &u.MarkingEventUsed)
	return u, err
}

// Difficulty returns the pending d12 count for the character's next roll (0 when none).
func (s *Store) Difficulty(ctx context.Context, campaignID, characterID string) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT d12 FROM roll_difficulties WHERE campaign_id = $1 AND character_id = $2`,
		campaignID, characterID).Scan(&n)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return n, err
}

// SetDifficulty stores the pending difficulty; 0 clears it.
func (s *Store) SetDifficulty(ctx context.Context, campaignID, characterID string, d12 int) error {
	if d12 == 0 {
		_, err := s.pool.Exec(ctx, `DELETE FROM roll_difficulties WHERE campaign_id = $1 AND character_id = $2`, campaignID, characterID)
		return err
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO roll_difficulties (campaign_id, character_id, d12) VALUES ($1,$2,$3)
		ON CONFLICT (campaign_id, character_id) DO UPDATE SET d12 = $3`, campaignID, characterID, d12)
	return err
}
