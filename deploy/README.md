# Deploy (VPS compartilhada `sheeps-vps`)

A VPS hospeda outros projetos (Sheeps, Gefen). Regras:

- **Nada é compilado na VPS** (1 vCPU, sem swap). `deploy/deploy.sh` compila as imagens aqui
  (linux/amd64), envia com `docker save | ssh docker load` e sobe a stack em `/opt/rpgmaker`.
- **Nenhuma porta no host.** O `rpgmaker-web` entra na rede `bootstrap_default` (a do Caddy da Sheeps)
  com nome fixo; o Caddy faz o TLS (Let's Encrypt).
- **Nomes exclusivos na rede compartilhada:** outros projetos atendem por `api` nessa rede, por isso o
  nginx fala com `rpgmaker-api` (alias da API na rede interna da stack).
- Limites de memória: db 384 MB, api 256 MB, web 96 MB.
- Sem backup automático (decisão do dono). Backup manual quando precisar:
  `ssh sheeps-vps 'cd /opt/rpgmaker && docker compose exec -T db pg_dump -U rpgmaker -d rpgmaker' | gzip > rpgmaker.sql.gz`

## Primeira vez
1. `/opt/rpgmaker/.env` (permissão 600) com `POSTGRES_PASSWORD=<aleatória>`, `RPG_VERSION=none`,
   `WS_ALLOWED_ORIGINS=https://<domínio>`.
2. `deploy/deploy.sh sheeps-vps`
3. Admin: `ssh sheeps-vps 'cd /opt/rpgmaker && docker compose exec api /api create-admin --email ... --name ... --password ...'`
4. Domínio: registro A do domínio → IP da VPS e, no Caddyfile da Sheeps
   (`/root/projects/sheeps/bootstrap/caddy/Caddyfile`), um bloco literal como os da Gefen:
   ```
   <domínio> {
   	encode gzip zstd
   	reverse_proxy rpgmaker-web:80 {
   		header_up X-Forwarded-For {remote_host}
   	}
   }
   ```
   e `docker exec sheeps-caddy caddy reload --config /etc/caddy/Caddyfile`.

## Atualizar
`deploy/deploy.sh sheeps-vps` a partir da `main` (a tag da imagem é o commit).
