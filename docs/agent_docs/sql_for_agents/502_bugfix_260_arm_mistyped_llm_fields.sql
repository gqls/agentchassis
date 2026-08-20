-- FILE: docs/agent_docs/sql_for_agents/502_bugfix_260_arm_mistyped_llm_fields.sql
--
-- bugs_open/260 — arm the PRE-RENDER declared-type refusal on BOTH
-- render_component steps of page-content-writer (the only live agent that has
-- any: measured 2026-08-19, 2 steps — render_section and render_from_template).
--
-- ✅ APPLIED BY HAND 2026-08-20 ~14:50Z, after the roll, and the `_HOLD` suffix
-- was dropped in the same commit to say so. Verified independently of the file's
-- own verify block, across the whole live roster rather than the one agent this
-- file names: both `render_component` steps report armed=true and no other live
-- agent has such a step. The re-measured population at apply time was unchanged
-- from the header below — 0 top-level refusals, 5 nested hits ALL of which are
-- the empty-string case the checker does not report.
--
-- The precondition was met and checked at the ARTEFACT, not at a tag: chassis
-- v1.0.1319 on BOTH replicas carries the added literal "refusing to emit output
-- that was not executed", and does NOT carry the deleted fallback's literal
-- "Go template execution failed, using regex fallback" — a removed-string
-- control, which is the strongest form available. Two further controls (a
-- long-lived string that must be present, a nonsense string that must be absent)
-- both behaved.
--
-- ⚠ WAS _HOLD: ORDERING-CRITICAL, THE RUNNER MUST NOT HAVE TAKEN THIS.
-- The Go half (refuseMistypedLLMFields, mistyped_llm_fields_gate.go, and
-- datahelpers.ContentTypeViolations) is committed in 80b9c6235 but INERT until a
-- chassis image built from that commit has rolled. This file only sets a config
-- key, and a config key naming behaviour the running binary does not have is a
-- no-op that LOOKS applied — the worst of both states, because the register
-- entry (STY-057) and the coverage report would then both read "armed" while
-- nothing is gated. Same reasoning, same shape, as 380 for the dead-URL guard.
--
-- APPLY ONLY AFTER the running chassis is confirmed to carry the Go half. Ask
-- the service what it is running — never `strings` (absent from the
-- debian-slim images, and behind the customary 2>/dev/null its failure is
-- indistinguishable from "not stamped"):
--   kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 \
--     | grep -m1 'build provenance'
--   git merge-base --is-ancestor 80b9c6235 <that sha>
-- Read the stamp of THIS service, not the fleet tag: until `release` pins one
-- REF a single tag can straddle several revisions (v1.0.1284 shipped three).
-- If the startup line has scrolled, probe the binary for a KNOWN value with a
-- control in the same breath (bugs_open/249; CLAUDE.md, "Ask the service").
--
-- WHAT IT DOES. Sets `refuse_mistyped_llm_fields: true` on page-content-writer's
-- `render_section` AND `render_from_template` steps, so a component whose
-- content contradicts its own input_schema is refused BEFORE the render rather
-- than after it. The refusal message names the field with an indexed path
-- ("steps[2].branches: declared array (items: object), got string").
--
-- WHAT IT DOES NOT DO, stated because it is the commonest misreading of this
-- change: the SEAM's hard error is already unconditional and is NOT gated by
-- this key. With this file unapplied, a component whose template cannot execute
-- still fails the step loudly and still carries the type diagnosis, because the
-- checker also runs as an unconditional ENRICHER on a render that has already
-- failed. This key buys EARLIER and BETTER-NAMED, not detection.
--
-- WHY IT IS OPT-IN AT ALL (owner ruling 2026-08-02 §2). The checker keys on the
-- SCHEMA, not on the template, so a mistyped field that this component's
-- template never references renders perfectly well today. Refusing it
-- unconditionally would therefore be new authority over content that currently
-- ships. Unsafe side is the default: unset or false == today's behaviour, byte
-- for byte.
--
-- WHAT IT COSTS. With this armed, a page whose stored content_data contradicts
-- the component's declared field types will REFUSE to render: the step fails,
-- save_page_sections never runs, and the stored row plus the live page are left
-- exactly as they were. That is the intent (a named failure beats a page nobody
-- can explain), but it IS a live blocking behaviour.
--
-- KNOWN POPULATION AT WRITING — RE-MEASURE BEFORE APPLYING, the query is below.
-- [MEASURED 2026-08-19] ZERO rows would be refused:
--   * top-level declared-array fields holding a non-array, over every stored
--     page_components row: 0.
--   * the NESTED case (an element of a declared array whose own declared-array
--     sub-field holds a non-array): 5 elements on ONE page —
--     fundamentallyai.com /production-backend-engineering, mechanism-flow,
--     steps[].branches — and every one of them is the EMPTY STRING, which the
--     checker does not report, because absent/nil/empty is the presence gate's
--     question at every declared type (datahelpers.IsEmptyContentValue, shared
--     with missingRequiredLLMFields so the two gates cannot disagree).
--     ⚠ THAT PAGE IS DEPLOYED AND SERVING CLEAN, and it is why the shared
--     emptiness predicate exists: the first version of the checker called ""
--     a violation, which would have refused a rebuild of a healthy live page.
--     It is the ONLY such row on the estate, so nothing but this census would
--     have found it. If a future version of this query returns rows that are
--     NOT empty strings, read them before applying — they are pages that will
--     stop rebuilding.
--
-- RE-MEASURE (read-only, safe to run any time):
--   WITH f AS (
--     SELECT c.id AS comp_id, c.name, e.k AS field
--       FROM content_components c, jsonb_each(c.input_schema->'fields') e(k,v)
--      WHERE c.is_active AND v->>'source'='llm' AND v->>'type' IN ('array','list'))
--   SELECT count(*) FROM page_components pc JOIN f ON f.comp_id = pc.component_id
--    WHERE pc.content_data ? f.field
--      AND jsonb_typeof(pc.content_data->f.field) NOT IN ('array','null')
--      AND btrim(COALESCE(pc.content_data->>f.field,'')) <> '';
--
-- Rollback: 502_bugfix_260_arm_mistyped_llm_fields_ROLLBACK.sql (still a hand-run sidecar) (sets both keys
-- back to false; the Go half then behaves exactly as it did before this file).

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('page-content-writer', '502_bugfix_260_arm_mistyped_llm_fields: pre-update');

\echo '=== BEFORE: the render_section step config ==='
SELECT jsonb_pretty(jsonb_path_query_first(default_config, '$.**.steps.render_section.config'))
  FROM agent_definitions
 WHERE type = 'page-content-writer' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- THE DUPLICATE-ROW GUARD (inherited from 380, council 98852baa). Four agent
-- types on this estate carry TWO active definition rows and only the HIGHER
-- VERSION is loaded at runtime, so an UPDATE ... WHERE type = '<x>' can silently
-- touch a stale duplicate while a verify block that re-reads "the row for this
-- type" confirms whichever one it finds. Measured for THIS type 2026-08-19:
-- exactly 1 active row. The guard exists so that if a second appears before this
-- HOLD is lifted, the file refuses instead of half-applying.
DO $$
DECLARE
    v_rows int;
    v_armed boolean;
BEGIN
    SELECT count(*) INTO v_rows
      FROM agent_definitions
     WHERE type = 'page-content-writer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '260/502: expected exactly 1 live page-content-writer row, found % — this type now carries duplicates and only the HIGHEST VERSION is loaded at runtime; target that row by id, do not update by type', v_rows;
    END IF;

    SELECT (jsonb_path_query_first(default_config,
              '$.**.steps.render_section.config.refuse_mistyped_llm_fields'))::text::boolean
      INTO v_armed
      FROM agent_definitions
     WHERE type = 'page-content-writer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF v_armed IS TRUE THEN
        RAISE EXCEPTION '260/502: already applied — refuse_mistyped_llm_fields is already true';
    END IF;
END $$;

-- PRE-FLIGHT: assert the paths EXIST and hold a render_component step BEFORE any
-- jsonb_set runs. jsonb_set(..., create_missing := true) on a wrong path inserts
-- a whole new branch and reports success — arming nothing while every downstream
-- reader says "armed". 380's round 2 was gated on exactly this.
DO $$
DECLARE
    v_a text;
    v_b text;
BEGIN
    SELECT default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_section,action}',
           default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_from_template,action}'
      INTO v_a, v_b
      FROM agent_definitions
     WHERE type = 'page-content-writer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_a IS DISTINCT FROM 'render_component' THEN
        RAISE EXCEPTION '260/502: render_section is % at the expected path (want render_component) — the workflow shape moved; re-derive the path before arming', COALESCE(v_a, '(absent)');
    END IF;
    IF v_b IS DISTINCT FROM 'render_component' THEN
        RAISE EXCEPTION '260/502: render_from_template is % at the expected path (want render_component) — the workflow shape moved; re-derive the path before arming', COALESCE(v_b, '(absent)');
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           jsonb_set(
               default_config,
               -- The steps live inside the process_sections_loop sub_workflow,
               -- not at the top level: a top-level jsonb_each finds nothing here
               -- and reads as "no such step".
               '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_section,config,refuse_mistyped_llm_fields}',
               'true'::jsonb,
               true),
           '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_from_template,config,refuse_mistyped_llm_fields}',
           'true'::jsonb,
           true),
       updated_at = now()
 WHERE type = 'page-content-writer' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

\echo '=== AFTER ==='
SELECT jsonb_pretty(jsonb_path_query_first(default_config, '$.**.steps.render_section.config'))
  FROM agent_definitions
 WHERE type = 'page-content-writer' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- VERIFY BY COUNTING, not by reading back the one path you just wrote.
-- (1) Every comparison is IS DISTINCT FROM / IS NOT TRUE, because a jsonb path
-- compared with =/<> sits GREEN for ever when the key is ABSENT (NULL <> 'true'
-- is NULL, not TRUE). (2) Counting ALL render_component steps means a future
-- THIRD step fails this file loudly instead of shipping silently unarmed.
DO $$
DECLARE
    v_steps  int;
    v_armed  int;
BEGIN
    SELECT count(*) FILTER (WHERE k.value->>'action' = 'render_component'),
           count(*) FILTER (WHERE k.value->>'action' = 'render_component'
                              AND (k.value->'config'->>'refuse_mistyped_llm_fields')::boolean IS TRUE)
      INTO v_steps, v_armed
      FROM agent_definitions ad,
           LATERAL jsonb_path_query(ad.default_config, 'strict $.**.steps') AS steps,
           LATERAL jsonb_each(steps) AS k
     WHERE ad.type = 'page-content-writer' AND ad.is_active
       AND COALESCE(ad.is_snapshot, false) = false AND ad.deleted_at IS NULL;

    IF v_steps = 0 THEN
        RAISE EXCEPTION '260/502: found ZERO render_component steps — the traversal is wrong, so a green result here would mean nothing; aborting';
    END IF;
    IF v_armed <> v_steps THEN
        RAISE EXCEPTION '260/502: % of % render_component step(s) armed — a step exists that this file does not know about; add it rather than shipping partial coverage the report will call complete', v_armed, v_steps;
    END IF;
    RAISE NOTICE '260/502: declared-type refusal ARMED on ALL % render_component step(s) of page-content-writer', v_steps;
END $$;

COMMIT;

-- ---------------------------------------------------------------------------
-- WATCH after arming. The refusal is a step failure, so it surfaces as one, and
-- the honest first question is whether it EVER fires:
--   SELECT count(*), max(occurred_at) FROM agent_error_log
--    WHERE message ILIKE '%do not match the declared field type%';
-- A sustained zero is the EXPECTED result on today's population (the census
-- above found nothing to refuse) and means the gate is armed and quiet, not
-- broken. What discriminates "armed and quiet" from "not armed" is the config
-- read-back in the verify block plus the binary probe in the header — not this
-- count. The unconditional half (seam error + enricher) is what to watch for
-- real occurrences:
--   SELECT count(*) FROM agent_error_log WHERE message ILIKE '%failed to render%bugs_open/260%';
-- ---------------------------------------------------------------------------
