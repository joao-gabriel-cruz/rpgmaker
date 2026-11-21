package vision

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/auth"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/maps"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/migrations"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/platform/database"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
)

// seed creates a campaign with one 16x16 scene and a token of a PC at (2,3).
func seed(t *testing.T) (*pgxpool.Pool, string) {
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
	ch, _ := characters.NewStore(pool).Create(c, characters.Character{CampaignID: cp.ID, OwnerID: &u.ID,
		Kind: characters.KindPC, Name: "Ana", Sheet: []byte(`{}`)})
	if _, err := tokens.NewStore(pool).Place(c, sc.ID, ch.ID, 2, 3); err != nil {
		t.Fatal(err)
	}
	return pool, sc.ID
}

func TestStoreRoundTrip(t *testing.T) {
	pool, scID := seed(t)
	c := context.Background()
	s := NewStore(pool)
	if f, _ := s.Fog(c, scID); f.Mode != ModeOff || f.Radius != DefaultRadius {
		t.Fatalf("defaults: %+v", f)
	}
	if err := s.SaveRevealed(c, scID, "AQ=="); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveConfig(c, scID, ModeDynamic, 7); err != nil {
		t.Fatal(err)
	}
	if f, _ := s.Fog(c, scID); f.Mode != ModeDynamic || f.Radius != 7 || f.Revealed != "AQ==" {
		t.Fatalf("upserts keep both halves: %+v", f)
	}
	if p, err := s.Positions(c, scID); err != nil || len(p) != 1 || p[0].X != 2 || p[0].OwnerID == nil {
		t.Fatalf("positions: %+v %v", p, err)
	}
}
