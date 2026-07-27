-- ============================================================================
-- 230_grounded_explainer_house_voice.sql
--
-- The grounded-explainer writes prose and never got the house voice.
--
-- Migration 228 made the humanised voice the default for `page-content-writer`
-- and `content-writer`, on the owner's instruction. It did not touch
-- `grounded-explainer`, which was written the day before and composes long-form
-- explainers — the most prose-heavy output on the platform.
--
-- The result was visible immediately. The first Thames Water draft arrived with
-- **20 em dashes** in 7,883 characters, and its closing sentence claimed the
-- page's own verification was complete. Both had to be fixed by hand before
-- publishing.
--
-- Same shape as everything else this week: the rule existed and the artefact
-- that predated it carried on regardless. 228 wired two writers because two
-- writers were what I was looking at.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config, '{workflow,steps,compose_explainer,config,prompt_template}',
      to_jsonb(
        (default_config->'workflow'->'steps'->'compose_explainer'->'config'->>'prompt_template')
        || $voice$

## HOUSE VOICE — follow this unless the site's own voice spec says otherwise

Write the way a knowledgeable person explains something out loud to one other person.

Start with the fact. Never open by saying what something is NOT before saying what it is. Fold any genuine contrast in afterwards, as a trailing clause.

One idea per sentence. Split anything chaining clauses with commas or dashes.

Do not use em dashes at all. Use two sentences, or a comma, or brackets.

Contractions in ordinary sentences: it's, isn't, doesn't, don't.

Match the word to the size of the fact, in both directions. Neither "critical" and "transformative" nor "nothing fancy" and "surprisingly simple".

Cut: crucially, genuinely, exactly, "which is the point", "what matters here is", "the real question is", at its core, in essence, seamless, robust, leverage, delve, furthermore, moreover.

Vary paragraph length. Never open consecutive sentences with "It is", "This is" or "There is". No exclamation marks.

On the closing note about the page's own limits: state plainly that the page can be wrong and that a reader should check the primary source. Do NOT claim the page's verification is complete, thorough, or that every gap has been declared — you cannot know that, and it is the exact overclaim the grounding audit flags.
$voice$
      ), false),
    updated_at = NOW()
WHERE type = 'grounded-explainer' AND deleted_at IS NULL
  AND default_config->'workflow'->'steps'->'compose_explainer'->'config'->>'prompt_template' NOT LIKE '%HOUSE VOICE%';

COMMIT;

-- Verify on OUTPUT, not on the prompt: count em dashes in the next draft.
-- The rule being present is not evidence the draft followed it — that is
-- precisely how the 20-em-dash draft happened with rule 3 already live in a
-- different writer.
