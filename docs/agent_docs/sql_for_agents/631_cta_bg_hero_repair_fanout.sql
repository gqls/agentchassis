-- 631 — fan 619's HERO repair out to the pages that are actually broken
--
-- bugs_open/398. RUN AFTER 619.
--
-- WHY THIS FILE EXISTS AT ALL, and it is the half 619 was missing. Nothing files
-- `template_changed` re-renders for a template edited by SQL: that fan-out lives
-- in `component-template-fixer.create_rerender`, keyed on the component_id the
-- FIXER just changed, and the fixer is an LLM repair agent, not something you
-- invoke to publish a hand-authored template. So a migration that edits
-- html_template ships NOTHING — every page keeps serving its stored
-- rendered_html, indefinitely, with a green status and no error anywhere
-- (bugs_open/283 §13; the same objection was raised and upheld against a
-- 2026-08-06 council round on this very builder).
--
-- CAUGHT AT THE ARTEFACT, not by reasoning: after 619 applied, a page-rerender
-- dispatch reported COMPLETED for /about.html, /approach.html and /contact.html
-- and all three still served the broken declaration, because a reason-less
-- rerender re-assembles STORED component HTML. `page_components.updated_at` for
-- those rows still read 2026-08-17 / 08-24 / 08-23. `template_changed` is in
-- page-rerender's `check_rerender_mode` reason list (migration 460), so ONE
-- re-render with that reason re-renders the new template. No LLM on this path.
-- Shape copied from 615, including the two filters the live fixer query lacks.
--
-- ── SCOPE: 9 PAGES, NOT 55, AND THE DIFFERENCE IS DELIBERATE ───────────────
-- [MEASURED 2026-08-25] 55 active pages still hold a stored hero carrying the
-- old two-declaration block. Only **9** of them are BROKEN — the ones whose site
-- theme puts a GRADIENT in --color-cta-bg, where the second declaration is
-- invalid at computed-value time and the band paints nothing:
--     finetuning.uk 3 · gaswholesalers.com 3 · robot-hands.com 3
-- On the other 46 the palette is a solid colour, so that declaration is VALID and
-- currently renders a subtle gradient sheen. Re-rendering them today would remove
-- that sheen — a cosmetic change to sites nobody reported, while a council round
-- on this very change is still open. They converge on their next natural
-- re-render. Re-run the census in bugs_open/398 §3 before widening this.
--
-- ⚠ TWO OF THE THREE SITES BELONG TO OTHER LANES (gaswholesalers.com,
-- robot-hands.com). They are included because the defect is live on them — an
-- invisible page heading — and the repair is a template re-render that does not
-- regenerate copy. Owner ruling 2026-07-29 §3: tell the other consumers. That is
-- owed as a CONTRIB in each lane's directory, not discharged by this comment.
--
-- NOT INCLUDED: any page carrying `call-to-action` or `tool-cta`. Their repair
-- needs `--color-cta-bg-ink`, which does not exist until the next chassis roll,
-- so fanning them out now would re-render them once for nothing and again after
-- the roll. One fan-out after the roll, filed separately.
--
-- Rollback: 631_..._ROLLBACK.sql cancels any of these items still OPEN.

BEGIN;

-- Guard: 619 must have landed, or this fans out the OLD template.
DO $$
DECLARE n_broken int;
BEGIN
    SELECT count(*) INTO n_broken FROM content_components
     WHERE name IN ('about-hero','contact-hero','services-hero')
       AND html_template LIKE '%linear-gradient(135deg, var(--color-cta-bg%';
    IF n_broken <> 0 THEN
        RAISE EXCEPTION '631: % hero component(s) still carry the invalid declaration — apply 619 first, or this fan-out re-renders the unchanged template', n_broken;
    END IF;
END $$;

CREATE TEMP TABLE _631_targets AS
WITH gradient_theme AS (
    SELECT source_domain
      FROM css_themes
     WHERE source_domain IS NOT NULL
       AND (regexp_match(css_content, '--color-cta-bg:\s*linear-gradient\([^,]+,\s*(#[0-9A-Fa-f]{3,8})'))[1] IS NOT NULL
)
SELECT DISTINCT p.id AS page_id, p.name AS page_name, p.site_id, s.domain
  FROM page_components pc
  JOIN pages p  ON p.id = pc.page_id
  JOIN sites s  ON s.id = p.site_id
  JOIN content_components cc ON cc.id = pc.component_id
  JOIN gradient_theme g ON g.source_domain = s.domain
 WHERE cc.name IN ('about-hero','contact-hero','services-hero')
   AND pc.rendered_html LIKE '%linear-gradient(135deg, var(--color-cta-bg%'  -- still stale
   AND pc.build_status <> 'removed'
   AND p.status = 'active'                                   -- bugs_open/098
   AND COALESCE(p.rebuild_policy, 'generic') <> 'owned';     -- save_sections refuses owned pages

INSERT INTO site_work_items (site_id, page_id, source, pipeline, item_type, severity, summary,
                             priority, handler_agent, status, created_by, spec, item_key, batch_id)
SELECT t.site_id, t.page_id, 'side_effect', 'build', 'page_rerender', 'medium',
       'Rerender ' || t.page_name || ' — hero band no longer substitutes a gradient into a <color> position (bugs_open/398, migration 619)',
       80, 'page-rerender', 'triaged', 'bugfix_398_cta_bg_hero_fanout',
       jsonb_build_object('reason', 'template_changed',
                          'page_id', t.page_id::text,
                          'page_name', t.page_name,
                          'domain', t.domain,
                          'cause', 'template_changed:cta_bg_hero_bands'),
       'page_rerender_' || t.page_name || '_' || t.site_id::text || '_template_changed',
       gen_random_uuid()
  FROM _631_targets t
 WHERE NOT EXISTS (
       SELECT 1 FROM site_work_items w
        WHERE w.site_id = t.site_id
          AND w.item_type = 'page_rerender'
          AND w.spec->>'page_id' = t.page_id::text
          AND w.spec->>'reason' = 'template_changed'
          AND w.status IN ('detected','triaged','claimed'))
ON CONFLICT DO NOTHING;

DO $$
DECLARE n_targets int; n_filed int; n_bad int;
BEGIN
    SELECT count(*) INTO n_targets FROM _631_targets;
    IF n_targets = 0 THEN
        RAISE EXCEPTION '631 verify: the target set is EMPTY — that contradicts the 9-page census this file was written against; re-measure before assuming it is already done';
    END IF;

    SELECT count(*) INTO n_filed FROM site_work_items
     WHERE created_by = 'bugfix_398_cta_bg_hero_fanout';

    IF EXISTS (
        SELECT 1 FROM _631_targets t
         WHERE NOT EXISTS (
            SELECT 1 FROM site_work_items w
             WHERE w.site_id = t.site_id
               AND w.item_type = 'page_rerender'
               AND w.spec->>'page_id' = t.page_id::text
               AND w.spec->>'reason' = 'template_changed'
               AND w.status IN ('detected','triaged','claimed'))) THEN
        RAISE EXCEPTION '631 verify: a target page has no open template_changed item — the INSERT dropped rows';
    END IF;

    SELECT count(*) INTO n_bad
      FROM site_work_items w JOIN pages p ON p.id = w.page_id
     WHERE w.created_by = 'bugfix_398_cta_bg_hero_fanout'
       AND (p.status <> 'active' OR COALESCE(p.rebuild_policy,'generic') = 'owned');
    IF n_bad > 0 THEN
        RAISE EXCEPTION '631 verify: % filed item(s) sit on an archived or owned page', n_bad;
    END IF;

    RAISE NOTICE '631: % target page(s), % item(s) filed under bugfix_398_cta_bg_hero_fanout', n_targets, n_filed;
END $$;

COMMIT;
