-- 749 — teach tool-generator's PLAN prompt the computed_values check, and when to REFUSE to write one
--
-- bugs_open/449 (§5 candidates 1+2, lane
-- docs/agent_docs/docs024_key_docs_latest/bugfix_449_fences_assert_no_number/).
--
-- THE DEFECT. `compose_plan`'s fence vocabulary is a CLOSED enumeration: four
-- liveness checks (selector_exists, no_console_errors, page_status_ok,
-- no_horizontal_overflow), one optional `interaction`, and the sentence "No
-- other check type exists for interactions". `computed_values` — the only check
-- type that judges what a tool COMPUTES rather than what it contains — is never
-- named anywhere in the row, so it is never a candidate.
--
-- MEASURED 2026-09-03 16:04Z, doc_plans WHERE subject_type='tool' AND is_current:
--   tool-generator          187 fences, uses_computed_values = 0, 91 drive inputs,
--                           55 DRIVE INPUTS AND ASSERT NOTHING
--   all other authors        54 fences, 30 with a value assertion
-- i.e. every value-asserting fence in the estate was written by an operator or a
-- lane by hand; the agent has authored ZERO in 187 attempts since 2026-07-14, and
-- max(created_at) is the same day this was measured, so it is a live intake.
-- The runner has supported the type all along: it is whitelisted at
-- run_checks_action.go:568 and dispatched at :708 — this is an authoring gap only.
--
-- WHY THE INSTRUCTION IS CONDITIONAL, AND WHY ITS DEFAULT IS REFUSAL.
-- `computed_values` is a REGRESSION check by design. Its own contract says so
-- (run_checks_action.go:790-808): values "are CAPTURED from the tool while it is
-- known good ... and then defended", and "a golden captured from an already-wrong
-- tool pins the wrong answer — the capture script therefore refuses to emit for a
-- tool whose outputs do not react to its inputs". At BIRTH nothing is known good,
-- and this step's only inputs are the spec and {{.generated_html}} — the very
-- artefact whose correctness is in question. So an unqualified "emit
-- computed_values" could only ever pin whatever the new code prints, which is
-- bugs_open/224 / bugs_closed/225: an expired £625k FTB SDLT cap certified green
-- for sixteen months. The one independent oracle actually present is the model's
-- own knowledge of a PUBLISHED formula, applied to inputs it chooses — so the
-- prompt licenses exactly that case and requires refusal otherwise, with the
-- refusal recorded in ## Dependencies where a reviewer can see it.
--
-- WHY THE CHARACTER CAP MOVES 3000 -> 3600. A computed_values check in the shape
-- taught below costs ~350-450 characters of the output document, plus the
-- Dependencies line and the two top-level keys. Left at 3000 the model must trade
-- the new assertion against prose it was also told to write, which is how a fix
-- like this ships as a truncated document instead. compose_plan runs
-- claude-sonnet-5 with max_tokens 4000; 3600 characters is ~900-1200 tokens, so
-- there is ample headroom and no truncation risk.
--
-- SCOPE. tool-generator ONLY. `experience-planner` carries a criteria fence in
-- THREE steps (compose, recompose, reframe) and has authored 3 fences EVER
-- (measured same run) — three more verbatim anchors for ~1% of the population, so
-- it is a deliberate follow-on, not a widening of this file. Component-subject
-- fences (55, zero value assertions) are recorded as a new candidate in the bug's
-- §9f and are not touched here.
--
-- COMPOSES WITH 732: that migration anchors on {…,generate_tool_html,…} and
-- {…,improve_tool,…}; this one on {…,compose_plan,…}. Different JSON paths in the
-- same row, so the two jsonb_sets apply in either order. The row is not the unit,
-- the path is.
--
-- SURGICAL: two verbatim anchors, each verified to match exactly 1 live row at
-- 2026-09-03 16:5xZ. ABORTS rather than writing a prompt it has not read.
-- Idempotent: re-running once the new text is present is a no-op.

BEGIN;

-- ---------------------------------------------------------------- guard: pre
DO $guard$
DECLARE
    n_vocab int;
    n_cap   int;
    n_known int;
BEGIN
    SELECT count(*) INTO n_known FROM agent_definitions
     WHERE type = 'tool-generator' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND default_config #>> '{workflow,steps,compose_plan,config,prompt_template}'
           LIKE '%"type":"computed_values"%';

    IF n_known > 0 THEN
        RAISE NOTICE '749: already applied (compose_plan already teaches computed_values), nothing to do';
        RETURN;
    END IF;

    SELECT count(*) INTO n_vocab FROM agent_definitions
     WHERE type = 'tool-generator' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND default_config #>> '{workflow,steps,compose_plan,config,prompt_template}'
           LIKE '%NEVER invent a selector — if unsure, omit the interaction check entirely. The JSON must be valid; ids lowercase-kebab.%';

    SELECT count(*) INTO n_cap FROM agent_definitions
     WHERE type = 'tool-generator' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND default_config #>> '{workflow,steps,compose_plan,config,prompt_template}'
           LIKE '%Keep the whole document under 3000 characters.%';

    IF n_vocab <> 1 THEN
        RAISE EXCEPTION '749 ABORT: expected exactly 1 live tool-generator whose compose_plan prompt carries the criteria-vocabulary anchor, found %. The prompt has been edited since 2026-09-03; RE-READ it and re-anchor rather than overwriting a prompt this migration has not seen. Note the anchor contains an em-dash and a semicolon and must match byte-for-byte.', n_vocab;
    END IF;
    IF n_cap <> 1 THEN
        RAISE EXCEPTION '749 ABORT: expected exactly 1 live tool-generator carrying the "under 3000 characters" anchor, found %. Do not apply the vocabulary half without the budget half: the new check costs ~400 characters of a document capped at 3000, so applying one without the other trades a value assertion for a truncated PLAN.', n_cap;
    END IF;
END
$guard$;

-- ------------------------------------------- the vocabulary gains ONE type...
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,compose_plan,config,prompt_template}',
        to_jsonb(replace(
            default_config #>> '{workflow,steps,compose_plan,config,prompt_template}',
            $anchor$NEVER invent a selector — if unsure, omit the interaction check entirely. The JSON must be valid; ids lowercase-kebab.$anchor$,
            $new$NEVER invent a selector — if unsure, omit the interaction check entirely. The JSON must be valid; ids lowercase-kebab.
ONE further check type exists and it is NOT an interaction, so the sentence above does not forbid it: "computed_values", which compares the EXACT TEXT a result element reads after inputs have been filled. It is the only check that judges what the tool COMPUTES; every check above passes a calculator that prints a confidently wrong number. Shape, exactly:
{"id":"<kebab-id>","type":"computed_values","profiles":["desktop"],"steps":[{"action":"fill","selector":"#realInput","value":"5000"}],"expect_values":{"#realResult":"$2,000.00"}}
"expect_values" maps a selector to the exact text it must read once "steps" have run. Whitespace is collapsed on both sides; everything else must match as the page renders it, including currency symbols, thousands separators and decimal places. Step actions are "fill" (with "value"), "click" and "select" (with "value").
EMIT IT ONLY IF YOU CAN DO THE ARITHMETIC YOURSELF WITHOUT READING ANY NUMBER OFF THE HTML ABOVE. That is possible only when the tool implements a rule that is published and checkable independently of this code: a standard financial formula, a statutory or tax rate, a unit conversion, a published index, or arithmetic that follows from the spec (volume x margin per unit). Choose the input values yourself, work each expected output out from that rule, and state the rule and your working in "## Dependencies" so a reviewer can check the expectation without trusting the tool.
OTHERWISE OMIT THE CHECK ENTIRELY and write one line in "## Dependencies" beginning "No value assertion:" and naming what is missing. Anything that scores, rates, ranks, grades or classifies by a heuristic invented for this tool has NO independent source: the only "expected" value available is whatever this code happens to print, and pinning that makes today's bug tomorrow's specification. A GUESSED EXPECTATION IS WORSE THAN NONE — omitting the check is the correct answer far more often than not, and it is not a failure to report.
WHEN you do include a computed_values check, add these two keys at the TOP LEVEL of the criteria object, beside "profiles" and "container": "no_auto_fix": true, and "no_auto_fix_reason": "arithmetic assertion — a failure means the formula or the law moved, which is a human's decision and not a rewriter's".$new$
        )),
        false)
WHERE type = 'tool-generator' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,compose_plan,config,prompt_template}'
      NOT LIKE '%"type":"computed_values"%';

-- ...and the document budget grows to pay for it
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,compose_plan,config,prompt_template}',
        to_jsonb(replace(
            default_config #>> '{workflow,steps,compose_plan,config,prompt_template}',
            $cap$Keep the whole document under 3000 characters.$cap$,
            $newcap$Keep the whole document under 3600 characters.$newcap$
        )),
        false)
WHERE type = 'tool-generator' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,compose_plan,config,prompt_template}'
      LIKE '%Keep the whole document under 3000 characters.%';

-- --------------------------------------------------------------- guard: post
-- A DO block that RAISES, never a list of bare SELECTs: ON_ERROR_STOP ignores a
-- non-empty result set, so a verify made of SELECTs cannot stop the COMMIT.
DO $verify$
DECLARE
    p text;
BEGIN
    SELECT default_config #>> '{workflow,steps,compose_plan,config,prompt_template}'
      INTO p
      FROM agent_definitions
     WHERE type = 'tool-generator' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF p IS NULL THEN
        RAISE EXCEPTION '749 ABORT: no live tool-generator row after the update.';
    END IF;

    -- the type is now nameable
    IF p NOT LIKE '%"type":"computed_values"%' THEN
        RAISE EXCEPTION '749 ABORT: compose_plan does not name computed_values after the update; nothing committed.';
    END IF;
    -- ...with its map key, or the shape is unusable
    IF p NOT LIKE '%expect_values%' THEN
        RAISE EXCEPTION '749 ABORT: compose_plan names the type but not expect_values; the shape is incomplete.';
    END IF;
    -- ...and the REFUSAL arm, which is the load-bearing half. A prompt that
    -- teaches the type without licensing refusal is WORSE than the defect: it
    -- pins whatever the new tool prints.
    IF p NOT LIKE '%WORSE THAN NONE%' THEN
        RAISE EXCEPTION '749 ABORT: the refusal arm is absent. Do not ship the capability without it — a guessed expectation pins garbage as specification (bugs_open/224).';
    END IF;
    IF p NOT LIKE '%No value assertion:%' THEN
        RAISE EXCEPTION '749 ABORT: the Dependencies refusal line is absent, so a refusal would leave no record a reviewer can see.';
    END IF;
    -- ...and no_auto_fix, so a red arithmetic fence reaches a human
    IF p NOT LIKE '%no_auto_fix%' THEN
        RAISE EXCEPTION '749 ABORT: no_auto_fix guidance absent; a failing arithmetic fence would be routed to tool-improver.';
    END IF;
    -- ...and the budget half actually landed
    IF p NOT LIKE '%under 3600 characters%' THEN
        RAISE EXCEPTION '749 ABORT: the document cap is still 3000; the new check would be traded against prose. Nothing committed.';
    END IF;
    IF p LIKE '%under 3000 characters%' THEN
        RAISE EXCEPTION '749 ABORT: both caps present — the replace matched twice or the text drifted. Re-read the prompt.';
    END IF;

    RAISE NOTICE '749 OK: compose_plan teaches computed_values WITH its refusal arm; document cap 3600';
END
$verify$;

COMMIT;
