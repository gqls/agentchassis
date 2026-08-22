#!/bin/bash
# =============================================================================
# bugs_open/342 — the post-roll CANARY for the editor absent-required REFUSAL
# =============================================================================
# Migration 551 armed `refuse_absent_required_fields` on section-editor's
# apply_edit step. This proves it PROTECTS rather than merely being configured.
#
# TWO ARMS, and the second is the one people skip:
#   refuse   — a section whose render leaves schema-required fields EMPTY must
#              be REFUSED: the stored rendered_html must be byte-identical and a
#              required_fields_missing item must exist naming the fields.
#   control  — a CLEAN edit must still persist. An arm that only stops edits is
#              not a fix, it is an outage. (bugs_open/348 §8 makes the same point.)
#
# Both targets are deliberately NOT `deployed`, so a canary that fails harms no
# live page. Do not "improve" this by pointing it at a deployed row.
#
# Usage:  ./CANARY_342_editor_refusal.sh refuse
#         ./CANARY_342_editor_refusal.sh control
#
# Then read the verdict with the SQL the script prints (it does not poll for
# you: the dispatch queues behind the fleet and a canary that appears to fail
# is usually latency — see CLAUDE.md's ~30-minute budget).
# =============================================================================
set -euo pipefail
ARM="${1:-}"

DOMAIN="leopardessconsulting.co.uk"

case "$ARM" in
  refuse)
    # ai-agent-roi-estimator / tool-cta. content_data is PRESENT but carries
    # none of headline, description, trust_note — all three declared
    # source:"llm" + required, and all three referenced as bare {{.field}}
    # placeholders, so the render leaves them empty and the seam reports.
    # build_status=pending as of 2026-08-22.
    PC_ID="0a1498b3-a066-4d50-8a9f-f97b281830a1"
    BASELINE_MD5="69a2f28c0715e525335b1ddff0bfd47d"   # length 9220, 2026-08-22
    # A field_updates merge that changes nothing semantically: the point is to
    # make the action RENDER, not to alter content. The absent required fields
    # come from the existing row, exactly as a real edit would inherit them.
    FIELD_UPDATES='{"tagline": "AI systems that do a defined job, run without supervision, and keep a record of every decision they make."}'
    EXPECT="REFUSED — the step must fail with 'refusing to persist', md5 unchanged, and an item filed"
    ;;
  control)
    # use-cases / use-cases-list. Its ONE required source:"llm" field
    # (headline) is present and non-empty, so the seam reports nothing and the
    # armed gate must not fire. build_status=approved as of 2026-08-22.
    PC_ID="9737d0d9-4670-41f7-a987-288d83bbe1e9"
    BASELINE_MD5="d0f24daf68d076fa74d2cc842e9bf621"   # length 5846, 2026-08-22
    # Set the required field to ITS OWN CURRENT VALUE: the write path is
    # exercised in full while the content is semantically unchanged.
    FIELD_UPDATES='{"headline": "What the platform is built to do"}'
    EXPECT="PERSISTED — the step must succeed; a refusal here means the arm is stopping healthy edits"
    ;;
  *)
    echo "Usage: $0 [refuse|control]" >&2
    exit 1
    ;;
esac

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

INPUT_DATA="{\"domain\":\"${DOMAIN}\",\"page_component_id\":\"${PC_ID}\",\"edit_type\":\"content_edit\",\"field_updates\":${FIELD_UPDATES}}"

WORKFLOW=$(cat <<'ENDWF'
{"start_step":"spawn_section_editor","processing_mode":"orchestrator","timeout_seconds":900,
 "steps":{
  "spawn_section_editor":{"action":"spawn_agent","config":{"role":"section_editor","agent_type":"section-editor"},"output_field":"section_editor_agent","next_step":"call_section_editor","description":"Spawn section-editor agent"},
  "call_section_editor":{"action":"call_agent","config":{"agent_type":"section-editor","target_role":"section_editor","input_mapping":{"domain":"input_data.domain","edit_type":"input_data.edit_type","page_component_id?":"input_data.page_component_id","field_updates?":"input_data.field_updates"},"timeout_seconds":600},"output_field":"edit_result","next_step":"complete","description":"Run section edit"},
  "complete":{"action":"complete_workflow","config":{"output_fields":["edit_result"]},"description":"Section edit complete"}}}
ENDWF
)
WORKFLOW_COMPACT=$(echo "$WORKFLOW" | tr -d '\n' | sed 's/  */ /g')

echo "========================================="
echo "342 CANARY — arm: $ARM"
echo "  page_component: $PC_ID"
echo "  baseline md5:   $BASELINE_MD5"
echo "  expect:         $EXPECT"
echo "  correlation:    $CORRELATION_ID"
echo "========================================="

MESSAGE_BODY="{\"headers\":{\"correlation_id\":\"${CORRELATION_ID}\",\"orchestration_id\":\"${ORCHESTRATION_ID}\",\"request_id\":\"${REQUEST_ID}\",\"message_id\":\"${MESSAGE_ID}\",\"message_type\":\"request\",\"client_id\":\"demo_client\",\"action\":\"process\",\"sender\":{\"agent_id\":\"cli-user\",\"agent_type\":\"cli\",\"pod_name\":\"cli\"},\"timestamp\":\"${TIMESTAMP}\"},\"config\":{\"workflow\":${WORKFLOW_COMPACT}},\"input_data\":${INPUT_DATA}}"

kubectl -n kafka run -i --rm "kcat-342-canary-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H message_type=request \
  -H client_id=demo_client \
  -H action=process \
  -H sender_agent_type=cli \
  -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TIMESTAMP <<ENDKAFKA
${MESSAGE_BODY}
ENDKAFKA

cat <<EOF

CORRELATION_ID=$CORRELATION_ID

READ THE VERDICT (all three, not just the first — the point of the canary):

-- 1. did the step refuse, and with WHICH message?
SELECT current_step, status, left(error, 400)
  FROM orchestration_states
 WHERE collected_data->'input_data'->>'page_component_id' = '$PC_ID'
 ORDER BY created_at DESC LIMIT 3;

-- 2. is the stored artefact byte-identical? (the protection itself)
SELECT md5(COALESCE(rendered_html,'')) = '$BASELINE_MD5' AS unchanged,
       length(rendered_html), updated_at
  FROM page_components WHERE id = '$PC_ID';

-- 3. was the defect RECORDED as well as refused? (refusing must never be why
--    a defect goes unrecorded — the emit runs BEFORE the gate by design)
SELECT status, summary, spec->>'missing_fields', created_at
  FROM site_work_items
 WHERE item_type = 'required_fields_missing'
   AND created_at > now() - interval '1 hour'
 ORDER BY created_at DESC LIMIT 5;

-- 4. and the DRIVING item's terminal status — READ it, do not assume it benign.
--    Expect 'complete' to be WRONG until bugs_open/344 lands (the dispatch
--    loop tramples a failure flag). Its fingerprint:
SELECT id, status, completed_at, retry_after, retry_after > completed_at AS trampled
  FROM site_work_items
 WHERE updated_at > now() - interval '1 hour' AND status = 'complete'
   AND retry_after IS NOT NULL AND completed_at IS NOT NULL
 ORDER BY updated_at DESC LIMIT 5;
EOF
