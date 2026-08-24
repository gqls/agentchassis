#!/usr/bin/env bash
# TRIGGER_claims_audit.sh <domain> — dispatch ONE claims-auditor pass at an EXISTING site.
#
# bugs_open/380 lane. The auditor audits every deployed page against the site's evidence
# register; with no register (or no facts) it runs COLD (migration 597): every assertion of
# fact about the business is unsupported, and first-person practice claims are the class it
# reports at severity high. Findings file ONE claims_unverified work item (HITL-terminal,
# item_key claims_llm*); every run writes a doc_notes receipt ('pipeline'/'claims-audit').
#
# Publishes via scripts/kafka-publish-lib.sh (OPP-009) — the receipt is ASSERTED, because
# `kubectl run -i --rm … kcat -P` drops ~4 publishes in 5 at exit 0 (LANDMINES).
# REFUSES a domain with no sites row: a hand audit must never create a site
# (ensure_site_record upserts by domain — handing it an unknown domain would mint a row).
#
# No dispatch within ~300s of a chassis pod restart (CLAUDE.md) — the spawn is dropped.
set -euo pipefail

DOMAIN="${1:?usage: TRIGGER_claims_audit.sh <domain>}"
REPO_ROOT="$(cd "$(dirname "$0")/../../../.." && pwd)"
PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At"

SITE_ID=$($PSQL -c "SELECT id FROM sites WHERE domain='${DOMAIN}'" | tr -d '[:space:]')
if [ -z "$SITE_ID" ]; then
  echo "REFUSED: no sites row for '${DOMAIN}' — a hand audit must never create a site." >&2
  exit 1
fi

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)

PAYLOAD=$(printf '{"action":"orchestrate","config":{"agent_type":"claims-auditor"},"input_data":{"domain":"%s","site_id":"%s"}}' "$DOMAIN" "$SITE_ID")

# shellcheck source=/dev/null
. "$REPO_ROOT/scripts/kafka-publish-lib.sh"

kafka_publish_checked \
  --topic system.agent.generic.requests \
  --payload "$PAYLOAD" \
  --correlation "$CORRELATION_ID" \
  --header "correlation_id=$CORRELATION_ID" \
  --header "request_id=$REQUEST_ID" \
  --header "message_id=$MESSAGE_ID" \
  --header "orchestration_id=$ORCHESTRATION_ID" \
  --header "orchestration_name=claims-audit-${DOMAIN}-$(date +%Y%m%d-%H%M%S)" \
  --header "step_name=start" \
  --header "client_id=demo_client" \
  --header "message_type=request" \
  --header "action=orchestrate" \
  --header "from_agent_type=user" \
  --header "from_agent_id=cli" \
  --header "responses_topic=system.agent.generic.responses"

echo "SAVE: CORRELATION_ID=${CORRELATION_ID}  SITE_ID=${SITE_ID}"
cat <<VERIFY

=== VERIFY (the publish receipt above is only the PUBLISH; a run takes minutes and
    queues behind the fleet — budget ~30, not ~2) ===

The run, by payload (never by the printed id alone):
  SELECT status, current_step, left(COALESCE(error,''),120) FROM orchestration_states
   WHERE owner_agent_type='claims-auditor'
     AND collected_data->'input_data'->>'domain'='${DOMAIN}'
   ORDER BY created_at DESC LIMIT 3;

Did the LLM actually LOOK (a [] with the offending sentence truncated out is not a verdict):
  SELECT created_at, input_tokens, output_tokens,
         position('NO VERIFIED FACT' in prompt_rendered) > 0 AS cold_arm_rendered,
         left(response_text,200)
    FROM llm_call_log WHERE agent_type='claims-auditor' ORDER BY created_at DESC LIMIT 1;

Findings (only findings file an item; a clean pass files none):
  SELECT item_key, status, left(summary,80) FROM site_work_items
   WHERE site_id='${SITE_ID}' AND item_type='claims_unverified' ORDER BY created_at DESC;

The per-run receipt (one row per run, clean or not — a MISSING row means it did not run):
  SELECT created_at, categories, left(body,120) FROM doc_notes
   WHERE subject_type='pipeline' AND subject_key='claims-audit' AND site_id='${SITE_ID}'
   ORDER BY created_at DESC LIMIT 3;
VERIFY
