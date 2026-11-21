package mcp

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

var (
	// ErrInvalidLink means the secret is unknown, revoked or no longer GM-bound.
	ErrInvalidLink = errors.New("invalid mcp link")
	// ErrNoLink means the user has no active link in the campaign.
	ErrNoLink = errors.New("no active mcp link")
	// ErrForbidden means the caller is not the campaign GM.
	ErrForbidden = errors.New("only the gm can manage the mcp link")
)

// Session is the identity an MCP request acts as.
type Session struct {
	CampaignID string
	Sender     realtime.Sender
}

// Link is a stored (hash-only) Mestre link.
type Link struct {
	ID         string     `json:"id"`
	CampaignID string     `json:"campaignId"`
	UserID     string     `json:"userId"`
	CreatedAt  time.Time  `json:"createdAt"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
}

// LinkRepo persists links.
type LinkRepo interface {
	Replace(ctx context.Context, campaignID, userID, hash string) (Link, error)
	Active(ctx context.Context, campaignID, userID string) (Link, error)
	ByHash(ctx context.Context, hash string) (Link, error)
	Revoke(ctx context.Context, campaignID, userID string) error
	Touch(ctx context.Context, id string) error
}

// Members resolves campaign roles.
type Members interface {
	Membership(ctx context.Context, campaignID, userID string) (campaigns.Member, error)
}

// Links manages the secret link a Mestre gives to their Claude.
type Links struct {
	repo    LinkRepo
	members Members
}

// NewLinks builds the service.
func NewLinks(repo LinkRepo, m Members) *Links { return &Links{repo: repo, members: m} }

func hashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func requireGM(m campaigns.Member) error {
	if m.Role != campaigns.RoleGM {
		return ErrForbidden
	}
	return nil
}

// Generate creates a new secret, revoking any previous one. The secret is returned only here.
func (l *Links) Generate(ctx context.Context, m campaigns.Member) (string, Link, error) {
	if err := requireGM(m); err != nil {
		return "", Link{}, err
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", Link{}, err
	}
	secret := base64.RawURLEncoding.EncodeToString(buf)
	link, err := l.repo.Replace(ctx, m.CampaignID, m.UserID, hashSecret(secret))
	if err != nil {
		return "", Link{}, err
	}
	return secret, link, nil
}

// Status reports the active link, if any.
func (l *Links) Status(ctx context.Context, m campaigns.Member) (Link, bool, error) {
	if err := requireGM(m); err != nil {
		return Link{}, false, err
	}
	link, err := l.repo.Active(ctx, m.CampaignID, m.UserID)
	if errors.Is(err, ErrNoLink) {
		return Link{}, false, nil
	}
	return link, err == nil, err
}

// Revoke invalidates the active link.
func (l *Links) Revoke(ctx context.Context, m campaigns.Member) error {
	if err := requireGM(m); err != nil {
		return err
	}
	return l.repo.Revoke(ctx, m.CampaignID, m.UserID)
}

// Resolve maps a secret to a GM session; the owner must still be the campaign GM.
func (l *Links) Resolve(ctx context.Context, secret string) (Session, error) {
	link, err := l.repo.ByHash(ctx, hashSecret(secret))
	if err != nil {
		return Session{}, ErrInvalidLink
	}
	m, err := l.members.Membership(ctx, link.CampaignID, link.UserID)
	if err != nil || m.Role != campaigns.RoleGM {
		return Session{}, ErrInvalidLink
	}
	_ = l.repo.Touch(ctx, link.ID)
	return Session{CampaignID: link.CampaignID, Sender: realtime.Sender{
		UserID: link.UserID, Role: campaigns.RoleGM, Name: "Claude"}}, nil
}
