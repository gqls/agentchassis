-- 148_supersede_quiz_plan_real_selectors.sql — the vonc quiz PLAN surrenders to
-- delivered reality (Option B, the 143 precedent). DB-only. doc_plans supersede
-- (never edit history in place); the retired row IS the rollback.
--
-- WHY. Tier-4's first run against tool-archetype-taster-quiz (vonc.com,
-- 2026-07-14) failed 3 of 7 checks. Two of those failures were NOT the tool's:
--
--   1. `boots` asserted `.tool-container` — the GENERATOR convention. This tool
--      ships via the page-section path and delivers
--      `.tool-archetype-taster-quiz-section`; `.tool-container` occurs ZERO
--      times on the live page. A composer-invented anchor, third sighting of
--      the class 143 was written for.
--   2. `asset` (asset_loads /tools/assets/<fn>.js) — the check Option B deleted
--      in 143/144. This PLAN predates 144, so it was born with the stale check.
--      The adapter skips it honestly ("not implemented"), never a fake fail, but
--      it has no business in a PLAN that describes what we DELIVER (inline JS).
--
-- The one genuine failure (mobile horizontal overflow) is NOT fixed here and is
-- NOT the tool's either: the offender is `div.footer-legal` (506px at a 390px
-- viewport) inside vonc's SITE FOOTER — it overflows every page on the site,
-- homepage included. It is routed to component-template-fixer as a
-- responsive_fix by the attribution logic in the judge, not to tool-improver.
--
-- ALSO: `quiz-flow-EDIT` was a placeholder that has never been filled in, so the
-- quiz's actual behaviour was untested. Its selectors are replaced with ones
-- VERIFIED against the live page in real Chromium first (clicking
-- `.quiz-option-btn` three times walks q1→q2→q3 and populates
-- `.result-archetype-name`) — a criterion is never authored unless it has been
-- watched passing. Dropping the -EDIT suffix makes Tier 4 evaluate it instead of
-- skipping it.
--
-- `container` is new (schema v0 addition): it names the tool's root so the judge
-- can tell a TOOL defect from SITE CHROME. Without it the adapter falls back to
-- a convention selector covering both delivery paths.

BEGIN;

DO $$
DECLARE
    fn        text := 'tool-archetype-taster-quiz';
    site      uuid := '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74';  -- vonc.com
    old_id    uuid;
    old_body  text;
    new_body  text;
    rest      text;
    s         int;
    e         int;
    -- Dollar-quoted: preserves backslashes (\\S) and newlines verbatim.
    new_fence text := $fence$```criteria
{ "profiles": ["desktop","mobile"],
  "container": ".tool-archetype-taster-quiz-section",
  "checks": [
    {"id":"boots","type":"selector_exists","selector":".tool-archetype-taster-quiz-section"},
    {"id":"console","type":"no_console_errors"},
    {"id":"status","type":"page_status_ok"},
    {"id":"mobile-fit","type":"no_horizontal_overflow","profiles":["mobile"]},
    {"id":"quiz-flow","type":"interaction",
      "steps":[{"action":"click","selector":".quiz-option-btn"},
               {"action":"click","selector":".quiz-option-btn"},
               {"action":"click","selector":".quiz-option-btn"}],
      "expect":{"selector":".result-archetype-name","text_matches":"\\S"}}
  ] }
```$fence$;
BEGIN
    SELECT id, body INTO old_id, old_body FROM doc_plans
    WHERE subject_type='tool' AND subject_key=fn AND is_current;
    IF old_id IS NULL THEN
        RAISE EXCEPTION '148: no current PLAN for %', fn;
    END IF;

    -- Guards: we are superseding the version we actually diagnosed. If any of
    -- these has moved, stop and re-inspect rather than clobber someone's edit.
    IF strpos(old_body, '"selector": ".tool-container"') = 0
       AND strpos(old_body, '"selector":".tool-container"') = 0 THEN
        RAISE EXCEPTION '148: the .tool-container anchor is not in the % PLAN — re-inspect before superseding', fn;
    END IF;
    IF strpos(old_body, 'asset_loads') = 0 THEN
        RAISE EXCEPTION '148: the asset_loads check is not in the % PLAN — re-inspect', fn;
    END IF;
    IF strpos(old_body, 'quiz-flow-EDIT') = 0 THEN
        RAISE EXCEPTION '148: the quiz-flow-EDIT placeholder is not in the % PLAN — re-inspect', fn;
    END IF;

    -- Replace the WHOLE criteria fence (surgery on individual lines is what
    -- produced half of these defects).
    s := strpos(old_body, '```criteria');
    IF s = 0 THEN
        RAISE EXCEPTION '148: no ```criteria fence in the % PLAN', fn;
    END IF;
    rest := substr(old_body, s);
    e := strpos(substr(rest, 12), chr(10) || '```');   -- closing fence, past the opener
    IF e = 0 THEN
        RAISE EXCEPTION '148: unterminated criteria fence in the % PLAN', fn;
    END IF;
    e := e + 11;                                       -- offset back into rest
    new_body := substr(old_body, 1, s - 1) || new_fence || substr(rest, e + 4);

    UPDATE doc_plans SET is_current = false, superseded_at = now() WHERE id = old_id;
    INSERT INTO doc_plans (subject_type, subject_key, body, source, created_by)
    VALUES ('tool', fn, new_body, 'human', '148_supersede_quiz_plan_real_selectors');

    INSERT INTO doc_notes (subject_type, subject_key, site_id, body, categories, source, created_by)
    VALUES ('tool', fn, site,
        '## PLAN superseded: real selectors + attribution — ' || fn || E'\n' ||
        'Observed: Tier-4''s first behavioural run (2026-07-14) failed boots on both profiles — the PLAN asserted .tool-container, but this tool ships via the page-section path and delivers .tool-archetype-taster-quiz-section (.tool-container occurs ZERO times on the live page). The PLAN also carried the asset_loads check that 143/144 retired, and quiz-flow-EDIT was never filled in, so the quiz flow itself was untested.' || E'\n' ||
        'Root cause: composer-invented anchor (third sighting of the class the anchor rule exists for) + a PLAN born before 144 fixed the composer.' || E'\n' ||
        'Fix: criteria fence superseded — boots anchored to the REAL section class; asset_loads removed (criteria describe what we DELIVER); quiz-flow-EDIT replaced with a real interaction (click .quiz-option-btn x3 -> .result-archetype-name) VERIFIED passing in real Chromium against the live page before being written here; container added so the judge can tell a tool defect from site chrome.' || E'\n' ||
        'Verified: live page probed 2026-07-14 (boots 1 match; quiz-flow passes; console clean).' || E'\n' ||
        'NOT fixed here: the page overflows horizontally on mobile — the offender is div.footer-legal (506px at 390px) in vonc''s SITE FOOTER, which overflows every page on the site (homepage included). That is site chrome, not this tool: it is routed to component-template-fixer as a responsive_fix. Blaming the tool would have sent the fixer to edit a component that cannot reach the footer.' || E'\n' ||
        'Categories: migration',
        '["migration"]'::jsonb, 'human', '148_supersede_quiz_plan_real_selectors');

    RAISE NOTICE '148: % superseded (old %, new body % chars)', fn, old_id, length(new_body);
END $$;

-- The parked ticket: record WHY it was cancelled, so nobody resurrects it.
UPDATE site_work_items
SET result = jsonb_build_object('resolution',
      'Cancelled 2026-07-14 before dispatch: the item bundled a FALSE boots failure (stale .tool-container anchor, fixed by PLAN supersede 148) with a GENUINE mobile-overflow failure that is NOT the tool''s (vonc site footer, div.footer-legal — routed to component-template-fixer as a responsive_fix). Dispatching it would have sent tool-improver chasing a stale contract.'),
    updated_at = now()
WHERE item_key = 'acceptance_fail:tool-archetype-taster-quiz:9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
  AND status = 'cancelled';

COMMIT;
