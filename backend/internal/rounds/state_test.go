package rounds

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/migrations"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/platform/database"
)

func TestTransitions(t *testing.T) {
	s := Start()
	if s != (State{true, 1, PhasePreparation}) {
		t.Fatalf("start: %+v", s)
	}
	if _, err := s.Prev(); !errors.Is(err, errAtStart) {
		t.Errorf("prev at start: %v", err)
	}
	for _, want := range []Phase{PhaseInfected, PhaseAssimilator, PhaseEnd} {
		s, _ = s.Next()
		if s.Phase != want || s.Round != 1 {
			t.Fatalf("next: %+v want %s", s, want)
		}
	}
	s, _ = s.Next()
	if s != (State{true, 2, PhasePreparation}) {
		t.Fatalf("after end: %+v", s)
	}
}

func TestPrevAndIdle(t *testing.T) {
	s := State{true, 2, PhasePreparation}
	s, _ = s.Prev()
	if s != (State{true, 1, PhaseEnd}) {
		t.Fatalf("prev from prep r2: %+v", s)
	}
	s, _ = s.Prev()
	if s.Phase != PhaseAssimilator {
		t.Fatalf("prev: %+v", s)
	}
	if _, err := Idle().Next(); !errors.Is(err, errInactive) {
		t.Errorf("next idle: %v", err)
	}
	if _, err := Idle().Prev(); !errors.Is(err, errInactive) {
		t.Errorf("prev idle: %v", err)
	}
}

type memStore map[string]State

func (m memStore) Get(_ context.Context, id string) (State, error) {
	if s, ok := m[id]; ok {
		return s, nil
	}
	return Idle(), nil
}
func (m memStore) Save(_ context.Context, id string, s State) error { m[id] = s; return nil }

func TestServiceHooksAndEnd(t *testing.T) {
	svc := NewService(memStore{})
	var prevs, nexts []State
	svc.OnPhaseChange(func(_ context.Context, id string, p, n State) {
		if id != "c1" {
			t.Errorf("campaign %s", id)
		}
		prevs, nexts = append(prevs, p), append(nexts, n)
	})
	ctx := context.Background()
	if _, err := svc.Next(ctx, "c1"); err == nil || len(nexts) != 0 {
		t.Fatalf("idle next must fail without hook: %v", err)
	}
	_, _ = svc.Start(ctx, "c1")
	_, _ = svc.Next(ctx, "c1")
	end, _ := svc.End(ctx, "c1")
	if end.Active || len(nexts) != 3 {
		t.Fatalf("end %+v hooks %d", end, len(nexts))
	}
	if prevs[1] != (State{true, 1, PhasePreparation}) || nexts[1].Phase != PhaseInfected {
		t.Errorf("hook args: %+v -> %+v", prevs[1], nexts[1])
	}
}

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	pool, err := database.Connect(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := migrations.Up(ctx, database.StdDB(pool)); err != nil {
		t.Fatal(err)
	}
	return pool
}

func TestPgStorePersists(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	var uid, cid string
	if err := pool.QueryRow(ctx, `INSERT INTO users (email, password_hash, display_name) VALUES ('rounds-'||gen_random_uuid()||'@x.com','h','R') RETURNING id`).Scan(&uid); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO campaigns (name, system_id, owner_id) VALUES ('r','assimilacao',$1) RETURNING id`, uid).Scan(&cid); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, uid) }()
	st := NewStore(pool)
	if got, _ := st.Get(ctx, cid); got != Idle() {
		t.Fatalf("default: %+v", got)
	}
	want := State{true, 3, PhaseAssimilator}
	if err := st.Save(ctx, cid, want); err != nil {
		t.Fatal(err)
	}
	if got, _ := NewStore(pool).Get(ctx, cid); got != want {
		t.Fatalf("persisted: %+v", got)
	}
}
