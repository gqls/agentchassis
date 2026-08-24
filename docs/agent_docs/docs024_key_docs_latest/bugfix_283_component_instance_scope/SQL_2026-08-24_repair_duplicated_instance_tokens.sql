-- SQL_2026-08-24_repair_duplicated_instance_tokens.sql
--
-- bugs_open/383 — repair the three pages whose stored rendered_html carries a
-- DUPLICATED per-instance element-id token.
--
-- WHY A RERENDER REPAIRS THIS EVEN THOUGH THE FIX IS NOT ROLLED YET, which is the
-- one thing to understand before running it. The defect lives in the two paths
-- that render ONE section at a time. `page_rerender` does not use them: it walks
-- the whole page through `loadStoredSections` + `InstanceCounter`, which has
-- always assigned occurrences correctly. That is how 9 of 12 pages were repaired
-- on 2026-08-23, on unfixed code.
--
-- ⚠ SO THIS REPAIRS, BUT IT DOES NOT HOLD. Until commit 364e80b7f rolls, the next
-- per-section render of a repaired page — a build, a `content_rewrite`, a section
-- edit — puts the collision straight back. Three of the twelve repaired on
-- 2026-08-23 were re-collided within hours by an unrelated lane's backfill. The
-- roll is what makes the repair durable; this file is what clears the standing
-- damage. Running it before the roll is deliberate (owner decision 2026-08-24)
-- and its value is bounded accordingly — do not read a green re-check as "fixed".
--
-- Population, measured live 2026-08-24 (re-measure before running; a page can
-- join or leave this set on any build):
--   gaswholesalers.com /pricing-transparency.html        c-generic-text-block ×2
--   gaswholesalers.com /wholesale-pricing-explained.html c-generic-text-block ×2
--   vetcomparison.uk   /how-it-works.html                c-generic-text-block ×2
--
-- ⚠ THE REASON IS LOAD-BEARING AND AN INVENTED ONE SILENTLY DOES NOTHING.
-- page-rerender's `check_rerender_mode` is a conditional over an ALLOW-LIST of
-- exactly five reasons (live config 2026-08-24): image_landed,
-- section_data_resolved, cta_links_stale, template_changed, literal_markdown.
-- A reason outside that list — or none at all — takes `else_step: render_page`,
-- which is `rerender_single_page`: "simple concatenation, no template
-- re-rendering", i.e. it RE-SHIPS THE STORED BYTES. For this repair that would
-- complete successfully and change nothing, which is the exact shape this lane
-- keeps filing bugs about. My first draft of this file invented
-- `instance_scope_383` and would have done precisely that.
--
-- `template_changed` is the right one of the five, and the reason is measured:
--   * it is the ONLY one with no Go branch keyed on it — it merely selects the
--     sections branch, which is what re-renders every stored section through
--     `loadStoredSections` + `InstanceCounter` (the canonical walk);
--   * `cta_links_stale` triggers a CTA recompute (rerender_page_sections_action.go:533)
--     with its own documented clobber landmine — avoided deliberately;
--   * `section_data_resolved` / `image_landed` become SCOPED to a single component
--     when a component_id is present (create_rerender_items_action.go:219);
--   * the estate already uses it in this sense: "generic pages take the conversion
--     on their template_changed rerender" (fix_component_template_action.go:1687).
--
-- Dedup: `idx_swi_dedup` is UNIQUE on (site_id, item_key) WHERE status is not one
-- of the 7 terminal values, and the key MUST be the canonical
-- `page_rerender_<page_name>_<site_id>_<reason>` that `pageRerenderItemKey` builds
-- — the mode is IN the key so a reason-less item can never suppress a reason-bearing
-- one (bugs_open/024 defect 6, the six-month-invisible delivery blocker). These
-- pages carry older `page_rerender` rows, but all are `unresolved`, which IS
-- terminal, so they do not hold the slot. Checked 2026-08-24: no non-terminal row
-- exists at any of the three keys.

BEGIN;

INSERT INTO site_work_items (
    site_id, page_id, affected_url, item_type, item_key, source, severity,
    summary, suggested_action, resolution_path,
    handler_agent, priority, status, approval_mode, created_by, pipeline, spec
)
SELECT
    p.site_id,
    p.id,
    p.url,
    'page_rerender',
    'page_rerender_' || regexp_replace(regexp_replace(p.url, '^/', ''), '\.html$', '')
        || '_' || p.site_id::text || '_template_changed',
    'bugs_open/383',
    'medium',
    'Rerender page: ' || regexp_replace(regexp_replace(p.url, '^/', ''), '\.html$', '')
        || ' — duplicated per-instance element id (bugs_open/383)',
    'Whole-page section re-render. The canonical walk (loadStoredSections + InstanceCounter) '
        || 'gives each instance its real occurrence, which the per-section render paths could '
        || 'not. No content change is intended: this re-renders stored content_data through '
        || 'the same templates.',
    'rerender',
    'page-rerender',
    80,
    'detected',
    'auto',
    'bugs_open/383',
    'build',
    jsonb_build_object(
        'reason',    'template_changed',
        'page_id',   p.id::text,
        'page_name', regexp_replace(regexp_replace(p.url, '^/', ''), '\.html$', ''),
        'filename',  regexp_replace(p.url, '^/', ''),
        'domain',    s.domain
    )
FROM pages p JOIN sites s ON s.id = p.site_id
WHERE p.id IN (
    '0ca123e2-2377-4ac3-990c-c1d255db8c3a',  -- gaswholesalers /pricing-transparency.html
    '08e34599-a130-417d-90d9-76370484d03a',  -- gaswholesalers /wholesale-pricing-explained.html
    '68307089-e741-44af-bbb8-93f2be49e537'   -- vetcomparison  /how-it-works.html
)
ON CONFLICT DO NOTHING;

-- VERIFY — a DO/RAISE block, not a bare SELECT.
-- `ON_ERROR_STOP` ignores a non-empty result set, so a verify made of SELECTs
-- cannot stop the COMMIT (RFC_006's lesson; 581's worked example).
DO $$
DECLARE
    filed   int;
    claimed int;
    routed  int;
BEGIN
    SELECT count(*) INTO filed
      FROM site_work_items
     WHERE created_by = 'bugs_open/383'
       AND status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled');
    IF filed <> 3 THEN
        RAISE EXCEPTION 'expected 3 live repair items, found % — the population moved or dedup '
                        'suppressed an insert. Re-measure before retrying.', filed;
    END IF;

    -- CONTROL 1, and it is the half that could come out otherwise: each item must
    -- point at a page that ACTUALLY carries a duplicated token right now. If a page
    -- repaired itself between measurement and run, this fails and the file is STALE.
    WITH live AS (
        SELECT pc.page_id, pc.rendered_html FROM page_components pc
         WHERE pc.build_status IS DISTINCT FROM 'removed'
    ), toks AS (
        SELECT page_id, (regexp_matches(rendered_html, 'id="(c-[^"]*)"', 'g'))[1] AS tok FROM live
    ), dup AS (
        SELECT page_id FROM toks GROUP BY page_id, tok HAVING count(*) > 1
    )
    SELECT count(DISTINCT w.page_id) INTO claimed
      FROM site_work_items w JOIN dup d ON d.page_id = w.page_id
     WHERE w.created_by = 'bugs_open/383';
    IF claimed <> 3 THEN
        RAISE EXCEPTION 'only % of the 3 filed items point at a page that still carries a '
                        'duplicated token — this file is STALE, re-measure', claimed;
    END IF;

    -- CONTROL 2 — the one that catches the mistake this file was rewritten for:
    -- a reason outside page-rerender's five-value allow-list routes to assemble-only,
    -- which re-ships the broken bytes and completes successfully.
    SELECT count(*) INTO routed
      FROM site_work_items
     WHERE created_by = 'bugs_open/383'
       AND spec->>'reason' IN ('image_landed','section_data_resolved','cta_links_stale',
                               'template_changed','literal_markdown');
    IF routed <> 3 THEN
        RAISE EXCEPTION 'only % of 3 items carry a reason page-rerender will route to the '
                        'SECTIONS branch — the rest would assemble stored HTML and repair nothing', routed;
    END IF;

    RAISE NOTICE 'VERIFY: PASS — 3 items filed, all on pages that currently carry a duplicated '
                 'token, all routed to the sections branch';
END $$;

COMMIT;
