# DESIGN (draft) — two evolutions of the fix loop: the FEATURE BUILDER and the COUNCIL GATE

*2026-07-17, "diagnosis fixloop 3" thread. Owner asked two questions; this records
the design answers. Both are DRAFTS for owner sign-off — nothing here is built.
Both reuse the proven chain (diagnose→plan→council→implement→gate→PR) rather than
inventing new machinery; that chain's cage (dedicated pod, writes only via
git-adapter, build gate, human merge) is non-negotiable in both.*

*Context that moved under us and helps both designs: the council is already
3 seats (bug-historian reviewer live 2026-07-16, fix-proposer v6 — the concept
register thread shipped stage-3 seat #1), and `make build-<service>-ref` now
exists, killing WIP-bundling deploys.*

---

## 1. The FEATURE BUILDER — multi-step capability construction

**Question:** can we reuse/extend this tool to BUILD complicated features that need
several steps — e.g. create a workflow AND the actions it calls?

**Answer: yes — the chain generalises; three components change, the cage doesn't.**

### What maps directly
| Fix loop | Feature builder |
|---|---|
| symptom (needs_diagnosis) | **spec** (`needs_capability`) — and triage ALREADY routes `capability_gap` items to the roadmap, so the intake pool exists and is deliberately kept out of the diagnosis loop |
| diagnosis (find the cause) | **design** (find the shape): a feature-designer agent reads the spec + relevant code and emits a STAGED PLAN |
| fix_plan (one constrained edit) | **staged plan**: an ordered list of stages, each itself a constrained edit plan — files allowlist, new-file manifest, expected symbols, per-stage gate criteria |
| council reviews the plan | council reviews the DESIGN first (stage list, boundaries, reuse), then optionally per-stage diffs |
| implementer: one branch, one commit, one PR | implementer LOOPS over stages: `create_branch` once, one `branch-commit` per stage via git-adapter, build gate per stage, ONE PR at the end |
| gofmt + targeted go build | per-stage build gate + an END gate that runs the affected packages' `go test` (features add behaviour; build-only is not enough) |

### The three real deltas
1. **Plan schema grows `stages[]`.** Today's `fix_plan` artifact is one edit plan;
   the builder's is a sequence with declared inter-stage dependencies. The council
   decision router needs no change — it judges the artifact it is given.
2. **The implementer must create files and loop.** PR #1 was a 2-deletion edit;
   features need new files (a new action file, its registration, a seed SQL for the
   agent def). The hard allowlist stays — it just includes to-be-created paths from
   the stage's manifest.
3. **Registration/seed discipline is encoded, not automated.** A new action is
   registered in code (ships with the image) but its agent-def seed is DB-side and
   LIVE immediately — the proven order is image FIRST, then seed. The builder
   therefore emits seeds as FILES IN THE PR (never executes them), with the PR
   description carrying the ordered apply checklist. A feature is "done" when the
   owner merges AND applies the seed — two human acts, deliberately.

### Human gates (more than the fix loop, because blast radius is bigger)
spec approval (owner) → design approval (council verdict + owner) → per-stage gates
(mechanical) → PR merge (owner) → seed apply (owner). Nothing self-merges, nothing
self-seeds.

### Build order when the owner green-lights
(1) stage-plan artifact schema + feature-designer agent (reuse fix-proposer's
prompt seams); (2) implementer stage-loop + new-file allowlist; (3) test gate;
(4) pilot on a SMALL real feature with a known shape — the natural candidate is one
this tool itself needs (e.g. the F1.2 "ref as per-run input" cleanup, or an
iteration-notes writer), so the first feature build is self-hosting and gradable.

---

## 2. The COUNCIL GATE — every thread's fixes through the council

**Question:** can we run ALL fixes from ALL threads through the council for
approval before they go live?

**Answer: yes — decouple the council from the proposer and open it as a service.
The council already judges an artifact by correlation_id; nothing in F2 cares who
authored the plan.**

### Shape
- New thin trigger (`09X_TRIGGER_council_review.sh`) + orchestrator: intake = a
  SUBMISSION — either a constrained edit plan, or (more realistically for human
  threads) a DIFF + rationale + files-touched list. A small wrapper step converts a
  diff submission into the fix_plan artifact shape on a fresh correlation.
- Then the EXISTING machinery runs unchanged: 3-seat council (right-edits,
  platform-safety, bug-historian; more seats as concept-register stage 3 lands),
  reviewer checks (read-only SQL under containment), decision router
  (approve / revise→verify→repropose / veto→reframe / escalate), live schema hint.
- Terminal = council_report artifact + a doc_note verdict; if the submission rode a
  `fix/*` branch pushed via the git-adapter, attach the verdict to the PR.

### The honest limits (state them up front)
1. **Advisory unless the workflow is PR-shaped.** The council cannot intercept a
   hand-commit to the shared branch — many concurrent sessions commit directly.
   Enforcement options, weakest→strongest: (a) convention: threads submit before
   committing; (b) visibility: the digest lists platform-code commits that carried
   no council verdict (deterministic — join git log against council_report
   artifacts); (c) structural: platform-code changes move to `fix/*` branches + PR
   + owner merges only council-green PRs. (c) is the real gate and aligns with the
   coordination thread's commit-per-task + build-from-ref rules; it is also a
   workflow change for every thread, so it is the owner's call, not this tool's.
2. **Cost.** Every council run spends credits. Scope the gate: platform code
   (`platform/`, `internal/`) yes; docs/site-content no. A cheap deterministic
   pre-filter (paths touched) keeps docs commits out.
3. **Latency.** A council round is minutes — fine for PRs, hostile to rapid
   iteration. Pair with (c): iterate freely on the branch, council once at PR time.

### Relationship to the widening track
This IS the council-widening track's delivery vehicle: seats added via concept
register stage 3 immediately serve BOTH the fix loop and the gate. `FIX-036` (the
wider-roster vision) becomes "the roster the gate runs". No competing design.

### Build order when green-lit
(1) submission wrapper (diff→fix_plan artifact); (2) trigger script + orchestrator
seed; (3) digest section "un-reviewed platform commits" (visibility BEFORE
enforcement, per the standing awareness-before-autonomy rule); (4) owner decides
whether/when to flip to PR-mode (c).

---

## Decision points for the owner (both designs)
- Feature builder: green-light the schema + designer-agent build? Pilot feature?
- Council gate: advisory (a/b) first, or straight to PR-mode (c)? Path scope?
- Both consume the same council seats — sequence seat-building (register stage 3)
  ahead of either, or grow seats as the pilots demand them?
