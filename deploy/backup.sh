#!/usr/bin/env bash
# Backup diário do Postgres do rpgmaker (cron em /etc/cron.d/rpgmaker-backup). Guarda 14 dias.
set -euo pipefail
cd /opt/rpgmaker
file="backups/rpgmaker-$(date +%F-%H%M).sql.gz"
docker compose exec -T db pg_dump -U rpgmaker -d rpgmaker | gzip > "$file"
find backups -name 'rpgmaker-*.sql.gz' -mtime +14 -delete
