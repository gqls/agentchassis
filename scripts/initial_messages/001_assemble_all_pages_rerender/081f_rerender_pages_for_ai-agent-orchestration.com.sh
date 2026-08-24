#!/usr/bin/env bash
set -euo pipefail
#
# AGENT_TYPE="rerender-pages"
#
# Propagates migration 456 (foreground --color-primary -> --color-primary-ink) into
# ai-agent-orchestration.com's rendered pages. The migration changed 12
# content_components.html_template rows; rendered placements keep the OLD html until
# re-rendered, so this is the step that makes the fix visible.
#
# ⚠ TWO TRAPS BEFORE YOU RUN IT:
#   1. No orchestration dispatch within ~300s of an agent-chassis (re)start — the
#      spawn is silently dropped. Check:
#        kubectl -n ai-persona-system get pods -l app=agent-chassis \
#          -o jsonpath='{range .items[*]}{.status.startTime}{"\n"}{end}'
#   2. `kcat -P` can exit 0 having published NOTHING. Do not treat a clean exit as
#      evidence; confirm with the orchestration_states query printed at the end.
#
# ⚠ Pages whose components have NULL content_data CANNOT be re-rendered — there is
# nothing to rebuild the section from (bugs_closed/194). On this site that is
# `pricing` (5/5) and 7 others; they need a framework rebuild, not this.

SITE_ID="2a8ebf9c-20a2-4c39-b191-840b012371da"
DOMAIN="ai-agent-orchestration.com"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Payload lifted out of the heredoc so it can be published with an asserted receipt
# (bugs_open/327). Delimiter left UNQUOTED, exactly as before, so ${VAR} still expands.
PAYLOAD_327=$(cat <<JSON
{"action":"orchestrate","config":{"agent_type":"rerender-pages"},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}","refresh_site_components":true}}
JSON
)

REPO_ROOT_PUB="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel 2>/dev/null || true)"
if [ -z "$REPO_ROOT_PUB" ] || [ ! -f "$REPO_ROOT_PUB/scripts/kafka-publish-lib.sh" ]; then
  echo "ERROR: scripts/kafka-publish-lib.sh not found — refusing to publish unverified (bugs_open/327)." >&2
  return 1 2>/dev/null || exit 1
fi
. "$REPO_ROOT_PUB/scripts/kafka-publish-lib.sh"

PUBLISH_RC=0
kafka_publish_checked \
  --topic system.agent.generic.requests \
  --correlation "$CORRELATION_ID" \
  --payload "$PAYLOAD_327" \
  --header "orchestration_id=$ORCHESTRATION_ID" \
  --header "request_id=$REQUEST_ID" \
  --header "message_id=$MESSAGE_ID" \
  --header "message_type=request" \
  --header "client_id=demo_client" \
  --header "action=orchestrate" \
  --header "sender_agent_type=cli" \
  --header "sender_agent_id=cli-user" \
  --header "responses_topic=system.agent.generic.responses" \
  --header "timestamp=$TIMESTAMP" || PUBLISH_RC=$?

if [ "$PUBLISH_RC" -ne 0 ]; then
  echo "NOT DISPATCHED — no rerender will run (bugs_open/327)." >&2
  exit "$PUBLISH_RC"
fi

echo "CORRELATION_ID=$CORRELATION_ID  ORCHESTRATION_ID=$ORCHESTRATION_ID  REQUEST_ID=$REQUEST_ID TIMESTAMP=$TIMESTAMP"
echo
echo "CONFIRM IT LANDED (a clean kcat exit is NOT evidence):"
echo "  SELECT current_step, status, created_at FROM orchestration_states"
echo "   WHERE orchestration_id = '${ORCHESTRATION_ID}';"
echo "VERIFY AT THE ARTEFACT afterwards, never at the status:"
echo "  python3 scripts/render_audit.py https://${DOMAIN}/index.html https://${DOMAIN}/about.html"
