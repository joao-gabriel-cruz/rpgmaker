package effects

import (
	"context"
	"errors"
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

func TestStoreRoundTrip(t *testing.T) {
	pool := testPool(t)
	c := context.Background()
	u, _ := auth.NewStore(pool).CreateUser(c, "gm@x.com", "h", "GM", false)
	cp, err := campaigns.NewStore(pool).CreateCampaign(c, "T", "assimilacao", u.ID)
	if err != nil {
		t.Fatal(err)
	}
	g, _ := maps.NewGrid(16, 16)
	sc, err := scenes.NewStore(pool).Create(c, cp.ID, "S", 16, 16, g.Serialize().Data)
	if err != nil {
		t.Fatal(err)
	}
	ch, err := characters.NewStore(pool).Create(c, characters.Character{CampaignID: cp.ID, OwnerID: &u.ID,
		Kind: characters.KindPC, Name: "Ana", Sheet: []byte(`{}`)})
	if err != nil {
		t.Fatal(err)
	}
	s := NewStore(pool)
	checkRules(t, s, cp.ID)
	checkSuggestions(t, s, cp.ID, ch.ID, sc.ID)
}

func checkRules(t *testing.T, s *PgStore, campID string) {
	t.Helper()
	c := context.Background()
	if err := s.SaveRules(c, campID, KindTerrain, "lava", []Rule{{TriggerEnter, EffectDamage, 4, "x"}}); err != nil {
		t.Fatal(err)
	}
	over, _ := s.Overrides(c, campID)
	if r := over["terrain:lava"]; len(r) != 1 || r[0].Amount != 4 {
		t.Fatalf("overrides %+v", over)
	}
	_ = s.DeleteRules(c, campID, KindTerrain, "lava")
	if over, _ = s.Overrides(c, campID); len(over) != 0 {
		t.Fatal("delete failed")
	}
}

func checkSuggestions(t *testing.T, s *PgStore, campID, charID, sceneID string) {
	t.Helper()
	c := context.Background()
	sg, err := s.Insert(c, Suggestion{CampaignID: campID, CharacterID: charID, CharacterName: "Ana", SceneID: sceneID,
		Source: "Lava", Trigger: TriggerEnter, Effect: EffectDamage, Amount: 2})
	if err != nil || sg.ID == "" {
		t.Fatal(err)
	}
	if list, _ := s.Pending(c, campID); len(list) != 1 {
		t.Fatal("expected pending")
	}
	if _, err := s.Resolve(c, campID, sg.ID, StatusApplied); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Resolve(c, campID, sg.ID, StatusApplied); !errors.Is(err, ErrNotFound) {
		t.Fatalf("second resolve must miss, got %v", err)
	}
	_ = s.Reopen(c, campID, sg.ID)
	if list, _ := s.Pending(c, campID); len(list) != 1 {
		t.Fatal("reopen failed")
	}
}
