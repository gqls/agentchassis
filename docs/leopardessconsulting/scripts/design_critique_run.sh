#!/bin/bash
# ============================================================================
# design_critique_run.sh — fire design-critique-agent (features_open/018 Phase 1,
# seed 645, register SQ-003) at ONE site, the way that actually gets it a storage
# client: as a site_work_items row that build-dispatch-loop SPAWNS.
#
# WHY NOT orchestrate_safe.sh: the generic topic runs an agent INLINE on a standing
# chassis pod, which deliberately has no storage client (owner ruling 2026-08-08),
# so execute_vision_prompt fails 'no storage client — cannot download screenshots'
# while the audit + measured-findings half reads green (complete_no_critique).
# bugs_open/243 candidate 2 option b documented this on 2026-08-11 for
# tool-acceptance-agent; the first manual critic run (corr 95f6b328, 2026-08-26)
# re-proved it. The isStorageEnabledAgent grant only reaches a SPAWNED pod.
#
# The dispatch is asynchronous: build-dispatch-loop claims within ~1 min and
# spawns agent-design-critique-agent-<hash>; the audit takes ~3 min, the vision
# call another minute or two. Follow it with the queries printed below.
#
# Report lands in doc_notes: subject_type='pipeline', subject_key='design-critique',
# categories ['design-report'], site_id set. Measured findings (contrast/broken
# images) file through the seed-301 drain with its dedup keys.
#
# Usage: ./design_critique_run.sh <site_id> <domain> [control_leg_label]
# ============================================================================
set -euo pipefail
SITE="${1:?site_id}"
DOMAIN="${2:?domain}"
LEG="${3:-manual}"
STAMP=$(date -u +%Y-%m-%dT%H%M%SZ)

ITEM_ID=$(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A <<SQL
INSERT INTO site_work_items
  (site_id, source, pipeline, item_type, severity, summary, spec,
   priority, handler_agent, status, created_by, item_key, triaged_at, approval_mode, max_attempts)
VALUES
  ('$SITE', 'manual', 'build', 'design_critique_run', 'low',
   '018 design critique: $DOMAIN ($LEG, via design_critique_run.sh)',
   jsonb_build_object('reason','manual_design_critique','control_leg','$LEG','domain','$DOMAIN'),
   90, 'design-critique-agent', 'triaged',
   'design_critique_run.sh ($(whoami 2>/dev/null || echo cli))',
   'design_critique_run:$DOMAIN:$STAMP', now(), 'auto', 2)
RETURNING id;
SQL
)
ITEM_ID=$(echo "$ITEM_ID" | head -1)
echo "site:       $DOMAIN ($SITE)"
echo "work item:  $ITEM_ID   (item_key design_critique_run:$DOMAIN:$STAMP)"
echo ""
echo "Follow it:"
echo "  kubectl -n ai-persona-system get pods | grep design-critique"
echo "  SELECT status, claimed_by, error FROM site_work_items WHERE id='$ITEM_ID';"
echo "  SELECT current_step, status FROM orchestration_states WHERE collected_data->'input_data'->>'item_id'='$ITEM_ID' ORDER BY updated_at DESC LIMIT 1;"
echo "  SELECT left(body,3000) FROM doc_notes WHERE categories ? 'design-report' AND site_id='$SITE' ORDER BY created_at DESC LIMIT 1;"
echo ""
echo "complete            = audit + measured findings + report"
echo "complete_no_critique = audit + measured findings, NO report (read __step_error)"
