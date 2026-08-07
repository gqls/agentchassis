# RFC 017 — the completion-verifier registry FAILS OPEN on error, so a verifier that cannot run waves the item through

**Filed 2026-08-07** by the `bugfix_201_page_content_writer_dispatch` lane.
**Raised independently by TWO council seats in the same APPROVED round** — `bug_historian`
(medium) and `architecture` (medium), correlation
`f14a8b64-4f71-4915-88d0-9587db845052`, r2, 15 reviewers, 2 abstained. The architecture seat's
words: *"Worth a standalone architecture item, not a blocker on this fix."* This file is that
item. **`bugs_open/201` is not blocked on it.**

It is also the answer to that seat's own r1 MISSING note — *"whether verifiers.go's
fail-open-on-error policy has been raised to the council as its own architecture item, separate
from per-verifier coverage gaps — a human should see this even though the author scoped it out."*
It had not been. It has now.

## The policy

`verifiers.go:60-63`, stated plainly in the source and therefore deliberate:

> `ItemVerifier` re-checks the defect described by a work item. Returning an error means the
> verification could not run at all — **the caller decides policy (CompleteWorkItemAction fails
> open and records the error in the result).**

So: **verifier errors → the item is stamped `complete` anyway.** The error is recorded; the
completion is not withheld.

## Why this is architecture-scope and not a bug in one verifier

The whole point of the registry (`bugs_open/021` §INSTANCE 2, `bugs_open/017`) is to stop a saga
that *says* it succeeded from being trusted. **The error path re-opens exactly that hole**, one
level down: "I could not check" is treated identically to "I checked and it is fixed."

It is generic. **Seven verifiers are registered today** — `truncated_component`,
`hardcoded_section_colors`, `empty_section`, `orphan_element_refs`, `content_duplication`,
`page_canonical_collision`, `dead_fragment_link`, plus `literal_markdown` as of this lane — and
every one inherits it, as will every future one. A verifier author who reaches for
`return VerifyResult{}, fmt.Errorf(...)` on an ambiguous case — the natural, cautious-feeling
thing to write — has silently chosen "complete it".

## The worked case, which is this lane's own near-miss

`bugs_open/201` symptom 2 is *a handler reporting success having written nothing*. Fixing it, I
wrote `VerifyLiteralMarkdownResolved` and gave its zero-rows branch an **error**: a page with no
scannable components cannot be distinguished from one whose content was **lost**
(`bugs_closed/194`'s class — 31 of 106 components NULL on one live site), so refusing to answer
felt like the honest move.

Under this policy that branch **stamped the item `complete`** — the precise outcome the verifier
existed to prevent, on the one input where the ambiguous case *is* content loss. `bug_historian`
gated the round on it and named the precedent: **`bugs_closed/032`,
"verifier reads a deleted target as a successful fix"**. So this shape has already shipped once,
been closed, and re-appeared.

I had *noticed* the fail-open and written "I am aware this means the caller fails open on that
path" into the submission — and shipped anyway. **The local fix was to return `Resolved:false`
instead**, which blocks completion. That is correct for `literal_markdown` and is now live in
code. It is also, exactly as both seats say, **routing around the policy rather than addressing
it.**

## The tension, honestly stated — fail-closed is not obviously right either

This is why it is an RFC and not a patch.

- **Fail open** loses the guarantee whenever verification is merely *broken* (a DB blip, a
  malformed spec). The item completes unverified and the only trace is a recorded error nobody
  reads.
- **Fail closed** means any verifier bug, transient DB error, or unparseable spec **burns the
  item's attempts and strands it in `failed`** — the `page_rerender` harm (1,849 items) arriving
  by a different door. A registry-wide flip would apply that to every verified type at once.

The interesting question is whether "could not run" deserves to be a **third outcome** rather
than being folded into either — an item that is neither completed nor failed but *parked* for a
human, with its attempts untouched.

## Options, costed

1. **Do nothing; keep the per-verifier discipline.** Cost: every future verifier author must
   independently rediscover that an error means "complete it". This lane's author did not, with a
   council seat gating the round to catch it, and `bugs_closed/032` says the estate did not
   before that either. **Two independent misses is the argument against this option.**
2. **A lint/test that forbids `return VerifyResult{}, err` for a non-infrastructural case** —
   cheapest real guard, in the spirit of `verifier_coverage_test.go`, which already source-scans
   this package. Catches the author-forgetfulness half; says nothing about genuine DB failures.
3. **A third outcome (`Indeterminate`)** — `VerifyResult` grows a state meaning "not verified,
   do not complete, do not burn an attempt", and `CompleteWorkItemAction` parks the item.
   Truest to the problem, widest blast radius: it changes a shared struct and the completion
   path for every verified item type. **Would need its own round.**
4. **Flip the default to fail-closed, per-verifier opt-out.** Middle cost. Risks the
   `page_rerender` harm on any verifier whose error path is noisier than its authors think, and
   nobody has measured how often verifiers error in practice — **which is the missing number
   below.**

## What is NOT established, and should be measured before choosing

`[UNMEASURED]` **How often do registered verifiers actually return an error in production?** If
the answer is ~never, option 4 is nearly free and option 3 is over-engineering. If it is common,
option 4 would strand items at scale and option 3 earns its cost. Nobody has this number,
including me, and **I would not choose between 3 and 4 without it.** The recorded errors are in
`CompleteWorkItemAction`'s result payload — that is where the census starts.

## Evidence

- `platform/orchestration/actions/discovery_checks/verifiers.go:60-63` (the policy, in the source).
- Council report, corr `f14a8b64-4f71-4915-88d0-9587db845052`, r2 — `bug_historian` [medium] and
  `architecture` [medium], both independently naming it; r1's gating [HIGH] on the same mechanism.
- `bugs_closed/032` — *"verifier reads a deleted target as a successful fix"*, the same shape,
  already closed once.
- `bugs_open/201` symptom 2 and this lane's `NOTES` 2026-08-06/07 — the near-miss in full.
- `verifier_coverage_test.go` — the existing guard, and the natural home for option 2.
