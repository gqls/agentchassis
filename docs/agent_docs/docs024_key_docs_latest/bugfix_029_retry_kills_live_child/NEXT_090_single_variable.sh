#!/usr/bin/env bash
# bugs_open/029 — the ONE-VARIABLE re-file. See HANDOFF_2026-08-19b, "the next re-file".
#
# Baseline is run d02a6958 (3 iterations, real Tier-1 citations, the lane's best result).
#
# ⚠ CORRECTED 2026-08-20 — THE RULE THIS SCRIPT USED TO ASSERT WAS WRONG, IN THE WRONG
# DIRECTION. It said "DO NOT add the previous run's NextScope symbols; that is the suspected
# cause of 5d1d8f1c's regression". The scope guard's arithmetic says the opposite. Both runs
# reconcile exactly against it (guard: pkg/diagnose/loop.go:432, `next.size() > prevSize+2`;
# size() = len(Symbols) :205; init PrevScopeSize = seed.size()+1, advance.go:68; on a stop
# advance.go returns at :104-111 BEFORE :120 overwrites it, so the persisted value is pre-guard):
#
#   d02a6958  seed 5 -> prev 6 | iter1 named  8: 8 > 8 false, PASSES -> prev 8
#                              | iter2 named  5: 5 > 10 false, passes -> prev 5
#                              | iter3 named 12: 12 > 7 TRIPS.   persisted prev_scope_size = 5 ✓
#   5d1d8f1c  seed 6 -> prev 7 | iter1 named 13: 13 > 9 TRIPS.   persisted prev_scope_size = 7 ✓
#
# Threshold = prevSize + 2, so a WIDER seed RAISES the allowance and is PROTECTIVE. Seed width
# cannot have caused 5d1d8f1c's trip. What differed is the model naming 13 symbols in iteration 1
# where the baseline named 8 - attributable to the symptom change or to plain variance, and NOT
# established either way. Note d02a6958 survived iteration 1 by exactly one (8 > 8 is false).
#
# So this script changes ONE thing - it appends the reconstruction query - and keeps the seed
# scope identical to the baseline for COMPARABILITY, not because widening is harmful.
# If this run also trips at iteration 1, the symptom addendum is implicated, not the seed.
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

# d02a6958's seed scope, unchanged - for comparability against a known baseline.
export SEED_SCOPE="platform/orchestration/coordinator.go:continueExecution,platform/orchestration/coordinator.go:handleCompleteResponse,platform/orchestration/coordinator.go:persistAwaitingStateWithRetry,platform/orchestration/loop_error_handler.go:skipToNextLoopIteration,platform/agentbase/client.go:processResponse"
export FORCE=1

echo "seed scope (must match d02a6958 exactly):"; echo "  $SEED_SCOPE"; echo
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/090_TRIGGER_needs_diagnosis_v1.sh "$BASE $ADDENDUM"
