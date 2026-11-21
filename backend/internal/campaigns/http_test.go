package campaigns

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/auth"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems"
)

type env struct {
	h    http.Handler
	svc  *Service
	auth *auth.Service
	now  time.Time
}

func newEnv(t *testing.T) *env {
	t.Helper()
	ar := &fakeAuthRepo{users: map[string]auth.User{}, sessions: map[string]string{}}
	authSvc := auth.NewService(ar)
	for _, n := range []string{"ana", "bob", "cid"} {
		if _, err := authSvc.CreateUser(context.Background(), n+"@x.com", n, n+"-password", false); err != nil {
			t.Fatal(err)
		}
	}
	svc := NewService(newFakeRepo(ar), authSvc, systems.Default())
	e := &env{svc: svc, auth: authSvc, now: time.Now()}
	svc.now = func() time.Time { return e.now }
	mux := http.NewServeMux()
	ah := auth.NewHandlers(authSvc)
	NewHandlers(svc, ah.RequireAuth).Register(mux)
	e.h = mux
	return e
}

func (e *env) login(t *testing.T, name string) *http.Cookie {
	t.Helper()
	_, token, _, err := e.auth.Login(context.Background(), name+"@x.com", name+"-password")
	if err != nil {
		t.Fatal(err)
	}
	//nolint:gosec // test-only request cookie
	return &http.Cookie{Name: auth.CookieName, Value: token}
}

func (e *env) do(method, path, body string, c *http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if c != nil {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	e.h.ServeHTTP(rec, req)
	return rec
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var out map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return out
}

// seedCampaign has ana as gm and returns its id and her cookie.
func (e *env) seedCampaign(t *testing.T) (string, *http.Cookie) {
	t.Helper()
	ana := e.login(t, "ana")
	rec := e.do("POST", "/api/campaigns", `{"name":"Teste","systemId":"assimilacao"}`, ana)
	if rec.Code != 201 {
		t.Fatalf("create = %d %s", rec.Code, rec.Body)
	}
	return decodeBody(t, rec)["id"].(string), ana
}

func TestCreateAndListCampaigns(t *testing.T) {
	e := newEnv(t)
	id, ana := e.seedCampaign(t)
	rec := e.do("GET", "/api/campaigns", "", ana)
	if !strings.Contains(rec.Body.String(), `"role":"gm"`) || !strings.Contains(rec.Body.String(), id) {
		t.Errorf("list = %s", rec.Body)
	}
	if rec := e.do("POST", "/api/campaigns", `{"name":"X","systemId":"nope"}`, ana); rec.Code != 400 {
		t.Errorf("bad system = %d", rec.Code)
	}
	if rec := e.do("POST", "/api/campaigns", `{"name":" ","systemId":"assimilacao"}`, ana); rec.Code != 400 {
		t.Errorf("blank name = %d", rec.Code)
	}
	if rec := e.do("GET", "/api/campaigns", "", nil); rec.Code != 401 {
		t.Errorf("anonymous = %d", rec.Code)
	}
}

func TestDetailOnlyForMembers(t *testing.T) {
	e := newEnv(t)
	id, ana := e.seedCampaign(t)
	if rec := e.do("GET", "/api/campaigns/"+id, "", ana); rec.Code != 200 ||
		!strings.Contains(rec.Body.String(), `"systemName":"Assimilação RPG"`) {
		t.Errorf("member = %d %s", rec.Code, rec.Body)
	}
	if rec := e.do("GET", "/api/campaigns/"+id, "", e.login(t, "bob")); rec.Code != 404 {
		t.Errorf("non-member = %d", rec.Code)
	}
	if rec := e.do("GET", "/api/campaigns/not-a-uuid", "", ana); rec.Code != 404 {
		t.Errorf("bad id = %d", rec.Code)
	}
}

func TestInviteSignupFlow(t *testing.T) {
	e := newEnv(t)
	id, ana := e.seedCampaign(t)
	rec := e.do("POST", "/api/campaigns/"+id+"/invites", "", ana)
	token, _ := decodeBody(t, rec)["token"].(string)
	if rec.Code != 201 || token == "" {
		t.Fatalf("invite = %d %s", rec.Code, rec.Body)
	}
	if rec := e.do("GET", "/api/invites/"+token, "", nil); rec.Code != 200 ||
		!strings.Contains(rec.Body.String(), `"campaignName":"Teste"`) {
		t.Errorf("preview = %d %s", rec.Code, rec.Body)
	}
	body := `{"email":"new@x.com","displayName":"Novo","password":"new-password"}`
	rec = e.do("POST", "/api/invites/"+token+"/signup", body, nil)
	var cookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		cookie = c
	}
	if rec.Code != 201 || cookie == nil || cookie.Name != auth.CookieName {
		t.Fatalf("signup = %d %s", rec.Code, rec.Body)
	}
	if rec := e.do("GET", "/api/campaigns", "", cookie); !strings.Contains(rec.Body.String(), `"role":"player"`) {
		t.Errorf("new user list = %s", rec.Body)
	}
	rec = e.do("POST", "/api/invites/"+token+"/signup", `{"email":"b@x.com","displayName":"B","password":"password1"}`, nil)
	if rec.Code != 404 || decodeBody(t, rec)["error"] != "Convite inválido ou expirado" {
		t.Errorf("reused = %d %s", rec.Code, rec.Body)
	}
}

func TestSignupWithoutValidInviteBlocked(t *testing.T) {
	e := newEnv(t)
	body := `{"email":"new@x.com","displayName":"Novo","password":"new-password"}`
	for _, token := range []string{"missing", "also-missing"} {
		rec := e.do("POST", "/api/invites/"+token+"/signup", body, nil)
		if rec.Code != 404 && rec.Code != 405 {
			t.Errorf("token %q = %d", token, rec.Code)
		}
	}
	if _, _, _, err := e.auth.Login(context.Background(), "new@x.com", "new-password"); err == nil {
		t.Error("user was created without invite")
	}
}

func TestExpiredInvite(t *testing.T) {
	e := newEnv(t)
	id, ana := e.seedCampaign(t)
	token, _ := decodeBody(t, e.do("POST", "/api/campaigns/"+id+"/invites", "", ana))["token"].(string)
	e.now = e.now.Add(InviteTTL + time.Minute)
	rec := e.do("GET", "/api/invites/"+token, "", nil)
	if rec.Code != 404 || decodeBody(t, rec)["error"] != "Convite inválido ou expirado" {
		t.Errorf("expired preview = %d %s", rec.Code, rec.Body)
	}
	body := `{"email":"new@x.com","displayName":"Novo","password":"new-password"}`
	if rec := e.do("POST", "/api/invites/"+token+"/signup", body, nil); rec.Code != 404 {
		t.Errorf("expired signup = %d", rec.Code)
	}
}

func TestAcceptInviteAsLoggedUser(t *testing.T) {
	e := newEnv(t)
	id, ana := e.seedCampaign(t)
	token, _ := decodeBody(t, e.do("POST", "/api/campaigns/"+id+"/invites", "", ana))["token"].(string)
	bob := e.login(t, "bob")
	if rec := e.do("POST", "/api/invites/"+token+"/accept", "", nil); rec.Code != 401 {
		t.Errorf("anonymous accept = %d", rec.Code)
	}
	if rec := e.do("POST", "/api/invites/"+token+"/accept", "", bob); rec.Code != 200 {
		t.Fatalf("accept = %d %s", rec.Code, rec.Body)
	}
	if rec := e.do("GET", "/api/campaigns/"+id, "", bob); !strings.Contains(rec.Body.String(), `"role":"player"`) {
		t.Errorf("bob detail = %s", rec.Body)
	}
	if rec := e.do("POST", "/api/invites/"+token+"/accept", "", e.login(t, "cid")); rec.Code != 404 {
		t.Errorf("used invite = %d", rec.Code)
	}
}

func TestOnlyGMInvitesAndRemoves(t *testing.T) {
	e := newEnv(t)
	id, ana := e.seedCampaign(t)
	token, _ := decodeBody(t, e.do("POST", "/api/campaigns/"+id+"/invites", "", ana))["token"].(string)
	bob := e.login(t, "bob")
	e.do("POST", "/api/invites/"+token+"/accept", "", bob)
	if rec := e.do("POST", "/api/campaigns/"+id+"/invites", "", bob); rec.Code != 403 {
		t.Errorf("player invite = %d", rec.Code)
	}
	bobID, anaID := "00000000-0000-0000-0000-000000000002", "00000000-0000-0000-0000-000000000001"
	if rec := e.do("DELETE", "/api/campaigns/"+id+"/members/"+anaID, "", bob); rec.Code != 403 {
		t.Errorf("player remove = %d", rec.Code)
	}
	if rec := e.do("DELETE", "/api/campaigns/"+id+"/members/"+anaID, "", ana); rec.Code != 400 {
		t.Errorf("gm self remove = %d", rec.Code)
	}
	if rec := e.do("DELETE", "/api/campaigns/"+id+"/members/"+bobID, "", ana); rec.Code != 204 {
		t.Errorf("gm remove player = %d", rec.Code)
	}
	if rec := e.do("GET", "/api/campaigns/"+id, "", bob); rec.Code != 404 {
		t.Errorf("removed player detail = %d", rec.Code)
	}
}

func TestRoleIsPerCampaign(t *testing.T) {
	e := newEnv(t)
	id, ana := e.seedCampaign(t)
	bob := e.login(t, "bob")
	e.do("POST", "/api/campaigns", `{"name":"Do Bob","systemId":"assimilacao"}`, bob)
	token, _ := decodeBody(t, e.do("POST", "/api/campaigns/"+id+"/invites", "", ana))["token"].(string)
	e.do("POST", "/api/invites/"+token+"/accept", "", bob)
	rec := e.do("GET", "/api/campaigns", "", bob)
	if !strings.Contains(rec.Body.String(), `"role":"gm"`) || !strings.Contains(rec.Body.String(), `"role":"player"`) {
		t.Errorf("roles = %s", rec.Body)
	}
}

func TestUpdateSettingsGMOnlyAndValidated(t *testing.T) {
	e := newEnv(t)
	id, ana := e.seedCampaign(t)
	rec := e.do("PATCH", "/api/campaigns/"+id+"/settings", `{"creationLimits":"block"}`, ana)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"creationLimits":"block"`) {
		t.Errorf("gm = %d %s", rec.Code, rec.Body)
	}
	if rec := e.do("PATCH", "/api/campaigns/"+id+"/settings", `{"creationLimits":"nope"}`, ana); rec.Code != 400 {
		t.Errorf("invalid = %d", rec.Code)
	}
	if rec := e.do("PATCH", "/api/campaigns/"+id+"/settings", `{"creationLimits":"warn"}`, e.login(t, "bob")); rec.Code != 404 {
		t.Errorf("non-member = %d", rec.Code)
	}
}
