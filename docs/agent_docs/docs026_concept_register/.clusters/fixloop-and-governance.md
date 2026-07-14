# Cluster: fixloop-and-governance
Categories included: fix-loop, new:autonomy-governance, new:autonomous-build-operate, new:autonomy-trust-model, reasoning, new:investigation-discipline, new:operating-doctrine, new:operator-practice


<!-- SOURCE: U08_travelling_docs.md -->
### fix-proposer agent (F1.1a) — read-only proposal writer
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** FYI addendum 2026-07-10: "a fix-proposer agent (F1.1a) now exists".
- **what:** An agent in the diagnosis→fix loop that reads only orchestration_states/diagnosis_artifacts and writes only kind='fix_plan' artifacts — no code writes, no git token. Noted here as a boundary fact for the travelling-docs surface owners.
- **sources:** FYI_from_fixloop_2026-07-10_verdict_prompt_symptom_check.md#addendum
- **relations:** diagnosis loop; fix-loop workstream (primary docs elsewhere).
- **verify-later:** fix-proposer agent_definitions row; diagnosis_artifacts table.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Diagnosis→fix council loop (overall system)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "PR #1 — github.com/gqls/agentchassis/pull/1 — APPROVED & MERGED" (NOTES(10) Turn 26-28); SUMMARY_where_we_are_2026-07-13.md: "Today it did the whole thing, end to end, for the first time."
- **what:** The end-to-end pipeline that turns a plain-English bug symptom into a human-reviewed pull request: intake → read-only diagnosis (cite-or-abstain, three evidence tiers) → constrained fix plan → two-reviewer council with deterministic decision → dedicated-pod implementer behind a hard file allowlist → containerized build gate → PR. Every stage writes to `diagnosis_artifacts`/`orchestration_states` keyed on one correlation_id so the whole run is auditable. Human review is the terminal; nothing merges itself.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#THE TASK, fixloop_eg_dartsonline/MILESTONE_diagnosis_fix_loop_2026-07-10.md, fixloop_eg_dartsonline/SUMMARY_where_we_are_2026-07-13.md, fixloop_eg_dartsonline/README_so_far.md
- **relations:** all other fixloop concepts in this extraction are sub-scopes of this one
- **verify-later:** PR #1 on github.com/gqls/agentchassis; `diagnosis_artifacts` table contents on clients_db

<!-- SOURCE: U13_docs024_small_dirs.md -->
### needs_diagnosis intake route (F0.1c)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "F0.1c — ✅ LANDED 2026-07-09" (PLAN_fixloop_pilot.md §1)
- **what:** The one documented way a bug enters the loop: `090_TRIGGER_needs_diagnosis_v1.sh` writes a durable `site_work_items` row (`pipeline='diagnose'`, `item_type='needs_diagnosis'`, `status='awaiting_diagnosis'`) and fires the `diagnose-orchestrator` Kafka envelope on the same correlation_id, so the intake record, the diagnosis_artifacts bundles, and the terminal doc_notes row all join on one key. `DISPATCH=0` records without firing; the older `084_TRIGGER_diagnose_v1.sh` remains for ad-hoc runs with no intake record. `item_key` (`needs_diagnosis:<slug>`) plus `idx_swi_dedup` makes re-running the same slug idempotent while an intake is open.
- **sources:** fixloop_eg_dartsonline/090_TRIGGER_needs_diagnosis_v1.sh, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md#F0.1c, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Q-B CORRECTION
- **relations:** system.internal pseudo-site anchor pattern; private inert pipeline statuses pattern; diagnose-dispatch-loop
- **verify-later:** site_work_items rows with pipeline='diagnose'; idx_swi_dedup index definition

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Superseded: null-site-allowed intake design
- **category:** fix-loop
- **status-signal:** superseded
- **status-evidence:** "Q-B CORRECTION (2026-07-09... 'Null-site allowed' is impossible" (RUNBOOK(10)#Q-B CORRECTION); original design in RUNBOOK(9)#QUESTIONS "null-site allowed"
- **what:** The 2026-07-07 owner decision originally specified that site-less code bugs would "ride null-site" in the new diagnose pipeline namespace. Reading the live schema on 2026-07-09 showed this was structurally impossible twice over (NOT NULL column; site-anchored loader query), and it was replaced by the system.internal pseudo-site anchor pattern.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#QUESTIONS, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Q-B CORRECTION, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 4
- **relations:** system.internal pseudo-site anchor pattern
- **verify-later:** n/a — superseded, no live code to check

<!-- SOURCE: U13_docs024_small_dirs.md -->
### diagnosis_artifacts table (unified egress store)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "F0.1a — ✅ LANDED 2026-07-09" (PLAN_fixloop_pilot.md); kind list extended live via ALTER in 0NN_fix_proposer.sql v4/v5
- **what:** The durable, correlation_id-keyed egress table for the whole loop, with `kind` growing over time: `bundle` and `iteration_note` (F0.1a) → `fix_plan` (F1.1a) → `council_report` (F2.1) → `escalation` (F2.3). `correlation_id` is deliberately `text` not `uuid` (ExecutionContext.CorrelationID has no guaranteed form). A partial unique index on `(correlation_id, iteration) WHERE kind='bundle'` gives retry-safe upsert for bundles while allowing multiple `iteration_note` rows per iteration. Carries a retention knob (`expires_at`/`pinned`) that is defined but has no sweep implemented yet.
- **sources:** fixloop_eg_dartsonline/0NN_diagnosis_artifacts.sql, fixloop_eg_dartsonline/0NN_fix_proposer.sql#§1, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §F0.1a
- **relations:** bundle write-through; fix-proposer's fix_plan artifacts; council_report; escalation artifact; retention sweep (unbuilt)
- **verify-later:** table DDL and CHECK constraints on clients_db; whether a retention sweep job exists

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Retention/expiry knob on diagnosis_artifacts
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** "Retention knob. NULL = keep indefinitely. Sweep deletes WHERE expires_at < now() AND NOT pinned" (0NN_diagnosis_artifacts.sql comment) — no sweep job found in scope
- **what:** `expires_at`/`pinned` columns exist on diagnosis_artifacts (bundles configured to expire sooner than notes; NULL = keep forever), and a partial index exists to support a future deletion sweep, but no sweep job/scheduled task was found anywhere in this extraction's file set. The mechanism is designed but not built.
- **sources:** fixloop_eg_dartsonline/0NN_diagnosis_artifacts.sql#idx_diagnosis_artifacts_expiry
- **relations:** diagnosis_artifacts table
- **verify-later:** search codebase/scheduled_tasks for any retention-sweep job referencing diagnosis_artifacts

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Known-answer benchmark methodology
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "Benchmark method validated end-to-end: same symptom, one variable cluster, measurable delta" (NOTES(10)#Turn 10)
- **what:** When a candidate bug's mechanism is dissolved by the mandatory cheap pre-check (three candidates in a row were), the pilot is not discarded but promoted from a "discovery run" to a "known-answer benchmark": the loop is run blind on the original symptom string and its output is scored against a pre-registered rubric of must/should/bonus claims fixed before the run, including a "refutation credit" penalizing confirmation of a known-false standing hypothesis. This produced a repeatable, gradable evaluation across five runs, each of which found and fixed a real engine defect.
- **sources:** fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §0, §3, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#AMENDMENT 2026-07-09, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 1
- **relations:** loop-worthiness test doctrine; blinding discipline; dartsonline guides defect (benchmark bug)
- **verify-later:** rubric table in PLAN_fixloop_pilot.md §3

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Abandoned pilot candidates (chrome bug, roadmap-phase gap, blank-guides fork)
- **category:** fix-loop
- **status-signal:** abandoned
- **status-evidence:** "Chrome pilot EVAPORATED (fixed live)"; "PREVIOUS: F0 PILOT #1 — DOWNGRADED... root cause found; not diagnosis-shaped" (RUNBOOK(9)#F0 PILOT ORIGINAL RECORD, #SUPERSEDED CANDIDATE 2)
- **what:** Three earlier pilot candidates were considered and dropped before the dartsonline guides bug was chosen: (1) dartsonline pages lacking site chrome (fixed before the loop ran); (2) a "no submission path produces a roadmap" gap, reclassified as a known platform gap rather than a diagnosis target since a human found it by reading two files; (3) a blank guides-index fork where the guide-writing mechanism's existence was unverified. Recorded because dropped/dissolved candidates are exactly the kind of idea worth remembering for future taxonomy work.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#F0 PILOT ORIGINAL RECORD, #SUPERSEDED CANDIDATE 2, #PREVIOUS F0 PILOT #1
- **relations:** loop-worthiness test doctrine; roadmap/phases mechanism
- **verify-later:** n/a — historical, dissolved candidates

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Blinding discipline for benchmark runs
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "BLINDING IS MANDATORY... Exclude .../fixloop_eg_dartsonline/ from the loop's corpus" (RUNBOOK(10)#★ F0 PILOT)
- **what:** Established that the diagnose-agent workflow structurally cannot read this docs directory (it walks Go source and DB rows only), so blinding is largely automatic; the only two leak vectors are the symptom string (must describe only observable behaviour) and `seed_scope` (must be omitted entirely for a benchmark run).
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#REGENERATING THE CONTEXT BUNDLE §BLINDING, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §3
- **relations:** known-answer benchmark methodology
- **verify-later:** diagnose-agent workflow JSON — confirm no doc-reading step was later added

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Standing hypothesis refuted (reconcile_site_plan routing table)
- **category:** fix-loop
- **status-signal:** superseded
- **status-evidence:** "THE STANDING HYPOTHESIS IS REFUTED — and it named the wrong file" (NOTES(10)#Turn 1)
- **what:** The original hypothesis (from RUNBOOK(9)) blamed `reconcile_site_plan`'s routing table for silently dropping "guide" pages. Hand-diagnosis on 2026-07-09 showed the routing table is real but lives in `WriteBuildItemsAction`, and absence from it does not drop a page (it defaults to page-build-handler); the actual drop mechanism is a separate `unavailableBuilders` map, and `reconcile_site_plan_action.go` has no type switch at all. Retained specifically because a hypothesis a loop should refuse to confirm is exactly the kind of "superseded" idea the register wants captured.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#★ F0 PILOT, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 1
- **relations:** dartsonline guides defect; two intake paths disagreement
- **verify-later:** grep/inspect `reconcile_site_plan`; `WriteBuildItemsAction`; `unavailableBuilders`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Two intake paths disagreement (WriteBuildItemsAction vs reconcile_site_plan)
- **category:** fix-loop
- **status-signal:** deployed (documented finding, not fixed — routed to builder thread)
- **status-evidence:** "Fourth finding (unlooked-for): the two intake paths disagree" (NOTES(10)#Turn 1)
- **what:** `WriteBuildItemsAction` deliberately skips `tool`/`entity-directory`/`entity-page` page types via an `unavailableBuilders` guard, while `reconcile_site_plan_action.go` hardcodes `handler_agent='page-build-handler'` for every plan page with no type switch at all, so it re-emits build items for the very types the other path skips. Flagged as a builder-thread decision, not fixed here.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 1, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#★ F0 PILOT
- **relations:** dartsonline guides defect; pipeline-blind dispatch surfaces
- **verify-later:** grep/inspect `WriteBuildItemsAction`; `tool`; `entity-directory`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### mark_no_sections — referenced-but-never-built step
- **category:** fix-loop
- **status-signal:** abandoned
- **status-evidence:** "`mark_no_sections` does not exist... appears nowhere in the repo but that one comment" (NOTES(10)#Turn 1)
- **what:** A code comment at `load_work_item_actions.go:750-756` names a remedy step `mark_no_sections` that would flag a sectionless page's work item `needs_human_review` — but the step was never implemented; it appears nowhere else in the repo. The completion guard that would consume its flag faithfully preserves a flag nothing ever sets.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 1, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#★ F0 PILOT
- **relations:** dartsonline guides defect
- **verify-later:** grep for mark_no_sections in current repo

<!-- SOURCE: U13_docs024_small_dirs.md -->
### fix-proposer agent / constrained edit plan (F1.1a)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "F1.1a — ✅ BUILT 2026-07-10" (PLAN_fixloop_pilot.md); live agent seeded in 0NN_fix_proposer.sql
- **what:** The first F1 slice: an `agent_definitions` workflow that loads a diagnosis by `fix_correlation_id`, refuses anything not `CONFIRMED`, and drafts a constrained edit plan (summary, ≤8 allowlisted edits with file/symbol/operation/rationale/sketch, `grounded_in` quotes required, risks) persisted to `diagnosis_artifacts` (`kind='fix_plan'`). It writes no code and needs no git token.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#propose step, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §3b F1.1a
- **relations:** diagnose_persist_fix_plan validator; two-reviewer council; CONFIRMED gate
- **verify-later:** fix-proposer agent_definitions row, workflow version

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Plan validation / hard allowlist for edit plans (diagnose_persist_fix_plan)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "structural validation... UNLIKE the bundle write-through this FAILS the step on bad input" (NOTES(10)#Turn 15)
- **what:** A deterministic Go action validating a proposer's plan before persisting it: non-empty summary/edits/rationale/sketch, operations restricted to `modify|add|remove|config_change`, repo-relative paths only, ≤8 edits, `grounded_in` quotes required, 32KB cap — plus (F1.1b(a)) rejection of explicit no-op phrases so edits that change nothing cannot pass as real edits. Fails closed; fired correctly on the first two real fix-proposer runs (max_tokens truncation).
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 15, #Turn 16, #Turn 18
- **relations:** fix-proposer agent; hard file allowlist (implementer's analogous gate)
- **verify-later:** grep/inspect `modify|add|remove|config_change`; `grounded_in`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Two-reviewer council (F2.1)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "F2.1 — ✅ PROVEN LIVE 2026-07-10" (PLAN_fixloop_pilot.md)
- **what:** Two sequential LLM reviewer steps plus a deterministic Go decision — not a third model opinion about two model opinions. `review_editquality` judges whether every edit changes something real and targets the actual causal path; `review_guardian` judges blast radius, architecture-change signals, and surface ownership, and alone holds the hard veto. Both attach optional `checks:[{sql,why}]`. First live run judged real objections correctly rather than rubber-stamping.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#review_editquality, #review_guardian, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 18-19
- **relations:** deterministic council decision + hard veto; verify step; schema hint for reviewers
- **verify-later:** grep/inspect `review_editquality`; `review_guardian`; `checks:[{sql,why}]`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Deterministic council decision + hard veto (diagnose_council_decide)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "diagnose_council_decide: ordered rules (hard veto → rejected; any veto → rejected; any objection → revise; all approve → approved)" (NOTES(10)#Turn 18)
- **what:** A pure Go action aggregates the two reviewers' verdicts deterministically, auditable and reproducible. `hard_veto_from` is currently a flag in the workflow step config naming the guardian reviewer as sole veto-holder; whether it should instead live on a reviewer/pipeline definition column remains an open sub-question. Malformed reviewer output fails closed.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#council_decide, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#STILL OPEN Q-D
- **relations:** two-reviewer council; hard_veto flag at multiple scopes (superseded early design); decision router (F2.3)
- **verify-later:** grep/inspect `hard_veto_from`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### hard_veto flag at multiple scopes (early design, superseded)
- **category:** fix-loop
- **status-signal:** superseded
- **status-evidence:** owner decision 2026-07-07: "hard_veto flag, attachable at multiple scopes (a reviewer agent, a pipeline, a specific tool/component)" (RUNBOOK(9)#Q-D) — actual shipped implementation is a single `hard_veto_from` list in one workflow step's config
- **what:** The original council design envisioned a hard_veto flag placeable at reviewer, pipeline, tool, or component scope, motivated by accessibility/legal review cases. What was actually built (F2.1) is narrower: a single `hard_veto_from: ["guardian"]` array in the fix-proposer's `council_decide` step config. The broader multi-scope placement remains an open sub-question of Q-D.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#Q-D, fixloop_eg_dartsonline/0NN_fix_proposer.sql#council_decide config
- **relations:** deterministic council decision + hard veto; future council roster
- **verify-later:** grep/inspect `hard_veto_from: ["guardian"]`; `council_decide`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Revise loop (F2.2)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "F2.2 REVISE LOOP — ✅ BUILT 2026-07-10" (PLAN_fixloop_pilot.md)
- **what:** On a `revise` decision with rounds remaining, feeds the diagnosis, prior plan, and both reviewers' objections back into a `repropose` step, then re-validates and re-reviews (capped, default 2, later 3). Exhausting the cap becomes a distinct terminal state `exhausted`. Round counting is scoped per `orchestration_id` (proposer run), not per `correlation_id` — an earlier per-correlation design flaw was caught mid-implementation and fixed, then found still live on the deployed binary for one further round (round-counting scope bug).
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#council_decide/check_revise, fixloop_eg_dartsonline/HANDOFF_turn21_2026-07-10.md#Session 20 Summary
- **relations:** decision router (F2.3); round-counting scope bug
- **verify-later:** grep/inspect `revise`; `repropose`; `exhausted`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Decision router (F2.3)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "F2.3 DECISION ROUTER + VERIFY + REFRAME + ESCALATION — ✅ CODE BUILT 2026-07-10... PROVEN live" v1.0.1108 (PLAN_fixloop_pilot.md)
- **what:** Replaces the single revise/complete branch with a full router: `approved`→complete; `revise` with rounds left→verify checks→repropose; `rejected` first time with rounds left→reframe-once; `rejected` again or exhausted→escalate. Motivated by two clean benchmark runs exposing two dead-ends. Flags are computed by a pure, directly-tested Go function `applyCouncilCaps`.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#check_approved/check_rejected/check_reframe/check_revise, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 22
- **relations:** revise loop; verify step; reframe step; escalation artifact
- **verify-later:** grep/inspect `approved`; `revise`; `rejected`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Verify step (diagnose_run_checks)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "Verify step ran 7 reviewer checks under the containment" (NOTES(10)#Turn 23)
- **what:** Reviewers attach `checks:[{sql,why}]` (SELECT/WITH only) to their verdicts; this action executes them under the same read-only containment the diagnosis loop's data_requests use (lint → READ ONLY transaction → statement_timeout → EXPLAIN gate → capped rows), and feeds results into the next repropose so fact-shaped objections are settled with evidence instead of another blind revision round. Capped at 8 checks by default.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#run_checks, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 22, #Turn 23
- **relations:** decision router; schema hint for reviewers
- **verify-later:** grep/inspect `checks:[{sql,why}]`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Schema hint for reviewers (F2.3b(a))
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "Verification run 1e221fb7: 8 checks run, 0 failures (prior run: 5 of 7 failed on hallucinated schema)" (NOTES(10)#Turn 24)
- **what:** A `load_schema_hint` query_database step pulls the live table/column list from `information_schema` at run time, and both reviewer prompts get this hint plus two named traps (workflow steps live in `agent_definitions.default_config` jsonb, not a steps table; a site's domain lives on `sites` joined via `pages.site_id`). Fixes a discovered defect where 5 of 7 reviewer-written verification SQL checks failed on hallucinated columns/tables.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#load_schema_hint, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 23, #Turn 24
- **relations:** verify step; two-reviewer council
- **verify-later:** grep/inspect `load_schema_hint`; `information_schema`; `agent_definitions.default_config`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Reframe step (post-veto)
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** "Reframe path is unit-tested but has never fired live (no veto since v4)" (HANDOFF_CURRENT_fixloop.md#State snapshot)
- **what:** After a guardian veto with rounds remaining, makes one attempt to reframe rather than reproposing the same shape: either a strictly narrower remediation (site-scoped interim fix allowed only if risks names the deferred structural fix) or an explicit "needs architecture review" declaration plus a minimal safe interim step. Capped at one attempt; built and unit-tested but has not fired live since v4 shipped.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#reframe, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 22, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md
- **relations:** decision router; escalation artifact; guardian veto (dartsonline run 8c770fd5)
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Escalation as first-class success terminal (diagnose_escalate)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "Escalation package persisted (kind=escalation)... The dead-end is now a hand-off" (NOTES(10)#Turn 23)
- **what:** When a plan is rejected-again or a revise budget is exhausted, the run persists a `kind='escalation'` artifact (decision, reason, round, diagnosis conclusion, final plan, both reviews) and completes via a distinct `complete_escalated` success terminal. Explicitly designed so "needs a human/architecture review" is a correct, successful output rather than a failure.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#escalate/complete_escalated, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 22
- **relations:** decision router; reframe step; dartsonline guides defect
- **verify-later:** grep/inspect `kind='escalation'`; `complete_escalated`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Write step / fix-implementer agent (F1.1b(c))
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "F1.1b(c): branch + PR — ✅ COMPLETE & PROVEN 2026-07-13 (PR #1 opened & merged)" (PLAN_fixloop_pilot.md)
- **what:** The loop's write organ. Given a `fix_correlation_id`, refuses anything whose latest council decision is not `approved`; reads current file bodies via the GitHub contents API; runs an LLM step (`sketch_to_files`) to turn the approved plan's sketches into complete new file bodies for ONLY the plan's named files; passes those through a deterministic hard allowlist; creates a `fix/<short-corr>` branch and commits via the git-adapter; gates on a build check; and on green opens a PR. `config_change` edits are deliberately NOT implemented by this agent.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §3b F1.1b(c), fixloop_eg_dartsonline/README_so_far.md
- **relations:** hard file allowlist; build gate; git-adapter write isolation; fix-implementer-orchestrator; PR as human terminal
- **verify-later:** grep/inspect `fix_correlation_id`; `approved`; `sketch_to_files`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Hard file allowlist (diagnose_prepare_fix_commit)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "Part 2a BUILT (chassis, commit a4c6cc63)... 7-case suite exercises the real logic" (NOTES(10)#Turn 25)
- **what:** A deterministic action sitting between the implementer's LLM step and the git-adapter: the approved plan's modify/add file list is a hard allowlist — a produced file outside the plan, a plan file the implementation is missing, or an empty/duplicate/no-op file all reject the whole implementation before anything touches git. Also assembles the branch name, commit message, and PR title/body (the "Q-H package"). This is the safety core that made the first live PR's diff exactly the approved plan.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql#prepare step, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 25, fixloop_eg_dartsonline/README_so_far.md
- **relations:** write step; build gate; plan validation
- **verify-later:** validateImplementation function; 7-case test suite

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Build gate (diagnose_build_gate)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "Build gate (golang Job): GREEN — === build gate: PASS ===" (README_so_far.md); "its first red correctly blocked a PR" (NOTES(10)#Turn 26-28)
- **what:** Before any PR is opened, changes must be built in a clean container (`gofmt` + targeted `go build`) in a short-lived golang-image k8s Job. Green routes to PR creation; red routes to a no-PR terminal with build log attached, branch left for human inspection — "no PRs for broken code." Chosen over GitHub Actions CI on the PR (Option B: broken implementations must never even become a visible PR). Its first live red catch was a genuine pre-existing bug, then fixed for real.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql#build_gate/check_gate, fixloop_eg_dartsonline/SUMMARY_write_step_position_2026-07-12.md#Option A/B/C, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 26-28
- **relations:** write step; hard file allowlist; PR as human terminal
- **verify-later:** diagnose_build_gate action; RBAC rbac-job-spawner.yaml pods/log grant

<!-- SOURCE: U13_docs024_small_dirs.md -->
### git_adapter_request generic adapter caller
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "git_adapter_request — ONE generic adapter caller (allowlisted verbs...)" (HANDOFF_CURRENT_fixloop.md#F1.1b(c) CODE COMPLETE)
- **what:** A single generic workflow action used for all git-adapter calls from the write step, with the adapter action name and data fields/literals supplied per-step from config, and an explicit note that `delete_repo` is unreachable through this path.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql#create_branch/commit_files/create_pr config, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md
- **relations:** git-adapter new actions; write step
- **verify-later:** grep/inspect `delete_repo`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### isRepoCloningAgent spawn gate / GITHUB_READ_TOKEN injection
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "the spawned implementer pod gets GITHUB_READ_TOKEN via the already-deployed isRepoCloningAgent gate" (HANDOFF_CURRENT_fixloop.md)
- **what:** An existing spawn-gate mechanism (already used for diagnose-agent) that injects a read-only GitHub token into a dedicated, ephemeral pod when the spawned agent type is listed in `isRepoCloningAgent`. `fix-implementer` was added to this list. Only works when the agent runs as a dedicated spawned pod — the generic in-chassis orchestrate path bypasses the gate entirely.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql#header point 1, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#FIRST END-TO-END RUN blocked
- **relations:** fix-implementer-orchestrator; git-adapter as sole write credential holder; diagnose_read_repo_files
- **verify-later:** grep/inspect `isRepoCloningAgent`; `fix-implementer`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### diagnose_read_repo_files action
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "diagnose_read_repo_files — plan's modify/add files via GitHub contents API (raw media type; read token from spawn gate; modify-404 = hard error)" (HANDOFF_CURRENT_fixloop.md)
- **what:** Fetches the current bodies of the approved plan's modify/add files via the GitHub contents API at an explicit ref, using the token from the spawn gate. A missing file for a "modify" operation is a hard error.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql#read_current_files, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md
- **relations:** isRepoCloningAgent spawn gate; write step
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### fix-implementer-orchestrator (dedicated-pod wrapper)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "0NN_fix_implementer_orchestrator.sql — F1.1b(c) fix: run the implementer as a DEDICATED POD" (header)
- **what:** A thin wrapper agent (`spawn_agent(fix-implementer)` → `call_agent` → `complete`) built to fix a real first-run failure: firing fix-implementer via the generic orchestrate path ran it IN the shared chassis pod, so the isRepoCloningAgent gate never fired. Mirrors the existing diagnose-orchestrator→diagnose-agent pattern exactly. Needed no image rebuild.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer_orchestrator.sql, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#FIRST END-TO-END RUN, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 26-28
- **relations:** isRepoCloningAgent spawn gate; write step
- **verify-later:** 092_TRIGGER_fix_implementer_v1.sh target agent type

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Whole-file rewrite strategy (implementer's LLM step)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "41KB whole-file rewrite → allowlist PASS" (README_so_far.md)
- **what:** The implementer's `sketch_to_files` LLM step outputs the COMPLETE new body of every plan-named file, never a diff/patch, with hard rules forbidding drive-by changes. Explicitly named as not scaling to very large files (32000 max_tokens gives headroom for one ~41KB file but not much more) — a diff/patch strategy is logged as future work (F1.2).
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql#sketch_to_files, fixloop_eg_dartsonline/README_so_far.md
- **relations:** write step; hard file allowlist
- **verify-later:** grep/inspect `sketch_to_files`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### PR as human terminal / nothing merges itself
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "PR — waits for you. Nothing merges itself." (README_so_far.md)
- **what:** A governing design principle: the platform's most autonomous act is opening a pull request; nothing ever merges its own work. Isolation model (2026-07-12): fix/* branches live on the same repo (no fork); the owner alone chooses what merges to main. This is why "escalation" is treated as a success, not a failure.
- **sources:** fixloop_eg_dartsonline/README_so_far.md, fixloop_eg_dartsonline/0NN_fix_implementer.sql#header, fixloop_eg_dartsonline/SUMMARY_where_we_are_2026-07-13.md
- **relations:** write step; NO FORK decision; escalation as first-class success terminal
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### NO FORK decision (abandoned fork idea)
- **category:** fix-loop
- **status-signal:** superseded
- **status-evidence:** "Decisions CLOSED... 4. NO FORK: isolation = fix/* branches + owner-gated merges on this repo" (HANDOFF_CURRENT_fixloop.md)
- **what:** The owner raised, and then explicitly closed, the idea of running the fix-loop's next phase against a separate forked repository. The decision landed instead on branch+PR isolation on the same repo, which is what was actually built.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 25, #Turn 26, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#Decisions CLOSED
- **relations:** PR as human terminal; write step
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Round-counting scope bug (correlation vs orchestration)
- **category:** fix-loop
- **status-signal:** superseded (fixed in source; deploy gap tracked separately)
- **status-evidence:** "the deployed v1.0.1107 binary counts council rounds per correlation... does NOT carry the orchestration_id-scoping fix" (NOTES(10)#Turn 22)
- **what:** Council-round counting was originally scoped per `correlation_id`, accumulating council_report rows across every proposer re-run — so a fresh proposer run on a correlation with review history would start mid-count and exhaust its revise budget without ever reproposing. Fixed in source to count per `orchestration_id`, but a same-tag deploy trap meant the fix did not reach the running binary for one further benchmark cycle.
- **sources:** fixloop_eg_dartsonline/HANDOFF_turn21_2026-07-10.md#Key accomplishments, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 22
- **relations:** revise loop; same-tag deploy trap gotcha
- **verify-later:** grep/inspect `correlation_id`; `orchestration_id`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### fixloop-digest / awareness surface
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** "IN FLIGHT (turn 29): the AWARENESS SURFACE... Built + committed, awaiting the next chassis image" (HANDOFF_CURRENT_fixloop.md)
- **what:** A deterministic (no-LLM-in-path) digest agent composing a window (default 24h) summary of fix-loop activity — status/terminal/gate/PR outcomes, decisions per correlation, and agent_definitions_backup snapshots — persisted to `doc_notes` (categories `["digest","fixloop"]`). Built to satisfy the owner's standing rule "more awareness before wider autonomy." v1 is manual-trigger only; a daily cadence is deliberately deferred.
- **sources:** fixloop_eg_dartsonline/0NN_fixloop_digest.sql, fixloop_eg_dartsonline/093_TRIGGER_fixloop_digest_v1.sh, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#IN FLIGHT
- **relations:** owner standing rule; diagnosis_artifacts table; future council roster
- **verify-later:** whether the chassis image carrying `fixloop_digest` action has shipped; `doc_notes` rows with categories ? 'digest'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Owner standing rule: awareness before autonomy
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "Owner standing rule (2026-07-12): 'more awareness BEFORE wider autonomy.'" (0NN_fixloop_digest.sql header)
- **what:** An explicit governance principle: before the council is widened with more reviewer perspectives or migration/feature-building agents, the owner must first have a reliable way to see what the loop has been doing and deciding. Directly produced the fixloop-digest slice being scheduled ahead of the F2 roster expansion.
- **sources:** fixloop_eg_dartsonline/0NN_fixloop_digest.sql#header, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#Decisions CLOSED, fixloop_eg_dartsonline/SUMMARY_where_we_are_2026-07-13.md
- **relations:** fixloop-digest; future council roster (deferred by this rule)
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Future council roster (aspirational reviewers)
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** "Initial roster: a guidelines agent... a reuse agent... a bug-historian... a compliance/legal eye; pipeline guardians... specialist knowledge agents" (RUNBOOK(9)#THE TASK) — none beyond the guardian/edit-quality pair were built
- **what:** The original council vision named a much wider roster than what was built: a guidelines agent, a reuse agent, a bug-historian, a compliance/legal eye, one pipeline-guardian per master workflow, and specialist knowledge agents — motivated by a real incident where a chat reinvented a trigger+triage SQL pair that already existed. Only a generic edit-quality reviewer and a single cross-pipeline guardian shipped.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#THE TASK, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#Direction
- **relations:** two-reviewer council; owner standing rule; architecture-change visibility (Q-E)
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Architecture-change visibility (Q-E signals)
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** "Q-E architecture-change signals: ... STILL OPEN" (RUNBOOK(10)#STILL OPEN)
- **what:** A standalone goal from the original task charter — make it loud when a proposed change is accidentally fundamental (touching platform contracts, message shapes, many packages, exported signatures) before it ships. Never built as a dedicated formal detector; what exists in practice is the pipeline-guardian reviewer's informal judgement, which has correctly identified architecture-level changes dressed as contained fixes.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#THE TASK, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#STILL OPEN Q-E
- **relations:** two-reviewer council; guardian veto; future council roster
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Guardian veto surfacing an architecture-level fix (dartsonline)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "the guardian vetoed all three edits as 'an architecture change dressed as a contained fix'" (NOTES(10)#Turn 22, orch 8c770fd5)
- **what:** A concrete, live-observed instance of the guardian reviewer correctly recognizing that a minimal-looking three-edit plan was actually architecture-level, vetoing it and proposing a safer alternative in its notes — deliberately NOT auto-applied since it fixes only one site while leaving the platform-wide cause live everywhere.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 22
- **relations:** architecture-change visibility; reframe step; platform-not-site-data fix philosophy; dartsonline guides defect
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Platform-not-site-data fix philosophy
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "Owner ruled: the F1 edit plan targets the PLATFORM, not dartsonline's data." (NOTES(10)#Turn 2)
- **what:** An owner ruling that any fix plan must target the platform mechanism rather than a single site's data rows — because the causes of the benchmark bug are relay-level and a data-only fix would fix one site while leaving every other site exposed. Directly shapes the proposer's prompt rules and the guardian's refusal to accept a scoped data-only remediation as a final answer.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 2, fixloop_eg_dartsonline/0NN_fix_proposer.sql#propose prompt rule 1
- **relations:** dartsonline guides defect; guardian veto surfacing architecture-level fix; reframe step
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### config_change edit operation type
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "config_change edits in a plan are NOT implemented by this agent... the PR body carries them for the human" (0NN_fix_implementer.sql header)
- **what:** A plan-edit operation type reserved for edits that target `agent_definitions` workflow-JSON configuration rather than repo files. The proposer's prompt requires such edits be explicitly labelled, but the fix-implementer deliberately does not apply them — they are left in the PR body for a human to apply by hand.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#propose prompt rule 5, fixloop_eg_dartsonline/0NN_fix_implementer.sql#header
- **relations:** write step; hard file allowlist; plan validation
- **verify-later:** grep/inspect `agent_definitions`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### F1.2 deferred work items (ref/base as input; fix_pr artifact; diff strategy)
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** "Open (F1.2): ref/base are live-set to the active working branch because origin/main is stale — make them a per-run INPUT" (PLAN_fixloop_pilot.md)
- **what:** A cluster of known-but-deferred improvements: the implementer's git ref/base/from_branch are hardcoded via a live jsonb_set patch rather than a per-run input field; a dedicated `kind='fix_pr'` diagnosis_artifacts row for the PR result is deferred; a diff/patch implementation strategy for large files.
- **sources:** fixloop_eg_dartsonline/PLAN_fixloop_pilot.md#F1.1b(c), fixloop_eg_dartsonline/0NN_fix_implementer.sql#header
- **relations:** whole-file rewrite strategy; write step
- **verify-later:** grep/inspect `kind='fix_pr'`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### F3 learning record / bug_records (never built)
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** "F3 — Learning... bug_records (category taxonomy, recurrence checks feeding the historian)" (RUNBOOK(10)#Phased plan) — no bug_records table or historian agent found
- **what:** The original phased plan's final stage: categorize confirmed bugs into a taxonomy so recurring classes are caught earlier, feed guideline-amendment proposals to the human, and enrich the corpus from what the loop learns. Never designed in detail or built.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Phased plan F3, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#Phased plan F3
- **relations:** future council roster (bug-historian); guideline-gap side-task mechanism
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Guideline-gap side-task mechanism (designed, unclear if built)
- **category:** fix-loop
- **status-signal:** unknown
- **status-evidence:** "Q-D completion — guideline-gap = SIDE-TASK (does not block the fix): a work item carrying the evidence; handler drafts a concrete amendment and opens a PR against the GUIDELINE DOCS" (NOTES(9)#DECISIONS)
- **what:** A 2026-07-07 decision that when a reviewer finds the guidelines themselves fell short, that finding becomes a side-task work item whose handler drafts a guideline-amendment PR, with gaps accumulating toward the F3 learning record. No implementation of this side-task handler was found in the files read.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(9).md#DECISIONS, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#QUESTIONS Q-D
- **relations:** F3 learning record; future council roster
- **verify-later:** search for a guideline-amendment work item type / handler agent in agent_definitions

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Q-G reviewer context (open design question, v1 answered narrowly)
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** "Q-G v1 = role prompts + plan + diagnosis (no per-reviewer corpora yet)" (PLAN_fixloop_pilot.md)
- **what:** The open question of how much context each council reviewer should see. What shipped is the narrowest option: both reviewers get the same role prompt, the persisted plan, the diagnosis conclusion, and (from F2.3b) a live schema hint — no per-reviewer curated corpus exists yet.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#STILL OPEN Q-G, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §F2.1
- **relations:** two-reviewer council; schema hint for reviewers; future council roster
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Q-H human-facing result package
- **category:** fix-loop
- **status-signal:** deployed (v1)
- **status-evidence:** "PR carrying the Q-H package" appears repeatedly as delivered (HANDOFF_CURRENT_fixloop.md)
- **what:** The decided shape of what a human ultimately sees: the PR body carries the diagnosis conclusion, the approved plan, and the council's decision/reviews together, so a human reviewing a fix-loop PR never has to go hunting through diagnosis_artifacts. The equivalent package for an escalated run is the escalation artifact.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql#prepare/create_pr, fixloop_eg_dartsonline/README_so_far.md
- **relations:** write step; escalation as first-class success terminal
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### SEED_first_writestep_diagnosis pattern (hand-authored diagnosis for downstream testing)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "the diagnosis is hand-written (the diagnosis LOOP is separately proven); proposer, council, and implementer that consume it all run for real" (SEED_first_writestep_diagnosis.sql header)
- **what:** A reusable technique for exercising downstream stages honestly without waiting for a live CONFIRMED diagnosis on a suitable bug: hand-author a CONFIRMED `orchestration_states` row for a real, tiny, zero-risk defect, then run the real proposer→council→implementer chain against it for real. Fabricating evidence rows was explicitly rejected as an option.
- **sources:** fixloop_eg_dartsonline/SEED_first_writestep_diagnosis.sql, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#First-run design decision
- **relations:** write step; tier-coverage guard
- **verify-later:** grep/inspect `orchestration_states`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### F0.3 per-iteration notes (never built)
- **category:** fix-loop
- **status-signal:** abandoned
- **status-evidence:** "Per-iteration notes — NOT MET, because F0.3 does not exist yet" (RUNBOOK(10)#F0 plumbing criteria)
- **what:** One of F0's four original acceptance criteria — writing the loop's per-iteration/per-step reasoning into task-specific running notes — was designed but never implemented across the entire workstream. The `diagnosis_artifacts.kind='iteration_note'` column value exists specifically to carry this, unused.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Phased plan F0.3, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md
- **relations:** doc_notes/travelling-docs integration boundary; diagnosis_artifacts table
- **verify-later:** grep/inspect `diagnosis_artifacts.kind='iteration_note'`

<!-- SOURCE: U14_docs019_runbooks.md -->
### Diagnosis→Fix Loop programme (F0–F3)
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "DISCUSSION COMPLETE for F0/F1 (2026-07-07): Q-A/B/C/D/F all decided — CUTOVER-READY … First slice: F0.1"; no build claimed.
- **what:** The v2 workstream turning the read-only diagnosis loop into a diagnosis→fix system, phased: F0 intake/observability/egress (documented route in and out, fetchable bundles, per-task running notes); F1 fix-on-a-branch; F2 council of reviewers + decision-maker with architecture-change visibility; F3 learning (bug records, guideline amendments, corpus enrichment). Mission: use everything available — code corpus, schemas, runtime, the guidelines themselves — with checks, balances and second opinions built in.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#the-task; docs019/RUNBOOK_diagnosis_fix_loop(9).md#phased-plan; docs019/RUNBOOK_diagnosis_fix_loop(9).md#current-position
- **relations:** read-only diagnosis loop (the base); council of reviewers; docs026 stage-3 council agents (this register's own consumer)
- **verify-later:** diagnosis_artifacts migration; needs_diagnosis items; fixer agent existence

<!-- SOURCE: U14_docs019_runbooks.md -->
### diagnosis_artifacts bundle egress
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "DECIDED 2026-07-07 (owner): Q-A diagnosis_artifacts table, written through inside assemble (unified-table refinement: kind ∈ {bundle, iteration_note})".
- **what:** Durable per-iteration bundle persistence: a diagnosis_artifacts table written through inside the assemble action (zero workflow-shape change, deliberately off the tools chat's emit-adjacent surface), with a documented fetch route. doc_notes was considered and set aside (notes are prose for humans; bundles are machine-replayable evidence with different retention). Sizing memory: bundles ~60KB × ≤5 iterations vs the 1.27MB collected_data incident.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-A); docs019/RUNBOOK_diagnosis_fix_loop(9).md#collision-surface
- **relations:** oversize delivery doctrine; per-task running notes; diagnose_assemble_bundle
- **verify-later:** diagnosis_artifacts table (exists?); assemble write-through code

<!-- SOURCE: U14_docs019_runbooks.md -->
### needs_diagnosis intake in a pipeline='diagnose' namespace
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "Q-B needs_diagnosis item in a NEW pipeline='diagnose' namespace (null-site allowed; envelope extends 084; manual trigger retained)"; "ENABLER CONFIRMED 2026-07-07: anchorless (site-less) diagnosis runs now SURVIVE".
- **what:** Task input rides the existing work-item dispatch + immune system: a needs_diagnosis site_work_items row in its own pipeline namespace, with null-site allowed for pure code bugs (enabled by the tools chat's load_runtime error-routing so anchorless runs degrade gracefully — ~26 min / 5 iterations observed). The canonical envelope adopts/extends the tools chat's 084 trigger with subject_type/subject_key.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-B); docs019/RUNBOOK_diagnosis_fix_loop(9).md#what-already-exists
- **relations:** build pump + immune system (the ride); generic envelope trigger (retained manual path)
- **verify-later:** pipeline='diagnose' rows; 084_TRIGGER_diagnose_v1.sh subject fields

<!-- SOURCE: U14_docs019_runbooks.md -->
### Fix-on-a-branch with an isolated fixer agent
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "Q-C separate fixer agent (isolated write token; constrained edit plan; gofmt+build in a spawned job pre-PR)" — decided 2026-07-07, not built.
- **what:** F1: a CONFIRMED diagnosis drives a proposed fix committed to a separate git branch via the git adapter, PR opened, human amends/ditches/applies. The loop's core stays read-only; the write surface is a SEPARATE fixer agent holding the only write token (the spawn token-gate pattern), producing a constrained edit plan validated by gofmt+build in a spawned job before the PR.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#the-task (item 2); docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-C)
- **relations:** repo-cloning token gate (the pattern); council of reviewers (gate before finalising)
- **verify-later:** fixer agent definition; git-adapter write paths

<!-- SOURCE: U14_docs019_runbooks.md -->
### Council of reviewers with a decision-maker
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) F2 "Independent reviewers (roster above), each a small agent with its own curated context … a decision-maker aggregates" — designed 2026-07-06/07, open Q-E/G/H.
- **what:** Before any fix is finalised, independent specialist agents each judge it from their own perspective and send structured opinions (verdict-wire-style: verdict + citations + objections + suggested alternative) to a decision-maker. Initial roster: guidelines agent (adherence to 000-0xx — or did the guideline fall short), reuse agent (code AND docs), bug-historian, compliance/legal, pipeline guardians (one per master workflow, seeded from the builder relay map), and specialist knowledge agents ("we already have one of these"). Precursor idea from the thin slice: build-time liability and MORALITY review contributors applying a configured, layered standard with contested calls routed to a human.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#the-task (item 3); docs019/RUNBOOK_thin_slice(27).md#next-improvements (item 3)
- **relations:** hard-veto semantics; three-tier citation (opinion contract); docs026 council-agents stage
- **verify-later:** reviewer agent definitions (none yet); Q-G reviewer-context decision

<!-- SOURCE: U14_docs019_runbooks.md -->
### Hard-veto flag semantics for reviewers
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "Q-D council topology — VETO SEMANTICS DECIDED (owner, 2026-07-07): parallel reviewers → decision-maker BY DEFAULT … a hard_veto flag, attachable at multiple scopes … converts that reviewer's negative verdict into a BLOCK".
- **what:** All council opinions are advisory by default and weighed together; a hard_veto flag — attachable per reviewer agent, per pipeline, or per tool/component, most-specific-scope contemplated — makes that reviewer's negative verdict blocking. Accessibility and legal are the motivating hard-veto cases. A guidelines-reviewer "the guideline itself fell short" finding leans side-task (gap, not violation), not block.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-D)
- **relations:** council of reviewers; learning layer (guideline-gap side-task)
- **verify-later:** where the flag lives (reviewer column vs council config)

<!-- SOURCE: U14_docs019_runbooks.md -->
### Architecture-change visibility detector
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "Q-E architecture-change signals: packages touched breadth; platform/ vs actions/; exported-signature diffs vs the corpus; message/topic/schema/contract changes; migration presence. Which are load-bearing?" (open, F2-phase).
- **what:** Make it loud when a proposed change is accidentally fundamental — touching platform contracts, message shapes, many packages, exported signatures — before it ships; runs as one council reviewer. Candidate signals enumerated; exported-signature diffs against the code_symbols corpus is the notable reuse of the diagnosis infrastructure.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#the-task (item 4); docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-E)
- **relations:** council of reviewers; code_symbols corpus
- **verify-later:** n/a (not built)

<!-- SOURCE: U14_docs019_runbooks.md -->
### Learning layer — bug records and guideline-amendment side-tasks
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) F3 "bug_records (category taxonomy, recurrence checks feeding the historian); guideline-amendment proposals routed to the human"; Q-D "guideline-gap SIDE-TASK (amendment PR against the guideline docs, human terminal, fix unblocked, F3 recurrence record)".
- **what:** The feedback layer: recorded bugs with a category taxonomy and recurrence checks (feeding the bug-historian reviewer so a class never repeats); when a fix exposes a guideline gap, a side-task raises an amendment PR against the guideline docs with the human as terminal approver while the fix itself proceeds; corpus and doc enrichment feed back into retrieval.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#phased-plan (F3); docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-D)
- **relations:** corpus enrichment policy; council (bug-historian); coverage baseline (guideline home)
- **verify-later:** bug_records table (absent); amendment-PR mechanism

<!-- SOURCE: U14_docs019_runbooks.md -->
### Loop-worthiness test (five-criteria intake doctrine)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** diagnosis_fix_loop(9) "LOOP-WORTHINESS TEST (doctrine — apply before every intake)" — applied three times in the same file (pilot #1 downgraded, candidate 2 forked, guides pilot confirmed).
- **what:** A task is loop material only when ALL hold: (1) a SYMPTOM about system behaviour, not a feature request; (2) a causal mechanism plausibly exists in code+data+runtime; (3) not answerable by one or two direct queries (mandatory cheap pre-check first); (4) bounded to one symptom; (5) verified CURRENT at intake — symptoms are perishable. Feature absences → build routes; quality judgements → council/auditors; one-query questions → the query. Demonstrated by downgrading the roadmap-gap "bug" (findable by reading two files) to a builder-queue item.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#loop-worthiness; docs019/RUNBOOK_diagnosis_fix_loop(9).md#previous-pilot-1
- **relations:** F0 guides pilot; falsification eval gate
- **verify-later:** n/a (doctrine)

<!-- SOURCE: U14_docs019_runbooks.md -->
### F0 pilot — the guides-route differential diagnosis
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "★ F0 PILOT — CONFIRMED 2026-07-07: nav links to a guides section that has no content" with pre-registered criteria; ordered after F0.1 plumbing.
- **what:** The chosen first fix-loop pilot: dartsonline published a Guides nav link and blank /guides/index.html while gamesdesign (same platform) has working guides — a two-site DIFFERENTIAL, the strongest evidence shape. Standing hypothesis for the loop to confirm/refute FROM CODE: reconcile_site_plan's routing table has no "guide" entry (blog-index present, tool commented out), so planner-emitted guide pages were silently dropped while nav — generated from the PLAN, not the built set — published the link. Two earlier pilot candidates were downgraded via the loop-worthiness test.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#f0-pilot; docs019/RUNBOOK_builder_route(21).md#queue (item 7)
- **relations:** loop-worthiness test; reconcile routing table; nav-grounded-in-built-set principle
- **verify-later:** load_work_item_actions.go routing table; the pilot's run artifacts once executed

<!-- SOURCE: U14_docs019_runbooks.md -->
### Per-task running notes via doc_notes (travelling docs reuse)
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** diagnosis_fix_loop(9) "Q-F DIRECTION SET (2026-07-07): REUSE doc_notes. The terminal-diagnosis note already exists on their side (pending their 3b subject threading)"; "the diagnose-agent workflow is ALREADY rewired by them: emit → persist_note → complete".
- **what:** Live monitoring of what the loop is doing and why: per-iteration and per-step reasoning written to a task-specific notes home. Decision: reuse the tools chat's doc_plans/doc_notes infrastructure (terminal diagnosis note already wired via persist_note with a strict no-guessing subject gate); per-iteration rows are additional doc_notes entries pending the owning thread's sign-off; category convention `diagnosis`.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-F); docs019/RUNBOOK_diagnosis_fix_loop(9).md#what-already-exists
- **relations:** doc_plans/doc_notes infrastructure; reasoning-state handoff; thread-boundary convention
- **verify-later:** doc_notes rows with category diagnosis; persist_diagnosis_note action

<!-- SOURCE: U15_docs019_running_notes.md -->
### Diagnosis→fix loop workstream (founding)
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** NOTES_running_fixloop(9).md THREAD STATE AT CUTOVER: "DECIDED AND RECORDED... STILL OPEN (F2-phase)... FIRST ACTION: slice F0.1 with pre-registered criteria."
- **what:** The founding thread (2026-07-06/07) that pivots the read-only diagnosis loop into a diagnosis→fix system: documented intake (`needs_diagnosis` work item, `pipeline='diagnose'` namespace), live per-iteration reasoning persisted to a new `diagnosis_artifacts` table (kind ∈ bundle|iteration_note, written through inside the assemble action — off the parallel tools-chat's `doc_notes` surface, only the terminal note relayed there), fixes produced by a SEPARATE fixer agent with an isolated git write token (spawn-gate pattern) producing a constrained edit plan validated by gofmt+build before any PR, and a council of parallel specialist reviewers feeding a decision-maker (see hard_veto flag concept below). This is the same workstream documented in far greater operational detail in `docs024_key_docs_latest/fixloop_eg_dartsonline/` — this file is its origin notes.
- **sources:** NOTES_running_fixloop(9).md (full); NOTES_running_synthesis_v4(39).md 2026-07-06/07 entries (same founding, condensed).
- **relations:** Loop-worthiness test doctrine; hard_veto council flag; diagnosis loop; roadmap-phase enforcement gap.
- **verify-later:** `diagnosis_artifacts` table; the fixer agent's isolated write-token/spawn-gate; cross-reference against `docs024_key_docs_latest/fixloop_eg_dartsonline/` for the fuller, later-stage version of this same concept.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Council hard_veto flag / decision-maker model
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** "Q-D veto semantics decided (owner)... Flag-based: DEFAULT = decision-maker weighs all opinions; a hard_veto flag at reviewer/pipeline/tool/component scope makes that reviewer's negative verdict a BLOCK" (NOTES_running_fixloop(9).md).
- **what:** The fix-loop's review-arbitration model: a parallel council of specialist reviewers (guidelines/reuse/bug-historian/compliance/per-pipeline guardians) feeds a decision-maker by default (advisory), except where a `hard_veto` flag is set at reviewer/pipeline/tool/component scope (accessibility and legal are the motivating cases), which makes that reviewer's negative verdict an unconditional block. A guideline-gap found during review is a SIDE-TASK (a work item that drafts an amendment PR against the guideline docs, human-terminal) rather than something that blocks the fix.
- **sources:** NOTES_running_fixloop(9).md "Q-D veto semantics decided" and "F0/F1 design settled" (DECISIONS).
- **relations:** Diagnosis→fix loop workstream founding.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Loop-worthiness test doctrine
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "Owner asked whether the loop fits the dartsonline quality problem. Answer: decomposed via the new LOOP-WORTHINESS TEST (symptom-not-feature; mechanism-plausible; not one-query-answerable; single-symptom)" (NOTES_running_fixloop(9).md); a fifth criterion (verify symptom currency at intake) added after a pilot candidate "evaporated" (was fixed live before the loop ran).
- **what:** A pre-registered five-criterion test for whether a candidate bug is worth running the diagnosis/fix loop on: it must be a genuine symptom (not a disguised feature request), the mechanism must be plausible from code, it must not be answerable by one query, it must be a single coherent symptom, and its currency must be reverified at intake (since bugs can be fixed out from under a pilot mid-triage — this happened twice in this thread alone).
- **sources:** NOTES_running_fixloop(9).md multiple 2026-07-07 pilot-selection entries.
- **relations:** Diagnosis→fix loop workstream founding.

<!-- SOURCE: U16_docs019_design_plans.md -->
### Diagnosis→fix loop programme (F0–F3)
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** HANDOFF_fixloop(8) (2026-07-07): "All F0/F1 design questions are DECIDED"; README_02: "a pipeline that turns 'something is wrong' into 'here is a reviewable, evidence-backed proposal' … exercised on exactly one bug. The write step (plan → PR) is half-built."
- **what:** The evolution of the diagnosis loop into a fix pipeline: symptom → cited diagnosis → constrained edit plan → adversarial council review → revision informed by reviewer-requested DB queries → approved plan or honest escalation. Phased F0 (persistence/intake) → F1 (write step) → F2 (council expansion) → F3, driven by open questions Q-A…Q-H resolved in the discussion thread. The valuable output is the general pattern, not the bug-fixing.
- **sources:** HANDOFF_fixloop_thread(8).md; README_02_evidence_backed_proposals.md; README_overview.md (F1.1b(c) status)
- **relations:** council pattern; fix-implementer; pilot worthiness test; docs026 concept-council mission
- **verify-later:** RUNBOOK_diagnosis_fix_loop.md + NOTES_running_fixloop.md (units U14/U15); fix-implementer seed

<!-- SOURCE: U16_docs019_design_plans.md -->
### Council pattern: adversarial multi-agent review with deterministic aggregation
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** README_02: "the roster of 2 (edit-quality, guardian) is explicitly a skeleton"; "the guardian vetoed an architecture change dressed as a fix"; veto semantics decided per HANDOFF_fixloop(8) delta history ("flag-based hard veto, default advisory → decision-maker").
- **what:** Multiple reviewer agents each examine the proposed fix plan from one lens; a deterministic rule (not a third model) aggregates their positions; specified veto semantics are flag-based hard veto with advisory as default. Reviewers can demand facts, and the loop runs the queries itself rather than letting the proposer argue (self-verification instead of self-belief). Three runs running, the council correctly ruled the test bug's proper fix beyond a constrained plan's mandate.
- **sources:** README_02_evidence_backed_proposals.md; HANDOFF_fixloop_thread(4)-(8).md deltas; README_comprehensive_documentation_categorisation.md (veto description)
- **relations:** expanded council bench; escalation as success; guardian-from-decision-record (Q-G)
- **verify-later:** council agent seeds + the aggregation rule in Go; fixloop_eg_dartsonline docs (docs024)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Expanded council bench (expert-per-area reviewers)
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** README_02: "the runbook's F2 design already names the full bench … Adding a reviewer is a seed change + prompt + curated context."
- **what:** The planned F2 roster beyond the two-agent skeleton: a guidelines agent (conformance to 000–0xx, or did the guidelines fall short), reuse agent (are we rebuilding something that exists, code and docs), bug-historian (has this class recurred), compliance eye, pipeline guardians one per master workflow, and specialist knowledge agents. Reviewer areas are expected to correlate with the docs024 documentation categories — the direct bridge to the docs026 concept register's council-agent goal.
- **sources:** README_02_evidence_backed_proposals.md#3; README_comprehensive_documentation_categorisation.md
- **relations:** council pattern; concept register mission; documentation categories as expertise areas
- **verify-later:** whether any bench agents beyond edit-quality/guardian were seeded

<!-- SOURCE: U16_docs019_design_plans.md -->
### Fix-implementer constrained write step
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** README_overview: "F1.1b(c) is code complete (b367a602) … validated as far as it can be without the deploys"; deploy checklist listed, first end-to-end run pending.
- **what:** A 15-step seeded agent: load plan/council/diagnosis → APPROVED gate (mirror of the CONFIRMED gate) → diagnose_read_repo_files fetches current file bodies via the GitHub contents API with a hard rule that a modify-file 404 is a refusal (whole-file rewrites of unseen files would be hallucination by construction) → sketch_to_files whole-file rewrites ("the diff a human reviews must contain ONLY the plan") → deterministic file allowlist → create fix/* branch → commit via git_adapter_request (one generic adapter caller; verbs allowlisted to commit/create_branch/create_pull_request so delete_repo is structurally unreachable) → build gate (golang Job): green → PR into main, red → NO PR, branch + build log left. Runs on the read-token spawn gate.
- **sources:** README_overview.md (landed pieces + deploy checklist); README_02_evidence_backed_proposals.md
- **relations:** hard deterministic gates; human-gate-never-moves; seeded-bug first run; build gate options A/B/C
- **verify-later:** 0NN_fix_implementer.sql; 092_TRIGGER_fix_implementer_v1.sh; git-adapter branch/PR ops; RBAC for pods/log

<!-- SOURCE: U16_docs019_design_plans.md -->
### Hard deterministic gates between every LLM step
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** README_02 lists them as built pattern: "CONFIRMED gate, plan validator, file allowlist. The models propose; plain Go code decides what proceeds."
- **what:** No LLM output passes into consequence unchecked: the diagnosis must be CONFIRMED (gate), the plan must validate, the files must be on a deterministic allowlist, the build must pass, before anything advances. Complexity and authority live in plain Go; the models only propose. The same shape as keeping convergence guards in the engine rather than in workflow conditionals.
- **sources:** README_02_evidence_backed_proposals.md#1; README_overview.md
- **relations:** council aggregation rule; fix-implementer; thin-workflows rule
- **verify-later:** the gate implementations in the fixloop actions

<!-- SOURCE: U16_docs019_design_plans.md -->
### The human gate never moves (nothing merges itself)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** README_02: "one structural commitment: the human gate never moves. More autonomy upstream … never past the PR. Nothing merges itself."
- **what:** Autonomy may widen upstream (diagnose, plan, revise, commit-to-branch) but the merge is permanently human. The PR is the fixed boundary of machine authority in the fix loop — a simpler, harder commitment than the graduated trust machinery, and orthogonal to it.
- **sources:** README_02_evidence_backed_proposals.md#2; README_overview.md (red build → NO PR)
- **relations:** trust ledger (graduated autonomy elsewhere); awareness surface; fork isolation
- **verify-later:** absence of any auto-merge path in the write step

<!-- SOURCE: U16_docs019_design_plans.md -->
### Escalation as a first-class success
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** README_02 pattern list: "'this is beyond my mandate' is a correct output, packaged for you"; the council produced exactly this on the test bug three times.
- **what:** When a fix exceeds the constrained plan's mandate (architecture-level causes), the loop's correct output is an honest escalation package for the human, not a forced plan. Treating refusal-to-proceed as success is the organisational analogue of UNVERIFIABLE-beats-guessing.
- **sources:** README_02_evidence_backed_proposals.md; README_02 §6 (the escalate decision explained)
- **relations:** cite-or-abstain; council pattern
- **verify-later:** escalation package format in the fixloop runbook

<!-- SOURCE: U16_docs019_design_plans.md -->
### Fix-loop value proposition: unattended, cited, consistent
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** README_02: "The value proposition (decided 2026-07-09): not 'the loop finds what humans can't' … The proposition is unattended, cited, consistent — the 3am diagnosis with a paper trail."
- **what:** A recorded decision reframing what the loop is for: on this platform bugs are legible to anyone with schema access and patience, so the differentiation is not superhuman insight but unattended operation with citations and consistency — a package instead of a hunch, reconstructible after the fact by one correlation id. Every design choice flows from it.
- **sources:** README_02_evidence_backed_proposals.md#2
- **relations:** falsification-first; awareness surface; diagnosis artifacts persistence
- **verify-later:** decision record in NOTES_running_fixloop (U15)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Awareness surface before wider autonomy
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** README_02: "the missing organ is a push surface: a periodic digest … before autonomy widens … the awareness surface gets built first" — a recommendation, not built.
- **what:** The named risk is not wrong action but unknown action — drift compounding silently while trails exist only pull-side. Proposed standing gate: before councils multiply or migration agents exist, build a push digest (what ran, what was decided and by which rule, what was escalated, what the council almost approved). "It must explain what it's doing, or it doesn't get to do more." The grown-up form of the parked F0.3 per-iteration notes.
- **sources:** README_02_evidence_backed_proposals.md#4
- **relations:** diagnosis artifacts persistence; decision log (the governance twin); human-gate-never-moves
- **verify-later:** whether any digest mechanism exists

<!-- SOURCE: U16_docs019_design_plans.md -->
### Fork isolation of the write surface
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** README_02 §5: "the strongest isolation is that the write surface points only at the fork … a designed slice, not huge" — proposed, not executed in these files.
- **what:** Point the loop's git-adapter credential, intake defaults and corpus indexing at a fork of the repo, making the main repo physically unwritable by the loop rather than protected by review discipline; the human pulls reviewed changes across. Folds in "mission and objectives correct in the first place": the fork's constitution/mission docs become the councils' curated context so conformance is checked against human-authored documents.
- **sources:** README_02_evidence_backed_proposals.md#5
- **relations:** human gate; guardian-from-decision-record; external rollback
- **verify-later:** git-adapter repo config; whether a fork exists

<!-- SOURCE: U16_docs019_design_plans.md -->
### Pilot worthiness test and the dartsonline guides pilot
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** HANDOFF_fixloop(8): "★ THE PILOT IS CONFIRMED (2026-07-07) … Two earlier candidates were rejected … that triage history is itself the worthiness test working."
- **what:** A five-criteria test for whether a bug deserves the diagnosis loop, exercised through three candidates: the chrome/nav defect (dropped — got fixed; perishability lesson), the nav-links-to-never-rendered-pages defect (downgraded — root cause found by direct code reading, a known platform gap, reclassified to the builder route), and the confirmed pilot: dartsonline published a Guides nav link and a blank /guides/index.html while gamesdesign has working guides — a broken route, not a missing feature, with a standing hypothesis (reconcile_site_plan's routing table omits "guide"; nav derives from the plan, not the built set), mandatory pre-check queries and a cross-site differential as evidence. Establishes "genuinely mechanism-unclear" as the admission bar.
- **sources:** HANDOFF_fixloop_thread(8).md; HANDOFF_fixloop_thread(3)-(5).md deltas (the triage history)
- **relations:** eval gate; site-plan reconciler routing table (the suspected mechanism)
- **verify-later:** reconcile_site_plan routing table in load_work_item_actions.go; the F0 PILOT section of the fixloop runbook (U14)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Seeded-bug strategy for the first end-to-end write run
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** README_02 §6 and README_overview both recommend it; "the system's first-ever PR will have earned every gate it passed" — proposed, awaiting deploys.
- **what:** Because the only real test bug never yields an approvable plan (correctly escalated as architecture-level), the write step is tested by planting a contained single-file defect with an obvious symptom on a low-stakes surface and running the full pipeline — diagnose → plan → council (genuine approval) → implementer → PR. Rejected alternatives: hand-approving a known-flawed plan (contradicts the reviewers), waiting for an organic small bug (unbounded).
- **sources:** README_02_evidence_backed_proposals.md#6; README_overview.md
- **relations:** fix-implementer; eval gate
- **verify-later:** whether the first PR happened (git history for fix/* branches)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Transferable machinery: legacy-migration and feature intakes
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** README_02 §3: migration agents "not built, but it's the same machinery with a different intake"; features from specs "honestly furthest away … plausible; not designed".
- **what:** The allowlist/gate/council scaffolding is intake-agnostic: a legacy migration is "pattern X supersedes pattern Y" (scanner finds Y-shaped code, proposer writes constrained plans, council reviews, PRs flow); feature-building from mission docs needs a new grounding tier ("cite the spec clause this serves") — same shape as causal citation but not designed.
- **sources:** README_02_evidence_backed_proposals.md#3
- **relations:** council pattern; hard gates
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U08_travelling_docs.md -->
### fix-proposer agent (F1.1a) — read-only proposal writer
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** FYI addendum 2026-07-10: "a fix-proposer agent (F1.1a) now exists".
- **what:** An agent in the diagnosis→fix loop that reads only orchestration_states/diagnosis_artifacts and writes only kind='fix_plan' artifacts — no code writes, no git token. Noted here as a boundary fact for the travelling-docs surface owners.
- **sources:** FYI_from_fixloop_2026-07-10_verdict_prompt_symptom_check.md#addendum
- **relations:** diagnosis loop; fix-loop workstream (primary docs elsewhere).
- **verify-later:** fix-proposer agent_definitions row; diagnosis_artifacts table.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Diagnosis→fix council loop (overall system)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "PR #1 — github.com/gqls/agentchassis/pull/1 — APPROVED & MERGED" (NOTES(10) Turn 26-28); SUMMARY_where_we_are_2026-07-13.md: "Today it did the whole thing, end to end, for the first time."
- **what:** The end-to-end pipeline that turns a plain-English bug symptom into a human-reviewed pull request: intake → read-only diagnosis (cite-or-abstain, three evidence tiers) → constrained fix plan → two-reviewer council with deterministic decision → dedicated-pod implementer behind a hard file allowlist → containerized build gate → PR. Every stage writes to `diagnosis_artifacts`/`orchestration_states` keyed on one correlation_id so the whole run is auditable. Human review is the terminal; nothing merges itself.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#THE TASK, fixloop_eg_dartsonline/MILESTONE_diagnosis_fix_loop_2026-07-10.md, fixloop_eg_dartsonline/SUMMARY_where_we_are_2026-07-13.md, fixloop_eg_dartsonline/README_so_far.md
- **relations:** all other fixloop concepts in this extraction are sub-scopes of this one
- **verify-later:** PR #1 on github.com/gqls/agentchassis; `diagnosis_artifacts` table contents on clients_db

<!-- SOURCE: U13_docs024_small_dirs.md -->
### needs_diagnosis intake route (F0.1c)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "F0.1c — ✅ LANDED 2026-07-09" (PLAN_fixloop_pilot.md §1)
- **what:** The one documented way a bug enters the loop: `090_TRIGGER_needs_diagnosis_v1.sh` writes a durable `site_work_items` row (`pipeline='diagnose'`, `item_type='needs_diagnosis'`, `status='awaiting_diagnosis'`) and fires the `diagnose-orchestrator` Kafka envelope on the same correlation_id, so the intake record, the diagnosis_artifacts bundles, and the terminal doc_notes row all join on one key. `DISPATCH=0` records without firing; the older `084_TRIGGER_diagnose_v1.sh` remains for ad-hoc runs with no intake record. `item_key` (`needs_diagnosis:<slug>`) plus `idx_swi_dedup` makes re-running the same slug idempotent while an intake is open.
- **sources:** fixloop_eg_dartsonline/090_TRIGGER_needs_diagnosis_v1.sh, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md#F0.1c, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Q-B CORRECTION
- **relations:** system.internal pseudo-site anchor pattern; private inert pipeline statuses pattern; diagnose-dispatch-loop
- **verify-later:** site_work_items rows with pipeline='diagnose'; idx_swi_dedup index definition

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Superseded: null-site-allowed intake design
- **category:** fix-loop
- **status-signal:** superseded
- **status-evidence:** "Q-B CORRECTION (2026-07-09... 'Null-site allowed' is impossible" (RUNBOOK(10)#Q-B CORRECTION); original design in RUNBOOK(9)#QUESTIONS "null-site allowed"
- **what:** The 2026-07-07 owner decision originally specified that site-less code bugs would "ride null-site" in the new diagnose pipeline namespace. Reading the live schema on 2026-07-09 showed this was structurally impossible twice over (NOT NULL column; site-anchored loader query), and it was replaced by the system.internal pseudo-site anchor pattern.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#QUESTIONS, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Q-B CORRECTION, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 4
- **relations:** system.internal pseudo-site anchor pattern
- **verify-later:** n/a — superseded, no live code to check

<!-- SOURCE: U13_docs024_small_dirs.md -->
### diagnosis_artifacts table (unified egress store)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "F0.1a — ✅ LANDED 2026-07-09" (PLAN_fixloop_pilot.md); kind list extended live via ALTER in 0NN_fix_proposer.sql v4/v5
- **what:** The durable, correlation_id-keyed egress table for the whole loop, with `kind` growing over time: `bundle` and `iteration_note` (F0.1a) → `fix_plan` (F1.1a) → `council_report` (F2.1) → `escalation` (F2.3). `correlation_id` is deliberately `text` not `uuid` (ExecutionContext.CorrelationID has no guaranteed form). A partial unique index on `(correlation_id, iteration) WHERE kind='bundle'` gives retry-safe upsert for bundles while allowing multiple `iteration_note` rows per iteration. Carries a retention knob (`expires_at`/`pinned`) that is defined but has no sweep implemented yet.
- **sources:** fixloop_eg_dartsonline/0NN_diagnosis_artifacts.sql, fixloop_eg_dartsonline/0NN_fix_proposer.sql#§1, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §F0.1a
- **relations:** bundle write-through; fix-proposer's fix_plan artifacts; council_report; escalation artifact; retention sweep (unbuilt)
- **verify-later:** table DDL and CHECK constraints on clients_db; whether a retention sweep job exists

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Retention/expiry knob on diagnosis_artifacts
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** "Retention knob. NULL = keep indefinitely. Sweep deletes WHERE expires_at < now() AND NOT pinned" (0NN_diagnosis_artifacts.sql comment) — no sweep job found in scope
- **what:** `expires_at`/`pinned` columns exist on diagnosis_artifacts (bundles configured to expire sooner than notes; NULL = keep forever), and a partial index exists to support a future deletion sweep, but no sweep job/scheduled task was found anywhere in this extraction's file set. The mechanism is designed but not built.
- **sources:** fixloop_eg_dartsonline/0NN_diagnosis_artifacts.sql#idx_diagnosis_artifacts_expiry
- **relations:** diagnosis_artifacts table
- **verify-later:** search codebase/scheduled_tasks for any retention-sweep job referencing diagnosis_artifacts

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Known-answer benchmark methodology
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "Benchmark method validated end-to-end: same symptom, one variable cluster, measurable delta" (NOTES(10)#Turn 10)
- **what:** When a candidate bug's mechanism is dissolved by the mandatory cheap pre-check (three candidates in a row were), the pilot is not discarded but promoted from a "discovery run" to a "known-answer benchmark": the loop is run blind on the original symptom string and its output is scored against a pre-registered rubric of must/should/bonus claims fixed before the run, including a "refutation credit" penalizing confirmation of a known-false standing hypothesis. This produced a repeatable, gradable evaluation across five runs, each of which found and fixed a real engine defect.
- **sources:** fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §0, §3, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#AMENDMENT 2026-07-09, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 1
- **relations:** loop-worthiness test doctrine; blinding discipline; dartsonline guides defect (benchmark bug)
- **verify-later:** rubric table in PLAN_fixloop_pilot.md §3

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Abandoned pilot candidates (chrome bug, roadmap-phase gap, blank-guides fork)
- **category:** fix-loop
- **status-signal:** abandoned
- **status-evidence:** "Chrome pilot EVAPORATED (fixed live)"; "PREVIOUS: F0 PILOT #1 — DOWNGRADED... root cause found; not diagnosis-shaped" (RUNBOOK(9)#F0 PILOT ORIGINAL RECORD, #SUPERSEDED CANDIDATE 2)
- **what:** Three earlier pilot candidates were considered and dropped before the dartsonline guides bug was chosen: (1) dartsonline pages lacking site chrome (fixed before the loop ran); (2) a "no submission path produces a roadmap" gap, reclassified as a known platform gap rather than a diagnosis target since a human found it by reading two files; (3) a blank guides-index fork where the guide-writing mechanism's existence was unverified. Recorded because dropped/dissolved candidates are exactly the kind of idea worth remembering for future taxonomy work.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#F0 PILOT ORIGINAL RECORD, #SUPERSEDED CANDIDATE 2, #PREVIOUS F0 PILOT #1
- **relations:** loop-worthiness test doctrine; roadmap/phases mechanism
- **verify-later:** n/a — historical, dissolved candidates

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Blinding discipline for benchmark runs
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "BLINDING IS MANDATORY... Exclude .../fixloop_eg_dartsonline/ from the loop's corpus" (RUNBOOK(10)#★ F0 PILOT)
- **what:** Established that the diagnose-agent workflow structurally cannot read this docs directory (it walks Go source and DB rows only), so blinding is largely automatic; the only two leak vectors are the symptom string (must describe only observable behaviour) and `seed_scope` (must be omitted entirely for a benchmark run).
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#REGENERATING THE CONTEXT BUNDLE §BLINDING, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §3
- **relations:** known-answer benchmark methodology
- **verify-later:** diagnose-agent workflow JSON — confirm no doc-reading step was later added

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Standing hypothesis refuted (reconcile_site_plan routing table)
- **category:** fix-loop
- **status-signal:** superseded
- **status-evidence:** "THE STANDING HYPOTHESIS IS REFUTED — and it named the wrong file" (NOTES(10)#Turn 1)
- **what:** The original hypothesis (from RUNBOOK(9)) blamed `reconcile_site_plan`'s routing table for silently dropping "guide" pages. Hand-diagnosis on 2026-07-09 showed the routing table is real but lives in `WriteBuildItemsAction`, and absence from it does not drop a page (it defaults to page-build-handler); the actual drop mechanism is a separate `unavailableBuilders` map, and `reconcile_site_plan_action.go` has no type switch at all. Retained specifically because a hypothesis a loop should refuse to confirm is exactly the kind of "superseded" idea the register wants captured.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#★ F0 PILOT, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 1
- **relations:** dartsonline guides defect; two intake paths disagreement
- **verify-later:** grep/inspect `reconcile_site_plan`; `WriteBuildItemsAction`; `unavailableBuilders`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Two intake paths disagreement (WriteBuildItemsAction vs reconcile_site_plan)
- **category:** fix-loop
- **status-signal:** deployed (documented finding, not fixed — routed to builder thread)
- **status-evidence:** "Fourth finding (unlooked-for): the two intake paths disagree" (NOTES(10)#Turn 1)
- **what:** `WriteBuildItemsAction` deliberately skips `tool`/`entity-directory`/`entity-page` page types via an `unavailableBuilders` guard, while `reconcile_site_plan_action.go` hardcodes `handler_agent='page-build-handler'` for every plan page with no type switch at all, so it re-emits build items for the very types the other path skips. Flagged as a builder-thread decision, not fixed here.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 1, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#★ F0 PILOT
- **relations:** dartsonline guides defect; pipeline-blind dispatch surfaces
- **verify-later:** grep/inspect `WriteBuildItemsAction`; `tool`; `entity-directory`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### mark_no_sections — referenced-but-never-built step
- **category:** fix-loop
- **status-signal:** abandoned
- **status-evidence:** "`mark_no_sections` does not exist... appears nowhere in the repo but that one comment" (NOTES(10)#Turn 1)
- **what:** A code comment at `load_work_item_actions.go:750-756` names a remedy step `mark_no_sections` that would flag a sectionless page's work item `needs_human_review` — but the step was never implemented; it appears nowhere else in the repo. The completion guard that would consume its flag faithfully preserves a flag nothing ever sets.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 1, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#★ F0 PILOT
- **relations:** dartsonline guides defect
- **verify-later:** grep for mark_no_sections in current repo

<!-- SOURCE: U13_docs024_small_dirs.md -->
### fix-proposer agent / constrained edit plan (F1.1a)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "F1.1a — ✅ BUILT 2026-07-10" (PLAN_fixloop_pilot.md); live agent seeded in 0NN_fix_proposer.sql
- **what:** The first F1 slice: an `agent_definitions` workflow that loads a diagnosis by `fix_correlation_id`, refuses anything not `CONFIRMED`, and drafts a constrained edit plan (summary, ≤8 allowlisted edits with file/symbol/operation/rationale/sketch, `grounded_in` quotes required, risks) persisted to `diagnosis_artifacts` (`kind='fix_plan'`). It writes no code and needs no git token.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#propose step, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §3b F1.1a
- **relations:** diagnose_persist_fix_plan validator; two-reviewer council; CONFIRMED gate
- **verify-later:** fix-proposer agent_definitions row, workflow version

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Plan validation / hard allowlist for edit plans (diagnose_persist_fix_plan)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "structural validation... UNLIKE the bundle write-through this FAILS the step on bad input" (NOTES(10)#Turn 15)
- **what:** A deterministic Go action validating a proposer's plan before persisting it: non-empty summary/edits/rationale/sketch, operations restricted to `modify|add|remove|config_change`, repo-relative paths only, ≤8 edits, `grounded_in` quotes required, 32KB cap — plus (F1.1b(a)) rejection of explicit no-op phrases so edits that change nothing cannot pass as real edits. Fails closed; fired correctly on the first two real fix-proposer runs (max_tokens truncation).
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 15, #Turn 16, #Turn 18
- **relations:** fix-proposer agent; hard file allowlist (implementer's analogous gate)
- **verify-later:** grep/inspect `modify|add|remove|config_change`; `grounded_in`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Two-reviewer council (F2.1)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "F2.1 — ✅ PROVEN LIVE 2026-07-10" (PLAN_fixloop_pilot.md)
- **what:** Two sequential LLM reviewer steps plus a deterministic Go decision — not a third model opinion about two model opinions. `review_editquality` judges whether every edit changes something real and targets the actual causal path; `review_guardian` judges blast radius, architecture-change signals, and surface ownership, and alone holds the hard veto. Both attach optional `checks:[{sql,why}]`. First live run judged real objections correctly rather than rubber-stamping.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#review_editquality, #review_guardian, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 18-19
- **relations:** deterministic council decision + hard veto; verify step; schema hint for reviewers
- **verify-later:** grep/inspect `review_editquality`; `review_guardian`; `checks:[{sql,why}]`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Deterministic council decision + hard veto (diagnose_council_decide)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "diagnose_council_decide: ordered rules (hard veto → rejected; any veto → rejected; any objection → revise; all approve → approved)" (NOTES(10)#Turn 18)
- **what:** A pure Go action aggregates the two reviewers' verdicts deterministically, auditable and reproducible. `hard_veto_from` is currently a flag in the workflow step config naming the guardian reviewer as sole veto-holder; whether it should instead live on a reviewer/pipeline definition column remains an open sub-question. Malformed reviewer output fails closed.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#council_decide, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#STILL OPEN Q-D
- **relations:** two-reviewer council; hard_veto flag at multiple scopes (superseded early design); decision router (F2.3)
- **verify-later:** grep/inspect `hard_veto_from`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### hard_veto flag at multiple scopes (early design, superseded)
- **category:** fix-loop
- **status-signal:** superseded
- **status-evidence:** owner decision 2026-07-07: "hard_veto flag, attachable at multiple scopes (a reviewer agent, a pipeline, a specific tool/component)" (RUNBOOK(9)#Q-D) — actual shipped implementation is a single `hard_veto_from` list in one workflow step's config
- **what:** The original council design envisioned a hard_veto flag placeable at reviewer, pipeline, tool, or component scope, motivated by accessibility/legal review cases. What was actually built (F2.1) is narrower: a single `hard_veto_from: ["guardian"]` array in the fix-proposer's `council_decide` step config. The broader multi-scope placement remains an open sub-question of Q-D.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#Q-D, fixloop_eg_dartsonline/0NN_fix_proposer.sql#council_decide config
- **relations:** deterministic council decision + hard veto; future council roster
- **verify-later:** grep/inspect `hard_veto_from: ["guardian"]`; `council_decide`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Revise loop (F2.2)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "F2.2 REVISE LOOP — ✅ BUILT 2026-07-10" (PLAN_fixloop_pilot.md)
- **what:** On a `revise` decision with rounds remaining, feeds the diagnosis, prior plan, and both reviewers' objections back into a `repropose` step, then re-validates and re-reviews (capped, default 2, later 3). Exhausting the cap becomes a distinct terminal state `exhausted`. Round counting is scoped per `orchestration_id` (proposer run), not per `correlation_id` — an earlier per-correlation design flaw was caught mid-implementation and fixed, then found still live on the deployed binary for one further round (round-counting scope bug).
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#council_decide/check_revise, fixloop_eg_dartsonline/HANDOFF_turn21_2026-07-10.md#Session 20 Summary
- **relations:** decision router (F2.3); round-counting scope bug
- **verify-later:** grep/inspect `revise`; `repropose`; `exhausted`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Decision router (F2.3)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "F2.3 DECISION ROUTER + VERIFY + REFRAME + ESCALATION — ✅ CODE BUILT 2026-07-10... PROVEN live" v1.0.1108 (PLAN_fixloop_pilot.md)
- **what:** Replaces the single revise/complete branch with a full router: `approved`→complete; `revise` with rounds left→verify checks→repropose; `rejected` first time with rounds left→reframe-once; `rejected` again or exhausted→escalate. Motivated by two clean benchmark runs exposing two dead-ends. Flags are computed by a pure, directly-tested Go function `applyCouncilCaps`.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#check_approved/check_rejected/check_reframe/check_revise, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 22
- **relations:** revise loop; verify step; reframe step; escalation artifact
- **verify-later:** grep/inspect `approved`; `revise`; `rejected`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Verify step (diagnose_run_checks)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "Verify step ran 7 reviewer checks under the containment" (NOTES(10)#Turn 23)
- **what:** Reviewers attach `checks:[{sql,why}]` (SELECT/WITH only) to their verdicts; this action executes them under the same read-only containment the diagnosis loop's data_requests use (lint → READ ONLY transaction → statement_timeout → EXPLAIN gate → capped rows), and feeds results into the next repropose so fact-shaped objections are settled with evidence instead of another blind revision round. Capped at 8 checks by default.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#run_checks, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 22, #Turn 23
- **relations:** decision router; schema hint for reviewers
- **verify-later:** grep/inspect `checks:[{sql,why}]`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Schema hint for reviewers (F2.3b(a))
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "Verification run 1e221fb7: 8 checks run, 0 failures (prior run: 5 of 7 failed on hallucinated schema)" (NOTES(10)#Turn 24)
- **what:** A `load_schema_hint` query_database step pulls the live table/column list from `information_schema` at run time, and both reviewer prompts get this hint plus two named traps (workflow steps live in `agent_definitions.default_config` jsonb, not a steps table; a site's domain lives on `sites` joined via `pages.site_id`). Fixes a discovered defect where 5 of 7 reviewer-written verification SQL checks failed on hallucinated columns/tables.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#load_schema_hint, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 23, #Turn 24
- **relations:** verify step; two-reviewer council
- **verify-later:** grep/inspect `load_schema_hint`; `information_schema`; `agent_definitions.default_config`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Reframe step (post-veto)
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** "Reframe path is unit-tested but has never fired live (no veto since v4)" (HANDOFF_CURRENT_fixloop.md#State snapshot)
- **what:** After a guardian veto with rounds remaining, makes one attempt to reframe rather than reproposing the same shape: either a strictly narrower remediation (site-scoped interim fix allowed only if risks names the deferred structural fix) or an explicit "needs architecture review" declaration plus a minimal safe interim step. Capped at one attempt; built and unit-tested but has not fired live since v4 shipped.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#reframe, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 22, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md
- **relations:** decision router; escalation artifact; guardian veto (dartsonline run 8c770fd5)
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Escalation as first-class success terminal (diagnose_escalate)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "Escalation package persisted (kind=escalation)... The dead-end is now a hand-off" (NOTES(10)#Turn 23)
- **what:** When a plan is rejected-again or a revise budget is exhausted, the run persists a `kind='escalation'` artifact (decision, reason, round, diagnosis conclusion, final plan, both reviews) and completes via a distinct `complete_escalated` success terminal. Explicitly designed so "needs a human/architecture review" is a correct, successful output rather than a failure.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#escalate/complete_escalated, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 22
- **relations:** decision router; reframe step; dartsonline guides defect
- **verify-later:** grep/inspect `kind='escalation'`; `complete_escalated`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Write step / fix-implementer agent (F1.1b(c))
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "F1.1b(c): branch + PR — ✅ COMPLETE & PROVEN 2026-07-13 (PR #1 opened & merged)" (PLAN_fixloop_pilot.md)
- **what:** The loop's write organ. Given a `fix_correlation_id`, refuses anything whose latest council decision is not `approved`; reads current file bodies via the GitHub contents API; runs an LLM step (`sketch_to_files`) to turn the approved plan's sketches into complete new file bodies for ONLY the plan's named files; passes those through a deterministic hard allowlist; creates a `fix/<short-corr>` branch and commits via the git-adapter; gates on a build check; and on green opens a PR. `config_change` edits are deliberately NOT implemented by this agent.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §3b F1.1b(c), fixloop_eg_dartsonline/README_so_far.md
- **relations:** hard file allowlist; build gate; git-adapter write isolation; fix-implementer-orchestrator; PR as human terminal
- **verify-later:** grep/inspect `fix_correlation_id`; `approved`; `sketch_to_files`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Hard file allowlist (diagnose_prepare_fix_commit)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "Part 2a BUILT (chassis, commit a4c6cc63)... 7-case suite exercises the real logic" (NOTES(10)#Turn 25)
- **what:** A deterministic action sitting between the implementer's LLM step and the git-adapter: the approved plan's modify/add file list is a hard allowlist — a produced file outside the plan, a plan file the implementation is missing, or an empty/duplicate/no-op file all reject the whole implementation before anything touches git. Also assembles the branch name, commit message, and PR title/body (the "Q-H package"). This is the safety core that made the first live PR's diff exactly the approved plan.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql#prepare step, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 25, fixloop_eg_dartsonline/README_so_far.md
- **relations:** write step; build gate; plan validation
- **verify-later:** validateImplementation function; 7-case test suite

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Build gate (diagnose_build_gate)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "Build gate (golang Job): GREEN — === build gate: PASS ===" (README_so_far.md); "its first red correctly blocked a PR" (NOTES(10)#Turn 26-28)
- **what:** Before any PR is opened, changes must be built in a clean container (`gofmt` + targeted `go build`) in a short-lived golang-image k8s Job. Green routes to PR creation; red routes to a no-PR terminal with build log attached, branch left for human inspection — "no PRs for broken code." Chosen over GitHub Actions CI on the PR (Option B: broken implementations must never even become a visible PR). Its first live red catch was a genuine pre-existing bug, then fixed for real.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql#build_gate/check_gate, fixloop_eg_dartsonline/SUMMARY_write_step_position_2026-07-12.md#Option A/B/C, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 26-28
- **relations:** write step; hard file allowlist; PR as human terminal
- **verify-later:** diagnose_build_gate action; RBAC rbac-job-spawner.yaml pods/log grant

<!-- SOURCE: U13_docs024_small_dirs.md -->
### git_adapter_request generic adapter caller
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "git_adapter_request — ONE generic adapter caller (allowlisted verbs...)" (HANDOFF_CURRENT_fixloop.md#F1.1b(c) CODE COMPLETE)
- **what:** A single generic workflow action used for all git-adapter calls from the write step, with the adapter action name and data fields/literals supplied per-step from config, and an explicit note that `delete_repo` is unreachable through this path.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql#create_branch/commit_files/create_pr config, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md
- **relations:** git-adapter new actions; write step
- **verify-later:** grep/inspect `delete_repo`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### isRepoCloningAgent spawn gate / GITHUB_READ_TOKEN injection
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "the spawned implementer pod gets GITHUB_READ_TOKEN via the already-deployed isRepoCloningAgent gate" (HANDOFF_CURRENT_fixloop.md)
- **what:** An existing spawn-gate mechanism (already used for diagnose-agent) that injects a read-only GitHub token into a dedicated, ephemeral pod when the spawned agent type is listed in `isRepoCloningAgent`. `fix-implementer` was added to this list. Only works when the agent runs as a dedicated spawned pod — the generic in-chassis orchestrate path bypasses the gate entirely.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql#header point 1, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#FIRST END-TO-END RUN blocked
- **relations:** fix-implementer-orchestrator; git-adapter as sole write credential holder; diagnose_read_repo_files
- **verify-later:** grep/inspect `isRepoCloningAgent`; `fix-implementer`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### diagnose_read_repo_files action
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "diagnose_read_repo_files — plan's modify/add files via GitHub contents API (raw media type; read token from spawn gate; modify-404 = hard error)" (HANDOFF_CURRENT_fixloop.md)
- **what:** Fetches the current bodies of the approved plan's modify/add files via the GitHub contents API at an explicit ref, using the token from the spawn gate. A missing file for a "modify" operation is a hard error.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql#read_current_files, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md
- **relations:** isRepoCloningAgent spawn gate; write step
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### fix-implementer-orchestrator (dedicated-pod wrapper)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "0NN_fix_implementer_orchestrator.sql — F1.1b(c) fix: run the implementer as a DEDICATED POD" (header)
- **what:** A thin wrapper agent (`spawn_agent(fix-implementer)` → `call_agent` → `complete`) built to fix a real first-run failure: firing fix-implementer via the generic orchestrate path ran it IN the shared chassis pod, so the isRepoCloningAgent gate never fired. Mirrors the existing diagnose-orchestrator→diagnose-agent pattern exactly. Needed no image rebuild.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer_orchestrator.sql, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#FIRST END-TO-END RUN, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 26-28
- **relations:** isRepoCloningAgent spawn gate; write step
- **verify-later:** 092_TRIGGER_fix_implementer_v1.sh target agent type

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Whole-file rewrite strategy (implementer's LLM step)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "41KB whole-file rewrite → allowlist PASS" (README_so_far.md)
- **what:** The implementer's `sketch_to_files` LLM step outputs the COMPLETE new body of every plan-named file, never a diff/patch, with hard rules forbidding drive-by changes. Explicitly named as not scaling to very large files (32000 max_tokens gives headroom for one ~41KB file but not much more) — a diff/patch strategy is logged as future work (F1.2).
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql#sketch_to_files, fixloop_eg_dartsonline/README_so_far.md
- **relations:** write step; hard file allowlist
- **verify-later:** grep/inspect `sketch_to_files`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### PR as human terminal / nothing merges itself
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "PR — waits for you. Nothing merges itself." (README_so_far.md)
- **what:** A governing design principle: the platform's most autonomous act is opening a pull request; nothing ever merges its own work. Isolation model (2026-07-12): fix/* branches live on the same repo (no fork); the owner alone chooses what merges to main. This is why "escalation" is treated as a success, not a failure.
- **sources:** fixloop_eg_dartsonline/README_so_far.md, fixloop_eg_dartsonline/0NN_fix_implementer.sql#header, fixloop_eg_dartsonline/SUMMARY_where_we_are_2026-07-13.md
- **relations:** write step; NO FORK decision; escalation as first-class success terminal
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### NO FORK decision (abandoned fork idea)
- **category:** fix-loop
- **status-signal:** superseded
- **status-evidence:** "Decisions CLOSED... 4. NO FORK: isolation = fix/* branches + owner-gated merges on this repo" (HANDOFF_CURRENT_fixloop.md)
- **what:** The owner raised, and then explicitly closed, the idea of running the fix-loop's next phase against a separate forked repository. The decision landed instead on branch+PR isolation on the same repo, which is what was actually built.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 25, #Turn 26, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#Decisions CLOSED
- **relations:** PR as human terminal; write step
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Round-counting scope bug (correlation vs orchestration)
- **category:** fix-loop
- **status-signal:** superseded (fixed in source; deploy gap tracked separately)
- **status-evidence:** "the deployed v1.0.1107 binary counts council rounds per correlation... does NOT carry the orchestration_id-scoping fix" (NOTES(10)#Turn 22)
- **what:** Council-round counting was originally scoped per `correlation_id`, accumulating council_report rows across every proposer re-run — so a fresh proposer run on a correlation with review history would start mid-count and exhaust its revise budget without ever reproposing. Fixed in source to count per `orchestration_id`, but a same-tag deploy trap meant the fix did not reach the running binary for one further benchmark cycle.
- **sources:** fixloop_eg_dartsonline/HANDOFF_turn21_2026-07-10.md#Key accomplishments, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 22
- **relations:** revise loop; same-tag deploy trap gotcha
- **verify-later:** grep/inspect `correlation_id`; `orchestration_id`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### fixloop-digest / awareness surface
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** "IN FLIGHT (turn 29): the AWARENESS SURFACE... Built + committed, awaiting the next chassis image" (HANDOFF_CURRENT_fixloop.md)
- **what:** A deterministic (no-LLM-in-path) digest agent composing a window (default 24h) summary of fix-loop activity — status/terminal/gate/PR outcomes, decisions per correlation, and agent_definitions_backup snapshots — persisted to `doc_notes` (categories `["digest","fixloop"]`). Built to satisfy the owner's standing rule "more awareness before wider autonomy." v1 is manual-trigger only; a daily cadence is deliberately deferred.
- **sources:** fixloop_eg_dartsonline/0NN_fixloop_digest.sql, fixloop_eg_dartsonline/093_TRIGGER_fixloop_digest_v1.sh, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#IN FLIGHT
- **relations:** owner standing rule; diagnosis_artifacts table; future council roster
- **verify-later:** whether the chassis image carrying `fixloop_digest` action has shipped; `doc_notes` rows with categories ? 'digest'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Owner standing rule: awareness before autonomy
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "Owner standing rule (2026-07-12): 'more awareness BEFORE wider autonomy.'" (0NN_fixloop_digest.sql header)
- **what:** An explicit governance principle: before the council is widened with more reviewer perspectives or migration/feature-building agents, the owner must first have a reliable way to see what the loop has been doing and deciding. Directly produced the fixloop-digest slice being scheduled ahead of the F2 roster expansion.
- **sources:** fixloop_eg_dartsonline/0NN_fixloop_digest.sql#header, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#Decisions CLOSED, fixloop_eg_dartsonline/SUMMARY_where_we_are_2026-07-13.md
- **relations:** fixloop-digest; future council roster (deferred by this rule)
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Future council roster (aspirational reviewers)
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** "Initial roster: a guidelines agent... a reuse agent... a bug-historian... a compliance/legal eye; pipeline guardians... specialist knowledge agents" (RUNBOOK(9)#THE TASK) — none beyond the guardian/edit-quality pair were built
- **what:** The original council vision named a much wider roster than what was built: a guidelines agent, a reuse agent, a bug-historian, a compliance/legal eye, one pipeline-guardian per master workflow, and specialist knowledge agents — motivated by a real incident where a chat reinvented a trigger+triage SQL pair that already existed. Only a generic edit-quality reviewer and a single cross-pipeline guardian shipped.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#THE TASK, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#Direction
- **relations:** two-reviewer council; owner standing rule; architecture-change visibility (Q-E)
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Architecture-change visibility (Q-E signals)
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** "Q-E architecture-change signals: ... STILL OPEN" (RUNBOOK(10)#STILL OPEN)
- **what:** A standalone goal from the original task charter — make it loud when a proposed change is accidentally fundamental (touching platform contracts, message shapes, many packages, exported signatures) before it ships. Never built as a dedicated formal detector; what exists in practice is the pipeline-guardian reviewer's informal judgement, which has correctly identified architecture-level changes dressed as contained fixes.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#THE TASK, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#STILL OPEN Q-E
- **relations:** two-reviewer council; guardian veto; future council roster
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Guardian veto surfacing an architecture-level fix (dartsonline)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "the guardian vetoed all three edits as 'an architecture change dressed as a contained fix'" (NOTES(10)#Turn 22, orch 8c770fd5)
- **what:** A concrete, live-observed instance of the guardian reviewer correctly recognizing that a minimal-looking three-edit plan was actually architecture-level, vetoing it and proposing a safer alternative in its notes — deliberately NOT auto-applied since it fixes only one site while leaving the platform-wide cause live everywhere.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 22
- **relations:** architecture-change visibility; reframe step; platform-not-site-data fix philosophy; dartsonline guides defect
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Platform-not-site-data fix philosophy
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "Owner ruled: the F1 edit plan targets the PLATFORM, not dartsonline's data." (NOTES(10)#Turn 2)
- **what:** An owner ruling that any fix plan must target the platform mechanism rather than a single site's data rows — because the causes of the benchmark bug are relay-level and a data-only fix would fix one site while leaving every other site exposed. Directly shapes the proposer's prompt rules and the guardian's refusal to accept a scoped data-only remediation as a final answer.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 2, fixloop_eg_dartsonline/0NN_fix_proposer.sql#propose prompt rule 1
- **relations:** dartsonline guides defect; guardian veto surfacing architecture-level fix; reframe step
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### config_change edit operation type
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "config_change edits in a plan are NOT implemented by this agent... the PR body carries them for the human" (0NN_fix_implementer.sql header)
- **what:** A plan-edit operation type reserved for edits that target `agent_definitions` workflow-JSON configuration rather than repo files. The proposer's prompt requires such edits be explicitly labelled, but the fix-implementer deliberately does not apply them — they are left in the PR body for a human to apply by hand.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#propose prompt rule 5, fixloop_eg_dartsonline/0NN_fix_implementer.sql#header
- **relations:** write step; hard file allowlist; plan validation
- **verify-later:** grep/inspect `agent_definitions`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### F1.2 deferred work items (ref/base as input; fix_pr artifact; diff strategy)
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** "Open (F1.2): ref/base are live-set to the active working branch because origin/main is stale — make them a per-run INPUT" (PLAN_fixloop_pilot.md)
- **what:** A cluster of known-but-deferred improvements: the implementer's git ref/base/from_branch are hardcoded via a live jsonb_set patch rather than a per-run input field; a dedicated `kind='fix_pr'` diagnosis_artifacts row for the PR result is deferred; a diff/patch implementation strategy for large files.
- **sources:** fixloop_eg_dartsonline/PLAN_fixloop_pilot.md#F1.1b(c), fixloop_eg_dartsonline/0NN_fix_implementer.sql#header
- **relations:** whole-file rewrite strategy; write step
- **verify-later:** grep/inspect `kind='fix_pr'`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### F3 learning record / bug_records (never built)
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** "F3 — Learning... bug_records (category taxonomy, recurrence checks feeding the historian)" (RUNBOOK(10)#Phased plan) — no bug_records table or historian agent found
- **what:** The original phased plan's final stage: categorize confirmed bugs into a taxonomy so recurring classes are caught earlier, feed guideline-amendment proposals to the human, and enrich the corpus from what the loop learns. Never designed in detail or built.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Phased plan F3, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#Phased plan F3
- **relations:** future council roster (bug-historian); guideline-gap side-task mechanism
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Guideline-gap side-task mechanism (designed, unclear if built)
- **category:** fix-loop
- **status-signal:** unknown
- **status-evidence:** "Q-D completion — guideline-gap = SIDE-TASK (does not block the fix): a work item carrying the evidence; handler drafts a concrete amendment and opens a PR against the GUIDELINE DOCS" (NOTES(9)#DECISIONS)
- **what:** A 2026-07-07 decision that when a reviewer finds the guidelines themselves fell short, that finding becomes a side-task work item whose handler drafts a guideline-amendment PR, with gaps accumulating toward the F3 learning record. No implementation of this side-task handler was found in the files read.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(9).md#DECISIONS, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#QUESTIONS Q-D
- **relations:** F3 learning record; future council roster
- **verify-later:** search for a guideline-amendment work item type / handler agent in agent_definitions

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Q-G reviewer context (open design question, v1 answered narrowly)
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** "Q-G v1 = role prompts + plan + diagnosis (no per-reviewer corpora yet)" (PLAN_fixloop_pilot.md)
- **what:** The open question of how much context each council reviewer should see. What shipped is the narrowest option: both reviewers get the same role prompt, the persisted plan, the diagnosis conclusion, and (from F2.3b) a live schema hint — no per-reviewer curated corpus exists yet.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#STILL OPEN Q-G, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §F2.1
- **relations:** two-reviewer council; schema hint for reviewers; future council roster
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Q-H human-facing result package
- **category:** fix-loop
- **status-signal:** deployed (v1)
- **status-evidence:** "PR carrying the Q-H package" appears repeatedly as delivered (HANDOFF_CURRENT_fixloop.md)
- **what:** The decided shape of what a human ultimately sees: the PR body carries the diagnosis conclusion, the approved plan, and the council's decision/reviews together, so a human reviewing a fix-loop PR never has to go hunting through diagnosis_artifacts. The equivalent package for an escalated run is the escalation artifact.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql#prepare/create_pr, fixloop_eg_dartsonline/README_so_far.md
- **relations:** write step; escalation as first-class success terminal
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### SEED_first_writestep_diagnosis pattern (hand-authored diagnosis for downstream testing)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "the diagnosis is hand-written (the diagnosis LOOP is separately proven); proposer, council, and implementer that consume it all run for real" (SEED_first_writestep_diagnosis.sql header)
- **what:** A reusable technique for exercising downstream stages honestly without waiting for a live CONFIRMED diagnosis on a suitable bug: hand-author a CONFIRMED `orchestration_states` row for a real, tiny, zero-risk defect, then run the real proposer→council→implementer chain against it for real. Fabricating evidence rows was explicitly rejected as an option.
- **sources:** fixloop_eg_dartsonline/SEED_first_writestep_diagnosis.sql, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#First-run design decision
- **relations:** write step; tier-coverage guard
- **verify-later:** grep/inspect `orchestration_states`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### F0.3 per-iteration notes (never built)
- **category:** fix-loop
- **status-signal:** abandoned
- **status-evidence:** "Per-iteration notes — NOT MET, because F0.3 does not exist yet" (RUNBOOK(10)#F0 plumbing criteria)
- **what:** One of F0's four original acceptance criteria — writing the loop's per-iteration/per-step reasoning into task-specific running notes — was designed but never implemented across the entire workstream. The `diagnosis_artifacts.kind='iteration_note'` column value exists specifically to carry this, unused.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Phased plan F0.3, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md
- **relations:** doc_notes/travelling-docs integration boundary; diagnosis_artifacts table
- **verify-later:** grep/inspect `diagnosis_artifacts.kind='iteration_note'`

<!-- SOURCE: U14_docs019_runbooks.md -->
### Diagnosis→Fix Loop programme (F0–F3)
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "DISCUSSION COMPLETE for F0/F1 (2026-07-07): Q-A/B/C/D/F all decided — CUTOVER-READY … First slice: F0.1"; no build claimed.
- **what:** The v2 workstream turning the read-only diagnosis loop into a diagnosis→fix system, phased: F0 intake/observability/egress (documented route in and out, fetchable bundles, per-task running notes); F1 fix-on-a-branch; F2 council of reviewers + decision-maker with architecture-change visibility; F3 learning (bug records, guideline amendments, corpus enrichment). Mission: use everything available — code corpus, schemas, runtime, the guidelines themselves — with checks, balances and second opinions built in.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#the-task; docs019/RUNBOOK_diagnosis_fix_loop(9).md#phased-plan; docs019/RUNBOOK_diagnosis_fix_loop(9).md#current-position
- **relations:** read-only diagnosis loop (the base); council of reviewers; docs026 stage-3 council agents (this register's own consumer)
- **verify-later:** diagnosis_artifacts migration; needs_diagnosis items; fixer agent existence

<!-- SOURCE: U14_docs019_runbooks.md -->
### diagnosis_artifacts bundle egress
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "DECIDED 2026-07-07 (owner): Q-A diagnosis_artifacts table, written through inside assemble (unified-table refinement: kind ∈ {bundle, iteration_note})".
- **what:** Durable per-iteration bundle persistence: a diagnosis_artifacts table written through inside the assemble action (zero workflow-shape change, deliberately off the tools chat's emit-adjacent surface), with a documented fetch route. doc_notes was considered and set aside (notes are prose for humans; bundles are machine-replayable evidence with different retention). Sizing memory: bundles ~60KB × ≤5 iterations vs the 1.27MB collected_data incident.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-A); docs019/RUNBOOK_diagnosis_fix_loop(9).md#collision-surface
- **relations:** oversize delivery doctrine; per-task running notes; diagnose_assemble_bundle
- **verify-later:** diagnosis_artifacts table (exists?); assemble write-through code

<!-- SOURCE: U14_docs019_runbooks.md -->
### needs_diagnosis intake in a pipeline='diagnose' namespace
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "Q-B needs_diagnosis item in a NEW pipeline='diagnose' namespace (null-site allowed; envelope extends 084; manual trigger retained)"; "ENABLER CONFIRMED 2026-07-07: anchorless (site-less) diagnosis runs now SURVIVE".
- **what:** Task input rides the existing work-item dispatch + immune system: a needs_diagnosis site_work_items row in its own pipeline namespace, with null-site allowed for pure code bugs (enabled by the tools chat's load_runtime error-routing so anchorless runs degrade gracefully — ~26 min / 5 iterations observed). The canonical envelope adopts/extends the tools chat's 084 trigger with subject_type/subject_key.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-B); docs019/RUNBOOK_diagnosis_fix_loop(9).md#what-already-exists
- **relations:** build pump + immune system (the ride); generic envelope trigger (retained manual path)
- **verify-later:** pipeline='diagnose' rows; 084_TRIGGER_diagnose_v1.sh subject fields

<!-- SOURCE: U14_docs019_runbooks.md -->
### Fix-on-a-branch with an isolated fixer agent
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "Q-C separate fixer agent (isolated write token; constrained edit plan; gofmt+build in a spawned job pre-PR)" — decided 2026-07-07, not built.
- **what:** F1: a CONFIRMED diagnosis drives a proposed fix committed to a separate git branch via the git adapter, PR opened, human amends/ditches/applies. The loop's core stays read-only; the write surface is a SEPARATE fixer agent holding the only write token (the spawn token-gate pattern), producing a constrained edit plan validated by gofmt+build in a spawned job before the PR.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#the-task (item 2); docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-C)
- **relations:** repo-cloning token gate (the pattern); council of reviewers (gate before finalising)
- **verify-later:** fixer agent definition; git-adapter write paths

<!-- SOURCE: U14_docs019_runbooks.md -->
### Council of reviewers with a decision-maker
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) F2 "Independent reviewers (roster above), each a small agent with its own curated context … a decision-maker aggregates" — designed 2026-07-06/07, open Q-E/G/H.
- **what:** Before any fix is finalised, independent specialist agents each judge it from their own perspective and send structured opinions (verdict-wire-style: verdict + citations + objections + suggested alternative) to a decision-maker. Initial roster: guidelines agent (adherence to 000-0xx — or did the guideline fall short), reuse agent (code AND docs), bug-historian, compliance/legal, pipeline guardians (one per master workflow, seeded from the builder relay map), and specialist knowledge agents ("we already have one of these"). Precursor idea from the thin slice: build-time liability and MORALITY review contributors applying a configured, layered standard with contested calls routed to a human.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#the-task (item 3); docs019/RUNBOOK_thin_slice(27).md#next-improvements (item 3)
- **relations:** hard-veto semantics; three-tier citation (opinion contract); docs026 council-agents stage
- **verify-later:** reviewer agent definitions (none yet); Q-G reviewer-context decision

<!-- SOURCE: U14_docs019_runbooks.md -->
### Hard-veto flag semantics for reviewers
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "Q-D council topology — VETO SEMANTICS DECIDED (owner, 2026-07-07): parallel reviewers → decision-maker BY DEFAULT … a hard_veto flag, attachable at multiple scopes … converts that reviewer's negative verdict into a BLOCK".
- **what:** All council opinions are advisory by default and weighed together; a hard_veto flag — attachable per reviewer agent, per pipeline, or per tool/component, most-specific-scope contemplated — makes that reviewer's negative verdict blocking. Accessibility and legal are the motivating hard-veto cases. A guidelines-reviewer "the guideline itself fell short" finding leans side-task (gap, not violation), not block.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-D)
- **relations:** council of reviewers; learning layer (guideline-gap side-task)
- **verify-later:** where the flag lives (reviewer column vs council config)

<!-- SOURCE: U14_docs019_runbooks.md -->
### Architecture-change visibility detector
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "Q-E architecture-change signals: packages touched breadth; platform/ vs actions/; exported-signature diffs vs the corpus; message/topic/schema/contract changes; migration presence. Which are load-bearing?" (open, F2-phase).
- **what:** Make it loud when a proposed change is accidentally fundamental — touching platform contracts, message shapes, many packages, exported signatures — before it ships; runs as one council reviewer. Candidate signals enumerated; exported-signature diffs against the code_symbols corpus is the notable reuse of the diagnosis infrastructure.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#the-task (item 4); docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-E)
- **relations:** council of reviewers; code_symbols corpus
- **verify-later:** n/a (not built)

<!-- SOURCE: U14_docs019_runbooks.md -->
### Learning layer — bug records and guideline-amendment side-tasks
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) F3 "bug_records (category taxonomy, recurrence checks feeding the historian); guideline-amendment proposals routed to the human"; Q-D "guideline-gap SIDE-TASK (amendment PR against the guideline docs, human terminal, fix unblocked, F3 recurrence record)".
- **what:** The feedback layer: recorded bugs with a category taxonomy and recurrence checks (feeding the bug-historian reviewer so a class never repeats); when a fix exposes a guideline gap, a side-task raises an amendment PR against the guideline docs with the human as terminal approver while the fix itself proceeds; corpus and doc enrichment feed back into retrieval.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#phased-plan (F3); docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-D)
- **relations:** corpus enrichment policy; council (bug-historian); coverage baseline (guideline home)
- **verify-later:** bug_records table (absent); amendment-PR mechanism

<!-- SOURCE: U14_docs019_runbooks.md -->
### Loop-worthiness test (five-criteria intake doctrine)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** diagnosis_fix_loop(9) "LOOP-WORTHINESS TEST (doctrine — apply before every intake)" — applied three times in the same file (pilot #1 downgraded, candidate 2 forked, guides pilot confirmed).
- **what:** A task is loop material only when ALL hold: (1) a SYMPTOM about system behaviour, not a feature request; (2) a causal mechanism plausibly exists in code+data+runtime; (3) not answerable by one or two direct queries (mandatory cheap pre-check first); (4) bounded to one symptom; (5) verified CURRENT at intake — symptoms are perishable. Feature absences → build routes; quality judgements → council/auditors; one-query questions → the query. Demonstrated by downgrading the roadmap-gap "bug" (findable by reading two files) to a builder-queue item.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#loop-worthiness; docs019/RUNBOOK_diagnosis_fix_loop(9).md#previous-pilot-1
- **relations:** F0 guides pilot; falsification eval gate
- **verify-later:** n/a (doctrine)

<!-- SOURCE: U14_docs019_runbooks.md -->
### F0 pilot — the guides-route differential diagnosis
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "★ F0 PILOT — CONFIRMED 2026-07-07: nav links to a guides section that has no content" with pre-registered criteria; ordered after F0.1 plumbing.
- **what:** The chosen first fix-loop pilot: dartsonline published a Guides nav link and blank /guides/index.html while gamesdesign (same platform) has working guides — a two-site DIFFERENTIAL, the strongest evidence shape. Standing hypothesis for the loop to confirm/refute FROM CODE: reconcile_site_plan's routing table has no "guide" entry (blog-index present, tool commented out), so planner-emitted guide pages were silently dropped while nav — generated from the PLAN, not the built set — published the link. Two earlier pilot candidates were downgraded via the loop-worthiness test.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#f0-pilot; docs019/RUNBOOK_builder_route(21).md#queue (item 7)
- **relations:** loop-worthiness test; reconcile routing table; nav-grounded-in-built-set principle
- **verify-later:** load_work_item_actions.go routing table; the pilot's run artifacts once executed

<!-- SOURCE: U14_docs019_runbooks.md -->
### Per-task running notes via doc_notes (travelling docs reuse)
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** diagnosis_fix_loop(9) "Q-F DIRECTION SET (2026-07-07): REUSE doc_notes. The terminal-diagnosis note already exists on their side (pending their 3b subject threading)"; "the diagnose-agent workflow is ALREADY rewired by them: emit → persist_note → complete".
- **what:** Live monitoring of what the loop is doing and why: per-iteration and per-step reasoning written to a task-specific notes home. Decision: reuse the tools chat's doc_plans/doc_notes infrastructure (terminal diagnosis note already wired via persist_note with a strict no-guessing subject gate); per-iteration rows are additional doc_notes entries pending the owning thread's sign-off; category convention `diagnosis`.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-F); docs019/RUNBOOK_diagnosis_fix_loop(9).md#what-already-exists
- **relations:** doc_plans/doc_notes infrastructure; reasoning-state handoff; thread-boundary convention
- **verify-later:** doc_notes rows with category diagnosis; persist_diagnosis_note action

<!-- SOURCE: U15_docs019_running_notes.md -->
### Diagnosis→fix loop workstream (founding)
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** NOTES_running_fixloop(9).md THREAD STATE AT CUTOVER: "DECIDED AND RECORDED... STILL OPEN (F2-phase)... FIRST ACTION: slice F0.1 with pre-registered criteria."
- **what:** The founding thread (2026-07-06/07) that pivots the read-only diagnosis loop into a diagnosis→fix system: documented intake (`needs_diagnosis` work item, `pipeline='diagnose'` namespace), live per-iteration reasoning persisted to a new `diagnosis_artifacts` table (kind ∈ bundle|iteration_note, written through inside the assemble action — off the parallel tools-chat's `doc_notes` surface, only the terminal note relayed there), fixes produced by a SEPARATE fixer agent with an isolated git write token (spawn-gate pattern) producing a constrained edit plan validated by gofmt+build before any PR, and a council of parallel specialist reviewers feeding a decision-maker (see hard_veto flag concept below). This is the same workstream documented in far greater operational detail in `docs024_key_docs_latest/fixloop_eg_dartsonline/` — this file is its origin notes.
- **sources:** NOTES_running_fixloop(9).md (full); NOTES_running_synthesis_v4(39).md 2026-07-06/07 entries (same founding, condensed).
- **relations:** Loop-worthiness test doctrine; hard_veto council flag; diagnosis loop; roadmap-phase enforcement gap.
- **verify-later:** `diagnosis_artifacts` table; the fixer agent's isolated write-token/spawn-gate; cross-reference against `docs024_key_docs_latest/fixloop_eg_dartsonline/` for the fuller, later-stage version of this same concept.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Council hard_veto flag / decision-maker model
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** "Q-D veto semantics decided (owner)... Flag-based: DEFAULT = decision-maker weighs all opinions; a hard_veto flag at reviewer/pipeline/tool/component scope makes that reviewer's negative verdict a BLOCK" (NOTES_running_fixloop(9).md).
- **what:** The fix-loop's review-arbitration model: a parallel council of specialist reviewers (guidelines/reuse/bug-historian/compliance/per-pipeline guardians) feeds a decision-maker by default (advisory), except where a `hard_veto` flag is set at reviewer/pipeline/tool/component scope (accessibility and legal are the motivating cases), which makes that reviewer's negative verdict an unconditional block. A guideline-gap found during review is a SIDE-TASK (a work item that drafts an amendment PR against the guideline docs, human-terminal) rather than something that blocks the fix.
- **sources:** NOTES_running_fixloop(9).md "Q-D veto semantics decided" and "F0/F1 design settled" (DECISIONS).
- **relations:** Diagnosis→fix loop workstream founding.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Loop-worthiness test doctrine
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** "Owner asked whether the loop fits the dartsonline quality problem. Answer: decomposed via the new LOOP-WORTHINESS TEST (symptom-not-feature; mechanism-plausible; not one-query-answerable; single-symptom)" (NOTES_running_fixloop(9).md); a fifth criterion (verify symptom currency at intake) added after a pilot candidate "evaporated" (was fixed live before the loop ran).
- **what:** A pre-registered five-criterion test for whether a candidate bug is worth running the diagnosis/fix loop on: it must be a genuine symptom (not a disguised feature request), the mechanism must be plausible from code, it must not be answerable by one query, it must be a single coherent symptom, and its currency must be reverified at intake (since bugs can be fixed out from under a pilot mid-triage — this happened twice in this thread alone).
- **sources:** NOTES_running_fixloop(9).md multiple 2026-07-07 pilot-selection entries.
- **relations:** Diagnosis→fix loop workstream founding.

<!-- SOURCE: U16_docs019_design_plans.md -->
### Diagnosis→fix loop programme (F0–F3)
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** HANDOFF_fixloop(8) (2026-07-07): "All F0/F1 design questions are DECIDED"; README_02: "a pipeline that turns 'something is wrong' into 'here is a reviewable, evidence-backed proposal' … exercised on exactly one bug. The write step (plan → PR) is half-built."
- **what:** The evolution of the diagnosis loop into a fix pipeline: symptom → cited diagnosis → constrained edit plan → adversarial council review → revision informed by reviewer-requested DB queries → approved plan or honest escalation. Phased F0 (persistence/intake) → F1 (write step) → F2 (council expansion) → F3, driven by open questions Q-A…Q-H resolved in the discussion thread. The valuable output is the general pattern, not the bug-fixing.
- **sources:** HANDOFF_fixloop_thread(8).md; README_02_evidence_backed_proposals.md; README_overview.md (F1.1b(c) status)
- **relations:** council pattern; fix-implementer; pilot worthiness test; docs026 concept-council mission
- **verify-later:** RUNBOOK_diagnosis_fix_loop.md + NOTES_running_fixloop.md (units U14/U15); fix-implementer seed

<!-- SOURCE: U16_docs019_design_plans.md -->
### Council pattern: adversarial multi-agent review with deterministic aggregation
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** README_02: "the roster of 2 (edit-quality, guardian) is explicitly a skeleton"; "the guardian vetoed an architecture change dressed as a fix"; veto semantics decided per HANDOFF_fixloop(8) delta history ("flag-based hard veto, default advisory → decision-maker").
- **what:** Multiple reviewer agents each examine the proposed fix plan from one lens; a deterministic rule (not a third model) aggregates their positions; specified veto semantics are flag-based hard veto with advisory as default. Reviewers can demand facts, and the loop runs the queries itself rather than letting the proposer argue (self-verification instead of self-belief). Three runs running, the council correctly ruled the test bug's proper fix beyond a constrained plan's mandate.
- **sources:** README_02_evidence_backed_proposals.md; HANDOFF_fixloop_thread(4)-(8).md deltas; README_comprehensive_documentation_categorisation.md (veto description)
- **relations:** expanded council bench; escalation as success; guardian-from-decision-record (Q-G)
- **verify-later:** council agent seeds + the aggregation rule in Go; fixloop_eg_dartsonline docs (docs024)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Expanded council bench (expert-per-area reviewers)
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** README_02: "the runbook's F2 design already names the full bench … Adding a reviewer is a seed change + prompt + curated context."
- **what:** The planned F2 roster beyond the two-agent skeleton: a guidelines agent (conformance to 000–0xx, or did the guidelines fall short), reuse agent (are we rebuilding something that exists, code and docs), bug-historian (has this class recurred), compliance eye, pipeline guardians one per master workflow, and specialist knowledge agents. Reviewer areas are expected to correlate with the docs024 documentation categories — the direct bridge to the docs026 concept register's council-agent goal.
- **sources:** README_02_evidence_backed_proposals.md#3; README_comprehensive_documentation_categorisation.md
- **relations:** council pattern; concept register mission; documentation categories as expertise areas
- **verify-later:** whether any bench agents beyond edit-quality/guardian were seeded

<!-- SOURCE: U16_docs019_design_plans.md -->
### Fix-implementer constrained write step
- **category:** fix-loop
- **status-signal:** partial
- **status-evidence:** README_overview: "F1.1b(c) is code complete (b367a602) … validated as far as it can be without the deploys"; deploy checklist listed, first end-to-end run pending.
- **what:** A 15-step seeded agent: load plan/council/diagnosis → APPROVED gate (mirror of the CONFIRMED gate) → diagnose_read_repo_files fetches current file bodies via the GitHub contents API with a hard rule that a modify-file 404 is a refusal (whole-file rewrites of unseen files would be hallucination by construction) → sketch_to_files whole-file rewrites ("the diff a human reviews must contain ONLY the plan") → deterministic file allowlist → create fix/* branch → commit via git_adapter_request (one generic adapter caller; verbs allowlisted to commit/create_branch/create_pull_request so delete_repo is structurally unreachable) → build gate (golang Job): green → PR into main, red → NO PR, branch + build log left. Runs on the read-token spawn gate.
- **sources:** README_overview.md (landed pieces + deploy checklist); README_02_evidence_backed_proposals.md
- **relations:** hard deterministic gates; human-gate-never-moves; seeded-bug first run; build gate options A/B/C
- **verify-later:** 0NN_fix_implementer.sql; 092_TRIGGER_fix_implementer_v1.sh; git-adapter branch/PR ops; RBAC for pods/log

<!-- SOURCE: U16_docs019_design_plans.md -->
### Hard deterministic gates between every LLM step
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** README_02 lists them as built pattern: "CONFIRMED gate, plan validator, file allowlist. The models propose; plain Go code decides what proceeds."
- **what:** No LLM output passes into consequence unchecked: the diagnosis must be CONFIRMED (gate), the plan must validate, the files must be on a deterministic allowlist, the build must pass, before anything advances. Complexity and authority live in plain Go; the models only propose. The same shape as keeping convergence guards in the engine rather than in workflow conditionals.
- **sources:** README_02_evidence_backed_proposals.md#1; README_overview.md
- **relations:** council aggregation rule; fix-implementer; thin-workflows rule
- **verify-later:** the gate implementations in the fixloop actions

<!-- SOURCE: U16_docs019_design_plans.md -->
### The human gate never moves (nothing merges itself)
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** README_02: "one structural commitment: the human gate never moves. More autonomy upstream … never past the PR. Nothing merges itself."
- **what:** Autonomy may widen upstream (diagnose, plan, revise, commit-to-branch) but the merge is permanently human. The PR is the fixed boundary of machine authority in the fix loop — a simpler, harder commitment than the graduated trust machinery, and orthogonal to it.
- **sources:** README_02_evidence_backed_proposals.md#2; README_overview.md (red build → NO PR)
- **relations:** trust ledger (graduated autonomy elsewhere); awareness surface; fork isolation
- **verify-later:** absence of any auto-merge path in the write step

<!-- SOURCE: U16_docs019_design_plans.md -->
### Escalation as a first-class success
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** README_02 pattern list: "'this is beyond my mandate' is a correct output, packaged for you"; the council produced exactly this on the test bug three times.
- **what:** When a fix exceeds the constrained plan's mandate (architecture-level causes), the loop's correct output is an honest escalation package for the human, not a forced plan. Treating refusal-to-proceed as success is the organisational analogue of UNVERIFIABLE-beats-guessing.
- **sources:** README_02_evidence_backed_proposals.md; README_02 §6 (the escalate decision explained)
- **relations:** cite-or-abstain; council pattern
- **verify-later:** escalation package format in the fixloop runbook

<!-- SOURCE: U16_docs019_design_plans.md -->
### Fix-loop value proposition: unattended, cited, consistent
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** README_02: "The value proposition (decided 2026-07-09): not 'the loop finds what humans can't' … The proposition is unattended, cited, consistent — the 3am diagnosis with a paper trail."
- **what:** A recorded decision reframing what the loop is for: on this platform bugs are legible to anyone with schema access and patience, so the differentiation is not superhuman insight but unattended operation with citations and consistency — a package instead of a hunch, reconstructible after the fact by one correlation id. Every design choice flows from it.
- **sources:** README_02_evidence_backed_proposals.md#2
- **relations:** falsification-first; awareness surface; diagnosis artifacts persistence
- **verify-later:** decision record in NOTES_running_fixloop (U15)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Awareness surface before wider autonomy
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** README_02: "the missing organ is a push surface: a periodic digest … before autonomy widens … the awareness surface gets built first" — a recommendation, not built.
- **what:** The named risk is not wrong action but unknown action — drift compounding silently while trails exist only pull-side. Proposed standing gate: before councils multiply or migration agents exist, build a push digest (what ran, what was decided and by which rule, what was escalated, what the council almost approved). "It must explain what it's doing, or it doesn't get to do more." The grown-up form of the parked F0.3 per-iteration notes.
- **sources:** README_02_evidence_backed_proposals.md#4
- **relations:** diagnosis artifacts persistence; decision log (the governance twin); human-gate-never-moves
- **verify-later:** whether any digest mechanism exists

<!-- SOURCE: U16_docs019_design_plans.md -->
### Fork isolation of the write surface
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** README_02 §5: "the strongest isolation is that the write surface points only at the fork … a designed slice, not huge" — proposed, not executed in these files.
- **what:** Point the loop's git-adapter credential, intake defaults and corpus indexing at a fork of the repo, making the main repo physically unwritable by the loop rather than protected by review discipline; the human pulls reviewed changes across. Folds in "mission and objectives correct in the first place": the fork's constitution/mission docs become the councils' curated context so conformance is checked against human-authored documents.
- **sources:** README_02_evidence_backed_proposals.md#5
- **relations:** human gate; guardian-from-decision-record; external rollback
- **verify-later:** git-adapter repo config; whether a fork exists

<!-- SOURCE: U16_docs019_design_plans.md -->
### Pilot worthiness test and the dartsonline guides pilot
- **category:** fix-loop
- **status-signal:** deployed
- **status-evidence:** HANDOFF_fixloop(8): "★ THE PILOT IS CONFIRMED (2026-07-07) … Two earlier candidates were rejected … that triage history is itself the worthiness test working."
- **what:** A five-criteria test for whether a bug deserves the diagnosis loop, exercised through three candidates: the chrome/nav defect (dropped — got fixed; perishability lesson), the nav-links-to-never-rendered-pages defect (downgraded — root cause found by direct code reading, a known platform gap, reclassified to the builder route), and the confirmed pilot: dartsonline published a Guides nav link and a blank /guides/index.html while gamesdesign has working guides — a broken route, not a missing feature, with a standing hypothesis (reconcile_site_plan's routing table omits "guide"; nav derives from the plan, not the built set), mandatory pre-check queries and a cross-site differential as evidence. Establishes "genuinely mechanism-unclear" as the admission bar.
- **sources:** HANDOFF_fixloop_thread(8).md; HANDOFF_fixloop_thread(3)-(5).md deltas (the triage history)
- **relations:** eval gate; site-plan reconciler routing table (the suspected mechanism)
- **verify-later:** reconcile_site_plan routing table in load_work_item_actions.go; the F0 PILOT section of the fixloop runbook (U14)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Seeded-bug strategy for the first end-to-end write run
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** README_02 §6 and README_overview both recommend it; "the system's first-ever PR will have earned every gate it passed" — proposed, awaiting deploys.
- **what:** Because the only real test bug never yields an approvable plan (correctly escalated as architecture-level), the write step is tested by planting a contained single-file defect with an obvious symptom on a low-stakes surface and running the full pipeline — diagnose → plan → council (genuine approval) → implementer → PR. Rejected alternatives: hand-approving a known-flawed plan (contradicts the reviewers), waiting for an organic small bug (unbounded).
- **sources:** README_02_evidence_backed_proposals.md#6; README_overview.md
- **relations:** fix-implementer; eval gate
- **verify-later:** whether the first PR happened (git history for fix/* branches)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Transferable machinery: legacy-migration and feature intakes
- **category:** fix-loop
- **status-signal:** aspirational
- **status-evidence:** README_02 §3: migration agents "not built, but it's the same machinery with a different intake"; features from specs "honestly furthest away … plausible; not designed".
- **what:** The allowlist/gate/council scaffolding is intake-agnostic: a legacy migration is "pattern X supersedes pattern Y" (scanner finds Y-shaped code, proposer writes constrained plans, council reviews, PRs flow); feature-building from mission docs needs a new grounding tier ("cite the spec clause this serves") — same shape as causal citation but not designed.
- **sources:** README_02_evidence_backed_proposals.md#3
- **relations:** council pattern; hard gates
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Confirm-not-initiate + the single central confirmer (one path to active)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_contract_set_review §2.1 "Resolution applied: a single central confirmer"; all contracts remain "contract specification" status.
- **what:** Agents propose (status=proposed rows + a work item); a human confirms; ONE component applies uniformly — flip to active, set last_verified_at/verified_by, deprecate the prior version, write the decision-log entry, emit the in-band change event — so confirm-not-initiate is a status-transition rule enforced in one auditable place, not a discipline reimplemented per agent. Hardening from the edge-case pass: the apply is one DB transaction with the change event in an outbox (crash-consistent, retry-safe), idempotent (re-applying an active version is a no-op), one live proposal per target extending down to layer rows (a new proposal replaces the proposed row; expiring a work item deprecates its row), and work items reference proposed rows by identity not pinned version.
- **sources:** FOCUS_contract_set_review.md#2.1,#2.3; FOCUS_pre_build_edge_cases(1).md#1.2,#1.3,#12; PLAN_config_work_items_contract(3).md#4
- **relations:** config_work_items; decision log; change layer in_band guard
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### config_work_items contract (mirror of site_work_items, tenant-scoped)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** PLAN_config_work_items_contract(3) "contract specification"; corrected against the real site_work_items shape (FOCUS_schema_verification_findings §2).
- **what:** The shared queue every onboarding/maintenance agent emits into and the tenant reads: a parallel table (site_id NOT NULL blocks direct reuse) mirroring the verified shape — item_type/spec/result naming, integer priority, the real status lifecycle (detected→triaged→approved|rejected→claimed→complete|failed), reuse of approval_mode (the pre-existing confirm-not-initiate field; config defaults manual, 'auto' only for graduated capabilities), item_key unique-partial dedup (one live item per target), depends_on/parent_item_id, retry machinery. Batch confirmation for the initial onboarding flood (approval granularity adapts; apply still honours dependency order). Explicit scope: gates config, not deliverables.
- **sources:** PLAN_config_work_items_contract(3).md; FOCUS_schema_verification_findings.md#2; FOCUS_pre_build_edge_cases(1).md#2.1,#15
- **relations:** central confirmer; two-gated-paths; site_work_items (the reuse source)
- **verify-later:** table existence; site_work_items approval_mode semantics in code

<!-- SOURCE: U16_docs019_design_plans.md -->
### Decision log (immutable; premise vs rule_trace; inputs_used)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** PLAN_decision_log_contract(2) "contract specification".
- **what:** The published-reasoning log given shape: append-only, one row per decision, carrying either a human-readable premise (judgement decisions) or a rule_trace (mechanical ones — exactly one of the two, so mechanical steps don't produce noise premises), plus inputs_used: the active-config slice in hand at decision time (compact atom id+version references + merged-view hashes by default; full snapshot inline for high-stakes kinds). Resolves freshness-vs-retrospect: compute on read for freshness, log at point of use for reconstruction. Write discipline: every decision logs, the entry precedes the apply, logging is not itself a logged decision. Read patterns: drift detection (premise vs current profile), heuristic invalidation, trust-ledger evidence, retrospective audit, compliance review. Open seam flagged: bundle assemblies would dominate a reasoning log by volume — bundle provenance may belong as the consuming decision's inputs_used instead.
- **sources:** PLAN_decision_log_contract(2).md; FOCUS_pre_build_edge_cases(1).md#4.4,#11; FOCUS_whole_plan_review.md#2.2
- **relations:** bundle provenance; trust ledger; work-item resolutions feed premises
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Trust ledger + bidirectional ratchet (asymmetric by design)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** PLAN_trust_ledger_contract(4) "contract specification".
- **what:** Mutable state, one row per (tenant, capability): trust_level ∈ confirm_every/confirm_exceptions/notify/autonomous plus derived gate_policy and evidence_summary; derived from but separate from the immutable decision log (different access patterns). Cold start is always confirm_every (trust earned per tenant; no cross-tenant inheritance — deferred deliberately). Graduation up is always confirm-not-initiate; de-graduation down may auto-apply with notification on severe evidence — losing trust shouldn't wait on a human, gaining it should — but de-graduation evidence must first pass the defect-vs-partition filter so a flaky test or infra blip can't drop a capability and trigger a confirmation flood. Cascade routers and gate policy engines read it at runtime.
- **sources:** PLAN_trust_ledger_contract(4).md; FOCUS_pre_build_edge_cases(1).md#2.2; FOCUS_whole_plan_review.md#2.1
- **relations:** capabilities catalog (the ceiling); maintenance agent (the evidence source); outcome-record gap
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Capabilities catalog: the ceiling lives on the capability (blast radius caps trust)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** PLAN_capabilities_catalog_contract(1) "contract specification. The sixth contract, closing the trust ledger's open dependency."
- **what:** A sibling table to agent_definitions (its existing capabilities jsonb holds free descriptive kebab tags for discovery, deliberately left alone; catalog capability_ids are snake_case dispatch keys — a recorded naming decision) holding per-capability ceiling, verifiability and containment. The ceiling is a judgement over the two factors (the weaker holds it); stored for cheap reads but the factors are authoritative — a factor change triggers a gated ceiling re-proposal. Capabilities aren't 1:1 with agents; the operation→capability mapping is declared at the action level. Seeding principle made explicit: the more a capability can break — especially chassis-editing ones — the lower its ceiling, regardless of verifiability; never fully autonomous for chassis-touching capabilities.
- **sources:** PLAN_capabilities_catalog_contract(1).md; FOCUS_pre_build_edge_cases(1).md#13; FOCUS_whole_plan_review.md#1.4
- **relations:** trust ledger; recursive self-improvement risk; cascade router
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Change-layer integration (change_events; in_band closes the self-modification loop)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** PLAN_change_layer_integration_contract(4) "contract specification. … Closes the final contract gap before implementation."
- **what:** Change events from four sources (git_webhook, polling, in_band, periodic_sweep, + manual) land as first-class change_events rows (at-least-once with commit dedup; no event silently dropped — triggers_fired=[] is an explicit record). The trigger filter mapping changed paths to typed maintenance triggers is computed from the mechanical config, not stored (compute-on-read applied to routing). in_band emission — the tool's own applies emit events — is what keeps self-modification visible to the drift detector and decision log; rule: state changes emit, computed-view refreshes don't. Guard: a confirmer apply doesn't re-trigger maintenance on the entry just confirmed, but genuine downstream effects (audit code against a newly-active convention) still fire, and generation-origin in_band events are never exempt. reuse_index_refresh is its own trigger because a stale reuse index fails silently.
- **sources:** PLAN_change_layer_integration_contract(4).md; FOCUS_whole_plan_review.md#1.2; FOCUS_pre_build_edge_cases(1).md#4.1,#2.4
- **relations:** config-maintenance agent; central confirmer; reuse-search freshness
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Two gated paths: config changes vs deliverables
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_pre_build_edge_cases §3 "a conceptual conflation to fix"; restated in the trust-ledger and work-items contracts.
- **what:** Changes to the tool's knowledge of the codebase (standards/objectives/mechanical) flow config_work_items → central confirmer → active config. The tool's outputs (generated code; edits to workflows/agent definitions — deliverables even though they're DB rows) flow cascade → trust-ledger gate → apply+commit+in_band event. The decision log spans both; the gates are not the same gate, and there are correspondingly two gated-mutation mechanisms (config confirmer; ledger ratchet-evaluator with asymmetric de-graduation).
- **sources:** FOCUS_pre_build_edge_cases(1).md#3; PLAN_trust_ledger_contract(4).md#1; PLAN_config_work_items_contract(3).md#5; FOCUS_whole_plan_review.md#2.1
- **relations:** trust ledger; config_work_items
- **verify-later:** n/a (design)

<!-- SOURCE: U16_docs019_design_plans.md -->
### The outcome-record gap (the loop runs on outcomes nobody sources)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_pre_build_edge_cases §9 "real gap … You can log every input and decision and still have no feedback signal."
- **what:** The contracts log decisions and inputs, but nothing records whether a deliverable succeeded — verification pass/fail, reverted, human-corrected, accepted-as-is — the raw signal evidence_summary must aggregate for the ratchet to move. Companion gap (§10): "the bundle helped" has no defined metric (candidate signals: fewer correction rounds, fewer convention violations, less manual context-gathering); both needed before Phase 2.
- **sources:** FOCUS_pre_build_edge_cases(1).md#9,#10
- **relations:** trust ledger; thin-slice premise test
- **verify-later:** n/a (gap)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Thin vertical slice before the six-contract infrastructure
- **category:** NEW:autonomy-governance
- **status-signal:** deployed
- **status-evidence:** FOCUS_pre_build_edge_cases §8 recommended it; the tool plan's status note shows it happened ("a thin slice of Phase 1 is built … deliberately ahead of [the contracts] to test the core thesis first").
- **what:** The whole design rests on one unproven premise — an assembled bundle beats paste-and-rot — and none of the six contracts test it. So: hand-write a minimal flat-file constitution, build analyser+schema extractor, assemble ONE bundle for ONE real task, paste it by hand; only if it visibly helps build the infrastructure. This sequencing was followed: the thin-slice harness shipped and was used on real bugs while the contracts stayed specifications.
- **sources:** FOCUS_pre_build_edge_cases(1).md#8,#16; PLAN_context_assembly_tool_and_service(2).md status
- **relations:** contextkit harness; the six contracts (their build deliberately deferred)
- **verify-later:** n/a (executed strategy)

<!-- SOURCE: U16_docs019_design_plans.md -->
### External rollback (the self-hosting trap) + recursive self-improvement as residual risk
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_pre_build_edge_cases §5, §14 — stated rules/risks, no implementation claimed.
- **what:** The tool runs on the chassis it modifies; a bad change could break the chassis badly enough that the tool can't run to fix it. Rule: rollback to known-good must be runnable externally, with no dependency on the agents/orchestrator being rolled back. And a self-improvement that passes verification can still degrade the tool's judgement gradually — not fully solvable; managed by conservative early trust, the human gate, external rollback and low ceilings for chassis-touching capabilities, and named as an accepted residual risk rather than assumed closed.
- **sources:** FOCUS_pre_build_edge_cases(1).md#5,#13,#14
- **relations:** capabilities-catalog ceilings; human gate; fork isolation
- **verify-later:** existence of any external rollback path

<!-- SOURCE: U16_docs019_design_plans.md -->
### Morality review as a configured, layered standard (not a baked-in view)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** MAPPING_tool_to_actions_and_agents(2) — design discussion; review contributors "(none yet)" in the thin-slice column.
- **what:** Distinct from liability ("will this get us sued" vs "is this right"): a build-time review contributor applying a layered standard held in the active config — an operator-chosen recognised base source (ASA/CAP Code, CMA guidance; OECD/UNESCO/NIST for the AI angle), operator values layered above it, jurisdiction/current-focus overlays later. Two altitudes: per-output, and a vertical-level gate at intake (should we build this site/industry at all). Contested cases route to HITL; the tool applies the configured standard and flags — it is not the moral authority.
- **sources:** MAPPING_tool_to_actions_and_agents(2).md (morality review section)
- **relations:** build-time review contributors; active-config standards layer; council compliance eye
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Contributors vs checkers (build-path reviews ≠ improvement-loop monitors)
- **category:** NEW:autonomy-governance
- **status-signal:** deployed
- **status-evidence:** MAPPING(2): "Checkers are a different concept — not these" — a settled terminology/architecture distinction (reuse overlap flagged to investigate).
- **what:** Context contributors assemble bundle slices (code/data/runtime/standards); build-time review contributors (reuse, near-duplicate, liability, morality, correctness) review a PROPOSED change before it ships, raising concerns that revise or HITL-gate; improvement-loop checkers (the check_*.go family) continuously monitor DEPLOYED sites against plan/spec in the operate layer. Two layers restated: the website-builder builds sites; the context tool builds reliable changes to the builder.
- **sources:** MAPPING_tool_to_actions_and_agents(2).md; PLAN_workflows_and_actions_migration(19).md (group-agent reviews)
- **relations:** council pattern (the reviews' fix-loop descendant); improvement-loop category
- **verify-later:** whether build-time reviews reuse check_*.go logic (flagged open)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Confirm-not-initiate + the single central confirmer (one path to active)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_contract_set_review §2.1 "Resolution applied: a single central confirmer"; all contracts remain "contract specification" status.
- **what:** Agents propose (status=proposed rows + a work item); a human confirms; ONE component applies uniformly — flip to active, set last_verified_at/verified_by, deprecate the prior version, write the decision-log entry, emit the in-band change event — so confirm-not-initiate is a status-transition rule enforced in one auditable place, not a discipline reimplemented per agent. Hardening from the edge-case pass: the apply is one DB transaction with the change event in an outbox (crash-consistent, retry-safe), idempotent (re-applying an active version is a no-op), one live proposal per target extending down to layer rows (a new proposal replaces the proposed row; expiring a work item deprecates its row), and work items reference proposed rows by identity not pinned version.
- **sources:** FOCUS_contract_set_review.md#2.1,#2.3; FOCUS_pre_build_edge_cases(1).md#1.2,#1.3,#12; PLAN_config_work_items_contract(3).md#4
- **relations:** config_work_items; decision log; change layer in_band guard
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### config_work_items contract (mirror of site_work_items, tenant-scoped)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** PLAN_config_work_items_contract(3) "contract specification"; corrected against the real site_work_items shape (FOCUS_schema_verification_findings §2).
- **what:** The shared queue every onboarding/maintenance agent emits into and the tenant reads: a parallel table (site_id NOT NULL blocks direct reuse) mirroring the verified shape — item_type/spec/result naming, integer priority, the real status lifecycle (detected→triaged→approved|rejected→claimed→complete|failed), reuse of approval_mode (the pre-existing confirm-not-initiate field; config defaults manual, 'auto' only for graduated capabilities), item_key unique-partial dedup (one live item per target), depends_on/parent_item_id, retry machinery. Batch confirmation for the initial onboarding flood (approval granularity adapts; apply still honours dependency order). Explicit scope: gates config, not deliverables.
- **sources:** PLAN_config_work_items_contract(3).md; FOCUS_schema_verification_findings.md#2; FOCUS_pre_build_edge_cases(1).md#2.1,#15
- **relations:** central confirmer; two-gated-paths; site_work_items (the reuse source)
- **verify-later:** table existence; site_work_items approval_mode semantics in code

<!-- SOURCE: U16_docs019_design_plans.md -->
### Decision log (immutable; premise vs rule_trace; inputs_used)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** PLAN_decision_log_contract(2) "contract specification".
- **what:** The published-reasoning log given shape: append-only, one row per decision, carrying either a human-readable premise (judgement decisions) or a rule_trace (mechanical ones — exactly one of the two, so mechanical steps don't produce noise premises), plus inputs_used: the active-config slice in hand at decision time (compact atom id+version references + merged-view hashes by default; full snapshot inline for high-stakes kinds). Resolves freshness-vs-retrospect: compute on read for freshness, log at point of use for reconstruction. Write discipline: every decision logs, the entry precedes the apply, logging is not itself a logged decision. Read patterns: drift detection (premise vs current profile), heuristic invalidation, trust-ledger evidence, retrospective audit, compliance review. Open seam flagged: bundle assemblies would dominate a reasoning log by volume — bundle provenance may belong as the consuming decision's inputs_used instead.
- **sources:** PLAN_decision_log_contract(2).md; FOCUS_pre_build_edge_cases(1).md#4.4,#11; FOCUS_whole_plan_review.md#2.2
- **relations:** bundle provenance; trust ledger; work-item resolutions feed premises
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Trust ledger + bidirectional ratchet (asymmetric by design)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** PLAN_trust_ledger_contract(4) "contract specification".
- **what:** Mutable state, one row per (tenant, capability): trust_level ∈ confirm_every/confirm_exceptions/notify/autonomous plus derived gate_policy and evidence_summary; derived from but separate from the immutable decision log (different access patterns). Cold start is always confirm_every (trust earned per tenant; no cross-tenant inheritance — deferred deliberately). Graduation up is always confirm-not-initiate; de-graduation down may auto-apply with notification on severe evidence — losing trust shouldn't wait on a human, gaining it should — but de-graduation evidence must first pass the defect-vs-partition filter so a flaky test or infra blip can't drop a capability and trigger a confirmation flood. Cascade routers and gate policy engines read it at runtime.
- **sources:** PLAN_trust_ledger_contract(4).md; FOCUS_pre_build_edge_cases(1).md#2.2; FOCUS_whole_plan_review.md#2.1
- **relations:** capabilities catalog (the ceiling); maintenance agent (the evidence source); outcome-record gap
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Capabilities catalog: the ceiling lives on the capability (blast radius caps trust)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** PLAN_capabilities_catalog_contract(1) "contract specification. The sixth contract, closing the trust ledger's open dependency."
- **what:** A sibling table to agent_definitions (its existing capabilities jsonb holds free descriptive kebab tags for discovery, deliberately left alone; catalog capability_ids are snake_case dispatch keys — a recorded naming decision) holding per-capability ceiling, verifiability and containment. The ceiling is a judgement over the two factors (the weaker holds it); stored for cheap reads but the factors are authoritative — a factor change triggers a gated ceiling re-proposal. Capabilities aren't 1:1 with agents; the operation→capability mapping is declared at the action level. Seeding principle made explicit: the more a capability can break — especially chassis-editing ones — the lower its ceiling, regardless of verifiability; never fully autonomous for chassis-touching capabilities.
- **sources:** PLAN_capabilities_catalog_contract(1).md; FOCUS_pre_build_edge_cases(1).md#13; FOCUS_whole_plan_review.md#1.4
- **relations:** trust ledger; recursive self-improvement risk; cascade router
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Change-layer integration (change_events; in_band closes the self-modification loop)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** PLAN_change_layer_integration_contract(4) "contract specification. … Closes the final contract gap before implementation."
- **what:** Change events from four sources (git_webhook, polling, in_band, periodic_sweep, + manual) land as first-class change_events rows (at-least-once with commit dedup; no event silently dropped — triggers_fired=[] is an explicit record). The trigger filter mapping changed paths to typed maintenance triggers is computed from the mechanical config, not stored (compute-on-read applied to routing). in_band emission — the tool's own applies emit events — is what keeps self-modification visible to the drift detector and decision log; rule: state changes emit, computed-view refreshes don't. Guard: a confirmer apply doesn't re-trigger maintenance on the entry just confirmed, but genuine downstream effects (audit code against a newly-active convention) still fire, and generation-origin in_band events are never exempt. reuse_index_refresh is its own trigger because a stale reuse index fails silently.
- **sources:** PLAN_change_layer_integration_contract(4).md; FOCUS_whole_plan_review.md#1.2; FOCUS_pre_build_edge_cases(1).md#4.1,#2.4
- **relations:** config-maintenance agent; central confirmer; reuse-search freshness
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Two gated paths: config changes vs deliverables
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_pre_build_edge_cases §3 "a conceptual conflation to fix"; restated in the trust-ledger and work-items contracts.
- **what:** Changes to the tool's knowledge of the codebase (standards/objectives/mechanical) flow config_work_items → central confirmer → active config. The tool's outputs (generated code; edits to workflows/agent definitions — deliverables even though they're DB rows) flow cascade → trust-ledger gate → apply+commit+in_band event. The decision log spans both; the gates are not the same gate, and there are correspondingly two gated-mutation mechanisms (config confirmer; ledger ratchet-evaluator with asymmetric de-graduation).
- **sources:** FOCUS_pre_build_edge_cases(1).md#3; PLAN_trust_ledger_contract(4).md#1; PLAN_config_work_items_contract(3).md#5; FOCUS_whole_plan_review.md#2.1
- **relations:** trust ledger; config_work_items
- **verify-later:** n/a (design)

<!-- SOURCE: U16_docs019_design_plans.md -->
### The outcome-record gap (the loop runs on outcomes nobody sources)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_pre_build_edge_cases §9 "real gap … You can log every input and decision and still have no feedback signal."
- **what:** The contracts log decisions and inputs, but nothing records whether a deliverable succeeded — verification pass/fail, reverted, human-corrected, accepted-as-is — the raw signal evidence_summary must aggregate for the ratchet to move. Companion gap (§10): "the bundle helped" has no defined metric (candidate signals: fewer correction rounds, fewer convention violations, less manual context-gathering); both needed before Phase 2.
- **sources:** FOCUS_pre_build_edge_cases(1).md#9,#10
- **relations:** trust ledger; thin-slice premise test
- **verify-later:** n/a (gap)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Thin vertical slice before the six-contract infrastructure
- **category:** NEW:autonomy-governance
- **status-signal:** deployed
- **status-evidence:** FOCUS_pre_build_edge_cases §8 recommended it; the tool plan's status note shows it happened ("a thin slice of Phase 1 is built … deliberately ahead of [the contracts] to test the core thesis first").
- **what:** The whole design rests on one unproven premise — an assembled bundle beats paste-and-rot — and none of the six contracts test it. So: hand-write a minimal flat-file constitution, build analyser+schema extractor, assemble ONE bundle for ONE real task, paste it by hand; only if it visibly helps build the infrastructure. This sequencing was followed: the thin-slice harness shipped and was used on real bugs while the contracts stayed specifications.
- **sources:** FOCUS_pre_build_edge_cases(1).md#8,#16; PLAN_context_assembly_tool_and_service(2).md status
- **relations:** contextkit harness; the six contracts (their build deliberately deferred)
- **verify-later:** n/a (executed strategy)

<!-- SOURCE: U16_docs019_design_plans.md -->
### External rollback (the self-hosting trap) + recursive self-improvement as residual risk
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_pre_build_edge_cases §5, §14 — stated rules/risks, no implementation claimed.
- **what:** The tool runs on the chassis it modifies; a bad change could break the chassis badly enough that the tool can't run to fix it. Rule: rollback to known-good must be runnable externally, with no dependency on the agents/orchestrator being rolled back. And a self-improvement that passes verification can still degrade the tool's judgement gradually — not fully solvable; managed by conservative early trust, the human gate, external rollback and low ceilings for chassis-touching capabilities, and named as an accepted residual risk rather than assumed closed.
- **sources:** FOCUS_pre_build_edge_cases(1).md#5,#13,#14
- **relations:** capabilities-catalog ceilings; human gate; fork isolation
- **verify-later:** existence of any external rollback path

<!-- SOURCE: U16_docs019_design_plans.md -->
### Morality review as a configured, layered standard (not a baked-in view)
- **category:** NEW:autonomy-governance
- **status-signal:** aspirational
- **status-evidence:** MAPPING_tool_to_actions_and_agents(2) — design discussion; review contributors "(none yet)" in the thin-slice column.
- **what:** Distinct from liability ("will this get us sued" vs "is this right"): a build-time review contributor applying a layered standard held in the active config — an operator-chosen recognised base source (ASA/CAP Code, CMA guidance; OECD/UNESCO/NIST for the AI angle), operator values layered above it, jurisdiction/current-focus overlays later. Two altitudes: per-output, and a vertical-level gate at intake (should we build this site/industry at all). Contested cases route to HITL; the tool applies the configured standard and flags — it is not the moral authority.
- **sources:** MAPPING_tool_to_actions_and_agents(2).md (morality review section)
- **relations:** build-time review contributors; active-config standards layer; council compliance eye
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Contributors vs checkers (build-path reviews ≠ improvement-loop monitors)
- **category:** NEW:autonomy-governance
- **status-signal:** deployed
- **status-evidence:** MAPPING(2): "Checkers are a different concept — not these" — a settled terminology/architecture distinction (reuse overlap flagged to investigate).
- **what:** Context contributors assemble bundle slices (code/data/runtime/standards); build-time review contributors (reuse, near-duplicate, liability, morality, correctness) review a PROPOSED change before it ships, raising concerns that revise or HITL-gate; improvement-loop checkers (the check_*.go family) continuously monitor DEPLOYED sites against plan/spec in the operate layer. Two layers restated: the website-builder builds sites; the context tool builds reliable changes to the builder.
- **sources:** MAPPING_tool_to_actions_and_agents(2).md; PLAN_workflows_and_actions_migration(19).md (group-agent reviews)
- **relations:** council pattern (the reviews' fix-loop descendant); improvement-loop category
- **verify-later:** whether build-time reviews reuse check_*.go logic (flagged open)

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### trust ledger (bidirectional trust ratchet, per-tenant per-capability)
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** "Status: contract specification... Now given concrete shape." — a design document, no implementation claimed. A later/fuller version of this exact contract exists live at docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/PLAN_trust_ledger_contract(4).md (outside this unit's scope).
- **what:** A `trust_ledger` table (one row per tenant×capability) holding a mutable `trust_level` (`confirm_every`/`confirm_exceptions`/`notify`/`autonomous`) that cascade routers read to floor the production tier and gate-policy engines read to decide autonomy — derived from, but distinct from, the immutable append-only `decision_log`. The capability's **ceiling** (max reachable trust, set by verifiability × containment) lives on a separate capability catalog, not the ledger row, so it's a property of the capability, not the tenant. Mutation is asymmetric: graduation (trust up) is always confirm-not-initiate via a `config_work_items` proposal; de-graduation (trust down) may auto-apply with notification on severe evidence — "losing trust is reversible; falsely gaining trust is what allows mistakes to apply unsupervised."
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/PLAN_trust_ledger_contract(1).md
- **relations:** change-layer integration contract (below, feeds the ledger's evidence_summary); governance/HITL principles (below); the fuller live PLAN_trust_ledger_contract(4).md (not in this unit's scope)
- **verify-later:** any `trust_ledger` or `capabilities` table in the live schema

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### change-layer integration contract (change_events, trigger filter, in-band emission)
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** "Status: contract specification... Closes the final contract gap before implementation." A fuller live version exists at docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/PLAN_change_layer_integration_contract(4).md (outside this unit's scope).
- **what:** Defines how code/doc diffs reach a maintenance agent: a first-class `change_events` table (source ∈ git_webhook/polling/in_band/periodic_sweep/manual, at-least-once with commit_id dedup) feeding a **trigger filter computed from the mechanical config** (not stored — so it self-updates when doc/code paths move) that fans out into typed triggers (`conventions_reextraction`, `schema_check`, `code_audit_refresh`, `reuse_index_refresh`, `intent_revalidation`, `freshness_check`). The `in_band` source is the mechanism that "closes the loop on self-modification" — when the tool's own bundle-builder or a layer agent applies a confirmed change, it emits its own change event so the drift detector doesn't go blind to its own effects; a scoped guard prevents a just-confirmed entry from re-triggering on itself while still letting genuine downstream effects (e.g. auditing existing code against a newly confirmed convention) fire.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/PLAN_change_layer_integration_contract(3).md (family-latest in this unit), PLAN_change_layer_integration_contract(1).md (delta-checked, no drops)
- **relations:** trust ledger (above); reuse-check retrieval pipeline (below, reuse_index_refresh trigger)
- **verify-later:** any `change_events` table in the live schema

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### context substrate model (authored vs derived, salience over presence)
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** Framed explicitly as "a list of synthesis points... captured as reusable lenses for further design work, not as a final theory."
- **what:** A framing for how LLM context should be built: documentation/standards/intent are operational inputs to generation, not passive reference, and split into two epistemic categories — **authored** (has an owner and lifecycle, can be wrong, needs maintenance) vs **derived** (no-owner true-right-now readout, can only be current or superseded; source code sits on this line). The **change layer** (diffs) is derived-but-narrative — the natural audit/learning surface. Authored layers should hold **references, not copies** of derived material so they don't drift when reality moves. Two staleness modes need two different fixes: authored drift is fixed by keeping authored content thin and pointer-rich; derived snapshot-staleness is fixed by fetching at reasoning time, not paste-time. LLMs lose the big picture from **salience, not window size** — local detail crowds out context mid-reasoning even when the text is still "in the window."
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/NOTES_running_synthesis_principles(20).md §Context substrate
- **relations:** contextkit toolchain (above); flat-file constitution (above)
- **verify-later:** none (a design framing, not a built artifact)

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### mediator model for competing design concerns ("right" as requirement-relative balance)
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** Same running-synthesis document, framed as an unbuilt lens.
- **what:** A model for resolving conflicting design dimensions (fast/secure/generic/simple/functional) as a requirement-relative balance rather than a pick-one-winner or naive-merge: authored solutions are treated as extremes that bound the solution space, and a mediator finds the point inside it that the requirement's priority profile dictates (ordered priority, not numeric weights, since real-world priority shifts arrive as "X now outranks Y"). A satisfied concern demotes from active author to passive checker (re-promoting if a later change breaks it) — unifying "checker" and "multi-author" as two modes of one process. Non-convergence among concerns is treated as the genuine escalation signal, isolating the one real tradeoff that needs human judgement from everything else that settles on its own. Multi-author surfaces tradeoffs vividly but cannot resolve value-laden conflicts — it's an option-generation engine, not a decision engine.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/NOTES_running_synthesis_principles(20).md §"Right" as balance, not a single answer
- **relations:** governance/HITL principles (below)
- **verify-later:** none (a design framing, not a built artifact)

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### governance/HITL principles (confirm-not-initiate, decision publishing, sealed inheritance)
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** Same running-synthesis document; principles, not shipped mechanisms.
- **what:** A cluster of governance rules for an autonomous build-and-operate system: **confirm-not-initiate** (agent-led reasoning, human confirms via a decision package, never authors from scratch); **every decision publishes its reasoning**, since drift detection is only possible because premises are logged and can be compared to the current premise; **two precedence directions in inheritance** — normal entries are child-wins (local refinement) but sealed constraints are ancestor-wins (legal floors, mission non-negotiables), so a leaf can't defeat a new law by prior relaxation; **three resolutions to a doc/code disagreement** (code drifted / doc drifted / legitimate exception) with a configurable default presumption that the human can always override; **one path to a privileged state transition** (e.g. `proposed → active`) routed through a single central confirmer rather than reimplemented per producer, so confirm-not-initiate is airtight in one place; **newer supersedes pending** — a fresh proposal for an already-pending target expires the older one rather than blocking on staleness.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/NOTES_running_synthesis_principles(20).md §Governance and HITL
- **relations:** trust ledger (above); change-layer integration contract (above); mediator model (above)
- **verify-later:** none (a design framing, not a built artifact)

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### autonomous-system building-block hardening checklist
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** Framed as "edge cases caught before building" in the running-synthesis notes — a pre-implementation checklist, not a verified implementation.
- **what:** A catalog of structural safety patterns for building the autonomous-operate machinery itself, distilled from design review: self-referential structures (trees, version chains) need a cycle guard on write plus a detect-and-fail walk; a multi-step apply (writes + an emitted event) must be all-or-nothing via one transaction with an outbox, or a mid-crash leaves a live row with no log/event; assembling from several tables needs one consistent point-in-time snapshot; "at most one live X per target" must be enforced at every layer down to the underlying row, not just the queue; bulk operations need bulk confirmation (per-item confirm doesn't scale to an onboarding flood); transient/infrastructure failures must be filtered out before they're allowed to lower a capability's trust; derived indexes/caches go stale silently and need a freshness stamp; recovery must not depend on the thing being recovered (the rollback path can't route through the agents it's rolling back); blast radius caps the trust ceiling regardless of verifiability, because self-modification is a residual risk that's managed (conservative early trust, human-in-the-loop, external rollback), not solved.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/NOTES_running_synthesis_principles(20).md §Building discipline
- **relations:** trust ledger (above); change-layer integration contract (above)
- **verify-later:** none (a design checklist, not a built artifact)

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### trust ledger (bidirectional trust ratchet, per-tenant per-capability)
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** "Status: contract specification... Now given concrete shape." — a design document, no implementation claimed. A later/fuller version of this exact contract exists live at docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/PLAN_trust_ledger_contract(4).md (outside this unit's scope).
- **what:** A `trust_ledger` table (one row per tenant×capability) holding a mutable `trust_level` (`confirm_every`/`confirm_exceptions`/`notify`/`autonomous`) that cascade routers read to floor the production tier and gate-policy engines read to decide autonomy — derived from, but distinct from, the immutable append-only `decision_log`. The capability's **ceiling** (max reachable trust, set by verifiability × containment) lives on a separate capability catalog, not the ledger row, so it's a property of the capability, not the tenant. Mutation is asymmetric: graduation (trust up) is always confirm-not-initiate via a `config_work_items` proposal; de-graduation (trust down) may auto-apply with notification on severe evidence — "losing trust is reversible; falsely gaining trust is what allows mistakes to apply unsupervised."
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/PLAN_trust_ledger_contract(1).md
- **relations:** change-layer integration contract (below, feeds the ledger's evidence_summary); governance/HITL principles (below); the fuller live PLAN_trust_ledger_contract(4).md (not in this unit's scope)
- **verify-later:** any `trust_ledger` or `capabilities` table in the live schema

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### change-layer integration contract (change_events, trigger filter, in-band emission)
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** "Status: contract specification... Closes the final contract gap before implementation." A fuller live version exists at docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/PLAN_change_layer_integration_contract(4).md (outside this unit's scope).
- **what:** Defines how code/doc diffs reach a maintenance agent: a first-class `change_events` table (source ∈ git_webhook/polling/in_band/periodic_sweep/manual, at-least-once with commit_id dedup) feeding a **trigger filter computed from the mechanical config** (not stored — so it self-updates when doc/code paths move) that fans out into typed triggers (`conventions_reextraction`, `schema_check`, `code_audit_refresh`, `reuse_index_refresh`, `intent_revalidation`, `freshness_check`). The `in_band` source is the mechanism that "closes the loop on self-modification" — when the tool's own bundle-builder or a layer agent applies a confirmed change, it emits its own change event so the drift detector doesn't go blind to its own effects; a scoped guard prevents a just-confirmed entry from re-triggering on itself while still letting genuine downstream effects (e.g. auditing existing code against a newly confirmed convention) fire.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/PLAN_change_layer_integration_contract(3).md (family-latest in this unit), PLAN_change_layer_integration_contract(1).md (delta-checked, no drops)
- **relations:** trust ledger (above); reuse-check retrieval pipeline (below, reuse_index_refresh trigger)
- **verify-later:** any `change_events` table in the live schema

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### context substrate model (authored vs derived, salience over presence)
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** Framed explicitly as "a list of synthesis points... captured as reusable lenses for further design work, not as a final theory."
- **what:** A framing for how LLM context should be built: documentation/standards/intent are operational inputs to generation, not passive reference, and split into two epistemic categories — **authored** (has an owner and lifecycle, can be wrong, needs maintenance) vs **derived** (no-owner true-right-now readout, can only be current or superseded; source code sits on this line). The **change layer** (diffs) is derived-but-narrative — the natural audit/learning surface. Authored layers should hold **references, not copies** of derived material so they don't drift when reality moves. Two staleness modes need two different fixes: authored drift is fixed by keeping authored content thin and pointer-rich; derived snapshot-staleness is fixed by fetching at reasoning time, not paste-time. LLMs lose the big picture from **salience, not window size** — local detail crowds out context mid-reasoning even when the text is still "in the window."
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/NOTES_running_synthesis_principles(20).md §Context substrate
- **relations:** contextkit toolchain (above); flat-file constitution (above)
- **verify-later:** none (a design framing, not a built artifact)

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### mediator model for competing design concerns ("right" as requirement-relative balance)
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** Same running-synthesis document, framed as an unbuilt lens.
- **what:** A model for resolving conflicting design dimensions (fast/secure/generic/simple/functional) as a requirement-relative balance rather than a pick-one-winner or naive-merge: authored solutions are treated as extremes that bound the solution space, and a mediator finds the point inside it that the requirement's priority profile dictates (ordered priority, not numeric weights, since real-world priority shifts arrive as "X now outranks Y"). A satisfied concern demotes from active author to passive checker (re-promoting if a later change breaks it) — unifying "checker" and "multi-author" as two modes of one process. Non-convergence among concerns is treated as the genuine escalation signal, isolating the one real tradeoff that needs human judgement from everything else that settles on its own. Multi-author surfaces tradeoffs vividly but cannot resolve value-laden conflicts — it's an option-generation engine, not a decision engine.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/NOTES_running_synthesis_principles(20).md §"Right" as balance, not a single answer
- **relations:** governance/HITL principles (below)
- **verify-later:** none (a design framing, not a built artifact)

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### governance/HITL principles (confirm-not-initiate, decision publishing, sealed inheritance)
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** Same running-synthesis document; principles, not shipped mechanisms.
- **what:** A cluster of governance rules for an autonomous build-and-operate system: **confirm-not-initiate** (agent-led reasoning, human confirms via a decision package, never authors from scratch); **every decision publishes its reasoning**, since drift detection is only possible because premises are logged and can be compared to the current premise; **two precedence directions in inheritance** — normal entries are child-wins (local refinement) but sealed constraints are ancestor-wins (legal floors, mission non-negotiables), so a leaf can't defeat a new law by prior relaxation; **three resolutions to a doc/code disagreement** (code drifted / doc drifted / legitimate exception) with a configurable default presumption that the human can always override; **one path to a privileged state transition** (e.g. `proposed → active`) routed through a single central confirmer rather than reimplemented per producer, so confirm-not-initiate is airtight in one place; **newer supersedes pending** — a fresh proposal for an already-pending target expires the older one rather than blocking on staleness.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/NOTES_running_synthesis_principles(20).md §Governance and HITL
- **relations:** trust ledger (above); change-layer integration contract (above); mediator model (above)
- **verify-later:** none (a design framing, not a built artifact)

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### autonomous-system building-block hardening checklist
- **category:** NEW:autonomous-build-operate
- **status-signal:** aspirational
- **status-evidence:** Framed as "edge cases caught before building" in the running-synthesis notes — a pre-implementation checklist, not a verified implementation.
- **what:** A catalog of structural safety patterns for building the autonomous-operate machinery itself, distilled from design review: self-referential structures (trees, version chains) need a cycle guard on write plus a detect-and-fail walk; a multi-step apply (writes + an emitted event) must be all-or-nothing via one transaction with an outbox, or a mid-crash leaves a live row with no log/event; assembling from several tables needs one consistent point-in-time snapshot; "at most one live X per target" must be enforced at every layer down to the underlying row, not just the queue; bulk operations need bulk confirmation (per-item confirm doesn't scale to an onboarding flood); transient/infrastructure failures must be filtered out before they're allowed to lower a capability's trust; derived indexes/caches go stale silently and need a freshness stamp; recovery must not depend on the thing being recovered (the rollback path can't route through the agents it's rolling back); blast radius caps the trust ceiling regardless of verifiability, because self-modification is a residual risk that's managed (conservative early trust, human-in-the-loop, external rollback), not solved.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/NOTES_running_synthesis_principles(20).md §Building discipline
- **relations:** trust ledger (above); change-layer integration contract (above)
- **verify-later:** none (a design checklist, not a built artifact)

<!-- SOURCE: U15_docs019_running_notes.md -->
### Trust ratchet & capability ceiling model
- **category:** NEW:autonomy-trust-model
- **status-signal:** aspirational
- **status-evidence:** "Bottleneck is trust, not capability... Automation is a per-capability ratchet, not a switch. Bidirectional ratchet." (NOTES_running_synthesis_principles(59) §Trust, reliability, and the ratchet — no implementation evidence anywhere in this file set).
- **what:** A design framework (never implemented, purely a framing document across all five families' shared preamble) for autonomous build/operate systems: trust is per-(tenant, capability), starts at the most conservative level, and moves on a bidirectional ratchet (losable, not just gainable) governed by a "trust ledger." A capability's ceiling is set by verifiability (can ground truth confirm it) × containment (blast radius), independent of how mature/trusted it currently is; the reliability cascade for any task is reuse → generate+verify → compete+judge → HITL, highest-reliability tier first; de-graduation (tightening) may auto-apply on severe evidence, but graduation (loosening) is always confirm-not-initiate — the core safety asymmetry.
- **sources:** NOTES_running_synthesis_principles(59), NOTES_running_synthesis_v2(36).md, v3(32), v4(39) — shared §Trust/§Build-vs-operate preamble (identical across all four).
- **relations:** Governance/HITL confirm-not-initiate model; onboarding/config three-layer model; requirement-mediation model.
- **verify-later:** No known code implements a "trust ledger" or per-capability ceiling table — treat as pure design framing pending stage-2 verification.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Requirement-mediation model ("right" as balance)
- **category:** NEW:autonomy-trust-model
- **status-signal:** aspirational
- **status-evidence:** "'Right' is a requirement-relative balance among conflicting dimensions (fast/secure/generic/simple/functional). Not pick, not merge." (principles(59) §"Right" as balance, not a single answer).
- **what:** A design framing for resolving competing quality dimensions in generated artifacts: authored solutions are treated as extremes that bound a solution space, a mediator finds the requirement-relative point inside it, priority is ordered (not numerically weighted) and modulated by direction-of-travel, and a satisfied concern demotes from "author" to passive "checker" (re-promoting if a later change breaks it) — unifying single-author and multi-author review as two modes of one process. Multi-author deliberation surfaces tradeoffs but cannot itself resolve value-laden conflicts; those still land with a human/authority model.
- **sources:** NOTES_running_synthesis_principles(59) §"Right" as balance (shared preamble across all four non-fixloop families).
- **relations:** Trust ratchet & capability ceiling model; governance/HITL confirm-not-initiate model.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Trust ratchet & capability ceiling model
- **category:** NEW:autonomy-trust-model
- **status-signal:** aspirational
- **status-evidence:** "Bottleneck is trust, not capability... Automation is a per-capability ratchet, not a switch. Bidirectional ratchet." (NOTES_running_synthesis_principles(59) §Trust, reliability, and the ratchet — no implementation evidence anywhere in this file set).
- **what:** A design framework (never implemented, purely a framing document across all five families' shared preamble) for autonomous build/operate systems: trust is per-(tenant, capability), starts at the most conservative level, and moves on a bidirectional ratchet (losable, not just gainable) governed by a "trust ledger." A capability's ceiling is set by verifiability (can ground truth confirm it) × containment (blast radius), independent of how mature/trusted it currently is; the reliability cascade for any task is reuse → generate+verify → compete+judge → HITL, highest-reliability tier first; de-graduation (tightening) may auto-apply on severe evidence, but graduation (loosening) is always confirm-not-initiate — the core safety asymmetry.
- **sources:** NOTES_running_synthesis_principles(59), NOTES_running_synthesis_v2(36).md, v3(32), v4(39) — shared §Trust/§Build-vs-operate preamble (identical across all four).
- **relations:** Governance/HITL confirm-not-initiate model; onboarding/config three-layer model; requirement-mediation model.
- **verify-later:** No known code implements a "trust ledger" or per-capability ceiling table — treat as pure design framing pending stage-2 verification.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Requirement-mediation model ("right" as balance)
- **category:** NEW:autonomy-trust-model
- **status-signal:** aspirational
- **status-evidence:** "'Right' is a requirement-relative balance among conflicting dimensions (fast/secure/generic/simple/functional). Not pick, not merge." (principles(59) §"Right" as balance, not a single answer).
- **what:** A design framing for resolving competing quality dimensions in generated artifacts: authored solutions are treated as extremes that bound a solution space, a mediator finds the requirement-relative point inside it, priority is ordered (not numerically weighted) and modulated by direction-of-travel, and a satisfied concern demotes from "author" to passive "checker" (re-promoting if a later change breaks it) — unifying single-author and multi-author review as two modes of one process. Multi-author deliberation surfaces tradeoffs but cannot itself resolve value-laden conflicts; those still land with a human/authority model.
- **sources:** NOTES_running_synthesis_principles(59) §"Right" as balance (shared preamble across all four non-fixloop families).
- **relations:** Trust ratchet & capability ceiling model; governance/HITL confirm-not-initiate model.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Chain-of-thought prompt pattern catalog
- **category:** reasoning
- **status-signal:** unknown
- **status-evidence:** Presented as a curated list of externally-sourced/community prompts with no indication any is wired into an actual chassis agent's system prompt
- **what:** A reference collection of five chain-of-thought prompting archetypes: (1) "Step Budget and Reflection" — scratchpad thinking with a numeric step budget and self-scored confidence driving continue/backtrack; (2) "Stream-of-Consciousness" — raw, marker-tagged unpolished reasoning trace; (3) "Panel of Experts" — simulated multi-domain-expert debate with per-claim correctness percentages; (4) "Enhanced Reasoning Protocol" — a two-stage consult-then-branch protocol; (5) classic baseline CoT ("Let's think step by step").
- **sources:** reasoning/001_chain_of_thought_prompts.md
- **relations:** n/a
- **verify-later:** whether any agent_definitions system prompt actually uses one of these five patterns

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Salience over presence (context bundle)
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §1 "an LLM loses the bigger picture not because the text left the window … but because local detail is more salient mid-reasoning"
- **what:** The reframe underpinning the whole salience thread: attention follows the concrete and immediate, so the lever is salience at the moment of decision, not mere presence in a task-scoped context bundle.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#1, ED/FOCUS_salience_and_multi_author_mediation(4).md#4
- **relations:** authored-vs-derived context; step-type-aware composition; checker model
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Four axes governing a development step
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §2 table (Purpose/How-well/Where-heading/What-is); "The dynamic axis was the gap"
- **what:** A dev step is governed by four axes — Purpose (why-chain, vertical), How-well (concern tree, horizontal), Where-heading (direction-of-travel, dynamic), What-is (code+state, local). The dynamic trajectory axis was the missing one: a snapshot says where things are, not where they're heading.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#2, ED/FOCUS_salience_and_multi_author_mediation(4).md#3
- **relations:** why-chain; direction-of-travel; concern tree
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Why-chain (objective-tree traversal)
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §3 "a traversal of the existing objective tree … Stable, low-churn, human-owned"
- **what:** The purpose axis rendered as a root-to-node path over the existing objective tree. Turned into a *question* at decision/gate points ("does this serve [why-chain]?") — described as the strongest, cheapest anti-drift mechanism.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#3, ED/FOCUS_salience_and_multi_author_mediation(4).md#4
- **relations:** four axes; priority profile; objective tree
- **verify-later:** existing objective/agent tree

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Direction-of-travel (trajectory layer)
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §3 "a vector, not a reason … the authored vector laid over the derived change-layer … freshness-stamped"
- **what:** A fast-churn dynamic layer capturing current heading, settled-don't-relitigate decisions, deliberately-temporary states, and what's in flux. Proposed by the system from recent diffs but only human confirmation makes it authored-by-record; kept thin, pointer-rich, freshness-stamped, surfaced flagged-stale rather than silently trusted.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#3, ED/FOCUS_salience_and_multi_author_mediation(4).md#9.6
- **relations:** why-chain; authored-vs-derived context; priority profile
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Step-type-aware prompt composition (altitude-aware)
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §4 "framing/routing → full why-chain + direction; generation → collapse to a one-line tether … depth is a virtue, not only a failure mode"
- **what:** Prompt composition made altitude-aware: framing/routing gets the full why-chain + direction; generation collapses to a one-line tether; conformance is local; fitness-check and gate get full why-chain.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#4
- **relations:** salience over presence; why-chain; prompt-composition pattern
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Checker model (single-axis parallel checkers)
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §5 "run several budgets, each narrow, each fully salient on one axis … curators and the advocate already are"
- **what:** Because one attention budget can't hold detail and breadth at once, run several narrow single-axis checkers fired *at decision points*, returning terse verdicts that are reconciled. Parallelism produces verdicts, not decisions — arbitration stays singular.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#5
- **relations:** curators/advocate; multi-author generation; mediator
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Multi-author generation (every concern authors a full solution)
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §6 "each perspective is an author, not a guardrail … generative competition along the concern axis"
- **what:** Instead of guardrails, each implicated concern authors its own maximally-on-axis solution; disagreements become worked demonstrations, not complaints. Reuses cascade tier-3/mediator/advocate but competes N attempts at *different* objectives. Bounded by routing (~2–4 implicated concerns) and counter-proposals-on-deltas.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#6, ED/FOCUS_salience_and_multi_author_mediation(4).md#5
- **relations:** reliability cascade; mediator as multi-objective optimiser; N-round convergence
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Mediator as multi-objective optimiser
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §7 "'Right' = a requirement-relative, defensible balance … not pick, not merge"; "a heuristic floor … full mediation"
- **what:** The mediator finds the requirement-relative balance point among conflicting dimensions using the priority profile, with authored solutions as the extremes that bound the space. Priority is not global; provenance informs weighting, not deference. A cheap heuristic floor settles the uncontested majority but must emit a decision + provenance and be auto-flagged for re-mediation when the why-chain no longer matches its baked-in assumptions.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#7, ED/FOCUS_salience_and_multi_author_mediation(4).md#10
- **relations:** priority profile; multi-author generation; N-round convergence; drift detection
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### N-round convergence (author/checker modes)
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §8 "replaces that one hard step with several easier ones … non-convergence is the escalation signal"
- **what:** Replaces single-round reconciliation of N whole solutions with rounds where each active concern reacts to a candidate (satisfied → checker mode; still-needs → author mode), so the active set shrinks. Non-convergence isolates the one genuine value-laden tradeoff to escalate; bounded by an audit-pass-style cap; a concern can withdraw or be dismissed.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#8, ED/FOCUS_salience_and_multi_author_mediation(4).md#12
- **relations:** multi-author generation; checker model; mediator; candidate ownership
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### N-round candidate ownership (owned by no concern; rival-base)
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §12 "Resolution: the candidate is owned by no concern … seeded deliberately from the profile … The honest limitation: this is path-dependent"
- **what:** The per-round candidate is a shared artifact plus a change log, seeded from known-good or the highest-priority dimension's solution, changed only by premised targeted proposals the mediator adjudicates. Path-dependent and biased toward the seed; rival-base proposals are the gated escape from seed bias.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#12, ED/FOCUS_salience_and_multi_author_mediation(4).md#10
- **relations:** N-round convergence; mediator reasoning trace; known-good library
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Self-development coding pipeline — coordination positions A/B/C
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_self_development(1) header "Status: exploratory. Nothing decided"; §5 "the live disagreement … How do area-owning agents coordinate a change that touches more than one area?"
- **what:** Use AI agents to develop the platform itself (one+ agent per area = one focus doc + its code slice). Unresolved crux: Position A work-item/ownership-serialized, Position B synchronous inter-agent negotiation, Position C a mediated go-between; current lean is C in a spawn-fresh variant (a per-change orchestrator that spawns current area-owners on demand, dissolving the ephemeral-worker staleness problem).
- **sources:** ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#5, ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#2, ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#6
- **relations:** mediator routing model; toolchain validator; MASTER control loop; spawn-fresh coordinator
- **verify-later:** existing spawn machinery; coordinator agent_category

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Chain-of-thought prompt pattern catalog
- **category:** reasoning
- **status-signal:** unknown
- **status-evidence:** Presented as a curated list of externally-sourced/community prompts with no indication any is wired into an actual chassis agent's system prompt
- **what:** A reference collection of five chain-of-thought prompting archetypes: (1) "Step Budget and Reflection" — scratchpad thinking with a numeric step budget and self-scored confidence driving continue/backtrack; (2) "Stream-of-Consciousness" — raw, marker-tagged unpolished reasoning trace; (3) "Panel of Experts" — simulated multi-domain-expert debate with per-claim correctness percentages; (4) "Enhanced Reasoning Protocol" — a two-stage consult-then-branch protocol; (5) classic baseline CoT ("Let's think step by step").
- **sources:** reasoning/001_chain_of_thought_prompts.md
- **relations:** n/a
- **verify-later:** whether any agent_definitions system prompt actually uses one of these five patterns

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Salience over presence (context bundle)
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §1 "an LLM loses the bigger picture not because the text left the window … but because local detail is more salient mid-reasoning"
- **what:** The reframe underpinning the whole salience thread: attention follows the concrete and immediate, so the lever is salience at the moment of decision, not mere presence in a task-scoped context bundle.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#1, ED/FOCUS_salience_and_multi_author_mediation(4).md#4
- **relations:** authored-vs-derived context; step-type-aware composition; checker model
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Four axes governing a development step
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §2 table (Purpose/How-well/Where-heading/What-is); "The dynamic axis was the gap"
- **what:** A dev step is governed by four axes — Purpose (why-chain, vertical), How-well (concern tree, horizontal), Where-heading (direction-of-travel, dynamic), What-is (code+state, local). The dynamic trajectory axis was the missing one: a snapshot says where things are, not where they're heading.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#2, ED/FOCUS_salience_and_multi_author_mediation(4).md#3
- **relations:** why-chain; direction-of-travel; concern tree
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Why-chain (objective-tree traversal)
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §3 "a traversal of the existing objective tree … Stable, low-churn, human-owned"
- **what:** The purpose axis rendered as a root-to-node path over the existing objective tree. Turned into a *question* at decision/gate points ("does this serve [why-chain]?") — described as the strongest, cheapest anti-drift mechanism.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#3, ED/FOCUS_salience_and_multi_author_mediation(4).md#4
- **relations:** four axes; priority profile; objective tree
- **verify-later:** existing objective/agent tree

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Direction-of-travel (trajectory layer)
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §3 "a vector, not a reason … the authored vector laid over the derived change-layer … freshness-stamped"
- **what:** A fast-churn dynamic layer capturing current heading, settled-don't-relitigate decisions, deliberately-temporary states, and what's in flux. Proposed by the system from recent diffs but only human confirmation makes it authored-by-record; kept thin, pointer-rich, freshness-stamped, surfaced flagged-stale rather than silently trusted.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#3, ED/FOCUS_salience_and_multi_author_mediation(4).md#9.6
- **relations:** why-chain; authored-vs-derived context; priority profile
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Step-type-aware prompt composition (altitude-aware)
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §4 "framing/routing → full why-chain + direction; generation → collapse to a one-line tether … depth is a virtue, not only a failure mode"
- **what:** Prompt composition made altitude-aware: framing/routing gets the full why-chain + direction; generation collapses to a one-line tether; conformance is local; fitness-check and gate get full why-chain.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#4
- **relations:** salience over presence; why-chain; prompt-composition pattern
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Checker model (single-axis parallel checkers)
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §5 "run several budgets, each narrow, each fully salient on one axis … curators and the advocate already are"
- **what:** Because one attention budget can't hold detail and breadth at once, run several narrow single-axis checkers fired *at decision points*, returning terse verdicts that are reconciled. Parallelism produces verdicts, not decisions — arbitration stays singular.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#5
- **relations:** curators/advocate; multi-author generation; mediator
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Multi-author generation (every concern authors a full solution)
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §6 "each perspective is an author, not a guardrail … generative competition along the concern axis"
- **what:** Instead of guardrails, each implicated concern authors its own maximally-on-axis solution; disagreements become worked demonstrations, not complaints. Reuses cascade tier-3/mediator/advocate but competes N attempts at *different* objectives. Bounded by routing (~2–4 implicated concerns) and counter-proposals-on-deltas.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#6, ED/FOCUS_salience_and_multi_author_mediation(4).md#5
- **relations:** reliability cascade; mediator as multi-objective optimiser; N-round convergence
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Mediator as multi-objective optimiser
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §7 "'Right' = a requirement-relative, defensible balance … not pick, not merge"; "a heuristic floor … full mediation"
- **what:** The mediator finds the requirement-relative balance point among conflicting dimensions using the priority profile, with authored solutions as the extremes that bound the space. Priority is not global; provenance informs weighting, not deference. A cheap heuristic floor settles the uncontested majority but must emit a decision + provenance and be auto-flagged for re-mediation when the why-chain no longer matches its baked-in assumptions.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#7, ED/FOCUS_salience_and_multi_author_mediation(4).md#10
- **relations:** priority profile; multi-author generation; N-round convergence; drift detection
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### N-round convergence (author/checker modes)
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §8 "replaces that one hard step with several easier ones … non-convergence is the escalation signal"
- **what:** Replaces single-round reconciliation of N whole solutions with rounds where each active concern reacts to a candidate (satisfied → checker mode; still-needs → author mode), so the active set shrinks. Non-convergence isolates the one genuine value-laden tradeoff to escalate; bounded by an audit-pass-style cap; a concern can withdraw or be dismissed.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#8, ED/FOCUS_salience_and_multi_author_mediation(4).md#12
- **relations:** multi-author generation; checker model; mediator; candidate ownership
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### N-round candidate ownership (owned by no concern; rival-base)
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §12 "Resolution: the candidate is owned by no concern … seeded deliberately from the profile … The honest limitation: this is path-dependent"
- **what:** The per-round candidate is a shared artifact plus a change log, seeded from known-good or the highest-priority dimension's solution, changed only by premised targeted proposals the mediator adjudicates. Path-dependent and biased toward the seed; rival-base proposals are the gated escape from seed bias.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#12, ED/FOCUS_salience_and_multi_author_mediation(4).md#10
- **relations:** N-round convergence; mediator reasoning trace; known-good library
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Self-development coding pipeline — coordination positions A/B/C
- **category:** reasoning
- **status-signal:** aspirational
- **status-evidence:** FOCUS_self_development(1) header "Status: exploratory. Nothing decided"; §5 "the live disagreement … How do area-owning agents coordinate a change that touches more than one area?"
- **what:** Use AI agents to develop the platform itself (one+ agent per area = one focus doc + its code slice). Unresolved crux: Position A work-item/ownership-serialized, Position B synchronous inter-agent negotiation, Position C a mediated go-between; current lean is C in a spawn-fresh variant (a per-change orchestrator that spawns current area-owners on demand, dissolving the ephemeral-worker staleness problem).
- **sources:** ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#5, ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#2, ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#6
- **relations:** mediator routing model; toolchain validator; MASTER control loop; spawn-fresh coordinator
- **verify-later:** existing spawn machinery; coordinator agent_category

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Abandoned "no owner" claim (checked and found false)
- **category:** NEW:investigation-discipline
- **status-signal:** abandoned
- **status-evidence:** "An earlier claim in this plan — 'no agent owns ensuring a tool page has a working widget...' — was checked and found false." (PLAN_tool_widget_clobber(9).md §2.6)
- **what:** During the M2 diagnosis, the investigation asserted that no agent owned ensuring an adopted tool page gets a working widget. Verification against `apply_adoption_plan_action.go`, `check_tool_completeness_action.go`, and the agent-definitions backup showed `tool-recreation-handler` is a real, registered, active agent that already owns exactly this responsibility. The claim was retracted before any redundant handoff was built.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md#2.6, tools/tool_widget_clobber/NOTES_running_tool_widget_investigation.md#Phase-3,#Phase-4
- **relations:** Adoption interactivity misroute; Verify-before-acting investigation discipline
- **verify-later:** confirm `tool-recreation-handler` agent_definition remains registered/active

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Verify-before-acting investigation discipline (diagnosis methodology)
- **category:** NEW:investigation-discipline
- **status-signal:** deployed
- **status-evidence:** NOTES_running_tool_widget_investigation.md, whole document, esp. "the diagnosis changed three times, and each change came from refusing to act on the current theory until it was checked"
- **what:** A recorded set of working principles used through the tool-widget investigation: don't jump to conclusions, verify architectural claims by code search before turning them into tasks, prefer structural fixes over quick hacks, reuse existing helpers rather than building parallel ones, check the schema before writing SQL, make falsifiable predictions rather than declarative claims. Framed as reusable guidance and as raw material for a fix-loop council member.
- **sources:** tools/tool_widget_clobber/NOTES_running_tool_widget_investigation.md#How-the-diagnosis-evolved,#Principles-that-actually-drove-the-work
- **relations:** Abandoned "no owner" claim; fix-loop / diagnosis-loop council concept
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Abandoned "no owner" claim (checked and found false)
- **category:** NEW:investigation-discipline
- **status-signal:** abandoned
- **status-evidence:** "An earlier claim in this plan — 'no agent owns ensuring a tool page has a working widget...' — was checked and found false." (PLAN_tool_widget_clobber(9).md §2.6)
- **what:** During the M2 diagnosis, the investigation asserted that no agent owned ensuring an adopted tool page gets a working widget. Verification against `apply_adoption_plan_action.go`, `check_tool_completeness_action.go`, and the agent-definitions backup showed `tool-recreation-handler` is a real, registered, active agent that already owns exactly this responsibility. The claim was retracted before any redundant handoff was built.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md#2.6, tools/tool_widget_clobber/NOTES_running_tool_widget_investigation.md#Phase-3,#Phase-4
- **relations:** Adoption interactivity misroute; Verify-before-acting investigation discipline
- **verify-later:** confirm `tool-recreation-handler` agent_definition remains registered/active

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Verify-before-acting investigation discipline (diagnosis methodology)
- **category:** NEW:investigation-discipline
- **status-signal:** deployed
- **status-evidence:** NOTES_running_tool_widget_investigation.md, whole document, esp. "the diagnosis changed three times, and each change came from refusing to act on the current theory until it was checked"
- **what:** A recorded set of working principles used through the tool-widget investigation: don't jump to conclusions, verify architectural claims by code search before turning them into tasks, prefer structural fixes over quick hacks, reuse existing helpers rather than building parallel ones, check the schema before writing SQL, make falsifiable predictions rather than declarative claims. Framed as reusable guidance and as raw material for a fix-loop council member.
- **sources:** tools/tool_widget_clobber/NOTES_running_tool_widget_investigation.md#How-the-diagnosis-evolved,#Principles-that-actually-drove-the-work
- **relations:** Abandoned "no owner" claim; fix-loop / diagnosis-loop council concept
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U14_docs019_runbooks.md -->
### Standing evidence rules (the working-method contract)
- **category:** NEW:operating-doctrine
- **status-signal:** deployed
- **status-evidence:** code_retrieval_route(21) header "Standing rules: user runs all SQL/kubectl/builds; read outcomes by correlation_id only; snapshot_agent before agent_definitions UPDATEs; schema before SQL; a 0-rows result is not decisive until the query itself is checked."
- **what:** The recurring operating contract of every runbook in this unit: the human runs all mutations/builds; outcomes are read by correlation_id, never `ORDER BY … LIMIT 1` (twice a red herring); `\d <table>` before every query; every agent_definitions change snapshots the row first (`snapshot_agent` = byte-exact revert path); a 0-rows result proves nothing until the query/selector is validated (wrong key, wrong label, wrong nesting all produced false zeros); migrations are self-guarded (UPDATE 0 = assumption wrong, nothing changed) and carry REVERT blocks.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#header; docs019/RUNBOOK(31)_diagnosis_loop.md#6C; docs019/RUNBOOK_gamesdesign_index_rebuild.md#7 (0-rows reminder)
- **relations:** instrument skepticism; repo-label bug (LIMIT-1 lesson); diagnostician seed→fix (snapshot rule)
- **verify-later:** snapshot_agent function; snapshots table growth

<!-- SOURCE: U14_docs019_runbooks.md -->
### Parallel-thread boundary and handoff convention
- **category:** NEW:operating-doctrine
- **status-signal:** deployed
- **status-evidence:** builder_route(21) "THREAD HANDED OFF 2026-07-06 → HANDOFF_builder_thread.md"; §B5 "BOUNDARY (adopted): the other chat owns everything INSIDE the tool pipeline …; This chat owns the RELAY …; The §B5 interface … is a JOINT decision, not taken unilaterally"; fix_loop "RULE: any fix-loop change to diagnose workflows is fetch-first against the CURRENT JSON and coordinated".
- **what:** Multiple concurrent working threads (builder, tools, quality, fix-loop, imagery) each own declared surfaces; runbooks record explicit boundaries, joint-decision seams, collision surfaces, and fetch-first rules for shared state; work moves between threads via handoff documents and "this item retains / that thread owns" dispositions. This is how the runbook families themselves relate — each family is one thread's travelling state.
- **sources:** docs019/RUNBOOK_builder_route(21).md#handoff-banner; docs019/RUNBOOK_builder_route(21).md#B5; docs019/RUNBOOK_diagnosis_fix_loop(9).md#boundaries; docs019/RUNBOOK_site_quality(1).md#boundaries
- **relations:** doc_plans/doc_notes; per-task notes; documentation-system travelling docs
- **verify-later:** HANDOFF_builder_thread.md; boundary sections across sibling runbooks

<!-- SOURCE: U14_docs019_runbooks.md -->
### Standing evidence rules (the working-method contract)
- **category:** NEW:operating-doctrine
- **status-signal:** deployed
- **status-evidence:** code_retrieval_route(21) header "Standing rules: user runs all SQL/kubectl/builds; read outcomes by correlation_id only; snapshot_agent before agent_definitions UPDATEs; schema before SQL; a 0-rows result is not decisive until the query itself is checked."
- **what:** The recurring operating contract of every runbook in this unit: the human runs all mutations/builds; outcomes are read by correlation_id, never `ORDER BY … LIMIT 1` (twice a red herring); `\d <table>` before every query; every agent_definitions change snapshots the row first (`snapshot_agent` = byte-exact revert path); a 0-rows result proves nothing until the query/selector is validated (wrong key, wrong label, wrong nesting all produced false zeros); migrations are self-guarded (UPDATE 0 = assumption wrong, nothing changed) and carry REVERT blocks.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#header; docs019/RUNBOOK(31)_diagnosis_loop.md#6C; docs019/RUNBOOK_gamesdesign_index_rebuild.md#7 (0-rows reminder)
- **relations:** instrument skepticism; repo-label bug (LIMIT-1 lesson); diagnostician seed→fix (snapshot rule)
- **verify-later:** snapshot_agent function; snapshots table growth

<!-- SOURCE: U14_docs019_runbooks.md -->
### Parallel-thread boundary and handoff convention
- **category:** NEW:operating-doctrine
- **status-signal:** deployed
- **status-evidence:** builder_route(21) "THREAD HANDED OFF 2026-07-06 → HANDOFF_builder_thread.md"; §B5 "BOUNDARY (adopted): the other chat owns everything INSIDE the tool pipeline …; This chat owns the RELAY …; The §B5 interface … is a JOINT decision, not taken unilaterally"; fix_loop "RULE: any fix-loop change to diagnose workflows is fetch-first against the CURRENT JSON and coordinated".
- **what:** Multiple concurrent working threads (builder, tools, quality, fix-loop, imagery) each own declared surfaces; runbooks record explicit boundaries, joint-decision seams, collision surfaces, and fetch-first rules for shared state; work moves between threads via handoff documents and "this item retains / that thread owns" dispositions. This is how the runbook families themselves relate — each family is one thread's travelling state.
- **sources:** docs019/RUNBOOK_builder_route(21).md#handoff-banner; docs019/RUNBOOK_builder_route(21).md#B5; docs019/RUNBOOK_diagnosis_fix_loop(9).md#boundaries; docs019/RUNBOOK_site_quality(1).md#boundaries
- **relations:** doc_plans/doc_notes; per-task notes; documentation-system travelling docs
- **verify-later:** HANDOFF_builder_thread.md; boundary sections across sibling runbooks

<!-- SOURCE: U25_leopardess_social.md -->
### In-chassis replicability requirement for operator work
- **category:** NEW:operator-practice
- **status-signal:** deployed
- **status-evidence:** REPLICATION_in_chassis.md (2026-07-10) maps every off-platform action to [chassis]/[human]/[gap]; RUNNING_NOTES turn 8: "everything done in this thread must be replicable inside the chassis … Either document the tool-free path or don't use the tool."
- **what:** Standing owner rule: interactive-agent work on sites must be reproducible by the chassis itself (or documented as a human-judgement or platform-gap item). The audit, spec rewrites, artifact verification and imagery deploys all map to normal platform operations; the genuinely human items (choosing a permanent logo, stating personal-history claims, setting engagement terms) are deliberately not automated; named gaps: pinned enforcement, favicon/OG derivation, chart capability.
- **sources:** docs/leopardessconsulting/REPLICATION_in_chassis.md (whole); docs/leopardessconsulting/RUNNING_NOTES.md#Turn-8
- **relations:** hitl; documentation-system; site_specs pinned gap
- **verify-later:** n/a (practice doc); checkpoint_for_review as the in-chassis review surface

<!-- SOURCE: U25_leopardess_social.md -->
### Operator discipline: verify-by-artifact, dated backups, kcat generic trigger
- **category:** NEW:operator-practice
- **status-signal:** deployed
- **status-evidence:** PLAN standing rule 2: "Verify by artifact, never by report. (This platform has a long history of builds reporting success while building nothing.)"; RUNBOOK landmines 2/17/18; VERDICT §7 "Verified by artifact, never by item status" with md5 predictions.
- **what:** The cross-workstream operating discipline: (1) never trust a `complete` work item — curl the page, read the DB row, diff the bytes (strongest form: predict output bytes offline and md5-compare); (2) back up before ANY change using dated `bak_*`/`_backup_YYYYMMDD` tables and never reuse a name (CREATE TABLE IF NOT EXISTS silently no-ops); (3) trigger any agent by producing to Kafka system.agent.generic.requests via kcat with the standard header set; (4) kubectl exec heredocs need `-i` or silently run nothing — prefer kubectl cp + psql -f; quote heredoc delimiters.
- **sources:** docs/leopardessconsulting/RUNBOOK.md#Landmines; docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#7; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#8; docs/leopardessconsulting/HANDOFF.md#5
- **relations:** silent no-op success class; fleet generalisation doctrine
- **verify-later:** n/a (doctrine); bak_* table inventory in clients_db

<!-- SOURCE: U25_leopardess_social.md -->
### In-chassis replicability requirement for operator work
- **category:** NEW:operator-practice
- **status-signal:** deployed
- **status-evidence:** REPLICATION_in_chassis.md (2026-07-10) maps every off-platform action to [chassis]/[human]/[gap]; RUNNING_NOTES turn 8: "everything done in this thread must be replicable inside the chassis … Either document the tool-free path or don't use the tool."
- **what:** Standing owner rule: interactive-agent work on sites must be reproducible by the chassis itself (or documented as a human-judgement or platform-gap item). The audit, spec rewrites, artifact verification and imagery deploys all map to normal platform operations; the genuinely human items (choosing a permanent logo, stating personal-history claims, setting engagement terms) are deliberately not automated; named gaps: pinned enforcement, favicon/OG derivation, chart capability.
- **sources:** docs/leopardessconsulting/REPLICATION_in_chassis.md (whole); docs/leopardessconsulting/RUNNING_NOTES.md#Turn-8
- **relations:** hitl; documentation-system; site_specs pinned gap
- **verify-later:** n/a (practice doc); checkpoint_for_review as the in-chassis review surface

<!-- SOURCE: U25_leopardess_social.md -->
### Operator discipline: verify-by-artifact, dated backups, kcat generic trigger
- **category:** NEW:operator-practice
- **status-signal:** deployed
- **status-evidence:** PLAN standing rule 2: "Verify by artifact, never by report. (This platform has a long history of builds reporting success while building nothing.)"; RUNBOOK landmines 2/17/18; VERDICT §7 "Verified by artifact, never by item status" with md5 predictions.
- **what:** The cross-workstream operating discipline: (1) never trust a `complete` work item — curl the page, read the DB row, diff the bytes (strongest form: predict output bytes offline and md5-compare); (2) back up before ANY change using dated `bak_*`/`_backup_YYYYMMDD` tables and never reuse a name (CREATE TABLE IF NOT EXISTS silently no-ops); (3) trigger any agent by producing to Kafka system.agent.generic.requests via kcat with the standard header set; (4) kubectl exec heredocs need `-i` or silently run nothing — prefer kubectl cp + psql -f; quote heredoc delimiters.
- **sources:** docs/leopardessconsulting/RUNBOOK.md#Landmines; docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#7; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#8; docs/leopardessconsulting/HANDOFF.md#5
- **relations:** silent no-op success class; fleet generalisation doctrine
- **verify-later:** n/a (doctrine); bak_* table inventory in clients_db
