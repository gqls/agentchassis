-- 434 ROLLBACK - surgical inverse of
-- 434_enable_finance_directory_and_structural_checks_b3f.sql: removes exactly
-- the eleven check names 434 appended to completeness-discovery-agent's
-- run_checks array, preserving every other entry and its order.
--
-- Deliberately NOT restore-from-backup (the backup row snapshots the whole
-- config; restoring it would clobber any edit another session has made since
-- 434 applied). Guards mirror the forward file: exactly-one-active-row, and
-- ALL eleven names must be present - a partial set means someone else edited
-- the array since 434 and a human should look before rolling back.

SELECT snapshot_agent('completeness-discovery-agent', '434_enable_finance_directory_and_structural_checks_b3f_ROLLBACK.sql: pre-rollback');

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
        RAISE EXCEPTION '434 ROLLBACK: completeness-discovery-agent does not have exactly one active row (found %) - resolve the ambiguity before editing', n;
    END IF;

    SELECT count(*) INTO n FROM agent_definitions_backup
    WHERE snapshot_reason = '434_enable_finance_directory_and_structural_checks_b3f_ROLLBACK.sql: pre-rollback'
      AND type = 'completeness-discovery-agent';
    IF n < 1 THEN
        RAISE EXCEPTION '434 ROLLBACK: no pre-rollback backup row - snapshot_agent did not run';
    END IF;

    SELECT default_config#>'{workflow,steps,run_checks,config,checks}' INTO checks
    FROM agent_definitions
    WHERE type = 'completeness-discovery-agent' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF checks IS NULL OR jsonb_typeof(checks) <> 'array' THEN
        RAISE EXCEPTION '434 ROLLBACK: no checks array at workflow.steps.run_checks.config.checks - the row has drifted';
    END IF;
    FOREACH nm IN ARRAY ARRAY[
        'missing_mortgage_lender_directory_section','missing_mortgage_lender_directory_page',
        'missing_savings_provider_directory_section','missing_savings_provider_directory_page',
        'missing_health_insurer_directory_section','missing_health_insurer_directory_page',
        'dead_internal_link_live','canonical_mismatch','structured_data_invalid',
        'head_essentials_missing','sitemap_entry_dead_live'
    ] LOOP
        IF NOT (checks ? nm) THEN
            RAISE EXCEPTION '434 ROLLBACK: check % is not enabled - 434 is not applied in full, or the array was edited since; inspect before rolling back', nm;
        END IF;
    END LOOP;
END $do$;

-- ── Apply inverse: rebuild the array without the eleven names ──────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,run_checks,config,checks}',
      COALESCE(
        (SELECT jsonb_agg(elem ORDER BY ord)
         FROM jsonb_array_elements(default_config#>'{workflow,steps,run_checks,config,checks}')
              WITH ORDINALITY AS t(elem, ord)
         WHERE elem #>> '{}' NOT IN (
             'missing_mortgage_lender_directory_section','missing_mortgage_lender_directory_page',
             'missing_savings_provider_directory_section','missing_savings_provider_directory_page',
             'missing_health_insurer_directory_section','missing_health_insurer_directory_page',
             'dead_internal_link_live','canonical_mismatch','structured_data_invalid',
             'head_essentials_missing','sitemap_entry_dead_live'
         )),
        '[]'::jsonb
      )
    ),
    updated_at = NOW()
WHERE type = 'completeness-discovery-agent' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND (default_config#>'{workflow,steps,run_checks,config,checks}')
      ? 'missing_mortgage_lender_directory_section';

-- ── Verify in-transaction (DO/RAISE) ───────────────────────────────────────
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
        IF checks ? nm THEN
            RAISE EXCEPTION '434 ROLLBACK verify: check % still present after removal', nm;
        END IF;
    END LOOP;

    -- The pre-434 baseline was 32; anything else means the array was edited
    -- since 434 in a way the name-guards above could not see (an ADDITION by
    -- another lane survives correctly - this check would then read >32 and
    -- refuse, which is the safe failure: inspect, then re-run with the count
    -- expectation corrected).
    IF jsonb_array_length(checks) <> 32 THEN
        RAISE EXCEPTION '434 ROLLBACK verify: checks array is % entries, expected 32 - another lane has edited the array since 434; verify their entries survived, then adjust this expectation and re-run', jsonb_array_length(checks);
    END IF;
END $do$;

COMMIT;
