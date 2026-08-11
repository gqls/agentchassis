-- 385 - build-site-planner: surface recompose_pages to the planner as a
-- prompt-visible field (features_open/012, the field-based fix; owner ruling
-- DECISIONS_2026-08-11 ruling 2 - scheduled immediately after the 151 census).
--
-- Why: the recompose release happens in validate (v3_site_actions.go:3105
-- filters existingPages before convergence), but seed 362 instructs the
-- PLANNER to re-emit every built page's realised list verbatim, with only a
-- prose escape ("only when the briefing explicitly asks..."). The planner is
-- never told WHICH pages are released, so a recompose_pages request whose
-- intent is not ALSO in the briefing silently no-ops (LANDMINES 2026-08-10;
-- durable tell RECOMPOSE_INTENT_NOT_REALISED, live since v1.0.1283). This seed
-- is the RFC_010 section 2 shape: authority ships as a field, not prose.
--
-- What it does - two anchored edits, derived byte-exact from the live row by
-- the session scratchpad generator (never retyped):
--   1. The existing-pages listing row gains a nested range over
--      $.input_data.spec.recompose_pages: a page named there gets an inline
--      "REDESIGN REQUESTED - propose a NEW composition..." marker on its own
--      listing line. input_data is ALREADY an input_field of plan_site, so no
--      input_fields change is needed; the spec travels at input_data.spec
--      (the 012 feature file's proven path).
--   2. The 362 instruction paragraph's escape sentence gains the flag
--      semantics: flagged pages are exactly the explicit-redesign case; their
--      shown sections are context only; unflagged pages keep preserve-exactly.
--
-- Opt-in safety, all measured/proven before this file was written:
--   - Go text/template treats the absent deep chain as falsy (proven
--     empirically against the platform's RenderPromptTemplate semantics:
--     plain text/template, no missingkey option, funcMap toJSON/placeholder/
--     rangeStart/rangeEnd): absent input_data.spec, absent recompose_pages,
--     and [] all render ZERO markers and no "<no value>".
--   - The full post-seed template parses and renders both ways (go run proof,
--     rendered 18,377 B unflagged / 18,527 B with one flagged page, exactly
--     one marked row).
--   - Exactly ONE active non-snapshot build-site-planner row.
--   - Zero needs_site_plan items in all history carry spec.recompose_pages,
--     and zero RECOMPOSE_INTENT_NOT_REALISED rows - nothing changes
--     retroactively.
--   - The other existing_pages consumer (content-gap-planner, its own
--     plan_gaps template) is untouched by this row-scoped UPDATE and does not
--     plan compositions; named per the 2026-07-29 ruling 3 duty to tell.
--
-- Follow-up (registered in features_open/012): once a live recompose run
-- proves the field, a further seed retires the prose escape's load-bearing
-- status in the 362 paragraph.

SELECT snapshot_agent('build-site-planner', '385_build_site_planner_recompose_pages_visible.sql: pre-update');

BEGIN;

CREATE TABLE IF NOT EXISTS agent_definitions_bak_385 AS
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
        RAISE EXCEPTION '385: build-site-planner plan_site prompt_template not found';
    END IF;
    -- Dual-active-row landmine guard (council round 62d2463f objection, same
    -- guard the APPROVED d1e8c36e round put on seeds 386/387/388): four agent
    -- types carry TWO active rows and only the higher version loads. Refuse
    -- unless build-site-planner has EXACTLY ONE active non-snapshot row, so
    -- "UPDATE 1 + verify passed" cannot describe a row the loader never reads.
    IF (SELECT count(*) FROM agent_definitions
         WHERE type = 'build-site-planner' AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL) <> 1 THEN
        RAISE EXCEPTION '385: build-site-planner does not have exactly one active row - resolve the duplicate before seeding';
    END IF;
    IF position('REDESIGN REQUESTED' in t) > 0 THEN
        RAISE EXCEPTION '385: already applied - the marker text is present';
    END IF;

    c1 := (length(t) - length(replace(t, $a1${{range .existing_pages}}- name: {{.name}} | page_type: {{.page_type}} | url: {{.url}}{{if .sections}} | sections: {{.sections}}{{end}}$a1$, ''))) / length($a1${{range .existing_pages}}- name: {{.name}} | page_type: {{.page_type}} | url: {{.url}}{{if .sections}} | sections: {{.sections}}{{end}}$a1$);
    c2 := (length(t) - length(replace(t, $a2$Only when the briefing explicitly asks for a page to be redesigned should you propose a different composition for it, and then its fact assignments must name the NEW sections.$a2$, ''))) / length($a2$Only when the briefing explicitly asks for a page to be redesigned should you propose a different composition for it, and then its fact assignments must name the NEW sections.$a2$);

    IF c1 <> 1 OR c2 <> 1 THEN
        RAISE EXCEPTION '385: anchor counts must both be exactly 1, got %/% - the live prompt has drifted; regenerate this file from a fresh dump', c1, c2;
    END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,prompt_template}',
        to_jsonb(
            replace(replace(
                default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template',
                $a1${{range .existing_pages}}- name: {{.name}} | page_type: {{.page_type}} | url: {{.url}}{{if .sections}} | sections: {{.sections}}{{end}}$a1$,
                $r1${{range .existing_pages}}- name: {{.name}} | page_type: {{.page_type}} | url: {{.url}}{{if .sections}} | sections: {{.sections}}{{end}}{{$pn := .name}}{{range $rp := $.input_data.spec.recompose_pages}}{{if eq $rp $pn}} | REDESIGN REQUESTED - propose a NEW composition for this page (its sections above are shown for context only; the re-emit rule does NOT apply to it){{end}}{{end}}$r1$),
                $a2$Only when the briefing explicitly asks for a page to be redesigned should you propose a different composition for it, and then its fact assignments must name the NEW sections.$a2$,
                $r2$Only when the briefing explicitly asks for a page to be redesigned should you propose a different composition for it, and then its fact assignments must name the NEW sections. Pages flagged "REDESIGN REQUESTED" in the listing below are exactly that case, named explicitly by the operator: for EACH flagged page, do NOT re-emit its current section list - propose a fresh composition from the Available Section Components, and put its fact assignments on the NEW section names. Re-emit a flagged page's current list unchanged only if you deliberately judge the current composition already right - that choice is recorded and reviewed. Pages NOT flagged keep the preserve-exactly rule above.$r2$)
        )
    ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Verify with DO/RAISE - a verify block of SELECTs cannot stop the COMMIT.
DO $$
DECLARE
    t text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template'
    INTO t
    FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF position($v1${{range $rp := $.input_data.spec.recompose_pages}}$v1$ in t) = 0 THEN
        RAISE EXCEPTION '385: verify failed - the per-row marker range is missing';
    END IF;
    IF position($v2$Pages flagged "REDESIGN REQUESTED" in the listing below$v2$ in t) = 0 THEN
        RAISE EXCEPTION '385: verify failed - the flag-semantics instruction is missing';
    END IF;
    IF length(t) <> 20445 THEN
        RAISE EXCEPTION '385: verify failed - post length % <> expected 20445', length(t);
    END IF;
END $$;

COMMIT;

-- ROLLBACK recipe (hand-run):
--   UPDATE agent_definitions ad
--   SET default_config = b.default_config
--   FROM agent_definitions_bak_385 b
--   WHERE ad.id = b.id;
