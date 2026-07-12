The overview now opens the runbook one level up from the mechanism — the why before the what. It explains that the wider system builds and operates marketing sites with autonomous agents; that when a build goes wrong, diagnosis means reading the right code, schema, and runtime evidence and discarding hypotheses until the real cause surfaces; and the two findings from this work that shape the tool — that symptom retrieval has a ceiling for infrastructure-layer causes, and that the hardest, most valuable debugging step is abandoning a confidently-wrong hypothesis, which LLMs do worst. Then it frames the loop as the response to both: it automates the debugging motion we performed by hand, read-only and human-gated, with safety in deterministic code and reasoning quality in the model verdict. I kept the existing mechanism description right below it, relabelled "mechanism", so the reader gets purpose first, then how it works.

## Overview — what this is for

The wider system builds and operates marketing websites with autonomous agents.
When one of those builds goes wrong — a page silently fails to update, a rebuild
reports success but changes nothing — diagnosing it means reading the right code,
the right schema, and the right runtime evidence, then forming and *discarding*
hypotheses until the real cause is found. This conversation's work established two
things that shape the tool: first, that a single retrieval from a symptom
description **cannot** reach a cause that lives in shared infrastructure named
nothing like the symptom (measured: all retrieval methods scored zero on exactly
such a bug); and second, that the hardest, most valuable step in real debugging is
the willingness to abandon a confidently-stated wrong hypothesis when the evidence
breaks it — the step LLMs do worst by default.

The diagnosis loop is the response to both. It automates the debugging *motion* we
performed by hand throughout this work: form a hypothesis, gather scoped evidence
(read-only), judge whether that evidence confirms, refutes, or fails to settle the
hypothesis — and, crucially, when it refutes, re-scope by **following what the
evidence names** (the call graph, the runtime fault site) rather than re-searching
the symptom. It is read-only and human-gated: it emits a diagnosis with a full
evidence trail for a person to act on, and never changes code or triggers a run.
This runbook is how to build, run, and test the loop today, and what remains to
wire it to a live model. The safety and auditability live in deterministic code
(tested here); the reasoning quality lives in the model verdict (gated on a
real-bug evaluation before it is trusted).

---

**What this is (mechanism).** The diagnosis loop (design:
`contextkit/docs/DESIGN_diagnosis_loop.md`) is an agent loop that wraps the
read-only `cmd/bundle` gather around a cite-or-abstain verdict, re-scoping by
FOLLOWING runtime/call-graph evidence (not re-searching the symptom — the B4a
ceiling finding). This runbook is how to RUN and TEST the parts that exist today.

**What exists (built + unit-tested in `internal/diagnose/`):**
- the deterministic **scaffold** (`loop.go`) — loop control, the four convergence
  guards, the evidence trail, re-scope-by-call-graph;
- the **BundleGatherer** (`gatherer.go`) — shells out to `cmd/bundle` (read-only);
- the **AnalysisCallGraph** (`callgraph.go`) — re-scope by following the
  analyser's `calls`, dropping ubiquitous names;
- the **`cmd/diagnose`** entrypoint wiring them together.

**What does NOT exist yet (chassis-side follow-on):**
- the **real Verdicter** — the LLM cite-or-abstain step (DESIGN §2). It needs a
  model; it can't run in the contextkit module alone. Until it's wired, the loop
  uses either a SCRIPTED verdict file (for testing a known reasoning path) or a
  STUB that abstains every iteration (so a model-less run never fabricates a
  conclusion). The runbook's "real-bug evaluation" (last section) is gated on
  this.

---


● F1.1b(c) is code complete (b367a602). The full write step exists end to end — every action, the workflow seed, the trigger — validated as far as it can be without the deploys.

What landed this turn

Three final pieces, each shaped by a reuse-or-safety argument:

1. diagnose_read_repo_files — fetches the plan's files via the GitHub contents API (no git binary needed, read token from the spawn gate). The rule that matters: a modify file that 404s is a hard error — whole-file
   rewrites from files the model has never seen would be hallucination by construction, so it refuses.
2. git_adapter_request — one generic adapter caller instead of three copies of the 200-line commit envelope. Verbs are allowlisted (commit, create_branch, create_pull_request) — delete_repo is structurally
   unreachable from any workflow using it.
3. fix-implementer seed (15 steps, graph verified, dry-run passed live):

load plan/council/diagnosis → APPROVED gate (mirror of the CONFIRMED gate)
→ read current file bodies → sketch_to_files (whole files, "no drive-by changes,
the diff a human reviews must contain ONLY the plan")
→ deterministic allowlist → create fix/* branch → commit
→ build gate (golang Job) → green: PR into main │ red: NO PR — branch + build log left

Plus fix-implementer on the read-token spawn gate, and 092_TRIGGER_fix_implementer_v1.sh.

Your deploy checklist (also in the living handoff, in order)

1. Chassis image (5 new actions + spawn-gate entry) — bump tag; I'll verify in-pod as before
2. git-adapter image (branch/PR ops)
3. kubectl apply the RBAC file (pods/log for build-gate failure reports)
4. Apply 0NN_fix_implementer.sql
5. Token smoke test — one create_branch on a throwaway name proves write scope on gqls/agentchassis

Then the last piece before the first end-to-end run: seeding the small bug. My suggestion when you're ready: I plant a contained single-file defect on a low-stakes surface (e.g. a wrong constant or inverted condition
in a leaf function with an obvious observable symptom), we run it through the full pipeline — diagnose → plan → council (which should genuinely approve) → implementer — and the system's first-ever PR will have
earned every gate it passed. When the images are up, say the word and I'll verify the binaries and take it from there.


