-- SQL_p17_home_browse_by_category.sql — webdesign.co.uk
--
-- Fix the two near-identical "What's here" sections on the home page, and give
-- the four category promises somewhere real to land.
--
-- THE PROBLEM, measured. The home page carries TWO `info-card-grid` components at
-- positions 2 and 3, both with the eyebrow "What's here" and near-identical
-- titles/subtitles — a generation artefact, and the owner will read it as
-- sloppiness the moment he looks:
--
--   pos 2  "What's here" / "Tools and guides for people who build websites"
--   pos 3  "What's here" / "Tools and guides for web practitioners"
--
-- They also differ in quality, which decides the fix. **Position 3 is the good
-- one**: every card points at a real, specific tool (`smart-contrast`,
-- `layout-generator`, `css-variables`, `fluid-typography`). **Position 2 is the
-- weak one**: five of its six cards point at the same generic `/tools/index.html`,
-- because it promises categories — "Colour and contrast tools", "Typography and
-- spacing tools", "Accessibility checks" — that exist as prose and nowhere else.
--
-- WHY NOT DELETE POSITION 2, AND WHY NOT BUILD CATEGORY PAGES. Deleting it loses
-- the browse-by-subject job entirely. Building four category pages is the answer
-- the earlier handoff proposed, and it is the expensive one: four new pages to
-- generate, verify and maintain, for navigation the site can already express.
--
-- **The tools index ALREADY has six real categories** as `<h2>` headings, and they
-- did not carry anchors. Adding six `id`s (done in `gqls/sites`, same change) turns
-- an existing page into a navigable one at no content cost. So position 2 becomes
-- a genuine "browse by category" section and position 3 keeps featuring specific
-- tools — two distinct jobs instead of two copies of one.
--
-- THE COUNTS ARE MEASURED, NOT ESTIMATED, and reconciled before use — this site
-- has shipped invented figures before (D7, `bugs_open/043`). Attributing each
-- `index-card` to the most recent `<h2>` in document order:
--
--     Creative & Workflow      9      Security & Engineering   6
--     Design & Visuals        20      Accessibility & UX       3
--     Performance & Code      12      AI, Growth & Algo       13
--                                                    TOTAL   63
--
-- 63 matches both the number of cards linked from the index and the number of tool
-- directories on disk. A first attempt gave 62 — a regex that required `<h3>` to
-- follow the anchor immediately missed one card — and the discrepancy is the only
-- reason it was caught. **A total that does not reconcile is the signal; publish
-- nothing until it does.**
--
-- BOTH SURFACES, as always: `content_data` is what a future render reads,
-- `rendered_html` is what assemble republishes. Fixing one alone changes nothing
-- a visitor sees.
--
-- ARTEFACT, NOT PROPERTY. Regenerating the home page or the tools index rebuilds
-- them from upstream and both changes are lost, because nothing upstream knows
-- about the categories. Same standing caveat as SQL_p10 and SQL_p11.

\set ON_ERROR_STOP on

BEGIN;

-- position 2 -> browse by category, pointing at the six anchors that now exist
UPDATE page_components pc
   SET content_data = jsonb_build_object(
         'section_eyebrow',  'Browse by category',
         'section_title',    'Sixty-three tools, grouped by what you are doing',
         'section_subtitle', 'Every tool runs in your browser. Nothing to install, no account, nothing uploaded.',
         'cards', $cards$[
           {"icon":"🎨","title":"Design & Visuals","body":"Twenty tools for colour, gradients, shadows, clip paths, favicons and image work.","link_url":"/tools/index.html#design-visuals","link_label":"20 tools"},
           {"icon":"🧠","title":"AI, Growth & Algo","body":"Thirteen tools for prompts, token costs, A/B significance and recommendation maths.","link_url":"/tools/index.html#ai-growth-algo","link_label":"13 tools"},
           {"icon":"⚡","title":"Performance & Code","body":"Twelve tools for minifying, optimising assets, cleaning JSON and inspecting regex.","link_url":"/tools/index.html#performance-code","link_label":"12 tools"},
           {"icon":"⚙️","title":"Creative & Workflow","body":"Nine tools for layout, design tokens, mind maps, pasteboards and flat-file content.","link_url":"/tools/index.html#creative-workflow","link_label":"9 tools"},
           {"icon":"🔒","title":"Security & Engineering","body":"Six tools for CSP headers, JWT inspection, password entropy and PII redaction.","link_url":"/tools/index.html#security-engineering","link_label":"6 tools"},
           {"icon":"♿","title":"Accessibility & UX","body":"Three tools for ARIA labels, focus rings and touch-target sizing.","link_url":"/tools/index.html#accessibility-ux","link_label":"3 tools"}
         ]$cards$::jsonb)
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/index.html' AND pc.position = 2;

-- position 3 keeps its (good) specific-tool cards; only the duplicated header goes
UPDATE page_components pc
   SET content_data = pc.content_data
        || jsonb_build_object(
             'section_eyebrow',  'Start here',
             'section_title',    'A few of the most-used tools',
             'section_subtitle', 'Plain-English guides sit alongside them, written for people who build sites for a living.')
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/index.html' AND pc.position = 3;

DO $verify$
DECLARE v2 jsonb; v3 jsonb; v_links int; v_dupe int;
BEGIN
    SELECT pc.content_data INTO v2 FROM page_components pc JOIN pages p ON p.id=pc.page_id
      JOIN sites s ON s.id=p.site_id
     WHERE s.domain='webdesign.co.uk' AND p.url='/index.html' AND pc.position=2;
    SELECT pc.content_data INTO v3 FROM page_components pc JOIN pages p ON p.id=pc.page_id
      JOIN sites s ON s.id=p.site_id
     WHERE s.domain='webdesign.co.uk' AND p.url='/index.html' AND pc.position=3;

    IF jsonb_array_length(v2->'cards') <> 6 THEN
        RAISE EXCEPTION 'position 2 should have 6 category cards, has %', jsonb_array_length(v2->'cards');
    END IF;
    IF jsonb_array_length(v3->'cards') < 1 THEN
        RAISE EXCEPTION 'position 3 lost its cards — the || merge overwrote instead of adding';
    END IF;

    -- the duplication that motivated this file must actually be gone
    IF (v2->>'section_eyebrow') = (v3->>'section_eyebrow') THEN
        RAISE EXCEPTION 'both sections still share the eyebrow "%"', v2->>'section_eyebrow';
    END IF;

    -- every category card must point at an anchor, not the bare index
    SELECT count(*) INTO v_links
      FROM jsonb_array_elements(v2->'cards') c
     WHERE c->>'link_url' LIKE '/tools/index.html#%';
    IF v_links <> 6 THEN
        RAISE EXCEPTION 'only % of 6 category cards carry an anchor', v_links;
    END IF;

    -- and no two cards may point at the same place (the defect being fixed)
    SELECT count(*) - count(DISTINCT c->>'link_url') INTO v_dupe
      FROM jsonb_array_elements(v2->'cards') c;
    IF v_dupe <> 0 THEN
        RAISE EXCEPTION '% duplicate link targets among the category cards', v_dupe;
    END IF;

    RAISE NOTICE 'positions 2 and 3 now differ; 6 anchored category cards, all distinct.';
    RAISE NOTICE 'NOT DONE: rendered_html still holds the OLD markup — assemble republishes STORED html. Re-render the home page, then read the live page.';
END
$verify$;

COMMIT;
