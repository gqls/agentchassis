# SUMMARY — 2026-07-31 — chrome component selection

## What we're trying to do

Make the platform give one answer to one question: *which component from the
library should serve a site's header, footer or head?* It had three answers.

## Where we've come from

On 27 July the relojistas lane hit something that looked like its own mistake. An
owner-approved fix to a site footer was applied to the footer component marked
active, the chrome was fully rebuilt, and no page changed. They dug in and filed
`bugs_open/118`: the code that picks a chrome component never looks at whether a
component is switched on. It takes whichever name sorts first. Three of the five
footers in the library are switched off, and the alphabetically-first one — a
switched-off one — was winning.

The bug then sat for four days, because the file said the fix "changes the
rendered footer on every site" and therefore needed the owner. That belief is what
parked it, and it turned out to be wrong.

## What we've done

Verified it first-hand, and the verification changed the fix three times.

There are not one but **three** places asking the question, and all three answer
differently. One has no filter at all and picks a deactivated component. One
filters on active-only, which means it would hand one client's *private forked
header* to every other site — because a fork carries its parent's function name.
The third filters correctly but does not sort its results, and what it returns is
a *page-section* component that happens to share the chrome name. Meanwhile the
correct predicate has been sitting in the codebase the whole time, in the section
selector, and nobody had copied it.

So the predicate is now one named string — active, not a fork, and actually a
chrome-level component — used by both places that assign chrome to a site, with a
test that fails if anyone writes a fourth copy. Committed, and **approved by the
council at round one** with five advisory objections, of which three earned code
changes and the rest were answered with measurements.

And the parked belief was measured rather than argued: the selection code only
runs for a site that has never been assigned a chrome component. All fourteen real
sites were assigned one long ago. **Live blast radius: one site — loancalculator,
created the day before. Zero pages re-rendered.** No owner call needed.

## Where we are now

The fix is committed and council-approved, and **inert until a chassis image
rolls**. The bug stays open until then, per the standing bar: a fix that has not
shipped leaves the defect reproducible.

Two things were deliberately not done, and both are now filed rather than left as
informal questions — which is precisely what three council seats asked for:

- **`bugs_open/166`** — the platform has been *detecting* this since 17 July. It
  raises a ticket saying "this site's footer points at a deactivated component"
  and routes it to a job that re-renders... the same deactivated component. The
  ticket can never be satisfied; two sit marked "unresolved after 2 attempts".
  Fixing it means repointing eleven sites, which changes how they look.
- **`bugs_open/167`** — the page-*building* path still resolves chrome through the
  unfiltered lookup, so it can still render a page-section component as a site
  header. Fixing it changes markup on every page built from now on.

Both are fleet-visible, which is exactly why they are separate from a fix that
changes nothing visible.

## Where we're going

One decision is owed to the owner: **do we repoint the eleven sites onto the
active footer?** It is a small database change plus a chrome re-render per site,
and it would let tickets that have been stuck since July close for the first time.
The risk is purely that the active footer looks different from what those sites
have been showing, so the way to do it is one site first, before and after.

Everything else in this lane is done. The next event is the chassis roll, after
which the fix is verified at the pod (not at git, not at the tag) and 118 moves to
`bugs_closed/`.
