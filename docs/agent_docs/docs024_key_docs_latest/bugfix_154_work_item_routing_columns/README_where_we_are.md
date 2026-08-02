# Where we are — bug 154, the work item whose own data was invisible

Plain-prose log, append-only, newest at the bottom.

---

## 2026-07-31, evening — what this bug actually was

The platform keeps a queue of jobs, one row per job. Different agents put jobs
on the queue; a dispatcher picks them up and hands each one to whichever agent
is meant to do it.

One of those agents, `tool-auditor`, files jobs that say "this tool on this site
needs fixing". Every single one of them died instantly, at the first step of the
agent meant to fix it, with an error saying it had not been told *which* tool.
Jobs filed by three other agents went through fine.

The odd part — and the previous session spotted this and wrote it down — is that
the jobs which failed were exactly the ones that **did** say which tool. The
column was filled in. The error said it was empty.

Both things were true, because there are two places a job can record the tool,
and only one of them was being read. The queue table has a proper column for it.
There is also a free-form blob attached to each job, and by convention most
agents drop a copy in there. The dispatcher was only ever looking in the blob.

The reason is a single line of code. The dispatcher builds its picture of a job
from a database query, and that query never asked for the column. It asked for a
neighbouring one — `page_id` — and stopped. So three real columns
(`component_id`, `entity_id`, `affected_url`) were invisible to every agent in
the system, no matter what was in them.

**`tool-auditor` was the only agent doing it properly, and that is precisely why
it was the only one broken.** Everyone else duplicated the value into the blob
and got away with it. Nothing warned about any of this: the dispatcher marks that
field "optional", and when it cannot find an optional field it quietly moves on.
The complaint surfaces two agents later, phrased as though the data was never
there.

## What I changed

The dispatcher now reads all three columns, and for each one it takes the column
if it has a value and falls back to the blob if it does not. That means a single
lookup path is correct for both kinds of job — the 235 existing ones that use the
blob, and the ones that use the column — so nothing has to be rewritten and the
next agent that fills in the column just works.

I checked this actually helps rather than just changing the error message: all
four stuck jobs point at tools that really exist, are live, and sit on a real
page — which is exactly what the fixing agent looks for. So they should now run.

## Two things I deliberately did not do

There was a tempting shortcut: copy the column's value into the blob, and then
nothing else needs changing at all. I nearly took it. It turns out another part
of the system reads that blob to decide whether a rebuild should cover the whole
site or just one component — so quietly putting a value in there could have
changed how sites get rebuilt, for reasons nobody would ever connect back to
this. I left the blob alone entirely.

The second: there is a fourth column, `page_id`, that is handled slightly
differently from the three I fixed, and it would be tidier to make it match. But
218 existing jobs would suddenly start carrying a value they do not carry today,
and the agents receiving them have not been written with that in mind. Tidiness
is not a good enough reason to change what several hundred jobs hand to their
handlers. I have written down the measurement so someone can revisit it with
evidence rather than instinct.

## Where it stands

The code is written, tested, and committed. I proved the test can actually fail
by deleting the fix and watching it go red — a test you have never seen fail
tells you nothing.

Two reviews are running: the automated diagnosis loop, which re-derives the cause
independently, and the reviewer council on the code itself. Neither had returned
when I committed, which is normal and expected here — the tree is shared, so
holding work back is not actually an option, and the commit is tagged with the
review's id so it gets credited automatically when the verdict lands.

The remaining step is that the fix is in the **program**, and the line of
configuration that uses it lives in the **database**. The database is live
instantly; the program only changes when a new image is built and rolled. So the
config change has to go second, and I have written the exact ordering down. Until
then this is inert — correct, committed, and not yet doing anything.

One more thing the previous session found and I have left alone: `tool-auditor`
names its jobs after the *site*, not the tool, so two jobs for two different
tools on one site are indistinguishable. That is a real problem but it is a
different one, in a different place, and fixing it would change how jobs are
de-duplicated across the whole fleet. It stays on the ticket rather than riding
along in a change about something else.

## 2026-07-31, late evening — the build went out, and both halves are now live

A fresh chassis image (`v1.0.1219`) was deployed, so I did the two things that
were waiting on it.

First I checked the fix was actually *in* it. That sounds like paperwork, but a
deployment tells you a new image is running — not that it contains your change,
because the image carries no record of which commit built it. So the check is to
look inside the running program for a piece of text my change added, on both
copies of it, alongside a second piece of text that was already there. The second
one matters: if it also came back missing, that would tell me my search was
broken rather than my fix absent. Both copies, both strings, present.

Then I applied the configuration half — the one line that tells the dispatcher to
read the proper column. It went in cleanly, and it took its own snapshot on the
way through so it can be put back.

So the bug is fixed and live. What is left is one observation: watching a real job
go through and get past the step that used to kill it.

## The bit that is still running

With your go-ahead I reset one stuck job on gamesdesign.co.uk so it would be
picked up again. It has not been picked up yet — about twenty minutes so far —
and I want to be clear that this is not a sign of trouble. I checked three things
that could have meant something was wrong, and all three came back healthy: the
dispatcher is alive and working, the site is not locked or blocked, and my job is
second in the queue for that site, well within the batch it takes.

What it is actually waiting for is the dispatcher's habit of picking one site at a
time, fairly arbitrarily, from a large list. So gamesdesign will come up; it just
has not come up yet. Nothing about that reflects on the fix either way.

I have left a watcher running and written a full handoff, so this can be finished
either here or in a fresh conversation without re-deriving anything. The one
remaining question is a single yes/no: does the job get past the step that used to
fail? Anything beyond that step is a pass, because that failure was the very first
thing the job did.

I have not closed the ticket yet, for that reason. Closing it now would mean
claiming a result I have not seen, and the whole point of that step is that it is
the one thing nobody has actually witnessed.

---

2026-08-02. The waiting is over, and it ended by fixing the thing that made the
waiting indefinite rather than by waiting harder.

A correction to my own paragraph above first: I said the dispatcher picks sites
"fairly arbitrarily" and that gamesdesign "will come up". Both wrong. When we
finally read the picking query, it always chooses the eligible site whose ID
sorts lowest — a fixed order, like being 14th in a queue where the first 13
places are permanently reserved. gamesdesign could only ever be picked if all 13
sites ahead of it happened to be busy at the same instant. It had the oldest
waiting work in the entire fleet — three and a half days — and that counted for
nothing, because the picker never looked at waiting time. It was never going to
"come up".

The owner looked at this and made two calls: fix the picking rule, and while in
that file, remove a leftover reference to a database column that no longer
exists (harmless to the live system, but a trap for anyone rebuilding from the
seed file — the rebuilt trigger would crash on its first run).

For the picking rule the owner chose, from four options laid out with their
trade-offs: serve whoever has waited longest. It is the simplest rule that
guarantees nobody waits forever — new work always joins the back of the queue,
so old work always gets to the front eventually. Priorities still decide the
order of jobs within a site, exactly as before; they never worked across sites
and still don't, deliberately — making them work across sites is a policy
decision for another day, written down where it can be found.

The change went in just after 9:30 this morning, with the platform's own
snapshot taken first so it can be undone in one step. Within two minutes, the
very next scheduling tick picked gamesdesign — the site nothing had picked in
three and a half days. Its queue is now being worked through, and the item we
have been trying to witness for two days is second in line in that very batch.

So the last step of this whole bug is now minutes-to-hours away, happening the
way it should: the platform doing its own work under a fair rule, not us
poking it by hand.

---

2026-08-02, later the same morning. Done — witnessed, and closed.

Four minutes after the fair-queue rule went live, the platform picked the
starved site on its own, worked down its queue, and reached our test item. The
job that used to die on its very first step ran all the way through in under a
minute: it loaded the tool, improved it, saved the improved version, and queued
the page update that follows. We checked the actual saved artefact, not just
the status flag — the tool's component was genuinely rewritten, full-sized, at
the exact moment the job's log says it saved it.

That was the one thing this ticket was still waiting for: seeing it happen for
real, not arguing from the code that it should. It happened the right way too —
the platform doing its own work under the new fair rule, nothing hand-fired.

The ticket is closed and moved to the done pile. A few genuinely separate
questions stay written down where the next person will find them: two sibling
fields that would hit the same bug if anyone ever starts using them (fix is
written down, waiting for a first real case), a question about whether page
routing should work the same way (touches 218 old items, so it's an owner
call), and a small labelling defect in how the auditor names its work items.
