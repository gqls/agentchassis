-- 471 — widen the finance directory's discovery, because the mortgage kind is
-- starved by SOURCE SHAPE, not by anything wrong with the pipeline.
--
-- THE MEASUREMENT THAT SETS THIS FILE'S DIRECTION (2026-08-18, all-history over
-- directory_claims joined to directory_entities — a query that could easily have
-- come out the other way and did not):
--
--   kind             source_domain                   claims  firms
--   savings-provider www.gov.uk                          14     12   <-- ONE page, twelve firms
--   health-insurer   www.mytribeinsurance.co.uk           8      7
--   health-insurer   www.drewberryinsurance.co.uk         7      7   <-- two pages, ten firms
--   mortgage-lender  www.fca.org.uk                       2      1
--   mortgage-lender  www.kbra.com                         2      1
--   mortgage-lender  www.familybuildingsociety.co.uk      2      1
--   mortgage-lender  www.mansfieldbs.co.uk                1      1   <-- four pages, TWO firms
--
-- Every high-yield source in the register's history is a MULTI-FIRM ENUMERATION
-- that states a fact per firm. Every single-firm page yields one firm. The
-- mortgage kind has never once had an enumeration page in its scrape set — its
-- four slots on 2026-08-18 went to two market-overview pages that named firms
-- without stating quotable facts about them (ukfinance.org.uk's largest-lenders
-- table, bsa.org.uk's HOMEPAGE rather than its member list) and two single-society
-- pages. Candidates that run: 2. Registered: 1.
--
-- So the three changes here, in order of expected effect:
--
--   1. MORE SLOTS. max_scrapes 4 -> 10, num_results 10 -> 20, max_snippets 5 -> 8.
--      Size checked against bugs_closed/062 (a batch_scrape reply over the broker's
--      ~1 MB max.message.bytes is dropped and the caller starves): the 2026-08-18
--      mortgage run carried 85 kB of scrape_results for 4 URLs with
--      formats=[markdown] + only_main_content, so 10 URLs projects to ~210 kB —
--      a ~5x margin. Measured, not assumed: orchestration a5ba225c.
--
--   2. AIM AT ENUMERATIONS. Four new mortgage-lender discovery tasks whose queries
--      hunt member lists and named-firm cohorts rather than the market. The
--      existing single weekly query is kept — it is the one that found both
--      current lenders.
--
--   3. STOP CALLING THE HIGH-YIELD SHAPE WEAK. The prompt currently says
--      "Third-party listicles are weak — prefer the register or the firm", which
--      is the opposite of what the register's own history shows. Narrowed to the
--      distinction that actually matters: a page ENUMERATING firms with a fact
--      per firm is a fine source; a page whose content is RANKINGS or PRICING is
--      not (and the competitor comparison sites stay excluded outright at
--      prepare_urls, which this does not touch).
--
-- NOT CHANGED, deliberately: the verbatim-citation gate, the closed field
-- vocabulary, the no-prices rule, the named-firm rule (migration 423), and the
-- exclude_domains list. This file widens INTAKE only. Every fact still has to
-- survive the same re-fetch.
--
-- Queries are all < 200 bytes — web_search drops a >=200-char query as "likely an
-- LLM error message" and the run then fails pointing at config keys, not length
-- (SEED_finance_directory_scheduled_tasks.sql's own warning; B4 run 2 died of it).

BEGIN;

-- ── 1. slots ───────────────────────────────────────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(jsonb_set(jsonb_set(
        default_config,
        '{workflow,steps,search_web,config,num_results}',  '20'::jsonb, true),
        '{workflow,steps,prepare_urls,config,max_scrapes}', '10'::jsonb, true),
        '{workflow,steps,prepare_urls,config,max_snippets}', '8'::jsonb, true),
    updated_at = now()
WHERE type = 'finance-directory-researcher'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── 3. the listicle line ───────────────────────────────────────────────────
-- Anchored on the verbatim sentence. The guard below aborts if it did not match
-- exactly one row, so a reworded prompt cannot be silently left unchanged.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,extract_claims,config,prompt_template}',
        to_jsonb(replace(
            default_config #>> '{workflow,steps,extract_claims,config,prompt_template}',
            'Third-party listicles are weak — prefer the register or the firm.',
            'A page that ENUMERATES many firms and states a fact about each is a strong source and historically the highest-yield one — a trade-body member list, a regulator or government list of authorised firms, or a specialist broker''s provider round-up. What is weak is a page whose substance is RANKINGS or PRICING rather than firm facts. Prefer, in order: the regulator or a government list, a trade-body member list, the firm''s own regulatory page, then a specialist round-up.'
        )),
        true),
    updated_at = now()
WHERE type = 'finance-directory-researcher'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,extract_claims,config,prompt_template}'
      LIKE '%Third-party listicles are weak%';

-- ── 2. four enumeration-shaped mortgage queries ────────────────────────────
INSERT INTO scheduled_tasks (
    name, description, interval_seconds, target_agent_type, target_topic,
    input_data, concurrency_group, max_concurrent, enabled, timeout_seconds)
SELECT v.name, v.descr, 604800, 'finance-directory-researcher',
       'system.agent.generic.requests',
       jsonb_build_object('research_query', v.q),
       'finance-directory-discovery', 1, true, 1200
FROM (VALUES
  ('mortgage-lender-directory-discovery-bsa',
   'Finance directory (DIR-001): mortgage lenders via the Building Societies Association member list — an enumeration page, the shape that yields many firms per scrape.',
   'Building Societies Association member list: named UK building societies and the mortgage ranges each offers'),
  ('mortgage-lender-directory-discovery-adverse',
   'Finance directory (DIR-001): specialist adverse-credit mortgage lenders. Directly serves adversecreditmortgage.co.uk, whose directory page has nothing to list.',
   'UK specialist mortgage lenders for adverse credit: named firms lending to borrowers with CCJs, defaults or a DMP'),
  ('mortgage-lender-directory-discovery-btl',
   'Finance directory (DIR-001): buy-to-let lenders — a cohort the single existing query has never surfaced.',
   'UK buy-to-let mortgage lenders: named banks and building societies and the landlord mortgages each offers'),
  ('mortgage-lender-directory-discovery-fscs',
   'Finance directory (DIR-001): FSCS/government protected-firm lists. gov.uk is the single highest-yield source in the register''s history (12 savings firms from one page).',
   'FSCS protected UK banks and building societies: named firms and the mortgage lending each of them does')
) AS v(name, descr, q)
WHERE NOT EXISTS (SELECT 1 FROM scheduled_tasks s WHERE s.name = v.name);

-- ── verification, as DO/RAISE ──────────────────────────────────────────────
-- A verify block made of SELECTs cannot stop the COMMIT: psql's ON_ERROR_STOP
-- ignores a non-empty result set, so the migration "verifies" and commits anyway
-- (LANDMINES). Every check below therefore RAISES.
DO $$
DECLARE
    n_scr   int; n_res int; n_snip int;
    n_prompt int; n_tasks int; n_bad_len int;
BEGIN
    SELECT (default_config #>> '{workflow,steps,prepare_urls,config,max_scrapes}')::int,
           (default_config #>> '{workflow,steps,search_web,config,num_results}')::int,
           (default_config #>> '{workflow,steps,prepare_urls,config,max_snippets}')::int
      INTO n_scr, n_res, n_snip
      FROM agent_definitions
     WHERE type='finance-directory-researcher' AND is_active
       AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    IF n_scr IS DISTINCT FROM 10 THEN RAISE EXCEPTION 'max_scrapes is %, expected 10', n_scr; END IF;
    IF n_res IS DISTINCT FROM 20 THEN RAISE EXCEPTION 'num_results is %, expected 20', n_res; END IF;
    IF n_snip IS DISTINCT FROM 8 THEN RAISE EXCEPTION 'max_snippets is %, expected 8', n_snip; END IF;

    -- the old sentence must be GONE and the new one PRESENT: checking only one
    -- side passes on a prompt that was never loaded.
    SELECT count(*) INTO n_prompt FROM agent_definitions
     WHERE type='finance-directory-researcher' AND is_active
       AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
       AND default_config #>> '{workflow,steps,extract_claims,config,prompt_template}' NOT LIKE '%listicles are weak%'
       AND default_config #>> '{workflow,steps,extract_claims,config,prompt_template}' LIKE '%ENUMERATES many firms%';
    IF n_prompt <> 1 THEN RAISE EXCEPTION 'prompt replacement did not take (matching rows: %)', n_prompt; END IF;

    SELECT count(*) INTO n_tasks FROM scheduled_tasks
     WHERE name LIKE 'mortgage-lender-directory-discovery%' AND enabled;
    IF n_tasks <> 5 THEN RAISE EXCEPTION 'expected 5 enabled mortgage discovery tasks, found %', n_tasks; END IF;

    -- the 200-byte trap, asserted rather than trusted
    SELECT count(*) INTO n_bad_len FROM scheduled_tasks
     WHERE target_agent_type='finance-directory-researcher'
       AND octet_length(input_data->>'research_query') >= 200;
    IF n_bad_len > 0 THEN RAISE EXCEPTION '% discovery query/queries are >= 200 bytes and web_search will drop them', n_bad_len; END IF;

    RAISE NOTICE '471 OK — scrapes=%, results=%, snippets=%, mortgage tasks=%', n_scr, n_res, n_snip, n_tasks;
END $$;

COMMIT;
