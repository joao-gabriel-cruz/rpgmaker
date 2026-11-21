package rolls

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

const (
	camp    = "00000000-0000-0000-0000-00000000000c"
	charID  = "22222222-2222-2222-2222-000000000001"
	npcID   = "22222222-2222-2222-2222-000000000002"
	anaID   = "ana"
	beaID   = "bea"
	gmUser  = "gm"
	anaName = "Ana"
)

type fakeRepo struct{ recs []Record }

func (f *fakeRepo) Insert(_ context.Context, r Record) (Record, error) {
	r.ID = fmt.Sprintf("33333333-3333-3333-3333-%012d", len(f.recs)+1)
	r.UserName, r.CreatedAt = r.UserID, time.Now()
	f.recs = append(f.recs, r)
	return r, nil
}

func (f *fakeRepo) Get(_ context.Context, _, id string) (Record, error) {
	for _, r := range f.recs {
		if r.ID == id {
			return r, nil
		}
	}
	return Record{}, ErrNotFound
}

func (f *fakeRepo) Finalize(_ context.Context, _, id string, res json.RawMessage) error {
	for i := range f.recs {
		if f.recs[i].ID == id {
			if f.recs[i].Status != StatusPending {
				return ErrAlreadyFinal
			}
			f.recs[i].Status, f.recs[i].Result = StatusFinal, res
			return nil
		}
	}
	return ErrNotFound
}

func (f *fakeRepo) Recent(_ context.Context, _ string, _ int, includeGM bool) ([]Record, error) {
	out := []Record{}
	for _, r := range f.recs {
		if includeGM || r.Visibility == VisPublic {
			out = append(out, r)
		}
	}
	return out, nil
}

// fakeChars serves one PC owned by ana and one NPC; damage sets the sheet health.
type fakeChars struct{ damage int }

func (f *fakeChars) Get(_ context.Context, m campaigns.Member, id string) (characters.View, error) {
	owner := anaID
	kind := characters.KindPC
	if id == npcID {
		kind = characters.KindNPC
		if m.Role != campaigns.RoleGM {
			return characters.View{}, characters.ErrNotFound
		}
	} else if id != charID {
		return characters.View{}, characters.ErrNotFound
	}
	sheet := fmt.Sprintf(`{"instincts":{"reaction":2,"potency":1,"resolution":1},"practices":{"athletics":1},"health":{"damage":%d}}`, f.damage)
	c := characters.Character{ID: id, OwnerID: &owner, Kind: kind, Name: "Rita", Sheet: json.RawMessage(sheet)}
	return characters.View{Character: c}, nil
}

type event struct {
	typ      string
	toGM     bool
	toPlayer bool
	payload  any
}

type fakeEvents struct{ got []event }

func (f *fakeEvents) Broadcast(_ string, to realtime.Filter, typ string, payload any) {
	f.got = append(f.got, event{typ,
		to(realtime.Sender{Role: campaigns.RoleGM}), to(realtime.Sender{Role: campaigns.RolePlayer}), payload})
}

// seqRNG returns a fixed face sequence, repeating the last value.
type seqRNG struct {
	vals []int
	i    int
}

func (r *seqRNG) Intn(n int) int {
	v := r.vals[min(r.i, len(r.vals)-1)] % n
	r.i++
	return v
}

func errNotFoundChar() error { return characters.ErrNotFound }
