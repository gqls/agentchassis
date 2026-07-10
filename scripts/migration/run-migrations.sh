#!/usr/bin/env bash
# run-migrations.sh — apply pending SQL migrations to clients_db, in order,
# recording each in schema_migrations.
#
# Migrations home:  docs/agent_docs/sql_for_agents/   (NNN_name.sql)
# Baseline:         124 — files 001–123 predate the tracking system and are
#                   history, never auto-applied. 124_schema_migrations.sql
#                   bootstraps the ledger (and backfills 125–139).
#
# Usage:
#   ./scripts/migration/run-migrations.sh            # list pending (dry run, default)
#   ./scripts/migration/run-migrations.sh --apply    # apply pending, in order
#
# Env overrides:
#   MIGRATIONS_DIR  (default: docs/agent_docs/sql_for_agents relative to repo root)
#   PSQL_CMD        (default: kubectl exec -i -n ai-persona-system postgres-clients-0
#                             -- psql -U clients_user -d clients_db)
#
# Conventions this shop already runs on (unchanged by this script):
#   * snapshot_agent('<type>','<file>: pre-update') opens every migration that
#     touches agent_definitions;
#   * every migration carries its own guard DO block and COMMITs only if the
#     guard passes; ON_ERROR_STOP below makes any failure abort that file;
#   * a failed file stops the run — nothing after it is attempted, and the
#     failed file is NOT recorded (psql wraps each -f in its own transaction
#     only if the file says BEGIN/COMMIT — keep writing them that way).
set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
MIGRATIONS_DIR="${MIGRATIONS_DIR:-$ROOT/docs/agent_docs/sql_for_agents}"
PSQL_CMD="${PSQL_CMD:-kubectl exec -i -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db}"
BASELINE=124
APPLY=0
[ "${1:-}" = "--apply" ] && APPLY=1

psql_scalar() {  # run one -tAc query, echo the result
  $PSQL_CMD -tAc "$1" 2>/dev/null
}

# Does the ledger exist yet? (bootstrap: if not, every file >= baseline is pending)
HAVE_LEDGER=$(psql_scalar "SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='schema_migrations';")

is_applied() {  # $1 = filename
  [ "$HAVE_LEDGER" = "1" ] || return 1
  [ "$(psql_scalar "SELECT 1 FROM schema_migrations WHERE filename='$1';")" = "1" ]
}

# Collect candidate files >= baseline, numerically ordered.
mapfile -t FILES < <(
  cd "$MIGRATIONS_DIR" && ls -1 2>/dev/null \
    | grep -E '^[0-9]{3}_[A-Za-z0-9_]+\.sql$' \
    | awk -v base="$BASELINE" '{ n=substr($0,1,3)+0; if (n >= base) print }' \
    | sort
)

# Loud warning for near-misses: a *.sql file whose 3-digit prefix is >= baseline
# but whose name breaks the pattern (hyphens, 'NNNb_' suffixes, spaces) would
# otherwise be SILENTLY skipped — this shop has used such names (081_..-fetcher).
while IFS= read -r odd; do
  echo "WARNING: '$odd' looks like a migration (number >= $BASELINE) but does not match" >&2
  echo "         NNN_name.sql (digits+underscores only) — it will NOT be applied. Rename it." >&2
done < <(
  cd "$MIGRATIONS_DIR" && ls -1 2>/dev/null \
    | grep -E '^[0-9]{3}.*\.sql$' \
    | grep -vE '^[0-9]{3}_[A-Za-z0-9_]+\.sql$' \
    | awk -v base="$BASELINE" '{ n=substr($0,1,3)+0; if (n >= base) print }'
)

PENDING=()
for f in "${FILES[@]}"; do
  is_applied "$f" || PENDING+=("$f")
done

if [ ${#PENDING[@]} -eq 0 ]; then
  echo "Up to date — no pending migrations (baseline $BASELINE, dir: $MIGRATIONS_DIR)."
  exit 0
fi

echo "Pending (${#PENDING[@]}):"
printf '  %s\n' "${PENDING[@]}"

if [ "$APPLY" != "1" ]; then
  echo ""
  echo "Dry run. Re-run with --apply to execute."
  exit 0
fi

echo ""
for f in "${PENDING[@]}"; do
  # Re-check right before applying: an earlier file this run (e.g. the 124
  # bootstrap backfill) may have recorded this one already.
  HAVE_LEDGER=$(psql_scalar "SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='schema_migrations';")
  if is_applied "$f"; then
    echo "== $f — already recorded (skipping)"
    continue
  fi
  echo "== applying $f"
  if ! $PSQL_CMD -v ON_ERROR_STOP=1 < "$MIGRATIONS_DIR/$f"; then
    echo "!! $f FAILED — stopping. Nothing recorded for it; later files not attempted." >&2
    exit 1
  fi
  SUM=$(md5sum "$MIGRATIONS_DIR/$f" | awk '{print $1}')
  $PSQL_CMD -tAc "INSERT INTO schema_migrations (filename, checksum, applied_by, notes)
                  VALUES ('$f', '$SUM', 'run-migrations.sh', NULL)
                  ON CONFLICT (filename) DO NOTHING;" >/dev/null
  echo "== $f recorded"
done
echo ""
echo "Done."
