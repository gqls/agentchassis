#!/usr/bin/env bash
# dispatch_content_feed_orchestrator.sh <domain> — one direct run of
# content-feed-orchestrator for ANY enrolled site, with a RECEIPT
# (scripts/kafka-publish-lib.sh, OPP-009 / bugs_open/327).
#
# Generalised 2026-09-04 from idea_uk_vm_site/scripts/dispatch_content_feed_orchestrator.sh,
# which is hardcoded to idea.uk. That copy is left alone (another lane owns it); this one
# resolves site_id from the domain so enabling a new site does not mint another copy.
#
# The framework's own route is content-feed-trigger (scheduled_tasks.content-feed-refresh,
# every 21600 s); this script exists so the first run after an enablement migration can be
# watched in-session instead of waiting up to six hours. The orchestrator is idempotent at
# the seeding step (ON CONFLICT DO NOTHING on (site_id, name)), so a run that overlaps the
# trigger's costs one duplicate fetch, not duplicate sources.
#
# Preconditions (read them, do not assume them):
#   - the site's classification spec carries content_features.news_feed.recommended=true;
#   - the site has at least one active content_sources row (checked below — the
#     orchestrator's seeder returns early on any existing active source and SKIPS rss
#     outright, so "enabled" without rows produces nothing: see LANDMINES, migration 746);
#   - no chassis pod (re)started in the last ~300 s — a dispatch in that window is dropped
#     silently (CLAUDE.md): kubectl -n ai-persona-system get pods -l app=agent-chassis
#
# Payload shape is 082_submit_domain_unified.sh's: the chassis resolves the agent_type's
# default_config, so no workflow is inlined here (080_test_content_feed_orchestrator.sh
# inlines one and uses the racing `kubectl run -i` form — do not copy it).
set -euo pipefail

DOMAIN="${1:?usage: $0 <domain>   e.g. $0 advertise.co.uk}"

REPO_ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel 2>/dev/null || true)"
if [ -z "$REPO_ROOT" ] || [ ! -f "$REPO_ROOT/scripts/kafka-publish-lib.sh" ]; then
  echo "ERROR: scripts/kafka-publish-lib.sh not found (repo root: ${REPO_ROOT:-<not in a git repo>})." >&2
  exit 1
fi
. "$REPO_ROOT/scripts/kafka-publish-lib.sh"

# Resolve the site and assert the two preconditions that are checkable from here.
# One round trip, pipe-separated, so a missing domain is distinguishable from a bad query.
PRE=$(kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -tAF'|' -c "
    SELECT s.id,
           COALESCE((SELECT (ss.data->'content_features'->'news_feed'->>'recommended')::boolean
                       FROM site_specs ss
                      WHERE ss.site_id = s.id AND ss.aspect = 'classification' AND ss.is_current), false),
           (SELECT count(*) FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active)
      FROM sites s WHERE lower(s.domain) = lower('${DOMAIN}');" 2>/dev/null | tr -d '\r')

[ -n "$PRE" ] || { echo "ERROR: no sites row for domain '${DOMAIN}'." >&2; exit 1; }
IFS='|' read -r SITE_ID RECOMMENDED N_SOURCES <<<"$PRE"

if [ "$RECOMMENDED" != "t" ]; then
  echo "ERROR: ${DOMAIN}'s current classification spec does not carry content_features.news_feed.recommended=true." >&2
  echo "       The trigger predicate would not select it either. Author the spec key AND the sources in one transaction." >&2
  exit 1
fi
if [ "${N_SOURCES:-0}" -eq 0 ]; then
  echo "ERROR: ${DOMAIN} has 0 active content_sources rows — the run would fetch nothing." >&2
  echo "       seed_content_sources_action skips rss and returns early on any existing active source; create the rows by hand." >&2
  exit 1
fi
echo "PRECONDITIONS OK  domain=${DOMAIN} site_id=${SITE_ID} recommended=true active_sources=${N_SOURCES}"

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
  echo "DISPATCH DID NOT GO OUT — nothing queued for ${DOMAIN}. Re-run: $0 ${DOMAIN}" >&2
  exit "$PUBLISH_RC"
fi

echo "SAVE: CORRELATION_ID=${CORRELATION_ID}  ORCHESTRATION_ID=${ORCHESTRATION_ID}"

LANDING_RC=0
kafka_verify_landing "$CORRELATION_ID" 120 || LANDING_RC=$?
echo "Find it later by:  SELECT status, current_step, error FROM orchestration_states WHERE correlation_id='${CORRELATION_ID}'::uuid;"
exit "$LANDING_RC"
