-- 416_auditor_prompts_drop_dead_work_item_type.sql
--
-- bugs_open/279 leg 1 (candidate 2a): two live auditor prompts REQUIRE a
-- `work_item_type` field in every finding ("Each finding MUST include ALL of
-- these fields") that NOTHING reads. Routing is classifyFinding
-- (platform/orchestration/actions/write_audit_findings_action.go), deterministic
-- in Go on `category` + page existence; the auditFinding struct has no such
-- field, so the value is dropped at parse, before routing is reached. Checked at
-- three layers on 2026-08-15 (the bug file quotes each): the struct, a
-- three-spelling grep over the Go tree (one hit, a WRITER), and a fleet-wide
-- agent_definitions scan (exactly these two prompts, no reader).
--
-- The field costs tokens on every finding of both auditors and does nothing —
-- and, worse, a prompt maintainer who "fixes routing" by editing this vocabulary
-- changes nothing and gets no signal. Deleting it is candidate 2a of the bug
-- file (the recommendation); parsing-and-honouring it (2b) was considered and
-- rejected there because it changes live routing for two currently-working
-- auditors and needs a reason nobody has.
--
-- SAFE TO APPLY INDEPENDENTLY OF THE GO FIX: nothing reads the field today, so
-- removing it from the prompts cannot change routing — image-first ordering does
-- not bind. Both auditors' category vocabularies are fully routable
-- (bugs_open/279 leg 3 table), so no behaviour changes at all; the LLM simply
-- stops being asked for a value that is discarded.
--
-- Each prompt carries the literal exactly TWICE (measured live 2026-08-15): once
-- in the required-fields schema line, once in the worked example. The
-- preconditions count occurrences and refuse on any other number, so prompt
-- drift between writing and applying this file aborts rather than half-applies.
--
-- ROLLBACK: 416_auditor_prompts_drop_dead_work_item_type_ROLLBACK.sql re-inserts
-- both fragments at their anchors.

SELECT snapshot_agent('site-review-agent',
                      '416_auditor_prompts_drop_dead_work_item_type.sql: pre-update');
SELECT snapshot_agent('content-quality-auditor',
                      '416_auditor_prompts_drop_dead_work_item_type.sql: pre-update');

BEGIN;

DO $$
DECLARE
    n_defs integer;
    occ_sr integer;
    occ_cq integer;
BEGIN
    SELECT count(*) INTO n_defs
    FROM agent_definitions
    WHERE type IN ('site-review-agent', 'content-quality-auditor')
      AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    IF n_defs <> 2 THEN
        RAISE EXCEPTION 'MIGRATION 416: expected exactly 2 live definitions (site-review-agent, content-quality-auditor), found %', n_defs;
    END IF;

    SELECT (length(default_config::text) - length(replace(default_config::text, 'work_item_type', ''))) / length('work_item_type')
      INTO occ_sr
    FROM agent_definitions
    WHERE type='site-review-agent'
      AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    SELECT (length(default_config::text) - length(replace(default_config::text, 'work_item_type', ''))) / length('work_item_type')
      INTO occ_cq
    FROM agent_definitions
    WHERE type='content-quality-auditor'
      AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    IF occ_sr = 0 AND occ_cq = 0 THEN
        RAISE EXCEPTION 'MIGRATION 416: work_item_type already absent from both prompts — already applied';
    END IF;

    IF occ_sr <> 2 OR occ_cq <> 2 THEN
        RAISE EXCEPTION 'MIGRATION 416: expected exactly 2 occurrences of work_item_type per prompt (schema line + example), found site-review=% content-quality=%. The prompt has drifted since this file was written — re-derive the replace literals before applying.', occ_sr, occ_cq;
    END IF;

    RAISE NOTICE 'migration 416 pre-conditions OK: 2 occurrences in each prompt';
END $$;

-- site-review-agent: schema line (five-value vocabulary) + worked example.
-- The literals operate on the jsonb::text rendering, where the prompt text's
-- inner quotes appear as \" — hence the backslashes.
UPDATE agent_definitions
SET default_config = replace(replace(default_config::text,
        ',\"work_item_type\":\"content_rewrite|needs_content_page|tone_shift|cta_improvement|nav_restructure\"', ''),
        ',\"work_item_type\":\"needs_content_page\"', '')::jsonb,
    version    = version + 1,
    updated_at = now()
WHERE type='site-review-agent'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- content-quality-auditor: schema line (four-value vocabulary) + worked example.
UPDATE agent_definitions
SET default_config = replace(replace(default_config::text,
        ',\"work_item_type\":\"content_rewrite|needs_content_page|tone_shift|cta_improvement\"', ''),
        ',\"work_item_type\":\"content_rewrite\"', '')::jsonb,
    version    = version + 1,
    updated_at = now()
WHERE type='content-quality-auditor'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- POST-CONDITIONS: zero occurrences left anywhere in either config, and the
-- workflow structure survived the text-level surgery (the ::jsonb cast above
-- already guarantees well-formed JSON; this asserts the steps are still there).
DO $$
DECLARE
    remaining integer;
    sr_steps  integer;
    cq_steps  integer;
BEGIN
    SELECT count(*) INTO remaining
    FROM agent_definitions
    WHERE type IN ('site-review-agent', 'content-quality-auditor')
      AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
      AND default_config::text LIKE '%work_item_type%';

    IF remaining <> 0 THEN
        RAISE EXCEPTION 'MIGRATION 416: % live definition(s) still carry work_item_type after the replace — a literal did not match', remaining;
    END IF;

    SELECT count(*) INTO sr_steps
    FROM jsonb_object_keys((SELECT default_config->'workflow'->'steps'
                            FROM agent_definitions
                            WHERE type='site-review-agent'
                              AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL)) k;
    SELECT count(*) INTO cq_steps
    FROM jsonb_object_keys((SELECT default_config->'workflow'->'steps'
                            FROM agent_definitions
                            WHERE type='content-quality-auditor'
                              AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL)) k;

    IF sr_steps = 0 OR cq_steps = 0 THEN
        RAISE EXCEPTION 'MIGRATION 416: workflow steps vanished (site-review=% content-quality=%) — the text surgery damaged the config', sr_steps, cq_steps;
    END IF;

    RAISE NOTICE 'migration 416 OK: work_item_type gone from both prompts; steps intact (% / %)', sr_steps, cq_steps;
END $$;

COMMIT;
