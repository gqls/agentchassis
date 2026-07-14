-- 091_cta_url_schema_source_flip.sql — remove the fleet-wide CTA misdirection writer
-- Created 2026-07-14.
--
-- ROOT CAUSE: hero.cta_url / hero.secondary_cta_url and
-- call-to-action.primary_cta_url / secondary_cta_url carry
-- "source": "pages.contact" / "pages.services" in content_components
-- .input_schema. The source resolver (plan_sections_action.go, case "pages")
-- is a literal page-name -> URL lookup with zero copy-awareness, and the
-- resolved value is written into resolved_data UNCONDITIONALLY on every
-- render — resolved_data merges LAST, so no content_data edit can ever win
-- (this is why migration 089's Gauntlet retargets silently reverted).
-- Blast radius at investigation time: 75/75 call-to-action + 68/69 hero CTA
-- instances across 8/10 sites all pointing at /contact.html.
--
-- FIX: flip those four fields' source to "renderer" wherever it is currently
-- a pages.* lookup. Verified against the field loop (plan_sections_action.go
-- ~1186-1201): a "renderer" source resolves to nil at plan time, on_missing
-- (default skip_field) omits the field, and the only remaining writers are
-- the internal-link-resolver (build time) and rerender_page_sections'
-- cta_links_stale recompute — both of which only ever write REAL, validated,
-- non-excluded pages. Existing authored content_data values keep rendering.
-- type / on_missing / llm_guidance / fallback are untouched.
--
-- SAFE: nothing changes in deployed HTML until a page's next render. Sites
-- whose stored content_data holds /contact.html keep serving it until a
-- cta_links_stale rerender or full rebuild recomputes it (chassis image with
-- the recompute must be live BEFORE dispatching those rerenders).
--
-- Reversal: _fleet_cta_backup_20260714_components holds the full prior rows.

BEGIN;

CREATE TABLE _fleet_cta_backup_20260714_components AS
  SELECT * FROM content_components
  WHERE is_active = true
    AND function IN ('hero', 'call-to-action');

-- ── Flip each URL field's pages.* source -> "renderer" ─────────────────────
UPDATE content_components
SET input_schema = jsonb_set(input_schema, '{fields,cta_url,source}', '"renderer"'),
    updated_at = NOW()
WHERE is_active = true AND function = 'hero'
  AND input_schema->'fields'->'cta_url'->>'source' LIKE 'pages.%';

UPDATE content_components
SET input_schema = jsonb_set(input_schema, '{fields,secondary_cta_url,source}', '"renderer"'),
    updated_at = NOW()
WHERE is_active = true AND function = 'hero'
  AND input_schema->'fields'->'secondary_cta_url'->>'source' LIKE 'pages.%';

UPDATE content_components
SET input_schema = jsonb_set(input_schema, '{fields,primary_cta_url,source}', '"renderer"'),
    updated_at = NOW()
WHERE is_active = true AND function = 'call-to-action'
  AND input_schema->'fields'->'primary_cta_url'->>'source' LIKE 'pages.%';

UPDATE content_components
SET input_schema = jsonb_set(input_schema, '{fields,secondary_cta_url,source}', '"renderer"'),
    updated_at = NOW()
WHERE is_active = true AND function = 'call-to-action'
  AND input_schema->'fields'->'secondary_cta_url'->>'source' LIKE 'pages.%';

-- ── Verify: no active hero/call-to-action URL field still page-name-sourced ─
DO $$
DECLARE remaining INT; flipped INT;
BEGIN
  SELECT COUNT(*) INTO remaining
  FROM content_components cc, jsonb_each(cc.input_schema->'fields') f
  WHERE cc.is_active = true
    AND cc.function IN ('hero', 'call-to-action')
    AND f.key IN ('cta_url', 'secondary_cta_url', 'primary_cta_url')
    AND f.value->>'source' LIKE 'pages.%';

  SELECT COUNT(*) INTO flipped
  FROM content_components cc, jsonb_each(cc.input_schema->'fields') f
  WHERE cc.is_active = true
    AND cc.function IN ('hero', 'call-to-action')
    AND f.key IN ('cta_url', 'secondary_cta_url', 'primary_cta_url')
    AND f.value->>'source' = 'renderer';

  IF remaining <> 0 THEN
    RAISE EXCEPTION 'verify failed: % active CTA url fields still pages.*-sourced (want 0)', remaining;
  END IF;
  IF flipped = 0 THEN
    RAISE EXCEPTION 'verify failed: 0 fields flipped to renderer — wrong function names or schema shape?';
  END IF;
  RAISE NOTICE 'verified: 0 pages.*-sourced CTA url fields remain; % fields now renderer-sourced', flipped;
END $$;

COMMIT;
