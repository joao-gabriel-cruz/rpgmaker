package items

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
)

const (
	maxLootQty        = 20
	maxLootDifficulty = 20
)

var errNoLoot = realtime.UserError("Item do mapa não encontrado")

// PlaceInput is the loot.place command payload; Hidden defaults to true.
type PlaceInput struct {
	SceneID    string  `json:"sceneId"`
	X          int     `json:"x"`
	Z          int     `json:"z"`
	ObjectID   *string `json:"objectId"`
	ItemID     string  `json:"itemId"`
	Qty        int     `json:"qty"`
	Difficulty int     `json:"difficulty"`
	Hidden     *bool   `json:"hidden"`
}

func (s *Service) checkPlace(ctx context.Context, campaignID string, in PlaceInput) error {
	sc, err := s.scenes.Get(ctx, campaignID, in.SceneID, campaigns.RoleGM)
	if errors.Is(err, scenes.ErrNotFound) {
		return realtime.UserError("Cena não encontrada")
	}
	if err != nil {
		return err
	}
	if in.X < 0 || in.Z < 0 || in.X >= sc.Width || in.Z >= sc.Depth {
		return realtime.UserError("Célula fora do mapa")
	}
	if in.ObjectID != nil {
		for _, o := range sc.Objects {
			if o.ID == *in.ObjectID {
				return nil
			}
		}
		return realtime.UserError("Objeto não encontrado no mapa")
	}
	return nil
}

// PlaceLoot hides (or drops) a compendium item on a cell or inside an object.
func (s *Service) PlaceLoot(ctx context.Context, campaignID string, in PlaceInput) (Loot, error) {
	if in.Qty == 0 {
		in.Qty = 1
	}
	if in.Qty < 1 || in.Qty > maxLootQty || in.Difficulty < 0 || in.Difficulty > maxLootDifficulty {
		return Loot{}, realtime.UserError("Quantidade ou dificuldade inválida")
	}
	it, err := s.compendiumItem(ctx, campaignID, in.ItemID)
	if err != nil {
		return Loot{}, err
	}
	if err := s.checkPlace(ctx, campaignID, in); err != nil {
		return Loot{}, err
	}
	hidden := in.Hidden == nil || *in.Hidden
	l, err := s.repo.InsertLoot(ctx, Loot{SceneID: in.SceneID, X: in.X, Z: in.Z, ObjectID: in.ObjectID, ItemID: &it.ID,
		Item: SheetItem(it), Qty: in.Qty, Difficulty: in.Difficulty, Hidden: hidden})
	if err != nil {
		return Loot{}, err
	}
	s.events.Broadcast(campaignID, lootAudience(l), EventLootPlaced, map[string]Loot{"loot": l})
	return l, nil
}

func lootAudience(l Loot) realtime.Filter {
	if l.Hidden {
		return realtime.ToGM
	}
	return realtime.ToAll
}

// RemoveLoot deletes a loot; players only hear about it when they could see it.
func (s *Service) RemoveLoot(ctx context.Context, campaignID, id string) error {
	l, err := s.loot(ctx, campaignID, id)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteLoot(ctx, campaignID, id); err != nil {
		return err
	}
	s.events.Broadcast(campaignID, lootAudience(l), EventLootRemoved, map[string]string{"id": id, "sceneId": l.SceneID})
	return nil
}

func (s *Service) loot(ctx context.Context, campaignID, id string) (Loot, error) {
	if !uuidRe.MatchString(id) {
		return Loot{}, errNoLoot
	}
	l, err := s.repo.GetLoot(ctx, campaignID, id)
	if errors.Is(err, ErrNotFound) {
		return Loot{}, errNoLoot
	}
	return l, err
}

// ListLoot returns the scene's loot; players never get hidden loot.
func (s *Service) ListLoot(ctx context.Context, m campaigns.Member, sceneID string) ([]Loot, error) {
	if !uuidRe.MatchString(sceneID) {
		return []Loot{}, nil
	}
	all, err := s.repo.ListLoot(ctx, m.CampaignID, sceneID)
	if err != nil || m.Role == campaigns.RoleGM {
		return all, err
	}
	out := []Loot{}
	for _, l := range all {
		if !l.Hidden {
			out = append(out, l)
		}
	}
	return out, nil
}

// settle records how many copies reached the finder: removes the loot or leaves the rest visible on the floor.
func (s *Service) settle(ctx context.Context, campaignID string, l Loot, charID string, delivered int) (Loot, error) {
	l.Qty -= delivered
	if l.Qty <= 0 {
		l.Qty = 0
		if err := s.repo.DeleteLoot(ctx, campaignID, l.ID); err != nil {
			return l, err
		}
		s.events.Broadcast(campaignID, lootAudience(l), EventLootRemoved, map[string]string{"id": l.ID, "sceneId": l.SceneID})
		return l, nil
	}
	l.Hidden, l.FoundBy = false, &charID
	if err := s.repo.SaveLoot(ctx, l); err != nil {
		return l, err
	}
	return l, nil
}

func (s *Service) deliver(ctx context.Context, campaignID, charID string, l Loot, places []string) (int, error) {
	var n int
	_, err := s.chars.EditSheet(ctx, campaignID, charID, func(sheet json.RawMessage) (json.RawMessage, error) {
		out, added, err := addCopies(sheet, l.Item, l.Qty, places)
		n = added
		return out, err
	})
	return n, mapChar(err)
}
