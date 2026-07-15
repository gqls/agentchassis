-- 149_composer_emits_container.sql — future PLANs are born ATTRIBUTABLE: the
-- compose_plan prompt now emits a top-level `container` naming the tool's root
-- element, and derives `boots` from the REAL root in the generated HTML instead
-- of asserting a class by convention. DB-only. Companion to the Tier-4
-- attribution work (adapter + judge, v1.0.1116).
--
-- WHY. `no_horizontal_overflow` measures the whole DOCUMENT, but an acceptance
-- run is scoped to ONE tool. Without knowing where the tool ENDS, the judge
-- cannot tell a tool defect from site chrome, and an overflowing site footer
-- raises an unfixable improve_tool ticket for every tool on the site. Live case
-- 2026-07-14: vonc.com's div.footer-legal (506px at a 390px viewport) failed the
-- archetype quiz's mobile-fit check, on every page of the site. `container` is
-- what lets the judge route that to component-template-fixer as a responsive_fix
-- instead of blaming the tool. The adapter falls back to a convention selector
-- when a PLAN omits `container` (i.e. every PLAN written before now), so this is
-- additive, not breaking.
--
-- AND: `boots` was hardcoded to `.tool-container`. That is TRUE for the tools
-- this generator writes — but stating it as a constant is how the same defect
-- keeps recurring (the quiz PLAN asserted `.tool-container` on a page-section
-- tool that ships `.tool-archetype-taster-quiz-section`, where `.tool-container`
-- occurs ZERO times — third sighting of the invented-anchor class). The prompt
-- now says: READ the root element out of the HTML you were given. If that HTML
-- uses `.tool-container`, the model still writes `.tool-container` — the same
-- output, honestly derived, and still correct the day the render path changes.
--
-- NOTE ON ANCHORING (144's lesson): prompt_template holds REAL newlines, so the
-- guards anchor on SINGLE-LINE substrings only; the newline for the new
-- container line is injected with chr(10).

BEGIN;

SELECT snapshot_agent('tool-generator', '149_composer_emits_container.sql: pre-update');

DO $$
DECLARE
    tmpl       text;
    old_profs  text := '{ "profiles": ["desktop","mobile"],';
    old_checks text := 'Always include these four checks verbatim: {"id":"boots","type":"selector_exists","selector":".tool-container"}, {"id":"console","type":"no_console_errors"}, {"id":"status","type":"page_status_ok"}, {"id":"mobile-fit","type":"no_horizontal_overflow","profiles":["mobile"]}.';
    new_checks text := 'The top-level "container" is the selector of the tool''s OUTERMOST element — COPY it from the generated HTML above, never assume it. It tells the behavioural checker where the tool ENDS, so a layout defect belonging to the site header or footer is not blamed on this tool. Always include these four checks, with boots anchored on that SAME root selector: {"id":"boots","type":"selector_exists","selector":"<the container selector>"}, {"id":"console","type":"no_console_errors"}, {"id":"status","type":"page_status_ok"}, {"id":"mobile-fit","type":"no_horizontal_overflow","profiles":["mobile"]}.';
    new_profs  text;
BEGIN
    new_profs := old_profs || chr(10) || '  "container": "<the tool''s outermost selector, copied from the HTML above>",';

    SELECT default_config #>> '{workflow,steps,compose_plan,config,prompt_template}' INTO tmpl
    FROM agent_definitions WHERE type='tool-generator' AND deleted_at IS NULL;

    IF tmpl IS NULL THEN
        RAISE EXCEPTION '149: no tool-generator compose_plan prompt found';
    END IF;
    IF strpos(tmpl, old_profs) = 0 THEN
        RAISE EXCEPTION '149: the criteria-fence example line was not found verbatim — prompt drifted, inspect before applying';
    END IF;
    IF strpos(tmpl, old_checks) = 0 THEN
        RAISE EXCEPTION '149: the four-checks sentence was not found verbatim (144 baseline expected) — prompt drifted, inspect before applying';
    END IF;
    IF strpos(tmpl, '"container"') > 0 THEN
        RAISE EXCEPTION '149: the prompt already mentions container — already applied?';
    END IF;

    tmpl := replace(tmpl, old_profs, new_profs);
    tmpl := replace(tmpl, old_checks, new_checks);

    UPDATE agent_definitions
    SET default_config = jsonb_set(default_config,
          '{workflow,steps,compose_plan,config,prompt_template}', to_jsonb(tmpl)),
        updated_at = now()
    WHERE type='tool-generator' AND deleted_at IS NULL;

    RAISE NOTICE '149: compose_plan prompt updated (% chars)', length(tmpl);
END $$;

-- Pipeline note (runbook §3).
INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('pipeline', 'build',
$note$## 149: composer emits `container` — PLANs are born attributable
Observed: Tier-4's no_horizontal_overflow measures the whole document, but an acceptance run is scoped to ONE tool. vonc.com's div.footer-legal (506px at a 390px viewport) overflowed EVERY page on the site and failed the archetype quiz's mobile-fit check — a site-chrome defect that would have raised an unfixable improve_tool ticket against the quiz, and against every other tool on the site, on every run.
Root cause: nothing in a PLAN said where the tool ENDS, so the judge could not tell a tool defect from site chrome. Separately, boots was hardcoded to .tool-container — true for generator tools, but stating it as a constant is how invented anchors keep recurring (third sighting; the quiz ships .tool-archetype-taster-quiz-section).
Fix: compose_plan now emits a top-level "container" (the tool's outermost selector, COPIED from the generated HTML) and anchors boots on that same root. The adapter+judge (v1.0.1116) use it to route chrome defects to component-template-fixer as responsive_fix items instead of blaming the tool; PLANs without container fall back to a convention selector covering both delivery paths.
Verified: guards assert the 144-baseline lines were replaced; post-check asserts the container instruction is present. The next tool creation writes a PLAN carrying container.
Categories: migration$note$,
'["migration"]'::jsonb, 'human', 'run-migrations.sh');

DO $$
DECLARE tmpl text;
BEGIN
    SELECT default_config #>> '{workflow,steps,compose_plan,config,prompt_template}' INTO tmpl
    FROM agent_definitions WHERE type='tool-generator' AND deleted_at IS NULL;
    IF strpos(tmpl, '"container"') = 0 THEN
        RAISE EXCEPTION '149: post-check failed — container instruction not present';
    END IF;
    IF strpos(tmpl, 'never assume it') = 0 THEN
        RAISE EXCEPTION '149: post-check failed — root-derivation instruction not present';
    END IF;
    RAISE NOTICE '149: post-check OK — composer emits container and anchors boots on the real root';
END $$;

COMMIT;
