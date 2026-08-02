#!/bin/bash
# Read-only. Which action that DECLARES single-ownership is carried by more than
# one live agent?
#
# RFC 006, owner ruling 2026-08-02; bugs_closed/150 is the instance. An action
# whose effect is "take everything on this site in state X" has two properties
# that are individually reasonable and jointly a trap: it is idempotent BY
# EMPTYING (a second run is a success returning zero, not an error), and its
# result describes what THAT CALL did while a branch reading it usually wants to
# know what is now TRUE. Those coincide only while exactly one caller exists.
# Add a second and the loser reports an honest zero that the branch reads as
# "nothing to do" — nothing errors, nothing logs, and a busy site is declared
# clean.
#
# Migration 286 removed the two duplicate `triage_detected_items` steps so the
# improvement loop owns promotion. That removal is a one-off; this check is the
# durable half, because the next agent to gain a triage step would otherwise
# re-create the bug with nothing reporting it. Same shape and same argument as
# audit-unregistered-actions.sh: the runtime cannot see a SIBLING agent's steps,
# so the check has to be offline and fleet-wide.
#
# Usage: scripts/audit-single-owner-actions.sh [--json]
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

# Identical export to audit-unregistered-actions.sh, and identical for a reason:
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

ERR=/tmp/audit-single-owner-actions.stderr
FINDINGS_JSON="$(printf '%s' "$WORKFLOWS_JSON" | (cd "$REPO_ROOT" && go run ./cmd/config-key-audit --single-owner-actions) 2>"$ERR")"
GO_EXIT=$?
if [[ $GO_EXIT -gt 1 ]]; then
    echo "ERROR: config-key-audit --single-owner-actions failed:" >&2
    cat "$ERR" >&2
    exit 2
fi
cat "$ERR" >&2
rm -f "$ERR"

if [[ "$JSON_OUT" == "1" ]]; then
    printf '%s\n' "$FINDINGS_JSON"
else
    COUNT="$(printf '%s' "$FINDINGS_JSON" | python3 -c 'import json,sys; print(len(json.load(sys.stdin)))')"
    echo "=== SINGLE-OWNER ACTIONS WITH MORE THAN ONE LIVE CARRIER ==="
    if [[ "$COUNT" == "0" ]]; then
        echo "  none"
    else
        printf '%s' "$FINDINGS_JSON" | python3 -c '
import json, sys
for f in json.load(sys.stdin):
    print("  action {!r} is carried by {} agents:".format(f["action"], len(f["owners"])))
    for p in f["paths"]:
        print("      {}".format(p))
'
        echo
        echo "  This action declares that its effect is wider than the run calling it,"
        echo "  so the copies race: whichever runs first takes everything and the rest"
        echo "  report an honest zero. Decide which agent OWNS the effect and remove the"
        echo "  step from the others — do not add a second signal per caller."
        echo "  Background: RFC 006 and bugs_closed/150."
    fi
fi

exit $GO_EXIT
