-- 539 — build-dispatch-loop's `mark_complete` step NAMES where `commit_sha` comes
--       from, with the `?` OPTIONAL-EXPLICIT marker: that path or ABSENCE, never
--       the whole-tree search. RFC_029 §10.13 step 5's LAST live blocker.
--       CONFIG ONLY — live on apply. NOT a _HOLD (see ORDERING).
--
-- ============================================================================
-- WHAT THE PROBLEM IS, in plain terms
-- ============================================================================
-- When a work item finishes, `complete_work_item` records which code change
-- deployed it. The action declares `commit_sha` as an Optional input
-- (load_work_item_actions.go:56) and writes it to the item's result
-- (:937-938). build-dispatch-loop's `mark_complete` step wires `result!` and
-- `work_item_id!` but says NOTHING about `commit_sha`, so the resolver falls
-- through to its last resort: collect every key of that name anywhere in the
-- collected data and pick a winner.
--
-- Inside a LOOP that tree accumulates one alias per iteration — `handler_result`,
-- `handler_result_0`, `handler_result_1`, … plus `process_item_iter_N_call_handler`
-- — and each iteration deployed something different, so the values genuinely
-- DIFFER. That is a conflict on every multi-item loop run, and it is this
-- estate's single largest live conflict class: **387 rows in the 24 h to
-- 2026-08-21 15:01Z**, still climbing.
--
-- ============================================================================
-- WHAT IT COSTS TODAY: PROBABLY NOTHING, BY LUCK — WHICH IS THE PROBLEM
-- ============================================================================
-- The search sorts shallowest-first, so the UNSUFFIXED `handler_result` wins,
-- and that alias holds the CURRENT iteration's result. So today's answer is
-- most likely correct. Nothing declares that, nothing tests it, and a reordered
-- collector would change it with no failing test. `bugs_open/334`.
--
-- ============================================================================
-- WHY IT MUST BE FIXED BEFORE THE FLIP, AND WHAT BREAKS IF IT IS NOT
-- ============================================================================
-- RFC_029 §9 D2's Phase 2 makes CONFLICTING candidates resolve to NOTHING. On
-- the day that ships, `commit_sha` inside a loop resolves to nothing, and
-- `result.commit_sha` **silently stops being recorded** — no error, because the
-- field is Optional. `bugs_open/315`'s page-stamping work depends on that value
-- ("3 of 19 git_commit steps feed a page stamp"). This is the pair the flip's
-- own precondition exists for.
--
-- ============================================================================
-- WHERE THE PATH COMES FROM — the 315 lane built it, and it is MEASURED here
-- ============================================================================
-- We asked the 315 lane for the path that is correct BY THEIR LIGHTS rather than
-- picking one from the shape (CONTRIB_2026-08-20_from_staged_component_build_…md).
-- They answered with three options, chose (b) "make the path stable at the
-- SOURCE", and BUILT it: migrations 519/521/522/523/527/528/534/535/536 convert
-- each handler's `complete` step from `output_fields` list mode to
-- `result_mapping`, hoisting `commit_sha` to a canonical top-level key. Their
-- handoff: "handler side COMPLETE — all 9 real handlers standardised, applied,
-- verified live. The wire is staged-component-build's remaining piece."
--
-- [MEASURED 2026-08-21 ~15:1xZ] on live build-dispatch-loop trees created AFTER
-- the last of those migrations (536, applied 14:17:47Z) — the boundary matters,
-- because over ALL trees the canonical path looks like a minority (31 of 175)
-- purely because most trees predate the migrations:
--
--   runs with handler_result            25
--   canonical  handler_result.response.commit_sha                        23
--   old nested handler_result.response.deploy_result.response.data...    20
--   canonical ONLY (canonical present, nested absent)                     3
--   nested ONLY  (nested present, canonical absent)                       0   <-- the key cell
--   neither                                                               2
--
-- **The canonical path strictly dominates**: there is no run where the old
-- nested path carries a sha and the canonical one does not. So naming the
-- canonical path loses nothing and gains 3.
-- (⚠ compute those cells with COALESCE(... ? 'commit_sha', false) — a jsonb `?`
-- on a NULL intermediate yields NULL, and `x AND NOT y` then drops the row
-- silently, which made an earlier run of this very query self-contradictory.)
--
-- ============================================================================
-- WHY `?` AND NOT `!` — the 315 lane recommended `!`, and their OWN measurement
-- refutes it
-- ============================================================================
-- `!` means explicit-path-or-FAIL (action_inputs.go:750, "strict enforcement at
-- the bottom fails `!` fields"). Their §3 suggested `commit_sha!`. Their §4 then
-- warns that ABSENCE IS CORRECT for items whose handler contains no `git_commit`
-- at all — most item types deploy nothing. Measured independently, and it is
-- larger than their figure:
--
--   SELECT count(*) , count(*) FILTER (WHERE result ? 'commit_sha')
--     FROM site_work_items WHERE status='complete' AND jsonb_typeof(result)='object'
--      AND completed_at >= now() - interval '3 days';
--   -> 2195 completed, 1115 with a sha, **1080 WITHOUT (49%)**
--
-- and 2 of the 25 post-boundary loop runs above carry neither path. So
-- `commit_sha!` would hard-fail about HALF of all work-item completions. `?` is
-- the correct marker: resolve from the named path, otherwise absent, no error.
--
-- ============================================================================
-- WHY `handler_result` IS A SAFE ROOT — and the LANDMINE that says otherwise
-- describes the PRE-FIX world (council round 1, guardian, HIGH / gating)
-- ============================================================================
-- The step ALREADY reads `result!: handler_result`, so this wire inherits an
-- assumption the step has always made rather than adding one.
--
-- Round 1's gating objection cited a LANDMINE keyed almost verbatim to this site:
-- "A `complete` work item's `result` may be the SPAWN RECORD, not the handler's
-- reply … `mark_complete` asks for `handler_result` (un-suffixed), finds nothing,
-- and the aggressive recursive search returns the spawn record." If that were
-- still true, the trust this file inherits would be exactly the trust that entry
-- says is misplaced. **The objection was right to raise it. The answer is that the
-- entry describes the world BEFORE its own fix, and its own footer says so.**
--
--  1. The defect is `bugs_closed/287`, and that same landmine records its closure:
--     "migs 448/452 + WFA-017, live and proven on v1.0.1307 — field=result
--     resolver rows 0, 11/11 completions carrying the reply".
--  2. **Migration 452 IS the `result!` wire** ("build_dispatch_loop_result_goes_
--     strict"). So the strict marker here is not incidental — it is the mechanism
--     that closed 287. `!` means explicit-path-or-FAIL, which is precisely what
--     stops the fall-through to the aggressive search that the landmine describes.
--  3. [MEASURED 2026-08-21 ~15:4xZ] the spawn-record shape is GONE from this slot.
--     Shape test: a spawn record carries `role`/`topics`/`agent_id` and no `response`.
--       all retained bdl trees carrying handler_result : 185 — spawn records **0**
--       post-536 window (created after 14:17:47Z)      :  35 — spawn records **0**,
--                                                          and 35 of 35 carry `response`
--
-- So the argument is STRONGER than round 1 stated it: the root is safe not merely
-- because the step reads it, but because it reads it **STRICTLY** — and that is the
-- fix. `handler_result` cannot silently become a spawn record again without
-- `result!` failing loudly first.
--
-- ⚠ The landmine stays VALID and must NOT be retired on the strength of this file:
-- ~939 historical rows hold the spawn record permanently (their parents were
-- reaped), and any run created before its agent's migration still carries the old
-- shape. What is established here is narrower — that THIS wire, on runs THIS
-- migration can affect, does not inherit the defect.
-- ============================================================================
-- ORDERING — why this is NOT a _HOLD file
-- ============================================================================
-- The `?` marker on the step-config surface parses only in a binary carrying
-- `ecc419bd1`. That binary is live (`v1.0.1321`, built from `0483e7f4e`), and
-- unlike migration 515 this no longer rests on a source-level argument at all:
-- **`?` has been PROVEN IN PRODUCTION.** 515 applied the first such key at
-- 13:19:19Z and at 14:24:28.457Z a live `plan_sections` extraction logged
--   requested_fields: ['section_facts', 'pipeline', 'site_type']
-- with `page_type` ABSENT (excluded by its marker) while `site_type` — same
-- Optional list, same action, also unwired — was still present as an internal
-- control. The marker works in the fleet.
--
-- ROLLBACK: 539_build_dispatch_loop_declares_commit_sha_ROLLBACK.sql
-- ============================================================================

BEGIN;

DO $$
DECLARE
    cfg      jsonb;
    n_active  int;
    n_missing int;
BEGIN
    SELECT count(*) INTO n_active FROM agent_definitions
     WHERE type = 'build-dispatch-loop' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n_active <> 1 THEN
        RAISE EXCEPTION '539: expected exactly 1 active build-dispatch-loop row, found % — a LANDMINE records four agent types carrying TWO active rows where only the higher version loads; resolve and pin the version before applying', n_active;
    END IF;

    -- The step is NESTED inside the process_item loop's sub_workflow. A
    -- top-level jsonb_each census cannot see it — that trap has bitten this lane
    -- four times, so the path is spelled out and asserted rather than assumed.
    SELECT default_config #> '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config}'
      INTO cfg
      FROM agent_definitions
     WHERE type = 'build-dispatch-loop' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '539: no mark_complete step at workflow.steps.process_item.config.sub_workflow.steps.mark_complete — the loop has been restructured; re-measure before applying';
    END IF;

    IF (SELECT default_config #>> '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,action}'
          FROM agent_definitions
         WHERE type = 'build-dispatch-loop' AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL) <> 'complete_work_item' THEN
        RAISE EXCEPTION '539: mark_complete no longer runs complete_work_item — every measurement in this file is about a different action';
    END IF;

    -- The root this wire depends on must already be in use by the same step.
    IF (cfg ->> 'result!') IS DISTINCT FROM 'handler_result' THEN
        RAISE EXCEPTION '539: result! is % , expected handler_result — this file argues the root is safe BECAUSE the step already reads it; if that changed, re-derive the path', COALESCE(cfg ->> 'result!', '<absent>');
    END IF;

    -- The canonical path only EXISTS because nine handler agents were converted
    -- to result_mapping (315 lane, migs 519/521/522/523/527/528/534/535/536).
    -- Council round 1 (guardian, low) noted the plan cited a HISTORICAL sample
    -- rather than live coverage. So assert it at APPLY TIME: if any handler has
    -- lost its canonical mapping — rolled back, re-snapshotted, re-seeded — its
    -- commit_sha would silently revert to absent under this wire, with no alarm.
    SELECT count(*) INTO n_missing
      FROM (VALUES ('page-rerender'),('rerender-pages'),('section-editor'),('nav-updater'),
                   ('webdesign-agent'),('css-patch-agent'),('page-build-handler'),
                   ('asset-deployer'),('image-build-handler')) AS h(t)
     WHERE NOT EXISTS (
        SELECT 1 FROM agent_definitions d
         WHERE d.type = h.t AND d.is_active
           AND COALESCE(d.is_snapshot, false) = false AND d.deleted_at IS NULL
           AND jsonb_path_query_array(d.default_config, '$.**.result_mapping.commit_sha') <> '[]'::jsonb);
    IF n_missing <> 0 THEN
        RAISE EXCEPTION '539: % of the 9 canonical handlers have no result_mapping.commit_sha — the path this wire names would be absent for them. Restore their mapping (315 lane migs 519/521/522/523/527/528/534/535/536) before applying', n_missing;
    END IF;
    IF cfg ? 'commit_sha?' THEN
        RAISE EXCEPTION '539: commit_sha? is already declared — already applied';
    END IF;
    IF cfg ? 'commit_sha' OR cfg ? 'commit_sha!' THEN
        RAISE EXCEPTION '539: a competing commit_sha wire already exists (unmarked or strict). Someone chose that deliberately — read their reason before converting it';
    END IF;

    -- Snapshot INSIDE the guard so an aborted re-run cannot leave a misleading
    -- second 'pre-update' row (council round 1 on migration 515, guidelines seat).
    PERFORM snapshot_agent('build-dispatch-loop',
                           '539_build_dispatch_loop_declares_commit_sha: pre-update');

    UPDATE agent_definitions
       SET default_config = jsonb_set(
               default_config,
               '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config,commit_sha?}',
               '"handler_result.response.commit_sha"'::jsonb,
               true),
           updated_at = NOW()
     WHERE type = 'build-dispatch-loop' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
END $$;

-- ============================================================================
-- VERIFY — a DO block, not SELECTs. ON_ERROR_STOP does not abort a COMMIT on a
-- non-empty result set, so a verify made of SELECTs cannot stop a bad apply.
-- ============================================================================
DO $$
DECLARE
    got text;
BEGIN
    SELECT default_config #>> '{workflow,steps,process_item,config,sub_workflow,steps,mark_complete,config,commit_sha?}'
      INTO got
      FROM agent_definitions
     WHERE type = 'build-dispatch-loop' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF got IS DISTINCT FROM 'handler_result.response.commit_sha' THEN
        RAISE EXCEPTION '539 VERIFY FAILED: commit_sha? reads % , expected handler_result.response.commit_sha', COALESCE(got, '<null>');
    END IF;

    IF (SELECT count(*) FROM agent_definitions_backup
         WHERE type = 'build-dispatch-loop'
           AND snapshot_reason LIKE '539_build_dispatch_loop_declares_commit_sha%') < 1 THEN
        RAISE EXCEPTION '539 VERIFY FAILED: no snapshot row in agent_definitions_backup';
    END IF;

    RAISE NOTICE '539 OK: mark_complete declares commit_sha? = handler_result.response.commit_sha';
    -- The REAL apply owes its own logged verify: this NOTICE plus an independent
    -- read-back, and then the runtime read with a DEMAND control (bdl loop runs
    -- in the window > 0). The conflict count alone is not the evidence.
END $$;

COMMIT;
