#!/bin/bash
# ==========================================================================
# Trigger the vet-pipeline-orchestrator
# ==========================================================================
#
# Usage:
#   bash trigger_vet_pipeline.sh                     # defaults: 50 areas, 500 promote, 100 verify
#   bash trigger_vet_pipeline.sh --area-code BT      # Belfast only
#   bash trigger_vet_pipeline.sh --limit 5           # only 5 areas to sweep
#   bash trigger_vet_pipeline.sh --verify-limit 20   # only verify 20 businesses
#   bash trigger_vet_pipeline.sh --dry-run            # show message without sending
#
# What happens:
#   1. Dispatches area-sweep-discoverer for unswept postcode districts (fire-and-forget)
#   2. Promotes pending discovery_candidates into businesses (from previous sweeps)
#   3. Dispatches vet-practice-verifier for pending businesses (fire-and-forget)
#
# This is a rolling pipeline — each run advances work from previous runs.
# ==========================================================================

set -euo pipefail

# Defaults
SWEEP_LIMIT=50
PROMOTE_LIMIT=500
VERIFY_LIMIT=100
AREA_CODE=""
DELAY_MS=200
DRY_RUN=false
CLIENT_ID="vetcomparison"

# Parse arguments
while [[ $# -gt 0 ]]; do
case $1 in
--area-code)    AREA_CODE="$2"; shift 2 ;;
--limit)        SWEEP_LIMIT="$2"; shift 2 ;;
--promote-limit) PROMOTE_LIMIT="$2"; shift 2 ;;
--verify-limit) VERIFY_LIMIT="$2"; shift 2 ;;
--delay-ms)     DELAY_MS="$2"; shift 2 ;;
--dry-run)      DRY_RUN=true; shift ;;
--client-id)    CLIENT_ID="$2"; shift 2 ;;
*) echo "Unknown option: $1"; exit 1 ;;
esac
done
----------------------
------- v 1  ------------

SWEEP_LIMIT=0
PROMOTE_LIMIT=500
VERIFY_LIMIT=200
AREA_CODE=""
DELAY_MS=5000
COUNTRY="GB"
BUSINESS_TYPE="veterinary practice"
VERTICAL_SLUG="veterinary"
DRY_RUN=false
CLIENT_ID="vetcomparison"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
ORCHESTRATION_NAME="vet-pipeline-$(date +%Y%m%d-%H%M%S)"

KAFKA_BOOTSTRAP="personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"
TOPIC="system.agent.generic.requests"

# Build input_data JSON
INPUT_DATA="{\"limit\":${SWEEP_LIMIT},\"promote_limit\":${PROMOTE_LIMIT},\"verify_limit\":${VERIFY_LIMIT},\"delay_ms\":${DELAY_MS},\"country\":\"${COUNTRY}\",\"business_type\":\"${BUSINESS_TYPE}\",\"vertical_slug\":\"${VERTICAL_SLUG}\""
if [ -n "$AREA_CODE" ]; then
  INPUT_DATA="${INPUT_DATA},\"area_code\":\"${AREA_CODE}\""
fi
INPUT_DATA="${INPUT_DATA}}"


echo "========================================="
echo "Vet Pipeline Orchestrator"
echo "========================================="
echo "  Sweep limit:     ${SWEEP_LIMIT}"
echo "  Promote limit:   ${PROMOTE_LIMIT}"
echo "  Verify limit:    ${VERIFY_LIMIT}"
echo "  Area code:       ${AREA_CODE:-all}"
echo "  Delay (ms):      ${DELAY_MS}"
echo "  Client:          ${CLIENT_ID}"
echo "  Orchestration:   ${ORCHESTRATION_NAME}"
echo "  Country:         ${COUNTRY}"
echo "  Business type:   ${BUSINESS_TYPE}"
echo "  Vertical:        ${VERTICAL_SLUG}"
echo "========================================="
echo ""
echo "SAVE THESE IDs:"
echo "  CORRELATION_ID=$CORRELATION_ID"
echo "  ORCHESTRATION_ID=$ORCHESTRATION_ID"
echo ""

if [ "$DRY_RUN" = true ]; then
echo "DRY RUN — message body:"
echo "$BODY" | python3 -m json.tool 2>/dev/null || echo "$BODY"
echo ""
echo "Headers:"
echo "  correlation_id=$CORRELATION_ID"
echo "  orchestration_id=$ORCHESTRATION_ID"
echo "  orchestration_name=$ORCHESTRATION_NAME"
echo "  client_id=$CLIENT_ID"
echo "  action=orchestrate"
exit 0
fi

kubectl -n kafka run -i --rm kcat-pipeline-$$ \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
    -b "$KAFKA_BOOTSTRAP" \
    -t "$TOPIC" \
    -H "correlation_id=$CORRELATION_ID" \
    -H "request_id=$REQUEST_ID" \
    -H "message_id=$MESSAGE_ID" \
    -H "orchestration_id=$ORCHESTRATION_ID" \
    -H "orchestration_name=$ORCHESTRATION_NAME" \
    -H "step_name=start" \
    -H "client_id=$CLIENT_ID" \
    -H "message_type=request" \
    -H "action=orchestrate" \
    -H "from_agent_type=user" \
    -H "from_agent_id=cli" \
    -H "responses_topic=system.agent.generic.responses" <<JSON
{"action":"orchestrate","config":{"agent_type":"vet-pipeline-orchestrator"},"input_data":${INPUT_DATA}}
JSON

echo ""
echo "========================================="
echo "Pipeline started!"
echo "========================================="
echo ""
echo "MONITORING:"
echo ""
echo "1. Watch logs:"
echo "   kubectl logs -n ai-persona-system -l app=agent-chassis -f | grep '$CORRELATION_ID'"
echo ""
echo "2. Check orchestration status:"
echo "   kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c \\"
echo "     \"SELECT status, current_step FROM orchestration_states WHERE orchestration_id = '$ORCHESTRATION_ID'\""
echo ""
echo "3. Check pipeline output (after completion):"
echo "   kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c \\"
echo "     \"SELECT final_result FROM orchestration_states WHERE orchestration_id = '$ORCHESTRATION_ID'\""
echo ""
echo "4. Monitor discovery candidates:"
echo "   kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c \\"
echo "     \"SELECT status, COUNT(*) FROM business_intel.discovery_candidates GROUP BY status\""
echo ""
echo "5. Monitor business verification:"
echo "   kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c \\"
echo "     \"SELECT verification_status, COUNT(*) FROM business_intel.businesses b JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id WHERE bv.slug = 'veterinary' GROUP BY verification_status\""
echo ""

echo "  -- Did ensure_tasks actually create them?"
echo "  SELECT status, COUNT(*) FROM business_intel.collection_tasks GROUP BY status;"
echo "  "
echo "  -- What did the batch processor's child orchestration do?"
echo "  SELECT orchestration_id, status, current_step, error"
echo "  FROM orchestration_states"
echo "  WHERE parent_orchestration_id = '$ORCHESTRATION_ID'"
echo "  ORDER BY created_at;"
echo "  "
echo "  -- And check the batch processor's own collected_data"
echo "  SELECT collected_data->'batch' as batch_data"
echo "  FROM orchestration_states"
echo "  WHERE owner_agent_type = 'vet-batch-processor'"
echo "  ORDER BY created_at DESC"
echo "  LIMIT 1;"
echo ""

echo ""
echo " -- Quick status check you can re-run"
echo " SELECT "
echo "     (SELECT COUNT(*) FROM business_intel.collection_tasks WHERE status = 'completed') as tasks_done,"
echo "     (SELECT COUNT(*) FROM business_intel.collection_tasks WHERE status = 'in_progress') as tasks_active,"
echo "     (SELECT COUNT(*) FROM business_intel.collection_tasks WHERE status = 'pending') as tasks_pending,"
echo "     (SELECT COUNT(*) FROM business_intel.businesses b JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id WHERE bv.slug = 'veterinary' AND b.verification_status = 'verified') as verified;"
echo "    (SELECT COUNT(*) FROM business_intel.businesses b JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id WHERE bv.slug = 'veterinary' AND b.verification_status = 'verified') as verified,"
echo "               (SELECT COUNT(*) FROM business_intel.business_prices WHERE is_current = TRUE) as current_prices;"
echo ""

echo " "
echo " SELECT orchestration_id, status, current_step, "
echo "        updated_at, responses_topic, requests_topic"
echo " FROM orchestration_states"
echo " WHERE orchestration_id = '$ORCHESTRATION_ID';"
echo " "
echo " -- Also check awaited_requests to see what it's waiting for"
echo " SELECT awaited_requests"
echo " FROM orchestration_states"
echo " WHERE orchestration_id = '$ORCHESTRATION_ID';"
echo ""


--
SELECT
    (SELECT COUNT(*) FROM business_intel.collection_tasks WHERE status = 'completed') as tasks_done,
    (SELECT COUNT(*) FROM business_intel.collection_tasks WHERE status = 'in_progress') as tasks_active,
    (SELECT COUNT(*) FROM business_intel.collection_tasks WHERE status = 'pending') as tasks_pending,
    (SELECT COUNT(*) FROM business_intel.businesses b JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id WHERE bv.slug = 'veterinary' AND b.verification_status = 'verified') as verified,
    (SELECT COUNT(*) FROM business_intel.business_prices WHERE is_current = TRUE) as current_prices;


-- View recently verified businesses with their details
SELECT
    b.name, b.website_url, b.town, b.postcode,
    b.phone, b.group_name, b.verification_status,
    b.updated_at
FROM business_intel.businesses b
JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id
WHERE bv.slug = 'veterinary'
  AND b.verification_status = 'verified'
ORDER BY b.updated_at DESC
LIMIT 10;

-- Check vet-specific details
SELECT
    b.name, vd.species_treated, vd.emergency_service,
    vd.num_vets, vd.head_vet_name, vd.accepting_new_clients
FROM business_intel.businesses b
JOIN business_intel.vet_details vd ON vd.business_id = b.id
ORDER BY vd.updated_at DESC
LIMIT 10;

-- Check extracted prices
clients_db=# SELECT
    b.name as business,
    pr.name as service,
    pr.category,
    pp.price_gbp,
    pp.price_qualifier,
    pp.observed_at
FROM business_intel.product_prices pp
JOIN business_intel.products pr ON pr.id = pp.product_id
JOIN business_intel.businesses b ON b.id = pp.business_id
ORDER BY pp.observed_at DESC
LIMIT 20;
 business | service | category | price_gbp | price_qualifier | observed_at
----------+---------+----------+-----------+-----------------+-------------
(0 rows)

-- Does business_prices exist?
SELECT COUNT(*) FROM business_intel.business_prices;

-- If it does, check recent prices
SELECT b.name, bp.service_category, bp.service_name,
       bp.price_gbp, bp.price_qualifier
FROM business_intel.business_prices bp
JOIN business_intel.businesses b ON b.id = bp.business_id
WHERE bp.is_current = TRUE
ORDER BY b.created_at DESC
LIMIT 20;

-- Check completed task summaries
SELECT
    b.name, ct.status, ct.completed_at,
    ct.result_summary::text
FROM business_intel.collection_tasks ct
JOIN business_intel.businesses b ON b.id = ct.business_id
WHERE ct.status = 'completed'
ORDER BY ct.completed_at DESC
LIMIT 5;

-- Species treated
SELECT
    b.name, vpd.species_treated
FROM business_intel.vet_practice_details vpd
JOIN business_intel.businesses b ON b.id = vpd.business_id
WHERE vpd.species_treated IS NOT NULL
  AND array_length(vpd.species_treated, 1) > 0
ORDER BY b.updated_at DESC
LIMIT 10;



clients_db=# -- How many areas did the sweep orchestrator load?
SELECT collected_data->'unswept_areas'->>'count' as areas_loaded
FROM orchestration_states
WHERE orchestration_id = '56bbae1b-a28b-4ea8-b17d-b16252a92b63';
 areas_loaded
--------------
 3402
(1 row)


-- Sweep progress: areas swept vs remaining
SELECT
    CASE WHEN last_swept_at IS NULL THEN 'unswept' ELSE 'swept' END as status,
    COUNT(*)
FROM business_intel.search_areas
GROUP BY (last_swept_at IS NULL);

-- Discovery candidates found this run
SELECT status, COUNT(*)
FROM business_intel.discovery_candidates
GROUP BY status;

-- Most recent sweep activity
SELECT district_code, area_name, last_swept_at, candidates_found
FROM business_intel.search_areas
WHERE last_swept_at IS NOT NULL
ORDER BY last_swept_at DESC
LIMIT 10;

-- Current sweep orchestration progress (iter number from the logs)
SELECT orchestration_id, status, current_step, updated_at
FROM orchestration_states
WHERE owner_agent_type = 'area-sweep-orchestrator'
  AND status NOT IN ('COMPLETED', 'FAILED')
ORDER BY created_at DESC
LIMIT 1;


-- stuck

UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,sweep_loop,config,max_iterations}',
    '3500'::jsonb
),
updated_at = NOW()
WHERE type = 'area-sweep-orchestrator';

           orchestration_id           |      orchestration_name       |       status       | current_step | error |          updated_at
--------------------------------------+-------------------------------+--------------------+--------------+-------+-------------------------------
 c88c9d21-c70e-4d1e-a260-952592c1518f | generic-process-0217-1916     | COMPLETED          | complete     |       | 2026-02-17 19:22:11.817349+00
 34c15d37-70c2-45ef-a5e1-dcb291de9732 | generic-orchestrate-0217-1839 | AWAITING_RESPONSES | run_sweeps   |       | 2026-02-17 18:39:44.369539+00
 c4461620-5166-41cf-b1f8-12146e59e061 | generic-process-0217-1829     | COMPLETED          | complete     |       | 2026-02-17 18:31:45.297514+00
 1337ed9c-8a80-4dde-8072-ed47ed6abc72 | generic-process-0217-1757     | COMPLETED          | complete     |       | 2026-02-17 18:00:17.24124+00
 9af79661-76e3-4053-bd69-9c5c073588de | generic-process-0217-1746     | COMPLETED          | complete     |       | 2026-02-17 17:52:59.845283+00
(5 rows)


There it is — 34c15d37 is the pipeline, stuck at run_sweeps in AWAITING_RESPONSES. The sweep orchestrator completed at 20:22 but the pipeline never got the completion message.

-- Check what the pipeline is waiting for
SELECT
    key as request_id,
    value->>'step_name' as step_name,
    value->>'timeout_at' as timeout_at,
    value->>'responses_topic' as responses_topic
FROM orchestration_states,
     jsonb_each(awaited_requests)
WHERE orchestration_id = '34c15d37-70c2-45ef-a5e1-dcb291de9732';

-- Check if the sweep orchestrator sent its completion to the right topic
SELECT
    collected_data->>'__parent_responses_topic__' as parent_topic,
    collected_data->>'__my_responses_topic__' as my_topic
FROM orchestration_states
WHERE orchestration_id = '56bbae1b-a28b-4ea8-b17d-b16252a92b63';

-- 5. Reset orphaned in_progress tasks
UPDATE business_intel.collection_tasks
SET status = 'pending', started_at = NULL, orchestration_id = NULL
WHERE status = 'in_progress';

--

Step 1: Find expired awaited requests
sqlSELECT ar.request_id, ar.orchestration_id, ar.correlation_id,
       ar.step_id, ar.step_name, ar.retry_version,
       ar.responses_topic, ar.requests_topic,
       ar.timeout_at, ar.target_agent_type
FROM awaited_requests ar
WHERE ar.status = 'waiting'
  AND ar.timeout_at < NOW() - INTERVAL '30 seconds'
ORDER BY ar.timeout_at ASC
LIMIT 20
FOR UPDATE SKIP LOCKED
The 30-second grace period avoids racing with in-process goroutines that
might still be about to fire. LIMIT 20 prevents one sweep from taking too
long.
Step 2: For each expired request, classify the situation
sql-- Check if the child orchestration completed
SELECT os.orchestration_id, os.status, os.final_result
FROM orchestration_states os
WHERE os.orchestration_id = (
    SELECT child_os.orchestration_id
    FROM orchestration_states child_os
    WHERE child_os.parent_orchestration_id = $parent_orch_id
      AND child_os.status IN ('COMPLETED', 'FAILED')
    ORDER BY child_os.updated_at DESC
    LIMIT 1
)

  -----

restart failed pipeline
-- 1. Fail stuck orchestrations
UPDATE orchestration_states
SET status = 'FAILED', error = 'manual reset', updated_at = NOW()
WHERE status = 'AWAITING_RESPONSES'
  AND updated_at < NOW() - INTERVAL '1 hour';

-- 2. Reset orphaned tasks
UPDATE business_intel.collection_tasks
SET status = 'pending', started_at = NULL, orchestration_id = NULL
WHERE status = 'in_progress';

-- 3. Re-run pipeline