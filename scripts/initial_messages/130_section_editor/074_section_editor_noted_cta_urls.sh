#!/usr/bin/env bash
# Re-render noted.co.uk's hero / call-to-action components so the CTA destinations
# set in content_data actually reach rendered_html.
#
# WHY THIS EXISTS — the thing that caught me out 2026-08-12:
#   `CTA_2026-08-12_noted_cta_destinations.sql` set the *_cta_url keys in
#   page_components.content_data (the source of truth). I then fired
#   `rerender-pages`, whose 5 page_rerender items all completed "success": true...
#   and rendered_html did not change by one byte. `page_rerender` RE-ASSEMBLES a
#   page from its EXISTING component HTML; it does not re-render a component from
#   content_data. The deploy even reported success with a commit message, because
#   the assembled page was byte-identical so git had nothing to commit.
#
#   The action that DOES re-render from content_data is the section editor's
#   `content_edit` (platform/orchestration/actions/section_editor_actions.go:215):
#     1. update content_data  2. re-render the component template with site
#     context  3. UPDATE page_components.rendered_html  4. reassemble the page
#     5. return assembled HTML for git_commit
#
# The field_updates below are the SAME values already in content_data, so the merge
# is a no-op and step 2 is the point. Idempotent: safe to re-run.
#
# SEQUENTIAL BY DESIGN. Two slots on one page each trigger a full page reassembly;
# firing them concurrently races two assemblies and two commits for the same file.
# One at a time, each waited on.
#
# Usage:  ./074_section_editor_noted_cta_urls.sh [index-hero|all]
#         Run the single canary FIRST — the spawn->call handshake on this path is
#         known-flaky (see memory `spawn-call-handshake-races`), and a failure is
#         not evidence that the edit itself is wrong. Never cancel a failed row
#         before diagnosing it.

set -euo pipefail

DOMAIN="noted.co.uk"
CLIENT_ID="demo_client"
MODE="${1:-index-hero}"

# page | slot | field_updates  — destinations match the copy the framework wrote.
# migrate's PRIMARY is absent on purpose: "Save everything" is a local-data rescue
# whose destination is the /legacy page that PLAN 4.3 has not built yet.
EDITS=(
  'index|hero|{"cta_url":"https://app.noted.co.uk/","secondary_cta_url":"/how-it-works.html"}'
  'index|call-to-action|{"primary_cta_url":"https://app.noted.co.uk/","secondary_cta_url":"/how-it-works.html"}'
  'how-it-works|hero|{"cta_url":"https://app.noted.co.uk/","secondary_cta_url":"/migrate.html"}'
  'how-it-works|call-to-action|{"primary_cta_url":"https://app.noted.co.uk/","secondary_cta_url":"/migrate.html"}'
  'migrate|hero|{"secondary_cta_url":"/how-it-works.html"}'
  'migrate|call-to-action|{"secondary_cta_url":"/how-it-works.html"}'
)

PSQL=(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA)

WORKFLOW_COMPACT='{"start_step":"spawn_section_editor","processing_mode":"orchestrator","timeout_seconds":900,"steps":{"spawn_section_editor":{"action":"spawn_agent","config":{"role":"section_editor","agent_type":"section-editor"},"output_field":"section_editor_agent","next_step":"call_section_editor","description":"Spawn section-editor agent"},"call_section_editor":{"action":"call_agent","config":{"agent_type":"section-editor","target_role":"section_editor","input_mapping":{"domain":"input_data.domain","edit_type":"input_data.edit_type","page_name?":"input_data.page_name","slot_name?":"input_data.slot_name","field_updates?":"input_data.field_updates"},"timeout_seconds":600},"output_field":"edit_result","next_step":"complete","description":"Run section edit"},"complete":{"action":"complete_workflow","config":{"output_fields":["edit_result"]},"description":"Section edit complete"}}}'

fire_one() {
  local page="$1" slot="$2" updates="$3"
  local CORRELATION_ID ORCHESTRATION_ID REQUEST_ID MESSAGE_ID TIMESTAMP INPUT_DATA
  CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
  ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
  REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
  MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
  TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
  INPUT_DATA="{\"domain\":\"${DOMAIN}\",\"page_name\":\"${page}\",\"slot_name\":\"${slot}\",\"edit_type\":\"content_edit\",\"field_updates\":${updates}}"

  echo "--- ${page} / ${slot}"
  echo "    corr=${CORRELATION_ID}"

  kubectl -n kafka run -i --rm --quiet "kcat-sed-$(date +%s)-$RANDOM" \
    --image=edenhill/kcat:1.7.1 --restart=Never -- \
    kcat -P -c 1 \
    -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.agent.generic.requests \
    -H correlation_id=$CORRELATION_ID -H orchestration_id=$ORCHESTRATION_ID \
    -H request_id=$REQUEST_ID -H message_id=$MESSAGE_ID \
    -H message_type=request -H client_id=$CLIENT_ID -H action=process \
    -H sender_agent_type=cli -H sender_agent_id=cli-user \
    -H responses_topic=system.agent.generic.responses -H timestamp=$TIMESTAMP \
    >/dev/null <<ENDKAFKA
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":${WORKFLOW_COMPACT}},"input_data":${INPUT_DATA}}
ENDKAFKA

  # Wait on the ORCHESTRATION ROW, not on kcat's exit code — kcat exits 0 without
  # publishing (see memory `kcat-publish-silently-drops`).
  local i st
  for i in $(seq 1 40); do
    st=$("${PSQL[@]}" -c "SELECT status||'/'||current_step FROM orchestration_states WHERE correlation_id='${CORRELATION_ID}'::uuid LIMIT 1;" 2>/dev/null | tr -d '[:space:]')
    if [ -n "$st" ]; then
      case "$st" in
        COMPLETED/*) echo "    -> $st"; return 0 ;;
        FAILED/*)    echo "    -> $st  (do NOT cancel; diagnose)"; return 1 ;;
      esac
    fi
    sleep 6
  done
  echo "    -> no terminal status after 4 min (last: ${st:-<no row>}) — latency or a dropped publish; check before re-firing"
  return 1
}

case "$MODE" in
  index-hero) IFS='|' read -r p s u <<< "${EDITS[0]}"; fire_one "$p" "$s" "$u" ;;
  all)
    rc=0
    for e in "${EDITS[@]}"; do
      IFS='|' read -r p s u <<< "$e"
      fire_one "$p" "$s" "$u" || rc=1
    done
    exit $rc ;;
  *) echo "Usage: $0 [index-hero|all]"; exit 1 ;;
esac
