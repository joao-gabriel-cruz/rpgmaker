package characters

import (
	"context"
	"fmt"
	"sort"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

type fakeRepo struct {
	byID map[string]Character
}

func newFakeRepo() *fakeRepo { return &fakeRepo{byID: map[string]Character{}} }

func (f *fakeRepo) Create(_ context.Context, c Character) (Character, error) {
	c.ID = fmt.Sprintf("22222222-2222-2222-2222-%012d", len(f.byID)+1)
	f.byID[c.ID] = c
	return c, nil
}

func (f *fakeRepo) Get(_ context.Context, cid, id string) (Character, error) {
	if c, ok := f.byID[id]; ok && c.CampaignID == cid {
		return c, nil
	}
	return Character{}, ErrNotFound
}

func (f *fakeRepo) List(_ context.Context, cid, kind string) ([]Character, error) {
	out := []Character{}
	for _, c := range f.byID {
		if c.CampaignID == cid && (kind == "" || c.Kind == kind) {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (f *fakeRepo) UpdateSheet(_ context.Context, cid, id string, n Character) (Character, error) {
	c, err := f.Get(context.Background(), cid, id)
	if err != nil {
		return c, err
	}
	c.Name, c.PortraitURL, c.Sheet = n.Name, n.PortraitURL, n.Sheet
	f.byID[id] = c
	return c, nil
}

func (f *fakeRepo) Delete(_ context.Context, cid, id string) error {
	if _, err := f.Get(context.Background(), cid, id); err != nil {
		return err
	}
	delete(f.byID, id)
	return nil
}

type fakeCampaigns struct{ settings string }

func (f *fakeCampaigns) Detail(_ context.Context, m campaigns.Member) (campaigns.Summary, error) {
	return campaigns.Summary{Campaign: campaigns.Campaign{ID: m.CampaignID, SystemID: "assimilacao", Settings: []byte(f.settings)}}, nil
}

func (f *fakeCampaigns) Membership(_ context.Context, cid, uid string) (campaigns.Member, error) {
	if uid == "outsider" {
		return campaigns.Member{}, campaigns.ErrNotFound
	}
	return campaigns.Member{CampaignID: cid, UserID: uid, Role: campaigns.RolePlayer}, nil
}

type event struct {
	typ     string
	gmOnly  bool
	payload any
}

type fakeEvents struct{ got []event }

func (f *fakeEvents) Broadcast(_ string, to realtime.Filter, typ string, payload any) {
	f.got = append(f.got, event{typ, !to(realtime.Sender{Role: campaigns.RolePlayer}), payload})
}
