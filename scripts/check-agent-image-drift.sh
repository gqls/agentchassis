#!/bin/bash
# Read-only. Three questions about one fact, answered separately:
#   1. what image is the agent-chassis Deployment RUNNING?
#   2. what do the agent_definitions rows SAY?
#   3. what will a spawned agent pod ACTUALLY run?
#
# bugs_open/066. (2) and (3) used to be the same question — a spawned dedicated
# agent pod took its image from the row, so a stale row put a stale binary into
# a pod and the deployment pod-grep was a FALSE GREEN for it. Since chassis
# v1.0.117x the spawn path resolves the image from the RUNNING chassis pod
# (platform/orchestration/actions/agent_image.go), so (3) follows (1).
#
# That is why this script exists: the census in the RUNBOOKs reads (2) and was
# being read as an answer to (3). After the fix a row can be stale and no pod is
# affected — the row is a RECORD, kept honest by scripts/deploy/update-agent-images.sh
# at deploy time. Drift here is a bookkeeping defect, not an outage, and this
# script says which.
#
# Usage: scripts/check-agent-image-drift.sh
# Exit:  0 = no drift · 1 = drift in the record · 2 = could not determine

set -uo pipefail

NAMESPACE="${NAMESPACE:-ai-persona-system}"
DEPLOYMENT="${DEPLOYMENT:-agent-chassis}"

# stderr is NOT swallowed. A `2>/dev/null` here would turn a broken query into
# an empty section that reads exactly like "nothing to report" — the same silent
# no-op this script exists to expose. (It did, on the first run of this script.)
psql_q() {
    kubectl -n "$NAMESPACE" exec -i postgres-clients-0 -- \
        psql -U clients_user -d clients_db -tAc "$1"
}

echo "=== 1. what the Deployment is running ==="
DEPLOY_IMAGE=$(kubectl -n "$NAMESPACE" get deployment "$DEPLOYMENT" \
    -o jsonpath='{.spec.template.spec.containers[0].image}' 2>/dev/null)
if [ -z "$DEPLOY_IMAGE" ]; then
    echo "  ERROR: could not read deployment/$DEPLOYMENT" >&2
    exit 2
fi
echo "  deployment spec : $DEPLOY_IMAGE"

# The spec is what was ASKED for; the pod is what is SERVING. During a rollout
# they differ, and the pod is the one that matters for what spawns inherit.
kubectl -n "$NAMESPACE" get pods -l app="$DEPLOYMENT" \
    -o custom-columns='  pod             :.metadata.name,IMAGE:.spec.containers[0].image,STARTED:.status.startTime' \
    --no-headers 2>/dev/null | sed 's/^/  running pod     : /'

DEPLOY_TAG="${DEPLOY_IMAGE##*:}"
DEPLOY_REPO="${DEPLOY_IMAGE%:*}"
if [[ "$DEPLOY_IMAGE" == *"@"* ]]; then
    echo "  NOTE: digest-pinned; there is no tag to compare against." >&2
    exit 2
fi

echo
echo "=== 2. what the agent_definitions rows say ==="
psql_q "
SELECT '  ' || COALESCE(image_repository,'(null)') || ':' || COALESCE(image_tag,'(null)')
       || '  x' || count(*)::text
       || CASE WHEN COALESCE(image_tag,'') <> '$DEPLOY_TAG' THEN '   <-- DRIFT' ELSE '' END
FROM agent_definitions
WHERE deleted_at IS NULL AND COALESCE(is_snapshot,false) = false
GROUP BY image_repository, image_tag
ORDER BY count(*) DESC;"

PINNED=$(psql_q "
SELECT count(*) FROM agent_definitions
WHERE deleted_at IS NULL
  AND COALESCE(default_config->'pin_image_tag','false'::jsonb) = 'true'::jsonb;")
echo "  deliberately pinned (default_config.pin_image_tag): ${PINNED:-?}"

DRIFTED=$(psql_q "
SELECT count(*) FROM agent_definitions
WHERE deleted_at IS NULL AND COALESCE(is_snapshot,false) = false
  AND image_repository = '$DEPLOY_REPO'
  AND COALESCE(image_tag,'') <> '$DEPLOY_TAG'
  AND COALESCE(default_config->'pin_image_tag','false'::jsonb) <> 'true'::jsonb;")

echo
echo "=== 3. what spawned pods are actually running ==="
SPAWNED=$(kubectl -n "$NAMESPACE" get pods -l app=dynamic-agent \
    -o jsonpath='{range .items[*]}{.spec.containers[0].image}{"\n"}{end}' 2>/dev/null | sort | uniq -c | sort -rn)
if [ -z "$SPAWNED" ]; then
    echo "  (no dynamic-agent pods alive right now — they are short-lived)"
else
    echo "$SPAWNED" | sed 's/^/  /'
fi

echo
echo "=== verdict ==="
if [ -z "${DRIFTED:-}" ]; then
    echo "  COULD NOT DETERMINE — the database query returned nothing."
    exit 2
fi
if [ "$DRIFTED" -eq 0 ]; then
    echo "  No drift: every unpinned chassis row records $DEPLOY_TAG."
    exit 0
fi
echo "  DRIFT: $DRIFTED unpinned row(s) record a tag other than $DEPLOY_TAG."
echo "  Effect on pods: NONE if the chassis is v1.0.117x or later — the spawn path"
echo "  takes the image from the running chassis pod, and the drift warning"
echo "  'bugs_open/066: agent_definitions.image_tag trails' appears in its log."
echo "  Confirm the fix is in the running binary before trusting that:"
echo "    kubectl exec -n $NAMESPACE <chassis-pod> -- sh -c \\"
echo "      'strings /app/agent-chassis | grep -c \"bugs_open/066: agent_definitions.image_tag trails\"'"
echo "  Repair the record with: scripts/deploy/update-agent-images.sh"
exit 1
