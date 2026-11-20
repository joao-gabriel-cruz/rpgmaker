package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakePinger struct{ err error }

func (f fakePinger) Ping(context.Context) error { return f.err }

func TestHandler(t *testing.T) {
	cases := []struct {
		name     string
		pingErr  error
		wantCode int
		want     Status
	}{
		{"database up", nil, http.StatusOK, Status{API: "ok", Database: "ok"}},
		{"database down", errors.New("connection refused"), http.StatusServiceUnavailable, Status{API: "ok", Database: "error"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			Handler(fakePinger{tc.pingErr})(rec, httptest.NewRequest(http.MethodGet, "/api/health", nil))

			if rec.Code != tc.wantCode {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantCode)
			}
			var got Status
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("invalid body: %v", err)
			}
			if got != tc.want {
				t.Fatalf("body = %+v, want %+v", got, tc.want)
			}
		})
	}
}
