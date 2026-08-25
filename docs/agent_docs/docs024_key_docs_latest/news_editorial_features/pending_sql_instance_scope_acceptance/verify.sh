#!/usr/bin/env bash
# news_editorial_features — verify the instance-scope acceptance, in the ONLY order
# that is safe (283 lane, 2026-08-24: three repairs complete, stored bytes correct,
# one page served the old version for hours).
#
# Read-only. Run after A_unlock_and_dispatch.sql; re-run until step 3 passes.
# Do NOT run B_relock.sql or C_close_lock_items.sql until this exits 0.

PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"
RH=https://robot-hands.com/insights/robot-demand-step-change.html
DO=https://dartsonline.com/insights/darts-calendar-density.html
fail=0

echo "=== 0. STYLESHEET CONTROL (08-21 handoff 9.1 — an audit against a broken"
echo "       stylesheet produces false PASSES; healthy is 25-27 KB) ==="
for d in robot-hands.com dartsonline.com; do
  b=$(curl -sL "https://$d/assets/css/styles.css" | wc -c)
  if [ "$b" -ge 25000 ] && [ "$b" -le 28000 ]; then echo "  OK   $d $b B"
  else echo "  FAIL $d $b B — out of range, nothing below is meaningful"; fail=1; fi
done

echo "=== 1. work items (necessary, NOT sufficient) ==="
$PSQL -c "SELECT left(item_key,58) AS item_key, status, attempt_count, left(COALESCE(error,''),60) AS err
            FROM site_work_items
           WHERE item_key IN ('section_edit_tplfix_0c297014-1d0d-444d-b673-a75f1ee706fc',
                              'section_edit_tplfix_cccc84fa-4b17-4dfb-aecd-fb776d220c33')
           ORDER BY created_at DESC LIMIT 4;"

echo "=== 2. stored rendered_html (necessary, NOT sufficient — stored is not served) ==="
# agent_writable is the column that decides, and its absence here is what hid the
# 2026-08-25 v1 defect: lock_type/locked_by read as 'unlocked' while locked_at
# still held the row shut. Between A and B expect t; after B expect f.
$PSQL -c "SELECT slot_name, locked_at, lock_type, locked_by,
                 (locked_at IS NULL OR (lock_type='timed' AND lock_expires_at IS NOT NULL
                                        AND lock_expires_at < NOW())) AS agent_writable,
                 rendered_html LIKE '%id=\"c-evidence-timeseries\"%' AS has_new_id,
                 rendered_html LIKE '%id=\"\"%'                      AS has_empty_id,
                 length(rendered_html) AS bytes,
                 component_version_id IS NOT NULL AS stamped, updated_at
            FROM page_components
           WHERE id IN ('d344585f-6f79-4a18-a1fe-3116b68a4c52',
                        'ea6b4ca7-7717-4e29-ae1c-88844040b0d2');"

echo "=== 3. THE ONE THAT DECIDES — the SERVED page ==="
echo "       baseline 2026-08-24: rh 94351 B, do 92883 B; predicted rh 94349, do 92872"
for pair in "$RH|rh|94351" "$DO|do|92883"; do
  url="${pair%%|*}"; rest="${pair#*|}"; tag="${rest%%|*}"; base="${rest##*|}"
  body=$(curl -sL "$url"); b=$(printf '%s' "$body" | wc -c)
  newid=$(printf '%s' "$body" | grep -c 'id="c-evidence-timeseries"')
  oldid=$(printf '%s' "$body" | grep -c 'id="evidence-timeseries-')
  empty=$(printf '%s' "$body" | grep -c 'id=""')
  # known-string control: the page must still BE the article, not an error body
  known=$(printf '%s' "$body" | grep -c 'class="ev-ts__inner"')
  printf "  %s bytes=%s (baseline %s, delta %+d)  new_id=%s old_id=%s empty_id=%s ev-ts_body=%s\n" \
         "$tag" "$b" "$base" "$((b-base))" "$newid" "$oldid" "$empty" "$known"
  [ "$newid" -eq 1 ] || { echo "    -> NOT YET SERVED (or failed). If 1+2 passed, this is a PUBLISH LAG, not a failed edit: wait, re-run. Do NOT re-dispatch."; fail=1; }
  [ "$oldid" -eq 0 ] || { echo "    -> still serving the OLD id"; fail=1; }
  [ "$empty" -eq 0 ] || { echo "    -> EMPTY ID SERVED — stop, this is the reEmptyElementID defect class"; fail=1; }
  [ "$known" -ge 1 ] || { echo "    -> chart body MISSING — a 200 is not an artefact"; fail=1; }
done

echo
if [ "$fail" -eq 0 ]; then
  echo "ALL PASS — now run B_relock.sql, re-run this to confirm the lock, then C_close_lock_items.sql"
else
  echo "NOT READY — do not re-lock, do not close the items, do not re-dispatch."
fi
exit $fail
