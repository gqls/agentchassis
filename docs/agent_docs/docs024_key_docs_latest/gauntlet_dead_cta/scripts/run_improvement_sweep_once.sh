#!/usr/bin/env bash
# run_improvement_sweep_once.sh — fire the `improvement-sweep` scheduled task ONCE,
# at a site you name, without enabling the task.
#
# WHY THIS EXISTS
#   `improvement-sweep` (agent `improvement-loop`) has been disabled since
#   2026-05-02. Its step `triage_findings` is the ONLY promoter of work items from
#   status='detected' to status='triaged' — and the build pipeline only ever
#   dispatches 'triaged'. So while it is off, every discovery finding parks
#   forever. That is bugs_open/083 BY SLUG (…detected_findings_never_reach_a_handler),
#   NOT the gauntlet-engine-503 083.
#
# WHY NOT JUST `UPDATE scheduled_tasks SET enabled = true`
#   interval_seconds is 180, so enabling fires it every 3 minutes at whatever site
#   its pre_query picks, until someone disables it again. The owner ruling of
#   2026-07-29 is that the loop is stopped DELIBERATELY. This runs it once and
#   changes no configuration.
#
# WHY YOU MUST PASS A SITE
#   The task's own pre_query ends `ORDER BY s.updated_at ASC NULLS FIRST LIMIT 1`
#   — one site per firing, the least-recently-touched eligible one. That is almost
#   never the site you are thinking about; worse, CREATING a work item bumps
#   sites.updated_at, so the site you just found a defect on sorts LAST. Naming the
#   site is the whole point of this script.
#
# BLAST RADIUS — read before firing. One run against SITE_ID will:
#   1. run five LLM agents on it (quality / design / completeness / design_audit /
#      site_review), each of which may CREATE new detected items — unless
#      get_audit_pass_count(site) >= 3, in which case discovery is skipped;
#   2. promote EVERY detected item on that site, of every item_type, to
#      status='triaged', pipeline='build' (triage_detect_items_action.go:108 —
#      the WHERE clause is site_id + status only, there is no type filter);
#   3. spawn build-dispatch-loop, which claims up to 5 and fires their handlers.
#   And independently of step 3, the `build-pipeline-trigger` scheduled task IS
#   enabled and ticks every 120s looking for exactly triaged+build — so promotion
#   alone is enough to start real fixers editing a LIVE site within ~2 minutes.
#   Count first:  SELECT item_type, count(*) FROM site_work_items
#                 WHERE site_id = '<id>' AND status='detected' GROUP BY 1;
#
# PRE-REQ: not within ~300s of an agent-chassis pod restart, or the spawn inside
#   the workflow is silently dropped.  kubectl -n ai-persona-system get pods -l app=agent-chassis
#
# The envelope below is a byte-faithful copy of cmd/scheduler/main.go fireTrigger()
# (:488-527) — same topic, same body, same headers — with ONE deliberate change:
# orchestration_name carries `-manual-` so the run is identifiable as hand-fired.
# The publish uses the hardened kcat form (payload in the container COMMAND,
# --command to beat the kcat ENTRYPOINT, `&& echo PUBLISH_OK`): the
# `kubectl run -i ... <<JSON` form sees EOF and produces NOTHING at exit 0.
set -u

SITE_ID="${1:-}"
DOMAIN="${2:-}"
if [ -z "${SITE_ID}" ] || [ -z "${DOMAIN}" ]; then
  echo "usage: $0 <site_id> <domain>" >&2
  echo "  e.g. $0 e33263f4-74f8-494f-b191-546845dbbddf gamesdesign.co.uk" >&2
  exit 2
fi

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCH_NAME="sched-improvement-sweep-manual-$(date -u +%Y%m%d-%H%M%S)"

BODY="{\"action\":\"orchestrate\",\"config\":{\"agent_type\":\"improvement-loop\"},\"input_data\":{\"site_id\":\"${SITE_ID}\",\"domain\":\"${DOMAIN}\"}}"

echo "========================================="
echo "improvement-sweep, ONE firing -> ${DOMAIN}"
echo "  site_id          = ${SITE_ID}"
echo "  SAVE: ORCHESTRATION_ID=${ORCHESTRATION_ID}"
echo "========================================="

kubectl -n kafka run "kcat-improvesweep-$(date +%s)-$RANDOM" --rm --restart=Never \
  --image=edenhill/kcat:1.7.1 --attach=true --quiet \
  --command -- sh -c "printf '%s' '${BODY}' | kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 -t system.agent.generic.requests \
  -k ${REQUEST_ID} \
  -H correlation_id=${CORRELATION_ID} \
  -H request_id=${REQUEST_ID} \
  -H message_id=${MESSAGE_ID} \
  -H orchestration_id=${ORCHESTRATION_ID} \
  -H orchestration_name=${ORCH_NAME} \
  -H step_name=start \
  -H client_id=system \
  -H message_type=request \
  -H action=orchestrate \
  -H from_agent_type=kafka-scheduler \
  -H from_agent_id=kafka-scheduler-singleton && echo PUBLISH_OK"

echo ""
echo "No PUBLISH_OK above means NOTHING was sent, whatever the exit code."
echo ""
echo "Find the run BY PAYLOAD, not by the printed id:"
echo "  SELECT orchestration_id, current_step, status, created_at FROM orchestration_states"
echo "   WHERE collected_data->'input_data'->>'site_id' = '${SITE_ID}'"
echo "   ORDER BY created_at DESC LIMIT 5;"
echo ""
echo "Did it promote anything?"
echo "  SELECT item_type, status, triaged_at FROM site_work_items"
echo "   WHERE site_id = '${SITE_ID}' AND triaged_at > now() - interval '15 minutes';"
