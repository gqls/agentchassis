#!/usr/bin/env bash
# Canary replan for the 407 planner widening (PLAN-049) — fires ONE direct
# build-site-planner orchestration at loancalculator.co.uk and prints the
# queries that judge it. Run AFTER the council verdict on 508fe8eb is read.
#
# What this exercises: the widened load_components menu (flag ON + placement
# gate), the site_record.site_id param binding, seed 362's converge-on-realised
# behaviour, and write_site_plan/sync_pages under url_shape=flat. What it does
# NOT exercise: the matchLockedRow identity arm — that fires at page BUILD time
# (save path), not at plan time. Plan-time success = the plan NAMES the tools.
#
# ⚠ kcat -P exits 0 having sent nothing — proof of dispatch is the
#   orchestration row, queried BY CORRELATION (never by now()-interval).
# ⚠ No dispatch within ~300s of a chassis pod (re)start.
set -euo pipefail

SITE_ID='0162cde4-633e-45e9-8ca6-87a6b2fe1d26'
DOMAIN='loancalculator.co.uk'
PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"

CORR=$(cat /proc/sys/kernel/random/uuid)
ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid)
MSG=$(cat /proc/sys/kernel/random/uuid)
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "== PRE-STATE (save this output) =="
$PSQL -tA -c "SELECT count(*), count(*) FILTER (WHERE status='active') FROM pages WHERE site_id='$SITE_ID';"
$PSQL -tA -c "SELECT md5(string_agg(name||'|'||url||'|'||page_type||'|'||status, E'\n' ORDER BY name)) FROM pages WHERE site_id='$SITE_ID';"
$PSQL -tA -c "SELECT count(*) FROM page_components pc JOIN pages p ON pc.page_id=p.id WHERE p.site_id='$SITE_ID' AND pc.locked_at IS NOT NULL;"
$PSQL -tA -c "SELECT id, created_at FROM site_plans WHERE site_id='$SITE_ID' ORDER BY created_at DESC LIMIT 1;"

echo "== FIRING corr=$CORR =="
kubectl -n kafka run -i --rm "kcat-canary-$(date +%s)" \
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
  -H sender_agent_type=cli \
  -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TS <<JSON
{"action":"orchestrate","config":{"agent_type":"build-site-planner"},"input_data":{"site_id":"$SITE_ID","domain":"$DOMAIN"}}
JSON

echo "CORR=$CORR"
cat <<EOF

== JUDGE IT (poll by correlation, budget minutes-to-tens-of-minutes) ==
# 1. The run (proof of dispatch; planner rows purge in ~2 days, read it soon):
$PSQL -c "SELECT orchestration_id, current_step, status FROM orchestration_states WHERE correlation_id='$CORR';"

# 2. THE SUCCESS CRITERION — the new plan NAMES the tools (expect the 12
#    locked placements' functions on their pages; component_name holds the
#    component's function-style name, measured on noted's 08-12 plan):
$PSQL -c "SELECT sps.page_name, sps.component_name FROM site_plans sp JOIN site_plan_sections sps ON sps.plan_id=sp.id WHERE sp.site_id='$SITE_ID' AND sp.is_current AND sps.component_name LIKE 'tool-%' ORDER BY 1;"

# 3. NO page identity moved (md5 must equal the pre-state value):
$PSQL -tA -c "SELECT md5(string_agg(name||'|'||url||'|'||page_type||'|'||status, E'\n' ORDER BY name)) FROM pages WHERE site_id='$SITE_ID';"

# 4. Locked rows untouched (count AND rendered_html lengths unchanged):
$PSQL -tA -c "SELECT count(*) FROM page_components pc JOIN pages p ON pc.page_id=p.id WHERE p.site_id='$SITE_ID' AND pc.locked_at IS NOT NULL;"

# 5. The menu the run actually saw, from its own log line (while the row lives):
$PSQL -c "SELECT collected_data->'available_components' IS NOT NULL, jsonb_array_length(collected_data->'available_components') FROM orchestration_states WHERE correlation_id='$CORR';"
#    Expect 140. An unflagged control site's next planner run must show 129.

# 6. Serving unchanged (the wire, guarded):
./check_site_serving.sh
EOF
