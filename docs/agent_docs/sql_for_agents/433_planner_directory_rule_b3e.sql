-- 433 - B3e: teach build-site-planner the directory vocabulary - the sixth of
-- the SEVEN places a directory kind lives (DIR-001: profiles x3, components,
-- checks-array enablement, planner prompt, publish-trigger VALUES map). Until
-- this migration the planner prompt was the UNFILLED place for ALL six kinds:
-- a greenfield site whose classification carries a content_features directory
-- key (written at plan time since 432 wired evaluate_directory_features into
-- domain-research-classifier) still planned no directory page and no homepage
-- section, so the site only acquired them via the discovery-check round trip
-- (missing_<x>_section / missing_<x>_page -> content-gap-planner), one full
-- sweep late.
--
-- What it adds, via 206's replace()-idiom (206_planner_news_index_page_type.sql,
-- the proven precedent for splicing a content_features-driven rule into this
-- exact prompt), but surgically on the prompt string itself rather than on the
-- whole config text:
--
--   1. A "Directory rule:" paragraph immediately after the News listing rule
--      (same section of the prompt, same conditional-on-content_features
--      shape). For each of the six spec keys with recommended=true: the
--      SNIPPET component on the homepage; and when separate_page=true, ONE
--      dedicated page whose name and page_type are EXACTLY the profile values,
--      composed hero -> <x>-listing -> call-to-action, in header+footer nav.
--   2. A RULES-list entry (rule 18, after rule 17) pointing at that paragraph,
--      because the RULES block is what the model re-reads while composing.
--
-- GROUNDING (all measured live, 2026-08-15 ~22:20Z):
--   - spec-key -> component/page mapping is verbatim from directoryCheckProfiles
--     (check_directory.go:72-149): SnippetComponent = bare name, ListingComponent
--     = -listing twin, PageName = PageType (model-directory, adoption-tracker,
--     protocol-tracker, mortgage-lenders, savings-providers, health-insurers).
--   - Composition matches the three LIVE directory pages on
--     ai-agent-orchestration.com: sections = ["hero","<x>-listing","call-to-action"],
--     and matches MissingDirectoryPageCheck's own suggestion text ("a hero
--     section, the <x> component, and a call-to-action ... header and footer").
--   - All 12 component rows exist, is_active, component_level='section', no
--     requires-backend tag => they already flow into available_components via
--     load_components; this rule tells the planner WHEN, not WHAT EXISTS.
--   - The six page_type values survive the Go plan validator untouched:
--     normaliseRole passes unknown roles through (page_role_validator.go:249),
--     CanonicalisePage's default arm gives name=slug, url=/<slug>.html and
--     PRESERVES the type (page_canonical.go:249-263) - exactly the shape of the
--     live pages. No -index suffix, no declared parent, no /<slug>/index.html
--     URL, so section-index flattening rules 2-4 cannot fire on them.
--   - Anchor counts on the live row (pre-flight re-asserts): news-rule tail
--     'orphans the page from all of it.' = 1; rule-17 tail 'When no Verified
--     Facts are listed, use plain strings only.' = 1; 'Directory rule:' = 0.
--   - build-site-planner active rows = 1 (version 1) - the two-active-rows
--     landmine measured clean; the pre-flight refuses on <> 1 anyway.
--
-- Deliberately NOT added to the Canonical Page Types table itself: the six
-- types are valid ONLY when the content_features key licenses them (an
-- unconditioned table row would invite a directory page on any vaguely
-- matching site, and a page for a kind the site has not opted into ships
-- empty - the publish leg fans out only to opted-in sites with a deployed
-- component). The paragraph states them as conditional additions instead,
-- mirroring how the flag itself is opt-in with the unsafe side OFF.
--
-- Config-only: live immediately, no image roll. In-flight planner runs carry
-- their own workflow snapshot and are unaffected.
--
-- ROLLBACK: 433_planner_directory_rule_b3e_ROLLBACK.sql - surgical inverse
-- (removes exactly the two inserted blocks), NOT restore-from-backup, so a
-- concurrent edit to any other part of the planner config survives a
-- rollback. Its guards mirror these (exactly-one-active-row, blocks present
-- exactly once).

SELECT snapshot_agent('build-site-planner', '433_planner_directory_rule_b3e.sql: pre-update');

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
        RAISE EXCEPTION '433: build-site-planner does not have exactly one active row (found %) - resolve the ambiguity before editing', n;
    END IF;

    -- Presence, not exactly-one: snapshot_agent runs OUTSIDE this txn, so a
    -- dry run (COMMIT swapped for ROLLBACK) or a refused-then-retried apply
    -- legitimately leaves earlier rows with this reason. Extra snapshots are
    -- harmless; a MISSING one is the failure this guard exists to catch.
    SELECT count(*) INTO n FROM agent_definitions_backup
    WHERE snapshot_reason = '433_planner_directory_rule_b3e.sql: pre-update'
      AND type = 'build-site-planner';
    IF n < 1 THEN
        RAISE EXCEPTION '433: no pre-update backup row - snapshot_agent did not run';
    END IF;

    SELECT default_config#>>'{workflow,steps,plan_site,config,prompt_template}' INTO p
    FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF p IS NULL THEN
        RAISE EXCEPTION '433: no prompt_template at workflow.steps.plan_site.config - the row has drifted, re-read before applying';
    END IF;
    IF position('Directory rule:' in p) > 0 THEN
        RAISE EXCEPTION '433: already applied - the prompt carries a Directory rule';
    END IF;
    IF (length(p) - length(replace(p, 'orphans the page from all of it.', '')))
       / length('orphans the page from all of it.') <> 1 THEN
        RAISE EXCEPTION '433: news-rule anchor is not present exactly once - the prompt has drifted, re-derive anchors from the live row';
    END IF;
    IF (length(p) - length(replace(p, 'When no Verified Facts are listed, use plain strings only.', '')))
       / length('When no Verified Facts are listed, use plain strings only.') <> 1 THEN
        RAISE EXCEPTION '433: rule-17 anchor is not present exactly once - the prompt has drifted, re-derive anchors from the live row';
    END IF;
END $do$;

-- ── Apply: two insertions on the prompt string, put back with jsonb_set ────
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,prompt_template}',
        to_jsonb(
            replace(
                replace(
                    default_config#>>'{workflow,steps,plan_site,config,prompt_template}',
                    'orphans the page from all of it.',
                    'orphans the page from all of it.'
                    || E'\n\n' ||
                    'Directory rule: content_features may carry one or more global directory keys — model_directory, adoption_tracker, protocol_tracker, mortgage_lender_directory, savings_provider_directory, health_insurer_directory. Each records that this site has opted into a fleet-wide, cited provider directory of ONE kind; several keys may be present and each is handled independently. For EACH such key with recommended=true: (a) add the matching SECTION component to the homepage sections array — model_directory uses model-directory, adoption_tracker uses adoption-tracker, protocol_tracker uses protocol-tracker, mortgage_lender_directory uses mortgage-lender-directory, savings_provider_directory uses savings-provider-directory, health_insurer_directory uses health-insurer-directory. (b) when that key also has separate_page=true, additionally plan exactly ONE dedicated page whose name and page_type are BOTH the exact value listed here — these are routing keys the directory machinery (publish trigger, discovery checks, renderers) selects on; never substitute entity-directory, content, or any other page_type: model_directory uses model-directory, adoption_tracker uses adoption-tracker, protocol_tracker uses protocol-tracker, mortgage_lender_directory uses mortgage-lenders, savings_provider_directory uses savings-providers, health_insurer_directory uses health-insurers. Compose that page with populated sections in this order: hero, then the LISTING component (the section component name plus a -listing suffix, e.g. mortgage-lender-directory-listing), then call-to-action; set in_header and in_footer true. Directory sections render from the global register, not from written content — when Verified Facts are in play, give every directory section facts: [] (a fact assigned to one is stated nowhere). These six page_type values are conditional additions to the Canonical Page Types table — valid ONLY when this rule fires, never without the content_features key. Do NOT plan a directory page or section for a kind whose key is absent or not recommended: a page for a kind the site has not opted into ships empty.'
                ),
                'When no Verified Facts are listed, use plain strings only.',
                'When no Verified Facts are listed, use plain strings only.'
                || E'\n' ||
                '18. If the classification data includes any content_features directory key (model_directory, adoption_tracker, protocol_tracker, mortgage_lender_directory, savings_provider_directory, health_insurer_directory) with recommended = true, apply the Directory rule above: the matching section component on the homepage, and when separate_page = true, the dedicated page with its exact name, exact page_type, and hero + listing + call-to-action sections'
            )
        )
    ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config#>>'{workflow,steps,plan_site,config,prompt_template}'
      LIKE '%orphans the page from all of it.%'
  AND default_config#>>'{workflow,steps,plan_site,config,prompt_template}'
      LIKE '%When no Verified Facts are listed, use plain strings only.%'
  AND default_config#>>'{workflow,steps,plan_site,config,prompt_template}'
      NOT LIKE '%Directory rule:%';

-- ── Verify in-transaction (DO/RAISE - a SELECT cannot stop the COMMIT) ─────
DO $do$
DECLARE
    p text;
BEGIN
    SELECT default_config#>>'{workflow,steps,plan_site,config,prompt_template}' INTO p
    FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF (length(p) - length(replace(p, 'Directory rule:', ''))) / length('Directory rule:') <> 1 THEN
        RAISE EXCEPTION '433 verify: Directory rule paragraph not present exactly once after update';
    END IF;
    IF (length(p) - length(replace(p, '18. If the classification data includes any content_features directory key', '')))
       / length('18. If the classification data includes any content_features directory key') <> 1 THEN
        RAISE EXCEPTION '433 verify: rule 18 not present exactly once after update';
    END IF;
    IF position('mortgage_lender_directory uses mortgage-lenders' in p) = 0 THEN
        RAISE EXCEPTION '433 verify: spec-key -> page mapping missing from the inserted rule';
    END IF;
    IF (length(p) - length(replace(p, 'orphans the page from all of it.', '')))
       / length('orphans the page from all of it.') <> 1 THEN
        RAISE EXCEPTION '433 verify: news listing rule damaged by the splice';
    END IF;
    IF (length(p) - length(replace(p, 'When no Verified Facts are listed, use plain strings only.', '')))
       / length('When no Verified Facts are listed, use plain strings only.') <> 1 THEN
        RAISE EXCEPTION '433 verify: rule 17 damaged by the splice';
    END IF;
END $do$;

COMMIT;
