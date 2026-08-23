-- 572 — the five one-shot BUILD-CHAIN handoffs declare `recurrence_expected: true`, so a
--       re-submitted domain queues work instead of reporting success and doing nothing.
--       CONFIG ONLY — live on apply, no image roll needed.
--
--       bugs_open/326. The key this sets is already declared in create_work_item's
--       ConfigKeys and already honoured by the running binary, which is why this half can
--       ship ahead of the Go half rather than behind it.
--
-- ============================================================================
-- WHAT THE PROBLEM IS, in plain terms
-- ============================================================================
-- A customer's build is one chain of five work items, each agent filing the next:
--
--   domain-submitter          -> needs_domain_research    (item_key `research_<domain>`)
--   domain-research-classifier-> needs_vertical_research  (`vertical_research_<domain>`)
--   vertical-exemplar-researcher -> needs_strategy        (`strategy_<domain>`)
--   domain-strategist         -> needs_briefing           (`briefing_<domain>`)
--   build-briefing-agent      -> needs_site_plan          (`site_plan_<domain>`)
--
-- Every one of those keys is `<prefix>_<domain>`: fixed for the life of the domain. That is
-- correct and deliberate — it is what stops two simultaneous submissions of one domain
-- running two builds over the same site.
--
-- The problem is a DIFFERENT mechanism that keys on the same thing. `writeWorkItem` runs an
-- anti-churn brake before it inserts: it counts rows on the same (site_id, item_key) that
-- reached `complete` or `failed` in the last 7 days, and if the newest is under three hours
-- old it holds the new request back. Until the Go half of this fix ships, "holds back"
-- means "returns no row and no error" — the caller is told `deduped: true`, which is the
-- same thing it would be told if the work were already queued and in hand.
--
-- So: submit a domain, watch the build fail halfway, re-submit within three hours, and the
-- orchestration completes with no error having queued nothing at all. Measured on
-- loanzy.uk, 2026-08-18: three submissions at 12:53:00Z, 15:21:17Z and 20:16:12Z (dated
-- from `site_specs`, aspect='submission', which the submitter writes BEFORE this step); the
-- middle one landed 2h28m after the first item was created and left no row anywhere.
--
-- ============================================================================
-- WHY `recurrence_expected` IS THE RIGHT KEY, AND NOT A WEAKENING OF DEDUP
-- ============================================================================
-- The brake was built for a DETECTED DEFECT: a check keeps finding the same fault, a fixer
-- keeps reporting `complete`, and something must break the loop. For that, a repeat is
-- evidence the fix is not working.
--
-- These five steps file ACTION REQUESTS instead. Nothing detected anything; an agent
-- finished its stage and is asking for the next one. A `complete` predecessor here means
-- the previous run SUCCEEDED, and asking again is the normal course of business — a new
-- customer build of a domain we built before, or a retry of one that broke.
--
-- `recurrence_expected` is exactly that distinction, and it was built for exactly this
-- failure in bugs_closed/024 (a tool re-render request that the same brake killed). Its own
-- doc comment states the limit of what it waives, and it is the reason this migration is
-- safe:
--
--     "Dedup is NOT waived by this flag — idx_swi_dedup still refuses a second OPEN item
--      for the same (site_id, item_key). Only the anti-churn heuristics are skipped."
--
-- So after this migration:
--   * re-submitting a domain whose previous build is FINISHED  -> queues a fresh item  ✅
--   * re-submitting while the previous build is STILL RUNNING  -> still refused by the
--     unique index, exactly as today                                                    ✅
--
-- That second line is the negative control bugs_open/326 itself demands, and it is enforced
-- by the database, not by this config.
--
-- ============================================================================
-- SCOPE — five steps, and deliberately not the other sixteen (owner ruling 2026-08-23)
-- ============================================================================
-- Twenty-one live `create_work_item` steps carry an `item_key_prefix`; nineteen of them
-- declare nothing either way, and the whole build chain is among those nineteen
-- (measured 2026-08-23, recursive walk over active non-snapshot definitions).
--
-- This migration touches ONLY the five above. The rest divide into action requests
-- (imagery, rewrite, re-render) and genuine detections (claims-auditor,
-- component-quality-auditor, tool-auditor, tool-suggester), and classifying them means
-- making calls inside lanes this change does not own. One of them is a worked example of
-- why that restraint is right: `claims-auditor.request_claims_review` NEEDS the counter,
-- because its revalidator-close loop writes `complete` into the two-strike window by design
-- (work_items_common.go, the workItemRevalidatableStatuses commentary). Setting it `true`
-- from here would have broken that silently.
--
-- The offline census `config-key-audit --undeclared-recurrence` (shipped with this fix) is
-- what surfaces the remaining sixteen, by name, for their owning lanes to rule on.
--
-- ============================================================================
-- APPLY
-- ============================================================================
BEGIN;

SELECT snapshot_agent('domain-submitter',
    'bugs_open/326: declare recurrence_expected on the build-chain handoff');
SELECT snapshot_agent('domain-research-classifier',
    'bugs_open/326: declare recurrence_expected on the build-chain handoff');
SELECT snapshot_agent('vertical-exemplar-researcher',
    'bugs_open/326: declare recurrence_expected on the build-chain handoff');
SELECT snapshot_agent('domain-strategist',
    'bugs_open/326: declare recurrence_expected on the build-chain handoff');
SELECT snapshot_agent('build-briefing-agent',
    'bugs_open/326: declare recurrence_expected on the build-chain handoff');

-- domain-submitter.create_research_item — the FRONT DOOR. A customer re-submitting after a
-- failed build is the motivating case: this is the step that reported COMPLETED and queued
-- nothing.
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
       '{workflow,steps,create_research_item,config,recurrence_expected}', 'true'::jsonb),
       updated_at = NOW()
 WHERE type = 'domain-submitter'
   AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

-- domain-research-classifier.create_next_item — a stage handoff. The classifier having
-- researched this domain before is not a reason to refuse to research it again.
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
       '{workflow,steps,create_next_item,config,recurrence_expected}', 'true'::jsonb),
       updated_at = NOW()
 WHERE type = 'domain-research-classifier'
   AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

-- vertical-exemplar-researcher.create_next_item — a stage handoff, same reasoning.
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
       '{workflow,steps,create_next_item,config,recurrence_expected}', 'true'::jsonb),
       updated_at = NOW()
 WHERE type = 'vertical-exemplar-researcher'
   AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

-- domain-strategist.create_next_item — a stage handoff, same reasoning.
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
       '{workflow,steps,create_next_item,config,recurrence_expected}', 'true'::jsonb),
       updated_at = NOW()
 WHERE type = 'domain-strategist'
   AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

-- build-briefing-agent.create_next_item — the last handoff before the site plan, and the
-- one furthest into a build, so the one most likely to be re-run after a partial failure.
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
       '{workflow,steps,create_next_item,config,recurrence_expected}', 'true'::jsonb),
       updated_at = NOW()
 WHERE type = 'build-briefing-agent'
   AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

-- ============================================================================
-- VERIFY — a DO block that RAISES, not a list of SELECTs.
--
-- ON_ERROR_STOP does not fire on a non-empty result set, so a verify block written as
-- plain SELECTs cannot stop the COMMIT: it prints its complaint and commits anyway. This
-- one aborts. (bugs_closed/RFC_006's recorded trap.)
-- ============================================================================
DO $$
DECLARE
    declared_true INT;
    step_name     TEXT;
BEGIN
    SELECT count(*) INTO declared_true
    FROM agent_definitions d
    CROSS JOIN LATERAL (
        SELECT CASE d.type WHEN 'domain-submitter' THEN 'create_research_item'
                           ELSE 'create_next_item' END AS s
    ) x
    WHERE d.type IN ('domain-submitter','domain-research-classifier',
                     'vertical-exemplar-researcher','domain-strategist','build-briefing-agent')
      AND d.is_active AND COALESCE(d.is_snapshot,false) = false AND d.deleted_at IS NULL
      AND d.default_config->'workflow'->'steps'->x.s->'config'->>'recurrence_expected' = 'true';

    IF declared_true <> 5 THEN
        RAISE EXCEPTION '572 FAILED: % of 5 build-chain steps carry recurrence_expected=true', declared_true;
    END IF;

    -- The step each UPDATE targeted must still BE a create_work_item step. jsonb_set on a
    -- path that does not exist creates it silently, so a renamed step would leave this
    -- migration "successful" having written a key onto nothing.
    FOR step_name IN
        SELECT d.type FROM agent_definitions d
        CROSS JOIN LATERAL (
            SELECT CASE d.type WHEN 'domain-submitter' THEN 'create_research_item'
                               ELSE 'create_next_item' END AS s
        ) x
        WHERE d.type IN ('domain-submitter','domain-research-classifier',
                         'vertical-exemplar-researcher','domain-strategist','build-briefing-agent')
          AND d.is_active AND COALESCE(d.is_snapshot,false) = false AND d.deleted_at IS NULL
          AND COALESCE(d.default_config->'workflow'->'steps'->x.s->>'action','') <> 'create_work_item'
    LOOP
        RAISE EXCEPTION '572 FAILED: %s targeted step is not a create_work_item step', step_name;
    END LOOP;

    RAISE NOTICE '572 OK: 5 build-chain create_work_item steps declare recurrence_expected=true';
END $$;

COMMIT;
