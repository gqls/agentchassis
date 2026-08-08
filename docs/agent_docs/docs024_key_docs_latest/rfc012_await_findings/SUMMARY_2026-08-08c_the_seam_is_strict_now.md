# SUMMARY — 2026-08-08c · the shared error-log seam stops guessing who wrote a row

*Third read-out of the day, and the last two were about delivering the owner's three rulings.
This one is about the loose end they left. New file, as always — the earlier summaries stand.*

## What we're trying to do

Make the platform's durable error record trustworthy enough to investigate from.

When something goes wrong inside a pipeline, an action writes a row to a table called
`agent_error_log`. That row is often the only thing left: certain steps pause and wait for an
outside service, and when they do, everything the action was holding in memory is thrown away. The
table is the one place a finding survives that pause. So the row has to be right — and "right"
includes saying **who** wrote it, because the first thing anybody does with an error is look up
the agent that produced it. The table is indexed on exactly that column.

## Where we've come from

Earlier this week we retired a mess: nineteen hand-copied copies of the same database INSERT,
which had drifted into five different shapes, with a third of them unable to be linked back to the
run that produced them. All nineteen now go through one shared writer. That was a real
improvement and it went live.

But the shared writer was helpful in a dangerous way. If the code calling it didn't say who the
row belonged to, it filled that in with whoever happened to be running at the time. Deliberate
omission was fine. **Forgetting was not, and forgetting looked exactly like succeeding** — no
error, no warning, and not one test in the package would have noticed, because they all check the
error code and the message and skip the author. Four independent reviewers spotted this in the
round that approved the work. All eighteen existing uses were correct; the nineteenth was the
risk. We wrote it down as the next job and handed it on.

## What we've done

Fixed it, at the source, and had it reviewed.

The shared writer's helpfulness is now split in two. The bookkeeping — which run, which machine,
which work item — is still filled in automatically, because there is no way to get those wrong.
The **authorship** is never filled in silently. Either the caller names it, or the caller asks for
"whoever is running" by calling a differently-named function, so that anyone reviewing that code
can see the choice was made deliberately. The old function that could do both no longer exists, so
this is a property of the structure rather than a warning in a comment — which matters on a tree
this many people work on at once.

If somebody does forget, the row is still written. That was a deliberate call and it was close:
there is a note elsewhere in our own code arguing that a row with the wrong name on it is worse
than no row, which reads like permission to refuse. But refusing would silently destroy a finding
in the one place findings survive, and a row that says **"unattributed"** isn't claiming the wrong
author — it's admitting it doesn't know. So the row lands, marked, with the running agent recorded
alongside as background rather than as a claim, plus a loud log line. One query finds every such
row. There are none.

Twelve tests, then fifteen. We proved they work by deliberately breaking the code six different
ways and checking that the right tests failed each time — and, just as important, that the wrong
ones didn't. That sounds excessive; it isn't. The original bug's defining feature was that the
whole suite passed while it was there, so "the tests pass" could not be the evidence this time.

It went to the review council and came back **approved on the first round**, ten reviewers, no
veto. Two of them raised the same fair objection: we'd been strict about the author but still
quietly borrowed the *step name*. We had argued that was out of scope. **They were right and the
argument was a false trade-off** — the two places affected didn't need the loose behaviour, they
needed to ask for it, and we'd just built the way to ask. So that's closed too, and no live record
changed as a result, which is itself pinned by a test rather than asserted.

## Where we are now

The seam is done and reviewed. It takes effect on the live system at the next fleet rebuild, and
the warning note about the old trap says exactly that rather than claiming victory — because on
every machine running right now, the old behaviour is still there.

One piece of honesty about that word "reviewed": the approval covers the main change. The
follow-up that closed the two reviewers' objection went back to the council on the same trail and
its verdict was still coming in as this was written, so the commit carrying it says "submitted",
not "reviewed" — a distinction the coverage report exists to catch people blurring, and one it
resolves by itself once the verdict lands.

Three things came out of the work that are worth more than the fix.

**The design proved itself while we were still arguing about it.** Another team added a
twenty-first place that writes to this table, and they did it *during* the review round, with no
knowledge of any of this. They reached for the obvious function and got the safe behaviour for
free. That is precisely the scenario the four reviewers were worried about, and it landed
correctly on its own.

**We found a bigger problem and left it alone on purpose.** When the code does ask for "whoever is
running", the answer is very often a useless placeholder — literally the word `generic`. All
twenty-five of one recorder's live rows say it; across the whole table it appears on 559 rows
spread over 25 unrelated steps, which is the fingerprint of a placeholder being passed around
rather than a real agent. And we already know: our own code carries a comment saying the value
we're reading "is often generic", pointing at a bug from July, with the correct answer sitting
right beside it — used by exactly one part of the system. Fixing that properly means moving a
piece of shared machinery, which is a second change of the same kind, and this same workstream had
a submission rejected two days ago for bundling two things into one. So it is measured, written
down in three places somebody will trip over, and left for its own review round.

**And I got caught out by our own shared tree, in a way worth reporting.** My commit message
claimed the warning note shipped alongside the code, as the rules require. It didn't: another
session had swept my edit to that shared file into *their* commit four minutes earlier, and mine
then picked up eleven lines of *their* unfinished work in exchange. Nothing was lost and the text
is in place — but the record now credits an unrelated change with writing it. Every signal I had
said fine: the file on disk was correct, the sync reported clean, the commit succeeded. It only
surfaced because a pre-commit check on my *next* commit flagged something else entirely. The
lesson is now logged fleet-wide: **"the file contains my edit" and "my commit contains my edit"
are separate facts, and almost every tool reports only the first.**

## Where we're going

One named next job, and it is the `generic` problem above: move the existing
resolve-the-real-agent-name logic somewhere both halves of the code can reach it, so error records
finally name the agent that produced them instead of a placeholder. It is measured, it has a
reviewed precedent to copy rather than reinvent, and it needs its own review round because it
changes what a column *means* for everyone already reading it — which may make it a question for
the architecture track rather than the ordinary gate. The handoff spells out the one trap in it.

Beyond that, this workstream's original three rulings are delivered and their follow-up list is
empty. The remaining open thread is not ours to close: a suspected fault in how deployed hero and
logo images reach a page, where five rounds of automated diagnosis and two humans' worth of code
reading have all hit the same wall — there is no retained evidence to examine. The next move there
is to watch one page build live. That is a twenty-minute job for whoever picks it up, and it will
settle in one run what reading the code cannot.
