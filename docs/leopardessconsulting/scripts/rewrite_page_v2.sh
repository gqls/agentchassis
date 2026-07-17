#!/bin/bash
# rewrite_page_v2.sh — framework-path plain-voice v2 rewrite of one page.
# Fires page-build-handler with the OWNER-APPROVED rewrite prompt (2026-07-17)
# in spec.suggestion → page-content-writer receives it as Rewrite Guidance,
# alongside the v2 voice spec. Output passes validate_page_content including
# the claims gate (banned claims = blocker, unregistered numbers = error).
# Usage: ./rewrite_page_v2.sh <page_name> [<page_name>...]
set -euo pipefail
SITE=4851f6fc-71cf-4160-a270-e03d6d3e0732
DOMAIN=leopardessconsulting.co.uk

read -r -d '' PROMPT <<'EOF' || true
Rewrite this page's copy in the site's plain register. This is a rewrite for readability: keep every fact, number, and claim exactly as given in the specs and section briefs. Do not add new facts, numbers, clients, or capabilities. If you do not have a number for something, describe it without one. Never round a number up or dramatise it.
How to write:
- One idea per sentence. If a sentence carries two ideas, split it. Keep most sentences under 20 words.
- Short paragraphs, one to three sentences.
- Use contractions: it's, we'd, you're, isn't.
- Use everyday words: use, not utilise; help, not facilitate.
- Active voice. Talk to the reader as "you"; call us "we". Start a sentence with And or But when it's natural.
- No em-dashes. Use a full stop or a comma instead.
- No hype words and no marketing register: never unlock, leverage, seamless, transform, cutting-edge, game-changing.
- No literary flourishes. If a phrase sounds quotable, simplify it until it just sounds clear. Never end a section with a summing-up line.
- Don't write lists of three by reflex. Two examples are fine.
- No rhetorical questions. No forced friendliness like "You know what?" or "honestly". Friendly here means calm and easy to read.
- If we haven't done a thing for a client, say so plainly: "We haven't done this for a client yet. The nearest thing we've built is X."
The test: read it aloud. It should sound like a person explaining their work to a smart friend, plainly, without performing. If a sentence sounds impressive, rewrite it until it sounds clear instead.
EOF

for PN in "$@"; do
  CID=$(cat /proc/sys/kernel/random/uuid)
  python3 - "$SITE" "$DOMAIN" "$PN" "$PROMPT" <<'PY' > /tmp/rw_payload.json
import json, sys
site, domain, pn, prompt = sys.argv[1:5]
print(json.dumps({"action":"orchestrate","config":{"agent_type":"page-build-handler"},
  "input_data":{"site_id":site,"domain":domain,"item_type":"needs_page",
    "spec":{"page_name":pn,"reason":"plain_voice_v2_rewrite","suggestion":prompt}}},
  separators=(",",":")))
PY
  cat /tmp/rw_payload.json | kubectl -n kafka run -i --rm "kcat-v2-${PN:0:8}-$(date +%s%N | tail -c 5)" --image=edenhill/kcat:1.7.1 --restart=Never -- \
    kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 -t system.agent.generic.requests \
    -H correlation_id=$CID -H orchestration_id=$(cat /proc/sys/kernel/random/uuid) -H request_id=$(cat /proc/sys/kernel/random/uuid) \
    -H message_id=$(cat /proc/sys/kernel/random/uuid) -H orchestration_name=v2rw-${PN:0:12}-$(date +%H%M%S) -H step_name=start \
    -H client_id=demo_client -H message_type=request -H action=orchestrate -H from_agent_type=user -H from_agent_id=cli \
    -H responses_topic=system.agent.generic.responses 2>&1 | grep -q "deleted from kafka" && echo "$PN fired ($CID)" || echo "$PN SEND FAILED"
done
