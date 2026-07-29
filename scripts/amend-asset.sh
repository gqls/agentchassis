#!/bin/bash
# ============================================================================
# amend-asset.sh — supply corrected image bytes for a site asset, through the
# platform, with no storage credentials on the operator side.
# ============================================================================
# The asset-amend path (bugs_open/131 og-card slug). Bytes travel:
#   file → sha256 + base64 → psql stdin → asset_ingest_staging (BYTEA)
#     → site_work_items ({mode:'ingest_upload', staging_id}) — same transaction
#     → build-dispatch (~2 min tick) → asset-deployer → ingest_staged_asset
#     → S3 (new key) + assets row amended in place.
#
# PARAMETERISATION (council round 1, constitution seat): every operator-
# controlled value reaches the SQL as a psql VARIABLE (:'var' — psql quotes
# it safely), never by shell interpolation into the statement text. The ONE
# exception is the base64 payload: it cannot ride argv (-v is argv; a
# megabyte of argv blows ARG_MAX — the lesson recorded in
# webdesign_publish_assets.sh), so it stays inline on stdin, made inert by
# CONSTRUCTION — the script refuses to proceed unless the blob matches
# ^[A-Za-z0-9+/=]+$, an alphabet that cannot contain a quote, a backslash,
# or a colon. Validation, not escaping discipline.
#
# The work item's dedup key is amend_asset:<asset_key> — a second amend for
# the same key while one is still in flight is REFUSED by idx_swi_dedup.
# That is the dedup working, not a bug; wait for the first to finish.
#
# A LOCKED assets row (locked_at set) is refused by the action by design —
# approved assets are never machine-overwritten. Clearing a lock is a
# deliberate, separate human step; do not fold it into this script.
#
# Usage:
#   ./scripts/amend-asset.sh <domain> <asset_key> <file> [--purpose <p>] [--note "<why>"] [--dry-run]
# Example:
#   ./scripts/amend-asset.sh gaswholesalers.com logo approved-logo.png \
#       --note "bugs_open/131: stored logo was a nine-up contact sheet; owner-approved replacement"
#
# Watch:   printed at the end. Result lands in site_work_items.result.
# Verify:  curl the presigned_url from the result — then LOOK at the image.
#          Every mechanical signal can be green while the picture is wrong;
#          that is how this workstream started.
# ============================================================================
set -euo pipefail

PSQL=(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db)
CREATED_BY="${CREATED_BY:-operator}"

DOMAIN="${1:?usage: amend-asset.sh <domain> <asset_key> <file> [--purpose p] [--note text] [--dry-run]}"
ASSET_KEY="${2:?asset_key required (e.g. logo)}"
FILE="${3:?file required}"
shift 3

PURPOSE="" NOTE="" DRY_RUN=""
while [ $# -gt 0 ]; do
  case "$1" in
    --purpose) PURPOSE="${2:?}"; shift 2 ;;
    --note)    NOTE="${2:?}";    shift 2 ;;
    --dry-run) DRY_RUN=1;        shift   ;;
    *) echo "unknown flag: $1" >&2; exit 2 ;;
  esac
done

[ -f "$FILE" ] || { echo "no such file: $FILE" >&2; exit 1; }

SIZE=$(wc -c < "$FILE")
if [ "$SIZE" -gt $((10 * 1024 * 1024)) ]; then
  echo "file is ${SIZE} bytes — the action refuses anything over 10MB" >&2; exit 1
fi
SHA=$(sha256sum "$FILE" | cut -d' ' -f1)
B64=$(base64 -w0 "$FILE")

# The blob is the only inline literal — refuse anything outside the base64
# alphabet (see header). Everything else travels as a psql variable.
case "$B64" in
  *[!A-Za-z0-9+/=]*) echo "base64 output contains bytes outside [A-Za-z0-9+/=] — refusing" >&2; exit 1 ;;
  "") echo "empty payload" >&2; exit 1 ;;
esac

SUMMARY="Amend asset '${ASSET_KEY}' with operator-supplied bytes (${SIZE} bytes, sha ${SHA})"
ITEM_KEY="amend_asset:${ASSET_KEY}"

SQL=$(cat <<EOSQL
\\set ON_ERROR_STOP on
BEGIN;

DO \$guard\$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM sites WHERE domain = current_setting('amend.domain')) THEN
        RAISE EXCEPTION 'no site with domain %', current_setting('amend.domain');
    END IF;
END
\$guard\$;

WITH staged AS (
    INSERT INTO asset_ingest_staging (site_id, asset_key, purpose, content, sha256, note, created_by)
    SELECT id, :'asset_key', NULLIF(:'purpose', ''), decode('$B64', 'base64'), :'sha', NULLIF(:'note', ''), :'created_by'
      FROM sites WHERE domain = :'domain'
    RETURNING id, site_id
)
INSERT INTO site_work_items (site_id, source, item_type, severity, summary, spec,
                             handler_agent, status, created_by, priority, pipeline,
                             item_key, triaged_at)
SELECT site_id, 'manual', 'amend_asset', 'medium', :'summary',
       jsonb_build_object('mode', 'ingest_upload', 'staging_id', id),
       'asset-deployer', 'triaged', :'created_by', 70, 'build',
       :'item_key', now()
  FROM staged
RETURNING id AS work_item_id, site_id, spec->>'staging_id' AS staging_id;

COMMIT;
EOSQL
)

PSQL_ARGS=(
  -v domain="$DOMAIN" -v asset_key="$ASSET_KEY" -v purpose="$PURPOSE"
  -v note="$NOTE" -v created_by="$CREATED_BY" -v sha="$SHA"
  -v summary="$SUMMARY" -v item_key="$ITEM_KEY"
)

if [ -n "$DRY_RUN" ]; then
  printf '%s\n' "$SQL" | sed "s/decode('[^']*'/decode('<BASE64 elided, $SIZE file bytes>'/"
  echo "-- dry run: nothing executed. psql vars: domain=$DOMAIN asset_key=$ASSET_KEY purpose=$PURPOSE created_by=$CREATED_BY" >&2
  exit 0
fi

# The guard block cannot read psql variables (server-side), so ship the domain
# once via set_config on the same connection, parameterised the same way.
{ printf "SELECT set_config('amend.domain', :'domain', false);\n"; printf '%s\n' "$SQL"; } \
  | "${PSQL[@]}" "${PSQL_ARGS[@]}"

cat >&2 <<EOF

Staged. Dispatch picks it up within ~2 minutes. Watch:
  ${PSQL[*]} -c "SELECT status, attempt_count, left(coalesce(error,'-'),80), result->'response'->'ingest_result' FROM site_work_items WHERE item_key='$ITEM_KEY' AND site_id=(SELECT id FROM sites WHERE domain='$DOMAIN') ORDER BY created_at DESC LIMIT 1;"

Then VERIFY: curl the presigned_url out of ingest_result and LOOK at the image.
EOF
