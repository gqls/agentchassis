#!/usr/bin/env bash
# D6 canary replan for loanandmortgagecalculator.co.uk — fires ONE direct
# build-site-planner orchestration and prints the queries that judge it.
#
# Adapted from loancalculator_couk/canary_replan_407.sh (same machinery, one
# site over). What differs here, and why the judging queries are not a copy:
#
#   * This site has ZERO locked page_components (B2 unlocked all of them), so
#     the sibling's "locked rows untouched" check carries no information here.
#     The protection is rebuild_policy='owned' (24 owned / 22 generic, measured
#     pre-fire) — so that census is checked instead.
#   * 38 of 45 pages are NOT fixed points of CanonicalisePage, so the identity
#     md5 is the load-bearing check, not a formality. honour_realised_identity
#     was seeded true immediately before this run for exactly that reason
#     (SEED_2026-08-17_identity_and_tools.sql; spec row 6ca809d6).
#   * Success is NOT "the plan is right". Success is: the plan names the 23
#     calculator slots, nothing shrank, no identity moved, arithmetic still
#     proves. Divergence from today's site is the OUTPUT of this run, per D6.
#
# ⚠ kcat -P exits 0 having sent nothing — proof of dispatch is the
#   orchestration row, queried BY CORRELATION (never by now()-interval).
# ⚠ No dispatch within ~300s of a chassis pod (re)start (checked: pods ~14h old).
set -euo pipefail

SITE_ID='ed633ada-f8af-424b-b4d4-8af79160dbcd'
DOMAIN='loanandmortgagecalculator.co.uk'
PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"

# Pre-state, measured 2026-08-17 11:55Z, before the fire:
PRE_IDENTITY_MD5='86f42aa5d5c122da938c970a01dc6ff4'   # name|url|page_type|status
PRE_ACTIVE=45; PRE_ARCHIVED=1; PRE_OWNED=24; PRE_GENERIC=22
# Snapshot tables: pages_bak_20260817_preplan_lmc (46), page_components_bak_20260817_preplan_lmc (82)

CORR=$(cat /proc/sys/kernel/random/uuid)
ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid)
MSG=$(cat /proc/sys/kernel/random/uuid)
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "== PRE-STATE RE-READ (must match the constants above; this tree moves in hours) =="
$PSQL -tA -c "SELECT md5(string_agg(name||'|'||url||'|'||page_type||'|'||status, E'\n' ORDER BY name)) FROM pages WHERE site_id='$SITE_ID';"
$PSQL -tA -F$'\t' -c "SELECT count(*) FILTER (WHERE status='active'), count(*) FILTER (WHERE status='archived') FROM pages WHERE site_id='$SITE_ID';"
$PSQL -tA -F$'\t' -c "SELECT COALESCE(rebuild_policy,'(null)'), count(*) FROM pages WHERE site_id='$SITE_ID' GROUP BY 1 ORDER BY 2 DESC;"
$PSQL -tA -F$'\t' -c "SELECT data->>'honour_realised_identity', data->>'plan_includes_tools' FROM site_specs WHERE site_id='$SITE_ID' AND aspect='structure' AND is_current;"

echo "== FIRING corr=$CORR =="
kubectl -n kafka run -i --rm "kcat-lmc-canary-$(date +%s)" \
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

== JUDGE IT (poll by correlation; budget tens of minutes for queue latency) ==
# 0. Proof of dispatch (planner rows purge in ~2 days — read it soon):
$PSQL -c "SELECT orchestration_id, current_step, status, created_at FROM orchestration_states WHERE correlation_id='$CORR';"

# 1. NO IDENTITY MOVED — the load-bearing check on this site. Must still be
#    $PRE_IDENTITY_MD5. A different digest means honour_realised_identity did
#    not hold and pages have been renamed/moved/archived:
$PSQL -tA -c "SELECT md5(string_agg(name||'|'||url||'|'||page_type||'|'||status, E'\n' ORDER BY name)) FROM pages WHERE site_id='$SITE_ID';"

# 2. NOTHING SHRANK, and nothing was minted: expect $PRE_ACTIVE active /
#    $PRE_ARCHIVED archived. MORE active pages is allowed by D6 (growth is
#    welcome) — but check WHICH, because 17 phantom tool-<name> rows are the
#    failure this run is guarding against:
$PSQL -tA -F\$'\t' -c "SELECT count(*) FILTER (WHERE status='active'), count(*) FILTER (WHERE status='archived') FROM pages WHERE site_id='$SITE_ID';"
$PSQL -tA -c "SELECT name, url, created_at FROM pages WHERE site_id='$SITE_ID' AND created_at > '2026-08-17' ORDER BY created_at;"

# 3. THE SUCCESS CRITERION — the plan names the calculator slots (23 expected;
#    compare against the list, not the count, RUNBOOK query 3):
$PSQL -c "SELECT sps.page_name, sps.ordering, sps.component_name FROM site_plans sp JOIN site_plan_sections sps ON sps.plan_id=sp.id WHERE sp.site_id='$SITE_ID' AND sp.is_current AND sps.component_name LIKE '%tool%' ORDER BY 1,2;"
$PSQL -tA -F\$'\t' -c "SELECT (SELECT count(*) FROM site_plan_pages WHERE plan_id=sp.id) AS planned_pages, (SELECT count(*) FROM site_plan_sections WHERE plan_id=sp.id) AS planned_sections FROM site_plans sp WHERE sp.site_id='$SITE_ID' AND sp.is_current;"

# 4. Ownership census unchanged (expect $PRE_OWNED owned / $PRE_GENERIC generic):
$PSQL -tA -F\$'\t' -c "SELECT COALESCE(rebuild_policy,'(null)'), count(*) FROM pages WHERE site_id='$SITE_ID' GROUP BY 1 ORDER BY 2 DESC;"

# 5. Arithmetic still proves, in the same session as the judgement:
python3 docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk/oracle.py
#    expect PASS 170 / FAIL 0 / CONVENTION 6 / N/A 0

# 6. The plan's own divergence from today's site (this is the D6 product):
$PSQL -c "SELECT spp.name, spp.role, spp.url FROM site_plan_pages spp JOIN site_plans sp ON sp.id=spp.plan_id WHERE sp.site_id='$SITE_ID' AND sp.is_current ORDER BY spp.role, spp.name;"

# ROLLBACK, if identity moved (forward-only; restore FROM the snapshot, do not
# drop rows): pages_bak_20260817_preplan_lmc holds all 46 rows as of 11:55Z.
EOF
