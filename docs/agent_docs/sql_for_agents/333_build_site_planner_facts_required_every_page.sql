-- 333 — build-site-planner: fact assignment becomes REQUIRED for every page
-- when a roster exists (bugs_open/151 candidate 1, Slice B round; RFC_016 §3a
-- option (a), DECIDED by the owner 2026-08-08).
--
-- Why: the 2026-08-07 Slice A observation (RFC_016 §3a) — the planner used the
-- object form only on the 5 pages it composed fresh and left every carried-over
-- page unscoped, including the pages holding the 9 fact-overlap pairs that
-- motivated 151. Three anchored edits, all derived byte-exact from the live row
-- by scratchpad/gen_seed_333.py (never retyped):
--   1. RULES rule 17: object form + explicit "facts" key mandatory on EVERY
--      section of EVERY page when a roster exists; ownership decided for every
--      roster fact; [] stays the deliberate-factless marker.
--   2. Roster block: the assignment instruction now says every-section-every-
--      page explicitly.
--   3. The JSON example's sections array becomes all-object form — the old
--      mixed example invited imitation of exactly the partial engagement we
--      measured.
--
-- Consequence for observation: after this applies, NULL assigned_fact_ids on a
-- fact-rich site's fresh plan means the planner DISOBEYED (yesterday it meant
-- "allowed"), so the follow-up replan read becomes a compliance check.
-- Safe against roll order for the same reason 329 was: nothing consumes
-- assignments until seeds 328/330 apply (held), and old binaries pass object
-- entries through name-only.

SELECT snapshot_agent('build-site-planner', '333_build_site_planner_facts_required_every_page.sql: pre-update');

BEGIN;

CREATE TABLE IF NOT EXISTS agent_definitions_bak_333 AS
SELECT id, type, default_config, now() AS backed_up_at
FROM agent_definitions
WHERE type = 'build-site-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
    t text;
    c1 int; c2 int; c3 int;
BEGIN
    SELECT default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template'
    INTO t
    FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF t IS NULL THEN
        RAISE EXCEPTION '333: build-site-planner plan_site prompt_template not found';
    END IF;
    IF position('EVERY section entry on EVERY page' in t) > 0 THEN
        RAISE EXCEPTION '333: already applied — mandatory-object rule present';
    END IF;

    c1 := (length(t) - length(replace(t, $a1$17. Section entries may be plain strings ("hero") or objects ({"name": "features", "facts": ["F1-…"]}). Use the object form when the Verified Facts section above lists facts: put every fact you want stated into exactly ONE section's "facts" list, using the IDs exactly as listed there (an ID not in that list is ignored). Spread facts across sections — never give two sections three or more of the same facts. "facts": [] marks a section that deliberately states no verified facts. When no Verified Facts are listed, use plain strings only.$a1$, ''))) / length($a1$17. Section entries may be plain strings ("hero") or objects ({"name": "features", "facts": ["F1-…"]}). Use the object form when the Verified Facts section above lists facts: put every fact you want stated into exactly ONE section's "facts" list, using the IDs exactly as listed there (an ID not in that list is ignored). Spread facts across sections — never give two sections three or more of the same facts. "facts": [] marks a section that deliberately states no verified facts. When no Verified Facts are listed, use plain strings only.$a1$);
    c2 := (length(t) - length(replace(t, $a2$Assign each fact you want the site to state to exactly ONE section, using the object form of section entries (RULES, rule 17).$a2$, ''))) / length($a2$Assign each fact you want the site to state to exactly ONE section, using the object form of section entries (RULES, rule 17).$a2$);
    c3 := (length(t) - length(replace(t, $a3$"sections": ["hero", {"name": "features", "facts": ["F1-example-id"]}, "testimonials", "call-to-action"]$a3$, ''))) / length($a3$"sections": ["hero", {"name": "features", "facts": ["F1-example-id"]}, "testimonials", "call-to-action"]$a3$);

    IF c1 <> 1 OR c2 <> 1 OR c3 <> 1 THEN
        RAISE EXCEPTION '333: anchor counts must all be exactly 1, got %/%/% — the live prompt has drifted; regenerate this file from a fresh dump', c1, c2, c3;
    END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,prompt_template}',
        to_jsonb(
            replace(replace(replace(
                default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template',
                $a1$17. Section entries may be plain strings ("hero") or objects ({"name": "features", "facts": ["F1-…"]}). Use the object form when the Verified Facts section above lists facts: put every fact you want stated into exactly ONE section's "facts" list, using the IDs exactly as listed there (an ID not in that list is ignored). Spread facts across sections — never give two sections three or more of the same facts. "facts": [] marks a section that deliberately states no verified facts. When no Verified Facts are listed, use plain strings only.$a1$,
                $r1$17. Section entries may be plain strings ("hero") or objects ({"name": "features", "facts": ["F1-…"]}). When the Verified Facts section above lists facts, EVERY section entry on EVERY page must be the object form with an explicit "facts" key — including pages you are carrying over unchanged; a plain-string entry means you have not decided that section. Put every fact you want stated into exactly ONE section's "facts" list, using the IDs exactly as listed there (an ID not in that list is ignored). Decide ownership for every fact in the roster: a fact left out of every list is stated nowhere on the site, so leave one out only deliberately. Spread facts across sections — never give two sections three or more of the same facts. "facts": [] marks a section that deliberately states no verified facts (the right value for most hero, call-to-action and navigation sections). When no Verified Facts are listed, use plain strings only.$r1$),
                $a2$Assign each fact you want the site to state to exactly ONE section, using the object form of section entries (RULES, rule 17).$a2$,
                $r2$Assign each fact you want the site to state to exactly ONE section, using the object form of section entries for EVERY section on EVERY page — including pages carried over unchanged (RULES, rule 17).$r2$),
                $a3$"sections": ["hero", {"name": "features", "facts": ["F1-example-id"]}, "testimonials", "call-to-action"]$a3$,
                $r3$"sections": [{"name": "hero", "facts": []}, {"name": "features", "facts": ["F1-example-id"]}, {"name": "testimonials", "facts": []}, {"name": "call-to-action", "facts": []}]$r3$)
        )
    ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Verify: new text present AND the old permissive spelling gone.
DO $$
DECLARE
    t text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template'
    INTO t
    FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF position('EVERY section entry on EVERY page' in t) = 0
       OR position('for EVERY section on EVERY page' in t) = 0
       OR position($v1$"sections": [{"name": "hero", "facts": []}$v1$ in t) = 0 THEN
        RAISE EXCEPTION '333: verify failed — a new edit is missing from the written-back template';
    END IF;
    IF position('Use the object form when the Verified Facts section above lists facts' in t) > 0 THEN
        RAISE EXCEPTION '333: verify failed — the old permissive rule 17 text survived';
    END IF;
END $$;

COMMIT;

-- ROLLBACK recipe (hand-run):
--   UPDATE agent_definitions ad
--   SET default_config = b.default_config
--   FROM agent_definitions_bak_333 b
--   WHERE ad.id = b.id;
