// Package campaigns owns campaigns, membership, invites and the per-campaign authorization helpers.
package campaigns

import (
	"encoding/json"
	"errors"
	"time"
)

// Roles a user can hold within a campaign.
const (
	RoleGM     = "gm"
	RolePlayer = "player"
)

// InviteTTL is how long an invite link stays valid.
const InviteTTL = 7 * 24 * time.Hour

var (
	// ErrNotFound means the campaign, member or invite does not exist.
	ErrNotFound = errors.New("not found")
	// ErrInvalidInvite covers unknown, expired and already used invites.
	ErrInvalidInvite = errors.New("invalid or expired invite")
	// ErrAlreadyMember means the user is already in the campaign.
	ErrAlreadyMember = errors.New("already a member")
	// ErrInvalidInput means a request field failed validation.
	ErrInvalidInput = errors.New("invalid input")
	// ErrUnknownSystem means the system_id is not registered.
	ErrUnknownSystem = errors.New("unknown system")
	// ErrCannotRemove means the target of a removal is not a removable player.
	ErrCannotRemove = errors.New("cannot remove this member")
)

// Campaign is a table of play.
type Campaign struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	SystemID  string          `json:"systemId"`
	Settings  json.RawMessage `json:"settings"`
	OwnerID   string          `json:"ownerId"`
	CreatedAt time.Time       `json:"createdAt"`
}

// Summary is a campaign with the caller's role in it.
type Summary struct {
	Campaign
	Role       string `json:"role"`
	SystemName string `json:"systemName"`
}

// Member is the authorization context of a user inside one campaign.
type Member struct {
	CampaignID string `json:"campaignId"`
	UserID     string `json:"userId"`
	Role       string `json:"role"`
}

// MemberInfo is a member as listed to the other members.
type MemberInfo struct {
	UserID      string `json:"userId"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
}

// Invite is a single-use link granting a role in a campaign.
type Invite struct {
	Token        string    `json:"token"`
	CampaignID   string    `json:"-"`
	CampaignName string    `json:"campaignName"`
	Role         string    `json:"-"`
	ExpiresAt    time.Time `json:"expiresAt"`
	UsedBy       *string   `json:"-"`
}
