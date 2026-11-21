package main

import (
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/auth"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/health"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/rounds"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems"
)

// deps carries the shared building blocks that table modules wire themselves onto.
type deps struct {
	mux       *http.ServeMux
	pool      *pgxpool.Pool
	hub       *realtime.Hub
	campaigns *campaigns.Handlers
	auth      *auth.Service
	systems   *systems.Registry
}

func buildMux(pool *pgxpool.Pool) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /api/health", health.Handler(pool))
	authSvc := auth.NewService(auth.NewStore(pool))
	authHandlers := auth.NewHandlers(authSvc)
	authHandlers.Register(mux)
	reg := systems.Default()
	systems.Register(mux, reg, authHandlers.RequireAuth)
	ch := campaigns.NewHandlers(campaigns.NewService(campaigns.NewStore(pool), authSvc, reg), authHandlers.RequireAuth)
	ch.Register(mux)
	hub := realtime.NewHub(realtime.Options{OriginPatterns: splitList(os.Getenv("WS_ALLOWED_ORIGINS"))})
	hub.Register(mux, ch.RequireMember)
	wireTable(deps{mux: mux, pool: pool, hub: hub, campaigns: ch, auth: authSvc, systems: reg})
	return mux
}

// wireTable registers every game-table module; each module gets its own wireX function below.
func wireTable(d deps) {
	wireRounds(d)
	wireScenes(d)
}

func wireRounds(d deps) {
	rounds.Register(d.mux, d.hub, rounds.NewService(rounds.NewStore(d.pool)), d.campaigns.RequireMember)
}

func wireScenes(d deps) {
	svc := scenes.NewService(scenes.NewStore(d.pool), d.hub)
	scenes.NewHandlers(svc, d.campaigns).Register(d.mux)
	scenes.RegisterCommands(d.hub, svc)
}
