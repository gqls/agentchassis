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
