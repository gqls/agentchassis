#!/usr/bin/env bash
# dispatch_content_feed_orchestrator.sh — one direct run of content-feed-orchestrator for
# idea.uk, with a RECEIPT (scripts/kafka-publish-lib.sh, OPP-009 / bugs_open/327).
#
# The framework's own route is content-feed-trigger (scheduled_tasks.content-feed-refresh,
# every 21600 s); this script exists so the first run after SQL_2026-08-25_arm_news_feed.sql
# can be watched in-session instead of waiting up to six hours. The orchestrator is
# idempotent at the seeding step (ON CONFLICT DO NOTHING on (site_id, name)), so a run
# that overlaps the trigger's costs one duplicate fetch, not duplicate sources.
#
# Preconditions (read them, do not assume them):
#   - classification spec carries content_features.news_feed.recommended=true (the seed);
#   - no chassis pod (re)started in the last ~300 s — a dispatch in that window is dropped
#     silently (CLAUDE.md): kubectl -n ai-persona-system get pods -l app=agent-chassis
#
# Payload shape is 082_submit_domain_unified.sh's: the chassis resolves the agent_type's
# default_config, so no workflow is inlined here (080_test_content_feed_orchestrator.sh
# inlines one and uses the racing `kubectl run -i` form — do not copy it).
set -euo pipefail

SITE_ID='1244516d-014d-421c-88c6-090bb1e9552a'   # idea.uk
DOMAIN='idea.uk'

REPO_ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel 2>/dev/null || true)"
if [ -z "$REPO_ROOT" ] || [ ! -f "$REPO_ROOT/scripts/kafka-publish-lib.sh" ]; then
  echo "ERROR: scripts/kafka-publish-lib.sh not found (repo root: ${REPO_ROOT:-<not in a git repo>})." >&2
  exit 1
fi
. "$REPO_ROOT/scripts/kafka-publish-lib.sh"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)

PUBLISH_RC=0
kafka_publish_checked \
  --topic system.agent.generic.requests \
  --correlation "$CORRELATION_ID" \
  --payload "{\"action\":\"orchestrate\",\"config\":{\"agent_type\":\"content-feed-orchestrator\"},\"input_data\":{\"site_id\":\"$SITE_ID\"}}" \
  --header "request_id=$REQUEST_ID" \
  --header "message_id=$MESSAGE_ID" \
  --header "orchestration_id=$ORCHESTRATION_ID" \
  --header "orchestration_name=content-feed-${DOMAIN}-$(date +%Y%m%d-%H%M%S)" \
  --header "step_name=start" \
  --header "client_id=demo_client" \
  --header "message_type=request" \
  --header "action=orchestrate" \
  --header "from_agent_type=user" \
  --header "from_agent_id=cli" \
  --header "responses_topic=system.agent.generic.responses" || PUBLISH_RC=$?

if [ "$PUBLISH_RC" -ne 0 ]; then
  echo "DISPATCH DID NOT GO OUT — nothing queued for ${DOMAIN}. Re-run: $0" >&2
  exit "$PUBLISH_RC"
fi

echo "SAVE: CORRELATION_ID=${CORRELATION_ID}  ORCHESTRATION_ID=${ORCHESTRATION_ID}"

LANDING_RC=0
kafka_verify_landing "$CORRELATION_ID" 120 || LANDING_RC=$?
echo "Find it later by:  SELECT status, current_step, error FROM orchestration_states WHERE correlation_id='${CORRELATION_ID}'::uuid;"
exit "$LANDING_RC"
