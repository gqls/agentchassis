-- 741_refuse_unknown_rerender_routing_key_HOLD.sql
--
-- RFC_062 PHASE 3 — THE FLIP. bugs_open/440: "a routing key nobody understands
-- completes green". This is the migration that ends that, and it is the one part of
-- the programme that changes what the shared page-rerender gate GUARANTEES.
--
-- ⚠⚠ _HOLD — DO NOT APPLY WITHOUT THE 404 LANE'S CO-SIGN (OWNER RULING D2). ⚠⚠
-- The livespec Declarations this migration's shape is asserted by belong to
-- bugs_open/404's lane, and D2 says they co-sign rather than being told afterwards.
-- Owner decision 2026-09-03: build it all, hold the apply. The release condition is the
-- co-sign, nothing else. Council round on this file: APPROVED r1, corr 56047b18.
-- ⚠ CORRECTED 2026-09-03 (night): this header used to say the 404 lane "has been dormant
-- since 2026-08-26". FALSE — their last own commit is 281c08bbe, 2026-09-02 16:24Z, and
-- their r4 verdict is APPROVED (2026-09-02 16:33:30Z) AND STILL UNREAD. Judged from
-- `ls -la` mtimes instead of `git log`; the newest file in their directory was one this
-- lane had written. See the lane handoff item 2 for the nudge that follows from it.
--
-- ── WHAT IT DOES ──────────────────────────────────────────────────────────────
--
-- Today `check_rerender_mode` is ONE conditional with only a then/else: five `==`
-- tests on `spec.reason`, else_step `render_page`, which re-ships the stored HTML and
-- completes GREEN. So an unrecognised routing key cannot be refused — it is
-- indistinguishable from "assemble this page", which is also the estate's safe,
-- cheap, correct default. That is the whole bug.
--
--   1. A NEW START STEP, `check_routing_key_known`, runs AHEAD of the gate. Routing
--      key absent, empty, or in the vocabulary -> on to the gate exactly as today.
--      Present-but-unknown -> the refusal branch. Present-but-unknown is the ONLY
--      refusing state, and after the phase 1a/1b/2 split it has no legitimate
--      counter-example: free prose and the deliberate assemble-only key
--      (`verbatim_adoption_deploy`) live in `spec.reason` and stamp nothing here.
--   2. `refuse_unknown_routing_key` parks the item at `needs_human_review` (OWNER
--      RULING D1 — not a silent assemble, and NOT a blunt orchestration failure), with
--      a message naming the OFFENDING VALUE and the vocabulary. The value comes from
--      fail_work_item's opt-in `error_message_template`; the static `error_message`
--      beside it is the fallback if that template ever fails to render.
--   3. `check_rerender_mode` moves to the TRANSITION condition: every vocabulary value
--      accepted under EITHER key. This is LOAD-BEARING, not a courtesy —
--      `[MEASURED 2026-09-03]` 1,804 pending page_rerender items carry an in-vocabulary
--      `reason` and only 12 carry a `routing_reason`. Narrowing to the new key today
--      would route ~1,792 items to assemble: this bug's own shape, fleet-wide, inside
--      its own fix. Narrowing is a LATER migration, gated on a drain census.
--   4. THE WRITE DOOR (OWNER RULING D3) IS NOT IN THIS FILE — IT IS 742, split out on a
--      council objection (guardian [medium], round 56047b18: the ACCESS EXCLUSIVE lock is
--      table-wide even though the CHECK's predicate is not, so a fleet-wide DDL and a
--      one-pipeline config edit want different windows). APPLY 741 THEN 742, and do not
--      stop after 741 — the read door alone is strictly better than today but it is the
--      weaker half of the pair, and raw-SQL INSERTs are the producer class only the CHECK
--      can reach.
--
-- ⚠ EVERY STRING BELOW IS PASTED FROM platform/livespec, NOT HAND-WRITTEN. DB config
-- cannot import Go, so the live objects are ASSERTED against the Go list by the daily
-- drift auditor. Regenerate, never retype:
--   livespec.CheckRoutingKnownConditionClause()          -> check_routing_key_known.condition
--   livespec.TransitionRerenderModeConditionClause()     -> check_rerender_mode.condition
--   livespec.RefuseUnknownRoutingKeyMessageTemplate()    -> error_message_template
--   livespec.RefuseUnknownRoutingKeyMessageFallback()    -> error_message
--   (livespec.RerenderSectionReasonNames() -> the CHECK's IN-list, in 742)
-- Hand-writing a sixth disjunct is what bugs_open/404 IS (two lanes appended a value
-- to this very gate on 2026-08-18 and neither touched Go).
--
-- ⚠ THE `== null` DISJUNCT IN STEP 1 IS LOAD-BEARING AND WAS MISSING FROM THE FIRST
-- CUT. Measured 2026-09-03 by EXECUTING the evaluator (not reading it): compareValues'
-- nil branch runs BEFORE quote-stripping, so a quoted '' never equals nil and a
-- MISSING key does NOT match `== ''`. A clause carrying only `== ''` evaluates FALSE
-- for every item minted before phase 2 — the entire legacy population — and would send
-- the fleet's normal re-render traffic to human review on the day this applied. Four-
-- state table pinned in rerender_routing_gate_clause_test.go against the real evaluator.
--
-- ── ⚠ THE APPLIER OWES ONE MORE COMMIT, IN THE SAME COMMIT ────────────────────
--
-- The daily auditor's own note says: "A finding means the live object and its
-- declaration have parted: fix whichever is wrong, IN THE SAME COMMIT AS THE MIGRATION
-- THAT MOVED IT." Today's row reads `probed 15 live object(s); 0 finding(s)` — a clean
-- signal worth keeping, so the livespec Declaration edits are NOT committed ahead of
-- this apply (they would go red every morning until it landed, masking real drift).
-- Apply this file and commit these together:
--
--   platform/livespec/livespec.go
--     (a) workflow.page-rerender.check_rerender_mode.reasons — the Fragment stays
--         CheckRerenderModeConditionClause() and STILL PASSES: verified by execution
--         2026-09-03, the old five-value clause is a substring of the transition clause
--         exactly once, so Min:1/Max:1 holds. Do not "fix" it.
--     (b) ⚠ ADD a count Declaration for 'input_data.spec.routing_reason ==' with
--         ExpectCount len(RerenderSectionReasons). THIS IS THE REAL GAP: both existing
--         404 Declarations stay GREEN through this flip (the `reason ==` count is still
--         5), so the five NEW routing_reason disjuncts arrive asserted by NOTHING, and a
--         sixth routing value appended to the live gate without touching Go would drift
--         exactly the way 404 drifted. A fragment sees loss and mutation; only a count
--         sees ADDITION.
--     (c) ADD a Declaration for check_routing_key_known.condition (FragmentMatch,
--         CheckRoutingKnownConditionClause(), Min:1 Max:1).
--     (d) ADD a Declaration for the CHECK constraint (Kind 'constraint'), asserting it
--         exists and lists the vocabulary — owed once 742 has applied, not before.
--     (e) BUMP LiveAuditOnlyDeclarations for whichever of the above no Go test reads,
--         and check MaxDeclarations (24) still holds.
--
--
-- ══ 404 LANE CO-SIGN (OWNER RULING D2) — GIVEN 2026-09-04, WITH ONE CONDITION ══
-- Signed by the bugs_open/404 lane itself (their r4 verdict is now read: APPROVED,
-- 2026-09-02 16:33:30Z, artifact e1abb1bc, 3 advisory objections none high; 404 has
-- been moved to bugs_closed/). Applier checklist steps (a) and (b) below were
-- re-derived BY EXECUTION and are confirmed and agreed.
--
-- ⚠ THE CONDITION, AND IT IS A DEFECT IN STEP (c) BELOW: a FragmentMatch on
-- CheckRoutingKnownConditionClause() is BLIND TO ADDITION for exactly the reason (b)
-- is. Mutation-proved 2026-09-04 with both controls: live == declared -> 0 findings;
-- literal_markdown REMOVED live -> 1 finding (so the guard IS armed); a sixth value
-- appended live -> 0 findings, SILENT. So (c) as written would re-create bugs_open/404
-- inside the change built to fix it, one clause along.
--
-- REQUIRED BEFORE APPLY: add a PAIRED CountEqual Declaration on
-- check_routing_key_known, needle 'input_data.spec.routing_reason ==', with
--   ExpectCount: strings.Count(CheckRoutingKnownConditionClause(), "input_data.spec.routing_reason ==")
-- ⚠ That is SEVEN, not five — the clause carries `== null` and `== ''` as well as the
-- five vocabulary values, and both are load-bearing (see the `== null` note above).
-- Derive it from the renderer; do NOT write len(RerenderSectionReasons), and do not
-- write a bare 7. Same mutated input then reports "live count is 8, declared 7".
--
-- Also for the applier, neither a condition: (d) will NOT match if you paste 742's own
-- text — Postgres normalises IN (...) to = ANY (ARRAY[...::text]), verified 2026-09-04
-- against the live doc_plans twin, so declare against pg_get_constraintdef's output.
-- (e) has room: 16 Declarations in the tree today against MaxDeclarations 24; (b), (c),
-- this added count and (d) take it to 20.
--
-- ⚠ CORRECTED 2026-09-04 (same day, same lane): the blindness above is real and mutation-
-- proved, but it CANNOT SHIP UNNOTICED. TestEveryFragmentMatchDeclarationIsGainVisibleOrWaived
-- (platform/livespec/livespec_test.go:364) already refuses any FragmentMatch Declaration with
-- no paired .value_count and no waiver — verified by EXECUTING it against step (c) as
-- prescribed, which turns it red with exactly the right message. So the condition is really:
-- WHEN THAT TEST STOPS YOU, TAKE THE COUNT DOOR, NOT THE WAIVER DOOR. The waiver door is open
-- and accepts 60 characters of prose; check_routing_key_known IS an enumerable vocabulary whose
-- size can grow, so a waiver there would be the wrong answer. The part no test gives you is
-- ExpectCount = 7, not 5.
-- ⚠ Residual risk that survives: platform/livespec has been RED at HEAD for NINE DAYS on
-- another lane's file (TestNoNewMigrationFileReadersOutsideTheAllowList, 405 lane), so that new
-- and correct failure arrives camouflaged among known breakage. Clear it, or run the livespec
-- package tests naming the specific test.
--
-- Full reasoning, evidence and the state of the 404 half at the artefact:
--   docs/agent_docs/docs024_key_docs_latest/bugfix_440_unknown_routing_key/
--     CONTRIB_2026-09-04_from_the_404_lane_cosign_GIVEN_with_one_condition.md
-- ═══════════════════════════════════════════════════════════════════════════════
-- Companions: 741_..._HOLD_ROLLBACK.sql (restores the pre-flip gate; it does NOT touch
-- the constraint — that is 742's own rollback) and 741_..._HOLD_VERIFY.sql, which covers
-- BOTH migrations: the flip's shape, 742's validate-safety census, and the drain.

BEGIN;

-- ⚠ lock_timeout: kept even though the fleet-wide DDL moved to 742. This file now
-- only UPDATEs one agent_definitions row, so the hazard is far smaller — but it is
-- not zero (another session's transaction can hold that row), and a bounded wait is
-- the right default for anything touching a live agent. The measurement that put it
-- here in the first place, and the reason 742 needs it far more, is recorded in 742's
-- own header: the same ADD CONSTRAINT statement measured 2 ms on one dry run and was
-- still waiting after 2 MINUTES on an earlier one. Failing fast is the safe direction
-- either way — one transaction, so a timeout rolls back cleanly and changes nothing.
-- If it fires, RE-RUN IN A QUIET WINDOW; do not raise the timeout to force it through.
SET LOCAL lock_timeout = '5s';

-- ── GUARDS. RAISE, never a SELECT: ON_ERROR_STOP ignores a non-empty result, so a
-- verify block of SELECTs cannot stop the COMMIT (LANDMINES / RFC_006).
DO $$
DECLARE
  n int;
  cur text;
  ss  text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='page-rerender' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '741 REFUSED: expected exactly 1 active page-rerender row, found %', n;
  END IF;

  SELECT default_config #>> '{workflow,start_step}' INTO ss FROM agent_definitions
   WHERE type='page-rerender' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF ss <> 'check_rerender_mode' THEN
    RAISE EXCEPTION '741 REFUSED: start_step is %, expected check_rerender_mode — another lane has restructured this workflow; re-read it before pasting', ss;
  END IF;

  -- The condition must be EXACTLY the five-value spec.reason clause this migration
  -- was generated against. If a sixth value was appended, the transition clause
  -- pasted below would silently DROP it — the 404 drift, committed by the fix for it.
  SELECT default_config #>> '{workflow,steps,check_rerender_mode,config,condition}' INTO cur
    FROM agent_definitions WHERE type='page-rerender' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF cur IS DISTINCT FROM 'input_data.spec.reason == ''image_landed'' OR input_data.spec.reason == ''section_data_resolved'' OR input_data.spec.reason == ''cta_links_stale'' OR input_data.spec.reason == ''template_changed'' OR input_data.spec.reason == ''literal_markdown''' THEN
    RAISE EXCEPTION '741 REFUSED: check_rerender_mode.condition is not the five-value clause this migration was generated against. Live value: %. Regenerate from livespec.TransitionRerenderModeConditionClause() against the CURRENT vocabulary and re-cut this file', cur;
  END IF;

  IF (SELECT default_config #> '{workflow,steps,check_routing_key_known}' FROM agent_definitions
        WHERE type='page-rerender' AND is_active
          AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) IS NOT NULL THEN
    RAISE EXCEPTION '741 REFUSED: check_routing_key_known already exists — this migration (or a hand edit) has already been applied; do not stack';
  END IF;

  PERFORM snapshot_agent('page-rerender',
    '741_refuse_unknown_rerender_routing_key_HOLD.sql: pre-flip (gate reads spec.reason only, unknown key assembles silently)');
END $$;

-- ── 1 + 2. The two new steps and the terminal.
--
-- ⚠ ORDER MATTERS AND THIS IS THE TRAP: the steps object is merged FIRST, and the
-- condition rewrite is a SEPARATE statement below. Nesting them as
-- jsonb_set(jsonb_set(default_config, '{...,condition}', ...), '{workflow,steps}',
-- (default_config #> '{workflow,steps}') || new) reads default_config from the ORIGINAL
-- row for the outer merge and silently DISCARDS the inner condition change.
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config, '{workflow,steps}',
         (default_config #> '{workflow,steps}') || $newsteps$
{
  "check_routing_key_known": {
    "action": "conditional",
    "config": {
      "condition": "input_data.spec.routing_reason == null OR input_data.spec.routing_reason == '' OR input_data.spec.routing_reason == 'image_landed' OR input_data.spec.routing_reason == 'section_data_resolved' OR input_data.spec.routing_reason == 'cta_links_stale' OR input_data.spec.routing_reason == 'template_changed' OR input_data.spec.routing_reason == 'literal_markdown'",
      "then_step": "check_rerender_mode",
      "else_step": "refuse_unknown_routing_key"
    },
    "description": "REFUSAL DOOR (bugs_open/440, RFC_062 phase 3). Absent/empty/known routing key -> on to the gate. Present-but-unknown -> human review, never a silent assemble. The `== null` disjunct is what allows the pre-phase-2 legacy population through; a missing key does NOT match `== ''`."
  },
  "refuse_unknown_routing_key": {
    "action": "fail_work_item",
    "config": {
      "work_item_id": "input_data.work_item_id",
      "status_override": "needs_human_review",
      "error_message": "This page_rerender item was REFUSED, not assembled: its spec.routing_reason is not in the sections-rerender vocabulary (image_landed, section_data_resolved, cta_links_stale, template_changed, literal_markdown). Read the offending value at spec.routing_reason on this row. If it was meant as a note for a human, move it to spec.reason, which is free prose and is never validated. bugs_open/440 / RFC_062.",
      "error_message_template": "This page_rerender item was REFUSED, not assembled: its spec.routing_reason = '{{.input_data.spec.routing_reason}}' is not in the sections-rerender vocabulary (image_landed, section_data_resolved, cta_links_stale, template_changed, literal_markdown). Before this refusal existed, an unrecognised routing key completed GREEN having changed nothing — bugs_open/440. If the value was meant as a note for a human, move it to spec.reason, which is free prose and is never validated. If it was meant to ROUTE, use a vocabulary value, or add one to RerenderSectionReasons in platform/livespec and paste the regenerated gate clause and this message into a migration. RFC_062."
    },
    "next_step": "complete_refused",
    "description": "OWNER RULING D1: park the item at needs_human_review with a message naming the bad key AND the vocabulary. error_message_template is fail_work_item's opt-in render (default OFF, zero other live consumers) and is what names the offending VALUE; error_message is the static fallback if it fails to render."
  },
  "complete_refused": {
    "action": "complete_workflow",
    "config": {
      "success_message": "Routing key refused — item parked at needs_human_review (bugs_open/440 / RFC_062)"
    },
    "description": "Terminal after a refusal. The ORCHESTRATION completes: D1 rules out a blunt orchestration failure, and the refusal is recorded on the ITEM, which is parked. Nothing was rendered and nothing was deployed."
  }
}
$newsteps$::jsonb),
       updated_at = now()
 WHERE type='page-rerender' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ── 3. The transition condition on the existing gate.
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config, '{workflow,steps,check_rerender_mode,config,condition}',
         to_jsonb($cond$input_data.spec.routing_reason == 'image_landed' OR input_data.spec.routing_reason == 'section_data_resolved' OR input_data.spec.routing_reason == 'cta_links_stale' OR input_data.spec.routing_reason == 'template_changed' OR input_data.spec.routing_reason == 'literal_markdown' OR input_data.spec.reason == 'image_landed' OR input_data.spec.reason == 'section_data_resolved' OR input_data.spec.reason == 'cta_links_stale' OR input_data.spec.reason == 'template_changed' OR input_data.spec.reason == 'literal_markdown'$cond$::text)),
       updated_at = now()
 WHERE type='page-rerender' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ── 4. The refusal door goes in FRONT of the gate.
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config, '{workflow,start_step}',
                                  '"check_routing_key_known"'::jsonb),
       updated_at = now()
 WHERE type='page-rerender' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ── 5. THE WRITE DOOR (D3) LIVES IN 742 NOW — SPLIT OUT, ON A COUNCIL OBJECTION.
-- `guardian` [medium], round 56047b18: "ALTER TABLE ... ADD CONSTRAINT ... NOT VALID
-- on site_work_items acquires an ACCESS EXCLUSIVE lock on the WHOLE table regardless
-- of the CHECK's item_type-scoped predicate. This table is the dispatch queue for
-- every pipeline, not just page-rerender ... bundling it with a single-pipeline
-- workflow-step insert removes the option to schedule the riskier DDL separately."
-- Conceded, and it is a BETTER argument than the one this file previously rebutted.
-- What I had rebutted was splitting in order to "get the rest in" — shipping a gate
-- and never coming back for the write door. Scheduling is a different reason: a
-- fleet-wide lock and a one-pipeline config edit want different windows, and bundled
-- in one transaction a lock timeout on the DDL rolls back the gate too.
-- ⚠ APPLY ORDER IS 741 THEN 742, and the intermediate state is SAFE rather than
-- merely tolerable: with the read door up and the write door not yet up, a bad key
-- can still be INSERTed but is REFUSED at the gate — strictly better than today,
-- where it assembles silently. The reverse order is also safe. What is NOT
-- acceptable is stopping after 741: see 742's header.

-- ── VERIFY, in a DO block so a failure actually stops the COMMIT.
DO $$
DECLARE
  cfg     jsonb;
  ss      text;
  cond    text;
  guard   text;
  n_reason  int;
  n_routing int;
BEGIN
  SELECT default_config #> '{workflow}' INTO cfg FROM agent_definitions
   WHERE type='page-rerender' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  ss := cfg #>> '{start_step}';
  IF ss <> 'check_routing_key_known' THEN
    RAISE EXCEPTION '741 VERIFY FAILED: start_step is %, expected check_routing_key_known', ss;
  END IF;

  FOR guard IN SELECT unnest(ARRAY['check_routing_key_known','refuse_unknown_routing_key','complete_refused','check_rerender_mode','render_page','rerender_sections']) LOOP
    IF cfg #> ARRAY['steps', guard] IS NULL THEN
      RAISE EXCEPTION '741 VERIFY FAILED: step % is missing', guard;
    END IF;
  END LOOP;

  -- The refusal must actually park, and must carry BOTH messages.
  IF cfg #>> '{steps,refuse_unknown_routing_key,config,status_override}' <> 'needs_human_review' THEN
    RAISE EXCEPTION '741 VERIFY FAILED: the refusal step does not park at needs_human_review (D1)';
  END IF;
  IF cfg #>> '{steps,refuse_unknown_routing_key,config,error_message_template}' NOT LIKE '%{{.input_data.spec.routing_reason}}%' THEN
    RAISE EXCEPTION '741 VERIFY FAILED: the refusal template does not interpolate the offending key — D1 asks for the bad key BY NAME, which a static literal cannot give';
  END IF;
  IF cfg #>> '{steps,refuse_unknown_routing_key,config,error_message}' IS NULL THEN
    RAISE EXCEPTION '741 VERIFY FAILED: no static error_message fallback — a template that fails to render would park the item mute';
  END IF;

  -- The guard clause must carry the null disjunct. Without it the whole legacy
  -- population goes to human review; this is the single most expensive way this
  -- migration could be wrong, so it is asserted rather than trusted.
  guard := cfg #>> '{steps,check_routing_key_known,config,condition}';
  IF guard NOT LIKE '%input_data.spec.routing_reason == null%' THEN
    RAISE EXCEPTION '741 VERIFY FAILED: the guard clause has no `== null` disjunct — every item minted before phase 2 would be refused';
  END IF;
  IF guard NOT LIKE '%input_data.spec.routing_reason == ''''%' THEN
    RAISE EXCEPTION '741 VERIFY FAILED: the guard clause has no empty-string disjunct';
  END IF;

  -- Both halves of the transition clause, counted. A count is the only thing that
  -- sees ADDITION, which is the direction this gate has actually drifted.
  cond := cfg #>> '{steps,check_rerender_mode,config,condition}';
  n_reason  := (length(cond) - length(replace(cond, 'input_data.spec.reason ==', ''))) / length('input_data.spec.reason ==');
  n_routing := (length(cond) - length(replace(cond, 'input_data.spec.routing_reason ==', ''))) / length('input_data.spec.routing_reason ==');
  IF n_reason <> 5 OR n_routing <> 5 THEN
    RAISE EXCEPTION '741 VERIFY FAILED: transition clause carries % spec.reason and % spec.routing_reason tests, expected 5 and 5', n_reason, n_routing;
  END IF;

  RAISE NOTICE '741 OK: refusal door live ahead of the gate; transition clause 5+5. ⚠ THE WRITE DOOR IS NOT UP YET — apply 742 next, then the VERIFY companion''s census, then VALIDATE (D3). The gate has NOT narrowed: 1,804 pending items still route on spec.reason.';
END $$;

COMMIT;
