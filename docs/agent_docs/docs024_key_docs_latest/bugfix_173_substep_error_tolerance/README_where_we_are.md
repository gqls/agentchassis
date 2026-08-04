# Where we are — per-substep error tolerance (bug 173)

Plain prose, append-only, newest at the bottom.

---

## 2026-08-04, morning — what this is and why it was worth doing

The job was to pick up the next open bug nobody else was working on and fix it properly —
at the framework level, not just for the one site that tripped over it.

**The bug, in plain terms.** A lot of our work happens in loops: build twelve pages, one
after another. Each pass through the loop does several things — write the content, save the
sections, extract the links, and so on. Today the platform has exactly one switch for what
happens when one of those things fails, and the switch covers the **whole loop**. Either
every failure inside the loop is fatal and the entire build stops, or every failure is
shrugged off and the item is quietly skipped. There is nothing in between.

That is a genuinely awkward position to be in, and it has already forced a bad decision. A
few days ago another thread needed one small step — recording links into a table that is
regenerable and currently empty — to be allowed to fail without killing an entire site
build. With no way to say "just this step", they made the step itself pretend it had
succeeded. Four separate reviewers on our review council rejected that, and they were right:
the honest word for it is that the code was made to lie about its own outcome. The reviewers
said, in effect, *the switch you actually need doesn't exist — go and add it, don't work
around it*. That is this bug.

**Why it matters beyond that one case.** When I went looking, the same shape turns up in
three more loops — the main page builders. In each, one page failing to save could take down
a build that had already produced a dozen good pages. So the missing switch is not an
oddity of one lane; it is a gap four live loops are standing on.

**What I'm doing about it.** Letting an individual step inside a loop declare its own
tolerance, and inheriting the loop's setting when it says nothing. It is a small change in
one file. The important properties are what it *doesn't* do:

- **Nothing changes for anything running today.** I checked the live configuration: of the
  79 steps sitting inside loops across the fleet, exactly **zero** currently use the new
  setting. So on the day this ships, every existing build behaves precisely as it does now.
  The new capability is invisible until somebody deliberately switches it on.
- **The unsafe direction is the one you have to ask for.** "Ignore failures" is the risky
  setting, and you only get it by writing it down, in the place a reviewer of that step will
  see it. Doing nothing keeps you exactly as strict as you are today.
- **It works in both directions.** Just as useful as making one step tolerant inside a strict
  loop is making one step *strict* inside a tolerant loop — several of our dispatch loops
  currently swallow everything, and this lets a step opt back out of that.

**One thing I want to flag as a small trap I found on the way.** It turns out you could
already write this setting on an individual step today. It just did nothing — the code
overwrote it a few lines later without a word. So anyone who tried this would have written
something reasonable, seen no error, and got no effect. I've recorded that as a landmine so
the next person doesn't lose an afternoon to it.

**Honest status.** The code and its tests are being written now. I can commit it, but I
can't make it live: rolling a new build out is a whole-fleet operation the owner runs, not
something one session does unilaterally. Our own rule is that a bug stays open until the fix
is genuinely running in production, not merely committed — so unless another session's roll
happens to carry it out today, this will stay open with the remaining step written down
precisely. I would rather leave it accurately open than tidily closed and wrong.

---

## 2026-08-04, later — approved first time, and the three things the reviewers asked for

The council approved it on the first round: eight seats in favour, three advisory notes, none
serious. Rather than nodding at the notes I went and checked each one, which turned out to be
worth doing because two of them had real answers.

**"Does a skipped step just vanish?"** The reviewer guarding against silent data loss asked
whether a step skipped by the new tolerance leaves any trace, or whether the page it should
have written simply disappears. It leaves three: a record written into the job's own state
saying which step failed, with the error and a timestamp; a warning in the logs naming the step
and which pass through the loop it was on; and a status of "error" against that item in the
loop's summary, deliberately distinguished from "nothing found". So it is not a silent loss.
The honest caveat I recorded is that none of that raises a ticket for a human, and the job
record is deleted after about a day — so the trace is real but it expires.

**"What about the interaction with those two other settings?"** I had flagged, honestly, that I
had not looked. Having looked: there is nothing to interact with. Those two settings appear
exactly twice in the entire codebase, both times in a list that only rewrites their *names*,
and nothing anywhere reads them to make a decision. No configuration in the fleet uses either.
They are decorative.

**"File the sibling bug properly."** I had noted, in my own risk section, that the same setting
read at a different level still has the old silent-failure behaviour, and said I would leave it
for another day. The reviewer's objection to that is the most useful sentence of the round:
burying a deferral in the document that defers it means it never reaches whoever looks at the
mechanism next, because they read the code, not my write-up. So it is now bug 193 with its own
measurements.

**One error of my own, worth writing down.** A reviewer noticed that I claimed to have
registered the change in our shared register "in the same commit" but had not listed that file
among the changes I showed them. The registration is genuinely there in the commit — but the
reviewers can only judge what I put in front of them, and I had made a claim they had no way to
check. That is my mistake, not theirs, and the fix for next time is trivial: list the file.

**Where that leaves us.** The fix is committed, approved, tested and *not live*. It goes live
on the next fleet build, which is an owner operation — I have not triggered one, because it
interrupts everyone else working today and this change is, by measurement, inert until somebody
deliberately opts into it. So the bug stays open, with the one remaining test written down
precisely: make the tolerant step fail for real and watch the job continue, then make the strict
step fail and watch it stop.

**One thing I would like you to confirm rather than take from me.** There is a rule about when a
change to shared machinery needs a heavier architecture review. I argued that this change does
not need one, and the reviewers agreed with my reasoning — but two of them pointed out that I am
the author, and the author is not really the right person to decide whether their own change
needs reviewing. I think the reading is right. I would rather you agreed than assumed.

---

## 2026-08-04, end of day — it's live, it's proven, and the bug is closed

The new build went out and carried the fix. I checked it at the running system rather than
trusting the build, on both copies, including a deliberately-wrong probe to prove the check
itself could come back negative — because a test that can only say "yes" isn't a test.

Then I ran the real thing, which is the part that actually mattered and the reason I wouldn't
close this earlier. I built two throwaway jobs, each with two steps, and made a step fail on
purpose:

- In the first, the job as a whole was set to "stop on any failure", but the failing step was
  marked "tolerate me". **Both passes through the loop were skipped and the job finished
  normally**, with a record of each skip.
- In the second, the job was set to "tolerate everything", but the failing step was marked
  "I must not be tolerated". **The job stopped dead at that step**, exactly as it should.

The neat part is that I set each job's overall setting to the *opposite* of the step's. So if
the fix hadn't worked, each run would have produced the other one's result — there's no way to
pass both by luck.

I also confirmed the failure genuinely happened rather than the job quietly skipping the whole
thing, which would have looked like success for entirely the wrong reason. That distinction is
the difference between a proof and a green tick.

**So bug 173 is closed and filed away.** The throwaway test jobs have been deleted.

**Two things I got wrong today, both now written up.** The first three attempts at the live
test produced nothing at all, and I put it down to the system being busy — there's even a
guideline saying that's usually the cause. It wasn't. My test job had one line written in the
wrong style and was being rejected instantly, every time. The message had been received and
thrown away within milliseconds, and one look at the system's own log would have told me on the
first attempt. A ready-made explanation that fits the evidence is the most expensive kind of
wrong, because it stops you looking. I'd also hidden the output of the command that sends the
job, "to keep things tidy" — which is precisely how you fail to notice it silently doing
nothing. Both mistakes are the same mistake: I threw away my own evidence and then reasoned
about the gap.

**Still open and still yours to answer if you want to**: the question from earlier about
whether I should be the one deciding that my own change didn't need a heavier review. Nothing
depends on it — the change is in and working — but two reviewers raised it and I'd rather it
were settled by you than quietly by me.
