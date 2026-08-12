#!/usr/bin/env bash
# Reconcile noted.co.uk's site plan against its realised pages — WITHOUT re-planning.
#
# WHY AN INLINE WORKFLOW AND NOT `build-site-planner`: reconcile_site_plan lives
# inside that agent's workflow, but the agent RE-DERIVES the plan first. This site
# has five working pages built from the current plan; re-planning them to add one
# page is the wrong trade. reconcile_site_plan is pure read-decide-write with no
# LLM, so it can be called on its own.
#
# What it does (reconcile_site_plan_action.go, decision order):
#   1. deployed at current plan version, or deployed older-but-same-composition -> skip
#   2. missing / not deployed / genuinely re-composed                           -> candidate
#   3. tool or game role, OR pages.rebuild_policy='owned'  -> owned_page_review
#      (needs_human_review, NO handler) because the generic builder clobbers
#      owned pages. This is a GUARD, not a failure.
#   4. an open work item already exists for the key                             -> skip
#   5. otherwise                                                                -> needs_page
#
# MEASURED 2026-08-12 18:45, first run after
# BACKFILL_2026-08-12_structure_spec_and_site_plan.sql:
#   pages_total 8, pages_emitted 2, pages_skipped_built 6,
#   pages_review_emitted 0, rerender_emitted true
#   - privacy                   -> needs_page   (no pages row at all)
#   - tool-legacy-rescue-guide  -> needs_page   (page exists, build_status='planned')
#   - the five deployed pages   -> skipped
#   - plus the terminal needs_rerender
#
# ⚠ I PREDICTED tool-legacy-rescue would raise an owned_page_review (rule 3, role
#   =tool) and it did NOT — rule 1 wins first: the page is deployed at the current
#   plan version, so it is skipped as built before the role is ever considered.
#   Recorded because the wrong version of this comment would have had the next
#   session hunting a human-review item that is not coming. Rule 3 fires on a
#   tool/owned page that is MISSING or NOT DEPLOYED, not on one that is fine.
#
# It is idempotent: idx_swi_dedup stops repeated runs duplicating items.

set -euo pipefail

SITE_ID="b50a8da1-25bd-4a6d-88f2-4d4c93f6acc8"
DOMAIN="noted.co.uk"
CLIENT_ID="demo_client"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

WORKFLOW='{"start_step":"reconcile","processing_mode":"orchestrator","timeout_seconds":600,"steps":{"reconcile":{"action":"reconcile_site_plan","config":{"target_site_id":"input_data.target_site_id"},"output_field":"reconcile_result","next_step":"complete","description":"Diff site_plan_pages vs realised pages; emit needs_page for the delta"},"complete":{"action":"complete_workflow","config":{"output_fields":["reconcile_result"]},"description":"done"}}}'

echo "=== reconcile_site_plan: ${DOMAIN} ==="
echo "  corr: ${CORRELATION_ID}"

kubectl -n kafka run -i --rm "kcat-reconcile-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -c 1 \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID -H orchestration_id=$ORCHESTRATION_ID \
  -H request_id=$REQUEST_ID -H message_id=$MESSAGE_ID \
  -H message_type=request -H client_id=$CLIENT_ID -H action=process \
  -H sender_agent_type=cli -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TIMESTAMP <<ENDKAFKA
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":${WORKFLOW}},"input_data":{"target_site_id":"${SITE_ID}","domain":"${DOMAIN}"}}
ENDKAFKA

cat <<NOTES

=== Verify at the ROW, not at kcat's exit code ===

  SELECT status, current_step, collected_data->'reconcile_result'
  FROM orchestration_states WHERE correlation_id='${CORRELATION_ID}';

  SELECT wi.item_type, wi.status, wi.spec->>'page_name' AS page, wi.created_at
  FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
  WHERE s.domain='noted.co.uk' AND wi.created_at > now()-interval '10 minutes'
  ORDER BY wi.created_at;
NOTES
