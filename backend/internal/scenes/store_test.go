package scenes

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/auth"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/migrations"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/platform/database"
)

func testStore(t *testing.T) (*Store, string) {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	c := context.Background()
	pool, err := database.Connect(c, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := migrations.Up(c, database.StdDB(pool)); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(c, `TRUNCATE users CASCADE`); err != nil {
		t.Fatal(err)
	}
	u, _ := auth.NewStore(pool).CreateUser(c, "gm@x.com", "h", "GM", false)
	camp, err := campaigns.NewStore(pool).CreateCampaign(c, "T", "assimilacao", u.ID)
	if err != nil {
		t.Fatal(err)
	}
	return NewStore(pool), camp.ID
}

func TestStoreRoundTrip(t *testing.T) {
	s, campID := testStore(t)
	c := context.Background()
	camp := struct{ ID string }{campID}
	a, err := s.Create(c, camp.ID, "A", 16, 16, "0100")
	if err != nil || a.IsActive || len(a.Objects) != 0 {
		t.Fatalf("create: %+v %v", a, err)
	}
	b, _ := s.Create(c, camp.ID, "B", 16, 16, "0100")
	for _, id := range []string{a.ID, b.ID} {
		if err := s.Activate(c, camp.ID, id); err != nil {
			t.Fatal(err)
		}
	}
	list, _ := s.List(c, camp.ID)
	if len(list) != 2 || list[0].IsActive || !list[1].IsActive {
		t.Errorf("list: %+v", list)
	}
	saveAndDelete(t, s, campID, a.ID, b.ID)
}

func saveAndDelete(t *testing.T, s *Store, campID, aID, bID string) {
	t.Helper()
	c := context.Background()
	camp := struct{ ID string }{campID}
	a, b := struct{ ID string }{aID}, struct{ ID string }{bID}
	if err := s.SaveObjects(c, a.ID, []maps.Object{{ID: "o", Prefab: "wall", X: 1, Z: 1}}); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveGrid(c, a.ID, "2a"); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(c, camp.ID, a.ID)
	if err != nil || got.Grid.Data != "2a" || len(got.Objects) != 1 || got.Objects[0].Prefab != "wall" {
		t.Errorf("get: %+v %v", got, err)
	}
	if err := s.Activate(c, camp.ID, "00000000-0000-0000-0000-000000000000"); !errors.Is(err, ErrNotFound) {
		t.Errorf("activate unknown: %v", err)
	}
	if l, _ := s.List(c, camp.ID); !l[1].IsActive {
		t.Error("failed activation must not deactivate the current scene")
	}
	if err := s.Delete(c, camp.ID, b.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(c, camp.ID, b.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("deleted: %v", err)
	}
}
