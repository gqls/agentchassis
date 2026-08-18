-- 477 (_HOLD) — bugs_open/312: point page-content-writer's select_sections at
-- the path the resolver's response ACTUALLY has, so link resolution stops
-- being computed and discarded on every fresh build.
--
-- ⚠⚠ THE INTERLOCK — this is the DANGEROUS one of the three, and the hold is
-- load-bearing, not procedural. The moment this applies, setCTAField's writes
-- go LIVE on every fresh build fleet-wide (config is immediate). On a binary
-- missing EITHER keep half, authored contact and phone CTAs are clobbered at
-- build time from that instant — the traced run (orch 05e3839d, 2026-08-18
-- 10:27) shows the discarded resolver output had already repointed the
-- authored "Get in touch" → /contact.html at a tool. Rename away the _HOLD
-- suffix ONLY after BOTH ancestry checks pass against the LIVE stamp
-- (per service, ask the binary, never git):
--   git merge-base --is-ancestor 53a8d3c1d <agent-chassis stamp>   # 248 keep (LNK-033) — TRUE on v1.0.1310
--   git merge-base --is-ancestor 757a0890a <agent-chassis stamp>   # 299 keep (LNK-034) — awaiting roll
-- and canary the FIRST post-apply build on a site with authored contact CTAs
-- (leopardessconsulting.co.uk, the 248 lane's suggestion: /index and
-- /how-it-works each carry two authored /contact.html CTAs) — diff the CTA
-- urls before/after; survival is the control that the keeps, not luck, made
-- this safe. setCTAField's keeps have NEVER executed in production (their
-- call site's output has never been consumed); they are unit/mutation-tested
-- only, which is exactly why the canary is owed.
--
-- History (LNK-013/LNK-014 + bugs_open/312): this same seam failed silently in
-- BOTH directions. June: envelope nested at response.link_resolution.…, config
-- read response.… → repointed TO the nested path (fixed 2026-06-26). Then the
-- lean-return follow-up LNK-014 itself asked for landed, the envelope became
-- {sections_ready, unresolved} directly under response, and the config went
-- stale again — masked both times by the silent fallback LNK-013 called
-- double-edged. Measured 2026-08-18: the configured path resolves in 0 of 150
-- retained runs; the lean shape is present in 149/150.
--
-- The fallback chain (paths 2..3) is KEPT: resolver failure must still degrade
-- to the un-augmented plan (resolve_links' error_step jumps straight here).
-- The loud-fallback and lockstep-test halves are 312's candidates 2/3 — code,
-- not this migration.

BEGIN;

CREATE TABLE IF NOT EXISTS _backup_477_select_sections AS
  SELECT id, type, default_config, now() AS backed_up_at
    FROM agent_definitions
   WHERE type = 'page-content-writer' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,select_sections,config,fields,sections_ready}',
         '["resolved_links.response.sections_ready",
           "input_data.section_plan.sections_ready",
           "section_plan.sections_ready"]'::jsonb)
 WHERE type = 'page-content-writer' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   -- anchor on the exact stale path, so this cannot double-apply or fire on a
   -- row somebody already fixed differently:
   AND (default_config #> '{workflow,steps,select_sections,config,fields,sections_ready}')
       @> '["resolved_links.response.link_resolution.sections_ready"]'::jsonb;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
    FROM agent_definitions
   WHERE type = 'page-content-writer' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
     AND (default_config #> '{workflow,steps,select_sections,config,fields,sections_ready}')
         @> '["resolved_links.response.sections_ready"]'::jsonb
     AND NOT (default_config #> '{workflow,steps,select_sections,config,fields,sections_ready}')
         @> '["resolved_links.response.link_resolution.sections_ready"]'::jsonb;
  IF n <> 1 THEN
    RAISE EXCEPTION '477: expected exactly 1 page-content-writer reading the lean resolver path (and not the stale one), found % — the config drifted from the 2026-08-18 live read; investigate before applying', n;
  END IF;
END $$;

COMMIT;

-- Post-apply verification (312 §How to verify): the next fresh
-- page-content-writer run must satisfy BOTH —
--   sections_for_render.sections_ready[*].resolved_data == resolved_links.response.sections_ready[*].resolved_data
--   AND the canary site's authored /contact.html CTAs survived.
-- Then bugs_open/312 (and with it 299's producer half) can move toward closed.
