package mcp

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
)

type fakeRepo struct {
	byHash  map[string]Link
	revoked map[string]bool
	active  map[string]string // campaign|user -> hash
	touched []string
	n       int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{byHash: map[string]Link{}, revoked: map[string]bool{}, active: map[string]string{}}
}

func (f *fakeRepo) Replace(_ context.Context, c, u, hash string) (Link, error) {
	if old, ok := f.active[c+"|"+u]; ok {
		f.revoked[old] = true
	}
	f.n++
	l := Link{ID: fmt.Sprintf("link-%d", f.n), CampaignID: c, UserID: u, CreatedAt: time.Now()}
	f.byHash[hash], f.active[c+"|"+u] = l, hash
	return l, nil
}

func (f *fakeRepo) Active(_ context.Context, c, u string) (Link, error) {
	h, ok := f.active[c+"|"+u]
	if !ok || f.revoked[h] {
		return Link{}, ErrNoLink
	}
	return f.byHash[h], nil
}

func (f *fakeRepo) ByHash(_ context.Context, h string) (Link, error) {
	l, ok := f.byHash[h]
	if !ok || f.revoked[h] {
		return Link{}, ErrInvalidLink
	}
	return l, nil
}

func (f *fakeRepo) Revoke(_ context.Context, c, u string) error {
	if h, ok := f.active[c+"|"+u]; ok {
		f.revoked[h] = true
	}
	return nil
}

func (f *fakeRepo) Touch(_ context.Context, id string) error {
	f.touched = append(f.touched, id)
	return nil
}

type fakeMembers struct{ role string }

func (f fakeMembers) Membership(_ context.Context, c, u string) (campaigns.Member, error) {
	if f.role == "" {
		return campaigns.Member{}, campaigns.ErrNotFound
	}
	return campaigns.Member{CampaignID: c, UserID: u, Role: f.role}, nil
}

var gm = campaigns.Member{CampaignID: "c1", UserID: "u1", Role: campaigns.RoleGM}

func TestGenerateReturnsSecretOnceAndStoresHash(t *testing.T) {
	repo := newFakeRepo()
	secret, _, err := NewLinks(repo, fakeMembers{campaigns.RoleGM}).Generate(context.Background(), gm)
	if err != nil {
		t.Fatal(err)
	}
	if len(secret) != 43 {
		t.Fatalf("secret length %d", len(secret))
	}
	if _, ok := repo.byHash[secret]; ok {
		t.Fatal("secret stored in clear")
	}
	if _, ok := repo.byHash[hashSecret(secret)]; !ok {
		t.Fatal("hash not stored")
	}
}

func TestGenerateRevokesPrevious(t *testing.T) {
	l := NewLinks(newFakeRepo(), fakeMembers{campaigns.RoleGM})
	first, _, _ := l.Generate(context.Background(), gm)
	second, _, _ := l.Generate(context.Background(), gm)
	if _, err := l.Resolve(context.Background(), first); !errors.Is(err, ErrInvalidLink) {
		t.Fatalf("first still valid: %v", err)
	}
	if _, err := l.Resolve(context.Background(), second); err != nil {
		t.Fatal(err)
	}
}

func TestResolveRejectsRevokedUnknownAndNonGM(t *testing.T) {
	ctx := context.Background()
	l := NewLinks(newFakeRepo(), fakeMembers{campaigns.RoleGM})
	secret, _, _ := l.Generate(ctx, gm)
	_ = l.Revoke(ctx, gm)
	if _, err := l.Resolve(ctx, secret); !errors.Is(err, ErrInvalidLink) {
		t.Fatalf("revoked: %v", err)
	}
	if _, err := l.Resolve(ctx, "nope"); !errors.Is(err, ErrInvalidLink) {
		t.Fatalf("unknown: %v", err)
	}
	for _, role := range []string{campaigns.RolePlayer, ""} {
		repo := newFakeRepo()
		s, _, _ := NewLinks(repo, fakeMembers{campaigns.RoleGM}).Generate(ctx, gm)
		demoted := NewLinks(repo, fakeMembers{role})
		if _, err := demoted.Resolve(ctx, s); !errors.Is(err, ErrInvalidLink) {
			t.Fatalf("role %q: %v", role, err)
		}
	}
}

func TestResolveBuildsClaudeSender(t *testing.T) {
	repo := newFakeRepo()
	l := NewLinks(repo, fakeMembers{campaigns.RoleGM})
	secret, link, _ := l.Generate(context.Background(), gm)
	s, err := l.Resolve(context.Background(), secret)
	if err != nil {
		t.Fatal(err)
	}
	if s.CampaignID != "c1" || s.Sender.UserID != "u1" || s.Sender.Role != campaigns.RoleGM || s.Sender.Name != "Claude" {
		t.Fatalf("session: %+v", s)
	}
	if len(repo.touched) != 1 || repo.touched[0] != link.ID {
		t.Fatalf("touched: %v", repo.touched)
	}
}

func TestGenerateRequiresGM(t *testing.T) {
	l := NewLinks(newFakeRepo(), fakeMembers{campaigns.RoleGM})
	player := campaigns.Member{CampaignID: "c1", UserID: "u2", Role: campaigns.RolePlayer}
	if _, _, err := l.Generate(context.Background(), player); !errors.Is(err, ErrForbidden) {
		t.Fatalf("got %v", err)
	}
}

type failingMembers struct{ err error }

func (f failingMembers) Membership(context.Context, string, string) (campaigns.Member, error) {
	return campaigns.Member{}, f.err
}

func TestResolveSurfacesInfraErrors(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	secret, _, _ := NewLinks(repo, fakeMembers{campaigns.RoleGM}).Generate(ctx, gm)
	boom := errors.New("db down")
	_, err := NewLinks(repo, failingMembers{boom}).Resolve(ctx, secret)
	if !errors.Is(err, boom) || errors.Is(err, ErrInvalidLink) {
		t.Fatalf("err = %v", err)
	}
}
