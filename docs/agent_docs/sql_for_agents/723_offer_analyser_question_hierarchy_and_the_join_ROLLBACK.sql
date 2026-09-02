-- ROLLBACK for 723 — remove the question hierarchy, its gate step, and the chain.
--
-- ⚠ WHAT A ROLLBACK CANNOT UNDO: any site re-analysed while 723 was live has a
-- `question_hierarchy` key in its CURRENT offer_ordering row. Removing the
-- prompt contract does NOT remove those keys, and because `ordering` is
-- deep-merged, a later run that omits the key leaves the stored one standing and
-- looking current. Census before deciding, and strip the key explicitly if you
-- mean it to be gone:
--   SELECT s.domain FROM site_specs ss JOIN sites s ON s.id=ss.site_id
--    WHERE ss.is_current AND ss.aspect='offer_ordering'
--      AND ss.data ? 'question_hierarchy';
--
-- Restores the chain to: repair_ordering_register -> write_offer_ordering.
BEGIN;

SELECT snapshot_agent('offer-analyser', '723_ROLLBACK: pre-revert');

UPDATE agent_definitions
SET default_config = jsonb_set(
      jsonb_set(
        (default_config #- '{workflow,steps,repair_hierarchy_register}'),
        '{workflow,steps,repair_ordering_register,next_step}', '"write_offer_ordering"'),
        '{workflow,steps,write_offer_ordering,config,spec_data}', '"ordering_register_checked.object"')
WHERE type = 'offer-analyser' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
      '{workflow,steps,run_offer_analysis,config,prompt}',
      to_jsonb(
        replace(
          regexp_replace(
            default_config->'workflow'->'steps'->'run_offer_analysis'->'config'->>'prompt',
            'THE QUESTION HIERARCHY\..*?OUTPUT\. Return ONE JSON object and nothing else',
            'OUTPUT. Return ONE JSON object and nothing else', 'ns'),
          '"question_hierarchy": [{"rank": 1, "question": "...", "why": "...", "from_field": "satisfaction_condition", "answered_by": 3, "unanswered": false}], "spec_version": 2}, "findings"',
          '"spec_version": 1}, "findings"')))
WHERE type = 'offer-analyser' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE p text; n text; s text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'run_offer_analysis'->'config'->>'prompt',
           default_config->'workflow'->'steps'->'repair_ordering_register'->>'next_step',
           default_config->'workflow'->'steps'->'write_offer_ordering'->'config'->>'spec_data'
      INTO p, n, s FROM agent_definitions
     WHERE type='offer-analyser' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF position('question_hierarchy' in p) > 0 THEN
        RAISE EXCEPTION '723 ROLLBACK: question_hierarchy still in the prompt';
    END IF;
    IF position('THE QUESTION HIERARCHY.' in p) > 0 THEN
        RAISE EXCEPTION '723 ROLLBACK: the guidance block was not removed';
    END IF;
    IF position('"spec_version": 1' in p) = 0 THEN
        RAISE EXCEPTION '723 ROLLBACK: spec_version was not restored to 1';
    END IF;
    IF n IS DISTINCT FROM 'write_offer_ordering' OR s IS DISTINCT FROM 'ordering_register_checked.object' THEN
        RAISE EXCEPTION '723 ROLLBACK: the chain was not restored (next=%, spec_data=%)', n, s;
    END IF;
    RAISE NOTICE '723 ROLLBACK OK: hierarchy removed, chain restored to the lead_with gate';
END $$;

COMMIT;
