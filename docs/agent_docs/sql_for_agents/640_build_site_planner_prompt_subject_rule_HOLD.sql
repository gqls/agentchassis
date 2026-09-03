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
-- RE-DERIVED 2026-09-02 against the live row: the first apply attempt was
-- correctly REFUSED by this file's own anchor guard — bugs_open/380's seed had
-- rewritten rule 17's tail ("use plain strings only" -> "still use the object
-- form with facts:[] on every section"). That change HELPS this one (object
-- form is now unconditional, so "subject" can ride everywhere); the subject
-- sentences now insert BEFORE the 380 sentence, which is kept verbatim.
--
-- WHY: pages.sections carried nothing but component names, so N same-named
-- slots got one identical brief — apis.uk served four solitary-bee sections
-- and a content_rewrite rewrote all six sections about the waggle dance.

--
-- ── VERIFY THE ROLL FIRST (council 4bd35ed8 r2, debug_historian: "post-roll" is
--    an assumption until the RUNNING BINARY says so; same-tag rebuilds have
--    shipped nothing before). Two probes, both against the pod, controls in the
--    same breath:
--      kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
--        -> git merge-base --is-ancestor 35905c547 <stamp sha>   (must exit 0)
--      # fallback if the startup line has scrolled (capability probe + controls):
--      P=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1)
--      kubectl -n ai-persona-system exec $P -- grep -ac 'section_subjects' /proc/1/exe   # >0 = shipped
--      kubectl -n ai-persona-system exec $P -- grep -ac 'section_facts' /proc/1/exe      # positive control, must be >0
--
-- ── AFTER HAND-APPLYING: _HOLD files never reach the migration ledger, so the
--    record is THIS FILE — append one line directly below this block and commit
--    it (pathspec):  -- APPLIED <date> by <session>; roll verified at <stamp sha>

-- ⚠ THE ANCHOR HAS EXTERNAL READERS (2026-09-03): migration 729 (bugs_open/450)
--   pins the literal   may also carry a "subject"   in its own verify block as a
--   neighbour-check (three lanes edit this prompt row; 729 refuses to apply if
--   rule 17 has gone missing). Verified 2026-09-03: count exactly 1 on the live
--   row. ANY future re-derivation of rule 17 must keep that literal verbatim, or
--   tell the 450 lane AND update this note. The 443 lane's live detector
--   REPEATED_COMPONENT_BUILT_WITHOUT_SUBJECT is also only interpretable while
--   the rule exists — losing the sentence makes its fire-rate read as planner
--   non-compliance.
-- APPLIED 2026-09-02 by the apis.uk session (SECOND attempt; the first was correctly REFUSED by the anchor guard against bugs_open/380's drift and the seed was re-derived, see header). Same pod verification as 639. Live post-check: subject rule + 380 sentence + example all present.

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

    c1 := (length(t) - length(replace(t, $a1$When no Verified Facts are listed, still use the object form with "facts": [] on every section — a plain string there leaves the writer unconstrained, which is the failure this rule exists to prevent (bugs_open/380).$a1$, ''))) / length($a1$When no Verified Facts are listed, still use the object form with "facts": [] on every section — a plain string there leaves the writer unconstrained, which is the failure this rule exists to prevent (bugs_open/380).$a1$);
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
                $a1$When no Verified Facts are listed, still use the object form with "facts": [] on every section — a plain string there leaves the writer unconstrained, which is the failure this rule exists to prevent (bugs_open/380).$a1$,
                $r1$Any object entry may also carry a "subject": one line saying what THIS section specifically covers, distinct from every sibling section's subject on the page. A "subject" is REQUIRED on every entry whose component name appears more than once on the same page — repeated components without subjects all receive the same brief, and the writer then produces the same section several times. When no Verified Facts are listed, still use the object form with "facts": [] on every section — a plain string there leaves the writer unconstrained, which is the failure this rule exists to prevent (bugs_open/380).$r1$),
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
       OR position('still use the object form with "facts": []' in t) = 0 THEN
        RAISE EXCEPTION '640: verify failed — the subject rule or example is missing, or the old sentence survived';
    END IF;
END $$;

COMMIT;

-- ROLLBACK recipe (hand-run): restore from agent_definitions_bak_640, or from
-- the snapshot_agent row this file took first.
