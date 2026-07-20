-- 180_tool_improver_rerender_request.sql
--
-- bugs_open/024 — a tool-improver fix is written to
-- content_components.html_template and NEVER rendered onto the page, so every
-- re-verification tests an unchanged page and the benchmark tool cannot go
-- green. Three defects in series; this migration closes the config half.
--
-- WHAT WAS WRONG WITH THIS STEP. tool-improver's create_rerender_item creates a
-- needs_rerender item, and rerender-pages' create_rerender_items step reads
-- reason <- input_data.spec.reason and component_id <- input_data.spec.component_id
-- off that item. It sets scoped = (reason IN ('section_data_resolved',
-- 'image_landed')) AND component_id <> '', and only a scoped run stamps the
-- reason onto the per-page page_rerender items — which is exactly what
-- page-rerender's check_rerender_mode gates the section re-render on. This step
-- supplied NEITHER value, so every request fell to else_step render_page =
-- rerender_single_page ("Simple concatenation - no template re-rendering"),
-- which re-shipped the stale stored HTML and reported success.
--
-- It also keyed the item site-wide (item_key_prefix + domain), so
-- insertWorkItem's anti-churn block counted two SUCCESSFUL predecessors as two
-- failed attempts and branded every later request 'unresolved'. Live evidence
-- on page f25dd4d8-6e25-44eb-a021-689d3057d7a3: needs_rerender complete on
-- 2026-07-17 15:11 and 20:00, then '[unresolved after 2 attempts]' on 07-18
-- 10:02 and 07-19 10:40.
--
-- THE GO HALF IS ALREADY LIVE. spec_literal, spec_paths, item_key_suffix_field
-- and recurrence_expected shipped in v1.0.1140 (commit bca5d8255) and are
-- verified present in the running chassis binary. They are additive and
-- default-off, so until this migration lands they do nothing. Applying this
-- file is what activates them, and ONLY for this step.
--
-- ORDERING IS SAFE EITHER WAY. This migration only ADDS keys to an existing
-- step's config: an old binary ignores unknown config keys, and a new binary
-- with the keys absent takes the default-off branches. Nothing is deleted, so
-- complete.config.output_fields' reference to "rerender_item" stays valid.
--
-- ROLLBACK: 180_tool_improver_rerender_request_ROLLBACK.sql
-- VERIFY:   180_tool_improver_rerender_request_VERIFY.sql

-- Defensive: a prior aborted statement in this session would leave a sticky
-- failed transaction, inside which everything below silently no-ops while
-- appearing to run. Harmless when there is no open transaction.
ROLLBACK;

BEGIN;

-- ── Pre-flight: count what we are about to touch ────────────────────────────
-- The needle gate below is silent about the difference between "already
-- applied" and "target missing", and those need different responses. Fail
-- loudly on anything but exactly one un-patched active row.
DO $$
DECLARE
    target_count int;
    already_patched int;
BEGIN
    SELECT count(*) INTO target_count
    FROM agent_definitions
    WHERE type = 'tool-improver'
      AND is_active = true
      AND default_config #> '{workflow,steps,create_rerender_item,config}' IS NOT NULL;

    SELECT count(*) INTO already_patched
    FROM agent_definitions
    WHERE type = 'tool-improver'
      AND is_active = true
      AND default_config #> '{workflow,steps,create_rerender_item,config}' ? 'spec_literal';

    IF already_patched > 0 THEN
        RAISE EXCEPTION '180: already applied (% row(s) carry spec_literal) — this file is not idempotent by design; re-running would mask a divergent state', already_patched;
    END IF;

    IF target_count <> 1 THEN
        RAISE EXCEPTION '180: expected exactly 1 active tool-improver row with a create_rerender_item config, found % — the roster moved, re-diff before applying', target_count;
    END IF;

    RAISE NOTICE '180: pre-flight OK — 1 target row, 0 already patched';
END $$;

-- Snapshot before any agent_definitions change (platform convention).
-- NOTE: tool-improver-guardtest is deliberately NOT patched. It is the write
-- guard's test fixture; giving it a live re-render request would have it
-- dispatch real work on every guard test.
SELECT snapshot_agent(
    'tool-improver',
    'bugs_open/024: stamp reason+component_id on the rerender request, scope the item_key per component, opt into recurrence'
);

-- ── The change ──────────────────────────────────────────────────────────────
-- Needle-gated on the pre-state so a re-run is a 0-row no-op rather than a
-- second, differently-shaped write. RETURNING asserts the post-conditions
-- inline so the applying session sees the result, not just a row count.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,create_rerender_item,config}',
        (default_config #> '{workflow,steps,create_rerender_item,config}')
          || jsonb_build_object(
               -- A 'complete' predecessor of an ACTION REQUEST is a SUCCESS,
               -- not a strike. Skips the anti-churn heuristics only; dedup
               -- (idx_swi_dedup) is NOT waived.
               'recurrence_expected', true,
               -- What create_rerender_items gates 'scoped' on, and what
               -- check_rerender_mode gates the section re-render on.
               'spec_literal', jsonb_build_object('reason', 'section_data_resolved'),
               -- The other half of 'scoped'. update_result is update_component's
               -- output_field, so this is the component whose write just
               -- SUCCEEDED (a failed write routes to refuse_mangled_write and
               -- never reaches this step). An unresolved path is a hard error
               -- in create_work_item by design: a spec missing component_id
               -- silently degrades to assemble-from-stale and still reports
               -- success, which is this very bug.
               'spec_paths', jsonb_build_object('component_id', 'update_result.component_id'),
               -- '<prefix>_<domain>' is site-wide: two tools fixed close
               -- together collide on idx_swi_dedup and one request is lost.
               'item_key_suffix_field', 'update_result.component_id'
             )
    ),
    updated_at = NOW()
WHERE type = 'tool-improver'
  AND is_active = true
  AND default_config #> '{workflow,steps,create_rerender_item,config}' IS NOT NULL
  AND NOT (default_config #> '{workflow,steps,create_rerender_item,config}' ? 'spec_literal')
RETURNING
    type,
    default_config #>> '{workflow,steps,create_rerender_item,config,spec_literal,reason}'   AS reason,
    default_config #>> '{workflow,steps,create_rerender_item,config,spec_paths,component_id}' AS component_path,
    default_config #>> '{workflow,steps,create_rerender_item,config,item_key_suffix_field}' AS key_suffix,
    (default_config #> '{workflow,steps,create_rerender_item,config}' ->> 'recurrence_expected')::boolean AS recurrence,
    -- The step must still exist and still be referenced, or we have recreated
    -- round 4's dangling-reference defect.
    (default_config #> '{workflow,steps}' ? 'create_rerender_item') AS step_present;

-- ── Post-condition: exactly one row changed, and it changed correctly ───────
DO $$
DECLARE
    patched int;
BEGIN
    SELECT count(*) INTO patched
    FROM agent_definitions
    WHERE type = 'tool-improver'
      AND is_active = true
      AND default_config #>> '{workflow,steps,create_rerender_item,config,spec_literal,reason}' = 'section_data_resolved'
      AND default_config #>> '{workflow,steps,create_rerender_item,config,spec_paths,component_id}' = 'update_result.component_id'
      AND (default_config #> '{workflow,steps,create_rerender_item,config}' ->> 'recurrence_expected')::boolean IS TRUE
      AND default_config #> '{workflow,steps}' ? 'create_rerender_item';

    IF patched <> 1 THEN
        RAISE EXCEPTION '180: post-condition failed — % fully-patched rows, expected 1', patched;
    END IF;

    RAISE NOTICE '180: post-condition OK — 1 row patched, step still present';
END $$;

-- Workflow-altering migrations leave a ('pipeline','build') note so the change
-- is legible from the travelling docs, not only from git.
INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES (
    'pipeline',
    'build',
    E'## tool-improver now REQUESTS a real re-render (bugs_open/024)\n\n'
    'Observed: three tool-improver cycles wrote a correct fix to '
    'content_components.html_template and reported success, while '
    'page_components.rendered_html stayed at its v1 born length (9,901 chars '
    'against a 10,626-char template). Every acceptance re-verification tested '
    'an unchanged page, so the benchmark tool could never go green.\n\n'
    'Root cause (config half): this step supplied no spec.reason and no '
    'spec.component_id, so rerender-pages'' create_rerender_items computed '
    'scoped=false and never stamped a reason; page-rerender''s '
    'check_rerender_mode then took else_step render_page (rerender_single_page, '
    '"Simple concatenation - no template re-rendering") and re-shipped the '
    'stale HTML. The item was also keyed site-wide, so insertWorkItem''s '
    'anti-churn block read two SUCCESSFUL predecessors as two failed attempts '
    'and branded later requests ''unresolved''.\n\n'
    'Fix: stamp reason=section_data_resolved (spec_literal) and component_id '
    '(spec_paths, from update_result — the write that actually succeeded), '
    'scope the item_key per component, and set recurrence_expected so a '
    'successful predecessor is not counted as a strike. Dedup is unchanged.\n\n'
    'New create_work_item affordances available to any step author: '
    'spec_literal (constants), spec_paths (per-key paths, unresolved = hard '
    'error), item_key_suffix_field (scopes the site-wide key), '
    'recurrence_expected (skips anti-churn only). All default-off; the Go '
    'shipped in v1.0.1140.\n\n'
    'Verified: pending — apply this migration, then re-run '
    'improve -> rerender -> acceptance and confirm rendered_html carries the '
    'minmax(0, 2fr) rule and mobile-fit@mobile passes.\n\n'
    'Categories: fix, migration',
    '["fix", "migration"]'::jsonb,
    'migration',
    '180_tool_improver_rerender_request.sql'
);

INSERT INTO schema_migrations (filename)
VALUES ('180_tool_improver_rerender_request.sql');

COMMIT;
