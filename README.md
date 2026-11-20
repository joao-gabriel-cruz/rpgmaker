# rpgmaker

Mesa virtual privada para Assimilação RPG (e outros sistemas). Spec: `docs/specs/0001-rpgmaker-mvp.md`;
tickets: `docs/tickets/0001-rpgmaker-mvp.md`.

## Estrutura

- `backend/` — API em Go (HTTP, migrations embutidas com goose, Postgres via pgx)
- `frontend/` — React + TypeScript + Tailwind (Vite)
- `docker-compose.yml` — Postgres + API + frontend (nginx)

## Rodar tudo

Requisito: Docker.

```bash
docker compose up --build
```

Abra http://localhost:3000 — a página deve mostrar **"API ok · banco ok"**. As migrations rodam
automaticamente quando a API sobe. Portas e credenciais podem ser trocadas copiando `.env.example`
para `.env`.

## Desenvolvimento

Requisitos: Go 1.26+, Node 20.19+, golangci-lint 2.x.

```bash
docker compose up db                     # só o banco (porta 5433)
cd backend && DATABASE_URL=postgres://rpgmaker:rpgmaker@localhost:5433/rpgmaker?sslmode=disable PORT=8081 go run ./cmd/api
cd frontend && npm install && npm run dev  # http://localhost:5173, /api vai para a porta 8081
```

## Qualidade

```bash
make lint    # golangci-lint + oxlint
make build   # go build + tsc/vite build
make test    # go test + vitest
make check   # os três
```

## Migrations

Arquivos SQL em `backend/internal/migrations/sql/`, nomeados `NNNNN_descricao.sql` com blocos
`-- +goose Up` / `-- +goose Down`. São embutidos no binário e aplicados na subida da API.
