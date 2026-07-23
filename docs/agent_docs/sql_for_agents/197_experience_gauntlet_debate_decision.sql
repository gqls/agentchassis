-- 197_experience_gauntlet_debate_decision.sql
--
-- Injects the owner's 2026-07-23 ruling into the experience-planner's compose
-- prompt: the vonc Gauntlet is now a REAL AI-debate tool backed by the fleet's
-- first public backend API (tools-api, built via the feature-builder), not the
-- earlier minimal-real client-side-only cut (old D1). Decisions live in the
-- compose prompt by design (seed 167 lineage: "Decisions already made" block) —
-- this is the channel a new owner ruling travels through.
--
-- Three surgical replace()s inside compose.config.prompt_template ONLY:
--   (1) D1 -> D1-REVISED (debate opponent + backend + honest degraded mode)
--   (2) diagnosis line 3 -> current gauntlet state (2026-07-22 partial remedy,
--       still hollow; owner-verified) so the planner designs from reality
--   (3) "no server" sentence -> the one-exception API contract with the EXACT
--       greenfield access paths (round/position/defend), which the 196 contracts
--       rule requires the plan to pin verbatim + back with failing §5 criteria.
--
-- Workstream: gauntlet_dead_cta (vonc AI-competitor debate). Config-only, live
-- immediately. ROLLBACK: snapshot below.

BEGIN;

SELECT snapshot_agent('experience-planner', '197_experience_gauntlet_debate_decision: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,compose,config,prompt_template}',
         to_jsonb(
           replace(replace(replace(
             default_config->'workflow'->'steps'->'compose'->'config'->>'prompt_template',
             $T1OLD$D1 — the Gauntlet ships as a MINIMAL-REAL playable round: a timed round against today's provocation, client-side scoring, NO leaderboard, zero fabricated numbers. Only demote it to an honest "coming soon" if it genuinely cannot be built both honest AND small — and then it must be ABSENT or labelled coming-soon, never simulated.$T1OLD$,
             $T1NEW$D1 (REVISED 2026-07-23 by owner ruling — supersedes the earlier minimal-real client-side-only cut) — the Gauntlet ships as a REAL DEBATE against an AI opponent, backed by the fleet's first live backend API (built in the platform concurrently with this plan). The flow: the visitor reads today's provocation, files a written Position, the AI opponent returns a REAL opposing Position plus a challenge to the visitor's, the visitor defends within the clock, and the AI returns an honest verdict with reasons on whether the take held up. The three objective checkboxes become REAL self-checking steps bound to these events (position filed / defence sent / verdict received before the clock runs out). The opponent is honestly labelled an AI competitor while the site has no human traffic. NO leaderboard in this round; zero fabricated numbers. DEGRADED MODE: if the API is unreachable the tool says so honestly and disables its controls (e.g. "the AI opponent is offline — try again later") — never simulate, never fall back to a mock.$T1NEW$),
             $T2OLD$3. /tools/gauntlet/index.html — a mock: both CTAs href="#", fabricated stats (12,847 competitors, 94,210 challenges, a "Live" leaderboard of invented users). build_status=needs_rebuild.$T2OLD$,
             $T2NEW$3. /tools/gauntlet/index.html — HISTORY: shipped as a mock (both CTAs href="#", fabricated stats, an invented "Live" leaderboard). PARTIALLY REMEDIED 2026-07-22: the dead hrefs and fabrications were removed — but the tool remains a hollow shell whose controls produce no outcome a visitor can perceive (owner-verified: the primary CTA starts an already-visible timer; checkboxes tick a progress bar bound to nothing). build_status=needs_rebuild. The REAL fix is D1's debate build.$T2NEW$),
             $T3OLD$Static site: no server; everything dynamic is client-side JS reading this JSON.$T3OLD$,
             $T3NEW$Static site with ONE exception (D1): the Gauntlet debate calls the fleet's tools API. Base URL is a configurable constant API_BASE (a hostname on apis.uk, owner-named; the built JS carries it as a literal constant). The EXACT access paths the new gauntlet JS must implement — greenfield pairs: pin them verbatim in §3 and back EACH with a §5 criterion that fails if it is never built: POST {API_BASE}/api/v1/tools/gauntlet/round -> {round_id, provocation:{headline, body}}; POST {API_BASE}/api/v1/tools/gauntlet/position with {round_id, position_text} -> {counter_position, challenge}; POST {API_BASE}/api/v1/tools/gauntlet/defend with {round_id, defence_text} -> {verdict, reasons}. Everything else dynamic remains client-side JS reading this JSON.$T3NEW$)
         ),
         true)
 WHERE type='experience-planner'
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE rp text;
BEGIN
  SELECT default_config->'workflow'->'steps'->'compose'->'config'->>'prompt_template' INTO rp
    FROM agent_definitions
   WHERE type='experience-planner' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF position('REAL DEBATE against an AI opponent' in rp) = 0 THEN
    RAISE EXCEPTION '197: D1-REVISED did not land';
  END IF;
  IF position('client-side scoring, NO leaderboard, zero fabricated numbers' in rp) > 0 THEN
    RAISE EXCEPTION '197: old D1 text still present';
  END IF;
  IF position('/api/v1/tools/gauntlet/position' in rp) = 0 THEN
    RAISE EXCEPTION '197: API access paths missing';
  END IF;
  IF position('hollow shell' in rp) = 0 THEN
    RAISE EXCEPTION '197: diagnosis line 3 not updated';
  END IF;
  IF position('DEGRADED MODE' in rp) = 0 THEN
    RAISE EXCEPTION '197: degraded-mode honesty rule missing';
  END IF;
  -- untouched neighbours must survive
  IF position('D2 (REVISED 2026-07-18' in rp) = 0 THEN
    RAISE EXCEPTION '197: D2 was disturbed';
  END IF;
  IF position('LENGTH AND QUOTING DISCIPLINE' in rp) = 0 THEN
    RAISE EXCEPTION '197: length discipline (176) was disturbed';
  END IF;
  IF position('<!-- END EXPERIENCE_PLAN -->' in rp) = 0 THEN
    RAISE EXCEPTION '197: output-format trailer was disturbed';
  END IF;
END $$;

COMMIT;
