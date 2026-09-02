-- gamedesign.uk — owner rulings 2026-09-02 ~19:20Z:
--   "cancel the article slot" · "we do want a favicon" · (no privacy/terms pages — nothing to do)
-- Applied out of band (psql), per-site, not a migration.
--
-- 1. ARTICLE SLOT. Page `article` (/blog/article.html) was planned with 0 sections; the build
--    handler parked its needs_page at needs_human_review (mark_no_ready_sections). Never
--    deployed, in_header=f, in_footer=f, linked from nowhere (link census 18:00Z). Retire with
--    the platform's own vocabulary: item -> cancelled (with a reason in result), page ->
--    status='archived' (bugs_open/266's ARCHIVED_PAGE_GUARD then refuses any deploy). The
--    site_plan_pages row is LEFT — the plan is written whole by write_site_plan and this lane
--    will not hand-edit a versioned plan. WATCH: if the next planner/discovery rotation re-files
--    work at the archived page, that is bugs_open/356's class and goes in NOTES, not here.
-- 2. FAVICON. /assets/images/favicon.png is referenced by the head and 404s; logo.png is 200.
--    derive_brand_head_assets (asset-deployer) resizes the approved logo — deterministic, no LLM.
--    The undeployed_assets discovery check files exactly this item; filing it by hand with the
--    SAME shape and item_key means the check dedupes against it (idx_swi_dedup) rather than
--    creating a twin. Shape copied from a completed fleet row.
\set ON_ERROR_STOP on
BEGIN;

UPDATE site_work_items
   SET status='cancelled',
       result = COALESCE(result,'{}'::jsonb) || '{"cancelled_by":"gamedesign_uk_rebuild lane","cancelled_at":"2026-09-02T19:20:00Z","reason":"owner ruling 2026-09-02: cancel the article slot — planned with 0 sections, no article exists yet"}'::jsonb
 WHERE site_id='8f17eb73-fc74-4718-8371-b3125bc4e414'
   AND item_type='needs_page' AND status='needs_human_review' AND spec->>'page_name'='article';

UPDATE pages SET status='archived', updated_at=now()
 WHERE id='2ea5d983-b798-4bb2-b30a-5e3047369561' AND site_id='8f17eb73-fc74-4718-8371-b3125bc4e414'
   AND name='article' AND deployed_at IS NULL;

INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary, spec, priority, handler_agent, status, created_by, item_key)
VALUES ('8f17eb73-fc74-4718-8371-b3125bc4e414', 'undeployed_assets', 'build', 'needs_brand_head_assets', 'medium',
  'Derive favicon from the approved logo — head references /assets/images/favicon.png which 404s (owner ruling 2026-09-02: we do want a favicon)',
  '{"mode":"brand_head","check":"undeployed_assets","purpose":"favicon","has_logo":true,"expected_path":"/assets/images/favicon.png","head_references":true,"original_pipeline":"design"}'::jsonb,
  60, 'asset-deployer', 'triaged', 'gamedesign_uk_rebuild lane 2026-09-02', 'needs_brand_head_assets:favicon')
ON CONFLICT DO NOTHING;

COMMIT;

SELECT 'item' AS what, item_type, status, left(spec->>'page_name','10') FROM site_work_items WHERE site_id='8f17eb73-fc74-4718-8371-b3125bc4e414' AND (item_type='needs_brand_head_assets' OR (item_type='needs_page' AND spec->>'page_name'='article'))
UNION ALL SELECT 'page', name, status, build_status FROM pages WHERE id='2ea5d983-b798-4bb2-b30a-5e3047369561';
