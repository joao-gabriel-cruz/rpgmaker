package campaigns

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/auth"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems"
)

const maxNameLen = 100

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Repository is the persistence the Service needs.
type Repository interface {
	CreateCampaign(ctx context.Context, name, systemID, ownerID string) (Campaign, error)
	ListByUser(ctx context.Context, userID string) ([]Summary, error)
	Campaign(ctx context.Context, id string) (Campaign, error)
	MergeSettings(ctx context.Context, id string, patch json.RawMessage) (Campaign, error)
	Member(ctx context.Context, campaignID, userID string) (Member, error)
	Members(ctx context.Context, campaignID string) ([]MemberInfo, error)
	RemoveMember(ctx context.Context, campaignID, userID string) error
	CreateInvite(ctx context.Context, inv Invite) error
	InviteByToken(ctx context.Context, token string) (Invite, error)
	ConsumeInvite(ctx context.Context, token, userID string, now time.Time) (Member, error)
}

// Accounts is the part of auth.Service used to register users from an invite.
type Accounts interface {
	CreateUser(ctx context.Context, email, name, password string, admin bool) (auth.User, error)
	Login(ctx context.Context, email, password string) (auth.User, string, time.Time, error)
}

// Service holds the campaign, membership and invite rules.
type Service struct {
	repo     Repository
	accounts Accounts
	systems  *systems.Registry
	now      func() time.Time
}

// NewService builds a Service.
func NewService(repo Repository, accounts Accounts, reg *systems.Registry) *Service {
	return &Service{repo: repo, accounts: accounts, systems: reg, now: time.Now}
}

// Create makes a campaign for a registered system; the creator becomes its gm.
func (s *Service) Create(ctx context.Context, ownerID, name, systemID string) (Campaign, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > maxNameLen {
		return Campaign{}, ErrInvalidInput
	}
	if _, ok := s.systems.Get(systemID); !ok {
		return Campaign{}, ErrUnknownSystem
	}
	return s.repo.CreateCampaign(ctx, name, systemID, ownerID)
}

// List returns the user's campaigns with their role.
func (s *Service) List(ctx context.Context, userID string) ([]Summary, error) {
	list, err := s.repo.ListByUser(ctx, userID)
	for i := range list {
		list[i].SystemName = s.systemName(list[i].SystemID)
	}
	return list, err
}

// Detail returns a campaign as seen by a member.
func (s *Service) Detail(ctx context.Context, m Member) (Summary, error) {
	c, err := s.repo.Campaign(ctx, m.CampaignID)
	if err != nil {
		return Summary{}, err
	}
	return Summary{Campaign: c, Role: m.Role, SystemName: s.systemName(c.SystemID)}, nil
}

// Creation-limit modes stored in settings.creationLimits.
const (
	LimitsWarn  = "warn"
	LimitsBlock = "block"
)

// UpdateSettings lets the gm change the campaign settings; today only creationLimits.
func (s *Service) UpdateSettings(ctx context.Context, m Member, creationLimits string) (Summary, error) {
	if creationLimits != LimitsWarn && creationLimits != LimitsBlock {
		return Summary{}, ErrInvalidInput
	}
	patch, err := json.Marshal(map[string]string{"creationLimits": creationLimits})
	if err != nil {
		return Summary{}, err
	}
	c, err := s.repo.MergeSettings(ctx, m.CampaignID, patch)
	if err != nil {
		return Summary{}, err
	}
	return Summary{Campaign: c, Role: m.Role, SystemName: s.systemName(c.SystemID)}, nil
}

func (s *Service) systemName(id string) string {
	if sys, ok := s.systems.Get(id); ok {
		return sys.Name()
	}
	return id
}

// Membership resolves the user's role in a campaign; non-members and bad ids yield ErrNotFound.
func (s *Service) Membership(ctx context.Context, campaignID, userID string) (Member, error) {
	if !uuidRe.MatchString(campaignID) {
		return Member{}, ErrNotFound
	}
	return s.repo.Member(ctx, campaignID, userID)
}

// Members lists the campaign's members.
func (s *Service) Members(ctx context.Context, campaignID string) ([]MemberInfo, error) {
	return s.repo.Members(ctx, campaignID)
}

// RemoveMember lets a gm remove a player, never themselves nor another gm.
func (s *Service) RemoveMember(ctx context.Context, gm Member, targetID string) error {
	if targetID == gm.UserID {
		return ErrCannotRemove
	}
	if !uuidRe.MatchString(targetID) {
		return ErrNotFound
	}
	target, err := s.repo.Member(ctx, gm.CampaignID, targetID)
	if err != nil {
		return err
	}
	if target.Role != RolePlayer {
		return ErrCannotRemove
	}
	return s.repo.RemoveMember(ctx, gm.CampaignID, targetID)
}

// CreateInvite issues a single-use player invite valid for InviteTTL.
func (s *Service) CreateInvite(ctx context.Context, campaignID string) (Invite, error) {
	token, err := randomToken()
	if err != nil {
		return Invite{}, err
	}
	inv := Invite{Token: token, CampaignID: campaignID, Role: RolePlayer, ExpiresAt: s.now().Add(InviteTTL)}
	return inv, s.repo.CreateInvite(ctx, inv)
}

// Invite returns an invite only while it is still usable.
func (s *Service) Invite(ctx context.Context, token string) (Invite, error) {
	inv, err := s.repo.InviteByToken(ctx, token)
	if errors.Is(err, ErrNotFound) {
		return Invite{}, ErrInvalidInvite
	}
	if err != nil {
		return Invite{}, err
	}
	if inv.UsedBy != nil || !inv.ExpiresAt.After(s.now()) {
		return Invite{}, ErrInvalidInvite
	}
	return inv, nil
}

// SignUp registers a user through a valid invite and opens their session.
func (s *Service) SignUp(ctx context.Context, token, email, name, password string) (auth.User, string, time.Time, error) {
	none := func(err error) (auth.User, string, time.Time, error) { return auth.User{}, "", time.Time{}, err }
	if _, err := s.Invite(ctx, token); err != nil {
		return none(err)
	}
	if strings.TrimSpace(email) == "" || strings.TrimSpace(name) == "" || len(password) < 8 {
		return none(ErrInvalidInput)
	}
	u, err := s.accounts.CreateUser(ctx, email, name, password, false)
	if err != nil {
		return none(err)
	}
	if _, err := s.repo.ConsumeInvite(ctx, token, u.ID, s.now()); err != nil {
		return none(err)
	}
	user, session, expires, err := s.accounts.Login(ctx, email, password)
	if err != nil {
		return none(fmt.Errorf("login after signup: %w", err))
	}
	return user, session, expires, nil
}

// Accept lets a logged-in user join through a valid invite.
func (s *Service) Accept(ctx context.Context, token, userID string) (Member, error) {
	if _, err := s.Invite(ctx, token); err != nil {
		return Member{}, err
	}
	return s.repo.ConsumeInvite(ctx, token, userID, s.now())
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate invite token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
