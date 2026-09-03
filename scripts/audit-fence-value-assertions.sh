#!/bin/bash
# Read-only. Which tool acceptance fences never assert a NUMBER?
#
# bugs_open/449. The ladder implements a check type that compares values
# (`computed_values`) and neither fence-authoring agent has ever been told it
# exists, so a calculator printing a confidently wrong figure passes Tier 4 and
# its record reads PASSED. Measured 2026-09-03: `tool-generator` had written
# 186 current fences, 115 asserting no expected value of any kind, 0 using
# `computed_values`.
#
# WHY A WINDOW IS THE POINT, AND A TOTAL IS NOT.
#
# 449 §6: "Compare by created_at window, not by total — the existing ones do not
# change themselves." A fix to the AUTHORING side shows up as a fall in NEW
# fences and leaves the total nearly flat for weeks, because ~115 blind fences
# sit there unchanged until something rewrites them. A report printing only
# totals would therefore read as "no improvement" for a month after a working
# fix, and would be quietly abandoned as useless. So the window is the headline
# and the total is context.
#
# THE DEMAND CONTROL IS BUILT IN, and it is not decoration. A census like this
# has two ways to return a comfortable number: the corpus is clean, or the
# query is blind (a changed fence shape, a renamed column, a `LIKE` that stopped
# matching). Those are the same bytes. So before reporting anything, the check
# requires that SOME author still shows a non-zero `uses_computed_values` — the
# operator lanes carry one on every fence (449 §2), so a fleet-wide zero there
# means this query no longer sees what it is looking for, and the check exits 2
# rather than reporting a clean corpus.
#
# ⚠ THE CONTROL IS DELIBERATELY NAME-FREE, and that was a correction made within
# minutes of the first run. It originally named `operator:mortgagecalculator-lane-a4`
# and `operator:bugfix224-session` explicitly, on the reasoning that a specific
# control is a stronger one. It is — and `created_by` is a FREE-TEXT LANE LABEL,
# not an identity. The mcalc lane re-keyed its eight fences the same morning this
# was written and that author became
# `operator:mortgagecalculator-lane-2026-09-03-701-rekey`; the control survived
# only because the OTHER hard-coded name happened to still exist. A control that
# fails when a lane renames itself does not detect a blind query, it detects a
# rename — and it fails CLOSED, so it would have exited 2 and looked like a
# broken census on a day when nothing was wrong.
#
# THE TRIGGER IS READ OFF THE FENCE, NEVER FROM A CLASSIFIER. "Is this tool a
# calculator" needs a judgement about tool kinds and inherits that judgement's
# gaps. A check carrying a `fill` or `select` step has DECLARED that the tool
# takes input, so `drives_but_asserts_nothing` is a fact in the document. This
# mirrors summariseCriteriaValueAssertions (criteria_value_assertions.go)
# exactly — but ⚠ it is a SQL approximation of it, not the same code: this uses
# LIKE over the extracted fence where the Go parses it. The SQL is the LOOSER
# of the two (a `"fill"` anywhere in the fence text counts), which is the safe
# direction for a report — it over-selects and a human decides — but it is NOT
# parity and must not be described as such.
#
# Usage: scripts/audit-fence-value-assertions.sh [--days N] [--json]
#        --days N   window for the headline (default 7)
# Exit:  0 = no NEW blind fences in the window · 1 = new blind fences found
#        2 = could not determine (includes the demand control failing)

set -uo pipefail

NAMESPACE="${NAMESPACE:-ai-persona-system}"
DAYS=7
JSON_OUT=0
while [[ $# -gt 0 ]]; do
    case "$1" in
        --days) DAYS="${2:-7}"; shift 2 ;;
        --json) JSON_OUT=1; shift ;;
        *) echo "unknown argument: $1" >&2; exit 2 ;;
    esac
done
[[ "$DAYS" =~ ^[0-9]+$ ]] || { echo "--days needs an integer, got $DAYS" >&2; exit 2; }

# stderr deliberately NOT swallowed — a kubectl/psql failure must look like a
# failure, not like an empty corpus (audit-config-keys.sh's note, same reason).
psql_q() {
    kubectl -n "$NAMESPACE" exec -i postgres-clients-0 -- \
        psql -U clients_user -d clients_db -tAF$'\t' -c "$1"
}

# The fence is the FIRST ```criteria block, non-greedy — mirroring
# extractCriteriaFence (check_tool_acceptance.go), which takes the first one.
# A greedy match swallows the rest of the document and every LIKE after it
# silently changes meaning.
read -r -d '' CENSUS_SQL <<'SQL'
WITH f AS (
  SELECT dp.created_by, dp.created_at,
         substring(dp.body from '```criteria(.*?)```') AS fence
    FROM doc_plans dp
   WHERE dp.subject_type = 'tool' AND dp.is_current
     AND dp.body LIKE '%```criteria%'
), g AS (
  SELECT COALESCE(created_by, '(null)') AS author,
         created_at,
         (fence LIKE '%"fill"%' OR fence LIKE '%"select"%')      AS drives,
         (fence LIKE '%text_matches%' OR fence LIKE '%expect_values%') AS asserts,
         (fence LIKE '%expect_values%')                          AS computed
    FROM f
)
SELECT author,
       count(*),
       count(*) FILTER (WHERE NOT asserts),
       count(*) FILTER (WHERE drives AND NOT asserts),
       count(*) FILTER (WHERE computed),
       count(*) FILTER (WHERE created_at > now() - ($1 || ' days')::interval),
       count(*) FILTER (WHERE created_at > now() - ($1 || ' days')::interval
                          AND drives AND NOT asserts),
       to_char(max(created_at), 'YYYY-MM-DD')
  FROM g GROUP BY 1 ORDER BY 2 DESC;
SQL
CENSUS_SQL="${CENSUS_SQL//\$1/$DAYS}"

ROWS="$(psql_q "$CENSUS_SQL")" || { echo "audit-fence-value-assertions: census query failed" >&2; exit 2; }
[[ -n "$ROWS" ]] || { echo "audit-fence-value-assertions: census returned NO ROWS — there is always at least one tool fence, so this is a broken query, not a clean corpus" >&2; exit 2; }

# ── the demand control, before any finding is reported ──────────────────────
CONTROL_OK=0
CONTROL_AUTHOR=""
while IFS=$'\t' read -r author total blind drivesblind computed win winblind last; do
    if [[ "${computed:-0}" -gt 0 ]]; then
        CONTROL_OK=1
        [[ -n "$CONTROL_AUTHOR" ]] || CONTROL_AUTHOR="$author"
    fi
done <<< "$ROWS"
if [[ $CONTROL_OK -eq 0 ]]; then
    echo "audit-fence-value-assertions: DEMAND CONTROL FAILED — NO author anywhere shows a computed_values fence." >&2
    echo "  Several operator lanes carry one on every fence (449 §2), so a fleet-wide zero means this query no" >&2
    echo "  longer sees what it is looking for — a changed fence shape, a renamed column, a LIKE that stopped" >&2
    echo "  matching. Refusing to report a clean corpus off a blind census." >&2
    exit 2
fi

TOTAL_NEW_BLIND=0
if [[ $JSON_OUT -eq 1 ]]; then
    printf '{"window_days":%s,"authors":[' "$DAYS"
    first=1
    while IFS=$'\t' read -r author total blind drivesblind computed win winblind last; do
        [[ $first -eq 1 ]] || printf ','
        first=0
        printf '{"author":"%s","fences":%s,"asserts_no_value":%s,"drives_but_asserts_nothing":%s,"uses_computed_values":%s,"created_in_window":%s,"new_blind_in_window":%s,"newest":"%s"}' \
            "$author" "$total" "$blind" "$drivesblind" "$computed" "$win" "$winblind" "$last"
        TOTAL_NEW_BLIND=$(( TOTAL_NEW_BLIND + winblind ))
    done <<< "$ROWS"
    printf '],"new_blind_in_window":%s,"demand_control":"passed","control_author":"%s"}\n' "$TOTAL_NEW_BLIND" "$CONTROL_AUTHOR"
else
    echo "Tool acceptance fences that never assert a number — bugs_open/449"
    echo "Window: last ${DAYS} day(s). THE WINDOW IS THE HEADLINE; the totals are context,"
    echo "because the standing stock does not change itself and would read as 'no improvement'."
    echo
    printf '%-40s %7s %7s %8s %9s | %6s %8s  %s\n' \
        "author" "fences" "blind" "drv+blind" "computed" "new" "new-blind" "newest"
    printf '%.0s-' {1..110}; echo
    while IFS=$'\t' read -r author total blind drivesblind computed win winblind last; do
        printf '%-40s %7s %7s %8s %9s | %6s %8s  %s\n' \
            "$author" "$total" "$blind" "$drivesblind" "$computed" "$win" "$winblind" "$last"
        TOTAL_NEW_BLIND=$(( TOTAL_NEW_BLIND + winblind ))
    done <<< "$ROWS"
    echo
    echo "demand control: PASSED — ${CONTROL_AUTHOR} still shows computed_values, so the census is not blind"
    if [[ $TOTAL_NEW_BLIND -gt 0 ]]; then
        echo
        echo "FINDING: ${TOTAL_NEW_BLIND} fence(s) created in the last ${DAYS} day(s) drive a tool's inputs and assert NO value."
        echo "  The intake is still open. Each is a tool whose Tier-4 PASS will mean 'it responded', not 'it is right'."
        echo "  Cause: neither fence-authoring agent knows the computed_values type (449 §3)."
    else
        echo
        echo "No NEW blind fences in the window. Note this says nothing about the standing stock above,"
        echo "which is repaired per site with a per-site oracle, not by this check."
    fi
fi

[[ $TOTAL_NEW_BLIND -gt 0 ]] && exit 1
exit 0
