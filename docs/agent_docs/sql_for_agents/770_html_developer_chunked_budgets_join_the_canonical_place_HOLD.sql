-- 770_html_developer_chunked_budgets_join_the_canonical_place_HOLD.sql
--
-- ⚠ _HOLD: THIS FILE HAS A REAL ORDERING CONSTRAINT AND MUST NOT BE APPLIED BY THE RUNNER.
--
-- APPLY BY HAND ONLY ONCE a chassis image containing the bugs_open/257 round-3 ladder is LIVE.
-- The condition is checkable, not a judgement:
--
--   kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
--   git merge-base --is-ancestor <the round-3 commit> <the sha that line reports> && echo SAFE
--
-- (If the provenance line has scrolled — it is a STARTUP line — probe the binary for a KNOWN sha
-- with a must-be-absent control, per CLAUDE.md. An empty grep means "not in range", not "absent".)
--
-- WHY THE ORDER MATTERS. html-developer-chunked's three steps declare their budget as a BARE
-- step-level key. That is the non-canonical spelling everywhere else in the estate — but here it
-- is the ONLY one that works: `getMaxTokens` (platform/orchestration/actions/html_actions.go)
-- reads `config["max_tokens"]` and does not look at the step's `ai_service` block at all. Move
-- these three today and all three silently take that function's hardcoded 16000 default instead:
--
--   generate_content    12000  ->  16000
--   generate_structure   4000  ->  16000
--   generate_styles      8000  ->  16000
--
-- The round-3 ladder makes `getMaxTokens` read the canonical spelling first and the bare one
-- second, so after it is live BOTH placements work and this move is behaviour-neutral. That is
-- the entire reason for the hold, and it is why 769 left these three alone while moving the
-- other fourteen.
--
-- [MEASURED 2026-09-04] html-developer-chunked has no rows in llm_call_log in the last 14 days,
-- so applying it early would not have been visible in traffic either — which is exactly the kind
-- of quiet that makes an ordering mistake permanent.
--
-- Rollback: 770_..._ROLLBACK.sql

BEGIN;

-- Guard: refuse if the three are not where this file expects them.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
    FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
   WHERE a.type='html-developer-chunked' AND a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
     AND s.value->'config' ? 'max_tokens';
  IF n <> 3 THEN
    RAISE EXCEPTION '770 ABORT: html-developer-chunked has % bare step budgets, expected 3 — re-census before applying', n;
  END IF;
END $$;

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
 WHERE a.type='html-developer-chunked'
   AND a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM agent_definitions a
     WHERE a.type='html-developer-chunked' AND a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
       AND a.default_config #>> '{workflow,steps,generate_content,config,ai_service,max_tokens}'   = '12000'
       AND a.default_config #>> '{workflow,steps,generate_structure,config,ai_service,max_tokens}' = '4000'
       AND a.default_config #>> '{workflow,steps,generate_styles,config,ai_service,max_tokens}'    = '8000')
  THEN
    RAISE EXCEPTION '770 VERIFY FAILED: the three budgets are not 12000/4000/8000 under ai_service';
  END IF;
  RAISE NOTICE '770 OK: every max_tokens in the fleet now sits in an ai_service block';
END $$;

COMMIT;
