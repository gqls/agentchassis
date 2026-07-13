-- 144_composer_inline_delivery.sql — future PLANs are born honest: the
-- compose_plan prompt stops asserting the never-built asset extraction.
-- Companion to 143 (which superseded the two existing PLANs). DB-only.
--
-- Two verbatim edits to tool-generator's compose_plan prompt_template:
--   1. the five standard checks lose the asset_loads entry (five → four);
--   2. the Delivery mechanism instruction changes from "Path 1 — extracted to
--      /tools/assets/<fn>.js on rerender" to inline reality.
-- Everything else (never-invent-selectors rule, size cap, section order)
-- untouched. Anchored on exact substrings with guards — abort if drifted.

BEGIN;

SELECT snapshot_agent('tool-generator', '144_composer_inline_delivery.sql: pre-update');

DO $$
DECLARE
    tmpl     text;
    old_five text := 'Always include these five checks verbatim: {"id":"boots","type":"selector_exists","selector":".tool-container"}, {"id":"console","type":"no_console_errors"}, {"id":"asset","type":"asset_loads","path":"/tools/assets/{{.input_data.spec.function}}.js"}, {"id":"status","type":"page_status_ok"}';
    new_four text := 'Always include these four checks verbatim: {"id":"boots","type":"selector_exists","selector":".tool-container"}, {"id":"console","type":"no_console_errors"}, {"id":"status","type":"page_status_ok"}';
    old_del  text := 'One line: Path 1 — component inline script, extracted to /tools/assets/{{.input_data.spec.function}}.js on rerender.';
    new_del  text := 'One line: inline — all JS and CSS ship inside the page HTML (no asset extraction; decision 2026-07-10).';
BEGIN
    SELECT default_config #>> '{workflow,steps,compose_plan,config,prompt_template}' INTO tmpl
    FROM agent_definitions WHERE type='tool-generator' AND deleted_at IS NULL;

    IF strpos(tmpl, old_five) = 0 THEN
        RAISE EXCEPTION '144: five-checks sentence not found verbatim — prompt drifted, inspect before applying';
    END IF;
    IF strpos(tmpl, old_del) = 0 THEN
        RAISE EXCEPTION '144: delivery instruction not found verbatim — prompt drifted, inspect before applying';
    END IF;

    tmpl := replace(tmpl, old_five, new_four);
    tmpl := replace(tmpl, old_del, new_del);

    UPDATE agent_definitions
    SET default_config = jsonb_set(default_config,
          '{workflow,steps,compose_plan,config,prompt_template}', to_jsonb(tmpl))
    WHERE type='tool-generator' AND deleted_at IS NULL;

    RAISE NOTICE '144: compose_plan prompt updated (% chars)', length(tmpl);
END $$;

-- Pipeline note (runbook §3).
INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('pipeline', 'build',
$note$## 144: composer stops asserting asset extraction (Option B, with 143)
Observed: every generated PLAN carried an asset_loads check for /tools/assets/<fn>.js and declared Path 1 extraction; the render path ships JS inline and the extraction was never built — Tier-2 failed every tool on `asset` by construction.
Root cause: the compose_plan prompt encoded an aspiration as an acceptance criterion.
Fix: standard checks five → four (asset_loads removed); Delivery mechanism instruction now states inline reality. Existing PLANs superseded by 143.
Verified: guard asserts both substrings replaced; next tool creation writes an asset-free PLAN.
Categories: migration$note$,
'["migration"]'::jsonb, 'human', 'run-migrations.sh');

DO $$
DECLARE tmpl text;
BEGIN
    SELECT default_config #>> '{workflow,steps,compose_plan,config,prompt_template}' INTO tmpl
    FROM agent_definitions WHERE type='tool-generator' AND deleted_at IS NULL;
    IF strpos(tmpl, 'asset_loads') > 0 THEN
        RAISE EXCEPTION '144: asset_loads still present in composer prompt';
    END IF;
    IF strpos(tmpl, 'these four checks') = 0 OR strpos(tmpl, 'inline — all JS and CSS ship inside the page HTML') = 0 THEN
        RAISE EXCEPTION '144: replacement text missing after update';
    END IF;
END $$;

COMMIT;

-- Verify: next organic tool creation writes a PLAN whose fence has four
-- standard checks and whose Delivery mechanism says inline.
-- Rollback: restore the snapshot taken at the top.
