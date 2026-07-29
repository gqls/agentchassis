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
# The base64 goes to psql on STDIN, never as an argv — a megabyte of argv
# blows ARG_MAX (the lesson recorded in webdesign_publish_assets.sh).
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
#   ./scripts/amend-asset.sh relojistas.com logo corrected-logo.png \
#       --note "bugs_open/131: stored logo was a two-up spec sheet; cropped to the light-variant wordmark"
#
# Watch:   printed at the end. Result lands in site_work_items.result.
# Verify:  curl the presigned_url from the result (or assets.url) — then LOOK
#          at the image. Every mechanical signal can be green while the
#          picture is wrong; that is how this workstream started.
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

# Identifiers reach the SQL inside single quotes — keep them to a charset that
# cannot escape one. The note is free text; escape it by doubling quotes.
case "$DOMAIN$ASSET_KEY$PURPOSE" in
  *[!a-zA-Z0-9._-]*) echo "domain/asset_key/purpose may only contain [a-zA-Z0-9._-]" >&2; exit 1 ;;
esac
NOTE_SQL=$(printf '%s' "$NOTE" | sed "s/'/''/g")

SIZE=$(wc -c < "$FILE")
if [ "$SIZE" -gt $((10 * 1024 * 1024)) ]; then
  echo "file is ${SIZE} bytes — the action refuses anything over 10MB" >&2; exit 1
fi
SHA=$(sha256sum "$FILE" | cut -d' ' -f1)
B64=$(base64 -w0 "$FILE")

PURPOSE_SQL="NULL"; [ -n "$PURPOSE" ] && PURPOSE_SQL="'$PURPOSE'"
NOTE_VAL="NULL";    [ -n "$NOTE" ]    && NOTE_VAL="'$NOTE_SQL'"

SQL=$(cat <<EOSQL
\\set ON_ERROR_STOP on
BEGIN;

DO \$guard\$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM sites WHERE domain = '$DOMAIN') THEN
        RAISE EXCEPTION 'no site with domain $DOMAIN';
    END IF;
END
\$guard\$;

WITH staged AS (
    INSERT INTO asset_ingest_staging (site_id, asset_key, purpose, content, sha256, note, created_by)
    SELECT id, '$ASSET_KEY', $PURPOSE_SQL, decode('$B64', 'base64'), '$SHA', $NOTE_VAL, '$CREATED_BY'
      FROM sites WHERE domain = '$DOMAIN'
    RETURNING id, site_id
)
INSERT INTO site_work_items (site_id, source, item_type, severity, summary, spec,
                             handler_agent, status, created_by, priority, pipeline,
                             item_key, triaged_at)
SELECT site_id, 'operator', 'amend_asset', 'medium',
       'Amend asset ''$ASSET_KEY'' with operator-supplied bytes ($SIZE bytes, sha $SHA)',
       jsonb_build_object('mode', 'ingest_upload', 'staging_id', id),
       'asset-deployer', 'triaged', '$CREATED_BY', 70, 'build',
       'amend_asset:$ASSET_KEY', now()
  FROM staged
RETURNING id AS work_item_id, site_id, spec->>'staging_id' AS staging_id;

COMMIT;
EOSQL
)

if [ -n "$DRY_RUN" ]; then
  printf '%s\n' "$SQL" | sed "s/decode('[^']*'/decode('<BASE64 elided, $SIZE file bytes>'/"
  echo "-- dry run: nothing executed" >&2
  exit 0
fi

printf '%s\n' "$SQL" | "${PSQL[@]}"

cat >&2 <<EOF

Staged. Dispatch picks it up within ~2 minutes. Watch:
  ${PSQL[*]} -c "SELECT status, attempt_count, left(coalesce(error,'-'),80), result->'response'->'ingest_result' FROM site_work_items WHERE item_key='amend_asset:$ASSET_KEY' AND site_id=(SELECT id FROM sites WHERE domain='$DOMAIN') ORDER BY created_at DESC LIMIT 1;"

Then VERIFY: curl the presigned_url out of ingest_result and LOOK at the image.
EOF
