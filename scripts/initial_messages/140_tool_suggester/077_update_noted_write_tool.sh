#!/usr/bin/env bash
# UPDATE the noted.co.uk EDITOR ("Write") in place — the replace_existing arm of
# create_tool_component (bugs_open/331, TL-047). 076 was the birth; this is the
# regeneration: the supplied html replaces the incumbent verbatim (no LLM), the
# live placements rendered_html is overwritten in the same transaction, and a
# visible-text shrink guard refuses a hollow replacement.
#
# WHY (2026-08-24): pasteboard stage 1 — media (images/GIFs/video/audio) joins
# the editor. DEPLOY ORDER: the ENGINE must be live first (the UI calls
# DELETE /api/media/{id}, kind=video and /api/me max_upload, all absent from the
# pre-2026-08-24 binary). Re-run test_editor_degraded.py after ANY edit to the
# html — all seven guards there are mutation-verified.
#
# ⚠ A REGENERATION DOES NOT ENQUEUE THE PAGES BUILD (stated in the action, ~line
# 300). After this completes, hand-file the page_rerender item for tool-write —
# copy a rows whole shape AND RESET THE page_id COLUMN (the handler reads the
# COLUMN, not your spec — this lane deployed the wrong page once already).
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
HTML_FILE="${REPO_ROOT}/docs/agent_docs/docs024_key_docs_latest/noted_rebuild/editor_tool/noted-write.html"

SITE_ID="b50a8da1-25bd-4a6d-88f2-4d4c93f6acc8"
DOMAIN="noted.co.uk"
FUNCTION="write"
DISPLAY_NAME="Write"
DESCRIPTION="The Noted editor: sign in or make an account, put a thought down, and find it again from any browser you sign in from. Says Saved only when the server has actually saved."

[ -f "$HTML_FILE" ] || { echo "missing $HTML_FILE"; exit 1; }

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "========================================="
echo "UPDATE tool component in place: ${FUNCTION}"
echo "  site:  ${DOMAIN} (${SITE_ID})"
echo "  html:  $(wc -c < "$HTML_FILE") bytes"
echo "  corr:  ${CORRELATION_ID}"
echo "========================================="

# Build the message with python so the HTML (quotes, newlines, <script>) is escaped
# properly. Hand-rolled shell quoting of a 13 KB HTML blob is how a payload arrives
# truncated and the action refuses it as "structurally incomplete".
BODY_FILE="$(mktemp)"
trap 'rm -f "$BODY_FILE"' EXIT

# Values go through the ENVIRONMENT and the heredoc is QUOTED ('PYEOF'), so the
# shell performs no expansion inside it. An unquoted heredoc here does not survive
# a body containing JSON, quotes and ${...}: the first attempt had bash trying to
# execute the Python as a command.
export LR_HTML_FILE="$HTML_FILE" LR_BODY_FILE="$BODY_FILE" \
       LR_CORRELATION_ID="$CORRELATION_ID" LR_ORCHESTRATION_ID="$ORCHESTRATION_ID" \
       LR_REQUEST_ID="$REQUEST_ID" LR_MESSAGE_ID="$MESSAGE_ID" \
       LR_CLIENT_ID="$CLIENT_ID" LR_TIMESTAMP="$TIMESTAMP" \
       LR_SITE_ID="$SITE_ID" LR_DOMAIN="$DOMAIN" LR_FUNCTION="$FUNCTION" \
       LR_DISPLAY_NAME="$DISPLAY_NAME" LR_DESCRIPTION="$DESCRIPTION"

python3 <<'PYEOF'
import json, os
html = open(os.environ["LR_HTML_FILE"], encoding="utf-8").read()
workflow = {
  "start_step": "create_tool",
  "processing_mode": "orchestrator",
  "timeout_seconds": 900,
  "steps": {
    "create_tool": {
      "action": "create_tool_component",
      "config": {
        "site_id":      "input_data.site_id",
        "function":     "input_data.function",
        "display_name": "input_data.display_name",
        "description":  "input_data.description",
        "html_content": "input_data.html_content",
        "replace_existing": True,
        "category":     "interactive",
        "in_header":    False,
        "in_footer":    False
      },
      "output_field": "create_result",
      "next_step": "complete",
      "description": "Regenerate the write tool component in place (replace_existing)"
    },
    "complete": {
      "action": "complete_workflow",
      "config": {"output_fields": ["create_result"]},
      "description": "done"
    }
  }
}
env = os.environ
msg = {
  "headers": {
    "correlation_id": env["LR_CORRELATION_ID"], "orchestration_id": env["LR_ORCHESTRATION_ID"],
    "request_id": env["LR_REQUEST_ID"], "message_id": env["LR_MESSAGE_ID"],
    "message_type": "request", "client_id": env["LR_CLIENT_ID"], "action": "process",
    "sender": {"agent_id": "cli-user", "agent_type": "cli", "pod_name": "cli"},
    "timestamp": env["LR_TIMESTAMP"]
  },
  "config": {"workflow": workflow},
  "input_data": {
    "site_id": env["LR_SITE_ID"], "domain": env["LR_DOMAIN"],
    "function": env["LR_FUNCTION"], "display_name": env["LR_DISPLAY_NAME"],
    "description": env["LR_DESCRIPTION"],
    "html_content": html
  }
}
open(env["LR_BODY_FILE"], "w", encoding="utf-8").write(json.dumps(msg, separators=(",", ":")))
PYEOF

echo "message: $(wc -c < "$BODY_FILE") bytes"

kubectl -n kafka run -i --rm "kcat-tool-write-upd-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -c 1 \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID -H orchestration_id=$ORCHESTRATION_ID \
  -H request_id=$REQUEST_ID -H message_id=$MESSAGE_ID \
  -H message_type=request -H client_id=$CLIENT_ID -H action=process \
  -H sender_agent_type=cli -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TIMESTAMP < "$BODY_FILE"

cat <<NOTES

=== Verify (kcat exit 0 is not delivery) ===

  SELECT owner_agent_type, status, current_step, error
  FROM orchestration_states WHERE correlation_id='${CORRELATION_ID}';

  -- the regeneration wrote the component AND the placement in one tx:
  SELECT cc.function, cc.updated_at, length(cc.html_content) AS html_len
  FROM content_components cc
  JOIN page_components pc ON pc.component_id=cc.id
  JOIN pages p ON p.id=pc.page_id
  WHERE p.site_id='${SITE_ID}' AND cc.function LIKE '%write%' AND cc.component_level='tool';

Then: hand-file the page_rerender for tool-write (RESET the page_id COLUMN),
wait for sitesync, and run the live smoke against the apex:
  /home/ant/.venvs/vonc_pw/bin/python docs/agent_docs/docs024_key_docs_latest/noted_rebuild/editor_tool/smoke_live_editor.py https://noted.co.uk
NOTES
