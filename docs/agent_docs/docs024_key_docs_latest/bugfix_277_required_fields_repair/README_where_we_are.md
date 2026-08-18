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
