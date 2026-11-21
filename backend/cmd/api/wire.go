package main

import (
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/auth"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/characters"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/conflict"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/drawings"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/effects"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/health"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/items"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/realtime"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/refuge"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/rolls"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/rounds"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/scenes"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/sessions"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/tokens"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/vision"
)

// deps carries the shared building blocks that table modules wire themselves onto.
type deps struct {
	mux         *http.ServeMux
	pool        *pgxpool.Pool
	hub         *realtime.Hub
	campaigns   *campaigns.Handlers
	campaignSvc *campaigns.Service
	auth        *auth.Service
	systems     *systems.Registry
	vision      *vision.Service
	effects     *effects.Service
}

func buildMux(pool *pgxpool.Pool) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /api/health", health.Handler(pool))
	authSvc := auth.NewService(auth.NewStore(pool))
	authHandlers := auth.NewHandlers(authSvc)
	authHandlers.Register(mux)
	reg := systems.Default()
	systems.Register(mux, reg, authHandlers.RequireAuth)
	campaignSvc := campaigns.NewService(campaigns.NewStore(pool), authSvc, reg)
	ch := campaigns.NewHandlers(campaignSvc, authHandlers.RequireAuth)
	ch.Register(mux)
	hub := realtime.NewHub(realtime.Options{OriginPatterns: splitList(os.Getenv("WS_ALLOWED_ORIGINS"))})
	hub.Register(mux, ch.RequireMember)
	wireTable(deps{mux: mux, pool: pool, hub: hub, campaigns: ch, campaignSvc: campaignSvc, auth: authSvc, systems: reg})
	return mux
}

// wireTable registers every game-table module; each module gets its own wireX function below.
func wireTable(d deps) {
	d.vision = vision.NewService(vision.NewStore(d.pool), scenes.NewService(scenes.NewStore(d.pool), d.hub), d.hub)
	d.effects = newEffects(d)
	wireRounds(d)
	wireScenes(d)
	wireCharacters(d)
	wireRolls(d)
	wireSessions(d)
	wireTokens(d)
	wireRefuge(d)
	wireDrawings(d)
	wireVision(d)
	wireConflict(d)
	wireEffects(d)
	wireItems(d)
}

func wireRounds(d deps) {
	svc := rounds.NewService(rounds.NewStore(d.pool))
	svc.OnPhaseChange(d.effects.StartOfTurn)
	rounds.Register(d.mux, d.hub, svc, d.campaigns.RequireMember)
}

func wireScenes(d deps) {
	svc := scenes.NewService(scenes.NewStore(d.pool), d.vision.Tap())
	scenes.NewHandlers(svc, d.campaigns).Register(d.mux)
	scenes.RegisterCommands(d.hub, svc)
}

func wireCharacters(d deps) {
	svc := characters.NewService(characters.NewStore(d.pool), d.campaignSvc, d.systems, d.hub)
	characters.NewHandlers(svc, d.campaigns.RequireMember).Register(d.mux)
}

func wireRolls(d deps) {
	chars := characters.NewService(characters.NewStore(d.pool), d.campaignSvc, d.systems, d.hub)
	svc := rolls.NewService(rolls.NewStore(d.pool), chars, sessions.NewService(sessions.NewStore(d.pool), d.hub), d.hub)
	rolls.RegisterCommands(d.hub, svc)
	rolls.RegisterHTTP(d.mux, svc, d.campaigns.RequireMember)
}

func wireSessions(d deps) {
	svc := sessions.NewService(sessions.NewStore(d.pool), d.hub)
	sessions.Register(d.mux, d.hub, svc, d.campaigns.RequireMember)
}

func wireTokens(d deps) {
	sceneSvc := scenes.NewService(scenes.NewStore(d.pool), d.hub)
	svc := tokens.NewService(tokens.NewStore(d.pool), sceneSvc, effects.NewTap(d.vision.Tap(), d.effects))
	svc.SetVisibility(d.vision)
	tokens.Register(d.mux, svc, d.campaigns)
	tokens.RegisterCommands(d.hub, svc)
}

func wireRefuge(d deps) {
	chars := characters.NewService(characters.NewStore(d.pool), d.campaignSvc, d.systems, d.hub)
	refuge.Register(d.mux, d.hub, refuge.NewService(refuge.NewStore(d.pool), chars, d.hub), d.campaigns)
}

func wireDrawings(d deps) {
	sceneSvc := scenes.NewService(scenes.NewStore(d.pool), d.hub)
	svc := drawings.NewService(drawings.NewStore(d.pool), sceneSvc, d.hub)
	drawings.Register(d.mux, svc, d.campaigns)
	drawings.RegisterCommands(d.hub, svc)
}

func wireVision(d deps) {
	vision.RegisterHTTP(d.mux, d.vision, d.campaigns)
	vision.RegisterCommands(d.hub, d.vision)
}

func wireConflict(d deps) {
	chars := characters.NewService(characters.NewStore(d.pool), d.campaignSvc, d.systems, d.hub)
	rollSvc := rolls.NewService(rolls.NewStore(d.pool), chars, sessions.NewService(sessions.NewStore(d.pool), d.hub), d.hub)
	svc := conflict.NewService(conflict.NewStore(d.pool), rollSvc, chars, d.hub)
	conflict.RegisterCommands(d.hub, svc)
	conflict.RegisterHTTP(d.mux, svc, d.campaigns.RequireMember)
}

func newEffects(d deps) *effects.Service {
	chars := characters.NewService(characters.NewStore(d.pool), d.campaignSvc, d.systems, d.hub)
	rollSvc := rolls.NewService(rolls.NewStore(d.pool), chars, sessions.NewService(sessions.NewStore(d.pool), d.hub), d.hub)
	sceneSvc := scenes.NewService(scenes.NewStore(d.pool), d.hub)
	return effects.NewService(effects.NewStore(d.pool), sceneSvc, tokens.NewStore(d.pool), chars, rollSvc, d.hub)
}

func wireEffects(d deps) {
	effects.RegisterHTTP(d.mux, d.effects, d.campaigns)
	effects.RegisterCommands(d.hub, d.effects)
}

func wireItems(d deps) {
	chars := characters.NewService(characters.NewStore(d.pool), d.campaignSvc, d.systems, d.hub)
	rf := refuge.NewService(refuge.NewStore(d.pool), chars, d.hub)
	sceneSvc := scenes.NewService(scenes.NewStore(d.pool), d.hub)
	tokenSvc := tokens.NewService(tokens.NewStore(d.pool), sceneSvc, d.vision.Tap())
	svc := items.NewService(items.NewStore(d.pool), chars, rf, d.effects, tokenSvc, sceneSvc, d.hub)
	items.RegisterHTTP(d.mux, svc, d.campaigns)
	items.RegisterCommands(d.hub, svc)
}
