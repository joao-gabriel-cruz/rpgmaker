package items

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/effects"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/refuge"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
)

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Characters is the part of characters.Service used here.
type Characters interface {
	Get(ctx context.Context, m campaigns.Member, id string) (characters.View, error)
	EditSheet(ctx context.Context, campaignID, id string, fn func(json.RawMessage) (json.RawMessage, error)) (characters.View, error)
}

// Refuge is the part of refuge.Service used here.
type Refuge interface {
	AddItem(ctx context.Context, campaignID string, item json.RawMessage) (refuge.Refuge, error)
	AddResource(ctx context.Context, campaignID, resource string, amount int) (refuge.Refuge, error)
}

// Effects is the part of effects.Service used to suggest healing to the Mestre.
type Effects interface {
	Suggest(ctx context.Context, in effects.Suggestion) (effects.Suggestion, error)
}

// Tokens is the part of the tokens service used to find a character on a scene.
type Tokens interface {
	List(ctx context.Context, m campaigns.Member, sceneID string) ([]tokens.Token, error)
}

// Scenes is the part of scenes.Service used here.
type Scenes interface {
	Get(ctx context.Context, campaignID, id, role string) (scenes.Scene, error)
}

// Events is the part of realtime.Hub used to notify clients.
type Events interface {
	Broadcast(campaignID string, to realtime.Filter, typ string, payload any)
}

// Service holds the compendium, delivery and loot rules.
type Service struct {
	repo    Repository
	chars   Characters
	refuge  Refuge
	effects Effects
	tokens  Tokens
	scenes  Scenes
	events  Events
}

// NewService builds a Service.
func NewService(repo Repository, ch Characters, rf Refuge, ef Effects, tk Tokens, sc Scenes, ev Events) *Service {
	return &Service{repo: repo, chars: ch, refuge: rf, effects: ef, tokens: tk, scenes: sc, events: ev}
}

func gmMember(campaignID string) campaigns.Member {
	return campaigns.Member{CampaignID: campaignID, Role: campaigns.RoleGM}
}

func (s *Service) changed(ctx context.Context, campaignID string) {
	list, err := s.repo.ListItems(ctx, campaignID)
	if err == nil {
		s.events.Broadcast(campaignID, realtime.ToGM, EventItemsChanged, map[string][]Item{"items": list})
	}
}

// List returns the compendium.
func (s *Service) List(ctx context.Context, campaignID string) ([]Item, error) {
	return s.repo.ListItems(ctx, campaignID)
}

// Create validates and stores a new item.
func (s *Service) Create(ctx context.Context, campaignID string, in Item) (Item, error) {
	it, err := Clean(in)
	if err != nil {
		return Item{}, err
	}
	out, err := s.repo.CreateItem(ctx, campaignID, it)
	if err == nil {
		s.changed(ctx, campaignID)
	}
	return out, err
}

// Update validates and replaces an item.
func (s *Service) Update(ctx context.Context, campaignID, id string, in Item) (Item, error) {
	if !uuidRe.MatchString(id) {
		return Item{}, ErrNotFound
	}
	it, err := Clean(in)
	if err != nil {
		return Item{}, err
	}
	it.ID = id
	out, err := s.repo.UpdateItem(ctx, campaignID, it)
	if err == nil {
		s.changed(ctx, campaignID)
	}
	return out, err
}

// Delete removes an item from the compendium.
func (s *Service) Delete(ctx context.Context, campaignID, id string) error {
	if !uuidRe.MatchString(id) {
		return ErrNotFound
	}
	if err := s.repo.DeleteItem(ctx, campaignID, id); err != nil {
		return err
	}
	s.changed(ctx, campaignID)
	return nil
}

// GiveInput is the item.give command payload.
type GiveInput struct {
	ItemID      string `json:"itemId"`
	CharacterID string `json:"characterId"`
	Target      string `json:"target"`
	Place       string `json:"place"`
	Qty         int    `json:"qty"`
}

func mapChar(err error) error {
	if errors.Is(err, characters.ErrNotFound) {
		return errNoChar
	}
	return err
}

func (s *Service) compendiumItem(ctx context.Context, campaignID, id string) (Item, error) {
	if !uuidRe.MatchString(id) {
		return Item{}, errNoItem
	}
	it, err := s.repo.GetItem(ctx, campaignID, id)
	if errors.Is(err, ErrNotFound) {
		return Item{}, errNoItem
	}
	return it, err
}

// Give copies a compendium item to a character's body/backpack or to the Refúgio, all or nothing.
func (s *Service) Give(ctx context.Context, campaignID string, in GiveInput) error {
	it, err := s.compendiumItem(ctx, campaignID, in.ItemID)
	if err != nil {
		return err
	}
	if in.Qty == 0 {
		in.Qty = 1
	}
	if in.Qty < 1 || in.Qty > BackpackSlots {
		return realtime.UserError("Quantidade inválida")
	}
	if in.Target == TargetRefuge {
		return s.giveRefuge(ctx, campaignID, SheetItem(it), in.Qty)
	}
	return s.giveCharacter(ctx, campaignID, SheetItem(it), in)
}

func (s *Service) giveRefuge(ctx context.Context, campaignID string, raw json.RawMessage, qty int) error {
	for i := 0; i < qty; i++ {
		if _, err := s.refuge.AddItem(ctx, campaignID, raw); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) giveCharacter(ctx context.Context, campaignID string, raw json.RawMessage, in GiveInput) error {
	_, err := s.chars.EditSheet(ctx, campaignID, in.CharacterID, func(sheet json.RawMessage) (json.RawMessage, error) {
		out, n, err := addCopies(sheet, raw, in.Qty, []string{in.Place})
		if err != nil {
			return nil, err
		}
		if n < in.Qty {
			return nil, fullError(in.Place)
		}
		return out, nil
	})
	return mapChar(err)
}
