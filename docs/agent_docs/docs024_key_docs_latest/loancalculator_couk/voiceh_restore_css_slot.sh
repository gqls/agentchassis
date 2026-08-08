#!/usr/bin/env bash
# Restore ONE slot on ONE page from the pre-run backup, then redeploy the page.
#
# Written for the CSS-in-a-prose-slot trap (see LANDMINES.md): on this site 8 of
# 51 `prose-*` rows hold the page's <style> block rather than prose, and the voice
# rewrite drops it about half the time — taking the calculator's layout with it
# while every guard in the platform passes.
#
# This is an exact REVERT of a row, not a hand edit: content_data, rendered_html
# and content_hash all come from page_components_bak_20260807_voiceh. That
# distinction matters, because hand-authoring a page_components row is precisely
# what the owner's 2026-08-04 ruling forbids.
#
# Usage: voiceh_restore_css_slot.sh <page-name> [slot]     (slot defaults to prose-0)
set -euo pipefail
SITE='0162cde4-633e-45e9-8ca6-87a6b2fe1d26'
PAGE="$1"; SLOT="${2:-prose-0}"
HERE="$(cd "$(dirname "$0")" && pwd)"

kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 <<SQL
BEGIN;
UPDATE page_components pc
SET content_data = b.content_data, rendered_html = b.rendered_html,
    content_hash = b.content_hash, updated_at = now()
FROM page_components_bak_20260807_voiceh b
WHERE b.page_id = pc.page_id AND b.slot_name = pc.slot_name
  AND pc.slot_name = '$SLOT'
  AND pc.page_id = (SELECT id FROM pages WHERE site_id='$SITE' AND name='$PAGE');

-- A verify block of bare SELECTs cannot stop a COMMIT (ON_ERROR_STOP ignores a
-- non-empty result) — it has to RAISE.
DO \$v\$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM pages p JOIN page_components pc ON pc.page_id=p.id
   JOIN page_components_bak_20260807_voiceh b ON b.page_id=pc.page_id AND b.slot_name=pc.slot_name
  WHERE p.name='$PAGE' AND pc.slot_name='$SLOT'
    AND pc.content_data = b.content_data AND pc.rendered_html IS NOT DISTINCT FROM b.rendered_html;
  IF n <> 1 THEN RAISE EXCEPTION 'restore of %/% did not match the backup (% rows)', '$PAGE', '$SLOT', n; END IF;
  RAISE NOTICE '% % restored from backup', '$PAGE', '$SLOT';
END \$v\$;
COMMIT;
SQL

PID=$(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A -c \
  "SELECT id FROM pages WHERE site_id='$SITE' AND name='$PAGE';")
bash "$HERE/../cta_link_integrity/scripts/049b_deploy_single_page.sh" "$PID" "$SITE" loancalculator.co.uk 2>&1 | grep -E "^(corr=|CORR=)"
echo "restored $PAGE/$SLOT and redeployed — CONFIRM AT THE SERVED PAGE, not here"
