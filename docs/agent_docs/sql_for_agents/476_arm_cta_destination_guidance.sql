-- 476 (_HOLD) — bug 299: arm the CTA destination stamp on internal-link-resolver.
--
-- ✅ HOLD DISCHARGED 2026-08-19 — RELEASED (_HOLD suffix off). The hold guarded
-- against the 380 trap: a config key naming behaviour the binary lacks reads as
-- applied and does nothing, which is worse than either state alone. So the check
-- that discharges it is not "which commit shipped" but "does the binary know
-- this key" — probed directly on the live pod (agent-chassis-5ddd9744-86nqf,
-- image v1.0.1316, 2026-08-19), because the build-provenance startup line had
-- already scrolled out of the retained log:
--     stamp_cta_destination_guidance   PRESENT in /proc/1/exe
--     "Destination (fixed)"            PRESENT   <- the phrase it writes
-- Both halves present: the key the config sets, and the literal the code emits
-- when it acts on it. See 475's header for the fuller note on why a capability
-- probe replaced the commit-ancestry check here.
--
-- ⚠⚠ READ THIS BEFORE YOU "VERIFY" THIS MIGRATION AND CONCLUDE IT FAILED.
-- Arming this is CORRECT and INERT AT THE SAME TIME. The resolver runs today and
-- stamps the guidance into its sections_ready — and select_sections then DISCARDS
-- that whole object (bugs_open/312; measured 2026-08-19: of 48 retained runs, 26
-- minted *_target_title and 0 survived into sections_for_render). So the
-- post-apply check below WILL READ ZERO until migration 477 repoints the path,
-- and that zero means "inert as designed", NOT "the stamp is broken".
-- The demand control that distinguishes them: 477 unapplied ⇒ expect 0; 477
-- applied ⇒ expect non-zero on the next fresh page-content-writer run. Do not
-- record a pre-477 zero as evidence about this migration.
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

-- README rule: every migration touching agent_definitions opens with a snapshot.
SELECT snapshot_agent('internal-link-resolver',
  '476_arm_cta_destination_guidance: pre-update');

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
