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
--
-- ⚠ QUERIES MUST STAY UNDER 200 BYTES. web_search's query_from path drops any query
-- >= 200 chars as a "likely LLM error message" (web_search_action.go extractSearchQuery,
-- the len(queryStr) < 200 sanity check) and the run then FAILS with "search query not
-- found - check 'query', 'topic', or 'query_field' config" — an error that misdirects to
-- config keys, not to length. B4 run 2 (42f72cd9, 2026-08-15 14:41Z) failed exactly this
-- way on a 275-byte query; the first re-aimed wording was shortened to fit the same day.
--
-- REVISED 2026-08-15 (later session, query iteration 2): regulator vocabulary ("FCA
-- authorisation, firm reference number") in the QUERY pulls the REGULATOR's own pages —
-- B4 run 3 (ffc22155) scraped 4/4 FCA/market pages and registered nothing. Regulator
-- words belong to EXTRACTION (the FCA-footer facts live on firm pages); the query must
-- hunt FIRMS, so all three queries below use the membership-list shape proven by runs
-- 4-5 ("list of ... named member firms (<trade body>) and each firm's <products>").
-- Mortgage was proven first, then savings/health mirrored (live rows updated same day).

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
        'list of UK mortgage lenders and building societies: named member firms (Building Societies Association, UK Finance) and each firm''s mortgage range: residential, buy-to-let, later life'
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
        'list of UK savings account providers: named member banks and building societies (Building Societies Association, UK Finance) and each firm''s accounts: easy access, fixed term, cash ISA'
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
        'list of UK private medical insurance providers: named member insurers (Association of British Insurers) and each firm''s cover: inpatient, outpatient, mental health, dental'
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
