-- 098_broaden_cta_recompute_source_unlock.sql — source-unlock the CTA url
-- fields of the components newly covered by applyCTARecompute.
-- Created 2026-07-16. The 091 pattern, applied to the broadened set.
--
-- CONTEXT: the misdirected_cta checker scans every button on every component,
-- but the cta_links_stale repair (rerender_page_sections' applyCTARecompute)
-- only covered hero/call-to-action — detection was broader than repair. The
-- Go side is fixed by broadening ctaFieldNames (commit e10c656f3) to add
-- archetype-grid, archetype-combinations, gauntlet-cta, content-block-about.
-- This migration is the schema half of the same fix: a site_specs.*-sourced
-- url field is re-resolved into resolved_data on EVERY render and merges
-- last, so no recompute or content edit can win until the source is flipped.
--
-- FIELDS FLIPPED (verified against live schema 2026-07-16 — no fallbacks set):
--   archetype-grid.cta_url                 site_specs.identity.primary_cta_url -> renderer
--   archetype-combinations.cta_primary_url site_specs.cta.primary_url          -> renderer
--   archetype-combinations.cta_secondary_url site_specs.cta.secondary_url      -> renderer
--   gauntlet-cta.cta_primary_url           site_specs.cta.primary_url          -> renderer
--   gauntlet-cta.cta_secondary_url         site_specs.cta.secondary_url        -> renderer
--
-- DELIBERATELY NOT TOUCHED:
--   content-block-about.cta_url — source "llm": content edits already win; it
--     only needed recompute coverage (Go side), no schema change.
--   provocations-archive-list.cta_url — static with fallback
--     /tools/arena/index.html is 097b's deliberate pin; not in the set.
--   required flags — a "renderer"/"static" source short-circuits planSection
--     BEFORE required-field handling (plan_sections_action.go ~1187), so
--     required:true on gauntlet-cta/archetype-combinations primary urls
--     cannot block section readiness. Verified in code 2026-07-16.
--   *_label fields still source:"static" on archetype-grid /
--     archetype-combinations — that is the 096 pattern (copy authoring), a
--     separate step-3 content-pass prerequisite, not link repair.
--
-- ORDER OF OPERATIONS (same discipline as 091, plus one sharper edge):
--   vonc's archetype-grid row has EMPTY content_data.cta_url — its rendered
--   /contact.html comes entirely from plan-time site_specs resolution. After
--   this flip, a rerender on a chassis image WITHOUT the broadened
--   ctaFieldNames would drop the button's destination. So:
--     1. chassis image with broadened ctaFieldNames must be live IN-POD first;
--     2. apply this migration;
--     3. immediately dispatch the detected cta_links_stale page_rerender items
--        (about, archetypes, index) — the recompute persists real urls into
--        content_data, closing the empty-field window.
--
-- Reversal: _fleet_098_backup_20260716_cta_components holds the full prior rows.

BEGIN;

CREATE TABLE _fleet_098_backup_20260716_cta_components AS
  SELECT * FROM content_components
  WHERE is_active = true
    AND function IN ('archetype-grid', 'archetype-combinations', 'gauntlet-cta');

-- ── Flip each url field's site_specs.* source -> "renderer" ────────────────
UPDATE content_components
SET input_schema = jsonb_set(input_schema, '{fields,cta_url,source}', '"renderer"'),
    updated_at = NOW()
WHERE is_active = true AND function = 'archetype-grid'
  AND input_schema->'fields'->'cta_url'->>'source' LIKE 'site_specs.%';

UPDATE content_components
SET input_schema = jsonb_set(input_schema, '{fields,cta_primary_url,source}', '"renderer"'),
    updated_at = NOW()
WHERE is_active = true AND function IN ('archetype-combinations', 'gauntlet-cta')
  AND input_schema->'fields'->'cta_primary_url'->>'source' LIKE 'site_specs.%';

UPDATE content_components
SET input_schema = jsonb_set(input_schema, '{fields,cta_secondary_url,source}', '"renderer"'),
    updated_at = NOW()
WHERE is_active = true AND function IN ('archetype-combinations', 'gauntlet-cta')
  AND input_schema->'fields'->'cta_secondary_url'->>'source' LIKE 'site_specs.%';

-- ── Verify: none of the five fields still site_specs-sourced; all renderer ──
DO $$
DECLARE remaining INT; flipped INT;
BEGIN
  SELECT COUNT(*) INTO remaining
  FROM content_components cc, jsonb_each(cc.input_schema->'fields') f
  WHERE cc.is_active = true
    AND cc.function IN ('archetype-grid', 'archetype-combinations', 'gauntlet-cta')
    AND f.key IN ('cta_url', 'cta_primary_url', 'cta_secondary_url')
    AND f.value->>'source' LIKE 'site_specs.%';

  SELECT COUNT(*) INTO flipped
  FROM content_components cc, jsonb_each(cc.input_schema->'fields') f
  WHERE cc.is_active = true
    AND cc.function IN ('archetype-grid', 'archetype-combinations', 'gauntlet-cta')
    AND f.key IN ('cta_url', 'cta_primary_url', 'cta_secondary_url')
    AND f.value->>'source' = 'renderer';

  IF remaining <> 0 THEN
    RAISE EXCEPTION 'verify failed: % url fields still site_specs-sourced (want 0)', remaining;
  END IF;
  IF flipped <> 5 THEN
    RAISE EXCEPTION 'verify failed: % url fields renderer-sourced (want 5) — wrong function names or schema shape?', flipped;
  END IF;
  RAISE NOTICE 'verified: 0 site_specs-sourced CTA url fields remain; % fields now renderer-sourced', flipped;
END $$;

COMMIT;
