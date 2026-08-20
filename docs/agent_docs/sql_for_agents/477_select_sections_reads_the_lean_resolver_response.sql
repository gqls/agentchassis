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
--   git merge-base --is-ancestor 53a8d3c1d <agent-chassis stamp>   # 248 keep (LNK-033)
--   git merge-base --is-ancestor 757a0890a <agent-chassis stamp>   # 299 keep (LNK-034)
--
-- ✅ HOLD DISCHARGED 2026-08-20 — RELEASED on the OWNER'S EXPLICIT DECISION
-- ("apply it with leopardessconsulting.co.uk watched"), which is the authority
-- this file was actually waiting on. The technical precondition was satisfied
-- first, and NOT by the two ancestry checks immediately above:
--
--   THOSE CHECKS WERE UNAVAILABLE. The provenance startup line scrolls (measured
--   2026-08-19: gone from a FULL `kubectl logs` three hours after a roll) and the
--   binary carries ONE stamp string rather than its ancestry — so probing it for
--   either commit returns ABSENT on a binary that certainly contains it. Full
--   trap in LANDMINES.md; generalised as RFC_040 (owner-ratified 2026-08-20).
--
--   SUBSTITUTED, and it is the stronger check: probe the CAPABILITY each keep
--   half provides, on EVERY pod, with a control that must come out absent.
--   Re-run against the CURRENT build after the 2026-08-19 22:26Z roll, because a
--   roll invalidates any earlier verification — agent-chassis **v1.0.1317**,
--   pods c7d6d875b-67cgh and c7d6d875b-x5tgn:
--     storedCTADestinationIsAuthored     PRESENT (both)   <- 248 keep, LNK-033
--     IsAuthoredNonPageCTADestination    PRESENT (both)   <- 299 keep, LNK-034
--     NormalizeTelHref                   PRESENT (both)
--     cta_nonpage_destination            PRESENT (both)
--     cta_nonpage_destination_NOTREAL    absent           <- control
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

-- README rule: every migration touching agent_definitions opens with a snapshot.
SELECT snapshot_agent('page-content-writer',
  '477_select_sections_reads_the_lean_resolver_response: pre-update');

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

-- ── CANARY BASELINE, captured 2026-08-20 BEFORE this file was applied ─────────
-- leopardessconsulting.co.uk authored /contact.html CTAs that MUST survive the
-- first post-apply build — survival is the control that the KEEPS, not luck,
-- made this safe:
--   index/hero                     cta_url           = /contact.html
--   index/call-to-action           primary_cta_url   = /contact.html
--   index/call-to-action           secondary_cta_url = /contact.html
--   how-it-works/hero              cta_url           = /contact.html
--   how-it-works/call-to-action    primary_cta_url   = /contact.html
--   ai-agent-roi-estimator/tool-…  cta_url           = /contact.html
--   tool-ai-vendor-trust-checklist/tool-… cta_url    = /contact.html
-- Full baseline + the diff query: bugfix_299_cta_dials_phone/RUNBOOK, canary section.
--
-- ROLLBACK if any of those move: re-run the UPDATE with the array's first two
-- entries swapped back (stale path first). Config is live immediately, so the
-- revert bites on the next build; pages already rebuilt wrong need a rerender.
-- Exposure is bounded and observable rather than fleet-wide-instant: measured
-- 2026-08-19, the fleet runs 1-7 page-content-writer builds per hour across 5
-- distinct sites.
