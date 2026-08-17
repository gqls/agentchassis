# NOTES — bugfix 291 (append-only, newest at the bottom)

## 2026-08-17 — session start, verification, and design

- Bug 291 taken on. Ownership checked: `who-owns.py 291` → filed by the 284 lane, which
  closed 284 the same morning ("no handoff needed"); live-transcript grep (the lagging-check
  memory) found the only heavy `hitl-review` session WAS the 284 lane. Unclaimed. Queue empty.
- `[MEASURED 12:27Z]` 14 rows `status='blocked' AND error='Handler agent not registered:
  hitl-review'`, 2 sites, newest 06:55Z. Same count re-measured 12:04Z pre-090.
- `[MEASURED]` `hitl-review` appears at exactly ONE live-config path fleet-wide:
  `tool-auditor.default_config.workflow.steps.create_items_loop.config.sub_workflow.steps.create_review_item.config.handler_agent`
  (recursive jsonb walk over the live row).
- **CORRECTION to the bug file** `[MEASURED]`: its "Measure before choosing" line says the
  other `needs_human_review` rows "carry an empty handler". They do not — the 7 non-291 rows
  all carry `handler_agent='human-review'` (checkpoint producers: tool-recreation-handler,
  image-url-404-handler, generic). The EMPTY-handler idiom is real but lives in the Go
  discovery checks (`check_unverified_claims.go`, `check_voice_tells.go`) and in the fleet
  census at `refresh_evidence_fact_drift.go:698-703` (544 `''` vs 22 `'human-review'`).
  Correction to be written into the bug file visibly.
- **CORRECTION to the bug file**: `resolve_composition_layout_action.go:390` is NOT "the same
  mistake waiting to fire" in the dispatch sense — :391 sets `status:"needs_human_review"`
  explicitly, so its items are never claimed. The wrong HANDLER value spread by copying; the
  safe STATUS did not. The bleeding difference is tool-auditor's missing status key.
- Root cause pinned: `create_work_item_action.go:208-211` defaults status to `'triaged'`;
  seed `088_tool_auditor_agent.sql:158` names the handler, sets no status; no later migration
  (349/350/425/434) added one. `hitl-review` documented 2026-04-19 as "a convention, not a
  registered agent" — the handler was a roadmap row never built.
- Design stress-test (Plan subagent) found the FATAL flaw in the first draft: the live binary
  refuses empty `handler_agent` in `create_work_item` config (:184-187 hard error), so a
  config flip to `''` today would silently lose every finding under `continue_on_error:true`.
  Verified first-hand at :184-187 before accepting. → three-phase plan (see PLAN).
- Also from the stress-test, verified against the tree: migration slot 446 already taken
  (tool-suggester lane) → ours are 447/448; the INSERT widening trap at
  `load_work_item_actions.go:1396-1408` (conditional-append idiom is mandatory for the new
  `error` column); ~25 sqlmock test files expect the shared INSERT and the guard's probe
  needs adding where their paths now see it.
- `[MEASURED]` 13 of the 14 rows carry `spec->>'check'='tool_auditor'` + `spec.issue`; the
  2026-08-14 finetuning.uk row predates migration 434's spec rewrite and lacks both (and its
  item_key has no page-id suffix — page_id was empty). Repair keys on created_by + exact
  error text, NOT spec.
- All 14: `attempt_count=0`, `claimed_at IS NULL`, `approval_mode='auto'` — claim's
  not-registered branch nulls the claim fields, so the repair need not touch them.
- tool-auditor is driven as a work-item handler by the dispatch loop (~hourly runs observed
  2026-08-17 06:15→10:27 on webdesign.co.uk); the bleed is live, hence config-first phasing.
- 090 filed 12:03Z: `RUN_CORRELATION_ID=3555b514-ca8f-4f31-9f55-e105ce73e961` (dispatch-loop
  correlation, the artifact key). Verdict pending at time of writing.

## 2026-08-17 (later) — Phase 1 LIVE, Phase 2 committed, 090 verdict in (with a self-inflicted caveat)

- Migration numbers moved TWICE under us on the shared tree (446 taken at plan time;
  447 and 449 taken by other sessions by write time) → ours are **450** (config,
  status-only) + **451** (repair). Both applied by hand + `--record-only` 12:07-12:08Z.
- Phase 1 verified at the artefact: config path carries `status: needs_human_review`
  (handler untouched, deliberately); snapshot proven PRE-update in
  `agent_definitions_backup` (no status key in backup, key in live); **0 rows** at the
  blocked predicate; all **14** repaired rows at `needs_human_review`/`''` with
  `result.repair_291`.
- Phase 2 committed **`c8400e452`** (guard + relaxed validation + resolve_composition
  flip + 6 tests + WDS-018 + index row). Full actions suite green.
  **Mutation proofs, 3, all bit**: (1) guard block disabled → born-blocked test fails
  on INSERT arg $12; (2) trigger set widened → `TestStatusRequiresRegisteredHandler_
  ExactlyCheck443sList` fails by name; (3) function call bypassed (probe
  unconditional) → the scripted-probe tripwire fails the parked shapes on a demoted
  INSERT.
- **MISSTEP (logged fleet-wide in WRONG_CALLS.md): the first version of the
  parked-shapes test PASSED under mutation (2).** A bare no-probe-expectation cannot
  catch a widened trigger set, because the guard's own probe-failure fall-through
  swallows sqlmock's unexpected-query error and inserts normally — my graceful-degrade
  design defeated my own negative test ("a mutation that passes usually hit a guard in
  series"). Fix: script the probe to answer "not registered" as a TRIPWIRE (wrongly
  probing build demotes → INSERT arg mismatch) + a function-level exact-set pin.
- **MISSTEP #2 (WRONG_CALLS.md): I repaired the state the 090 run was about to read.**
  Filed 12:03; migrations applied 12:07-08; diagnoser queried ~12:10+. Verdict
  **CONFIRMED** on the core (phantom handler, config route) but three legs read
  "contradicted" — all three caused by 450/451 landing first, all three true at filing
  time (dated first-hand evidence above). The verdict artifact will mislead a cold
  reader; the bug file now carries the timeline caveat.
- 090 bonus: one parked `hitl-review` row from 2026-08-12
  (`needs_new_layout_candidate`, site-design-planner) — resolve_composition_layout IS
  live and reachable, contra the repo-side "no workflow wires it" uncertainty. Staged
  Phase 3 file extended to sweep parked rows' handler to `''` (scoped to
  `status='needs_human_review'`, refuses if any NON-parked row still carries the name).
- Council submitted 12:20Z: `SUBMISSION_CORR=4d1ed8a5-20c4-420f-b619-6197ab9af1b2`,
  committed with `Council-Submitted:` trailer per the 2026-07-30 rule. Verdict pending
  (~30 min budget); READ IT and act on REVISE/REJECTED — the code is on the shared
  branch.
- Docs done: 027/020/016 hitl-review retirement (docs024 + docs014 twins, 6 files);
  016b §9 entry ("an item TYPE named for a state is not the STATE"); LANDMINES entry
  (create_work_item status default) + `landmines-verify-dispatch.sh` run (dispatched 1).
- Supporting evidence the live binary honours a config-set `status` on this exact action
  (450's whole mechanism): `grounded-explainer` rows sit at `needs_human_review` (created
  via its `create_work_item` step, seed 224, which sets the status key explicitly) —
  `grounded_draft_review` ×2 + `citation_unverified` ×1. `[MEASURED 2026-08-17]`. The
  definitive artefact proof is still the next NATURAL tool-auditor run (no dispatchable
  `audit_tool` items at 12:45Z, so it waits for the next discovery/rotation pass — check
  per RUNBOOK, and run the straggler sweep after).

## 2026-08-17 (later still) — council round 1 REVISE, revised, round 2 in

- Round 1 verdict 12:31Z: **REVISE, gated by guardian [high]** (shared-door change:
  wants consumer sign-off evidence, probe cost measured not asserted, and a rollback
  lever for the guard itself). Substantive objections also from `editquality` +
  `bug_historian` (both correct and the sharpest of the round): **demote-to-blocked
  does not close the dedup-slot loss** — `audit_review_<page_id>` is one slot per PAGE,
  so any OPEN item (blocked before, parked after) swallows later DISTINCT findings;
  that is a KEY GRANULARITY defect no routing fix can close. `debug_historian` [medium]
  judged the migrations lacking needle-gate discipline — from the one-line description;
  the actual files carry snapshot+pinned-id+DO/RAISE gates+stamped repair+rollbacks,
  requoted verbatim in round 2. `reuse_agent` [low] wants the probe/demote control flow
  unified with claim's — deferred citing WDS-017's recorded owner-call precedent (two
  seats pulled opposite ways on exactly that unification in 284's round).
- Revision committed `f629f4530`: **kill-switch** env `DISABLE_UNREGISTERED_HANDLER_DEMOTION`
  (ships ARMED; disarm = exact pre-guard behaviour; disarm test added) + **probe cost
  measured**: EXPLAIN ANALYZE 0.107ms actual, index-only scan on
  `idx_agent_definitions_type_active` (210 rows), × 613 site_work_items inserts/24h
  fleet-wide ≈ 0.07 s/day. WDS-018 limits now name the dedup residual's owner:
  per-FINDING review keys, the bugfix-285 lane's recorded follow-on.
- Round 2 submitted ~12:50Z, same trail: `RESUBMIT_CORR=4d1ed8a5`, run orch
  `a89b64c8-c7ac-44f3-8ef8-61272b5acba5`. Verdict watcher armed.
