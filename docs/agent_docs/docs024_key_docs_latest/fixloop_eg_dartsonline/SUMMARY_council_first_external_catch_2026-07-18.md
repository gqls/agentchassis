# The council gate's first catch on someone else's change (2026-07-18)

*One page. Written by the imagery thread, which submitted D14 to the gate and
was told to revise. The point of the record: this is the first submission from
a thread that did not build the council, so it is the first evidence about
whether the gate is worth its credits to an outsider.*

## What was submitted

Imagery D14 — per-article "content heroes" moved to their own kind routed to a
provider that honours style anchors, plus a per-kind style-guide override map,
plus an eligibility predicate stopping a listing from showing 404s. Five files
across `platform/` and `internal/`, already committed (`4e35c8064`) and live:
the CLAUDE.md council section landed mid-session, after the commit, so the
submission was **retrospective and said so**.

## What it cost and what came back

Three rounds, all returning a verdict: **revise** (6 approve / 3 object) →
**revise** (8/1) → **revise** (7/2).

**Three real defects, none of which the author had spotted:**

1. **`jsonb_array_length(p.sections) > 0` was unguarded** — raised
   independently by five seats, one at HIGH. On an object-shaped value the
   function raises a Postgres ERROR, so a *single* malformed row would abort
   the whole discovery sweep for that site rather than skip one page; on NULL
   it returns NULL, which is silently falsy — a silent drop. Fixed with a
   `jsonb_typeof(...) = 'array'` guard. (All 269 live pages are array-shaped,
   so this was latent — but latent in the "one bad row takes a site down"
   sense.)
2. **The eligibility predicate existed as two hand-copied strings** in the
   listing resolver and the imagery check, bound only by comments — four seats
   named it as the dedup-index ↔ `workItemTerminalStatuses` drift shape this
   platform has already paid for. Fixed by making it one exported constant,
   `queryresolve.ListedPageEligibilitySQL`, imported by both.
3. **A genuinely missing guard**: `referenceKeysForKind` lacked the logo lock
   that `directionForKind` and `avoidForKind` both had, so a stray
   `kinds.logo` entry could have fed style anchors to an approved-and-frozen
   logo. Fixed and pinned by test.

Plus one architecture note worth keeping: the adapter's provider-routing
switch is fail-*silent* — a new kind nobody adds to it loses reference-anchor
support with no error anywhere, which is exactly how the original bug shipped.
A WARN now makes that visible. All fixes: `358e14af6`.

## What the rounds cost in signal

Each round also produced objections that were **evidentiary, not defects** —
"the plan shows two sites, show the fleet", "did you check the existing
`sites.style_overrides` column?". Both were worth answering (fleet-wide there
are exactly three consumers and only one changes; `style_overrides` turned out
to be a dead column with zero Go references and zero populated rows), and
answering them cleared the guardian's objection in round 3. But they are the
reason a submission converges slowly: **the gate reviews the plan document, not
the repository**, so anything the plan does not *state* reads as unchecked even
when it was checked.

## Honest verdict on the gate

**Worth it.** The `jsonb_array_length` catch alone justifies the run: it is a
whole-site outage waiting for one malformed row, in code that passed review by
its author, tests, and a build.

**Two things to know before you submit:**

- **Front-load the evidence.** Put the fleet-wide counts, the STEP-ZERO reuse
  check, and the pod-verification you performed *in the plan*. Half the
  objections across three rounds were the plan failing to show work that had
  been done.
- **Approval is not guaranteed and should not be chased.** This submission
  stopped at three rounds with the code-level findings all fixed and two
  disclosed items surviving (a deliberately out-of-scope structural fix, and a
  plan-hygiene note). It was committed **without** a `Council-Reviewed:`
  trailer, because the verdict was not APPROVED and the 098 report buckets
  false trailers separately on purpose. A revise you have genuinely acted on is
  a good outcome; a trailer you did not earn is a lie in the audit trail.

**Trail:** submission correlation `098b29b8-9a57-4cbd-8d2f-7f21c7495b0e`
(3 rounds on the one correlation); fixes in `358e14af6`; the change itself in
`4e35c8064`, live and pod-verified on v1.0.1135.
