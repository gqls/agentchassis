#!/usr/bin/env bash
# noted-pg-backup.sh — nightly dump of the `noted` database.
#
# Installed on webdesign.vs.mythic-beasts.com 2026-08-10, when Postgres was
# installed for noted.co.uk. Run by noted-pg-backup.timer (daily, 03:20 + jitter).
#
# SCOPE, STATED HONESTLY: this writes to the SAME DISK as the database. That
# protects against the failure that actually happens most — a bad migration, a
# DELETE without a WHERE, a corrupted table — and against NOTHING ELSE. If this
# box is lost, these dumps are lost with it.
#
# An off-box copy is REQUIRED before real user notes land here, and is
# deliberately not set up in this script: it needs a credential on the box, and
# this box's whole posture is that it holds none and dials nothing in. A B2
# application key scoped write-only to one backup prefix is the intended answer
# (it dials OUT, like cloudflared, so it fits the posture) but it is the owner's
# call to issue one. Until then, treat this as crash-and-fat-finger insurance,
# not disaster recovery.
set -euo pipefail

DB=noted
DEST=/var/backups/noted
KEEP_DAYS=14
STAMP=$(date -u +%Y%m%dT%H%M%SZ)
OUT="$DEST/${DB}-${STAMP}.dump"

install -d -m 700 -o root -g root "$DEST"

# -Fc: custom format — compressed, and restorable selectively with pg_restore.
# Dump as the postgres superuser so the backup does not depend on the service
# role's grants (a role that loses a privilege must not silently stop backing up).
#
# Write to STDOUT and redirect as root, rather than `pg_dump -f "$OUT"`. The
# dump directory is 700 root:root because a dump contains every note in the
# database, and `-f` would make pg_dump — running as the `postgres` user —
# open the file itself, which it cannot do. The redirect is performed by this
# script's own shell, which is root. Getting this wrong fails at 03:20 nightly
# with "Permission denied" and nobody watching.
sudo -u postgres pg_dump -Fc -d "$DB" > "$OUT.partial"

# Only name it as a real backup once it is complete. A half-written dump that
# looks like a backup is worse than no backup, because retention will rotate a
# good one out to make room for it.
mv "$OUT.partial" "$OUT"
chmod 600 "$OUT"

SIZE=$(stat -c %s "$OUT")
# A pg_dump of an empty-but-valid database is ~1-2 KB of header. Anything at or
# below a few hundred bytes means the dump failed in a way that still exited 0.
if [ "$SIZE" -lt 300 ]; then
    echo "noted-pg-backup: FAILED — dump is only ${SIZE} bytes, refusing to rotate" >&2
    exit 1
fi

# Prune only AFTER a verified-good new dump exists, never before.
find "$DEST" -name "${DB}-*.dump" -type f -mtime "+${KEEP_DAYS}" -delete
find "$DEST" -name "*.partial"    -type f -mtime +1 -delete

COUNT=$(find "$DEST" -name "${DB}-*.dump" -type f | wc -l)
echo "noted-pg-backup: wrote $OUT (${SIZE} bytes); ${COUNT} dump(s) retained, keeping ${KEEP_DAYS} days"
