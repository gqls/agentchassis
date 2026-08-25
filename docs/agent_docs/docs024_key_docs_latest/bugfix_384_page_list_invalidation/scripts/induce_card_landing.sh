#!/bin/bash
# ============================================================================
# induce_card_landing.sh — bugs_open/384's post-roll acceptance test.
#
# Dispatches asset-deployer in content_card mode for ONE page that already has
# a card. Re-deriving is idempotent in effect (derive_card_asset re-crops the
# page's current hero and upserts the same asset_key), so this is a real card
# landing without inventing an asset.
#
# WHAT IT PROVES, and the disconfirming result to insist on: within the run,
# EXACTLY N page_rerender rows must appear carrying spec.cause='card_landed:<page>',
# where N is the site's consumer-page count under the SHIPPED lookup (printed
# below). Zero rows = the seam is not wired or not rolled. N != consumers, or a
# row on a rebuild_policy='owned' page = the lookup is wrong. A listing that
# merely looks right proves nothing: on dartsonline both listings already show
# 12/12 after the 2026-08-24 hand repair, so the ROWS are the measurement.
#
# BEFORE RUNNING: no orchestration dispatch within ~300s of a chassis pod
# (re)start, or the spawn is silently dropped:
#   kubectl -n ai-persona-system get pods -l app=agent-chassis \
#     -o custom-columns='NAME:.metadata.name,START:.status.startTime'
#
# Usage: ./induce_card_landing.sh <domain> <page_name>
# ============================================================================
set -euo pipefail
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
. "$REPO_ROOT/scripts/kafka-publish-lib.sh"

DOMAIN="${1:?domain}"
PAGE="${2:?page_name (pages.name, not the url)}"
PSQL=(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -X -q -At)

read -r SITE PAGE_ID <<<"$("${PSQL[@]}" -c "SELECT s.id, p.id FROM pages p JOIN sites s ON s.id=p.site_id WHERE s.domain='$DOMAIN' AND p.name='$PAGE';" | tr '|' ' ')"
[ -n "${SITE:-}" ] && [ -n "${PAGE_ID:-}" ] || { echo "no such page $PAGE on $DOMAIN" >&2; exit 2; }

# N, computed by the SHIPPED predicate (queryresolve/consumers.go pageListConsumerSQL).
N=$("${PSQL[@]}" -c "
SELECT count(DISTINCT p.id) FROM page_components pc
  JOIN pages p ON p.id=pc.page_id JOIN content_components cc ON cc.id=pc.component_id
 WHERE p.site_id='$SITE' AND p.status IN ('active','deployed') AND pc.build_status<>'removed'
   AND COALESCE(p.rebuild_policy,'generic')<>'owned'
   AND cc.input_schema IS NOT NULL AND cc.input_schema::text LIKE '%query.%'
   AND cc.html_template ~* '\.image\y'
   AND EXISTS (SELECT 1 FROM jsonb_each(coalesce(cc.input_schema->'fields','{}'::jsonb)) f
                WHERE lower(f.value->>'source') ~ '^query\.(blog_posts|pages_where_type|pages_under_section)');")
echo "site=$SITE page=$PAGE page_id=$PAGE_ID  EXPECTED consumer items N=$N"

# Pre-count, so the assertion is about THIS run and not about history.
BEFORE=$("${PSQL[@]}" -c "SELECT count(*) FROM site_work_items WHERE site_id='$SITE' AND item_type='page_rerender' AND spec->>'cause'='card_landed:$PAGE';")
echo "existing items with this cause (before): $BEFORE"

CID=$(cat /proc/sys/kernel/random/uuid)
PAYLOAD=$(python3 - "$SITE" "$DOMAIN" "$PAGE" "$PAGE_ID" <<'PY'
import json, sys
site, domain, page, page_id = sys.argv[1:5]
# spec shape = discovery_checks.ContentImageSpecJSON (the ONE spelling derive_card_asset reads)
msg = {"action": "orchestrate", "config": {"agent_type": "asset-deployer"},
       "input_data": {"site_id": site, "domain": domain,
                      "spec": {"mode": "content_card", "check": "induced_384_acceptance",
                               "entity_type": "page", "entity_id": page_id,
                               "page_name": page, "purpose": "card"}}}
line = json.dumps(msg, separators=(",", ":")); assert "\n" not in line
sys.stdout.write(line)
PY
)
kafka_publish_checked --topic system.agent.generic.requests --payload "$PAYLOAD" --correlation "$CID" \
  --header "orchestration_id=$(cat /proc/sys/kernel/random/uuid)" \
  --header "request_id=$(cat /proc/sys/kernel/random/uuid)" \
  --header "message_id=$(cat /proc/sys/kernel/random/uuid)" \
  --header "orchestration_name=induce384-${PAGE}-$(date +%H%M%S)" \
  --header "step_name=start" --header "client_id=demo_client" \
  --header "message_type=request" --header "action=orchestrate" \
  --header "from_agent_type=user" --header "from_agent_id=cli" \
  --header "responses_topic=system.agent.generic.responses"

echo
echo "correlation: $CID"
echo "follow the run:   SELECT current_step, status, error FROM orchestration_states WHERE correlation_id='$CID';"
echo "THE ASSERTION (expect exactly $N new rows):"
cat <<SQL
  SELECT p.name AS consumer_page, w.status, w.spec->>'reason' AS reason, w.item_key, w.created_at
    FROM site_work_items w JOIN pages p ON p.id=w.page_id
   WHERE w.site_id='$SITE' AND w.item_type='page_rerender'
     AND w.spec->>'cause'='card_landed:$PAGE' ORDER BY w.created_at DESC;
SQL
echo "and NO row may sit on a rebuild_policy='owned' page."
