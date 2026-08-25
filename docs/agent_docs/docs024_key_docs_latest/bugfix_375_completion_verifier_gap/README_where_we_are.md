# Where we are — the unguarded completion writer (`bugs_open/375`)

Append-only, newest at the bottom. Plain prose, for the owner.

---

**2026-08-24, evening.** Picked this bug up fresh — nobody had worked it, though the ownership
tool said otherwise (it was reading mentions from a lane that has since closed). Claimed it first
so the next person gets a straight answer.

Here is the bug in ordinary terms. When one of our agents finishes trying to fix something on a
site, a row somewhere gets stamped "done". There are two bits of code that can do that stamping.
One of them, before it stamps, re-checks that the problem is actually gone — that's the whole
point of it, because an agent can report success without having changed anything. The other one
just stamps. Which check a job gets depends entirely on which of the two bits of code the job's
configuration happened to name, and nothing anywhere tells you that.

I re-ran the measurement the previous session had left. It holds: four of our two hundred agents
use the unchecked stamp, across six places, covering five kinds of problem — and **none of those
five kinds has a re-check written for it yet**. So nothing is being skipped today. The damage is
zero.

That sounds like a reason to close the bug. It isn't, and the reason why is the interesting part.
We keep a maintained list of problem types that *ought* to have a re-check written, and it says so
in its own words — "this is the actionable backlog, not an excuse list". Two of these five are on
that list. So the situation is: somebody will eventually sit down to write one of these re-checks,
they will register it, our test suite will go green, and it will protect nothing at all, because
the code that stamps those particular jobs never asks.

**Then I found the part that makes it worse than the handover said.** We keep a register of how
each mechanism works, and its entry for the relevant router carries an explicit warning to whoever
writes that re-check: *"registering this will cause one of the router's closing paths to refuse to
complete — read the close paths first."* That warning is wrong. It's wrong *because of this bug* —
that closing path uses the unchecked stamp, so registering the re-check wouldn't cause a refusal,
it would cause nothing whatsoever. So the person we're warning is being told to brace for one
wrong outcome and will walk into a different one, quietly.

That single fact decided the fix for me. The tempting fix is to make the unchecked stamp start
checking, always. But that would make the register's warning come true — it would break a live
route, as a side effect of switching on a guard nobody had asked for. So instead the checking is
**opt-in, one closing path at a time**: the code gains the ability, and each place that stamps has
to say "yes, check me". Nothing changes anywhere until somebody arms it deliberately, which is
also exactly where the decision is visible to whoever reviews it. That's the shape you ruled for
this kind of change on 2 August.

The obvious objection to opt-in is that a safety switch nobody turns on is worthless — and you've
been bitten by mechanisms rotting unexercised before. So there's a second half: when the unchecked
stamp runs on a job whose type *does* have a re-check registered, it will complete the job exactly
as it does now, but **write down on the row that it skipped a check that existed**. No behaviour
change, no risk, but the moment the trap is sprung it announces itself somewhere we can query,
instead of being invisible.

And I'm correcting both documents that currently mislead — the test file's header, which reads as
though registering a re-check protects a type, and the register entry with the wrong warning.

What I am deliberately *not* doing: merging the two stamping code paths into one. That's the real
structural fix, it's a bigger change, and it's the kind that goes to architecture review on its own
merits rather than riding inside a bug patch. This change should make it easier — after it, both
paths share one implementation of the check itself.

Next: write the code, prove the guard is load-bearing by deliberately breaking it and requiring
the test to fail, and put it through the reviewer council.

---

**2026-08-24, later.** Done and committed. Two commits: the code, and the documents that were
telling people the wrong thing. Both carry the council submission id, so they get credited
automatically when the verdict lands — no rewriting history, which isn't allowed here anyway.

Two things worth your time out of the day.

**The first is that the warning was worse than the silence.** I said earlier that the trap here
is that somebody will write one of these re-checks and it will protect nothing. What I found
afterwards is that our own register *warns* that person — and warns them wrong. It tells them
to expect a specific side effect that cannot happen, precisely because of the bug. So they'd
plan around a fail-close, get a silent no-op, and have no reason to look. A wrong warning is
more expensive than none, because it stops you asking.

**The second is a bit of an embarrassment that turned into the most useful finding of the day.**
I got a test fixture wrong — registered a fake problem type into a shared registry — and a guard
I'd never heard of failed the build. My first instinct was to make the guard stop complaining by
adding my fake type to a real production list. I didn't, but I thought about it for longer than
I'd like, and I've written that up honestly in the fleet-wide log of wrong calls, because "a
guard is objecting to me, therefore the guard needs adjusting" is a reliable way to break
something.

What the guard turned out to be protecting is the interesting part. There is a **third** thing
that stamps rows done — a timeout sweep — and it bypasses *both* checks. Somebody solved that,
months ago, with a written-down list plus a test that fails the build if the list and reality
disagree. Same problem as ours, already solved once, and nobody had joined the two up. That's
now written down as the exact shape for the piece of work still outstanding here.

So: the fix is in, the documents are honest, the mechanism is registered so the next lane can
find it, and the remaining work has a known-good pattern to copy rather than a design question.

---

**2026-08-24, end of the evening.** The council approved it first time round — twelve reviewers,
about fourteen minutes end to end. But the useful part is that "approved" did not mean "nothing to
do", and I want to be straight with you about one of the two things it found, because it was mine.

**They caught a measurement I had got wrong and already published in six places.** The whole
argument for doing this the contained way rests on a count: how many agents use the unguarded door,
over how many kinds of problem. I measured it, controlled it, wrote it up carefully, marked it as
measured, and put it in the bug file, the register, the index, a landmine, the review submission
and two source files.

The table I measured it from is a **rolling window**. Completed rows get moved out to an archive
after a while, and the archive is exactly where the completions I was counting had gone. Over both
tables together it is **seven** kinds of problem and **578** completions, not five and 134. Two
entire categories were invisible to every query I ran, because every one of them had finished and
been archived.

Three things about that are worth more than the correction itself.

**The control I was pleased with had the same blind spot.** I'd deliberately run a check to prove
my zero wasn't a typo — and it queried the same table, so it tested my spelling and not my window.
A check drawn from the same place as the measurement cannot see a problem with the place.

**The conclusion survived, and that is luck rather than skill.** All seven categories are still
unprotected, so the answer didn't change and the fix didn't change. But if either of the two hidden
ones had been protected, my query would have printed the same reassuring zero, and a claim I made
to a review board would have been resting on it.

**And the warning was already in front of me.** The note that says this exact thing about this
exact table loads automatically at the start of every session I run. I had it and didn't apply it.
That is now logged in the fleet-wide record of wrong calls, along with the observation that the
handover note I was working from made the identical mistake about the same table on the same day —
which suggests the fix is the query, not more care.

Everything is corrected at source, including the two code files, because a stale "measured" comment
in code outlives every document. The second objection — about a shared helper I'd refactored — I
answered by deliberately breaking it, which proved the existing tests do catch drift, and in the
process exposed a gap where a column could have been silently dropped. That's closed too, though it
took two attempts: my first test looked right and proved nothing.

Fix committed, documents honest, figures corrected, verdict read and acted on. The remaining piece
of work has a known-good pattern to copy rather than a design question.

---

**2026-08-25, morning.** The new chassis build went out overnight, so I checked whether the change
actually reached the running system rather than assuming the tag meant anything. It has: both pods
are on the new build and both carry the new code, confirmed by looking inside the running binary for
two things that should be there and two that shouldn't. One of those checks is worth mentioning
because it is a trap I'd written down in advance and it behaved exactly as predicted — if you look
for the *name we use in the source code* rather than the *text that actually ends up in the
binary*, you get "not shipped" back while the feature works perfectly.

**The honest position on whether it is doing anything: it isn't, and that's the design.** Nothing
has switched it on, and none of the problem types it can reach has a re-check written yet. There
are zero recorded bypasses — and I've been careful to write down, everywhere that number appears,
that the zero means nothing on its own. Nothing *can* produce that record until somebody writes one
of those re-checks. It is not a pass and it is not a fail, and a future reader will meet the zero
before they meet the explanation, so the explanation now sits next to it in three places.

**I've also gone back and fixed our own status lines,** which had quietly become the most dangerous
thing in the lane. Three documents said "inert until the next roll". That was true when written and
misleading by this morning — the kind of sentence that makes the correct next step look premature
to whoever reads it. They now say the gate is live, while keeping the separate and still-true point
that it deliberately does nothing yet. Those are two different claims and running them together is
how "live" starts to mean "working".

**Now the question you asked: can we close it?** No, and I'd push back on closing it. Our bar is
*fixed and live*. It's live. It isn't fixed. What the bug actually says is that one of the three
things that stamp work "done" has no safety check at all — and after this change it still doesn't,
unless somebody switches it on per closing path, which nobody has. What shipped is the mechanism
and a tripwire, plus honest documentation where there was misleading documentation. That's real
progress and it is not the fix. Closing it now would be the same mistake the bug is about: treating
"a mechanism exists" as "the problem is gone".

**What's left is three things, and only one of them is a real decision for you** — I've set them out
in the handoff and in my reply.

Everything is committed, the review verdict is in and acted on, and the next session has a
cold-start document that tells it what is proven, what is merely true today, and what would let it
close this properly.
