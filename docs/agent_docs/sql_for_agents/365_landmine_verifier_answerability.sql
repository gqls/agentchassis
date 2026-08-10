-- 365_landmine_verifier_answerability.sql
--
-- bugs_open/223, the CONFIG half. The Go half (chassis, this lane's commit) makes
-- the code index state what it CANNOT represent; this seed stops the verifier
-- converting that blindness into a verdict against the corpus.
--
-- WHY A BRANCH AND NOT BETTER PROMPT WORDING. The bug's own correction settles it:
-- given identical 0-row input, three of four recorded verdicts abstained correctly
-- and the fourth declared three of this repo's scripts non-existent — in a note
-- delivered BY those scripts. Improving the average changes nothing, because the
-- reader cannot tell which run they are holding. So the STALE-bearing prompt is
-- made UNREACHABLE for a round that confirmed nothing against indexed code:
--
--   run_checks --> gate_evidence (conditional_branch on lookup.no_code_evidence)
--                    |-- true  --> verify_unverifiable   (no STALE in its vocabulary)
--                    `-- false --> verify               (existing step, prompt amended)
--
-- and, whichever branch ran, persist_verdict now appends the action's own
-- mechanically-composed evidence census to the stored body, so a future reader of
-- the row can always see what was and was not checkable.
--
-- SAFE BEFORE THE IMAGE ROLLS — verified, not assumed (2026-08-10):
--   * conditional_branch resolves an absent field to nil (resolveFieldValue), and
--     compareValues(nil,"true") is false, so on the CURRENT binary the gate takes
--     else_step and the pipeline behaves exactly as it does today. No error path.
--   * append_doc_note declares neither ConfigKeys nor CheckConfig nor StrictConfig,
--     so it is not opted into unknown-config-key detection: note_body_suffix_field
--     is inert on the old binary, not even a warning.
-- The degraded window is therefore "verdicts carry no evidence suffix and no
-- gate until the roll", which is today's behaviour, stated rather than hidden.
-- No ordering constraint is claimed (owner ruling 2026-07-29, condition 1 retired).
--
-- ROLLBACK: 365_landmine_verifier_answerability_ROLLBACK.sql restores the
-- three-key delta only (start_step and every other step are untouched here).

BEGIN;

-- Snapshot first: the roster rule of this estate is that a live agent_definitions
-- row is the fact and the seed file is only history, so the pre-change row must be
-- recoverable independently of this file.
INSERT INTO agent_definitions (type, name, description, default_config, version, is_active, is_snapshot, created_at)
SELECT type, name || ' (pre-365 snapshot)', 'Snapshot before bugs_open/223 answerability gate', default_config,
       version, false, true, now()
  FROM agent_definitions
 WHERE type = 'landmine-verifier' AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions SET default_config = jsonb_set(
  jsonb_set(
    jsonb_set(
      jsonb_set(default_config,

        -- 1. run_checks now hands off to the gate rather than straight to verify.
        '{workflow,steps,run_checks,next_step}', '"gate_evidence"'::jsonb),

      -- 2. The gate itself. `no_code_evidence` is the action's own boolean:
      --    true when NOT ONE check in the round matched a code-tier row —
      --    unanswerable checks and honest misses alike. A bool is used
      --    deliberately: compareValues handles bool robustly and numerics less so.
      '{workflow,steps,gate_evidence}', '{
        "action": "conditional_branch",
        "config": {
          "condition": "lookup.no_code_evidence == true",
          "then_step": "verify_unverifiable",
          "else_step": "verify"
        },
        "output_field": "evidence_gate",
        "description": "bugs_open/223: a round that confirmed nothing against indexed code cannot reach the STALE-bearing verdict prompt"
      }'::jsonb),

    -- 3. The no-evidence branch. STALE is ABSENT from its vocabulary, and so is
    --    STILL_VALID: the bug records that with zero checkable evidence BOTH
    --    directions are uninformative about the footprint, and a STILL_VALID
    --    earned by prose reasoning alone has been mistaken for verification.
    --    UNVERIFIABLE says what actually happened — the wrong question was asked
    --    of this index — which is not the same as "the entry is wrong".
    '{workflow,steps,verify_unverifiable}', '{
      "action": "execute_llm_prompt",
      "config": {
        "ai_service": {
          "model": "claude-opus-4-6",
          "provider": "anthropic",
          "max_tokens": 1200,
          "api_key_env_var": "ANTHROPIC_API_KEY"
        },
        "input_fields": ["entry", "lookup", "input_data"],
        "output_format": "json",
        "prompt_template": "You are recording that a LANDMINES.md entry COULD NOT be mechanically verified, and why.\n\nThe entry:\n\n{{.entry.body}}\n\nThe code-lookup results. Read them: not one check in this round matched indexed code, and the lookup itself states which of them could never have matched, because the index it searches does not hold that class of thing (it indexes Go symbols only, and not every kind of Go declaration).\n\n{{.lookup.results_text}}\n\nThis is a bounded task with two permitted outcomes and STALE is NOT one of them. An index that cannot represent a footprint tells you NOTHING about whether that footprint exists, was removed, was renamed or was inlined. Do not infer a rename. Do not infer removal. Do not report the entry as stale, obsolete or unfounded on this evidence, because there is no evidence here in either direction.\n\nReturn ONLY this JSON shape:\n{\n  \"status\": \"UNVERIFIABLE | NEEDS_HUMAN_REVIEW\",\n  \"rationale\": \"one paragraph: which footprint items could not be checked and why (quote the lookup lines), plus anything you can say about the entry TEXT alone — is it internally consistent, does its footprint match its fires-when clause\",\n  \"body\": \"a short markdown note, formatted as: **last verified (landmine-verifier): <STATUS>.** <one-sentence rationale>. Checked against {{.input_data.ref}}.\"\n}\n\nUNVERIFIABLE means the wrong question was asked of this index — the entry stands unexamined, and a human or a non-index check is needed. NEEDS_HUMAN_REVIEW means the entry TEXT itself is unclear or self-contradictory, which is a judgement you CAN make from the entry alone. Never state or imply that a footprint is absent from the system."
      },
      "output_field": "verdict",
      "next_step": "persist_verdict",
      "description": "bugs_open/223: the no-evidence branch — UNVERIFIABLE or NEEDS_HUMAN_REVIEW, never STALE"
    }'::jsonb),

  -- 4. The persist step appends the action's OWN census to whatever the model
  --    wrote. Mechanical, so it cannot be softened or omitted by the model whose
  --    verdict it qualifies. Both branches write output_field "verdict", so this
  --    one step serves both.
  '{workflow,steps,persist_verdict,config,note_body_suffix_field}', '"lookup.evidence_line"'::jsonb)
WHERE type = 'landmine-verifier' AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- 5. The EVIDENCE branch's prompt is amended in place: it keeps its full
--    vocabulary (a STALE backed by rows that contradict the entry is exactly what
--    this mechanism is for) but may no longer rest one on an unanswerable check.
--    This is the MIXED-footprint case, which the bug names as the dangerous one
--    because a partial confirmation reads as diligence.
UPDATE agent_definitions SET default_config = jsonb_set(
  default_config,
  '{workflow,steps,verify,config,prompt_template}',
  to_jsonb(
    (default_config #>> '{workflow,steps,verify,config,prompt_template}')
    || E'\n\nEVIDENCE RULES (bugs_open/223 — these bound what your verdict may rest on).\n'
    || E'1. A check line marked "NOT ANSWERABLE BY THIS INDEX" is evidence in NEITHER direction. The query ran and could not have matched. It does not support STALE, and it does not support STILL_VALID either.\n'
    || E'2. STALE may rest ONLY on a check that returned indexed rows contradicting the entry, or on a 0-row answer the lookup itself rendered as in-scope. Quote the line you are relying on.\n'
    || E'3. Never assert that a symbol was renamed or inlined without a row showing where it went. If the lookup says the index holds no declarations of that kind, the absence is a property of the INDEX and you must say so.\n'
    || E'4. Some footprint items resolve and others cannot be checked: that is the common case, not an edge case. Confirm what you confirmed, and name what you could not check, in the same verdict. A partly-checked entry is not a verified entry.\n'
  )
)
WHERE type = 'landmine-verifier' AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Verify INSIDE the transaction, and RAISE rather than SELECT: a verify block of
-- bare SELECTs cannot stop the COMMIT (ON_ERROR_STOP ignores a non-empty result),
-- which is a trap this estate has already paid for once (RFC_006).
DO $$
DECLARE
  cfg jsonb;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type = 'landmine-verifier' AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF cfg #>> '{workflow,steps,run_checks,next_step}' <> 'gate_evidence' THEN
    RAISE EXCEPTION 'run_checks does not hand off to gate_evidence (got %)', cfg #>> '{workflow,steps,run_checks,next_step}';
  END IF;
  IF cfg #>> '{workflow,steps,gate_evidence,action}' <> 'conditional_branch' THEN
    RAISE EXCEPTION 'gate_evidence is not a conditional_branch';
  END IF;
  IF cfg #>> '{workflow,steps,gate_evidence,config,then_step}' <> 'verify_unverifiable'
     OR cfg #>> '{workflow,steps,gate_evidence,config,else_step}' <> 'verify' THEN
    RAISE EXCEPTION 'gate_evidence branches are not wired to (verify_unverifiable, verify)';
  END IF;
  IF cfg #>> '{workflow,steps,persist_verdict,config,note_body_suffix_field}' <> 'lookup.evidence_line' THEN
    RAISE EXCEPTION 'persist_verdict does not append the evidence line';
  END IF;
  -- The load-bearing assertion: STALE must be UNREACHABLE on the no-evidence
  -- branch. Checking the topology without checking this would pass a seed whose
  -- new prompt still offered the verdict the branch exists to remove.
  IF cfg #>> '{workflow,steps,verify_unverifiable,config,prompt_template}' LIKE '%STALE |%'
     OR cfg #>> '{workflow,steps,verify_unverifiable,config,prompt_template}' LIKE '%| STALE%' THEN
    RAISE EXCEPTION 'the no-evidence branch still offers STALE in its status vocabulary';
  END IF;
  IF cfg #>> '{workflow,steps,verify,config,prompt_template}' NOT LIKE '%NOT ANSWERABLE BY THIS INDEX%' THEN
    RAISE EXCEPTION 'the evidence branch was not told what an unanswerable check means';
  END IF;
  -- And the amendment must be idempotent-safe to REVIEW: appending twice would be
  -- harmless but noisy, so fail loudly if this seed has already run.
  IF (length(cfg #>> '{workflow,steps,verify,config,prompt_template}')
      - length(replace(cfg #>> '{workflow,steps,verify,config,prompt_template}', 'EVIDENCE RULES', ''))) / 14 > 1 THEN
    RAISE EXCEPTION 'EVIDENCE RULES appear more than once — this seed has already been applied';
  END IF;

  RAISE NOTICE 'bugs_open/223 gate wired: run_checks -> gate_evidence -> {verify_unverifiable | verify}, suffix on persist_verdict';
END $$;

COMMIT;
