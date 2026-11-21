package effects

import (
	"context"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
)

// Rules lists the effective rules of every terrain type and prefab of the campaign.
func (s *Service) Rules(ctx context.Context, campaignID string) ([]RuleSet, error) {
	over, err := s.store.Overrides(ctx, campaignID)
	if err != nil {
		return nil, err
	}
	out := make([]RuleSet, 0, len(maps.TerrainOrder)+16)
	for _, t := range maps.TerrainOrder {
		out = append(out, ruleSet(over, KindTerrain, t.ID, t.BlocksMovement, t.BlocksVision))
	}
	for _, p := range maps.PrefabIDs() {
		r := maps.RuleFor(maps.Object{Prefab: p})
		out = append(out, ruleSet(over, KindPrefab, p, r.BlocksMovement, r.BlocksVision))
	}
	return out, nil
}

func ruleSet(over map[string][]Rule, kind, key string, bm, bv bool) RuleSet {
	_, custom := over[id(kind, key)]
	return RuleSet{Kind: kind, Key: key, Label: labelOf(kind, key), Rules: cloneRules(effective(over, kind, key)),
		IsDefault: !custom, BlocksMovement: bm, BlocksVision: bv}
}

// SetRules replaces the rules of one terrain or prefab for the campaign.
func (s *Service) SetRules(ctx context.Context, campaignID, kind, key string, rules []Rule) (RuleSet, error) {
	if !knownKey(kind, key) {
		return RuleSet{}, badInput("Terreno ou objeto desconhecido")
	}
	clean, err := normalize(rules)
	if err != nil {
		return RuleSet{}, err
	}
	if err := s.store.SaveRules(ctx, campaignID, kind, key, clean); err != nil {
		return RuleSet{}, err
	}
	return s.one(ctx, campaignID, kind, key)
}

// RestoreRules drops the campaign override so the default rules apply again.
func (s *Service) RestoreRules(ctx context.Context, campaignID, kind, key string) (RuleSet, error) {
	if !knownKey(kind, key) {
		return RuleSet{}, badInput("Terreno ou objeto desconhecido")
	}
	if err := s.store.DeleteRules(ctx, campaignID, kind, key); err != nil {
		return RuleSet{}, err
	}
	return s.one(ctx, campaignID, kind, key)
}

func (s *Service) one(ctx context.Context, campaignID, kind, key string) (RuleSet, error) {
	all, err := s.Rules(ctx, campaignID)
	if err != nil {
		return RuleSet{}, err
	}
	for _, r := range all {
		if r.Kind == kind && r.Key == key {
			return r, nil
		}
	}
	return RuleSet{}, ErrNotFound
}
