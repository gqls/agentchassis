# Where we are — bug 284, the findings that were marked as failures

Plain prose, append-only, newest at the bottom.

---

**2026-08-16, afternoon.**

Picked up bug 284 off the open pile. It had been filed yesterday by the session
working bug 279, which finished and said in as many words that 284 was unowned and
needed a fresh pair of hands. Checked that no one else was on it — the ownership
script said "owned", but the only thing it had to go on was the commit that filed
the bug, which is a known blind spot, so I read the live session logs instead.
Nobody was on it.

**What the bug is.** The platform finds problems on our sites and writes each one
down as a "work item". Most of those are jobs: something is wrong, and we have an
agent that can fix it, so the item names that agent and the machinery sends it
along. But some findings are not jobs at all. Nobody can automatically repaint a
client's brand colours, restart a customer's virtual machine, or decide which page
a duplicated paragraph really belongs on. For those, the checker deliberately
leaves the "who fixes this" field empty and the item just sits there, visible, for
a person to read.

The trouble is that the step which moves new findings into the work queue never
looked at that field. It swept up everything on a site, including the ones with
nobody to send them to. The queue then picked them up, discovered there was no one
to hand them to, and stamped them **blocked — "cannot be routed to any agent"**.

So a perfectly correct observation ends up filed as a machinery failure. And it is
worse than untidy in two ways. Nothing ever unblocks those rows, because the
recovery job only rescues items whose named agent has since appeared, and these
never named one. And a blocked item still counts as "open" for de-duplication
purposes, so the checker that found the problem in the first place is not allowed
to report it again. The finding is frozen in the wrong state and the fresh evidence
is silently dropped.

**How big.** The bug file said 18 rows on 14 sites, all of one kind. When I asked
the question by the error message rather than by the kind of item, it came back
**60 rows across four kinds on at least fifteen sites** — the biggest group being
broken image references, at 40. Another 37 are sitting in the queue today waiting
to be swept up the same way. The reason the original count was low is worth
recording: the search that found the producers looked for a line of code that sets
the field to empty, and the worst offender does not set the field at all — the
programming language fills it in as empty by itself. Invisible to that search.

**What I have done.** The step that promotes findings now asks the same question
the dispatcher will ask a moment later: does this item name someone, and does that
someone exist? If not, it leaves the item alone. Both places now get that question
from one shared piece of code rather than each spelling it out, so they cannot
drift apart later — which is the failure this codebase keeps paying for. It also
now says out loud how many items it held back and of what kind, because a filter
that quietly does less looks exactly like a quiet week.

I also ran it past the diagnosis loop before asserting the cause, as the rules
require, and I am glad I did. It did not reach a verdict, but it found a mistake in
my evidence: I had said a certain marker on the rows could only have been written
by the one step I was accusing, and it turned out three different places write that
marker. The accusation still holds — but on a sharper test, because the other two
always write a fixed value and these rows carry the value only the accused can
produce. That check also turned up something I had missed entirely: two of the 60
rows were not created by the machinery at all. They were inserted by hand, by other
sessions, already in the queue and with no agent named. My fix does not stop that,
and I have said so rather than quietly rounding 58 up to 60.

**Where it stands.** The code is committed and has gone to the review council. It
does nothing until the next chassis release — Go changes only take effect when a
new image is built and rolled. I have deliberately **not** repaired the 60 damaged
rows yet, because until the fix is actually running they would simply be blocked
again within the hour.

**What is left, and one of it is a judgement call.** After the release: repair the
60 rows, then add a database-level rule that makes the bad combination impossible
to write at all — that is the only thing that catches the hand-inserted case,
because roughly twenty places in the code write these rows directly and bypass the
shared front door. That rule has to go in **after** the new code is running, not
before: database changes take effect instantly while code changes wait for a
release, so adding it first would make the old code fail on every site that has one
of these findings, and that would stop the improvement loop across the whole fleet.

One small note in passing: you asked for the plan to be prepared with Fable, and
Fable came back with "you've reached your Fable 5 limit" and produced nothing. I
did the planning myself rather than stall, and said so in the notes so nobody later
mistakes it for Fable's work.
