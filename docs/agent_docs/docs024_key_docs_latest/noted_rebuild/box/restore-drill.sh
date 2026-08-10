#!/usr/bin/env bash
# restore-drill.sh — can we get noted's data back from nothing but the B2 bucket
# and the age identity?
#
# Run quarterly, and after ANY change to the backup scripts.
#
# WHAT A DRILL IS FOR, AND WHY THIS ONE IS SHAPED LIKE THIS
#   The question is never "does a file exist in the bucket". It is "on the day
#   the box is gone, can a person holding only the identity read the notes
#   back". So every step here runs OFF the box, and the identity used must be
#   the one from the owner's stored copy — not a copy left lying around from
#   setup, because the step that fails in a real recovery is always the key
#   nobody could find.
#
# USAGE
#   ./restore-drill.sh [path-to-age-identity]
#   AGE=/path/to/age  PGHOST=... ./restore-drill.sh ~/id.txt
#
# Needs: an ADMIN b2 key (the backup key cannot read or list, by design), and
# the `age` binary. A local postgres client is optional — without one the drill
# still proves retrieval, decryption and archive validity, and says plainly that
# it could not prove the final restore.
set -uo pipefail

BUCKET=${NOTED_BACKUP_BUCKET:-personae-noted-backups}
PREFIX=noted/pg/
IDENTITY=${1:-$HOME/noted-backup-age-identity.txt}
AGE=${AGE:-age}
WORK=$(mktemp -d); trap 'rm -rf "$WORK"' EXIT
fail=0

step() { printf '\n== %s\n' "$*"; }
ok()   { printf '   PASS  %s\n' "$*"; }
bad()  { printf '   FAIL  %s\n' "$*"; fail=1; }

step "1. Identity available?"
if [ -r "$IDENTITY" ]; then
    ok "$IDENTITY readable ($(stat -c %a "$IDENTITY") $(stat -c %U "$IDENTITY"))"
    [ "$(stat -c %a "$IDENTITY")" = "600" ] || bad "identity is not mode 600"
else
    bad "identity not readable at $IDENTITY — THIS IS THE WHOLE DRILL. Without it every backup is unreadable."
    exit 1
fi

step "2. Newest backup in the bucket (listing with --versions, so a HIDE is visible)"
listing=$(b2 ls --recursive --long --versions "b2://${BUCKET}/${PREFIX}" 2>&1) || {
    bad "cannot list the bucket — $listing"; exit 1; }
hidden=$(printf '%s\n' "$listing" | awk '$5 == 0 {print $NF}' | sort -u)
[ -n "$hidden" ] && bad "hide markers present — run: b2 file unhide b2://${BUCKET}/<name>"
OBJ=$(b2 ls --recursive "b2://${BUCKET}/${PREFIX}" 2>/dev/null | tail -1)
[ -n "$OBJ" ] || { bad "no visible backup objects"; exit 1; }
ok "newest: $OBJ"

step "3. Download it (admin key — the backup key deliberately cannot)"
b2 file download --no-progress "b2://${BUCKET}/${OBJ}" "$WORK/b.age" >/dev/null 2>&1 \
    && ok "downloaded $(stat -c %s "$WORK/b.age") bytes" || { bad "download failed"; exit 1; }

step "4. Is it actually encrypted? (a plaintext backup of people's notes is a breach)"
if head -c 21 "$WORK/b.age" | grep -q '^age-encryption.org'; then
    ok "age header present — not plaintext"
else
    bad "NOT age-encrypted — investigate immediately, this is a disclosure"
fi

step "5. Decrypt with the identity"
if "$AGE" -d -i "$IDENTITY" -o "$WORK/b.dump" "$WORK/b.age" 2>"$WORK/err"; then
    ok "decrypted to $(stat -c %s "$WORK/b.dump") bytes"
else
    bad "DECRYPT FAILED — $(head -1 "$WORK/err"). Wrong identity, or the backups are unrecoverable."
    exit 1
fi

step "6. Is it a valid, complete PostgreSQL archive?"
ft=$(file -b "$WORK/b.dump")
case "$ft" in
    *"PostgreSQL custom database dump"*) ok "$ft" ;;
    *) bad "not a pg dump: $ft" ;;
esac

step "7. Restore into a throwaway database"
# Deliberately does NOT fall back to the box's postgres. The scenario being
# rehearsed is "the box is gone", so borrowing its database would rehearse the
# one case that cannot happen.
if command -v pg_restore >/dev/null 2>&1 && command -v psql >/dev/null 2>&1; then
    DB=noted_drill_$$
    createdb "$DB" 2>/dev/null && pg_restore -d "$DB" "$WORK/b.dump" 2>/dev/null
    tables=$(psql -d "$DB" -tAc "SELECT count(*) FROM information_schema.tables WHERE table_schema='public';" 2>/dev/null)
    ok "restored via local client; $tables table(s) in public"
    psql -d "$DB" -tAc "SELECT table_name FROM information_schema.tables WHERE table_schema='public';" 2>/dev/null | sed 's/^/         - /'
    dropdb "$DB" 2>/dev/null
elif docker ps >/dev/null 2>&1; then
    C=noted-drill-$$
    docker run -d --rm --name "$C" -e POSTGRES_PASSWORD=drill postgres:16 >/dev/null 2>&1
    for _ in $(seq 30); do
        docker exec "$C" pg_isready -U postgres >/dev/null 2>&1 && break
        sleep 1
    done
    docker exec -u postgres "$C" createdb drill >/dev/null 2>&1
    docker exec -i -u postgres "$C" pg_restore -d drill < "$WORK/b.dump" >/dev/null 2>&1
    tables=$(docker exec -u postgres "$C" psql -d drill -tAc \
        "SELECT count(*) FROM information_schema.tables WHERE table_schema='public';" 2>/dev/null | tr -d ' ')
    if [ "${tables:-0}" -ge 0 ] 2>/dev/null && [ -n "$tables" ]; then
        ok "restored in a throwaway postgres:16 container; $tables table(s) in public"
        docker exec -u postgres "$C" psql -d drill -tAc \
            "SELECT table_name FROM information_schema.tables WHERE table_schema='public';" 2>/dev/null \
            | sed '/^$/d; s/^ */         - /'
        # Read actual ROWS back, not just table names — a schema-only restore
        # would satisfy every check above and contain none of anyone's notes.
        docker exec -u postgres "$C" psql -d drill -tAc \
            "SELECT '         rows in '||relname||': '||n_live_tup FROM pg_stat_user_tables ORDER BY relname;" 2>/dev/null \
            | sed '/^$/d'
    else
        bad "container restore produced no readable schema"
    fi
    docker stop "$C" >/dev/null 2>&1
else
    bad "SKIPPED — no local postgres client and no usable docker, so the final restore was NOT proven."
    printf '         Weaker evidence follows; this is not a substitute.\n'
    strings "$WORK/b.dump" | grep -E '^CREATE TABLE|^COPY ' | head -10 | sed 's/^/         /'
fi

printf '\n'
[ "$fail" -eq 0 ] && printf 'DRILL PASSED\n' || printf 'DRILL INCOMPLETE OR FAILED — see FAIL lines above\n'
exit $fail
