package actions

import (
	"context"
	"encoding/json"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
)

// spot is the active scene with the actor's token and everyone else's.
type spot struct {
	sc    scenes.Scene
	actor tokens.Token
	toks  []tokens.Token
}

// locate finds the active scene and the character's token on it.
func (s *Service) locate(ctx context.Context, campaignID, charID string) (spot, error) {
	list, err := s.scenes.List(ctx, campaignID)
	if err != nil {
		return spot{}, err
	}
	for _, sum := range list {
		if !sum.IsActive {
			continue
		}
		sc, err := s.scenes.Get(ctx, campaignID, sum.ID, campaigns.RoleGM)
		if err != nil {
			return spot{}, err
		}
		toks, err := s.tokens.List(ctx, campaignID, sum.ID)
		if err != nil {
			return spot{}, err
		}
		for _, t := range toks {
			if t.CharacterID == charID {
				return spot{sc: sc, actor: t, toks: toks}, nil
			}
		}
	}
	return spot{}, deny("Coloque o personagem na cena ativa para agir.")
}

func dist(ax, az, bx, bz int) int { return max(abs(ax-bx), abs(az-bz)) }

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// check validates the requested target against the scene and the action's range; it returns the
// normalized target (positions filled in) and the targeted token, if any.
func (p spot) check(act Action, in Target) (Target, *tokens.Token, error) {
	tgt, victim, err := p.resolve(act, in)
	if err != nil || tgt.Kind == TargetNone || act.Range == 0 {
		return tgt, victim, err
	}
	return tgt, victim, p.reach(act, *tgt.X, *tgt.Z)
}

func (p spot) resolve(act Action, in Target) (Target, *tokens.Token, error) {
	switch act.Target {
	case TargetToken:
		for i, t := range p.toks {
			if t.ID == in.TokenID {
				return Target{Kind: TargetToken, TokenID: t.ID, X: &p.toks[i].X, Z: &p.toks[i].Z}, &p.toks[i], nil
			}
		}
		return Target{}, nil, deny("Escolha um alvo válido na cena.")
	case TargetCell:
		return p.cell(in)
	case TargetObject:
		if in.Kind == TargetCell {
			return p.cell(in)
		}
		for _, o := range p.sc.Objects {
			if o.ID == in.ObjectID && o.ID != "" {
				return Target{Kind: TargetObject, ObjectID: o.ID, X: &o.X, Z: &o.Z}, nil, nil
			}
		}
		return Target{}, nil, deny("Escolha um objeto ou uma célula da cena.")
	}
	return Target{Kind: TargetNone}, nil, nil
}

func (p spot) cell(in Target) (Target, *tokens.Token, error) {
	if in.X == nil || in.Z == nil || *in.X < 0 || *in.Z < 0 || *in.X >= p.sc.Width || *in.Z >= p.sc.Depth {
		return Target{}, nil, deny("Escolha uma célula dentro do mapa.")
	}
	x, z := *in.X, *in.Z
	return Target{Kind: TargetCell, X: &x, Z: &z}, nil, nil
}

// reach enforces the range (Chebyshev, in cells) and, for longer ranges, a simple line of sight.
func (p spot) reach(act Action, x, z int) error {
	if dist(p.actor.X, p.actor.Z, x, z) > act.Range {
		return deny("Alvo fora de alcance.")
	}
	if act.Range <= 2 {
		return nil
	}
	grid, err := maps.Deserialize(p.sc.Grid)
	if err != nil {
		return nil
	}
	seen := maps.Visible(grid, p.sc.Objects, maps.Coord{X: p.actor.X, Z: p.actor.Z}, eyeHeight, act.Range)
	if _, ok := seen[maps.Coord{X: x, Z: z}]; !ok {
		return deny("Sem linha de visão para o alvo.")
	}
	return nil
}

// safe tells whether the actor rests somewhere safe: the scene is the Refúgio or a campfire is near.
func (p spot) safe() bool {
	var st struct {
		Refuge bool `json:"refuge"`
	}
	if len(p.sc.Settings) > 0 && json.Unmarshal(p.sc.Settings, &st) == nil && st.Refuge {
		return true
	}
	for _, o := range p.sc.Objects {
		if o.Prefab == "campfire" && dist(p.actor.X, p.actor.Z, o.X, o.Z) <= safeRadius {
			return true
		}
	}
	return false
}
