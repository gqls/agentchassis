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
#   ./scripts/migration/run-migrations.sh --no-probe # dry run / apply without the
#                                                    # already-applied probe (below)
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
#
# The PROBE (bugs_open/007 residual 1, added 2026-07-24): by default each
# pending file is asked whether it has already been applied, by executing it
# verbatim inside a DOOMED transaction — a deferred UNIQUE violation planted
# first makes any COMMIT the file issues fail (Postgres-enforced, no SQL
# parsing), and ON_ERROR_STOP stops psql at that failure so nothing after it
# can run in autocommit. A file whose own guard RAISEs '... already ...'
# (P0001) is reported LIKELY ALREADY APPLIED; under --apply it is SKIPPED and
# the run continues past it instead of halting — but it is never RECORDED
# (recording stays a human act: verify artifacts, then --record-only). A raw
# duplicate key (23505) never skips — it can be genuinely wrong SQL — so that
# still halts loudly; since 2026-07-25 the halt lands BEFORE the file is
# executed (the probe already saw the duplicate; the verdict names the table
# and key). Files the doomed transaction cannot safely contain (their
# own ROLLBACK/ABORT, psql metas, setval, ...) are listed "not probed" and
# behave exactly as before. NOTE a probing run EXECUTES pending SQL then rolls
# it back: brief row locks are taken and sequences may advance. --no-probe
# restores the previous behaviour byte for byte.
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

# Idempotency lint (bugs_open/007 candidate c). Tables where a duplicate INSERT
# on replay is HARMLESS, so an unguarded insert into them is not a hazard worth
# a warning: append-only log tables (doc_notes, doc_plans) and the runner's own
# ledger (schema_migrations, always written ON CONFLICT DO NOTHING). An unguarded
# INSERT into any OTHER table is the non-idempotency that springs the trap on
# replay — 151 inserted into content_components with no guard and that is what
# blocked the runner for 3 days. This allowlist is grounded in the live corpus,
# NOT the handoff's literal "any bare INSERT" spec: nearly every migration writes
# a doc_notes row, so linting those would be pure noise and the warning would be
# learned-ignored. Optional schema prefix (public.doc_notes) tolerated.
IDEMPOTENT_SINK_RE='(^|\.)(doc_notes|doc_plans|schema_migrations)$'

APPLY=0
RECORD_ONLY=""
NOTE=""
PROBE=1

while [ $# -gt 0 ]; do
  case "$1" in
    --apply)        APPLY=1; shift ;;
    --no-probe)     PROBE=0; shift ;;
    --record-only)  RECORD_ONLY="${2:-}"; shift 2 || { echo "--record-only needs a filename" >&2; exit 2; } ;;
    --note)         NOTE="${2:-}"; shift 2 || { echo "--note needs text" >&2; exit 2; } ;;
    -h|--help)      sed -n '10,19p' "${BASH_SOURCE[0]}"; exit 0 ;;
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

# Advisory idempotency lint (bugs_open/007 candidate c). Echoes one warning line
# for a file that INSERTs into a non-log table while carrying no guard DO block,
# no ON CONFLICT and no WHERE NOT EXISTS — the shape that errors on replay. If
# the file has ANY guard we trust the author's handling and stay silent (the
# check is file-global, matching the handoff's spec and the shop convention that
# a migration guards itself). Warning only: some INSERTs are meant to fail loudly
# on a duplicate, so this never blocks and never changes the exit code.
lint_idempotency() {  # $1 = path to a .sql file; echoes a warning tail or nothing
  local file="$1" flat sinks t risky=()
  flat=$(tr '\n' ' ' < "$file")
  grep -qiE 'ON CONFLICT|WHERE NOT EXISTS' <<<"$flat" && return 0
  grep -qE 'DO \$\$|DO \$[A-Za-z_]' <<<"$flat" && return 0
  sinks=$(grep -oiP 'INSERT\s+INTO\s+\K[A-Za-z_][A-Za-z0-9_.]*' <<<"$flat" | sort -u)
  [ -z "$sinks" ] && return 0
  while IFS= read -r t; do
    [ -n "$t" ] || continue
    grep -qiE "$IDEMPOTENT_SINK_RE" <<<"$t" || risky+=("$t")
  done <<<"$sinks"
  [ ${#risky[@]} -eq 0 ] && return 0
  echo "INSERT into ${risky[*]} with no guard DO block, ON CONFLICT or WHERE NOT EXISTS"
}

# ---------------------------------------------------------------------- probe
# (bugs_open/007 residual 1.) Ask a pending file itself whether it has already
# been applied: run it verbatim inside a transaction that CANNOT commit. A
# deferred UNIQUE violation is planted before the file, so any COMMIT the file
# issues fails with 23505 naming probe_commit_poison and rolls everything back;
# ON_ERROR_STOP then stops psql so no later statement can run in autocommit.
# 86 of the 92 live migrations carry their own BEGIN/COMMIT, which is why a
# plain BEGIN..ROLLBACK wrap would be UNSAFE here — the poison closes that hole
# without parsing anyone's SQL. Mechanics proven against the live DB 2026-07-24
# (deferred violation fires at COMMIT; DO-block COMMIT errors 2D000; guard
# RAISE surfaces as "ERROR:  P0001: ..." under VERBOSITY verbose).
#
# The one structural blind spot: a top-level ROLLBACK/ABORT in the file would
# abort the doomed transaction, and a later BEGIN..COMMIT of its own would then
# run FOR REAL — four live files carry a defensive leading "ROLLBACK;"
# (136/180/195/205). Those, and anything else the doomed transaction cannot
# safely contain, are refused fail-closed: listed "not probed", never executed.

probe_refusal() {  # $1 = path; echoes a reason (=> do not probe) or nothing
  local flat
  flat=$(sed 's/--.*//' "$1")
  if grep -qiE '\b(ROLLBACK|ABORT)\b' <<<"$flat"; then
    echo "contains its own ROLLBACK/ABORT — would escape the doomed transaction"; return 0; fi
  if grep -vE '^[[:space:]]*\\set[[:space:]]+ON_ERROR_STOP\b' <<<"$flat" \
       | grep -qE '^[[:space:]]*\\|\\gexec|\\gset|\\copy'; then
    echo "contains psql meta-commands"; return 0; fi
  if grep -qiE 'PREPARE[[:space:]]+TRANSACTION' <<<"$flat"; then
    echo "uses two-phase commit"; return 0; fi
  if grep -qiE '\b(dblink|postgres_fdw)\b' <<<"$flat"; then
    echo "may write through a foreign connection the rollback cannot reach"; return 0; fi
  if grep -qiE 'SET[[:space:]]+CONSTRAINTS' <<<"$flat"; then
    echo "re-times constraint checks — could fire the poison early"; return 0; fi
  if grep -qiE '\bsetval[[:space:]]*\(' <<<"$flat"; then
    echo "setval is non-transactional and would persist through the rollback"; return 0; fi
  return 0
}

probe_file() {  # $1 = path; sets PROBE_VERDICT + PROBE_DETAIL
  local file="$1" out rc reason err tbl key
  PROBE_VERDICT=""; PROBE_DETAIL=""
  reason=$(probe_refusal "$file")
  if [ -n "$reason" ]; then
    PROBE_VERDICT="NOT_PROBED"; PROBE_DETAIL="$reason"; return 0
  fi
  out=$(
    {
      printf '%s\n' \
        '\set VERBOSITY verbose' \
        'BEGIN;' \
        'CREATE TEMP TABLE _probe_commit_poison(x int, CONSTRAINT probe_commit_poison UNIQUE(x) DEFERRABLE INITIALLY DEFERRED);' \
        'INSERT INTO _probe_commit_poison VALUES (1),(1);'
      cat "$file"
      printf '\n%s\n' 'ROLLBACK;'
    } | timeout 90 $PSQL_CMD -v ON_ERROR_STOP=1 2>&1
  )
  rc=$?
  err=$(grep -m1 '^ERROR:' <<<"$out" || true)
  if [ "$rc" = "124" ]; then
    PROBE_VERDICT="INCONCLUSIVE"; PROBE_DETAIL="probe exceeded 90s (transaction aborted server-side)"
  elif grep -q '^ERROR:  23505:' <<<"$out" && grep -q 'probe_commit_poison' <<<"$out"; then
    PROBE_VERDICT="CLEAN"; PROBE_DETAIL="ran to its own COMMIT without error (everything rolled back)"
  elif grep -qiE '^ERROR:  P0001:.*already' <<<"$out"; then
    PROBE_VERDICT="ALREADY"; PROBE_DETAIL="${err#ERROR:  P0001: }"
  elif grep -q '^ERROR:  23505:' <<<"$out"; then
    # VERBOSITY verbose makes Postgres name the table and key in the error
    # fields — surface them so the human check starts at the right row.
    tbl=$(grep -m1 '^TABLE NAME:' <<<"$out" | sed 's/^TABLE NAME:[[:space:]]*//')
    key=$(grep -m1 '^DETAIL:' <<<"$out" | sed 's/^DETAIL:[[:space:]]*//')
    PROBE_VERDICT="DUP"
    PROBE_DETAIL="${err#ERROR:  }${tbl:+ [table ${tbl}${key:+; ${key}}]} — already applied out of band, OR genuinely wrong SQL (key/numbering collision); inspect the existing row before deciding"
  elif [ -n "$err" ]; then
    PROBE_VERDICT="INCONCLUSIVE"; PROBE_DETAIL="$err"
  elif [ "$rc" != "0" ]; then
    PROBE_VERDICT="INCONCLUSIVE"; PROBE_DETAIL="probe exited $rc"
  else
    PROBE_VERDICT="CLEAN"; PROBE_DETAIL="ran to end without error (no COMMIT of its own; rolled back)"
  fi
  return 0
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

# Advisory idempotency lint over the pending set (bugs_open/007 candidate c).
# Doubles as an early warning for the ledger trap itself: a pending file that
# looks non-idempotent is exactly the one that errors if it was already applied
# by hand. Warnings go to stderr; they never change what runs or the exit code.
LINT=()
for f in "${PENDING[@]}"; do
  msg=$(lint_idempotency "$MIGRATIONS_DIR/$f")
  [ -n "$msg" ] && LINT+=("  $f — $msg")
done
if [ ${#LINT[@]} -gt 0 ]; then
  echo "" >&2
  echo "IDEMPOTENCY WARNING (advisory, bugs_open/007) — these pending files may error on replay:" >&2
  printf '%s\n' "${LINT[@]}" >&2
  echo "  If a file above is ALREADY applied and merely unrecorded, this IS the replay hazard —" >&2
  echo "  verify its artifacts and use --record-only, do not --apply." >&2
fi

# Probe every pending file once, report, and cache the verdicts — --apply
# reuses them rather than probing twice. See the probe block above for why a
# probing run is not purely read-only (locks + rolled-back execution).
declare -A PROBE_V=() PROBE_D=()
SKIPPED=()
# Files this run applied that touch the truncation contract (bugs_closed/076 R1).
# Collected so the pointer at the end names them; see the note by that pointer for
# why this is a pointer and not a check.
TRUNC_TOUCHED=()
if [ "$PROBE" = "1" ]; then
  echo ""
  echo "Probe (each pending file executed in a doomed transaction, rolled back; --no-probe skips):"
  for f in "${PENDING[@]}"; do
    probe_file "$MIGRATIONS_DIR/$f"
    PROBE_V["$f"]="$PROBE_VERDICT"
    PROBE_D["$f"]="$PROBE_DETAIL"
    case "$PROBE_VERDICT" in
      ALREADY)
        echo "  !! $f — LIKELY ALREADY APPLIED; its own guard raised:"
        echo "       ${PROBE_DETAIL}"
        echo "       Verify its artifacts in the DB, then: $0 --record-only $f --note '<what you checked>'" ;;
      CLEAN)        echo "  ok $f — ${PROBE_DETAIL}" ;;
      DUP)          echo "  !! $f — DUPLICATE KEY during probe: ${PROBE_DETAIL}"
                    echo "       --apply will REFUSE this file. If it was applied out of band: verify artifacts, then --record-only." ;;
      NOT_PROBED)   echo "  -- $f — not probed: ${PROBE_DETAIL}" ;;
      INCONCLUSIVE) echo "  ?? $f — probe inconclusive: ${PROBE_DETAIL}" ;;
    esac
  done
fi

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
  # Skip-not-record (bugs_open/007 residual 1): the file's OWN guard declared it
  # already applied, so replaying it would only halt the run. Skipping is safe —
  # the halt it replaces produced the identical end state, minus every file
  # after it. Recording is NOT safe to automate, so the file stays pending (and
  # gets skipped again) until a human verifies artifacts and records it.
  if [ "${PROBE_V[$f]:-}" = "ALREADY" ]; then
    echo "== $f SKIPPED — its own guard says it has already been applied:"
    echo "   ${PROBE_D[$f]}"
    echo "   NOT recorded. It stays pending until you verify its artifacts and run:"
    echo "     $0 --record-only $f --note 'applied by hand <date>; <what you checked>'"
    SKIPPED+=("$f")
    continue
  fi
  # A DUP probe verdict can NEVER auto-skip — the duplicate may be an out-of-
  # band apply OR genuinely wrong SQL (numbering collisions are live in this
  # directory: two 203s exist), and only a human inspecting the existing row
  # can tell. But executing the file would just crash into the same duplicate
  # the probe already hit seconds ago. Halt BEFORE running it, with the
  # classification done and the table/key named. Same exit code, same
  # die-loudly, later files equally not attempted; --no-probe --apply
  # restores the old crash-into-it behaviour.
  if [ "${PROBE_V[$f]:-}" = "DUP" ]; then
    echo "!! $f REFUSED — the probe hit a duplicate key before this file's own COMMIT:" >&2
    echo "   ${PROBE_D[$f]}" >&2
    echo "   Nothing applied (the probe's execution rolled back); later files not attempted." >&2
    echo "   If it was applied out of band: verify its artifacts in the DB, then" >&2
    echo "     $0 --record-only $f --note 'applied by hand <date>; <what you checked>'" >&2
    echo "   If the SQL is genuinely wrong (key collision): fix the file or the colliding row." >&2
    echo "   Only record it if you CHECKED — a ledger row for a migration that never" >&2
    echo "   ran skips it for ever. See bugs_open/007." >&2
    exit 1
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
  if grep -q 'tolerate_truncation\|accepts_truncated' "$MIGRATIONS_DIR/$f" 2>/dev/null; then
    TRUNC_TOUCHED+=("$f")
  fi
done
if [ ${#SKIPPED[@]} -gt 0 ]; then
  echo ""
  echo "Skipped as likely-already-applied — each stays PENDING until verified and recorded:"
  printf '  %s\n' "${SKIPPED[@]}"
fi
# Truncation contract pointer (bugs_closed/076 R1). A step may only KEEP an LLM
# response the model cut at max_tokens where another step in ITS workflow reads the
# __truncated marker; otherwise a fragment is indistinguishable from a complete
# answer. The runtime guard only says so when a response is ACTUALLY cut, in
# production, by failing that run.
#
# A POINTER, NOT A CHECK, deliberately. The offending config usually arrives here
# as a jsonb_set patch (all three files in the repo that arm the flag are patches),
# and a patch names the flag but not the workflow — so nothing this script can read
# from the FILE settles it. The live lint resolves the real workflow in one
# read-only command; this just says "now is the moment". Running it from here was
# rejected: it would put a python + kubectl dependency in the middle of an apply
# that has already committed, and a failure there would read as a migration failure.
if [ ${#TRUNC_TOUCHED[@]} -gt 0 ]; then
  echo ""
  echo "Applied file(s) touching the truncation contract (bugs_closed/076):"
  printf '  %s\n' "${TRUNC_TOUCHED[@]}"
  echo "  Confirm every tolerating step still sits in a workflow that reads the marker:"
  echo "    python3 docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/103_LINT_truncation_consumer.py"
fi
echo ""
echo "Done."
