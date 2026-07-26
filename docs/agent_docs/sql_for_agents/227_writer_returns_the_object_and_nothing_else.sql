-- 227_writer_returns_the_object_and_nothing_else.sql — bugs_open/088, the
-- prompt-side half (candidate A). DB-only; effective immediately, no image roll.
--
-- ── WHY ────────────────────────────────────────────────────────────────────
-- 2026-07-26 14:26Z, correlation d9fd6ed2: page-content-writer returned a complete
-- hero object, then this, out loud —
--
--     Wait — I must scan for em dashes before returning. Found one in the
--     headline. Rewriting now.
--
-- — and then a corrected object. Two complete objects and a paragraph between
-- them. json.Unmarshal rejects the document, the raw-text envelope swallows it,
-- and the required-field gate fails the whole page build at iteration 0 with
-- "likely LLM truncation", which it is not. A model-directory build died on it.
--
-- The prompt invited both halves of that:
--
--   * the Voice & Style block says "Before returning, scan your draft for the —
--     character". An instruction to self-check "before returning", in a prompt
--     whose output contract is not fenced, is an invitation to emit the check.
--     The model did exactly as told — and its correction line contains an em dash.
--   * the Output Format block says "Return a JSON object with exactly these keys"
--     and never says ONLY that object. Nothing forbade the commentary.
--
-- ── SCOPE: WHY THIS AGENT AND NOT THE FLEET ────────────────────────────────
-- Measured, not assumed. 5,844 stored llm_call_log responses (2026-03-25 →
-- 07-26) that carry a '{' and were not clean single objects; today's parser
-- rejects 647 of them. Of the 34 that are COMPLETE answers buried in commentary
-- (the class this prompt change addresses), 33 are page-content-writer and 1 is
-- component-creator. The remaining 613 hold no complete value at all — they are
-- truncation, which is bugs_open/076's problem and not this one.
--
-- Deliberately NOT patched:
--   * experience-planner / tool-generator / generic — their steps ASK for markdown
--     containing a fenced JSON block ("Output the whole plan as markdown … the
--     ```criteria fence … <!-- END EXPERIENCE_PLAN -->"). Their responses are not
--     malformed; the whole document is the answer. Fencing their output format
--     would break them.
--   * content-writer — already says "Return ONLY valid JSON", and contributed 0
--     of the 34.
--   * component-creator — 1 case, and its 177 failures are overwhelmingly
--     truncation. Not worth a prompt edit on this evidence.
--
-- ── THE GO HALF ────────────────────────────────────────────────────────────
-- This migration stops the writer emitting commentary. It cannot repair a
-- response that arrives with it anyway, so the platform-side recovery lands
-- separately in json_envelope.go (ParseLLMJSONWithProvenance tier 3) — INERT until
-- the next chassis roll. The two are independent: either alone reduces the
-- failure, both together close it.
--
-- Anchors verified unique against the LIVE row 2026-07-26 (each: exactly 1
-- occurrence; "and nothing else" absent, so this has not already been applied).
-- Standing rule: snapshot_agent opens the transaction; the DO block fails the
-- whole transaction if any anchor did not match.

BEGIN;

SELECT snapshot_agent('page-content-writer', '227_writer_returns_the_object_and_nothing_else.sql: pre-update');

-- (1) Fence the output contract: the object, and nothing else.
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
      to_jsonb(replace(
        default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
        '## Output Format
Return a JSON object with exactly these keys:',
        $of$## Output Format
Your entire reply must be the JSON object and nothing else. No preamble before it, no explanation after it, no second copy of it, no markdown fence around it. If a field has no honest value, give it an empty string — do not explain the omission in prose. If you want to revise a draft, revise it before you write the JSON out: a correction written after the object ("Wait, let me redo that…") makes the whole reply unreadable to the renderer, and the page build fails.

Return a JSON object with exactly these keys:$of$
      ))
    ),
    updated_at = now()
WHERE type = 'page-content-writer' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

-- (2) Make the em-dash self-check silent. Same rule, same strictness — it just
--     must not be narrated into the response.
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
      to_jsonb(replace(
        default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
        'Before returning, scan your draft for the',
        $em$Do this check silently as you compose, never in the reply itself: scan your draft for the$em$
      ))
    ),
    updated_at = now()
WHERE type = 'page-content-writer' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

-- Verify: one live row, both new texts present exactly once, both old anchors gone.
DO $$
DECLARE pt text; n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
      WHERE type='page-content-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF n <> 1 THEN RAISE EXCEPTION 'page-content-writer: expected exactly one live row, found %', n; END IF;

    SELECT default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
      INTO pt FROM agent_definitions
      WHERE type='page-content-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    IF (length(pt)-length(replace(pt,'must be the JSON object and nothing else','')))/length('must be the JSON object and nothing else') <> 1 THEN
        RAISE EXCEPTION '227: output-format fence not present exactly once';
    END IF;
    IF (length(pt)-length(replace(pt,'Do this check silently as you compose','')))/length('Do this check silently as you compose') <> 1 THEN
        RAISE EXCEPTION '227: silent em-dash check not present exactly once';
    END IF;
    IF position('Before returning, scan your draft' in pt) <> 0 THEN
        RAISE EXCEPTION '227: the old "Before returning" phrasing is still present';
    END IF;

    -- The rules this change must NOT have disturbed: migration 201's
    -- anti-fabrication rule 14, and the field list the writer works from.
    IF position('is NOT permission to invent one' in pt) = 0 THEN
        RAISE EXCEPTION '227: migration 201 rule 14 is missing — this edit damaged it';
    END IF;
    IF position('Return a JSON object with exactly these keys' in pt) = 0 THEN
        RAISE EXCEPTION '227: the field-list contract went missing';
    END IF;

    RAISE NOTICE '227 applied: page-content-writer returns the object and nothing else, and self-checks silently';
END $$;

INSERT INTO schema_migrations (filename, notes)
VALUES ('227_writer_returns_the_object_and_nothing_else.sql',
        'bugs_open/088 prompt half: page-content-writer''s Output Format now says the reply must be the object and nothing else, and the em-dash self-check is explicitly silent. Scope chosen from 5,844 stored responses — 33 of the 34 complete-answers-buried-in-commentary are this agent. Config-only, live on apply. The platform-side recovery (json_envelope.go tier 3) is separate and needs a roll.')
ON CONFLICT (filename) DO NOTHING;

COMMIT;
