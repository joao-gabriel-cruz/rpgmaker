package characters

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
)

// editMu serializes read-modify-write sheet edits made by other modules (inventory changes).
var editMu sync.Mutex

// EditSheet applies fn to the stored sheet and announces character.updated; it skips the creation-limit
// checks of Save so a full sheet issue elsewhere never blocks an inventory change.
func (s *Service) EditSheet(ctx context.Context, campaignID, id string, fn func(json.RawMessage) (json.RawMessage, error)) (View, error) {
	if !uuidRe.MatchString(id) {
		return View{}, ErrNotFound
	}
	editMu.Lock()
	defer editMu.Unlock()
	c, err := s.repo.Get(ctx, campaignID, id)
	if err != nil {
		return View{}, err
	}
	if c.Sheet, err = fn(c.Sheet); err != nil {
		return View{}, err
	}
	saved, err := s.repo.UpdateSheet(ctx, campaignID, id, c)
	if err != nil {
		return View{}, err
	}
	info, err := s.load(ctx, campaigns.Member{CampaignID: campaignID, Role: campaigns.RoleGM})
	if err != nil {
		return View{}, err
	}
	v := info.view(saved)
	s.notify(saved, "character.updated", v)
	return v, nil
}
