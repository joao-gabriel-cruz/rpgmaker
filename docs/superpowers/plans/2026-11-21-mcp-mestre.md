# MCP do Mestre — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** The Mestre connects their own Claude to a campaign through a secret link and changes the live table (terrain, objects, tokens, ambience, conditions, scenes) just by narrating, with one-click undo.

**Architecture:** New Go package `backend/internal/mcp` serves a stateless Streamable HTTP MCP endpoint at `/api/mcp/{secret}` using the official Go SDK. Tools call the *same* `scenes`/`tokens`/`ambience` service instances the editor uses, so broadcasts, locks, fog and terrain effects behave exactly as for manual edits. An in-memory journal keeps the last 20 Claude changes per campaign with their inverses; the table shows them to the Mestre with "Desfazer".

**Tech Stack:** Go 1.26, `github.com/modelcontextprotocol/go-sdk v1.8.0` (`mcp` package), pgx, goose; React + TS + Tailwind + react-query, Vitest.

**Spec:** `docs/specs/0004-mcp-mestre.md`

## Global Constraints

- File ≤ 300 lines, function ≤ 30 lines (commit hook blocks); comments only for non-obvious *why*, ≤ 2 lines.
- Code, identifiers, logs, errors-in-code, test names and commit titles in English; every user-visible string (UI and tool error/summary text Claude relays to the Mestre) in PT-BR.
- HTTP errors as `{"error": "<PT-BR>"}`; routes registered with `"METHOD /api/..."` patterns; GM routes behind `campaigns.Handlers.RequireGM`.
- Wiring only in `backend/cmd/api/wire.go` via `wireMCP(d deps)` + one line in `wireTable`.
- Migration file `backend/internal/migrations/sql/NNNNN_mcp_links.sql` where NNNNN is the GitHub issue number of this work (use `00069` if there is none); never edit merged migrations.
- Random secrets from `crypto/rand`; only the sha256 hash of the secret is stored.
- Limits: `paint_terrain` ≤ 4096 cells per call; `place_objects` ≤ 200 objects per call; journal keeps the newest 20 changes per campaign; heights clamp to 0–10; circle radius ≤ 32; line width 1–5.
- Terrain ids: `floor grass dirt sand water stone snow lava ruin`. Times of day: `dawn day dusk night`. Weathers: `clear rain storm fog snow ash`. Conditions: `assimilacao.Conditions`.
- Write tools act on the **active scene**; if it has an open draft they refuse with `há um rascunho aberto nesta cena; aplique ou descarte no editor antes`.
- `make check` green; `docker compose up --build` still boots.

## Review Focus

1. **Shape partly off the map** ("lava opens around Junior" near an edge) → clip to the grid silently; only a shape with zero in-grid cells is an error. *(test in Task 3)*
2. **Ambiguous or differently-cased character name** ("junior" vs "Junior", two "Ana") → case-insensitive, accent-sensitive exact match on name, then token id; ambiguity returns an error listing the candidates. *(test in Task 8)*
3. **Blocking object dropped on a token** ("wall to defend Junior" placed on Junior's cell) → refuse that placement with `não dá para colocar <prefab> em (x,z): há um token ali`. *(test in Task 7)*
4. **Link used after revocation, regeneration, or the user losing the Mestre role** → 404 with no hint the link existed. *(test in Task 2)*
5. **Undo of an already-undone change, or of a change whose scene was deleted** → PT-BR user error, journal stays consistent, no panic; undo of a change in a now-inactive scene still applies to that scene. *(test in Task 5)*

---

## File Structure

Backend `backend/internal/mcp/`:
- `links.go` — `Link`, `Session`, `Links` service (generate/status/revoke/resolve).
- `links_store.go` — pgx `LinkRepo`.
- `links_http.go` — REST for the link + recent changes.
- `shapes.go` — `Shape` + `Rasterize` (pure).
- `textmap.go` — text rendering of a scene for `get_table` (pure).
- `ports.go` — narrow interfaces over scenes/tokens/ambience/characters + `Table` bundle.
- `journal.go` — change log, inverses, undo.
- `undo_cmd.go` — realtime `mcp.undo` command.
- `server.go` — HTTP handler resolving the secret and serving the SDK handler; `newServer`.
- `tools_read.go`, `tools_terrain.go`, `tools_objects.go`, `tools_tokens.go`, `tools_scene.go` — tool registration and handlers.
- `resolve.go` — active scene lookup, draft guard, character/token name resolution.
- Tests next to each file; shared fakes in `fake_test.go`, MCP client helper in `client_test.go`.

Frontend `frontend/src/features/mcp/`: `mcp.ts`, `ClaudePanel.tsx`, `ClaudeChangesList.tsx`, `ClaudeChangeToast.tsx`, `mcp.test.tsx`.

---

### Task 1: Share the editor's service instances in wiring

**Files:**
- Modify: `backend/cmd/api/wire.go` (`deps`, `wireTable`, `wireScenes`, `wireTokens`, `wireAmbience`)

**Interfaces:**
- Produces: `deps.scenes *scenes.Service` (the one built with `d.vision.Tap()`), `deps.tokens *tokens.Service` (the one with the effects tap + `SetVisibility(d.vision)`), `deps.ambience *ambience.Service`. `wireScenes/wireTokens/wireAmbience` use these instead of constructing their own.

- [ ] **Step 1:** In `wireTable`, after `d.rolls`, build `d.scenes`, `d.tokens`, `d.ambience` exactly as `wireScenes`, `wireTokens`, `wireAmbience` build them today; make those three functions use the fields. Other modules keep their private instances (out of scope).
- [ ] **Step 2:** Run `cd backend && go build ./... && go test ./...` → PASS, no behavior change.
- [ ] **Step 3:** Commit `refactor: share scene, token and ambience services in wiring`.

---

### Task 2: Mestre link (migration, store, service, REST)

**Files:**
- Create: `backend/internal/migrations/sql/NNNNN_mcp_links.sql`, `backend/internal/mcp/links.go`, `links_store.go`, `links_http.go`
- Test: `backend/internal/mcp/links_test.go`, `links_store_test.go` (skips without `TEST_DATABASE_URL`)

**Interfaces:**
- Produces:
  ```go
  type Session struct{ CampaignID string; Sender realtime.Sender } // Sender{UserID: gm, Role: campaigns.RoleGM, Name: "Claude"}
  type Link struct{ ID, CampaignID, UserID string; CreatedAt time.Time; LastUsedAt *time.Time }
  var ErrInvalidLink, ErrNoLink error
  type LinkRepo interface {
    Replace(ctx context.Context, campaignID, userID, hash string) (Link, error) // revoke active + insert, one tx
    Active(ctx context.Context, campaignID, userID string) (Link, error)        // ErrNoLink
    ByHash(ctx context.Context, hash string) (Link, error)                      // only non-revoked; ErrInvalidLink
    Revoke(ctx context.Context, campaignID, userID string) error
    Touch(ctx context.Context, id string) error
  }
  type Members interface{ Membership(ctx context.Context, campaignID, userID string) (campaigns.Member, error) }
  func NewLinks(repo LinkRepo, m Members) *Links
  func (l *Links) Generate(ctx context.Context, m campaigns.Member) (secret string, link Link, err error)
  func (l *Links) Status(ctx context.Context, m campaigns.Member) (Link, bool, error)
  func (l *Links) Revoke(ctx context.Context, m campaigns.Member) error
  func (l *Links) Resolve(ctx context.Context, secret string) (Session, error) // ErrInvalidLink
  ```
  REST (all `RequireGM`): `GET /api/campaigns/{id}/mcp-link` → `{exists, createdAt?, lastUsedAt?}`; `POST` → 201 `{url, createdAt}` where `url = <scheme>://<host>/api/mcp/<secret>` (scheme from `X-Forwarded-Proto`, else TLS); `DELETE` → 204. Route registration function: `func RegisterLinkHTTP(mux *http.ServeMux, l *Links, cmp *campaigns.Handlers)`.

- [ ] **Step 1: Migration**
  ```sql
  -- +goose Up
  CREATE TABLE mcp_links (
      id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
      campaign_id  uuid NOT NULL REFERENCES campaigns (id) ON DELETE CASCADE,
      user_id      uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
      secret_hash  text NOT NULL UNIQUE,
      created_at   timestamptz NOT NULL DEFAULT now(),
      last_used_at timestamptz,
      revoked_at   timestamptz
  );
  CREATE UNIQUE INDEX mcp_links_one_active ON mcp_links (campaign_id, user_id) WHERE revoked_at IS NULL;
  -- +goose Down
  DROP TABLE mcp_links;
  ```
  (Confirm the users table name in earlier migrations.)
- [ ] **Step 2: Failing service tests** (fake `LinkRepo`/`Members`):
  - `TestGenerateReturnsSecretOnceAndStoresHash`: secret is 43 chars base64url; repo stored `sha256hex(secret)`, never the secret.
  - `TestGenerateRevokesPrevious`: second Generate → resolving the first secret returns `ErrInvalidLink`.
  - `TestResolveRejectsRevokedUnknownAndNonGM`: after `Revoke`, unknown secret, and Membership returning `RoleMember` → all `ErrInvalidLink`.
  - `TestResolveBuildsClaudeSender`: Session `{CampaignID, Sender{UserID, Role: RoleGM, Name: "Claude"}}`; `Touch` called.
  - `TestGenerateRequiresGM`: member with `RoleMember` → error.
- [ ] **Step 3:** Run `go test ./internal/mcp/ -run Link` → FAIL (undefined).
- [ ] **Step 4:** Implement `links.go`, `links_store.go`, `links_http.go`.
- [ ] **Step 5:** Store test `TestLinkStoreReplaceKeepsOneActive` against Postgres (skips without `TEST_DATABASE_URL`).
- [ ] **Step 6:** `go test ./internal/mcp/...` → PASS. Commit `feat: mestre MCP link storage and endpoints`.

---

### Task 3: Shape rasterizer

**Files:** Create `backend/internal/mcp/shapes.go`; Test `shapes_test.go`

**Interfaces:**
- Produces:
  ```go
  type Point struct{ X int `json:"x"`; Z int `json:"z"` }
  type Shape struct {
    Kind   string  `json:"kind"`             // rect | circle | line | cells
    From   *Point  `json:"from,omitempty"`   // rect corner / line start / circle center
    To     *Point  `json:"to,omitempty"`     // rect opposite corner / line end
    Radius int     `json:"radius,omitempty"` // circle, 0–32
    Width  int     `json:"width,omitempty"`  // line thickness 1–5, default 1
    Cells  []Point `json:"cells,omitempty"`
  }
  func Rasterize(s Shape, w, d int) ([]maps.Coord, error) // clipped, deduped, row-major order
  ```
  Errors (PT-BR): unknown kind, missing `from`/`to`, radius/width out of range, zero cells inside the grid (`a forma fica toda fora do mapa`), more than 4096 cells.

- [ ] **Step 1: Failing tests** (table-driven):
  - rect `(2,3)-(4,4)` on 10×10 → 6 cells; corners given in any order give the same result.
  - circle center `(5,5)` radius 1 → 5 cells (`dx²+dz² ≤ r²`); radius 0 → 1 cell.
  - line `(0,0)-(4,2)` width 1 → Bresenham cells, 5 cells; width 3 adds every in-grid cell within Chebyshev distance 1 of a line cell.
  - circle center `(0,0)` radius 3 on 10×10 → only in-grid cells, no error (**Review Focus 1**).
  - rect `(20,20)-(25,25)` on 10×10 → error `a forma fica toda fora do mapa`.
  - cells with duplicates → deduped.
  - rect 70×70 clipped to 64×64 = 4096 → ok; `cells` list of 4097 distinct in-grid cells on a 64×64 → error.
- [ ] **Step 2:** `go test ./internal/mcp/ -run Rasterize` → FAIL.
- [ ] **Step 3:** Implement.
- [ ] **Step 4:** PASS. Commit `feat: rasterize MCP terrain shapes`.

---

### Task 4: Text map and catalog data

**Files:**
- Create `backend/internal/mcp/textmap.go`; Test `textmap_test.go`
- Modify `backend/internal/maps/objects.go` (add `func PrefabRules() map[string]PrefabRule`, returns a copy); `backend/internal/ambience/model.go` (add `func TimesOfDay() []string`, `func Weathers() []string` returning copies)

**Interfaces:**
- Produces:
  ```go
  type TokenInfo struct{ ID, Name string; X, Z int; Player bool; Conditions []string }
  type TableView struct {
    Scene    scenes.Scene
    Ambience ambience.Ambience
    Tokens   []TokenInfo
  }
  func RenderTable(v TableView) (string, error)
  func RenderCatalog() string
  ```
  `RenderTable` output, in order: header (name, `W×D`, time of day, weather); axis line `x cresce para a direita (leste); z cresce para baixo (sul); norte = z menor`; legend; terrain grid (one line per z, prefixed `zNN `, exactly W glyphs); a heights grid with the same layout (digits `0`–`9`, `X` = 10) **only if any height ≠ 0**; token list `- <name> (jogador|NPC) em (x,z) [condições]`; object list `- <id> <prefab> em (x,z) rot <r>[ aberta]`.
  Terrain glyphs: `floor . grass " dirt : sand s water ~ stone # snow * lava ! ruin %`.
  `RenderCatalog`: terrains with blocks-movement/vision flags, prefabs with footprint/height/blocks, times, weathers, conditions.

- [ ] **Step 1: Failing tests:**
  - 3×2 grid all floor → 2 grid lines `z00 ...`, `z01 ...`; no heights block.
  - one lava cell at (1,0) height 2 → `z00 .!.` and heights line `z00 020`.
  - token and open door appear in the lists with coordinates; NPC labelled `NPC`.
  - `RenderCatalog` contains every terrain id, every prefab from `maps.PrefabRules()`, and every condition.
- [ ] **Step 2:** FAIL. **Step 3:** Implement (verify against the frontend camera that `z` smaller is the side the default camera shows as "top"; if not, fix only the axis wording).
- [ ] **Step 4:** PASS. Commit `feat: text rendering of the table for MCP`.

---

### Task 5: Ports, journal, undo command and recent-changes endpoint

**Files:**
- Create `backend/internal/mcp/ports.go`, `journal.go`, `undo_cmd.go`; Modify `links_http.go` (add changes route)
- Test `journal_test.go`, `fake_test.go`

**Interfaces:**
- Produces:
  ```go
  type Scenes interface {
    List(ctx context.Context, campaignID string) ([]scenes.Summary, error)
    Get(ctx context.Context, campaignID, id, role string) (scenes.Scene, error)
    Paint(ctx context.Context, campaignID, sceneID, origin string, p maps.Patch) error
    EditObjects(ctx context.Context, campaignID, sceneID, origin string, p scenes.ObjectPatch) error
    Create(ctx context.Context, campaignID, name string, w, d int) (scenes.Scene, error)
    Activate(ctx context.Context, campaignID, id string) error
  }
  type Tokens interface {
    List(ctx context.Context, m campaigns.Member, sceneID string) ([]tokens.Token, error)
    Move(ctx context.Context, snd realtime.Sender, campaignID, tokenID string, x, z int, force bool) error
    Place(ctx context.Context, snd realtime.Sender, campaignID, sceneID, charID string, x, z int) error
  }
  type Ambience interface {
    Get(ctx context.Context, campaignID, sceneID string) (ambience.Ambience, error)
    Set(ctx context.Context, campaignID string, a ambience.Ambience) error
    SetCondition(ctx context.Context, snd realtime.Sender, campaignID, charID, cond string, active bool) error
  }
  type Characters interface{ List(ctx context.Context, m campaigns.Member) ([]characters.View, error) }
  type Table struct{ Scenes Scenes; Tokens Tokens; Ambience Ambience; Characters Characters }

  type Inverse struct {
    Cells     []maps.CellPatch     // previous terrain+height
    Objects   *scenes.ObjectPatch  // patch that restores previous objects
    Token     *TokenPos            // {TokenID string; X, Z int}
    Ambience  *ambience.Ambience
    Condition *ConditionState      // {CharacterID, Condition string; Active bool}
  }
  type ChangeView struct{ ID, SceneID, Summary string; At time.Time } // json: id, sceneId, summary, at
  type Notifier interface{ Broadcast(campaignID string, to realtime.Filter, typ string, payload any) }
  func NewJournal(n Notifier) *Journal
  func (j *Journal) Record(campaignID, sceneID, summary string, inv Inverse) ChangeView
  func (j *Journal) List(campaignID string) []ChangeView                 // newest first
  func (j *Journal) Undo(ctx context.Context, t Table, s Session, id string) (string, error) // id "" = newest; returns summary
  func RegisterUndo(r scenes.Registrar, j *Journal, t Table)              // command "mcp.undo" {changeId}, realtime.GMOnly
  ```
  Event `mcp.changed` (ToGM): `{changes: []ChangeView, latest?: ChangeView}` — `latest` only on `Record`, absent after undo. `GET /api/campaigns/{id}/mcp-changes` (RequireGM) → `{changes}`, registered by `func RegisterChangesHTTP(mux *http.ServeMux, j *Journal, cmp *campaigns.Handlers)`.
  Undo applies, in this order and only the present fields: `Cells` via `Paint`, `Objects` via `EditObjects`, `Token` via `Move(force=true)`, `Ambience` via `Set`, `Condition` via `SetCondition`; scene edits use origin `""`. It refuses (keeping the entry) when the change's scene has an open draft. Missing id → `UserError("Essa mudança não existe mais (talvez já tenha sido desfeita)")`; scene gone → entry dropped and `UserError("A cena dessa mudança não existe mais")`.

- [ ] **Step 1: Failing tests** with fakes in `fake_test.go` (in-memory scene with grid/objects/draft, tokens, ambience, characters, recording notifier):
  - `TestRecordCapsAtTwenty`: 21 records → `List` has 20, newest first, oldest dropped.
  - `TestRecordBroadcastsLatestToGM`: notifier got `mcp.changed` with `latest` and filter `realtime.ToGM`.
  - `TestUndoRestoresCellsAndObjects`: inverse applied through fake Scenes, entry removed, broadcast without `latest`.
  - `TestUndoAnyOrder`: undo the 2nd of 3 entries leaves 1st and 3rd.
  - `TestUndoTwiceFails` and `TestUndoDeletedScene` (**Review Focus 5**).
  - `TestUndoInactiveSceneStillApplies` (**Review Focus 5**).
  - `TestUndoRefusesOpenDraft`.
- [ ] **Step 2:** FAIL. **Step 3:** Implement. **Step 4:** PASS.
- [ ] **Step 5:** Commit `feat: journal and undo for MCP changes`.

---

### Task 6: MCP endpoint and read tools

**Files:**
- Create `backend/internal/mcp/server.go`, `resolve.go`, `tools_read.go`; Test `client_test.go`, `tools_read_test.go`
- Modify `backend/cmd/api/wire.go` (`wireMCP`), `backend/go.mod`/`go.sum` (`go get github.com/modelcontextprotocol/go-sdk@v1.8.0`), `frontend/nginx.conf`

**Interfaces:**
- Consumes: `Links.Resolve` (Task 2), `RenderTable`/`RenderCatalog` (Task 4), `Table`, `Journal` (Task 5).
- Produces:
  ```go
  func NewHandler(l *Links, t Table, j *Journal) http.Handler     // resolves r.PathValue("secret"); 404 {"error":"Link inválido"} on ErrInvalidLink
  func newServer(t Table, j *Journal, s Session) *mcp.Server       // registers every tool; Implementation{Name: "rpgmaker-mesa", Version: "1"}
  func activeScene(ctx context.Context, t Table, s Session) (scenes.Scene, error) // UserError-style message "nenhuma cena ativa"
  func requireNoDraft(sc scenes.Scene) error
  func tableView(ctx context.Context, t Table, s Session, sc scenes.Scene) (TableView, error)
  ```
  Handler: `http.MaxBytesReader` 1 MiB; SDK handler `mcp.NewStreamableHTTPHandler(getServer, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true})`, where `getServer` reads the `Session` the outer handler put in the request context. Route: `mux.Handle("/api/mcp/{secret}", h)` (all methods). Tool handlers are `mcp.ToolHandlerFor[In, any]` returning a single text content; domain/user errors are returned as tool errors (`IsError`) carrying the PT-BR message.
  Read tools: `get_table {}` → `RenderTable` of the active scene (no active scene → text explaining there is none plus `list_scenes` hint, not an error); `list_scenes {}` → one line per scene `id · nome · WxD · ativa?`; `get_catalog {}` → `RenderCatalog()`. Tool descriptions (English) explain coordinates (x→east, z→south), that changes are live for players, and that every write can be undone.
  `wireMCP(d deps)`: `j := mcp.NewJournal(d.hub)`, `t := mcp.Table{d.scenes, d.tokens, d.ambience, characters.NewService(...)}`, register handler, link REST, `RegisterUndo(d.hub, j, t)`.
  nginx: add before `location /api/` a `location /api/mcp/ { access_log off; ...same proxy directives... }`.

- [ ] **Step 1:** `go get github.com/modelcontextprotocol/go-sdk@v1.8.0`.
- [ ] **Step 2: `client_test.go` helper** `callTool(t *testing.T, tbl Table, j *Journal, name string, args any) (text string, isErr bool)` using `mcp.NewInMemoryTransports()` + `mcp.NewClient` against `newServer`.
- [ ] **Step 3: Failing tests:**
  - `TestGetTableRendersActiveScene` (text contains the scene name and a `z00 ` line).
  - `TestGetTableWithoutActiveScene` (not an error; mentions `list_scenes`).
  - `TestHandlerRejectsUnknownSecret` via `httptest`: POST `/api/mcp/nope` → 404.
  - `TestHandlerServesInitialize` via `httptest` with a fake resolving `Links`: JSON-RPC `initialize` POST → 200 `application/json`.
- [ ] **Step 4:** FAIL. **Step 5:** Implement + wire + nginx. **Step 6:** `cd backend && go test ./... && go build ./...` → PASS.
- [ ] **Step 7:** Commit `feat: MCP endpoint with table read tools`.

---

### Task 7: Terrain and object tools

**Files:** Create `backend/internal/mcp/tools_terrain.go`, `tools_objects.go`; Test `tools_terrain_test.go`, `tools_objects_test.go`

**Interfaces:**
- Consumes: `Rasterize`, `activeScene`, `requireNoDraft`, `Journal.Record`.
- Tools and inputs:
  - `paint_terrain {shape: Shape, terrain?: string, height?: int, heightDelta?: int}` — at least one of the three; per cell new terrain = `terrain` or current, new height = clamp(`height` or current+`heightDelta`, 0, 10). Builds `maps.Patch`, inverse = previous cells. Summary: `<n> células → <terreno PT-BR>[, altura …] perto de (x,z)` (centroid).
  - `place_objects {objects?: [{prefab, x, z, rotation?, props?}], line?: {prefab, from: Point, to: Point}}` — ids `mcp-` + 12 hex from `crypto/rand`; `line` places one object per Bresenham cell and only for 1×1 prefabs; ≤ 200 total; a prefab whose `maps.PrefabRules()` rule blocks movement may not cover a token's cell (**Review Focus 3**). Inverse = remove the new ids.
  - `remove_objects {ids?: []string, area?: Shape}` — area removes objects whose `maps.Footprint` intersects it; inverse = add them back.
  - `set_door {objectId, open: bool}` — only prefab `door`; inverse = update with previous props.
  All refuse on open draft; all return the summary as text.

- [ ] **Step 1: Failing tests:**
  - `TestPaintLavaCircleIsLiveAndUndoable`: Paint called with origin `""`, journal has 1 entry, `undo_last` restores (after Task 8 adds `undo_last`, keep the undo assertion via `Journal.Undo` here).
  - `TestPaintHeightDeltaClamps`: height 9 + delta 3 → 10.
  - `TestPaintRequiresSomething`: no terrain/height/delta → tool error.
  - `TestPaintRefusesDraft`: draft scene → tool error with the draft message.
  - `TestPlaceWallLine`: line `(2,2)-(2,5)` wall → 4 objects with `mcp-` ids.
  - `TestPlaceBlockingOnTokenRefused` (**Review Focus 3**); bush (non-blocking) on token cell is allowed.
  - `TestRemoveByAreaAndUndo`, `TestSetDoorOnlyDoors`.
- [ ] **Step 2:** FAIL. **Step 3:** Implement. **Step 4:** PASS.
- [ ] **Step 5:** Commit `feat: MCP terrain and object tools`.

---

### Task 8: Token, ambience, condition and undo tools

**Files:** Create `backend/internal/mcp/tools_tokens.go`; Modify `resolve.go`; Test `tools_tokens_test.go`

**Interfaces:**
- Produces in `resolve.go`: `func findToken(list []tokens.Token, ref string) (tokens.Token, error)` and `func findCharacter(list []characters.View, ref string) (characters.View, error)` — exact id first, then `strings.EqualFold` on name; none → `não encontrei "<ref>"`; several → `"<ref>" é ambíguo: <names with ids>`.
- Tools:
  - `move_token {token: string, x, z}` — target must pass `maps.CanEnter(g, sc.Objects, to, to, 0)` (else its PT-BR reason as tool error), then `Tokens.Move(..., force=true)`; inverse = previous position.
  - `set_ambience {timeOfDay?, weather?}` — merges with current; inverse = previous ambience.
  - `set_condition {character, condition, active}` — reads current state from the character sheet's `conditions`; no change → text `já estava assim`, nothing recorded; inverse = previous state.
  - `undo_last {}` — `Journal.Undo(ctx, t, s, "")`, returns `desfeito: <summary>`.

- [ ] **Step 1: Failing tests:** `TestMoveTokenByNameCaseInsensitive`, `TestMoveTokenAmbiguousName` (**Review Focus 2**), `TestMoveTokenIntoBlockedCellRefused`, `TestSetAmbienceMergesAndUndo`, `TestSetConditionNoopNotRecorded`, `TestUndoLastEmptyJournal` (tool error).
- [ ] **Step 2:** FAIL. **Step 3:** Implement. **Step 4:** PASS.
- [ ] **Step 5:** Commit `feat: MCP token, ambience and condition tools`.

---

### Task 9: Scene tools (create, activate, bring the party)

**Files:** Create `backend/internal/mcp/tools_scene.go`; Test `tools_scene_test.go`

**Interfaces:**
- Tools (not recorded in the journal):
  - `create_scene {name, width, depth, baseTerrain?}` — `Scenes.Create`; if `baseTerrain` ≠ `floor`, one `Paint` over every cell (allowed above the 4096 tool limit since it is server-built). Returns id and name; the scene stays inactive.
  - `activate_scene {scene: string (id or name), arrival?: Point}` — `Scenes.Activate`, then for every character with `OwnerID != nil` that has no token in that scene: `Tokens.Place` on the free cell nearest to `arrival` (default center), searching rings by Chebyshev distance, skipping cells failing `maps.CanEnter` or already holding a token. Characters already present stay. Returns who was brought and where; if no free cell remains, reports who could not be placed.

- [ ] **Step 1: Failing tests:** `TestCreateSceneWithBaseTerrain`, `TestActivateBringsPlayersNearArrival` (2 players, 1 NPC → 2 tokens placed on distinct free cells, NPC not placed), `TestActivateKeepsExistingTokens`, `TestActivateSkipsBlockedCells`.
- [ ] **Step 2:** FAIL. **Step 3:** Implement. **Step 4:** PASS.
- [ ] **Step 5:** Commit `feat: MCP scene tools that bring the party`.

---

### Task 10: "Claude" settings tab with link and recent changes

**Files:**
- Create `frontend/src/features/mcp/mcp.ts`, `ClaudePanel.tsx`, `ClaudeChangesList.tsx`; Test `frontend/src/features/mcp/mcp.test.tsx`
- Modify `frontend/src/features/table/SettingsDialog.tsx` (tab `{ id: 'claude', label: 'Claude' }` → `<ClaudePanel campaignId={campaign.id} />`)

**Interfaces:**
- Produces in `mcp.ts`:
  ```ts
  export interface LinkStatus { exists: boolean; createdAt?: string; lastUsedAt?: string | null }
  export interface ClaudeChange { id: string; sceneId: string; summary: string; at: string }
  export const getLink: (campaignId: string) => Promise<LinkStatus>
  export const createLink: (campaignId: string) => Promise<{ url: string; createdAt: string }>
  export const revokeLink: (campaignId: string) => Promise<void>
  export const getChanges: (campaignId: string) => Promise<{ changes: ClaudeChange[] }>
  export function useClaudeChanges(campaignId: string): ClaudeChange[] // query ['mcp-changes', id] kept fresh by 'mcp.changed'
  ```
- UI copy (PT-BR): title `Conectar Claude`; button `Gerar link` / `Gerar novo link` (confirm: `O link atual vai parar de funcionar. Continuar?`); after generating, read-only input with the URL + `Copiar` and the warning `Guarde este link como uma senha: quem tiver ele age como Mestre nesta campanha. Ele só aparece agora.`; `Último uso: …` / `Nunca usado`; `Revogar link`; steps: `claude.ai: Configurações → Conectores → Adicionar conector personalizado → cole o link.` and `Claude Desktop: Configurações → Conectores → Adicionar conector personalizado → cole o link.`; section `Mudanças recentes do Claude` listing `summary` + relative time + `Desfazer` (`send('mcp.undo', { changeId })`), empty state `Nenhuma mudança ainda.`

- [ ] **Step 1: Failing tests** (Testing Library, mocked api + realtime context): generating shows the URL once and hides it after remount; revoke calls DELETE; an `mcp.changed` event updates the list; `Desfazer` sends `mcp.undo` with the id.
- [ ] **Step 2:** `cd frontend && npx vitest run src/features/mcp` → FAIL. **Step 3:** Implement. **Step 4:** PASS.
- [ ] **Step 5:** Commit `feat: claude connection tab with recent changes`.

---

### Task 11: Live toast for the Mestre, docs, full check

**Files:**
- Create `frontend/src/features/mcp/ClaudeChangeToast.tsx`; Modify `frontend/src/features/table/slots/ToastSlot.tsx` (`{isGm && <ClaudeChangeToast />}`); extend `mcp.test.tsx`
- Modify `CLAUDE.md` (one bullet under Backend describing `internal/mcp`: endpoint, link REST, tools, `mcp.undo`/`mcp.changed`), `docs/specs/0004-mcp-mestre.md` status → `DONE` only after merge (leave as is otherwise)

**Interfaces:**
- Consumes: `mcp.changed` payload `{changes, latest?}` (Task 5).
- Toast: on an event with `latest`, show `Claude: <summary>` with `Desfazer` for 8 s (emerald, same placement style as `AppliedToast` but not overlapping it — offset below); events without `latest` (undo) never open a toast and close an open toast whose change was undone.

- [ ] **Step 1: Failing tests:** event with `latest` shows the toast; `Desfazer` sends `mcp.undo`; event without `latest` shows nothing; non-GM `ToastSlot` does not render it.
- [ ] **Step 2:** FAIL. **Step 3:** Implement + CLAUDE.md bullet. **Step 4:** PASS.
- [ ] **Step 5:** Run `make check` → green. Run `docker compose up --build`, then manually: generate a link in the Claude tab, `curl -s -X POST <url> -H 'Content-Type: application/json' -H 'Accept: application/json, text/event-stream' -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"curl","version":"0"}}}'` → 200 with `serverInfo.name` `rpgmaker-mesa`.
- [ ] **Step 6:** Commit `feat: live claude change toast for the mestre`.
