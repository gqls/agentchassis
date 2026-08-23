#!/bin/bash
# Read-only. Which keyed create_work_item step has never declared whether the
# item it files is an ACTION REQUEST or a DETECTED DEFECT?
#
# writeWorkItem runs an anti-churn brake over (site_id, item_key): siblings that
# reached complete/failed inside 7 days slow the key down. Right for a DETECTED
# DEFECT, where a repeat means the fix is not working. Wrong for an ACTION
# REQUEST — a stage handoff, a re-render request, a re-submission — where a
# `complete` predecessor means the previous one SUCCEEDED.
#
# THE FINDING IS A MISSING DECLARATION, NOT A WRONG GUESS. This check does not
# know which answer is right for a step and does not try. Either explicit value
# of `recurrence_expected` is clean; silence is the finding, because silence is
# how the estate paid for this twice. bugs_closed/024 established the rule for a
# tool re-render request the brake killed, and the remedy reached only the Go
# call sites that lane touched; two years later a customer's domain
# re-submission died the same way (bugs_open/326), because nothing counted
# adoption. At the commit that shipped this check: 19 of 21 keyed steps had
# never declared.
#
# Declaring never weakens dedup: recurrence_expected waives only the
# heuristics, and idx_swi_dedup still refuses a second OPEN item for the same
# (site_id, item_key). A step declared `true` still cannot run twice at once.
#
# Usage: scripts/audit-undeclared-recurrence.sh [--json]
# Exit:  0 = no findings · 1 = findings found · 2 = could not determine

set -uo pipefail

NAMESPACE="${NAMESPACE:-ai-persona-system}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
JSON_OUT=0
[[ "${1:-}" == "--json" ]] && JSON_OUT=1

# stderr is deliberately NOT swallowed — see audit-config-keys.sh's identical note.
psql_q() {
    kubectl -n "$NAMESPACE" exec -i postgres-clients-0 -- \
        psql -U clients_user -d clients_db -tAc "$1"
}

# Identical export to audit-loop-sitewide-item-keys.sh, and identical for the
# same reason: WalkSteps (in the Go binary) descends into sub-workflows itself,
# so this query must not try to. bugs_open/144 is what a second hand-written
# descent costs.
WORKFLOWS_JSON="$(psql_q "
SELECT jsonb_agg(jsonb_build_object('type', type, 'workflow', default_config->'workflow'))
FROM agent_definitions
WHERE deleted_at IS NULL
  AND COALESCE(is_snapshot,false) = false
  AND is_active
  AND default_config ? 'workflow';
")"
if [[ -z "$WORKFLOWS_JSON" || "$WORKFLOWS_JSON" == "null" ]]; then
    echo "ERROR: no live workflow definitions returned — query failed or the fleet is empty" >&2
    exit 2
fi

ERR=/tmp/audit-undeclared-recurrence.stderr
FINDINGS_JSON="$(printf '%s' "$WORKFLOWS_JSON" | (cd "$REPO_ROOT" && go run ./cmd/config-key-audit --undeclared-recurrence) 2>"$ERR")"
GO_EXIT=$?
# Read refusal from EMPTY stdout, never the exit code alone: `go run` folds the
# tool's exit 2 into its own 1 (the WFA-013 gotcha), so exit codes cannot
# separate "findings" from "refused to run".
if [[ -z "$FINDINGS_JSON" ]]; then
    echo "ERROR: config-key-audit --undeclared-recurrence produced no output:" >&2
    cat "$ERR" >&2
    rm -f "$ERR"
    exit 2
fi
cat "$ERR" >&2
rm -f "$ERR"

if [[ "$JSON_OUT" == "1" ]]; then
    printf '%s\n' "$FINDINGS_JSON"
else
    printf '%s' "$FINDINGS_JSON" | python3 -c '
import json, sys
out = json.load(sys.stdin)
fs = out["findings"]
print("=== KEYED create_work_item STEPS WITH NO recurrence_expected DECLARATION ===")
print("  ({} live agents scanned, {} undecoded)".format(out["agents_scanned"], out["agents_undecoded"]))
if not fs:
    print("  none")
else:
    for f in fs:
        extra = "  <-- DECLARED IN A SHAPE THE ACTION CANNOT READ" if f["declared_unhonoured"] else ""
        print("  {} {}{}".format(f["agent"], f["path"], extra))
        print("      item_type {!r}, prefix {!r}".format(f["item_type"], f["item_key_prefix"]))
    print()
    print("  For each: is the item an ACTION REQUEST (an agent asking for the next")
    print("  stage, a re-render, a retry) or a DETECTED DEFECT (a check that found")
    print("  something)? Set recurrence_expected true for the first, false for the")
    print("  second. Declaring true does NOT weaken dedup — idx_swi_dedup still")
    print("  refuses a second OPEN item for the key. Worked example: migration 572.")
    print("  Background: bugs_open/326, and bugs_closed/024 for the original rule.")
'
fi

exit $GO_EXIT
