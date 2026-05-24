Post-run verification SQL
Run these after re-adopting gamedesign.uk → gamesdesign.co.uk and the plan lands (site_plan_* populated). They're split by what each proves.
1. Did the section indexes survive correctly? (Part A working end to end)
' ||
                                          'SELECT spp.name, spp.role, spp.url
FROM site_plan_pages spp
JOIN site_plans sp ON sp.id = spp.plan_id AND sp.is_current
WHERE sp.site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND (spp.name LIKE '%-index' OR spp.name IN ('games','tools','guides'))
ORDER BY spp.name;' ||
                                          '
Pass: games-index and tools-index present, role=section-index, url=/games/index.html and /tools/index.html. No flat games/tools rows. guides-index likewise at /guides/index.html. Fail: any flat games/tools, or -index names missing.
2. Same check on the realised pages (what actually got built/deployed):' ||
                                          '
SELECT name, page_type, build_status, url
FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND status = 'active'
  AND (name LIKE '%-index' OR name IN ('games','tools','guides'))
ORDER BY name;' ||
                                          '
Pass: games-index → page_type=section-index, url=/games/index.html (compare last run's broken content / /games-index.html). This is the before/after that confirms the fix reached the pages table, not just the plan.
3. The flavour-collapse evidence — did guides-index keep blog-index or collapse to section-index?
                                 
SELECT name, page_type, role_note
FROM (
  SELECT p.name, p.page_type,
         (SELECT spp.role FROM site_plan_pages spp
          JOIN site_plans sp ON sp.id = spp.plan_id AND sp.is_current
          WHERE sp.site_id = p.site_id AND spp.name = p.name) AS role_note
  FROM pages p
  WHERE p.site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
    AND p.status = 'active'
    AND p.name IN ('guides-index','games-index','tools-index')
) x
ORDER BY name;

This shows, per hub, the stored page_type (pages) alongside the plan role. If all three read section-index/section_index, the flavour collapsed (the Tension #2 residual is live); if guides-index retained blog-index, it didn't. Either way it's the input to the rendering question below.
Where to look for whether the guides hub got its list component
The flavour question only matters if it changes what renders. Two places tell you, in order of directness:
A. The deployed/rendered HTML for the guides hub — does it contain a guide/blog list section at all?

SELECT name, build_status,
       (content_data ? 'sections')                       AS has_sections_key,
       jsonb_array_length(COALESCE(content_data->'sections','[]'::jsonb)) AS n_sections,
       LEFT(content_data::text, 600)                      AS content_head
FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND name = 'guides-index';

Look in content_head / the sections array for a guide-list (or blog-list) component. If it's there, the hub renders its list regardless of the collapsed page_type → the flavour loss is cosmetic, no fix needed. If sections is empty or carries no list component, the hub came out bare → the flavour collapse (or the empty-sections the LLM emitted) is the cause, and it's a real bug.

