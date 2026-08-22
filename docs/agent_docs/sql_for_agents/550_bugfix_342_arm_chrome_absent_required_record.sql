-- 550_bugfix_342_arm_chrome_absent_required_record.sql
--
-- bugs_open/342 — arm the RENDER-TIME required_fields_missing record on every
-- live step whose action is render_site_components (the chrome path). The Go
-- half (recordAbsentRequiredFields + emitRequiredFieldsMissing, council trail
-- bb7f5d0e) is LIVE: verified 2026-08-22 at the artefact — commits cd90e8b27 /
-- 65f1b0b95 / af4743464 are ancestors of the v1.0.1323 build stamp 70e7b4f9c,
-- and the stamp was probed IN THE BINARY on both replicas with a nonsense
-- control. So this file is appliable on sight: no ordering constraint, no _HOLD.
--
-- WHAT IT DOES. Sets record_absent_required_fields: true on the step config of
-- every live render_site_components step. Enumerated 2026-08-22 (top level AND
-- one sub_workflow level deep — 502's steps were nested and a top-level-only
-- scan would have missed them; these are not):
--   nav-link-fixer.rerender_site_components,   nav-updater.render_site_components,
--   pageflow-builder.render_site_components,   rerender-chrome.render_site_components,
--   rerender-pages.render_site_components,     rerender-site.render_site_components,
--   site-work-orchestrator.render_site_components
--
-- MEASURED BEFORE ARMING (bugs_open/342 §5's own instruction), 2026-08-22:
-- the chrome store (site_components) references only site-header / site-footer
-- / head class components, which declare ZERO required source:"llm" fields —
-- so this arm fires on 0 rows today. ⚠ That census was VACUOUS on the first
-- attempt (0 candidate pairs — no chrome row references ANY component with
-- required fields); the zero is real but means "no chrome component in use has
-- required fields", not "all fields are supplied". Arming while the population
-- is zero is deliberate: it is free today, and it closes the door BEFORE a
-- chrome component that does declare required fields (five exist in the
-- library: footer-with-disclaimer 17, header-with-categories 16, ...) is ever
-- adopted — the alternative is the door standing open precisely when it starts
-- to matter.
--
-- WHY OPT-IN AT ALL: a new DB write on a shared render path is new authority
-- whatever its size (three seats, council 98852baa; owner ruling 2026-08-02
-- §2). This migration IS the visible, reviewable act of opting in.
--
-- Snapshot per agent first; pre-conditions refuse a double-apply; post-verify
-- is DO/RAISE (a bare SELECT cannot stop the COMMIT) and asserts the key at
-- EVERY step carrying the action, so a step added between enumeration and
-- apply is caught rather than skipped.

SELECT snapshot_agent(t.type, 'migration 550: pre-update (bugs_open/342 chrome record arm)')
FROM (SELECT DISTINCT a.type
      FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') s(k,v)
      WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
        AND s.v->>'action' = 'render_site_components') t;

BEGIN;

-- ── Pre-conditions ──────────────────────────────────────────────────────────
DO $$
DECLARE
    n_steps   integer;
    n_armed   integer;
    n_nested  integer;
BEGIN
    SELECT count(*) INTO n_steps
    FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') s(k,v)
    WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
      AND s.v->>'action' = 'render_site_components';
    IF n_steps < 1 THEN
        RAISE EXCEPTION 'MIGRATION 550: found no live render_site_components steps — the action has moved or been renamed. Re-derive.';
    END IF;

    -- Double-apply / concurrent-arm refusal: no such step may already carry the key.
    SELECT count(*) INTO n_armed
    FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') s(k,v)
    WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
      AND s.v->>'action' = 'render_site_components'
      AND s.v->'config' ? 'record_absent_required_fields';
    IF n_armed > 0 THEN
        RAISE EXCEPTION 'MIGRATION 550: % step(s) already carry record_absent_required_fields — another session armed first. Refusing to overwrite; re-derive against the live rows.', n_armed;
    END IF;

    -- The enumeration above scanned one sub_workflow level too and found none;
    -- refuse if that has changed, because the UPDATE below only writes top level.
    SELECT count(*) INTO n_nested
    FROM agent_definitions a,
         jsonb_each(a.default_config->'workflow'->'steps') s(k,v),
         jsonb_each(COALESCE(s.v->'config'->'sub_workflow'->'steps','{}'::jsonb)) ss(k,v)
    WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
      AND ss.v->>'action' = 'render_site_components';
    IF n_nested > 0 THEN
        RAISE EXCEPTION 'MIGRATION 550: % render_site_components step(s) now live inside a sub_workflow, which this UPDATE cannot reach. Extend the migration (the 502 nested-path shape) before applying.', n_nested;
    END IF;
END $$;

-- ── The write: per matching step, dynamically (step names differ per agent) ──
DO $$
DECLARE
    r record;
    n integer := 0;
BEGIN
    FOR r IN
        SELECT a.id, s.k AS step
        FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') s(k,v)
        WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
          AND s.v->>'action' = 'render_site_components'
    LOOP
        UPDATE agent_definitions
        SET default_config = jsonb_set(default_config,
                ARRAY['workflow','steps', r.step, 'config', 'record_absent_required_fields'],
                'true'::jsonb),
            version    = version + 1,
            updated_at = now()
        WHERE id = r.id;
        n := n + 1;
    END LOOP;
    RAISE NOTICE 'MIGRATION 550: armed % step(s)', n;
END $$;

-- ── Post-conditions: every step carrying the action must now be armed ───────
DO $$
DECLARE
    n_unarmed integer;
    n_armed   integer;
BEGIN
    SELECT count(*) INTO n_unarmed
    FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') s(k,v)
    WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
      AND s.v->>'action' = 'render_site_components'
      AND COALESCE((s.v->'config'->>'record_absent_required_fields')::boolean, false) IS DISTINCT FROM true;
    IF n_unarmed > 0 THEN
        RAISE EXCEPTION 'MIGRATION 550: % render_site_components step(s) still unarmed after the write.', n_unarmed;
    END IF;

    SELECT count(*) INTO n_armed
    FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') s(k,v)
    WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
      AND s.v->>'action' = 'render_site_components'
      AND (s.v->'config'->>'record_absent_required_fields')::boolean IS TRUE;
    IF n_armed < 1 THEN
        RAISE EXCEPTION 'MIGRATION 550: zero armed steps after the write — the loop matched nothing.';
    END IF;
    RAISE NOTICE 'MIGRATION 550: verified — % armed step(s), 0 unarmed', n_armed;
END $$;

COMMIT;
