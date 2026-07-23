-- 196_experience_contracts_greenfield_split.sql
--
-- Unsticks the experience loop. Runs 9/10/11 all escalated (5-seat council,
-- max_rounds exhausted, never converged) because the `review_contracts` seat
-- (migration 174) applies one rule to two different cases:
--
--   "You may NOT approve a pair by reasoning about what a name implies. Quote
--    the consumer's own source from the context below. And if the source
--    needed to settle a pair is NOT in your context, that is itself an
--    OBJECTION..."
--
-- That is correct when the plan's claim contradicts EXISTING code (run 9 caught
-- a real §3<->loader mismatch this way — worth keeping). It is wrong when the
-- "consumer" is code the plan itself proposes to CREATE: there is no source to
-- quote because the code does not exist yet, which is what a plan IS. Run 11's
-- own words: "Step 2 creates a new provocations-archive-detail hydrator ...
-- this consumer's source is not in context and does not exist yet" / "Step 1
-- adds a new fetch ... cannot be checked against any real source" — both true,
-- neither a defect. See docs/agent_docs/docs024_key_docs_latest/experience_loop/
-- RUNNING_NOTES_experience_loop.md, "2026-07-19 — runs 9/10/11" entry, for the
-- full evidence and the rule as the owner approved it verbatim (2026-07-23,
-- vonc gauntlet AI-competitor workstream).
--
-- Fix: split the rule in the prompt only (config-only; no image; no other
-- seat touched). Everything else about the seat (its 4 judged pairs, its
-- verdict shape, its non-veto-but-blocking status, council wiring, recompose
-- visibility) is unchanged and re-asserted below.
--
-- ORDER: none — config-only, live immediately, no image dependency.
-- ROLLBACK: snapshot_agent below; restore via UPDATE ... default_config =
-- (SELECT default_config FROM agent_definitions WHERE is_snapshot AND
-- default_config->>'_snapshot_of' = 'experience-planner' ORDER BY created_at
-- DESC LIMIT 1) WHERE type='experience-planner' AND NOT is_snapshot.

BEGIN;

SELECT snapshot_agent('experience-planner', '196_experience_contracts_greenfield_split: pre-update');

-- ── replace ONLY the strictness paragraph inside review_contracts' prompt_template ──
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,review_contracts,config,prompt_template}',
         to_jsonb(replace(
           default_config->'workflow'->'steps'->'review_contracts'->'config'->>'prompt_template',
           '## THE RULE THAT MAKES THIS SEAT WORTH ITS COST' || chr(10) ||
           'You may NOT approve a pair by reasoning about what a name implies. Quote the consumer''s own source from the context below. And if the source needed to settle a pair is NOT in your context, that is itself an OBJECTION — say exactly which source you needed and could not see. Never approve an unverifiable pair; an unseen consumer is how a mismatched contract reaches production.',
           '## THE RULE THAT MAKES THIS SEAT WORTH ITS COST — split by whether the consumer EXISTS' || chr(10) ||
           'First decide, for each pair, whether the consumer is EXISTING code (already in the context below) or NEW code this plan proposes to create or change.' || chr(10) ||
           'EXISTING consumer: you may NOT approve by reasoning about what a name implies. Quote the consumer''s own source from the context below. If the source needed to settle the pair is NOT in your context, that absence is itself an OBJECTION — say exactly which source you needed and could not see. Never approve an unverifiable pair against existing code; an unseen consumer is how a mismatched contract reaches production.' || chr(10) ||
           'NEW consumer (the plan itself creates or changes it): there is no source to quote — that is expected, not a defect, and is NOT grounds for an objection by itself. Approve the pair ONLY IF BOTH hold: (1) the plan states the EXACT access path the new code must implement (the literal field names / selector / endpoint shape it will read or call — not just "it will fetch the data"), AND (2) the plan''s §5 criteria fence contains an acceptance criterion that would FAIL if that access path is never implemented as stated. If either is missing — the access path is vague, or no criterion would catch its absence — THAT is the objection: say precisely which of the two is missing. Do not object merely because the code does not exist yet.'
         )),
         true)
 WHERE type='experience-planner'
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ── Assert: the split landed, nothing else moved ────────────────────────────
DO $$
DECLARE
  w jsonb; rp text;
BEGIN
  SELECT default_config->'workflow' INTO w
    FROM agent_definitions
   WHERE type='experience-planner'
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  rp := w->'steps'->'review_contracts'->'config'->>'prompt_template';

  IF position('split by whether the consumer EXISTS' in rp) = 0 THEN
    RAISE EXCEPTION '196: greenfield split did not land in review_contracts prompt';
  END IF;
  IF position('EXACT access path' in rp) = 0 THEN
    RAISE EXCEPTION '196: greenfield allow-condition missing';
  END IF;
  IF position('acceptance criterion that would FAIL' in rp) = 0 THEN
    RAISE EXCEPTION '196: greenfield criterion condition missing';
  END IF;
  IF position('Never approve an unverifiable pair against existing code' in rp) = 0 THEN
    RAISE EXCEPTION '196: existing-consumer strictness was lost, not just extended';
  END IF;
  -- unchanged invariants from 174 must still hold
  IF w->'steps'->'review_contracts' IS NULL THEN
    RAISE EXCEPTION '196: review_contracts seat missing (unexpected — 174 should have applied first)';
  END IF;
  IF w->'steps'->'review_mvp'->>'next_step' <> 'review_contracts' THEN
    RAISE EXCEPTION '196: review_mvp no longer chains into review_contracts';
  END IF;
  IF w->'steps'->'review_contracts'->>'next_step' <> 'council_decide' THEN
    RAISE EXCEPTION '196: review_contracts no longer chains into council_decide';
  END IF;
  IF NOT (w->'steps'->'council_decide'->'config'->'review_fields' @> '["review_contracts.result"]'::jsonb) THEN
    RAISE EXCEPTION '196: council_decide stopped counting review_contracts';
  END IF;
  IF position('{{.review_contracts}}' in (w->'steps'->'recompose'->'config'->>'prompt_template')) = 0 THEN
    RAISE EXCEPTION '196: recompose can no longer see review_contracts output';
  END IF;
  IF w->'steps'->'review_honesty'->'config'->>'error_step' <> 'complete_refused' THEN
    RAISE EXCEPTION '196: review_honesty lost its fail-closed error_step (171) — unrelated regression';
  END IF;
END $$;

COMMIT;
