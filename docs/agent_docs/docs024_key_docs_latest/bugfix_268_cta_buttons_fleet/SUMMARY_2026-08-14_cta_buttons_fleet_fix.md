# SUMMARY 2026-08-14 — the CTA buttons fix (bug 268)

## What we're trying to do

Across our live sites, call-to-action buttons — the "get in touch", "try the
tool" buttons that pages exist to serve — were silently disappearing. The
button's text was still stored, but the link it should point at was gone, and
our templates rightly refuse to draw a button that goes nowhere. Nothing
errored and five of our six quality checks stayed green throughout. This
workstream exists to stop the loss happening again, restore what was lost,
and say honestly how big the damage actually was.

## Where we've come from

The bug was found and measured on 2026-08-12 by the site-selling lane:
a routine content rewrite on webdesign.uk deleted the links from seven
components in one afternoon, while the one page not included in the rewrite
kept its links — a clean controlled experiment. That lane repaired
webdesign.uk by hand, locked its components as a tourniquet, counted 216
label-without-link buttons across 19 sites, and handed the fleet-wide fix to
a fresh thread — this one — with two already-refuted theories written down so
nobody would re-walk them.

## What we've done

We found the actual mechanism. Years-old plumbing has a shortcut for fields
that are "someone else's job to fill in": it skips them during page planning.
Months later, a rescue mechanism was built (for an earlier bug in this same
family) that preserves stored values through a rewrite — but fields taking
the old shortcut never reach the rescue. So every full rewrite of a page
threw away its button links while keeping the labels. The fix is one line in
the right place: the shortcut now tries the rescue first, and only falls back
to its old behaviour when there is nothing to rescue. Four new tests pin the
behaviour — including proof that the tests fail against the unfixed code —
and the change went through the review council, which approved it first time.
It shipped in this morning's release, and we verified at the running binaries
themselves that both servers carry it.

Two things went sideways and were handled. The independent diagnosis run we
always fire before committing to a root cause broke on its own plumbing (its
final write-up outgrew a size limit), so we did its checking by hand against
the database's change history and wrote down exactly what we checked. And
while our edits sat uncommitted, three of them were swept into other
sessions' commits — the known hazard of this shared tree; nothing was lost,
and each case is recorded.

The most important discovery came from that by-hand checking: **the damage
count was two different problems wearing one number.** Of the 217 buttons
missing fleet-wide, only about ten ever had a link that a rewrite deleted —
those we can restore from history. The other two hundred or so never had a
destination at all: the part of the system that picks where buttons point
never found an answer for them, said so at the time in its own queue, and
the button was born label-only. We had written "expect the count to fall to
zero after repair" into our own plan; that was wrong, it is corrected
visibly, and the lesson is logged in the fleet's wrong-calls ledger.

## Where we are now

The leak is plugged in production: as of today's release, a page rewrite
cannot delete a button's destination. Ten genuinely-deleted links across five
sites are identified, dated, and recoverable from history. webdesign.uk
remains repaired and locked. The council's approval came with two sensible
advisory asks — check the neighbouring code paths the same way, and add one
more test — which are queued, not yet done. Nothing has yet been proven on a
live rewrite since the release, and the ten rows are not yet restored.

## Where we're going

In order: do the council's two follow-ups; run one controlled rewrite on a
real page and watch the links survive (the proof the fix works in anger);
restore the ten deleted links and rewrite one of those pages again to prove
fix-plus-repair holds; then unlock webdesign.uk as the final step. The
two-hundred-odd buttons that never had destinations are a separate decision
that belongs to the owner: resolve destinations for them site by site, accept
them as label-only, or open a new lane for it — the handoff lays out the
options. A fresh session picks all of this up from
HANDOFF_2026-08-14_canary_and_repair.md.
