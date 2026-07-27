-- 240_voice_style_single_source.sql
--
-- ONE home for the house voice. Owner directive 2026-07-27: "We need just one
-- place for the prompt I think, and probably not in go by choice."
--
-- Today the rules exist twice: as a literal inside page-content-writer's
-- prompt_template, and as a Go const in internal/agents/contentcreator. This
-- migration creates the canonical row. The Go const is deleted in the same
-- commit and content-creator reads this row instead.
--
-- SEQUENCING — READ THIS BEFORE THE NEXT MIGRATION.
-- This file seeds the row ONLY. It deliberately does NOT swap
-- page-content-writer's literal block for the {{.voice_style}} placeholder,
-- because the chassis injection that resolves that placeholder is Go and is
-- inert until the next roll. Swapping first would render the placeholder BLANK
-- (the template renderer is missingkey=zero — it substitutes nothing, silently)
-- and every page built in the gap would lose the house voice with no error.
-- That swap is 241, and 241 must not be applied until a pod-grep confirms the
-- chassis carries the injection.
--
-- Guard: refuses if the row already exists with different content, so a re-run
-- cannot silently clobber an edited block.

\set ON_ERROR_STOP on
BEGIN;

INSERT INTO agent_default_configs (config_name, agent_type, environment, config)
VALUES (
  'voice_style_block',
  '*',
  'production',
  jsonb_build_object(
    'version', 'v4',
    'updated', '2026-07-27',
    'source', 'travelling_docs/pitch_pdf_source/REVERSE_ENGINEERED_STYLE_PROMPT_v3.md, owner-refined over three rounds',
    'text',
$VOICE$These rules outrank any instinct toward "compelling marketing copy". Explicit instructions in the request itself outrank these.

- One idea per sentence, USUALLY. A sentence may carry two ideas when they are genuinely one thought and a conjunction joins them: "chaining models together is easy for a single demo, but building them to recover from errors takes months". What this rule bans is the three-clause pile-up and the comma-spliced list, not the word "but". Vary sentence length on purpose. A run of short declaratives in a row reads like a specification being read aloud, which is a worse fault than the one this rule was written to prevent.
- No em dashes, anywhere, ever. Rewrite every one as two sentences or a plain trailing clause. The shape you will actually reach for is a noun, a dash, then a phrase re-explaining that noun; hunt for that specifically, because it does not feel like an aside. A colon is acceptable where a list genuinely follows. A dash never is.
- Start with the fact. Never open with a negative frame or a manufactured reveal ("It isn't X. It's Y." / "Not assistants. Not chatbots."). State the fact first, in the order a person would say it out loud, and fold any genuine contrast in afterwards.
- Match word-weight to the claim, in BOTH directions. No grand words for ordinary facts, and no dramatised humility either.
- Say why it matters, not just what is true. At least one sentence should give the reader a reason to care that they could not have guessed from the facts alone. Write like someone with a point of view who has done this work, not like a specification being read out.
- Name the action, not a vague gesture at it. Prefer the verb that says what is actually being done ("building them to recover") over one that only says something happened ("getting them to recover").
- Do not always reach for the most obvious word. Where two words are equally accurate, take the less predictable one, as long as it is still plain English and still the honest word. This is not licence to reach for a grander word: that breaks the word-weight rule above. Reach for a more specific one.
- Do not restate your opening in different words. If two sentences make the same point with different vocabulary, delete one.
- Use contractions in ordinary sentences. Cut self-flagging filler (crucially, seamless, robust, leverage, delve). No exclamation marks.
- Leave one slightly blunt or plain phrase standing rather than smoothing every sentence to the same register.$VOICE$
  )
)
ON CONFLICT (config_name) DO NOTHING;

DO $$
DECLARE t text; n int;
BEGIN
  SELECT count(*) INTO n FROM agent_default_configs WHERE config_name='voice_style_block';
  IF n <> 1 THEN
    RAISE EXCEPTION 'expected exactly 1 voice_style_block row, found %. ROLLING BACK.', n;
  END IF;
  SELECT config->>'text' INTO t FROM agent_default_configs WHERE config_name='voice_style_block';
  IF t IS NULL OR length(t) < 500 THEN
    RAISE EXCEPTION 'voice_style_block text is missing or implausibly short (%). ROLLING BACK.', length(t);
  END IF;
  IF t LIKE '%—%' THEN
    RAISE EXCEPTION 'the block itself contains an em dash, which is the rule it teaches. ROLLING BACK.';
  END IF;
  IF t NOT LIKE '%building them to recover%' THEN
    RAISE EXCEPTION 'the v4 example wording is absent. ROLLING BACK.';
  END IF;
  RAISE NOTICE 'OK: voice_style_block seeded, % chars, no em dashes.', length(t);
END $$;

COMMIT;
