-- 648 — the owner's comparison rule, into the site brief BEFORE the queued rebuilds read it
--
-- OWNER, 2026-08-26, verbatim: "I think perhaps, as a trial, whenever we want to write the second
-- half of one of these sentences, we should just stop before the negative (or the 'not' or the
-- 'instead of') and leave that part of the comparison out all together. We don't need to sound
-- competitive like this. There is no hidden competition. We offer what we offer straight up."
--
-- WHY IT IS BETTER THAN THE BAN IT REPLACES IN PRACTICE. 646/647 removed the DEMONSTRATIONS of the
-- construction from this brief. This adds the POSITIVE instruction — what to do instead of what to
-- avoid — which is the one shape a writer cannot follow by imitation and get wrong. It is also
-- stated in his register, deliberately: the rule and the voice it asks for are the same thing.
--
-- TIMING IS THE POINT. Four needs_page items (the copy rebuilds for the canary's nine pages) are
-- sitting at 'triaged' and have NOT run. content_direction is live immediately, so this rule
-- reaches them before the writer does. Applied now rather than after, which is the difference
-- between the canary testing the rule and the canary testing its absence.
--
-- ⚠ 'formatted' is DERIVED (FormatContentDirection). The rule is appended to writing_rules AND the
-- same text spliced into formatted, which is byte-equivalent to a regeneration — see 646's header.
-- Rollback: 648_..._ROLLBACK.sql

BEGIN;

UPDATE site_specs ss SET is_current = false, superseded_at = now()
  FROM sites s WHERE s.id = ss.site_id AND s.domain = 'finetuning.uk'
   AND ss.aspect = 'content_direction' AND ss.is_current;

INSERT INTO site_specs (site_id, aspect, data, source, created_by, is_current, notes)
SELECT ss.site_id, ss.aspect,
       jsonb_set(
         jsonb_set(ss.data, '{writing_rules}',
                   (ss.data->'writing_rules') || to_jsonb('When a sentence sets up a comparison, write the first half and stop. Leave the second half out altogether - the "not", the "instead of", the "rather than". We do not need to sound competitive. There is no hidden competition. We offer what we offer, straight up. (Owner instruction, 2026-08-26.)'::text)),
         '{formatted}',
         to_jsonb((ss.data->>'formatted') || E'\n- ' || 'When a sentence sets up a comparison, write the first half and stop. Leave the second half out altogether - the "not", the "instead of", the "rather than". We do not need to sound competitive. There is no hidden competition. We offer what we offer, straight up. (Owner instruction, 2026-08-26.)')),
       'owner_ruling', 'claude-finetuning-uk-lane', true,
       'Owner comparison rule, 2026-08-26: write the first half of a comparison and stop.'
  FROM site_specs ss JOIN sites s ON s.id = ss.site_id
 WHERE s.domain = 'finetuning.uk' AND ss.aspect = 'content_direction'
   AND ss.superseded_at IS NOT NULL ORDER BY ss.superseded_at DESC LIMIT 1;

DO $$
DECLARE n_rules int; has_rule bool; has_fmt bool; n_rt int;
BEGIN
  SELECT jsonb_array_length(ss.data->'writing_rules'),
         (ss.data->>'writing_rules') LIKE '%There is no hidden competition%',
         (ss.data->>'formatted')     LIKE '%There is no hidden competition%',
         (length(ss.data::text)-length(replace(ss.data::text,'rather than','')))/length('rather than')
    INTO n_rules, has_rule, has_fmt, n_rt
    FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='finetuning.uk' AND ss.aspect='content_direction' AND ss.is_current;
  IF NOT has_rule THEN RAISE EXCEPTION '648: the rule did not land in writing_rules'; END IF;
  IF NOT has_fmt  THEN RAISE EXCEPTION '648: the rule did not land in formatted — the writer reads formatted, so this half is the one that matters'; END IF;
  IF n_rules < 13 THEN RAISE EXCEPTION '648: writing_rules has % entries, expected the existing 12 plus 1', n_rules; END IF;
  -- The rule NAMES the constructions it forbids; that is a ban naming what it bans, and is the
  -- one place "rather than" is allowed back. It lands TWICE because the rule text goes into
  -- writing_rules AND into the derived formatted block, so the document carries two copies of one
  -- rule. Pinned at 2 so nothing else creeps in. (Pinned at 1 on the first attempt and the guard
  -- refused — the assertion had been written from the headline expectation instead of from where
  -- the text actually lands.)
  IF n_rt <> 2 THEN RAISE EXCEPTION '648: % "rather than" in the brief, want exactly 2 (the rule naming what it bans, in writing_rules and in formatted)', n_rt; END IF;
  RAISE NOTICE '648 OK: rule live in writing_rules (%) and in formatted; rather-than pinned at 2 (the rule naming what it bans, in writing_rules and in formatted)', n_rules;
END $$;

COMMIT;
