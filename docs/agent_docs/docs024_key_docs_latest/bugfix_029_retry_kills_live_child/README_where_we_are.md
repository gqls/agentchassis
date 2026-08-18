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
