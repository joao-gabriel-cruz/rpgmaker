#!/usr/bin/env bash
# Compila as imagens localmente (linux/amd64), envia para a VPS e sobe a stack.
# Uso: deploy/deploy.sh [host-ssh]   (padrão: sheeps-vps)
# Nada é compilado na VPS: ela é compartilhada com outros projetos e tem 1 vCPU.
set -euo pipefail
host="${1:-sheeps-vps}"
root="$(cd "$(dirname "$0")/.." && pwd)"
version="$(git -C "$root" rev-parse --short HEAD)"
remote=/opt/rpgmaker

docker build --platform linux/amd64 -t "rpgmaker-api:$version" "$root/backend"
docker build --platform linux/amd64 -t "rpgmaker-web:$version" "$root/frontend"
docker save "rpgmaker-api:$version" "rpgmaker-web:$version" | gzip | ssh "$host" 'gunzip | docker load'

ssh "$host" "mkdir -p $remote/backups"
scp -q "$root/deploy/docker-compose.prod.yml" "$host:$remote/docker-compose.yml"
scp -q "$root/deploy/backup.sh" "$host:$remote/backup.sh"
ssh "$host" "cd $remote && chmod +x backup.sh && sed -i 's/^RPG_VERSION=.*/RPG_VERSION=$version/' .env && docker compose up -d --remove-orphans"
# Remove só versões antigas do próprio rpgmaker (a VPS tem imagens de outros projetos).
ssh "$host" "for r in rpgmaker-api rpgmaker-web; do docker images \$r --format '{{.Tag}}' | grep -vx $version | xargs -r -I{} docker rmi \$r:{} >/dev/null; done"
echo "deploy ok: versão $version em $host"
