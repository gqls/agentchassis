#!/bin/bash
# Staged dispatch: remove the orphan /assets/images/logo.jpg from
# gaswholesalers.com through retract_asset_files (DGH-010). Owner-directed
# 2026-08-16; capability committed 2026-08-17 (37303a6dd), INERT until a roll.
#
# PRECONDITIONS — all three, in order, no skipping:
#  1. The roll: the running chassis carries the action.
#     POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
#     kubectl -n ai-persona-system exec $POD -- grep -aq "retract_asset_files requires a database handle" /proc/1/exe && echo PRESENT
#     kubectl -n ai-persona-system exec $POD -- grep -aq "ZZZ_ABSENT_CONTROL" /proc/1/exe || echo control-ok
#  2. The seed: rename sql_for_agents/446_asset_retraction_agent_HOLD.sql (drop _HOLD),
#     apply via the migration runner SCOPED TO THAT FILE's directory.
#  3. The dry run (this script's default): read result.refusals before arming.
#
# Usage:   ./RETRACT_gaswholesalers_logo_jpg.sh            # dry run (audit only)
#          ARM=1 ./RETRACT_gaswholesalers_logo_jpg.sh      # the real deletion
#
# VERIFY AT THE WIRE, CONTROL PAIR — a check asserting only the 404 would pass
# if the wrong file were deleted:
#   curl -s -o /dev/null -w 'logo.jpg %{http_code}\n' https://gaswholesalers.com/assets/images/logo.jpg   # want 404
#   curl -s -o /dev/null -w 'logo.png %{http_code}\n' https://gaswholesalers.com/assets/images/logo.png   # MUST stay 200
# And the audit row:
#   SELECT error_code, error_message FROM agent_error_log
#   WHERE error_code LIKE 'ASSET_RETRACTION%' ORDER BY created_at DESC LIMIT 5;
set -euo pipefail
SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
DOMAIN="gaswholesalers.com"
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)

if [ "${ARM:-0}" = "1" ]; then
  CONFIG='{"agent_type":"asset-retraction","step_overrides":{"retract":{"dry_run":false}}}'
  echo "ARMED — this run deletes."
else
  CONFIG='{"agent_type":"asset-retraction"}'
  echo "Dry run (default). ARM=1 to delete."
fi

PAYLOAD=$(printf '{"action":"orchestrate","config":%s,"input_data":{"site_id":"%s","paths":["/assets/images/logo.jpg"]}}' "$CONFIG" "$SITE_ID")
echo "$PAYLOAD" | kubectl -n kafka run -i --rm "kcat-retract-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID -H request_id=$(cat /proc/sys/kernel/random/uuid) \
  -H message_id=$(cat /proc/sys/kernel/random/uuid) -H orchestration_id=$(cat /proc/sys/kernel/random/uuid) \
  -H orchestration_name=asset-retract-${DOMAIN}-$(date +%Y%m%d-%H%M%S) \
  -H step_name=start -H client_id=demo_client -H message_type=request \
  -H action=orchestrate -H from_agent_type=user -H from_agent_id=cli \
  -H responses_topic=system.agent.generic.responses
echo "correlation: $CORRELATION_ID  — kcat can exit 0 having sent NOTHING; check orchestration_states."
# ⚠ step_overrides reaching the step's config is [UNVERIFIED] for this agent
# shape — if the armed run still reports dry_run:true, the fallback is a
# one-off agent_definitions config UPDATE (snapshot first), or extend the seed
# with an armed variant. The DRY RUN needs no override and works as-is.
