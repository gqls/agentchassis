# Where we are — the unread finding codes (bug 358)

Plain-prose log for the owner. Append-only, newest at the bottom.

## 2026-08-22 — picked up, and the bug is still real

The bug, in a sentence: lots of parts of the platform, when they notice something wrong that
they won't fix themselves, write a note about it into a database table — and for sixteen
different kinds of note, nothing ever reads the note, and a cleanup job deletes it after
30 days. We pay to detect things and then throw the detection away unread.

Today I checked the bug is still valid before planning anything, because things move fast
here. It is. Two small things changed since it was filed this morning, and both are actually
good news: the other team working on content loss shipped their checker, and it is the
first thing EVER to use the table's "resolved" tick-box (45,000 rows and nobody had ever
ticked it before today). That team's checker is the model citizen: it writes its findings
AND reads them back AND acts on them. The sixteen orphaned kinds of note are all still
orphaned — and a seventeenth was added by another session this very morning, which is the
whole point of the bug: nothing stops a new one arriving without a reader.

Next: research how the one similar guard-rail we already have was built (the "optional key
budget" checker), then write the plan and put it past the council.

## 2026-08-22 (later) — the fix is built and running, and I broke the build on the way

**What the fix is, in plain terms.** Every one of those notes the platform writes to itself now
has to say, in one file, what it is FOR: something automatic reads it, or it is a deliberate
measurement with an owner and an end date, or it is there for a human to look at by hand and we
accept it will be deleted in 30 days, or it is ordinary error plumbing. A note that turns up in
the database without any of those is a failure that someone has to answer for. Adding a new kind
of note without saying which it is now trips the check the next morning.

**The important design decision, and it was not the obvious one.** There is a piece of code in
this system that describes itself as "the ONE writer" of these notes — so the obvious place for
the check was inside it. I counted, and it is not the one writer: there are five, one of which
*cannot* use it for a technical reason that will not go away. A check in the obvious place would
have covered one writer in five and looked complete. So the check asks the database what notes
have actually been written, which is the one question no amount of clever code-reading can get
wrong. It found 43 kinds; 32 of them nobody has yet decided about, and that number going down is
the measure of progress.

**Three things in the original bug report turned out to be wrong**, and I have corrected them in
place rather than quietly working around them. The biggest: the report's own headline example —
the loudest, most frequent note, 9,617 of them in five days — is described as something "nobody
has ruled on". Somebody had. It is a deliberate measurement with an owner, six dated readings and
a decision you made on 18 August. Filing it as waste would have been my first mistake. Another:
a note the report holds up as a *success* story turns out to have never been switched on at all —
the database change that would have installed it was never applied, so its "zero occurrences"
reads as "quiet" and actually means "absent".

**Now the bad part, and it is mine.** Committing my work broke the build for everybody for about
twenty minutes. Several of us edit the same files at the same time, and the safe way to commit is
to name your files explicitly — which I did, correctly. What I had not appreciated is that naming
a file takes whatever is in it *at that moment*, including another session's half-finished
sentence. I took two of them: a comment they were mid-way through typing, and a reference to a
piece of code they had written but not yet committed. Neither was visible to me, because on my
machine all the files existed and everything compiled. The version that was broken was the one I
never tested — the committed one.

I found it because an unrelated automated check complained in a way that did not match its own
description, and I read past the headline. I fixed it in two steps and verified the real thing
this time by extracting the committed version and building *that*. I chose to fix it by removing
my own premature reference rather than by committing their unfinished work, which would have put
a whole feature of theirs under my name — repairing a mistake should not make it bigger. I have
messaged that session directly, left a note where their code was so they cannot miss it, and
written the whole thing up in the shared log of mistakes, including the one-line check that would
have caught it in seconds.

**Where this sits.** The reviewers said "revise" on my first submission and they were right — I
had described two changes that depended on a piece I had not listed. That is exactly what the
review is for, it cost one resubmission, and the second round found a second omission of the same
kind. The revised version is with them now.

---

**2026-08-22, evening — the notes that were about to be deleted, and what they turned out to say.**

You made four calls this afternoon. Save the expiring evidence and find out whether the problem
it describes is fixed or whether the detector has gone deaf. Give deliberate findings a longer
life than ordinary error plumbing, and stop marking something "resolved" making it die sooner.
Let me propose what each of the thirty-two undecided notes is for, with the evidence attached, and
you ratify in batches. And leave the backlog count visible rather than enforced until we have seen
how hard a batch actually is.

The first one is done, and it did not come out the way I expected.

The forty-one notes are saved, in full, before Tuesday's deletion. They cover twenty-four review
rounds and twelve reviewer seats, and the great majority came from the fix-loop's own council
rather than the one that reviews platform changes.

Then the interesting part. These notes are filed under a name that says a reviewer's answer was
**cut off** — ran out of room mid-sentence. That is what the code believes about itself, and it is
wrong for forty of the forty-one. I went and looked at what the reviewers actually said, which we
keep word-for-word: in the whole retained history, only **five** reviewer answers have ever run out
of room, four of them in mid-July before this detector even existed, and one on 2 August. Nearly
twelve thousand answers end properly. So what the notes really record is "this reviewer's answer
came back unreadable" — and the reason was never established. The name asserted a cause nobody
measured.

On the question you actually asked — fixed, or deaf? **Fixed.** The same piece of code that writes
these notes also writes a short report on every single review round, and that report counts
unreadable reviewers whether or not there are any. Those reports are still being written — two
hundred and forty-eight of them last week — and the count of unreadable reviewers goes seventeen,
twenty-four, then zero, zero, zero. That is the difference between a thing that is quiet and a
thing that is broken: I can see the mechanism running and reporting nothing, rather than just
seeing nothing. If it had gone deaf, those reports would still be finding unreadable reviewers
while no notes were being written. They aren't.

What appears to have fixed it is that the reviewers' room to answer was doubled over the same few
weeks. The timing lines up closely and the reason is sensible — more room, complete answers,
readable results — but I could not find the change that did it, so I have written that down as a
strong hunch rather than a fact. There was also one reviewer seat on 2 August configured with
almost no room at all, which guaranteed it would fail; that has since been corrected and is not a
live problem.

One thing I got wrong on the way and want on the record. My first attempt to measure "did the
reviewers run out of room" compared two columns that are, it turns out, completely empty for these
records. It returned a clean zero every week and looked like a result. It was a question that could
not have produced any other answer. I only caught it because I checked whether the columns had any
data in them before believing what they said — which is the check I should have run first, not
second.

Next: the retention change, which is a single guarded database edit and goes to the reviewers
because it is live the moment it applies; then the first batch of proposed rulings for you.

---

**2026-08-22, ~18:50 — the retention change is live, and then the whole estate stopped talking to
the model.**

Your second ruling is done. Deliberate findings now live a year; ordinary error plumbing still goes
at thirty days; and marking something "resolved" no longer makes it die sooner. I tested it
end-to-end by putting four fake rows into the table inside a transaction I threw away afterwards —
one piece of plumbing and one finding, both thirty-one days old, plus a resolved finding twenty
days old. The two that should have gone, went. The two that the old rule would have destroyed,
survived. Then I rolled the whole thing back so nothing real was touched.

I found one thing wrong with my own work on the way, and it is the kind worth telling you about.
The migration carries a block of self-checks that are supposed to refuse if the change is wrong. I
tested those checks by deliberately breaking the change — and they let it through. Twice over: they
were comparing against their own private copy of the list rather than the list actually being
installed, and the sample they tested against had no rows old enough to matter, so the answer was
always the same regardless. I rewrote them to read the real thing, and now both kinds of breakage
are caught. I have written that up, because "the check passed" meant nothing until I tried to make
it fail.

**Now the thing you actually need to know, and it is not about this piece of work.** At 18:15 UTC
today, every part of the platform that talks to the AI model stopped being able to. The provider is
returning "you have reached your specified API usage limits, you will regain access on 1 September
at midnight UTC". Fifteen different parts of the system have hit it. The last call that worked was
at 18:15:51; there have been none since and twenty-two failures in the twenty minutes after.

This has happened five times in the last fortnight, and every previous time it cleared within the
same day — the tell is that work carried on around it. This one is different: the successes stop
dead fifteen seconds after the first refusal and do not resume. So I do not think this is the usual
blip, though I am reading that off the shape of it and the provider's own message rather than
anything I can verify from here.

What it means practically: no reviews, no diagnosis runs, no content being written, none of the
checking agents. Anything queued will fail at its first step. It is a billing or plan matter, so it
is yours rather than something I can fix.

It also means my change went live without its review. I submitted it before applying, it was
accepted, and then no reviewer could run. I have recorded that plainly rather than letting the
"submitted" note read as "approved", and it should be resubmitted once the cap clears.

---

**2026-08-22, ~19:05 — correcting myself on the outage, within the hour.**

I told you above that this looked worse than the usual interruption, because on the previous five
occasions the system kept working around it and this time it stopped dead. **That was wrong, and it
was wrong because I measured it a day at a time.** Looking day by day, all I could actually see was
"it was working again by the evening" — which I read as "it never really stopped". Hour by hour,
the previous times stopped just as dead as this one: on the 10th there were two solid hours with
not a single successful call before it came back, and on the 14th, one. Today is at exactly the
same point in that pattern.

We have a written record of this happening three times before. Every time the error named a reset
date weeks away, and every time it came back within one to three hours — because you raised the
limit. Our own notes say, in terms, not to repeat the mistake of writing "we are down for weeks"
into documents, because other people then read it as fact. That is precisely what I did, twice, in
this folder.

So the honest version: **the cap is hit; the useful action is to raise it; the precedent is hours,
not days.** I have no basis for the ten-day figure beyond the provider's message, and the record
says that message is the worst case rather than a forecast.

Another session spotted it and pointed me at the note. I re-did the measurement myself rather than
take their word, and they were right. I have logged the mistake in the shared log of wrong calls,
because the interesting part is not the outage — it is that I built a comparison at a resolution
too coarse to show the thing I was comparing, and then found it convincing.

---

**2026-08-23, evening — the check now has a clock, and you ratified the first batch.**

Short version: the thing we built yesterday only ran when somebody remembered to run it. That is
now fixed — it is packaged as a daily job at 07:30, and the image is built and pushed. It is not
deployed yet, because deploying is your whole-fleet release, not something I should do to one
service on its own.

There is a nice irony worth stating plainly, because it is the reason this mattered rather than
being tidy-up. The whole point of this piece of work is that the system writes down things it
notices and then nobody reads them. Our own check was in exactly that position: it noticed things,
and it only spoke when a person went and asked it. Putting it on a clock is what makes it a check
rather than a thing we could run.

**One real obstacle, and it is the interesting part.** The check has two questions it can only
answer by opening our own source code — "does the file you claim reads this code actually mention
it?". A scheduled job runs from a small image that contains the program and nothing else, so those
two questions would have failed every single morning, on a perfectly healthy setup. I measured it
rather than guessing: five failures, every day, for nothing.

I did not solve that by shipping our source code into the image. Those two questions compare two
things that both only change when somebody commits — so asking them again every morning cannot
possibly produce a different answer than it did when the image was built. What I did instead was
move them to the moment they can actually change: they now run when someone commits. And the daily
job says, in its own words, in every record it writes, which questions it did not ask and where
those questions are asked instead — because a check that quietly skips something looks exactly
like a check that passed.

While doing that I found the two questions **had no automatic runner at all**, which surprised me.
There is an existing commit-time check that compiles the very same code, but it runs only four
named tests, and the one that grades our registry is not among them. I proved that by running it
and listing what came out, rather than by reading the script and believing myself.

**Your batch 1 is applied.** All seven codes are now recorded as human-evidence, each with the
reason and the retention window it accepts written into the file. The undecided backlog went from
32 to 25, and I lowered the cap to 25 in the same commit — otherwise the ground we just gained
gets quietly given back. Two of the seven carry corrections that travel with them: one is read
automatically but from a *different* record, so its row here is still unread; and one is misnamed
for what its own rows actually contain.

**One thing went wrong and it went wrong in the right direction.** The image build refused,
because a file listing what may travel into images did not include our registry. The comment in
that file predicts this exact mistake and says the loud failure is preferable to the quiet one. It
is: a build that refuses cannot ship a check that silently cannot run — which is precisely what
happened to another team's check the day before, deployed and green-looking and dead.

**The reviewers sent it back once, and they were right to.** The report the daily job writes names
the commit-time script as the place those two questions now get asked. I had described that script
in prose rather than listing it as part of the change — so from the reviewer's side, the job was
pointing at something they could not confirm existed. It does exist and shipped in the same commit,
but they could not see that, and "points at something that may not be real" is exactly what that
reviewer is there to catch. Resubmitted with the script shown in full.

**What I need from you:** the deploy. The image `finding-code-registry-check:v1.0.1331` is built
and pushed; it needs the fleet release to actually start running:

    date; make release redeploy-agents ENVIRONMENT=production REGION=uk001; date

After that I will trigger one run by hand rather than waiting for the morning, and check the job
actually wrote its row — a missing row has to mean "it did not run", never "nothing was wrong".

---

**2026-08-23, later — approved, and a separate thing I found on the way that matters more.**

The review passed on the fourth attempt. Worth being straight about why it took four: the
reviewers never found anything wrong with the code. Every time they sent it back, it was because
my write-up claimed something that was true but that they could not check from what I had shown
them — I kept leaving files out of the submission on the grounds that they were "out of scope for
review", which meant the reviewers were being asked to take my word for the part that mattered.
That is a fair thing to refuse. I have written it up as a mistake of mine, because it happened
three times and I only fixed the specific instance each time instead of the habit.

One of their comments was worth more than the review. A reviewer said I had bumped a
version number that the whole fleet shares without showing evidence that this was safe. They were
right that I had not shown it, so I measured it — and it turned up a genuine trap nobody had
written down. **Deploying does not use the version each service records; it rewrites every
service to the number in the makefile and then deploys.** So in the window between adding a new
service and the next full release, exactly one image exists at the new number and thirty-two do
not. Anyone taking the shortcut of deploying without building first would have quietly broken
thirty-two services — and the way that failure shows up, the cluster reports them as *running*.
The full release command builds everything first, which is exactly why it is the procedure. Now
written down.

**And one real defect, found by accident.** After the review passed I went to check that the
system had credited my work as reviewed — and it had not listed it at all. Not "unreviewed":
absent. It turns out the report that tracks which changes have been reviewed keeps its own
private list of which folders to look in, and when you widened the review scope this morning to
include the folder holding our daily checks, that private list was not updated with it. So for
the last fortnight, **twenty-two changes across four different workstreams have been invisible to
that report** — including other people's, not just mine. The report was printing a confident
total that meant "of the ones I can see". Fixed in both places, with a note left where the next
person changing the scope will actually be standing.

**Still outstanding, and it is the same ask as before:** the deploy. The image is built and
pushed; the job will not run until the fleet release goes:

    date; make release redeploy-agents ENVIRONMENT=production REGION=uk001; date

Please run the whole thing rather than a deploy on its own — that is precisely the trap above.

---

**2026-08-24 — it is running, it found something on day one, and what it found was partly my own
sloppiness.**

The daily check is live. I did not wait for its 7:30 slot — I ran it by hand, which is the whole
point of the "never let the schedule tell you it works" rule, and it is why we know all this today
instead of tomorrow.

**It failed on its first run, and it was right to.** A brand-new error code had appeared in the
system about two hours earlier with nobody having declared what it is for. That is exactly the
thing this whole piece of work exists to notice, so the failure is the mechanism working, not a
fault.

**And what it caught is not paperwork.** That new code was recording something real: twice today,
the part of the system that writes page content could not load the list of pages it is allowed to
link to — a database timeout — so it was told to write those pages **with no internal links at
all**. Two pages went out degraded, the system dutifully wrote down that it had happened, and
nothing anywhere reads that note. That belongs to another workstream, but it is the clearest
example we have of why this matters: the record existed, and it was going to expire unread in
thirty days.

A second new code turned up within the hour. Its own comment says it is "a fourth code beside" three
existing ones and copies their shape — which is precisely the spreading pattern the original bug
report described: the habit of writing a finding down propagates, and the habit of not reading it
propagates with it.

**Now the part that is on me.** Chasing how those codes got in without being declared, I found that
a safeguard I have been citing for two days **does not exist**. A comment in our own code says a
check runs "at commit time" and names the file it lives in. That file was never written. The
sentence was composed while I was planning to write it, and after that it read like a description of
something real — it went through a review, into the concept register, and into the handoff for this
work, and nobody, me included, ever checked that the file was there. The check that *was* running
instead was a hand-typed list of eleven names, which can only catch mistakes somebody remembered to
add to it.

It is written now, and properly: it reads the code and works out for itself which error codes exist,
rather than being told. I have also gone back and marked the false claim as false where it was made,
and logged it in the shared record of wrong calls, because "a comment naming a file is a claim" is a
lesson worth more than the fix.

**One thing I nearly got wrong and did not.** Having found the hand-typed list, my instinct was to
delete it as redundant. I checked it entry by entry first — and one of the eleven is written in a
way the new automatic check genuinely cannot see. Deleting the list would have quietly dropped that
one while looking like tidying up. It is kept, cut from eleven entries to one, and there is now a
test that complains if it ever grows back.

**Nothing needs deciding by you right now.** The two new codes are recorded as "human evidence,
nothing reads them" — which is measurably true — and you can overrule either. The review of the new
check is with the reviewers as I write this.

---

**2026-08-24, evening — the loop is closed, and here is what your decisions did.**

The fresh build you rolled out carries yesterday's declarations, and I proved it rather than
assumed it: the running job names the exact commit it was built from, your two new codes are in
that commit's history, and a hand-triggered run in the cluster came back clean — zero findings,
fifty-five codes declared. So the full story of the first thing this check ever caught is: it went
red within two hours of a new code appearing, the code got declared the same day, your next
release carried the declaration, and the check went green. That is the whole design working end to
end, on its first real case.

Since you asked me to explain your decisions — here is what each one actually did, because the
effects are worth seeing joined up:

**You chose to ratify in batches rather than delegate the rulings.** Right call, and not as a
formality: a disposition is a decision about what the business accepts losing — "nothing reads
this and we accept that" — and a session can measure what is true but cannot accept a loss on
your behalf. The mechanics honour that split: I propose with the evidence attached, you say yes
or no, one commit applies it.

**You said "cap it".** I built that as a one-way gate rather than a fixed ceiling: the undecided
pile can never grow, and every batch you ratify locks the new, lower number in the same commit.
A fixed ceiling below the existing pile would have made the check fail every day from day one,
and a check that is always red is a check everyone learns to ignore — that is this project's
founding observation, so the cap could not be allowed to recreate it.

**You said "Phase 2 first, then batch 1".** That ordering is the reason we caught anything at
all: the clock went live a day earlier than it otherwise would have, and it was the clock — not
any ruling — that noticed two brand-new codes arriving unread. If we had spent that day ruling
the backlog instead, both would have sat invisible.

**You ratified batch 1 in full.** Seven codes now honestly say "a human wrote this deliberately
and nothing reads it", each with its reason and the retention window it accepts written down.
The undecided pile went from thirty-two to twenty-five.

**You approved the sink rule.** "Consumed" now has to name the table the reader actually reads.
Without that, one code would today be wearing a green "consumed" badge over a row nothing has
ever read — the exact defect this whole effort exists to end, dressed in its own compliance.

**And your standing rules did quiet work throughout.** Your whole-fleet release rule is why the
deploy waited for you — and it turns out to be load-bearing: deploying one service alone in that
window would have silently broken thirty-two others. Your rule that new switches ship OFF by
default is why the one new flag here cannot change anyone else's behaviour. And your widening of
review scope last week is the only reason any of this code was reviewable — and following that
thread found the coverage report had been blind to the whole widened area, which is now fixed.

**Still open, both yours, no urgency:** the remaining twenty-five undecided codes (batch at a
time, as before), and the thirteen codes that are written in the source but have never fired —
same treatment when you want it. A second review pass by another model is running at your
request; if it finds anything real I will fix it and say so.
