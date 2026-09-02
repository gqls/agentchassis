#!/usr/bin/env bash
# check_deploy.sh — "has a given code change actually reached the running chassis?"
#
# Answers the ONE question that keeps costing this lane time, and answers it with
# controls so a wrong answer cannot look like a right one.
#
#   ./check_deploy.sh <literal-the-change-REMOVED> [literal-the-change-ADDED]
#
# WHY A REMOVED LITERAL IS THE BEST PROBE. If your change DELETED a symbol, then
# "absent from the binary" is proof the change is in — and unlike an added literal
# it cannot be satisfied by an older build that happened to contain something similar.
# CLAUDE.md calls this "a genuine REMOVED-string control — the strongest kind".
#
# ⚠ DO NOT probe for a commit SHA. A binary carries the sha it was BUILT from, not
# every ancestor commit, so a miss proves nothing. That mistake was made in this lane
# on 2026-09-02 and the timestamp comparison had to settle it instead.
#
# ⚠ DO NOT rely on the `build provenance` log line. It is a STARTUP line and scrolls
# out of reach within hours on a busy service; a session's note in the chassis logs
# on 2026-09-02 reported it "matches NOTHING on a backend service".
#
# Worked example (2026-09-02, site-design-planner's layout-resolver fix bd8e45aba):
#   ./check_deploy.sh extractClassificationTags
#   -> removed literal 0, present control 1, absent control 0  ==> LIVE
set -uo pipefail
NS=ai-persona-system
REMOVED="${1:?usage: check_deploy.sh <removed-literal> [added-literal]}"
ADDED="${2:-}"
POD=$(kubectl -n "$NS" get pods -l app=agent-chassis --no-headers 2>/dev/null | awk 'NR==1{print $1}')
[ -n "$POD" ] || { echo "no agent-chassis pod found (expired kubeconfig? owner refreshes it)"; exit 2; }
IMG=$(kubectl -n "$NS" get deploy agent-chassis -o jsonpath='{.spec.template.spec.containers[0].image}' 2>/dev/null)
AGE=$(kubectl -n "$NS" get pod "$POD" --no-headers 2>/dev/null | awk '{print $5}')
echo "pod $POD   image ${IMG##*:}   age $AGE"
probe() { kubectl -n "$NS" exec "$POD" -- grep -ac "$1" /proc/1/exe 2>/dev/null | head -1; }
R=$(probe "$REMOVED"); P=$(probe "color-cta-bg-ink"); A=$(probe "zzz-definitely-not-present")
printf "  removed literal   %-32s %s\n" "$REMOVED" "$R"
[ -n "$ADDED" ] && printf "  added literal     %-32s %s\n" "$ADDED" "$(probe "$ADDED")"
printf "  CONTROL present   %-32s %s   (must be >0, else the probe is blind)\n" "color-cta-bg-ink" "$P"
printf "  CONTROL absent    %-32s %s   (must be 0, else it matches everything)\n" "zzz-definitely-not-present" "$A"
echo
if [ "${P:-0}" = "0" ] || [ "${A:-0}" != "0" ]; then
  echo "  VERDICT: PROBE IS UNRELIABLE — a control failed. Do not read the result above."
elif [ "${R:-1}" = "0" ]; then
  echo "  VERDICT: the change IS in this build."
else
  echo "  VERDICT: NOT rolled — the removed literal is still present."
fi
