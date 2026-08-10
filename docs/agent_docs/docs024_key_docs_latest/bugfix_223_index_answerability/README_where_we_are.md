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

## 2026-08-10, afternoon — built, committed, and one uncomfortable moment

The fix is in and the config half is already live. Here is what happened, including the bit
I got wrong.

**The design.** I had a design pass run by a second model before writing anything, and it
agreed with the shape I had gone in with, which was reassuring rather than surprising. It
also caught a number of mine: I had sized the follow-on work using a quick text search that
counted 930 declarations, and the real figure is 1,173, because a grouped declaration block
counts as one line however many things it declares. Worth recording because it is the same
class of error as the bug itself — a search tells you what it can see, not what is there.

**What we actually changed.** Three things, in order of how much they matter.

The smallest and most important: every verdict the robot writes now carries a line, composed
by the machinery rather than by the robot, saying how many of its checks could not be
answered by the index at all. That matters because the verdict is read months later by
people and by other machines, long after the run that produced it is gone. A caveat the robot
is merely *asked* to include arrives most of the time — and "most of the time" is exactly the
problem we are fixing.

The second: the robot can no longer *reach* the "this note is stale" verdict when nothing in
its round was actually confirmed against real code. Not discouraged from it — unable to get
there. The route through the workflow now forks on a plain true/false fact, and the branch
for "we confirmed nothing" offers only two answers, neither of which is "stale". I put this
second because a determined model could still type the word into free text; the line above is
what protects the reader either way.

The third, and the one that helps most widely: the lookup itself now says what it cannot
see. It reads its own contents — which file types, which kinds of declaration — and when you
ask it about something it structurally cannot hold, it says so instead of saying "we ran your
query and found nothing". Crucially, all of that wording is *computed* from what the index
actually contains, so when we widen the index later the warnings retire themselves. Hardcoding
"we only hold Go" would have been quicker and would have quietly become a lie.

That third change benefits four different consumers, not just the landmine robot — including
the diagnosis loop, whose own source comments had already asked for exactly this and never
got it.

**Two things we found that the original bug report did not know.**

The blind spot does not only make things look absent — it can make them look *present*. If a
note points at a folder, and that folder happens to contain some Go files in its
subdirectories, the check comes back with a generous list and reads as confirmation, while
every file the note actually named is invisible. That is worse than a wrong accusation: a
wrong accusation makes you go and look. A flattering half-answer looks like diligence.

And the careful runs are not free either. I fired the robot at a fresh note on purpose this
morning, before changing anything, to bank the "before" picture. It drew the careful branch
— and still spent two model calls and eight queries to arrive at "either these files don't
exist or they aren't indexed, I can't tell", which one census query answers outright. It then
guessed the answer correctly and hung its whole verdict on the guess.

**The uncomfortable moment.** I wrote twelve tests, watched them all pass, and then did what
this project's rules require: broke each guard deliberately to check a test would notice.
Six of seven noticed. The seventh did not — I deleted the folder-listing fix entirely, at the
point where it is used, and every test stayed green, because they all tested the *function*
and none tested that anything *calls* it. And that was the guard for the false-confirmation
problem I had discovered myself an hour earlier. Finding a failure mode does not protect you
from it. Three more tests now go through the real path and the deliberate break is caught.
It is written up in the shared log of wrong calls, because the useful part is the question:
which test fails if I delete the *call*, not the function?

**Where we are.** Committed. The database half is applied and I proved it is safe to have
applied it early rather than assuming: I fired the robot again afterwards, on the unchanged
program, and watched the new fork evaluate to "no" and take the old route, complete
normally, and produce the same verdict as before. It could have failed three ways and did
not.

The code itself does nothing until the next build of the service goes out — that is normal
here, and I have left a marker so anyone can check in one command whether the running program
contains this change or not. It is currently absent, which is the useful thing about having
banked that check before making it.

**What is still owed.** The review council's verdict, which is running now and which I will
read and act on. Then, after the next build: re-run the robot on the two notes that started
this and compare against the banked "before" — including that the checks which *did* work
must still work, because a fix that buys quiet by checking less would look like success.

And then the larger piece, deliberately kept separate: actually teaching the index about the
~1,170 declarations it currently cannot see. That widens what every diagnosis in the system
searches, so it deserves its own review rather than riding in on this one's coat-tails.
