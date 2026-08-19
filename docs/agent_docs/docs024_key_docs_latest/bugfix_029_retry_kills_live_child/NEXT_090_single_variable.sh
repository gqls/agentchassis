#!/usr/bin/env bash
# bugs_open/029 — the ONE-VARIABLE re-file. See HANDOFF_2026-08-19b, "the next re-file".
#
# Baseline is run d02a6958 (3 iterations, real Tier-1 citations, the lane's best result).
# Run 5d1d8f1c regressed to 1 iteration (scope-not-narrowing) after THREE simultaneous
# changes, so nothing could be attributed. This script changes exactly ONE thing:
# it appends the reconstruction query to the baseline symptom and keeps the seed scope
# BYTE-IDENTICAL to d02a6958's five symbols.
#
# DO NOT add the previous run's NextScope symbols. That is the suspected cause of the
# regression and is the variable under test.
set -euo pipefail
cd "$(dirname "$0")/../../../.." || exit 1
LANE="docs/agent_docs/docs024_key_docs_latest/bugfix_029_retry_kills_live_child"

BASE="$(cat "$LANE/SYMPTOM_d02a6958_baseline.txt")"
# NOTE the grep -v: the .sql file carries leading `--` comments, and flattening it to ONE
# line without stripping them comments out the ENTIRE query. It then returns zero rows with
# NO error - a silent failure that looks exactly like "the query found nothing". Verified
# both ways 2026-08-19: commented flatten -> 0 rows silently; stripped -> 20 rows.
QUERY_ONE_LINE="$(grep -v '^[[:space:]]*--' "$LANE/RECONSTRUCTION_QUERY.sql" | tr '\n' ' ' | tr -s ' ')"
ADDENDUM="RUN THIS QUERY FIRST, IN ITERATION 1 - a prior run exhausted its iterations rediscovering it, and awaited_requests results are capped at 200 rows so always filter and order by sent_at, never by orchestration_id: $QUERY_ONE_LINE"

# d02a6958's seed scope, unchanged.
export SEED_SCOPE="platform/orchestration/coordinator.go:continueExecution,platform/orchestration/coordinator.go:handleCompleteResponse,platform/orchestration/coordinator.go:persistAwaitingStateWithRetry,platform/orchestration/loop_error_handler.go:skipToNextLoopIteration,platform/agentbase/client.go:processResponse"
export FORCE=1

echo "seed scope (must match d02a6958 exactly):"; echo "  $SEED_SCOPE"; echo
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/090_TRIGGER_needs_diagnosis_v1.sh "$BASE $ADDENDUM"
