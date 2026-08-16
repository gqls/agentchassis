#!/usr/bin/env bash
# Bind noted.co.uk's two experience patterns to the BUILT components, through
# bind_site_experience so its four refusal doors run (unclosed binding, empty
# value, anchorless selector, dead page). PLAN §4 step 1.
#
# Both patterns are `draft`, so the fork is recorded `proposed` — that is the
# first rung of proposed → bound → verified, not a shortcut past it.
#
# Every selector below is a real element id on a deployed page (read from
# editor_tool/noted-write.html and legacy_tool/noted-legacy-rescue.html).
# The schema change that made adopt_control optional is in
# EXPERIENCES_2026-08-15_bind_noted_patterns.sql — apply it first.
set -euo pipefail

SITE_ID="b50a8da1-25bd-4a6d-88f2-4d4c93f6acc8"
CLIENT_ID="demo_client"
PSQL=(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA)

bind() {
  local pattern="$1" bindings="$2"
  local C=$(cat /proc/sys/kernel/random/uuid) O=$(cat /proc/sys/kernel/random/uuid) \
        R=$(cat /proc/sys/kernel/random/uuid) M=$(cat /proc/sys/kernel/random/uuid) \
        T=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
  echo "--- bind ${pattern}   corr=${C}"
  kubectl -n kafka run -i --rm --quiet "kcat-bind-$(date +%s)-$RANDOM" \
    --image=edenhill/kcat:1.7.1 --restart=Never -- \
    kcat -P -c 1 -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.agent.generic.requests \
    -H correlation_id=$C -H orchestration_id=$O -H request_id=$R -H message_id=$M \
    -H message_type=request -H client_id=$CLIENT_ID -H action=process \
    -H sender_agent_type=cli -H sender_agent_id=cli-user \
    -H responses_topic=system.agent.generic.responses -H timestamp=$T >/dev/null <<EOF
{"headers":{"correlation_id":"$C","orchestration_id":"$O","request_id":"$R","message_id":"$M","message_type":"request","client_id":"$CLIENT_ID","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"$T"},"config":{"workflow":{"start_step":"bind","processing_mode":"orchestrator","timeout_seconds":300,"steps":{"bind":{"action":"bind_site_experience","config":{"created_by":"noted-lane 2026-08-15","bindings_field":"input_data.experience_bindings"},"output_field":"bind_result","next_step":"complete"},"complete":{"action":"complete_workflow","config":{"output_fields":["bind_result"]}}}}},"input_data":{"site_id":"$SITE_ID","pattern_name":"$pattern","experience_bindings":$bindings}}
EOF
  for i in $(seq 1 30); do
    st=$("${PSQL[@]}" -c "SELECT status||'/'||current_step||' '||COALESCE(left(error,300),'') FROM orchestration_states WHERE correlation_id='$C'::uuid;" 2>/dev/null)
    case "$st" in
      COMPLETED/complete*) echo "    -> $st"; return 0;;
      *FAILED*|*complete_error*) echo "    -> $st"; "${PSQL[@]}" -c "SELECT collected_data->'__step_error'->>'message' FROM orchestration_states WHERE correlation_id='$C'::uuid;"; return 1;;
    esac
    sleep 6
  done
  echo "    -> no terminal status in 3 min"; return 1
}

# Re-running is safe: the action upserts ON CONFLICT (site_id, pattern_name, instance_key).
bind authenticated-note-sync '{"tool_section":"#noted-write","sign_in_form":"#nw-auth-form","email_input":"#nw-email","password_input":"#nw-password","sign_in_submit":"#nw-signin","note_list":"#nw-list","note_editor":"#nw-content","save_indicator":"#nw-status","api_base":"/api","sample_email":"noted-check@example.invalid","sample_password":"check-0123456789-noted"}'

bind legacy-local-data-adoption '{"tool_section":"#legacy-rescue","summary_region":"#lr-counts","download_control":"#lr-download"}'

echo; echo "=== result ==="
"${PSQL[@]}" -c "SELECT pattern_name||' | '||status||' | '||left(bindings::text,90) FROM site_experiences WHERE site_id='$SITE_ID';"
