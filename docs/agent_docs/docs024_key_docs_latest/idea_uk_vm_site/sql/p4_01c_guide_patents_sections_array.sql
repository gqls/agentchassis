-- p4_01c_guide_patents_sections_array.sql — backfill pages.sections for the patents guide.
--
-- WHY. The page deployed correctly (live 200, build_status='deployed', deployed_at stamped), but
-- `pages.sections` was left `[]`. The page-rerender path does not write it: save_page_sections
-- reported {"sections_found":3,"sections_saved":3} — it saves page_components.rendered_html, not
-- the page-level slot list. Every other guide in the fleet has it populated by the ORIGINAL build
-- path (page-build-handler), which this page bypassed by design.
--
-- WHY IT MATTERS (not cosmetic). `ListedPageEligibilitySQL` (queryresolve.go:162) is
--   deployed_at IS NOT NULL AND jsonb_typeof(sections)='array' AND jsonb_array_length(sections) > 0
-- and it is the SHARED literal used by both the article listing and the imagery sweep
-- (discovery_checks.ContentImageMissingCheck) precisely so the two cannot disagree. A deployed page
-- with an empty sections array is invisible to everything on that contract. The guides hub itself is
-- safe (pages_where_type uses the weaker FetchablePageEligibilitySQL — deployed_at OR
-- build_status='deployed' — which this page already passes), but leaving the field empty makes this
-- page a permanent odd-one-out and a candidate for some future "deployed but has no sections"
-- sweep to flag or rebuild — which would overwrite authored legal content.
--
-- The value is simply the ordered slot-name list, verified against the fleet:
--   /guides/rng-design/index.html -> ["hero","generic-text-block"]
--   idea.uk /guides/index.html    -> ["hero","content-listing"]

\set ON_ERROR_STOP on

BEGIN;

UPDATE pages p
SET sections = (
      SELECT COALESCE(jsonb_agg(pc.slot_name ORDER BY pc.position), '[]'::jsonb)
      FROM page_components pc
      WHERE pc.page_id = p.id AND COALESCE(pc.slot_name,'') <> ''
    ),
    updated_at = now()
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/patents/index.html';

DO $guard$
DECLARE n int;
BEGIN
  SELECT jsonb_array_length(sections) INTO n FROM pages
   WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
     AND url = '/guides/patents/index.html';
  IF COALESCE(n,0) <> 3 THEN
    RAISE EXCEPTION 'ABORT: expected 3 slot names in pages.sections, got %.', COALESCE(n,0);
  END IF;
END
$guard$;

COMMIT;

SELECT url, build_status, deployed_at IS NOT NULL AS stamped, sections
FROM pages
WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND url = '/guides/patents/index.html';
