# SUMMARY — 2026-07-31b — chrome build path: approved, live, and the approved version was the smallest one

*(A second summary the same day because three of the five headings genuinely
changed after the first was written: the guard it describes as shipped has been
deleted, the council verdict it lists as pending is in, and the fix it calls "not
live" is live and pod-verified. The first summary is left exactly as it was.)*

## What we're trying to do

Make the platform give one answer to one question — *which library component may
serve a site's header, footer or head?* — at every place that asks it, and make a
wrong answer unrepresentable rather than merely unlikely.

## Where we've come from

`bugs_closed/118` built the predicate and fixed the two places that *assign* chrome.
It left the three that *build a page* and filed them as `bugs_open/167`, parked as an
owner decision on the grounds that fixing them would change header and footer markup
fleet-wide.

That premise had already expired when it was written. Both predicates return the same
component today, because 118's own fleet repoint activated `header-theme-chrome` at
12:39 that afternoon — hours before it filed the bug. Re-running the measurement,
rather than quoting it, is what turned an owner decision back into an ordinary fix.

## What we've done

The three build-path renderers now resolve through `ResolveChromeComponent` and use
its answer **only when it reports the component eligible**. That gate is the fix, not
a detail: the resolver always returns something, and for `<head>` — where the library
has nothing eligible at all — that something is an 8,500-character page section. The
"one line each" change the bug file describes would have put a page section into every
page's `<head>`.

Alongside it, the door: a scan that fails if a chrome function name is handed to the
section-shaped lookup again, covering the file that 118's own scan structurally
exempts.

**The council took three rounds, and the version it approved was smaller than the
version it rejected.** Round one objected that we had found a fourth unguarded path
and shipped anyway, so we added a per-render error log for it. Round two rejected
*that*, on four independent grounds — no reader, unbounded noise, a home-made reporter
where a platform mechanism already existed, and a swallowed diagnostic as the only
gate. Round three deleted it and passed, 11 seats of 13.

**Deleting it is what produced the best finding of the lane.** Going to look at the
mechanism we should have reused showed that `deactivated_site_components` — the
platform's own detector for "this site is using a switched-off component" — joins
`site_components` only and **has never looked at style-collection pins**. That blind
spot is precisely why three deployed sites have been serving a deactivated header with
no work item, no finding and no alert. It is worth more than the guard was, and it
would not have been found if the guard had been let through.

## Where we are now

**167 is fixed, council-approved and LIVE** on `v1.0.1225`, pod-verified on both
replicas with the new strings present, a positive control intact and — the
load-bearing one — a **negative** control proving the old code is gone rather than
merely accompanied. It therefore meets this repo's *fixed AND live* bar on its own
terms, closing the gap left when it was moved to `bugs_closed/` ahead of the roll.

Two things are honestly unfinished, and both are written down rather than smoothed
over:

- **`bugs_open/170`** — the style-collection pin path still dereferences by id with no
  eligibility predicate of any kind, and three deployed sites are on a switched-off
  header. Two council seats still hold, advisorily, that deleting the guard left that
  path with *no* signal at all. They are right that it did. The counter-argument is a
  judgement, not a proof, and 170 tells whoever picks it up to treat those objections
  as the brief.
- **The verification command we first published was wrong** — it grepped the wrong
  case and returned a confident false "not shipped" on a binary that had the fix. It
  is corrected in the runbook and the bug file, and the general form is now a
  fleet-wide landmine: a *positive* control proves the pipeline works and can never
  prove your pattern is spelled right; only a *negative* control can.

## Where we're going

Nothing further is owed in this lane. Three things sit downstream, all filed, none
ours: `bugs_open/170` (the pin path, needing a decision on three live sites);
`bugs_open/166` (the repair that could not repair); and `bugs_open/149`'s
page-rerender queue, which is why a fix in the binary and a page on the internet can
still disagree for days.
