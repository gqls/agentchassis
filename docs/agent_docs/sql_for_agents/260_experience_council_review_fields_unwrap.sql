-- ============================================================================
-- 260_experience_council_review_fields_unwrap.sql
--
-- Migration 259 seated five reviewers and then discarded every one of them.
--
-- WHAT HAPPENED. `execute_llm_prompt` writes its output WRAPPED:
--   {"type": "json", "result": { ...the reviewer's actual JSON... }}
-- `diagnose_council_decide` reads each name in `review_fields` literally, so it
-- must be pointed at `review_x.result`, not `review_x`. The live code council
-- has always done this. I wrote the bare names, and the first run came back:
--
--   "no reviewer produced a readable opinion (0 abstained, 5 unreadable: ...)
--    — a council with no opinions cannot decide"
--
-- Five seats ran, five LLM calls were paid for, five sound opinions were
-- produced, and all five were thrown away. The decide action's refusal to
-- proceed is the only reason this surfaced as a failure rather than as an
-- approval nobody had reviewed.
--
-- WHY THE 259 GUARD DID NOT CATCH IT — the part worth keeping. That guard
-- asserted every seat's `output_field` appears in `review_fields`. Both sides of
-- that comparison were MINE, so it verified my list against my own list and
-- passed while the contract with the CONSUMER was broken. A guard that checks a
-- thing against itself is self-consistent by construction; the same shape as a
-- test fixture written to match the code it tests.
--
-- So the guard below checks the shape the CONSUMER requires — each field must be
-- exactly `<output_field>.result` — and additionally compares against the live
-- code council, which is the only independent witness available in SQL.
-- ============================================================================

BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,council_decide,config,review_fields}',
      '["review_observable_outcome.result",
        "review_honesty.result",
        "review_checkability.result",
        "review_deferral_honesty.result",
        "review_prior_art.result"]'::jsonb),
    updated_at = now()
WHERE type = 'experience-approval-council'
  AND COALESCE(is_snapshot, false) = false
  AND deleted_at IS NULL;

DO $guard$
DECLARE
    steps jsonb; fields jsonb; f text; seat text;
    code_council_ok boolean;
BEGIN
    SELECT default_config->'workflow'->'steps' INTO steps FROM agent_definitions
     WHERE type = 'experience-approval-council'
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF steps IS NULL THEN
        RAISE EXCEPTION '260: the experience-approval-council definition is missing';
    END IF;
    fields := steps->'council_decide'->'config'->'review_fields';

    IF jsonb_array_length(fields) <> 5 THEN
        RAISE EXCEPTION '260: expected 5 review_fields, got %', jsonb_array_length(fields);
    END IF;

    -- The consumer's contract: an unwrapped path, and a real step behind it.
    FOR f IN SELECT jsonb_array_elements_text(fields) LOOP
        IF f NOT LIKE '%.result' THEN
            RAISE EXCEPTION '260: review_field % does not end in .result — execute_llm_prompt wraps its output, so the seat would be read as UNREADABLE and silently discarded', f;
        END IF;
        seat := left(f, length(f) - length('.result'));
        IF steps->seat IS NULL THEN
            RAISE EXCEPTION '260: review_field % names step %, which does not exist', f, seat;
        END IF;
        IF steps->seat->>'output_field' <> seat THEN
            RAISE EXCEPTION '260: step % writes output_field %, not % — the field read and the field written must be the same', seat, steps->seat->>'output_field', seat;
        END IF;
    END LOOP;

    -- Independent witness: the live code council has been decided this way for
    -- weeks. If it does NOT use the .result form, my premise is wrong and this
    -- migration should not be trusted.
    SELECT bool_and(v LIKE '%.result') INTO code_council_ok
      FROM agent_definitions a,
           jsonb_array_elements_text(a.default_config->'workflow'->'steps'->'council_decide'->'config'->'review_fields') v
     WHERE a.type = 'council-gate' AND COALESCE(a.is_snapshot,false) = false AND a.deleted_at IS NULL;
    IF code_council_ok IS DISTINCT FROM true THEN
        RAISE EXCEPTION '260: the live council-gate does NOT use the .result form — the premise of this fix is wrong, stop and re-read diagnose_council_decide';
    END IF;
END
$guard$;

COMMIT;

-- Verify
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'council_decide'->'config'->'review_fields')
FROM agent_definitions
WHERE type = 'experience-approval-council'
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
