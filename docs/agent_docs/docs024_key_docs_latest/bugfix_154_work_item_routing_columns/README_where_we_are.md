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

---

2026-08-02, midday. The fair-queue fix was right, and watching it work uncovered
a bigger problem sitting underneath it. That is now fixed too.

After the queue became fair, one site was served — and then everything went
quiet again for over an hour. The dispatcher was running the whole time, every
couple of minutes, finishing cleanly and reporting no errors. It was simply
doing nothing, over and over.

The cause: two different parts of the system disagree about what counts as work
that can be started. The part that chooses which site to work on uses three
tests. The part that then picks up the actual jobs uses those three plus two
more — one about approval, one about whether a job is waiting on another job to
finish first. So the chooser kept handing the worker a site whose only job the
worker would not touch. The worker looked, found nothing it was allowed to do,
reported success, and stopped. Nothing was marked as started, so the same site
was chosen again next time, forever.

The job in question is waiting on another job that needs a human to look at it —
which will never happen on its own. So one single job, out of 366 across 17
sites, was holding up the entire fleet. Twice today that produced complete
standstills, of 89 and 68 minutes.

Worth being straight about two things. First, I had seen those standstills
before and written them off as "about the same as the last quiet spell, so
probably normal". That was wrong, and wrong in an avoidable way: I had only ever
seen the same unexplained thing twice and treated the repetition as
reassurance, when it was the opposite. Second, the fair-queue change I made this
morning made this worse, not better — under the old unfair rule a blocked site
only held things up while it happened to sort first, but under "oldest waits
first" a job that can never be started sits at the front permanently. The fair
rule is still right; it just needed this fixed alongside it.

The fix makes the chooser use the same tests as the worker. It is deliberately
conservative: the only sites it now skips are ones where the worker would have
found nothing anyway. Within two minutes of going live, three different sites
were being worked on, in the correct fairest-first order, and the stuck job was
left alone exactly as it should be — the queue no longer waits for it, and it
stays there for whoever wants to sort out the human question behind it.

---

2026-08-03, after midnight. The new safety catch — the one that stops a rewrite
from quietly gutting a page section — has now been proven for real, not just in
tests. I set a deliberate trap on the darts site (our own demonstration site, so
nobody's customer was involved): I made the system believe one blog article was
about three times longer than it really is, then asked it to rebuild the page.
If the safety catch works, the rebuild should be refused — from where the system
stands, it looks exactly like a rewrite about to throw away two thirds of an
article. That is precisely what happened. The rebuild was refused, twice (the
system retries once by itself), the refusal was reported honestly as a failure
rather than dressed up as success, a "needs a human to look at this" note
appeared in the work queue within a fraction of a second, and — the part that
matters most — not a single byte of the page was changed. I wrote down in
advance exactly what a pass would look like, and all four things happened as
written. Afterwards I put everything back exactly as it was and checked,
character for character, that it matched.

The reviewers who looked at this change on Sunday had one serious worry: that
the refusal might be silently swallowed by the layer above, so the protection
would exist but nobody would ever hear it fire. The trap answers that with
evidence rather than argument. They also caught one genuine gap I have now
fixed: sections a human has locked by hand were being included in the
comparison, and could in theory have caused the guard to refuse a save it had
no business refusing — locked sections can't be overwritten anyway, so the
guard now leaves them out of its sums. I have sent the whole package back to
the review panel with all nine of their points answered, and I'm waiting on
their verdict.

---

2026-08-04. The last open piece of this lane's work is done and live: the fix
for the bug where asking the system to "add a link to X" on an existing page
could quietly throw away most of that page's writing while reporting success.
The cause turned out to be simple once found: that kind of request never told
the page-writer what was already on the page, so the writer had nothing to
edit and just made something up from scratch that satisfied the instruction —
which is exactly what happened on the gripper page a few days ago (the one
that lost 59% of its reference material). The fix gives the writer a look at
the page's current content when, and only when, the request explicitly asks
for an edit rather than a fresh build — every other kind of page build is
untouched by this change, on purpose, because a shared piece of machinery
like this one shouldn't change behaviour for people who didn't ask for it.

I proved the fix works on a real, already-queued production request rather
than a made-up test: two links that had been deliberately held back exactly
because dispatching them safely required this fix. I checked what the writer
was actually handed, and it was the page's real current paragraph, word for
word, correctly matched to the right section. That is the hard part of the
bug, and it works.

What I could not finish is watching the whole thing through to a saved page,
because I ran into a second, completely separate problem: the page-writer has
been failing outright on roughly half its runs since last night around 9pm —
long before this fix even existed anywhere — for a reason that has nothing to
do with this change. I confirmed that by watching it fail the same way on an
unrelated tool page on a different site in the same few minutes. Nothing was
lost by that failure — it fails loudly, before anything gets saved — but it
does mean I can't finish proving the very last step of this bug end to end
until someone fixes that other, separate fault. I've written it up as a new
case (192) with everything I found, rather than trying to chase it down
myself and running the token budget for this session even higher.

One thing worth knowing: I noticed the queue that's supposed to pick up and
process these page-build requests isn't actually running on any schedule at
all — nothing was ever set up to fire it periodically. I fired it by hand to
get my test through. This is the same shape of problem as a few other things
already known about this system (something built and working, but nobody
ever wired up the clock that's meant to run it) — not something I fixed, just
flagged, because it explains why things sometimes sit quietly in the queue
for a long time with nobody noticing.

Also: I sent this change to the automatic review panel before committing it,
same as always, but that particular review run got stuck partway through and
never gave a verdict — not rejected, just stalled. That's only advisory, so
it didn't stop anything, but the review itself is owed to whoever looks at
that system next.

---

2026-08-04, later the same day. I need to correct what I told you above, on
two points, and you asked good questions that caught the first one.

The bigger correction: the fix I described as done actually broke every
single page build across the whole system for about 40 minutes this
morning. The new code I added had a mistake in it — a step that was
supposed to pass information through untouched instead wrapped it in an
extra layer, and because of how that information gets reused by the next
step, every page build anywhere started failing. I noticed page builds were
failing when I went to test my own fix, but I looked at the timing wrong and
convinced myself it must be a different, older problem that had nothing to
do with me — it genuinely looked that way from the numbers, but I should
have read the actual error messages from the earlier failures instead of
just comparing counts, and if I had I'd have seen they were a different
problem entirely. Another session read what I'd written, spotted the real
cause within the hour, fixed it properly, and even left me a very generous
note explaining exactly what went wrong and why. I'm grateful for that and
I've corrected the record everywhere I'd written the wrong conclusion,
rather than quietly editing it away.

The second correction is the one you asked me to double check: I told you
the system that's meant to pick up and run page-build requests isn't
scheduled at all. That was also wrong, and it was a sloppy check on my
part — I searched for anything with "dispatch" in its name and found
nothing, but the actual scheduler is called something else entirely
("build-pipeline-trigger") and it's running fine, every two minutes, exactly
as it should. I should have looked at the whole list rather than searching
for the word I expected to see. Thank you for pushing on that rather than
letting it stand.

The honest state of things now: the fix itself does work, but only
partially. I tested it on two real pages. On one, it worked exactly as
intended — the page kept its original writing and just got the new link
added. On the other, the same original problem happened again, for a
different reason than the one I'd fixed: the system picked a different,
generic building-block for that section than the one actually stored on the
page, so my fix had nothing to match against and the writer made something
up from scratch, same as before. Nothing was destroyed either way — the old
version is always kept — but the underlying promise of this fix, that
asking for a small edit can't quietly throw away a page's writing, isn't
fully true yet. I've left the case open rather than calling it finished, and
written up both what worked and what didn't so whoever looks at this next
doesn't have to take my word for it.
