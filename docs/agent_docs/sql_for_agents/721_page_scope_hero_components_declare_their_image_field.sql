-- 721_page_scope_hero_components_declare_their_image_field.sql
--
-- SIX page-scope hero components render an image key their schema never
-- declares, so the per-page hero the pipeline generated for them is ORPHANED
-- and the page wears the SITE-WIDE homepage image instead.
--
-- ══ THE MECHANISM, and it is already written down in the code ════════════════
-- `plan_sections_action.go:2897` ("Authoritative hero aliasing"), verbatim:
--
--   "when this section declares an image-typed field, also write the resolved
--    page hero under the legacy alias keys (hero_url, background_image) unless
--    the schema declares them itself ... this is what lets the per-page hero
--    defeat the site-wide hero_url that BuildRenderContext still injects for
--    legacy templates: without it, {{or .hero_url .background_image}} picks the
--    site-wide value and every page shows the same image."
--
-- That last clause IS the defect. BuildRenderContext injects a site-wide
-- hero_url for legacy templates; the aliasing block exists to override it per
-- page; and the block is GATED on the section declaring an IMAGE-TYPED field.
-- These six declare none, so the override never runs and the site-wide value
-- wins every time.
--
-- ⚠ THE FIELD MUST BE type:"image" (or "image_url") — THIS IS THE WHOLE FIX.
-- The gate is `sectionHasImageField` (`plan_sections_action.go:2936`):
--     if t, _ := def["type"].(string); t == "image" || t == "image_url"
-- Declared as "text" or "string" the gate stays FALSE, the aliasing still does
-- not run, hero_url still wins, and this migration applies cleanly, re-renders
-- cleanly and changes NOTHING — while every "does the page have a hero image"
-- check goes on passing, because a WRONG image is indistinguishable from a
-- right one to a presence check. That is the same shape as the defect itself.
--
-- ══ SCOPE, enumerated BY PREDICATE rather than by name ═══════════════════════
-- [MEASURED 2026-09-02] components whose html_template reads `.background_image`
-- or `.hero_url` while their input_schema declares no `background_image` source:
--
--   hero-tool        76 instances     services-hero       6
--   about-hero       43               case-studies-hero   5
--   contact-hero     25               use-cases-hero      2      = 157 live
--
-- `hero` itself (638 instances) already declares `site_assets.hero` and is the
-- exemplar this copies verbatim — it is deliberately NOT touched.
--
-- DELIBERATELY EXCLUDED: "webdesign.co.uk Two-Column Hero" (`webdesign-couk-hero`)
-- matches the template half of the predicate but has ZERO live instances, so it
-- cannot be a live defect; adding a field to a component nothing renders would
-- be inert change on a shared library. Re-run the predicate before assuming that
-- is still true.
--
-- ══ WHY NOT THE CHEAPER ROUTE THAT WAS PROPOSED ══════════════════════════════
-- A counter-instance was offered as possibly generalising: leopardessconsulting
-- .co.uk/tool-automation-savings-estimator renders its OWN hero asset
-- (`/assets/images/hero-tool-automation-savings-estimator.jpg`) through the
-- canonical `hero-tool` with no declared field, because `background_image` sits
-- in that instance's stored content_data. If that route were general it would
-- beat editing six components.
-- [MEASURED 2026-09-02] IT IS NOT GENERAL: across the 157 instances, exactly
-- ONE per component carries `background_image` in content_data (1 of 76, 1 of
-- 43, 1 of 25, 1 of 6, 1 of 5, 0 of 2) — five instances in all. That is an
-- anomaly, not a mechanism, and nothing can be built on it.
--
-- ══ WHAT CHANGES, AND WHAT DOES NOT ══════════════════════════════════════════
-- Additive: a field the templates ALREADY read. A page with no page-scope hero
-- asset falls through to the declared fallback exactly as it does today. A page
-- WITH one starts showing its own instead of the homepage's.
-- The change is CONFIG — live on apply — but no stored page changes until its
-- next render. Pages must be re-rendered to pick it up, and a completed
-- page_rerender is NOT evidence: read the served `url(...)` with a
-- filename-anchored control (bugs_open/425 learned this the hard way).
--
-- Reversible: 721_..._ROLLBACK.sql restores both fields verbatim.
-- Source: bugs_open/446 §3.6 (gamedesign.uk lane, filename-anchored census),
-- bugs_open/114 second CONTRIB (inline guide imager, the seven-component
-- census), routed to this lane 2026-09-02.

BEGIN;

-- DRIFT GUARD. Abort rather than clobber if the library is not in the state
-- this migration was written against.
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n
      FROM content_components
     WHERE is_active
       AND function IN ('hero-tool','hero-about','hero-contact',
                        'hero-services','hero-case-studies','hero-use-cases')
       AND input_schema->'fields' ? 'background_image';
    IF n <> 0 THEN
        RAISE EXCEPTION
            'ABORT: % of the six target components already declare background_image — '
            'another session has edited them, or this migration has ALREADY applied. '
            'Re-read before re-running.', n;
    END IF;

    SELECT count(*) INTO n
      FROM content_components
     WHERE is_active AND name = 'hero'
       AND input_schema->'fields'->'background_image'->>'type' = 'image';
    IF n <> 1 THEN
        RAISE EXCEPTION
            'ABORT: the `hero` exemplar this copies no longer declares an image-typed '
            'background_image (matched % rows). The shape being copied has moved.', n;
    END IF;
END $$;

-- Pre-image for the positive control below. Counting KEYS rather than asserting
-- a named field, because the six do not share a field vocabulary: hero-tool
-- declares hero_headline/hero_subheadline where the other five declare
-- headline/subheadline. The first cut of this control asserted `headline` and
-- failed the dry run on hero-tool — the control was over-strict, not the
-- migration wrong, and a shape-agnostic count is what it should always have been.
CREATE TEMP TABLE _pre_721 ON COMMIT DROP AS
SELECT id, (SELECT count(*) FROM jsonb_object_keys(input_schema->'fields')) AS n_fields
  FROM content_components
 WHERE is_active
   AND function IN ('hero-tool','hero-about','hero-contact',
                    'hero-services','hero-case-studies','hero-use-cases');

UPDATE content_components
   SET input_schema = jsonb_set(
           input_schema,
           '{fields,background_image}',
           jsonb_build_object(
               'type',       'image',            -- LOAD-BEARING: sectionHasImageField
               'source',     'site_assets.hero',
               'fallback',   '/assets/images/hero.jpg',
               'required',   false,
               'on_missing', 'use_fallback'
           ),
           true),
       updated_at = now()
 WHERE is_active
   AND function IN ('hero-tool','hero-about','hero-contact',
                    'hero-services','hero-case-studies','hero-use-cases');

-- VERIFY. DO/RAISE, not SELECTs: ON_ERROR_STOP does not fire on a non-empty
-- result set, so a block of SELECTs cannot stop the COMMIT.
DO $$
DECLARE
    targets  int;
    declared int;
    typed    int;
    lost     int;
BEGIN
    SELECT count(*) INTO targets FROM content_components
     WHERE is_active AND function IN ('hero-tool','hero-about','hero-contact',
                                      'hero-services','hero-case-studies','hero-use-cases');

    SELECT count(*) INTO declared FROM content_components
     WHERE is_active AND function IN ('hero-tool','hero-about','hero-contact',
                                      'hero-services','hero-case-studies','hero-use-cases')
       AND input_schema->'fields'->'background_image'->>'source' = 'site_assets.hero';

    -- The load-bearing assertion. A field of the wrong TYPE satisfies "declared"
    -- and leaves the gate false, so the migration would report success and
    -- change nothing.
    SELECT count(*) INTO typed FROM content_components
     WHERE is_active AND function IN ('hero-tool','hero-about','hero-contact',
                                      'hero-services','hero-case-studies','hero-use-cases')
       AND input_schema->'fields'->'background_image'->>'type' = 'image';

    -- Positive control: every pre-existing field must SURVIVE and exactly ONE
    -- must be added. jsonb_set on the wrong path satisfies every check above
    -- while destroying the schema, so this compares against the pre-image
    -- captured before the UPDATE rather than against a named field.
    SELECT count(*) INTO lost
      FROM content_components cc JOIN _pre_721 pre ON pre.id = cc.id
     WHERE (SELECT count(*) FROM jsonb_object_keys(cc.input_schema->'fields'))
           <> pre.n_fields + 1;

    IF targets = 0 THEN
        RAISE EXCEPTION 'ABORT: 0 target components matched — the function names have moved';
    END IF;
    IF declared <> targets THEN
        RAISE EXCEPTION 'ABORT: % of % targets carry the source', declared, targets;
    END IF;
    IF typed <> targets THEN
        RAISE EXCEPTION 'ABORT: only % of % carry type=image — the gate '
                        '(sectionHasImageField) stays FALSE for the rest and the fix is inert',
                        typed, targets;
    END IF;
    IF lost > 0 THEN
        RAISE EXCEPTION 'ABORT: % component(s) do not have exactly one MORE field than before — '
                        'jsonb_set hit the wrong path and altered the schema', lost;
    END IF;

    RAISE NOTICE '721: % components now declare an IMAGE-TYPED background_image; '
                 'every pre-existing field intact, exactly one added', targets;
END $$;

COMMIT;
