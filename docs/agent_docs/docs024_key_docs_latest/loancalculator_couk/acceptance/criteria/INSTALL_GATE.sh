#!/usr/bin/env bash
# INSTALL_GATE.sh — refuse to install a computed_values fence into a fleet that
# cannot run it.
#
# WHY THIS EXISTS. The council gate returned REVISE on the computed_values
# submission (corr 1056cf11-7693-4fb6-a9fe-f67ee9f28bca, 2026-07-31) and FOUR
# independent seats raised the same hazard, debug_historian gating it at HIGH:
#
#   A criteria check type the running browser-runner binary does not know is
#   SKIPPED, not failed — and an all-skipped fence PASSES.
#
# So a computed_values fence installed into doc_plans before the browser-runner
# adapter has rolled would report a clean green run having asserted NOTHING.
# That is not a generic deploy-ordering nag: it is the exact false-green failure
# this whole check type was written to eliminate, reproduced by the fix for it.
# The plan's answer was "no fences are installed yet", which is a fact about
# today rather than a guard, and nothing stopped another lane installing one.
#
# This is the guard. Run it and read it before any fence naming computed_values
# is written into doc_plans.
#
# TWO CHECKS, because either alone can lie:
#
#   1. POD-GREP WITH A POSITIVE CONTROL. Grepping the running binary for a symbol
#      your change ADDED proves the image carries it. Grepping for it ALONE does
#      not: a mistyped symbol, a stripped binary or the wrong container all
#      produce "0 matches", which is indistinguishable from "not deployed". The
#      control is a string that must be present in ANY build of this binary, in
#      the SAME exec — if the control is also 0, the check itself is broken and
#      the run says so rather than reporting "not deployed".
#
#   2. A REAL RUN THAT REPORTS ZERO "not implemented" SKIPS. The binary carrying
#      the string is necessary, not sufficient — the type must also be reachable
#      through splitByProfile. Tier 4 reports an unknown type as
#      "SKIPPED: <type> not implemented", so a first in-cluster run whose skip
#      list is free of that phrase is the only positive evidence that the fence
#      is being executed rather than waved through. Do this once per fence, and
#      per TL-036 watch the 120s whole-request deadline: three vectors is
#      untested in-cluster.
#
# Usage:  ./INSTALL_GATE.sh
# Exit 0 = the deployed binary knows computed_values. Exit 1 = do not install.

set -uo pipefail

NS=ai-persona-system
DEPLOY=browser-runner-adapter
NEW_SYMBOL=computed_values
# Present in every build of this binary regardless of this change. If THIS comes
# back 0 the grep is not looking at what it thinks it is.
CONTROL_SYMBOL=no_horizontal_overflow

echo "== 1. pod-grep, with a positive control in the same exec =="

POD=$(kubectl get pods -n "$NS" -l app="$DEPLOY" \
        -o jsonpath='{.items[?(@.status.phase=="Running")].metadata.name}' 2>/dev/null \
      | tr ' ' '\n' | head -1)
if [ -z "${POD:-}" ]; then
  echo "FAIL: no running $DEPLOY pod found in $NS."
  echo "      Without a pod there is nothing to verify — do NOT install a fence."
  exit 1
fi
echo "   pod: $POD"

# Both greps in ONE exec so they cannot disagree about which binary they read.
#
# `grep -a`, NOT `strings | grep`. The recipe in CLAUDE.md is
#     strings /app/agent-chassis | grep -c "<symbol>"
# and it DOES NOT WORK IN THIS CONTAINER: browser-runner-adapter's image has no
# `strings` binary. The pipeline then fails silently and BOTH counts come back 0,
# which reads exactly like "the symbol is not deployed" — a false negative that
# would have had someone rebuild and re-roll a perfectly good image. The control
# is what turned that into a legible error instead, on this script's first run.
# `grep -a` treats the binary as text and needs nothing installed.
# `|| true` because grep exits 1 on zero matches, which is a legitimate answer
# here, not a failure of the exec.
COUNTS=$(kubectl exec -n "$NS" "$POD" -- sh -c \
  "grep -ac '$NEW_SYMBOL' /app/$DEPLOY || true; \
   grep -ac '$CONTROL_SYMBOL' /app/$DEPLOY || true" 2>/dev/null)
NEW_COUNT=$(echo "$COUNTS" | sed -n 1p)
CTL_COUNT=$(echo "$COUNTS" | sed -n 2p)
NEW_COUNT=${NEW_COUNT:-0}
CTL_COUNT=${CTL_COUNT:-0}

echo "   $NEW_SYMBOL: $NEW_COUNT    $CONTROL_SYMBOL (control): $CTL_COUNT"

if [ "$CTL_COUNT" -eq 0 ]; then
  echo
  echo "FAIL: the CONTROL symbol is absent too, so this grep proves nothing."
  echo "      The binary path or the container is wrong, or the grep tool is not"
  echo "      present — fix the CHECK, not the image."
  echo "      Do NOT read this as 'computed_values is not deployed'."
  exit 1
fi
if [ "$NEW_COUNT" -eq 0 ]; then
  echo
  echo "FAIL: the running binary does NOT carry computed_values (control present,"
  echo "      so the grep is sound)."
  echo "      Installing a fence now would make it SKIP — and an all-skipped fence"
  echo "      PASSES, reporting green having asserted nothing. Roll the"
  echo "      browser-runner-adapter first, then re-run this."
  exit 1
fi

echo
echo "PASS: the deployed browser-runner binary knows computed_values."
echo
echo "== 2. STILL OWED, and this script cannot do it for you =="
cat <<'OWED'
   Carrying the string is necessary, not sufficient. Run each fence ONCE in the
   cluster and confirm its skip list contains no "not implemented":

     SELECT collected_data->'request_run'->'response'->'summary'
     FROM orchestration_states WHERE id = '<run id>';

   Read it properly: a FAILED run reports status=COMPLETED with
   current_step='complete_error' and the real message in __step_error.
   Watch the 120s whole-request deadline — TL-036 hit it at 36 evaluations
   in-cluster while the same fence took 10.6s locally, and these fences carry
   three vectors each. If it times out, profile-gate the extra vectors to
   desktop rather than dropping them.
OWED
