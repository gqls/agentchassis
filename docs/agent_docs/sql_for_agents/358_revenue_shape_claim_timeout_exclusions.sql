-- 358_revenue_shape_claim_timeout_exclusions.sql
--
-- Adds 'revenue_shape_cta' and 'missing_conversion_path' to the claimed-item-
-- timeout sweep's item_type exclusion list (scheduled_tasks.pre_query), so
-- VerifyRevenueShapeCTAResolved / VerifyConversionPathResolved can gate
-- completion instead of being bypassed.
--
-- vigilant_designer_offer_analysis Programme B, phase B3 — the two offer checks
-- (check_premise_incomplete.go, check_revenue_shape.go) register verifiers for
-- these two NEW item types, and the moment they did,
-- TestRegisteredVerifiersMatchClaimTimeoutExclusion named this obligation.
-- The Go test reads the DECLARED list in 220 (updated in the same commit);
-- this file makes the LIVE column agree. Both halves required — 305's header
-- records the 151 lane declaring in 220 only, leaving their verifier
-- bypassable for two days. Precedent chain: 269 → 305 → 322 → 331 → this.
--
-- LIVE STATE READ BEFORE WRITING (2026-08-09, per the LANDMINES rule that a
-- replace() must name the exact current string):
--   claimed-item-timeout | item_type NOT IN ('truncated_component',
--     'hardcoded_section_colors', 'empty_section', 'orphan_element_refs',
--     'content_duplication', 'page_canonical_collision', 'dead_fragment_link',
--     'literal_markdown', 'unbuilt_internal_link')
--   Exactly one row, nine entries, IDENTICAL to 220's declared list at this
--   commit's parent — no drift to carry.
--
-- URGENCY, honestly: ZERO items of either type exist (the checks are unbuilt
-- until an image rolls, and unenabled until named in quality-discovery-agent's
-- checks array after that). Until the roll there is no verifier to bypass, so
-- applying now is safe and closes the window before it opens — 331's position
-- exactly.
--
-- ROLLBACK: the inverse replace, same shape.

\set ON_ERROR_STOP on

BEGIN;

DO $pre$
DECLARE v_rows int;
BEGIN
    SELECT count(*) INTO v_rows FROM scheduled_tasks
    WHERE name = 'claimed-item-timeout'
      AND pre_query LIKE '%''dead_fragment_link'', ''literal_markdown'', ''unbuilt_internal_link''%'
      AND pre_query NOT LIKE '%revenue_shape_cta%';
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '358: pre-state mismatch (% rows) — the live exclusion list is not the string this replace targets; re-read it and recompose', v_rows;
    END IF;
END $pre$;

UPDATE scheduled_tasks
SET pre_query = replace(
    pre_query,
    '''dead_fragment_link'', ''literal_markdown'', ''unbuilt_internal_link''',
    '''dead_fragment_link'', ''literal_markdown'', ''unbuilt_internal_link'', ''revenue_shape_cta'', ''missing_conversion_path'''
)
WHERE name = 'claimed-item-timeout';

DO $post$
DECLARE v_rows int;
BEGIN
    SELECT count(*) INTO v_rows FROM scheduled_tasks
    WHERE name = 'claimed-item-timeout'
      AND pre_query LIKE '%''revenue_shape_cta'', ''missing_conversion_path''%';
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '358: post-state mismatch (% rows) — the replace did not land', v_rows;
    END IF;
END $post$;

COMMIT;
