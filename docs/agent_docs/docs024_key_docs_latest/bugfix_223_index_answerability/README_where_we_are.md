# Where we are — bug 223, the code index's blind spots being read as "it doesn't exist"

Plain prose, append-only, newest at the bottom. This is the running log.

---

## 2026-08-10, morning — picking the bug up, and what it is really about

We have a small robot whose job is to check the landmine notes we keep for each other.
When someone appends a note ("touch this file and here is the trap"), the robot looks up
the files and symbols the note names, reads them, and records a verdict: still true, gone
stale, or needs a human. The verdicts go into the database where the review council and
the next session can read them.

The robot looks things up in a code index — a table of every symbol in the codebase. The
trouble is that the index only contains **Go**. It has 5,837 entries this morning and
every single one is a `.go` file. But most landmine notes are not about Go: they are about
scripts, SQL migrations, database tables, config values, commands. So when the robot looks
those up, it gets nothing back — and it has been writing that nothing down as "this does
not exist", sometimes in so many words: *"the entire described workflow has no footprint."*

The best example of how wrong that is: one of those verdicts declared that three of our
own scripts, and a whole category of database rows, did not exist anywhere. It was
delivered **by those scripts**, and stored **in that category**. It disproved itself on
arrival.

The reason this matters more than one bad verdict is what a verdict is *for*. A "stale"
verdict is the signal a future session uses to delete a note. So a false stale does not
just fail to check something — it actively argues for throwing away a warning that was
correct.

## What we found today that changes the shape of the fix

Two things, both from looking rather than assuming.

**First**, the person who filed the bug thought the problem was "non-Go footprints". It is
wider than that. The index does not merely lack other languages — it also lacks two
categories of *Go* declaration: package-level `var`s and `const`s. There are around 930 of
those in the codebase and none is indexed. So when a note points at one, the robot says
things like *"no longer resolves as a standalone symbol (possibly inlined or renamed)"* —
which is worse than saying nothing, because it invents a plausible story a human will then
waste an afternoon chasing. Nothing was renamed. The index simply cannot hold that kind of
thing. The same gap stops our diagnosis loop cold: it can find every *use* of one of those
declarations and never the declaration itself, so it gives up with "unverifiable" while
naming exactly the thing it cannot see.

**Second**, and this is the useful discovery: the database table was *built* expecting
those two categories. Its own constraint permits them. The half of the code that reads the
index already treats them as ordinary code. Only the half that writes it never learned to
collect them. So this is not a design gap, it is an unfinished job.

## What is actually broken, in one sentence

The sentence the lookup prints when it finds nothing. It currently says, in effect, *"we
ran your query and it matched nothing; this is a real answer, not silence."* That wording
was itself a fix, from an earlier bug where empty answers were read as silence — and it
worked. But now it overshoots: for a thing the index cannot represent, "we ran it and
found nothing" is true and deeply misleading. A guard that is too confident is still a
defect.

And the deeper point, which the bug's own author corrected himself into: the blindness is
perfectly consistent, but the *conclusion drawn from it* varies run to run. Four verdicts
on identical empty input ranged from a careful "cannot be mechanically verified" to a flat
assertion of non-existence. You cannot tell which one you are holding. So asking the robot
more nicely to be careful is not a fix — three times in four it already is careful. The
one flat wrong answer is what does the damage, and only something structural removes it.

## Where we are right now

Bug confirmed still live, re-measured this morning rather than taken on trust. Ownership
checked two ways: no lane owns it, and no live session is working it. The mechanism is
read end to end and the four places that consume the same lookup are identified — this is
shared machinery, so a change here is seen by the review council's own seats, not just by
the landmine robot.

A design pass is running now to rank the candidate fixes. My own preference, going in, is
to fix it in the shared machinery rather than in the one agent: teach the lookup to know
what it cannot see, say so in words that cannot be read as absence, and hand the consumer
a machine-readable fact ("nothing you asked about was checkable") so the decision stops
being a matter of how the model felt that run. Then the separate, larger question of
actually indexing the missing Go declarations gets its own round, because it widens what
every diagnosis run searches and that deserves to be reviewed on its own merits.

Next entry will record what the design pass recommended and what the council said.
