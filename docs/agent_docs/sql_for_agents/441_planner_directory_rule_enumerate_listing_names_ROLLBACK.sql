-- 441 ROLLBACK - surgical inverse of
-- 441_planner_directory_rule_enumerate_listing_names.sql: puts the derivation
-- clause back and removes the exact-copy consequence sentence, leaving 433's
-- Directory rule otherwise untouched.
--
-- RUN THIS BEFORE 433's ROLLBACK if undoing the pair. 441 edits text INSIDE the
-- block 433 inserted, so while 441 is applied, 433's inverse cannot match its
-- literal and refuses (by design). Running this first restores 433's block to
-- its as-inserted bytes, after which 433's ROLLBACK matches again.
--
-- Deliberately NOT restore-from-backup (the backup row snapshots the WHOLE
-- config; restoring it would clobber later unrelated edits by other sessions -
-- and this prompt is edited by several lanes; migration 439 landed in the same
-- agent's config on the same day).

SELECT snapshot_agent('build-site-planner', '441_planner_directory_rule_enumerate_listing_names_ROLLBACK.sql: pre-rollback');

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
        RAISE EXCEPTION '441 ROLLBACK: build-site-planner does not have exactly one active row (found %) - resolve the ambiguity before editing', n;
    END IF;

    SELECT count(*) INTO n FROM agent_definitions_backup
    WHERE snapshot_reason = '441_planner_directory_rule_enumerate_listing_names_ROLLBACK.sql: pre-rollback'
      AND type = 'build-site-planner';
    IF n < 1 THEN
        RAISE EXCEPTION '441 ROLLBACK: no pre-rollback backup row - snapshot_agent did not run';
    END IF;

    SELECT default_config#>>'{workflow,steps,plan_site,config,prompt_template}' INTO p
    FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF p IS NULL THEN
        RAISE EXCEPTION '441 ROLLBACK: no prompt_template at workflow.steps.plan_site.config - the row has drifted';
    END IF;
    IF (length(p) - length(replace(p, 'model_directory uses model-directory-listing', '')))
       / length('model_directory uses model-directory-listing') <> 1 THEN
        RAISE EXCEPTION '441 ROLLBACK: the listing enumeration is not present exactly once - 441 is not applied, or the block was edited since; inspect before rolling back';
    END IF;
    IF position('Copy every component name in this rule EXACTLY as written' in p) = 0 THEN
        RAISE EXCEPTION '441 ROLLBACK: the exact-copy sentence is absent - 441 is only half applied; inspect before rolling back';
    END IF;
END $do$;

-- ── Apply inverse ──────────────────────────────────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,prompt_template}',
        to_jsonb(
            replace(
                replace(
                    default_config#>>'{workflow,steps,plan_site,config,prompt_template}',
                    'then the LISTING component named here — model_directory uses model-directory-listing, adoption_tracker uses adoption-tracker-listing, protocol_tracker uses protocol-tracker-listing, mortgage_lender_directory uses mortgage-lender-directory-listing, savings_provider_directory uses savings-provider-directory-listing, health_insurer_directory uses health-insurer-directory-listing — then call-to-action',
                    'then the LISTING component (the section component name plus a -listing suffix, e.g. mortgage-lender-directory-listing), then call-to-action'
                ),
                'a page for a kind the site has not opted into ships empty. Copy every component name in this rule EXACTLY as written: component names are matched verbatim against the Available Section Components list, and a paraphrased name (an underscore for a hyphen, a singular for a plural, a display title for a component name) is DROPPED from the plan silently — the page then ships without that section and nothing reports it.',
                'a page for a kind the site has not opted into ships empty.'
            )
        )
    ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config#>>'{workflow,steps,plan_site,config,prompt_template}'
      LIKE '%model_directory uses model-directory-listing%';

-- ── Verify in-transaction (DO/RAISE) ───────────────────────────────────────
DO $do$
DECLARE
    p text;
BEGIN
    SELECT default_config#>>'{workflow,steps,plan_site,config,prompt_template}' INTO p
    FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF position('model_directory uses model-directory-listing' in p) > 0 THEN
        RAISE EXCEPTION '441 ROLLBACK verify: the listing enumeration is still present - the inverse replace found a drifted block; ROLLBACK the transaction and inspect';
    END IF;
    IF position('Copy every component name in this rule EXACTLY as written' in p) > 0 THEN
        RAISE EXCEPTION '441 ROLLBACK verify: the exact-copy sentence is still present - the inverse replace found a drifted block; ROLLBACK the transaction and inspect';
    END IF;
    IF (length(p) - length(replace(p, 'then the LISTING component (the section component name plus a -listing suffix', '')))
       / length('then the LISTING component (the section component name plus a -listing suffix') <> 1 THEN
        RAISE EXCEPTION '441 ROLLBACK verify: the derivation clause was not restored exactly once - 433''s block is NOT back to its as-inserted bytes, so 433''s ROLLBACK will still refuse; inspect';
    END IF;
    IF (length(p) - length(replace(p, 'Directory rule:', ''))) / length('Directory rule:') <> 1 THEN
        RAISE EXCEPTION '441 ROLLBACK verify: Directory rule is no longer present exactly once';
    END IF;
END $do$;

COMMIT;
