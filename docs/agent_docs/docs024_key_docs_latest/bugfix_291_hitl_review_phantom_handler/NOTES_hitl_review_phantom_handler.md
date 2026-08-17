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

## 2026-08-17 (close of session) — council APPROVED round 2

- Round 2 verdict ~12:55Z: **APPROVED — "approved with 2 advisory objection(s), none
  high-severity"**. `editquality`, `bug_historian`, `reuse_agent` and 8 others clean.
  Advisories, all already answered by design or process:
  - guardian [medium]: register+council is "documentation not sign-off" — recorded; the
    estate has no per-pipeline sign-off authority short of the owner (2026-07-29 ruling
    #3's "tell them" is the register + this trail).
  - guardian [low]: probe latency under agent_definitions lock contention not ruled out
    by an EXPLAIN — the kill-switch is the mitigation (disarm without a roll).
  - debug_historian [medium]: rollout evidence must be read from the RUNNING binary,
    never commit hashes — agreed; that is already the bug file's closure criterion
    (provenance stamp + merge-base before the staged flip, RUNBOOK section).
  - prior_art_librarian [low ×2]: helper existence + kill-switch commit — both verified
    in-tree (`workItemHandlerRegisteredSQL` pre-exists from 284 at work_items_common.go;
    kill-switch in `f629f4530`).
- Commits `c8400e452` + `f629f4530` carry `Council-Submitted: 4d1ed8a5…`; 098 credits
  them automatically now the verdict is approved (no amend — forward-only).
- **Remaining to close 291** (fixed-AND-live bar): (1) next fleet roll carries the two
  Go commits — verify per SERVICE at the provenance stamp, `git merge-base
  --is-ancestor c8400e452 <stamp>`; (2) apply the STAGED phase-3 flip (renumber into
  sql_for_agents/ at the then-free number); (3) one natural tool-auditor run files
  `''`+`needs_human_review` review items; (4) straggler sweep + next-day census.

## 2026-08-17 17:2x — "a fresh chassis build has been deployed" — IT DID NOT SHIP MY CODE

- **Checked before acting, and the ordering gate on the staged Phase 3 file did its
  job: PHASE 3 NOT APPLIED.** Applying it against this binary would hard-error every
  review-item filing inside `continue_on_error` — every finding silently lost.
- Evidence, in order taken:
  - Pods `agent-chassis-5bd56bdd9b-{6sb8t,jzmns}` restarted 14:43Z, image
    `v1.0.1305` — **the same tag as before my commits**; `makefile` `IMAGE_TAG` still
    `v1.0.1305`, never bumped.
  - `build provenance` startup line had scrolled (busy service — the known trap; an
    empty grep means "out of range", not "unstamped").
  - **Binary probe, controls in the same breath and both discriminating:** my
    kill-switch literal `DISABLE_UNREGISTERED_HANDLER_DEMOTION` **ABSENT** (0);
    positive control `Handler agent not registered: ` PRESENT (1); negative control
    `ZZZ_NOT_A_REAL_SYMBOL_291` absent (0).
  - **The decisive one — the build WORKED, the delivery did not.** The LOCAL image
    `aqls/agent-chassis:v1.0.1305`, built 14:30Z, **contains** my literal (1, with the
    same two controls passing). Digests differ: local `sha256:6039e19c…` vs running
    `sha256:f90a7e88…`. `imagePullPolicy: IfNotPresent` + an unchanged tag = the node
    serves its cached older image for ever.
- **Not my lane's problem alone — it is fleet-wide today.** Another lane independently
  measured the running chassis still at commit `6a782274b` with **203 commits in HEAD
  but not in it**. So every Go change committed on 2026-08-17 is inert, mine included.
- **Remedy is owner-run and is a TAG BUMP, not a redeploy** (a re-apply at the same tag
  re-serves the same cache): `make release IMAGE_TAG=v1.0.1306` (the variable is `?=`,
  so no makefile edit is needed; releases are whole-fleet by ruling).
- Fleet-wide LANDMINE appended: the 10-second digest comparison (`docker inspect
  RepoDigests` vs pod `imageID`) — sharper than a sha probe and needs no exec. ⚠ It
  went in under another session's commit (`07229196e`) as a **same-file passenger** —
  their message, my text; content intact, attribution off. Forward-only, left alone.
- **291's close-out is therefore UNCHANGED and still owed**: (1) a roll that actually
  ships `c8400e452`+`f629f4530` — verify by DIGEST, then `git merge-base
  --is-ancestor c8400e452 <stamp>`; (2) staged Phase 3 flip; (3) one natural
  tool-auditor run filing `''`+`needs_human_review`; (4) straggler sweep + census.

## 2026-08-17 ~17:2xZ — v1.0.1307 IS the real roll: guard LIVE + EXERCISED, Phase 3 applied, 291 CLOSED

- **Roll proven three ways before touching anything** (the 14:43Z "fresh build" failed
  the first of these; this one passes all three):
  1. **DIGEST match** — local `aqls/agent-chassis:v1.0.1307` RepoDigest
     `sha256:8339bdbd7999…` **==** the running pods' `imageID`. (Tag was bumped this
     time: `IMAGE_TAG` now `v1.0.1307`.)
  2. **BINARY probe, both controls discriminating** — `DISABLE_UNREGISTERED_HANDLER_DEMOTION`
     PRESENT (1); positive control `Handler agent not registered: ` present (2);
     negative control `ZZZ_NOT_A_REAL_SYMBOL_291` absent (0).
  3. **ANCESTRY with a control** — image OCI `revision=a6d1c53c068a5df421479cc9e8801f251f80d539`;
     `git merge-base --is-ancestor` YES for `c8400e452` (guard + relaxed validation)
     AND `f629f4530` (kill-switch), and correctly NO for a commit made after the build.
- **Phase 3 applied**: staged file moved in as **migration 457** (+ ROLLBACK written
  first), applied by hand + `--record-only`. Post-state verified at the artefact:
  config `handler_agent=''` / `status='needs_human_review'`; **0 rows anywhere in
  `site_work_items` carry `hitl-review`** (the one parked 08-12 row swept, stamped
  `result.handler_291`); 0 blocked at the phantom.
- **The guard EXERCISED, not merely deployed** — `PROBE_write_door_guard.sh` (new, in
  this dir), corr `a40d69b7-b91b-431f-9187-b7c8f0ee9220`, run `COMPLETED|finish`.
  Two `create_work_item` steps in ONE inline-workflow dispatch (no `agent_definitions`
  row written), both at `status='claimed'` — in the guard's trigger set but NOT
  dispatchable, so the probe structurally cannot cause real work to run:
  - **TEST** (`zzz-unregistered-probe-291`, precondition-checked as unregistered):
    born **`blocked`**, `error='Handler agent not registered: zzz-unregistered-probe-291'`,
    `claimed_at IS NULL` — **blocked AT WRITE, never claimed**. Under the pre-guard
    binary this row is born `claimed` with a NULL error, so the observable discriminates.
  - **CONTROL** (`tool-improver`, registered): born **`claimed`**, error NULL —
    the guard is selective, not a blanket block. Without this arm a
    block-everything bug would have read as a pass.
  - Both probe rows cancelled by the script at the end.
- **291 CLOSED**: fixed AND live on every axis — config (450/451/457), Go
  (`c8400e452`+`f629f4530` on v1.0.1307), and exercised in production.
- Residuals carried forward, unchanged and owned elsewhere: the dedup KEY-GRANULARITY
  loss (per-FINDING review keys — the 285 lane's follow-on); the ~41 raw-INSERT sites
  that bypass the door (claim remains their backstop); the five `'human-review'`
  pseudo-handler producers (inert; recorded in WDS-018).
