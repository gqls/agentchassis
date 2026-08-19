-- 481 — tool-generator: promote six recurring tool defects from per-brief prose into the
--       generator's own quality contract (rules 15-20)
--       (OWNER RULING 2026-08-19: "all fixes should extend into the framework so the
--        problems don't recur, and we have checkers and handlers that can find and fix
--        the problems")
--
-- CONFIG ONLY. No Go half, no roll required. Live for the next generation.
--
-- ============================================================================
-- WHAT THE PROBLEM IS, in plain terms
-- ============================================================================
-- The `webdesign_tool_rebuilds` lane has rebuilt 8 of webdesign.co.uk's 63 imported
-- tools. Five of the nine ported tools it read were measurably broken in production,
-- and the lane fixed each one by writing a requirement into THAT tool's work item
-- description. That fixes one tool. The next tool the generator builds, for any site,
-- by any lane, is born with the same defect - because the requirement lived in one
-- session's prose and not in the platform.
--
-- The cost of that is not hypothetical. The owner reported a copy button that said
-- "Copied" when nothing had been copied, on a tool THIS LANE had built and graded PASS
-- hours earlier. The rule ("report success only when the copy actually succeeded") was
-- in the author's head; the generator had never been told it.
--
-- ============================================================================
-- WHAT THIS MIGRATION DOES
-- ============================================================================
-- Appends six rules to `generate_tool_html.prompt_template`, continuing the existing
-- numbered list (which ends at 14). Each one is generalised from a defect MEASURED in a
-- live tool on webdesign.co.uk:
--
--   15  unverified success   <- copy buttons calling showCopied() unconditionally, with
--                               document.execCommand's boolean return discarded and the
--                               exception swallowed (html-minifier, svg-optimizer)
--   16  inline handlers      <- onclick="cleanJson()" style binding, which also requires
--                               the function to be global (42 of the 55 remaining ported
--                               tools carry this shape)
--   17  blocking dialogs     <- alert("Copied!") (25 of 55)
--   18  unvalidated numbers  <- parseInt() of a blank field yielding NaN, so every
--                               comparison is false and the tool silently does nothing,
--                               which is indistinguishable from success (json-cleaner)
--   19  errors destroy work  <- writing a parse error INTO the output textarea, losing
--                               the result the user was about to copy (json-cleaner)
--   20  invisible no-op      <- a transformer that legitimately changes almost nothing
--                               looking identical to one that is broken (html-minifier:
--                               2.9% on its own page, because 71% of it is inside script
--                               and style elements it must not touch)
--
-- Rules 18 and 20 are the same underlying failure at two levels: a tool that does
-- nothing must never look like a tool that succeeded.
--
-- ============================================================================
-- WHY THE PROMPT AND NOT A CHECKER (both are wanted; this is the cheap half)
-- ============================================================================
-- A prompt rule binds what is generated NEXT, costs nothing per tool, and is live
-- immediately with no image. It cannot see what is already deployed - that is Track 2,
-- a discovery check in platform/orchestration/actions/discovery_checks/, which needs the
-- council gate and a roll. This migration is deliberately the half that can ship today.
--
-- ============================================================================
-- SAFETY
-- ============================================================================
-- Idempotent: the UPDATE is guarded on the rule-15 marker being ABSENT, so re-running
-- appends nothing. A verify block RAISEs if the marker is not present afterwards, or if
-- the template shrank. Rollback: 481_..._ROLLBACK.sql restores the pre-change template
-- from the snapshot this migration takes.
-- ============================================================================

BEGIN;

-- snapshot the pre-change template so the rollback has a true pre-image
CREATE TABLE IF NOT EXISTS mig_481_tool_generator_prompt_backup (
  taken_at        timestamptz NOT NULL DEFAULT now(),
  agent_id        uuid        NOT NULL,
  prompt_template text        NOT NULL
);

INSERT INTO mig_481_tool_generator_prompt_backup (agent_id, prompt_template)
SELECT id, default_config#>>'{workflow,steps,generate_tool_html,config,prompt_template}'
FROM agent_definitions
WHERE type='tool-generator' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config#>>'{workflow,steps,generate_tool_html,config,prompt_template}' NOT LIKE '%15. Never report success%';

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,generate_tool_html,config,prompt_template}',
      to_jsonb(
        replace(
          default_config#>>'{workflow,steps,generate_tool_html,config,prompt_template}',
          E'\n\n## Structure',
          E'\n15. Never report success you have not verified. If an action can fail (a clipboard write, a download, a permission), check its actual result: await the promise and handle rejection, and use document.execCommand''s boolean return value. Show a distinct failure state that tells the user what to do instead. Never show a success label when the action did not succeed.'
          || E'\n16. Bind every behaviour with addEventListener on the element itself. Never use an inline onclick, oninput or onchange attribute, and never depend on a function being global.'
          || E'\n17. Never use alert(), confirm() or prompt(). Show messages inline, next to the control they concern, and clear them when the condition passes.'
          || E'\n18. Validate every value you parse before using it. A blank, non-numeric or out-of-range entry must produce a short visible message. It must NEVER silently produce no change: a tool that quietly does nothing is indistinguishable from one that worked, which is the worst failure a tool can have.'
          || E'\n19. An error must never destroy the user''s work. Show it in its own area and leave any existing output exactly where it is, then clear it on the next successful run. Never write an error message into the output field.'
          || E'\n20. If the tool transforms input into output, always show both sizes and the change between them (for example characters in, characters out, percent saved). Some inputs legitimately change very little, and without a readout the user cannot tell that from a tool that is broken.'
          || E'\n\n## Structure'
        )
      ),
      false
    ),
    updated_at = now()
WHERE type='tool-generator' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config#>>'{workflow,steps,generate_tool_html,config,prompt_template}' NOT LIKE '%15. Never report success%';

-- verify: a DO block, because a bare SELECT cannot stop the COMMIT (ON_ERROR_STOP
-- ignores a non-empty result set) - see LANDMINES on migration verify blocks
DO $$
DECLARE
  tmpl text;
  before_len int;
BEGIN
  SELECT default_config#>>'{workflow,steps,generate_tool_html,config,prompt_template}'
    INTO tmpl
  FROM agent_definitions
  WHERE type='tool-generator' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF tmpl IS NULL THEN
    RAISE EXCEPTION '481: tool-generator prompt_template is NULL - wrong path or wrong agent';
  END IF;
  IF tmpl NOT LIKE '%15. Never report success%' THEN
    RAISE EXCEPTION '481: rule 15 absent after update - the ## Structure anchor did not match';
  END IF;
  IF tmpl NOT LIKE '%20. If the tool transforms input%' THEN
    RAISE EXCEPTION '481: rule 20 absent after update';
  END IF;
  IF tmpl NOT LIKE '%## Structure%' THEN
    RAISE EXCEPTION '481: the ## Structure section was consumed rather than preserved';
  END IF;
  IF tmpl LIKE '%15. Never report success%15. Never report success%' THEN
    RAISE EXCEPTION '481: rules appended twice - guard failed';
  END IF;

  SELECT length(prompt_template) INTO before_len
  FROM mig_481_tool_generator_prompt_backup ORDER BY taken_at DESC LIMIT 1;
  IF before_len IS NOT NULL AND length(tmpl) <= before_len THEN
    RAISE EXCEPTION '481: template did not grow (% -> %)', before_len, length(tmpl);
  END IF;

  RAISE NOTICE '481 OK: prompt_template % -> % chars, rules 15-20 present', before_len, length(tmpl);
END $$;

COMMIT;
