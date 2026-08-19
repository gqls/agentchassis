# Where we are — bug 315, "the database says the page was published and it wasn't"

Plain-prose running log for the owner. Append only, newest at the bottom.

## 2026-08-19, morning — starting on it

**The bug in one sentence.** When a page is rebuilt, the database writes a timestamp called
`deployed_at` that is supposed to mean "this page is now live" — but nothing in the system ever
checks whether the page actually reached the website, so the timestamp gets written either way.

The lane that found it (the one rebuilding the webdesign.co.uk tools) hit the real-world version:
a tool page was rebuilt four separate times, all four rebuilds reported success, the timestamp was
refreshed each time, and the public website carried on serving the *old* tool for about six hours.
Then it published itself with nobody doing anything. A second lane found the same shape on a
different site — a page marked active that has never existed on the web at all, with three
successful rebuilds behind it.

**Is it still real?** Yes. I re-measured everything this morning rather than trusting the file:
42 live pages across 14 sites have no content in them at all, and two of those are marked as
successfully deployed. Those numbers are unchanged from yesterday.

**Whose is it?** The lane that filed it says plainly, twice, that it is not theirs to fix — they
found it while doing something else. So nobody is working on it and I am not stepping on anyone.

## What I have found so far, and one thing I did not expect

I traced the whole path a page takes from "rebuilt" to "on the website". It goes: the platform
writes the new page, hands it to a small service that commits it to a shared GitHub repository, and
GitHub then copies the changed folders up to the storage bucket the public site is served from.
The phrase the platform's own documentation uses is **"commit is deploy"** — there is no separate
publishing step, which is why the timestamp has nothing to wait for.

Two things stand out.

**First, the timestamp is written in the wrong place — and in two cases, at the wrong time.** There
are five places in the system that stamp a page as deployed. Three of them stamp it just after
handing the page to the commit service, but they never look at what that service came back and
said — so if it came back saying "there was nothing to commit", the page is still marked deployed.
The other two stamp the page as deployed **before the commit has even been requested.** So this is
not a subtle race. There is no arrangement of these five workflows in which that timestamp could
honestly mean "this page is live".

**Second, and this is the one I did not expect: the machinery to do this properly was already built
and then never plugged in.** The database has a column on pages for a content fingerprint, and
another on page sections for the commit that deployed them. Both exist. Both are completely empty —
zero rows out of 786 and zero out of 1,775 respectively — and searching the entire codebase,
including the tests, finds *no code that writes to either of them*. Somebody designed exactly the
traceability this bug is asking for, and it was never wired up.

That changes what the fix is. It is not "invent a way to prove a page published". It is "start
writing the two things the schema already has room for, then compare them against what the website
is actually serving". That is a much cheaper piece of work, and it is what the house rules ask for
anyway — reuse what exists before building something new.

**A related thing I corrected while I was in there.** The platform's own concept register — the
document other parts of the system, including the automated code reviewers, read as authoritative —
states that commit references *are* recorded on pages and work items for traceability. That is
wrong in all three of its parts: there is no such column on pages, none on work items, and the one
column that does exist has never been written to. I have marked the correction in place, because a
reviewer reading that line today would conclude the traceability already exists and would push back
on a proposal to add it.

## One operational note that is not about this bug

At about 10:25 this morning the fleet's AI provider started refusing requests with "you have
reached your specified API usage limits". It knocked over my first diagnosis run. Before reporting
that as an outage I checked the history: the same message has appeared on five separate days over
the last month and the fleet carried on each time, and the system re-queued my run by itself within
two minutes. So it looks like an intermittent spend ceiling rather than the hard lockout the wording
suggests. Worth knowing about, not worth acting on from here.

## Where I am going next

I have asked for a full fix plan and I will bring it back with a recommendation rather than a menu.
The shape it will take: make the commit step report what it actually committed, write that down in
the columns that already exist, and then add a periodic check that compares what we believe we
published against what the website is really serving. The last of those is the piece that would
have caught this bug the first time it happened, six hours before anyone noticed.

## 2026-08-19, late morning — the plan, and what changed my mind twice

The plan is written and it is at the review council now. Two things I found along the way changed
what the fix should be, and both of them are the kind of thing worth saying out loud.

**The first is that part of the reported symptom turned out to be correct behaviour.** The bug's
evidence table lists three pages whose "published" timestamp was newer than the file actually on the
website, and treats all three as the same problem. They are not. The platform's own notes record
that when a page is rebuilt and comes out **byte-for-byte identical**, the commit that carries it is
an *empty* commit — nothing changes, so nothing gets copied to the website, so the file's date
rightly stays where it was. Two of those three pages are exactly that case. They were fine.

That matters because the bug file's suggested fix number four was "alert whenever the timestamp is
newer than the website's file". That check would have raised the alarm on both healthy pages. And
when I ran it properly across forty pages, it flagged **all forty** — because the website is updated
in one batch per site, tens of minutes after the rebuild, so at any given moment most perfectly
healthy pages look "stale" by that test. So the cheap check does not work, and the more careful one
(compare a fingerprint of what we meant to publish against what is actually being served) is not an
alternative to the fingerprint work — it *depends* on it.

**The second is that I nearly reported an outage that was not one.** For about an hour I could see a
steady stream of successful commits and no corresponding updates to the website, and both deployment
machines sitting idle. That reads like a broken pipeline. Then I looked at the machines' own history:
they do not run continuously, they run in bursts twenty-five to fifty minutes apart, and the gap I
was staring at was inside that normal range. So I stopped and set a background check running instead
of writing it up. **At the time of writing that gap has reached seventy minutes**, which is longer
than any normal spacing I can see in the history, so it may yet turn out to be real — but it will be
reported as a measurement, not as an alarm.

**What the fix does.** In plain terms: at the moment we hand a page to the deployment system, we
write down two things we currently throw away — the identifier of the commit that carried it, and a
fingerprint of the exact bytes we sent. Both have a place waiting for them in the database already.
Then the timestamp is only written if the deployment step actually reported doing something, and
separately, on a schedule, we compare the fingerprint against what the website is really serving and
raise a flag when they disagree for longer than the normal delivery window.

The first step of that needs no new code at all: two of the five places that mark a page as deployed
do it *before* the deployment is even requested, and both of them go on to call a routine that marks
it properly afterwards. Deleting the premature step is a configuration change that takes effect
immediately, and on its own it removes the worst half of the problem.

**One more correction, and it is about our own records.** The platform's internal reference document
states that commit identifiers are recorded against pages for traceability. They are not — there is
no such field, and the one related field that does exist has never been written to by any code in
the repository. I have corrected that entry, because the automated reviewers treat it as fact, and
it would have led one of them to object that we were adding something that already existed.

**A note on what I could not do.** The final "why was this one page skipped" question cannot be
answered from here: that logic lives in a private repository we have no read access to from this
machine. The plan works around it deliberately — the fingerprint check detects the failure from our
side without needing to see inside the deployment machinery. And the automated diagnosis service I
would normally run this past refused twice on an API spending limit, so I have said plainly in the
record that I substituted my own first-hand checks and listed exactly what they were.

## 2026-08-19, midday — the review came back "revise", and it was worth every minute

The council reviewed the plan and asked for changes. Five of its seats approved, five objected. I
want to record what they caught, because two of the objections were right in ways I would not have
found on my own.

**The first: I overclaimed.** My summary said the change would mean the "deployed" timestamp is only
ever written after a real deployment. That was not true of the plan I submitted. The part that does
the checking was designed to be switched on per-agent, and I had not included the step that switches
it on for any of them — so three of the five offending agents would have carried on exactly as
before, with a new but dormant safety mechanism sitting beside them. The reviewer put it plainly and
was right. The fix is not better wording; it is saying honestly what the change does and does not do.

**The second: I was about to undo somebody's deliberate decision.** My plan added a column to the
pages table to hold the deployment reference. One reviewer objected that this rested on an unchecked
assumption. Checking it found something better than either of us expected: that column was removed
on purpose years-of-commits ago, with the stated reason that it belongs on the page-sections table
instead — and the migration that records this goes further, saying that deciding whether to wire this
up at all "is an owner call, not a bug fix". So that part is out of my plan and is a question for
you (below).

**The third came from the architecture seat**, which ruled that changing what the deployment service
reports back is not a bug fix at all — that response is consumed by nineteen different steps across
sixteen agents, and I had checked none of them for how they read it. It told me to ship the small
safe part now and take the rest to architecture review. I have done exactly that: the small part is a
configuration change, and the rest is written up as RFC 038.

## And then I caught myself being wrong about something bigger

Earlier I reported that forty pages looked stale — the database saying "published" while the website's
files had not been touched for nearly an hour and a half. I hedged it and set a check running.

**All forty were fine.** I proved it by pulling a page's content out of the database, cutting a
distinctive chunk out of it, fetching the live page, and searching for that chunk. It was there. The
website is serving exactly what the database holds. Those pages had simply been rebuilt into
*identical* content, which produces an empty change that correctly copies nothing.

This is the most useful thing I found all morning, because of what it took. Four steps and a
judgement call, for one page — and until I did it, "these pages never needed republishing" and "these
pages failed to republish" looked **completely identical** in every signal the platform produces.
That is the bug, stated properly. It is not that pages fail to publish. It is that nothing we have
can tell those two situations apart.

It also kills the cheap version of the fix outright. The bug report suggested alerting whenever the
timestamp is newer than the website's file. I ran precisely that, and it produced forty confident
false alarms on our busiest site, and they stayed false for eighty-five minutes — longer than any
sensible "give it time to settle" allowance. So the fingerprint approach is not a refinement of that
idea; it is the only version that works.

## What I need from you

**One decision.** The platform has two unused database fields designed for exactly this — a content
fingerprint on pages, and a deployment reference on page sections. Both are empty, neither has ever
been written to by any code, and this is now the *third* time somebody has independently discovered
they are empty and walked away. The note left by the last person says wiring them up is your call,
not a bug-fixer's. **Do we wire them up, or drop them?** Everything else in this fix depends on that
answer, and I have deliberately not taken it myself.

**One thing you may want sooner.** There is a configuration-only change ready that removes the worst
half of the problem: two agents currently mark a page "deployed" *before* they have even asked for it
to be deployed, and both already call a routine that marks it properly afterwards. Deleting the
premature step needs no new code and takes effect immediately. The architecture reviewer looked at
this specific piece and said "clean point fix — proceed". I have not applied it, because it changes
the live build pipeline the moment it runs and that felt like your call to make, not mine.

## 2026-08-19, afternoon — the premature stamp is gone, and the decision you asked me to explain

**Done first: the change you authorised is live.** Migration 491, applied 15:20Z. The two agents
that marked a page "deployed" before asking for it to be deployed no longer have that step; the job
now falls to the routine they already call, which does it after the deployment has actually been
requested. Verified at the live configuration, not just at the migration's say-so: all four
remaining places that write that timestamp now come *after* a deployment request. It was two out of
five; it is now none. The snapshots taken beforehand were checked to hold the *old* configuration —
a snapshot that exists but holds the new value would restore nothing.

One honest caveat: this is proven at the configuration, not yet watched at runtime, because no page
has been built through the new path in the twenty minutes since. If it were wrong, the failure would
be pages staying *un*-marked rather than falsely marked, which is the recoverable direction, and the
rollback is written and ready.

---

## The decision, explained

### What these two things actually are

**A content fingerprint** (the `content_hash` field on pages) is a short string calculated from a
file's exact contents. Change a single character anywhere in the file and the string changes
completely. It is the ordinary way to answer "are these two files identical?" without comparing them
character by character. There is a field for one on every page in the database. It is a 64-character
text field — exactly the size the standard calculation produces — so somebody set it up for precisely
this and never filled it in.

**A deployment reference** (the `deploy_commit` field on page sections) is meant to hold the
identifier of the specific save that carried that section to the website — a reference you could look
up afterwards to see exactly what was sent.

Both fields exist on every relevant row. Both are **completely empty**, everywhere, and no code
anywhere in the system — including the tests — has ever written to either.

### Why it is your decision and not mine

Because a previous session wrote it down as yours. When a related unused setting was cleaned up,
whoever did it left a note saying that the column is already there, that its being empty means "never
built" rather than "never deployed", and that **deciding whether to wire it up or drop it is an owner
call, not a bug fix**. I found that note only because a reviewer refused to let my plan assume
otherwise — my original plan would have quietly taken the decision by adding a similar column back to
a table it had been deliberately removed from.

### The thing that makes the decision easier — they are not a pair

I assumed at first that these two fields were two halves of one idea. They are not, and it matters:

- **The website serves one file per page.** So a fingerprint stored *on the page* is exactly the
  right shape for the question this bug is about: take the fingerprint of what we sent, later
  fingerprint what the website is actually serving, compare. One step, no judgement.
- **A section is not a file.** So a deployment reference stored *on the section* cannot answer "is
  this page published" at all. It answers a different and narrower question — "which save carried
  this particular section's last change" — which is useful for tracing history but is not this bug.

So they can be decided separately, and I would decide them differently.

### What each option costs

**Wire up the page fingerprint — my recommendation.** This is the piece the whole fix rests on.
Without a record of what we intended to publish, "this page never needed republishing" and "this page
failed to publish" are indistinguishable — that is the bug, and no amount of cleverness about
timestamps gets round it, as I demonstrated by producing forty confident false alarms this morning.
It cannot land on its own, though: it needs the deployment service to hand back the fingerprint of
what it actually sent, which is the piece now in architecture review as RFC 038, because that
response is consumed by nineteen different steps and nobody has surveyed them.

**The section-level deployment reference — I would drop it, or leave it.** It does not help this bug,
nothing has ever needed it in the months it has existed, and keeping an empty field that looks
purposeful is exactly what has now cost three separate sessions time. If you want deployment
traceability later, it is a one-line addition at that point. The only argument for keeping it is that
adding columns back is mildly annoying, which is not much of an argument.

**Do nothing.** This is the option that has effectively been chosen three times already — once in
early August by the lane that measured the fingerprint field empty and worked around it, once by
whoever wrote the cleanup note, and once by me if I had not been stopped. The cost is not that
anything breaks; it is that the same discovery gets made again, and that this bug stays unfixable in
principle rather than merely unfixed.

### What I need from you

Just one answer, really: **wire up the page fingerprint, yes or no?** If yes, RFC 038 becomes worth
pushing through and the rest follows from it. The section-level field is a tidy-up either way and I
am happy to take silence on it as "leave it alone".
