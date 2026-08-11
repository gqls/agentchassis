#!/bin/bash
# ============================================================================
# tool_acceptance_run.sh — fire tool-acceptance-agent at ONE tool: S6 of the
# staged build ladder (real clicks, real Chromium, desktop + mobile).
#
# REWRITTEN 2026-08-11 (owner decision, bugs_open/243 candidate 2 option b).
# This used to kcat the generic topic, which ran the agent INLINE on a standing
# chassis pod — a pod that deliberately has no storage client (owner ruling
# 2026-08-08), so the vision half of acceptance ('look') failed on every manual
# run, silently, while the check half read green. Now it raises the SAME
# site_work_items row the due-sweep raises (`check_tool_acceptance_due.go`),
# and build-dispatch-loop spawns a dedicated agent-tool-acceptance-agent pod —
# which since v1.0.1284 gets the storage config and secretKeyRef credentials,
# so BOTH halves run. Proven end-to-end 2026-08-11 (bugs_open/243 update:
# work item ae33ed59, run 0ee53904, first-ever vision rows in llm_call_log).
#
# The dispatch is asynchronous: the loop claimed the proof item within ~1
# minute, the run took ~3 more. This script prints the queries to follow it.
#
# THREE THINGS MUST STILL LINE UP OR THE RUN IS A NO-OP, two failing QUIETLY:
#
#  1. doc_plans.subject_key MUST equal <function>. If no PLAN is found the
#     fence is empty and request_browser_run SKIPS with reason=needs_criteria.
#     That is honest by design ("no fake pass") but it is not a failure either,
#     so a mistyped key looks like a clean run that asserted nothing.
#
#  2. THE NAMING GATE NO LONGER APPLIES HERE (2026-08-11). `url_field` is live
#     on the request_run step (migration 384) and is checked BEFORE the name
#     lookup, so this script resolves the page BY COMPONENT PLACEMENT and puts
#     its url in spec.page_url. A page whose name differs from the function —
#     the eight loancalculator.co.uk tools, including the one on `index` — is
#     now testable without renaming anything (owner decision 2026-08-11). It
#     prints a note when it uses that route, and refuses if pages.url is empty
#     (the one case neither route can resolve).
#
#  3. Every check TYPE in the fence must be one the RUNNING browser-runner
#     binary implements. An unknown type is SKIPPED, not failed, and an
#     all-skipped result set reads as PASS plus a 7-day cooldown. Grep the pod
#     with a LONG marker before trusting a green run (RUNBOOK §4).
#
# ALSO KNOW: a FAILING verdict on a fence without `no_auto_fix` raises ONE
# improve_tool item at the page. Ask the owning lane before firing at a page
# you do not own (rebuild_policy='owned' pages especially).
#
# Usage: ./tool_acceptance_run.sh <site_id> <domain> <function>
#   (domain is display-only; site_id + function drive everything)
# ============================================================================
set -euo pipefail

SITE="${1:?site_id}"
DOMAIN="${2:?domain (display only)}"
FUNCTION="${3:?function (must match doc_plans.subject_key AND pages.name)}"

PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A -v ON_ERROR_STOP=1"

# ── Preflight: find the page BY COMPONENT PLACEMENT, not by name ─────────────
# The name lookup is no longer the only route: `url_field` is live on the
# request_run step (migration 384) and is checked BEFORE the name lookup, so a
# spec carrying `page_url` resolves whatever the page is called. We therefore
# resolve by placement — which is what "the page this tool is on" actually
# means — and always put page_url in the spec. The exact-name page wins the tie
# so behaviour is unchanged for the tools that already resolved.
PAGE=$($PSQL <<SQL
SELECT p.id || '|' || p.name || '|' || cc.id || '|' || COALESCE(NULLIF(p.url,''),'')
  FROM pages p
  JOIN page_components pc ON pc.page_id = p.id
  JOIN content_components cc ON cc.id = pc.component_id
 WHERE p.site_id = '$SITE' AND p.status = 'active'
   AND cc.function = '$FUNCTION'
 ORDER BY (p.name IN ('$FUNCTION', 'tool-' || '$FUNCTION')) DESC,
          (p.deployed_at IS NOT NULL) DESC
 LIMIT 1;
SQL
)
if [ -z "$PAGE" ]; then
  echo "REFUSED: no active page on site $SITE carries a component with function"
  echo "'$FUNCTION'. Check the function spelling and that the placement is active."
  exit 1
fi
IFS='|' read -r PAGE_ID PAGE_NAME COMPONENT_ID PAGE_URL <<<"$PAGE"
if [ -z "$PAGE_URL" ]; then
  echo "REFUSED: page '$PAGE_NAME' has an empty pages.url, so neither route can"
  echo "resolve a target (the name route would also hard-error inside the run)."
  exit 1
fi
case "$PAGE_NAME" in
  "$FUNCTION"|"tool-$FUNCTION") ;;
  *) echo "note: page '$PAGE_NAME' does NOT match the Tier-4 name lookup" \
          "($FUNCTION / tool-$FUNCTION) — using the url_field route via spec.page_url." ;;
esac

PLAN_OK=$($PSQL -c "SELECT 1 FROM doc_plans WHERE subject_type='tool' AND subject_key='$FUNCTION' AND is_current;")
if [ -z "$PLAN_OK" ]; then
  echo "REFUSED: no current doc_plans row with subject_key='$FUNCTION' (header item 1)."
  echo "The run would SKIP everything with reason=needs_criteria and assert nothing."
  exit 1
fi

OPEN=$($PSQL -c "SELECT id FROM site_work_items WHERE site_id='$SITE' AND item_type='acceptance_run' AND spec->>'function'='$FUNCTION' AND status NOT IN ('complete','cancelled','failed') LIMIT 1;")
if [ -n "$OPEN" ]; then
  echo "REFUSED: an open acceptance_run for '$FUNCTION' already exists: $OPEN"
  echo "Follow that one instead — a duplicate insert would trip the dedup index anyway."
  exit 1
fi

# ── Insert the work item (the due-sweep's exact shape) ───────────────────────
ITEM_ID=$($PSQL <<SQL
INSERT INTO site_work_items
  (site_id, source, pipeline, item_type, severity, summary, spec,
   priority, handler_agent, status, created_by, item_key, triaged_at,
   approval_mode, max_attempts)
VALUES
  ('$SITE', 'discovery', 'build', 'acceptance_run', 'low',
   'Tier-4 acceptance run: $FUNCTION (manual, via tool_acceptance_run.sh)',
   jsonb_build_object('check','tool_acceptance_due','function','$FUNCTION',
                      'page_id','$PAGE_ID','page_name','$PAGE_NAME',
                      'component_id','$COMPONENT_ID','page_url','$PAGE_URL'),
   90, 'tool-acceptance-agent', 'triaged',
   'tool_acceptance_run.sh ($(whoami 2>/dev/null || echo cli))',
   'acceptance_run:$FUNCTION:$SITE', now(), 'auto', 3)
RETURNING id;
SQL
)
# psql -t -A still prints the INSERT command tag after the RETURNING row;
# keep only the id (first line). Caught live on this script's first run.
ITEM_ID=$(echo "$ITEM_ID" | head -1)

echo "function:   $FUNCTION  ($DOMAIN, page $PAGE_NAME)"
echo "work item:  $ITEM_ID"
echo ""
echo "build-dispatch-loop claims it within minutes and SPAWNS the agent pod."
echo "Follow it:"
echo "  SELECT status, claimed_by FROM site_work_items WHERE id='$ITEM_ID';"
echo "  SELECT correlation_id, processing_node, current_step,"
echo "         collected_data->'__step_error'->>'message'"
echo "    FROM orchestration_states"
echo "   WHERE collected_data->'input_data'->>'work_item_id'='$ITEM_ID';"
echo ""
echo "Read the verdict honestly — SKIPS ARE NOT PASSES (RUNBOOK §10), and the"
echo "vision result lives in collected_data->'look'."
