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
#   ./scripts/migration/run-migrations.sh --record-only <file> --note "<why>"
#                                                    # register an out-of-band apply
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
#
# UPPERCASE-suffixed sidecars are NOT migrations (see SIDECAR_RE below):
#   NNN_name_ROLLBACK.sql / NNN_name_VERIFY.sql are run by hand, deliberately,
#   never by this runner. See bugs_open/007.
set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
MIGRATIONS_DIR="${MIGRATIONS_DIR:-$ROOT/docs/agent_docs/sql_for_agents}"
PSQL_CMD="${PSQL_CMD:-kubectl exec -i -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db}"
BASELINE=124

# A migration is NNN_lowercase_words.sql. A file whose trailing _TOKEN is
# UPPERCASE is a hand-run companion to the migration of the same number —
# a rollback or a verification script — and must never be auto-applied.
# This is the live convention: as of 2026-07-20 the only two files of 176 in
# the directory containing ANY uppercase letter are 180's ROLLBACK and VERIFY.
SIDECAR_RE='_[A-Z][A-Z0-9_]*\.sql$'

APPLY=0
RECORD_ONLY=""
NOTE=""

while [ $# -gt 0 ]; do
  case "$1" in
    --apply)        APPLY=1; shift ;;
    --record-only)  RECORD_ONLY="${2:-}"; shift 2 || { echo "--record-only needs a filename" >&2; exit 2; } ;;
    --note)         NOTE="${2:-}"; shift 2 || { echo "--note needs text" >&2; exit 2; } ;;
    -h|--help)      sed -n '10,17p' "${BASH_SOURCE[0]}"; exit 0 ;;
    *)              echo "unknown argument: $1" >&2; exit 2 ;;
  esac
done

if [ -n "$RECORD_ONLY" ] && [ "$APPLY" = "1" ]; then
  echo "--record-only and --apply are mutually exclusive." >&2
  exit 2
fi

sql_lit() {  # escape a value for a single-quoted SQL literal
  printf '%s' "$1" | sed "s/'/''/g"
}

psql_scalar() {  # run one -tAc query, echo the result
  $PSQL_CMD -tAc "$1" 2>/dev/null
}

# Fail loudly if the database is unreachable. Previously every psql call
# swallowed stderr and an empty result was indistinguishable from "the ledger
# does not exist" — which reads as "nothing is applied", i.e. --apply would
# replay the entire history against a live database. Refuse instead.
if [ "$(psql_scalar 'SELECT 1;')" != "1" ]; then
  echo "ERROR: cannot reach the database — refusing to run." >&2
  echo "       PSQL_CMD=$PSQL_CMD" >&2
  echo "       Diagnose with: $PSQL_CMD -c 'SELECT 1;'" >&2
  exit 1
fi

HAVE_LEDGER=$(psql_scalar "SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='schema_migrations';")

is_applied() {  # $1 = filename — live check, one round trip
  [ "$HAVE_LEDGER" = "1" ] || return 1
  [ "$(psql_scalar "SELECT 1 FROM schema_migrations WHERE filename='$(sql_lit "$1")';")" = "1" ]
}

# ---------------------------------------------------------------- record-only
# Register a migration that was applied out of band (psql -f by hand). The
# ledger only records what THIS script applies, so a hand-applied file stays
# "pending" for ever and eventually gets replayed — failing on a duplicate key
# while wearing the look of someone else's broken SQL. That trap cost ~3 days
# in 2026-07-16 and sprang again on 2026-07-20. See bugs_open/007.
#
# The artifact check is yours to do and is the load-bearing part: a ledger row
# for a migration that never ran skips it FOR EVER, which is strictly worse
# than a replay.
if [ -n "$RECORD_ONLY" ]; then
  f="$(basename "$RECORD_ONLY")"

  if [ ! -f "$MIGRATIONS_DIR/$f" ]; then
    echo "ERROR: no such migration: $MIGRATIONS_DIR/$f" >&2
    exit 1
  fi
  if [[ "$f" =~ $SIDECAR_RE ]]; then
    echo "ERROR: '$f' is an UPPERCASE-suffixed sidecar (rollback/verify), not a migration." >&2
    echo "       Sidecars are never applied by the runner, so recording one is meaningless." >&2
    exit 1
  fi
  if [ -z "$NOTE" ]; then
    echo "ERROR: --note is required. Say what was verified, e.g.:" >&2
    echo "       --note 'applied by hand 2026-07-20; checked content_components.tool-list carries {{if .image}}'" >&2
    exit 2
  fi
  if [ "$HAVE_LEDGER" != "1" ]; then
    echo "ERROR: schema_migrations does not exist — run 124_schema_migrations.sql first." >&2
    exit 1
  fi
  if is_applied "$f"; then
    echo "== $f already recorded — nothing to do."
    exit 0
  fi

  SUM=$(md5sum "$MIGRATIONS_DIR/$f" | awk '{print $1}')
  $PSQL_CMD -tAc "INSERT INTO schema_migrations (filename, checksum, applied_by, notes)
                  VALUES ('$(sql_lit "$f")', '$SUM', 'record-only', '$(sql_lit "$NOTE")')
                  ON CONFLICT (filename) DO NOTHING;" >/dev/null || {
    echo "!! failed to record $f" >&2; exit 1; }

  if is_applied "$f"; then
    echo "== $f recorded (applied_by='record-only')"
    echo "   note: $NOTE"
    exit 0
  fi
  echo "!! $f still not in the ledger after the insert — investigate." >&2
  exit 1
fi

# ------------------------------------------------------------------ scan
# Collect candidate files >= baseline, numerically ordered.
mapfile -t FILES < <(
  cd "$MIGRATIONS_DIR" && ls -1 2>/dev/null \
    | grep -E '^[0-9]{3}_[A-Za-z0-9_]+\.sql$' \
    | grep -vE "$SIDECAR_RE" \
    | awk -v base="$BASELINE" '{ n=substr($0,1,3)+0; if (n >= base) print }' \
    | sort
)

# Sidecars are reported, never silently dropped — a skip nobody can see is how
# the previous silent skip (odd filenames, below) went unnoticed for so long.
mapfile -t SIDECARS < <(
  cd "$MIGRATIONS_DIR" && ls -1 2>/dev/null \
    | grep -E '^[0-9]{3}_[A-Za-z0-9_]+\.sql$' \
    | grep -E "$SIDECAR_RE" \
    | awk -v base="$BASELINE" '{ n=substr($0,1,3)+0; if (n >= base) print }' \
    | sort
)
if [ ${#SIDECARS[@]} -gt 0 ]; then
  echo "Sidecars (hand-run only, NOT applied by this runner):"
  printf '  %s\n' "${SIDECARS[@]}"
  echo ""
fi

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

# Fetch the whole ledger in ONE round trip. The per-file query this replaces
# made a dry run take minutes through `kubectl exec` — and the dry run is the
# check bugs_open/007 asks every thread to make before trusting the queue, so
# its cost is the reason it gets skipped.
declare -A APPLIED=()
if [ "$HAVE_LEDGER" = "1" ]; then
  while IFS= read -r row; do
    [ -n "$row" ] && APPLIED["$row"]=1
  done < <(psql_scalar "SELECT filename FROM schema_migrations;")
fi

PENDING=()
for f in "${FILES[@]}"; do
  [ -n "${APPLIED[$f]:-}" ] || PENDING+=("$f")
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
  echo "NOTE: pending files may belong to other threads, and 'pending' can also"
  echo "      mean 'applied by hand and never recorded' — verify before applying."
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
    echo "" >&2
    echo "   If the error above is a duplicate key (SQLSTATE 23505), the likeliest" >&2
    echo "   cause is NOT broken SQL: this file has ALREADY been applied by hand and" >&2
    echo "   was never recorded in schema_migrations, so the runner replayed it." >&2
    echo "   Verify its artifacts exist in the database, then register it:" >&2
    echo "" >&2
    echo "     $0 --record-only $f --note 'applied by hand <date>; <what you checked>'" >&2
    echo "" >&2
    echo "   Only record it if you CHECKED — a ledger row for a migration that never" >&2
    echo "   ran skips it for ever. See bugs_open/007." >&2
    exit 1
  fi
  SUM=$(md5sum "$MIGRATIONS_DIR/$f" | awk '{print $1}')
  $PSQL_CMD -tAc "INSERT INTO schema_migrations (filename, checksum, applied_by, notes)
                  VALUES ('$(sql_lit "$f")', '$SUM', 'run-migrations.sh', NULL)
                  ON CONFLICT (filename) DO NOTHING;" >/dev/null
  echo "== $f recorded"
done
echo ""
echo "Done."
