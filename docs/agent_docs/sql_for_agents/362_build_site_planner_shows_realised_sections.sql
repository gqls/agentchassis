-- 362 — build-site-planner: SHOW the realised section list, and scope fact
-- assignment to it (bugs_open/151 candidate 1b (i); RFC_016 §3b).
--
-- HELD: this is the prompt half of candidate 1b and belongs to the Slice B
-- council round, which is not submitted yet. _HOLD in the filename because a
-- banner cannot hold a file — run-migrations.sh's SIDECAR_RE excludes _HOLD
-- while still listing it, and a migration's guard checks DRIFT, not ORDER.
--
-- Why: the planner is already GIVEN existing_pages (it is an input_field of
-- plan_site, and load_existing_pages already selects p.sections), but the
-- prompt only ever printed name | page_type | url. So the planner could not
-- know what a built page is composed of, re-composed it freely, and — since
-- seed 333 made fact assignment mandatory per section — attached its fact
-- assignments to section names the built page does not have. validate_plan
-- then restores the realised composition and, before candidate 1b (ii), those
-- assignments were discarded with the entries that carried them. Measured on
-- fundamentallyai 2026-08-08, corr 1cb17b11 (RFC_016 §3b, finding 2).
--
-- (ii), the Go half, carries assignments onto the restored sections by name and
-- records the misses (FACT_CARRY_UNMATCHED_SECTION). This seed is what makes
-- that match near-total rather than coincidental: shown the realised list, the
-- planner re-emits it, the composition compares EQUAL, and nothing has to be
-- restored or carried at all.
--
-- Two anchored edits, both derived byte-exact from the live row by
-- scratchpad/gen_seed_362.py (never retyped):
--   1. The existing-pages listing line gains "| sections: [...]" when the page
--      has a composition. Rendered from p.sections, which query_database
--      stringifies, so it prints as a JSON array of component names.
--   2. The existing-pages instruction paragraph gains the rule: re-emit that
--      exact list, put this page's fact assignments on THOSE names, treat
--      "sections: []" as rendered-elsewhere, and redesign only on an explicit
--      brief.
--
-- Deliberately NOT absolute: the last sentence preserves the explicit-redesign
-- path, because `recompose_pages` releases a page from the preservation guard
-- and an unconditional "always re-emit" would defeat it. The planner is not
-- told which pages those are, so the escape has to live in the instruction.
--
-- Safe against roll order, for the same reason 329 and 333 were: nothing
-- consumes fact assignments until seeds 328/330 apply (both still _HOLD at
-- HEAD). Verified 2026-08-10 — agent_definitions matching section_facts /
-- facts_scoped / assigned_fact_ids = 0 / 0 / 0, with the LIKE predicate proved
-- able to match (workflow 186, evidence_base 9) against 185 live agents.

SELECT snapshot_agent('build-site-planner', '362_build_site_planner_shows_realised_sections_HOLD.sql: pre-update');

BEGIN;

CREATE TABLE IF NOT EXISTS agent_definitions_bak_362 AS
SELECT id, type, default_config, now() AS backed_up_at
FROM agent_definitions
WHERE type = 'build-site-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
    t text;
    c1 int; c2 int;
BEGIN
    SELECT default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template'
    INTO t
    FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF t IS NULL THEN
        RAISE EXCEPTION '362: build-site-planner plan_site prompt_template not found';
    END IF;
    IF position($g1${{if .sections}} | sections: {{.sections}}{{end}}$g1$ in t) > 0 THEN
        RAISE EXCEPTION '362: already applied — the sections column is present in the listing line';
    END IF;

    c1 := (length(t) - length(replace(t, $a1${{range .existing_pages}}- name: {{.name}} | page_type: {{.page_type}} | url: {{.url}}$a1$, ''))) / length($a1${{range .existing_pages}}- name: {{.name}} | page_type: {{.page_type}} | url: {{.url}}$a1$);
    c2 := (length(t) - length(replace(t, $a2$Do not invent placeholder or example pages such as "post".$a2$, ''))) / length($a2$Do not invent placeholder or example pages such as "post".$a2$);

    IF c1 <> 1 OR c2 <> 1 THEN
        RAISE EXCEPTION '362: anchor counts must both be exactly 1, got %/% — the live prompt has drifted; regenerate this file from a fresh dump', c1, c2;
    END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,prompt_template}',
        to_jsonb(
            replace(replace(
                default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template',
                $a1${{range .existing_pages}}- name: {{.name}} | page_type: {{.page_type}} | url: {{.url}}$a1$,
                $r1${{range .existing_pages}}- name: {{.name}} | page_type: {{.page_type}} | url: {{.url}}{{if .sections}} | sections: {{.sections}}{{end}}$r1$),
                $a2$Do not invent placeholder or example pages such as "post".$a2$,
                $r2$Do not invent placeholder or example pages such as "post". Each existing page is listed with the "sections:" it is actually built from. That composition is preserved when the plan is validated, so for a page shown with a section list, re-emit that exact list, in that order, as its "sections" array — do not re-compose it and do not substitute similar component names. This matters for fact assignment (RULES, rule 17): an assignment travels inside the section entry you emit, so a fact you attach to a section name this page does not have is discarded silently. Put each carried-over page's fact assignments on the section names shown for it. A page shown as "sections: []" is rendered by another part of the system — emit an empty "sections" array for it and assign it no facts. Only when the briefing explicitly asks for a page to be redesigned should you propose a different composition for it, and then its fact assignments must name the NEW sections.$r2$)
        )
    ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Verify: both additions present in the written-back template. A DO/RAISE, not
-- a SELECT — ON_ERROR_STOP ignores a non-empty result, so a verify block of
-- SELECTs cannot stop the COMMIT.
DO $$
DECLARE
    t text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template'
    INTO t
    FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF position($v1${{if .sections}} | sections: {{.sections}}{{end}}$v1$ in t) = 0 THEN
        RAISE EXCEPTION '362: verify failed — the listing line did not gain the sections column';
    END IF;
    IF position($v2$re-emit that exact list, in that order$v2$ in t) = 0 THEN
        RAISE EXCEPTION '362: verify failed — the re-emit instruction is missing';
    END IF;
END $$;

COMMIT;

-- ROLLBACK recipe (hand-run):
--   UPDATE agent_definitions ad
--   SET default_config = b.default_config
--   FROM agent_definitions_bak_362 b
--   WHERE ad.id = b.id;
