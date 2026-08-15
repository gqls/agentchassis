-- 434_spec_data_map_is_never_read_four_producers_file_empty_specs.sql
--
-- Four live create_work_item steps pass `spec_data` as an INLINE MAP of
-- {key: path}. The action's contract says spec_data is a PATH STRING to a map
-- (create_work_item_action.go:34, :257); a map value resolves through NO
-- extraction strategy — Strategy 0/4 require strings, Strategy 5 deliberately
-- leaves composites alone ("no evidence they were ever intended as literals",
-- action_inputs.go) — so GetMap("spec_data") returns nil and every item these
-- steps file carries spec = '{}'. Measured: 233/233 audit_fix improve_tool
-- items ever created have no spec.page_id; the map shape the authors wrote is
-- exactly the contract of `spec_paths` ({key: path}, per-key resolution), with
-- constants belonging in `spec_literal`. Note the strict config validation of
-- bugs_open/234 checks KEY NAMES, not value types, which is why seed-time
-- validation passed; bugs_open/234's own text describes this failure class
-- ("every item they filed carried spec = '{}'") for the retired `spec` key.
--
-- The four steps (census 2026-08-15, object-typed spec_data, live rows only):
--   tool-auditor              create_items_loop > create_improve_item  (improve_tool / tool-improver)
--   tool-auditor              create_items_loop > create_review_item   (needs_human_review / hitl-review)
--   internal-linker           create_items_loop > create_rewrite_item  (content_rewrite / page-build-handler)
--   component-quality-auditor create_regen_items > create_work_item    (needs_component_regeneration / component-creator)
--
-- Live damage being repaired here: tool-improver's load_tool (seed 426) reads
-- input_data.spec.page_id and the dispatcher maps issue <- current_item.spec.issue,
-- so every audit_fix item hard-fails at load_tool ("query param path
-- 'input_data.spec.page_id' resolved to nil"). Four actionable items on
-- webdesign.co.uk are backfilled from their own row columns + summary below
-- (row page_id/component_id came from the SIBLING config keys, which resolve
-- from the same tool_data paths — the intended spec values are recoverable).
--
-- Deliberately DROPPED from the new spec shape: category / confidence /
-- fix_suggestion. They are LLM-emitted (llm_audit is output_format json with no
-- schema) so presence is not guaranteed, spec_paths hard-errors on a missing or
-- empty value (by design — bugs_open/024), and no live consumer reads them
-- (tool-improver reads spec.page_id + spec.issue only; hitl-review reads the
-- summary). Losing an occasional whole finding to keep optional garnish would
-- be the wrong trade.
--
-- internal-linker note: content_rewrite consumers are the bugs_open/271 lane's
-- surface (their reader fix, commit 9a7d23c49, is at HEAD but not yet rolled).
-- On the live binary the newly-populated spec is inert (271's premise is that
-- nothing reads the brief); when their fix rolls, the brief this seed restores
-- is what travels. Producer-side fix only; noted in their bug file.
--
-- Config is live immediately; no Go ships with it.

ROLLBACK;

BEGIN;

DO $$
DECLARE
    n int;
BEGIN
    -- Drift guards: each target step still carries an object-typed spec_data
    -- and no spec_paths (i.e. nobody fixed or half-fixed it since the census).
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type='tool-auditor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
      AND jsonb_typeof(default_config #> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_improve_item,config,spec_data}') = 'object'
      AND NOT (default_config #> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_improve_item,config}') ? 'spec_paths'
      AND jsonb_typeof(default_config #> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config,spec_data}') = 'object'
      AND NOT (default_config #> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config}') ? 'spec_paths';
    IF n <> 1 THEN
        RAISE EXCEPTION '434: expected exactly 1 active tool-auditor with map-valued spec_data on both loop steps, found % — re-census before applying', n;
    END IF;

    SELECT count(*) INTO n FROM agent_definitions
    WHERE type='internal-linker' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
      AND jsonb_typeof(default_config #> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_rewrite_item,config,spec_data}') = 'object'
      AND NOT (default_config #> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_rewrite_item,config}') ? 'spec_paths';
    IF n <> 1 THEN
        RAISE EXCEPTION '434: expected exactly 1 active internal-linker with map-valued spec_data, found % — re-census before applying', n;
    END IF;

    SELECT count(*) INTO n FROM agent_definitions
    WHERE type='component-quality-auditor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
      AND jsonb_typeof(default_config #> '{workflow,steps,create_regen_items,config,sub_workflow,steps,create_work_item,config,spec_data}') = 'object'
      AND NOT (default_config #> '{workflow,steps,create_regen_items,config,sub_workflow,steps,create_work_item,config}') ? 'spec_paths';
    IF n <> 1 THEN
        RAISE EXCEPTION '434: expected exactly 1 active component-quality-auditor with map-valued spec_data, found % — re-census before applying', n;
    END IF;

    -- Backfill guards: the four webdesign items still have the empty spec.
    SELECT count(*) INTO n FROM site_work_items
    WHERE id IN ('949c9421-12eb-42fc-8205-742a36875bee',
                 'bdf0f88e-c5ae-42e8-a56d-a067a6a93157',
                 '79ff36c0-0970-4063-8eaa-16a44e61f1a6',
                 '28901c68-6640-4134-81dc-dff035858690')
      AND spec = '{}'::jsonb AND page_id IS NOT NULL AND component_id IS NOT NULL;
    IF n <> 4 THEN
        RAISE EXCEPTION '434: expected the 4 webdesign audit_fix items with spec={} and row-level ids, found % — another session may have touched them; re-read before applying', n;
    END IF;

    RAISE NOTICE '434: pre-flight OK — 3 agent rows un-migrated, 4 items awaiting backfill';
END $$;

SELECT snapshot_agent('tool-auditor',
    '434: spec_data passed as a map is never read (contract: path string) — both loop create steps move to spec_paths/spec_literal');
SELECT snapshot_agent('internal-linker',
    '434: spec_data passed as a map is never read — create_rewrite_item moves to spec_paths/spec_literal');
SELECT snapshot_agent('component-quality-auditor',
    '434: spec_data passed as a map is never read — create_work_item moves to spec_paths');

UPDATE agent_definitions
SET default_config =
    jsonb_set(
        jsonb_set(
            default_config,
            '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_improve_item,config}',
            (default_config #> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_improve_item,config}') - 'spec_data'
            || '{"spec_literal": {"check": "tool_auditor"},
                 "spec_paths": {"issue": "current_finding.description",
                                "page_id": "tool_data.page_id",
                                "page_name": "tool_data.page_name",
                                "component_id": "tool_data.component_id"}}'::jsonb
        ),
        '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config}',
        (default_config #> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config}') - 'spec_data'
        || '{"spec_literal": {"check": "tool_auditor"},
             "spec_paths": {"issue": "current_finding.description",
                            "page_id": "tool_data.page_id",
                            "page_name": "tool_data.page_name",
                            "component_id": "tool_data.component_id",
                            "tool_function": "tool_data.function"}}'::jsonb
    ),
    updated_at = NOW()
WHERE type='tool-auditor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND jsonb_typeof(default_config #> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_improve_item,config,spec_data}') = 'object';

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_rewrite_item,config}',
        (default_config #> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_rewrite_item,config}') - 'spec_data'
        || '{"spec_literal": {"source": "internal-linker"},
             "spec_paths": {"page_name": "current_link.source_page",
                            "suggestion": "current_link.guidance",
                            "anchor_text": "current_link.anchor_text",
                            "link_target_url": "target_page.url",
                            "link_target_title": "target_page.title"}}'::jsonb
    ),
    updated_at = NOW()
WHERE type='internal-linker' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND jsonb_typeof(default_config #> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_rewrite_item,config,spec_data}') = 'object';

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,create_regen_items,config,sub_workflow,steps,create_work_item,config}',
        (default_config #> '{workflow,steps,create_regen_items,config,sub_workflow,steps,create_work_item,config}') - 'spec_data'
        || '{"spec_paths": {"function": "current_component.function",
                            "component_id": "current_component.component_id",
                            "quality_score": "current_component.quality_score",
                            "quality_issues": "current_component.quality_issues"}}'::jsonb
    ),
    updated_at = NOW()
WHERE type='component-quality-auditor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND jsonb_typeof(default_config #> '{workflow,steps,create_regen_items,config,sub_workflow,steps,create_work_item,config,spec_data}') = 'object';

-- Backfill the four webdesign audit_fix items from their own row columns; the
-- summary is byte-identical to the intended spec.issue (both were configured
-- from current_finding.description).
UPDATE site_work_items swi
SET spec = jsonb_build_object(
        'check', 'tool_auditor',
        'issue', swi.summary,
        'page_id', swi.page_id::text,
        'page_name', p.name,
        'component_id', swi.component_id::text),
    updated_at = now()
FROM pages p
WHERE p.id = swi.page_id
  AND swi.id IN ('949c9421-12eb-42fc-8205-742a36875bee',
                 'bdf0f88e-c5ae-42e8-a56d-a067a6a93157',
                 '79ff36c0-0970-4063-8eaa-16a44e61f1a6',
                 '28901c68-6640-4134-81dc-dff035858690')
  AND swi.spec = '{}'::jsonb;

-- Re-arm the exhausted items whose only defect was the empty spec. The ab-test
-- one (bdf0f88e) is DELIBERATELY left 'failed': its page is mid-repair
-- (ported slot retired, owner-gate rerender queued) and its finding was taken
-- against the pre-repair page — re-assess it at the served artefact first.
UPDATE site_work_items
SET status='triaged', attempt_count=0, error=NULL, updated_at=now()
WHERE id IN ('949c9421-12eb-42fc-8205-742a36875bee',
             '79ff36c0-0970-4063-8eaa-16a44e61f1a6')
  AND status='failed';

UPDATE site_work_items
SET attempt_count=0, error=NULL, updated_at=now()
WHERE id='28901c68-6640-4134-81dc-dff035858690' AND status='triaged';

DO $$
DECLARE
    n int;
BEGIN
    -- No live step, at any nesting depth we know of, still carries an
    -- object-typed spec_data.
    SELECT count(*) INTO n
    FROM agent_definitions ad,
         jsonb_each(ad.default_config->'workflow'->'steps') outer_steps,
         jsonb_each(coalesce(outer_steps.value->'config'->'sub_workflow'->'steps', '{}'::jsonb)) inner_steps
    WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
      AND jsonb_typeof(inner_steps.value->'config'->'spec_data') = 'object';
    IF n <> 0 THEN
        RAISE EXCEPTION '434: post-condition failed — % nested step(s) still carry map-valued spec_data', n;
    END IF;

    SELECT count(*) INTO n
    FROM agent_definitions ad,
         jsonb_each(ad.default_config->'workflow'->'steps') steps
    WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
      AND jsonb_typeof(steps.value->'config'->'spec_data') = 'object';
    IF n <> 0 THEN
        RAISE EXCEPTION '434: post-condition failed — % top-level step(s) still carry map-valued spec_data', n;
    END IF;

    -- The improver contract keys are now wired: the improve step's spec_paths
    -- carries the two keys with live readers.
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type='tool-auditor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
      AND default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_improve_item,config,spec_paths,page_id}' = 'tool_data.page_id'
      AND default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_improve_item,config,spec_paths,issue}' = 'current_finding.description';
    IF n <> 1 THEN
        RAISE EXCEPTION '434: post-condition failed — tool-auditor improve step spec_paths not as expected';
    END IF;

    -- All four items carry a resolvable spec.page_id; two are re-armed.
    SELECT count(*) INTO n FROM site_work_items
    WHERE id IN ('949c9421-12eb-42fc-8205-742a36875bee',
                 'bdf0f88e-c5ae-42e8-a56d-a067a6a93157',
                 '79ff36c0-0970-4063-8eaa-16a44e61f1a6',
                 '28901c68-6640-4134-81dc-dff035858690')
      AND coalesce(spec->>'page_id','') <> '' AND coalesce(spec->>'issue','') <> '';
    IF n <> 4 THEN
        RAISE EXCEPTION '434: post-condition failed — % of 4 items backfilled', n;
    END IF;

    SELECT count(*) INTO n FROM site_work_items
    WHERE id IN ('949c9421-12eb-42fc-8205-742a36875bee','79ff36c0-0970-4063-8eaa-16a44e61f1a6')
      AND status='triaged' AND attempt_count=0;
    IF n <> 2 THEN
        RAISE EXCEPTION '434: post-condition failed — % of 2 items re-armed', n;
    END IF;

    RAISE NOTICE '434: post-condition OK — 0 map-valued spec_data steps live, 4 items backfilled, 2 re-armed';
END $$;

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES (
    'pipeline',
    'build',
    E'## spec_data passed as a map is never read — four producers filed empty specs (migration 434)\n\n'
    'create_work_item''s spec_data key is a PATH STRING to a map; four live steps passed an inline '
    '{key: path} map instead, which no extraction strategy resolves, so every item they filed carried '
    'spec = ''{}''. All four now use spec_paths (per-key paths, unresolved = hard error) + spec_literal '
    'for constants: tool-auditor''s create_improve_item and create_review_item, internal-linker''s '
    'create_rewrite_item, component-quality-auditor''s create_work_item. The four actionable webdesign '
    'audit_fix improve_tool items were backfilled from their row columns and (except the ab-test one, '
    'held for re-assessment) re-armed. LLM-optional finding fields (category/confidence/fix_suggestion) '
    'are deliberately no longer copied into specs — no live consumer reads them and spec_paths '
    'hard-errors on absence.',
    '["build-pipeline", "tool-auditor", "internal-linker", "component-quality-auditor"]'::jsonb,
    'migration',
    '434_spec_data_map_is_never_read_four_producers_file_empty_specs.sql'
);

INSERT INTO schema_migrations (filename)
VALUES ('434_spec_data_map_is_never_read_four_producers_file_empty_specs.sql');

COMMIT;
