#!/usr/bin/env bash
# Re-fire the page-rebuild for a domain — the copy rewrite this lane has staged.
#
# WHY THIS EXISTS: the rewrite was dispatched 2026-08-23 and FAILED on the account's
# Anthropic usage limit ("regain access on 2026-09-01"). Everything it needs is already
# in place — the corrected `identity` and `content_direction` are live and pinned, and
# pages.build_status is still 'needs_rebuild', which is the precondition the agent reads.
# So this is one command, and it should not have to be reconstructed from a transcript.
#
# The repo's own trigger (scripts/initial_messages/110_page_rebuild/072_page_rebuild)
# hardcodes leopardessconsulting.co.uk. This is the same envelope, parameterised, with
# the guards that cost this lane time.
#
# ⚠ DO NOT hand-write the copy instead. The owner's instruction (2026-08-23) was
# explicitly to put the copywriting through the framework, and the 2026-08-04 ruling says
# the same. The live copy is in the wrong voice but is honest and asserts no quantities,
# so it is safe to leave standing until this can run.
set -euo pipefail

# GUARD PROVENANCE [2026-08-23]: guard 1's REFUSE arm is proven in isolation, not by
# observation — when it was first run the claude endpoint had just recovered, so the live
# run exercised only the ALLOW path. Both arms were then driven directly: HEALTHY=f and
# HEALTHY='' (a failed query) both refuse, HEALTHY=t proceeds. The empty case matters: a
# psql failure must not read as healthy. Guard 2 likewise (N=0 refuses, N=1 proceeds).

DOMAIN="${1:?usage: $0 <domain>   e.g. $0 apis.uk}"
NS=ai-persona-system
PSQL="kubectl -n $NS exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"

# GUARD 1 — the AI endpoint. This is the exact thing that killed the 2026-08-23 run, and
# it fails deep inside a child orchestration where the error is easy to misread as a bug.
HEALTHY=$($PSQL -tAc "SELECT healthy FROM ai_endpoint_health WHERE name='claude';" | tr -d '[:space:]')
if [ "$HEALTHY" != "t" ]; then
  echo "REFUSING: the claude endpoint is not healthy — a rebuild will fail mid-run." >&2
  $PSQL -tAc "SELECT left(coalesce(error,'(no error recorded)'),200) FROM ai_endpoint_health WHERE name='claude';" >&2
  exit 4
fi

# GUARD 2 — something must actually be flagged, or the agent runs and rebuilds nothing
# while reporting success.
N=$($PSQL -tAc "SELECT count(*) FROM pages WHERE site_id=(SELECT id FROM sites WHERE domain='$DOMAIN') AND build_status='needs_rebuild';" | tr -d '[:space:]')
if [ "$N" = "0" ]; then
  echo "REFUSING: no page on $DOMAIN is flagged build_status='needs_rebuild'." >&2
  echo "  Flag one first:  UPDATE pages SET build_status='needs_rebuild', updated_at=NOW()" >&2
  echo "                     WHERE site_id=(SELECT id FROM sites WHERE domain='$DOMAIN') AND name='index';" >&2
  exit 5
fi

# GUARD 3 — a spawn within ~300s of a chassis restart is silently dropped.
kubectl -n $NS rollout status deploy/agent-chassis --timeout=20s >/dev/null || {
  echo "REFUSING: agent-chassis is mid-rollout — the spawn would be dropped." >&2; exit 3; }

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"   # ⚠ NEVER a hyphenated invented id: client_id is interpolated
                          # UNQUOTED as a schema name (spawn_actions.go), and a hyphen dies
                          # as SQLSTATE 42601 in a way that reads like a platform fault.

echo "page-rebuild: $DOMAIN  ($N page(s) flagged)"
echo "SAVE: CORRELATION_ID=$CORRELATION_ID  ORCHESTRATION_ID=$ORCHESTRATION_ID"

kubectl -n kafka run -i --rm "kcat-page-rebuild-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -c 1 -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID -H orchestration_id=$ORCHESTRATION_ID \
  -H request_id=$REQUEST_ID -H message_id=$MESSAGE_ID -H message_type=request \
  -H client_id=$CLIENT_ID -H action=process -H sender_agent_type=cli -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses -H timestamp=$TIMESTAMP <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_rebuilder","processing_mode":"orchestrator","timeout_seconds":1200,"steps":{"spawn_rebuilder":{"action":"spawn_agent","config":{"role":"rebuilder","agent_type":"page-rebuild"},"output_field":"rebuilder_agent","next_step":"call_rebuilder","description":"Spawn page-rebuild agent"},"call_rebuilder":{"action":"call_agent","config":{"agent_type":"page-rebuild","target_role":"rebuilder","input_mapping":{"domain":"input_data.domain"},"timeout_seconds":900},"output_field":"rebuild_result","next_step":"complete","description":"Run page rebuild"},"complete":{"action":"complete_workflow","config":{"output_fields":["rebuild_result"]},"description":"Page rebuild complete"}}}},"input_data":{"domain":"${DOMAIN}"}}
JSON

cat <<EOF

⚠ kcat exits 0 having published nothing. The orchestration row is the only proof:
  SELECT status, current_step, left(coalesce(error,'-'),200) FROM orchestration_states
   WHERE orchestration_id='${ORCHESTRATION_ID}'::uuid;

THEN VERIFY AT THE ARTEFACT, NOT THE STATUS — this lane got that wrong once:
  1. read the page as PROSE:  curl -sS https://${DOMAIN}/ | sed 's/<[^>]*>//g' | head -60
  2. BOTH columns clean (a rerender can reuse the cached render):
     SELECT position, (content_data::text ~* '<thing>'), (coalesce(rendered_html,'') ~* '<thing>')
       FROM page_components pc JOIN pages p ON p.id=pc.page_id
      WHERE p.site_id=(SELECT id FROM sites WHERE domain='${DOMAIN}') ORDER BY position;
  3. python3 \$(dirname \$0)/check_evidence_base.py ${DOMAIN}
  4. for apis.uk ONLY — the API is an independent fact, check it separately:
     curl -sS -o /dev/null -w '%{http_code}\\n' -X POST https://tools.apis.uk/api/v1/tools/gauntlet/round \\
       -H 'Origin: https://vonc.com' -H 'Content-Type: application/json' -d '{}'   # expect 200
EOF
