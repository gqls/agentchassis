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

## The missing number — **MEASURED 2026-08-08**, and it does not say what either side expected

~~`[UNMEASURED]` **How often do registered verifiers actually return an error in production?** If
the answer is ~never, option 4 is nearly free and option 3 is over-engineering. If it is common,
option 4 would strand items at scale and option 3 earns its cost. Nobody has this number,
including me, and **I would not choose between 3 and 4 without it.**~~ **Measured below.**

**The rate, and it is not the useful part.** Over the whole life of the gate (live since
`e1b8e1f84`, 2026-07-14, v1.0.1116), the current `result` payloads carry **11** verifier
consultations — 8 `verified`, **2 `error`**, 1 `defect_persists`. All 11 fall between 2026-08-03
10:37Z and 2026-08-07 08:37Z.

| verifier | consultations | errors |
|---|---|---|
| `hardcoded_section_colors` | 8 | 0 |
| `empty_section` | 2 | **2 (100%)** |
| `literal_markdown` | 1 | 0 |
| the other 5 registered | **0** | — |

So: **2/11 = 18% of consultations, and 100% of one verifier's.** Neither "~never" nor "common".

```sql
-- the census. Enumerates statuses rather than asserting the three in the source,
-- so a shape change cannot hide behind a named-key read.
SELECT result->'_verification'->>'status' AS v_status, count(*)
FROM site_work_items WHERE result ? '_verification' GROUP BY 1 ORDER BY 2 DESC;
-- verified 8 | error 2 | defect_persists 1
```

**⚠ Two caveats, both of which make 2 a FLOOR and not a count.** `result` is **overwritten** on
every completion attempt (the complete path and `failUnverifiedCompletion` both write it), so this
is a **current-state census, not a history**: an item that errored and later verified now reads
`verified`. And only 3 of 8 registered verifiers have ever been consulted at all, so the sample is
`n=11` over five days. The direction below is suggestive; it is not a rate estimate.

### What the errors ARE, which decides more than how many

**Zero infrastructural errors.** No DB blip, no timeout, no malformed spec. Both errors are one
deliberate branch, `check_empty_sections.go:412`:

> `cannot verify: component %s no longer exists (genuinely fixed or silently deleted —
> indistinguishable here)`

That branch is `bugs_closed/032`'s own accepted fix, and its comment states the trade plainly:
error → fail open → *"a false success becomes a visible unknown."* **The identical branch is
duplicated in a second verifier** — `check_truncated_component.go:272` — which has never been
consulted (0 completions of that type). So the error path is armed on **2 of 8** verifiers, by two
authors, both reasoning correctly from the documented policy. This is not forgetfulness, and
**option 2 (a lint) would have forbidden a reasoned, documented choice** rather than caught a slip.

### The part that inverts the expected answer: on both occasions the honest verdict was `Resolved:false`

`bugs_closed/032` named the disambiguator itself — *"if the page still EXPECTS this component (a
`plan_sections` entry, a slot reference), absence is not ambiguous at all — it is deletion."*
**Both items pass that test**, so on both the ambiguity the error branch exists to express **was
not actually present**:

- Items `177bbb2e…` (page `ai-guides`) and `8c4b10f1…` (page `insights`), site
  `1368e337…`, both slot `featured-article`, both stamped **`complete`** 2026-08-03 with
  `attempt_count` **0**.
- Both pages **still declare the slot**: `pages.sections` = `[…, "featured_article", …]` on each
  (note the snake spelling — `pages.sections` is an array of bare strings, not objects).
- Both pages now carry a **deployed 334-byte shell** in slot `featured-content` — a `<section>`
  wrapping an empty `<h1>` and nothing else. Verbatim, so the reader need not take this on trust:

  ```html
  <section class="section section--featured"><div class="container"><article class="featured-article">
    <div class="featured-article__content"><h1 class="featured-article__title"></h1></div>
  </article></div></section>
  ```

So the defect these two items describe **is still live on two production pages**, and the gate
recorded it as `complete` with a "cannot verify" note attached. The recorded unknown is real; what
032 could not know is that it would land on the deletion horn both times it fired.

### And the backstop the policy rests on has not fired

`complete_work_item_verification.go:14-21` justifies failing open on the grounds that *"discovery
re-detection + two-strike is the backstop."* Measured, five days on:

- **No work item for a `featured-content` slot has ever existed, fleet-wide** (`item_key LIKE
  '%featured-content%'` → 0 rows). The only items for these pages' `featured-article` key are the
  two completed above, plus April rows long since parked `unresolved`.
- **Yet the detector's own predicate matches both components right now.** Running
  `findEmptySections`' SQL verbatim (`check_empty_sections.go:158-189`) against the site returns
  both, classified `empty_heading`, `locked=f`, `suppressed=f` — and `build_status='deployed'`, so
  `bugs_open/185`'s deployed-only blindness does **not** explain it.
- The `empty_sections` check demonstrably ran on this site *after* the rebuild (it retracted 10
  items at 08-03 19:41Z, 21:03Z, 08-04 08:36Z and 10:35Z, stamping
  `result.resolved_by='empty_sections'`) and filed nothing for the empty slot.

**Why detection did not re-file is NOT established here** and is deliberately left unasserted — the
dedup index excludes `unresolved`, so the April zombies do not block a new key, and `bugs_closed/041`
is cleared too (the site's four `needs_new_component` rows are `category_section`/`article_grid`
with `already_exists=f`, i.e. genuinely absent components, not 041's snake-case miss). That points
somewhere else and wants the `090` loop, not another hour of grep.

### What this does to the four options

- **Option 4 (fail-closed, per-verifier opt-out)** — the measurement **favours it, against
  expectation.** Its feared harm is stranding legitimately-removed items; **0 of 2** observed errors
  were legitimate removals, and fail-closed would have produced the correct outcome on **both**. Its
  other feared harm, transient infrastructure errors at scale, is **unobserved** (0 of 11).
- **Option 3 (`Indeterminate` / park)** — still the truest model of "could not run", but on the
  observed evidence it buys nothing over option 4: both cases had a knowable right answer, so
  parking them for a human would have been slower and no more correct. Its cost (a shared struct
  plus the completion path for every verified type) is now harder to justify on `n=11`.
- **Option 2 (lint)** — would have fired on both, and been **wrong** about why. These authors did
  not reach for `err` carelessly; they followed the documented policy and cited it.
- **Option 1 (do nothing)** — weakest it has looked: the branch is on 2 of 8 verifiers already, and
  its first two live firings both completed a live defect.

**Cheapest thing that fixes the observed cases without touching the registry:** take
`bugs_closed/032`'s own "stronger option" — check whether the page still declares the slot, and
return `Resolved:false` when it does. That is per-verifier, needs no shared-struct change, and is
correct on 2 of 2. It leaves the generic policy exactly as it is, which is what this RFC is about,
so it is a **complement to a decision here, not a substitute for one.**

**Measured by** the `bugfix_201_page_content_writer_dispatch` lane, 2026-08-08. Every figure above
is a live query; the queries are in that lane's `RUNBOOK` (R8) and its `NOTES` records the two
wrong turns taken getting here.

## Evidence

- `platform/orchestration/actions/discovery_checks/verifiers.go:60-63` (the policy, in the source).
- Council report, corr `f14a8b64-4f71-4915-88d0-9587db845052`, r2 — `bug_historian` [medium] and
  `architecture` [medium], both independently naming it; r1's gating [HIGH] on the same mechanism.
- `bugs_closed/032` — *"verifier reads a deleted target as a successful fix"*, the same shape,
  already closed once.
- `bugs_open/201` symptom 2 and this lane's `NOTES` 2026-08-06/07 — the near-miss in full.
- `verifier_coverage_test.go` — the existing guard, and the natural home for option 2.
