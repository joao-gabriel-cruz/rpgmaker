package items

import (
	"context"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

// Target says where a search looked: some cells and/or one object.
type Target struct {
	Cells    []maps.Coord
	ObjectID string
}

func (t Target) matches(l Loot) bool {
	if t.ObjectID != "" && l.ObjectID != nil && *l.ObjectID == t.ObjectID {
		return true
	}
	for _, c := range t.Cells {
		if c.X == l.X && c.Z == l.Z {
			return true
		}
	}
	return false
}

// TryFind reveals the hidden loot of the scene inside target whose difficulty is <= successes (called by a
// Perceber/Vasculhar action). Each find goes to the finder's backpack; what does not fit stays on the floor,
// visible. Emits loot.found to everyone for each loot.
func (s *Service) TryFind(ctx context.Context, campaignID, characterID, sceneID string, target Target, successes int) ([]Found, error) {
	view, err := s.chars.Get(ctx, gmMember(campaignID), characterID)
	if err != nil {
		return nil, mapChar(err)
	}
	all, err := s.repo.ListLoot(ctx, campaignID, sceneID)
	if err != nil {
		return nil, err
	}
	found := []Found{}
	for _, l := range all {
		if !l.Hidden || l.Difficulty > successes || !target.matches(l) {
			continue
		}
		f, err := s.reveal(ctx, campaignID, view.ID, view.Name, l)
		if err != nil {
			return found, err
		}
		found = append(found, f)
	}
	return found, nil
}

func (s *Service) reveal(ctx context.Context, campaignID, charID, charName string, l Loot) (Found, error) {
	n, err := s.deliver(ctx, campaignID, charID, l, []string{PlaceBackpack})
	if err != nil {
		return Found{}, err
	}
	rest, err := s.settle(ctx, campaignID, l, charID, n)
	if err != nil {
		return Found{}, err
	}
	f := Found{Loot: rest, Delivered: n, Left: rest.Qty}
	s.events.Broadcast(campaignID, realtime.ToAll, EventLootFound, map[string]any{
		"loot": rest, "characterId": charID, "characterName": charName, "delivered": n, "left": rest.Qty, "item": l.Item})
	return f, nil
}

// Pickup moves visible loot next to the character's token into its inventory (backpack, then body).
func (s *Service) Pickup(ctx context.Context, m campaigns.Member, lootID, charID string) error {
	l, err := s.loot(ctx, m.CampaignID, lootID)
	if err != nil {
		return err
	}
	if l.Hidden && m.Role != campaigns.RoleGM {
		return errNoLoot
	}
	charID, err = s.authorizePickup(ctx, m, l, charID)
	if err != nil {
		return err
	}
	n, err := s.deliver(ctx, m.CampaignID, charID, l, []string{PlaceBackpack, PlaceBody})
	if err != nil {
		return err
	}
	if n == 0 {
		return realtime.UserError("Seu inventário está cheio")
	}
	rest, err := s.settle(ctx, m.CampaignID, l, charID, n)
	if err == nil && rest.Qty > 0 {
		s.events.Broadcast(m.CampaignID, realtime.ToAll, EventLootPlaced, map[string]Loot{"loot": rest})
	}
	return err
}

// authorizePickup checks ownership and (for players) adjacency; it returns the validated character id.
func (s *Service) authorizePickup(ctx context.Context, m campaigns.Member, l Loot, charID string) (string, error) {
	view, err := s.chars.Get(ctx, m, charID)
	if err != nil {
		return "", mapChar(err)
	}
	if m.Role == campaigns.RoleGM {
		return view.ID, nil
	}
	if view.OwnerID == nil || *view.OwnerID != m.UserID {
		return "", realtime.UserError("Você só pode pegar itens com seus personagens")
	}
	return view.ID, s.checkAdjacent(ctx, m.CampaignID, l, view.ID)
}

func (s *Service) checkAdjacent(ctx context.Context, campaignID string, l Loot, charID string) error {
	toks, err := s.tokens.List(ctx, gmMember(campaignID), l.SceneID)
	if err != nil {
		return err
	}
	for _, t := range toks {
		if t.CharacterID != charID {
			continue
		}
		if abs(t.X-l.X) <= 1 && abs(t.Z-l.Z) <= 1 {
			return nil
		}
		return realtime.UserError("Chegue mais perto do item")
	}
	return realtime.UserError("Seu personagem não está nesta cena")
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
