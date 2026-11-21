package characters

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/auth"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/migrations"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/platform/database"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems"
)

type dbEnv struct {
	h      http.Handler
	store  *Store
	camp   campaigns.Campaign
	gm     auth.User
	ana    auth.User
	cookie map[string]*http.Cookie
}

func newDBEnv(t *testing.T) *dbEnv {
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
	if _, err := pool.Exec(ctx, `TRUNCATE users CASCADE`); err != nil {
		t.Fatal(err)
	}
	return seedEnv(t, pool)
}

func seedEnv(t *testing.T, pool *pgxpool.Pool) *dbEnv {
	t.Helper()
	authSvc := auth.NewService(auth.NewStore(pool))
	campStore := campaigns.NewStore(pool)
	reg := systems.Default()
	campSvc := campaigns.NewService(campStore, authSvc, reg)
	e := &dbEnv{store: NewStore(pool), cookie: map[string]*http.Cookie{}}
	e.gm, e.ana = e.user(t, authSvc, "gm"), e.user(t, authSvc, "ana")
	e.camp, _ = campStore.CreateCampaign(ctx, "T", "assimilacao", e.gm.ID)
	inv, _ := campSvc.CreateInvite(ctx, e.camp.ID)
	if _, err := campSvc.Accept(ctx, inv.Token, e.ana.ID); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	ch := campaigns.NewHandlers(campSvc, auth.NewHandlers(authSvc).RequireAuth)
	ch.Register(mux)
	NewHandlers(NewService(e.store, campSvc, reg, realtime.NewHub(realtime.Options{})), ch.RequireMember).Register(mux)
	e.h = mux
	return e
}

func (e *dbEnv) user(t *testing.T, authSvc *auth.Service, name string) auth.User {
	t.Helper()
	u, err := authSvc.CreateUser(ctx, name+"@x.com", name, name+"-password", false)
	if err != nil {
		t.Fatal(err)
	}
	_, tok, _, err := authSvc.Login(ctx, name+"@x.com", name+"-password")
	if err != nil {
		t.Fatal(err)
	}
	e.cookie[name] = &http.Cookie{Name: auth.CookieName, Value: tok} //nolint:gosec // test cookie
	return u
}

func (e *dbEnv) do(who, method, path, body string) (int, map[string]any) {
	req := httptest.NewRequest(method, "/api/campaigns/"+e.camp.ID+path, strings.NewReader(body))
	req.AddCookie(e.cookie[who])
	rec := httptest.NewRecorder()
	e.h.ServeHTTP(rec, req)
	var out map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

func TestStoreRoundTrip(t *testing.T) {
	e := newDBEnv(t)
	own := e.ana.ID
	c, err := e.store.Create(ctx, Character{CampaignID: e.camp.ID, OwnerID: &own, Kind: KindPC, Name: "Rita", Sheet: json.RawMessage(`{"a":1}`)})
	if err != nil || c.OwnerName != "ana" || c.PortraitURL != nil {
		t.Fatalf("create = %+v %v", c, err)
	}
	url := "http://x/p.png"
	u, err := e.store.UpdateSheet(ctx, e.camp.ID, c.ID, Character{Name: "Rita2", PortraitURL: &url, Sheet: json.RawMessage(`{"a":2}`)})
	if err != nil || u.Name != "Rita2" || *u.PortraitURL != url || !u.UpdatedAt.After(c.UpdatedAt) {
		t.Fatalf("update = %+v %v", u, err)
	}
	if l, _ := e.store.List(ctx, e.camp.ID, KindNPC); len(l) != 0 {
		t.Errorf("npc list = %d", len(l))
	}
	if l, _ := e.store.List(ctx, e.camp.ID, ""); len(l) != 1 {
		t.Errorf("all list = %d", len(l))
	}
	if _, err := e.store.Get(ctx, "00000000-0000-0000-0000-000000000000", c.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("other campaign = %v", err)
	}
	if err := e.store.Delete(ctx, e.camp.ID, c.ID); err != nil {
		t.Fatal(err)
	}
	if err := e.store.Delete(ctx, e.camp.ID, c.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("double delete = %v", err)
	}
}

func TestHTTPFlowWithSettings(t *testing.T) {
	e := newDBEnv(t)
	code, ch := e.do("ana", "POST", "/characters", `{"name":"Rita"}`)
	if code != 201 {
		t.Fatalf("create = %d %v", code, ch)
	}
	id := ch["id"].(string)
	sheet := `{"sheet":{"identity":{"name":"Rita"},"instincts":{"potency":2,"resolution":1}}}`
	code, out := e.do("ana", "PUT", "/characters/"+id, sheet)
	if d, _ := out["derived"].(map[string]any); code != 200 || d["boxesPerLevel"] != float64(4) {
		t.Errorf("save = %d %v", code, out)
	}
	if code, _ := e.do("gm", "PUT", "/characters/"+id, sheet); code != 200 {
		t.Errorf("gm save = %d", code)
	}
	if code, _ := e.do("gm", "POST", "/characters", `{"ownerId":"`+e.gm.ID+`"}`); code != 201 {
		t.Errorf("gm own = %d", code)
	}
	if code, _ := e.do("ana", "PUT", "/characters/"+id+"x", sheet); code != 404 {
		t.Errorf("bad id = %d", code)
	}
	e.checkBlockMode(t, id)
}

func (e *dbEnv) checkBlockMode(t *testing.T, id string) {
	t.Helper()
	if code, _ := e.do("gm", "PATCH", "/settings", `{"creationLimits":"block"}`); code != 200 {
		t.Fatal("settings")
	}
	bad := `{"sheet":{"identity":{"name":"R"},"instincts":{"influence":3,"perception":3,"potency":3,"reaction":3,"resolution":3,"sagacity":3}}}`
	if code, out := e.do("ana", "PUT", "/characters/"+id, bad); code != 400 || out["issues"] == nil {
		t.Errorf("block = %d %v", code, out)
	}
	if code, _ := e.do("ana", "PATCH", "/settings", `{"creationLimits":"warn"}`); code != 403 {
		t.Errorf("player settings = %d", code)
	}
	if code, _ := e.do("ana", "DELETE", "/characters/"+id, ""); code != 204 {
		t.Errorf("delete = %d", code)
	}
}
