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
- Névoa/visão (`internal/vision`, tabela `fog`): por cena `mode` (`off` padrão | `manual` | `dynamic`), `visionRadius` (12) e `revealed` (bitset base64, bit `z*w+x`, LSB primeiro). Comandos GMOnly `fog.reveal|fog.hide {sceneId, cells:[{x,z}]}` e `fog.configure {sceneId, mode, visionRadius}`; eventos `fog.updated {sceneId, mode, visionRadius, revealed}` (ToGM) e `vision.updated {sceneId, mode, visible?}` (ToUser, só a visão do jogador; manual = revelado, dynamic = união dos tokens dos seus personagens + revelado). `GET /api/campaigns/{id}/scenes/{sceneId}/vision` devolve o mesmo por papel. `svc.Tap()` é o Notifier de scenes/tokens: filtra `token.moved|placed` por jogador (move para fora da visão vira `token.removed`), recalcula em `map.patched|objects.patched|replaced`/movimento do dono e `tokens.Service.SetVisibility` filtra o `GET .../tokens`. Limitação do MVP: o GET da cena ainda entrega o grid/objetos completos a jogadores (o cliente só renderiza o visível); só os tokens são filtrados no servidor. "Explorado" não implementado.
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
