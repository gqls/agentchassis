-- 092_enable_link_checks_and_cta_rerender.sql — wire the link-integrity loop
-- Created 2026-07-14. Requires the chassis image carrying misdirected_cta /
-- incomplete_page_group / the cta_links_stale recompute (registered-but-absent
-- check names only warn+skip, so this SQL is safe to land first — the checks
-- simply do nothing until the image is live).
--
-- Three independent wirings:
--   1. completeness-discovery-agent: enable phantom_internal_links (written
--      long ago, enabled nowhere), misdirected_cta, incomplete_page_group.
--   2. page-rerender: accept spec.reason == 'cta_links_stale' into the
--      section re-render path (rerender_page_sections), which gates the CTA
--      recompute on exactly that reason.
--   3. hero / call-to-action CTA text fields: llm_guidance addition so the
--      writer authors CTA copy FOR the destination the resolver chose
--      (resolved_data.cta_target_title / *_cta_target_title), instead of
--      inventing a destination the link doesn't go to.
--
-- All three are idempotent (guarded by NOT-already-present tests).
-- Reversal: _fleet_092_backup_20260714_agentdefs + the 091 component backup
-- (091 must run first; its backup predates this file's component edits).

BEGIN;

CREATE TABLE _fleet_092_backup_20260714_agentdefs AS
  SELECT * FROM agent_definitions
  WHERE type IN ('completeness-discovery-agent', 'page-rerender');

-- ── 1. Enable the three checks (append missing names, preserve order) ──────
UPDATE agent_definitions ad
SET default_config = jsonb_set(
      ad.default_config,
      '{workflow,steps,run_checks,config,checks}',
      (ad.default_config #> '{workflow,steps,run_checks,config,checks}')
      || COALESCE((
           SELECT jsonb_agg(to_jsonb(newcheck))
           FROM unnest(ARRAY['phantom_internal_links',
                             'misdirected_cta',
                             'incomplete_page_group']) AS newcheck
           WHERE NOT (ad.default_config #> '{workflow,steps,run_checks,config,checks}') ? newcheck
         ), '[]'::jsonb)
    ),
    updated_at = NOW()
WHERE ad.type = 'completeness-discovery-agent'
  AND jsonb_typeof(ad.default_config #> '{workflow,steps,run_checks,config,checks}') = 'array';

-- ── 2. page-rerender: gate the section re-render on cta_links_stale too ────
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,check_rerender_mode,config,condition}',
      to_jsonb(
        (default_config #>> '{workflow,steps,check_rerender_mode,config,condition}')
        || ' OR input_data.spec.reason == ''cta_links_stale'''
      )
    ),
    updated_at = NOW()
WHERE type = 'page-rerender'
  AND default_config #>> '{workflow,steps,check_rerender_mode,config,condition}'
      NOT LIKE '%cta_links_stale%';

-- ── 3. CTA text guidance: copy follows the resolved destination ────────────
-- Appends to llm_guidance on whichever CTA text fields each active hero /
-- call-to-action component actually declares (field naming varies across
-- component generations).
DO $$
DECLARE
  comp RECORD;
  fld TEXT;
  guidance_add CONSTANT TEXT :=
    ' When resolved_data provides a companion *_target_title (e.g. '
    || 'cta_target_title for cta_url), the link destination is already fixed: '
    || 'write this CTA text FOR that destination — name it or clearly promise '
    || 'it. Never write copy promising a page the URL does not point to.';
  updated INT := 0;
BEGIN
  FOR comp IN
    SELECT id, input_schema FROM content_components
    WHERE is_active = true AND function IN ('hero', 'call-to-action')
  LOOP
    FOREACH fld IN ARRAY ARRAY['cta_text', 'primary_cta_text', 'secondary_cta_text', 'secondary_cta']
    LOOP
      IF comp.input_schema->'fields' ? fld
         AND COALESCE(comp.input_schema->'fields'->fld->>'llm_guidance', '') NOT LIKE '%_target_title%'
      THEN
        UPDATE content_components
        SET input_schema = jsonb_set(
              input_schema,
              ARRAY['fields', fld, 'llm_guidance'],
              to_jsonb(COALESCE(input_schema->'fields'->fld->>'llm_guidance', '') || guidance_add)
            ),
            updated_at = NOW()
        WHERE id = comp.id;
        updated := updated + 1;
      END IF;
    END LOOP;
  END LOOP;
  RAISE NOTICE '092: CTA text guidance appended to % field instances', updated;
END $$;

-- ── Verify ──────────────────────────────────────────────────────────────────
DO $$
DECLARE checks JSONB; cond TEXT;
BEGIN
  SELECT default_config #> '{workflow,steps,run_checks,config,checks}' INTO checks
  FROM agent_definitions WHERE type = 'completeness-discovery-agent';
  IF checks IS NULL
     OR NOT (checks ? 'phantom_internal_links'
             AND checks ? 'misdirected_cta'
             AND checks ? 'incomplete_page_group') THEN
    RAISE EXCEPTION 'verify failed: completeness checks array = %', checks;
  END IF;

  SELECT default_config #>> '{workflow,steps,check_rerender_mode,config,condition}' INTO cond
  FROM agent_definitions WHERE type = 'page-rerender';
  IF cond IS NULL OR cond NOT LIKE '%cta_links_stale%'
     OR cond NOT LIKE '%image_landed%' OR cond NOT LIKE '%section_data_resolved%' THEN
    RAISE EXCEPTION 'verify failed: page-rerender condition = %', cond;
  END IF;

  RAISE NOTICE 'verified: 3 checks enabled; page-rerender accepts cta_links_stale';
END $$;

COMMIT;
