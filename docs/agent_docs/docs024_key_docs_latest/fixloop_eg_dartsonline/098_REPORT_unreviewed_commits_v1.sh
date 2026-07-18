#!/usr/bin/env bash
# 098_REPORT_unreviewed_commits_v1.sh — council-gate VISIBILITY report
# (build-order step 3: visibility BEFORE enforcement, the standing rule).
#
# WHAT. A deterministic join of git history against council verdicts: every
# commit in the window touching the review scope (platform/, internal/, pkg/ —
# owner ruling 2026-07-17) is bucketed as
#   REVIEWED   — carries a "Council-Reviewed: <correlation>" trailer AND the
#                latest council_report on that correlation says approved;
#   MISMATCH   — carries the trailer but the report is missing or not approved
#                (a claim of review the artifacts do not back);
#   UNREVIEWED — no trailer at all.
# No LLM anywhere (digest rule: an awareness surface that could hallucinate
# what happened would defeat itself). No enforcement: this script only reports.
#
# WHY REPO-SIDE, not a digest section: the digest action runs in the chassis
# pod, which has no git repository — the join can only happen where git is.
# PERSIST=1 writes the report to doc_notes (pipeline/diagnose, categories
# digest+council-gate) so it rides the same awareness channel the digest does.
#
# HONEST LIMITS (say them, don't design around them silently):
#   * The trailer is voluntary in advisory mode — this report is exactly the
#     visibility that makes the gap measurable, nothing more.
#   * Runs against the CURRENT branch's history. Another session may commit
#     while this runs; the report names its window and HEAD so it is a
#     snapshot, not a ledger.
#   * DB verification needs kubectl access to postgres-clients-0
#     (ai-persona-system). NO_DB=1 skips it: trailered commits then bucket as
#     TRAILER-UNVERIFIED rather than silently passing as REVIEWED.
#
# Usage:
#   ./098_REPORT_unreviewed_commits_v1.sh                # last 7 days
#   ./098_REPORT_unreviewed_commits_v1.sh 14             # last 14 days
#   NO_DB=1 ./098_REPORT_unreviewed_commits_v1.sh        # no verdict lookup
#   PERSIST=1 ./098_REPORT_unreviewed_commits_v1.sh      # also write doc_note
set -euo pipefail

WINDOW_DAYS="${1:-7}"
SCOPE_PATHS=(platform internal pkg)
NS='ai-persona-system'
PG_POD='postgres-clients-0'

REPO_ROOT=$(git rev-parse --show-toplevel 2>/dev/null) \
  || { echo "ERROR: not inside a git repository" >&2; exit 1; }
cd "$REPO_ROOT"
BRANCH=$(git rev-parse --abbrev-ref HEAD)
HEAD_SHA=$(git rev-parse --short HEAD)
NOW_UTC=$(date -u '+%Y-%m-%d %H:%M UTC')

# Resolve a trailer id to its latest council verdict. The id may be EITHER a
# submission/diagnosis correlation_id (what 097 prints) OR the orchestration id
# of a council RUN (what a fix-proposer run prints as RUN_ORCH_ID) — threads
# legitimately have both, and v1 only matched the first, so it reported a
# genuinely-approved fix-proposer commit as a false claim of review (f32b208e5,
# 2026-07-18: trailer 53da3a30 was the run whose council approved after two
# revise rounds). Prefix matching too: threads paste short ids.
# Prints "<decision>|<matched-on>", or "" if nothing matched.
db_decision() { # $1 = trailer id (raw, from a commit message — sanitised here)
  local id
  id=$(printf '%s' "$1" | tr -cd '0-9a-fA-F-')   # commit text never reaches SQL unsanitised
  [ "${#id}" -ge 8 ] || { echo ""; return; }
  # NO `-i`, and stdin closed: this runs INSIDE the `while read` loop below, and
  # `kubectl exec -i` consumes that loop's stdin — the report then stops dead at
  # the FIRST commit carrying a trailer. It was silently reporting 4 of 41
  # in-scope commits (proven 2026-07-18: NO_DB=1, which skips this call, printed
  # the full 41). The SQL is passed with -c, so stdin was never needed.
  kubectl -n "$NS" exec "$PG_POD" -- \
    psql -U clients_user -d clients_db -tA -c \
    "SELECT COALESCE(metadata->>'decision','') || '|' ||
            CASE WHEN correlation_id LIKE '${id}%' THEN 'correlation' ELSE 'run' END
     FROM diagnosis_artifacts
     WHERE kind = 'council_report'
       AND (correlation_id LIKE '${id}%' OR orchestration_id LIKE '${id}%')
     ORDER BY created_at DESC LIMIT 1;" 2>/dev/null | tr -d '[:space:]'
}

reviewed=(); mismatch=(); missing=(); unreviewed=(); unverified=()
count=0

while IFS='|' read -r sha short date subject; do
  [ -n "$sha" ] || continue
  count=$((count+1))
  corr=$(git show -s --format=%B "$sha" | sed -n 's/^[Cc]ouncil-[Rr]eviewed:[[:space:]]*//p' | head -1 | tr -d '[:space:]')
  line="$short  $date  $subject"
  if [ -z "$corr" ]; then
    unreviewed+=("$line")
  elif [ "${NO_DB:-0}" = "1" ]; then
    unverified+=("$line  [trailer: $corr]")
  else
    resolved=$(db_decision "$corr" || true)
    decision="${resolved%%|*}"; matched="${resolved#*|}"
    if [ "$decision" = "approved" ]; then
      reviewed+=("$line  [${corr:0:8}, by ${matched}]")
    elif [ -z "$resolved" ]; then
      # No report row AT ALL is not the same accusation as "reviewed and not
      # approved". Council reports are deletable, and a documented practice
      # deletes them (091's "clear them first for a fair run"), so an honestly
      # reviewed commit can lose its evidence later — proven 2026-07-18:
      # f32b208e5 resolved as approved at 12:03 and as NO REPORT at 13:29,
      # because its run's reports were cleared in between. Do not call that a
      # false claim.
      missing+=("$line  [trailer: $corr -> evidence gone (cleared or expired)]")
    else
      mismatch+=("$line  [trailer: $corr -> $decision]")
    fi
  fi
done < <(git log --since="${WINDOW_DAYS} days ago" --pretty='%H|%h|%ad|%s' --date=short -- "${SCOPE_PATHS[@]}")

print_bucket() { # $1 = title, then lines from the named array (via nameref)
  local title="$1"; shift
  local -n arr="$1"
  echo "### ${title}: ${#arr[@]}"
  if [ "${#arr[@]}" -eq 0 ]; then
    echo "(none in window)"
  else
    printf '%s\n' "${arr[@]}"
  fi
  echo ""
}

REPORT=$(
  echo "# Un-reviewed platform commits — council-gate visibility report"
  echo ""
  echo "Window: last ${WINDOW_DAYS} days, ending ${NOW_UTC}. Branch: ${BRANCH} @ ${HEAD_SHA}."
  echo "Scope: commits touching ${SCOPE_PATHS[*]} (owner ruling 2026-07-17)."
  echo "In-scope commits found: ${count}. Advisory mode: this is visibility, not enforcement."
  echo ""
  print_bucket "REVIEWED (trailer + approved council_report)" reviewed
  if [ "${NO_DB:-0}" = "1" ]; then
    print_bucket "TRAILER-UNVERIFIED (NO_DB=1 — verdicts not checked)" unverified
  else
    print_bucket "MISMATCH (trailer + a report that did NOT approve)" mismatch
    print_bucket "EVIDENCE GONE (trailer, but no report row — cleared/expired, NOT a false claim)" missing
  fi
  print_bucket "UNREVIEWED (no Council-Reviewed trailer)" unreviewed
)

printf '%s\n' "$REPORT"

if [ "${PERSIST:-0}" = "1" ]; then
  if [ "${NO_DB:-0}" = "1" ]; then
    echo "ERROR: PERSIST=1 needs the DB — drop NO_DB" >&2; exit 1
  fi
  TAG='$cgate098$'
  kubectl -n "$NS" exec -i "$PG_POD" -- \
    psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 <<SQL
INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('pipeline', 'diagnose', ${TAG}${REPORT}${TAG},
        '["digest","council-gate"]'::jsonb, '098_report', 'cli');
SQL
  echo ""
  echo "Persisted to doc_notes (categories digest+council-gate)."
fi
