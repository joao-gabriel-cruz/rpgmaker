package tokens

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/auth"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/migrations"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/platform/database"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
)

func testPool(t *testing.T) *pgxpool.Pool {
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
	return pool
}

// seed creates a campaign with one scene and one player character, returning their ids.
func seed(t *testing.T, pool *pgxpool.Pool) (campID, sceneID, charID string) {
	t.Helper()
	c := context.Background()
	u, _ := auth.NewStore(pool).CreateUser(c, "gm@x.com", "h", "GM", false)
	cp, err := campaigns.NewStore(pool).CreateCampaign(c, "T", "assimilacao", u.ID)
	if err != nil {
		t.Fatal(err)
	}
	sc, err := scenes.NewStore(pool).Create(c, cp.ID, "S", 16, 16, scenes16())
	if err != nil {
		t.Fatal(err)
	}
	ch, err := characters.NewStore(pool).Create(c, characters.Character{CampaignID: cp.ID, OwnerID: &u.ID,
		Kind: characters.KindPC, Name: "Ana", Sheet: []byte(`{}`)})
	if err != nil {
		t.Fatal(err)
	}
	return cp.ID, sc.ID, ch.ID
}

func TestStoreRoundTrip(t *testing.T) {
	pool := testPool(t)
	c := context.Background()
	campID, scID, chID := seed(t, pool)
	s := NewStore(pool)
	tok, err := s.Place(c, scID, chID, 2, 3)
	if err != nil || tok.Name != "Ana" || tok.OwnerID == nil {
		t.Fatalf("place: %+v %v", tok, err)
	}
	if _, err := s.Place(c, scID, chID, 1, 1); !errors.Is(err, ErrExists) {
		t.Fatalf("want ErrExists, got %v", err)
	}
	if err := s.Move(c, tok.ID, 4, 5); err != nil {
		t.Fatal(err)
	}
	got, _ := s.Get(c, campID, tok.ID)
	list, _ := s.List(c, campID, scID)
	if got.X != 4 || got.Z != 5 || len(list) != 1 {
		t.Fatalf("get/list: %+v %v", got, list)
	}
	if err := s.Remove(c, tok.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(c, campID, tok.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestStoreAppearancePayload(t *testing.T) {
	pool := testPool(t)
	c := context.Background()
	campID, scID, chID := seed(t, pool)
	s := NewStore(pool)
	tok, _ := s.Place(c, scID, chID, 1, 1)
	if string(tok.Appearance) != "null" && tok.Appearance != nil || tok.AssimilationLevel != nil {
		t.Fatalf("empty sheet: %s %v", tok.Appearance, tok.AssimilationLevel)
	}
	sheet := `{"appearance":{"hair":{"style":"long"}},"tugOfWar":{"assimilationLevel":7}}`
	if _, err := pool.Exec(c, `UPDATE characters SET sheet = $2::jsonb WHERE id = $1`, chID, sheet); err != nil {
		t.Fatal(err)
	}
	list, _ := s.List(c, campID, scID)
	if len(list) != 1 || string(list[0].Appearance) != `{"hair": {"style": "long"}}` || list[0].AssimilationLevel == nil || *list[0].AssimilationLevel != 7 {
		t.Fatalf("list: %+v", list)
	}
}

func scenes16() string {
	b := make([]byte, 16*16*2)
	for i := range b {
		b[i] = '0'
	}
	return string(b)
}
