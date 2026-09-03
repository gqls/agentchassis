-- 740_info_card_grid_carousel_defaults_on.sql
--
-- OWNER RULING 2026-09-03 ("switch the switches"): `info-card-grid`'s `carousel`
-- flag is to be DEFAULT-ON. [MEASURED 2026-09-03 12:41:06Z] it is set on
-- 1 of 40 live instances (deployed+active pages) across 21 sites.
--
-- ══ THE OPEN DESIGN QUESTION, ANSWERED ═══════════════════════════════════════
-- The handoff left it open: "where does the default live — schema default plus a
-- backfill of the instances, or resolution-time?" It is RESOLUTION-TIME, and the
-- mechanism already exists. `plan_sections_action.go:2886` (the renderer/static
-- branch), verbatim:
--
--     if source == "renderer" || source == "static" || ... {
--         if !carryStored() && fallback != nil {
--             resolvedData[fieldName] = fallback
--         }
--         continue
--     }
--
-- So a `fallback` on a `source: static` field IS the default, applied per render,
-- and a stored value BEATS it. No backfill is needed and none is done here.
--
-- ⚠ NO BACKFILL IS NEEDED ONLY BECAUSE THE KEY IS ABSENT, NOT FALSE.
-- [MEASURED 2026-09-03 12:41:06Z] of the 40 live instances: 1 carries
-- `carousel: true`, 0 carry `carousel: false`, 39 do not carry the key at all.
-- That matters because `carryStored()` -> `storedFieldValue()` rejects only
-- values `IsEmptyContentValue` calls empty, and a Go `bool` hits that function's
-- default arm — so a stored `false` is NOT empty, IS carried, and WOULD beat this
-- fallback. Had the 39 stored an explicit false, this migration would have
-- applied cleanly, verified green and changed nothing. Re-run the census before
-- assuming that is still true.
--
-- ══ THE MECHANISM IS PROVEN LIVE, NOT JUST READ ══════════════════════════════
-- [MEASURED 2026-09-03] `ai-readiness-quiz` is the existing worked example on the
-- SECTION path (not chrome): 12 `source: static` fields with fallbacks, and both
-- live instances (finetuning.uk, leopardessconsulting.co.uk) carry the fallback
-- text persisted into `page_components.content_data` byte-for-byte —
-- `quiz_back_label = 'Back'`, `quiz_badge_label = 'AI Readiness Assessment'`.
-- Static fields are never offered to the writer, so those values can only have
-- come from this path.
--
-- ══ THE PRE-FLIGHT GATE — the arrows would be inert without the JS ═══════════
-- The field's own guidance says the layout "requires the hero-card-carousel
-- js_snippet to be bundled for the site ... without it ... only the arrows are
-- inert". Turning arrows on for a site that cannot drive them is worse than the
-- grid, so this was measured AT THE ARTEFACT before writing the migration:
--   * `js_snippets` has ONE matching row, `hero-card-carousel`, is_active,
--     applies_to = ["hero-card-carousel", "info-card-grid"] — so the bundle
--     follows the component, not the flag.
--   * [MEASURED 2026-09-03] all 21 sites carrying info-card-grid serve
--     /assets/js/snippets.js at HTTP 200 with 15 `data-hcc` occurrences each.
--   * NEGATIVE CONTROL, because a constant 15 could mean the grep matches
--     something always present: 6 sites WITHOUT info-card-grid serve 0 — and one
--     of them (fundamentallyai.com) has a 10,928-byte bundle, so it is not
--     "small bundle = zero". The instrument discriminates.
--
-- ══ ⚠ THE ACCEPTANCE TEST IN THE HANDOFF IS WRONG. DO NOT USE IT ═════════════
-- HANDOFF_2026-09-03 §3.2 names `overflow-x` as the NEGATIVE control — "it reads
-- 2 on both ... a flip that moves `overflow-x` is doing something other than what
-- it says". That is FALSE, and following it would make a CORRECT flip read as a
-- defect.
-- The template's ONLY `overflow-x` (line 204) sits INSIDE the `{{if $.carousel}}`
-- style block, so a correct flip ADDS ONE per flipped instance. The equal count
-- of 2 was a coincidence of unrelated CSS on two different pages of two different
-- sites, [MEASURED 2026-09-03] at the served bytes:
--   * leopardess/services.html: 1 from `--trp-track-gap` (another component) +
--     1 from the info-card-grid carousel block itself (`--icg-track-gap`).
--   * designblog/index.html:   2 from `.category-strip`, emitted twice.
-- CORRECTED TEST — before/after on the SAME page:
--   POSITIVE (0 -> n): data-hcc-track, data-hcc-prev, data-hcc-next,
--                      data-hcc-slide, info-card-grid__grid--carousel,
--                      scroll-snap-type
--   EXPECTED TO MOVE:  overflow-x, +1 per flipped instance
--   NEGATIVE (must NOT move): the count of `info-card-grid__card` articles and
--                      the card titles — this is a LAYOUT change; the content
--                      must be byte-stable. A flip that changes the card count
--                      is doing something other than what it says.
--
-- ══ WHAT CHANGES, AND WHEN ═══════════════════════════════════════════════════
-- Config: live on apply, no image build. But nothing on a served page moves
-- until that page re-renders. The re-render path DOES apply this —
-- `rerender_page_sections_action.go:1450` calls the same `planSection`, and its
-- own comment records that `plan.ResolvedData` merges LAST and wins over stored
-- content_data. So a page-rerender is sufficient; a full rebuild is not needed.
-- A completed rerender is NOT evidence: read the served bytes with the test above.
--
-- IDEMPOTENT BY CONSTRUCTION. `jsonb_set` to a fixed value; running it twice
-- leaves the same single key. Stated because THIS LANE shipped the opposite two
-- days ago: 723 used `replace()` on text that re-embedded its own anchor, so a
-- second run would stack a second copy of the guidance. A JSON path write cannot
-- do that.
--
-- Reversible: 740_..._ROLLBACK.sql removes the fallback key.
-- Source: HANDOFF_2026-09-03 §3.2, owner ruling relayed via designblog.co.uk.

BEGIN;

-- DRIFT GUARD. Abort rather than clobber if the component is not in the state
-- this migration was written against.
DO $$
DECLARE n int; tpl text;
BEGIN
    SELECT count(*) INTO n
      FROM content_components
     WHERE is_active AND name = 'info-card-grid'
       AND input_schema->'fields'->'carousel'->>'source' = 'static'
       AND input_schema->'fields'->'carousel'->>'type'   = 'boolean';
    IF n <> 1 THEN
        RAISE EXCEPTION
            'ABORT: expected exactly 1 active info-card-grid declaring a static boolean '
            '`carousel` field, found %. The component or its schema has moved.', n;
    END IF;

    SELECT count(*) INTO n
      FROM content_components
     WHERE is_active AND name = 'info-card-grid'
       AND input_schema->'fields'->'carousel' ? 'fallback';
    IF n <> 0 THEN
        RAISE EXCEPTION
            'ABORT: `carousel` already declares a fallback — another session has edited '
            'it, or this migration has ALREADY applied. Re-read before re-running.';
    END IF;

    -- The template gate is the whole point: a fallback on a field no template
    -- reads is inert config, and inert config verifies green.
    SELECT html_template INTO tpl FROM content_components
     WHERE is_active AND name = 'info-card-grid';
    IF position('{{if $.carousel}}' in tpl) = 0 THEN
        RAISE EXCEPTION
            'ABORT: the info-card-grid template no longer gates on {{if $.carousel}} — '
            'setting the default would change nothing while verifying green.';
    END IF;
END $$;

-- Pre-image: the schema must gain exactly ONE key, inside the carousel
-- descriptor, and the `fields` set itself must NOT change. jsonb_set on the
-- wrong path satisfies every named assertion below while destroying the schema.
CREATE TEMP TABLE _pre_740 ON COMMIT DROP AS
SELECT id,
       (SELECT count(*) FROM jsonb_object_keys(input_schema->'fields')) AS n_fields,
       (SELECT count(*) FROM jsonb_object_keys(input_schema->'fields'->'carousel')) AS n_carousel_keys,
       input_schema->'fields'->'carousel'->>'llm_guidance' AS guidance
  FROM content_components
 WHERE is_active AND name = 'info-card-grid';

UPDATE content_components
   SET input_schema = jsonb_set(
           input_schema,
           '{fields,carousel,fallback}',
           'true'::jsonb,       -- JSON boolean, NOT the string "true": a Go
                                -- template {{if}} treats ANY non-empty string as
                                -- truthy, so "false" would render a carousel too.
           true),
       updated_at = now()
 WHERE is_active AND name = 'info-card-grid';

-- VERIFY. DO/RAISE, not SELECTs: ON_ERROR_STOP does not fire on a non-empty
-- result set, so a block of SELECTs cannot stop the COMMIT.
DO $$
DECLARE
    rows_seen  int;
    is_true    boolean;
    is_boolean boolean;
    bad_fields int;
    bad_keys   int;
    lost_guid  int;
BEGIN
    SELECT count(*) INTO rows_seen
      FROM content_components WHERE is_active AND name = 'info-card-grid';

    SELECT (input_schema->'fields'->'carousel'->'fallback') = 'true'::jsonb,
           jsonb_typeof(input_schema->'fields'->'carousel'->'fallback') = 'boolean'
      INTO is_true, is_boolean
      FROM content_components WHERE is_active AND name = 'info-card-grid';

    -- The `fields` SET must be untouched: this adds a key one level deeper.
    SELECT count(*) INTO bad_fields
      FROM content_components cc JOIN _pre_740 pre ON pre.id = cc.id
     WHERE (SELECT count(*) FROM jsonb_object_keys(cc.input_schema->'fields')) <> pre.n_fields;

    -- The carousel descriptor must gain exactly one key (source/type/required/
    -- llm_guidance all survive).
    SELECT count(*) INTO bad_keys
      FROM content_components cc JOIN _pre_740 pre ON pre.id = cc.id
     WHERE (SELECT count(*) FROM jsonb_object_keys(cc.input_schema->'fields'->'carousel'))
           <> pre.n_carousel_keys + 1;

    -- Named survivor, because a count can be satisfied by a swap.
    SELECT count(*) INTO lost_guid
      FROM content_components cc JOIN _pre_740 pre ON pre.id = cc.id
     WHERE cc.input_schema->'fields'->'carousel'->>'llm_guidance' IS DISTINCT FROM pre.guidance;

    IF rows_seen <> 1 THEN
        RAISE EXCEPTION 'ABORT: % active info-card-grid rows after the UPDATE, expected 1', rows_seen;
    END IF;
    -- NULL here means the ROW is present but the PATH is not — i.e. jsonb_set
    -- wrote somewhere else. Distinguished from the missing-row case above,
    -- because "no row" is the wrong thing to tell someone whose path is wrong.
    IF is_true IS NULL THEN
        RAISE EXCEPTION 'ABORT: fields.carousel.fallback does not resolve after the UPDATE — '
                        'jsonb_set wrote to the wrong path';
    END IF;
    IF NOT is_boolean THEN
        RAISE EXCEPTION 'ABORT: the fallback is not a JSON boolean — a string "false" '
                        'renders a carousel, because {{if}} is truthy on any non-empty string';
    END IF;
    IF NOT is_true THEN
        RAISE EXCEPTION 'ABORT: the fallback did not land as true';
    END IF;
    IF bad_fields > 0 THEN
        RAISE EXCEPTION 'ABORT: the `fields` set changed size — jsonb_set hit the wrong path';
    END IF;
    IF bad_keys > 0 THEN
        RAISE EXCEPTION 'ABORT: the carousel descriptor does not have exactly one MORE key than before';
    END IF;
    IF lost_guid > 0 THEN
        RAISE EXCEPTION 'ABORT: the carousel llm_guidance did not survive the write';
    END IF;

    RAISE NOTICE '740: info-card-grid carousel now defaults ON at resolution time. '
                 '39 of 40 live instances carry no stored value and will flip on their next '
                 'render; the 1 with a stored true is unchanged. Read the SERVED bytes, and '
                 'note overflow-x is EXPECTED to move +1 (the handoff said otherwise).';
END $$;

COMMIT;
