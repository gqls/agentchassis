#!/bin/bash
# Read-only. Which live conditional dots into a field whose producer cannot
# carry that path?
#
# bugs_open/313: `internal-linker.check_candidates` tested
# `candidate_pages.count > 0` while its producer declared
# `output_format: "array"` — a bare slice with no `count` key. The path
# resolves through no strategy, the numeric arm returns false silently, and
# every run for four months took else_step: 57 link jobs read `complete` with
# no link ever planned. The mismatch is fully visible in config — producer
# shape on one step, consumer path two lines below — so it is checkable
# offline, before a run pays for it. Same shape and same argument as
# audit-single-owner-actions.sh: a step cannot see at execution time what its
# producer DECLARED, so the check has to be offline and fleet-wide.
#
# Skips (deliberate — this asserts "can never resolve", not "looks odd"):
# numeric indexes into the array (WFA-012), fields with no known producer,
# non-query_database producers, and fields any object-format producer writes.
#
# Usage: scripts/audit-array-producer-conditions.sh [--json]
# Exit:  0 = no findings · 1 = findings found · 2 = could not determine
#
# NOTE on exit codes (LANDMINES.md, `go run` collapses the child's exit status):
# the refusal is discriminated by EMPTY OUTPUT where JSON belongs, never by
# branching on exit code 2 — that branch would be dead code under `go run`.

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

# Identical export to audit-single-owner-actions.sh, and identical for a reason:
# WalkSteps (in the Go binary) descends into loop sub-workflows itself, so this
# query must not try to. bugs_open/144 is what a second hand-written descent
# costs.
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

ERR=$(mktemp)
FINDINGS_JSON="$(printf '%s' "$WORKFLOWS_JSON" | (cd "$REPO_ROOT" && go run ./cmd/config-key-audit --array-producer-conditions) 2>"$ERR")"
cat "$ERR" >&2
rm -f "$ERR"

# Refusals (broken export, zero conditionals decoded) print NO JSON — that is
# the discriminator, not the collapsed exit code.
if [[ -z "$FINDINGS_JSON" ]]; then
    echo "ERROR: config-key-audit --array-producer-conditions produced no report — see stderr above" >&2
    exit 2
fi

if [[ "$JSON_OUT" == "1" ]]; then
    printf '%s\n' "$FINDINGS_JSON"
else
    COUNT="$(printf '%s' "$FINDINGS_JSON" | python3 -c 'import json,sys; print(len(json.load(sys.stdin)))')"
    echo "=== CONDITIONS THAT CAN NEVER RESOLVE AGAINST THEIR PRODUCER'S SHAPE ==="
    if [[ "$COUNT" == "0" ]]; then
        echo "  none"
    else
        printf '%s' "$FINDINGS_JSON" | python3 -c '
import json, sys
for f in json.load(sys.stdin):
    print("  {}: {} tests {!r}".format(f["agent"], f["step_path"], f["condition_path"]))
    print("      but {} ({}) declares output_format={!r} — an array carries no named keys".format(
        f["producer_path"], f["producer_action"], f["output_format"] or "(defaulted: array)"))
'
    fi
fi

COUNT="${COUNT:-$(printf '%s' "$FINDINGS_JSON" | python3 -c 'import json,sys; print(len(json.load(sys.stdin)))')}"
[[ "$COUNT" == "0" ]] && exit 0 || exit 1
