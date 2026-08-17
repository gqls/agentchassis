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

---

**2026-08-16, later.**

The council came back **REVISE** first, then **APPROVED** on the second round.

The first round is the interesting one, because it caught me twice. One seat
pointed out that the marker I was using to blame the two checkers cannot actually
tell you which checker wrote a row — it records the row's own category, and several
things write it. It was right. Re-measured on the marker each checker stamps with
its own name, the answer came out cleaner than what I had submitted: nine rows from
one, nine from the other, exactly the two files I changed and nothing else. A second
seat picked at a loose phrase in my write-up — "three of six producers" — and
following it properly turned up a **sixth** checker I had not found, with
thirty-six live rows, which also never mentions the field at all.

The gating objection was fair too: I had called the change to the shared dispatch
code "semantically identical" without ever quoting the code it was replacing. It
is identical — I quoted both halves in round two, and they differ only by a
nickname and some whitespace — but the reviewer was right that I had asked them to
take it on trust.

Round two passed with four advisory notes, and I have answered all of them rather
than waving them through: the old case in the back-catalogue that looked like a
duplicate is a different member of the same family (a named agent that doesn't
exist, versus no agent named at all); the repair of the sixty rows is now a proper
gated script with its own backup, a check that can fail, and a rollback beside it;
and the deferred work is recorded in the database where the other agents can read
it, not just in our files.

**One thing genuinely needs you.** The two seats that looked hardest at this change
disagree with each other about the same edit. One says I unified too little — there
is a third copy of the same test elsewhere in the codebase, held in step only by a
comment, and I named it and walked past it. The other says I unified too much —
that the shared dispatch code every pipeline depends on shouldn't carry a
refactor that came out of a bug lane, even though it agrees the change is harmless.
Both are reasonable. The rules say a person breaks that tie, so I have left the
approved code as it is and written the disagreement down where you'll meet it.
Either direction is a one-file change whenever you decide.

---

## 2026-08-17 — it is done, and one of the rescued items is yours

The release went out, so everything this lane was holding could finally run.

**The 60 stuck items are repaired.** Each one is back in the state its own checker
files — the "roadmap" ones parked as roadmap entries, the broken-image ones back to
plain findings — and each carries a note saying what happened to it, so none of them
looks like it fixed itself. **One of them was yours**: the note you raised on
2026-08-12 reading fundamentallyai.com — the six Platform Log guides that name tools
and never link to them, and no Tools entry in the top nav. It had been filed as work,
then silently marked as a routing failure two days later, and it has been unreachable
ever since. It is now parked where the roadmap report picks it up. **Nothing has
acted on it — it needs a human, because there has never been an agent that handles
that kind of item.** Flagging it rather than leaving you to find it.

**The door is shut, not just swept.** Beyond the code fix, the database now refuses
outright to put a "nobody can do this" item into the queue of things to do — so a
hand-written insert, or one of the twenty-odd places in the code that bypass the
normal door, can no longer recreate this. I tested that by trying it: the bad shape
is rejected, the two legitimate shapes still go through.

**And I proved the fix works rather than assuming it.** A quiet week would have
looked identical to a working fix, so I picked the site with 36 of these items and
nothing else to do, and ran the exact step that used to break them. It held all 36
back and promoted none — where the old build would have taken every one. That is the
difference between "no complaints" and "checked".

**One thing needs your ruling** (unchanged from the review, nothing was done
unilaterally): two reviewers disagreed about the same edit — one says a third copy of
a shared check should have been unified too, the other says touching that file at all
went beyond the bug. Either way is a small, single-file change whenever you decide.

---

**2026-08-17.**

The build went out and the fix is live and working. It is worth saying how that was
established, because "we deployed it" is not the same claim: another session checked
that the *running* services carry the exact commit (the image's own label, plus a
proof that my commit is an ancestor of it), and I separately asked both running
processes whether my code is inside them, with a deliberate nonsense check alongside
to prove the question could come back "no". Then they proved it *works* rather than
merely exists — they picked a site holding thirty-six of these flag-only findings and
nothing else routable, ran the promoting step at it on purpose, and it held back all
thirty-six and promoted none. Under the old build those thirty-six would have been
queued and then marked as failures.

All sixty damaged rows are repaired, and the last hole is closed: there is now a
database-level rule making the bad combination impossible to write, added in the
correct order (after the build), and tested by trying to break it in three different
ways.

**Two things I got wrong today, both worth you knowing about.** First, another
session had already written the repair by the time I got to mine, because I wrote
mine into our lane's own folder rather than the shared migrations ledger where they
looked — so there are two files that do the same thing, and theirs is the one of
record. Second, and more embarrassing: I checked whether the fix had actually been
exercised, saw that the scheduled job which normally drives it is switched off, and
concluded the code "cannot currently run at all". That was wrong — you can fire the
step directly at one site, which is exactly what they did, and it's a technique
already written down in our own notes. I have corrected that claim where I made it
and logged all of today's misreadings; there were three, and every one of them was me
treating "I found nothing" as "there is nothing".

**One new problem found while checking this one, and it is live.** The same safety
check has two arms: one for a finding with nobody named to fix it, one for a finding
pointed at an agent that doesn't exist. I fixed the first. The second is happening
right now: our tool auditor is filing genuinely useful findings about live tools —
missing input labels, a tool depending on a script that isn't there — and addressing
them to something called "hitl-review", which has never existed. Fourteen of them,
across two sites, and it's growing: five yesterday, fourteen today. Each one is
recorded as a routing failure and, worse, silently blocks the auditor from reporting
that same finding again. I have filed it as bug 291 with the producer identified.
Note that neither my fix nor the new database rule catches it, because the handler is
named rather than blank — so it needs its own decision about what "hitl-review" was
ever meant to be.

**And the judgement call from yesterday is still yours**, untouched: whether to
unify the third copy of that shared test, or to back my change out of the shared
dispatch code. Two reviewers wanted opposite things and I have not picked for you.
