# SUMMARY — 2026-08-02b · loancalculator.co.uk is now ours to rebuild

*Written to be read aloud. Second summary of the day, and a separate one because
the state it describes is different: `SUMMARY_2026-08-02_the_site_becomes_editable.md`
records the site being taken apart, this records it being handed to the loop.*

## What we're trying to do

Take a hand-built site of twenty-seven pages and twelve working calculators and
end up with something the platform can improve on its own — restyle it, rewrite
the prose, add pages — without any calculator ever quietly starting to produce a
different number.

The owner's words remain the specification: it must "evolve and improve like the
other sites will, just as long as it starts similarly enough with working tools."
Both halves of that sentence are load-bearing, and this step is where they finally
pull against each other.

## Where we've come from

We adopted the site frozen: each page one complete document, served back
byte-for-byte. Safe, and useless for the goal — nothing editable, nothing
restylable.

We then rewrote all eleven distinct calculators as components and proved each one
computes what the original computed. Earlier today we took every page apart into
its constituent pieces: **sixty-three of them across twenty-seven pages, fifty-one
blocks of prose and twelve calculators**, with nothing frozen left. All twenty-seven
rebuilt pages went live and every one matched, byte for byte, a prediction written
before any of it was stored.

But the site was still marked **ours-to-hold**. That flag is what stops the
platform's improvement machinery from touching a site, and while it was set, all
the decomposition had bought was the *ability* to edit — nothing was going to.

## What we've done

Flipped all twenty-seven pages to **ours-to-rebuild**, and locked the twelve
calculators.

Before flipping, we read what the flag was actually holding back, because it turns
out to be holding back two specific things, both named after incidents:

- The planner treats an ours-to-hold page as off-limits and files it for a human
  instead of rebuilding it. Without that, a page whose plan says it is missing gets
  handed to the generic page builder — which is documented to produce "a
  widget-less prose page where an interactive tool belongs".
- The section-saver refuses outright to touch an ours-to-hold page, because the
  way it saves is to delete every piece of a page and re-insert them.

Both of those would land squarely on our calculators. So three things were
measured rather than assumed:

**There is no plan for this site.** Zero rows. The planner iterates over a plan's
pages, so it currently has nothing to act on. That risk is real but dormant, and
wakes up when a plan is created.

**Nothing is scheduled to walk sites.** Of twenty-six enabled scheduled jobs, the
only one that touches sites dispatches work that already exists — it does not
create any. So the flip is not a starting gun. Like every step before it, it does
nothing until something makes work for this site.

**The section-saver's delete respects locks.** This is the fact the whole approach
turns on: its delete statement carries "and this row is not locked". A locked piece
survives, and the blocked write raises a review item rather than disappearing.

So the twelve calculators carry a permanent lock, and the fifty-one prose blocks
do not — because being rewritable is the entire point of having taken the pages
apart. We proved the lock bites rather than assuming it, running the exact
condition the delete carries: on the standard calculator page, all five prose
blocks come back writable and the calculator comes back not.

The site is byte-for-byte unchanged by the flip. A policy change is not a render.

## Where we are now

```
pages                27, all ours-to-rebuild
prose blocks         51, all writable by the loop
calculators          12, all permanently locked
live and verified    27 of 27, byte-identical to prediction
calculators correct  12 of 12 against a baseline that is itself
                     clean against the hand-built original
```

The site is open to the improvement loop for everything that is text, and closed
for everything that is arithmetic. That is the sharpest expression we can give of
"evolve and improve, with working tools" — and it is enforced by the database
rather than by anyone remembering.

## Where we're going

The lock has a cost we should name rather than discover: the queue of real defects
the rewrite surfaced — money shown to three decimal places, a car-finance tool that
computes nothing at zero per cent, a consolidation checker that counts a debt
towards a balance but not towards interest, a verdict distinguished only by colour
— now each need an explicit unlock before the calculator can be changed. That is
the intended behaviour. Changing a calculator whose arithmetic is proven should be
a deliberate act, and the lock makes the attempt visible instead of silent.

The next real decision is whether to create a **plan** for the site. That is what
wakes the planner, and it is the thing that would let the loop add pages and
reshape the site rather than only improve the pages that exist. It is also the
moment the first guard rail stops being dormant, so it wants its own look.

Still with the owner: the **GitHub token cannot see the repository holding the
site's source**, which needs admin.

And the neighbouring sites, loanandmortgagecalculator.co.uk and loancash.co.uk,
are where this one was a week ago, with **no site furniture at all**. If either is
decomposed before that is built, the platform's fallback links a stylesheet
neither of them serves and every page ships unstyled, silently. Both lanes have
been told, with the measurements.
