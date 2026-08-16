-- 441 - B3e refinement, from council round 1 on 433 (corr 53ae1501, REVISE):
-- replace the DERIVED listing-component name with an explicit enumeration, and
-- state the consequence of paraphrasing a component name.
--
-- THE OBJECTION (bug_historian, medium, and it is right): 433's rule told the
-- planner to compose the directory page with "the LISTING component (the section
-- component name plus a -listing suffix, e.g. mortgage-lender-directory-listing)".
-- That is a DERIVATION INSTRUCTION - it asks a model composing a ~22KB prompt to
-- CONSTRUCT a component name rather than COPY one. The seat's cited landmine is
-- the reason that matters: validate_site_plan silently DROPS a section whose name
-- does not resolve to a real component, with no error and no work item - the page
-- just ships without the section, invisible until the next discovery sweep files
-- missing_<x>_section, which is the very one-sweep-late gap 433 exists to close.
-- Every other name in 433's rule is enumerated; this was the one exception, and
-- the exception was the derivable one. Now enumerated like the rest.
--
-- WHAT THE SAME ROUND ESTABLISHED, so a future reader does not re-litigate it
-- (all measured, in 433's round-2 resubmission's grounded_in):
--   - The names 433 teaches are content_components.FUNCTION values, and function
--     is the canonical resolution target: componentNameResolver.resolve()'s FIRST
--     arm is "already a valid function -> return unchanged"
--     (v3_site_actions.go:3896-3898). Teaching the function string is the safest
--     possible input, not a lucky one. name <> function for all 12 directory rows
--     (name is prose: "UK Mortgage Lender Directory"), so teaching `name` would
--     have been the defect the seat feared.
--   - Since migration 439 (another lane, landed 2026-08-16 mid-session)
--     validate_plan also carries menu_field=available_components, so the
--     planner's own menu rows are added to the valid set as well - both paths
--     now cover these 12.
--
-- ORDERING NOTE, and it is load-bearing: this migration edits text INSIDE the
-- block 433 inserted, so 433's surgical-inverse ROLLBACK will no longer match its
-- literal and WILL REFUSE (by design - it refuses rather than mis-splices).
-- **To undo the pair, run 441's ROLLBACK first, then 433's.** Stated here because
-- the refusal message alone would send a reader hunting for drift that is not there.
--
-- Config-only, live immediately. Guards: exactly-one-active-row, anchor
-- exactly-once, already-applied refusal; verify is DO/RAISE.

SELECT snapshot_agent('build-site-planner', '441_planner_directory_rule_enumerate_listing_names.sql: pre-update');

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
        RAISE EXCEPTION '441: build-site-planner does not have exactly one active row (found %) - resolve the ambiguity before editing', n;
    END IF;

    -- Presence, not exactly-one: snapshot_agent runs OUTSIDE this txn, so a dry
    -- run or a refused-then-retried apply legitimately leaves earlier rows.
    SELECT count(*) INTO n FROM agent_definitions_backup
    WHERE snapshot_reason = '441_planner_directory_rule_enumerate_listing_names.sql: pre-update'
      AND type = 'build-site-planner';
    IF n < 1 THEN
        RAISE EXCEPTION '441: no pre-update backup row - snapshot_agent did not run';
    END IF;

    SELECT default_config#>>'{workflow,steps,plan_site,config,prompt_template}' INTO p
    FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF p IS NULL THEN
        RAISE EXCEPTION '441: no prompt_template at workflow.steps.plan_site.config - the row has drifted, re-read before applying';
    END IF;
    IF position('Directory rule:' in p) = 0 THEN
        RAISE EXCEPTION '441: 433 is not applied (no Directory rule in the prompt) - apply 433 first';
    END IF;
    IF position('model_directory uses model-directory-listing' in p) > 0 THEN
        RAISE EXCEPTION '441: already applied - the listing names are already enumerated';
    END IF;
    IF (length(p) - length(replace(p, 'then the LISTING component (the section component name plus a -listing suffix, e.g. mortgage-lender-directory-listing), then call-to-action', '')))
       / length('then the LISTING component (the section component name plus a -listing suffix, e.g. mortgage-lender-directory-listing), then call-to-action') <> 1 THEN
        RAISE EXCEPTION '441: the derivation clause is not present exactly once - 433''s block has been edited since; re-derive the anchor from the live row';
    END IF;
END $do$;

-- ── Apply ──────────────────────────────────────────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,prompt_template}',
        to_jsonb(
            replace(
                default_config#>>'{workflow,steps,plan_site,config,prompt_template}',
                'then the LISTING component (the section component name plus a -listing suffix, e.g. mortgage-lender-directory-listing), then call-to-action',
                'then the LISTING component named here — model_directory uses model-directory-listing, adoption_tracker uses adoption-tracker-listing, protocol_tracker uses protocol-tracker-listing, mortgage_lender_directory uses mortgage-lender-directory-listing, savings_provider_directory uses savings-provider-directory-listing, health_insurer_directory uses health-insurer-directory-listing — then call-to-action'
            )
        )
    ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config#>>'{workflow,steps,plan_site,config,prompt_template}'
      LIKE '%then the LISTING component (the section component name plus a -listing suffix%'
  AND default_config#>>'{workflow,steps,plan_site,config,prompt_template}'
      NOT LIKE '%model_directory uses model-directory-listing%';

-- Second edit: state the consequence of paraphrasing a component name, at the
-- end of the Directory rule paragraph. The rule already says to use the exact
-- page_type; nothing said what happens to a mistyped COMPONENT name, and what
-- happens is a silent drop.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,prompt_template}',
        to_jsonb(
            replace(
                default_config#>>'{workflow,steps,plan_site,config,prompt_template}',
                'a page for a kind the site has not opted into ships empty.',
                'a page for a kind the site has not opted into ships empty. Copy every component name in this rule EXACTLY as written: component names are matched verbatim against the Available Section Components list, and a paraphrased name (an underscore for a hyphen, a singular for a plural, a display title for a component name) is DROPPED from the plan silently — the page then ships without that section and nothing reports it.'
            )
        )
    ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config#>>'{workflow,steps,plan_site,config,prompt_template}'
      LIKE '%a page for a kind the site has not opted into ships empty.%'
  AND default_config#>>'{workflow,steps,plan_site,config,prompt_template}'
      NOT LIKE '%Copy every component name in this rule EXACTLY as written%';

-- ── Verify in-transaction (DO/RAISE) ───────────────────────────────────────
DO $do$
DECLARE
    p text;
    nm text;
BEGIN
    SELECT default_config#>>'{workflow,steps,plan_site,config,prompt_template}' INTO p
    FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    -- All six listing names enumerated, each exactly once in the new clause.
    FOREACH nm IN ARRAY ARRAY[
        'model_directory uses model-directory-listing',
        'adoption_tracker uses adoption-tracker-listing',
        'protocol_tracker uses protocol-tracker-listing',
        'mortgage_lender_directory uses mortgage-lender-directory-listing',
        'savings_provider_directory uses savings-provider-directory-listing',
        'health_insurer_directory uses health-insurer-directory-listing'
    ] LOOP
        IF position(nm in p) = 0 THEN
            RAISE EXCEPTION '441 verify: listing enumeration missing "%"', nm;
        END IF;
    END LOOP;

    IF position('then the LISTING component (the section component name plus a -listing suffix' in p) > 0 THEN
        RAISE EXCEPTION '441 verify: the derivation clause is still present - the replace did not fire';
    END IF;
    IF position('Copy every component name in this rule EXACTLY as written' in p) = 0 THEN
        RAISE EXCEPTION '441 verify: the exact-copy consequence sentence is missing';
    END IF;
    -- 433's block and its neighbours must be intact.
    IF (length(p) - length(replace(p, 'Directory rule:', ''))) / length('Directory rule:') <> 1 THEN
        RAISE EXCEPTION '441 verify: Directory rule is no longer present exactly once';
    END IF;
    IF (length(p) - length(replace(p, 'orphans the page from all of it.', '')))
       / length('orphans the page from all of it.') <> 1 THEN
        RAISE EXCEPTION '441 verify: news listing rule damaged';
    END IF;
    IF (length(p) - length(replace(p, 'When no Verified Facts are listed, use plain strings only.', '')))
       / length('When no Verified Facts are listed, use plain strings only.') <> 1 THEN
        RAISE EXCEPTION '441 verify: rule 17 damaged';
    END IF;
END $do$;

COMMIT;
