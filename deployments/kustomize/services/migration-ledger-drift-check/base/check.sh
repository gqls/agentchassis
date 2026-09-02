#!/usr/bin/env bash
# check.sh — migration-ledger-drift-check (bugs_open/426).
#
# WHY THIS EXISTS. `scripts/migration/run-migrations.sh`'s default, no-argument,
# DRY RUN already computes exactly what this estate needs: every appliable
# migration with no `schema_migrations` row, each probed (executed in a doomed,
# self-rolling-back transaction) to say whether its own guard thinks it has
# already been applied by hand. CLAUDE.md already mandates running that dry run
# "per session and after every roll". Nobody drove it on a schedule, so the
# signal only ever surfaced when a human happened to run it interactively — seven
# real instances of "applied by hand, never recorded" sat unnoticed for a month
# before one lane found them by chance auditing itself (bugs_open/426). A live
# re-run the day this was built found THIRTY-FOUR, not seven — see this
# workstream's NOTES for the measurement. This job puts that free check on a
# clock and writes ONE doc_notes row every day, including on a clean day, so a
# MISSING row means "the job didn't run" and is never mistaken for "all clear"
# (the RFC_022 lesson, learned the hard way elsewhere on this estate).
#
# WHY NOT A SECOND IMPLEMENTATION OF THE MIGRATION VOCABULARY. bugs_open/426 §4
# rules this out explicitly, and bugs_closed/314 is the standing case for why: the
# naming pattern, sidecar exclusion and baseline number have already drifted
# across independent copies once. So this script does NOT decide what counts as a
# migration, a sidecar, or "pending" — it fetches the REAL, CURRENT
# `run-migrations.sh` and the REAL, CURRENT migrations directory from the pinned
# ref and runs the UNMODIFIED script, then only PARSES its already-existing
# output vocabulary (the `!!`/`??`/`--`/`ok`/`Pending (N)` line prefixes the
# runner itself prints). A person who changes the runner's rules never has to
# remember a second place to change them.
#
# WHY A SPARSE GIT CLONE, NOT THE GITHUB CONTENTS API (bugs-open-staleness-
# sweep's own mechanism). That sweep only ever needs file EXISTENCE and LINE
# COUNTS — metadata. This check needs to EXECUTE the real SQL text of every
# pending file (that's how the probe gets its signal), and there is no way to
# know in advance which files are pending without asking the runner itself,
# which means materialising the WHOLE directory, not a fetched-per-file subset.
# `docs/agent_docs/sql_for_agents/` is 16MB / ~1,150 files (measured when this
# was built) — a `--filter=blob:none --sparse --depth 1` clone of just that dir
# plus `scripts/migration/` fetches only those blobs, no history, and completed
# in 1.8s against the real repo from inside this namespace when this was tested.
# That is nothing like the 262M-`.git` problem `bugs-open-staleness-sweep`'s own
# comment rules a full clone out for — that concern is about HISTORY, and
# `--depth 1` fetches none.
#
# WHY DIRECT psql, NOT `kubectl exec` (`run-migrations.sh`'s own PSQL_CMD
# default). This job's service account has no pods/exec RBAC in
# ai-persona-system — same constraint every sibling CronJob in this directory
# already works around by connecting to Postgres directly with a secret. This
# script overrides PSQL_CMD and MIGRATIONS_DIR, the two env vars
# `run-migrations.sh` already exposes for exactly this purpose, and otherwise
# touches nothing.
#
# NEEDS_REVIEW bucketing (added at the request of the dispatch_throughput lane,
# 2026-09-02, credited in this workstream's NOTES): the runner's ALREADY verdict
# only fires when a guard's RAISE message happens to match `/already/i` — an
# undocumented vocabulary contract (bugs_open/426 §3's own worked case, `672`:
# a truthful "drifted, investigate" message that actually means "applied by
# hand, fine"). So a hand-applied-unrecorded file can just as easily land in
# INCONCLUSIVE as in ALREADY. This script treats ALREADY + DUP + INCONCLUSIVE
# as one NEEDS REVIEW bucket (sub-labelled, so a human can still tell them
# apart), and only CLEAN + NOT_PROBED as non-actionable. Fixing the runner's
# probe to signal structurally instead of by prose is a real, separate, higher-
# risk change (it touches every migration guard's convention) — deferred, see
# bugs_open/426 §5 candidate 2 and this workstream's PLAN.
#
# This script NEVER runs --record-only and NEVER calls --apply. Recording stays
# a human act (verify the artefact, then record) — this is the runner's own
# stated contract and this job only ever calls the default dry run.
set -uo pipefail

env_required() {
  local name="$1" val="${!1:-}"
  if [ -z "$val" ]; then
    echo "REFUSING TO RUN: $name is not set." >&2
    exit 2
  fi
}
env_required MIGRATION_CHECK_REF
env_required REPO_OWNER
env_required REPO_NAME
env_required PG_CLIENTS_HOST
env_required CLIENTS_DB_PASSWORD
env_required GITHUB_READ_TOKEN

REPO_DIR=/tmp/repo
REPORT=/tmp/run-migrations-report.txt

echo "== cloning ${REPO_OWNER}/${REPO_NAME}@${MIGRATION_CHECK_REF} (sparse, shallow, migrations dir only) =="
AUTH_B64=$(printf 'x-access-token:%s' "$GITHUB_READ_TOKEN" | base64 | tr -d '\n')
git -c http.extraHeader="Authorization: Basic ${AUTH_B64}" \
  clone --filter=blob:none --sparse --depth 1 --branch "$MIGRATION_CHECK_REF" \
  "https://github.com/${REPO_OWNER}/${REPO_NAME}.git" "$REPO_DIR" \
  || { echo "REFUSING TO RUN: clone failed (bad ref, bad token, or network)." >&2; exit 2; }
(cd "$REPO_DIR" && git sparse-checkout set docs/agent_docs/sql_for_agents scripts/migration)

RUNNER="$REPO_DIR/scripts/migration/run-migrations.sh"
if [ ! -f "$RUNNER" ]; then
  echo "REFUSING TO RUN: $RUNNER not found after sparse checkout." >&2
  exit 2
fi

echo "== running the unmodified dry run =="
export PSQL_CMD="psql -h ${PG_CLIENTS_HOST} -p 5432 -U clients_user -d clients_db"
export PGPASSWORD="$CLIENTS_DB_PASSWORD"
export MIGRATIONS_DIR="$REPO_DIR/docs/agent_docs/sql_for_agents"
# The dry run (no --apply) always exits 0 regardless of findings — see
# run-migrations.sh's own final block — so this script's own exit code is
# decided below, from parsing the report, not from $?.
bash "$RUNNER" >"$REPORT" 2>&1
echo "== dry run complete, report is $(wc -l <"$REPORT") lines =="

# A successful dry run always ends with one of these two lines (its own final
# block). Anything else — DB unreachable, a probe timeout cascading into a
# hard stop, an unexpected script error — means there is no real report to
# summarise. Refuse to write a doc_notes row in that case: a false "up to
# date" from an empty/garbage report would be exactly the failure this whole
# job exists to prevent (a broken check reading identical to a clean one).
if ! grep -qE '^(Pending \([0-9]+\):|Up to date —)' "$REPORT"; then
  echo "REFUSING TO WRITE A REPORT: the dry run did not end the way a successful" >&2
  echo "run always does. Raw output:" >&2
  cat "$REPORT" >&2
  exit 2
fi

# ---------------------------------------------------------------- parse
PENDING_N=$(grep -oE 'Pending \([0-9]+\)' "$REPORT" | grep -oE '[0-9]+' || true)
[ -n "$PENDING_N" ] || PENDING_N=0

CLEAN_N=$(grep -c '^  ok ' "$REPORT" || true)
NOT_PROBED_N=$(grep -c '^  -- ' "$REPORT" || true)

# BUCKET<TAB>filename<TAB>detail, one row per NEEDS-REVIEW entry. Multi-line
# entries (ALREADY/DUP: verdict on one line, guard detail on the next) and
# single-line entries (INCONCLUSIVE: detail on the same line) both handled.
PARSED=$(awk '
  /^  !! / {
    line = $0
    getline detail
    gsub(/^[ \t]+/, "", detail)
    n = split(line, parts, " — ")
    fn = parts[1]
    sub(/^  !! /, "", fn)
    bucket = (line ~ /DUPLICATE KEY/) ? "DUP" : "ALREADY"
    printf "%s\t%s\t%s\n", bucket, fn, detail
    next
  }
  /^  \?\? / {
    line = $0
    idx = index(line, " — probe inconclusive: ")
    fn = substr(line, 1, idx - 1)
    sub(/^  \?\? /, "", fn)
    detail = substr(line, idx + length(" — probe inconclusive: "))
    printf "INCONCLUSIVE\t%s\t%s\n", fn, detail
    next
  }
' "$REPORT")

NEEDS_REVIEW_N=0
[ -n "$PARSED" ] && NEEDS_REVIEW_N=$(printf '%s\n' "$PARSED" | grep -c . || true)

# ------------------------------------------------------------- render
{
  echo "MIGRATION LEDGER DRIFT CHECK (bugs_open/426)"
  echo ""
  echo "ref checked:        ${MIGRATION_CHECK_REF}"
  echo "pending migrations: ${PENDING_N}"
  echo "  clean (genuinely not yet applied): ${CLEAN_N}"
  echo "  not probed (refused fail-closed):  ${NOT_PROBED_N}"
  echo "  NEEDS REVIEW (see below):          ${NEEDS_REVIEW_N}"
  echo ""
  if [ "$NEEDS_REVIEW_N" -eq 0 ]; then
    if [ "$PENDING_N" -eq 0 ]; then
      echo "Up to date — nothing pending."
    else
      echo "All ${PENDING_N} pending file(s) look genuinely not-yet-applied — no probe"
      echo "suggests any is already applied by hand."
    fi
    echo ""
    echo "This row exists on a clean run ON PURPOSE: a MISSING row means the job did"
    echo "not run, which is not the same as 'nothing is wrong', and the two must not"
    echo "look alike (bugs_open/426, RFC_022's lesson)."
  else
    echo "NEEDS REVIEW — a pending file's own guard suggests it may already have been"
    echo "applied by hand and never recorded. Verify each one's artefacts in the DB,"
    echo "then: scripts/migration/run-migrations.sh --record-only <file> --note '<what you checked>'"
    echo ""
    echo "Three sub-buckets, most to least confident (bugs_open/426 §3; the second and"
    echo "third are as actionable as the first — a hand-applied-unrecorded file can"
    echo "land in any of them depending on its guard's exact wording, not just ALREADY):"
    echo ""
    for b in ALREADY DUP INCONCLUSIVE; do
      rows=$(printf '%s\n' "$PARSED" | awk -F'\t' -v b="$b" '$1==b')
      [ -z "$rows" ] && continue
      echo "  -- ${b} --"
      printf '%s\n' "$rows" | while IFS=$'\t' read -r bucket fn detail; do
        echo "  ${fn}"
        echo "      ${detail}"
      done
      echo ""
    done
  fi
} >/tmp/rendered-report.txt

cat /tmp/rendered-report.txt

# ------------------------------------------------------------ doc_notes
SQLFILE=/tmp/write-note.sql
{
  printf "INSERT INTO doc_notes (subject_type, subject_key, body, categories, source)\n"
  printf "VALUES ('pipeline', 'migration-ledger-drift', \$mldcheck\$"
  cat /tmp/rendered-report.txt
  printf "\$mldcheck\$, '[\"migration-ledger-drift\"]'::jsonb, 'migration-ledger-drift-check');\n"
} >"$SQLFILE"
if ! $PSQL_CMD -v ON_ERROR_STOP=1 -f "$SQLFILE" >/dev/null; then
  echo "" >&2
  echo "REFUSING TO CLAIM SUCCESS: the doc_notes INSERT failed — this run's whole" >&2
  echo "point is a row a missing-row check can trust, so a failed write must not" >&2
  echo "report as though it happened." >&2
  exit 2
fi
echo ""
echo "doc_notes row written (subject_type='pipeline', subject_key='migration-ledger-drift')."
echo "NEEDS_REVIEW_N=${NEEDS_REVIEW_N}"

# Exit 0 whenever the check itself completed and the report was written,
# REGARDLESS of findings. Findings are the ordinary, expected, daily state on
# this estate (measured 2026-09-02: 72 of 158 pending, worked out live on this
# very job's first real run — see NOTES) — they are not this job failing, and
# treating them as failure was tried first and rejected: it made
# concurrencyPolicy Forbid + the jobTemplate's own backoffLimit re-run the
# WHOLE clone-and-probe a second time, every single day, and left the Job
# permanently red in a way nobody would ever clear (findings don't go away on
# retry, so every run "fails" for ever). The doc_notes row — not the Job's own
# exit code — is the durable, query-able signal this estate's checks are
# already built to be read by (RFC_022's convention: missing row = didn't run;
# present row = ran, read its body for findings). Genuine operational failure
# (clone/env/DB/write) already exits 2 above, well before this point, and
# THAT is what backoffLimit's retry exists to recover from.
exit 0
