#!/usr/bin/env bash
# CHECK_naming_contract.sh — the three-way naming contract for acceptance-testable tools.
#
# WHY THIS IS THE FIRST THING THE LANE BUILT, ahead of its own substrate work.
# An acceptance run resolves a subject through THREE values that must agree:
#
#     doc_plans.subject_key  ==  content_components.function
#     pages.name             IN  (function, 'tool-' || function)
#
# `load_docs` keys the fence on `input_data.spec.function`; a mismatch returns an EMPTY
# fence and `request_browser_run` SKIPS with `needs_criteria`. That is honest — it does
# not fake a pass — but it is not a failure either, so **it reads as a clean run that
# asserted nothing**. The page lookup is `name IN ($2::text, 'tool-' || $2::text)`
# (tool_acceptance_actions.go:142), so a page named the function MINUS its prefix matches
# neither and the step hard-errors instead.
#
# Measured by the leopardess lane 2026-07-30 on hosted tools, and independently here on
# canonical tool components: a real and non-trivial population is affected. Until this
# passes, every other gate's green result is untrustworthy — which is why it outranks
# everything else this lane is doing.
#
# READ-ONLY. Prints what it measured, never just "no failures". Exit 1 if any tool that
# HAS a fence cannot be resolved (the actively-misleading class); exit 0 otherwise, with
# the authoring backlog reported as context rather than as failure.
set -uo pipefail

NS=ai-persona-system
PSQL=(kubectl -n "$NS" exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A -F'|')

q() { "${PSQL[@]}" -c "$1" 2>/dev/null; }

echo "=== The three-way naming contract ============================================="
echo "Population: content_components where component_level='tool', is_active,"
echo "forked_from IS NULL — the canonical tool set (idx_cc_tool_function_unique's scope)."
echo
echo "DENOMINATOR NOTE, because this figure has already been quoted two different ways:"
echo "this counts tool COMPONENTS. The leopardess lane counted HOSTED tools per site and"
echo "got a different number. Neither is wrong; they answer different questions. Say which"
echo "you mean when you quote either, and re-run rather than carrying a figure forward."
echo

TOTAL=$(q "SELECT count(DISTINCT function) FROM content_components
            WHERE component_level='tool' AND is_active AND forked_from IS NULL;")
echo "canonical tool components: ${TOTAL:-?}"
echo

echo "--- breakdown (fence present? / page resolvable?) ---"
q "WITH tools AS (
     SELECT DISTINCT function FROM content_components
      WHERE component_level='tool' AND is_active AND forked_from IS NULL
   ), r AS (
     SELECT t.function,
            EXISTS (SELECT 1 FROM doc_plans dp
                     WHERE dp.subject_type='tool' AND dp.subject_key=t.function
                       AND COALESCE(dp.is_current,true))                       AS has_fence,
            EXISTS (SELECT 1 FROM pages p
                     WHERE p.name = t.function OR p.name = 'tool-'||t.function) AS resolves
       FROM tools t
   )
   SELECT CASE WHEN has_fence AND resolves      THEN 'testable now'
               WHEN has_fence AND NOT resolves  THEN 'BROKEN: fence exists, page unresolvable'
               WHEN NOT has_fence AND resolves  THEN 'authoring backlog: page fine, no fence'
               ELSE 'neither: no fence and no resolvable page' END AS state,
          count(*)
     FROM r GROUP BY 1 ORDER BY 2 DESC;"

echo
echo "--- the BROKEN class, named, with its remedy ---"
echo "(a fence that can never run. This is the class that produces a misleading result"
echo " rather than an honest absence, so it is the only one that fails this check.)"
BROKEN=$(q "WITH tools AS (
              SELECT DISTINCT function FROM content_components
               WHERE component_level='tool' AND is_active AND forked_from IS NULL
            )
            SELECT t.function
              FROM tools t
             WHERE EXISTS (SELECT 1 FROM doc_plans dp
                            WHERE dp.subject_type='tool' AND dp.subject_key=t.function
                              AND COALESCE(dp.is_current,true))
               AND NOT EXISTS (SELECT 1 FROM pages p
                                WHERE p.name = t.function OR p.name = 'tool-'||t.function)
             ORDER BY 1;")

if [ -z "$BROKEN" ]; then
  echo "  none — every tool carrying a fence resolves to a page."
else
  # ⚠ READ INTO AN ARRAY FIRST — do NOT `while read` over a here-string here.
  # The first version of this script did, and it printed ONE of the two broken tools
  # while the RESULT line correctly said 2. Cause: `kubectl exec -i` inside the loop
  # body READS STDIN, so it swallowed the rest of the here-string and the loop ended
  # after one iteration. Silent, and it under-reports — the worst direction for a
  # detector, because a shorter list looks like better news.
  # It was caught ONLY because the summary count is computed separately from the list.
  # That is this lane's own rule ("print the count you measured") catching its own bug,
  # and it is the sixth instance in two days of the one class the ladder exists to defeat.
  mapfile -t BROKEN_ARR <<< "$BROKEN"
  for fn in "${BROKEN_ARR[@]}"; do
    [ -z "$fn" ] && continue
    STRIPPED=${fn#tool-}
    CAND=$(q "SELECT p.name || '  (' || s.domain || p.url || ')'
                FROM pages p JOIN sites s ON s.id = p.site_id
               WHERE p.name = '$STRIPPED' OR p.url LIKE '%/$STRIPPED.html'
               ORDER BY p.name LIMIT 3;")
    echo "  ✗ $fn"
    if [ -n "$CAND" ]; then
      echo "      page exists under a NON-MATCHING name:"
      printf '        %s\n' "$CAND"
      echo "      REMEDY: UPDATE pages SET name='$fn' WHERE name='$STRIPPED';"
      echo "      Safe — the deployed filename derives from pages.url, not pages.name."
    else
      echo "      no page found under '$STRIPPED' either — this is an ORPHANED tool"
      echo "      component with no page at all, a different defect. Do not rename;"
      echo "      decide whether the component should exist."
    fi
  done
  # Self-check: the list and the count must agree, or this script is under-reporting.
  LISTED=${#BROKEN_ARR[@]}
  COUNTED=$(printf '%s\n' "$BROKEN" | grep -c .)
  if [ "$LISTED" -ne "$COUNTED" ]; then
    echo "  ⚠ INTERNAL: listed $LISTED but counted $COUNTED — this script is dropping rows."
  fi
fi

echo
echo "--- the authoring backlog, for context (NOT a failure of this check) ---"
q "WITH tools AS (
     SELECT DISTINCT function FROM content_components
      WHERE component_level='tool' AND is_active AND forked_from IS NULL
   )
   SELECT count(*) || ' tool(s) resolve to a page but carry no fence — they cannot be'
       || ' acceptance-tested either, but honestly so: nothing claims they were.'
     FROM tools t
    WHERE NOT EXISTS (SELECT 1 FROM doc_plans dp
                       WHERE dp.subject_type='tool' AND dp.subject_key=t.function
                         AND COALESCE(dp.is_current,true))
      AND EXISTS (SELECT 1 FROM pages p
                   WHERE p.name = t.function OR p.name = 'tool-'||t.function);"

echo
if [ -n "$BROKEN" ]; then
  N=$(printf '%s\n' "$BROKEN" | grep -c .)
  echo "RESULT: FAIL — $N tool(s) carry a fence that can never resolve to a page."
  echo "Each one currently produces a skipped-but-not-failed run, which reads as clean."
  exit 1
fi
echo "RESULT: PASS — no tool carries a fence it cannot run."
exit 0
