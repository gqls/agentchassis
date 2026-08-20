-- 501 — ask the writer for a SHORTER sentence (bugs_open/320)
--
-- TWO REASONS, and the second one is a mitigation for a defect that is NOT fixed here.
--
-- 1. IT IS BETTER COPY ON ITS OWN MERITS. A meta description is displayed at roughly
--    155 characters and the house style's first rule is "one idea, one sentence". The
--    prompt asked for 120-155 chars without bounding WORDS, and 155 characters of
--    long words is a 28-word sentence. Asking for <= 20 words lands the same
--    information in the same display budget and reads faster. This edit would be
--    worth making with no gate involved at all.
--
-- 2. IT ALSO STOPS AN HOURLY RETRY ON TWO SITES — and that half is a WORKAROUND, said
--    plainly rather than dressed up as a fix.
--
-- ── THE DEFECT IT WORKS AROUND ──────────────────────────────────────────────
--
-- `save_page_meta_description` runs the site's `VoiceGate.ScanVoice` over the
-- candidate. That gate carries two KINDS of rule:
--   * CONTENT rules — banned phrases. Correct to apply to any string, and they work:
--     proven 2026-08-20, the writer was refused for "trust" and, once told the rules
--     (499/500), avoided it and wrote "what builds and breaks confidence" instead.
--   * DENSITY / DISTRIBUTION rules — mean sentence length, long-sentence share,
--     em-dash per 1000 words, triads per page, contraction expectation. These are
--     STATISTICS OVER A CORPUS. `ScanVoiceTells` computes them across a page's
--     rendered blocks. **Over a single sentence they are not measurements.** "Mean
--     sentence length" of one sentence is just its length.
--
-- `[MEASURED 2026-08-20]` the second kind bites on exactly TWO of 27 sites.
-- Of the 9 sites with an enabled voice gate, SEVEN set `mean_sentence_words: 100000`
-- and `long_sentence_words: 10000` — thresholds high enough to disable the length
-- checks deliberately while keeping the banned-phrase list. Only
-- `leopardessconsulting.co.uk` and `oufe.com` leave them unset and so inherit the Go
-- defaults (mean 22, long 25 — `voicetells.go:179-180`). Those two refused a
-- perfectly good 24-word description.
--
-- **THE REAL FIX IS IN GO AND IS NOT IN THIS FILE**: `metaDescriptionFailsCopyGates`
-- should apply the CONTENT rules to a single-sentence field and skip the
-- distribution ones, because a one-sentence sample cannot support them. That needs a
-- build and a roll. Recorded in `bugs_open/320` so it is not lost.
--
-- ⚠ AND THE THING TO BE HONEST ABOUT: tightening a prompt so output clears a gate is
-- one move away from "relax the checker to agree with the content", which this estate
-- has a standing rule against. The distinction claimed here, which a reader should
-- test rather than accept: **the gate is not being changed, and the new instruction
-- improves the copy whether or not any gate exists.** If either half of that stopped
-- being true, this file would be the wrong shape.
--
-- ROLLBACK: 501_backfiller_shorter_sentences_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('meta-description-backfiller', '501_shorter_sentences: pre-update');

DO $$
DECLARE p text;
BEGIN
  SELECT default_config#>>'{workflow,steps,write_descriptions,config,prompt_template}'
    INTO p FROM agent_definitions WHERE type='meta-description-backfiller' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF p IS NULL THEN
    RAISE EXCEPTION '501: no prompt found';
  END IF;
  IF position('- 120-155 characters.' in p) = 0 THEN
    RAISE EXCEPTION '501: the length rule is not in its expected form — prompt changed under me';
  END IF;
  IF position('AT MOST 20 words' in p) > 0 THEN
    RAISE EXCEPTION '501: already applied';
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,write_descriptions,config,prompt_template}',
         to_jsonb(
           replace(
             default_config#>>'{workflow,steps,write_descriptions,config,prompt_template}',
             '- 120-155 characters.',
             E'- 110-150 characters AND AT MOST 20 words. Both limits, not either. A search result shows about 155 characters, and 20 short words carry more than 28 long ones.\n- 120-155 characters.'
           )
         )
       ),
       updated_at = now()
 WHERE type='meta-description-backfiller' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- remove the now-superseded original line so the prompt does not carry two limits
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,write_descriptions,config,prompt_template}',
         to_jsonb(
           replace(
             default_config#>>'{workflow,steps,write_descriptions,config,prompt_template}',
             E'- 120-155 characters. Shorter wastes the slot; longer is cut off mid-word.\n',
             ''
           )
         )
       ),
       updated_at = now()
 WHERE type='meta-description-backfiller' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE p text;
BEGIN
  SELECT default_config#>>'{workflow,steps,write_descriptions,config,prompt_template}'
    INTO p FROM agent_definitions WHERE type='meta-description-backfiller' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF position('AT MOST 20 words' in p) = 0 THEN
    RAISE EXCEPTION '501 VERIFY: the word bound was not inserted';
  END IF;
  IF position('- 120-155 characters. Shorter wastes' in p) > 0 THEN
    RAISE EXCEPTION '501 VERIFY: the superseded line survives — the prompt now carries two conflicting limits';
  END IF;
  IF position('{{range .voice_rules.rows}}' in p) = 0 THEN
    RAISE EXCEPTION '501 VERIFY: 500''s rules block was damaged';
  END IF;
  RAISE NOTICE '501 OK: <=20 words and 110-150 chars, single limit, rules block intact';
END $$;

COMMIT;
