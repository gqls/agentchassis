-- 433 ROLLBACK - surgical inverse of 433_planner_directory_rule_b3e.sql:
-- removes exactly the two blocks 433 inserted into build-site-planner's
-- plan_site prompt (the Directory rule paragraph and RULES entry 18),
-- restoring the two anchor sentences to their pre-433 form.
--
-- Deliberately NOT restore-from-backup: the backup row snapshots the WHOLE
-- config, so restoring it would clobber any edit another session has made to
-- any other part of the planner config since 433 applied. This file touches
-- only the prompt string, and only the inserted text.
--
-- The two long literals below MUST stay byte-identical to the insertion
-- literals in the forward file - the replace() finds nothing otherwise and
-- the guards below refuse (which is the correct failure: it means the prompt
-- was edited after 433 and a human should look before rolling back).
--
-- Guards mirror the forward file (the 432 round-1 council lesson: a rollback
-- without the exactly-one-active-row guards is exactly as dangerous as a
-- forward file without them - it runs at the worst possible moment).

SELECT snapshot_agent('build-site-planner', '433_planner_directory_rule_b3e_ROLLBACK.sql: pre-rollback');

BEGIN;

-- ── Pre-flight guards ──────────────────────────────────────────────────────
DO $do$
DECLARE
    n int;
    p text;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '433 ROLLBACK: build-site-planner does not have exactly one active row (found %) - resolve the ambiguity before editing', n;
    END IF;

    -- Presence, not exactly-one: snapshot_agent runs OUTSIDE this txn, so a
    -- dry run or refused-then-retried attempt legitimately leaves earlier
    -- rows with this reason. A MISSING row is the failure this guard catches.
    SELECT count(*) INTO n FROM agent_definitions_backup
    WHERE snapshot_reason = '433_planner_directory_rule_b3e_ROLLBACK.sql: pre-rollback'
      AND type = 'build-site-planner';
    IF n < 1 THEN
        RAISE EXCEPTION '433 ROLLBACK: no pre-rollback backup row - snapshot_agent did not run';
    END IF;

    SELECT default_config#>>'{workflow,steps,plan_site,config,prompt_template}' INTO p
    FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF p IS NULL THEN
        RAISE EXCEPTION '433 ROLLBACK: no prompt_template at workflow.steps.plan_site.config - the row has drifted';
    END IF;
    IF (length(p) - length(replace(p, 'Directory rule:', ''))) / length('Directory rule:') <> 1 THEN
        RAISE EXCEPTION '433 ROLLBACK: Directory rule not present exactly once - 433 is not applied, or the prompt was edited since; inspect before rolling back';
    END IF;
    IF (length(p) - length(replace(p, '18. If the classification data includes any content_features directory key', '')))
       / length('18. If the classification data includes any content_features directory key') <> 1 THEN
        RAISE EXCEPTION '433 ROLLBACK: rule 18 not present exactly once - 433 is not applied, or the prompt was edited since; inspect before rolling back';
    END IF;
END $do$;

-- ── Apply inverse: replace (anchor + inserted block) with (anchor) ─────────
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,prompt_template}',
        to_jsonb(
            replace(
                replace(
                    default_config#>>'{workflow,steps,plan_site,config,prompt_template}',
                    'orphans the page from all of it.'
                    || E'\n\n' ||
                    'Directory rule: content_features may carry one or more global directory keys — model_directory, adoption_tracker, protocol_tracker, mortgage_lender_directory, savings_provider_directory, health_insurer_directory. Each records that this site has opted into a fleet-wide, cited provider directory of ONE kind; several keys may be present and each is handled independently. For EACH such key with recommended=true: (a) add the matching SECTION component to the homepage sections array — model_directory uses model-directory, adoption_tracker uses adoption-tracker, protocol_tracker uses protocol-tracker, mortgage_lender_directory uses mortgage-lender-directory, savings_provider_directory uses savings-provider-directory, health_insurer_directory uses health-insurer-directory. (b) when that key also has separate_page=true, additionally plan exactly ONE dedicated page whose name and page_type are BOTH the exact value listed here — these are routing keys the directory machinery (publish trigger, discovery checks, renderers) selects on; never substitute entity-directory, content, or any other page_type: model_directory uses model-directory, adoption_tracker uses adoption-tracker, protocol_tracker uses protocol-tracker, mortgage_lender_directory uses mortgage-lenders, savings_provider_directory uses savings-providers, health_insurer_directory uses health-insurers. Compose that page with populated sections in this order: hero, then the LISTING component (the section component name plus a -listing suffix, e.g. mortgage-lender-directory-listing), then call-to-action; set in_header and in_footer true. Directory sections render from the global register, not from written content — when Verified Facts are in play, give every directory section facts: [] (a fact assigned to one is stated nowhere). These six page_type values are conditional additions to the Canonical Page Types table — valid ONLY when this rule fires, never without the content_features key. Do NOT plan a directory page or section for a kind whose key is absent or not recommended: a page for a kind the site has not opted into ships empty.',
                    'orphans the page from all of it.'
                ),
                'When no Verified Facts are listed, use plain strings only.'
                || E'\n' ||
                '18. If the classification data includes any content_features directory key (model_directory, adoption_tracker, protocol_tracker, mortgage_lender_directory, savings_provider_directory, health_insurer_directory) with recommended = true, apply the Directory rule above: the matching section component on the homepage, and when separate_page = true, the dedicated page with its exact name, exact page_type, and hero + listing + call-to-action sections',
                'When no Verified Facts are listed, use plain strings only.'
            )
        )
    ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config#>>'{workflow,steps,plan_site,config,prompt_template}'
      LIKE '%Directory rule:%';

-- ── Verify in-transaction (DO/RAISE) ───────────────────────────────────────
DO $do$
DECLARE
    p text;
BEGIN
    SELECT default_config#>>'{workflow,steps,plan_site,config,prompt_template}' INTO p
    FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    -- NULL-before-arithmetic guard (council advisory, debug_historian, corr
    -- 53ae1501 round 2). Without it this whole block is a check that CANNOT
    -- FAIL: if the SELECT returns no row, p is NULL, every length()/position()
    -- below is NULL, and `NULL <> 1` is NULL - so no IF fires and the verify
    -- PASSES while having inspected nothing. The pre-flight already refuses on
    -- a NULL prompt; the post-check has to as well, or the pair is asymmetric
    -- in exactly the direction that matters.
    IF p IS NULL THEN
        RAISE EXCEPTION '433 ROLLBACK verify: no active build-site-planner prompt_template found after the update - cannot verify, refusing to commit';
    END IF;

    IF position('Directory rule:' in p) > 0 THEN
        RAISE EXCEPTION '433 ROLLBACK verify: Directory rule still present - the inverse replace found a drifted block; ROLLBACK the transaction and inspect';
    END IF;
    IF position('18. If the classification data includes any content_features directory key' in p) > 0 THEN
        RAISE EXCEPTION '433 ROLLBACK verify: rule 18 still present - the inverse replace found a drifted block; ROLLBACK the transaction and inspect';
    END IF;
    IF (length(p) - length(replace(p, 'orphans the page from all of it.', '')))
       / length('orphans the page from all of it.') <> 1 THEN
        RAISE EXCEPTION '433 ROLLBACK verify: news-rule anchor not restored to exactly once';
    END IF;
    IF (length(p) - length(replace(p, 'When no Verified Facts are listed, use plain strings only.', '')))
       / length('When no Verified Facts are listed, use plain strings only.') <> 1 THEN
        RAISE EXCEPTION '433 ROLLBACK verify: rule-17 anchor not restored to exactly once';
    END IF;
END $do$;

COMMIT;
