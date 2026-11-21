package sessions

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/auth"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/migrations"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/platform/database"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
)

var ctx = context.Background()

type memStore struct{ active *Session }

func (m *memStore) Active(context.Context, string) (Session, error) {
	if m.active == nil {
		return Session{}, ErrNoneActive
	}
	return *m.active, nil
}

func (m *memStore) Start(context.Context, string) (Session, error) {
	if m.active != nil {
		return Session{}, ErrAlreadyActive
	}
	m.active = &Session{ID: "s1", StartedAt: time.Now()}
	return *m.active, nil
}

func (m *memStore) End(context.Context, string) error {
	if m.active == nil {
		return ErrNoneActive
	}
	m.active = nil
	return nil
}

type events struct{ got []State }

func (e *events) Broadcast(_ string, to realtime.Filter, _ string, p any) {
	if to(realtime.Sender{Role: campaigns.RolePlayer}) {
		e.got = append(e.got, p.(State))
	}
}

func TestStartEndBroadcastsToEveryone(t *testing.T) {
	ev := &events{}
	svc := NewService(&memStore{}, ev)
	if err := svc.Start(ctx, "c"); err != nil {
		t.Fatal(err)
	}
	if id, _ := svc.ActiveID(ctx, "c"); id != "s1" {
		t.Errorf("active id = %q", id)
	}
	if err := svc.Start(ctx, "c"); !errors.Is(err, ErrAlreadyActive) {
		t.Errorf("double start = %v", err)
	}
	if err := svc.End(ctx, "c"); err != nil {
		t.Fatal(err)
	}
	if err := svc.End(ctx, "c"); !errors.Is(err, ErrNoneActive) {
		t.Errorf("double end = %v", err)
	}
	if len(ev.got) != 2 || !ev.got[0].Active || ev.got[0].StartedAt == nil || ev.got[1].Active {
		t.Errorf("events = %+v", ev.got)
	}
	if id, _ := svc.ActiveID(ctx, "c"); id != "" {
		t.Errorf("after end = %q", id)
	}
}

func pgStore(t *testing.T) (*PgStore, string) {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	pool, err := database.Connect(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := migrations.Up(ctx, database.StdDB(pool)); err != nil {
		t.Fatal(err)
	}
	_, _ = pool.Exec(ctx, `TRUNCATE users CASCADE`)
	u, _ := auth.NewService(auth.NewStore(pool)).CreateUser(ctx, "gm@x.com", "gm", "gm-password", false)
	c, _ := campaigns.NewStore(pool).CreateCampaign(ctx, "T", "assimilacao", u.ID)
	return NewStore(pool), c.ID
}

func TestPgStore(t *testing.T) {
	st, campID := pgStore(t)
	if _, err := st.Active(ctx, campID); !errors.Is(err, ErrNoneActive) {
		t.Fatalf("active = %v", err)
	}
	first, err := st.Start(ctx, campID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.Start(ctx, campID); !errors.Is(err, ErrAlreadyActive) {
		t.Errorf("double start = %v", err)
	}
	if err := st.End(ctx, campID); err != nil {
		t.Fatal(err)
	}
	if second, err := st.Start(ctx, campID); err != nil || second.ID == first.ID {
		t.Errorf("restart = %+v %v", second, err)
	}
}
