package characters

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems"
)

const (
	defaultName = "Novo personagem"
	unnamed     = "Sem nome"
	maxNameLen  = 100
)

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Repository is the persistence the Service needs.
type Repository interface {
	Create(ctx context.Context, c Character) (Character, error)
	Get(ctx context.Context, campaignID, id string) (Character, error)
	List(ctx context.Context, campaignID, kind string) ([]Character, error)
	UpdateSheet(ctx context.Context, campaignID, id string, c Character) (Character, error)
	Delete(ctx context.Context, campaignID, id string) error
}

// Campaigns is the part of campaigns.Service used here.
type Campaigns interface {
	Detail(ctx context.Context, m campaigns.Member) (campaigns.Summary, error)
	Membership(ctx context.Context, campaignID, userID string) (campaigns.Member, error)
}

// Events is the part of realtime.Hub used to notify clients.
type Events interface {
	Broadcast(campaignID string, to realtime.Filter, typ string, payload any)
}

// Service holds the character rules: permissions, validation policy and notifications.
type Service struct {
	repo      Repository
	campaigns Campaigns
	systems   *systems.Registry
	events    Events
}

// NewService builds a Service.
func NewService(repo Repository, c Campaigns, reg *systems.Registry, ev Events) *Service {
	return &Service{repo: repo, campaigns: c, systems: reg, events: ev}
}

// ctxInfo is the campaign context a request runs in.
type ctxInfo struct {
	system   systems.System
	settings json.RawMessage
}

func (s *Service) load(ctx context.Context, m campaigns.Member) (ctxInfo, error) {
	d, err := s.campaigns.Detail(ctx, m)
	if err != nil {
		return ctxInfo{}, err
	}
	sys, ok := s.systems.Get(d.SystemID)
	if !ok {
		return ctxInfo{}, ErrNoSheetSupport
	}
	return ctxInfo{system: sys, settings: d.Settings}, nil
}

func canEdit(m campaigns.Member, c Character) bool {
	return m.Role == campaigns.RoleGM || (c.OwnerID != nil && *c.OwnerID == m.UserID)
}

func visibleTo(m campaigns.Member, c Character) bool {
	return m.Role == campaigns.RoleGM || c.Kind == KindPC
}

func (i ctxInfo) view(c Character) View {
	v := View{Character: c, Issues: []systems.Issue{}}
	if n, ok := i.system.(systems.NPCSheeter); ok && c.Kind == KindNPC {
		v.Derived, _ = n.DeriveNPCSheet(c.Sheet)
		v.Issues = n.ValidateNPCSheet(c.Sheet)
		return v
	}
	if d, ok := i.system.(systems.SheetDeriver); ok {
		if out, err := d.DeriveSheet(c.Sheet); err == nil {
			v.Derived = out
		}
	}
	if val, ok := i.system.(systems.SheetValidator); ok {
		v.Issues = val.ValidateSheet(c.Sheet, i.settings)
	}
	return v
}

// List returns the campaign's characters; players only see PCs.
func (s *Service) List(ctx context.Context, m campaigns.Member) ([]View, error) {
	info, err := s.load(ctx, m)
	if err != nil {
		return nil, err
	}
	kind := KindPC
	if m.Role == campaigns.RoleGM {
		kind = ""
	}
	list, err := s.repo.List(ctx, m.CampaignID, kind)
	if err != nil {
		return nil, err
	}
	out := make([]View, len(list))
	for i, c := range list {
		out[i] = info.view(c)
	}
	return out, nil
}

// Get returns one character; NPCs are hidden from players.
func (s *Service) Get(ctx context.Context, m campaigns.Member, id string) (View, error) {
	info, c, err := s.fetch(ctx, m, id)
	if err != nil {
		return View{}, err
	}
	return info.view(c), nil
}

func (s *Service) fetch(ctx context.Context, m campaigns.Member, id string) (ctxInfo, Character, error) {
	if !uuidRe.MatchString(id) {
		return ctxInfo{}, Character{}, ErrNotFound
	}
	info, err := s.load(ctx, m)
	if err != nil {
		return ctxInfo{}, Character{}, err
	}
	c, err := s.repo.Get(ctx, m.CampaignID, id)
	if err != nil {
		return ctxInfo{}, Character{}, err
	}
	if !visibleTo(m, c) {
		return ctxInfo{}, Character{}, ErrNotFound
	}
	return info, c, nil
}

// Save replaces the whole sheet. Blocking issues always reject; limit issues only in "block" mode.
func (s *Service) Save(ctx context.Context, m campaigns.Member, id string, sheet json.RawMessage) (View, error) {
	info, c, err := s.fetch(ctx, m, id)
	if err != nil {
		return View{}, err
	}
	if !canEdit(m, c) {
		return View{}, ErrForbidden
	}
	c.Sheet = sheet
	c.Name, c.PortraitURL = identityOf(sheet)
	if v := info.view(c); rejects(v.Issues, info.settings) {
		return View{}, &RejectedError{Issues: v.Issues}
	}
	saved, err := s.repo.UpdateSheet(ctx, m.CampaignID, id, c)
	if err != nil {
		return View{}, err
	}
	v := info.view(saved)
	s.notify(saved, "character.updated", v)
	return v, nil
}

// Announce broadcasts character.updated for a character changed outside this service (e.g. by the refuge).
func (s *Service) Announce(ctx context.Context, campaignID, id string) error {
	c, err := s.repo.Get(ctx, campaignID, id)
	if err != nil {
		return err
	}
	info, err := s.load(ctx, campaigns.Member{CampaignID: campaignID, Role: campaigns.RoleGM})
	if err != nil {
		return err
	}
	s.notify(c, "character.updated", info.view(c))
	return nil
}

// Delete removes a character (owner or gm).
func (s *Service) Delete(ctx context.Context, m campaigns.Member, id string) error {
	_, c, err := s.fetch(ctx, m, id)
	if err != nil {
		return err
	}
	if !canEdit(m, c) {
		return ErrForbidden
	}
	if err := s.repo.Delete(ctx, m.CampaignID, id); err != nil {
		return err
	}
	s.notify(c, "character.deleted", map[string]string{"id": id})
	return nil
}

func (s *Service) notify(c Character, typ string, payload any) {
	to := realtime.Filter(realtime.ToAll)
	if c.Kind != KindPC {
		to = realtime.ToGM
	}
	s.events.Broadcast(c.CampaignID, to, typ, payload)
}

func rejects(issues []systems.Issue, settings json.RawMessage) bool {
	var cfg struct {
		CreationLimits string `json:"creationLimits"`
	}
	_ = json.Unmarshal(settings, &cfg)
	for _, i := range issues {
		if i.Severity == systems.SeverityBlocking || (i.Severity == systems.SeverityLimit && cfg.CreationLimits == campaigns.LimitsBlock) {
			return true
		}
	}
	return false
}

// identityOf reads the denormalized name and portrait from the sheet's identity block.
func identityOf(sheet json.RawMessage) (string, *string) {
	var in struct {
		Identity struct {
			Name        string  `json:"name"`
			PortraitURL *string `json:"portraitUrl"`
		} `json:"identity"`
	}
	_ = json.Unmarshal(sheet, &in)
	name := strings.TrimSpace(in.Identity.Name)
	if name == "" {
		name = unnamed
	}
	if len(name) > maxNameLen {
		name = name[:maxNameLen]
	}
	if in.Identity.PortraitURL != nil && *in.Identity.PortraitURL == "" {
		return name, nil
	}
	return name, in.Identity.PortraitURL
}
