
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
