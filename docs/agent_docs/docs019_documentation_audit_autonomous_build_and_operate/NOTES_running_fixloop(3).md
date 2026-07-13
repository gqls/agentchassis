# RUNNING NOTES — Diagnosis→Fix Loop (v1)

Chronological; newest entries appended under DISCUSSION LOG; decisions
promoted to DECISIONS with rationale.

## 2026-07-06 — Workstream founded

- Origin: the diagnosis loop (read-only, three-tier citing, human-gated) is
  closed and working; the owner wants it developed into diagnosis→fix with:
  documented intake, live per-iteration/per-step reasoning into task-specific
  notes, fetchable bundles, fixes on a git branch, a council of specialist
  reviewers (guidelines / reuse incl. docs / bug-historian / compliance /
  per-pipeline guardians / named specialists e.g. trigger + site-work-items
  experts) feeding a decision-maker, architecture-change visibility, and a
  learning record of bugs. Motivating example: a chat re-invented a trigger +
  triage SQL that already existed — a specialist would have said so.
- Assets inventoried (see runbook): live loop, contextkit CLI (+ the real
  example_bundle.txt invocation), code_symbols corpus, the work-item relay,
  the tools chat's doc_notes infra (coordinate; their docs arrive next turn),
  the builder thread's pipeline map for guardians.
- Three documents created: RUNBOOK_diagnosis_fix_loop.md (task + phased plan
  F0–F3 + open questions Q-A…Q-H), this notes file, HANDOFF_fixloop_thread.md
  (manifest + the cmd/bundle invocation for code context).
- Method carried over: thin slices, pre-registered criteria, evidence first,
  snapshots, reuse before recreate; the loop's READ-ONLY core is preserved —
  the write surface (F1 fixer) is isolated by design.

## DISCUSSION LOG
(appended per exchange in the originating chat; handoff updated in step)

### 2026-07-07 — tools-chat rev-22 absorbed (their runbook + running notes)
- THEIR LIVE STATE: doc tables shipped; diagnose-agent workflow rewired
  emit→persist_note→complete with a skip-don't-guess subject gate (run-3
  verified the skip); load_runtime error-routing applied — ANCHORLESS RUNS
  SURVIVE (≈26 min / 5 iterations); 3b (subject threading) in flight; first
  tool PLAN live 2026-07-07 12:32; Stage 5 = static Tier-2 criteria check;
  Stage 6 = browser-runner adapter (Playwright, 035-conformant contract).
- WHAT IT ANSWERS HERE: Q-F → reuse doc_notes (per-iteration rows pending
  their volume sign-off; intake adopts their 084 envelope + subject fields).
  Q-B → the site-less-bug loop side is DONE by their routing; only the item
  namespace remains. Q-A refined: egress as write-through inside assemble
  (Go-side) — keeps us entirely off their emit-adjacent surface.
- COLLISION RULE ADOPTED: diagnose workflow changes are fetch-first +
  coordinated; their surface is active until 3b closes.
- GOTCHAS ADOPTED: error_step INSIDE config (step-level silently ignored —
  001 §16; fix adjacent instances when touching a workflow, noted); ~3600s
  pod reap → ProcessingHistory state-dump as evidence substitute.
- SHARED-COMPONENT REGISTER: 084 trigger; doc_notes; browser-runner (future
  F1 verification + a council instrument); criteria-fence pattern.
- RELAY QUESTION for the tools chat: per-iteration diagnosis notes in
  doc_notes — acceptable volume/shape? proposed: one note per iteration,
  category 'diagnosis', body = hypothesis/scope/requests/verdict grounds.

### 2026-07-07 — Q-D veto semantics decided (owner)
- Flag-based: DEFAULT = decision-maker weighs all opinions; a hard_veto flag
  at reviewer/pipeline/tool/component scope makes that reviewer's negative
  verdict a BLOCK (accessibility, legal = motivating cases). Open detail for
  the new thread: flag placement (definition column vs council config vs
  both; most-specific wins?) and the guideline-gap case (block vs side-task —
  leaning side-task).
- Per-iteration note volume issue explained to the owner + relay message
  DRAFTED for the tools chat (three shape options: filterable category /
  is_current-versioned single note / own table + terminal note only).
- Standing unconfirmed at the time: Q-A/Q-B/Q-C — ALL CONFIRMED/DECIDED by
  the owner on 2026-07-07 (see DECISIONS): diagnosis_artifacts; (c) own
  working-notes table + terminal-only in doc_notes; pipeline='diagnose';
  separate fixer agent w/ constrained edit plan + gofmt/build gate.
  Guideline-gap = side-task, with the mechanism answered (amendment PR
  against the guideline docs; human terminal; F3 recurrence record).
  REMAINING OPEN (F2-phase; may open the new thread's discussion): Q-E
  architecture-change signals, Q-G reviewer context, Q-H result format.

## DECISIONS (with rationale)

### 2026-07-07 — F0/F1 design settled (owner decisions)
- **Q-A — bundle egress = `diagnosis_artifacts` table**, written through from
  INSIDE the assemble action (Go-side; zero workflow-shape change; off the
  tools chat's emit-adjacent surface). Rationale: DB is the durable shared
  memory; one-SQL fetch by correlation; cd/Kafka size class already bitten.
- **Q-F shape = (c)**: working notes in OUR OWN table; only the terminal note
  lands in the tools chat's doc_notes — via THEIR existing persist_note
  wiring (we write nothing into their table). Integration duty: the intake
  envelope carries subject_type/subject_key so their gate opens post-3b.
  REFINEMENT (proposed, start unified): fold working notes into
  diagnosis_artifacts with kind ∈ {bundle, iteration_note}; split later only
  if retention diverges. Relay to tools chat downgrades to a courtesy FYI.
- **Q-B — intake = `needs_diagnosis` work item, `pipeline='diagnose'`**
  namespace (site-less code bugs ride null-site in the new namespace; the
  loop side already survives anchorless via their load_runtime routing);
  envelope extends their 084 trigger; manual trigger retained for ad-hoc.
- **Q-C — the fixer = a SEPARATE agent**: distinct responsibility; the git
  WRITE token isolated behind the existing spawn-gate pattern; diff produced
  as a CONSTRAINED EDIT PLAN and validated by gofmt+build in a spawned job
  before any PR opens.
- **Q-D completion — guideline-gap = SIDE-TASK** (does not block the fix):
  a work item carrying the evidence; handler drafts a concrete amendment and
  opens a PR against the GUIDELINE DOCS (docs-branch symmetry with F1);
  human review terminal; gaps accumulate in the F3 learning record
  (recurrence = restructuring signal). Decision-maker sees "gap raised" as
  context only.
