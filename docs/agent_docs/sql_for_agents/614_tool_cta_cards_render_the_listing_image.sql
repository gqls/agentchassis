-- 614_tool_cta_cards_render_the_listing_image.sql
--
-- WHAT: the `tool-cta` component's card tiles gain a gated thumbnail, and the
-- component's schema DECLARES the `image` item key it is now rendering.
--
-- WHY (bugs_open/384 decision 4, OWNER RULING 2026-08-25). tool-cta's stored
-- item arrays carry an `image` key that the component never declared and never
-- rendered. It is there only because plan_sections stores the resolver's full
-- item map verbatim (plan_sections_action.go:2402) — nothing asked for it. That
-- made 14 entries on 2 sites (leopardess 9, finetuning 5, [MEASURED 2026-08-25])
-- STALE and invisible: the 384 seam deliberately skips consumers whose template
-- does not render the image, so nothing refreshes them.
--
-- THE FRAMING THAT DECIDED IT, and it is the useful part for the next reader:
-- a stored-but-unrendered key RE-STALES after any repair, because the resolver
-- always returns it and the seam always skips it. Only two states are stable —
-- the key is RENDERED, or it is NOT STORED. The owner chose RENDERED.
--
-- "NOT STORED" WAS MEASURED AND IS UNSAFE TODAY, so do not reach for it as the
-- tidier fix: of 28 live (component, query-array-field) declarations,
-- **17 render an item key their own schema omits** ([MEASURED 2026-08-25]) —
-- every directory listing renders `.url` without declaring it, and all three
-- blog listings declare no item keys at all. Projecting stored items down to
-- declared keys would blank live content on 17 of 28.
--
-- ── WHAT THIS COSTS, STATED BECAUSE IT IS BOUGHT DELIBERATELY ────────────────
-- Rendering `.image` puts tool-cta INSIDE queryresolve.PageListConsumerPages'
-- predicate (`cc.html_template ~* '\.image\y'`). So **40 pages across 10
-- domains** ([MEASURED 2026-08-25]) join the page-image consumer set, and every
-- card landing on those sites will file a page_rerender from now on.
-- (60 tool-cta instances exist; the consumer lookup excludes 16 on ARCHIVED
-- pages and 4 on `owned` pages, which is why the number is 40 and not 60 —
-- quote 40, not the instance count.) That is
-- exactly the cost the council's round-2 bound removed from the seam
-- (correlation 2005a846), and the owner has knowingly bought it back to make
-- the images visible. It is recorded here, in PBP-048 and in bugs_open/384 —
-- not only in a chat reply.
--
-- ── WHAT IT WILL ACTUALLY SHOW, RE-MEASURED AFTER THE CARDS LANDED ──────────
-- 228 tool-cta entries fleet-wide: **206 a purpose-built card crop, 0 heroes,
-- 22+20=42 nothing** ([MEASURED 2026-08-25, after the D1 derives]).
-- Before those derives it was 62 crops / 144 full-bleed page HEROES / 42
-- nothing — all 144 heroes on loancalculator.co.uk, whose 10 tool pages had no
-- card at all. The owner ruled "derive the missing cards first"; 10 of 10
-- landed and loancalculator now resolves 144/144 to crops.
-- loanandmortgagecalculator.co.uk's 12 entries stay blank: derive_card_asset
-- CROPS AN EXISTING HERO and those pages have none ("no hero asset to derive
-- from: no active page, content, or site hero" — the action's own words, in
-- the completed items' result). Blank is the gated branch, so they render
-- nothing rather than an empty box.
--
-- ── THE GATE IS LOAD-BEARING ────────────────────────────────────────────────
-- `{{if .image}}` wraps the whole element. An UNGATED template is one of the
-- two independent causes of an empty image attribute, and it is the one that
-- leaves a blank box when the value is legitimately absent — which is the case
-- for 42 of the 228 entries here, permanently. Same shape as the sibling
-- `tool-list` component, which shares this exact source
-- (query.pages_where_type:tool) and limit (6).
--
-- ── ONE DELIBERATE DEVIATION FROM THE SIBLING, and why ──────────────────────
-- tool-list renders `alt="{{.title}}"`. This uses `alt=""`. The tile here is
-- wrapped in an `<a>`, so alt text is concatenated into the LINK's accessible
-- name and a screen reader announces the title twice ("Compare Loans, Compare
-- Loans"). tool-list's img sits in an `<article>`, where that does not happen.
-- The image is decorative in this context — the adjacent <h3> already names the
-- destination — so the empty alt is correct here and is NOT drift.
--
-- ── APPLYING / ROLLING BACK ─────────────────────────────────────────────────
-- Pre-image is copied to content_components_bak_20260825_toolcta_image (the
-- estate's per-migration convention). Rollback: 614_..._ROLLBACK.sql restores
-- html_template and input_schema from that table.
--
-- The UPDATE is anchored on VERBATIM pre-image fragments and this file ABORTS
-- if either anchor is absent or occurs more than once — so a concurrent edit by
-- another session refuses this migration rather than being silently reverted.
-- Verification is DO/RAISE: a SELECT-only verify cannot stop the COMMIT.
--
-- AFTER APPLYING, the pages still serve their STORED rendered_html. The
-- re-render fan-out is the companion step — see the RUNBOOK section
-- "D3 — fan out the template_changed re-renders". A template change with no
-- fan-out ships nothing, silently, with a green status (bugs_open/283 §13).

BEGIN;

-- ── Pre-image ───────────────────────────────────────────────────────────────
DROP TABLE IF EXISTS content_components_bak_20260825_toolcta_image;
CREATE TABLE content_components_bak_20260825_toolcta_image AS
SELECT * FROM content_components WHERE name = 'tool-cta';

-- ── Guards: exactly one live row, each anchor exactly once ──────────────────
DO $$
DECLARE
    n_rows int;
    n_css  int;
    n_card int;
    tmpl   text;
BEGIN
    SELECT count(*) INTO n_rows FROM content_components
     WHERE name = 'tool-cta' AND is_active;
    IF n_rows <> 1 THEN
        RAISE EXCEPTION '614: expected exactly 1 active tool-cta component, found %', n_rows;
    END IF;

    SELECT html_template INTO tmpl FROM content_components
     WHERE name = 'tool-cta' AND is_active;

    IF tmpl ~ '\.image\y' THEN
        RAISE EXCEPTION '614: tool-cta already renders .image — already applied, or another session got there first';
    END IF;

    n_css := (length(tmpl) - length(replace(tmpl, '  .tool-cta-section .tool-cta-card-title {', '')))
             / length('  .tool-cta-section .tool-cta-card-title {');
    IF n_css <> 1 THEN
        RAISE EXCEPTION '614: the CSS anchor occurs % times, expected exactly 1 — the template has been edited; re-derive the anchor before applying', n_css;
    END IF;

    n_card := (length(tmpl) - length(replace(tmpl, '        <a href="{{.url}}" class="tool-cta-card">', '')))
              / length('        <a href="{{.url}}" class="tool-cta-card">');
    IF n_card <> 1 THEN
        RAISE EXCEPTION '614: the card-markup anchor occurs % times, expected exactly 1 — the template has been edited; re-derive the anchor before applying', n_card;
    END IF;
END $$;

-- ── 1. The thumbnail's CSS, inserted immediately before the card title rule ─
UPDATE content_components
   SET html_template = replace(
         html_template,
         '  .tool-cta-section .tool-cta-card-title {',
         '  .tool-cta-section .tool-cta-card-thumb {' || E'\n' ||
         '    display: block;' || E'\n' ||
         '    width: 100%;' || E'\n' ||
         '    aspect-ratio: 16 / 9;' || E'\n' ||
         '    object-fit: cover;' || E'\n' ||
         '    border-radius: calc(var(--border-radius) * 0.75);' || E'\n' ||
         '    margin-bottom: 0.25rem;' || E'\n' ||
         '  }' || E'\n' || E'\n' ||
         '  .tool-cta-section .tool-cta-card-title {'),
       updated_at = NOW()
 WHERE name = 'tool-cta' AND is_active;

-- ── 2. The gated <img>, first child of the card ─────────────────────────────
UPDATE content_components
   SET html_template = replace(
         html_template,
         '        <a href="{{.url}}" class="tool-cta-card">' || E'\n',
         '        <a href="{{.url}}" class="tool-cta-card">' || E'\n' ||
         '          {{if .image}}<img class="tool-cta-card-thumb" src="{{.image}}" alt="" loading="lazy">{{end}}' || E'\n'),
       updated_at = NOW()
 WHERE name = 'tool-cta' AND is_active;

-- ── 3. Declare the key the template now renders ─────────────────────────────
-- Without this, tool-cta joins the 17-of-28 declarations that render an item
-- key their schema omits — the exact mess this lane measured and recorded.
-- Spelling copied from the sibling tool-list: {"type": "image"}.
UPDATE content_components
   SET input_schema = jsonb_set(
         input_schema,
         '{fields,items,items,image}',
         '{"type": "image"}'::jsonb,
         true),
       updated_at = NOW()
 WHERE name = 'tool-cta' AND is_active;

-- ── Verify, or refuse the COMMIT ────────────────────────────────────────────
DO $$
DECLARE
    tmpl text;
    sch  jsonb;
    bak  text;
BEGIN
    SELECT html_template, input_schema INTO tmpl, sch
      FROM content_components WHERE name = 'tool-cta' AND is_active;
    SELECT html_template INTO bak
      FROM content_components_bak_20260825_toolcta_image WHERE is_active;

    -- The whole point: the consumer lookup's predicate must now match.
    IF tmpl !~* '\.image\y' THEN
        RAISE EXCEPTION '614 verify: tool-cta still does not render .image — PageListConsumerPages would not see it';
    END IF;
    IF tmpl NOT LIKE '%{{if .image}}%' THEN
        RAISE EXCEPTION '614 verify: the image is rendered UNGATED — 42 of 228 entries have no image and would get an empty box';
    END IF;
    IF tmpl NOT LIKE '%tool-cta-card-thumb {%' THEN
        RAISE EXCEPTION '614 verify: the thumbnail CSS rule was not inserted';
    END IF;

    -- Nothing else in the card may have been lost by the replace().
    IF tmpl NOT LIKE '%{{.title}}%' OR tmpl NOT LIKE '%{{.meta_description}}%'
       OR tmpl NOT LIKE '%{{.nav_label}}%' OR tmpl NOT LIKE '%{{.url}}%' THEN
        RAISE EXCEPTION '614 verify: a pre-existing card field vanished from the template';
    END IF;
    IF length(tmpl) <= length(bak) THEN
        RAISE EXCEPTION '614 verify: template did not grow (% -> %) — a replace() matched nothing or removed text', length(bak), length(tmpl);
    END IF;

    -- The schema must declare the new key AND keep the four it already had.
    IF NOT (sch #> '{fields,items,items}' ? 'image') THEN
        RAISE EXCEPTION '614 verify: input_schema does not declare the image item key';
    END IF;
    IF NOT (sch #> '{fields,items,items}' ?& array['url','title','nav_label','meta_description']) THEN
        RAISE EXCEPTION '614 verify: a pre-existing declared item key was lost';
    END IF;
    IF sch #>> '{fields,items,source}' IS DISTINCT FROM 'query.pages_where_type:tool' THEN
        RAISE EXCEPTION '614 verify: the items source changed — expected query.pages_where_type:tool, got %', sch #>> '{fields,items,source}';
    END IF;
END $$;

COMMIT;
