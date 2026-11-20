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
- `cmd/api` só faz wiring. Domínios em `internal/<domain>` (ex.: `auth`, `campaigns`, `realtime`,
  `dice`, `systems/assimilacao`, `maps`), cada um com `store.go` (SQL via pgx), `service.go`
  (regras) e `http.go` (handlers) quando fizer sentido.
- HTTP: `net/http` ServeMux com padrões `"METHOD /api/..."`; JSON in/out; erros como
  `{"error": "<mensagem PT-BR para o usuário>"}` com status correto.
- Auth: sessão em cookie httpOnly `rpg_session` (SameSite=Lax), sessões no Postgres; middleware
  injeta o usuário no context.
- Migrations: goose em `internal/migrations/sql/NNNNN_desc.sql`, nunca editar migration já
  mergeada. IDs `uuid` com `gen_random_uuid()`. Dados específicos de sistema em `jsonb`.
- Tempo real: WebSocket por campanha em `GET /api/campaigns/{id}/ws`; servidor autoritativo;
  mensagens `{type, payload, reqId?}`; comandos validados no servidor; eventos filtrados por
  destinatário.
- RNG de dados: `crypto/rand`.
- Testes: `go test ./...`; testes de store/integração podem usar o Postgres do compose via
  `TEST_DATABASE_URL` e devem ser pulados (`t.Skip`) quando a variável não existir.

## Frontend (React + TS + Tailwind, `frontend/`)
- Pastas por feature em `src/features/<feature>/`; utilitários compartilhados em `src/lib/`.
- Roteamento: `react-router`. Dados do servidor: `@tanstack/react-query` + cliente `src/lib/api.ts`
  (fetch com `credentials: 'include'`). Tempo real: hook `useCampaignSocket`.
- 3D: `three` + `@react-three/fiber` + `@react-three/drei`.
- Tema escuro (zinc/emerald), desktop-first.
- Testes: Vitest + Testing Library.

## Qualidade
- `make check` (lint + build + testes, backend e frontend) precisa ficar verde.
- `docker compose up --build` precisa continuar subindo tudo.
