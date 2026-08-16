-- 434 - B3f: enable ELEVEN discovery checks on completeness-discovery-agent -
-- the three finance directory pairs (Phase B) and the five Phase-A structural
-- validity checks. Same statement family as 150/188/189/190/194/215 (the
-- checks-array append), hardened with this lane's DO/RAISE guard posture.
--
--   Finance directory pairs (check_directory.go, profiles committed B1/B2):
--     missing_mortgage_lender_directory_section / _page
--     missing_savings_provider_directory_section / _page
--     missing_health_insurer_directory_section  / _page
--   Phase-A structural checks (check_site_structural_validity.go, A1):
--     dead_internal_link_live, canonical_mismatch, structured_data_invalid,
--     head_essentials_missing, sitemap_entry_dead_live
--
-- WHY SAFE TO ENABLE NOW:
--   - Directory checks self-gate TWICE: on the site's opt-in flag
--     (content_features.<spec_key>, absent everywhere today - DIR-001 records
--     zero opted-in finance sites) AND on the register holding current found
--     claims OF THAT KIND (B4 populated all three finance kinds 2026-08-15).
--     Enabling produces ZERO work items until the Phase C pilot opts in -
--     which is exactly the order the plan wants (checks armed before the
--     pilot builds, so the pilot proves them).
--   - Structural checks are flag-only routing per A1's council round; they
--     run only when a completeness sweep is dispatched at a site (per-site
--     on demand, owner ruling 2026-07-24).
--
-- IMAGE PRECONDITION PROVEN AT THE ARTEFACT (2026-08-15 22:35Z, pod
-- agent-chassis-584b6fcf-9mtqd): an unregistered check NAME is silently
-- skipped, not an error (discovery_checks.go registry lookup), so the binary
-- was probed with controls in the same breath:
--     grep -aq missing_mortgage_lender_directory_section /proc/1/exe -> PRESENT
--     grep -aq sitemap_entry_dead_live                   /proc/1/exe -> PRESENT
--     positive control missing_model_directory_section   -> PRESENT
--     negative control missing_zzz_nonexistent_check_qqq -> ABSENT
--
-- Config-only: live immediately. Checks array measured 32 entries pre-apply;
-- 43 after. completeness-discovery-agent active rows measured exactly 1.
--
-- ROLLBACK: 434_enable_finance_directory_and_structural_checks_b3f_ROLLBACK.sql
-- (surgical: removes exactly these eleven names, refuses on drift).

SELECT snapshot_agent('completeness-discovery-agent', '434_enable_finance_directory_and_structural_checks_b3f.sql: pre-update');

BEGIN;

-- ── Pre-flight guards ──────────────────────────────────────────────────────
DO $do$
DECLARE
    n int;
    checks jsonb;
    nm text;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'completeness-discovery-agent' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '434: completeness-discovery-agent does not have exactly one active row (found %) - resolve the ambiguity before editing', n;
    END IF;

    -- Presence, not exactly-one (snapshot_agent runs outside this txn; dry
    -- runs and refused retries legitimately accumulate rows with this reason).
    SELECT count(*) INTO n FROM agent_definitions_backup
    WHERE snapshot_reason = '434_enable_finance_directory_and_structural_checks_b3f.sql: pre-update'
      AND type = 'completeness-discovery-agent';
    IF n < 1 THEN
        RAISE EXCEPTION '434: no pre-update backup row - snapshot_agent did not run';
    END IF;

    SELECT default_config#>'{workflow,steps,run_checks,config,checks}' INTO checks
    FROM agent_definitions
    WHERE type = 'completeness-discovery-agent' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF checks IS NULL OR jsonb_typeof(checks) <> 'array' THEN
        RAISE EXCEPTION '434: no checks array at workflow.steps.run_checks.config.checks - the row has drifted, re-read before applying';
    END IF;
    -- Refuse if ANY of the eleven is already enabled (partial application is
    -- the state a human should look at, not one to silently complete).
    FOREACH nm IN ARRAY ARRAY[
        'missing_mortgage_lender_directory_section','missing_mortgage_lender_directory_page',
        'missing_savings_provider_directory_section','missing_savings_provider_directory_page',
        'missing_health_insurer_directory_section','missing_health_insurer_directory_page',
        'dead_internal_link_live','canonical_mismatch','structured_data_invalid',
        'head_essentials_missing','sitemap_entry_dead_live'
    ] LOOP
        IF checks ? nm THEN
            RAISE EXCEPTION '434: check % is already enabled - partially applied or applied twice; inspect before re-running', nm;
        END IF;
    END LOOP;
END $do$;

-- ── Apply ──────────────────────────────────────────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,run_checks,config,checks}',
      (default_config#>'{workflow,steps,run_checks,config,checks}')
        || '["missing_mortgage_lender_directory_section", "missing_mortgage_lender_directory_page", "missing_savings_provider_directory_section", "missing_savings_provider_directory_page", "missing_health_insurer_directory_section", "missing_health_insurer_directory_page", "dead_internal_link_live", "canonical_mismatch", "structured_data_invalid", "head_essentials_missing", "sitemap_entry_dead_live"]'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'completeness-discovery-agent' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND NOT (default_config#>'{workflow,steps,run_checks,config,checks}')
          ? 'missing_mortgage_lender_directory_section';

-- ── Verify in-transaction (DO/RAISE - a SELECT cannot stop the COMMIT) ─────
DO $do$
DECLARE
    checks jsonb;
    nm text;
BEGIN
    SELECT default_config#>'{workflow,steps,run_checks,config,checks}' INTO checks
    FROM agent_definitions
    WHERE type = 'completeness-discovery-agent' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    FOREACH nm IN ARRAY ARRAY[
        'missing_mortgage_lender_directory_section','missing_mortgage_lender_directory_page',
        'missing_savings_provider_directory_section','missing_savings_provider_directory_page',
        'missing_health_insurer_directory_section','missing_health_insurer_directory_page',
        'dead_internal_link_live','canonical_mismatch','structured_data_invalid',
        'head_essentials_missing','sitemap_entry_dead_live'
    ] LOOP
        IF NOT (checks ? nm) THEN
            RAISE EXCEPTION '434 verify: check % missing after update', nm;
        END IF;
    END LOOP;

    IF jsonb_array_length(checks) <> 43 THEN
        RAISE EXCEPTION '434 verify: checks array is % entries, expected 43 (32 measured pre-apply + 11) - the array drifted between measurement and apply; re-measure', jsonb_array_length(checks);
    END IF;
END $do$;

COMMIT;
