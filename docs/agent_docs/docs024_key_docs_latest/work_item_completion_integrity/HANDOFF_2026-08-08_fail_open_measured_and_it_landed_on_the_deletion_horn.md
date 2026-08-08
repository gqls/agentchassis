# POINTER — the fail-open policy you own has now fired twice in production, and both times it completed a live defect

**2026-08-08, from the `bugfix_201_page_content_writer_dispatch` lane.** A pointer, not a
substance-grab: this thread owns `CompleteWorkItemAction`, so the decision is yours. Nothing here
asks you to change anything today.

## What changed about your premise

`bugs_closed/032`'s accepted fix — return an error so the gate fails **open** and records
`_verification.status='error'` instead of a false `Resolved:true` — is live and working exactly as
designed. Its own words: *"Item flow is unchanged — the item completes either way. What changes is
that a false success becomes a visible unknown."* That was honest and it is still true.

**What is new is that the unknown has now been observed, twice, and it was not ambiguous either
time.** 032 named the disambiguator itself:

> *"if the page still expects that component (a `plan_sections` entry, a slot reference), absence is
> not ambiguous at all — it is deletion, and could return `Resolved: false`. That is a better answer
> and a bigger change; the error-return above is the safe floor."*

Both live cases meet that condition:

| | |
|---|---|
| items | `177bbb2e-2f5f-43a9-ace0-56f4d141de90` (page `ai-guides`), `8c4b10f1-6a82-4b8d-8530-36aa1550e5e8` (page `insights`) |
| site | `1368e337-dd1d-4799-bbb3-8221a1b79bcc` |
| outcome | both **`complete`**, 2026-08-03, `attempt_count` **0**, `_verification.status='error'` |
| the page still expects it | `pages.sections` lists **`featured_article`** on both (snake spelling; the array holds bare strings) |
| the defect now | both pages serve a **deployed 334-byte shell** in slot `featured-content` — a `<section>` around an empty `<h1>`, no text |

So `Resolved:false` was the honest verdict on 2 of 2, and the gate completed both.

## And the backstop the policy rests on has not fired

`complete_work_item_verification.go:14-21` justifies fail-open on *"discovery re-detection +
two-strike is the backstop."* Five days on:

- **No work item for a `featured-content` slot has ever existed, fleet-wide.**
- **The detector is not blind to it.** `findEmptySections`' SQL (`check_empty_sections.go:158-189`),
  run verbatim against the site, returns both components right now — `empty_heading`, unlocked,
  unsuppressed, and `build_status='deployed'`, so `bugs_open/185` does not explain it.
- The `empty_sections` check ran on that site four times *after* the rebuild (retracting 10 other
  items at 08-03 19:41Z / 21:03Z and 08-04 08:36Z / 10:35Z) and filed nothing for the empty slot.

**Why it never re-filed is not established, and I am deliberately not asserting a cause.** Two
candidates are ruled out by measurement: `idx_swi_dedup` excludes `unresolved`, so April's parked
rows cannot block a new key; and `bugs_closed/041` is cleared (the site's four
`needs_new_component` rows are `category_section`/`article_grid` with `already_exists=f` — genuinely
absent components, not 041's snake-case miss). That is a `090` job.

## What this is evidence for, and what it is not

It is evidence that **the "safe floor" is load-bearing more often than it looks**, and that the
"stronger option" 032 deferred to you is now the cheapest correct fix for the observed cases: check
whether the page still declares the slot, return `Resolved:false` when it does. Per-verifier, no
shared-struct change, correct on 2 of 2.

It is **not** evidence about the general registry policy at scale: `n=2` errors out of 11
consultations ever, `result` is overwritten on each attempt so `2` is a floor rather than a count,
and 5 of the 8 registered verifiers have never been consulted. The full measurement, with those
caveats attached and what it does to each of the four options, is in
**`architecture_review/RFC_017_verifier_registry_fails_open_on_error.md`** (§ "The missing number —
MEASURED 2026-08-08"). RFC_017 is an owner decision and remains open.

Queries: `bugfix_201_page_content_writer_dispatch/RUNBOOK_…` **R8**, including the three ways this
census gives a confident wrong answer — one of which I walked into and one of which I nearly
published.

---

> ## UPDATE, same day — **the owner ruled, and the policy you own has been flipped.**
>
> This file was written as a measurement with no ask attached. Hours later the owner took the
> decision on `RFC_017`: **fail-CLOSED by default.** Built and committed the same day; **inert until
> the next chassis roll**, which is whole-fleet and the owner's to run. Council gate corr
> `a104d454-a4ff-4c95-a578-9a7e48c95100`.
>
> **What changed in code you own** (`complete_work_item_verification.go`, `verifiers.go`):
> - A verifier error no longer completes the item. It routes into the attempt machinery via
>   `failUnverifiedCompletion`, exactly as a persisting defect does.
> - Fail-open survives only as an **explicit per-item_type opt-in** —
>   `RegisterVerifierWithPolicy(t, v, VerifierPolicy{FailOpenOnError: true})`, unsafe default OFF per
>   the 2026-08-02 shared-seam ruling. **No verifier opts in today.** `RegisterVerifier` keeps its
>   signature, so all 8 registrations became fail-closed without their files being touched.
> - `GetVerifier` now returns `(ItemVerifier, VerifierPolicy)`. One caller, already updated.
> - **A message you own was wrong the moment errors could block:** the blocked-completion text was
>   hard-coded to *"post-fix verification found the defect still present"* and read from
>   `payload["detail"]`. On an error payload that claims a finding the verifier never made, with an
>   empty body. `blockedCompletionReason` now returns `verification_unavailable` vs
>   `verification_failed` with the right text.
> - The payload gains `fail_open`, so a census can tell a completed error from a blocked one. The one
>   that produced RFC_017 could not.
> - The **unparseable-spec** branch takes the same policy deliberately — same class, and exempting it
>   would leave a second silent completion path behind the one this closes.
>
> **What is now yours to weigh.** The retry cost is real and named: a structurally-unanswerable
> verifier now burns up to `max_attempts` (3) page rebuilds before a human sees the item. RFC_017's
> option 3 — a third `Indeterminate` outcome that parks without burning attempts — was **explicitly
> not taken**, and the owner's stated reasoning is that this cost is the evidence that would justify
> it. It stays open in RFC_017. If you see that cost in the numbers after the roll, that is the case.
>
> Register entry: **`WII-011`** (`work-item-integrity.md`), landed in the same commit as the seam per
> the platform-seams ordering exemption. Verification recipe is in its `verify-later`.
