#!/bin/bash
# Read-only. Which SHARED action's optional-key set has accumulated past the
# budget?
#
# RFC 022, owner ruling 2026-08-11 (option 3 with option 1 interim). The interim
# exempted every individually-inert opt-in field (unsafe default OFF, no live
# consumer naming it) from architecture review — which deliberately gave up the
# ACCUMULATION signal: ten such fields are a shared action nobody understands,
# and the per-change trigger was the only thing that would have noticed the
# tenth. This report is what notices the tenth. Same shape and argument as
# audit-single-owner-actions.sh (RFC 006): the runtime cannot see a sibling
# agent's steps, so the count has to be offline and fleet-wide.
#
# Usage: scripts/audit-optional-key-budget.sh [--json] [--census] [BUDGET]
#   The budget is RULED (owner, 2026-08-14): N = 10 — the default when no BUDGET
#   is given, so a bare run enforces the ruling. A number overrides it for
#   what-if sizing; --census runs the no-budget census (no findings, exit 0).
#   Sharing itself is estate design, not the defect (owner, same ruling): a
#   finding means an action's ACCUMULATED optional surface owes one review as
#   a whole, never that its reuse is a problem.
# Exit:  0 = no findings (or census-only) · 1 = findings · 2 = could not determine
#
# NOTE on exit codes (LANDMINES.md, `go run` collapses the child's exit status):
# the refusal is discriminated by EMPTY OUTPUT where JSON belongs, never by
# branching on exit code 2 — that branch would be dead code under `go run`.

set -uo pipefail

NAMESPACE="${NAMESPACE:-ai-persona-system}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
JSON_OUT=0
CENSUS=0
BUDGET=""
for arg in "$@"; do
    if [[ "$arg" == "--json" ]]; then
        JSON_OUT=1
    elif [[ "$arg" == "--census" ]]; then
        CENSUS=1
    else
        BUDGET="$arg"
    fi
done
# The ruled default (owner, 2026-08-14). An explicit number still overrides for
# what-if sizing; --census suppresses the budget entirely.
if [[ "$CENSUS" == "0" && -z "$BUDGET" ]]; then
    BUDGET=10
fi

# stderr is deliberately NOT swallowed — see audit-config-keys.sh's identical note.
psql_q() {
    kubectl -n "$NAMESPACE" exec -i postgres-clients-0 -- \
        psql -U clients_user -d clients_db -tAc "$1"
}

# Identical export to audit-single-owner-actions.sh, and identical for a reason:
# WalkSteps (in the Go binary) descends into loop sub-workflows itself, so this
# query must not try to. bugs_open/144 is what a second hand-written descent costs.
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

ERR="$(mktemp)"
CENSUS_JSON="$(printf '%s' "$WORKFLOWS_JSON" | (cd "$REPO_ROOT" && go run ./cmd/config-key-audit --optional-key-budget ${BUDGET:+"$BUDGET"}) 2>"$ERR")"
GO_EXIT=$?
cat "$ERR" >&2
rm -f "$ERR"

# The refusal path: empty stdout where JSON belongs. Exit-code branching cannot
# see it — `go run` folds the tool's exit 2 into 1 (LANDMINES.md).
if [[ -z "$CENSUS_JSON" ]]; then
    echo "ERROR: config-key-audit --optional-key-budget printed nothing — refusal or broken build; see stderr above" >&2
    exit 2
fi

if [[ "$JSON_OUT" == "1" ]]; then
    printf '%s\n' "$CENSUS_JSON"
else
    printf '%s' "$CENSUS_JSON" | python3 -c '
import json, sys
out = json.load(sys.stdin)
budget, rows = out["budget"], out["actions"]
findings = [r for r in rows if r["over_budget"]]
shared   = [r for r in rows if r["consumers"] >= 2]
print("=== OPTIONAL-KEY SURFACE PER ACTION (RFC 022) ===")
print("  {} actions declare optional keys; {} of them are SHARED (>=2 live carriers)".format(len(rows), len(shared)))
if budget is None:
    print("  census only — no budget given, so no findings by construction")
else:
    print("  budget: {} — {} shared action(s) over it".format(budget, len(findings)))
    for f in findings:
        print("  OVER BUDGET: {!r} declares {} optional keys, carried by {} agents: {}".format(
            f["action"], f["optional_keys"], f["consumers"], ", ".join(f["agents"])))
print()
print("  widest shared surfaces (accumulation is the signal — the tenth field, not the first):")
for r in shared[:10]:
    print("    {:3d} optional keys  {:2d} carriers  {}".format(r["optional_keys"], r["consumers"], r["action"]))
'
fi

if [[ -n "$BUDGET" ]]; then
    OVER="$(printf '%s' "$CENSUS_JSON" | python3 -c 'import json,sys; print(sum(1 for r in json.load(sys.stdin)["actions"] if r["over_budget"]))')"
    [[ "$OVER" != "0" ]] && exit 1
fi
exit 0
