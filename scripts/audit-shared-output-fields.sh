#!/usr/bin/env bash
# audit-shared-output-fields.sh — RFC_012 (d): find step pairs in one live
# workflow sharing an output_field, transitively reachable over the FULL
# routing graph (13 config keys + next_step/error_step), with DIFFERENT
# actions. Hand-run form; the standing form is the shared-output-fields
# CronJob (same binary, same image as component-render-check's pattern).
#
# Both naive detectors return 0 on bugs_open/192 — the check's own tests pin
# that case (cmd/config-key-audit/sharedoutputs_test.go), so a green run of
# the SUITE is the proof the detector can fire; a clean run HERE then means
# the fleet, not the detector.
#
# Exit: 0 clean, 1 findings, 2 unusable input.
set -uo pipefail

NAMESPACE="${NAMESPACE:-ai-persona-system}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

psql_q() {
    kubectl -n "$NAMESPACE" exec -i postgres-clients-0 -- \
        psql -U clients_user -d clients_db -tAc "$1"
}

# Identical export to the sibling audits, identical for a reason: the Go side
# descends into sub-workflows itself; a second hand-written descent is
# bugs_open/144's cost.
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

printf '%s' "$WORKFLOWS_JSON" | (cd "$REPO_ROOT" && go run ./cmd/config-key-audit --shared-output-fields)
