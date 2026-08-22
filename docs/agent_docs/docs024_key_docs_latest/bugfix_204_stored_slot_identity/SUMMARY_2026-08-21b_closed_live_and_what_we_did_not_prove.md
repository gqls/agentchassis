# Closed — live, verified, and honest about the one thing we did not prove

**2026-08-21 (second read-out, same day).** Written to be read aloud. The first
summary was written when the code was committed but not yet running; this one is
written after the release, and the state has genuinely changed.

## What we're trying to do

Stop the site planner deleting the record of what a page is made of — and, separately,
stop anything at all overwriting that record with a blank.

## Where we've come from

Some sites we took over have pages that were chopped into pieces, and those pieces are
named by position: "the first block of prose", "the second calculator". A checker that
runs over every new site plan deletes any section name it doesn't recognise, and it only
ever knew how to recognise catalogue names. So it deleted all of them. On 20 August one
routine replan emptied the composition record of **41 of one site's 45 live pages**.

Two of the three places with this blindness had been fixed in early August. The third —
the one that *deletes* rather than postpones — had not.

## What we've done

Fixed all four places that had it (there turned out to be four, not one), left the one
place where deleting is genuinely correct, and added a second, independent protection so
that an empty section list can no longer overwrite a real one whatever causes it.

Both changes went through the review council and both were approved. **Both reviews still
found real defects**, which is the argument for the council in one line:

- The first found that our own failure-handling left no permanent trace — a run that kept
  everything because the database was unreachable would have looked exactly like a clean
  run. That is the disease we were curing, reproduced inside the cure.
- The second objected that our list of "other write paths that are safe" was *asserted
  rather than measured*, and cited three past cases where this same guard was fixed for
  one path and found incomplete on a sibling weeks later. It was right: one of the five
  was not safe, and it was a path where somebody had already fixed the *identical*
  omission on the *identical* database statement for a neighbouring column and left ours.

## Where we are now

**Closed.** Live on the current build, proven in the running program itself — and proven
with a control, meaning we also checked that our test could come out negative, because a
check that always says yes says nothing.

The protection is proven against the real database, which our normal tests structurally
cannot do: we ran the real instruction against the live database inside a transaction and
undid it. All four cases behaved correctly.

Coverage is measured rather than argued: of the section names at risk across the estate,
**every single one that should be kept is now protected**. The handful that are not are
names referring to nothing at all — no component matches them and the pages don't have
sections by those names — so deleting those is the checker doing its actual job.

**The one thing we did not prove, and it is stated at the top of the closed file rather
than buried:** the fix has not yet run for real. No site planner has run anywhere since
the incident, so the current clean readings prove nothing. Making one run means triggering
a replan on a site with these positional names, and every such site was being actively
worked that day — triggering a replan there is exactly what caused the original incident.
We were not prepared to repeat it to test our own fix. The team working the most affected
site has been told what to check when they next replan it.

We also corrected a mistake of our own along the way: the query everyone had been using to
count the exposed pages — including the one written into the original bug report — has
been over-counting since the day it was filed, because it compares names exactly while the
real code first tidies them up. That has been fixed at source, in the runbook, so the next
person doesn't inherit it.

## Where we're going

Nothing on this lane. Two checks remain and both will happen naturally the next time
anyone replans one of those sites: confirm the protection actually fired, and — this is
the important one — **prove the detector still works before trusting any clean reading.**
A broken detector and a working fix produce identical numbers, and the whole lane is a
lesson in how easily that fools you.
