package campaigns

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store persists campaigns, members and invites in Postgres.
type Store struct{ pool *pgxpool.Pool }

// NewStore wraps a pgx pool.
func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

const campaignCols = "c.id, c.name, c.system_id, c.settings, c.owner_id, c.created_at"

func scanCampaign(row pgx.Row, extra ...any) (Campaign, error) {
	var c Campaign
	err := row.Scan(append([]any{&c.ID, &c.Name, &c.SystemID, &c.Settings, &c.OwnerID, &c.CreatedAt}, extra...)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return Campaign{}, ErrNotFound
	}
	return c, err
}

// CreateCampaign inserts a campaign and its owner as gm in one transaction.
func (s *Store) CreateCampaign(ctx context.Context, name, systemID, ownerID string) (Campaign, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Campaign{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	c, err := scanCampaign(tx.QueryRow(ctx,
		`INSERT INTO campaigns AS c (name, system_id, owner_id) VALUES ($1,$2,$3) RETURNING `+campaignCols,
		name, systemID, ownerID))
	if err != nil {
		return Campaign{}, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO campaign_members (campaign_id, user_id, role) VALUES ($1,$2,$3)`,
		c.ID, ownerID, RoleGM); err != nil {
		return Campaign{}, err
	}
	return c, tx.Commit(ctx)
}

// ListByUser returns the campaigns a user belongs to, with their role.
func (s *Store) ListByUser(ctx context.Context, userID string) ([]Summary, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+campaignCols+`, m.role FROM campaigns c
		JOIN campaign_members m ON m.campaign_id = c.id WHERE m.user_id = $1 ORDER BY c.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Summary{}
	for rows.Next() {
		var sum Summary
		sum.Campaign, err = scanCampaign(rows, &sum.Role)
		if err != nil {
			return nil, err
		}
		out = append(out, sum)
	}
	return out, rows.Err()
}

// Campaign loads a campaign by id.
func (s *Store) Campaign(ctx context.Context, id string) (Campaign, error) {
	return scanCampaign(s.pool.QueryRow(ctx, `SELECT `+campaignCols+` FROM campaigns c WHERE c.id = $1`, id))
}

// Member returns the user's membership, or ErrNotFound.
func (s *Store) Member(ctx context.Context, campaignID, userID string) (Member, error) {
	m := Member{CampaignID: campaignID, UserID: userID}
	err := s.pool.QueryRow(ctx, `SELECT role FROM campaign_members WHERE campaign_id = $1 AND user_id = $2`,
		campaignID, userID).Scan(&m.Role)
	if errors.Is(err, pgx.ErrNoRows) {
		return Member{}, ErrNotFound
	}
	return m, err
}

// Members lists everyone in a campaign, gm first.
func (s *Store) Members(ctx context.Context, campaignID string) ([]MemberInfo, error) {
	rows, err := s.pool.Query(ctx, `SELECT u.id, u.display_name, m.role FROM campaign_members m
		JOIN users u ON u.id = m.user_id WHERE m.campaign_id = $1
		ORDER BY (m.role = 'gm') DESC, m.joined_at`, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []MemberInfo{}
	for rows.Next() {
		var mi MemberInfo
		if err := rows.Scan(&mi.UserID, &mi.DisplayName, &mi.Role); err != nil {
			return nil, err
		}
		out = append(out, mi)
	}
	return out, rows.Err()
}

// RemoveMember deletes a membership; ErrNotFound if absent.
func (s *Store) RemoveMember(ctx context.Context, campaignID, userID string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM campaign_members WHERE campaign_id = $1 AND user_id = $2`,
		campaignID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// CreateInvite stores a new invite.
func (s *Store) CreateInvite(ctx context.Context, inv Invite) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO invites (token, campaign_id, role, expires_at) VALUES ($1,$2,$3,$4)`,
		inv.Token, inv.CampaignID, inv.Role, inv.ExpiresAt)
	return err
}

// InviteByToken loads an invite with its campaign name, or ErrNotFound.
func (s *Store) InviteByToken(ctx context.Context, token string) (Invite, error) {
	inv := Invite{Token: token}
	err := s.pool.QueryRow(ctx, `SELECT i.campaign_id, c.name, i.role, i.expires_at, i.used_by
		FROM invites i JOIN campaigns c ON c.id = i.campaign_id WHERE i.token = $1`, token).
		Scan(&inv.CampaignID, &inv.CampaignName, &inv.Role, &inv.ExpiresAt, &inv.UsedBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return Invite{}, ErrNotFound
	}
	return inv, err
}

// ConsumeInvite atomically marks a valid invite used and adds the user to the campaign.
// It returns ErrInvalidInvite if unusable and ErrAlreadyMember (invite kept) if already in.
func (s *Store) ConsumeInvite(ctx context.Context, token, userID string, now time.Time) (Member, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Member{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	m := Member{UserID: userID}
	err = tx.QueryRow(ctx, `UPDATE invites SET used_by = $2 WHERE token = $1 AND used_by IS NULL AND expires_at > $3
		RETURNING campaign_id, role`, token, userID, now).Scan(&m.CampaignID, &m.Role)
	if errors.Is(err, pgx.ErrNoRows) {
		return Member{}, ErrInvalidInvite
	}
	if err != nil {
		return Member{}, err
	}
	tag, err := tx.Exec(ctx, `INSERT INTO campaign_members (campaign_id, user_id, role) VALUES ($1,$2,$3)
		ON CONFLICT DO NOTHING`, m.CampaignID, userID, m.Role)
	if err != nil {
		return Member{}, err
	}
	if tag.RowsAffected() == 0 {
		return Member{}, ErrAlreadyMember
	}
	return m, tx.Commit(ctx)
}
