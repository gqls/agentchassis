-- ============================================================================
-- 228_writers_default_to_humanised_voice.sql
--
-- Owner directive, 2026-07-27: "the writer should default to the humanised voice
-- if it is not overridden."
--
-- WHY, WITH THE MEASUREMENT
--   oufe.com's copy reads as machine-written. Counted on the two live pages the
--   day after they shipped:
--     /            3 em dashes,  7 negative-frame constructions
--     /about.html  3 em dashes, 19 negative-frame constructions,
--                  6 sentences opening "It is…" / "It does…"
--
--   `page-content-writer` already carried several of the style rules, including
--   the one against negative framing. It did not work. The rules sat inside a
--   12.5 KB prompt among the schema, the section spec, the research findings and
--   four other rule blocks, and a rule competing with that much context loses.
--
--   `content-writer` carried no style guidance at all.
--
-- THE SOURCE
--   `docs024_key_docs_latest/travelling_docs/pitch_pdf_source/
--    REVERSE_ENGINEERED_STYLE_PROMPT_v3.md` — built 2026-07-17 by comparing AI
--   copy against a hand-edited rewrite the owner judged more readable, then
--   refined across three rounds of the owner critiquing the prompt's own output.
--   It has never been wired into any writer. This migration wires it in.
--
--   Rule 3 is the one the owner named directly: don't state what a thing ISN'T
--   before saying what it is. Round 3 of that document exists because rounds 1
--   and 2 both fixed the literal words and missed the underlying move.
--
-- DEFAULT ON, OVERRIDABLE
--   The block below opens by deferring to the site's own `voice` spec where one
--   exists, so a deliberate house voice still wins. Absent that, this is the
--   voice. Same shape as `globalTellPhrases()` in `datahelpers/voicetells.go:109`
--   — a fleet default unioned with per-site config, rather than a per-site
--   setting that defaults to nothing.
--
-- WHAT THIS DOES NOT DO
--   It does not detect violations after the fact. That is `check_voice_tells`,
--   which is enabled on **1 of 15 sites**. Writing the rule and checking the
--   rule are different jobs; see the companion arming in this migration's notes.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

-- page-content-writer: prompt lives under the sections loop --------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
      to_jsonb(
        (default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template')
        || $voice$

## HOUSE VOICE — THIS IS THE DEFAULT. Follow it unless this site's own voice spec says otherwise.

Write the way a knowledgeable person would explain this out loud to one other person. Preserve every fact, number and name exactly.

Start with the fact. Never open by saying what something is NOT, or what it is not about, before saying what it is. "This isn't about the technology, it's about where the record lives" and "Nothing here is unusual — one choice matters" are the same move wearing different grammar: a negative first, then the real point as a reveal. Say the real point first. If a contrast genuinely helps, fold it in after the fact, as a trailing clause. Two or three sentences built this way in one page means go back and say things in the order a person would say them.

One idea per sentence. Split anything chaining clauses with commas, semicolons or dashes.

Do not use em dashes. Make it two sentences, or fold the aside in as a trailing clause.

Use contractions in ordinary sentences: it's, isn't, doesn't, that's, don't.

Match the word to the size of the fact. Don't reach for "critical", "essential", "powerful", "truth", "transformative" when a plain word carries it. The reverse counts too: "nothing fancy", "surprisingly simple", "no magic here" still ask the reader to be impressed, just by humility. Usually the plainest move is to state what it is and skip characterising it.

Cut these outright: crucially, genuinely, exactly, deliberately, "which is the point", "what matters here is", "the real question is", "at its core", "in essence", seamless, robust, leverage, delve, furthermore, moreover, "load-bearing" as a metaphor. Don't announce that a fact is important before giving it. State it and let it sit.

Don't repeat a sentence shape for rhythm. "Every agent reads them. Every agent writes to them." is one idea said twice because it sounds good. Combine it.

A very short sentence can close a thought, and a matched contrasting pair can land a point the reader might have guessed wrong. Both are earned once or twice per page at most. A page that lands a beat on every paragraph teaches the reader to skip them.

Vary paragraph length. One sentence here, four there. Leave one plain, slightly blunt phrase unpolished rather than smoothing every line to the same register.

Never open consecutive sentences with "It is", "It does", "This is", "There is".

No exclamation marks. No hype adjectives in either direction.
$voice$
      ),
      false
    ),
    updated_at = NOW()
WHERE type = 'page-content-writer'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template' NOT LIKE '%HOUSE VOICE — THIS IS THE DEFAULT%';

-- content-writer: prompt lives one level up (verified 2026-07-26 — the two
-- writers keep their prompts at DIFFERENT paths, and a single UPDATE covers one)
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,generate_content,config,prompt_template}',
      to_jsonb(
        (default_config->'workflow'->'steps'->'generate_content'->'config'->>'prompt_template')
        || $voice2$

## HOUSE VOICE — THIS IS THE DEFAULT. Follow it unless this site's own voice spec says otherwise.

Write the way a knowledgeable person would explain this out loud to one other person. Preserve every fact, number and name exactly.

Start with the fact. Never open by saying what something is NOT before saying what it is. Say the real point first, and fold any genuine contrast in afterwards as a trailing clause.

One idea per sentence. No em dashes. Contractions in ordinary sentences.

Match the word to the size of the fact, in both directions: neither "transformative" nor "nothing fancy here". Cut crucially, genuinely, exactly, "which is the point", "what matters here is", at its core, seamless, robust, leverage, delve, furthermore.

Don't repeat a sentence shape for rhythm. Vary paragraph length. Never open consecutive sentences with "It is" or "This is". No exclamation marks, no hype adjectives.
$voice2$
      ),
      false
    ),
    updated_at = NOW()
WHERE type = 'content-writer'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config->'workflow'->'steps'->'generate_content'->'config'->>'prompt_template' IS NOT NULL
  AND default_config->'workflow'->'steps'->'generate_content'->'config'->>'prompt_template' NOT LIKE '%HOUSE VOICE — THIS IS THE DEFAULT%';

COMMIT;

-- Verify BOTH, by type — one UPDATE never covers both writers:
--   SELECT type, (default_config::text LIKE '%HOUSE VOICE — THIS IS THE DEFAULT%') AS has_voice
--     FROM agent_definitions WHERE type IN ('page-content-writer','content-writer')
--      AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
--
-- The rule is now written. Whether it is FOLLOWED is a separate question, and the
-- answer for oufe was "no" even with rule 3 already present. Check the output, not
-- the prompt: count em dashes and negative-frame openers on the rendered page.
