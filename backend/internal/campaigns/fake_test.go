package campaigns

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/auth"
)

type fakeAuthRepo struct {
	users    map[string]auth.User
	sessions map[string]string
}

func (f *fakeAuthRepo) CreateUser(_ context.Context, email, hash, name string, admin bool) (auth.User, error) {
	for _, u := range f.users {
		if strings.EqualFold(u.Email, email) {
			return auth.User{}, auth.ErrEmailTaken
		}
	}
	id := fmt.Sprintf("00000000-0000-0000-0000-%012d", len(f.users)+1)
	u := auth.User{ID: id, Email: email, PasswordHash: hash, DisplayName: name, IsAdmin: admin}
	f.users[id] = u
	return u, nil
}

func (f *fakeAuthRepo) UserByEmail(_ context.Context, email string) (auth.User, error) {
	for _, u := range f.users {
		if strings.EqualFold(u.Email, email) {
			return u, nil
		}
	}
	return auth.User{}, auth.ErrNotFound
}

func (f *fakeAuthRepo) UserByID(context.Context, string) (auth.User, error) {
	return auth.User{}, auth.ErrNotFound
}
func (f *fakeAuthRepo) ListUsers(context.Context) ([]auth.User, error)    { return nil, nil }
func (f *fakeAuthRepo) SetPassword(context.Context, string, string) error { return nil }

func (f *fakeAuthRepo) CreateSession(_ context.Context, id, userID string, _ time.Time) error {
	f.sessions[id] = userID
	return nil
}

func (f *fakeAuthRepo) UserBySession(_ context.Context, id string) (auth.User, error) {
	if uid, ok := f.sessions[id]; ok {
		return f.users[uid], nil
	}
	return auth.User{}, auth.ErrNotFound
}

func (f *fakeAuthRepo) DeleteSession(context.Context, string) error { return nil }

type fakeRepo struct {
	users     *fakeAuthRepo
	campaigns map[string]Campaign
	members   map[string]map[string]string // campaign -> user -> role
	invites   map[string]*Invite
}

func newFakeRepo(users *fakeAuthRepo) *fakeRepo {
	return &fakeRepo{users: users, campaigns: map[string]Campaign{},
		members: map[string]map[string]string{}, invites: map[string]*Invite{}}
}

func (f *fakeRepo) CreateCampaign(_ context.Context, name, systemID, ownerID string) (Campaign, error) {
	id := fmt.Sprintf("11111111-1111-1111-1111-%012d", len(f.campaigns)+1)
	c := Campaign{ID: id, Name: name, SystemID: systemID, OwnerID: ownerID, Settings: []byte("{}")}
	f.campaigns[id] = c
	f.members[id] = map[string]string{ownerID: RoleGM}
	return c, nil
}

func (f *fakeRepo) ListByUser(_ context.Context, userID string) ([]Summary, error) {
	out := []Summary{}
	for id, m := range f.members {
		if role, ok := m[userID]; ok {
			out = append(out, Summary{Campaign: f.campaigns[id], Role: role})
		}
	}
	return out, nil
}

func (f *fakeRepo) Campaign(_ context.Context, id string) (Campaign, error) {
	if c, ok := f.campaigns[id]; ok {
		return c, nil
	}
	return Campaign{}, ErrNotFound
}

func (f *fakeRepo) Member(_ context.Context, cid, uid string) (Member, error) {
	if role, ok := f.members[cid][uid]; ok {
		return Member{CampaignID: cid, UserID: uid, Role: role}, nil
	}
	return Member{}, ErrNotFound
}

func (f *fakeRepo) Members(_ context.Context, cid string) ([]MemberInfo, error) {
	out := []MemberInfo{}
	for uid, role := range f.members[cid] {
		out = append(out, MemberInfo{UserID: uid, DisplayName: f.users.users[uid].DisplayName, Role: role})
	}
	return out, nil
}

func (f *fakeRepo) RemoveMember(_ context.Context, cid, uid string) error {
	delete(f.members[cid], uid)
	return nil
}

func (f *fakeRepo) CreateInvite(_ context.Context, inv Invite) error {
	inv.CampaignName = f.campaigns[inv.CampaignID].Name
	f.invites[inv.Token] = &inv
	return nil
}

func (f *fakeRepo) InviteByToken(_ context.Context, token string) (Invite, error) {
	if inv, ok := f.invites[token]; ok {
		return *inv, nil
	}
	return Invite{}, ErrNotFound
}

func (f *fakeRepo) ConsumeInvite(_ context.Context, token, uid string, now time.Time) (Member, error) {
	inv, ok := f.invites[token]
	if !ok || inv.UsedBy != nil || !inv.ExpiresAt.After(now) {
		return Member{}, ErrInvalidInvite
	}
	if _, in := f.members[inv.CampaignID][uid]; in {
		return Member{}, ErrAlreadyMember
	}
	inv.UsedBy = &uid
	f.members[inv.CampaignID][uid] = inv.Role
	return Member{CampaignID: inv.CampaignID, UserID: uid, Role: inv.Role}, nil
}
