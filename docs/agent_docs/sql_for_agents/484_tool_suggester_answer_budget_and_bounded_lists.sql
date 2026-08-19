-- 484 — tool-suggester: raise the ANSWER budget to 6000 and BOUND the two lists
--       that scale with the library (bugs_open/319)
--
-- WHY. Migration 445 (bugs_open/275) removed `LIMIT 30` from `load_library_tools`,
-- so the model is now shown the whole library — 80 tools instead of 30, proven at
-- the artefact 2026-08-19 (80 of 81 present, 51 past rank 30, highest rank 81).
-- The FIRST post-fix run then FAILED:
--
--   step suggest_tools failed: AI call failed with unhandled error:
--   response truncated: stop_reason=max_tokens (output_tokens=3000 reached the
--   configured cap, 11090 chars recovered); raise max_tokens or shorten the prompt
--
-- THE CAUSE IS NOT "ANSWERS GOT LONGER" — IT IS ONE INSTRUCTION THAT ENUMERATES
-- THE LIBRARY. The prompt says *"Also list tools from the library that you
-- considered but rejected, with a reason"*, so `rejected_tools` carries roughly one
-- entry per tool SHOWN. Measured across the five most recent successful answers:
--
--   output_tokens | resp_chars | rejected_listed | suggestions | rejected_chars
--            2438 |       9394 |              28 |           4 |          4665  (50%)
--            1749 |       6630 |              30 |           1 |          4399  (66%)
--            2679 |      10104 |              26 |           5 |          4103  (41%)
--            2921 |      11103 |              30 |           8 |          4105  (37%)
--            2254 |       8455 |              19 |           6 |          3276  (39%)
--
-- The rejected list is 37–66% of every answer and its COUNT tracks the menu (28,
-- 30, 26, 30, 19). Tripling the menu triples that section: ~4,400 chars becomes
-- ~11,500 — which is exactly the 11,090 chars recovered before truncation.
--
-- The suggestions themselves were never the problem: 1–8 per run, and the
-- add_tool items they create run 1–2 per batch.
--
-- ⚠ THE MARGIN WAS ALREADY GONE BEFORE 445, AND NOBODY LOOKED. Across all 59
-- historical calls output ran 1,178–2,921 against the 3,000 cap — an all-time
-- high-water mark of 97.4%, with 5 calls within 15% of the ceiling, while the menu
-- was a THIRD of today's size. One query over llm_call_log.output_tokens would
-- have shown that before 445 was applied. Recorded in bugs_open/319 as the check
-- to run before widening any prompt payload.
--
-- WHAT THIS DOES, and why both halves:
--   1. max_tokens 3000 -> 6000. Restores function. Owner-authorised 2026-08-19.
--   2. Bounds the two lists in the prompt. A ceiling alone would be reached again
--      the next time the library grows — the library went 71 -> 81 in the three
--      days this lane was open. Bounding the term that SCALES is what closes the
--      door; raising the cap is what reopens it today.
--
-- CHOICE OF N, from the data above rather than from taste:
--   * suggestions:    AT MOST 8 — equal to the highest any run has ever produced,
--                     so it bounds a pathological tail without cutting behaviour
--                     that has actually occurred.
--   * rejected_tools: AT MOST 5 — this is the real fix. An exhaustive rejection
--                     audit is not information anyone consumes (nothing reads
--                     `rejected_tools`; it is not persisted by save_tool_spec),
--                     and it is 37–66% of the bill.
--   Worst case after this: ~8 suggestions (~320 tokens each) + 5 rejections +
--   reasoning ~= 3,500 tokens against a 6,000 cap.
--
-- Scoped by id, pre-state gated, DO/RAISE verify, snapshot first, rollback
-- sidecar — the shape 445/446 used. Config-only, LIVE ON APPLY, no roll needed.
-- Not council-submitted: agent_definitions config is outside the gate's
-- platform/ internal/ pkg/ scope (and see bugs_open/314).
--
-- ROLLBACK: 484_tool_suggester_answer_budget_and_bounded_lists_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('tool-suggester',
  '484_tool_suggester_answer_budget_and_bounded_lists: pre-update');

DO $$
DECLARE n int; mt text; p text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'tool-suggester'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '484: expected exactly 1 live tool-suggester row, found % — a second active row would make the id-scoped UPDATE a silent no-op', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,suggest_tools,config,ai_service,max_tokens}',
         default_config#>>'{workflow,steps,suggest_tools,config,prompt_template}'
    INTO mt, p
    FROM agent_definitions WHERE id = 'c0756913-04b1-489d-86b4-9ec249dc804d';

  IF mt IS DISTINCT FROM '3000' THEN
    RAISE EXCEPTION '484: pre-state max_tokens is %, expected 3000 — refusing to overwrite a value someone else set', mt;
  END IF;
  IF p IS NULL OR position('Also list tools from the library that you considered but rejected, with a reason.' in p) = 0 THEN
    RAISE EXCEPTION '484: the unbounded rejection instruction is not present — prompt has changed under me, refusing';
  END IF;
  IF position('Return ONLY valid JSON:' in p) = 0 THEN
    RAISE EXCEPTION '484: anchor "Return ONLY valid JSON:" missing — refusing to insert blind';
  END IF;
  IF position('HARD LIMITS ON THE ANSWER' in p) > 0 THEN
    RAISE EXCEPTION '484: already applied (HARD LIMITS block present) — refusing to double-apply';
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           default_config,
           '{workflow,steps,suggest_tools,config,ai_service,max_tokens}',
           to_jsonb(6000)
         ),
         '{workflow,steps,suggest_tools,config,prompt_template}',
         to_jsonb(
           replace(
             replace(
               default_config#>>'{workflow,steps,suggest_tools,config,prompt_template}',
               $old$Also list tools from the library that you considered but rejected, with a reason.$old$,
               $new$Also list AT MOST 5 library tools you considered but rejected — the closest calls only, not an audit of the whole library — each with a one-line reason.$new$
             ),
             $anchor$Return ONLY valid JSON:$anchor$,
             $limits$HARD LIMITS ON THE ANSWER — these bound its SIZE, not its ambition:
- AT MOST 8 entries in "suggestions".
- AT MOST 5 entries in "rejected_tools" — the closest calls only, never one per library tool.
An empty "suggestions" array remains a valid and often correct answer.

Return ONLY valid JSON:$limits$
           )
         )
       ),
       updated_at = now()
 WHERE id = 'c0756913-04b1-489d-86b4-9ec249dc804d'
   AND type = 'tool-suggester'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE mt text; p text;
BEGIN
  SELECT default_config#>>'{workflow,steps,suggest_tools,config,ai_service,max_tokens}',
         default_config#>>'{workflow,steps,suggest_tools,config,prompt_template}'
    INTO mt, p
    FROM agent_definitions WHERE id = 'c0756913-04b1-489d-86b4-9ec249dc804d';

  IF mt IS DISTINCT FROM '6000' THEN
    RAISE EXCEPTION '484: max_tokens is % after the update, expected 6000', mt;
  END IF;
  IF position('AT MOST 8 entries in "suggestions"' in p) = 0 THEN
    RAISE EXCEPTION '484: the suggestions bound is absent after the update';
  END IF;
  IF position('AT MOST 5 library tools you considered but rejected' in p) = 0 THEN
    RAISE EXCEPTION '484: the rejection bound is absent after the update';
  END IF;
  IF position('Also list tools from the library that you considered but rejected, with a reason.' in p) > 0 THEN
    RAISE EXCEPTION '484: the OLD unbounded rejection instruction survives — both instructions would be live at once';
  END IF;
  -- controls: the parts this migration must NOT have disturbed
  IF position('Available in Library (can be forked and customised)' in p) = 0
     OR position('related_pages' in p) = 0
     OR position('experience_plan' in p) = 0
     OR position('Return ONLY valid JSON:' in p) = 0 THEN
    RAISE EXCEPTION '484: the prompt lost a load-bearing section — the replace() hit more than it should have';
  END IF;
  IF (SELECT default_config#>>'{workflow,steps,load_library_tools,config,query}'
        FROM agent_definitions WHERE id = 'c0756913-04b1-489d-86b4-9ec249dc804d') ~* 'LIMIT[[:space:]]+[0-9]+' THEN
    RAISE EXCEPTION '484: a row LIMIT reappeared on load_library_tools — 445 must stay undone';
  END IF;
END $$;

COMMIT;
