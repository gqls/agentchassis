-- SQL_p10_fix_dead_homepage_links.sql — webdesign.co.uk
--
-- The owner found that no link on the home page to a tool or a guide worked.
-- Measured live 2026-07-27: 10 of 13 hrefs on https://webdesign.co.uk/ returned
-- 404. The only survivors were the three nav links. EVERY content link on the
-- page was dead — all 12 cards across two `info-card-grid` components.
--
-- SCOPE, checked before writing rather than assumed:
--   * page_components with dead links: the home page ONLY (the other 97 pages
--     match none of the bad patterns);
--   * site_components (header/footer/head): CLEAN — every internal href already
--     uses a full `/index.html` path.
-- So this file touches exactly two rows.
--
-- TWO INDEPENDENT FAULTS, which is why no single substitution covers them:
--
-- 1. INVENTED SLUGS. `/tools/colour-contrast-checker` and
--    `/tools/css-layout-generator` name pages that do not exist; the real pages
--    are `smart-contrast` and `layout-generator`. `/tools/spacing-scale-calculator`
--    and `/tools/typography-scale` name tools that do not exist in ANY form among
--    the 63 built. The slugs are absent from `cmd/webdesignport`, so they came
--    from generation, not the port — `bugs_open/092`'s mechanism (the writer gets
--    no link constraints, so the model invents).
--
-- 2. WRONG PATH SHAPE. `/tools`, `/guides` and the four category links have no
--    `/index.html`. The sites are served from an S3-compatible bucket behind
--    Cloudflare, and an object store does not resolve directory indexes. Measured:
--        /tools/ 404   /tools 404   /tools/smart-contrast/ 404
--        /tools/smart-contrast/index.html 200
--    This is fleet-wide, not a site misconfiguration: /about/ and /about 404 on
--    relojistas.com, robot-hands.com and gaswholesalers.com too. The platform-side
--    half of this (NormalizePagePath treats `/tools` as equivalent to
--    `/tools/index.html`, so the deploy gate and the audit both call it valid) is
--    a separate code change; this file only stops the bleeding on one page.
--
-- WHY DATA AND NOT A RE-RENDER. The hrefs are structured fields
-- (`content_data->'cards'->[]->'link_url'`), not prose. Correcting them is
-- deterministic. Re-running the writer would re-enter the code path that invented
-- them in the first place.
--
-- BOTH FIELDS ARE UPDATED, and this is load-bearing. `content_data` is what a
-- future render reads; `rendered_html` is what is actually served. The standing
-- landmine is that assemble republishes STORED rendered_html, so fixing
-- content_data alone would change nothing a visitor sees. The same replacement
-- list is applied to both so they cannot drift apart.
--
-- COPY IS CORRECTED WHERE IT DESCRIBED A TOOL WE DO NOT HAVE. Repointing a card
-- titled "Spacing scale calculator" at a design-token tool would fix the 404 and
-- leave the card lying about the destination — the same class of defect, quieter.
-- The two replacement descriptions were taken from the live tools:
--   css-variables    "CSS Variable Architect — Define your design tokens once,
--                     generate a scalable theme file instantly."
--   fluid-typography "Fluid Typography Composer — ... traditionally responsive
--                     design uses breakpoints where text jumps abruptly ..."
--
-- KNOWN RESIDUE, deliberately NOT fixed here — both are content decisions:
--   (a) The two components are near-duplicates: same eyebrow ("What's here"),
--       near-identical titles/subtitles, and overlapping cards ("Front-end
--       guides"/"Practical front-end guides", "Full tool library"/"Find the tool
--       you need"). Two "What's here" sections on one home page is a visible
--       defect, but removing a section is the owner's call, not a link fix.
--   (b) NO CATEGORY PAGES EXIST. After this file, five of the six cards in the
--       first component point at /tools/index.html, because there is nothing more
--       specific to point at and the index carries a global search. That is
--       functional and honest but redundant. Building real category pages
--       (colour / CSS / typography / accessibility) is the better answer.

\set ON_ERROR_STOP on

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. content_data — operate on the JSON text so one list covers every card.
--    Patterns include the surrounding quotes, which is what makes them exact:
--    "/tools" cannot match inside "/tools/colour" because of the closing quote.
-- ---------------------------------------------------------------------------
UPDATE page_components pc
   SET content_data = replace(replace(replace(replace(replace(replace(replace(
                      replace(replace(replace(replace(replace(replace(replace(
                          pc.content_data::text,
                          '"/tools/colour-contrast-checker"',  '"/tools/smart-contrast/index.html"'),
                          '"/tools/css-layout-generator"',     '"/tools/layout-generator/index.html"'),
                          '"/tools/spacing-scale-calculator"', '"/tools/css-variables/index.html"'),
                          '"/tools/typography-scale"',         '"/tools/fluid-typography/index.html"'),
                          '"/tools/accessibility"',            '"/tools/index.html"'),
                          '"/tools/typography"',               '"/tools/index.html"'),
                          '"/tools/colour"',                   '"/tools/index.html"'),
                          '"/tools/css"',                      '"/tools/index.html"'),
                          '"/guides"',                         '"/learn/index.html"'),
                          '"/tools"',                          '"/tools/index.html"'),
                          'Spacing scale calculator',          'CSS variable architect'),
                          'Builds a modular spacing scale from a base unit and ratio. Outputs CSS custom properties you can paste directly.',
                          'Define your design tokens once and generate a scalable CSS theme file, with custom properties you can paste straight in.'),
                          'Typography scale tool',             'Fluid typography composer'),
                          'Calculates a type scale from a chosen ratio and base size. Shows the values in rem, px and a copy-ready CSS block.',
                          'Builds type that scales smoothly between screen sizes using clamp(), instead of jumping at breakpoints. Outputs copy-ready CSS.'
                      )::jsonb,
       updated_at = now()
  FROM pages p
 WHERE p.id = pc.page_id
   AND p.site_id = '6b49db8e-d447-4467-8277-4f3018af9897'
   AND p.name = 'index'
   AND pc.slot_name = 'info-card-grid';

-- ---------------------------------------------------------------------------
-- 2. rendered_html — the same corrections against href="..." and the copy.
-- ---------------------------------------------------------------------------
UPDATE page_components pc
   SET rendered_html = replace(replace(replace(replace(replace(replace(replace(
                       replace(replace(replace(replace(replace(replace(replace(
                           pc.rendered_html,
                           'href="/tools/colour-contrast-checker"',  'href="/tools/smart-contrast/index.html"'),
                           'href="/tools/css-layout-generator"',     'href="/tools/layout-generator/index.html"'),
                           'href="/tools/spacing-scale-calculator"', 'href="/tools/css-variables/index.html"'),
                           'href="/tools/typography-scale"',         'href="/tools/fluid-typography/index.html"'),
                           'href="/tools/accessibility"',            'href="/tools/index.html"'),
                           'href="/tools/typography"',               'href="/tools/index.html"'),
                           'href="/tools/colour"',                   'href="/tools/index.html"'),
                           'href="/tools/css"',                      'href="/tools/index.html"'),
                           'href="/guides"',                         'href="/learn/index.html"'),
                           'href="/tools"',                          'href="/tools/index.html"'),
                           'Spacing scale calculator',               'CSS variable architect'),
                           'Builds a modular spacing scale from a base unit and ratio. Outputs CSS custom properties you can paste directly.',
                           'Define your design tokens once and generate a scalable CSS theme file, with custom properties you can paste straight in.'),
                           'Typography scale tool',                  'Fluid typography composer'),
                           'Calculates a type scale from a chosen ratio and base size. Shows the values in rem, px and a copy-ready CSS block.',
                           'Builds type that scales smoothly between screen sizes using clamp(), instead of jumping at breakpoints. Outputs copy-ready CSS.'
                       ),
       updated_at = now()
  FROM pages p
 WHERE p.id = pc.page_id
   AND p.site_id = '6b49db8e-d447-4467-8277-4f3018af9897'
   AND p.name = 'index'
   AND pc.slot_name = 'info-card-grid';

-- ---------------------------------------------------------------------------
-- 3. Verify — assert against the REAL pages table, not against a restatement
--    of the substitution list. A check that only re-asserts what step 1 wrote
--    would pass even if every target were fictional.
-- ---------------------------------------------------------------------------
DO $verify$
DECLARE
    v_site   uuid := '6b49db8e-d447-4467-8277-4f3018af9897';
    v_bad    int;
    v_html   int;
    v_cards  int;
BEGIN
    -- (a) every link_url in the cards must resolve to a real, deployed page row
    SELECT count(*) INTO v_bad
      FROM pages p
      JOIN page_components pc ON pc.page_id = p.id,
      LATERAL jsonb_array_elements(pc.content_data->'cards') c
     WHERE p.site_id = v_site AND p.name = 'index' AND pc.slot_name = 'info-card-grid'
       AND NOT EXISTS (
           SELECT 1 FROM pages t
            WHERE t.site_id = v_site
              AND t.url = c->>'link_url'
              AND t.deployed_at IS NOT NULL
       );
    IF v_bad > 0 THEN
        RAISE EXCEPTION '% card link(s) still do not resolve to a deployed page', v_bad;
    END IF;

    -- (b) no dead href shape survives in what is actually served
    SELECT count(*) INTO v_html
      FROM pages p JOIN page_components pc ON pc.page_id = p.id
     WHERE p.site_id = v_site AND p.name = 'index'
       AND pc.rendered_html ~ 'href="/(guides|tools)("|/[a-z-]+")';
    IF v_html > 0 THEN
        RAISE EXCEPTION '% component(s) still carry an extensionless internal href', v_html;
    END IF;

    -- (c) nothing was lost: still 12 cards across the two components
    SELECT count(*) INTO v_cards
      FROM pages p JOIN page_components pc ON pc.page_id = p.id,
      LATERAL jsonb_array_elements(pc.content_data->'cards') c
     WHERE p.site_id = v_site AND p.name = 'index' AND pc.slot_name = 'info-card-grid';
    IF v_cards <> 12 THEN
        RAISE EXCEPTION 'expected 12 cards, found % — the replacement damaged the JSON', v_cards;
    END IF;

    RAISE NOTICE 'all 12 home page cards now resolve to deployed pages; no extensionless hrefs remain';
    RAISE NOTICE 'NEXT: redeploy the page, then verify against the LIVE URL — a DB row is not a deploy';
END
$verify$;

COMMIT;
