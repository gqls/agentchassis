-- APPLIED 2026-09-02 direct to clients_db (site content, not a platform migration).
-- Kept as the worked record for the RUNBOOK. Do NOT re-apply: the spec ids it
-- retires are already superseded; a re-run's in-transaction guard would abort it.
-- gamesdesign.co.uk brand rename: 'GameDesign.uk' -> 'GamesDesign.co.uk'
-- Owner ruling 2026-09-02 (relayed via Portfolio positioning + gamedesign.uk sessions);
-- class bug: bugs_open/439 (adoption carried source brand verbatim). Instance remediation.
-- Case-SENSITIVE replace throughout: 'gamedesign.uk' (lowercase) is a historical fact
-- (identity.adopted_from) and a real cross-link (guide-p2p-architecture) and must survive.
\set ON_ERROR_STOP on
BEGIN;

-- ==== backups (site-scoped, affected rows only) ====
CREATE TABLE bak_gdcouk_rename_20260902_pages AS
  SELECT * FROM pages
  WHERE site_id='e33263f4-74f8-494f-b191-546845dbbddf'
    AND (title LIKE '%GameDesign.uk%' OR meta_description LIKE '%GameDesign.uk%');

CREATE TABLE bak_gdcouk_rename_20260902_plan_pages AS
  SELECT spp.* FROM site_plan_pages spp
  JOIN site_plans sp ON sp.id=spp.plan_id
  WHERE sp.site_id='e33263f4-74f8-494f-b191-546845dbbddf'
    AND spp.title LIKE '%GameDesign.uk%';

CREATE TABLE bak_gdcouk_rename_20260902_page_components AS
  SELECT pc.* FROM page_components pc
  JOIN pages p ON p.id=pc.page_id
  WHERE p.site_id='e33263f4-74f8-494f-b191-546845dbbddf'
    AND (pc.content_data::text LIKE '%GameDesign.uk%' OR pc.rendered_html LIKE '%GameDesign.uk%');

-- ==== site_specs: supersede, never update in place ====
-- retire (separate statement from the insert; chained CTE hits the partial unique index)
UPDATE site_specs SET is_current=false, superseded_at=now()
WHERE id IN ('7814e6b0-89ad-48de-ae97-9d54bf6660b2',  -- briefing
             'bf4ef822-3d4d-4dae-9571-de5dc0ff02b1',  -- design_intent
             'b4b25fa1-e5c3-491f-8596-e4afb1f9e34b',  -- identity
             '528c662c-965d-4b33-a871-3774f2aa0e8a'); -- tools

-- insert replacements: identical data except the brand string
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, notes, is_current, created_by)
SELECT site_id, aspect,
       replace(data::text, 'GameDesign.uk', 'GamesDesign.co.uk')::jsonb,
       'operator', NULL,
       'Brand rename GameDesign.uk -> GamesDesign.co.uk (owner ruling 2026-09-02; bugs_open/439 instance). Supersedes ' || id || '; only change is the brand string.',
       true, 'claude-session-gamesdesign-couk-20260902'
FROM site_specs
WHERE id IN ('7814e6b0-89ad-48de-ae97-9d54bf6660b2',
             'bf4ef822-3d4d-4dae-9571-de5dc0ff02b1',
             'b4b25fa1-e5c3-491f-8596-e4afb1f9e34b',
             '528c662c-965d-4b33-a871-3774f2aa0e8a');

-- ==== plan (the source the build reads) ====
UPDATE site_plan_pages spp SET title = replace(spp.title, 'GameDesign.uk', 'GamesDesign.co.uk')
FROM site_plans sp
WHERE sp.id=spp.plan_id
  AND sp.site_id='e33263f4-74f8-494f-b191-546845dbbddf'
  AND spp.title LIKE '%GameDesign.uk%';

-- ==== pages (materialised titles + one meta) ====
UPDATE pages SET
  title = replace(title, 'GameDesign.uk', 'GamesDesign.co.uk'),
  meta_description = replace(meta_description, 'GameDesign.uk', 'GamesDesign.co.uk')
WHERE site_id='e33263f4-74f8-494f-b191-546845dbbddf'
  AND (title LIKE '%GameDesign.uk%' OR meta_description LIKE '%GameDesign.uk%');

-- ==== components: content_data (feeds every rerender) + rendered_html (the stored artefact) ====
UPDATE page_components pc SET
  content_data = replace(pc.content_data::text, 'GameDesign.uk', 'GamesDesign.co.uk')::jsonb,
  rendered_html = replace(pc.rendered_html, 'GameDesign.uk', 'GamesDesign.co.uk')
FROM pages p
WHERE p.id=pc.page_id
  AND p.site_id='e33263f4-74f8-494f-b191-546845dbbddf'
  AND (pc.content_data::text LIKE '%GameDesign.uk%' OR pc.rendered_html LIKE '%GameDesign.uk%');

-- ==== verify inside the transaction: residuals must be zero, else abort ====
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id='e33263f4-74f8-494f-b191-546845dbbddf' AND is_current
     AND data::text LIKE '%GameDesign.uk%';
  IF n > 0 THEN RAISE EXCEPTION 'site_specs residual: %', n; END IF;

  SELECT count(*) INTO n FROM pages
   WHERE site_id='e33263f4-74f8-494f-b191-546845dbbddf'
     AND (title LIKE '%GameDesign.uk%' OR meta_description LIKE '%GameDesign.uk%');
  IF n > 0 THEN RAISE EXCEPTION 'pages residual: %', n; END IF;

  SELECT count(*) INTO n FROM page_components pc JOIN pages p ON p.id=pc.page_id
   WHERE p.site_id='e33263f4-74f8-494f-b191-546845dbbddf'
     AND (pc.content_data::text LIKE '%GameDesign.uk%' OR pc.rendered_html LIKE '%GameDesign.uk%');
  IF n > 0 THEN RAISE EXCEPTION 'page_components residual: %', n; END IF;

  -- demand control: the NEW name must now be present (a zero here means the replace never ran)
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id='e33263f4-74f8-494f-b191-546845dbbddf' AND is_current
     AND data::text LIKE '%GamesDesign.co.uk%';
  IF n <> 4 THEN RAISE EXCEPTION 'expected 4 current specs carrying GamesDesign.co.uk, got %', n; END IF;

  -- the lowercase historical fact must survive
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id='e33263f4-74f8-494f-b191-546845dbbddf' AND is_current AND aspect='identity'
     AND data->>'adopted_from' = 'gamedesign.uk';
  IF n <> 1 THEN RAISE EXCEPTION 'identity.adopted_from lost its lowercase historical value'; END IF;
END $$;

COMMIT;
