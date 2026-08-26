#!/usr/bin/env bash
# Hand-run wrapper for `config-key-audit --ungraded-completions` (bugs_open/393).
# Builds the stdin object the mode demands: the grouped
# NO_CHANGE_GATE_UNREADABLE_RESULT rows PLUS the aliveness total, in one shape —
# the groups alone cannot prove the read was not blind.
#
# ⚠ `go run` collapses the child's exit status (documented landmine on the
# sibling wrappers): discriminate an exit-2 refusal by its EMPTY stdout, never
# by branching on $?.
set -euo pipefail
cd "$(dirname "$0")/.."

PSQL() {
  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
    psql -U clients_user -d clients_db -tA -c "$1"
}

GROUPS_JSON=$(PSQL "
SELECT COALESCE(jsonb_agg(t), '[]'::jsonb) FROM (
  SELECT COALESCE(NULLIF(context->>'item_type',''), '(no item_type in context)') AS item_type,
         count(*)                AS rows,
         min(occurred_at)::text  AS first_seen,
         max(occurred_at)::text  AS last_seen
    FROM agent_error_log
   WHERE error_code = 'NO_CHANGE_GATE_UNREADABLE_RESULT'
   GROUP BY 1
   ORDER BY 1) t;")
TOTAL=$(PSQL "SELECT count(*) FROM agent_error_log;")

printf '{"groups": %s, "error_log_rows": %s}' "$GROUPS_JSON" "$TOTAL" \
  | go run ./cmd/config-key-audit --ungraded-completions "$@"
