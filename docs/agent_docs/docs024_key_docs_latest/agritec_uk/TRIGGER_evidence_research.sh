#!/bin/bash
# ============================================================================
# TRIGGER_evidence_research.sh — dispatch evidence-researcher at one question
# ============================================================================
# agritec.uk Phase 2 (PLAN section 4). Owner ruling D4: source everything before
# any page is written. This fires ONE research question; run it once per data
# domain and read the result before firing the next.
#
# Envelope copied from 082_submit_domain_unified.sh's kcat block (same topic,
# headers, responses_topic). The agent's own contract, read from the live
# agent_definitions row 2026-08-22:
#   search_web          query_from = input_data.research_query, num_results 10
#   prepare_urls        max_scrapes 4; prefers .gov/.org/.edu (so GOV.UK matches)
#   extract_claims      up to 10 candidates, each needing a VERBATIM quote
#   verify_and_register re-fetches the url and REJECTS the claim unless the quote
#                       still appears in it. A paraphrase fails. This is the
#                       whole point: it is why a registered fact is worth more
#                       than a cited one.
#
# ⚠ kcat -P CAN EXIT 0 HAVING SENT NOTHING (LANDMINES). Never treat the exit code
#   as proof of dispatch. This script therefore prints the verify query, and you
#   MUST run it. A missing orchestration row within ~30 min means it did not land;
#   a missing row at 5 min means nothing at all, because publish->run start was
#   measured at 29 minutes under normal fleet load. Do NOT retry on a 5-minute
#   silence — that costs a duplicate round.
#
# Usage:  ./TRIGGER_evidence_research.sh "<research question>"
# ============================================================================
set -euo pipefail

QUERY="${1:?Usage: $0 \"<research question>\"}"
DOMAIN="${DOMAIN:-agritec.uk}"

# JSON-escape the query (no jq dependency, same approach as 082's --mission-file).
ESCAPED=$(printf '%s' "$QUERY" | python3 -c 'import json,sys; print(json.dumps(sys.stdin.read())[1:-1])')

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)

echo "========================================="
echo "evidence-researcher: ${DOMAIN}"
echo "  Question:      ${QUERY}"
echo "  Correlation:   ${CORRELATION_ID}"
echo "========================================="
echo "SAVE: CORRELATION_ID=${CORRELATION_ID}"
echo ""

kubectl -n kafka run -i --rm kcat-evres-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H orchestration_name=evres-${DOMAIN}-$(date +%Y%m%d-%H%M%S) \
  -H step_name=start \
  -H client_id=demo_client \
  -H message_type=request \
  -H action=orchestrate \
  -H from_agent_type=user \
  -H from_agent_id=cli \
  -H responses_topic=system.agent.generic.responses <<JSON
{"action":"orchestrate","config":{"agent_type":"evidence-researcher"},"input_data":{"domain":"${DOMAIN}","research_query":"${ESCAPED}"}}
JSON

cat <<VERIFY

=== VERIFY — the exit code above proves nothing (kcat landmine) ===

Did it land? (allow ~30 min; a 5-minute silence means nothing)
  SELECT status, current_step, left(COALESCE(error,''),100) AS error
    FROM orchestration_states WHERE correlation_id='${CORRELATION_ID}'::uuid;

Or find it by payload rather than by id:
  SELECT current_step, status, created_at FROM orchestration_states
   WHERE owner_agent_type='evidence-researcher'
     AND collected_data->'input_data'->>'domain'='${DOMAIN}'
   ORDER BY created_at DESC LIMIT 5;

What actually got REGISTERED (the only thing that counts):
  SELECT jsonb_array_length(data->'facts') AS facts
    FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='${DOMAIN}' AND ss.aspect='evidence_base' AND ss.is_current;

  SELECT f->>'id', f->>'claim', f->'source'->'citation'->>'url'
    FROM site_specs ss JOIN sites s ON s.id=ss.site_id,
         LATERAL jsonb_array_elements(ss.data->'facts') f
   WHERE s.domain='${DOMAIN}' AND ss.aspect='evidence_base' AND ss.is_current;

⚠ ZERO facts is a LEGITIMATE outcome, not a failure: it means no source page
  literally contained a sentence stating the fact. That is the machine check
  doing its job. Read the run before concluding the dispatch broke.

⚠ A registered fact is not a TRUE fact (bugs_open/161). The check proves the
  quote is really on that page — not that we read it correctly, nor that the
  page is right. Read what landed before letting a writer use it.
VERIFY
