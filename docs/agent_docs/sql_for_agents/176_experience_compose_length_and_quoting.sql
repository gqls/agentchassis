-- 176_experience_compose_length_and_quoting.sql — fix a regression I caused.
--
-- WHAT HAPPENED (run 20a1581b, 2026-07-19). `compose` failed:
--   "response truncated: stop_reason=max_tokens (output_tokens=32000 reached
--    the configured cap)"
-- The run terminated at complete_refused with ZERO council reports.
--
-- CAUSE — mine, not the model's. Migrations 174 + 175 grew load_context from
-- ~13 KB to ~39 KB by adding two blocks of real JavaScript source (js_snippets
-- loaders + component-owned js_content). `compose` renders that context inline
-- and had NO length rule and NO quoting rule: LENGTH DISCIPLINE was added to
-- `recompose` only, by the run-6 truncation fix, because with a small context
-- compose had never needed one. Given 39 KB of source to look at, it wrote a
-- plan that ran past the 32000-token output cap.
--
-- WHY IT SURFACED NOW RATHER THAN SILENTLY: the fresh chassis build v1.0.1138
-- decodes `stop_reason` and hard-errors on a capped completion
-- (`aiservice/anthropic.go:180`). On the previous image the same overrun would
-- have returned a TRUNCATED plan that persisted and then drew truncation
-- objections nobody could clear — the exact death spiral of run 6. So the new
-- image converted a silent corruption into a loud failure. That is the fix
-- working, and it caught my regression on its first run.
--
-- FIX: give `compose` the discipline `recompose` already has, plus a quoting
-- rule aimed specifically at the new context. NOT raising max_tokens: runs 8
-- and 9 both produced complete, approved-quality plans at ~14 KB, so 32000 is
-- ample. The plan overran because nothing told it not to, and raising the cap
-- would just buy a bigger truncation later — the same reasoning that made the
-- run-6 fix pair a raised ceiling WITH a discipline rule rather than trusting
-- the ceiling alone.
--
-- Deliberately NOT shrinking the 8000-char source cap in 174/175: the guards
-- that matter (`if (!data.today) return;`, `Array.isArray(data.archive.entries)`,
-- `Array.isArray(a.cards)`) are what the contract critic quotes, and trimming
-- risks cutting exactly the evidence the seat exists to read.
--
-- Config-only: live on commit, no image roll. Seed 167 synced in-place.

BEGIN;

SELECT snapshot_agent('experience-planner', 'pre-update: 176 compose length + quoting discipline')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='experience-planner' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,compose,config,prompt_template}',
         to_jsonb(replace(
           default_config->'workflow'->'steps'->'compose'->'config'->>'prompt_template',
           '## Output format (IMPORTANT)',
           '## LENGTH AND QUOTING DISCIPLINE (a run has already died here)' || chr(10) ||
           'The site context above now includes REAL JAVASCRIPT SOURCE for the runtime loaders and component-owned scripts. That source is REFERENCE MATERIAL FOR YOU TO READ. It must NEVER be reproduced in the plan. Do not paste loader bodies, function definitions, or long code blocks into any section. When a data field or selector matters, name it and give the one-line access path (e.g. `data.archive.entries[]`, `.gauntlet-interface__timer`) — never the surrounding code.' || chr(10) || chr(10) ||
           'The whole plan must come in WELL UNDER the output cap: aim for about 14,000 characters, and never exceed 20,000. A plan that hits the cap is DESTROYED, not shortened — the run fails outright with stop_reason=max_tokens and produces nothing. Approved plans for this experience have been ~14,000 characters, so this is comfortable, not tight.' || chr(10) || chr(10) ||
           'Priority if you are running long: the ```criteria fence in §5 has ABSOLUTE priority and must always be complete, valid, closed JSON followed by the END trailer. Cut narrative, prose rationale and LATER-list detail first. A terse plan is fine; a cut-off plan is worthless.' || chr(10) || chr(10) ||
           '## Output format (IMPORTANT)'
         )),
         false)
 WHERE type='experience-planner'
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE t text;
BEGIN
  SELECT default_config->'workflow'->'steps'->'compose'->'config'->>'prompt_template' INTO t
    FROM agent_definitions
   WHERE type='experience-planner'
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF position('LENGTH AND QUOTING DISCIPLINE' in t) = 0 THEN
    RAISE EXCEPTION 'compose discipline rule not inserted';
  END IF;
  IF position('## Output format (IMPORTANT)' in t) = 0 THEN
    RAISE EXCEPTION 'compose output-format section lost';
  END IF;
  IF position('END EXPERIENCE_PLAN' in t) = 0 THEN
    RAISE EXCEPTION 'compose output contract lost';
  END IF;
END $$;

COMMIT;

-- Rollback: restore the snapshot taken above.
