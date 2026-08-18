-- 468_brand_head_claim_timeout_exclusion.sql
--
-- Adds 'needs_brand_head_assets' to the claimed-item-timeout sweep's item_type
-- exclusion list (scheduled_tasks.pre_query), so its new Go verifier can gate
-- completion instead of being bypassed.
--
-- WHY (the lockstep, not a judgement call) — bugs_open/131 (og-card slug)
-- ----------------------------------------------------------------------
-- check_undeployed_assets.go now registers VerifyBrandHeadAssetsResolved for
-- needs_brand_head_assets: 21 such items had been stamped 'complete' off a
-- deploy_image_asset brand-head refusal that derived nothing (the items were
-- filed without spec.mode, so asset-deployer's chain fell through to the
-- deploy branch, whose refusal-as-result completes the workflow). The moment
-- the verifier registered, TestRegisteredVerifiersMatchClaimTimeoutExclusion
-- named this obligation: a verifier the auto-complete branch can walk past is
-- not a gate. The Go test reads the DECLARED list in
-- 220_claimed_item_timeout_generic_evidence.sql (amended in the same commit);
-- this file is what makes the LIVE column agree.
--
-- Precedent and mechanism: 269, then 305, then 322 — targeted replace against
-- the exact current string, with a before-assertion and an after-assertion.
-- Never retype the 84-line pre_query column.
--
-- LIVE STATE READ BEFORE WRITING THIS (2026-08-18):
--   SELECT name, substring(pre_query from 'item_type NOT IN \([^)]*\)')
--     FROM scheduled_tasks WHERE pre_query LIKE '%item_type NOT IN%';
--   claimed-item-timeout | item_type NOT IN ('truncated_component',
--     'hardcoded_section_colors', 'empty_section', 'orphan_element_refs',
--     'content_duplication', 'page_canonical_collision', 'dead_fragment_link',
--     'literal_markdown', 'unbuilt_internal_link', 'revenue_shape_cta',
--     'missing_conversion_path', 'decision_regression')
-- Exactly one row, twelve entries, identical to 220's declared list — no drift
-- to carry.
--
-- URGENCY, stated honestly: LOW, and one hazard window stated rather than
-- hidden. This config is live on apply; the verifier it protects is Go and
-- rides the next chassis roll. Between the two, a STUCK CLAIM of this type
-- (claimed, handler orchestration COMPLETED, claim never released) falls to
-- the 40-minute reset instead of auto-completing. Measured 2026-08-18: zero
-- open claimed items of this type exist, so the window is expected to bite
-- nothing; it self-heals at the roll.
--
-- ROLLBACK: the inverse replace, same shape.

\set ON_ERROR_STOP on

BEGIN;

DO $pre$
DECLARE v_rows int;
BEGIN
    SELECT count(*) INTO v_rows FROM scheduled_tasks
     WHERE pre_query LIKE '%item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'', ''orphan_element_refs'', ''content_duplication'', ''page_canonical_collision'', ''dead_fragment_link'', ''literal_markdown'', ''unbuilt_internal_link'', ''revenue_shape_cta'', ''missing_conversion_path'', ''decision_regression'')%';
    IF v_rows <> 1 THEN
        RAISE EXCEPTION 'expected exactly 1 scheduled_task carrying the known 12-entry exclusion list, found % — the live pre_query has drifted; read it before applying this', v_rows;
    END IF;
END $pre$;

UPDATE scheduled_tasks
   SET pre_query = replace(
         pre_query,
         $old$item_type NOT IN ('truncated_component', 'hardcoded_section_colors', 'empty_section', 'orphan_element_refs', 'content_duplication', 'page_canonical_collision', 'dead_fragment_link', 'literal_markdown', 'unbuilt_internal_link', 'revenue_shape_cta', 'missing_conversion_path', 'decision_regression')$old$,
         $new$item_type NOT IN ('truncated_component', 'hardcoded_section_colors', 'empty_section', 'orphan_element_refs', 'content_duplication', 'page_canonical_collision', 'dead_fragment_link', 'literal_markdown', 'unbuilt_internal_link', 'revenue_shape_cta', 'missing_conversion_path', 'decision_regression', 'needs_brand_head_assets')$new$),
       updated_at = NOW()
 WHERE pre_query LIKE '%item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'', ''orphan_element_refs'', ''content_duplication'', ''page_canonical_collision'', ''dead_fragment_link'', ''literal_markdown'', ''unbuilt_internal_link'', ''revenue_shape_cta'', ''missing_conversion_path'', ''decision_regression'')%';

DO $verify$
DECLARE v_after int; v_old int;
BEGIN
    SELECT count(*) INTO v_after FROM scheduled_tasks
     WHERE pre_query LIKE '%''decision_regression'', ''needs_brand_head_assets''%';
    IF v_after <> 1 THEN
        RAISE EXCEPTION 'exclusion list not extended: % rows carry the new list', v_after;
    END IF;

    -- The replacement must have CONSUMED the old list, not appended a second
    -- one. The closing paren is load-bearing: the old 12-entry string ends
    -- ...'decision_regression')  and the new one continues with a comma, so
    -- this pattern matches ONLY a surviving old list.
    SELECT count(*) INTO v_old FROM scheduled_tasks
     WHERE pre_query LIKE '%item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'', ''orphan_element_refs'', ''content_duplication'', ''page_canonical_collision'', ''dead_fragment_link'', ''literal_markdown'', ''unbuilt_internal_link'', ''revenue_shape_cta'', ''missing_conversion_path'', ''decision_regression'')%';
    IF v_old <> 0 THEN
        RAISE EXCEPTION 'the old exclusion list is still present in % row(s)', v_old;
    END IF;

    RAISE NOTICE 'needs_brand_head_assets excluded from the claim-timeout auto-complete; VerifyBrandHeadAssetsResolved now gates its completion';
END $verify$;

COMMIT;
