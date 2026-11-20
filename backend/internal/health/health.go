// Package health exposes the API health check.
package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Pinger is any dependency whose availability is part of the health check.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Status is the health check response body.
type Status struct {
	API      string `json:"api"`
	Database string `json:"database"`
}

const (
	ok      = "ok"
	failure = "error"
)

// Handler responds 200 when the API and database are up, and 503 when the database is unreachable.
func Handler(db Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		status := Status{API: ok, Database: ok}
		code := http.StatusOK
		if err := db.Ping(ctx); err != nil {
			status.Database = failure
			code = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(status)
	}
}
