-- 454 ROLLBACK — remove the "honest" stop word (rule 19) from page-content-writer.
--
-- Removes ONLY the rule 19 sentence this migration added, by deleting the exact
-- text it inserted. It does not restore from the snapshot, deliberately: the
-- snapshot is a whole-config revert and would silently discard any other change
-- another session has made to this agent since 454 applied.
--
-- If you want the whole pre-454 config back instead, take it from
-- agent_definitions_backup (note '454_page_content_writer_honest_stop_word:
-- pre-update') -- and read the landmine first: snapshot_agent has TWO overloads
-- writing to TWO different tables, and the two-arg form used by 454 writes to
-- agent_definitions_backup, NOT to an is_snapshot row in agent_definitions.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='page-content-writer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '454 ROLLBACK: expected exactly 1 live page-content-writer row, found %', n;
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
         to_jsonb(
           replace(
             default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
             E'\n'
             || '19. Never write the words "honest", "honestly" or "honesty" in page copy. This does NOT soften rule 18: keep BEING straight with the reader, and keep preferring a general truth to a specific invention - just never LABEL it. Show it instead, by naming the limit, the failure mode, or what the thing cannot do. Say "we cannot tell you X" rather than "an honest assessment". (Owner ruling: the word was overused across the estate. The single blessed exception is idea.uk''s report hero, protected by decision record D-005, and you are not writing that sentence.)',
             ''
           )
         ),
         false)
 WHERE type='page-content-writer' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE t text;
BEGIN
  SELECT default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
    INTO t FROM agent_definitions
   WHERE type='page-content-writer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF t LIKE '%19. Never write the words%' THEN
    RAISE EXCEPTION '454 ROLLBACK: rule 19 is still present -- the stored text did not match the literal being removed';
  END IF;
  IF t NOT LIKE '%18. It is ALWAYS better to be honest and general than specific and fabricated.%' THEN
    RAISE EXCEPTION '454 ROLLBACK: rule 18 is missing after the removal';
  END IF;
  RAISE NOTICE '454 ROLLBACK: rule 19 removed; rule 18 intact';
END $$;

COMMIT;
