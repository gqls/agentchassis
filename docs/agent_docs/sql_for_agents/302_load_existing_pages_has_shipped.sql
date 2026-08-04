-- 302_load_existing_pages_has_shipped.sql
--
-- bugs_open/185 tranche 3 — the planner's empty-page gate asked build_status
-- when it meant "has this page been served".
--
-- WHY. reconcilePlanWithRealised (v3_site_actions.go) gates sectionless realised
-- pages on realisedPageHasShipped (formerly realisedPageIsBuilt): a page that
-- SERVED sectionless renders through another subsystem and must not receive an
-- injected generic layout; a page that never shipped is merely uncomposed and
-- must stay composable. The old test — build_status = 'deployed' — conflates the
-- two for needs_rebuild rows: a shipped needs_rebuild page (still serving its
-- previous artefact; bugs_closed/037's population, 35 of 46 rows carry a
-- deployed_at) reads as "never composed", so a re-plan may inject a layout into
-- a live page. The Go side now reads `has_shipped`; this migration surfaces it.
--
-- The predicate below is PageHasShippedPredicateFor("p")'s exact output, copied
-- verbatim. A migration is SQL text and cannot CALL the Go builder, so the tie
-- between them is a TEST, not a compiler:
-- datahelpers.TestMigration302CarriesTheCanonicalPredicateVerbatim reads this
-- file and asserts the builder's output appears in it exactly once. Edit either
-- side without the other and that test goes red (proved by mutating both sides).
-- If this file is ever archived or renumbered, move that assertion onto its
-- successor rather than deleting it — the LIVE agent_definitions row was written
-- from this text, and the test is the only thing tying the row to the builder.
--
-- The predicate is the estate's ONE definition of "has been served" —
-- datahelpers.NeverDeployedPagePredicate, negated — NOT a status list. Naming
-- needs_rebuild in the predicate is the trap this whole bug is about
-- (datahelpers/links_deployment_test.go forbids it; dartsonline's brands-index
-- is a needs_rebuild page that must NOT be gated, and deployed_at is what tells
-- it apart from one that shipped).
--
-- SCOPE: build-site-planner ONLY. content-gap-planner also carries a step named
-- load_existing_pages, but it is a registered Go action with a different config
-- shape (checked live 2026-08-03: its config is {"site_id": ...}, no query key)
-- and does not feed reconcilePlanWithRealised.
--
-- ORDERING IS FREE, same contract as 173: this migration is inert on its own (a
-- chassis without the Go change ignores the extra column), and the Go change
-- falls back to the old build_status test when has_shipped is absent. DB-first
-- and image-first are both safe; the fix is complete once both have landed.
--
-- NOTE the query below is written against the LIVE query as read from
-- agent_definitions on 2026-08-03 (post-173, post the site_has_no_current_plan
-- rename). It is NOT built on 173's own text, whose CASE column still said
-- adoption_locked.

BEGIN;

SELECT snapshot_agent('build-site-planner',
                      'load_existing_pages surfaces has_shipped for the empty-page gate (bugs_open/185 tranche 3)');

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_existing_pages,config,query}',
        to_jsonb($Q$
SELECT p.name, p.page_type, p.url, p.title, p.nav_label, p.in_header, p.in_footer,
       p.sections, p.meta_description, p.nav_order, p.build_status,
       NOT (p.deployed_at IS NULL AND COALESCE(p.build_status, '') <> 'deployed') AS has_shipped,
       CASE
         WHEN NOT EXISTS (
             SELECT 1 FROM site_plans sp
             WHERE sp.site_id = p.site_id AND sp.is_current = true
         ) THEN true
         ELSE false
       END AS site_has_no_current_plan
FROM pages p
WHERE p.site_id = $1 AND p.status = 'active'
ORDER BY p.name
$Q$::text),
        true
    ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND deleted_at IS NULL AND is_active = true
  AND COALESCE(is_snapshot, false) = false;

-- Verification INSIDE the transaction, and it RAISEs rather than SELECTs — a
-- verify block of SELECTs cannot stop the COMMIT (fleet memory: ON_ERROR_STOP
-- ignores a non-empty result set).
--
-- > **THIS BLOCK IS UNSOUND FOR ONE CASE, AND THE SQL BELOW IS LEFT AS IT RAN.**
-- > Three council seats (editquality, debug_historian, guardian — corr c881ef22)
-- > flagged it against the known trap: a `#>>` path that is ABSENT yields NULL,
-- > and every comparison below is then NULL, so no `IF` fires and the block
-- > passes SILENTLY. Reproduced against the live row on 2026-08-04 by pointing
-- > the SELECT at a nonexistent path: `q IS NULL: t`, `would any guard have
-- > FIRED? f`. **So these guards prove the query's CONTENT when the path
-- > resolves, and prove nothing at all when it does not** — exactly the case a
-- > path typo produces.
-- >
-- > The migration itself is fine: the path was right, and the column was verified
-- > independently against the live row and against production data (brands-index
-- > computes has_shipped=f). The file is annotated rather than edited because it
-- > is the record of what ran, and it is RECORDED in schema_migrations.
-- >
-- > **IF YOU COPY THIS BLOCK — and it is otherwise a good pattern — add the
-- > null-guard FIRST:**
-- >     IF q IS NULL THEN
-- >       RAISE EXCEPTION 'NNN: jsonb path did not resolve — nothing was verified';
-- >     END IF;
-- > Without it, a verify block reads as belt-and-braces while being a decoration
-- > for the one failure it most needs to catch.
DO $$
DECLARE
  n int;
  q text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
  WHERE type = 'build-site-planner' AND deleted_at IS NULL AND is_active = true
    AND COALESCE(is_snapshot, false) = false;
  IF n <> 1 THEN
    RAISE EXCEPTION '302: expected exactly 1 live build-site-planner row, found %', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,load_existing_pages,config,query}' INTO q
  FROM agent_definitions
  WHERE type = 'build-site-planner' AND deleted_at IS NULL AND is_active = true
    AND COALESCE(is_snapshot, false) = false;

  IF (length(q) - length(replace(q, 'has_shipped', ''))) / length('has_shipped') <> 1 THEN
    RAISE EXCEPTION '302: has_shipped must appear exactly once in the query, got: %', q;
  END IF;
  IF q NOT LIKE '%site_has_no_current_plan%' THEN
    RAISE EXCEPTION '302: the site_has_no_current_plan column was lost — the query was overwritten from a stale base';
  END IF;
  IF q NOT LIKE '%p.build_status,%' THEN
    RAISE EXCEPTION '302: build_status was lost — the fallback and 173''s consumers still need it';
  END IF;
END $$;

COMMIT;

-- ---------------------------------------------------------------------------
-- Post-apply checks (run by hand; the DO block above is what actually guards)
-- ---------------------------------------------------------------------------
-- 1. The step query carries has_shipped exactly once and still runs:
--    SELECT p.name, p.build_status,
--           NOT (p.deployed_at IS NULL AND COALESCE(p.build_status, '') <> 'deployed') AS has_shipped
--    FROM pages p WHERE p.site_id = '<site>' AND p.status = 'active' ORDER BY p.name;
--
-- 2. Functional, once the chassis with realisedPageHasShipped is rolled: re-plan
--    a site carrying a SHIPPED needs_rebuild sectionless page and confirm the
--    validate log line "forced deployed sectionless page back to empty" fires
--    for it; dartsonline's brands-index (never shipped) must keep taking the
--    LLM's proposal.
