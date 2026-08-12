#!/usr/bin/env bash
# Create the noted.co.uk /legacy rescue tool as a framework tool component.
#
# WHAT IT IS: a page that reads the PREVIOUS version of Noted's IndexedDB
# (`NotedDB`, same origin) and hands the person their notes, recordings, photos and
# history back as the same file the old app's "Save everything" produces. This is
# PLAN §4 step 3 — the migration itself, not a fallback: IndexedDB is keyed by
# origin, so at cutover everyone's notes are still in their browser.
#
# WHY THE FRAMEWORK PATH AND NOT A HAND-BUILT PAGE: owner ruling 2026-08-04, every
# site goes through the framework. `create_tool_component` also canonicalises the
# page identity — the flat /tools/<fn>.html hand-roll is exactly what bugs_open/080
# is about, so the URL is the framework's to decide, not ours.
#
# ⚠ THE RESULTING URL IS NOT /legacy.html. CanonicalisePage(role="tool") yields
#   name="tool-legacy-rescue", url=/tools/legacy-rescue.html (flat-URL site).
#   migrate's CTAs currently point at /legacy.html and therefore render NO button
#   (the renderer drops a destination that is not a real page). After this lands,
#   re-point them with:
#     scripts/initial_messages/130_section_editor/074_section_editor_noted_cta_urls.sh migrate
#   having first updated that script's EDITS entries to the real URL printed below.
#
# THE HTML+JS IS TESTED, and the tests are proven to fail:
#   docs/.../noted_rebuild/legacy_tool/test_legacy_rescue.py  (24 checks, Playwright)
#   Mutation-verified 2026-08-12: removing the abort() that prevents creating a
#   stray database fails "LEAVES NO DATABASE BEHIND"; ignoring the legacy
#   single-blob record shape fails the recording-count and rescue checks.
#   Re-run it after ANY edit to the html — the runner cannot seed IndexedDB, so
#   this probe is the only thing that exercises the code against data (HANDOFF §5).

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
HTML_FILE="${REPO_ROOT}/docs/agent_docs/docs024_key_docs_latest/noted_rebuild/legacy_tool/noted-legacy-rescue.html"

SITE_ID="b50a8da1-25bd-4a6d-88f2-4d4c93f6acc8"
DOMAIN="noted.co.uk"
FUNCTION="legacy-rescue"
DISPLAY_NAME="Your notes from the old Noted"
DESCRIPTION="Finds notes, voice recordings and photos saved by the previous version of Noted in this browser and saves them to a file. Read-only: it never changes or deletes anything, and no account is needed."

[ -f "$HTML_FILE" ] || { echo "missing $HTML_FILE"; exit 1; }

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "========================================="
echo "Create tool component: ${FUNCTION}"
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
        "category":     "interactive",
        "in_header":    False,
        "in_footer":    False
      },
      "output_field": "create_result",
      "next_step": "complete",
      "description": "Create the legacy-rescue tool component and its page"
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

kubectl -n kafka run -i --rm "kcat-tool-legacy-$(date +%s)" \
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

  -- the component, the page, and the REAL url to point the CTAs at
  SELECT cc.function, cc.component_level, length(cc.js_content) AS js_len
  FROM content_components cc WHERE cc.function LIKE '%legacy-rescue%';

  SELECT p.name, p.url, p.page_type, p.build_status
  FROM pages p JOIN sites s ON s.id=p.site_id
  WHERE s.domain='noted.co.uk' AND p.name LIKE 'tool-%';

Then re-point migrate's CTAs at that url (074 script) and re-run the Playwright
probe against the DEPLOYED page, not just the local file.
NOTES
