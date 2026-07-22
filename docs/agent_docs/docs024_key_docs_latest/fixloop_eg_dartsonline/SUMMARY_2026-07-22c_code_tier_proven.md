# SUMMARY — the diagnosis code tier, proven on a real bug (2026-07-22c)

*A milestone read-out, written to be read aloud. Current state only — the
chronology is in NOTES and README_so_far.*

## What we're trying to do

Build a diagnosis loop that can be trusted to find the real cause of a bug on its
own — cited, consistent, and honest enough to say "I don't know" rather than
guess. Part of that loop is the **code tier**: the ability, mid-diagnosis, to stop
and ask "let me look at the actual source code" when a hypothesis turns on what the
code does, not just on what the live database shows. Without it, the loop can only
reason from runtime signals; a whole class of bug (does this mechanism exist? is it
handled at every call site?) is invisible to it.

## Where we've come from

The code tier was designed and built weeks ago and wired end-to-end — the plumbing
was verified live in the running binary — but it had **never once fired on a real
diagnosis**. It was the single biggest thing the last handoff listed as still owed.
Everything else in the loop (cited read-only diagnosis, the reviewer council, the
build gate, the PR path) had been exercised; this had not.

## What we've done

We proved it, and fixed a blocker on the way. The blocker: the loop reads code from
a pre-built search index, and that index was three weeks stale — nothing refreshes
it automatically, and the documented manual refresh was itself out of date and
failed. We found the correct path, refreshed the index to current code, and filed
the staleness as a bug in its own right, because the same stale index also feeds
the review council's "does this already exist?" seat — so for three weeks, anything
built recently looked nonexistent to both.

Then we ran a real, code-shaped diagnosis. The loop stopped, fetched the actual
source, and **corrected the person who asked it**: the functions we had pointed at
all already reported their dropped items properly, so instead of confirming our
guess it surveyed the rest of the package and found the *real* offender we hadn't
named — a function that silently caps a list and leaves only a vague text note
instead of a proper count. It then checked the live data to confirm that cap has
actually been hit in real runs, not merely that it could be. It reached a firm,
fully-sourced conclusion in four passes.

## Where we are now

The code tier is proven end-to-end on a real case: it emits a code question, that
question is answered from the (now-fresh) index in the next pass, and the answer
drives a cited, CONFIRMED diagnosis that overruled a wrong premise and found a new
defect. The loop's core promise — it reads the code and follows the evidence to the
real cause, even against the asker — held. The defect it surfaced is minor and
left unfixed pending a decision; the machinery is the deliverable.

## Where we're going

Three things are open, all owner-gated. The index staleness needs a lasting fix —
a refresh cadence (best tied to image roll) or a freshness guard so a stale answer
reads as "unknown", not "absent"; that is filed and waiting. The route↔load_runtime
wiring guard shipped this cycle but is inert until the next image roll and could
still go through the reviewer council if we want the extra eyes. And the minor
capping defect the loop just found is a candidate fix if we judge it worth the
change. None of these is blocking; the loop itself now stands on all its tiers.
