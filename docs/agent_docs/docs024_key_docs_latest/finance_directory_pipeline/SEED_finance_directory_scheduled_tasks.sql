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
--
-- REVISED 2026-08-15 (B4 supervised run 1): the three research_query values below were
-- re-aimed at NAMED firms and updated on the LIVE rows the same day. The originals
-- described the market ("UK mortgage lenders FCA authorised: banks, building societies
-- and specialist lenders; …") and retrieved 3/4 market-level pages, which produced
-- category-shaped entities. Old values preserved in
-- portfolio_positioning/NOTES 2026-08-15. This file now matches live so a re-apply
-- cannot resurrect the market-shaped queries. Companion prompt fix: sql_for_agents/423.

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
        'named UK mortgage lenders — individual banks, building societies and specialist lenders (Building Societies Association and UK Finance member firms): each firm''s mortgage product range (residential, buy-to-let, later life), FCA authorisation statement and firm reference number'
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
        'named UK savings providers — individual banks and building societies: each firm''s savings products (easy access, fixed term, cash ISA), FSCS protection statement, FCA authorisation and firm reference number'
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
        'named UK private medical insurers — individual insurance companies and underwriters: each firm''s cover options (inpatient, outpatient, mental health, dental), FCA/PRA authorisation and firm reference number'
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
