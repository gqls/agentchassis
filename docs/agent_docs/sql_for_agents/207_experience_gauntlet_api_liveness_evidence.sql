-- 207_experience_gauntlet_api_liveness_evidence.sql
--
-- Follow-on to 197: the D1-REVISED API-contract sentence still described
-- API_BASE as "a hostname on apis.uk, owner-named" — a plan-time placeholder,
-- written before the backend existed. As of 2026-07-25 it is no longer a
-- placeholder: tools-api is BUILT (feature-builder PR #3, merged), DEPLOYED
-- (island VM, Route B1), and VERIFIED with a full real round-trip through the
-- public internet — genuine AI-generated content, not a mock. This is the
-- 196-contracts-rule "failing §5 criterion" case turned real: the access
-- paths now resolve for real, so the plan should be written knowing that,
-- not hedging it.
--
-- One surgical replace() inside compose.config.prompt_template: the API_BASE
-- clause gains the concrete base URL + a compact liveness citation (the exact
-- verified call shapes and one real response fragment), so the feasibility
-- seat can ground its judgement in evidence instead of trusting an assertion.
--
-- Workstream: gauntlet_dead_cta. Config-only, live immediately.
-- ROLLBACK: snapshot below.

BEGIN;

SELECT snapshot_agent('experience-planner', '207_experience_gauntlet_api_liveness_evidence: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,compose,config,prompt_template}',
         to_jsonb(
           replace(
             default_config->'workflow'->'steps'->'compose'->'config'->>'prompt_template',
             $OLD$Base URL is a configurable constant API_BASE (a hostname on apis.uk, owner-named; the built JS carries it as a literal constant).$OLD$,
             $NEW$Base URL is API_BASE = https://tools.apis.uk — DECIDED and LIVE, not a placeholder (the built JS carries it as a literal constant). VERIFIED LIVE 2026-07-25: a full real round-trip was run against this exact URL through the public internet (Cloudflare -> tunnel -> island VM -> tools-api), not a mock or a local test. POST https://tools.apis.uk/api/v1/tools/gauntlet/round with Origin: https://vonc.com -> HTTP 200, real body incl. round_id and today's actual provocation JSON. POST .../position with {round_id, position_text} -> HTTP 200, a genuine Anthropic-generated counter_position and challenge (not templated). POST .../defend with {round_id, defence_text} -> HTTP 200, a genuine Anthropic-generated verdict and reasons; two full rounds completed this way, both verdict="opponent wins" (the AI judges honestly, not a pushover). Denied-origin POSTs return 403; a missing round_id returns 404; a malformed/oversized body is rejected; the AI-unavailable path returns 503 with a clean JSON error body (Cloudflare replaces raw 502 bodies, so 503 is the status that survives to the browser — write the front-end's degraded-mode handling against 503, not 502). Every one of these was exercised for real, not asserted.$NEW$
           )
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

  IF position('https://tools.apis.uk — DECIDED and LIVE' in rp) = 0 THEN
    RAISE EXCEPTION '207: liveness evidence did not land';
  END IF;
  IF position('a hostname on apis.uk, owner-named' in rp) > 0 THEN
    RAISE EXCEPTION '207: old placeholder text still present';
  END IF;
  IF position('two full rounds completed this way, both verdict="opponent wins"' in rp) = 0 THEN
    RAISE EXCEPTION '207: round-trip evidence missing';
  END IF;
  -- untouched neighbours must survive (197's own guard set, re-checked)
  IF position('REAL DEBATE against an AI opponent' in rp) = 0 THEN
    RAISE EXCEPTION '207: D1-REVISED (197) was disturbed';
  END IF;
  IF position('/api/v1/tools/gauntlet/position' in rp) = 0 THEN
    RAISE EXCEPTION '207: API access paths were disturbed';
  END IF;
  IF position('hollow shell' in rp) = 0 THEN
    RAISE EXCEPTION '207: diagnosis line 3 (197) was disturbed';
  END IF;
  IF position('D2 (REVISED 2026-07-18' in rp) = 0 THEN
    RAISE EXCEPTION '207: D2 was disturbed';
  END IF;
  IF position('LENGTH AND QUOTING DISCIPLINE' in rp) = 0 THEN
    RAISE EXCEPTION '207: length discipline (176) was disturbed';
  END IF;
  IF position('<!-- END EXPERIENCE_PLAN -->' in rp) = 0 THEN
    RAISE EXCEPTION '207: output-format trailer was disturbed';
  END IF;
END $$;

COMMIT;
