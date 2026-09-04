-- 769 — the two copywriter directory components (snippet + listing), cloned from the live
-- health-insurer pair so the rendering, styling and client-refresh cannot drift from a pair known
-- to work. Kind `copywriter`, DIR-001. Owner instruction 2026-09-04.
--
-- ⚠ _HOLD — APPLY ONLY AFTER v1.0.1361 IS LIVE AND POD-VERIFIED. These components declare
-- `query.directory:copywriter` / `query.directory_full:copywriter`, and those parameterised arms
-- ship in that image (commit 48bff098d, cut 06c0b18f2, council 32c75bc5 APPROVED). On the OLD binary
-- `queryresolve.Resolve` refuses an unknown base and the server render fails — the finance seed's own
-- header records the same precondition, and its lesson is that the failure is at RENDER time, not at
-- apply time. Gate before applying:
--   SELECT pod_name, git_commit FROM service_binary_capabilities
--    WHERE kind='build' AND pod_name LIKE 'agent-chassis-%' ORDER BY started_at DESC;
--   git merge-base --is-ancestor 48bff098d <that commit>     # must be true on EVERY chassis pod
--
-- WHY THE GENERIC SOURCE AND NOT `query.copywriter_directory`: the parameterised arm is the whole
-- point of 48bff098d — a seventh kind should need no Go. Declaring the generic form here makes this
-- pair the first proof that it works end to end. The twelve literal per-kind arms stay live for the
-- components that already declare them; nothing is migrated by this file.
--
-- ORGANISATIONS, NOT PEOPLE: the listing renders `entity_name` and the organisation's own claims.
-- The register only holds organisations (765/767's prompt rule), so this is inherited, not enforced
-- here — if that rule is ever relaxed, this page is where it becomes public.

BEGIN;

DO $g$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM content_components
   WHERE function IN ('health-insurer-directory','health-insurer-directory-listing') AND is_active;
  IF n <> 2 THEN RAISE EXCEPTION '769 REFUSED: expected the 2 live health-insurer components to clone, found %', n; END IF;
  SELECT count(*) INTO n FROM content_components
   WHERE function IN ('copywriter-directory','copywriter-directory-listing');
  IF n <> 0 THEN RAISE EXCEPTION '769 REFUSED: copywriter components already exist (% rows)', n; END IF;
END $g$;

-- Snippet (section) and listing (full page), cloned then re-pointed. Every literal that names the
-- kind is replaced; the structure, styling and JS are inherited verbatim.
INSERT INTO content_components (
    name, display_name, description, function, category, component_level,
    is_active, created_from, input_schema, html_template, js_content
)
SELECT
    replace(replace(src.name, 'Health Insurer', 'Copywriter'), 'UK Health Insurer', 'UK Copywriter'),
    replace(src.display_name, 'Health Insurer', 'Copywriter'),
    'Cited directory of UK copywriting organisations (agencies, studios, consultancies and marketplaces) with their specialisms, sectors served and location. Server-rendered from the global register via the parameterised query.directory arm; organisations only, never individuals.',
    replace(src.function, 'health-insurer', 'copywriter'),
    src.category, src.component_level, true,
    'cloned from '||src.function||' by migration 769 (DIR-001 kind copywriter)',
    -- re-point the item source at the generic arm for this kind
    jsonb_set(
      src.input_schema,
      '{fields,entries,source}',
      to_jsonb(CASE WHEN src.function = 'health-insurer-directory'
                    THEN 'query.directory:copywriter'
                    ELSE 'query.directory_full:copywriter' END)
    ),
    replace(replace(src.html_template, 'health-insurer', 'copywriter'), 'Health Insurer', 'Copywriter'),
    replace(replace(src.js_content, 'health-insurer', 'copywriter'), 'Health Insurer', 'Copywriter')
FROM content_components src
WHERE src.function IN ('health-insurer-directory','health-insurer-directory-listing') AND src.is_active;

DO $v$
DECLARE r record; n int;
BEGIN
  SELECT count(*) INTO n FROM content_components WHERE function IN ('copywriter-directory','copywriter-directory-listing') AND is_active;
  IF n <> 2 THEN RAISE EXCEPTION '769 VERIFY: expected 2 new components, found %', n; END IF;
  FOR r IN SELECT function, input_schema #>> '{fields,entries,source}' AS src, html_template, js_content
             FROM content_components WHERE function IN ('copywriter-directory','copywriter-directory-listing') AND is_active LOOP
    IF r.function = 'copywriter-directory' AND r.src <> 'query.directory:copywriter' THEN
      RAISE EXCEPTION '769 VERIFY: snippet source is %', r.src; END IF;
    IF r.function = 'copywriter-directory-listing' AND r.src <> 'query.directory_full:copywriter' THEN
      RAISE EXCEPTION '769 VERIFY: listing source is %', r.src; END IF;
    IF position('health-insurer' in r.html_template) > 0 OR position('health-insurer' in COALESCE(r.js_content,'')) > 0 THEN
      RAISE EXCEPTION '769 VERIFY: % still carries the health-insurer literal', r.function; END IF;
    IF position('Health Insurer' in r.html_template) > 0 THEN
      RAISE EXCEPTION '769 VERIFY: % still carries the Health Insurer label', r.function; END IF;
  END LOOP;
  -- the source pair we cloned from must be untouched
  SELECT count(*) INTO n FROM content_components
   WHERE function IN ('health-insurer-directory','health-insurer-directory-listing') AND is_active
     AND input_schema #>> '{fields,entries,source}' LIKE 'query.health_insurer_directory%';
  IF n <> 2 THEN RAISE EXCEPTION '769 VERIFY: the health-insurer originals were modified'; END IF;
END $v$;

COMMIT;
