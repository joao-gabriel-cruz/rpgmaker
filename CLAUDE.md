# rpgmaker — convenções

Mesa virtual privada para Assimilação RPG. Spec: `docs/specs/0001-rpgmaker-mvp.md` (fonte de verdade
das regras e decisões). Tickets: `docs/tickets/0001-rpgmaker-mvp.md` + issues do GitHub.

## Idioma
- **Código em inglês:** identificadores, comentários, logs, erros, nomes de teste, commits (título).
- **UI em PT-BR:** todo texto visível ao usuário.
- Docs/specs em PT-BR.

## Limites (hook de commit bloqueia)
- Arquivo ≤ 300 linhas; função ≤ 30 linhas. Quebre por responsabilidade.
- Comentários só para o *porquê* não óbvio, ≤ 2 linhas.

## Backend (Go, `backend/`)
- `cmd/api` só faz wiring: módulos da mesa entram em `cmd/api/wire.go` com uma função `wireX(d deps)` e uma linha em `wireTable` (evita conflito entre tickets). Domínios em `internal/<domain>` (ex.: `auth`, `campaigns`, `realtime`,
  `dice`, `systems/assimilacao`, `maps`), cada um com `store.go` (SQL via pgx), `service.go`
  (regras) e `http.go` (handlers) quando fizer sentido.
- HTTP: `net/http` ServeMux com padrões `"METHOD /api/..."`; JSON in/out; erros como
  `{"error": "<mensagem PT-BR para o usuário>"}` com status correto.
- Auth: sessão em cookie httpOnly `rpg_session` (SameSite=Lax), sessões no Postgres; middleware
  injeta o usuário no context.
- Migrations: goose em `internal/migrations/sql/NNNNN_desc.sql`, nunca editar migration já
  mergeada. Como tickets rodam em paralelo, a versão NNNNN é o **número do ticket** (ex.:
  `00009_campaigns.sql`); precisando de mais de uma, use 10009, 20009… O provider usa
  `WithAllowOutofOrder(true)` para aceitar versões mergeadas fora de ordem.
  IDs `uuid` com `gen_random_uuid()`. Dados específicos de sistema em `jsonb`.
- Tempo real: WebSocket por campanha em `GET /api/campaigns/{id}/ws`; servidor autoritativo;
  mensagens `{type, payload, reqId?}`; comandos validados no servidor; eventos filtrados por
  destinatário.
- API do tempo real (`internal/realtime`, um `*Hub` criado em `main` e injetado nos domínios):
  - Comando (cliente→servidor): `hub.Handle("token.move", func(ctx, cmd realtime.Command) error, realtime.GMOnly)`;
    `Command{CampaignID, Sender{UserID,Name,Role}, Type, Payload json.RawMessage, ReqID}`. Guards:
    `GMOnly`, `AnyMember` (padrão) ou `func(Sender) error`. Erro `realtime.UserError("msg PT-BR")` vai ao
    remetente como evento `error {reqId,message}`; outros erros viram mensagem genérica. Sucesso → `ack {reqId}`.
  - Evento (servidor→clientes), de handlers WS ou HTTP: `hub.Broadcast(campaignID, filter, "type", payload)`,
    nunca bloqueia. Filtros: `ToAll`, `ToGM`, `ToUser(id)`, `Except(id)`, `Or(...)` ou qualquer `func(Sender) bool`.
    Eco: `cmd.ConnID` + `ExceptConn(id)` excluem só a conexão (aba) de origem; `Except(userID)` exclui todas as abas do usuário.
  - Presença: evento `presence.updated {users:[{userId,displayName,role}]}`; `hub.Online(campaignID)`. `ping` → `pong`.
  - Registre handlers no wiring (`main`), antes de servir; tipo duplicado dá panic.
- Rodadas (`internal/rounds`): comandos GMOnly `round.start|next|prev|end` → evento `round.updated {state:{active,round,phase}}` (ToAll); estado atual via `GET /api/campaigns/{id}/round`; persistido em `campaign_rounds`.
  Gancho: `svc.OnPhaseChange(func(ctx, campaignID, prev, next rounds.State))` roda síncrono após persistir cada transição (ex.: efeitos de terreno "ao começar o turno"); registre antes de servir.
- Ações (`internal/actions`, tabela `campaign_actions`): catálogo padrão do Assimilação (11 ações em `defaults.go`) + overrides/ações próprias da campanha. REST `GET /api/campaigns/{id}/actions` (membros; o Mestre vê as desativadas), `POST|PUT /{actionId}|DELETE /{actionId}` (GM; DELETE de padrão = restaurar). Comando `action.perform {characterId, actionId, target{kind,x?,z?,tokenId?,objectId?}, rollChoice, modifiers{effort,useOccupation,useMarkingEvent,secondInstinct}}` (dono ou Mestre; alvo na cena ativa, alcance Chebyshev e linha de visão se alcance > 2) cria a rolagem pelo `rolls.RequestEntry` e emite `action.started` (ToAll); `rolls.OnFinalized` dispara `action.resolved` (ToAll, com `outcome{kind,text}`) e as consequências (sugestões do `effects`, `items.TryFind`, teste de Descanso com `action.cards` ToGM e comando GMOnly `action.card.accept`). Bônus `effects.keptBonus` dos itens do corpo entram via `RequestInput.BonusKept` (só uso interno). Runs pendentes ficam em memória. `action.started {actionRunId, actionId, actionName, actorName, animation, actorCharacterId, actorTokenId, actor{x,z}, target, rollId, variant?}`; chaves de `animation`: melee_strike, ranged_shot, look_around, search, sneak, repair, first_aid, rest, dance, persuade, intimidate (+ `hit_react` no alvo). No front, `src/lib/fxBus.ts` leva `fx:<animation>`, `fx:hit_react`, `fx:loot_found` e `fx:number` com `{x, z, ...}` (sons e efeitos assinam esses nomes).
- Rolagem do Mestre sem personagem: comando GMOnly `roll.pool {d6,d10,d12}` (até 20 de cada) registra uma rolagem `threat` pública.
- Névoa/visão (`internal/vision`, tabela `fog`): por cena `mode` (`off` padrão | `manual` | `dynamic`), `visionRadius` (12) e `revealed` (bitset base64, bit `z*w+x`, LSB primeiro). Comandos GMOnly `fog.reveal|fog.hide {sceneId, cells:[{x,z}]}` e `fog.configure {sceneId, mode, visionRadius}`; eventos `fog.updated {sceneId, mode, visionRadius, revealed}` (ToGM) e `vision.updated {sceneId, mode, visible?}` (ToUser, só a visão do jogador; manual = revelado, dynamic = união dos tokens dos seus personagens + revelado). `GET /api/campaigns/{id}/scenes/{sceneId}/vision` devolve o mesmo por papel. `svc.Tap()` é o Notifier de scenes/tokens: filtra `token.moved|placed` por jogador (move para fora da visão vira `token.removed`), recalcula em `map.patched|objects.patched|replaced`/movimento do dono e `tokens.Service.SetVisibility` filtra o `GET .../tokens`. Limitação do MVP: o GET da cena ainda entrega o grid/objetos completos a jogadores (o cliente só renderiza o visível); só os tokens são filtrados no servidor. "Explorado" não implementado.
- MCP do Mestre (`internal/mcp`): endpoint `/api/mcp/{secret}` (JSON-RPC stateless, sem sessão; só o sha256 do segredo é guardado em `mcp_links`, resolve a campanha). Link: `GET|POST|DELETE /api/campaigns/{id}/mcp-link` (GM; POST gera/regenera e devolve a URL uma única vez). Tools: `get_table`, `list_scenes`, `get_catalog`, `create_scene`, `activate_scene`, `paint_terrain`, `place_objects`, `remove_objects`, `set_door`, `move_token`, `set_ambience`, `set_condition`, `undo_last`. Escrita age na cena ativa e recusa se há rascunho aberto (`há um rascunho aberto nesta cena…`). Journal dos 20 últimos changes por campanha (`journal.go`): evento `mcp.changed {changes, latest?}` (ToGM), comando `mcp.undo {changeId}` e `GET /api/campaigns/{id}/mcp-changes` (GM). As tools reusam as mesmas instâncias de service de scenes/tokens/ambience/characters (`wireMCP`), então os eventos normais chegam ao mapa ao vivo.
- RNG de dados: `crypto/rand`.
- Testes: `go test ./...`; testes de store/integração podem usar o Postgres do compose via
  `TEST_DATABASE_URL` e devem ser pulados (`t.Skip`) quando a variável não existir.

## Frontend (React + TS + Tailwind, `frontend/`)
- Pastas por feature em `src/features/<feature>/`; utilitários compartilhados em `src/lib/`.
- Roteamento: `react-router`. Dados do servidor: `@tanstack/react-query` + cliente `src/lib/api.ts`
  (fetch com `credentials: 'include'`). Tempo real (`src/features/realtime/`): `useCampaignSocket(campaignId)` → `{status, send(type,payload): Promise<void>, subscribe(type,handler): unsubscribe}` (socket único por campanha, reconexão com backoff); `<CampaignSocketProvider>` + `useRealtime()` para telas filhas; `usePresence()`.
- 3D: `three` + `@react-three/fiber` + `@react-three/drei`.
- Tema escuro (zinc/emerald), desktop-first.
- Testes: Vitest + Testing Library.

## Qualidade
- `make check` (lint + build + testes, backend e frontend) precisa ficar verde.
- `docker compose up --build` precisa continuar subindo tudo.
