\set ON_ERROR_STOP on
-- Acceptance item 2: name the component in the planner prompt, or the planner
-- never selects it. This is the item that has bitten this workstream five times
-- — all five existing components are registered, visible in the library the
-- planner is shown, and still never chosen, because the prompt's "Common
-- patterns" prose only names the old set.
--
-- WHY THE PROMPT AND NOT THE REGISTRATION: `site-architect` is the only agent
-- that runs `load_component_library` (its `load_components` step), and its
-- prompt injects the whole active section library as
-- {{.component_library.for_prompt}} — 105 components as of today. Registration
-- therefore puts the component in front of the model already; what it lacks is
-- any reason to pick this one out of a hundred. That is what a named pattern
-- with a selection condition supplies.
--
-- SCOPE: `build-site-planner` is deliberately NOT edited — it does not load the
-- component library and names no components at all, so section choice is not
-- its decision to make.
--
-- The edit is additive, anchored on a string that occurs exactly once, and
-- applied through jsonb_set on the prompt_template path so JSON escaping is
-- handled by to_jsonb rather than by hand.

BEGIN;

DROP TABLE IF EXISTS bak_agent_definitions_sitearch_20260726;
CREATE TABLE bak_agent_definitions_sitearch_20260726 AS
SELECT * FROM agent_definitions
 WHERE type = 'site-architect' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE n int; hits int;
BEGIN
  SELECT count(*) INTO n FROM bak_agent_definitions_sitearch_20260726;
  IF n <> 1 THEN
    RAISE EXCEPTION 'expected exactly 1 live site-architect row, found %', n;
  END IF;
  SELECT (length(p) - length(replace(p, E'\n\n## Output JSON Format', ''))) /
         length(E'\n\n## Output JSON Format') INTO hits
    FROM (SELECT default_config#>>'{workflow,steps,design,config,prompt_template}' AS p
            FROM bak_agent_definitions_sitearch_20260726) q;
  IF hits <> 1 THEN
    RAISE EXCEPTION 'anchor must occur exactly once, found % — the prompt has changed, re-read it before editing', hits;
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,design,config,prompt_template}',
         to_jsonb(replace(
           default_config#>>'{workflow,steps,design,config,prompt_template}',
           E'\n\n## Output JSON Format',
           E'\n6. Evidence charts: when the business has real, audited figures to show, include one "evidence-chart" section on the page that argues from numbers — usually the index or a capabilities/approach page, and at most one per page. It draws code-rendered charts from the site''s own verified figures; it never invents a number, and it renders nothing at all if the site has no audited figures, so choosing it is always safe.\n7. Prefer a specific component over a generic one where the content suits it: "stat-band" for a few headline figures as text, "evidence-chart" for figures worth comparing, "hero-card-carousel" for a set of offerings with images, "people-feature-block" for a statement about the people, "swipeable-insight-carousel" for a series of short insights, "image-hover-card-grid" for a grid of illustrated cards.\n\n## Output JSON Format')))
 WHERE type = 'site-architect' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

COMMIT;

\echo '--- verify: the six components are now named, and the JSON still parses ---'
SELECT position('evidence-chart' in p) > 0 AS names_evidence_chart,
       position('stat-band' in p) > 0      AS names_stat_band,
       position('hero-card-carousel' in p) > 0 AS names_carousel,
       length(p) AS prompt_bytes
  FROM (SELECT default_config#>>'{workflow,steps,design,config,prompt_template}' AS p
          FROM agent_definitions WHERE type='site-architect' AND is_active
            AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) q;
