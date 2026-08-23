-- 568 — the editquality seat must stop gating on a change ALREADY BEING COMMITTED.
--
-- OWNER RULING 2026-08-23. Raised by bugs_open/309's round 2, where editquality gated
-- the submission at HIGH severity on this ground, verbatim:
--
--   "The plan states 'Committed 747e717a1 / effd08fff / 62f187442 under
--    Council-Submitted: a092d7d8' and verifies deployment facts via `kubectl kustomize`
--    against a live overlay — i.e. it describes work already applied outside this
--    review, not a set of edits awaiting review"
--
-- That is the workflow the owner MANDATED, being scored as a fault. CLAUDE.md, owner
-- ruling 2026-07-29 §2: "review here is after the fact, by design … Do not claim an
-- ordering constraint you do not have; do not pretend you could have waited." And the
-- `Council-Submitted:` trailer exists precisely so a commit can precede its verdict and
-- be credited automatically when the verdict lands (no amend; forward-only forbids one).
--
-- WHY A PROMPT CHANGE AND NOT A DECISION-RULE CHANGE. The seat is not malfunctioning —
-- its brief tells it it is reviewing "a proposed fix plan", so already-applied work
-- genuinely reads as out of contract. The defect is in what it was told, so that is what
-- moves. The decision rule stays untouched: a real editquality objection must still be
-- able to gate.
--
-- ⚠ WHY THIS IS AN UPDATE AND NOT THE 099 MIRROR. `099_SYNC_gate_roster.py --apply` is
-- SUSPENDED (CLAUDE.md): it regenerates all 17 gate prompts and its transform predates
-- migration 377, so it would revert the hoisted shared-evidence block and the
-- <!--CACHE_BREAKPOINT--> that makes prefix caching fire — a measured 68% saving on what
-- was ~85% of fleet LLM spend. The documented workaround is the 381+383 pair: migrate
-- fix-proposer, then mirror into council-gate with a surgical update anchored on a
-- verbatim line and guarded. This file does both rosters in one transaction because the
-- anchor and the replacement are byte-identical for both.
--
-- ⚠ THE INSERTION IS AFTER THE CACHE BREAKPOINT, DELIBERATELY. Measured 2026-08-23: the
-- breakpoint sits at line 9 of the editquality prompt and the anchor at line 16, so the
-- SHARED CACHED PREFIX IS UNTOUCHED and 377's saving is preserved. The verify block
-- asserts that explicitly — it aborts if the 17-seat / 1-prefix invariant breaks.
--
-- ANCHOR SAFETY, measured before writing this: the anchor line occurs exactly ONCE in
-- each roster, and in exactly ONE seat per roster (council-gate 1, fix-proposer 1), so
-- it cannot splash onto a sibling seat.
--
-- APPLY BY HAND (the runner takes EVERY pending file and 560-567 are other sessions'):
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
--     < docs/agent_docs/sql_for_agents/568_editquality_already_committed_is_not_an_objection.sql
--   then record it:
--   scripts/migration/run-migrations.sh --record-only 568_editquality_already_committed_is_not_an_objection.sql

BEGIN;

-- Backup, per the estate's convention for a live-config rewrite.
CREATE TABLE IF NOT EXISTS bak_568_editquality_20260823 AS
SELECT id, type, default_config, now() AS backed_up_at
FROM agent_definitions
WHERE type IN ('council-gate','fix-proposer') AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ── PRE-STATE GUARD ───────────────────────────────────────────────────────────
-- Gate on the WHOLE pre-state, not just the tail (316's lesson): the anchor must be
-- present exactly once per roster, and the amendment must not already be there.
DO $guard$
DECLARE
  n_anchor int;
  n_already int;
BEGIN
  SELECT count(*) INTO n_anchor
  FROM agent_definitions
  WHERE type IN ('council-gate','fix-proposer') AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
    AND default_config->'workflow'->'steps'->'review_editquality'->'config'->>'prompt_template'
        LIKE '%Verdicts: approve (sound), object (fixable problems — list them), veto (fundamentally wrong: fixes a different bug, or all edits are no-ops).%';
  IF n_anchor <> 2 THEN
    RAISE EXCEPTION '568 ABORT: expected the editquality anchor in exactly 2 rosters, found %. The prompt has drifted — re-read both seats before editing.', n_anchor;
  END IF;

  SELECT count(*) INTO n_already
  FROM agent_definitions
  WHERE type IN ('council-gate','fix-proposer') AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
    AND default_config->'workflow'->'steps'->'review_editquality'->'config'->>'prompt_template'
        LIKE '%ALREADY COMMITTED IS NORMAL%';
  IF n_already <> 0 THEN
    RAISE EXCEPTION '568 ABORT: the amendment is already present in % roster(s) — this migration is not idempotent by design, so re-applying would double the paragraph.', n_already;
  END IF;
END
$guard$;

-- ── THE CHANGE ────────────────────────────────────────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,review_editquality,config,prompt_template}',
      to_jsonb(replace(
        default_config->'workflow'->'steps'->'review_editquality'->'config'->>'prompt_template',
        'Verdicts: approve (sound), object (fixable problems — list them), veto (fundamentally wrong: fixes a different bug, or all edits are no-ops).',
        'Verdicts: approve (sound), object (fixable problems — list them), veto (fundamentally wrong: fixes a different bug, or all edits are no-ops).'
        || E'\n\n'
        || 'ALREADY COMMITTED IS NORMAL, AND IS NEVER AN OBJECTION. On this tree there is one shared branch, and any session''s build ships whatever is at HEAD — so a change CANNOT be held back pending your verdict, and the owner ruled (2026-07-29) that review here is AFTER THE FACT by design. The `Council-Submitted:` commit trailer exists precisely so a commit can precede its verdict and be credited when it lands. A submission that says its edits are committed, cites shas, or reports facts verified against live state or a running cluster is FOLLOWING that rule, not evading you. Judge the CHANGE on its merits exactly as you would an unapplied proposal. Do NOT object that the work "is already applied", that it is "not a set of edits awaiting review", or that it was verified outside the review — and never let any of those reduce a verdict or contribute to a gate. Your objections still land: forward-only means the remedy is another commit, never a revert, so an objection on the SUBSTANCE changes what happens next just as much as it would have before the commit.'
      )),
      false)
WHERE type IN ('council-gate','fix-proposer') AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config->'workflow'->'steps'->'review_editquality'->'config'->>'prompt_template'
      LIKE '%Verdicts: approve (sound), object (fixable problems%';

-- ── VERIFY (DO/RAISE, never a bare SELECT) ────────────────────────────────────
-- A verify block made of SELECTs cannot stop the COMMIT: ON_ERROR_STOP ignores a
-- non-empty result set. RFC_006's lesson, and this estate has been bitten by it.
DO $verify$
DECLARE
  n_amended int;
  n_prefix_distinct int;
  n_marked int;
BEGIN
  SELECT count(*) INTO n_amended
  FROM agent_definitions
  WHERE type IN ('council-gate','fix-proposer') AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
    AND default_config->'workflow'->'steps'->'review_editquality'->'config'->>'prompt_template'
        LIKE '%ALREADY COMMITTED IS NORMAL%';
  IF n_amended <> 2 THEN
    RAISE EXCEPTION '568 VERIFY FAILED: amendment present in % roster(s), expected 2.', n_amended;
  END IF;

  -- 377's prefix-caching invariant must survive: 17 seats marked, ONE distinct prefix.
  SELECT count(*), count(DISTINCT split_part(p,'<!--CACHE_BREAKPOINT-->',1))
    INTO n_marked, n_prefix_distinct
  FROM (
    SELECT v->'config'->>'prompt_template' AS p
    FROM agent_definitions, LATERAL jsonb_each(default_config->'workflow'->'steps') AS e(k,v)
    WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false
      AND deleted_at IS NULL AND v->'config'->>'prompt_template' LIKE '%<!--CACHE_BREAKPOINT-->%'
  ) s;
  IF n_marked <> 17 OR n_prefix_distinct <> 1 THEN
    RAISE EXCEPTION '568 VERIFY FAILED: prefix-cache health is % seats / % distinct prefixes, expected 17 / 1 — migration 377''s shared prefix has been fragmented.', n_marked, n_prefix_distinct;
  END IF;

  RAISE NOTICE '568 OK: editquality amended in 2 rosters; cache prefix intact (17 seats, 1 prefix).';
END
$verify$;

COMMIT;
