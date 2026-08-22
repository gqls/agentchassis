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

## 2026-08-19, late afternoon — the fingerprint is built, and it needs your build

**Your decision is implemented, both halves, tested and committed.**

**One:** the deployment service now reports what it actually did. It used to hand back the
repository's web address — the same string every time, for every page, for ever — while throwing
away the reference to the save it had just made. It now returns that reference, plus a fingerprint
of the exact bytes of each file it was given.

**Two:** the page-marking step now reads that. When a deployment reports it did nothing, the page is
no longer marked as deployed. When it succeeds, the fingerprint is written to the page. If nothing
can be read at all it marks the page anyway and logs a countable warning — deliberately, because a
typo in configuration must not freeze the whole fleet's deployments, and because a *silent*
fall-through is precisely how this bug survived four rebuilds.

**Two places I narrowed my own plan, both for reasons I'd rather state than bury.** I dropped a
"nothing changed" flag I had promised: producing it needs an extra call to GitHub on *every* save
across nineteen places in the system, to fill in a field the reviewers had already ruled
advisory-only — and the fingerprint answers the same question better. And I graded the "couldn't read
the evidence" warning lower than planned, because until *both* services are rebuilt the newer one
will fail to read the older one on every single deployment; logging that at high severity would have
made the error log useless for a day.

**The riskiest thing I found while building it** is that files sent as images are wrapped in an
encoding, and fingerprinting the wrapper instead of the file would produce a value that could never
match what the website serves — silently, permanently, on every image. That is live code, not a
hypothetical. I wrote the test first, then deliberately broke the fix to confirm the test catches it,
then restored it.

**The earlier change is confirmed working in the wild.** Thirty-one pages have been built since this
morning's configuration change and every one of them is marked deployed, with the mark coming from
the rebuild step as intended. None stranded. No rollback needed. (A second session, working at your
request, spotted this first; I re-ran it myself before recording it, because a report from another
session is still just a report.)

**What I need from you: a build.** Nothing above is live. Two images need rebuilding and rolling —
the deployment service and the main chassis — and by our own rules releases are yours to run, not
mine. Everything is committed, so a build will pick it up.

**And one thing must happen in an unusual order.** The configuration that switches the new checking
on is *deliberately held back* and will not be applied by the normal migration run. It has to wait
until the rebuilt chassis is actually running, because the page-marking step rejects configuration
keys its own code does not recognise — so applying it early would break page-marking everywhere at
once. The file carries the exact commands and the verification step at the top, and the migration
tool lists it as held rather than pending, so it cannot go in by accident.

Once that is done, the platform will finally be recording *what it sent*. The last piece — a
scheduled check comparing that against what the website is actually serving — is designed and not
yet built, and it is the piece that would have caught this bug six hours before anyone noticed.

## 2026-08-19, evening — the build shipped it, the reviewers found a real flaw in it, and it is fixed

**First, the good news about your build.** Both services rolled at 17:13 as `v1.0.1316`, and I
confirmed they genuinely carry my code — not by trusting the version number, which proves nothing,
but by reading the commit reference the deployment service prints about itself at startup and
checking my commits are behind it. (The other method, searching the running binary, was actively
misleading here: it reported my code absent *and* reported a deliberately fake control as present,
which is a known trap.)

**Now the important part: the reviewers found a genuine flaw and they were right.**

I had justified a piece of the design by relying on a promise made by a shared helper elsewhere in
the codebase — that when it finds two conflicting answers it returns nothing rather than picking one.
A reviewer cross-checked that against our own hazard register, which said the opposite, and objected
that both could not be true.

They were right and I was wrong. The helper's comment states that rule as the *intention* — and then,
four lines further down, states that the current version still picks a winner, and that the refusing
version is a later piece of work that has not been done. **I read far enough to find a sentence that
supported what I wanted, and stopped.**

That mattered more here than it usually would. A fingerprint taken from the wrong deployment is
silently and permanently wrong, and every future check would then report a perfectly healthy page as
broken. No fingerprint at all is recoverable; someone else's fingerprint is not.

**The fix is to make the promise true rather than restate it.** My code now collects the candidates
itself and refuses outright when they disagree. I proved it by deliberately restoring the old
guessing behaviour and confirming the new tests fail — then restoring the fix and confirming they
pass.

**And re-deriving that turned up a second mistake nobody had objected to.** I had been careful to
*preserve* an existing fingerprint when a new one could not be worked out. That is backwards: marking
a page deployed means new bytes went out, so the old fingerprint is out of date by definition, and
keeping it would make the future check compare live content against a superseded record — convicting
a healthy page, which is exactly the false-alarm problem I spent the morning proving was fatal to the
cheap version of this fix. It now records "unknown" instead, which the check is designed to skip.

**One objection I could answer rather than act on.** A reviewer worried that only three deployment
steps get the new protection while sixteen others do not. I measured it: of the nineteen, exactly
three are followed by a step that marks a page as deployed — and those three are precisely the ones
being switched on. The other sixteen publish stylesheets, scripts and feeds, and make no claim about
a page at all. Coverage is complete, not partial.

## What this means for the next step

**The build you just did carries the flawed version.** It is completely harmless as things stand,
because the switch that would activate it is still held back and set nowhere — the code cannot run.
But it does mean:

**Please don't apply the held migration yet, and one more build is needed.** Once a build carrying
this evening's fix has rolled, the held configuration can go in and the fingerprint starts being
recorded for real. I've put the exact verification command in the migration file itself, so whoever
applies it can confirm the running image is the right one first rather than trusting the version
number.

## 2026-08-19, night — approved

The reviewers passed it on the third round: **approved**, with two advisory notes and nothing
serious. The trail is worth recording because it is the argument for bothering with review at all —
round one reviewed the *plan* and found I had promised more than I was delivering; round two reviewed
the *code* and found a claim in it that was flatly untrue; round three approved the corrected
version. The second round is the one that earned its keep.

**I acted on one of the two advisories, and it was a real if latent bug.** The statement that marks a
page includes an optional piece, and I had written the position of one of its values as a fixed
number. It happens to be right today. It is a trap for whoever next adds a value to that statement,
and it would fail at run time, on the deployment path, invisible to the compiler.

**More usefully: the test I wrote for it was worthless and I caught that myself.** It contained its
own copy of the code rather than calling the real thing — so it would have passed cheerfully while
the actual code was broken, which is exactly what it existed to prevent. What caught it was writing
the commit message: I typed "pins the invariant" and had to ask *pins it where*. The answer was "in
the test file". It now calls the real code, and I proved the test can fail at all by simulating the
future it protects against.

**One commit is stuck, deliberately.** Another session is working inside the same file as me and has
called a function whose definition they haven't committed yet. Because a commit takes the whole file
as it stands, committing mine would carry their half-finished call without its other half — and since
builds are made from committed code, that would break the build for every session on the machine. So
it waits. I have a watcher on it and it will go in the moment their side lands. Nothing is at risk:
the code involved cannot run at all until the held configuration is applied, and that is still held.

**Where that leaves you:** the position is unchanged from this evening — one more build, then the held
migration. The only difference is that the fix inside that build has now been through review and
come out the other side.

## 2026-08-20, early morning — it is switched on

**Your build did it.** `v1.0.1317` carries the corrected, reviewed version, and I have switched the
new checking on for all three places that mark a page as deployed.

I want to be precise about how I checked, because the version number proves nothing on its own.
The deployment service prints the exact code revision it was built from when it starts, and the fix
is behind that. For the main chassis that line had already scrolled out of reach, so I searched the
running program for two function names instead — one that exists only in the corrected version, and
one that exists only in work I have *not* shipped yet. The first was present, the second absent. A
check that can only ever say "yes" is worth nothing; this one could have said no, and didn't.

Two of my own checks along the way were worthless and I threw them out: my first "this should be
absent" comparison used a commit that turned out to predate the build, and the second compared the
build against itself. Recorded, because a bad control is worse than no control — it manufactures
confidence.

**The switch-on is confirmed in the live configuration**: all three agents now name the right field,
and they are genuinely three different names — one of them would have been missed entirely by the
obvious approach of hard-coding a single name.

**One instruction of mine was wrong and is now fixed.** I had told whoever applied the held migration
to register it with the migration tool afterwards. The tool refuses — held files are deliberately
outside its remit. Harmless, but the instruction has to be right, so both the file and the runbook
now say to record it in the working notes instead, which is what I did.

## What is not yet proven, and I would rather say so

**The fingerprint count is still zero, and that is because nothing has run.** No page has been
rebuilt since three o'clock yesterday afternoon, so the new code has not been exercised once. A zero
with no traffic behind it is exactly the kind of green light this whole bug is about — I am not going
to report it as a pass.

I have a watcher running, and the fleet rebuilt thirty-one pages yesterday afternoon, so ordinary
traffic should answer this within hours rather than days.

**I could force it, and chose not to without asking you.** The documented tool for rebuilding a single
page regenerates that page's content from source rather than re-publishing what is already stored —
so it is not guaranteed to leave the page identical. Every page I could point it at belongs to another
lane's live customer site. Testing my fix by risking their page is not a trade I will make on my own
initiative. If you would like it proven now rather than on the next natural rebuild, say so and I will
do it — ideally on a page you are happy to have rebuilt.

**When it does fire**, I have written down in the notes what each of the three possible outcomes would
mean, so whoever sees it first — me or another session — reads it the same way.

## 2026-08-20, mid-morning — I have to correct my last entry: switching it on broke the fleet

My previous entry said the new checking was switched on and simply hadn't been exercised yet. **That
was wrong, and wrong in the worst available direction.** Switching it on stopped every page in the
estate from being able to publish, for thirty-three minutes.

**What I got wrong.** The file that describes what settings each part of the system accepts contains
two such descriptions, forty lines apart, for two different parts. I added the new setting to the
wrong one. Because the part that actually reads the setting declares its list of accepted settings to
be complete, an unrecognised entry is treated as a definition error and refuses the whole job rather
than ignoring it — which is correct behaviour and exactly why it caught me.

**The damage:** eight jobs failed outright, and a hundred and twenty-three queued page rebuilds
stopped draining. Armed at 07:49, first failure at 08:01, service restored at 08:22 by another
session, who found it as a blocked page check of their own, traced it to the line, wrote it up, and
ran **my own rollback file** to fix it. I would rather they had done exactly that than waited for me.

**The part I want you to see, because it is the same failure this whole bug is about.** After
switching it on, I checked that the setting was present in all three places, got the three expected
answers, and wrote "verified". That is a status check. I never asked the only question that mattered —
*can a page still be published?* — and it already couldn't as I typed the word verified. My own notes
ninety minutes earlier read: *"config being right is not the artefact — that is this bug's entire
lesson, and it applies to the fix as much as to the defect."* I wrote that sentence and then did not
do it.

And the number I was watching lied to me in a way I had specifically warned myself about. I was
looking at the count of fingerprints recorded, saw zero, and correctly refused to call it a success —
but read it as *"nothing has run yet"* when it meant *"nothing can run"*. Those look identical in that
column. The query that would have told them apart was one line away, and I never ran it, because I was
looking for evidence that my change had worked rather than evidence that it had broken something.

**Where things actually stand.** The mistake is fixed in the code — by another session, not me — and
the fleet is confirmed healthy: no errors of that kind remain and the queue is draining normally. The
new checking is **switched off and must stay off** until the corrected code is in a running build;
switching it on today would reproduce the outage exactly. I have rewritten the switch-on instructions
so they now check the right thing, and so the first thing anyone does after switching it on is look
for damage rather than for success.

**Nothing needs a decision from you.** The next build picks up the fix, and I have left the arming
instructions in a state where they cannot be followed against the wrong build. I am sorry for the
disruption; the honest summary is that I was careful about the wrong things and one of the checks I
skipped was the one this entire piece of work exists to teach.

## 2026-08-20, late afternoon — it works, and here is the proof in one line

**The fingerprint is being recorded, and it matches what the website actually serves.**

```
robot-hands.com/product-detail.html
what we recorded sending:  e9d7090facaaddd3733d11885982979b9710d855df97297c062099bb5b09940b
what the site serves now:  e9d7090facaaddd3733d11885982979b9710d855df97297c062099bb5b09940b
```

Thirty-eight pages across four sites now carry one, where the number was zero for the entire history
of the estate until this afternoon. Nothing has gone wrong: no validation errors, no unreadable
deployments, and the switch survived the latest build.

**What that line replaces.** Yesterday, answering "is this page serving what we sent?" for a *single*
page meant pulling its content out of the database, cutting a distinctive fragment from it, fetching
the live page, searching for the fragment, and then making a judgement call. That is why the original
fault went unnoticed for six hours — nobody was going to do that routinely, and the cheap alternatives
all lie. It is now a string comparison.

It also settles three design choices that could each have been sensible and still wrong: that image
files had to be fingerprinted after decoding rather than as sent; that the two halves had to agree on
how a page's filename is written; and that nothing rewrites the bytes between us and the visitor. A
mismatch in any of those would have shown up as a mismatch here, and it didn't.

**Where that leaves the work.** The thing this bug asked for is delivered and live. The last piece —
a scheduled check that compares the recorded fingerprint against the served page and raises a flag
when they disagree — is designed and not built, and for the first time it is *buildable*, because the
thing it needs to compare against finally exists.

**I have written a handoff** so this can be picked up in a fresh conversation without re-reading
everything:

`docs/agent_docs/docs024_key_docs_latest/bugfix_315_deployed_at_without_publication/HANDOFF_2026-08-20_continue_here.md`

It records what is done, the one substantive item left, and — the part I would most want a successor
to read — the nine traps this lane walked into for real, including the one where I broke every page
publish in the estate for thirty-three minutes by checking that the configuration was right and never
asking whether a page could still be published.

---

## 2026-08-21 — the last piece is built, and the fleet turns out to be healthy

The check is written. It does the obvious thing: for every page we have a fingerprint for, fetch the
live page, fingerprint what came back, and compare. If they disagree, raise a flag. That is the whole
idea, and until two days ago it was impossible, because we had nothing to compare against.

**First, the good news, and it is the main news.** Before writing a line of it I ran the comparison
by hand across every page that has a fingerprint — 228 pages across 12 sites. **All 228 matched.**
Every page we believe we published is, right now, serving exactly the bytes we sent. So the check
ships as a smoke alarm rather than a repair: there is no fire today.

That sweep quietly settled something else too. It is 228 independent confirmations that the
fingerprint we record and the bytes a visitor receives are the same thing — no compression, no
rewriting, no filename mismatch anywhere in between. One matching page could have been luck. Two
hundred and twenty-eight is the mechanism working.

**Then the part I want to flag, because it is the difference between a useful alarm and an annoying
one.** The obvious way to build this is to check a page the moment we publish it. I measured what
that would do: I watched every freshly-published page, every two minutes, for nearly three hours —
1,099 measurements over 85 pages. Three times a page genuinely did not match. All three were
seconds old, all three were completely healthy, and all three were serving correctly within about
two minutes. Publishing is not instant; it goes out in batches.

So an eager check would have raised **three false alarms in under three hours**, on the very sites we
care most about. The check therefore ignores anything published in the last thirty minutes. That is
not caution for its own sake — it is the difference between an alarm people trust and one they learn
to ignore. And it costs us nothing that matters: the fault this whole bug is about lasted **six
hours**, so we still catch that kind with hours to spare.

Two more would-be false alarms turned up in the same window: twice, a healthy page briefly returned
"not found" — the same error page both times, fine before and fine after. The check treats "the site
didn't answer properly" as a separate question it does not judge, which is already another check's
job. So those file nothing either.

**What I am less comfortable about, stated plainly.** The check trusts that every route that marks a
page as published also records its fingerprint. I checked: there are exactly three such routes today
and all three do. But that is a fact about our current configuration, not something the code
guarantees — if someone adds a fourth and forgets, this check would start accusing perfectly healthy
pages, and it would look convincing. There is a one-line fix that closes it properly. I have **not**
made it, because it reverses a decision that went through review two days ago, and slipping that into
an unrelated change is exactly the habit that got a previous change vetoed. It is written up as the
next decision for someone to take deliberately.

**And a confession about my own work, since this lane keeps finding these.** The test file claims
each safety guard is proven by deliberately breaking the code and watching a specific test fail. I
wrote that claim from the design and then actually ran it — and four of the nine claims were false.
Three tests were passing against broken code because a *different* safety net further down caught the
fault; one could never have failed at all, because it measured the limit it was testing against
itself. Fixed, re-run, and now true. Two other guards turned out not to be provable in isolation, so
the file says that instead of claiming a proof it does not have.

**Where this leaves us.** The check is committed and registered. It is not switched on: it needs the
next routine release, and then a one-line configuration change I have written and tested but
deliberately held back, because switching it on before the code ships would break the checks that are
already running. That ordering is the same trap that cost this lane thirty-three minutes of downtime
last time, so it is spelled out in the file itself, along with the instruction to check what you have
broken before checking whether it worked.

---

## 2026-08-21, evening — the new build carries the check, and the last obstacle got much smaller

The release went out and I checked, rather than assumed, that it actually contains the new check: the
running service reports which version of the code it was built from, that version includes everything
this lane wrote today, and the check's name is present inside the running program. I also ran the
checks backwards to make sure the test could have said "no" — twice today a test of mine would have
said "yes" no matter what, so that is now a habit rather than a flourish.

**The remaining obstacle turned out to be far smaller than I thought.** This morning I found three
publishing routes that mark a page as published without recording what they sent — the flaw that would
have made the new check accuse healthy pages. I described fixing them as risky, because they are part
of the main page-building path and the last time we touched that path we caused half an hour of
downtime.

Then I measured it, and the caution was theoretical: **all three routes have never run. Not recently —
never, in the entire recorded history.** Nothing schedules them and no work is routed to them. Exactly
one other part of the system is even able to call one of them, and it never has.

Finding that took one careful step worth mentioning, because the obvious check was wrong. Searching
the system's configuration for these three names finds four places — but three of those are simply the
names appearing in the *text of review prompts*, not actual wiring. Searching properly, for the names
used as instructions rather than as words, leaves one. That is the same mistake in miniature that
started today's whole correction: a text search answering a structural question and giving a plausible
wrong answer.

So the fix is now a small, reversible change to three dormant routes: harmless today, protective the
moment any of them is used. It is written, it has been tested three separate ways against the live
system without changing anything, and it is with the reviewers.

**What is left is one action, in one order.** Apply the small fix, then apply the setting that switches
the check on. If someone does them the wrong way round, the second one refuses — I built that refusal
in this morning, and watched it refuse, precisely so that nobody has to remember the order.

After that, this piece of work is finished, and the original bug can be closed.

---

## 2026-08-22, morning — the alarm went off, and it was a real fire

Overnight the new check ran 86 times across the estate. It raised **one** item, and that one is worth
the whole exercise.

**vetcomparison.uk's home page was serving the wrong version for at least nine hours.** We published
it at 20:49 last night. At 21:53 the check noticed the live page was not the page we had sent, and
raised a flag. It checked again at 01:54 and again at 05:56 — still wrong, and serving *exactly* the
same old bytes each time, so this was not a page caught mid-update. By the time I looked this morning
it had finally corrected itself, somewhere between 05:56 and 08:40 — nine to twelve hours after we
published it.

Nothing else in the platform noticed. The database said the page was deployed. The publish had
committed successfully. The job that did it completed without error. This is precisely the fault the
whole investigation started from — a page quietly serving yesterday's content while every internal
signal reads green — and it is the first time we have ever *seen* it happen rather than inferred it
afterwards.

**What that tells us, in order of importance.** First, the check works, on a real fault, unprompted.
Second, the fault is not a one-off from last week: it happened again last night, on a different site,
and it lasted longer than the original six-hour case that prompted all this. Third, and this is the
part I would not have believed without the evidence — **the page fixed itself**. We did not republish
it. The same publish eventually arrived, many hours late. So the delivery step is not losing pages
outright, it is delivering some of them extraordinarily slowly, which is a different problem with
different fixes.

**One correction to what I wrote earlier this morning.** I said the page had not been republished —
true when I looked, and it was republished ten minutes later, at 08:50, with identical content. The
nine-hour finding stands for the period it describes; the note just needed the boundary saying out
loud, because anyone checking now sees the later timestamp and would think I had it wrong.

**And then it went wrong again.** An hour after that republish, the live page was showing the *old*
version once more, before correcting itself by mid-morning. So this is not simply "delivery is slow" —
on this one page, in one morning, we saw new content fail to arrive for nine hours, arrive, and then be
replaced by the old content again for a while. Whatever is happening at the publishing step is less
reliable than "sometimes slow" suggests.

**What I have not done.** I have not chased *why* delivery took nine hours, because that happens
inside the publishing runner, which lives in a repository we do not control and cannot read. That was
already the known boundary. What has changed is that we can now hand whoever does have access an exact
case: this site, this page, published at this timestamp, still wrong nine hours later. Before today we
could not have named one.

**The one caveat, stated plainly.** The check is designed to clear its own flag once the page comes
right. That has not happened yet — the next scheduled pass for that site is late this morning, and I
have not watched that half of the mechanism work. Until I have, treat "it will clean up after itself"
as designed-but-unproven rather than a fact.
