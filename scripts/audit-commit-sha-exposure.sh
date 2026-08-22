#!/bin/bash
# Read-only. Which live handler can produce a commit of its own but does not
# expose it at the standard path `response.commit_sha` via its complete step's
# `result_mapping`?
#
# The STANDING form of migration 537's apply-time guard (bugs_closed/334): 537
# wired build-dispatch-loop's `mark_complete` to `commit_sha?` — the handler's
# own reply or ABSENCE, never the whole-tree search. Absence is the contract,
# so a NEW commit-producing handler that does not expose the sha will simply
# never record `result.commit_sha`: no error, no log, no row. The guard proved
# the estate ready once, at apply time; this asks the same question on demand
# (the daily half is CronJob commit-sha-exposure-check, 06:45 UTC).
#
# The three sets are 537's, verbatim (see cmd/config-key-audit/
# commitshaexposure.go for why each window and predicate is what it is):
#   producers  — agents whose OWN orchestrations carry a commit_sha, 30 days
#   handlers   — site_work_items.handler_agent, 7 days
#   exposed    — complete_workflow steps with result_mapping.commit_sha,
#                walked by the Go binary (WalkSteps, nested included)
#
# Usage: scripts/audit-commit-sha-exposure.sh [--json]
# Exit:  0 = clean · 1 = findings · 2 = could not determine
#
# NOTE on exit codes (LANDMINES.md, `go run` collapses the child's exit status):
# the refusal is discriminated by EMPTY OUTPUT where JSON belongs, never by
# branching on exit code 2 — that branch would be dead code under `go run`.

set -uo pipefail

NAMESPACE="${NAMESPACE:-ai-persona-system}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
JSON_OUT=0
[[ "${1:-}" == "--json" ]] && JSON_OUT=1

psql_q() {
    kubectl -n "$NAMESPACE" exec -i postgres-clients-0 -- \
        psql -U clients_user -d clients_db -tAc "$1"
}

# One round trip building the whole stdin object server-side, so the three
# reads are from one consistent snapshot and the shell never string-assembles
# JSON (escape-sequence emission is a recorded trap on this estate).
INPUT_JSON="$(psql_q "
SELECT jsonb_build_object(
  'agents', (
     SELECT jsonb_agg(jsonb_build_object('type', type, 'workflow', default_config->'workflow'))
       FROM agent_definitions
      WHERE deleted_at IS NULL
        AND COALESCE(is_snapshot,false) = false
        AND is_active
        AND default_config ? 'workflow'),
  'producers', (
     SELECT COALESCE(jsonb_agg(DISTINCT owner_agent_type), '[]'::jsonb)
       FROM orchestration_states
      WHERE created_at > now() - interval '30 days'
        AND collected_data::text LIKE '%commit_sha%'),
  'handlers', (
     SELECT COALESCE(jsonb_agg(DISTINCT handler_agent), '[]'::jsonb)
       FROM site_work_items
      WHERE handler_agent IS NOT NULL AND handler_agent <> ''
        AND created_at > now() - interval '7 days')
);
")"
if [[ -z "$INPUT_JSON" || "$INPUT_JSON" == "null" ]]; then
    echo "ERROR: the combined export returned nothing — query failed or no DB access" >&2
    exit 2
fi

ERR="$(mktemp)"
ACKS_FILE="$REPO_ROOT/docs/agent_docs/docs024_key_docs_latest/architecture_review/commit_sha_exposure_acks.json"
FINDINGS_JSON="$(printf '%s' "$INPUT_JSON" | (cd "$REPO_ROOT" && go run ./cmd/config-key-audit --commit-sha-exposure --acks "$ACKS_FILE") 2>"$ERR")"
cat "$ERR" >&2
rm -f "$ERR"

if [[ -z "$FINDINGS_JSON" ]]; then
    echo "ERROR: config-key-audit --commit-sha-exposure printed nothing — refusal or broken build; see stderr above" >&2
    exit 2
fi

if [[ "$JSON_OUT" == "1" ]]; then
    printf '%s\n' "$FINDINGS_JSON"
fi

# The exit-code decision mirrors the binary's: any member of the intersection
# that is neither exposed nor acknowledged is a finding.
BAD=$(printf '%s' "$FINDINGS_JSON" | python3 -c '
import json,sys
fs = json.load(sys.stdin)
print(sum(1 for f in fs if not f["exposed"] and not f["acknowledged"]))
' 2>/dev/null)
if [[ -z "$BAD" ]]; then
    echo "ERROR: could not parse the findings JSON" >&2
    exit 2
fi
if [[ "$BAD" -gt 0 ]]; then
    exit 1
fi
exit 0
