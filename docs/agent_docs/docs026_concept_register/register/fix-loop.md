# Register — fix-loop

> **covers-through: 2026-07-16** · FIX-051/052/053 added 2026-07-16 (post-freeze hand-patch).
> Everything else dates from the 2026-07-13 extraction freeze — absence
> here is not evidence of absence in the platform. See `bugs_open/106`.

53 concepts (50 from stage 1 consolidation + 3 added 2026-07-16 for the
triage/escalation subsystem that shipped after extraction froze — FIX-051/052/053),
consolidated from 123 raw extractions across units U08, U13, U14, U15, U16
(note: U13's ~47 blocks and U16's autonomy-governance-tagged blocks appeared byte-identically
twice within the cluster input file — treated as duplicate copies of the same extraction, not
independent corroboration, and collapsed accordingly).

### FIX-001 — Diagnosis→fix loop programme / council loop (F0–F3)
- **status:** deployed
- **status-evidence:** Originally a design proposal (2026-07-06/07, "DISCUSSION COMPLETE for F0/F1... CUTOVER-READY", no build claimed); by 2026-07-13 "PR #1 — github.com/gqls/agentchassis/pull/1 — APPROVED & MERGED" and "Today it did the whole thing, end to end, for the first time" (SUMMARY_where_we_are_2026-07-13.md).
- **what:** The overarching workstream and its shipped end-to-end system: turning a plain-English bug symptom into a human-reviewed pull request via intake → read-only diagnosis (cite-or-abstain, three evidence tiers) → constrained fix plan → two-reviewer council with deterministic decision → dedicated-pod implementer behind a hard file allowlist → containerized build gate → PR. Phased as F0 (intake/observability/egress) → F1 (write step) → F2 (council + decision-maker) → F3 (learning, never built). Every stage writes to diagnosis_artifacts/orchestration_states keyed on one correlation_id so the whole run is auditable; human review is the terminal, nothing merges itself. Documented across multiple founding-to-milestone snapshots spanning the 2026-07-06 founding notes through the 2026-07-13 first fully proven end-to-end run.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#THE TASK, fixloop_eg_dartsonline/SUMMARY_where_we_are_2026-07-13.md, fixloop_eg_dartsonline/README_so_far.md, docs019/RUNBOOK_diagnosis_fix_loop(9).md#the-task, NOTES_running_fixloop(9).md, HANDOFF_fixloop_thread(8).md
- **relations:** all other fix-loop concepts in this register are sub-scopes of this one; docs026 concept-register's own consumer (council-agents stage)
- **verify-later:** PR #1 on github.com/gqls/agentchassis; diagnosis_artifacts table contents on clients_db

### FIX-002 — fix-proposer agent / constrained edit plan (F1.1a)
- **status:** deployed
- **status-evidence:** "F1.1a — ✅ BUILT 2026-07-10" (PLAN_fixloop_pilot.md); live agent seeded in 0NN_fix_proposer.sql; FYI addendum 2026-07-10 "a fix-proposer agent (F1.1a) now exists."
- **what:** An agent_definitions workflow that loads a diagnosis by fix_correlation_id, refuses anything not CONFIRMED, and drafts a constrained edit plan (summary, ≤8 allowlisted edits with file/symbol/operation/rationale/sketch, grounded_in quotes required, risks) persisted to diagnosis_artifacts (kind='fix_plan'). It reads only orchestration_states/diagnosis_artifacts and writes no code — no git token needed.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#propose step, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §3b F1.1a, FYI_from_fixloop_2026-07-10_verdict_prompt_symptom_check.md#addendum
- **relations:** diagnose_persist_fix_plan validator; two-reviewer council; CONFIRMED gate
- **verify-later:** fix-proposer agent_definitions row, workflow version

### FIX-003 — needs_diagnosis intake route (F0.1c)
- **status:** deployed
- **status-evidence:** "F0.1c — ✅ LANDED 2026-07-09" (PLAN_fixloop_pilot.md §1); earlier design decision "Q-B needs_diagnosis item in a NEW pipeline='diagnose' namespace" from 2026-07-07.
- **what:** The one documented way a bug enters the loop: 090_TRIGGER_needs_diagnosis_v1.sh writes a durable site_work_items row (pipeline='diagnose', item_type='needs_diagnosis', status='awaiting_diagnosis') and fires the diagnose-orchestrator Kafka envelope on the same correlation_id, so the intake record, diagnosis_artifacts bundles, and terminal doc_notes row all join on one key. DISPATCH=0 records without firing; the older 084_TRIGGER_diagnose_v1.sh remains for ad-hoc runs with no intake record. item_key (needs_diagnosis:<slug>) plus idx_swi_dedup makes re-running the same slug idempotent while an intake is open. Rides the existing work-item dispatch + immune system, with anchorless (site-less) runs surviving via load_runtime error-routing (~26 min / 5 iterations observed).
- **sources:** fixloop_eg_dartsonline/090_TRIGGER_needs_diagnosis_v1.sh, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md#F0.1c, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Q-B CORRECTION, docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-B)
- **relations:** system.internal pseudo-site anchor pattern; private inert pipeline statuses pattern; diagnose-dispatch-loop; superseded null-site-allowed intake design
- **verify-later:** site_work_items rows with pipeline='diagnose'; idx_swi_dedup index definition

### FIX-004 — Superseded: null-site-allowed intake design
- **status:** superseded
- **status-evidence:** "Q-B CORRECTION (2026-07-09... 'Null-site allowed' is impossible" (RUNBOOK(10)#Q-B CORRECTION); original design in RUNBOOK(9)#QUESTIONS "null-site allowed".
- **what:** The 2026-07-07 owner decision originally specified that site-less code bugs would "ride null-site" in the new diagnose pipeline namespace. Reading the live schema on 2026-07-09 showed this was structurally impossible twice over (NOT NULL column; site-anchored loader query), and it was replaced by the system.internal pseudo-site anchor pattern.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#QUESTIONS, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Q-B CORRECTION, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 4
- **relations:** system.internal pseudo-site anchor pattern; needs_diagnosis intake route
- **verify-later:** n/a — superseded, no live code to check

### FIX-005 — diagnosis_artifacts table (unified egress store)
- **status:** deployed
- **status-evidence:** "F0.1a — ✅ LANDED 2026-07-09" (PLAN_fixloop_pilot.md); kind list extended live via ALTER in 0NN_fix_proposer.sql v4/v5; originally decided 2026-07-07 as "Q-A diagnosis_artifacts table, written through inside assemble."
- **what:** The durable, correlation_id-keyed egress table for the whole loop, with kind growing over time: bundle and iteration_note (F0.1a) → fix_plan (F1.1a) → council_report (F2.1) → escalation (F2.3). correlation_id is deliberately text not uuid (ExecutionContext.CorrelationID has no guaranteed form). A partial unique index on (correlation_id, iteration) WHERE kind='bundle' gives retry-safe upsert for bundles while allowing multiple iteration_note rows per iteration. doc_notes was considered and set aside for this purpose (notes are prose for humans; bundles are machine-replayable evidence with different retention). Carries a retention knob (expires_at/pinned) that is defined but has no sweep implemented yet.
- **sources:** fixloop_eg_dartsonline/0NN_diagnosis_artifacts.sql, fixloop_eg_dartsonline/0NN_fix_proposer.sql#§1, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §F0.1a, docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-A)
- **relations:** bundle write-through; fix-proposer's fix_plan artifacts; council_report; escalation artifact; retention sweep (unbuilt)
- **verify-later:** table DDL and CHECK constraints on clients_db; whether a retention sweep job exists

### FIX-006 — Retention/expiry knob on diagnosis_artifacts
- **status:** aspirational
- **status-evidence:** "Retention knob. NULL = keep indefinitely. Sweep deletes WHERE expires_at < now() AND NOT pinned" (0NN_diagnosis_artifacts.sql comment) — no sweep job found in scope.
- **what:** expires_at/pinned columns exist on diagnosis_artifacts (bundles configured to expire sooner than notes; NULL = keep forever), and a partial index exists to support a future deletion sweep, but no sweep job/scheduled task was found anywhere in this extraction's file set. The mechanism is designed but not built.
- **sources:** fixloop_eg_dartsonline/0NN_diagnosis_artifacts.sql#idx_diagnosis_artifacts_expiry
- **relations:** diagnosis_artifacts table
- **verify-later:** search codebase/scheduled_tasks for any retention-sweep job referencing diagnosis_artifacts

### FIX-007 — Known-answer benchmark methodology
- **status:** convention
- **status-evidence:** "Benchmark method validated end-to-end: same symptom, one variable cluster, measurable delta" (NOTES(10)#Turn 10).
- **stage2-verified (2026-07-14):** deployed → convention — Benchmark methodology/doctrine, no concrete artifact named beyond a rubric in a doc; process not code.
- **what:** When a candidate bug's mechanism is dissolved by the mandatory cheap pre-check (three candidates in a row were), the pilot is promoted from a "discovery run" to a "known-answer benchmark": the loop is run blind on the original symptom string and scored against a pre-registered rubric of must/should/bonus claims fixed before the run, including a "refutation credit" penalizing confirmation of a known-false standing hypothesis. Produced a repeatable, gradable evaluation across five runs, each of which found and fixed a real engine defect.
- **sources:** fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §0, §3, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#AMENDMENT 2026-07-09, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 1
- **relations:** loop-worthiness test doctrine; blinding discipline; dartsonline guides pilot selection history
- **verify-later:** rubric table in PLAN_fixloop_pilot.md §3

### FIX-008 — Dartsonline guides pilot selection history (three candidates + confirmed pilot)
- **status:** convention
- **status-evidence:** "★ THE PILOT IS CONFIRMED (2026-07-07)... Two earlier candidates were rejected... that triage history is itself the worthiness test working" (HANDOFF_fixloop_thread(8).md); "Chrome pilot EVAPORATED (fixed live)" (RUNBOOK(9)).
- **stage2-verified (2026-07-14):** deployed → convention — Narrative history of pilot-candidate selection (a process record), not a code/infra claim — no artifact to verify beyond docs prose.
- **what:** Three earlier pilot candidates were considered and dropped before the dartsonline guides bug was chosen as the F0 pilot: (1) dartsonline pages lacking site chrome, fixed live before the loop ran (a perishability lesson); (2) a "no submission path produces a roadmap" gap, reclassified as a known platform gap and routed to the builder queue since a human found it by reading two files; (3) a blank guides-index fork where the guide-writing mechanism's existence was unverified. The confirmed pilot: dartsonline published a Guides nav link and blank /guides/index.html while gamesdesign (same platform) has working guides — a two-site differential, the strongest evidence shape — with a standing hypothesis that reconcile_site_plan's routing table has no "guide" entry so planner-emitted guide pages were silently dropped while nav (generated from the plan, not the built set) still published the link.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#F0 PILOT ORIGINAL RECORD, #SUPERSEDED CANDIDATE 2, #PREVIOUS F0 PILOT #1, docs019/RUNBOOK_diagnosis_fix_loop(9).md#f0-pilot, HANDOFF_fixloop_thread(8).md, HANDOFF_fixloop_thread(3)-(5).md
- **relations:** loop-worthiness test doctrine; standing hypothesis refuted; roadmap/phases mechanism
- **verify-later:** load_work_item_actions.go routing table; the pilot's run artifacts

### FIX-009 — Blinding discipline for benchmark runs
- **status:** convention
- **status-evidence:** "BLINDING IS MANDATORY... Exclude .../fixloop_eg_dartsonline/ from the loop's corpus" (RUNBOOK(10)#★ F0 PILOT).
- **stage2-verified (2026-07-14):** deployed → convention — A benchmarking discipline/rule about symptom wording and seed_scope omission; no artifact named to check, correctly a process concept.
- **what:** Established that the diagnose-agent workflow structurally cannot read this docs directory (it walks Go source and DB rows only), so blinding is largely automatic; the only two leak vectors are the symptom string (must describe only observable behaviour) and seed_scope (must be omitted entirely for a benchmark run).
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#REGENERATING THE CONTEXT BUNDLE §BLINDING, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §3
- **relations:** known-answer benchmark methodology
- **verify-later:** diagnose-agent workflow JSON — confirm no doc-reading step was later added

### FIX-010 — Standing hypothesis refuted (reconcile_site_plan routing table)
- **status:** superseded
- **status-evidence:** "THE STANDING HYPOTHESIS IS REFUTED — and it named the wrong file" (NOTES(10)#Turn 1).
- **what:** The original hypothesis blamed reconcile_site_plan's routing table for silently dropping "guide" pages. Hand-diagnosis on 2026-07-09 showed the routing table is real but lives in WriteBuildItemsAction, and absence from it does not drop a page (it defaults to page-build-handler); the actual drop mechanism is a separate unavailableBuilders map, and reconcile_site_plan_action.go has no type switch at all. Retained specifically because a hypothesis a loop should refuse to confirm is exactly the kind of "superseded" idea the register wants captured.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#★ F0 PILOT, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 1
- **relations:** dartsonline guides pilot selection history; two intake paths disagreement
- **verify-later:** grep/inspect reconcile_site_plan; WriteBuildItemsAction; unavailableBuilders

### FIX-011 — Two intake paths disagreement (WriteBuildItemsAction vs reconcile_site_plan)
- **status:** deployed (documented finding, not fixed — routed to builder thread)
- **status-evidence:** "Fourth finding (unlooked-for): the two intake paths disagree" (NOTES(10)#Turn 1).
- **what:** WriteBuildItemsAction deliberately skips tool/entity-directory/entity-page page types via an unavailableBuilders guard, while reconcile_site_plan_action.go hardcodes handler_agent='page-build-handler' for every plan page with no type switch at all, so it re-emits build items for the very types the other path skips. Flagged as a builder-thread decision, not fixed here.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 1, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#★ F0 PILOT
- **relations:** dartsonline guides pilot selection history; pipeline-blind dispatch surfaces
- **verify-later:** grep/inspect WriteBuildItemsAction; tool; entity-directory

### FIX-012 — mark_no_sections — referenced-but-never-built step
- **status:** abandoned
- **status-evidence:** "mark_no_sections does not exist... appears nowhere in the repo but that one comment" (NOTES(10)#Turn 1).
- **what:** A code comment at load_work_item_actions.go:750-756 names a remedy step mark_no_sections that would flag a sectionless page's work item needs_human_review — but the step was never implemented; it appears nowhere else in the repo. The completion guard that would consume its flag faithfully preserves a flag nothing ever sets.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 1, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#★ F0 PILOT
- **relations:** dartsonline guides pilot selection history
- **verify-later:** grep for mark_no_sections in current repo

### FIX-013 — Plan validation / hard allowlist for edit plans (diagnose_persist_fix_plan)
- **status:** deployed
- **status-evidence:** "structural validation... UNLIKE the bundle write-through this FAILS the step on bad input" (NOTES(10)#Turn 15).
- **what:** A deterministic Go action validating a proposer's plan before persisting it: non-empty summary/edits/rationale/sketch, operations restricted to modify|add|remove|config_change, repo-relative paths only, ≤8 edits, grounded_in quotes required, 32KB cap — plus (F1.1b(a)) rejection of explicit no-op phrases so edits that change nothing cannot pass as real edits. Fails closed; fired correctly on the first two real fix-proposer runs (max_tokens truncation).
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 15, #Turn 16, #Turn 18
- **relations:** fix-proposer agent; hard file allowlist (implementer's analogous gate)
- **verify-later:** grep/inspect modify|add|remove|config_change; grounded_in

### FIX-014 — Two-reviewer council (F2.1)
- **status:** deployed
- **status-evidence:** "F2.1 — ✅ PROVEN LIVE 2026-07-10" (PLAN_fixloop_pilot.md).
- **what:** Two sequential LLM reviewer steps plus a deterministic Go decision — not a third model opinion about two model opinions. review_editquality judges whether every edit changes something real and targets the actual causal path; review_guardian judges blast radius, architecture-change signals, and surface ownership, and alone holds the hard veto. Both attach optional checks:[{sql,why}]. First live run judged real objections correctly rather than rubber-stamping.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#review_editquality, #review_guardian, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 18-19
- **relations:** deterministic council decision + hard veto; verify step; schema hint for reviewers; council roster expansion vision
- **verify-later:** grep/inspect review_editquality; review_guardian; checks:[{sql,why}]

### FIX-015 — Deterministic council decision + hard veto (diagnose_council_decide)
- **status:** deployed
- **status-evidence:** "diagnose_council_decide: ordered rules (hard veto → rejected; any veto → rejected; any objection → revise; all approve → approved)" (NOTES(10)#Turn 18).
- **what:** A pure Go action aggregates the two reviewers' verdicts deterministically, auditable and reproducible. hard_veto_from is currently a flag in the workflow step config naming the guardian reviewer as sole veto-holder; whether it should instead live on a reviewer/pipeline definition column remains an open sub-question (Q-D). Malformed reviewer output fails closed.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#council_decide, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#STILL OPEN Q-D
- **relations:** two-reviewer council; hard-veto flag at multiple scopes (superseded early design); decision router (F2.3)
- **verify-later:** grep/inspect hard_veto_from

### FIX-016 — Hard-veto flag at multiple scopes — early design (superseded)
- **status:** superseded
- **status-evidence:** owner decision 2026-07-07: "hard_veto flag, attachable at multiple scopes (a reviewer agent, a pipeline, a specific tool/component)" (RUNBOOK(9)#Q-D) — actual shipped implementation (F2.1, 2026-07-10) is a single hard_veto_from list in one workflow step's config.
- **what:** The original council design envisioned a hard_veto flag placeable at reviewer, pipeline, tool, or component scope — parallel reviewers feed a decision-maker by default (advisory) except where the flag converts a negative verdict into a block — motivated by accessibility/legal review cases. What was actually built (F2.1) is narrower: a single hard_veto_from: ["guardian"] array in the fix-proposer's council_decide step config. A guidelines-reviewer "the guideline itself fell short" finding was designed to lean side-task (gap, not violation), not block. The broader multi-scope placement remains an open sub-question of Q-D.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-D), NOTES_running_fixloop(9).md "Q-D veto semantics decided", fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#Q-D, fixloop_eg_dartsonline/0NN_fix_proposer.sql#council_decide config
- **relations:** deterministic council decision + hard veto; council roster expansion vision; learning layer (guideline-gap side-task)
- **verify-later:** grep/inspect hard_veto_from: ["guardian"]; council_decide

### FIX-017 — Revise loop (F2.2)
- **status:** deployed
- **status-evidence:** "F2.2 REVISE LOOP — ✅ BUILT 2026-07-10" (PLAN_fixloop_pilot.md).
- **what:** On a revise decision with rounds remaining, feeds the diagnosis, prior plan, and both reviewers' objections back into a repropose step, then re-validates and re-reviews (capped, default 2, later 3). Exhausting the cap becomes a distinct terminal state exhausted. Round counting is scoped per orchestration_id (proposer run), not per correlation_id — an earlier per-correlation design flaw was caught mid-implementation and fixed, then found still live on the deployed binary for one further round (see round-counting scope bug).
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#council_decide/check_revise, fixloop_eg_dartsonline/HANDOFF_turn21_2026-07-10.md#Session 20 Summary
- **relations:** decision router (F2.3); round-counting scope bug
- **verify-later:** grep/inspect revise; repropose; exhausted

### FIX-018 — Decision router (F2.3)
- **status:** deployed
- **status-evidence:** "F2.3 DECISION ROUTER + VERIFY + REFRAME + ESCALATION — ✅ CODE BUILT 2026-07-10... PROVEN live" v1.0.1108 (PLAN_fixloop_pilot.md).
- **what:** Replaces the single revise/complete branch with a full router: approved→complete; revise with rounds left→verify checks→repropose; rejected first time with rounds left→reframe-once; rejected again or exhausted→escalate. Motivated by two clean benchmark runs exposing two dead-ends. Flags are computed by a pure, directly-tested Go function applyCouncilCaps.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#check_approved/check_rejected/check_reframe/check_revise, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 22
- **relations:** revise loop; verify step; reframe step; escalation artifact
- **verify-later:** grep/inspect approved; revise; rejected

### FIX-019 — Verify step (diagnose_run_checks)
- **status:** deployed
- **status-evidence:** "Verify step ran 7 reviewer checks under the containment" (NOTES(10)#Turn 23).
- **what:** Reviewers attach checks:[{sql,why}] (SELECT/WITH only) to their verdicts; this action executes them under the same read-only containment the diagnosis loop's data_requests use (lint → READ ONLY transaction → statement_timeout → EXPLAIN gate → capped rows), and feeds results into the next repropose so fact-shaped objections are settled with evidence instead of another blind revision round. Capped at 8 checks by default.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#run_checks, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 22, #Turn 23
- **relations:** decision router; schema hint for reviewers
- **verify-later:** grep/inspect checks:[{sql,why}]

### FIX-020 — Schema hint for reviewers (F2.3b(a))
- **status:** deployed
- **status-evidence:** "Verification run 1e221fb7: 8 checks run, 0 failures (prior run: 5 of 7 failed on hallucinated schema)" (NOTES(10)#Turn 24).
- **what:** A load_schema_hint query_database step pulls the live table/column list from information_schema at run time, and both reviewer prompts get this hint plus two named traps (workflow steps live in agent_definitions.default_config jsonb, not a steps table; a site's domain lives on sites joined via pages.site_id). Fixes a discovered defect where 5 of 7 reviewer-written verification SQL checks failed on hallucinated columns/tables.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#load_schema_hint, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 23, #Turn 24
- **relations:** verify step; two-reviewer council
- **verify-later:** grep/inspect load_schema_hint; information_schema; agent_definitions.default_config

### FIX-021 — Reframe step (post-veto)
- **status:** partial
- **status-evidence:** "Reframe path is unit-tested but has never fired live (no veto since v4)" (HANDOFF_CURRENT_fixloop.md#State snapshot).
- **what:** After a guardian veto with rounds remaining, makes one attempt to reframe rather than reproposing the same shape: either a strictly narrower remediation (site-scoped interim fix allowed only if risks names the deferred structural fix) or an explicit "needs architecture review" declaration plus a minimal safe interim step. Capped at one attempt; built and unit-tested but has not fired live since v4 shipped.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#reframe, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 22, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md
- **relations:** decision router; escalation artifact; guardian veto surfacing architecture-level fix
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

### FIX-022 — Escalation as first-class success terminal (diagnose_escalate)
- **status:** deployed
- **status-evidence:** "Escalation package persisted (kind=escalation)... The dead-end is now a hand-off" (NOTES(10)#Turn 23); design principle stated earlier: "'this is beyond my mandate' is a correct output, packaged for you" (README_02).
- **what:** When a plan is rejected-again or a revise budget is exhausted, the run persists a kind='escalation' artifact (decision, reason, round, diagnosis conclusion, final plan, both reviews) and completes via a distinct complete_escalated success terminal. Explicitly designed so "needs a human/architecture review" is a correct, successful output rather than a failure — the organisational analogue of UNVERIFIABLE-beats-guessing. The council produced exactly this on the test bug three times before the design was formalised.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#escalate/complete_escalated, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 22, README_02_evidence_backed_proposals.md
- **relations:** decision router; reframe step; dartsonline guides pilot selection history
- **verify-later:** grep/inspect kind='escalation'; complete_escalated

### FIX-023 — Write step / fix-implementer agent (F1.1b(c))
- **status:** deployed
- **status-evidence:** "F1.1b(c): branch + PR — ✅ COMPLETE & PROVEN 2026-07-13 (PR #1 opened & merged)" (PLAN_fixloop_pilot.md); originally decided 2026-07-07 as "Q-C separate fixer agent (isolated write token; constrained edit plan; gofmt+build in a spawned job pre-PR)"; by 2026-07-12 "F1.1b(c) is code complete (b367a602)... validated as far as it can be without the deploys" (README_overview.md).
- **what:** The loop's write organ, evolved from design (Q-C, 2026-07-07) through code-complete-pending-deploy (2026-07-12) to proven live (2026-07-13). Given a fix_correlation_id, refuses anything whose latest council decision is not approved; reads current file bodies via the GitHub contents API (a modify-file 404 is a hard error — a hallucination-by-construction guard); runs an LLM step (sketch_to_files) to turn the approved plan's sketches into complete new file bodies for ONLY the plan's named files; passes those through a deterministic hard allowlist; creates a fix/<short-corr> branch and commits via the git-adapter; gates on a build check; and on green opens a PR. config_change edits are deliberately NOT implemented by this agent — left for a human.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §3b F1.1b(c), fixloop_eg_dartsonline/README_so_far.md, README_overview.md, docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-C)
- **relations:** hard file allowlist; build gate; git-adapter write isolation; fix-implementer-orchestrator; PR as human terminal
- **verify-later:** grep/inspect fix_correlation_id; approved; sketch_to_files

### FIX-024 — Hard file allowlist (diagnose_prepare_fix_commit)
- **status:** deployed
- **status-evidence:** "Part 2a BUILT (chassis, commit a4c6cc63)... 7-case suite exercises the real logic" (NOTES(10)#Turn 25).
- **what:** A deterministic action sitting between the implementer's LLM step and the git-adapter: the approved plan's modify/add file list is a hard allowlist — a produced file outside the plan, a plan file the implementation is missing, or an empty/duplicate/no-op file all reject the whole implementation before anything touches git. Also assembles the branch name, commit message, and PR title/body (the "Q-H package"). This is the safety core that made the first live PR's diff exactly the approved plan.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql#prepare step, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 25, fixloop_eg_dartsonline/README_so_far.md
- **relations:** write step; build gate; plan validation
- **verify-later:** validateImplementation function; 7-case test suite

### FIX-025 — Build gate (diagnose_build_gate)
- **status:** deployed
- **status-evidence:** "Build gate (golang Job): GREEN — === build gate: PASS ===" (README_so_far.md); "its first red correctly blocked a PR" (NOTES(10)#Turn 26-28).
- **what:** Before any PR is opened, changes must be built in a clean container (gofmt + targeted go build) in a short-lived golang-image k8s Job. Green routes to PR creation; red routes to a no-PR terminal with build log attached, branch left for human inspection — "no PRs for broken code." Chosen over GitHub Actions CI on the PR (Option B: broken implementations must never even become a visible PR). Its first live red catch was a genuine pre-existing bug, then fixed for real.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql#build_gate/check_gate, fixloop_eg_dartsonline/SUMMARY_write_step_position_2026-07-12.md#Option A/B/C, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 26-28
- **relations:** write step; hard file allowlist; PR as human terminal; hard deterministic gates between every LLM step
- **verify-later:** diagnose_build_gate action; RBAC rbac-job-spawner.yaml pods/log grant

### FIX-026 — git_adapter_request generic adapter caller
- **status:** deployed
- **status-evidence:** "git_adapter_request — ONE generic adapter caller (allowlisted verbs...)" (HANDOFF_CURRENT_fixloop.md#F1.1b(c) CODE COMPLETE).
- **what:** A single generic workflow action used for all git-adapter calls from the write step, with the adapter action name and data fields/literals supplied per-step from config, and an explicit note that delete_repo is unreachable through this path — verbs are allowlisted to commit/create_branch/create_pull_request.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql#create_branch/commit_files/create_pr config, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md
- **relations:** git-adapter new actions; write step
- **verify-later:** grep/inspect delete_repo

### FIX-027 — isRepoCloningAgent spawn gate / GITHUB_READ_TOKEN injection
- **status:** deployed
- **status-evidence:** "the spawned implementer pod gets GITHUB_READ_TOKEN via the already-deployed isRepoCloningAgent gate" (HANDOFF_CURRENT_fixloop.md).
- **what:** An existing spawn-gate mechanism (already used for diagnose-agent) that injects a read-only GitHub token into a dedicated, ephemeral pod when the spawned agent type is listed in isRepoCloningAgent. fix-implementer was added to this list. Only works when the agent runs as a dedicated spawned pod — the generic in-chassis orchestrate path bypasses the gate entirely.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql#header point 1, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#FIRST END-TO-END RUN blocked
- **relations:** fix-implementer-orchestrator; write step; diagnose_read_repo_files action
- **verify-later:** grep/inspect isRepoCloningAgent; fix-implementer

### FIX-028 — diagnose_read_repo_files action
- **status:** deployed
- **status-evidence:** "diagnose_read_repo_files — plan's modify/add files via GitHub contents API (raw media type; read token from spawn gate; modify-404 = hard error)" (HANDOFF_CURRENT_fixloop.md).
- **what:** Fetches the current bodies of the approved plan's modify/add files via the GitHub contents API at an explicit ref, using the token from the spawn gate. A missing file for a "modify" operation is a hard error.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql#read_current_files, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md
- **relations:** isRepoCloningAgent spawn gate; write step
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

### FIX-029 — fix-implementer-orchestrator (dedicated-pod wrapper)
- **status:** deployed
- **status-evidence:** "0NN_fix_implementer_orchestrator.sql — F1.1b(c) fix: run the implementer as a DEDICATED POD" (header).
- **what:** A thin wrapper agent (spawn_agent(fix-implementer) → call_agent → complete) built to fix a real first-run failure: firing fix-implementer via the generic orchestrate path ran it IN the shared chassis pod, so the isRepoCloningAgent gate never fired. Mirrors the existing diagnose-orchestrator→diagnose-agent pattern exactly. Needed no image rebuild.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer_orchestrator.sql, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#FIRST END-TO-END RUN, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 26-28
- **relations:** isRepoCloningAgent spawn gate; write step
- **verify-later:** 092_TRIGGER_fix_implementer_v1.sh target agent type

### FIX-030 — Whole-file rewrite strategy (implementer's LLM step)
- **status:** deployed
- **status-evidence:** "41KB whole-file rewrite → allowlist PASS" (README_so_far.md).
- **what:** The implementer's sketch_to_files LLM step outputs the COMPLETE new body of every plan-named file, never a diff/patch, with hard rules forbidding drive-by changes ("the diff a human reviews must contain ONLY the plan"). Explicitly named as not scaling to very large files (32000 max_tokens gives headroom for one ~41KB file but not much more) — a diff/patch strategy is logged as future work (F1.2).
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql#sketch_to_files, fixloop_eg_dartsonline/README_so_far.md, README_overview.md
- **relations:** write step; hard file allowlist; F1.2 deferred work items
- **verify-later:** grep/inspect sketch_to_files

### FIX-031 — PR as human terminal / nothing merges itself
- **status:** convention
- **status-evidence:** "PR — waits for you. Nothing merges itself." (README_so_far.md); "one structural commitment: the human gate never moves. More autonomy upstream … never past the PR." (README_02_evidence_backed_proposals.md).
- **stage2-verified (2026-07-14):** deployed → convention — A governing design principle ('PR is the human terminal'), not itself a checkable artifact; correctly consistent with confirmed absence of any auto-merge action in the codebase.
- **what:** A governing design principle running through the whole fix-loop: autonomy may widen upstream (diagnose, plan, revise, commit-to-branch), but the merge is permanently human — the PR is the fixed boundary of machine authority, simpler and harder than the graduated trust machinery elsewhere, and orthogonal to it. Isolation model (2026-07-12): fix/* branches live on the same repo (no fork); the owner alone chooses what merges to main. This is why "escalation" is treated as a success, not a failure.
- **sources:** fixloop_eg_dartsonline/README_so_far.md, fixloop_eg_dartsonline/0NN_fix_implementer.sql#header, fixloop_eg_dartsonline/SUMMARY_where_we_are_2026-07-13.md, README_02_evidence_backed_proposals.md#2, README_overview.md
- **relations:** write step; fork isolation / NO FORK decision; escalation as first-class success terminal
- **verify-later:** absence of any auto-merge path in the write step

### FIX-032 — Fork isolation / NO FORK decision (superseded)
- **status:** superseded
- **status-evidence:** "Decisions CLOSED... 4. NO FORK: isolation = fix/* branches + owner-gated merges on this repo" (HANDOFF_CURRENT_fixloop.md, 2026-07-12) — closes a proposal raised earlier the same week: "the strongest isolation is that the write surface points only at the fork" (README_02_evidence_backed_proposals.md#5).
- **what:** A design proposal considered pointing the loop's git-adapter credential, intake defaults and corpus indexing at a fork of the repo, making the main repo physically unwritable by the loop rather than protected by review discipline (with the fork's constitution/mission docs becoming the councils' curated context). The owner raised, then explicitly closed this idea; the decision landed instead on branch+PR isolation on the same repo, which is what was actually built.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 25, #Turn 26, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#Decisions CLOSED, README_02_evidence_backed_proposals.md#5
- **relations:** PR as human terminal; write step; external rollback (residual risk framing, autonomy-governance)
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

### FIX-033 — Round-counting scope bug (correlation vs orchestration)
- **status:** superseded (fixed in source; deploy gap tracked separately)
- **status-evidence:** "the deployed v1.0.1107 binary counts council rounds per correlation... does NOT carry the orchestration_id-scoping fix" (NOTES(10)#Turn 22).
- **what:** Council-round counting was originally scoped per correlation_id, accumulating council_report rows across every proposer re-run — so a fresh proposer run on a correlation with review history would start mid-count and exhaust its revise budget without ever reproposing. Fixed in source to count per orchestration_id, but a same-tag deploy trap meant the fix did not reach the running binary for one further benchmark cycle.
- **sources:** fixloop_eg_dartsonline/HANDOFF_turn21_2026-07-10.md#Key accomplishments, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 22
- **relations:** revise loop; same-tag deploy trap gotcha
- **verify-later:** grep/inspect correlation_id; orchestration_id

### FIX-034 — fixloop-digest / awareness surface
- **status:** deployed
- **status-evidence:** "IN FLIGHT (turn 29): the AWARENESS SURFACE... Built + committed, awaiting the next chassis image" (HANDOFF_CURRENT_fixloop.md) — realizing an earlier proposal: "the missing organ is a push surface: a periodic digest … before autonomy widens … the awareness surface gets built first" (README_02_evidence_backed_proposals.md#4).
- **stage2-verified (2026-07-14):** partial → deployed — platform/orchestration/actions/fixloop_digest_action.go + registry.go:1180 register 'fixloop_digest'; git log shows f95004aaf (v1.0.1117) postdates the doc's 'awaiting next chassis image' snapshot (v1.0.1114/1113); docs/fixloop_digests/DIGEST_latest.md + archive/DIGEST_2026-07-13.md exist as delivered output; README...
- **2026-07-16 addition (Phase 4 — escalation section):** live since v1.0.1120/1121 (commits d827d6334, 2887247b2, "Phase 4 LIVE — digest escalation channel delivered (DIGEST 2026-07-15)"). Extends this same digest with `digestGatherImmune` (`fixloop_digest_action.go:291`, called from `renderDigest` at line 378) to also surface: sweep counts, the **entire** open diagnosis queue on every digest (so parked items — decisions waiting on the owner — never silently fade out of view, and new ones are flagged), silent-check findings (FIX-052) both open and closed, and standing capability gaps (from the triage router's `triageGatherCapabilityGaps`, FIX-051). This is the piece that answers "how does the owner see the whole triage/escalation layer at a glance," not just fix-loop run outcomes — the original 2026-07-14 verification below only covered the narrower base digest.
- **what:** A deterministic (no-LLM-in-path) digest agent composing a window (default 24h) summary of fix-loop activity — status/terminal/gate/PR outcomes, decisions per correlation, and agent_definitions_backup snapshots — persisted to doc_notes (categories ["digest","fixloop"]). Built to satisfy the owner's standing rule "more awareness before wider autonomy" and to be the grown-up form of the parked F0.3 per-iteration notes. v1 is manual-trigger only; a daily cadence is deliberately deferred; as of the source snapshot it awaits the next chassis image before going live.
- **sources:** fixloop_eg_dartsonline/0NN_fixloop_digest.sql, fixloop_eg_dartsonline/093_TRIGGER_fixloop_digest_v1.sh, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#IN FLIGHT, README_02_evidence_backed_proposals.md#4, fixloop_eg_dartsonline/SUMMARY_where_we_are_2026-07-16.md#Where we are, platform/orchestration/actions/fixloop_digest_action.go:291,378
- **relations:** owner standing rule: awareness before autonomy; diagnosis_artifacts table; council roster expansion vision; F0.3 per-iteration notes; triage router (FIX-051); silent-check verifier (FIX-052)
- **verify-later:** whether the chassis image carrying fixloop_digest action has shipped; doc_notes rows with categories ? 'digest'; digestGatherImmune output on a live DIGEST file

### FIX-035 — Owner standing rule: awareness before autonomy
- **status:** deployed
- **status-evidence:** "Owner standing rule (2026-07-12): 'more awareness BEFORE wider autonomy.'" (0NN_fixloop_digest.sql header).
- **what:** An explicit governance principle: before the council is widened with more reviewer perspectives or migration/feature-building agents, the owner must first have a reliable way to see what the loop has been doing and deciding. Directly produced the fixloop-digest slice being scheduled ahead of the F2 roster expansion. Named risk: not wrong action but unknown action — drift compounding silently while trails exist only pull-side.
- **sources:** fixloop_eg_dartsonline/0NN_fixloop_digest.sql#header, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#Decisions CLOSED, fixloop_eg_dartsonline/SUMMARY_where_we_are_2026-07-13.md, README_02_evidence_backed_proposals.md#4
- **relations:** fixloop-digest / awareness surface; council roster expansion vision (deferred by this rule)
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

### FIX-036 — Council roster expansion vision (guidelines/reuse/bug-historian/compliance/pipeline-guardians)
- **status:** aspirational
- **status-evidence:** "Initial roster: a guidelines agent... a reuse agent... a bug-historian... a compliance/legal eye; pipeline guardians... specialist knowledge agents" (RUNBOOK(9)#THE TASK) — none beyond the guardian/edit-quality pair were built; "the roster of 2 (edit-quality, guardian) is explicitly a skeleton... Adding a reviewer is a seed change + prompt + curated context" (README_02).
- **what:** The original council vision named a much wider roster than what was built: a guidelines agent (adherence to 000-0xx, or did the guideline fall short), a reuse agent (code and docs), a bug-historian (has this class recurred), a compliance/legal eye, one pipeline-guardian per master workflow (seeded from the builder relay map), and specialist knowledge agents — motivated by a real incident where a chat reinvented a trigger+triage SQL pair that already existed. Only a generic edit-quality reviewer and a single cross-pipeline guardian shipped (see two-reviewer council); reviewer areas are expected to correlate with the docs024 documentation categories, the direct bridge to this concept register's own council-agent goal.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#THE TASK, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#Direction, docs019/RUNBOOK_diagnosis_fix_loop(9).md#the-task (item 3), README_02_evidence_backed_proposals.md#3, README_comprehensive_documentation_categorisation.md
- **relations:** two-reviewer council; owner standing rule: awareness before autonomy; architecture-change visibility; hard-veto flag at multiple scopes
- **verify-later:** whether any bench agents beyond edit-quality/guardian were seeded

### FIX-037 — Architecture-change visibility (Q-E signals / detector)
- **status:** partial
- **status-evidence:** "Q-E architecture-change signals: ... STILL OPEN" (RUNBOOK(10)#STILL OPEN); originally "packages touched breadth; platform/ vs actions/; exported-signature diffs vs the corpus; message/topic/schema/contract changes; migration presence. Which are load-bearing?" (docs019 Q-E, open, F2-phase).
- **what:** A standalone goal from the original task charter — make it loud when a proposed change is accidentally fundamental (touching platform contracts, message shapes, many packages, exported signatures) before it ships, running as one council reviewer. Never built as a dedicated formal detector; what exists in practice is the pipeline-guardian reviewer's informal judgement, which has correctly identified architecture-level changes dressed as contained fixes (see the concrete dartsonline instance).
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#THE TASK, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#STILL OPEN Q-E, docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-E)
- **relations:** two-reviewer council; guardian veto surfacing architecture-level fix; council roster expansion vision
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

### FIX-038 — Guardian veto surfacing an architecture-level fix (dartsonline)
- **status:** deployed
- **status-evidence:** "the guardian vetoed all three edits as 'an architecture change dressed as a contained fix'" (NOTES(10)#Turn 22, orch 8c770fd5).
- **what:** A concrete, live-observed instance of the guardian reviewer correctly recognizing that a minimal-looking three-edit plan was actually architecture-level, vetoing it and proposing a safer alternative in its notes — deliberately NOT auto-applied since it fixes only one site while leaving the platform-wide cause live everywhere.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 22
- **relations:** architecture-change visibility; reframe step; platform-not-site-data fix philosophy; dartsonline guides pilot selection history
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

### FIX-039 — Platform-not-site-data fix philosophy
- **status:** deployed
- **status-evidence:** "Owner ruled: the F1 edit plan targets the PLATFORM, not dartsonline's data." (NOTES(10)#Turn 2).
- **what:** An owner ruling that any fix plan must target the platform mechanism rather than a single site's data rows — because the causes of the benchmark bug are relay-level and a data-only fix would fix one site while leaving every other site exposed. Directly shapes the proposer's prompt rules and the guardian's refusal to accept a scoped data-only remediation as a final answer.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 2, fixloop_eg_dartsonline/0NN_fix_proposer.sql#propose prompt rule 1
- **relations:** dartsonline guides pilot selection history; guardian veto surfacing architecture-level fix; reframe step
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

### FIX-040 — config_change edit operation type
- **status:** deployed
- **status-evidence:** "config_change edits in a plan are NOT implemented by this agent... the PR body carries them for the human" (0NN_fix_implementer.sql header).
- **what:** A plan-edit operation type reserved for edits that target agent_definitions workflow-JSON configuration rather than repo files. The proposer's prompt requires such edits be explicitly labelled, but the fix-implementer deliberately does not apply them — they are left in the PR body for a human to apply by hand.
- **sources:** fixloop_eg_dartsonline/0NN_fix_proposer.sql#propose prompt rule 5, fixloop_eg_dartsonline/0NN_fix_implementer.sql#header
- **relations:** write step; hard file allowlist; plan validation
- **verify-later:** grep/inspect agent_definitions

### FIX-041 — F1.2 deferred work items (ref/base as input; fix_pr artifact; diff strategy)
- **status:** aspirational
- **status-evidence:** "Open (F1.2): ref/base are live-set to the active working branch because origin/main is stale — make them a per-run INPUT" (PLAN_fixloop_pilot.md).
- **what:** A cluster of known-but-deferred improvements: the implementer's git ref/base/from_branch are hardcoded via a live jsonb_set patch rather than a per-run input field; a dedicated kind='fix_pr' diagnosis_artifacts row for the PR result is deferred; a diff/patch implementation strategy for large files (the whole-file rewrite strategy doesn't scale beyond ~41KB).
- **sources:** fixloop_eg_dartsonline/PLAN_fixloop_pilot.md#F1.1b(c), fixloop_eg_dartsonline/0NN_fix_implementer.sql#header
- **relations:** whole-file rewrite strategy; write step
- **verify-later:** grep/inspect kind='fix_pr'

### FIX-042 — F3 learning layer: bug_records taxonomy + guideline-amendment side-tasks (never built)
- **status:** aspirational
- **status-evidence:** "F3 — Learning... bug_records (category taxonomy, recurrence checks feeding the historian)" (RUNBOOK(10)#Phased plan) — no bug_records table or historian agent found; the guideline-gap side-task specifically: "Q-D completion — guideline-gap = SIDE-TASK... a work item carrying the evidence; handler drafts a concrete amendment and opens a PR against the GUIDELINE DOCS" (NOTES(9)#DECISIONS) — no implementation of this side-task handler was found in the files read, so whether any piece of F3 was built remains genuinely unconfirmed rather than definitively absent.
- **stage2-verified (2026-07-14):** unknown → aspirational — grep -rn 'bug_records' across .go/.sql: 0 hits. grep -rn 'guideline.amendment|guideline_gap|guideline-gap' across .go/.sql: 0 hits. Both the taxonomy table and the guideline-amendment side-task handler are confirmed absent repo-wide, resolving the doc's 'unknown/genuinely unconfirmed' to a concrete aspirational/neve...
- **what:** The original phased plan's final stage: categorize confirmed bugs into a taxonomy (bug_records) so recurring classes are caught earlier and feed the bug-historian reviewer, feed guideline-amendment proposals to the human as a side-task that doesn't block the fix (human-terminal PR against the guideline docs), and enrich the corpus from what the loop learns. Decided in outline on 2026-07-07 (Q-D) but never designed in detail or built as far as any of the source documents show.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Phased plan F3, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#Phased plan F3, fixloop_eg_dartsonline/NOTES_running_fixloop(9).md#DECISIONS, docs019/RUNBOOK_diagnosis_fix_loop(9).md#phased-plan (F3)
- **relations:** council roster expansion vision (bug-historian); hard-veto flag at multiple scopes
- **verify-later:** search for a guideline-amendment work item type / handler agent in agent_definitions; bug_records table (absent)

### FIX-043 — Q-G reviewer context (open design question, v1 answered narrowly)
- **status:** partial
- **status-evidence:** "Q-G v1 = role prompts + plan + diagnosis (no per-reviewer corpora yet)" (PLAN_fixloop_pilot.md).
- **what:** The open question of how much context each council reviewer should see. What shipped is the narrowest option: both reviewers get the same role prompt, the persisted plan, the diagnosis conclusion, and (from F2.3b) a live schema hint — no per-reviewer curated corpus exists yet.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#STILL OPEN Q-G, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §F2.1
- **relations:** two-reviewer council; schema hint for reviewers; council roster expansion vision
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

### FIX-044 — Q-H human-facing result package
- **status:** deployed (v1)
- **status-evidence:** "PR carrying the Q-H package" appears repeatedly as delivered (HANDOFF_CURRENT_fixloop.md).
- **what:** The decided shape of what a human ultimately sees: the PR body carries the diagnosis conclusion, the approved plan, and the council's decision/reviews together, so a human reviewing a fix-loop PR never has to go hunting through diagnosis_artifacts. The equivalent package for an escalated run is the escalation artifact.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql#prepare/create_pr, fixloop_eg_dartsonline/README_so_far.md
- **relations:** write step; escalation as first-class success terminal
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

### FIX-045 — SEED_first_writestep_diagnosis pattern / seeded-bug strategy
- **status:** deployed
- **status-evidence:** "the diagnosis is hand-written (the diagnosis LOOP is separately proven); proposer, council, and implementer that consume it all run for real" (SEED_first_writestep_diagnosis.sql header) — executing an earlier recommendation: "the system's first-ever PR will have earned every gate it passed" (README_02 §6).
- **what:** A reusable technique for exercising downstream stages honestly without waiting for a live CONFIRMED diagnosis on a suitable bug: hand-author a CONFIRMED orchestration_states row for a real, tiny, zero-risk defect, then run the real proposer→council→implementer chain against it for real. Fabricating evidence rows was explicitly rejected as an option, as was hand-approving a known-flawed plan (contradicts the reviewers) or waiting for an organic small bug (unbounded).
- **sources:** fixloop_eg_dartsonline/SEED_first_writestep_diagnosis.sql, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#First-run design decision, README_02_evidence_backed_proposals.md#6, README_overview.md
- **relations:** write step; tier-coverage guard
- **verify-later:** grep/inspect orchestration_states

### FIX-046 — F0.3 per-iteration notes / per-task running notes via doc_notes
- **status:** partial
- **status-evidence:** "Per-iteration notes — NOT MET, because F0.3 does not exist yet" (RUNBOOK(10)#F0 plumbing criteria) — following an earlier direction-setting decision: "Q-F DIRECTION SET (2026-07-07): REUSE doc_notes... the diagnose-agent workflow is ALREADY rewired by them: emit → persist_note → complete" (docs019 Q-F).
- **what:** One of F0's four original acceptance criteria — writing the loop's per-iteration/per-step reasoning into task-specific running notes — was designed (reuse the tools chat's doc_plans/doc_notes infrastructure, category convention diagnosis) but never fully implemented: the terminal-diagnosis note is wired via persist_note with a strict no-guessing subject gate, but per-iteration rows remained pending the owning thread's sign-off and were still NOT MET as of the later docs024 dossier. The diagnosis_artifacts.kind='iteration_note' column value exists specifically to carry this, unused.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Phased plan F0.3, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md, docs019/RUNBOOK_diagnosis_fix_loop(9).md#questions (Q-F)
- **relations:** doc_notes/travelling-docs integration boundary; diagnosis_artifacts table; fixloop-digest / awareness surface
- **verify-later:** grep/inspect diagnosis_artifacts.kind='iteration_note'; doc_notes rows with category diagnosis

### FIX-047 — Loop-worthiness test doctrine (five-criteria intake test)
- **status:** convention
- **status-evidence:** "LOOP-WORTHINESS TEST (doctrine — apply before every intake)" — applied three times in the same file (pilot #1 downgraded, candidate 2 forked, guides pilot confirmed) (docs019 RUNBOOK(9)); "Owner asked whether the loop fits the dartsonline quality problem. Answer: decomposed via the new LOOP-WORTHINESS TEST" (NOTES_running_fixloop(9).md).
- **stage2-verified (2026-07-14):** deployed → convention — A five-criterion intake doctrine described in docs — explicitly 'n/a (doctrine)' per its own verify-later field; process, not code.
- **what:** A pre-registered five-criterion test for whether a candidate bug is worth running the diagnosis/fix loop on: (1) a genuine SYMPTOM about system behaviour, not a disguised feature request; (2) a causal mechanism plausibly exists in code+data+runtime; (3) not answerable by one or two direct queries (mandatory cheap pre-check first); (4) bounded to a single coherent symptom; (5) verified CURRENT at intake, since symptoms are perishable (added after a pilot candidate "evaporated," fixed live before the loop ran — this happened twice in the founding thread alone). Feature absences route to build queues; quality judgements to council/auditors; one-query questions to the query itself.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#loop-worthiness, docs019/RUNBOOK_diagnosis_fix_loop(9).md#previous-pilot-1, NOTES_running_fixloop(9).md multiple 2026-07-07 pilot-selection entries
- **relations:** dartsonline guides pilot selection history; known-answer benchmark methodology
- **verify-later:** n/a (doctrine)

### FIX-048 — Hard deterministic gates between every LLM step
- **status:** convention
- **status-evidence:** README_02 lists them as built pattern: "CONFIRMED gate, plan validator, file allowlist. The models propose; plain Go code decides what proceeds."
- **stage2-verified (2026-07-14):** deployed → convention — A cross-cutting design pattern description ('gates between every LLM step'); underlying individual gates (build gate, council decide, plan validator) independently confirmed as code, but this concept itself is the pattern/principle, not a single artifact.
- **what:** No LLM output passes into consequence unchecked: the diagnosis must be CONFIRMED (gate), the plan must validate, the files must be on a deterministic allowlist, the build must pass, before anything advances. Complexity and authority live in plain Go; the models only propose. The same shape as keeping convergence guards in the engine rather than in workflow conditionals.
- **sources:** README_02_evidence_backed_proposals.md#1, README_overview.md
- **relations:** deterministic council decision + hard veto; write step; hard file allowlist; build gate
- **verify-later:** the gate implementations in the fixloop actions

### FIX-049 — Fix-loop value proposition: unattended, cited, consistent
- **status:** convention
- **status-evidence:** README_02: "The value proposition (decided 2026-07-09): not 'the loop finds what humans can't' … The proposition is unattended, cited, consistent — the 3am diagnosis with a paper trail."
- **stage2-verified (2026-07-14):** deployed → convention — A recorded design decision/value proposition narrative; no code artifact claimed.
- **what:** A recorded decision reframing what the loop is for: on this platform bugs are legible to anyone with schema access and patience, so the differentiation is not superhuman insight but unattended operation with citations and consistency — a package instead of a hunch, reconstructible after the fact by one correlation id. Every design choice flows from it.
- **sources:** README_02_evidence_backed_proposals.md#2
- **relations:** escalation as first-class success terminal; fixloop-digest / awareness surface; diagnosis_artifacts table
- **verify-later:** decision record in NOTES_running_fixloop

### FIX-050 — Transferable machinery: legacy-migration and feature intakes
- **status:** aspirational
- **status-evidence:** README_02 §3: migration agents "not built, but it's the same machinery with a different intake"; features from specs "honestly furthest away … plausible; not designed".
- **what:** The allowlist/gate/council scaffolding is intake-agnostic: a legacy migration is "pattern X supersedes pattern Y" (scanner finds Y-shaped code, proposer writes constrained plans, council reviews, PRs flow); feature-building from mission docs needs a new grounding tier ("cite the spec clause this serves") — same shape as causal citation but not designed.
- **sources:** README_02_evidence_backed_proposals.md#3
- **relations:** council roster expansion vision; hard deterministic gates between every LLM step
- **verify-later:** n/a (unbuilt)

### FIX-051 — Triage router (Phase 1): deterministic failure sorter (2026-07-16 addition)
- **status:** deployed
- **status-evidence:** Live since v1.0.1117 (commit f95004aaf, "triage LIVE on v1.0.1117 — channel closed, dedup proven"); `SUMMARY_where_we_are_2026-07-16.md`: "Its first live run confirmed the value: ~half of all 'failures' were operational noise it correctly kept out." Independently confirmed 2026-07-16: `platform/orchestration/actions/diagnose_triage_action.go` (526 lines) exists, registered as `diagnose_triage` (`registry.go:1198`).
- **what:** A deterministic router — no LLM in the classification path — that reads every recorded failure across the fleet and sorts it four ways: genuine code bugs escalate to the diagnosis queue (deduped by pattern via `triageItemKey`, hard-capped per sweep, inserted by `triageInsertNeedsDiagnosis`); operational blips (timeouts, dead pods) get re-queued and never reach the loop; failures with no error text are held for a human; missing-capability signals go to the roadmap (`triageGatherCapabilityGaps`), never the loop. `triageRoute` (line 63) is the classification function; `DiagnoseTriageAction` (line 115) is the entry point.
- **sources:** fixloop_eg_dartsonline/SUMMARY_where_we_are_2026-07-16.md#Where we are; commit f95004aaf; platform/orchestration/actions/diagnose_triage_action.go:63,115,327,356,383,403
- **relations:** council roster expansion vision (FIX-036); feedback close-out (FIX-053, same file); fixloop-digest (FIX-034, consumes triage's capability-gap findings)
- **verify-later:** registry.go:1198 diagnose_triage entry; scheduled_tasks cadence firing the triage sweep

### FIX-052 — Silent-check verifier (Phase 2): the class no work item ever records (2026-07-16 addition)
- **status:** deployed
- **status-evidence:** Live since v1.0.1118 (commit b2736a457, "Phase 2 silent-check LIVE on v1.0.1118 — proven end to end incl. cross-thread close-out"); `SUMMARY_where_we_are_2026-07-16.md`: "It found the darts bug on two sites and routed it through triage into the queue." Independently confirmed: `platform/orchestration/actions/diagnose_silent_check_action.go` (532 lines) exists, registered as `diagnose_silent_check` (`registry.go:1204`).
- **what:** A verification checker for the failure class no work item ever records — the "darts signature": a page referenced in a site's navigation that was never built, with nothing anywhere flagging it. Emits inert findings **only for what the immune system cannot already see** (if any existing work item references the page, it stays out — avoiding duplicate noise), and groups every affected site into one platform-level pattern so the root cause gets fixed once rather than per-site. Routes confirmed findings through the Phase 1 triage router into the diagnosis queue.
- **sources:** fixloop_eg_dartsonline/SUMMARY_where_we_are_2026-07-16.md#Where we are; commit b2736a457; platform/orchestration/actions/diagnose_silent_check_action.go, diagnose_silent_check_test.go
- **relations:** triage router (FIX-051, the downstream consumer of its findings); orphan-pages / nav-drift concepts (link-management.md)
- **verify-later:** registry.go:1204 diagnose_silent_check entry; whether it has since found any silent-failure class beyond the darts nav-page signature

### FIX-053 — Feedback close-out (Phase 3): all-time resolution recheck + auto-reescalation (2026-07-16 addition)
- **status:** deployed
- **status-evidence:** Live since v1.0.1122 (commit b869469c8, "fixloop triage Phase 3 close-out: parked escalations close when their failure pattern resolves (all-time check, never window aging); re-escalation automatic via dedup index"); `SUMMARY_where_we_are_2026-07-16.md`: "Proven both ways in production: a real sweep closed nothing (all patterns still real), and a synthetic probe closed itself while the real ones stayed open." Independently confirmed: `triageCloseResolved` at `diagnose_triage_action.go:262`.
- **what:** Each triage sweep re-checks whether a parked escalation's failure pattern still exists among currently-failed items, using an **all-time** check rather than a recency window — so a pattern that simply aged out of a lookback window is never mistaken for having resolved. `triageCloseResolved` (line 262) does the recheck; `triageResolvedKeys` (line 314) computes which parked keys no longer appear live. Closes genuinely-resolved escalations and re-escalates automatically (via the same `triageItemKey` dedup index as Phase 1) if a closed pattern returns.
- **sources:** fixloop_eg_dartsonline/SUMMARY_where_we_are_2026-07-16.md#Where we are; commit b869469c8; platform/orchestration/actions/diagnose_triage_action.go:262,314
- **relations:** triage router (FIX-051, same file/mechanism, same dedup index); the "also open" note in SUMMARY_where_we_are_2026-07-16.md that re-*queuing* (vs. just closing) the original items after a fix ships is still a deliberate human action, not automated
- **verify-later:** diagnose_triage_action.go:262-326; confirm the auto-reescalation path has fired on a real recurrence (as of 2026-07-16 only proven via synthetic probe + one real sweep finding nothing to close)

### FIX-054 — Forward-fitness council seat on the FIX lane: `review_architecture` on `fix-proposer` + `council-gate` (2026-07-28 addition)
- **status:** deployed (config-only; live immediately, no image)
- **status-evidence:** Seated 2026-07-28 ~21:55Z. Verified on the live rows: `fix-proposer` and `council-gate` both report `review_fields` **17** (was 16), `review_architecture` and `gate_architecture` present, `architecture` present in the `select_panel` footprint map, `hard_veto_from` still `["guardian"]`, `max_rounds` 3. Reachability by BFS over `then_step`/`else_step`/`error_step` (a linear `next_step` walk reports false orphans on this branching workflow): **fix-proposer 50/50, council-gate 44/44, zero orphans.** Rollback rows exist in `agent_definitions_backup` at 16 seats for both lanes (`council-gate` 21:56:37Z via the 099 mirror, `fix-proposer` 21:59:59Z). ~~**NOT yet exercised — 0 reviews as of seating; presence is not evidence it works.**~~ **DISCHARGED 2026-07-29: the seat has fired.** 12 reviews after the truncation fix (07-29 cutover 07:19:36Z), 0 truncated, 11 of 12 carrying `ARCHITECTURE_SIGNAL`, and one of its objections became `RFC_002` on owner instruction. **Its first three reviews were 2/3 TRUNCATED, which made it look like a seat that objects too much — see FIX-055 and `bugs_open/138`. Object rate 2-of-3 → 2-of-12 once it stopped being cut off; the `verify-later` note below called this exactly and is why the seat was not pulled.**
- **what:** The council's only seat that argues the **cost of not changing**, extended from design time to the fix lane. Judges four things in order: forward fitness, cost of not changing, point-fix-vs-architecture-change (the RFC trigger test), and whether the same site has been deflected upward before. **Advisory: its prompt never offers `veto`**, which is what makes it advisory — `hard_veto_from` is only an audit label. Its routing signal is the first line of `notes`, `ARCHITECTURE_SIGNAL: point_fix|needs_rfc|insufficient | DEFLECTIONS: <n|unknown>`, because the decider persists only `{reviewer, verdict, objections, missing, notes}` and discards every other field. Adapted from the live `feature-designer` seat rather than written fresh; the fix-lane copy adds one paragraph naming the failure mode it exists to catch — *a shared mechanism arriving inside a bug patch* — and the 099 mirror swaps the diagnosis context for the submitter's rationale on the gate. Footprint deliberately broad (`platform/`, `internal/`, `pkg/`, `cmd/`, `.sql`, `migration`, `_action.go`, `coordinator`, `contract`, `namespace`, …) per `select_review_panel_action.go:34-38`: a seat with no footprint fails **open**, so the real choice was broad-vs-always-on, not broad-vs-narrow.
- **why it exists:** `bugs_closed/124` and `bugs_closed/129` both drew guardian SCOPE vetoes within two days, and both were told to "route the seam to architecture review" — **a venue reachable from neither lane they arrived on.** Owner reversal of decision D9 (2026-07-28), triggered by D9's own reversal condition, which had been recorded `[UNMEASURED]` and never run: **13 affirmative architecture-level escalations across 10 distinct submissions**, against a threshold of 3.
- **sources:** architecture_review/DECISIONS_open_for_owner_2026-07-26_architecture_seat.md §D9; bugs_closed/129_HANDOFF_2026-07-28...md; docs024_key_docs_latest/WRONG_CALLS.md (2026-07-28, the vacuous trigger query); platform/orchestration/actions/select_review_panel_action.go:34-38; fixloop_eg_dartsonline/099_SYNC_gate_roster.py
- **relations:** the design-time instance on `feature-designer` (same seat, different placement — a duplicate of PLACEMENT, not of REMIT); the 099 gate-roster mirror (how it reached `council-gate`); the guardian seat it counterweights; `bugs_closed/124` + `bugs_closed/129` (the two vetoes that proved the routing hole); the architecture-review RFC track (`PROCESS_architecture_review.md`), which is where a `needs_rfc` signal is meant to land
- **verify-later:** **The gate has NO `code_lookup`** (verified: `fix-proposer` t, `council-gate` **f**, `feature-designer` t), so on the gate this seat's `code_checks` are raised and never answered — deliberate per `099_SYNC_gate_roster.py:24-29`, but it leaves the seat half-sighted on the very lane 124 and 129 arrived on. Open question, not an oversight. Also: read the `ARCHITECTURE_SIGNAL` line of its first real reviews rather than counting verdicts — the kill switch is a high object-rate with no signal line.

### FIX-055 — Truncation-gate attribution: `gated_by_truncation` on every council report (2026-07-29 addition)
- **status:** built, shipped to HEAD, **NOT yet live** (Go — inert until the next chassis image is rolled). Committed `3a59b5012`.
- **status-evidence:** `platform/orchestration/actions/diagnose_council_decide_action.go` — `hasGatingObjection`, `gatesOnlyBecauseTruncated`, `decideCouncil` third return value; tests `TestGatesOnlyBecauseTruncated` and 7 new `TestDecideCouncil` cases, all passing against `git archive HEAD` + the change (the working tree would not compile — another session's WIP in `datahelpers`). **Behaviour measured, not asserted:** the new rule replayed over 14 days of stored `reviews[]` gives 63 gated revise rounds → 10 would now read TRUNCATED, 3 mixed, 50 unchanged, and exactly **1** round changes which seat it names. Query: `bugfix_138_degraded_gates/RUNBOOK_degraded_gates.md` §2.
- **what:** A council round that returns `revise` now says **why it gated**. `decided_by` distinguishes `gating objection from X` (a judgement) from `gating TRUNCATED objection from X — cut off at max_tokens …` (a token-budget overrun), and a boolean `gated_by_truncation` is persisted on **every** `council_report` artifact — in `body` (full record) and in `metadata` (jsonb, so a rate alert is an indexed one-liner needing no cast and no `jsonb_array_elements`). The flag is emitted unconditionally, true or false, so its **absence means "written before 2026-07-29", not "measured and clean"**. The gate itself is unchanged: `objectionGates` is the same rule byte-for-byte and its pre-existing test passes unmodified — a Degraded object still always gates, because a high objection may have been cut off with the tail.
- **why it exists:** `bugs_open/138`. A reviewer that exceeds `max_tokens` is marked `Degraded`, and a `Degraded` object gates unconditionally (correctly — that is `bugs_closed/076`'s carve-out). But the verdict named the SEAT, so an **advisory seat that ran long became a blocking seat, silently**, 17 times in 14 days. The live harm is not the wasted round: **a high object-rate with no signal line is also the documented kill-switch for retiring a seat**, so a working seat can be pulled for being cut off. Demonstrated rather than argued — `review_architecture` went from 2-of-3 objections to 2-of-12 the moment it stopped truncating.
- **sources:** bugs_open/138_HANDOFF_2026-07-29...md; docs024_key_docs_latest/bugfix_138_degraded_gates/{PLAN,RUNBOOK,NOTES,README_where_we_are}.md; commit `3a59b5012`; platform/orchestration/actions/diagnose_council_decide_action.go
- **relations:** `bugs_closed/076` (added the `Degraded` carve-out this attributes — 076 behaving correctly, not a regression of it); `bugs_open/119` (sibling: `unreadable` voids a round, `degraded` silently decides one — different field, different `decided_by`, different fix); FIX-054 (`review_architecture`, the seat whose apparent noisiness exposed this); FIX-034 fixloop-digest and `diagnose_escalate_action` (consumers of `decided_by`, which is free text at every one — grepped, none parse it)
- **LANDMINE:** `max_tokens` for a council seat lives at **`config.ai_service.max_tokens`**, not `config.max_tokens`. Querying the wrong depth does not error — it returns a confident `(unset→default)` for every seat, which reads exactly like "nobody has right-sized these" and is false. Three seats had already been raised.
- **open review question:** the TRUNCATED label is only assigned when **nothing else** gated the round; a merits gate is named in preference even when the truncated seat came first in review order. That is deliberate (labelling a round TRUNCATED while a real high objection also gated would invite the author to dismiss it), but it means the flag under-counts rounds *containing* truncation — 10 of the 13 such rounds. If the question is "how often does truncation cost us", read the flag; if it is "how often are seats truncating", read `llm_call_log` or the RUNBOOK §1 seat-level query.
- **verify-later:** confirm on the first post-roll council round that a `council_report` carries `metadata->>'gated_by_truncation'`, and that a genuine truncation produces the TRUNCATED wording. Induce it (drop a scratch seat to `max_tokens: 500`) rather than waiting — **a green round proves nothing here, the failing branch is the whole bug.**

### FIX-056 — `Council-Submitted:` trailer: review credit that survives committing before the verdict (2026-07-30 addition)
- **status:** deployed (repo-side tooling + convention; live immediately, no image)
- **status-evidence:** `097_TRIGGER_council_review_v1.sh` prints the new trailer alongside `Council-Reviewed:`; `098_REPORT_unreviewed_commits_v1.sh` reads it, resolves it, and buckets it. Commit `fc5b790d3`. **Verified by inducing every branch in a scratch git repo, not by running the happy path** — 7 synthetic commits covering reviewed+approved, submitted+approved (the new credit path), submitted+revise (the new AWAITING bucket), no trailer, reviewed+revise (MISMATCH must still fire — it does), submitted+garbage-corr, and both-trailers (`Council-Reviewed:` wins — it does). All 7 bucketed correctly; live 2-day run 14 REVIEWED / 0 MISMATCH, no regression.
- **what:** A second commit trailer for the council gate. `Council-Reviewed: <corr>` **asserts** the change was approved — if the artifacts disagree, `098` calls it a MISMATCH, and that is the report's dishonesty surface. `Council-Submitted: <corr>` **asserts nothing**: it records which correlation covers this commit and leaves the verdict to the report. Because `db_decision()` resolves a correlation to its **latest** verdict at REPORT time (not commit time), a commit that records its correlation while the verdict is still pending is **credited automatically once approval lands, with no amend** — which matters because forward-only forbids amending a trailer in. New `AWAITING` bucket keeps not-yet-approved submissions out of MISMATCH; both trailers on one commit resolves to `Council-Reviewed:`, so a thread cannot assert approval and then be graded on the softer trailer.
- **why it exists:** Two standing rules were **mutually unsatisfiable**. Owner feedback 2026-07-20 (from a real incident: a thread held the `bugs_open/011` fix across four council rounds and the owner's own sweep commit `bca5d8255` took it to production with the verdict still REVISE) says commit the moment work is coherent and never hold code for a verdict. But `Council-Reviewed:` can only be written *after* approval. So a thread that complied, submitted, and was **approved** still read as un-reviewed for ever. Measured 2026-07-29: 8 REVIEWED against 40 UNREVIEWED in one day, at least one of the 40 (`3a59b5012`) approved on corr `919a05bf`. **The failure mode was a metric that understates coverage by an amount growing with compliance** — the one direction a visibility metric must not drift, because it makes a norm look ignored exactly when it is being followed.
- **sources:** commit `fc5b790d3`; bugs_open/138 (the round that surfaced it); fixloop_eg_dartsonline/{097,098,RUNBOOK_council_gate.md}; NOTES_running_council_gate.md 2026-07-29/30; CLAUDE.md council-gate section
- **relations:** the `Council-Reviewed:` trailer it complements (FIX-036/gate machinery); `098` coverage report; the 2026-07-20 commit-early feedback that created the tension; the 2026-07-29 ordering-exemption ruling (**related but NOT the cause** — that ruling is about default-OFF switches and whether a change can be held out of the fleet, a different axis; mis-attributing this to it was corrected in `c34967a3b`)
- **LANDMINE:** **neither trailer proves the correlation describes THIS change.** A pasted or stale id resolves and is credited. That was already true of `Council-Reviewed:` and is why this is a visibility report, not a gate. Also: **commits before 2026-07-30 cannot be resolved retrospectively**, so any `098` window spanning that date still mixes "never submitted" with "approved, but committed first" — the script says so in its own output rather than leaving the reader to infer it.
- **open review question:** whether `AWAITING` should age. A submission that never returns a verdict sits there indefinitely and reads as in-progress; there is currently no distinction between "queued 20 minutes" and "abandoned three days ago". `created_at` on the report is available, so it is cheap — deliberately not built, because a threshold guessed now is worse than one chosen after watching the bucket for a week.

### FIX-057 — Recoverable structural plan refusal: `repair_step` on `diagnose_persist_fix_plan` (2026-07-30 addition)
- **status:** built, committed to HEAD, **NOT yet live** (Go — inert until the next chassis image is rolled). Config half is migration `272`, deliberately NOT applied until the image is verified in the pod.
- **status-evidence:** `platform/orchestration/actions/diagnose_persist_fix_plan_action.go` — `planValidationRefusal`, `recordPlanRefusal`, `planRefusalNoteKind`, `planRefusalErrorCode`, `plan_valid` on the success path; `diagnose_artifact_count.go` — `countRunArtifacts`, the counter now shared with **both** of `diagnose_council_decide`'s loops; `sql_for_agents/272_…sql` (feature-designer, opted in) and `273_…sql` (fix-proposer, **written and dry-run clean but deliberately NOT applied**), both dry-run against live rows with `COMMIT`→`ROLLBACK`. **15 tests** in `diagnose_plan_refusal_test.go`, all passing, including four end-to-end through `DiagnosePersistFixPlanAction` itself — the helper-only tests hand `problems` in and so cannot prove the action REACHES the refusal path.
- **council trail:** corr `f4a4628f-3b90-4054-a875-f2cf72b83e72`. **Round 1 REVISE** (13 seats, 8 approve, 5 object, `gated_by_truncation:false`, gated by `llm_reliability`). Three objections CONFIRMED and fixed — a **wrong rollback block** (`debug_historian`), a **duplicated bounded-retry counter** (`reuse_agent`), and **recoverable-but-not-visible** (`bug_historian`) — plus a sibling migration for the consumer left exposed. Two refuted by measurement with the check added anyway: a root `ai_service` cannot shadow the step block here (no root block exists, *and* that behaviour was `bugs_open/009` and is fixed — `resolveAIServiceConfig` is a per-key overlay), and an `ai_service.thinking` key would be **dead config** (0 live steps set it; `execute_llm_prompt` reads `budget_tokens`). Round 2 resubmitted on the same trail.
- **what:** A step running this action may name a `repair_step` (+ `max_repair_attempts`, default 1). With it set, a **structural** validation failure stops being terminal: the rejected plan and the validator's exact problems are persisted as a durable artefact, and the action returns a RESULT carrying `plan_valid: false`, `should_repair_plan: true`, `validation_problems_text` and `rejected_plan_json`, so a `conditional` step can route it back to a repair prompt for a bounded number of rounds before failing as before. **With `repair_step` unset the action is byte-for-byte the old behaviour**, which is why the other two consumers need no change. The gate is NOT lowered: an invalid plan is still never persisted as a `fix_plan` on either path, and truncated/invalid JSON stays terminal (a cut completion is a `max_tokens` fault, not a repairable plan — `bugs_open/012`, `138`).
- **why it exists:** `bugs_open/099`. The action returned a bare Go error on any validation failure, which fails the step, routes to `complete_refused`, and **discards a completed, good design** over a rule the producing prompt was never told — with `orchestration_states.error` NULL and the reason only in `collected_data->>'__step_error'`, so a dashboard keyed on `error` reports the run as clean. Fix candidate 1 (migration `222`) stated ONE of the rules in the prompt; the validator has a dozen more, and that shape has to be repeated per rule, per agent, and drifts the moment the validator changes. This is rule-agnostic: every rule that exists and every rule added later becomes recoverable with no prompt edit.
- **sources:** bugs_open/099_HANDOFF_2026-07-26_feature_designer_plans_die_on_a_rule_it_is_never_told.md; docs024_key_docs_latest/bugfix_099_plan_refusal_recoverable/{PLAN,RUNBOOK,NOTES,README_where_we_are}.md; sql_for_agents/272_feature_designer_plan_repair_loop.sql; sql_for_agents/222_… (candidate 1, which this supersedes as the durable fix without replacing it)
- **relations:** `diagnose_council_decide` (FIX-055's file) — its round counter is the idiom reused here, including the `orchestration_id` scoping and the fail-closed-on-count-error rule; the `repropose`/`reframe` council loop (a DIFFERENT loop — see the open question); FIX-054/FIX-055 council seats, unaffected because `council-gate` is not opted in
- **LANDMINE (two, both silent):** (1) **`diagnosis_artifacts.kind` carries a CHECK constraint** — `bundle|iteration_note|fix_plan|council_report|escalation`. A new kind fails at RUNTIME, not at build, and `go build` cannot see it. The refusal note therefore reuses the allowed-but-previously-unused `iteration_note` slot and is discriminated by `metadata->>'note_kind' = 'plan_validation_refusal'`. **Any reader of `iteration_note` must filter on that key** or it will count refusals as whatever it thought `iteration_note` meant. (2) **A NULL `orchestration_id` never satisfies `= $2`**, so an unscoped refusal count returns 0 every time and the repair loop would never terminate. Handled here by refusing terminally when the run id is absent — but the same `orchestration_id = $2` shape appears in `diagnose_council_decide_action.go:514-517`, where a NULL would read as round 0 for ever. **Not measured there; see the open question.**
- **CORRECTED 2026-07-31 — "rule-agnostic" is TOO STRONG, and the entry above overstated it.** Reading the validator's own messages before running a live induction, its rules split into two classes. **STRUCTURAL** — duplicate file in a stage, modify-before-add, create-then-delete, forward `depends_on`, empty goal, bad stage id, contradictory `artifact_role`, missing checklist entry, and the per-stage `max_edits` cap (`"stage N: X edits exceeds the per-stage cap Y"`) — are all repairable by rearranging, with **no scope lost**, and the repair loop genuinely closes that class. **SIZE CAPS** — `max_stages` (`"…a build this broad needs splitting into more than one feature"`) and `max_total_edits` (`"…a build this broad needs splitting"`) — literally ask for *less scope*, while the repair prompt says *"do not drop scope"*. That is a contradiction: such a refusal burns its one repair round and then goes terminal. **Not a correctness bug** — the loop is bounded and lands exactly where the platform lands today, so the cost is one extra LLM call, and a genuinely oversized plan arguably *should* reach a human rather than be silently shrunk by a model. But the honest claim is "the structural class becomes recoverable", not "any validator rule does". Open: whether the repair prompt should detect a size-cap problem and escalate explicitly rather than attempt a repair it is forbidden to make.
- **open review question:** the bug file's candidate 2 says to route the refusal into the existing `repropose`, "which exists". **Checked, and it does not work as written** — `persist_plan` runs BEFORE any council, so on a first-pass refusal `repropose`'s prompt renders `{{.council_reviews.body}}` and `{{.check_results.results_text}}` against nothing and frames a structural problem as a council objection. Hence a dedicated `repair_plan`. Whether the two loops should later converge is open. Separately: whether `fix-proposer` should be opted in (it has the same defect and the same fix is one migration) is deliberately left for its owning lane rather than decided from outside it.
- **verify-later:** after the roll, pod-grep `plan_validation_refusal` **with a positive control in the same exec** (a tag bump does not imply a rebuild — `bugs_open/153`; `v1.0.1206` and `1207` shared an image id on 07-30), then apply `272` and **induce** the failing branch. ⚠ **099's own stated verification CANNOT prove this and returns a false PASS** — it says re-fire work item `7b89fb35` and require a `fix_plan` artifact plus no `complete_refused`, but that procedure was written for candidate 1, candidate 1 *worked*, and its rule is still live in the design step (path-qualified check returns `t`). That run now takes the **success** path: both conditions satisfied, repair loop never fires. Induce instead — set `persist_plan.config.max_edits=1`, which is repairable without losing scope — and require **four** things, the last two being what make it discriminating: (1) a `fix_plan` artifact exists, (2) the run does not end at `complete_refused`, (3) a refusal note EXISTS for the correlation (`kind='iteration_note' AND metadata->>'note_kind'='plan_validation_refusal'`), and (4) the run actually reached `repair_plan`. Commands in the workstream RUNBOOK. **A green run proves nothing — the failing branch is the whole bug.**

### FIX-058 — Council seat token-pressure instrument: a pull report and a CTE-only push alert (2026-07-30 addition)
- **status:** **deployed and exercised.** The report runs; the scheduled task fired within a minute of being seeded and wrote its first `doc_notes` row.
- **status-evidence:** `SELECT last_triggered_at FROM scheduled_tasks WHERE name='council-seat-token-pressure'` → 2026-07-30 15:21:22, `last_completed_at` equal; `SELECT subject_key FROM doc_notes WHERE categories ? 'seat-token-pressure'` → `council-seat-token-pressure:cc470e99…`, body listing 5 flagged pairs. The no-flag branch was self-tested by substituting impossible thresholds into the same pre_query: **zero rows**, which is the path the scheduler takes 23 hours in 24.
- **what:** Two halves that share no code and deliberately share no threshold. **Pull** — `104_REPORT_seat_token_pressure_v1.sh`: per (seat, cap) at each seat's CURRENT live cap, the p95 and peak of `output_tokens/max_tokens`, truncation count, round-level truncation-gate rate, and cross-council cap divergence. **Push** — `scheduled_tasks` row `council-seat-token-pressure`, `fire_message=false` so the `pre_query` IS the work (no Kafka message, no orchestration, no LLM, no credits), every 6h, inserting **one** `doc_notes` row when the flagged set changes. `subject_key` carries an md5 of the flagged set, so a persisting condition is announced once and an escalation announces itself — an event, not a heartbeat. Thresholds live in the pre_query and nowhere else; the report points readers at it rather than re-encoding it.
- **why it exists:** `bugs_open/138` candidate 2 — "the query is a one-line check; nothing runs it". The non-obvious part is WHICH rate. Counting truncations reads ~0 for ever, because candidate 3 raised the cap on every seat that had truncated; but a cap raise **moves** the cliff rather than closing it (`review_architecture` reintroduced truncation against its new 16000 cap within hours, on a longer prompt). So the instrument is built on **headroom**, the leading indicator, with two separately-named thresholds: near-miss (peak ≥ 95% of cap — truncation is a tail event, so the maximum is the primary signal) and pressure (p95 ≥ 85%). Both anchored on the live distribution, not chosen: the two populations that have truncated peak at 100% and sit at p95 96.1/85.7, and nothing below either cut has ever truncated.
- **sources:** bugs_open/138_HANDOFF_2026-07-29_…md; docs024_key_docs_latest/bugfix_138_degraded_gates/{RUNBOOK §7–8, NOTES, README_where_we_are}.md; fixloop_eg_dartsonline/104_REPORT_seat_token_pressure_v1.sh + 104_TASK_seat_token_pressure_v1.sql
- **relations:** FIX-055 (`gated_by_truncation`, whose flag section 3 reads); 102_LINT_council_seat_parity (compares a seat against its OWN council's family and declines cross-council comparison — this reports divergence as information, not as drift, which is a different claim); 099_SYNC_gate_roster (mirrors fix-proposer→council-gate only, which is *why* section 4 exists); `bugs_closed/076` (the `Degraded` carve-out being measured)
- **LANDMINE (two, both silent, both hit during construction):** (1) **`llm_call_log.agent_type` cannot attribute a review call to its council.** Everything before 2026-07-26 14:54 logged `generic`; from 15:03 the same calls log `council-gate`; **`fix-proposer` has never appeared at all.** `WHERE agent_type='council-gate'` therefore discards 1,798 rows of the same population without erroring. Key on (seat, cap) and report `n_holder` — exact for feature-designer and the experience councils, a lower bound for the fix lane. (2) **The denominator changes inside the window.** `max_tokens` is per call and caps were raised mid-window, so a p95 of `output/max(max_tokens)` mixes populations: it renders `review_editquality` as "95% of a 16000 cap" when the 16000-cap rows peak at 62.9%. Compute the ratio per row and join on the seat's current cap.
- **open review question:** whether the near-miss threshold should scale with the cap (a 5% margin is 400 tokens at 8000 and 800 at 16000, and the risk is surely absolute-ish, not proportional). Left alone deliberately — every 16000-cap seat currently sits below 63%, so the question has no live consequence and picking a rule now would be guessing. Revisit when a 16000 seat first crosses.

### FIX-059 — Seat length budget applier: one block, many seats, snapshot-then-write (2026-07-30 addition)
- **status:** **APPLIED 2026-07-31 to 10 seat/council targets** (owner approved the write). ~~built, self-tested, NOT applied~~ — the earlier `--apply` refusal by the session's permission classifier was lifted.
- **status-evidence:** `--verify` reports **10 of 10 APPLIED**; a second `--apply` reports "nothing to do", so idempotence is exercised not claimed; `099_SYNC_gate_roster.py` dry run reports **drift (none)**, which is the check that the mirrored pair was not edited on one council only. Cutovers taken from `agent_definitions.updated_at`, not scrollback: `15:12:49`–`15:12:58` (guardian ×3, improvement_guardian ×2, debug_historian ×2) and `15:39:26`–`15:39:30` (editquality ×3). The classifier was exercised against four real live seats and returned `HAND-WRITTEN` (council-gate/review_architecture, correctly refusing to overwrite a hand-authored block), `NEEDS-BLOCK` ×2, and `MISSING` for a non-existent seat. **BEHAVIOUR CHANGE UNVERIFIED:** an orchestration keeps the workflow it loaded at spawn, so only rounds spawned after each cutover carry the block, and none had run at time of writing. Peaks to beat: guardian 99.2%, debug_historian 99.8%, improvement_guardian 96.6% (all of 8000), **editquality 98.3% of 16000**. If they do not move, the block is being IGNORED and that is the finding, not a null result.
- **what:** A single copy of a LENGTH-BUDGET prompt block, inserted before the `## Output` anchor (the only heading present in all 51 live `review_*` templates), idempotent via a start phrase plus an `— end length budget —` sentinel, refusing to touch a hand-authored block that has no sentinel. `--apply` calls `snapshot_agent(<type>, …)` once per agent type before writing, updates via `jsonb_set(..., create_if_missing=false)` and asserts exactly one row per target — the row count is the check, because a wrong path is a silent no-op there. Targets carry their own evidence string.
- **why it exists:** `bugs_open/138` candidate 4, which asked for the load-bearing field FIRST in every seat's schema. **Measurement mostly refuted that**: `reviewer`/`verdict` are already first in 51 of 51; 0 of 2,713 stored objections lack a severity, so the severity-last theory never fires; and moving `notes` to the head would put `objections` — which survive 80% of truncations and carry both the gate's severities and the proposer's revision content — into the tail instead. What generalises is the OTHER half of the architecture-seat fix: the length budget, which is the half the evidence credits (outputs got *shorter*, peak 4,443 tokens = 28% of the new cap). Deliberately does NOT generalise that block's "at most 3 objections" clause — budgeting coverage across every council loses real objections invisibly, so this block budgets prose and says explicitly to cut words, never findings.
- **sources:** bugs_open/138_HANDOFF_2026-07-29_…md (the 2026-07-30 entry, incl. the refutation table); docs024_key_docs_latest/bugfix_138_degraded_gates/{NOTES, RUNBOOK §9}.md; scripts/apply-seat-length-budget.py
- **relations:** FIX-058 (the report that chose the target list); FIX-054 (the architecture seat, whose hand-written block this generalises and deliberately does not overwrite); 099_SYNC_gate_roster (fix-proposer and council-gate are both targeted precisely so the mirror stays clean — a block on one only would BE the drift); 102_LINT_council_seat_parity
- **LANDMINE:** **`prompt_template` and `max_tokens` sit at DIFFERENT depths and neither wrong path errors.** The cap is `config.ai_service.max_tokens`; the prompt is `config.prompt_template`, a SIBLING of `ai_service`. Reading the prompt at the cap's depth returns NULL for **all 51 seats**, which reads as "these seats have no prompts". This was hit on 2026-07-30 by a thread that already knew about the cap-depth trap from the day before — knowing "watch the depth" does not tell you which keys are nested.
- **open review question:** whether the block should eventually reach all 51 seats. Not done, on the owner's own criterion for the sibling change (raise the seats that actually truncate, leave the rest) — applied here to the leading indicator instead of the lagging one. Extending is one line in `TARGETS`, which is why it is a script, and **3 of the 10 targets were chosen by the FIX-058 alert rather than by hand** (`debug_historian`, peak 99.8% at a p95 of 62.2% — invisible to a p95 rule; then `editquality@16000`). **The sharper open question is whether a prompt budget can hold a seat at all:** `review_editquality` grew into its DOUBLED cap in three days with no prompt change (13,115 → 15,721 tokens, peak 98.3%, 52/52 attributable), which is this bug's "a raise moves the cliff" claim measured free of `review_architecture`'s confound. If its post-cutover peaks stay near 98%, the answer is a per-seat instruction, not a third raise.
