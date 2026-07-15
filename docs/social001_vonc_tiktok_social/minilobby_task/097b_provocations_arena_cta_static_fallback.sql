-- 097b_provocations_arena_cta_static_fallback.sql — correct 097's mechanism
-- Created 2026-07-15.
--
-- 097 flipped provocations-archive-list.cta_url static -> renderer and authored
-- /tools/arena/index.html into content_data. It DID NOT hold: this component has
-- no CTA-recompute path (unlike hero/call-to-action), so on the rerender the
-- resolver saw source=renderer WITH a fallback and wrote the fallback
-- (/index.html) into resolved_data, which merged over content_data and reverted
-- it. Lesson: "renderer + fallback" is not a content_data-wins source unless a
-- recompute path also writes the field.
--
-- Deterministic fix: keep the field source=static (a static field always renders
-- its schema value) and point the fallback itself at the Arena. vonc-only
-- component (1 instance), so this sets the day's-CTA default to the Arena for
-- vonc's provocations archive — coherent with the "Enter today's Arena" label.
--
-- Reversal: _vonc_097_backup_20260715_component holds the pre-097 schema.

BEGIN;

UPDATE content_components
SET input_schema = jsonb_set(
      jsonb_set(input_schema, '{fields,cta_url,source}', '"static"'),
      '{fields,cta_url,fallback}', '"/tools/arena/index.html"'),
    updated_at = NOW()
WHERE function = 'provocations-archive-list';

DO $$
DECLARE src TEXT; fb TEXT;
BEGIN
  SELECT input_schema->'fields'->'cta_url'->>'source',
         input_schema->'fields'->'cta_url'->>'fallback'
    INTO src, fb
  FROM content_components WHERE function='provocations-archive-list';
  IF src <> 'static' OR fb <> '/tools/arena/index.html' THEN
    RAISE EXCEPTION 'verify failed: source=% fallback=%', src, fb;
  END IF;
  RAISE NOTICE 'verified: provocations-archive-list cta_url static, fallback -> Arena';
END $$;

COMMIT;
