-- 640 — build-site-planner: rule 17 gains "subject", required on repeated
-- components.
--
-- ⚠ _HOLD: apply BY HAND, AFTER the image carrying the subject rails has
-- rolled AND seed 639 is applied. Before that, a planner that emits subjects
-- is harmless (extractSectionEntries and the normalise pass ignore unknown
-- keys on old binaries; on the new binary the value is stored and simply
-- never consumed until 639/641) — but holding it keeps one ordering to
-- reason about: image -> 638 (already applied) -> 639 -> 640 -> 641.
--
-- WHAT THIS CHANGES (surgical, anchored on the LIVE text, aborts on drift):
--  1. RULES rule 17: object entries MAY carry "subject" (one line, what THIS
--     section specifically covers); REQUIRED when the same component name
--     appears more than once on a page; and the no-facts arm now permits
--     {"name","subject"} objects instead of forcing plain strings.
--  2. The JSON example shows a repeated component with two DISTINCT subjects.
--
-- WHY: pages.sections carried nothing but component names, so N same-named
-- slots got one identical brief — apis.uk served four solitary-bee sections
-- and a content_rewrite rewrote all six sections about the waggle dance.

SELECT snapshot_agent('build-site-planner', '640_build_site_planner_prompt_subject_rule_HOLD.sql: pre-update');

BEGIN;

CREATE TABLE IF NOT EXISTS agent_definitions_bak_640 AS
SELECT id, type, default_config, now() AS backed_up_at
FROM agent_definitions
WHERE type = 'build-site-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
    t text;
    c1 int; c2 int; nrows int;
BEGIN
    -- Pre-flight (council 4bd35ed8 round 1): SELECT INTO takes the FIRST of N
    -- rows silently; a duplicate active row would half-apply. Count first.
    SELECT count(*) INTO nrows FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF nrows <> 1 THEN
        RAISE EXCEPTION '640: expected exactly 1 active build-site-planner row BEFORE writing, found %', nrows;
    END IF;
    SELECT default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template'
    INTO t
    FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF t IS NULL THEN
        RAISE EXCEPTION '640: build-site-planner plan_site prompt_template not found';
    END IF;
    IF position('may also carry a "subject"' in t) > 0 THEN
        RAISE EXCEPTION '640: already applied — subject rule present';
    END IF;

    c1 := (length(t) - length(replace(t, $a1$When no Verified Facts are listed, use plain strings only.$a1$, ''))) / length($a1$When no Verified Facts are listed, use plain strings only.$a1$);
    c2 := (length(t) - length(replace(t, $a2$"sections": [{"name": "hero", "facts": []}, {"name": "features", "facts": ["F1-example-id"]}, {"name": "call-to-action", "facts": []}]$a2$, ''))) / length($a2$"sections": [{"name": "hero", "facts": []}, {"name": "features", "facts": ["F1-example-id"]}, {"name": "call-to-action", "facts": []}]$a2$);
    IF c1 <> 1 OR c2 <> 1 THEN
        RAISE EXCEPTION '640: live prompt has drifted (anchor counts %/%; both must be 1) — re-derive this seed from the live row', c1, c2;
    END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,prompt_template}',
        to_jsonb(
            replace(replace(
                default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template',
                $a1$When no Verified Facts are listed, use plain strings only.$a1$,
                $r1$Any object entry may also carry a "subject": one line saying what THIS section specifically covers, distinct from every sibling section's subject on the page. A "subject" is REQUIRED on every entry whose component name appears more than once on the same page — repeated components without subjects all receive the same brief, and the writer then produces the same section several times. When no Verified Facts are listed, use plain strings, or {"name": …, "subject": …} objects where a component repeats.$r1$),
                $a2$"sections": [{"name": "hero", "facts": []}, {"name": "features", "facts": ["F1-example-id"]}, {"name": "call-to-action", "facts": []}]$a2$,
                $r2$"sections": [{"name": "hero", "facts": []}, {"name": "features", "facts": ["F1-example-id"], "subject": "What the platform does"}, {"name": "features", "facts": [], "subject": "How a team adopts it"}, {"name": "call-to-action", "facts": []}]$r2$)
        )
    ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE t text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template'
    INTO t
    FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF position('may also carry a "subject"' in t) = 0
       OR position('"subject": "How a team adopts it"' in t) = 0
       OR position('When no Verified Facts are listed, use plain strings only.' in t) > 0 THEN
        RAISE EXCEPTION '640: verify failed — the subject rule or example is missing, or the old sentence survived';
    END IF;
END $$;

COMMIT;

-- ROLLBACK recipe (hand-run): restore from agent_definitions_bak_640, or from
-- the snapshot_agent row this file took first.
