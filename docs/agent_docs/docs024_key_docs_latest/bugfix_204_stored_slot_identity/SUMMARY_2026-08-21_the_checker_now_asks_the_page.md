# The checker now asks the page — bug 204's remaining half, fixed

**2026-08-21.** Written to be read aloud.

## What we're trying to do

Stop the site planner from deleting the record of what a page is made of.

When we take over someone's existing site and break their hand-written pages into
pieces, each piece gets a **positional name** — "the first block of prose on this
page", "the second calculator" — rather than a name describing what kind of thing it
is. That is deliberate: a page with three prose blocks needs three distinguishable
names, and calling them all "prose" would collide. The real identity of each piece is
a link stored alongside it in the database.

Several parts of the platform take a section name and try to work out which component
it means. All of them did it by looking the name up in a catalogue of components. A
positional name is not in that catalogue and never will be.

## Where we've come from

This was found in August and fixed twice — in the page-rebuild path and in the
re-render path — both times by teaching the code to consult the stored link first. A
third place was never fixed, and it behaves worse than the other two: instead of
postponing a section it cannot identify, it **deletes** it from the plan. The plan is
then written to the database, and the page's section list is overwritten with the
shortened version.

The page keeps serving perfectly well, because the actual content is untouched. What
is destroyed is the *record* of what the page is made of — so the next rebuild has
nothing to rebuild, and would build a blank page over a live one.

On the 20th of August another session fired a routine replan at one site to prove an
unrelated fix, and **41 of that site's 45 live pages had their section lists
emptied**. It was caught within the hour and restored from a snapshot. The same run
also queued twenty jobs to "rebuild" those now-empty pages.

## What we've done

Found that the blind lookup is used in **four** places, not the one the bug report
named — and that two of them write straight to live pages.

Established the scale from the platform's own records rather than by argument. A
permanent log of every deleted section started on the 17th of August. Since then it
has recorded **140 deletions across 41 pages, and all 140 are positional names**. The
mistakes that checker exists to catch — typos, renamed components — account for
**none** of them.

Fixed three of the four places. The fourth is deliberately left alone, with a test to
keep it that way: a genuinely new page has no stored sections, so a positional name
proposed for it really does point at nothing.

The fix had to work around a written warning of our own: *do not fix this by widening
the checker's list of known components*, because three of the four places that use
that list belong to a path where widening it was an explicit decision not taken. So
the list is untouched. Instead the question is asked per page — *does this page
already have a section called this?* — which can only ever preserve what a page
already has, and can never let the planner place something new. A test fails if a
section stored on a different page were ever enough.

We also added a second, independent protection: the step that saves a plan can no
longer write an empty section list over a real one. Its two neighbouring columns in
the very same database statement were given exactly that protection two days earlier;
this one had been left out.

Both changes went through the review council. The first round was approved and still
found a genuine defect — our own failure-handling left no permanent trace, which is
the same disease we were curing. Fixed.

## Where we are now

Everything is committed. **None of it is doing anything yet** — this kind of change
only takes effect when a new image is built and released. The second review round is
still running.

Two things are honestly imperfect and are written down rather than smoothed over.
Some of the tests establish that the database is asked the right question but cannot
establish that it answers correctly, because the test framework does not run real
SQL — that needs a live check after release, and the verification list says so. And a
page whose stored name differs from its canonical name may not match the "release this
page for redesign" instruction; the failure direction is the safe one and it is
recorded, but it is real.

Two mistakes this session are logged in the shared record of wrong calls. Once, a
change of mine sat half-finished in the shared working tree and another session's
commit picked up half of it — the shared codebase would not have compiled for 33
seconds. And a figure in a delegated plan turned out to be invented; I caught it by
re-deriving it, and the lesson is that the correct figures either side of it were what
made it believable.

## Where we're going

After the next release: prove the change is actually in the running service, then run
the canary on the site that was damaged — taking the snapshot first and cancelling the
job queue before any repair, because the queue is what turns a database problem into a
deployed one.

Then the check that matters most: the count of deleted sections should go to zero, but
**a zero must not be trusted until the detector is proven still to fire.** We will
induce one deliberate deletion and confirm it still appears. A blind detector and a
working fix produce identical numbers.
