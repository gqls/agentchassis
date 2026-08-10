-- 363_experience_plan_persists_only_after_the_council_approves.sql
--
-- bugs_open/227, SECOND defect — `persist_plan` runs immediately after `compose`,
-- so an EXPERIENCE_PLAN is written `is_current=true` BEFORE the council votes and
-- nothing demotes it when the verdict is a veto. On 2026-08-08 the council vetoed
-- a plan at 18:25 and the reframed version was persisted as current 8 seconds
-- before the run ended `complete_refused`: a council-rejected, fabricated plan was
-- the plan of record until it was demoted by hand.
--
-- OWNER DECISION 2026-08-10: the CONFIG-ONLY rewire (route a), not the
-- write_doc_plan Go seam (route b). Config only — live on apply, no image, no roll,
-- and the council gate does not apply (DB config is refused client-side).
--
-- WHY THIS IS THE WHOLE FIX AND NOT A PATCH: today three of the graph's steps
-- (`compose`, `recompose`, `reframe`) each hand straight to `persist_plan`, so the
-- number of plan rows a run writes is the number of compose rounds it took, and
-- every one of them is `is_current` in turn regardless of any verdict. Moving the
-- single persist onto the approved branch makes the bad state UNREPRESENTABLE
-- rather than repaired: there is no edge from an unapproved outcome to a write.
--
-- THE REWIRE — six edits, no prompt text touched:
--   compose.next_step        persist_plan  -> review_journeys
--   recompose.next_step      persist_plan  -> review_journeys
--   reframe.next_step        persist_plan  -> review_journeys
--   check_approved.then_step complete      -> persist_plan
--   persist_plan.next_step   review_journeys -> complete
--   complete_escalated.config.output_fields  drop "plan_persisted"   (see below)
--
-- Resulting graph (unchanged edges omitted):
--   compose/recompose/reframe -> review_journeys -> review_feasibility ->
--   review_honesty -> review_mvp -> review_contracts -> council_decide ->
--   append_council_note -> check_approved
--     approved            -> persist_plan -> complete
--     rejected            -> check_reframe -> reframe (once) | complete_escalated
--     revise, rounds left -> check_revise  -> run_checks -> recompose
--     exhausted           -> complete_escalated        (NOTHING PERSISTED)
--     error in persist    -> complete_refused          (error_step, unchanged)
--
-- WHY MOVING THE WRITE IS SAFE — the fact that had to be checked first, because
-- persist reads `plan_body_field: proposal.result` and it is now read LATER:
-- `compose`, `recompose` AND `reframe` all declare the SAME `output_field:
-- proposal`, so `proposal.result` always holds the most recent composition at the
-- moment the council approves. Verified against the two real runs of 2026-08-09
-- rather than assumed: `length(collected_data->'proposal'->>'result')` equals the
-- length of the plan each run actually persisted — 11,442 b for corr c3976aab
-- (loancalculator, one compose + two recomposes) and 13,840 b for corr 72f540d3
-- (vonc). Had recompose written to its own output field, this rewire would have
-- persisted the FIRST draft on every revise round, which is a silent wrong-content
-- failure rather than a missing row.
--
-- THE ONE DELIBERATE LOSS, stated because it is a behaviour change and not a
-- refactor: `complete_escalated` currently lists `plan_persisted` in its
-- `output_fields`, i.e. the escalation path is meant to surface a plan for a human
-- to choose from — and after this change nothing is persisted on that path, so the
-- field would reference a value that no longer exists. It is dropped. The
-- escalated plan is NOT lost: it remains in the run's own
-- `collected_data->'proposal'->>'result'` and in `llm_call_log`, keyed by the
-- correlation id, which is where a human reading an escalation already is.
--   Rejected alternative: persist an escalated plan under a derived subject_key
--   (e.g. '<key>:escalated') so it survives in doc_plans without colliding with
--   the plan of record. That invents a key convention nothing else knows to read,
--   which is the folklore shape the concept register exists to prevent. If an
--   escalated plan must be durable in doc_plans, the honest way is route (b) —
--   an opt-in `set_current_when` on write_doc_plan, so it can be written
--   NOT-current — and that is a platform seam owing a council round and a
--   register entry.
--
-- CONSUMERS TOLD (owner ruling 2026-07-29 §3): this changes only the
-- `experience-planner` row's own workflow graph; `write_doc_plan` itself is
-- untouched, so every other caller (`tool`, `pipeline`, `action`,
-- `experience-pattern`, `component` subjects) keeps today's behaviour exactly.
-- The only consumer of this agent's output is `doc_plans subject_type='experience'`
-- — today `vonc-spark-game` and `debt-difficulty-help`. What changes for a reader
-- of that table: it now holds ONLY council-approved plans, and a run that ends
-- `complete_escalated` or `complete_refused` leaves no row at all. That is a
-- strictly stronger guarantee than the one it replaces.
--
-- VERIFY (after apply) — the structural check is in the transaction; the
-- behavioural one is a run, and it has a cheap positive signal that could not have
-- come out the same before:
--   ./092_TRIGGER_experience_plan.sh loancalculator.co.uk debt-difficulty-help \
--     "getting help when you cannot keep up with a loan repayment"
-- An APPROVED run that takes N compose rounds wrote N plan rows before this change
-- (the 08-09 run wrote THREE for one approval: 11,631 / 10,857 / 11,442 b) and must
-- write EXACTLY ONE after it:
--   SELECT count(*), count(*) FILTER (WHERE is_current) AS current
--     FROM doc_plans WHERE subject_type='experience' AND subject_key='debt-difficulty-help'
--      AND created_at > '<run start>';           -- expect 1, 1
--
-- ⚠ CORRECTED 2026-08-10, SAME DAY, BY THE APPLYING SESSION — THE ROW-COUNT CHECK
-- ABOVE ONLY DISCRIMINATES IF THE RUN TAKES TWO OR MORE COMPOSE ROUNDS. A run
-- approved on round 1 writes exactly ONE row under the OLD graph too, so "1 row"
-- is the same answer either way and proves nothing. The first verification run
-- (corr 9150dd54-6129-464b-8600-771e0a84408a, 2026-08-10) was approved on round 1
-- — compose ×1, no recompose, no reframe — so its clean "baseline 5 -> 6, one
-- current" result is NOT evidence for this fix. Check the round count before
-- reading the row count:
--   SELECT step_name, count(*) FROM llm_call_log WHERE correlation_id='<CID>'
--     AND step_name IN ('compose','recompose','reframe') GROUP BY 1;
--
-- THE CHECK THAT DISCRIMINATES ON ANY RUN, INCLUDING A SINGLE-ROUND ONE, IS THE
-- ORDERING — and it is what actually proved this fix live. The OLD edge was
-- `compose -> persist_plan -> review_journeys`, so under the old graph a row
-- EXISTS by the time the run is executing any review step. Sample the count while
-- the run is mid-flight:
--   SELECT (SELECT current_step FROM orchestration_states WHERE correlation_id='<CID>'
--             AND status='EXECUTING_STEP'),
--          (SELECT count(*) FROM doc_plans WHERE subject_type='experience'
--             AND subject_key='<key>');
-- Observed 2026-08-10 on 9150dd54: `review_journeys` with the count still at the
-- pre-run baseline of 5 — i.e. the run had passed the point where it used to
-- persist and had written nothing. That is the disconfirmable observation.
-- STILL OWED after that, and it is the arm this migration exists for: a REJECTED
-- round must leave NO new row. It cannot be induced on demand — both 08-09 runs
-- were approved — so the honest verification is either to wait for a natural veto
-- or to seed a deliberately unbuildable experience. Until that arm is observed,
-- this fix is proven only for the approved path. ⚠ Do NOT record it as fully
-- proven on the approved run alone: that is the check-that-cannot-fail shape this
-- same bug already produced once (WRONG_CALLS 2026-08-09).
--
-- ✅ THE VETOED-ROUND ARM IS NOW OBSERVED — 2026-08-10 afternoon, corr
-- d81aa5f4-a732-4fb3-b438-4ff496ef7ba2. A deliberately unbuildable experience was
-- seeded through 345's brief channel (doc_notes keyed by subject_key, so nothing
-- else was touched: fixture docs024_key_docs_latest/loancalculator_couk/
-- probe_363_veto_arm_brief.sql), and drew a real `veto from feasibility` on round
-- 1 — no server for its write endpoint, no cross-device store, an API key in
-- client JS. The mid-flight count for that subject was 0 while the run executed
-- review_journeys (past where the old graph had already persisted), 0 across the
-- veto, 0 across the whole reframe round, and reached 1 only after check_approved
-- routed to persist on the approved round 2. UNDER THE OLD GRAPH THAT RUN WRITES
-- TWO ROWS AND THE FIRST IS THE VETOED ONE, is_current — i.e. bugs_open/227's
-- second defect verbatim, reproduced and now absent.
--
-- AND IT SETTLED AN ASSUMPTION THIS HEADER COULD ONLY CHECK HALFWAY. The
-- persisted body was 7,661 b = the REFRAME response exactly, not compose's
-- 12,189 b. The "compose, recompose AND reframe all write proposal" claim above
-- was verified against compose+recompose runs only; the reframe branch had never
-- been measured, and it is the branch where being wrong would have persisted the
-- VETOED draft on approval.
--
-- WHAT IS STILL OWED, narrowed: a run that ENDS non-approved leaving no row.
-- Note the reason it stayed open, because "wait for a natural veto" is not a
-- plan: A VETO IS NOT TERMINAL BY DESIGN. reframe's prompt tells it to demote the
-- vetoed feature to a coming-soon label ("that is an acceptable honest MVP"), and
-- applyCouncilCaps (diagnose_council_decide_action.go:663) escalates only on a
-- SECOND rejection. The way to force it: set council_decide.config.max_rounds to
-- 1 — then any non-approved round-1 verdict routes straight to complete_escalated
-- — fire an unbuildable experience, assert no new row, and RESTORE max_rounds to
-- 5. Attempted 14:51Z; the run died at compose on a fleet-wide Anthropic usage cap
-- before reaching the council. ⚠ That failed run READS AS A PASS AND IS NOT ONE
-- (complete_refused, no new row, plan of record unchanged — but compose never
-- returned, so the old graph writes nothing either).
--
-- ROLLBACK: 363_..._ROLLBACK.sql restores from agent_definitions_backup, picking
-- by snapshot_taken_at DESC with snapshot_reason LIKE 'pre-update: 227 persist%'.
-- Every backup row for one agent shares the source row's id and created_at, so
-- ordering by created_at returns an arbitrary snapshot — order by
-- snapshot_taken_at (LANDMINES 2026-07-30).

-- ============================================================================
-- Probe guard: refuse a second application.
-- ============================================================================
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM agent_definitions
        WHERE type = 'experience-planner'
          AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
          AND default_config #>> '{workflow,steps,check_approved,config,then_step}' = 'persist_plan'
    ) THEN
        RAISE EXCEPTION '227/363: already applied — check_approved.then_step is already persist_plan';
    END IF;
END $$;

-- ============================================================================
-- Drift guard: every edge this file rewrites, pinned to what it was composed
-- against at 2026-08-10 10:2xZ. A raise means another session changed the graph —
-- re-read it and recompose, do not force. The row is bulk-touched by the DEPLOY on
-- every roll (scripts/deploy/update-agent-images.sh stamps image_tag; measured
-- 2026-08-10: 189 rows, image_tag + updated_at only, default_config untouched), so
-- a fresh updated_at is NOT evidence of a config change — check the edges.
-- ============================================================================
DO $$
DECLARE
    c_next text; r_next text; f_next text; p_next text; a_then text; esc_of text; brief text;
BEGIN
    SELECT default_config #>> '{workflow,steps,compose,next_step}',
           default_config #>> '{workflow,steps,recompose,next_step}',
           default_config #>> '{workflow,steps,reframe,next_step}',
           default_config #>> '{workflow,steps,persist_plan,next_step}',
           default_config #>> '{workflow,steps,check_approved,config,then_step}',
           default_config #>> '{workflow,steps,complete_escalated,config,output_fields}',
           default_config #>> '{workflow,steps,load_schema_hint,next_step}'
      INTO c_next, r_next, f_next, p_next, a_then, esc_of, brief
      FROM agent_definitions
     WHERE type = 'experience-planner'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF c_next IS DISTINCT FROM 'persist_plan'
       OR r_next IS DISTINCT FROM 'persist_plan'
       OR f_next IS DISTINCT FROM 'persist_plan'
       OR p_next IS DISTINCT FROM 'review_journeys'
       OR a_then IS DISTINCT FROM 'complete' THEN
        RAISE EXCEPTION '227/363 drift: graph edges are not what this file was composed against '
                        '(compose=%, recompose=%, reframe=%, persist=%, approved_then=%)',
                        c_next, r_next, f_next, p_next, a_then;
    END IF;

    IF esc_of IS DISTINCT FROM '["plan_persisted", "council"]' THEN
        RAISE EXCEPTION '227/363 drift: complete_escalated.output_fields is % — recompose the drop', esc_of;
    END IF;

    -- 345 must already be live: this file assumes the brief-as-data chain.
    IF brief IS DISTINCT FROM 'load_brief' THEN
        RAISE EXCEPTION '227/363: expected 345 applied (load_schema_hint.next_step=load_brief), found %', brief;
    END IF;
END $$;

BEGIN;

SELECT snapshot_agent('experience-planner',
    'pre-update: 227 persist-after-approval rewire (loancalculator_couk lane, owner decision 2026-08-10)');

-- ============================================================================
-- The six edits.
-- ============================================================================
UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           jsonb_set(
             jsonb_set(
               jsonb_set(
                 jsonb_set(default_config,
                   '{workflow,steps,compose,next_step}',   '"review_journeys"'::jsonb, false),
                 '{workflow,steps,recompose,next_step}',   '"review_journeys"'::jsonb, false),
               '{workflow,steps,reframe,next_step}',       '"review_journeys"'::jsonb, false),
             '{workflow,steps,check_approved,config,then_step}', '"persist_plan"'::jsonb, false),
           '{workflow,steps,persist_plan,next_step}',      '"complete"'::jsonb, false),
         '{workflow,steps,complete_escalated,config,output_fields}', '["council"]'::jsonb, false)
 WHERE type = 'experience-planner'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Keep the descriptions honest — a stale description is how a graph becomes folklore.
UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(default_config,
           '{workflow,steps,persist_plan,description}',
           '"Supersede-write the EXPERIENCE_PLAN (doc_plans, subject_type=experience). Reached ONLY from check_approved, so doc_plans holds approved plans exclusively: a vetoed or escalated run leaves no row (bugs_open/227 second defect, migration 363). Reads proposal.result, which compose, recompose AND reframe all write."'::jsonb, false),
         '{workflow,steps,check_approved,description}',
         '"Router 1/3: approved is the only path to a persisted plan, then complete. Everything else routes away WITHOUT writing (migration 363)."'::jsonb, false)
 WHERE type = 'experience-planner'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ============================================================================
-- Verify inside the transaction: assert the whole reachable graph, not just the
-- fields we set — a jsonb_set with a wrong path silently ADDS a key rather than
-- failing, so "the value I wrote is there" is not the same as "the edge moved".
-- ============================================================================
DO $$
DECLARE
    c_next text; r_next text; f_next text; p_next text; a_then text; a_else text;
    esc_of jsonb; n_persist int;
BEGIN
    SELECT default_config #>> '{workflow,steps,compose,next_step}',
           default_config #>> '{workflow,steps,recompose,next_step}',
           default_config #>> '{workflow,steps,reframe,next_step}',
           default_config #>> '{workflow,steps,persist_plan,next_step}',
           default_config #>> '{workflow,steps,check_approved,config,then_step}',
           default_config #>> '{workflow,steps,check_approved,config,else_step}',
           default_config #>  '{workflow,steps,complete_escalated,config,output_fields}'
      INTO c_next, r_next, f_next, p_next, a_then, a_else, esc_of
      FROM agent_definitions
     WHERE type = 'experience-planner'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF c_next <> 'review_journeys' OR r_next <> 'review_journeys' OR f_next <> 'review_journeys' THEN
        RAISE EXCEPTION '227/363 verify: a compose-family edge still points at persist (%, %, %)',
                        c_next, r_next, f_next;
    END IF;
    IF a_then <> 'persist_plan' OR p_next <> 'complete' THEN
        RAISE EXCEPTION '227/363 verify: approved path is not check_approved->persist_plan->complete (%, %)',
                        a_then, p_next;
    END IF;
    IF a_else <> 'check_rejected' THEN
        RAISE EXCEPTION '227/363 verify: check_approved.else_step changed to % — the veto routing moved', a_else;
    END IF;
    IF esc_of <> '["council"]'::jsonb THEN
        RAISE EXCEPTION '227/363 verify: complete_escalated.output_fields is % not ["council"]', esc_of;
    END IF;

    -- NO step other than check_approved may reach persist_plan. This is the
    -- property the fix actually claims, and it is the one a future edit will break.
    SELECT count(*) INTO n_persist
      FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
     WHERE a.type = 'experience-planner'
       AND a.is_active AND COALESCE(a.is_snapshot, false) = false AND a.deleted_at IS NULL
       AND s.key <> 'check_approved'
       AND (s.value->>'next_step' = 'persist_plan'
            OR s.value->'config'->>'then_step' = 'persist_plan'
            OR s.value->'config'->>'else_step' = 'persist_plan');
    IF n_persist <> 0 THEN
        RAISE EXCEPTION '227/363 verify: % step(s) other than check_approved still reach persist_plan', n_persist;
    END IF;

    RAISE NOTICE '227/363 OK: persist_plan is reachable only from check_approved, and leads to complete.';
END $$;

COMMIT;
