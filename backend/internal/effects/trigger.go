package effects

import (
	"context"
	"log/slog"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
)

// source is one thing affecting a cell; ident tells apart two objects of the same prefab.
type source struct{ ident, kind, key string }

func sourcesAt(sc scenes.Scene, x, z int) []source {
	var out []source
	if g, err := maps.Deserialize(sc.Grid); err == nil {
		if c, ok := g.At(x, z); ok {
			out = append(out, source{id(KindTerrain, c.Terrain), KindTerrain, c.Terrain})
		}
	}
	for _, o := range sc.Objects {
		for _, c := range maps.Footprint(o) {
			if c.X == x && c.Z == z {
				out = append(out, source{id(KindPrefab, o.Prefab) + "#" + o.ID, KindPrefab, o.Prefab})
				break
			}
		}
	}
	return out
}

// OnEntered handles a token that was placed (from == nil) or moved. It fires on_enter for every source
// of the destination cell that the origin cell does not share, so moving inside one lava pool is silent.
func (s *Service) OnEntered(ctx context.Context, campaignID, tokenID string, x, z int, from *maps.Coord) {
	tok, err := s.tokens.Get(ctx, campaignID, tokenID)
	if err != nil {
		slog.Error("effects: load token failed", "err", err)
		return
	}
	sc, err := s.scenes.Get(ctx, campaignID, tok.SceneID, campaigns.RoleGM)
	if err != nil {
		slog.Error("effects: load scene failed", "err", err)
		return
	}
	known := map[string]bool{}
	if from != nil {
		for _, src := range sourcesAt(sc, from.X, from.Z) {
			known[src.ident] = true
		}
	}
	var fresh []source
	for _, src := range sourcesAt(sc, x, z) {
		if !known[src.ident] {
			fresh = append(fresh, src)
		}
	}
	s.suggest(ctx, campaignID, tok, x, z, TriggerEnter, fresh)
}

// OnTurnStart fires on_turn_start for every token of the active scene standing on a cell with such a rule.
func (s *Service) OnTurnStart(ctx context.Context, campaignID string) {
	list, err := s.scenes.List(ctx, campaignID)
	if err != nil {
		slog.Error("effects: list scenes failed", "err", err)
		return
	}
	for _, sum := range list {
		if sum.IsActive {
			s.turnStartScene(ctx, campaignID, sum.ID)
		}
	}
}

func (s *Service) turnStartScene(ctx context.Context, campaignID, sceneID string) {
	sc, err := s.scenes.Get(ctx, campaignID, sceneID, campaigns.RoleGM)
	if err != nil {
		slog.Error("effects: load scene failed", "err", err)
		return
	}
	toks, err := s.tokens.List(ctx, campaignID, sceneID)
	if err != nil {
		slog.Error("effects: list tokens failed", "err", err)
		return
	}
	for _, t := range toks {
		s.suggest(ctx, campaignID, t, t.X, t.Z, TriggerTurnStart, sourcesAt(sc, t.X, t.Z))
	}
}

func (s *Service) suggest(ctx context.Context, campaignID string, tok tokens.Token, x, z int, trigger string, srcs []source) {
	if len(srcs) == 0 {
		return
	}
	over, err := s.store.Overrides(ctx, campaignID)
	if err != nil {
		slog.Error("effects: load rules failed", "err", err)
		return
	}
	for _, src := range srcs {
		for _, r := range effective(over, src.kind, src.key) {
			if r.Trigger == trigger {
				s.emit(ctx, Suggestion{CampaignID: campaignID, CharacterID: tok.CharacterID, CharacterName: tok.Name,
					SceneID: tok.SceneID, X: x, Z: z, Source: labelOf(src.kind, src.key), Trigger: trigger,
					Effect: r.Effect, Amount: r.Amount, Message: r.Message})
			}
		}
	}
}

func (s *Service) emit(ctx context.Context, in Suggestion) {
	sg, err := s.store.Insert(ctx, in)
	if err != nil {
		slog.Error("effects: save suggestion failed", "err", err)
		return
	}
	s.events.Broadcast(sg.CampaignID, realtime.ToGM, EventSuggested, map[string]Suggestion{"suggestion": sg})
}
