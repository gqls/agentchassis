#!/usr/bin/env bash
# VERIFY_075_post_roll.sh — proof for bugs_open/075 (dead-pod ownership discards
# responses; the F2 retry driver then loops for ever).
#
# Run AFTER a chassis roll that carries the 075 commit. The Go change is inert
# until then, so running this against an older image proves nothing — step 1
# refuses to continue in that case, which is the whole point of a control.
#
# WHAT IT CHECKS
#   1. pod-grep, BOTH directions. The new literals must be present AND the
#      REMOVED literal ("owned by different pod") must grep 0. The removed
#      string is the load-bearing half: a marker the change merely uses would
#      be satisfied by any image, which is how a vacuous pod-grep gets
#      published as evidence (see WRONG_CALLS 2026-07-26).
#   2. INDUCED FAULT A — the takeover branch. Stamp a live AWAITING_RESPONSES
#      orchestration with a pod name that does not exist, then require the next
#      response to be APPLIED. Pre-fix that response was discarded and the
#      offset committed (2026-07-25, orchestrations dc853e38 / 14bedb94), so
#      this is a controlled comparison, not a bare green.
#   3. INDUCED FAULT B — the cap. Manual, needs a harness agent whose adapter
#      step targets a topic nobody consumes; the recipe is printed, not run,
#      because it seeds config.
#   4. Orphan census: non-terminal orchestrations stamped to pods that no
#      longer exist. These should now heal themselves as responses arrive.
#
# GOTCHAS, each one paid for elsewhere in this repo:
#   - Give a restarted chassis pod ~300s before dispatching anything at it;
#     spawns inside that window are silently dropped.
#   - `strings /app/agent-chassis | grep -c` is the only trustworthy check that
#     a deploy carries your code. git and the image tag both lie (a same-tag
#     rebuild ships the node's cached binary).
#   - Step 2 mutates ONE orchestration row you choose. Do not point it at
#     another thread's production build unless you are content for it to be
#     driven by whichever pod answers — which, if the fix works, is harmless.
set -euo pipefail

NS=ai-persona-system
PSQL="kubectl -n $NS exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"

POD=$(kubectl -n "$NS" get pods -o name | grep -m1 'agent-chassis' | cut -d/ -f2)
[ -n "$POD" ] || { echo "ERROR: no agent-chassis pod found" >&2; exit 1; }

echo "=== 1. pod-grep: $POD ==="
kubectl -n "$NS" get deploy agent-chassis -o jsonpath='{.spec.template.spec.containers[0].image}'; echo

NEW_PRESENT=$(kubectl exec -n "$NS" "$POD" -- sh -c 'strings /app/agent-chassis | grep -c "ORCHESTRATION_TAKEN_OVER"' || echo 0)
OLD_GONE=$(kubectl exec -n "$NS" "$POD" -- sh -c 'strings /app/agent-chassis | grep -c "owned by different pod"' || echo 0)

kubectl exec -n "$NS" "$POD" -- sh -c '
  for s in ORCHESTRATION_TAKEN_OVER ORCHESTRATION_TAKEOVER_RACED ORCHESTRATION_TAKEOVER_FAILED \
           ADAPTER_RETRY_CAP_REACHED "failed to take over orchestration"; do
    printf "  %-40s %s   (new, expect >=1)\n" "$s" "$(strings /app/agent-chassis | grep -c "$s")"
  done
  printf "  %-40s %s   (REMOVED, expect 0)\n" "owned by different pod" \
    "$(strings /app/agent-chassis | grep -c "owned by different pod")"
  printf "  %-40s %s   (positive control, expect >=1)\n" RETRY_TICKER_CLAIMED \
    "$(strings /app/agent-chassis | grep -c RETRY_TICKER_CLAIMED)"
  printf "  %-40s %s   (negative control, expect 0)\n" zzz_not_a_real_symbol_075 \
    "$(strings /app/agent-chassis | grep -c zzz_not_a_real_symbol_075)"' || true

if [ "${NEW_PRESENT:-0}" -lt 1 ] || [ "${OLD_GONE:-1}" -ne 0 ]; then
  echo
  echo "STOP: this pod does NOT carry the 075 fix (new=$NEW_PRESENT old=$OLD_GONE)."
  echo "Everything below would test the OLD code. Roll first."
  exit 1
fi

echo
echo "=== 2. induced fault A — takeover ==="
echo "Pick a live AWAITING_RESPONSES orchestration (or dispatch one and wait for it):"
$PSQL -c "SELECT orchestration_id, orchestration_name, current_step, processing_node, last_activity
          FROM orchestration_states
          WHERE status='AWAITING_RESPONSES'
          ORDER BY last_activity DESC LIMIT 10;"

if [ -n "${ORCH:-}" ]; then
  echo "-- stamping $ORCH with a pod that does not exist --"
  $PSQL -c "UPDATE orchestration_states SET processing_node='induced-dead-pod-075'
            WHERE orchestration_id='$ORCH' AND status='AWAITING_RESPONSES';"
  echo "-- now WAIT for the awaited response to arrive, then: --"
  echo "   kubectl -n $NS logs $POD --since=10m | grep -E 'ORCHESTRATION_TAKEN_OVER|$ORCH' | head"
  echo "   PASS = ORCHESTRATION_TAKEN_OVER logged with previous_pod=induced-dead-pod-075,"
  echo "          the orchestration advances past its current step, and:"
  $PSQL -c "SELECT orchestration_id, status, current_step, processing_node
            FROM orchestration_states WHERE orchestration_id='$ORCH';"
  echo "   FAIL (the pre-fix behaviour) = nothing applied, the request expires,"
  echo "          RETRY_TICKER_CLAIMED re-executes the step ~3 min later, for ever."
else
  echo "(set ORCH=<orchestration_id> to run the induced fault against a chosen row)"
fi

echo
echo "=== 3. induced fault B — the adapter retry cap (recipe, not run) ==="
cat <<'RECIPE'
  Seed a throwaway agent whose workflow has ONE adapter step pointed at a topic
  nobody consumes (e.g. requests_topic system.adapter.nowhere.process) with a
  short timeout, dispatch it, then watch:

    SELECT orchestration_id, status, current_step,
           execution_metadata->'retry_count' AS retry_count
    FROM orchestration_states WHERE orchestration_name LIKE '<your-agent>%'
    ORDER BY created_at DESC LIMIT 5;

  PASS = retry_count for the step climbs 1 -> 2 -> 3 and then the orchestration
         goes FAILED (or routes to its error_step), with
         ADAPTER_RETRY_CAP_REACHED in the chassis log.
  FAIL = retry_count pinned at 1 while awaited_requests keeps gaining rows —
         that is the pre-fix assignment (`= retry_version + 1`) and the
         unbounded loop of 2026-07-25.
RECIPE

echo
echo "=== 4. orphan census: non-terminal rows stamped to pods that are gone ==="
LIVE=$(kubectl -n "$NS" get pods --no-headers -o custom-columns=:metadata.name | paste -sd"','" -)
$PSQL -c "SELECT status, processing_node, count(*), min(last_activity) AS oldest
          FROM orchestration_states
          WHERE status NOT IN ('COMPLETED','FAILED','CANCELLED')
            AND processing_node NOT IN ('$LIVE')
          GROUP BY 1,2 ORDER BY 3 DESC;"
echo "AWAITING_RESPONSES rows here should now heal on their next response."
echo "EXECUTING_STEP rows are F1's >4h reaper; INITIALIZED rows are covered by"
echo "nothing yet (bugs_open/075 deferred fix 4)."

echo
echo "=== 5. the kill test (bug 003's owed re-run) ==="
cat <<'KILL'
  Start a real orchestration with an adapter step, wait for AWAITING_RESPONSES,
  then: kubectl -n ai-persona-system delete pod <agent-chassis-pod>

  PASS = the F2 ticker claims the expired request (RETRY_TICKER_CLAIMED), the
         step is re-executed AT MOST ONCE, the response is applied by the new
         pod via ORCHESTRATION_TAKEN_OVER, and the orchestration COMPLETES with
         no manual containment and no repeated external side effects.
  On 2026-07-25 the same test produced a GitHub commit to vm-sites every ~3
  minutes until processing_node was cleared by hand — that is the comparison.
KILL
