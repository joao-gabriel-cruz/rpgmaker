package refuge

import (
	"context"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
)

// edit runs fn over the refuge as the Mestre (gm true) or as any member.
func (s *Service) edit(ctx context.Context, m campaigns.Member, gm bool, fn func(*Refuge) error) (Refuge, error) {
	if gm && m.Role != campaigns.RoleGM {
		return Refuge{}, errGMOnly
	}
	return s.Mutate(ctx, m.CampaignID, fn)
}

// AddBuilding plans a building directly (Mestre).
func (s *Service) AddBuilding(ctx context.Context, m campaigns.Member, kind string) (Refuge, error) {
	return s.edit(ctx, m, true, func(r *Refuge) error { return r.plan(kind) })
}

// ProposeBuilding lets any member suggest a building; the Mestre decides.
func (s *Service) ProposeBuilding(ctx context.Context, m campaigns.Member, by, kind string) (Refuge, error) {
	return s.edit(ctx, m, false, func(r *Refuge) error { return r.propose(kind, m.UserID, by) })
}

// ResolveProposal approves (becoming a planned building) or rejects a proposal (Mestre).
func (s *Service) ResolveProposal(ctx context.Context, m campaigns.Member, id string, approve bool) (Refuge, error) {
	return s.edit(ctx, m, true, func(r *Refuge) error { return r.resolve(id, approve) })
}

// UpgradeBuilding starts the next level of a leveled building (Mestre).
func (s *Service) UpgradeBuilding(ctx context.Context, m campaigns.Member, id string) (Refuge, error) {
	return s.edit(ctx, m, true, func(r *Refuge) error { return r.upgrade(id) })
}

// CancelBuilding drops a not yet finished building (Mestre).
func (s *Service) CancelBuilding(ctx context.Context, m campaigns.Member, id string) (Refuge, error) {
	return s.edit(ctx, m, true, func(r *Refuge) error { return r.cancel(id) })
}

// Allocate distributes the available workers among buildings (Mestre).
func (s *Service) Allocate(ctx context.Context, m campaigns.Member, alloc map[string]int) (Refuge, error) {
	return s.edit(ctx, m, true, func(r *Refuge) error { return r.allocate(alloc) })
}
