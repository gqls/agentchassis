# HANDOFF — start the COUNCIL GATE thread

**Filed 2026-07-17** from "diagnosis fixloop 3". Cold-start for a NEW thread that
opens the F2 reviewer council as a service every thread can run its fixes
through. **Full design: `DESIGN_feature_builder_and_council_gate.md` §2 — read it
first.** This doc adds thread-boot context, the build order, and the owner
decisions to collect before building.

## Working rules
Same as the feature-builder handoff (Go; schema first; reuse before recreate;
deploy from committed ref; commit per task with explicit paths and a staged-file
check immediately before commit).

## What already exists (the service IS mostly built)
- The council: 3 reviewer seats (right-edits, platform-safety, bug-historian —
  the last added by the concept-register thread 2026-07-16, fix-proposer v6),
  deterministic decision router (approved | revise→verify→repropose |
  first-veto→reframe-once | rejected/exhausted→escalate), reviewer `checks[]`
  (read-only SQL run under the diagnosis containment), live schema hint.
- It judges a `fix_plan` artifact by correlation_id and DOES NOT CARE who
  authored it — that is the whole decoupling.
- Verdict surfaces: `council_report` artifact + doc_notes; the digest
  (`fixloop_digest`) is the awareness channel to extend.

## Build order (design §2; visibility BEFORE enforcement — standing rule)
1. **Submission wrapper**: diff + rationale + files-touched → `fix_plan`-shaped
   artifact on a fresh correlation. This is the only genuinely new code.
2. **Trigger** (`09X_TRIGGER_council_review.sh`, clone the 091 envelope) +
   orchestrator seed that runs ONLY the council portion of the v6 workflow.
3. **Digest section "un-reviewed platform commits"**: deterministic join of git
   log (paths under `platform/`, `internal/`) against council_report artifacts —
   visibility first, no enforcement.
4. **PR-mode** (the real gate): platform changes ride `fix/*` branches via the
   git-adapter; council verdict attaches to the PR; owner merges only green.
   This is a fleet workflow change — DO NOT build past step 3 without the
   owner's explicit go.

## Owner decisions to collect FIRST (bring answers to the build)
- Scope: which paths trigger review (proposal: `platform/`, `internal/`,
  `pkg/`; never docs/site content)?
- Mode at launch: advisory (steps 1–3) or straight to PR-mode?
- Credit policy: council on every submission, or batch per PR?
- Seat roster for the gate: the 3 live seats, or wait for more concept-register
  stage-3 seats?

## Honest limits (state them to the owner, not around them)
- Advisory mode cannot intercept a hand-commit to the shared branch. Many
  concurrent sessions commit directly — proven repeatedly (index races on
  2026-07-16/17 even swept foreign files into other threads' commits).
- Each council run spends credits and takes minutes — fine at PR cadence,
  hostile mid-iteration. Gate at PR time, not per-commit.

## Cross-links
- `DESIGN_feature_builder_and_council_gate.md` §2 (the design this executes).
- Concept register `docs026_concept_register/` — stage 3 = the seat roster;
  FIX-036 is the wider-roster vision; seat-wiring is owner-sign-off-gated.
- Multi-session coordination workstream — the commit-per-task and
  build-from-ref rules this gate composes with.
