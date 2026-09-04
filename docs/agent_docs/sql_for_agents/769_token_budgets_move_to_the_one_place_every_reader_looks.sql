-- 769_token_budgets_move_to_the_one_place_every_reader_looks.sql
--
-- bugs_open/257 round 3 — the CONFIGURATION half of owner decision 4, 2026-09-04:
-- "the limits are set in each individual's config, sometimes it has been set in the wrong place
--  and sometimes the agent reads the wrong place, please fix it properly."
--
-- The CODE half (a six-level precedence ladder in platform/orchestration/actions) is inert until
-- the next chassis roll. THIS FILE IS LIVE THE MOMENT IT APPLIES and fixes every misplacement it
-- touches under the CURRENT binary as well as the next one. The two halves are deliberately
-- independent: neither waits for the other, and after both, every reader agrees.
--
-- ─── WHAT IS WRONG, MEASURED 2026-09-04 AGAINST LIVE agent_definitions ────────────────────────
--
-- 208 active non-snapshot agents carry 171 declarations of `max_tokens`, in five shapes:
--
--   .workflow.steps.<step>.config.ai_service.max_tokens        149   the canonical place
--   .max_tokens                            (agent, bare)        10   READ FIRST — see (1)
--   .workflow.steps.<step>.config.max_tokens                     7   READ BY NOBODY — see (2)
--   .ai_service.max_tokens                 (agent, service)      3   fine
--   …<nested loop step>.config.ai_service.max_tokens             2   fine (arrives as StepConfig)
--
-- (1) SHADOWED. `ExecuteLLMPromptAction` reads `agentConfig["max_tokens"]` BEFORE the ai_service
--     block, so an agent carrying a bare root key caps every one of its steps at it. All ten such
--     agents declare 8000 on their single LLM step, in the canonical place, and are pinned to
--     500-2000 by a leftover:
--
--       content-creator-about                  root 2000  vs generate_about_content   8000
--       content-creator-contact (ec74d095)     root 2000  vs generate_content         8000
--       content-creator-contact (92b207b3)     root 1500  vs generate_contact_content 8000
--       content-creator-cta                    root 2000  vs generate_content         8000
--       content-creator-features               root 2000  vs generate_content         8000
--       content-creator-hero                   root 2000  vs generate_hero_content    8000
--       content-creator-hero-without-research  root 1500  vs generate_hero_content    8000
--       content-creator-testimonials           root 2000  vs generate_content         8000
--       content_researcher                     root 1500  vs process                  8000
--       simple-content-writer-with-approval    root  500  vs generate_draft           8000
--
--     ⚠ NONE of these ten has EVER appeared in llm_call_log (all-history, checked 2026-09-04), so
--     this is a latent cap on agents nobody is currently running — not live damage today. It is
--     fixed here because the next caller of one of them would inherit a 500-token content writer
--     and have no way to see why.
--
-- (2) UNREAD. Nothing looks at a STEP's bare key. site-adoption-agent's four steps each carry an
--     `ai_service` block holding model/provider/api_key_env_var — which ARE read — and
--     `max_tokens` as a SIBLING of that block, one brace outside where it was meant to go:
--
--       analyze_site              asks 32000, runs at the root 16000  (asked for double, got half)
--       derive_content_direction  asks  6000, runs at the root 16000
--       classify_archetype        asks  4000, runs at the root 16000
--       generate_design_intent    asks  4000, runs at the root 16000
--
--     This agent IS live: 8 calls in the 14 days to 2026-09-04, every one at 16000.
--
-- ─── WHAT THIS MIGRATION DOES, AND WHAT CHANGES AT THE WIRE ───────────────────────────────────
--
-- It MOVES each misplaced number into the canonical `ai_service` block at the SAME level. It
-- invents no numbers and changes no declared value; every figure below is the one an operator
-- already wrote. After it:
--
--   * the ten agents' single LLM step sends 8000 (its own declaration) instead of 500-2000.
--   * site-adoption-agent's four steps send 32000 / 6000 / 4000 / 4000 instead of 16000.
--
-- COST AND HEADROOM, stated because two of those four go DOWN. Largest observed output on each of
-- the four in the 14 days to 2026-09-04: analyze_site 843, derive_content_direction 4167,
-- classify_archetype 1017, generate_design_intent 1493 — so the tightest new ceiling
-- (derive_content_direction, 6000) still carries 1.4x its observed maximum, and the other three
-- 2.7x or better. Zero truncations on this agent all-history. If one of them starts truncating,
-- the cause is the operator's declared number and the lever is one jsonb write, which is the whole
-- point of moving it somewhere a reader honours.
--
-- ─── WHAT IS DELIBERATELY NOT TOUCHED ─────────────────────────────────────────────────────────
--
-- html-developer-chunked's three bare step keys (generate_content 12000, generate_structure 4000,
-- generate_styles 8000) STAY WHERE THEY ARE. They are the one place in the estate where the bare
-- spelling is the ONLY one that works: `getMaxTokens` (html_actions.go) reads the bare key and
-- ignores the ai_service block, so moving them TODAY would silently give all three the hardcoded
-- 16000 default. The code half fixes that reader; migration 770 (_HOLD) moves them, by hand, once
-- a chassis image carrying it is live. This is an ordering constraint that genuinely exists, and
-- it is named rather than assumed.
--
-- Rollback: 769_..._ROLLBACK.sql (restores every key to the place it was in).

BEGIN;

-- Guard 1: the ten shadowing agents must still be exactly the ten this file names. A census does
-- not go wrong, it goes STALE BY ADDITION — an eleventh agent given a bare root key since
-- 2026-09-04 would be silently skipped by an UPDATE that names ids, and silently mis-described by
-- this header. Abort and re-census rather than half-fix.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     AND default_config ? 'max_tokens';
  IF n <> 10 THEN
    RAISE EXCEPTION '769 ABORT: % agents carry a bare root max_tokens, this file was written for 10. Re-run the census before applying.', n;
  END IF;
END $$;

-- Guard 2: same, for the step-level bare key. 7 = site-adoption-agent's 4 + html-developer-chunked's 3.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
    FROM agent_definitions a, LATERAL jsonb_each(COALESCE(a.default_config->'workflow'->'steps','{}'::jsonb)) s
   WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
     AND jsonb_typeof(a.default_config->'workflow'->'steps')='object'
     AND s.value->'config' ? 'max_tokens';
  IF n <> 7 THEN
    RAISE EXCEPTION '769 ABORT: % steps carry a bare step-level max_tokens, this file was written for 7. Re-run the census before applying.', n;
  END IF;
END $$;

-- Guard 3: no agent declares BOTH spellings at one level. If one did, moving the bare key would
-- OVERWRITE a canonical declaration, silently replacing an operator's number with another of
-- their numbers — the only shape in which this migration could destroy information.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     AND default_config ? 'max_tokens' AND default_config->'ai_service' ? 'max_tokens';
  IF n > 0 THEN
    RAISE EXCEPTION '769 ABORT: % agents declare max_tokens at BOTH root spellings; moving would clobber one.', n;
  END IF;

  SELECT count(*) INTO n
    FROM agent_definitions a, LATERAL jsonb_each(COALESCE(a.default_config->'workflow'->'steps','{}'::jsonb)) s
   WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
     AND jsonb_typeof(a.default_config->'workflow'->'steps')='object'
     AND s.value->'config' ? 'max_tokens' AND s.value->'config'->'ai_service' ? 'max_tokens';
  IF n > 0 THEN
    RAISE EXCEPTION '769 ABORT: % steps declare max_tokens at BOTH step spellings; moving would clobber one.', n;
  END IF;
END $$;

-- Snapshot every agent this file rewrites, before it rewrites it. House convention
-- for any migration touching agent_definitions, and the practical rollback path: the
-- ROLLBACK file restores the values, these rows restore the whole definition if the
-- jsonb surgery below turns out to have taken something with it.
DO $$
DECLARE t text;
BEGIN
  FOR t IN
    SELECT DISTINCT type FROM agent_definitions
     WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
       AND (default_config ? 'max_tokens' OR type = 'site-adoption-agent')
  LOOP
    PERFORM snapshot_agent(t, '769_token_budgets_move_to_the_one_place_every_reader_looks.sql: pre-update');
  END LOOP;
END $$;

-- ── (1) the agent-level bare key moves into the agent's ai_service block ──────────────────────
-- Every one of the ten already HAS a root ai_service block (checked 2026-09-04), so this adds a
-- key to an existing object rather than creating one; jsonb_set with create_if_missing does both.
UPDATE agent_definitions
   SET default_config =
         (default_config - 'max_tokens')
         || jsonb_build_object('ai_service',
              COALESCE(default_config->'ai_service','{}'::jsonb)
              || jsonb_build_object('max_tokens', default_config->'max_tokens')),
       updated_at = now()
 WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
   AND default_config ? 'max_tokens';

-- ── (2) site-adoption-agent's four step-level bare keys move into each step's ai_service ──────
-- Named by agent rather than applied to all 7, because html-developer-chunked's three must NOT
-- move yet (see the header). Rebuilding the steps object key-by-key is the only way to address a
-- varying step name in one statement.
UPDATE agent_definitions a
   SET default_config = jsonb_set(a.default_config, '{workflow,steps}', (
         SELECT jsonb_object_agg(s.key,
                  CASE WHEN s.value->'config' ? 'max_tokens'
                       -- ⚠ ((s.value->'config') - 'max_tokens'), not s.value->'config' - 'max_tokens':
                       -- PostgreSQL binds subtraction TIGHTER than ->, so the unparenthesised form
                       -- parses as s.value -> ('config' - 'max_tokens') and fails with
                       -- "operator is not unique: unknown - unknown". Caught by executing this file
                       -- with COMMIT replaced by ROLLBACK before applying it.
                       THEN jsonb_set(s.value, '{config}',
                              ((s.value->'config') - 'max_tokens')
                              || jsonb_build_object('ai_service',
                                   COALESCE(s.value->'config'->'ai_service','{}'::jsonb)
                                   || jsonb_build_object('max_tokens', s.value->'config'->'max_tokens')))
                       ELSE s.value END)
           FROM jsonb_each(a.default_config->'workflow'->'steps') s)),
       updated_at = now()
 WHERE a.type = 'site-adoption-agent'
   AND a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL;

-- ── VERIFY: a DO block, not a SELECT ─────────────────────────────────────────────────────────
-- ON_ERROR_STOP ignores a non-empty result set, so a verify block made of SELECTs cannot stop the
-- COMMIT (LANDMINES.md). These RAISE.
DO $$
DECLARE bare_root int; bare_step int; moved int;
BEGIN
  SELECT count(*) INTO bare_root FROM agent_definitions
   WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     AND default_config ? 'max_tokens';
  IF bare_root <> 0 THEN
    RAISE EXCEPTION '769 VERIFY FAILED: % agents still carry a bare root max_tokens', bare_root;
  END IF;

  -- The ten values must have SURVIVED the move, not vanished: 10 agents must now declare a root
  -- ai_service.max_tokens where 3 did before.
  SELECT count(*) INTO moved FROM agent_definitions
   WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     AND default_config->'ai_service' ? 'max_tokens';
  IF moved <> 13 THEN
    RAISE EXCEPTION '769 VERIFY FAILED: % agents declare root ai_service.max_tokens, expected 13 (3 pre-existing + 10 moved)', moved;
  END IF;

  -- site-adoption-agent: four steps, each now declaring its number in the canonical place, and
  -- the numbers themselves unchanged.
  SELECT count(*) INTO bare_step
    FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
   WHERE a.type='site-adoption-agent' AND a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
     AND s.value->'config' ? 'max_tokens';
  IF bare_step <> 0 THEN
    RAISE EXCEPTION '769 VERIFY FAILED: site-adoption-agent still has % bare step-level max_tokens', bare_step;
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM agent_definitions a
     WHERE a.type='site-adoption-agent' AND a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
       AND a.default_config #>> '{workflow,steps,analyze_site,config,ai_service,max_tokens}' = '32000'
       AND a.default_config #>> '{workflow,steps,derive_content_direction,config,ai_service,max_tokens}' = '6000'
       AND a.default_config #>> '{workflow,steps,classify_archetype,config,ai_service,max_tokens}' = '4000'
       AND a.default_config #>> '{workflow,steps,generate_design_intent,config,ai_service,max_tokens}' = '4000')
  THEN
    RAISE EXCEPTION '769 VERIFY FAILED: site-adoption-agent budgets are not 32000/6000/4000/4000 in ai_service';
  END IF;

  -- And html-developer-chunked's three must be UNTOUCHED. A migration that quietly took them too
  -- would look identical to this one succeeding.
  SELECT count(*) INTO bare_step
    FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
   WHERE a.type='html-developer-chunked' AND a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
     AND s.value->'config' ? 'max_tokens';
  IF bare_step <> 3 THEN
    RAISE EXCEPTION '769 VERIFY FAILED: html-developer-chunked has % bare step budgets, expected 3 left in place for migration 770', bare_step;
  END IF;

  RAISE NOTICE '769 OK: 10 root keys and 4 step keys moved into ai_service; html-developer-chunked left for 770_HOLD';
END $$;

COMMIT;
