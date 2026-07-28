# HANDOFF — architecture seat, continue here (2026-07-28, late evening)

**COLD-START ENTRY POINT. Supersedes `HANDOFF_2026-07-28b_continue_here.md`**,
which is still correct about everything except §2 (layer 1b is no longer "in
council" — it is PAUSED on an owner decision) and §4 (whose lesson now has a
limit, see §3 below). Go to `HANDOFF_2026-07-27_continue_here.md` §5 and `…-27b`
§5 for the Go-contract landmines, unchanged and still the most expensive here.

Prose: `README_where_we_are.md` (owner's log, append-only) ·
`NOTES_architecture_seat.md` (technical log + every misstep) ·
`DECISIONS_open_for_owner_2026-07-26_architecture_seat.md` (**§ D12 is the live
one**) · `RUNBOOK_architecture_seat.md`.

---

## 1. State

| thing | state |
|---|---|
| D11 layer 2 (routing) | **LIVE** |
| D11 layer 1 (symbol bodies) | **LIVE & PROVEN** — 4,992 rows / 4,992 bodies, survived `v1.0.1194` |
| **D11 layer 1b (markdown)** | **PAUSED — blocked on owner decision D12.** NOT revising, NOT resubmitting |
| `bugs_open/135` (prune has no floor) | **OPEN, UNOWNED** — pre-existing, independent of 1b |
| `review_architecture` | **still 0 reviews** — rate limit, not fault |

**⚠ REVERSAL TRIGGER, unchanged:** migration 252 pins the indexer's ref to
`086_experience_loop`. **Change that literal to `'main'` AS PART OF the merge.**
A no-rows `pre_query` makes the scheduler SKIP the task — no fallback, silent.

## 2. Layer 1b: rounds 7 and 8, and why it stopped

`SUBMISSION_CORR=7ba5b8c4-0e10-46db-9fc4-2bd0584e943a`. Round 7 plan at
`SUBMISSION_2026-07-28_markdown_into_the_index.json`, round 8 at
`SUBMISSION_2026-07-28b_…json` (new file, so the 7→8 diff is auditable).

Eight rounds: **5/3 → 5/3 → 8/2 → 10/1 → 6/4 → 9/3 → 8/2 → 8/1.**

**Round 8 proved the method.** Round 7's six objections were *all* requests for
measurement, not design faults. Round 8 changed **no code, no globs, no schema, no
helper** — it added eight query results to `grounded_in` and one missing
verification step. Every seat that objected in round 7 approved in round 8 with
zero objections (`guardian` ×2 → 0, `reuse_agent` → 0, `prior_art_librarian` ×2 →
0, `debug_historian` → approve).

**It stopped on one `medium` from `diagnosis_guardian`, and that objection is
correct.** Widening `kind` to admit `'doc'` lets a markdown excerpt satisfy the
**static tier** of a CONFIRMED verdict — the tier that means *code was read*.
Five live-code sites in `DECISIONS… § D12`; the sharpest is
`diagnose_assemble_bundle_action.go:266-270`, which prints
`## Code search results (… STATIC tier)` and the literal sentence **"They are
CODE"** over everything the index returns.

**Nothing is broken today** — 0 markdown rows, the CHECK still refuses them.

## 3. THE TWO RULES THIS SEQUENCE PRODUCED — and the limit on the first

`…-28b` §4 said: *when a plan's objections migrate onto machinery added during
review, split it — do not iterate it.* That is still right, and round 7's split
(→ `bugs_open/135`) is still the worked example. **Round 8 found its limit:**

> **You may only split out what review surfaced that is NOT YOURS.** The prune had
> no floor with or without layer 1b — pre-existing, independent, someone else's
> bug that my call site exposed. D12's hazard **does not exist until layer 1b
> ships**. Filing a bug for a defect your own unshipped change creates, and then
> shipping the change, is not a split — it is shipping a known defect with a paper
> trail.

And the second, which round 8 is the cleanest evidence for anywhere in this repo:

> **Answer an evidence objection with EVIDENCE, never with code.** Rounds 4–6
> answered by building and never converged (5/3 → 6/4 → 9/3). Rounds 7–8 answered
> with queries and went 8/2 → 8/1, clearing every objecting seat in one round.
> Each mechanism you add in response to review becomes fresh surface to object to.

## 4. What the next thread does

**Do NOT resubmit layer 1b.** It is not blocked on better writing — for the first
time in eight rounds. It is blocked on **D12**, which is a design question about
what an evidence tier *means*, with four options costed in the DECISIONS file.

The cleanest fix (D12 option 1b) touches `diagnose_assemble_bundle_action.go` —
**the diagnosis loop's citation contract, blast radius every diagnosis run.** That
is a platform seam; the 2026-07-28 owner ruling makes it architecture-scope, and
the ordering-constraint exemption **does not apply** (nothing is broken, nothing is
waiting on a binary). Answering it inside round 9 would repeat rounds 4–6 exactly,
with the citation contract as the accreted mechanism, and the `guardian` that
approved round 8 would be right to veto it on scope.

So: **put D12 to the owner.** When it is answered, fold in `debug_historian`'s
one-line low objection (a defensive `ROLLBACK;` preamble before the migration's
`BEGIN;` — an aborted transaction from a prior failed run in the same session is
sticky) and resubmit on the same `RESUBMIT_CORR`.

## 5. Landmines earned in rounds 7–8

- **A count you report must be the count the tool RETURNED, not the count you
  KEPT.** Rounds 6–7 said the branch-on-`kind` grep returned "one hit". It returns
  four (none is the column — three are the `code_check` kind namespace, one is the
  analyser's `typeDef.Kind`). I had discarded the irrelevant hits in my head and
  reported the remainder as the census. **The conclusion was right, which is what
  makes it dangerous:** a wrong conclusion gets caught; a right one propped up by
  an uncounted count gets quoted forward. Logged in `WRONG_CALLS.md`.
- **A reviewer saying "I cannot verify this" is not a weaker objection than "this
  is wrong."** `guardian` never disputed the conclusion — it said a human grep is
  something its own tier cannot re-check (`code_checks` see declarations, never
  switch bodies) and asked a *different* question instead. Going to answer that is
  what exposed the miscount. Concede the epistemic limit in the resubmission; the
  honest form is shorter and more persuasive than a defence.
- **The same measurement can answer one question and raise another.** "No reader
  switches on `kind`" is a blast-radius reassurance AND a citation-integrity
  problem. `diagnosis_guardian` re-framed round 8's own new evidence rather than
  disputing it. That objection was not reachable in rounds 1–7.
- **`rag_lookup` is wired to `rag-test-agent` and nothing else** — no diagnostic or
  council seat can reach `knowledge_base`. Useful whenever someone proposes the RAG
  surface as a home for anything the seats need to find.
- Unchanged from `…-28b`: a wedged council run looks exactly like a queued one (the
  exit condition is the peer check); **always read `unreadable`, never
  `abstained`** — round 8 had 1 unreadable and 6 abstained of 9.

## 6. Still open, unchanged

1. **The seat still has 0 reviews.** `review_architecture` sits only on
   `feature-designer`, which refuses anything without an owner-approved
   `capability_gap` spec. **Do NOT manufacture a review by firing at another
   thread's ticket.**
2. **`council-gate` still gets no code answers** — deliberate (no reproposer); the
   fix is surfacing code results into its verdict note.
3. **D11 layer 3** — a seat cannot look things up *while reasoning*. `[UNSCOPED]`.
4. **Two orphaned council orchestrations from ~16:55 on 07-28** never recovered
   across a roll. `[UNDIAGNOSED]`, not chased.
