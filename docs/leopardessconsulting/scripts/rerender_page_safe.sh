#!/bin/bash
# ============================================================================
# rerender_page_safe.sh — re-render ONE page from stored content_data, with a
# publish receipt.
#
# Same payload as rerender_pages.sh, and the same load-bearing detail:
# page-rerender only REGENERATES section HTML from content_data when
# input_data.spec.reason is 'section_data_resolved' (or 'image_landed').
# Without it, check_rerender_mode re-assembles the STORED (old) section HTML.
# rerender_page_sections also requires spec.page_name — the page's `name`, not
# its url. Both are set below.
#
# WHY THIS EXISTS ALONGSIDE rerender_pages.sh. That script publishes with
# `echo "$PAYLOAD" | kubectl -n kafka run -i --rm … kcat -P` and sends both
# streams to /dev/null. `kubectl run -i` attaches stdin asynchronously, so if the
# container reaches kcat before stdin is wired it sees EOF and publishes nothing
# at exit 0 — measured 2026-07-26 at four of five publishes lost. With output
# discarded there is then NO receipt either way, and a silently dropped render
# looks exactly like queue latency. Here the payload travels in the container
# COMMAND and the publisher prints PUBLISH_OK.
#
# `--command` is load-bearing: the image's ENTRYPOINT is kcat, so without it the
# `sh -c …` arrives as arguments TO kcat and nothing is published.
#
# BEFORE RUNNING: no orchestration dispatch within ~300s of a chassis pod
# (re)start, or the spawn is silently dropped.
#   kubectl -n ai-persona-system get pods -l app=agent-chassis \
#     -o custom-columns='NAME:.metadata.name,START:.status.startTime'
#
# Usage: ./rerender_page_safe.sh <site_id> <domain> <page_name>
# ============================================================================
set -euo pipefail

SITE="${1:?site_id}"
DOMAIN="${2:?domain}"
PAGE="${3:?page_name (pages.name, not the url)}"

CID=$(cat /proc/sys/kernel/random/uuid)
OID=$(cat /proc/sys/kernel/random/uuid)
RID=$(cat /proc/sys/kernel/random/uuid)
MID=$(cat /proc/sys/kernel/random/uuid)
BROKER="personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"

PAYLOAD_B64=$(python3 - "$SITE" "$DOMAIN" "$PAGE" <<'PY'
import base64, json, sys
site, domain, page = sys.argv[1:4]
msg = {
    "action": "orchestrate",
    "config": {"agent_type": "page-rerender"},
    "input_data": {
        "site_id": site,
        "domain": domain,
        "spec": {"reason": "section_data_resolved", "page_name": page},
    },
}
line = json.dumps(msg, separators=(",", ":"))
assert "\n" not in line
sys.stdout.write(base64.b64encode(line.encode()).decode())
PY
)

echo "page:        $PAGE"
echo "correlation: $CID"
echo "publishing to system.agent.generic.requests ..."

kubectl -n kafka run "kcat-rr-$(date +%s)-$RANDOM" \
  --rm --restart=Never --image=edenhill/kcat:1.7.1 --attach=true --quiet \
  --command -- sh -c "echo '$PAYLOAD_B64' | base64 -d | kcat -P \
    -b $BROKER \
    -t system.agent.generic.requests \
    -H correlation_id=$CID \
    -H orchestration_id=$OID \
    -H request_id=$RID \
    -H message_id=$MID \
    -H orchestration_name=rr-$PAGE \
    -H step_name=start \
    -H client_id=demo_client \
    -H message_type=request \
    -H action=orchestrate \
    -H from_agent_type=user \
    -H from_agent_id=cli \
    -H responses_topic=system.agent.generic.responses && echo PUBLISH_OK"

echo ""
echo "No PUBLISH_OK above means NOTHING was published — re-run now."
echo "Follow it (a missing row is usually latency, not a drop):"
echo "  SELECT current_step, status, error FROM orchestration_states"
echo "   WHERE correlation_id = '$CID';"
echo "Then check it did NOT escalate to the writer:"
echo "  SELECT item_type, status, spec FROM site_work_items"
echo "   WHERE site_id='$SITE' AND item_type='needs_page'"
echo "     AND created_at > now() - interval '15 minutes';"
