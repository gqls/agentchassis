-- 577 — seed the `adopted-fragment` component (RFC_046 phase 2, bugs_open/357)
--
-- HELD ON PURPOSE. `_HOLD` keeps this out of the migration runner's sweep so it
-- is applied BY HAND, at the same moment the `adopt_unidentified_fragments` step
-- config is switched on. Nothing needs it before then: with the flag OFF no code
-- path looks this component up, and with the flag ON but the component absent,
-- `adoptFragmentSection` returns false and the section is left honestly
-- unidentified — a safe degradation, not a failure.
--
-- WHY HOLD IT RATHER THAN JUST APPLY IT. `content_components` is a SHARED library
-- and `LoadComponentLibraryAction` enumerates it (`load_component_library_actions.go:145`),
-- filtering on `component_level`, `category` and `is_active`. A row that exists is
-- a row something can offer to a planner. The column values below make that
-- structurally unlikely — `component_level='fragment'` is excluded by every
-- caller asking for `'section'`, and `category='internal'` by every caller passing
-- a category list — but a caller passing `component_level='all'` with no category
-- filter WOULD see it. That residual is stated rather than hidden, and holding the
-- seed until the flag is armed means the library gains nothing until something
-- actually needs it.
--
-- WHAT IT IS FOR. `bugs_open/357`: a tool page is one fragment with no <section>,
-- so it is stored as a single section whose identity is invented from POSITION in
-- the page's plan — 22 live rows as of 2026-08-23 declare themselves the shared
-- `hero` while storing a whole interactive tool, and 12 of those were born that
-- day. The estate has no way to say "these bytes ARE the content": measured
-- 2026-08-23, `content_components` holds ZERO components whose template is a bare
-- `{{.body}}` and none named adopted-fragment/passthrough/verbatim. This is that
-- missing representation.
--
-- WHY `{{.body}}` IS SAFE HERE: `RenderTemplate` executes with **text/template**
-- (`component_library.go` imports "text/template", not "html/template"), so the
-- stored markup is reproduced verbatim rather than escaped. `adoptFragmentSection`
-- does not take this on trust — it renders this template and refuses to bind
-- unless the output is byte-identical to the fragment, so editing this template
-- into something that wraps or escapes STOPS adoption rather than corrupting rows.
--
-- `input_schema` declares `body` with NO `source`. That is deliberate and it is
-- tolerated by design: an empty source is in `bareSourceValues`
-- (`component_source_guard.go`), so neither the birth gate nor the daily
-- `component-source-vocabulary-check` (CLC-025, whose baseline is frozen at
-- 2026-08-22 and would go RED on a new offending field) has anything to say
-- about it.
--
-- Idempotent: re-running is a no-op, so applying it twice cannot mint a second
-- component and split the fleet's adopted rows across two identities.

BEGIN;

INSERT INTO content_components (
    name, display_name, description,
    html_template, function,
    input_schema,
    component_level, category, render_mode,
    suitable_site_types, suitable_page_types,
    is_active
)
SELECT
    'Adopted Fragment',
    'Adopted Fragment',
    'Identity-function component for markup this platform stores but did not compose: '
      || 'its template is exactly {{.body}}, so a row bound to it regenerates byte-identically '
      || 'from its own content_data. Exists so that an unidentified fragment can be typed '
      || 'TRUTHFULLY instead of inheriting the identity of whatever its page plan listed first '
      || '(RFC_046, bugs_open/357). Not for authoring: nothing should ever plan a page section '
      || 'onto this component.',
    '{{.body}}',
    'adopted-fragment',
    '{"fields": {"body": {"type": "html", "required": true, "description": "The stored markup, reproduced verbatim."}}}'::jsonb,
    'fragment',
    'internal',
    'template',
    '[]'::jsonb,
    '[]'::jsonb,
    true
WHERE NOT EXISTS (
    SELECT 1 FROM content_components WHERE function = 'adopted-fragment'
);

-- Refuse to commit unless exactly one exists afterwards. A plain SELECT here
-- could not stop the COMMIT (ON_ERROR_STOP ignores a non-empty result set), so
-- the check RAISEs — the rule from RFC_006's verify-block trap.
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM content_components
     WHERE function = 'adopted-fragment' AND is_active = true;
    IF n <> 1 THEN
        RAISE EXCEPTION 'expected exactly 1 active adopted-fragment component, found %', n;
    END IF;
END $$;

COMMIT;
