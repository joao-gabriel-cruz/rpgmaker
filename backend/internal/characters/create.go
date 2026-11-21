package characters

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems"
)

// CreateInput is a request to create a character; OwnerID, Name and Kind are optional (Kind defaults to pc).
type CreateInput struct {
	OwnerID string
	Name    string
	Kind    string
}

// Create makes a character with a blank sheet. Players create their own PCs; the gm may pick any
// member, and only the gm creates NPCs (owned by nobody).
func (s *Service) Create(ctx context.Context, m campaigns.Member, in CreateInput) (View, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = defaultName
	}
	if len(name) > maxNameLen || (in.Kind != "" && in.Kind != KindPC && in.Kind != KindNPC) {
		return View{}, ErrInvalidInput
	}
	var owner *string
	if in.Kind == KindNPC {
		if m.Role != campaigns.RoleGM {
			return View{}, ErrForbidden
		}
	} else {
		o, err := s.resolveOwner(ctx, m, in.OwnerID)
		if err != nil {
			return View{}, err
		}
		owner = &o
	}
	return s.insert(ctx, m, Character{CampaignID: m.CampaignID, OwnerID: owner, Kind: kindOrPC(in.Kind), Name: name})
}

func kindOrPC(k string) string {
	if k == "" {
		return KindPC
	}
	return k
}

func (s *Service) insert(ctx context.Context, m campaigns.Member, c Character) (View, error) {
	info, err := s.load(ctx, m)
	if err != nil {
		return View{}, err
	}
	if c.Sheet, err = info.blankSheet(c.Kind, c.Name); err != nil {
		return View{}, err
	}
	if c, err = s.repo.Create(ctx, c); err != nil {
		return View{}, err
	}
	v := info.view(c)
	s.notify(c, "character.created", v)
	return v, nil
}

func (i ctxInfo) blankSheet(kind, name string) (json.RawMessage, error) {
	if kind == KindNPC {
		n, ok := i.system.(systems.NPCSheeter)
		if !ok {
			return nil, ErrNoSheetSupport
		}
		return n.NewNPCSheet(name)
	}
	t, ok := i.system.(systems.SheetTemplater)
	if !ok {
		return nil, ErrNoSheetSupport
	}
	return t.NewSheet(name)
}

func (s *Service) resolveOwner(ctx context.Context, m campaigns.Member, requested string) (string, error) {
	if requested == "" || requested == m.UserID {
		return m.UserID, nil
	}
	if m.Role != campaigns.RoleGM {
		return "", ErrForbidden
	}
	if _, err := s.campaigns.Membership(ctx, m.CampaignID, requested); err != nil {
		return "", ErrInvalidInput
	}
	return requested, nil
}
