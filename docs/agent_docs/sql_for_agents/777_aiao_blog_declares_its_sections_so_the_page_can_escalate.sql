-- 777 — ai-agent-orchestration.com /blog: declare the sections the page already has
--
-- Raised by the bugs_open/384 lane (2026-09-04), verified first-hand by the
-- site_ai_agent_orchestration lane the same day.
--
-- THE HOLE. /blog's `hero` and `call-to-action` slots carry content_data = NULL.
-- rerender_page_sections_action.go:428-459 pre-checks every stored section and
-- refuses to render the page rather than blank it ("no stored content_data").
-- Having refused, it calls escalateRerenderToWriter, which asks
-- pageSectionsSatisfiable — and that fails ALL THREE of its sources for this
-- page, so the disposition is `skipped_sectionless_page` and NO needs_page item
-- is raised. The run then reports COMPLETED. The page cannot be re-rendered and
-- nothing asks for it to be rebuilt, silently.
--
-- WHAT THIS MIGRATION DOES, AND WHAT IT DELIBERATELY DOES NOT.
-- It repairs source 3 only: `pages.sections`, which is [] while the page has
-- THREE built component slots. That is a demonstrable data defect on this site,
-- not a convention — 39 of 45 active pages declare their sections, and every
-- sibling checked declares exactly its slot_names in position order
-- (index, about, contact, services all match their slots verbatim).
--
-- ⚠ IT DOES NOT REPAIR THE MISSING CONTENT, and must not be read as doing so
-- (the 384 lane's own correction, which it had to make against itself): the
-- empty `sections` is not why the page fails to render, it only suppresses the
-- fallback. This converts a SILENT SKIP into a `needs_page` item, and the repair
-- then depends on that item draining. The primary defect — three required
-- source:"llm" fields with no stored value (hero.headline,
-- call-to-action.headline, call-to-action.primary_cta) — is for the framework to
-- write, never for a migration to hand-author (CLAUDE.md, owner ruling
-- 2026-08-04).
--
-- ⚠ SOURCES 1 AND 2 ARE DEAD SITE-WIDE AND ARE NOT TOUCHED HERE: this site has
-- ZERO current `site_plans` rows and no `site_specs` aspect='site_plan', so
-- declaredPageSections' plan lookups and pageInCurrentPlan return nothing for
-- EVERY page on it. 39 pages are safe only because they carry their own
-- `pages.sections`. Five more pages sit on empty `sections` today
-- (`ai-readiness-quiz`, `roi-estimator`, `tool-ai-agent-roi-estimator`,
-- `tool-build-vs-buy-analyzer`, `tool-llm-cost-calculator`) and are one content
-- loss away from this same hole. Deliberately out of scope: four are single-slot
-- tool pages where the declaration may legitimately be empty, and `roi-estimator`
-- has NO components at all, which is a different defect.
--
-- DERIVED, NOT HARDCODED: the value is built from the page's own slots in
-- position order, so it cannot go stale against a slot that is added or renamed
-- between writing and applying. The guards pin what it is allowed to produce.

BEGIN;

-- ---------------------------------------------------------------- guard: pre
DO $guard$
DECLARE
    n_page      int;
    n_empty     int;
    n_slots     int;
    v_slots     text;
BEGIN
    SELECT count(*) INTO n_page FROM pages
     WHERE site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da' AND name = 'blog' AND status = 'active';
    IF n_page <> 1 THEN
        RAISE EXCEPTION '777 ABORT: expected exactly 1 active blog page on this site, found %.', n_page;
    END IF;

    SELECT count(*) INTO n_empty FROM pages
     WHERE site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da' AND name = 'blog' AND status = 'active'
       AND jsonb_typeof(sections) = 'array' AND jsonb_array_length(sections) = 0;
    IF n_empty <> 1 THEN
        RAISE NOTICE '777: blog.sections is no longer an empty array — another session has repaired it. Nothing to do.';
        RETURN;
    END IF;

    SELECT count(*), string_agg(pc.slot_name, ',' ORDER BY pc.position)
      INTO n_slots, v_slots
      FROM page_components pc JOIN pages p ON p.id = pc.page_id
     WHERE p.site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da' AND p.name = 'blog';

    -- Pin the shape this migration was written against. A different slot set is
    -- a different page and wants a human, not a derived write.
    IF v_slots IS DISTINCT FROM 'hero,blog-listing,call-to-action' THEN
        RAISE EXCEPTION '777 ABORT: blog slots are "%" (n=%), not the "hero,blog-listing,call-to-action" this migration was written against. Re-read the page before writing its declaration.', v_slots, n_slots;
    END IF;
END
$guard$;

-- ------------------------------------------------------------------- the write
UPDATE pages p
SET sections = COALESCE((
        SELECT jsonb_agg(pc.slot_name ORDER BY pc.position)
        FROM page_components pc WHERE pc.page_id = p.id
    ), '[]'::jsonb)
WHERE p.site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da'
  AND p.name = 'blog' AND p.status = 'active'
  AND jsonb_typeof(p.sections) = 'array' AND jsonb_array_length(p.sections) = 0;

-- --------------------------------------------------------------- guard: post
DO $verify$
DECLARE
    v_declared text;
BEGIN
    SELECT sections::text INTO v_declared FROM pages
     WHERE site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da' AND name = 'blog' AND status = 'active';

    IF v_declared IS DISTINCT FROM '["hero", "blog-listing", "call-to-action"]' THEN
        RAISE EXCEPTION '777 ABORT: blog.sections is % after the update, not the expected three slots in position order; nothing committed.', v_declared;
    END IF;
    RAISE NOTICE '777 OK: blog declares %; the page can now escalate to the writer instead of skipping silently. The MISSING CONTENT is still missing — expect a needs_page item, not a repaired page.', v_declared;
END
$verify$;

COMMIT;
