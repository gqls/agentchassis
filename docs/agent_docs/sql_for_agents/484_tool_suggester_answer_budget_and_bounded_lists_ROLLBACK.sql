-- ROLLBACK for 484 — restores max_tokens 3000 and the two unbounded list instructions.
--
-- ⚠ ROLLING THIS BACK RESTORES A BROKEN AGENT. 3000 is the cap the first post-fix
-- run hit (stop_reason=max_tokens, step FAILED, bugs_open/319) while the library is
-- ~80 tools. Only use this if 484 itself proves harmful — and if you do, expect
-- suggest_tools to fail again until either the answer budget or the library size
-- changes. Reverting 484 does NOT revert 445; the menu stays wide.
--
-- Gated the same way as the forward migration: refuses unless the live row is the
-- one 484 wrote, still carrying 484's exact text.

BEGIN;

SELECT snapshot_agent('tool-suggester',
  '484_ROLLBACK: pre-revert');

DO $$
DECLARE n int; mt text; p text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'tool-suggester'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '484_ROLLBACK: expected exactly 1 live tool-suggester row, found %', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,suggest_tools,config,ai_service,max_tokens}',
         default_config#>>'{workflow,steps,suggest_tools,config,prompt_template}'
    INTO mt, p
    FROM agent_definitions WHERE id = 'c0756913-04b1-489d-86b4-9ec249dc804d';

  IF mt IS DISTINCT FROM '6000' THEN
    RAISE EXCEPTION '484_ROLLBACK: max_tokens is %, expected 6000 — 484 is not the live state, refusing', mt;
  END IF;
  IF position('HARD LIMITS ON THE ANSWER' in p) = 0 THEN
    RAISE EXCEPTION '484_ROLLBACK: 484''s HARD LIMITS block is absent — nothing of 484 to revert, refusing';
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           default_config,
           '{workflow,steps,suggest_tools,config,ai_service,max_tokens}',
           to_jsonb(3000)
         ),
         '{workflow,steps,suggest_tools,config,prompt_template}',
         to_jsonb(
           replace(
             replace(
               default_config#>>'{workflow,steps,suggest_tools,config,prompt_template}',
               $new$Also list AT MOST 5 library tools you considered but rejected — the closest calls only, not an audit of the whole library — each with a one-line reason.$new$,
               $old$Also list tools from the library that you considered but rejected, with a reason.$old$
             ),
             $limits$HARD LIMITS ON THE ANSWER — these bound its SIZE, not its ambition:
- AT MOST 8 entries in "suggestions".
- AT MOST 5 entries in "rejected_tools" — the closest calls only, never one per library tool.
An empty "suggestions" array remains a valid and often correct answer.

Return ONLY valid JSON:$limits$,
             $anchor$Return ONLY valid JSON:$anchor$
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
  IF mt IS DISTINCT FROM '3000' THEN
    RAISE EXCEPTION '484_ROLLBACK: max_tokens is % after revert, expected 3000', mt;
  END IF;
  IF position('HARD LIMITS ON THE ANSWER' in p) > 0 THEN
    RAISE EXCEPTION '484_ROLLBACK: the HARD LIMITS block survives the revert';
  END IF;
  IF position('Also list tools from the library that you considered but rejected, with a reason.' in p) = 0 THEN
    RAISE EXCEPTION '484_ROLLBACK: the original rejection instruction was not restored';
  END IF;
  IF position('Return ONLY valid JSON:' in p) = 0 OR position('related_pages' in p) = 0 THEN
    RAISE EXCEPTION '484_ROLLBACK: the prompt lost a load-bearing section';
  END IF;
END $$;

COMMIT;
