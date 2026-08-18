-- 476 (_HOLD) — bug 299: arm the CTA destination stamp on internal-link-resolver.
--
-- ⚠ HELD until the image carrying commit 757a0890a is POD-VERIFIED live on
-- agent-chassis. A config key naming behaviour the binary lacks reads as
-- applied and does nothing (the 380 trap) — worse than either state alone.
-- Rename away the _HOLD suffix only after the provenance check passes.
--
-- What it arms: stamp_cta_destination_guidance (opt-in, default OFF in code —
-- the 2026-08-02 owner-ruling shape). When ON, resolve_internal_links appends
-- "Destination (fixed): <title>. Write this CTA's text to name or clearly
-- promise this destination…" to the paired LABEL field's llm_field_specs
-- description — the pipe the page-content-writer prompt already renders — for
-- BOTH page and non-page destinations. Closes the producer half of 299:
-- measured 2026-08-18, the *_target_title VALUE reached 0 of 182 sampled
-- writer prompts; the writer was told (llm_guidance) to consult a datum it
-- could not see, and invented destinations instead.
--
-- Step key verified against the LIVE row 2026-08-18 (seed-vs-live drift trap):
-- internal-link-resolver's step is resolve_links, action resolve_internal_links.
--
-- NOTE: the stamp's effect on real prompts is only observable AFTER
-- bugs_open/312's repoint (477) makes the resolver's output reach the writer
-- pipe at all — but arming it before 477 is safe and correct: the resolver
-- runs today, its sections_ready is simply discarded downstream.
-- Verify at llm_call_log.prompt_rendered: a value-shaped
-- "Destination (fixed):" occurrence, with the pre-arm 0/182 as the baseline.

BEGIN;

CREATE TABLE IF NOT EXISTS _backup_476_cta_stamp AS
  SELECT id, type, default_config, now() AS backed_up_at
    FROM agent_definitions
   WHERE type = 'internal-link-resolver' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,resolve_links,config,stamp_cta_destination_guidance}',
         'true'::jsonb)
 WHERE type = 'internal-link-resolver' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #> '{workflow,steps,resolve_links,config}' IS NOT NULL
   AND default_config #>> '{workflow,steps,resolve_links,action}' = 'resolve_internal_links';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
    FROM agent_definitions
   WHERE type = 'internal-link-resolver' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
     AND (default_config #> '{workflow,steps,resolve_links,config,stamp_cta_destination_guidance}') = 'true'::jsonb;
  IF n <> 1 THEN
    RAISE EXCEPTION '476: expected exactly 1 armed internal-link-resolver, found % — the step key or action drifted from the 2026-08-18 live read; investigate before applying', n;
  END IF;
END $$;

COMMIT;
