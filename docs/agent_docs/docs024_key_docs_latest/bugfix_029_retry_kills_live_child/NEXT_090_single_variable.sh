#!/usr/bin/env bash
# Bug 343, slug parent_freezes_silently_after_an_abandoned_await — the ONE-VARIABLE re-file.
# Lane: docs024_key_docs_latest/bugfix_029_retry_kills_live_child/ (see HANDOFF_2026-08-21).
#
# ⚠ NOTE THE POINTER SHAPE, AND DO NOT "HELPFULLY" RESTORE THE DIRECTORY PREFIX. This header named
# `bugs_open/029` when written, was retargeted to `bugs_open/343` on 08-20, and would have needed a
# THIRD edit on 08-22 when 343 closed. Three rewrites of one pointer in three days is the SHAPE
# failing, not three unlucky addresses — so it now names NUMBER + SLUG + LANE DIR and no directory,
# which is what LANDMINES.md's "A pointer that hardcodes `bugs_open/` rots at exactly the moment it
# starts to matter" prescribes. Resolve a bug by slug across BOTH bugs_ dirs; a bare number is
# ambiguous anyway (029 names two unrelated cases). Measured fleet-wide 2026-08-22 by the 040 lane:
# ~71% of resolvable `bugs_open/NNN` pointers in this repo are ALREADY DEAD, so a pointer that does
# not resolve is the base rate — not evidence the reference was wrong.
#
# ⚠ DO NOT FIRE THIS. Two independent reasons, and the second is new:
#   1. The standing wait-for-the-burst instruction (unchanged): the 08-17 cohort is explained as an
#      external GitHub outage, the evidence is preserved so nothing expires, and the capture cron
#      (RSH-011) takes the next occurrence automatically.
#   2. 343 WAS CLOSED 2026-08-22 by owner ruling — an explicit override of the fixed-AND-live bar,
#      with the first death still unexplained. A 090 costs real credits and a diagnosis run against
#      a closed bug is almost certainly not what anyone wants. If the freeze recurs OUTSIDE an
#      outage window, the honest first move is to re-open under a NEW number and ask the owner,
#      not to fire this script at a closed case.
# The script is kept because the symptom text and seed scope are still correct and were expensive
# to get right — it is a ready instrument, not a recommendation.
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
