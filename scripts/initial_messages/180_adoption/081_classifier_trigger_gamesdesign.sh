#!/usr/bin/env bash
# trigger-classifier-gamesdesign.sh
#
# Manually trigger domain-research-classifier for gamesdesign.co.uk —
# a direct smoke test of migration 008's prompt and the
# read_layout_taxonomy action. Bypasses the work-item dispatch loop
# entirely by posting straight to the generic entry point
# (system.agent.generic.requests with config.agent_type).
#
# Use this when:
#   - You need to prove the classifier's 008 workflow runs end-to-end.
#   - The dispatch loop has been marking needs_domain_research items
#     complete without actually running the classifier (as of 2026-04-23).
#
# Don't use this for routine classification — it runs in-chassis rather
# than in a dedicated pod (per doc 001, the generic entry point runs
# the workflow inside one of the long-lived agent-chassis pods). Fine
# for a smoke test; not appropriate for long-running production work.
# ────────────────────────────────────────────────────────────────────────

set -euo pipefail

SITE_ID='166cda1c-1f76-4013-a9b5-2da1dcdb1a18'
DOMAIN='gamesdesign.co.uk'

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID='demo_client'

echo "========================================="
echo "Manual classifier trigger — smoke test"
echo "========================================="
echo "  Site:             $DOMAIN"
echo "  Site ID:          $SITE_ID"
echo "  Correlation ID:   $CORRELATION_ID"
echo "  Orchestration ID: $ORCHESTRATION_ID"
echo "========================================="
echo ""
echo "SAVE THESE IDs:"
echo "  CORRELATION_ID=$CORRELATION_ID"
echo "  ORCHESTRATION_ID=$ORCHESTRATION_ID"
echo ""

kubectl -n kafka run -i --rm "kcat-classifier-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "orchestration_name=manual-classifier-$(date +%Y%m%d-%H%M%S)" \
  -H "step_name=start" \
  -H "client_id=$CLIENT_ID" \
  -H "message_type=request" \
  -H "action=orchestrate" \
  -H "from_agent_type=user" \
  -H "from_agent_id=cli" \
  -H "responses_topic=system.agent.generic.responses" <<JSON
{"action":"orchestrate","config":{"agent_type":"domain-research-classifier"},"input_data":{"site_id":"$SITE_ID","domain":"$DOMAIN","objective":"Smoke test 008: classify adopted gamesdesign.co.uk. Respect adopted site_archetype (interactive platform, dark practitioner aesthetic). Emit category and industry_tags so the layout matcher picks tool-portal-dark."}}
JSON

echo ""
echo "========================================="
echo "Classifier triggered. Running in a chassis pod (generic entry point)."
echo "========================================="
echo ""
echo "Tail chassis logs for this correlation:"
echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=500 | grep '$CORRELATION_ID'"
echo ""
echo "Watch key workflow steps:"
echo "  kubectl -n ai-persona-system logs -l app=agent-chassis --tail=500 | grep '$CORRELATION_ID' | grep -E 'search_domain|scrape_site|read_site_specs|read_layout_taxonomy|classify_and_extract|write_classification_spec'"
echo ""
echo "Watch for the taxonomy action output (confirms 008's new step ran):"
echo "  kubectl -n ai-persona-system logs -l app=agent-chassis --tail=500 | grep '$CORRELATION_ID' | grep -E 'read_layout_taxonomy|layout_count|loaded taxonomy'"
echo ""
echo "Check orchestration state:"
echo "  psql -c \"SELECT status, current_step, EXTRACT(EPOCH FROM (NOW() - last_activity))::int AS since_s, substring(COALESCE(error,''), 1, 300) AS err FROM orchestration_states WHERE correlation_id = '$CORRELATION_ID'::uuid ORDER BY created_at;\""
echo ""
echo "========================================="
echo "Once the classifier completes (workflow reaches 'complete' step), verify:"
echo "========================================="
echo ""
echo "1. LLM call was made with the 008 prompt + rendered taxonomy:"
cat <<'SQL'

SELECT created_at, step_name, success, length(prompt_rendered) AS prompt_len,
       CASE WHEN prompt_rendered LIKE '%Current library categories%' THEN 'YES' ELSE 'NO' END AS has_taxonomy_block,
       substring(prompt_rendered FROM 'Current library categories[^\[]*(\[[^\]]+\])') AS rendered_categories
FROM llm_call_log
WHERE correlation_id = '$CORRELATION_ID'
ORDER BY created_at;

SQL
echo ""
echo "2. New classification spec was written with category + industry_tags:"
cat <<'SQL'

SELECT is_current, source_agent, created_at,
       data->>'site_type'  AS site_type,
       data->>'category'   AS category,
       jsonb_array_length(COALESCE(data->'industry_tags','[]'::jsonb)) AS n_tags,
       data->'industry_tags' AS industry_tags,
       data->>'confidence' AS confidence
FROM site_specs
WHERE site_id = '166cda1c-1f76-4013-a9b5-2da1dcdb1a18'
  AND aspect = 'classification'
ORDER BY created_at DESC
LIMIT 3;

SQL
echo ""
echo "3. Classifier created the next work item (needs_strategy):"
cat <<'SQL'

SELECT item_type, status, handler_agent,
       EXTRACT(EPOCH FROM (NOW() - created_at))::int AS age_s
FROM site_work_items
WHERE site_id = '166cda1c-1f76-4013-a9b5-2da1dcdb1a18'
  AND item_type = 'needs_strategy'
  AND created_at > NOW() - INTERVAL '20 minutes'
ORDER BY created_at DESC;

SQL
echo ""
echo "Expected outcomes for a healthy run:"
echo "  (1) One LLM row, success=true, has_taxonomy_block=YES, rendered_categories is a JSON array containing brochure/interactive/social/etc."
echo "  (2) New classification row with category populated (likely 'interactive') and n_tags between 4 and 10."
echo "  (3) One needs_strategy work item in 'triaged' status."
echo ""
echo "If (1) returns 0 rows, the classifier workflow never reached classify_and_extract — check orchestration_states for where it stopped."
echo "If (2) shows category is NULL or n_tags=0, 008's prompt didn't elicit the new fields — inspect prompt_rendered and response_text in the llm_call_log row."
echo "If (3) is missing, create_next_item step didn't run — check orchestration_states.current_step at failure."