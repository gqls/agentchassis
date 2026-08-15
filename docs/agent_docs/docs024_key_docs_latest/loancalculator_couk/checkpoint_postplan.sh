#!/usr/bin/env bash
# Post-planner checkpoint for loancalculator phase 1 (fired 2026-08-15 07:54:41Z DB time).
# Run the moment the new current plan lands. Read-only except the guarded un-defer.
set -euo pipefail
SITE='0162cde4-633e-45e9-8ca6-87a6b2fe1d26'
FIRE='2026-08-15 07:54:00Z'
PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"

echo "== 1. Q2 INVENTION CHECK — page rows created after fire (expect NONE; the two restored rows are 08-14 19:14Z) =="
$PSQL -c "SELECT id, name, url, page_type, status, created_at FROM pages WHERE site_id='$SITE' AND created_at > '$FIRE';"

echo "== 2. The new plan and its shape =="
$PSQL -c "SELECT id, created_at FROM site_plans WHERE site_id='$SITE' AND is_current;"
$PSQL -c "SELECT spp.role, count(*) FROM site_plans sp JOIN site_plan_pages spp ON spp.plan_id=sp.id WHERE sp.site_id='$SITE' AND sp.is_current GROUP BY 1 ORDER BY 2 DESC;"
$PSQL -c "SELECT count(*) AS total_plan_pages FROM site_plans sp JOIN site_plan_pages spp ON spp.plan_id=sp.id WHERE sp.site_id='$SITE' AND sp.is_current;"

echo "== 3. Compositions proposed for the two restored pages (expect >0 sections each) =="
$PSQL -c "SELECT sps.page_name, count(*) AS sections, string_agg(sps.component_name, ', ' ORDER BY sps.ordering) FROM site_plans sp JOIN site_plan_sections sps ON sps.plan_id=sp.id WHERE sp.site_id='$SITE' AND sp.is_current AND sps.page_name IN ('about','guides-index') GROUP BY 1;"

echo "== 4. Tool placements in the plan (the placement gate's output) =="
$PSQL -c "SELECT sps.page_name, sps.component_name FROM site_plans sp JOIN site_plan_sections sps ON sps.plan_id=sp.id WHERE sp.site_id='$SITE' AND sp.is_current AND sps.component_name LIKE 'tool-%' ORDER BY 1;"

echo "== 5. RECOMPOSE tell rows since fire (phase 1 carries no spec — expect NONE) =="
$PSQL -c "SELECT error_code, count(*) FROM agent_error_log WHERE COALESCE(domain,'')='loancalculator.co.uk' AND error_code='RECOMPOSE_INTENT_NOT_REALISED' AND occurred_at > '$FIRE' GROUP BY 1;"

echo "== 6. Page identity md5 of the 27 pre-fire built pages (name|url|page_type|status, restored pair + newcomers excluded) =="
$PSQL -tA -c "SELECT md5(string_agg(name||'|'||url||'|'||page_type||'|'||status, E'\n' ORDER BY name)) FROM pages WHERE site_id='$SITE' AND name NOT IN ('about','guides-index') AND created_at < '$FIRE';"

echo "== 7. UN-DEFER the three owner-mandated items (idempotent; RETURNING shows what moved) =="
$PSQL -c "UPDATE site_work_items SET status='detected' WHERE id IN ('222ecf94-d8f2-4ae1-b689-8ba12e08d953','a52e59d8-3c6c-4f3e-8f2b-121c295af66f','ad289c0e-30c1-4646-8870-20e14887d952') AND status='deferred' RETURNING left(id::text,8), item_key, status;"

echo "== 8. Locked calculator rows still 12/12 untouched =="
$PSQL -tA -c "SELECT count(*) FROM page_components pc JOIN pages p ON pc.page_id=p.id WHERE p.site_id='$SITE' AND pc.locked_by='decompose_20260802_proven_calculators' AND pc.locked_at IS NOT NULL;"

echo "== 9. New work items since fire =="
$PSQL -c "SELECT item_type, status, item_key, created_at FROM site_work_items WHERE site_id='$SITE' AND created_at > '$FIRE' ORDER BY created_at;"
