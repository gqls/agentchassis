# 010 FEATURE — council decision adjudicator (severity/nature-weighted, not "any object → revise")

**Raised:** 2026-07-21, by the owner ("maybe the agent that evaluates council
decisions should have decided this"), during the fundamentallyai.com contamination
fix (bugs_open/055, council corr `03908b72-2471-474e-baaf-7952d1903460`).
**Status:** specified, not built.
**Owner:** fixloop / council-gate workstream (this touches the gate's own
decider, shared by both councils).
**Related:** `docs024_key_docs_latest/fixloop_eg_dartsonline/RUNBOOK_council_gate.md`,
`platform/orchestration/actions/diagnose_council_decide_action.go`.

## The observation (a real case)

The 055 contamination-allowlist fix went through the gate **twice**. Both rounds:
the code itself was affirmed sound by the clear majority (round 2: 7 approve),
never vetoed. Both rounds returned **REVISE**, each `decided_by` a single seat's
objection. But the objections were, on inspection:

- **Round 1 (substantive, useful):** the plan fixed the mechanism but omitted the
  switch-flip, deploy-verification, and page regeneration → genuinely incomplete.
  This is the gate working as intended.
- **Round 2 (form / operational / scoped-out):** "prove your single-caller claim
  with a formal lookup not a grep" (the claim was true), "why content_data not
  settings" (content_data is the established convention — 14 Go reads, same-file
  precedent; settings is empty and unread), "make the re-queue a concrete script"
  (operational, best validated against the live system post-roll), "the deeper
  silent-drop fault is still platform-wide" (true — but it's bugs_open/056,
  correctly kept separate). **None said the code was wrong.**

A human then had to adjudicate "the code is sound, the remaining objections are
non-blocking → proceed." **That adjudication is exactly what should have been
automated.** The gate is advisory (it cannot block), but with no adjudicator the
raw REVISE lands on the human every time a single seat raises any nit.

## Root cause in the current decider

`diagnose_council_decide_action.go` decides purely by presence of an objection
(header comment, lines ~13-16):

```
1. any hard-veto reviewer says veto        → "rejected"
2. any reviewer says veto (advisory veto)  → "rejected"
3. any reviewer says object                → "revise"   ← no severity/nature test
4. all approve                             → "approved"
```

The wire contract **already carries `severity` (low|medium|high)** on every
objection (`councilObjection.Severity`) — but the decision ignores it. One
low-severity form nit from one seat is indistinguishable, at the decision layer,
from a high-severity correctness defect.

## Proposed improvement

An adjudication step (either inside `diagnose_council_decide` or a dedicated
downstream "decision-evaluator" seat) that weighs the **nature** of objections
rather than their mere presence:

- **veto (either kind) → rejected.** Unchanged — a guardian veto still blocks.
- **any high-severity object, or ≥N medium → revise.** A real defect still forces
  a revise round, as today.
- **only low-severity / operational / already-scoped-out objections, on
  otherwise-approved code → a new terminal state** (e.g. `approved_with_notes` or
  `advise_proceed`): record the notes, do NOT force another human-adjudicated
  revise round. The notes still travel with the artifact trail.

The adjudicator must be **conservative and auditable** — it decides on the
declared severity + a cheap classifier of "is this about code correctness, or
about form/rollout/evidence?", and it must log *why* it downgraded, so the
downgrade is reviewable and can't silently rubber-stamp a real objection. It must
never override a veto, and never upgrade approve→proceed if a medium+ correctness
objection stands.

## Why this is the right altitude

This is the same lesson the gate itself teaches applied one level up: the gate
adjudicates *plans*; nothing adjudicates the gate's own *verdicts*. Today a
"revise" is a raw signal a human must interpret. An evaluator that classifies
"blocking vs advisory" is the structural fix — it lets the advisory gate be
genuinely advisory (proceed on non-blocking notes) without a human in the loop
for every form nit, while still stopping on real defects and vetoes.

## Caveats / what a reviewer of THIS feature should check

- Severity is self-reported by each seat; a seat could under-rate a real defect.
  Mitigate by classifying nature (correctness vs form) independently, not trusting
  severity alone, and by keeping the veto path untouched (a seat that believes a
  defect is blocking escalates to veto/object-high, not low).
- Must not become a path to ship unreviewed change: `approved_with_notes` still
  records the full objection set; the notes are a to-do, not a dismissal.
- Interacts with the round counter (`should_revise` routing, F2.3) — an
  `advise_proceed` terminal must not loop.
