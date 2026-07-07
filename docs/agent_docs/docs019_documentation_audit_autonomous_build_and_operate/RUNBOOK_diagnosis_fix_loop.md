# RUNBOOK — Diagnosis→Fix Loop (v2 of the diagnosis loop)

## THE TASK (read this first if you are new)

The platform already has a working, read-only **diagnosis loop**: given a bug
symptom, an agent forms a hypothesis, gathers scoped evidence (real code
bodies from an indexed corpus, read-only database rows, runtime records),
issues a verdict that must CITE evidence or ABSTAIN, and re-scopes by
FOLLOWING what the evidence names — until it confirms a cause with citations
across all three tiers (static code / live data reads / runtime records). It
is deliberately human-gated: it emits a diagnosis and changes nothing.

**This workstream develops it into a diagnosis→fix system** with, in order:
1. **An easy, documented route in and out**: one clear way to input a task or
   bug; live monitoring of what the loop is doing and why (per-iteration —
   and per-step — reasoning written to a task-specific running-notes file);
   and a usable result out, including the ability to
   **consume/download/fetch the bundles** the loop builds each iteration
   (today they are ephemeral, in-memory).
2. **Fixes on a branch**: the confirmed diagnosis drives a proposed fix
   committed to a separate git branch, so the human can amend, ditch, or
   apply it. The loop's core stays read-only; the write surface is isolated.
3. **A council of reviewers** before any fix is finalised: independent
   specialist agents each judging the proposal from their own perspective and
   sending opinions to a **decision-maker** that weighs them all. Initial
   roster (from the problem owner): a **guidelines agent** (does the fix
   adhere to guidelines 000-0xx — or did the guidelines fall short?); a
   **reuse agent** (are we building a new route where a tried-and-tested
   solution exists — checking BOTH code and docs); a **bug-historian**
   (catch early; record bug categories so the same class never repeats); a
   **compliance/legal eye**; **pipeline guardians** — one per master
   workflow/pipeline (seeded from the builder thread's relay map) — checking
   the fix doesn't infringe on another workflow; and **specialist knowledge
   agents** (e.g. a trigger expert, a site-work-items triage expert) that
   answer "we already have one of these" — the motivating example being a
   chat that composed a trigger + triage SQL which already existed.
4. **Architecture-change visibility**: make it loud when a proposed change is
   accidentally fundamental — touching platform contracts, message shapes,
   many packages, exported signatures — before it ships.
5. **Learning**: recorded bugs, proposed guideline amendments, corpus and doc
   enrichment feeding back in.

**Mission for the tool**: use everything available to reach the right result
— the code corpus, schemas, runtime records, the guidelines themselves — with
checks, balances and second opinions built in.

## What already exists (do not rebuild)

- The **live loop** (chassis `pkg/diagnose` + diagnose_* actions): three-tier
  CONFIRMED diagnosis achieved; engine guards (named-scope narrowing, capped
  call-graph expansion, cite-or-abstain, SQL guard read-only allowlist);
  the §7D resolver (fuzzy scope → real symbols, incl. basename
  canonicalisation). RUNBOOK_code_retrieval_route.md is the closed record.
- **contextkit CLI** (`cmd/bundle` + `cmd/analyse`, RUNBOOK_31_.md): manual
  paste-ready bundles from explicit -scope flags (example_bundle.txt is a
  real invocation). The live loop's assembler is its descendant.
- The **code_symbols corpus** (3,7xx symbols, single current commit; the
  index-orchestrator reindex route) + vector/trigram lookup.
- The **work-item relay + immune system** (builder thread §B2/§B3): a proven
  intake/dispatch/retry mechanism a diagnosis task could ride.
- The **tools chat's travelling-docs infrastructure** (doc_plans/doc_notes,
  persist_diagnosis_note after diagnose_emit, KB tool_docs): per-subject
  PLAN/NOTES persistence — the natural home for per-task loop notes.
  COORDINATION REQUIRED; their runbook/notes arrive next turn.
- The **builder thread's pipeline map** (RUNBOOK_builder_route.md §B0–§B3):
  the seed material for pipeline guardians.

## Phased plan (thin slices; pre-registered criteria per slice)

**F0 — Intake, observability, egress (first; test on the user's next real bug)**
- F0.1 Bundle egress: persist each iteration's bundle durably + one
  documented fetch route. (Design question Q-A below.)
- F0.2 Task input: one documented way in. (Q-B.)
- F0.3 Per-task running notes: the loop writes its reasoning per iteration
  AND per step (hypothesis, scope chosen and why, requests issued, verdict
  grounds, resolver substitutions) to a task-specific notes doc. (Q-F —
  likely REUSE the tools chat's doc_notes.)
- Success criteria: a bug goes in via the documented route; the human can
  watch reasoning appear; every bundle is fetchable afterwards; the diagnosis
  lands as today.

**F1 — Fix on a branch**
- F1.1 Fix-proposer (Q-C: who/where) turns a CONFIRMED diagnosis into a
  patch on a new branch via the git adapter; PR opened; human amends/ditches/
  applies. Write-token security isolated to the proposer (the spawn token
  gate pattern exists).
- F1.2 The per-task notes gain the proposal rationale + diff summary.

**F2 — The council**
- Independent reviewers (roster above), each a small agent with its own
  curated context (Q-G), producing a structured opinion (verdict-wire-style
  contract: verdict + citations + objections + suggested alternative);
  a decision-maker aggregates; the human sees diagnosis + proposal + council
  report (Q-H format). Architecture-change detector runs as one reviewer
  (Q-E signals).

**F3 — Learning**
- bug_records (category taxonomy, recurrence checks feeding the historian);
  guideline-amendment proposals routed to the human; corpus/doc enrichment.

## Boundaries
- Tools chat: owns doc_plans/doc_notes + tool docs + its diagnose_load_runtime
  draft — F0.3 reuses rather than reinvents; align next turn on their notes.
- Builder thread: owns the relay/spine; the pipeline map is INPUT here;
  guardian findings that imply relay changes route back through it.
- Quality thread: a future consumer of fixes; no overlap now.

## OPEN QUESTIONS — the critical-discussion agenda (positions welcome)
- **Q-A egress medium** for bundles: diagnosis_artifacts table vs
  orchestration-state collected_data vs object storage vs a git branch of
  artefacts. (Constraint memory: cd bloat already caused a 1.27MB Kafka
  incident; bundles are ~60KB × ≤5 iterations.)
- **Q-B intake**: a `needs_diagnosis` work item (rides dispatch + immune
  system + claims; wrinkle: pure code bugs have no site_id — pipeline
  namespace or null-site allowance needed) vs the manual trigger only vs both.
- **Q-C the fixer**: new agent (distinct responsibility, isolated write
  token) vs extending diagnose-agent; how the diff is produced (LLM patch vs
  constrained edit plan) and validated (gofmt/build in a spawned job?).
- **Q-D council topology**: parallel reviewers + decision-maker vs staged
  gates; opinion contract shape; quorum/veto semantics (does compliance veto?
  does the guidelines agent's "guideline fell short" open a side-task?).
- **Q-E architecture-change signals**: packages touched breadth; platform/ vs
  actions/; exported-signature diffs vs the corpus; message/topic/schema/
  contract changes; migration presence. Which are load-bearing?
- **Q-F per-task notes**: reuse doc_plans/doc_notes keyed by task/correlation
  (coordinate with tools chat) vs a new table.
- **Q-G reviewer context**: per-reviewer docselect/contextkit bundles vs one
  shared bundle + role prompts vs curated RAG corpora per specialist.
- **Q-H the human-facing result**: what exactly lands (PR link + diagnosis +
  council report + task notes link) and where.

## CURRENT POSITION — 2026-07-06
Documents created (this runbook; NOTES_running_fixloop.md;
HANDOFF_fixloop_thread.md incl. the file manifest and a ready-to-run
contextkit bundle command). DISCUSSION PHASE is live in the originating chat;
the notes log it and this trio updates as positions are taken, so the handoff
always carries the current state. Next turn: the tools chat's runbook/notes
arrive for the Q-F/coordination alignment.
