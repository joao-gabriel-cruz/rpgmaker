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

## Primeira conta admin

Não há cadastro público: o admin é criado por comando (a senha deve ter ao menos 8 caracteres).

```bash
# com a stack do compose rodando
docker compose exec api /api create-admin --email admin@exemplo.com --name "Admin" --password 'troque-esta-senha'
# ou, sem a API no ar (só o banco)
docker compose run --rm api create-admin --email admin@exemplo.com --name "Admin" --password '...'
# em desenvolvimento local
cd backend && DATABASE_URL=... go run ./cmd/api create-admin --email ... --name ... --password ...
```

Depois, entre em http://localhost:3000/login. Em `/admin` o admin lista usuários e gera senhas temporárias.

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

Testes de store usam Postgres real: `TEST_DATABASE_URL=postgres://rpgmaker:rpgmaker@localhost:5433/rpgmaker_test?sslmode=disable`
(sem a variável eles são pulados; os testes limpam a tabela `users`, use um banco de testes).

## Migrations

Arquivos SQL em `backend/internal/migrations/sql/`, nomeados `NNNNN_descricao.sql` com blocos
`-- +goose Up` / `-- +goose Down`. São embutidos no binário e aplicados na subida da API.
