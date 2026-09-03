-- 749 ROLLBACK — remove the computed_values vocabulary from tool-generator's PLAN prompt
--
-- Reverses both halves of 749 by swapping back the two VERBATIM literals it
-- inserted. It deliberately does NOT use a regexp: the first cut of this file
-- did, and failed on both counts a regexp can fail here — the block spans
-- newlines, and its tail contained an apostrophe the pattern had dropped. It
-- reported UPDATE 1 while substituting nothing, and only 749's post-verify DO
-- block caught it. The literals below are EXTRACTED FROM 749 ITSELF, so they are
-- byte-identical by construction rather than by transcription.
--
-- ⚠ REGENERATE THIS FILE, never hand-edit it, whenever 749's prompt text changes:
--     anchor = re.search(r"\$anchor\$(.*?)\$anchor\$", src, re.S).group(1)
--     new    = re.search(r"\$new\$(.*?)\$new\$", src, re.S).group(1)
--     assert new.startswith(anchor)
--
-- NOTE ON PROVENANCE: since round 2, 749 also calls
-- snapshot_agent('tool-generator', ...) before mutating, so a queryable pre-image
-- row exists independently of this file. This remains the targeted reversal;
-- the snapshot is the general one. Prefer sharpening the prompt to reverting it —
-- with this reverted the generator is blind again by construction (bugs_open/449).
--
-- WHEN TO USE THIS. If generated fences start carrying value assertions read off
-- the tool's own output rather than derived. Symptom to look for BEFORE
-- reverting: a new fence whose expect_values match the tool exactly while
-- ## Dependencies names no rule and shows no working.

BEGIN;

DO $guard$
DECLARE
    n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
     WHERE type = 'tool-generator' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND default_config #>> '{workflow,steps,compose_plan,config,prompt_template}'
           LIKE '%"type":"computed_values"%';

    IF n = 0 THEN
        RAISE NOTICE '749 ROLLBACK: compose_plan does not teach computed_values, nothing to do';
        RETURN;
    END IF;
    IF n <> 1 THEN
        RAISE EXCEPTION '749 ROLLBACK ABORT: expected exactly 1 live tool-generator carrying the 749 text, found %.', n;
    END IF;

    PERFORM snapshot_agent('tool-generator', '749_..._ROLLBACK.sql: pre-revert');
END
$guard$;

-- the vocabulary block, back out — exact literal swap, the reverse of 749's
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,compose_plan,config,prompt_template}',
        to_jsonb(replace(
            default_config #>> '{workflow,steps,compose_plan,config,prompt_template}',
            $new$NEVER invent a selector — if unsure, omit the interaction check entirely. The JSON must be valid; ids lowercase-kebab.
ONE further check type exists and it is NOT an interaction, so the sentence above does not forbid it: "computed_values", which compares the EXACT TEXT a result element reads after inputs have been filled. It is the only check that judges what the tool COMPUTES; every check above passes a calculator that prints a confidently wrong number. Shape, exactly:
{"id":"<kebab-id>","type":"computed_values","profiles":["desktop"],"steps":[{"action":"fill","selector":"#realInput","value":"5000"}],"expect_values":{"#realResult":"$2,000.00"}}
"expect_values" maps a selector to the exact text it must read once "steps" have run. Whitespace is collapsed on both sides; everything else must match as the page renders it, including currency symbols, thousands separators and decimal places. Step actions are "fill" (with "value"), "click" and "select" (with "value").
EMIT IT ONLY IF YOU CAN DO THE ARITHMETIC YOURSELF WITHOUT READING ANY NUMBER OFF THE HTML ABOVE. That is possible only when the tool implements a rule that is published and checkable independently of this code: a standard financial formula, a statutory or tax rate, a unit conversion, a published index, or arithmetic that follows from the spec (volume x margin per unit). Choose the input values yourself, work each expected output out from that rule, and state the rule and your working in "## Dependencies" so a reviewer can check the expectation without trusting the tool.
DERIVE THE VALUE FROM THE RULE; READ THE FORMAT OFF THE CODE. These are two different acts and only the first must be independent. The number itself must come from the published rule, but how the page renders it — currency symbol, thousand separators, decimal places, any suffix — you must copy from the tool's own formatting code above, because the comparison collapses whitespace and is otherwise exact. A correct answer written in the wrong format fails as loudly as a wrong one.
WRITE THE INPUTS AS LITERAL VALUES in "steps" — never "the page default", never a reference for the checker to resolve later. The fence is run against the DEPLOYED page, not the draft markup you are reading, and a reference resolves differently on the two sides; the input vector is part of the expectation's identity, so it must travel with it.
OTHERWISE OMIT THE CHECK ENTIRELY and write one line in "## Dependencies" beginning "No value assertion:" and naming what is missing. That line is the honest label for "we could only establish that the tool responds", and it is the correct answer far more often than not. Anything that scores, rates, ranks, grades or classifies by a heuristic invented for this tool has NO independent source: the only "expected" value available is whatever this code happens to print, and pinning that makes today's bug tomorrow's specification. A GUESSED EXPECTATION IS WORSE THAN NONE — omitting the check is the correct answer far more often than not, and it is not a failure to report.
WHEN you do include a computed_values check, add these two keys at the TOP LEVEL of the criteria object, beside "profiles" and "container": "no_auto_fix": true, and "no_auto_fix_reason": "arithmetic assertion — a failure means the formula or the law moved, which is a human's decision and not a rewriter's".$new$,
            $anchor$NEVER invent a selector — if unsure, omit the interaction check entirely. The JSON must be valid; ids lowercase-kebab.$anchor$
        )),
        false)
WHERE type = 'tool-generator' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,compose_plan,config,prompt_template}'
      LIKE '%"type":"computed_values"%';

-- the budget, back to 3000
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,compose_plan,config,prompt_template}',
        to_jsonb(replace(
            default_config #>> '{workflow,steps,compose_plan,config,prompt_template}',
            $newcap$Keep the whole document under 3600 characters.$newcap$,
            $cap$Keep the whole document under 3000 characters.$cap$
        )),
        false)
WHERE type = 'tool-generator' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,compose_plan,config,prompt_template}'
      LIKE '%Keep the whole document under 3600 characters.%';

DO $verify$
DECLARE p text;
BEGIN
    SELECT default_config #>> '{workflow,steps,compose_plan,config,prompt_template}'
      INTO p FROM agent_definitions
     WHERE type = 'tool-generator' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF p LIKE '%computed_values%' OR p LIKE '%expect_values%' THEN
        RAISE EXCEPTION '749 ROLLBACK ABORT: the vocabulary block is still present — the literal did not match, so 749 has been edited since this file was generated. REGENERATE this file from 749 (see header). Nothing committed.';
    END IF;
    IF p NOT LIKE '%under 3000 characters%' THEN
        RAISE EXCEPTION '749 ROLLBACK ABORT: document cap not restored to 3000.';
    END IF;
    IF p NOT LIKE '%NEVER invent a selector%' THEN
        RAISE EXCEPTION '749 ROLLBACK ABORT: the original anchor sentence did not survive the removal — the prompt is now missing text 749 did not add. Nothing committed.';
    END IF;
    RAISE NOTICE '749 ROLLBACK OK: compose_plan is back to the liveness-only vocabulary and a 3000-character cap';
END
$verify$;

COMMIT;
