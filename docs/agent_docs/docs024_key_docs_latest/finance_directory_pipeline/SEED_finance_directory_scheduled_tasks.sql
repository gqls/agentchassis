-- SEED — finance directory discovery tasks (Phase B, DIR-001): three weekly
-- discovery sweeps, one per kind, all driving finance-directory-researcher.
--
-- Three TASKS on one agent, sharing ONE concurrency_group — the exact
-- adoption/protocol pattern and its stated rationale: the three searches
-- want different words, and a single weekly slot would starve two of them.
--
-- ⚠ ENABLED = FALSE ON PURPOSE — see the researcher seed's header. The
-- registration-time non-price allowlist ships in the Phase B binary; running
-- research on the old binary would enforce the owner's compliance ruling by
-- prompt alone, which the council explicitly objected to (corr 69a619e6).
-- ENABLE ONLY AFTER the pod-grep gate passes on BOTH chassis replicas:
--   kubectl -n ai-persona-system exec <pod> -- grep -aq "per kind (Phase B kind-scoped keys)" /proc/1/exe
-- (with an absent-string control), then:
--   UPDATE scheduled_tasks SET enabled = true
--   WHERE name IN ('mortgage-lender-directory-discovery',
--                  'savings-provider-directory-discovery',
--                  'health-insurer-directory-discovery');
-- Force an immediate first run per task with:
--   UPDATE scheduled_tasks SET last_triggered_at = NULL WHERE name = '<task>';
--
-- No finance freshness task is added: the existing daily
-- model-directory-freshness task's refresh_directory_claims step runs with
-- EMPTY config, which sweeps every kind (loadDueDirectoryClaims: `$1 = ''
-- OR de.kind = $1`) — verify at enable time rather than trusting this
-- comment, then leave it be. A second kindless sweep would double-fetch.

BEGIN;

INSERT INTO scheduled_tasks (
    name, description, interval_seconds, target_agent_type, target_topic,
    input_data, concurrency_group, max_concurrent, enabled, timeout_seconds
)
SELECT
    'mortgage-lender-directory-discovery',
    'Finance directory (Phase B, DIR-001): discover UK mortgage lenders and register their cited NON-PRICE facts (FCA reference, regulator status, product types, lender type, established year) via finance-directory-researcher.',
    604800,
    'finance-directory-researcher',
    'system.agent.generic.requests',
    jsonb_build_object(
        'research_query',
        'UK mortgage lenders FCA authorised: banks, building societies and specialist lenders; residential, buy-to-let and later-life product ranges; regulator status and firm reference numbers'
    ),
    'finance-directory-discovery',
    1,
    false,
    900
WHERE NOT EXISTS (SELECT 1 FROM scheduled_tasks WHERE name = 'mortgage-lender-directory-discovery');

INSERT INTO scheduled_tasks (
    name, description, interval_seconds, target_agent_type, target_topic,
    input_data, concurrency_group, max_concurrent, enabled, timeout_seconds
)
SELECT
    'savings-provider-directory-discovery',
    'Finance directory (Phase B, DIR-001): discover UK savings providers and register their cited NON-PRICE facts (FCA reference, regulator status, product types, FSCS protection scheme, established year) via finance-directory-researcher.',
    604800,
    'finance-directory-researcher',
    'system.agent.generic.requests',
    jsonb_build_object(
        'research_query',
        'UK savings account providers: banks and building societies, FSCS protection, product types (easy access, fixed term, ISA), regulator status and firm reference numbers'
    ),
    'finance-directory-discovery',
    1,
    false,
    900
WHERE NOT EXISTS (SELECT 1 FROM scheduled_tasks WHERE name = 'savings-provider-directory-discovery');

INSERT INTO scheduled_tasks (
    name, description, interval_seconds, target_agent_type, target_topic,
    input_data, concurrency_group, max_concurrent, enabled, timeout_seconds
)
SELECT
    'health-insurer-directory-discovery',
    'Finance directory (Phase B, DIR-001): discover UK private medical insurers and register their cited NON-PRICE facts (FCA reference, regulator status, cover types, underwriter, established year) via finance-directory-researcher.',
    604800,
    'finance-directory-researcher',
    'system.agent.generic.requests',
    jsonb_build_object(
        'research_query',
        'UK private medical insurance providers: insurers and underwriters, cover types (inpatient, outpatient, mental health, dental), regulator status and firm reference numbers'
    ),
    'finance-directory-discovery',
    1,
    false,
    900
WHERE NOT EXISTS (SELECT 1 FROM scheduled_tasks WHERE name = 'health-insurer-directory-discovery');

COMMIT;

-- ── Post-apply verification ────────────────────────────────────────────────
--   SELECT name, enabled, interval_seconds, concurrency_group FROM scheduled_tasks
--   WHERE name LIKE '%-directory-discovery' ORDER BY name;
-- Expect the three finance rows enabled=f alongside the model/adoption rows.
