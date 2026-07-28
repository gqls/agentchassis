#!/usr/bin/env bash
# VERIFY_110_post_roll.sh — is bugs_open/110 candidate 2 actually live?
#
# Two questions, and they are NOT the same one:
#   1. is the code in the binary?      (pod-grep, below)
#   2. is it on the feature's path?    (a real Gemini row with non-NULL columns)
# A pod-grep answers only the first. 107 stayed open for a day on exactly that gap.
#
# THE MARKER TRAP THIS SCRIPT EXISTS TO AVOID. Three of the four new column names are
# SUBSTRINGS of option keys that bugs_open/107 put in the binary weeks ago:
#   wire_max_output_tokens   <- matches __sent_wire_max_output_tokens    (pre-existing)
#   thinking_reserve_tokens  <- matches __sent_thinking_reserve_tokens   (pre-existing)
#   thinking_tokens          <- matches __usage_thinking_tokens          (pre-existing)
# Grepping any of those returns >0 on a binary WITHOUT this change. Only
# `total_output_tokens` discriminates, because __usage_total_tokens does not contain it.
#
# Usage: ./VERIFY_110_post_roll.sh
set -uo pipefail

NS=ai-persona-system
PSQL="kubectl -n $NS exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc"

POD=$(kubectl -n $NS get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
IMAGE=$(kubectl -n $NS get pods -l app=agent-chassis -o jsonpath='{.items[0].spec.containers[0].image}')
START=$(kubectl -n $NS get pods -l app=agent-chassis -o jsonpath='{.items[0].status.startTime}')
echo "pod=$POD"
echo "image=$IMAGE"
echo "started=$START"
echo

g() { kubectl -n $NS exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c '$1'" 2>/dev/null; }

echo "--- 1. binary: POSITIVE CONTROLS (same INSERT, pre-existing — must be >=1) ---"
for s in rag_context_used prompt_variant "work_item_id, vertical"; do
  printf "  %-26s %s\n" "$s" "$(g "$s")"
done

echo "--- 2. binary: THE ONLY DISCRIMINATING MARKER (must be >=1 if this shipped) ---"
DISC=$(g total_output_tokens)
printf "  %-26s %s\n" "total_output_tokens" "$DISC"

echo "--- 3. binary: NEGATIVE CONTROL (invented — must be 0) ---"
printf "  %-26s %s\n" "zzz_not_a_real_symbol" "$(g zzz_not_a_real_symbol)"

echo "--- 4. binary: VACUOUS markers, shown ONLY to prove they are useless ---"
for s in wire_max_output_tokens thinking_reserve_tokens; do
  printf "  %-26s %s  (matches a pre-existing __sent_* key — ignore)\n" "$s" "$(g "$s")"
done
echo

if [ "${DISC:-0}" -lt 1 ]; then
  echo "RESULT: the change is NOT in this binary. 110 stays OPEN."
  echo "        Needs an image built at or after commit 4ca21d7d6."
  exit 1
fi
echo "RESULT: the code IS in the binary. That is necessary and NOT sufficient — continue."
echo

echo "--- 5. path: Gemini rows since this pod started ---"
$PSQL "
SELECT created_at||' | '||agent_type||' | max='||COALESCE(max_tokens::text,'-')
       ||' wire='||COALESCE(wire_max_output_tokens::text,'NULL')
       ||' reserve='||COALESCE(thinking_reserve_tokens::text,'NULL')
       ||' thinking='||COALESCE(thinking_tokens::text,'NULL')
       ||' total='||COALESCE(total_output_tokens::text,'NULL')
FROM llm_call_log
WHERE provider='gemini' AND created_at > '$START'
ORDER BY created_at DESC LIMIT 10;"

N=$($PSQL "SELECT count(*) FROM llm_call_log WHERE provider='gemini' AND created_at > '$START' AND thinking_tokens IS NOT NULL;")
echo
if [ "${N:-0}" -lt 1 ]; then
  echo "NO populated Gemini row yet. This is EXPECTED until a Gemini call runs on the new"
  echo "pod — it is not evidence of failure. Wait for an organic page build (they happen"
  echo "on their own; one ran unprompted at 02:29 on 07-28), or trigger one."
  echo "NOTE: no orchestration dispatch within ~300s of a pod restart — silently dropped."
  exit 2
fi

echo "RESULT: $N Gemini row(s) with non-NULL thinking_tokens — 110 candidate 2 is LIVE."
echo "Sanity: thinking_tokens should land near 2,764-2,878 for the writer's real prompt."
echo "        A wildly different number is worth reading, not assuming away."
