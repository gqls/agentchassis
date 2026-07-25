#!/bin/sh
# Island: /opt/island/backup_pg.sh — nightly via /etc/cron.d/island-backup.
# Dumps the island DB locally (14-day retention) and mirrors to Mythic Beasts
# backup space (mirrored to a second UK site per MB).
set -eu
STAMP=$(date +%F)
DUMP=/opt/island/backups/tools_api_$STAMP.sql.gz
mkdir -p /opt/island/backups
docker compose -f /opt/island/docker-compose.yml exec -T postgres \
  pg_dump -U tools_api tools_api | gzip > "$DUMP"
find /opt/island/backups -name 'tools_api_*.sql.gz' -mtime +14 -delete
# Off-box mirror to the MB backup account (20GB, mirrored to a second UK site).
# Auth is key-only: the island's /root/.ssh/id_ed25519.pub must be installed on
# the backup account via the MB control panel. Until then this line fails loudly
# in cron mail while the local dump above still succeeds — it self-heals the
# moment the key is installed.
rsync -a --delete /opt/island/backups/ 32950_toolsapisuk@backup-sov-a.mythic-beasts.com:tools-api-backups/
