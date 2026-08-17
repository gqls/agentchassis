-- 454 — page-content-writer: one stop word for "honest", at the writer
-- (OWNER RULING 2026-08-17)
--
-- WHY. The 2026-08-12 fleet ban on the word was implemented by cleaning
-- INSTANCES: 16 site specs, 124 page sections, 12 sites. It worked and it did
-- not hold. Measured 2026-08-16, same predicate as the original census: the
-- reader-visible count went 53 -> 18 -> back up to 30 pages across 11 sites,
-- and 23 of the 37 matching components were CREATED AFTER the sweep, the
-- newest dated that day. The regrown copy is ordinary prose on sites the
-- sweeping lane never touched -- "an honest read of your own credit file",
-- "the more honest way to see which loan costs less", "the failure modes named
-- honestly" -- with no shared spec term and no shared component behind it.
-- That is a writer habit, and a habit cannot be swept.
--
-- OWNER RULING 2026-08-17: "I think we have dealt with the honesty problem
-- enough. It doesn't need any more sweeps. just stop the overuse probably by a
-- single stop word in the content writer agent would do it sufficiently."
-- So: no more remediation passes, and the rule goes in ONE place, at the point
-- of writing, rather than into 23 per-site `content_direction.avoid` lists --
-- which would itself have been a sweep, and would drift the moment a 24th site
-- is created.
--
-- WHY THIS AGENT. `page-content-writer` is the live page-prose writer: 1,826
-- llm_call_log rows in the 7 days to 2026-08-17, newest that day. The
-- similarly named `content-writer` is DORMANT (zero calls in the same window),
-- so editing it would have changed nothing and looked identical afterwards.
--
-- WHY IT IS RULE 19 AND WHY THE WORDING IS DEFENSIVE. This prompt already
-- contains four uses of the word as an instruction about the MODEL'S OWN
-- truthfulness, the load-bearing one being rule 18: "It is ALWAYS better to be
-- honest and general than specific and fabricated." Those are ANTI-FABRICATION
-- rules on an estate whose whole evidence-gating apparatus exists to stop
-- invented content, and a naive "remove the word from the prompt" pass would
-- have deleted them -- a grep on a prompt cannot see polarity. Rule 19
-- therefore states the distinction the owner himself drew on 2026-07-18 (show
-- the honesty, do not label it): keep BEING straight, never LABEL it. It is
-- placed immediately after 18 so the two are read together.
--
-- THE BLESSED EXCEPTION is named so the writer does not "fix" it: idea.uk's
-- report hero keeps the word by owner ruling 2026-08-16 and is protected by
-- decision record D-005 (covers report/hero, guard asserts the served page
-- contains "honest assessment").
--
-- SCOPE. Config only -- live immediately, no build needed. Reversible via the
-- ROLLBACK companion, and a pre-change snapshot is taken below
-- (snapshot_agent note '454_page_content_writer_honest_stop_word: pre-update').
-- Not a guarantee change: it constrains word choice in generated copy and adds
-- no capability, names no new key and changes no contract.

BEGIN;

SELECT snapshot_agent('page-content-writer',
  '454_page_content_writer_honest_stop_word: pre-update');

-- Guard 1: exactly one live row. A shadowed sibling would make the UPDATE
-- below hit the wrong row (or none) while reporting success.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='page-content-writer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '454: expected exactly 1 live page-content-writer row, found % -- re-read before applying', n;
  END IF;
END $$;

-- Guard 2: the anchor exists, exactly once, and the rule is not already there.
DO $$
DECLARE t text; hits int;
BEGIN
  SELECT default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
    INTO t FROM agent_definitions
   WHERE type='page-content-writer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF t IS NULL THEN
    RAISE EXCEPTION '454: prompt_template not found at the expected path -- the sub_workflow shape has changed';
  END IF;

  hits := (length(t) - length(replace(t, 'A real visitor will trust a general statement of capability more than a fabricated testimonial.', ''))) / length('A real visitor will trust a general statement of capability more than a fabricated testimonial.');
  IF hits <> 1 THEN
    RAISE EXCEPTION '454: anchor (end of rule 18) found % times, expected exactly 1 -- re-read the prompt before applying', hits;
  END IF;

  IF t LIKE '%19. Never write the words%' THEN
    RAISE EXCEPTION '454: rule 19 is already present -- this migration has been applied';
  END IF;

  -- The anti-fabrication rules this file must NOT disturb.
  IF t NOT LIKE '%18. It is ALWAYS better to be honest and general than specific and fabricated.%' THEN
    RAISE EXCEPTION '454: rule 18 is missing -- refusing to add a word ban to a prompt that has lost its anti-fabrication rule';
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
         to_jsonb(
           replace(
             default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
             'A real visitor will trust a general statement of capability more than a fabricated testimonial.',
             'A real visitor will trust a general statement of capability more than a fabricated testimonial.'
             || E'\n'
             || '19. Never write the words "honest", "honestly" or "honesty" in page copy. This does NOT soften rule 18: keep BEING straight with the reader, and keep preferring a general truth to a specific invention - just never LABEL it. Show it instead, by naming the limit, the failure mode, or what the thing cannot do. Say "we cannot tell you X" rather than "an honest assessment". (Owner ruling: the word was overused across the estate. The single blessed exception is idea.uk''s report hero, protected by decision record D-005, and you are not writing that sentence.)'
           )
         ),
         false)
 WHERE type='page-content-writer' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- Post-state verify. DO/RAISE, never a bare SELECT: ON_ERROR_STOP ignores a
-- non-empty result, so a verify block of SELECTs cannot stop the COMMIT.
DO $$
DECLARE t text;
BEGIN
  SELECT default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
    INTO t FROM agent_definitions
   WHERE type='page-content-writer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF t NOT LIKE '%19. Never write the words "honest", "honestly" or "honesty" in page copy.%' THEN
    RAISE EXCEPTION '454: rule 19 did not land';
  END IF;

  -- The four anti-fabrication uses must ALL survive. This is the assertion that
  -- matters: the failure this migration must not cause is a prompt that stopped
  -- telling the model to prefer truth to invention.
  IF t NOT LIKE '%18. It is ALWAYS better to be honest and general than specific and fabricated.%' THEN
    RAISE EXCEPTION '454: rule 18 was damaged by the splice';
  END IF;
  IF t NOT LIKE '%If a field has no honest value, give it an empty string%' THEN
    RAISE EXCEPTION '454: the empty-string rule was damaged by the splice';
  END IF;

  -- Order: 19 must follow 18, or they read as contradictory rather than paired.
  IF position('19. Never write the words' in t) < position('18. It is ALWAYS better' in t) THEN
    RAISE EXCEPTION '454: rule 19 landed before rule 18';
  END IF;

  RAISE NOTICE '454: stop word live; rule 18 and the empty-string rule intact';
END $$;

COMMIT;
