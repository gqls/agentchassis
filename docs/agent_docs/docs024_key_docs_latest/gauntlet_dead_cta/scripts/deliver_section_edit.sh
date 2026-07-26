#!/usr/bin/env bash
# deliver_section_edit.sh <page_component_id> <field_updates.json>
#
# Generalised from deliver_gauntlet_section_edit.sh (which hardcodes both the
# component and its copy inline). Same envelope, same agent, same guarantees —
# the only change is that the component and its field_updates are arguments, so
# the script can deliver any owned-page section on this site without being
# edited and re-edited.
#
# WHY section-editor: these pages are rebuild_policy='owned'. The generic
# rerender->save_page_sections path is HARD-REFUSED for owned pages
# (bugs_closed/024). apply_section_edit (content_edit) is the sanctioned path:
# it re-renders exactly ONE page_component from its current template, merges
# field_updates into content_data, reassembles the page from the other stored
# sections, and commits. Blast radius: one section.
#
# WHY the direct orchestrator envelope (spawn_agent + call_agent, action=process)
# rather than the bare action=orchestrate one: the bare envelope has silently
# failed to ingest on this page before — no error, no orchestration row, no work
# item (049b, a kubectl-run stdin race).
#
# AFTER this: apply_section_edit does NOT republish js_content. Follow with
# republish_gauntlet_js.sh (assemble-only rerender), then set the component's
# build_status back to 'deployed' — section-editor leaves it 'approved'.
set -euo pipefail

PAGE_COMPONENT_ID="${1:?usage: deliver_section_edit.sh <page_component_id> <field_updates.json>}"
FIELD_UPDATES_FILE="${2:?usage: deliver_section_edit.sh <page_component_id> <field_updates.json>}"

SITE_ID="${SITE_ID:-9ec3b9ee-5b08-461b-b4f8-9e1e03579c74}"
DOMAIN="${DOMAIN:-vonc.com}"
CLIENT_ID="${CLIENT_ID:-demo_client}"

[ -f "$FIELD_UPDATES_FILE" ] || { echo "no such file: $FIELD_UPDATES_FILE" >&2; exit 1; }

# applyContentEdit requires a non-empty merge, and the envelope is a single-line
# heredoc — so compact the JSON and refuse an empty object up front rather than
# discovering it as a silent no-op downstream.
FIELD_UPDATES="$(python3 -c '
import json, sys
d = json.load(open(sys.argv[1]))
if not isinstance(d, dict) or not d:
    raise SystemExit("field_updates must be a non-empty JSON object")
print(json.dumps(d, separators=(",", ":"), ensure_ascii=False))' "$FIELD_UPDATES_FILE")"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "SAVE: CORRELATION_ID=$CORRELATION_ID"
echo "      page_component=$PAGE_COMPONENT_ID  fields=$(python3 -c 'import json,sys;print(len(json.load(open(sys.argv[1]))))' "$FIELD_UPDATES_FILE")"

kubectl -n kafka run -i --rm "kcat-section-edit-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "message_type=request" \
  -H "client_id=$CLIENT_ID" \
  -H "action=process" \
  -H "sender_agent_type=cli" \
  -H "sender_agent_id=cli-user" \
  -H "responses_topic=system.agent.generic.responses" \
  -H "timestamp=$TIMESTAMP" <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_editor","processing_mode":"orchestrator","timeout_seconds":300,"steps":{"spawn_editor":{"action":"spawn_agent","config":{"role":"section_editor","agent_type":"section-editor"},"output_field":"editor_agent","next_step":"call_editor","description":"Spawn section-editor"},"call_editor":{"action":"call_agent","config":{"agent_type":"section-editor","target_role":"section_editor","input_mapping":{"domain":"input_data.domain","site_id":"input_data.site_id","page_component_id":"input_data.page_component_id","edit_type":"input_data.edit_type","field_updates":"input_data.field_updates"},"timeout_seconds":260},"output_field":"edit_result","next_step":"complete","description":"Re-render the section from its current template, reassemble, deploy"},"complete":{"action":"complete_workflow","config":{"output_fields":["edit_result"]},"description":"Section edit complete"}}}},"input_data":{"domain":"${DOMAIN}","site_id":"${SITE_ID}","page_component_id":"${PAGE_COMPONENT_ID}","edit_type":"content_edit","field_updates":${FIELD_UPDATES}}}
JSON

echo ""
echo "Watch (find it by PAYLOAD, not by the printed id — and treat a missing row"
echo "as QUEUED, not dropped; ingest under load has been measured in minutes):"
echo "  SELECT status,current_step FROM orchestration_states WHERE correlation_id='$CORRELATION_ID'::uuid;"
