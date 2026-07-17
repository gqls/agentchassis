# RUNNING NOTES — Council Gate thread

Turn-by-turn record, newest at the bottom (house practice, matching
`docs026_concept_register/RUNNING_NOTES_concept_register.md`). Standing state
lives in `RUNBOOK_council_gate.md`; this file records how it got there.

---

## Turn 1 — 2026-07-17 — Decisions collected; advisory gate built end to end (NOT applied)

Cold-started from `HANDOFF_2026-07-17_council_gate_thread.md` + design §2.

**Recon before building** (schema-first rule): read the v6 fix-proposer
workflow, `diagnose_persist_fix_plan_action.go`,
`diagnose_council_decide_action.go`, `append_doc_note_action.go`,
`fixloop_digest_action.go`, and the 091 trigger envelope. Three mechanical
facts shaped the build:
1. `plan_field` resolves over collected data, where `input_data` lives — so a
   submission can carry a ready-made fix_plan JSON and the **wrapper needs
   zero new Go**; the whole gate is config + scripts, live on seed apply, no
   image build.
2. Council round-counting is **orchestration-scoped** in the deployed source,
   so resubmissions on the same correlation are judged fresh; the shared
   correlation is for the artifact trail, not the cap.
3. The digest action runs **in the pod, which has no git repo** — the
   "un-reviewed platform commits" join can only run repo-side. Built as a
   script that can persist its report to doc_notes (same awareness channel),
   not as a digest section.

**Owner decisions collected** (the handoff's four questions): scope =
`platform/` + `internal/` + `pkg/`; advisory first; credits per submission
(= per task/commit); roster = **wait for more stage-3 seats** — which makes
seat-building the launch gate.

**Built** (files only, nothing applied):
- `0NN_council_gate.sql` — 17-step orchestrator: schema hint → persist
  submission (same structural validation as proposer plans, 64KB cap) → the
  three v6 reviewer prompts with the diagnosis context swapped for the
  author's rationale → deterministic council_decide (guardian hard veto,
  all-three review_fields) → no repropose loop (the author revises; a revise
  verdict runs ALL three reviewers' checks first so objections come back
  settled with evidence — v6 only ran two reviewers' checks, an omission not
  copied) → deterministic SQL-composed verdict doc_note (no LLM in an
  awareness surface) → three terminals. Literal balance verified with a
  comment-aware tokenizer (validator agreed with the known-good v6 file).
- `097_TRIGGER_council_review_v1.sh` — submission wrapper: client-side
  validation (thin-rationale refusal — missingkey=zero would render a missing
  rationale as silent blank in the prompts, so it must fail HERE; scope
  pre-filter so docs never spend credits; ≤8 edits, ≤64KB), single-line jq -c
  payload (kcat -P line-split trap), RESUBMIT_CORR support. Dry-run tested
  against a stubbed kubectl: good/docs-only/thin-rationale all behave.
- `098_REPORT_unreviewed_commits_v1.sh` — deterministic git↔verdict join on
  the `Council-Reviewed: <correlation>` commit trailer (exact join, not
  file-overlap heuristics; MISMATCH bucket makes false claims visible;
  NO_DB degrades to TRAILER-UNVERIFIED, never silently to REVIEWED).
  **First live run: 28 in-scope commits in 3 days, 0 reviewed — the
  baseline.**
- `RUNBOOK_council_gate.md` — standing state + launch checklist.

**Seat #4 (reuse-agent) — a live cross-thread collision, resolved by
addendum, not overwrite.** The owner's roster ruling put seat-building on the
gate's critical path, so this thread set out to draft the reuse-agent spec —
and found `docs026_concept_register/PILOT_reuse_agent_reviewer.md` already on
disk, written minutes earlier by the concurrently-running concept-register
thread, complete and better-grounded (FIX-036's founding incident: a
reinvented trigger+triage SQL pair). Did NOT overwrite it. Appended an
attributed §6 addendum carrying the two facts only this thread knew: (i)
there are now TWO council definitions (fix-proposer + the gate seed) and any
seat migration must patch both or the rosters silently drift; (ii) v6's
`run_checks.check_fields` omits the bug-historian's checks — their four-edit
v7 patch would repeat that omission for the new seat; the v7 migration should
carry all four reviewers' checks (the gate seed already does); (iii) the
owner's four gate rulings, as context for their roster-scaling question. The
file was modified on disk even between this thread's read and its edit — the
concurrency is not theoretical.

**Read-aloud summary written** (user request, mid-turn):
`SUMMARY_council_gate_2026-07-17.md` — what we're doing / where we are /
what's next, plain language, house read-aloud convention. Standing practice
from here: this notes file is updated every turn.

**Deliberately not done:** applying the seed (roster ruling + named-target
permission gate), PR-mode (build order step 4, owner's explicit go), any edit
to the live fix-proposer, any overwrite of the other thread's pilot spec.

**LATE-TURN DEVELOPMENT — the roster arrived while we worked.** The
concept-register thread, in its own conversation with the owner, built and
APPLIED seats #4 and #5 to the live fix-proposer the same afternoon:
reuse-agent (v7) and guidelines-agent (v8) — the live council is now **5
sequential reviewers**, and it also produced `DESIGN_relevance_filter.md`
(seat-skipping needs a small chassis Go change: `diagnose_council_decide`
hard-fails on an absent reviewer field, so "skip" must become "abstain").
Consequences handled here: (1) synced `0NN_council_gate.sql` to the v8
roster — both new steps mirrored verbatim with the rationale-context swap,
`review_fields` and `check_fields` now carry all five; 19 steps, literal
balance re-verified; (2) flagged (RUNBOOK + pilot §6(ii)) that the live v8
`run_checks.check_fields` still lists only editquality + guardian — three
advisory seats' checks are solicited but never executed; a one-line config
fix belonging to the fixloop/concept-register surface, not this thread's;
(3) the owner's launch condition ("more seats first") is now plausibly MET —
reframed as launch-checklist step 0: launch on 5 seats, or wait for the
relevance filter? (4) updated the read-aloud summary to match reality.

<!-- Append new turns below this line. Format: ## Turn N — date — one-line summary -->
