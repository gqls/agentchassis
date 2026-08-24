#!/usr/bin/env bash
# Full site rerender for noted.co.uk.
# DOESN'T REWRITE CONTENT — re-assembles pages from their existing components.
#
# Why this:
#   2026-08-12: the CTA destinations were set by hand in page_components.content_data
#   (CTA_2026-08-12_noted_cta_destinations.sql) because the framework could not
#   resolve them — this product's CTA target is a different origin
#   (app.noted.co.uk), not an internal content hub, so the resolver found nothing
#   and both templates, which gate on `{{if and .cta_text .cta_url}}`, rendered
#   ZERO anchors on every hero and call-to-action.
#   content_data is now correct but rendered_html is not, so the pages need
#   re-assembling for the buttons to exist.
#
#   A RERENDER is the right tool and a REGENERATION is not: a rerender MERGES
#   content_data (the hand-set *_cta_url keys survive), whereas a regeneration
#   REPLACES it and would silently drop them. See memory `bugfix 238`.
#
# refresh_site_components=false — only the page bodies changed. Header/footer/head
#   are untouched and there is no reason to re-render them.
#
# Expected outcome:
#   - 5 page_rerender work items created with status='triaged' (one per deployed
#     page), claimed by build-dispatch-loop.
#   - Commits "Rerender: <page>.html" to gqls/vm-sites (NOT gqls/sites — this site
#     is a VM-target site; see sites.github_repo='vm-sites', which is load-bearing
#     safety: the default repo would b2-sync-with-delete over the live legacy app).
#   - Within 5 min of each commit, sitesync pulls it onto the box.
#
# ⚠ TIMING: these items are filed NOW, so they sort to the BACK of the estate-wide
#   dispatch queue (find_dispatchable_site orders by wi.created_at ASC across ALL
#   sites). Expect hours, not minutes, unless the dispatch loop is pinned at this
#   site. That is fine here: the framework build is NOT public — noted.co.uk still
#   serves the legacy app from B2 — so nothing user-facing is waiting on this.
#   See LANDMINES: "A correctly-filed `triaged` build item can sit for HOURS".

set -euo pipefail

AGENT_TYPE="rerender-pages"
SITE_ID="b50a8da1-25bd-4a6d-88f2-4d4c93f6acc8"
DOMAIN="noted.co.uk"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "=== Triggering: ${AGENT_TYPE} ==="
echo "  Site:           ${DOMAIN} (${SITE_ID})"
echo "  Correlation:    ${CORRELATION_ID}"
echo "  Orchestration:  ${ORCHESTRATION_ID}"
echo "  Request:        ${REQUEST_ID}"
echo "  Time:           ${TIMESTAMP}"
echo ""
echo ""

# Payload lifted from the heredoc; delimiter left UNQUOTED so ${VAR} still expands.
PAYLOAD_327=$(cat <<JSON
{"action":"orchestrate","config":{"agent_type":"rerender-pages"},"input_data":{"site_id":"$SITE_ID","domain":"$DOMAIN","refresh_site_components":false}}
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
  --header "client_id=$CLIENT_ID" \
  --header "action=orchestrate" \
  --header "sender_agent_type=cli" \
  --header "sender_agent_id=cli-user" \
  --header "responses_topic=system.agent.generic.responses" \
  --header "timestamp=$TIMESTAMP" || PUBLISH_RC=$?

if [ "$PUBLISH_RC" -ne 0 ]; then
  echo "NOT DISPATCHED — no site rerender will run (bugs_open/327)." >&2
  exit "$PUBLISH_RC"
fi

echo "  SAVE THIS: CORRELATION_ID=${CORRELATION_ID}"

cat <<'NOTES'

=== Verify the DISPATCH landed (kcat exit 0 is NOT delivery) ===

  SELECT status, count(*) FROM site_work_items
  WHERE site_id='b50a8da1-25bd-4a6d-88f2-4d4c93f6acc8' AND item_type='page_rerender'
    AND created_at > now() - interval '10 minutes'
  GROUP BY status;

Expect 5 rows in 'triaged' within ~30s. Nothing there after a few minutes means the
message was dropped, not that it is queued — re-check before re-firing.

=== Verify the OUTCOME at the artefact, not the status ===

  -- buttons must actually exist in the rendered HTML
  SELECT p.name, pc.slot_name,
         (length(pc.rendered_html)-length(replace(pc.rendered_html,'<a ','')))/3 AS anchors
  FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
  WHERE s.domain='noted.co.uk' AND pc.slot_name IN ('hero','call-to-action')
  ORDER BY p.name, pc.position;

Expect >0 anchors everywhere EXCEPT migrate's two primaries, which are deliberately
destination-less until /legacy is built (they keep their secondary button).

Then on the box, after sitesync's next 5-minute tick:
  ssh -i ~/.ssh/webdesign_box_ed25519 root@webdesign.vs.mythic-beasts.com \
    'grep -c "app.noted.co.uk" /var/www/noted.co.uk/index.html'
NOTES
