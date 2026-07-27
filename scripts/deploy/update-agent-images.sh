#!/bin/bash
# Sync agent_definitions.image_tag with the tag the agent-chassis Deployment is
# actually running.
#
# bugs_open/066. A dedicated agent pod takes its image from the
# agent_definitions row, not from the Deployment, so a chassis roll never
# reached it: on 2026-07-24 the fleet sat at v1.0.1151 while the Deployment ran
# v1.0.1155, and a run failed on a bug that had already shipped.
#
# WHAT THIS SCRIPT IS AND IS NOT. Since chassis v1.0.117x the spawn path
# resolves the image from the RUNNING chassis pod
# (platform/orchestration/actions/agent_image.go), so a stale row can no longer
# put a stale binary into a pod. This script keeps the recorded value HONEST —
# it is the fallback if that lookup ever fails, and it is what every census in
# the RUNBOOKs reads. It is wired into the tail of `make deploy-agents` (target
# `sync-agent-image-tags`), AFTER the apply, so it records what actually rolled
# rather than what was asked for. Written Aug 2025 and never wired in until
# 2026-07-27 — which is the whole reason 066 happened.
#
# Read-only preview: DRY_RUN=1 scripts/deploy/update-agent-images.sh

set -uo pipefail

NAMESPACE="${NAMESPACE:-ai-persona-system}"
DEPLOYMENT="${DEPLOYMENT:-agent-chassis}"
DRY_RUN="${DRY_RUN:-0}"

# The tag comes from the live Deployment, not from $(IMAGE_TAG): the point is to
# record what IS running, and a make variable is a request, not an outcome.
CURRENT_IMAGE=$(kubectl -n "$NAMESPACE" get deployment "$DEPLOYMENT" \
    -o jsonpath='{.spec.template.spec.containers[0].image}' 2>/dev/null)

if [ -z "$CURRENT_IMAGE" ]; then
    echo "ERROR: could not read the image of deployment/$DEPLOYMENT in $NAMESPACE." >&2
    echo "       (On a fresh cluster the Deployment does not exist yet — bootstrap" >&2
    echo "        uses make deploy-100-bootstrap-agents, which seeds from \$(IMAGE_TAG).)" >&2
    exit 1
fi

# A digest reference has no tag to record; refuse rather than write nonsense.
if [[ "$CURRENT_IMAGE" == *"@"* ]]; then
    echo "ERROR: $DEPLOYMENT runs a digest-pinned image ($CURRENT_IMAGE)." >&2
    echo "       There is no tag to record. Nothing written." >&2
    exit 1
fi

if [[ $CURRENT_IMAGE =~ ^(.+):([^:/]+)$ ]]; then
    IMAGE_REPO="${BASH_REMATCH[1]}"
    IMAGE_TAG="${BASH_REMATCH[2]}"
else
    echo "ERROR: could not parse image: $CURRENT_IMAGE" >&2
    exit 1
fi

echo "deployment/$DEPLOYMENT is running:"
echo "  repository: $IMAGE_REPO"
echo "  tag:        $IMAGE_TAG"
echo

# The scope is the load-bearing part. The original of this script said
# `WHERE 1=1`, which also rewrote:
#   - is_snapshot rows  — version+1000 rollback copies (021_model_swap_and_rollback.sql);
#                         rewriting their tag destroys what a rollback restores;
#   - soft-deleted rows — resurrecting a dead row's image;
#   - image_repository  — silently converting an agent that deliberately runs
#                         some OTHER image onto the chassis image.
# It touches only rows already on this repository, and only their tag.
#
# pin_image_tag is the deliberate opt-out and is matched EXACTLY as the Go
# resolver matches it (a JSON boolean true, not the string "true"), so a pin
# means one thing in both places.
#
# is_active / status are deliberately NOT in the scope. Enumerated 2026-07-27
# before trusting the blast radius (the sites.status lesson — count by the
# column before assuming it means nothing): 105 experimental+active,
# 70 active+active, 5 experimental+inactive = the 180 rows in scope. `status` is
# informational here — both values are live agents — and the 5 inactive rows are
# INCLUDED on purpose: an inactive row can be reactivated, and leaving it on an
# ancient tag would re-open 066 for exactly that row.
#
# The values below are passed to psql as BOUND VARIABLES (-v … referenced as
# :'name'), not spliced into the SQL text. They come from a trusted deploy
# pipeline, but "SCHEMA FIRST, PARAMETERISED ALWAYS" is written without that
# exception, and psql's own quoting is more reliable than ours anyway.
# (Council gate 3e146ef2, constitution seat, medium.)
SQL_WHERE="deleted_at IS NULL
      AND COALESCE(is_snapshot, false) = false
      AND image_repository = :'img_repo'
      AND COALESCE(default_config->'pin_image_tag', 'false'::jsonb) <> 'true'::jsonb"

if [ "$DRY_RUN" = "1" ]; then
    echo "DRY_RUN=1 — nothing will be written."
    # Must be fed on STDIN, not with -c: psql performs :'var' interpolation only
    # on input it lexes (stdin/-f), and silently sends a -c string as-is, so the
    # parameterised form fails with `syntax error at or near ":"`. Verified both
    # ways 2026-07-27.
    kubectl -n "$NAMESPACE" exec -i postgres-clients-0 -- \
        psql -U clients_user -d clients_db \
        -v img_repo="$IMAGE_REPO" -v img_tag="$IMAGE_TAG" <<EOF
SELECT
  count(*)                                                                   AS in_scope,
  count(*) FILTER (WHERE COALESCE(image_tag,'') <> :'img_tag')               AS would_change,
  count(*) FILTER (WHERE COALESCE(image_tag,'')  = :'img_tag')               AS already_correct
FROM agent_definitions WHERE $SQL_WHERE;
EOF
    exit $?
fi

kubectl -n "$NAMESPACE" exec -i postgres-clients-0 -- \
    psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
    -v img_repo="$IMAGE_REPO" -v img_tag="$IMAGE_TAG" <<EOF
\\set ON_ERROR_STOP on

-- Before: what the rows say now.
SELECT COALESCE(image_tag,'(null)') AS tag_before, count(*)
FROM agent_definitions WHERE $SQL_WHERE
GROUP BY 1 ORDER BY 2 DESC;

UPDATE agent_definitions
SET image_tag = :'img_tag',
    updated_at = NOW()
WHERE $SQL_WHERE
  AND COALESCE(image_tag,'') <> :'img_tag';

-- Rows deliberately left alone, so the exclusions are visible rather than implied.
SELECT 'pinned'       AS left_alone, count(*) FROM agent_definitions
  WHERE deleted_at IS NULL AND COALESCE(default_config->'pin_image_tag','false'::jsonb) = 'true'::jsonb
UNION ALL
SELECT 'snapshot',    count(*) FROM agent_definitions WHERE COALESCE(is_snapshot,false)
UNION ALL
SELECT 'soft-deleted', count(*) FROM agent_definitions WHERE deleted_at IS NOT NULL
UNION ALL
SELECT 'other image', count(*) FROM agent_definitions
  WHERE deleted_at IS NULL AND COALESCE(is_snapshot,false) = false
    AND image_repository IS DISTINCT FROM :'img_repo';

-- After: this must be a single row at the deployed tag.
SELECT COALESCE(image_tag,'(null)') AS tag_after, count(*)
FROM agent_definitions WHERE $SQL_WHERE
GROUP BY 1 ORDER BY 2 DESC;
EOF

rc=$?
if [ $rc -ne 0 ]; then
    echo "ERROR: the image_tag sync FAILED (exit $rc)." >&2
    echo "       Spawned pods still follow the running chassis (066 fix), so this is" >&2
    echo "       a stale RECORD, not a stale pod. Re-run this script when the DB is" >&2
    echo "       reachable: scripts/deploy/update-agent-images.sh" >&2
    exit $rc
fi

echo
echo "agent_definitions.image_tag synced to $IMAGE_TAG"
