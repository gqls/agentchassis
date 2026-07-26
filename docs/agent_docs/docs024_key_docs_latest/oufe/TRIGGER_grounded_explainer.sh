#!/usr/bin/env bash
# ============================================================================
# TRIGGER_grounded_explainer.sh — fire the high-attention content lane
# ============================================================================
# GENERIC: any site, any topic. Seeded by
# docs/agent_docs/sql_for_agents/224_grounded_explainer_agent.sql
#
#   search → prepare_urls → scrape → extract atomic claims with verbatim quotes
#   → verify_and_register_citations (RE-FETCHES every url, discards any quote
#   that is not literally there) → compose from survivors only → independent
#   grounding audit → needs_human_review.
#
# It cannot publish. That is the point, not an omission — read 224's header.
#
# Usage:
#   ./TRIGGER_grounded_explainer.sh <domain> "<topic>" "<research query>" ["<audience>"]
#
# Example:
#   ./TRIGGER_grounded_explainer.sh oufe.com \
#     "How a restructuring plan can bind a class that voted against it" \
#     "Companies Act 2006 Part 26A restructuring plan cross-class cram down conditions" \
#     "a professional new to restructuring who wants the mechanism"
#
# TUNING THAT MATTERS: prefer_domains lives in the agent definition and points at
# UK primary legal publishers (legislation.gov.uk, judiciary.uk, bailii). For a
# different field, change it there rather than here — pointing the acquisition
# step at the primary instrument is most of what makes the output trustworthy.
#
# WHAT SUCCESS LOOKS LIKE
#   A work item of type grounded_draft_review carrying the draft AND the audit.
#   Read the audit FIRST: `ungrounded` should be empty. If it is not, the draft
#   asserted something it could not support, and the honest move is to cut those
#   sentences, not to go and find a source that agrees with them.
#
#   An empty registration (nothing survived verification) is a legitimate
#   outcome, not a failure. It means the open web did not yield a quotable
#   primary source for the question as asked — usually a sign to sharpen the
#   research query toward the instrument itself.
# ============================================================================
set -euo pipefail

DOMAIN="${1:?Usage: $0 <domain> \"<topic>\" \"<research query>\" [\"<audience>\"]}"
TOPIC="${2:?missing topic}"
QUERY="${3:?missing research query}"
AUDIENCE="${4:-}"

PSQL=(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc)

SITE_ID=$("${PSQL[@]}" "SELECT id FROM sites WHERE domain='${DOMAIN}';" | tr -d ' ')
[ -n "$SITE_ID" ] || { echo "no site row for ${DOMAIN}" >&2; exit 2; }

esc() { printf '%s' "$1" | sed -e 's/\\/\\\\/g' -e 's/"/\\"/g' | tr '\n\t' '  '; }

INPUT="{\"domain\":\"$DOMAIN\",\"site_id\":\"$SITE_ID\",\"topic\":\"$(esc "$TOPIC")\",\"research_query\":\"$(esc "$QUERY")\""
[ -n "$AUDIENCE" ] && INPUT="${INPUT},\"audience\":\"$(esc "$AUDIENCE")\""
INPUT="${INPUT}}"

if command -v python3 >/dev/null; then
  printf '%s' "$INPUT" | python3 -c 'import json,sys; json.load(sys.stdin)' \
    || { echo "input_data is not valid JSON — aborting" >&2; exit 3; }
fi

CORR=$(cat /proc/sys/kernel/random/uuid)
ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid)
MSG=$(cat /proc/sys/kernel/random/uuid)

echo "corr=$CORR site=$DOMAIN"
echo "topic=$TOPIC"

kubectl -n kafka run -i --rm "kcat-ge-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -c 1 \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORR \
  -H orchestration_id=$ORCH \
  -H request_id=$REQ \
  -H message_id=$MSG \
  -H message_type=request \
  -H client_id=demo_client \
  -H action=orchestrate \
  -H from_agent_type=cli \
  -H from_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses <<JSON
{"action":"orchestrate","config":{"agent_type":"grounded-explainer"},"input_data":$INPUT}
JSON

cat <<EOF

=== Monitor ===
  SELECT status, current_step, LEFT(COALESCE(error,''),160)
    FROM orchestration_states WHERE correlation_id='$CORR'::uuid;

  -- what actually survived verification (read this before the draft)
  SELECT summary, jsonb_pretty(spec->'grounding_audit') FROM site_work_items wi
    JOIN sites s ON s.id=wi.site_id
   WHERE s.domain='$DOMAIN' AND wi.item_type='grounded_draft_review'
   ORDER BY wi.created_at DESC LIMIT 1;

An absent orchestration row means QUEUED, not dropped — see bugs_open/096, a long
council run head-of-line blocks this lane. Check consumer-group lag before
re-firing; a duplicate does the work twice.
EOF
