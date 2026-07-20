#!/usr/bin/env bash
# 049_TRIGGER_chrome_refresh.sh — re-render site chrome, then reassemble+deploy every page.
#
# WHY (bugs_open/049): the chrome renderer's hardcoded {/privacy.html, /terms.html}
# legal slice was fixed 2026-06-10 (0681e1542), but chrome is only re-rendered when
# something explicitly asks. finetuning.uk / ai-agent-orchestration.com /
# gaswholesalers.com carry chrome from 28 Apr / 21 May, so every page of all three
# still ships two 404 legal links (204 broken anchors of the 312 measured fleet-wide).
#
# These sites HAVE been page-rerendered since (07-10 / 07-17) and it changed nothing,
# because a rerender REASSEMBLES pages and injects the STORED chrome. The one thing
# those runs lacked is the flag below.
#
#   refresh_site_components=true  -> check_refresh_components -> render_site_components
#                                    (slots header/footer/head, force_rerender:true)
#
# This does NOT rewrite content: rerender-pages reassembles stored page_components
# HTML. It is not a re-plan and not a rebuild, so it is outside the bugs_open/001 /
# 038 clobber class.
#
# ROLLBACK: bak_site_components_chrome_20260720 holds all 9 pre-change rows
# (site_id, slot_name, rendered_html, domain, snapshot_at).
#
# USAGE:  ./049_TRIGGER_chrome_refresh.sh <domain>
# Sites are dispatched ONE AT A TIME on purpose — each produces ~26-37 git commits to
# gqls/sites, each firing a B2 sync + Cloudflare purge. Three at once floods the queue.

set -euo pipefail

DOMAIN="${1:-}"
if [[ -z "$DOMAIN" ]]; then
  echo "usage: $0 <domain>" >&2; exit 2
fi

case "$DOMAIN" in
  finetuning.uk)              SITE_ID="1368e337-dd1d-4799-bbb3-8221a1b79bcc" ;;
  ai-agent-orchestration.com) SITE_ID="2a8ebf9c-20a2-4c39-b191-840b012371da" ;;
  gaswholesalers.com)         SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002" ;;
  *) echo "unknown domain: $DOMAIN (add its site_id deliberately, do not guess)" >&2; exit 2 ;;
esac

# One domain -> one site_id, resolved by case. The script this replaces
# (scripts/initial_messages/001_assemble_all_pages_rerender/081_*.sh) declared
# AGENT_TYPE/SITE_ID/DOMAIN twice in sequence, so the second block silently won and
# it targeted a different site from the one in its own filename. Do not reintroduce that.

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "=== rerender-pages + chrome refresh ==="
echo "  domain:        $DOMAIN"
echo "  site_id:       $SITE_ID"
echo "  correlation:   $CORRELATION_ID"
echo "  time:          $TIMESTAMP"
echo ""

PAYLOAD="{\"action\":\"orchestrate\",\"config\":{\"agent_type\":\"rerender-pages\"},\"input_data\":{\"site_id\":\"$SITE_ID\",\"domain\":\"$DOMAIN\",\"refresh_site_components\":true}}"

# The payload is echoed INSIDE the pod, not piped to `kubectl run -i` on stdin.
#
# WHY (learned the hard way, 2026-07-20): the inherited pattern
#   kubectl run -i --rm ... -- kcat -P -c 1 ... <<JSON ... JSON
# is a race. `kubectl run -i` attaches stdin only after the container starts; if
# kcat reaches EOF first it produces NOTHING, exits 0, and the pod is deleted —
# so the script prints its full "=== Monitor ===" success banner having sent no
# message. That is exactly what happened on the first attempt here: no
# orchestration row, no chassis log line for the correlation, and the correlation
# absent from the topic's last 20 messages, while the chassis logged "No activity
# for 5 minutes". Read as queue latency it looks identical to a slow dispatch —
# which is the trap in the council queue-latency note, inverted: there, absence of
# a row meant QUEUED; here it meant NEVER SENT. Distinguish them by reading the
# TOPIC, not the orchestration table.
#
# `sh -c 'echo ... | kcat'` removes the stdin dependency entirely.
# `-i` is kept only because `--rm` requires an attached container ("--rm should
# only be used for attached containers"). Nothing reads stdin — the payload is
# echoed inside the pod — so the race described above cannot apply.
kubectl -n kafka run -i --rm --restart=Never "kcat-chrome-refresh-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --command -- \
  sh -c "echo '$PAYLOAD' | kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H message_type=request \
  -H client_id=demo_client \
  -H action=orchestrate \
  -H sender_agent_type=cli \
  -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TIMESTAMP"

echo ""
echo "CORRELATION_ID=$CORRELATION_ID"
echo ""
echo "=== FIRST: prove the message was actually PRODUCED (never skip this) ==="
echo "kubectl -n kafka run -i --rm kcat-verify-\$(date +%s) --image=edenhill/kcat:1.7.1 --restart=Never -- \\"
echo "  kcat -C -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \\"
echo "  -t system.agent.generic.requests -o -20 -e -f '%h\\n' | grep $CORRELATION_ID"
echo "   no match = it was never sent; do NOT sit waiting for an orchestration row."
echo ""
echo "=== verify, in order ==="
echo "1. chrome actually re-rendered (updated_at must move off Apr/May):"
echo "   SELECT slot_name, updated_at, rendered_html LIKE '%/privacy.html%' AS still_phantom"
echo "     FROM site_components WHERE site_id='$SITE_ID';"
echo "2. page_rerender items created and draining:"
echo "   SELECT status, count(*) FROM site_work_items WHERE site_id='$SITE_ID'"
echo "     AND item_type='page_rerender' AND created_at > now()-interval '30 min' GROUP BY 1;"
echo "3. the LIVE artefact (this is the only one that counts) — poll per RUNBOOK R13,"
echo "   status + good marker + bad marker absent, never just 'bad marker gone':"
echo "   curl -s https://$DOMAIN/index.html | grep -c 'href=\"/privacy.html\"'   # want 0"
echo "4. re-run the full audit: scripts/live_link_audit.sh"
