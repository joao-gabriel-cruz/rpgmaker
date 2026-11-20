package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newServer(t *testing.T) (http.Handler, *Service) {
	t.Helper()
	svc, _ := seeded(t)
	if _, err := svc.CreateUser(context.Background(), "bob@x.com", "Bob", "bob-password", false); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	NewHandlers(svc).Register(mux)
	return mux, svc
}

func do(h http.Handler, method, path, body string, c *http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if c != nil {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func loginCookie(t *testing.T, h http.Handler, email, pass string) *http.Cookie {
	t.Helper()
	rec := do(h, "POST", "/api/auth/login", `{"email":"`+email+`","password":"`+pass+`"}`, nil)
	for _, c := range rec.Result().Cookies() {
		if c.Name == CookieName {
			return c
		}
	}
	t.Fatalf("no cookie, status %d", rec.Code)
	return nil
}

func TestLoginWrongPasswordGenericMessage(t *testing.T) {
	h, _ := newServer(t)
	for _, body := range []string{
		`{"email":"ana@example.com","password":"bad"}`,
		`{"email":"ghost@example.com","password":"secret-pass"}`,
	} {
		rec := do(h, "POST", "/api/auth/login", body, nil)
		var out map[string]string
		_ = json.Unmarshal(rec.Body.Bytes(), &out)
		if rec.Code != 401 || out["error"] != "E-mail ou senha inválidos" {
			t.Errorf("%s -> %d %v", body, rec.Code, out)
		}
	}
}

func TestLoginSetsPersistentHttpOnlyCookie(t *testing.T) {
	h, _ := newServer(t)
	c := loginCookie(t, h, "ana@example.com", "secret-pass")
	if !c.HttpOnly || c.Expires.IsZero() || c.SameSite != http.SameSiteLaxMode {
		t.Errorf("bad cookie %+v", c)
	}
}

func TestMeAndLogout(t *testing.T) {
	h, _ := newServer(t)
	if rec := do(h, "GET", "/api/auth/me", "", nil); rec.Code != 401 {
		t.Errorf("anonymous me = %d", rec.Code)
	}
	c := loginCookie(t, h, "ana@example.com", "secret-pass")
	rec := do(h, "GET", "/api/auth/me", "", c)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"displayName":"Ana"`) ||
		strings.Contains(rec.Body.String(), "argon2") {
		t.Errorf("me = %d %s", rec.Code, rec.Body)
	}
	if rec := do(h, "POST", "/api/auth/logout", "", c); rec.Code != 204 {
		t.Errorf("logout = %d", rec.Code)
	}
	if rec := do(h, "GET", "/api/auth/me", "", c); rec.Code != 401 {
		t.Errorf("me after logout = %d", rec.Code)
	}
}

func TestAdminEndpointsRequireAdmin(t *testing.T) {
	h, _ := newServer(t)
	if rec := do(h, "GET", "/api/admin/users", "", nil); rec.Code != 401 {
		t.Errorf("anonymous = %d", rec.Code)
	}
	bob := loginCookie(t, h, "bob@x.com", "bob-password")
	if rec := do(h, "GET", "/api/admin/users", "", bob); rec.Code != 403 {
		t.Errorf("non-admin = %d", rec.Code)
	}
	ana := loginCookie(t, h, "ana@example.com", "secret-pass")
	if rec := do(h, "GET", "/api/admin/users", "", ana); rec.Code != 200 {
		t.Errorf("admin list = %d", rec.Code)
	}
}

func TestAdminResetPassword(t *testing.T) {
	h, _ := newServer(t)
	ana := loginCookie(t, h, "ana@example.com", "secret-pass")
	rec := do(h, "POST", "/api/admin/users/id-bob@x.com/reset-password", "", ana)
	var out map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	if rec.Code != 200 || len(out["temporaryPassword"]) != 12 {
		t.Fatalf("reset = %d %s", rec.Code, rec.Body)
	}
	loginCookie(t, h, "bob@x.com", out["temporaryPassword"])
	if rec := do(h, "POST", "/api/admin/users/nope/reset-password", "", ana); rec.Code != 404 {
		t.Errorf("missing user = %d", rec.Code)
	}
}
