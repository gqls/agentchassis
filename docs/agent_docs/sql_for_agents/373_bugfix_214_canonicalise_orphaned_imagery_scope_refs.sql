-- FILE: docs/agent_docs/sql_for_agents/373_bugfix_214_canonicalise_orphaned_imagery_scope_refs.sql
--
-- bugs_open/214 — repair the imagery scope_refs that name a page nothing can
-- resolve, because WriteSitePlanAction renamed the page and never carried the
-- reference across.
--
-- THE CODE FIX IS THE FIX. This file only repairs rows already written; the
-- write path is fixed in write_site_plan_imagery_scope.go and no new orphan can
-- be minted once that ships. Applying this without the code fix buys one plan
-- generation, because the next replan re-mints the same refs.
--
-- WHAT IT TOUCHES, PRECISELY. Only rows that are BOTH (a) unresolvable today —
-- the scope_ref's page part matches no pages.name on that site, which is the
-- vocabulary every consumer actually joins against — and (b) rescuable, i.e.
-- their own plan holds a page named "<ref>-index" which DOES resolve in pages.
-- A row that works today fails predicate (a) and cannot be touched by this
-- script. A row with no canonical variant fails (b) and is deliberately left
-- alone rather than repointed at a guess.
--
-- The "-index" join is deliberate rather than a reimplementation of
-- CanonicalisePage in SQL: the section-index collapse is the only family
-- present in the live 10, and a second implementation of page canonicalisation
-- is exactly the drift class this bug is an instance of.
--
-- EXPECTED OUTCOME, measured 2026-08-10 on current plans:
--   before: 10 invisible rows (9 page-scope, ... see the census below)
--   after :  1 invisible row  (mortgagecalculator.co.uk 'tools-index', which
--                              has no canonical variant — left for a human)
-- The DO block at the end ABORTS the transaction on any other outcome. That
-- matters in both directions: 10 remaining means the script was inert, and 0
-- remaining means it touched the row it was told to leave. A verify block of
-- bare SELECTs cannot stop a COMMIT — ON_ERROR_STOP ignores a non-empty result
-- — so the guard is a RAISE, not a SELECT.
--
-- Rollback: 373_..._ROLLBACK.sql, generated from this script's RETURNING output.

\set ON_ERROR_STOP on

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. Pre-flight census. Read this before letting the transaction commit.
-- ---------------------------------------------------------------------------
\echo '=== BEFORE: imagery rows on current plans that resolve to no page ==='
WITH cur AS (
    SELECT sp.id AS plan_id, sp.site_id, si.domain
      FROM site_plans sp
      JOIN sites si ON si.id = sp.site_id
     WHERE sp.is_current
)
SELECT cur.domain, spi.scope, spi.scope_ref, spi.key,
       (SELECT spp.name FROM site_plan_pages spp
         WHERE spp.plan_id = spi.plan_id
           AND spp.name = split_part(spi.scope_ref, ':', 1) || '-index') AS canonical_candidate
  FROM site_plan_imagery spi
  JOIN cur ON cur.plan_id = spi.plan_id
 WHERE spi.scope IN ('page', 'section')
   AND NOT EXISTS (SELECT 1 FROM pages p
                    WHERE p.site_id = cur.site_id
                      AND p.name = split_part(spi.scope_ref, ':', 1))
 ORDER BY cur.domain, spi.scope_ref, spi.key;

-- ---------------------------------------------------------------------------
-- 2. Page scope. scope_ref is the bare page name.
--    chk_scope_ref_consistency forbids a colon here; a canonical page name has
--    none, and the WHERE clause below can only ever select one.
-- ---------------------------------------------------------------------------
\echo '=== repairing page-scope refs (save this output for the rollback) ==='
WITH cur AS (
    SELECT sp.id AS plan_id, sp.site_id FROM site_plans sp WHERE sp.is_current
)
UPDATE site_plan_imagery spi
   SET scope_ref = spp.name
  FROM cur
  JOIN site_plan_pages spp ON spp.plan_id = cur.plan_id
 WHERE spi.plan_id = cur.plan_id
   AND spi.scope   = 'page'
   AND spp.name    = spi.scope_ref || '-index'
   -- (a) unresolvable today
   AND NOT EXISTS (SELECT 1 FROM pages p
                    WHERE p.site_id = cur.site_id AND p.name = spi.scope_ref)
   -- (b) the canonical variant does resolve
   AND EXISTS     (SELECT 1 FROM pages p
                    WHERE p.site_id = cur.site_id AND p.name = spp.name)
   -- (c) the unique index (plan_id, scope, COALESCE(scope_ref,''), key) must
   --     survive — never collapse onto an identity that already exists
   AND NOT EXISTS (SELECT 1 FROM site_plan_imagery x
                    WHERE x.plan_id = spi.plan_id
                      AND x.scope   = spi.scope
                      AND COALESCE(x.scope_ref, '') = spp.name
                      AND x.key     = spi.key)
RETURNING spi.id, spi.key, spi.scope_ref AS new_ref;

-- ---------------------------------------------------------------------------
-- 3. Section scope. scope_ref is "<page>:<ordinal>".
--    The ordinal is NOT touched — no consumer parses it, the correct value is
--    unknowable here, and rewriting it would risk the unique index for no
--    behavioural gain (see write_site_plan_imagery_scope.go's header). Taking
--    the remainder from the FIRST colon onward preserves both the ordinal and
--    the colon chk_scope_ref_consistency requires of a section ref.
-- ---------------------------------------------------------------------------
\echo '=== repairing section-scope refs (save this output for the rollback) ==='
WITH cur AS (
    SELECT sp.id AS plan_id, sp.site_id FROM site_plans sp WHERE sp.is_current
)
UPDATE site_plan_imagery spi
   SET scope_ref = spp.name || substr(spi.scope_ref, strpos(spi.scope_ref, ':'))
  FROM cur
  JOIN site_plan_pages spp ON spp.plan_id = cur.plan_id
 WHERE spi.plan_id = cur.plan_id
   AND spi.scope   = 'section'
   AND strpos(spi.scope_ref, ':') > 0
   AND spp.name    = split_part(spi.scope_ref, ':', 1) || '-index'
   AND NOT EXISTS (SELECT 1 FROM pages p
                    WHERE p.site_id = cur.site_id
                      AND p.name = split_part(spi.scope_ref, ':', 1))
   AND EXISTS     (SELECT 1 FROM pages p
                    WHERE p.site_id = cur.site_id AND p.name = spp.name)
   AND NOT EXISTS (SELECT 1 FROM site_plan_imagery x
                    WHERE x.plan_id = spi.plan_id
                      AND x.scope   = spi.scope
                      AND COALESCE(x.scope_ref, '') =
                          spp.name || substr(spi.scope_ref, strpos(spi.scope_ref, ':'))
                      AND x.key     = spi.key)
RETURNING spi.id, spi.key, spi.scope_ref AS new_ref;

-- ---------------------------------------------------------------------------
-- 4. The guard. A RAISE, not a SELECT — ON_ERROR_STOP does not fire on a
--    non-empty result set, so a verify block of SELECTs cannot stop a COMMIT.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
    remaining int;
BEGIN
    SELECT count(*) INTO remaining
      FROM site_plan_imagery spi
      JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current
     WHERE spi.scope IN ('page', 'section')
       AND NOT EXISTS (SELECT 1 FROM pages p
                        WHERE p.site_id = sp.site_id
                          AND p.name = split_part(spi.scope_ref, ':', 1));

    IF remaining <> 1 THEN
        RAISE EXCEPTION
            'bugfix 214 backfill: expected exactly 1 unresolvable imagery row to remain (mortgagecalculator tools-index, which has no canonical variant), got %. '
            'More than 1 means the repair was inert or the fleet changed since 2026-08-10; 0 means it touched the row it was told to leave. '
            'Re-run the pre-flight census, reconcile, and only then re-apply.', remaining;
    END IF;

    RAISE NOTICE 'bugfix 214 backfill: OK — 1 unresolvable row remains, as expected.';
END $$;

COMMIT;

-- ---------------------------------------------------------------------------
-- 5. Post-apply verification (run AFTER commit; not part of the transaction).
--
--    This proves the repair at the CONSUMER's own predicate rather than at the
--    row: for each repaired page-scope hero, the join plan_sections actually
--    performs must now return the asset.
-- ---------------------------------------------------------------------------
\echo '=== AFTER: repaired rows now resolve through the consumer join ==='
-- SELECT si.domain, p.name AS page, spi.scope_ref, a.asset_key
--   FROM site_plan_imagery spi
--   JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
--   JOIN sites si ON si.id = sp.site_id
--   JOIN pages p ON p.site_id = sp.site_id AND p.name = spi.scope_ref
--   JOIN assets a ON a.site_id = sp.site_id AND a.asset_key = spi.key AND a.status = 'active'
--  WHERE spi.scope = 'page' AND spi.kind = 'hero'
--    AND si.domain IN ('gamesdesign.co.uk','fundamentallyai.com','mortgagecalculator.co.uk')
--  ORDER BY si.domain, p.name;
