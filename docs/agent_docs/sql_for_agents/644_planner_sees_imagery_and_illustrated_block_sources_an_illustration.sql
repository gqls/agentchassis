-- 644 — teach the planner the word "image", and point Illustrated Text Block at
-- an ILLUSTRATION instead of the page hero.
--
-- Two halves of one defect, and they must ship together (see COUPLING below).
--
-- ── HALF 1: THE MISSING WORD ────────────────────────────────────────────────
-- The three planner menus (build-site-planner, content-gap-planner,
-- site-planner) describe each component to the model with
-- component_expresses(html_template, input_schema). That function derives four
-- tokens — html-block, list, table, items — and NONE of them is an image.
-- bugs_open/381 taught the planner to see lists, tables and item sets; imagery
-- was never added to the vocabulary.
--
-- The visible consequence: `Generic Text Block` and `Illustrated Text Block`
-- both read [html-block, list, table]. IDENTICAL. The planner is choosing
-- between two components it has been told are the same thing and picks the
-- plain one. [MEASURED 2026-08-26] Generic Text Block has 208 live instances
-- across 23 sites; Illustrated Text Block has 6, all on ONE site (apis.uk).
--
-- So this was never a missing component. It is a missing WORD.
--
-- ⚠ DERIVED FROM THE SCHEMA'S `source`, NOT FROM `<img` IN THE TEMPLATE, and
-- the difference is the whole precision of it. [MEASURED 2026-08-26] a template
-- grep reports 47 of 386 components — every header, hero and card thumbnail,
-- chrome the writer cannot influence. The schema predicate reports 14, of which
-- 11 are active and section-level, i.e. actually reachable in a menu.
--
-- `site_assets.logo` is the ONE exclusion, and it is a judgement worth stating
-- rather than burying: a logo is site identity, present on 11 components, and
-- is never the illustration a planner is reaching for when it wants an image.
-- Every other site_assets.* image-typed field counts, heroes included — a hero
-- component genuinely does offer a server-resolved image slot, and excluding it
-- would need a hand-kept deny-list, which is the drift class this estate keeps
-- filing.
--
-- ── HALF 2: THE SOURCE THAT RESOLVES TO THE PAGE HERO ───────────────────────
-- Illustrated Text Block's `image_url` declares source `site_assets.image`.
-- That path is NOT a literal asset key. imageryplan.imageRoleAliases maps
-- "image" -> "hero", and plan_sections' site_assets arm consults the alias when
-- the literal key misses — which it always does, because nothing ever populates
-- r.assets["image"] (the content_data fallback fills only "hero" and "logo").
--
-- So `site_assets.image` resolves to THE PAGE'S OWN HERO, always. Measured at
-- the artefact, not inferred: [MEASURED 2026-08-26] every populated
-- site_assets.image value in the estate is a hero asset — hero-about.jpg,
-- hero-home.jpg, content-hero-*.jpg — and 20 of 52 populated instances show an
-- image ALREADY displayed elsewhere on the same page. On
-- leopardessconsulting.co.uk /about.html the about-hero at position 1 and
-- content-block-about at position 2 carry the identical file.
--
-- Shipping half 1 alone would therefore have made the planner reach for a
-- component whose figure repeats the hero, mid-page, fleet-wide.
--
-- ⚠ AND THERE IS A LIVE LATENT REGRESSION THIS FIXES. apis.uk/index has an
-- active `hero_home` asset, so `site_assets.image` RESOLVES there — and live
-- resolution always beats the stored-content carry (plan_sections_action.go,
-- carryStored: "Live resolution always wins"). On the next plan_sections run
-- for that page, all six Illustrated Text Blocks would have their distinct
-- illustrations OVERWRITTEN with hero-home.jpg. The repoint below prevents it:
-- `site_assets.illustration` is not in the alias map, so it is a literal key,
-- it resolves nothing on apis.uk (whose illustration plan rows are scope=page,
-- while the resolver's section loop reads scope=section) — and carryStored then
-- preserves the six good values from the page's own deployed content_data.
--
-- `image_alt` is repointed to `llm` for a plainer reason: it is typed `text` but
-- sourced `site_assets.image`, so the resolver hands it the image URL. A screen
-- reader would read out the file path. [MEASURED 2026-08-26] the estate's
-- convention for alt text is source `llm` — 13 fields across 9 components — and
-- this field is the only one in the estate sourced otherwise. Its own
-- llm_guidance ("Describe what the image SHOWS") already assumes a model writes
-- it. This is the outlier being brought back to the convention, not a new idea.
--
-- ── ⚠ COUPLING: WHY THESE CANNOT BE TWO MIGRATIONS ──────────────────────────
-- The obvious narrow predicate is `source = 'site_assets.image'` (exact). Under
-- half 2 the field stops carrying that value, so the narrow predicate would
-- make Illustrated Text Block INVISIBLE again — the two halves would silently
-- cancel. The predicate is written over `site_assets.%` for that reason. Apply
-- them together or neither.
--
-- ── SCOPE / BLAST RADIUS ────────────────────────────────────────────────────
-- Changes what EVERY planner is shown, on every subsequent build, fleet-wide.
-- No Go change: DB config, live the moment it applies. Nothing re-renders on
-- its own; existing pages are untouched until something rebuilds them.
--
-- ⚠ WHAT THIS DOES NOT DO — say it out loud, because the ask was "imagery
-- between paragraphs" and this is only half an answer. It makes the illustrated
-- component SELECTABLE. It does not create a single asset. [MEASURED
-- 2026-08-26] the estate holds 206 active hero assets across 28 sites and 26
-- illustration assets across 5; only 4 section-scope illustration plan rows
-- across 3 sites exist for the resolver to read. Everywhere else `on_missing:
-- skip_field` means the section renders as plain prose, silently and correctly.
-- The supply question is the bigger half and is NOT addressed here.
--
-- REPLAY-SAFE: CREATE OR REPLACE and jsonb_set are both idempotent.

BEGIN;

-- Capture the CURRENT vocabulary for every component BEFORE anything changes,
-- so the verify block can assert the change is purely ADDITIVE against this
-- exact population. Deliberately NOT a hard-coded count: components are created
-- by other lanes continuously (five appeared during this migration's own
-- measurement run on 2026-08-26), so a literal N is a spurious failure waiting
-- to happen, while "no component lost a token" is true regardless of who adds
-- what in between.
CREATE TEMP TABLE _644_before ON COMMIT DROP AS
SELECT id,
       name,
       component_expresses(html_template, input_schema) AS tok
  FROM content_components;

-- ── HALF 1 ─────────────────────────────────────────────────────────────────
CREATE OR REPLACE FUNCTION public.component_expresses(p_html_template text, p_input_schema jsonb)
 RETURNS text[]
 LANGUAGE sql
 IMMUTABLE
AS $function$
  SELECT COALESCE(array_agg(x ORDER BY x), ARRAY[]::text[]) FROM (
    SELECT 'html-block'::text AS x WHERE EXISTS (
      SELECT 1 FROM jsonb_each(COALESCE(p_input_schema->'fields', '{}'::jsonb)) f
       WHERE f.value->>'source' = 'llm' AND f.value->>'type' = 'html')
    UNION
    SELECT 'list' WHERE p_html_template ~* '<(ul|ol)[\s>]' OR EXISTS (
      SELECT 1 FROM jsonb_each(COALESCE(p_input_schema->'fields', '{}'::jsonb)) f
       WHERE f.value->>'source' = 'llm' AND f.value->>'type' = 'html')
    UNION
    SELECT 'table' WHERE p_html_template ~* '<table[\s>]' OR EXISTS (
      SELECT 1 FROM jsonb_each(COALESCE(p_input_schema->'fields', '{}'::jsonb)) f
       WHERE f.value->>'source' = 'llm' AND f.value->>'type' = 'html')
    UNION
    SELECT 'items' WHERE p_html_template ~* '\{\{[-\s]*range' AND EXISTS (
      SELECT 1 FROM jsonb_each(COALESCE(p_input_schema->'fields', '{}'::jsonb)) f
       WHERE f.value->>'source' = 'llm' AND f.value->>'type' IN ('array', 'list'))
    UNION
    -- 644: a server-resolved image slot the component offers as CONTENT.
    -- site_assets.logo excluded: site identity, not illustration.
    SELECT 'image' WHERE EXISTS (
      SELECT 1 FROM jsonb_each(COALESCE(p_input_schema->'fields', '{}'::jsonb)) f
       WHERE f.value->>'source' LIKE 'site_assets.%'
         AND f.value->>'source' <> 'site_assets.logo'
         AND f.value->>'type' IN ('url', 'image', 'image_url'))
  ) s;
$function$;

-- ── HALF 2 ─────────────────────────────────────────────────────────────────
UPDATE content_components
   SET input_schema = jsonb_set(
                        jsonb_set(input_schema,
                          '{fields,image_url,source}', '"site_assets.illustration"'),
                          '{fields,image_alt,source}',  '"llm"'),
       updated_at = NOW()
 WHERE name = 'Illustrated Text Block'
   AND input_schema->'fields'->'image_url' IS NOT NULL
   AND input_schema->'fields'->'image_alt' IS NOT NULL;

-- ── VERIFY ─────────────────────────────────────────────────────────────────
-- DO/RAISE, never a bare SELECT: ON_ERROR_STOP does not abort a COMMIT on a
-- non-empty result set, so a SELECT-shaped verify block cannot stop a bad
-- migration (the RFC_006 lesson).
DO $$
DECLARE
    lost      integer;
    reshuffle integer;
    unearned  integer;
    gained    integer;
    gtb_same  boolean;
    itb       text[];
    repointed integer;
BEGIN
    -- (1) THE CONTROL THAT MATTERS: purely additive. No component may lose a
    -- token, and no component may change by anything other than gaining
    -- 'image'. A count of changed rows would NOT catch this — verified by
    -- mutation on 2026-08-26: an arm that also suppressed 'list' changed the
    -- same number of rows and was caught only by these two assertions.
    SELECT count(*) INTO lost
      FROM _644_before b JOIN content_components c ON c.id = b.id
     WHERE EXISTS (SELECT 1 FROM unnest(b.tok) t
                    WHERE t <> ALL (component_expresses(c.html_template, c.input_schema)));
    IF lost <> 0 THEN
        RAISE EXCEPTION '644: % component(s) LOST a capability token — not a widening, aborting', lost;
    END IF;

    SELECT count(*) INTO reshuffle
      FROM _644_before b JOIN content_components c ON c.id = b.id
     WHERE array_remove(component_expresses(c.html_template, c.input_schema), 'image')
           IS DISTINCT FROM array_remove(b.tok, 'image');
    IF reshuffle <> 0 THEN
        RAISE EXCEPTION '644: % component(s) changed by something other than gaining image — aborting', reshuffle;
    END IF;

    -- (2) PRECISION: nothing gains the token without a qualifying field.
    SELECT count(*) INTO unearned
      FROM content_components c
     WHERE 'image' = ANY (component_expresses(c.html_template, c.input_schema))
       AND NOT EXISTS (SELECT 1 FROM jsonb_each(COALESCE(c.input_schema->'fields','{}'::jsonb)) f
                        WHERE f.value->>'source' LIKE 'site_assets.%'
                          AND f.value->>'source' <> 'site_assets.logo'
                          AND f.value->>'type' IN ('url','image','image_url'));
    IF unearned <> 0 THEN
        RAISE EXCEPTION '644: % component(s) express image without a qualifying field — aborting', unearned;
    END IF;

    -- (3) The word actually arrived, and the control component did not move.
    SELECT count(*) INTO gained
      FROM content_components c
     WHERE 'image' = ANY (component_expresses(c.html_template, c.input_schema));
    IF gained = 0 THEN
        RAISE EXCEPTION '644: no component expresses image — the arm is inert, aborting';
    END IF;

    SELECT component_expresses(html_template, input_schema) = (SELECT tok FROM _644_before b WHERE b.id = c.id)
      INTO gtb_same FROM content_components c WHERE c.name = 'Generic Text Block';
    IF gtb_same IS DISTINCT FROM true THEN
        RAISE EXCEPTION '644: Generic Text Block changed — it must be byte-identical, aborting';
    END IF;

    SELECT component_expresses(html_template, input_schema) INTO itb
      FROM content_components WHERE name = 'Illustrated Text Block';
    IF itb IS NULL OR NOT ('image' = ANY (itb)) THEN
        RAISE EXCEPTION '644: Illustrated Text Block does not express image — the coupling failed, aborting';
    END IF;

    -- (4) The repoint landed on both fields.
    SELECT count(*) INTO repointed FROM content_components
     WHERE name = 'Illustrated Text Block'
       AND input_schema->'fields'->'image_url'->>'source' = 'site_assets.illustration'
       AND input_schema->'fields'->'image_alt'->>'source' = 'llm';
    IF repointed <> 1 THEN
        RAISE EXCEPTION '644: repoint applied to % Illustrated Text Block rows, expected 1 — aborting', repointed;
    END IF;

    RAISE NOTICE '644 OK: % components now express image; additive (0 lost, 0 reshuffled, 0 unearned); Generic Text Block unchanged; Illustrated Text Block repointed to site_assets.illustration + llm alt', gained;
END $$;

COMMIT;

-- POST-APPLY READING, with the demand control that makes a zero informative.
--
--   SELECT name, array_to_string(component_expresses(html_template, input_schema), ', ')
--     FROM content_components
--    WHERE is_active AND component_level IN ('section','element')
--      AND 'image' = ANY (component_expresses(html_template, input_schema))
--    ORDER BY name;
--
-- ⚠ A ZERO IN "new Illustrated Text Block instances" IS NOT A FAILURE OF THIS
-- MIGRATION. Nothing re-plans a page on its own — the token only reaches a
-- model on the next build/content-gap run for a site. And even then, a site with
-- no section-scope illustration row renders the block as plain prose by design.
-- Before reading adoption, confirm the DEMAND side: that a planner has actually
-- run since this applied.
--
--   SELECT count(*) FROM orchestration_states
--    WHERE owner_agent_type IN ('build-site-planner','site-planner','content-gap-planner')
--      AND created_at > '<the apply time>';
