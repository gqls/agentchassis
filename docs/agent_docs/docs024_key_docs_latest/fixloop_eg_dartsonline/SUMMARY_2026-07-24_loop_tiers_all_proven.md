# SUMMARY — the diagnosis loop, with all its tiers now proven (2026-07-24)

*A milestone read-out, written to be read aloud. Current state only — the
chronology is in NOTES and README_so_far. A NEW file, not an edit of
SUMMARY_2026-07-22c (which was narrowly the code-tier proof); this is the wider
read-out of where the whole loop stands, and the state has moved since: the
wiring guard is now live, and the stale-index gap is already re-demonstrated.*

## What we're trying to do

Build a self-healing system that finds and fixes its own bugs, and is trustworthy
because it **refuses** — it will not confirm a cause without evidence, will not
bless a fix that only covers part of a problem, and will say "I don't know" rather
than guess. The diagnosis half reads the real code and the live database, forms a
cited theory, and follows the evidence to the actual cause — which often lives
somewhere the symptom never points. The fix half turns a confirmed diagnosis into
a constrained edit, argues it before a reviewer council, and opens a pull request
for a human.

## Where we've come from

Over prior sessions every part of this was built and, one by one, exercised on
real cases — the cited read-only diagnosis, the reviewer councils, the build gate,
the first automatically-opened pull request. Two things were still outstanding.
First, the **code tier** — the loop's ability to stop mid-diagnosis and fetch the
actual source when a question turns on what the code does — had never once fired on
a real bug; it was the last unproven tier. Second, a **recurring class of mistake**
kept re-opening bugs we thought were fixed: a guard written against the members of
a set that existed at the time, silently missing members added later (a new council
seat without a setting its siblings had; a fix applied to one of two twins).

## What we've done

The code tier is proven. Given a genuinely code-shaped question, the loop stopped,
fetched the real source, **corrected the premise it was handed** — the functions we
pointed it at all turned out to be already fixed — then searched the rest of the
code and found the actual offender we had not named, and checked the live data to
confirm the fault has really occurred, not merely that it could. A firm, fully-
sourced conclusion in four passes. On the way we found and fixed a blocker: the
code search index the loop reads had gone three weeks stale, which also quietly
degraded a reviewer seat; we refreshed it and filed the underlying gap.

And we hardened that recurring class three ways: a small checker that reads the
live councils and flags any seat that has drifted from its siblings; a guard, now
shipped and live, that makes two loosely-coupled diagnosis steps fail loudly if
their wiring is ever mismatched rather than going silently wrong; and we brought a
lagging council up onto the current model, which had been left behind in an earlier
upgrade.

## Where we are now

Every tier of the diagnosis loop has now run end-to-end on a real case — the code
tier was the last, and it did more than pass, it out-reasoned the person who
prompted it. The wiring guard went live in yesterday's image roll. The councils are
consistent and on the current model. The one real soft spot is the code search
index: it has no automatic refresh, so it is already drifting behind the freshly
deployed code again — exactly the gap we filed. Nothing is broken; the loop stands
on all its tiers.

## Where we're going

Four things are open, all owner-gated and none blocking. The index needs a lasting
fix — a refresh on every image roll, or a freshness check so a stale answer reads
as "unknown" rather than "absent" — because that same index also feeds the
council's existence checks. The wiring guard is live but has not yet been watched
catching a real induced fault, and it can still go through the reviewer council for
a second opinion if we want it. The minor capping defect the loop found is a
candidate fix if we judge it worth the change. And, most importantly, the code tier
being proven means we can now point the loop at the harder, genuinely code-shaped
bugs it was built for.
