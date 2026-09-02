#!/usr/bin/env bash
# Hand-run wrapper for `config-key-audit --render-truncation` (bugs_open/394).
#
# Builds the stdin object the mode demands: the RENDER_AUDIT_TRUNCATED rows PLUS
# the aliveness total, in ONE shape — the rows alone cannot prove the read was
# not blind, and zero rows of this code is a HEALTHY reading (a fleet whose sites
# all fit inside their caps writes none), so blindness has to be guarded one
# level down.
#
# The query is the same text as renderTruncationRunsQuery in
# cmd/config-key-audit/rendertruncation.go, deliberately: a wrapper that fetches
# a DIFFERENT shape from the one the CronJob reads is a second definition of the
# check, and it drifts. If you change one, change both.
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

RUNS_JSON=$(PSQL "
SELECT COALESCE(jsonb_agg(t), '[]'::jsonb) FROM (
  SELECT COALESCE(domain, '(no domain)')                              AS domain,
         COALESCE(agent_type, '(no agent_type)')                      AS agent_type,
         COALESCE(NULLIF(context->>'coverage_mode',''), '(absent)')   AS coverage_mode,
         COALESCE(context->>'window_first', '')                       AS window_first,
         COALESCE((context->>'priority_not_live')::int, 0)            AS priority_not_live,
         occurred_at::text                                            AS occurred_at
    FROM agent_error_log
   WHERE error_code = 'RENDER_AUDIT_TRUNCATED'
   ORDER BY domain, agent_type, occurred_at DESC) t;")
TOTAL=$(PSQL "SELECT count(*) FROM agent_error_log;")

printf '{"runs": %s, "error_log_rows": %s}' "$RUNS_JSON" "$TOTAL" \
  | go run ./cmd/config-key-audit --render-truncation "$@"
