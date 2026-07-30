#!/bin/bash
# Read-only. Which live steps name an action that is neither locally registered
# nor topic-routed?
#
# bugs_open/148: platform/validation/workflow.go (top-level steps) and
# subworkflow.go (nested steps) already reject a step in exactly that state with
# "requires a topic" — but only when a MESSAGE arrives at it. A definition can
# carry the defect for weeks with nothing reporting it, because the validator
# only speaks at execution time. This is the offline half, over every live step
# whether it has run recently or not — the same shape as audit-config-keys.sh,
# and for the same reason: two things that ask the SAME question the SAME way
# cannot go blind in different directions and agree with each other.
#
# Usage: scripts/audit-unregistered-actions.sh [--json]
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

# Same shape the runtime executes: type + full workflow, is_active / live / non-
# snapshot. WalkSteps (called by the Go binary below) descends into loop
# sub-workflows itself, so this query does not need to — see bugs_open/144 for
# why a hand-written descent here would be the exact mistake this tool avoids.
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

FINDINGS_JSON="$(printf '%s' "$WORKFLOWS_JSON" | (cd "$REPO_ROOT" && go run ./cmd/config-key-audit --unregistered-actions) 2>/tmp/audit-unregistered-actions.stderr)"
GO_EXIT=$?
if [[ $GO_EXIT -gt 1 ]]; then
    echo "ERROR: config-key-audit --unregistered-actions failed:" >&2
    cat /tmp/audit-unregistered-actions.stderr >&2
    exit 2
fi
cat /tmp/audit-unregistered-actions.stderr >&2
rm -f /tmp/audit-unregistered-actions.stderr

if [[ "$JSON_OUT" == "1" ]]; then
    printf '%s\n' "$FINDINGS_JSON"
else
    COUNT="$(printf '%s' "$FINDINGS_JSON" | python3 -c 'import json,sys; print(len(json.load(sys.stdin)))')"
    echo "=== UNREGISTERED-ACTION STEPS (rejected on every message; nothing else says so) ==="
    if [[ "$COUNT" == "0" ]]; then
        echo "  none"
    else
        printf '%s' "$FINDINGS_JSON" | python3 -c '
import json, sys
for f in json.load(sys.stdin):
    nested = " (nested)" if f["nested"] else ""
    print("  {}: {} -> action {!r}{}".format(f["agent"], f["path"], f["action"], nested))
'
        echo
        echo "  Each of these is rejected on every message it receives. Check whether any"
        echo "  live builder still dispatches at the agent named above before deciding"
        echo "  between repointing the action and retiring the dispatch edge — see"
        echo "  bugs_open/148 (or its closed successor) fix candidates 2 and 3."
    fi
fi

exit $GO_EXIT
