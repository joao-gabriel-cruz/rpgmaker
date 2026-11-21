package ambience

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/auth"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/migrations"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/platform/database"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
)

// seedScene creates a campaign with one scene and returns the pool and the scene id.
func seedScene(t *testing.T) (*pgxpool.Pool, string) {
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
	_, _ = pool.Exec(c, `TRUNCATE users CASCADE`)
	u, _ := auth.NewStore(pool).CreateUser(c, "gm@x.com", "h", "GM", false)
	cp, _ := campaigns.NewStore(pool).CreateCampaign(c, "T", "assimilacao", u.ID)
	g, _ := maps.NewGrid(16, 16)
	sc, err := scenes.NewStore(pool).Create(c, cp.ID, "S", 16, 16, g.Serialize().Data)
	if err != nil {
		t.Fatal(err)
	}
	return pool, sc.ID
}

func TestStoreRoundTrip(t *testing.T) {
	pool, sceneID := seedScene(t)
	c := context.Background()
	s := NewStore(pool)
	if a, _ := s.Get(c, sceneID); a != Default(sceneID) {
		t.Fatalf("defaults: %+v", a)
	}
	want := Ambience{SceneID: sceneID, TimeOfDay: "dusk", Weather: "ash"}
	if err := s.Save(c, want); err != nil {
		t.Fatal(err)
	}
	if a, _ := s.Get(c, sceneID); a != want {
		t.Fatalf("persisted: %+v", a)
	}
}
