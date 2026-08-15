#!/usr/bin/env bash
# PHASE 2 of the 2026-08-14/15 rebuild fire — targeted recompose of the 26
# pre-fire built pages (owner ruling 2026-08-14: explicit recompose, all 26).
#
# FIRE ONLY AFTER PHASE 1 SETTLES: new plan landed, the two restored pages
# built, serving 29/29, checkpoint_postplan findings dispositioned.
#
# Mechanism (NOTES 2026-08-14 late night): 082 has no spec plumbing, so the
# recompose list rides a DIRECT build-site-planner dispatch, exactly the
# canary_replan_407.sh shape. The redesign intent PROSE is already standing in
# the mission spec ("reconsider every existing page's layout and recompose it
# ... do not simply re-emit the current layouts"), so the recompose no-op
# landmine's both-places rule is satisfied.
#
# EXPECTED OUTPUT (not failure): the reconciler parks the 11 tool-role pages'
# rebuilds at owned_page_review (human gate, TP-004); ~15 non-tool pages get
# needs_page items and rebuild unattended; 1 needs_rerender at the end.
#
# ⚠ kcat -P exits 0 having sent nothing — proof is the orchestration row BY
#   CORRELATION. ⚠ No dispatch within ~300s of a chassis pod (re)start.
set -euo pipefail

SITE_ID='0162cde4-633e-45e9-8ca6-87a6b2fe1d26'
DOMAIN='loancalculator.co.uk'
PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"

CORR=$(cat /proc/sys/kernel/random/uuid)
ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid)
MSG=$(cat /proc/sys/kernel/random/uuid)
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

RECOMPOSE='["guide-can-i-overpay","guide-car-finance-explained","guide-debt-consolidation-explained","guide-debt-help-uk","guide-document-checklist","guide-finance-damage-and-insurance","guide-fixed-vs-variable-loans","guide-hidden-loan-fees","guide-how-loans-are-calculated","guide-jargon-buster","guide-loan-eligibility-uk","guide-secured-vs-unsecured","guide-uk-lending-landscape","index","legal","tool-application-tracker","tool-car-finance-calculator","tool-compare-loans","tool-consolidation","tool-credit-health-check","tool-credit-roadmap","tool-damage-checker","tool-interest-rate-stress-test","tool-loan-vs-savings","tool-overpayment-calculator","tool-settlement-calculator"]'

echo "== PRE-STATE (save this output) =="
$PSQL -tA -c "SELECT now() AS db_fire_time;"
$PSQL -tA -c "SELECT count(*), count(*) FILTER (WHERE status='active') FROM pages WHERE site_id='$SITE_ID';"
$PSQL -tA -c "SELECT md5(string_agg(name||'|'||url||'|'||page_type||'|'||status, E'\n' ORDER BY name)) FROM pages WHERE site_id='$SITE_ID';"
$PSQL -tA -c "SELECT count(*) FROM page_components pc JOIN pages p ON pc.page_id=p.id WHERE p.site_id='$SITE_ID' AND pc.locked_at IS NOT NULL;"
$PSQL -tA -c "SELECT id, created_at FROM site_plans WHERE site_id='$SITE_ID' AND is_current;"

echo "== FIRING corr=$CORR =="
kubectl -n kafka run -i --rm "kcat-recompose-$(date +%s)" \
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
{"action":"orchestrate","config":{"agent_type":"build-site-planner"},"input_data":{"site_id":"$SITE_ID","domain":"$DOMAIN","spec":{"recompose_pages":$RECOMPOSE}}}
JSON

echo "CORR=$CORR"
cat <<EOF

== JUDGE IT (poll by correlation; planner rows purge in ~2 days, read soon) ==
# 1. Proof of dispatch:
$PSQL -c "SELECT orchestration_id, current_step, status FROM orchestration_states WHERE correlation_id='$CORR';"

# 2. THE RECOMPOSE TELL — proposed_verbatim = the no-op landmine fired for that
#    page; absent_from_plan = dropped/renamed (investigate). Absence of a row
#    for a page = it was genuinely recomposed:
$PSQL -c "SELECT error_message, count(*) FROM agent_error_log WHERE error_code='RECOMPOSE_INTENT_NOT_REALISED' AND created_at > now()-interval '1 hour' GROUP BY 1;"

# 3. Compositions actually changed (compare against the pre-phase-2 baseline
#    captured by the runbook):
$PSQL -c "SELECT sps.page_name, count(*), string_agg(sps.component_name, ',' ORDER BY sps.position) FROM site_plans sp JOIN site_plan_sections sps ON sps.plan_id=sp.id WHERE sp.site_id='$SITE_ID' AND sp.is_current GROUP BY 1 ORDER BY 1;"

# 4. Tool placements kept (THE placement test of this whole lane):
$PSQL -c "SELECT sps.page_name, sps.component_name FROM site_plans sp JOIN site_plan_sections sps ON sps.plan_id=sp.id WHERE sp.site_id='$SITE_ID' AND sp.is_current AND sps.component_name LIKE 'tool-%' ORDER BY 1;"

# 5. What the reconciler emitted (expect ~15 needs_page + 11 owned_page_review + 1 needs_rerender):
$PSQL -c "SELECT item_type, status, count(*) FROM site_work_items WHERE site_id='$SITE_ID' AND created_at > now()-interval '1 hour' GROUP BY 1,2;"

# 6. Locked rows untouched (12/12):
$PSQL -tA -c "SELECT count(*) FROM page_components pc JOIN pages p ON pc.page_id=p.id WHERE p.site_id='$SITE_ID' AND pc.locked_by='decompose_20260802_proven_calculators' AND pc.locked_at IS NOT NULL;"
EOF
