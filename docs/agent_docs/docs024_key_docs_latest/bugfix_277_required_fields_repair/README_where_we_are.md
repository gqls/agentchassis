# README — where we are (plain prose, append-only, newest at the bottom)

## 2026-08-15 — the repair handler you asked for this morning is built

This morning you ruled that `required_fields_missing` — the "this page is missing required
content fields" finding — should get a repair handler fleet-wide instead of piling up in the
human-review queue. This session built it.

Before building, we measured what the 44 open findings actually are, and that changed the
design. Most of them (35) are NOT pages missing content — they are pages that serve perfectly
well today, but whose content lives as one stored block of HTML rather than as structured
fields. Automatically "repairing" those would regenerate the section from a template and
throw away the served page — the exact accident we've had before. Six more point at
components that no longer exist. Only a handful are genuinely repairable by a rebuild: one
component with fields that are truly empty, one generic page with no section plan, and your
gas unit converter (a tool page, which the platform deliberately refuses to rebuild with the
generic builder, because that clobbers tools).

So the handler is a router, on the same pattern as the image routers you asked for on the
12th. For each finding it asks the database what is true NOW and takes one of four actions:
if the finding is out of date, it closes it with the evidence written on it; if the fields
are genuinely empty, it files a targeted rewrite that edits the existing copy rather than
replacing it; if the page has no plan and is safe to rebuild, it files that rebuild; and for
the two classes that genuinely need a human (the stored-HTML pages, and tool pages like the
gas converter), it parks the finding back in your review queue — but now carrying its
classification, the evidence, and the safe options, instead of sitting there as a bare
mystery. Parked findings are pinned so the system cannot keep re-raising duplicates of them.

State right now: the classification was dry-run over all 44 findings and every prediction
checked out; the change went to the council for review; the handler is written and ready to
seed. Next steps are: seed it (inert), run four representative findings through it as a
canary, then point the remaining 40 at it. The gas converter itself will come back to you as
a parked decision naming the tool pipeline as the repair route — this handler routes it
honestly rather than overriding the no-generic-rebuild-of-tools rule.

## 2026-08-15 evening — built, live, backlog routed; two decisions need you

The repair handler is done and working. Every one of the 44 parked findings has now been
through it or is queued for it: the dead ones are being closed with the evidence written on
them, and the ones a machine must not touch (the stored-HTML pages, the tool pages, the
image-slot fields) are parked back in your review queue carrying their classification, the
danger, and the safe options — instead of sitting there as bare mysteries. New findings of
this type route themselves from birth now (that change rode another lane's release this
afternoon). Along the way the review process caught two genuine design errors before they
could do harm — a repair that would have asked a prose writer to invent image addresses
(the system refused it, and we made that refusal a routing rule), and a page rebuild that
would have quietly produced nothing (measured, and made a routing rule too).

The council reviewed this four times and improved it each time, but it has not approved it,
and the remaining objections are ones the reviewers themselves disagree about — which by our
own rules means they are yours to settle, not mine to argue again:

1. **One reviewer insists new findings should be born "unclaimable" and promoted by the
   triage stage; another accepts what we did.** The trouble is the triage stage's machinery
   has been switched off since May (bug 083) — honouring the contract today means findings
   sit stranded forever, which is what you ruled against. If you want the contract honoured,
   the real fix is rebuilding the promoter, and that is a separate piece of work. My change
   is one line to revert the day that happens.
2. **This is the third single-purpose router of its kind** (the two image routers from the
   12th are its siblings). Several reviewers want one shared engine instead of a growing
   family. I filed that as RFC_030 with the case for doing it as a consolidation of all
   three; it needs a decision on whether and when to schedule it.

---

## 2026-08-17, late morning — the promoter was letting one broken handler waste work, and we caught it by not believing a graph

The fresh build you deployed is `v1.0.1305`. I checked it properly rather than trusting the
version number: the image itself records which commit it was built from, and I confirmed that
same commit is inside the running program, using a second commit that had to come back *absent*
as a control. It carries the change this lane was waiting on. Nothing here was blocked on it.

Then I re-measured everything before acting on it, and two of the numbers turned out to mean
the opposite of what they looked like.

**The queue looked like it had refilled.** The count of unclaimed findings had gone from 4 to
82 overnight, which reads like the new promoter had stalled. It hadn't. Seventy-seven of those
82 are a kind of finding that deliberately has no handler attached — they are flags for a human
to read, and sitting in that queue is where they are *supposed* to live. Forty of them were put
back there on purpose this morning by another session working a neighbouring bug. Once you
count only the findings that actually have a handler, the number is five, and all five are the
promoter correctly refusing to act until a person has run one by hand. So the mechanism is fine
— but the *measurement* we had written down for judging it was not, because it lumped together
two groups that mean opposite things. I've rewritten that test in the bug file so the next
person isn't misled the way I nearly was.

**The promoter was, however, doing something genuinely wrong.** When we built it we wrote down a
risk: it decides a handler is trustworthy if that handler has ever succeeded at that kind of
job even once. We flagged that as thin at the time. It has now bitten. One handler had
succeeded once and failed twenty-eight times, and the promoter — seeing the one success — kept
handing it more work: six items, five of which failed. That is not a disaster, but it is
wasted machine time and a queue full of failures that look like new problems.

I've closed that door. A handler now has to be succeeding at least a quarter of the time before
the promoter will feed it, and only once it has a real track record — so a brand-new handler
still gets its careful first run, exactly as before. I did not pick the "quarter" out of the
air: I listed every handler's success rate, and there is a clean gap in the data — one handler
at 3%, and the next worst at 41%. Anywhere between those two isolates the broken one and
touches nothing else. As it stands the new rule blocks nothing at all today; it is a door, not
a repair.

**The part worth telling you about is how nearly we missed it.** My first attempt to measure
handler reliability said every single handler had a 100% success rate. That is obviously too
good, which is the only reason I looked again. The cause was a genuine trap in the database:
failed jobs never record a "finished at" time, only successful ones do — so a query that asks
"how did this handler do before now?" using that timestamp silently counts *zero failures for
everybody*. It doesn't error, it doesn't come back empty; it comes back as a clean, uniformly
excellent table. I've written that up as a warning other sessions will see automatically before
they touch the same column, because it would flatter any handler-reliability report we ever run.

**The other thing I want to correct on the record.** One of the three tests we set for declaring
this bug fixed was already passed six days before we built the fix — and both of the last two
updates said it was still pending, because nobody re-checked it, they just copied it forward. If
I'd re-measured a day later I would probably have written it up as proof our fix worked. It
isn't; it was the *old* mechanism, run by hand back in July. The fix is still well-evidenced —
I verified four repaired pages by actually loading them in a browser and reading them, and in
every case the page was written seconds *before* the job was marked done, which is the
fingerprint of real work — but that particular test proves nothing about it and I've said so in
the bug file rather than quietly banking it.

(Small thing, same theme: my first attempt to load those four pages returned "not found" on all
four, which looked briefly like the repairs had never reached the live site. It was my own
mistake in how I typed the addresses. Loading the site's front page first — which worked
fine — is what told me the problem was my question, not the site.)

**Where that leaves things.** The repair router and the promoter are both live and behaving.
The promoter now has both door-closers the review round asked for, plus the one it needed and
nobody asked for. Next: the review round itself, which I'm submitting now with all of this as
evidence.

---

## 2026-08-17, early afternoon — approved; and two things I want your decision on rather than mine

The review round came back **approved** — twelve of the reviewers in favour, including the two
that had objected most strongly last time. Two of them left non-blocking notes, and I checked
both rather than filing the approval and moving on, because one of them was a good question.

That reviewer's point was: *your new rule decides a handler is trustworthy by counting how often
it has succeeded — but this platform has a documented case of a job reporting success while
actually deploying nothing useful. If the one success your rule is built on was fake, the rule
is built on sand.* That is exactly the right question. So I opened the page that one success
claims to have repaired and read it: clean, no defect. Then — and this is the part that matters
— I ran the same test on three pages where the same handler *failed*, to prove my test can
actually detect the problem it was looking for. It found the defect on all three. So the success
is real, the failures are real, and the rule is sound. A test that finds nothing is worthless
until you have shown it can find something.

**The good news on the original bug.** Another team's mechanism — built for a different problem
entirely, knowing nothing about ours — has started re-checking the findings we parked. It agrees
with us row for row: on the twenty-nine pages we said "a machine must not touch this", it
independently declines to judge them, and gives the same reason we did. That matters more than
anything I could measure myself, because the judgement it confirms is the one this whole piece of
work rests on and the one you'd be right to doubt. It could easily have disagreed. The review
queue for this finding type now contains exactly **one** genuinely live item, where it used to
contain forty-four indistinguishable ones.

### Decision 1 — a part of the repair machine has never actually run

One reviewer flagged that our repair router has two branches that rebuild page content, and that
rebuilding is a known way to silently lose content on this platform. I have now measured
something nobody had: **those two branches have never once fired.** Every finding the router has
ever handled went down a "park it for a human" or "close it, it's already fixed" branch. The two
occasions it tried to convert something, the conversion was cancelled.

So the risk isn't disproven — it's untested. And the first time those branches do run, it will be
on a live customer page, unrehearsed.

We have a house rule for exactly this shape: a risky new capability ships switched off, and
someone turns it on deliberately the first time. Our own promoter works that way — it refuses to
hand work to a handler type nobody has watched at least once. I did **not** apply that to the
router myself, because it changes the safety posture of something you've already signed off, and
there are two sensible ways to do it. **Your call:** leave it as is, or make those two branches
park-for-a-human until someone has run one by hand and watched it.

### Decision 2 — five findings are stuck waiting for a human, and nobody is that human

The promoter deliberately holds back any finding type nobody has ever supervised. Right now five
findings are held that way — four since 10 August. That is the design working, but there's a
gap: nothing prompts anyone to go and do the first supervised run, so a held finding can sit
indefinitely. That is a smaller version of the original bug we set out to fix. It needs either an
owner for those types, or a rule that says how long something may sit held before it's escalated.

### One small thing worth knowing

Our review system refuses to look at changes that are "just configuration" — it only reviews Go
code. But on this platform a great deal of real behaviour lives in configuration, including the
whole mechanism reviewed today. It only got reviewed the first time because it happened to have
one Go file attached. I overrode the refusal deliberately and said so, but it's a real blind spot
in the review process rather than a one-off inconvenience.

---

## 2026-08-17, afternoon — you were right to push the review round, and I've logged why I was wrong

You told me to run the fifth review round anyway. **It passed** — approved by every reviewer that
had previously blocked it, including the two that objected most strongly and the one that killed
the fourth attempt.

My reasoning was half right and half lazy, and it's worth being precise about which half. The half
that held up: a round that merely *repeats your rulings* would have been thrown out, because the
fourth attempt died on exactly that — listing work already done as though it were new. So I
structured this one around a single genuine change: the code fix that answers the reviewers'
strongest objection was committed **three and a half hours after** the fourth verdict was written,
so no round had ever actually looked at it. Your rulings became the *grounding* for that change
rather than the submission itself.

The half I got wrong is the part worth recording. I concluded there was nothing new to submit
without checking whether anything had changed since the last verdict — and I'd had that very commit
open earlier the same day for another reason. "There's nothing new to say" is a claim about the
world, and I stated it with no evidence behind it while being careful about evidence everywhere
else that morning. One command would have settled it. It's in the wrong-calls log with the check
attached, because that particular mistake — reasoning about the argument instead of looking at what
changed — is the kind that repeats.

**One reviewer earned its fee.** It couldn't see the promoter's configuration from where it sat, so
it asked: *how do you know the promoter actually handles this type of finding? If it doesn't,
switching back to the old behaviour silently recreates the exact bug you were fixing.* Fair, and I
checked it four ways. It does. But checking it surfaced something about **my own** change from
earlier today: the safety rule I added this morning only lets through findings tagged with one of
three categories, and if this lane's findings had carried a different tag — two other real ones
exist — **my own safety rule would have quietly stranded the very findings this project exists to
handle.** They carry the right tag. It's fine. But it was fine by luck until I measured it, and
that's the sort of thing that only ever gets caught by someone asking an awkward question.

The clean proof, for the record: one finding has been created since the fix went live. It appeared
at 10:02, was picked up 42 minutes later, correctly re-tagged, routed, and parked with its
explanation attached. Every hop worked.

**Still waiting on you:** the two decisions from this morning — whether the router's never-yet-used
rebuild branches should be switched off until someone has watched one run, and who is responsible
for the five findings held awaiting a first supervised run. On the first, a **second** reviewer has
now independently said those branches want a fail-loud guard, so that decision has more weight
behind it than when I raised it.

---

## 2026-08-17, late afternoon — the build you deployed didn't actually ship, and I got something wrong that I want to flag properly

**First, the thing you need to know: this afternoon's chassis deployment shipped no new code.** It
was rebuilt under the *same version number* as this morning's build, and when that happens the
server quietly reuses the copy it already had. The pods restarted and look perfectly healthy. They
are running this morning's code.

I checked this three separate ways before saying it: the image on disk is genuinely new (built at
14:30), the program actually running contains this morning's fingerprint and not the new one, and
the pods are running a different image file from the one that was built. The restarts happened
*after* the build, so it isn't a timing coincidence.

**What that costs:** 252 commits are sitting undeployed, 26 of which change actual program code. So
26 pieces of work whose authors reasonably believe are live are not. Nothing is broken — it simply
hasn't shipped. The fix is to bump the version number in the makefile and run a release; that's a
whole-fleet action you run, not me. I've recorded the incident against the existing bug that owns
this trap, including one honest correction to that bug's proposed fixes: one of the three fixes on
its list would *not* have caught today's case, and I've said so rather than let it look better
supported than it is.

None of my own work today is affected — it's all database configuration, which takes effect the
moment it's committed rather than needing a build.

**Second: the time-limit mechanism you asked for is working.** It ran at 12:57 and escalated the
four findings that had been waiting seven days, while correctly leaving the one-day-old finding
alone. That's the clock actually discriminating rather than just emptying a backlog. Each escalated
finding now carries how long it waited, what to do about it, and — for the unowned type — the words
"UNASSIGNED, claim this" along with the evidence that nobody has touched that check since July.

**Third, and I'd rather you heard this from me: the safety rule I added this morning was wrong, and
I've fixed it.** I built it to hold back any handler succeeding less than a quarter of the time. It
turns out this system records success in **two** different ways — a job can be marked "done", and
then later marked "verified" once it's been checked. I was only counting the first. So every time
the platform *verified* a piece of work, that handler's apparent success rate went **down**. I had
built a scorecard that gets worse the better we do.

I caught it by accident: I looked at a handler that read 46% this morning and 20% this afternoon,
which is impossible unless something is wrong. Nothing had gone wrong — nine of its successes had
simply been verified in between. Counting both, it's 50%, and my rule had been about to block it
for no reason.

Two things worth saying about that. The damage was small and I've fixed it while it was small — that
second status is currently rare, nine records in total. But the version of this that would have hurt
is the one that hadn't happened yet: a handler whose successes had *all* been verified would have
looked like it had never worked at all, and been blocked permanently — which is precisely the
original bug this whole project exists to fix.

And: **the review panel could not have caught this.** Twelve reviewers approved that rule this
morning. One of them came close — it asked whether "done" could really be trusted — and my answer
to it was correct but aimed at the wrong question. A review reads the plan; it can't check a detail
the plan never mentions. That one needed a query, not a reviewer. It's logged as a mistake of mine,
alongside a near-identical one earlier the same day, because twice in one session is a pattern
rather than bad luck: both times I filtered on a status column without first asking what values that
column can hold.

---

## 2026-08-17, evening — somebody finally ran the first one by hand, and it worked

A fresh session picked this up. Three things happened worth telling you about, and one of them is a
mistake of mine caught before it did any harm.

**First, a housekeeping point about your own instructions.** You told me to gate the two
content-rebuilding branches of the repair router. You'd told the earlier session the opposite —
leave them alone — about three and a half hours before, and that session had already written your
decision into a commit so nobody would reopen it. Rather than quietly reverse you, I put both
answers side by side with their timestamps and explained the trade-off again, including the argument
*against* gating that I hadn't given you the first time: a safety gate that nothing ever reaches is
a mechanism that rots unused, which you've ruled against before. You confirmed: leave them alone.
Nothing changed. I mention it because with this many sessions running, the same question can reach
you twice in an afternoon wearing different clothes, and the honest thing is to say so rather than
act on whichever answer came last.

**Second, the actual work: the first supervised run of a repair type that had never once run.**
The promoter deliberately refuses to hand work to a handler nobody has ever watched, until a person
runs one and watches it. Yesterday we built the clock that nags when something has been waiting too
long. Nobody had yet been the person. Now someone has.

The type in question had twenty findings, none of which had *ever* been dispatched, the oldest
waiting a week. I promoted exactly one by hand and watched it: it was picked up 64 seconds later,
the thing it was complaining about was repaired, and the job closed — with the repair landing
eleven seconds *before* the job was marked done, which is the fingerprint of real work rather than a
job that just says it's finished.

Then the interesting part happened on its own. That one success is all the promoter needed to start
trusting the type, so on its next scheduled run, four minutes later, **it picked up the remaining
findings and repaired them without anyone asking**. That is the whole point of this piece of work,
end to end, for the first time: a person watches one, the machine takes the rest. I checked the
live page before and afterwards — identical, to the byte, with all its text intact. This kind of
repair fixes a status, and it did not touch anyone's content.

**Third, and this is the one I want you to notice: I nearly ran the wrong one.** Before promoting
anything I checked whether the four findings were still true. One wasn't. It was complaining about a
piece of a page that no longer exists — the page had been rebuilt five days after the complaint was
filed, and the replacement was already fine. Had I promoted *that* one, it would have failed, for a
reason that has nothing to do with whether the repair machinery works. And because we added a rule
yesterday that stops trusting a handler once it starts failing, **that failure would have been
scored against the very thing the run was meant to prove.** The exercise designed to earn trust
would have destroyed it instead.

That's now written down in three places so the next person doesn't step on it, and the underlying
cause is filed as its own bug: this kind of finding remembers a page's part by an identifier that
the platform already knows is not stable when a page is rebuilt. There's a rule in our own debugging
guide saying don't do that, complete with the fix. This code is on the wrong side of it.

**A separate thing that turned out to be worse than advertised.** The review round left one open
complaint: when the promoter refuses to hand out work, it does so silently. The suggested fix was to
have it report a count. I went to add the count and found it would have gone nowhere — the
scheduler throws away everything these background jobs report. So for the last two days the
promoter has been writing a little summary of what it did on every run, and nothing has ever read
it. Every run has logged one identical line whether it moved twenty items or none.

So I fixed it at the level where it was actually broken — one change that makes *every* job of this
kind report what it did, not just this one — plus the counting the reviewer asked for. It's
committed and goes live at the next rebuild of that service, and I've said clearly in the bug file
that it is **not** verified until I can see a real run's numbers in the log, rather than assuming
the commit was enough.

Worth knowing what it says on its first run: it is holding two things back, and one of them is
yesterday's new safety rule doing its job on a live item for the first time. Yesterday that rule was
recorded as "blocking nothing today — it's a door, not a repair". It's now load-bearing, and until
today nobody could have known.

**My mistake, logged.** Having found one stale finding, I went looking for the general pattern,
intending to build a guard. I measured it with what looked like a sensible test — do findings whose
page was rebuilt since filing fail more often? — and the answer came back saying the stale ones
succeed *twice as often*, which is nonsense. The reason is that the column I was measuring gets
written by the repair itself, so I was using the outcome to predict the outcome. The measurement
couldn't have told me anything. What saved me is that it came out absurd; had it come out merely
plausible, I'd have built a guard on it and written the number into a bug file. I've logged it, and
what shipped instead is the small, directly-checkable finding rather than the grand one.

**Where it leaves things.** The promoter is doing what it was built to do, and now has evidence
rather than argument behind that claim. Bug 083 stays open on two narrow points: the logging change
needs the next rebuild before it's real, and yesterday's safety doors should sit a week before
anyone says they're behaving. One caveat inherited from earlier today, which I've flagged rather
than papered over: another session found that completed work records have been disappearing from
the database and nobody yet knows why. All of this counting reads that same table, so if records
can vanish, the promoter can decide a handler "has never succeeded" when it has. That needs the
diagnosis run someone has already recommended.

---

## 2026-08-18 — the logging change went live, and the first thing it did was find fourteen problems we could not see

The fresh build carried the scheduler change. I checked it at the service rather than trusting the
deploy: the pod states which commit built it, and ours is in it. (One honest wobble: my first sanity
check was meaningless. I checked that yesterday's last commit was *absent* from the build — it
wasn't, because the build is newer than everything I did yesterday, so nothing of mine could have
been absent. A test that cannot fail isn't a test. Redone properly against a commit made after the
build, it behaved.)

**What the change was for, and what it found.** Until today, this background job wrote a short
summary of what it did on every run, and the scheduler threw it away unread. Every run logged the
same sentence whether it moved twenty items or none. It now reports itself — and a job with nothing
to do now says so, which is a different sentence from a job that did something. Those two cases were
indistinguishable before.

**Then it told us something we did not know.** Yesterday I measured, by hand, that the promoter was
holding back two findings. Today its own log says **sixteen**, across five kinds of repair. That is
an eight-fold increase overnight, and it was completely invisible the whole time. The mechanism is
behaving exactly as designed — it refuses to hand work to a repair type nobody has supervised — but
"the machine is fine" and "sixteen real complaints about live pages are parked" are both true at
once, and only the second one is a problem you can now see.

Ten of the sixteen are one kind: a page-formatting defect whose repair handler succeeds about one
time in nine. Yesterday's new safety rule is deliberately holding those back rather than feeding a
handler that mostly fails — which is right, and it is also why those ten pages stay broken. That
handler is already someone else's open bug.

**One thing has a clock on it, which is why I am flagging rather than filing it.** Yesterday we added
a rule that nags when something has been held too long: after three days it moves the finding into
the queue a human reads. Three of these sixteen cross that line **tomorrow**. The problem is that the
move is one-way — the promoter can never pick those rows up again, even if we later prove the repair
type works. So tomorrow, three real findings quietly leave the automated path for a human queue with
829 items in it. It is a small version of the exact bug this whole piece of work was about, and the
fix is not mine to choose because the three-day rule was built on purpose by another session.

**On the original bug: it is done.** Everything 083 was filed for has been answered and checked at
the artefact rather than asserted. What is left in that file is not its own machinery but the
questions the new visibility raised, each of which has a named owner or needs a decision from you.

---

## 2026-08-18, evening — I found what was eating the history, and it was worse than I'd filed it

Yesterday I told you some completed work records were vanishing and I couldn't find what was doing
it. I found it, and the honest summary is that one query would have answered it and my three
searches were each looking in a place it could not have been.

**There is a housekeeping job that moves finished work older than a week into an archive table.**
It's been running daily, it's supposed to be there, and nothing is broken about it. The problem is
that the safety rules I built read only the *live* table — and the archive turns out to hold more
than twice as much history as the live table does. So when my rules asked "has this handler ever
succeeded?", they were really asking "has it succeeded in the last seven days?"

**That was already doing damage, not waiting to.** One handler was being blocked as "never
succeeded" when it had in fact succeeded nine times out of fourteen — the record was simply older
than a week. Another, with 316 successes behind it, looked marginal. Both are fixed: the rules now
read the archive as well, and I watched the blocked one get picked up and start work within seconds
of the change going in.

**Why I want to flag how I missed it.** The job moves records rather than deleting them, so
searching for deletions found nothing. Its configuration lives somewhere other than where I was
looking. Its description says "archives", and I was searching for "retention". And the reassurance
I gave myself — "the oldest surviving record is from March, so nothing is sweeping the table" — was
worthless, because that record is one the job would never touch. Four checks, all clean, none of
them capable of finding it. What actually found it was the job's name appearing in a list in someone
else's file that I happened to be reading for another reason.

**I also corrected three faults in the time-limit mechanism I built for you yesterday**, one of
which another session found and left for me. The important one: it had drifted out of step with the
rule it describes, so it would have asked a person to go and hand-test a handler that was already
working fine. Both now use one shared definition rather than two copies that agree until they don't.

**One pattern is worth naming, because it's now happened five times in this piece of work and always
the same way:** every mistake was me measuring a population I hadn't checked the boundaries of.
Not sloppy arithmetic — well-formed queries answering a slightly different question than the one I
meant. Three of them were caught only because a number moved when it shouldn't have. Twelve
reviewers approved the rule that had the first one in it; a review reads the plan, and none of these
was visible in a plan.

**Still with you or ahead of us:** nothing needs a decision from you right now. The queue mechanism
should escalate its first overdue items around the 19th and 20th, which will be the real test of it.
The bigger remaining piece is the shared router engine you ruled on — its groundwork is done and its
design round is the next real job.

## 2026-08-18, evening — the escalation dates I gave you are wrong, and the rule that stops bad handlers is judging them on the wrong evidence

Two things this session, and the first is a correction to what I told you a couple of hours ago.

**The dates I gave you for the queue escalation were both a day early.** I said the first overdue
items would surface around the 19th and 20th. They will actually be the 20th and 21st. The reason is
small and worth knowing, because it will happen again: the job runs **once a day, at 12:57**, and it
asks "has this been waiting more than three days?" at that moment and no other. An item that arrived
at 19:17 in the evening is still six hours short of three days when the job looks at lunchtime, so it
waits another full day. A "three-day limit" on a once-a-day job is really three-to-four days. I had
worked the dates out on a calendar, which throws away the time of day.

**What makes this worth a paragraph rather than a footnote:** tomorrow the job will run and escalate
**nothing**, and that is the correct behaviour. But it looks identical to the mechanism being broken
on its first run after I changed it. I had written the wrong dates into the handover notes as
instructions, so the next person to look would have seen a zero on the day I told them to expect
action, and reasonably concluded the change had failed. I have corrected the notes and written the
trap up where sessions read it.

**The second thing is more substantial, and it is a fault in my own design.** The rule I built stops
sending work to a handler that keeps failing — below a one-in-four success rate, it stops. I have now
gone and looked at *why* those failures happened, across every failure on record rather than the last
week's. **Only about one in six is the handler actually failing.** Nearly half are the handler
**correctly refusing** — declining to overwrite a page that a tool owns, which is exactly what we
want it to do. A quarter are infrastructure: timeouts, a pod dying, a message not delivered. And
about one in eight were never an attempt at all — old records tidied up by hand or backfilled by a
previous fix. My rule counts all of them identically as "this handler failed".

**Nothing is currently being blocked wrongly** — I checked every handler pairing on the system and
none of them changes verdict if you exclude the refusals. The one that is currently blocked is still
correctly blocked. So this is not a fire. But it is getting worse quickly: before July, refusals were
**zero percent** of all failures; in August they are **sixty-two percent**, and today they were
sixty-six out of seventy-four. They also never expire, so one bad batch run against tool-owned pages
permanently damages a handler's record, and a handler that has been stopped gets no more work and so
can never earn its way back.

**What I changed, and what I deliberately did not.** I did **not** change the rule itself. The
obvious fix is to teach it to recognise refusals by reading the error message — and that would mean
that anyone who ever rewords that error message silently changes which work the system dispatches,
with nothing failing to warn them. The right answer is for a handler to *say* "I refused" in a way
the rule can read, rather than us guessing from its prose, and that is a change to a shared interface
that needs proper review rather than being slipped in. I have written it up as such.

What I did change is the message a person receives when the system escalates one of these. It used to
say "fix the handler". On the 21st that message was going to go to someone about a handler whose
failures are nearly half correct refusals. It now tells them to break the failures down first, gives
them the query, and says that if refusals dominate then the handler is fine and the real problem is
the page needing a different route.

**And I got that message wrong on the first attempt**, which is the third correction in this entry. I
pointed the reader at a bug file that had been closed the day before, having copied the reference out
of my own handover notes without checking it still existed. I have fixed it to point at the part that
is genuinely still open. The lesson I took: a stale reference sitting in prose is untidy, but the same
reference inside a message that only fires when someone is already stuck is a dead end at the worst
possible moment.

**Nothing needs a decision from you.** The next real job is still the shared router engine design
round.

---

## 2026-08-18, evening — your six decisions, and the answer to the tool question

You asked why the tool fix isn't working properly. The answer is not what I expected and it is
better news than it sounds.

**Nothing is broken and nothing is missing.** The machinery for repairing a tool or widget page
exists and is the best-performing repair path we have: it has succeeded 220 times and failed 5. The
problem is that these particular complaints never get sent to it.

**What decides success is the kind of repair, not the page.** The general page-rebuilder succeeds on
tool-owned pages seventy-four times — whenever the job is to *add* something: fill an empty section,
write a missing page. It has never once succeeded, across thirty-nine attempts, when the job is to
*change* something that is already there: fix stray formatting characters, repair a link that points
nowhere, replace placeholder contact details. That is not a coincidence — changing existing content
is precisely what the safety guard exists to refuse on a page that belongs to a tool.

**And here is why we cannot simply redirect them.** The editor that can safely change a tool page
needs to be told *what the corrected content should be*. Our detectors only report that something is
wrong — that stray asterisks reached the page, not what the sentence should say instead. So the
missing piece is a small translation step between "this is broken" and "here is the fix". That is a
genuine gap, it is a design job rather than a patch, and roughly 134 waiting complaints sit behind
it.

**The immediate protection you asked for is cheap and I have a concrete way to do it.** When the
guard refuses, it currently records the refusal as a failure of the repair handler — and our new
safety rule counts failures to decide whether to keep trusting that handler. So refusals for work it
was never allowed to attempt could switch off a repair path that works two times in three on
ordinary pages. The fix is to record the refusal as "not applicable" rather than "failed". I checked:
that status is already excluded from the trust calculation *and* it releases the record so the
complaint comes back naturally once routing exists. No new machinery, and reversible.

**On the other decisions.** I am fixing the one-way door so that findings held back can be reclaimed
once we prove a repair type works — this matters more than it did yesterday, because we now have a
real case of a repair type being wrongly held while it had nine successes to its name. I have left
both canaries alone as you said; one would have failed for a documented reason and the other would
have rebuilt a live tool page unprotected. And I have written down a small worry for later: that tool
page is marked as an ordinary page in our records, so the guard that protects other tool pages would
not have protected it. If that marking is inconsistent across the estate, the routing work inherits
the problem.

**The order I intend to work in**, so you can redirect me if it is wrong: the one-way door first
because it has a clock on it (three findings cross the line tomorrow); then the "not applicable"
change and the unstable-identifier fix together, since they share a rebuild; then closing 083 next
week; and the translation step last, with a design review first, because it is the only one that
actually repairs those 134 pages.

---

**2026-08-18, later that evening — the "not applicable" change is written, tested and half-live.**

The second job on my list is done and committed. Here is what it does, in plain terms.

When one of our tools or widgets has its own page, the system refuses to let the ordinary page
builder rewrite it — a generic rewrite would wipe the working tool. That refusal is right and I have
not touched it. The problem was the *record* it left: the job was filed as "failed", which is the
same word we use for a handler that tried and could not manage it. Somewhere else entirely, a safety
rule counts those "failed" records to decide whether a repair type is still worth dispatching. So
refusals for work the handler was never permitted to start were quietly voting to switch off a repair
type that works about two times in three on ordinary pages.

From now on that refusal will be filed as "not applicable" instead, and the safety rule ignores it —
which is honest, because the handler was never given a chance to try. A genuine failure on the very
same step still files as "failed" and still counts. I made a point of testing that second half: I
deliberately broke the code two different ways and checked the tests caught both, because a version
of this that marked *every* failure "not applicable" would pass a careless test and would be worse
than the bug I set out to fix.

**One thing I got wrong yesterday and want to correct.** In yesterday's write-up I gave two reasons
for choosing this status. The first — that the safety rule ignores it — is right, and I have now read
the rule itself rather than trusting my note. The second reason was that it "releases the record so
the complaint comes back naturally". That is true, but it is not a *reason*, because the old "failed"
status released it too. So there is really only one reason, and it is a good one. I would rather say
so than leave a sentence standing that sounds like evidence and is not.

**Where it stands.** The database half is live now. The code half sits waiting for the next time the
services are rebuilt, which happens routinely — nothing needs doing to make it happen, and until then
everything behaves exactly as it did this morning. I have also put it through the review council; the
verdict has not come back yet, and I have labelled the commit honestly as "submitted" rather than
"reviewed" so nobody reads it as approved.

**What I could not settle, and am not going to pretend I did.** This change repairs no page. It stops
a filing error from switching off a working repair path; that is all it is for. The ~134 findings
sitting on tool pages still need the translation step — turning "this page has a defect" into "here is
the corrected text" — and that is the design piece I want to write up properly before anyone builds it.

**One loose end worth someone's time.** The page *re-render* path saves to tool pages nearly four
thousand times without being refused, while the page *build* path is refused every time. Both go
through the same guard, so one of those two facts needs explaining. I have not chased it — it is not
this job — but nobody should conclude the guard covers every save until somebody does.

---

**2026-08-18, late evening — the review board sent my change back, and it was right to.**

I mentioned earlier that I had put the "not applicable" change through the review council. The
verdict came back **revise**, not approve, and the objection was a good one. I want to write it down
properly because it is the most useful thing that happened this evening.

The reviewer asked a question I had not asked myself. My change has the page handler write a status
onto the job record. But after the handler finishes, the machinery that dispatched it runs one more
step — and there is a note in our own trap file saying that step **overwrites** what a handler
wrote. So: does my change do anything at all?

I had not checked. I had been thorough about one half of the question — I listed every part of the
system that *reads* this status, and there are several — and I had not looked at what else might
*write* it afterwards. The list I did make was long and careful, which is exactly what made the
gap invisible.

**The answer turned out to be a split.** There are three pieces of code that can set a job's final
status. Two of them explicitly refuse to overwrite a decision a handler has made — so my change
stands. The third has no such protection. Counting the actual jobs: **113 of 115 refusals go through
the protected route and 2 do not.** So the change works for about 98% of cases, and for the other 2%
nothing gets worse — they behave exactly as they do today.

Two things worth saying about that.

**First, the missing protection is a real defect and it is not mine to fix.** The unprotected piece
is shared by every dispatch loop in the system, and adding the same protection there would change
how retries work fleet-wide. Another thread is already working on that exact function this week for
a different reason, so I have written the finding into their file rather than starting a competing
fix.

**Second, and this is the part I find interesting:** nobody could have found this defect before
today, including me, because every route currently writes the same word — "failed". You cannot tell
"the handler said failed and it stood" from "something overwrote it with failed". My change is what
introduces a second possible answer, and that is what makes the overwrite visible at all. The defect
was not hidden; it was unobservable.

**Also done this evening.** I fixed the unstable-identifier bug (the one where a finding points at a
component that a routine rebuild has since replaced). While measuring it I found the exposure is
larger than we recorded — 12 of 82 findings already point at something that no longer exists, not 1
of 20 — and that five of them died in the single day since that bug was written. The stable way of
identifying the component works for all 82. There was one trap: the stable identifier is not always
unique (17 places on the estate have two components in the same slot), and none of them is one of
these findings today — so a fix that ignored it would have been correct on every case that exists
and wrong on the first one that did not. That case now has a test.

And I filed the last small item you asked for: our review council can't be given a
configuration-only change, because it decides what to review by looking at where a file lives, and
our configuration happens to live under a documentation folder. Two thirds of the changes that alter
live behaviour are in that category. I nearly did not file it — there is an older note in our own
records saying that refusal is correct, and it is, but it was arguing about written documents, not
about configuration. Reading what the old note was actually about is what let me file this one.

---

**2026-08-18, closing the evening — both changes approved, and the approval caught me one more time.**

Both went through the review council in the end. The unstable-identifier fix passed first time. The
"not applicable" change took three rounds, and I want to record that plainly rather than round it up:
**both rejections found something real.** The first was the overwrite question I described above. The
second was harder to take, because it was a sentence I had written in four places.

I had been saying that repairing these tool pages needs a piece of machinery nobody has built — that
we can turn a defect report into a page fix only after someone designs the translation step. **The
reviewer asked whether I had actually looked. I had not.** It exists. Another thread built it
yesterday, and it already produces exactly the format the repair tool consumes, with the careful
parts — don't invent a fact, don't lose a link, don't change a field's type — already thought
through.

**And then the approval caught me a third time, on my own correction.** Having found the thing I
said didn't exist, I described it in one line as barely used — and the reviewer asked where those
numbers came from. They came from me, and two were wrong: I had compared two different things, and
I had called a one-day-old mechanism dormant. Finding what you missed is not the end of it; the
first sentence you write about it comes out of the same hurry that produced the mistake.

**What that means practically:** the big remaining piece is smaller than I told you, and it starts
with a conversation rather than a design. The thread that owns that machinery is still changing it
daily. Writing a design around it tonight would be writing against something that will have moved by
the time anyone reads it.

Nothing is running differently in production yet — both fixes wait for the next routine rebuild of
the services, and I have written down exactly what to check afterwards, including the control that
stops a good-looking result from being mistaken for a working one.

---

**2026-08-19 — the rebuild went out, both fixes are live, and checking them cost me two numbers.**

The services were rebuilt this morning, so both changes are now actually running. I checked that
properly rather than trusting the version tag: I asked each of the two running copies of the service
whether the new code is inside it, and included a deliberately fake thing to look for as a control —
so that a "yes" means something. Both copies: yes to both changes, no to the fake. That part is
settled.

**What is not settled is whether they behave correctly, and I want to be straight about why.** No
page has been refused since the rebuild — but no page build has run either, so the silence tells us
nothing. It is the difference between "the alarm didn't go off" and "nobody opened the door". These
refusals happen about four times an hour on ordinary traffic, so the answer will arrive on its own
by tomorrow. I could force one, but forcing one wastes exactly the expensive work that the related
bug exists to complain about, so I would rather wait a day.

**Two things I told you were wrong, and one of them matters.**

The first is arithmetic. I said one of our repair types was running at 47% success and was close to
being switched off automatically. It is at 62.7%, and the automatic cut-off is 25% — it would need
another 243 failures to get there. The two numbers I quoted either side of it were right; I simply
divided them wrong and then repeated it in four places.

The second is worse, and I caught it only because I went looking for something specific. I had been
treating "this page is owned by a tool **and** the job failed" as meaning "our guard refused it". It
does not always. Two of those failures were the content writer genuinely failing, not the guard
refusing — and those two were the entire case I was about to build a follow-up fix around. **So the
honest position is that yesterday's change protects us from future noise but does not release
anything that is currently stuck.** That is still worth having — there are 85 refusals already
recorded and 134 more jobs queued behind the same refusal — but I had it sounding more like a rescue
than it is.

**On whether we can close this lane: not yet, and the reason is worth your attention.**

The original job here was "a page with missing content has no repair handler anywhere". We built the
router, it works, and every stuck job now carries a label saying why it is stuck — 27 say "there is
no content to work with", two say "the picture was sourced", one says "no plan, owned page". Nothing
strands silently any more.

**But nothing repairs them.** Of the jobs that reached "done", forty-four got there because a sweep
noticed the problem had gone away by itself — the page acquired content some other way. None was
repaired by what we built. The specific page the bug was filed about is still sitting there,
labelled, unrepaired. Closing on the strength of a tidy queue would be exactly the mistake we keep
writing down: the list looks healthy and the page is still broken.

**And the missing piece is probably the same missing piece as the other big job** — something that
turns "this is wrong" into "here is the corrected text". Another thread built the beginnings of that
yesterday. My recommendation is that the next step is a conversation with them, not a design
document written past them.
