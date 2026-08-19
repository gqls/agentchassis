# Where we are — bug 029, "hung spawns"

Plain prose, append-only, newest at the bottom. For the owner.

---

## 2026-08-18, early afternoon — what this bug turns out to be

You asked me to pick up bug 029 and I've now got a mechanism for it that I don't
think anyone has had before. I want to explain it plainly, because the short
version is a bit surprising: **the platform's own retry is what's killing the work.**

### First, what a "retry" is here

When one agent asks another to do a job, it doesn't wait on the phone. It sends a
message and writes down "I am waiting for an answer to request X, and I'll give it
up to N minutes." If the answer doesn't come in time, it re-sends the same request
and waits again. It does that three times, then gives up and reports a failure. The
failure message you see everywhere in this bug's history — *"Request … timed out
after 3 retries"* — is that giving-up moment.

### The rule the platform is meant to follow

Each step in a workflow can say how long it should be allowed. The build dispatcher
says 900 seconds — fifteen minutes. That's a deliberate number: dispatching a site's
work genuinely takes a while.

### The rule it actually follows

It honours that fifteen minutes **on the first attempt only**. On every retry it
quietly replaces it with **five minutes**. And if a step asks for more than thirty
minutes, the retry gets **three** minutes instead. So the longer a job says it needs,
the *less* time it's given when things are already going badly. That's backwards, and
it isn't written down anywhere — it's two lines in one function that nobody has
revisited since they were added as an improvement on an even worse hardcoded value.

I checked how much this matters rather than assuming. For the build dispatcher, about
**a quarter of its perfectly normal, successful runs take longer than five minutes**,
but only about one in twenty takes longer than the fifteen it actually asked for. So
the shortened window turns a retry that would nearly always have worked into one that
fails about a quarter of the time. For the page builder it's worse — half a percent of
its runs exceed its real limit, but nearly eighteen percent exceed the five minutes it
gets on retry.

This isn't confined to those two. **Thirty-three steps across twenty-five different
agents** ask for more than five minutes, so all of them are silently getting less than
they asked for. One of them is a step that waits for a *human being* to approve
something and asks for twenty-four hours. On a retry it would be given three minutes.

### And now the part that actually destroys work

Because the window is short, the caller burns through its three retries in about
twenty-five minutes instead of an hour. So it sends its final re-send while the job it
is chasing **is still running perfectly well**.

I looked at every case where this happened over the last few days. In eleven out of
twelve, the job being chased **stopped dead between eleven and twenty-two seconds after
that final re-send arrived**. Not slowed down — stopped, mid-task, and never moved
again. (I'm reporting the twelfth as well: it stopped before the re-send, so it doesn't
fit, and I'd rather you saw that than a tidy story.)

So the sequence is: the job is running fine → the caller loses patience early because
of the shortened window → the caller pokes it again → **the poke kills it.**

### Why one of these costs hours rather than minutes

A job that dies this way dies in a state nothing is watching. The two mechanisms that
normally rescue stuck work both key on "is something waiting for an answer?" — and this
job isn't waiting for anything, it's mid-task. So neither can see it. The only thing that
ever touches it is a cleanup sweep that looks for jobs stuck for **more than four hours**.
In the meantime the site it was working on stays locked and can't be dispatched again for
forty minutes.

That's why this bug reads, from the outside, as "builds mysteriously stop and then
mysteriously start again."

### What I've done so far, and what's next

I've put the diagnosis through the platform's own diagnosis loop before committing to it,
which is what our own rules require for a claim this structural — I'd rather be told I'm
wrong by that than by you. Separately I've asked Fable to design the fix, with a brief to
prefer something that fixes the framework rather than patching this one dispatcher, and
explicitly to contradict me if the code doesn't support what I've found.

There are really three separate faults tangled together here, and I want to keep them
separate because they need different answers:

1. **The shortened retry window.** Cheapest to fix, widest reach, and it's the thing that
   causes the caller to give up early in the first place.
2. **A re-send being able to kill a job that's still working.** This is the one that
   actually destroys work, and fixing (1) alone would only make it rarer, not safe.
3. **A job frozen mid-task being invisible for four hours.** This is why one incident is
   expensive rather than trivial.

I'll come back to you with the plan once the diagnosis run and Fable have both reported.
One thing I should flag now: I have *not* proved **why** the re-send kills the job. I have
proved that it does, to within about fifteen seconds, eleven times out of twelve. The
"why" is still a hypothesis and I've marked it as one.

### One thing I got wrong along the way

My first measurement of how long these jobs survived gave a beautifully consistent answer —
every one lasted about four and a half hours. That was nonsense: I was reading the column
that records when the *cleanup sweep* wrote to the row, not when the job actually stopped.
The real answer was twenty-five minutes, and it's the twenty-five minutes that led to the
whole mechanism. A uniform, plausible number from the wrong column is exactly the kind of
thing that looks like a finding.

---

## 2026-08-18, later the same afternoon — I was wrong about the most striking part, and here is what actually happens

I need to correct the account above before anything else, because the headline of it was
wrong and I'd rather you heard that from me than discovered it later.

### What I said, and what is actually true

I told you: **the platform's own retry is killing work that was still running.** The
evidence was that the jobs stopped dead within about fifteen seconds of the caller's final
re-send. That measurement was right and I still stand behind it. What I got wrong was what
it meant.

The truth is: **the job was already dead when the re-send arrived — by about ten minutes.**

I should have caught this myself, because the code says so. The "poke" can only happen at
all if the job has gone quiet for more than five minutes, and a healthy job cannot be quiet
that long — every time it makes progress it stamps the clock. So going quiet for five
minutes *is* death. It was the precondition for the poke, not the result of it. I read that
guard and then asserted something it forbids.

### What caught it

I had asked Fable to design the fix, and briefed it explicitly to contradict me if the code
didn't support what I'd found. It did. Then I checked its claim myself rather than taking
its word, and the check was decisive in a way I want to flag because it turned on something
small.

If Fable was right, each dead job should show **two** records of starting the same piece of
work — one for the real attempt, one for the pointless re-run. If I was right, there should
be one. Seventeen of the eighteen showed two.

And the eighteenth showed one — and it is exactly the odd-one-out I mentioned in my earlier
note, the single case that didn't fit my pattern and which I reported rather than quietly
dropping. It doesn't fit because in that one case the re-run never happened. **The outlier I
kept out of honesty turned out to be the case that settled the question.** That is the best
argument I have ever had for not tidying away the inconvenient data point.

### So what IS happening

Most of what I told you stands, and one part is worse than I said:

- **The shortened retry window is real, and it's worse than I described.** I said it makes
  the caller give up early. It does — but it also fires one level *further down*, inside the
  job itself, where it abandons real page-building work. So it's causing damage at two
  levels, not one.
- **The job dies on its own, shortly after starting a piece of work.** This is the actual
  "hung spawn" this bug is named for, and **I do not yet know what kills it.** That is now
  the open question at the centre of the bug. Fable has a candidate — a recovery path given
  only sixty seconds to do work that takes minutes — which is plausible and unproven.
- **The re-send is still harmful, just differently.** Instead of killing live work, it
  re-runs work that's already dead. That costs a duplicate build agent each time (real
  money and real side effects), and — this is the sting — it resets the four-hour cleanup
  timer. So poking the corpse makes it lie there four hours longer.

### What this changes about the fix

Less than you might expect, and this is the useful part. The fix I was heading towards is
still the right one, and the reasons are now firmer: don't send a re-run at a job that
already exists, and stop silently shortening the window everyone configured. What changes is
the *claim I can make for it*: I can no longer say it prevents work being destroyed. I can
say it stops us paying for duplicate agents, stops corpses being kept alive, and stops real
page work being abandoned one level down.

I've told the other affected thread. They had already, on their own initiative, reduced my
mechanism to a pointer at my notes rather than restating it in their own handoff — which
means my error did not propagate into their cold-start document. That was their good
judgement, not my carefulness, and it is worth recording as such.

---

## 2026-08-18, end of the day — approved, and what it did and didn't buy

The retry-window fix went through the review council and was **approved on the third
attempt**. The first two attempts were sent back, and both times the reviewers found
something real rather than something procedural, so it's worth telling you what.

**First rejection: I contradicted myself and hadn't noticed.** My submission argued that a
different, bigger fix was the important one and that shipping only the small fix "would be
the classic mistake" — and then it shipped only the small fix. That sentence was left over
from *before* I corrected myself that afternoon. When the correction landed I updated the
evidence and left the conclusion standing above it. The document carried its own refutation
and its stale headline in the same breath, and the reviewer read both.

**Second rejection: I claimed something was unused without checking.** There's a second,
duplicate copy of the same defect elsewhere in the code. I said it was unreachable and didn't
need fixing — and I hadn't actually verified that, I'd assumed it. The reviewer refused it on
exactly that ground. I then ran the check properly, and the claim held. But it held by luck
of being right, not by my having earned it, and the check turned up something genuinely
useful: two different functions share a name, one live and one dead, so anyone searching for
that name finds the live one and concludes the dead code is running. That's now written down
as a trap for the next person.

**What I changed because of the review:** a reviewer pointed out I'd written a second way of
reading a step's declared timeout when the codebase already had one. In a fix whose entire
subject is "two bits of code disagreed about a timeout", writing a third reader would have
been faintly absurd. It now goes through the existing one.

**What this buys, honestly.** Thirty-three steps across twenty-five agents stop being given
less time than they asked for, and one of them is a step that waits on a human. That's a real
improvement and it will reduce how often the failure fires. **It is not why builds stop.**
The thing that actually freezes a job is still unexplained, and I'd rather say that plainly
than let an approved fix imply the bug is dealt with. Everything I've written — the commit,
the register entry, the bug file — says "part A, does not close it" in those words.

The fix is committed but does nothing until the fleet next rebuilds, which isn't mine to
trigger. When it does, there's a check written down that can tell whether it worked, and it's
designed so that "no change" and "the test didn't run" can't be confused with each other —
which, after the day I've had with measurements that couldn't come out any other way, felt
like the least I could do.

---

## 2026-08-18, evening — the fix is proven working, and the real bug finally has a shape

**The check we designed came out positive.** At 18:28 a build dispatch step retried, and it was
given the full fifteen minutes it asks for instead of the five it used to get — and the job it
was waiting on answered inside that window rather than being given up on. That is the fix
working in production, on a real job, not in a test.

I want to be clear about why one single case is enough here, because normally it wouldn't be.
Before the fix, every retry of this kind in the entire recorded history — two hundred and
nineteen of them — got five minutes. Not most of them. All of them. So fifteen minutes isn't a
lucky roll of the dice; it is a result the old code was incapable of producing. We also checked
that nothing *else* moved: first attempts still get exactly what they always got. If those had
changed too, it would have meant we'd broken something broader.

It took two hours and forty-three minutes after the fleet updated for a suitable case to come
along. During that wait there was plenty of traffic — hundreds of jobs — just none of the
particular kind that could tell the two versions of the code apart. It would have been very easy
to point at all that traffic and call the fix verified. Every number in that claim would have
been true and the conclusion would have been unearned, so we sat on our hands instead.

**The bug itself is still open, and this is the part worth reading.** I went looking at every
frozen build job we still have records of — eighteen of them. They are not eighteen different
problems. They are one problem, repeated with a consistency I did not expect:

- Every single one froze after a *previous* step had already failed. Never during normal running.
- Every single one had successfully started its next job — the handshake worked — and then simply
  stopped, about twelve seconds later, without ever recording that it was waiting for anything.
- Seventeen of the eighteen had asked for that next job **twice**.
- And they all died at almost exactly twenty-five minutes old, within a few seconds of each other.

Things that die at the same age are killed by a clock, not by bad luck. That is a much better
question to ask than the one we had this morning, which was essentially "why do jobs sometimes
freeze".

**Two things I got wrong today and corrected before they went anywhere.** I wrote a query
comparing what each step asks for against what it gets, and it reported two steps being
short-changed. They weren't. The same step name is used by two different agents that ask for
different amounts, and my query had mashed them together. Second, I noticed all eighteen frozen
jobs were from yesterday and started to conclude the problem had stopped — until I checked how
far back our records actually go, and found the answer is about a day, with the oldest record we
have being the first frozen job itself. "It started yesterday" was the filing cabinet talking,
not the system.

**One thing I need a decision on.** To investigate the frozen jobs properly I'd hand the problem
to our automated diagnosis system. It reads the code from the shared server rather than from this
machine — and the fix we just proved working isn't on that server. Nor is the exact version
currently running in production. There are 233 commits sitting here unpublished, all from today,
belonging to a dozen different pieces of work going on in parallel. If I send the diagnosis off
now it will read the old code and very likely tell us, with confident citations, that the cause
is the thing we already fixed this afternoon.

Publishing all 233 isn't something I think one session should decide on its own, so I've stopped
and left it with you. **There is a clock on it:** we only keep about a day of these records, so
the eighteen frozen jobs disappear tomorrow, and after that the investigation would be looking at
an empty table.

---

## 2026-08-19, morning — the new build checked out, and I caught myself claiming a win we haven't got

**The new build is fine and our fix is still in it.** I checked it the reliable way — asking the
running program which version of the code it was built from, with a deliberately wrong answer
mixed in to prove the check can fail — and both copies agree. First attempts still get the times
they ask for, so nothing has regressed.

**The records of the frozen jobs are gone.** We only keep about a day of them, and yesterday I
flagged that they'd disappear this morning. They have. Eighteen frozen jobs, all from Monday,
now unrecoverable from the system — though everything we learned from them is written down.

**Meanwhile the other obstacle cleared itself.** The reason I couldn't hand the investigation to
our automated diagnosis yesterday was that it reads code from the shared server and our fix
wasn't there. Someone published overnight, so it is now. The two problems have swapped places:
yesterday we had evidence and no way to use it; today we have the means and no evidence. I'm not
spending a diagnosis run on an empty table.

**And here is the bit I'd rather tell you than bury.** I measured whether the failure that
precedes a freeze had stopped happening since our fix, got a clean zero, and started writing it
up as encouraging. Then I checked how those failures were spread out over time, and the answer
was that thirty of the thirty-one all happened on Monday. Every other day was zero — **including
Sunday and Tuesday, before our fix was anywhere near production.** So "it hasn't happened since"
describes six of the last eight days and tells us nothing at all about whether we helped. I had
even hedged it carefully, which made it worse: a cautious sentence about weak evidence still
reads as evidence.

**What that means for the bug.** It is rare and it comes in bursts — one bad afternoon on Monday,
which happens to line up with a GitHub outage we'd already noticed. We can't say we fixed it and
we can't say it's still biting. What we can say is that the one real defect we found on the way
is fixed, proven, and worth having on its own terms.

**So: not closeable, and the honest reason isn't "it's still broken" — it's that we never
explained it, and it hasn't shown itself since.** The next burst is what would settle it, and
there's a catch worth a decision from you: we'd have about a day to notice and capture it before
the records vanish, exactly as they just did. If this bug matters, something needs to grab that
evidence automatically. If it doesn't, we let it lie and reopen when it bites.

---

## 2026-08-19, late morning — the capture is built, and the investigation told us honestly that it had nothing to work with

**The capture exists and it works.** There is now a job that runs every hour and takes a full
snapshot of any frozen build job — the job itself and every request it was waiting on — and
files it somewhere retention cannot reach. The next time this happens we will still have the
evidence next week, instead of losing it overnight as we just did.

I did not want to install something that had only ever reported "nothing to see", so I forced
it: dropped its threshold to zero so every job in flight counted as frozen, and it correctly
captured five real ones with their complete history, then captured nothing on a second run
because it recognised it had already stored them. I deleted those test records afterwards, so
what is running now starts clean.

**Building it turned up something I did not know and did not like.** When a job freezes, two
different processes reach it at exactly the same age — four hours. One writes a record saying
"this froze"; the other simply deletes it. Whichever gets there first wins. So the eighteen
frozen jobs we studied on Monday are the ones where the recording process happened to win the
race, and there may have been more that were quietly deleted having never been recorded at
all. **Our count of this failure has always been a floor, not a measurement.** The new capture
takes its snapshot three and a half hours before either process gets near the job, which is the
point of doing it that way rather than the easy way.

**I also sent the investigation off, and it came back "not confirmed".** That is the honest
outcome and I would rather have it than a confident guess. It ran properly this time — the last
attempt was killed by the very bug we fixed on Monday — and it did something I thought was
impressive: it recognised that the evidence it needed had expired, went looking for a live
example to use instead, found a candidate, checked it, and rejected it. I checked the candidate
myself afterwards and it had simply finished normally a few minutes later. It was never frozen.
The investigation was right to refuse.

**So we are where we were on the bug itself, with two things gained:** the fix we proved on
Monday is still live and still correct on the new build, and the reason this bug has been
impossible to investigate — that its evidence evaporates within a day — is now fixed. The next
occurrence is the one we will finally be able to look at.
