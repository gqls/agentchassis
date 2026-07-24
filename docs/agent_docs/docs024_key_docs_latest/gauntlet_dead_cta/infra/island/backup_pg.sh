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
# TODO(owner): fill in from the Mythic Beasts control panel ("Backup account"
# host + username), install the island's SSH key there, then uncomment:
# rsync -a --delete /opt/island/backups/ <BACKUP_USER>@<BACKUP_HOST>:tools-api-backups/
