# Where we are — design-repair completion verification (`bugs_open/302`, and a check on `201`)

Plain prose, append-only, newest at the bottom. The owner maintains this too — append, never
rewrite.

---

## 2026-08-18 — what I picked up, and what I found before writing any code

I was asked to take the next open bug nobody else is on, and 302 and 201 were the two named.

**On 201 there is nothing to fix, and I checked that rather than taking the lane's word for it.**
Its two faults were fixed, shipped and proven live back on the 7th of August, and the bigger
question they exposed — what a checker should do when it cannot run at all — was decided by you
on the 8th and has been working ever since. Today I went and looked at the live data instead of
reading the lane's own summary: the checker has refused fifteen completions where the defect was
still there, and certified one where the repair was genuine. That is the important shape, because
a checker that only ever passes things tells you nothing. Two more items had closed with no check
recorded at all, which looked like a hole for a few minutes; it is not one. Those closed because
the detector itself went back, re-scanned the page, and found the problem gone — and when the
detector measures something directly there is nothing for a completion check to add. So 201 is
genuinely finished. Under your ruling of the 12th of August, a bug that is fixed and live belongs
in the closed pile, so that is where I am proposing to move it, with today's evidence written into
it first.

**On 302 the bug is real, but the file that raised it gets an important part wrong, and the fix it
recommends is the expensive one.**

The background in plain terms: when a repair job finishes, the platform asks two questions before
it stamps the job as done. The first is "did the repair agent actually change anything?" — it
reads a couple of counters out of whatever the agent reported. The second is "is the defect
actually gone?" — that one re-runs the original detector. The second question has a firm rule, and
it is yours: if the check cannot run, the job does not get stamped done. The first question does
the opposite. If it cannot make sense of what the agent reported, it shrugs, writes a note in a
log table, and stamps the job done anyway. Those two rules contradict each other, and the file
holding the second one says in its own comments that leaving a gap like the first one is exactly
what it refuses to do.

That gap is not theoretical. It fired eleven times between the 14th and the 17th of August.

Here is where the filing is wrong, and it changes the story. It says those eleven happened because
the design-repair agents hand back an analysis instead of doing the work. I looked at all eleven.
Seven of them were not the agent's report at all — they were a completely unrelated record that a
separate bug was stuffing into that field, and **that bug was fixed and shipped yesterday
afternoon.** Three were the design blob the filing describes, and one was a stray decision about a
different page entirely. So the majority cause was removed at source a day before I started.

And since yesterday's release, no job of this kind has come through at all. The fleet has been
busy — nearly two thousand jobs completed in the same window — but this particular kind has had no
traffic. So I cannot tell you what the rate is now, in either direction, and I am not going to
pretend the leak is still gushing when I have nothing to measure it with.

**What I think is still worth fixing, and why it is not just tidiness.** The first question's
whole justification, written into its own code, is that for this kind of job "nothing changed
cannot possibly mean it was repaired". An unreadable report silently excuses the job from the very
rule its owners opted it into — permanently, and for every future repair type somebody adds. That
is worth closing whether or not it is firing this week, and it costs very little.

**What I am not going to do is the fix the filing recommends.** It suggests writing a proper
before-and-after checker for each kind of design repair. I ran the check this estate insists on
before anyone does that, and it says don't: one of those job types is filed by four different
producers who each mean something different by it. A single checker over that population is a
mistake we have already made once and written up — the checker answers its own question correctly,
says "all clear", and the defect closes untouched. Doing it properly means a per-type declaration
of who each checker actually speaks for, which is a much bigger job than the filing suggests and
is not the cheapest thing that closes the door.

I have asked fable to design the fix against all of that evidence, including an instruction to
tell me if my preferred approach is wrong. The plan and the decision on whether this needs your
architecture review or just the council gate will follow in `PLAN_2026-08-18_*.md`.

---

## 2026-08-18, later — what the measuring turned up, including two places I was wrong

I said above that the filing gets an important part wrong. Having now measured all of it, it gets
three things wrong, and the third one is the fix it recommends. None of this means the bug isn't
real — it is — but it changes what to build and how loudly to claim it.

**One.** It says the checker registry covers eleven kinds of job. It covers thirteen. The two it
missed are registered through a slightly different function call, and one of them is a *design*
job type — which matters, because the filing's headline is "no design job type has a checker".
The accurate version is narrower: the design **audit** family has none, the design **discovery**
aggregate has one, and that one is in better shape than anybody claimed (more on that below).

**Two.** It says the eleven bad cases happened because the design agents hand back an analysis
instead of doing the work. Seven of the eleven were a different bug entirely, and that bug was
fixed and shipped at five past five on the afternoon of the 17th. I checked what the shipping
actually did: across the whole fleet, the broken shape appears 939 times before that release and
**zero** times in the 1,880 jobs completed after it. That is about as clean a result as this
estate ever gets. So the leak I was sent to plug had already been plugged upstream by somebody
else, the day before.

What remains is still worth fixing, and I want to be precise about why, because it would be easy
to oversell. The first of the two questions the platform asks — "did the agent change anything?"
— is only asked for job types whose owners have explicitly signed them up, and signing up means
asserting, with a measurement attached, that for this job type "nothing changed" cannot possibly
mean "repaired". If the agent's report is unreadable, that assertion is quietly waived. Not
logged-and-blocked: waived, and the job passes. It will be waived for every future job type
somebody signs up, by default. That is a door worth closing whether or not anything is currently
walking through it — and at the moment nothing is, which I would rather tell you plainly than
dress up.

**Three, and this is the one that changes the work.** The filing's recommended fix is to write
proper before-and-after checkers for the design repairs. There is a file in the codebase whose
entire purpose is to record, with reasons, which job types deliberately have no checker — so that
these gaps are decisions rather than oversights. It already covers this whole family, and it has
already decided against. For two of them the recorded reason is that checking would need a real
browser on the completion path, which this estate has refused three times for other job types.
For three more the recorded reason is that "fixed" is an aesthetic opinion with nothing to re-run.
On top of that, one of those job types is filed by four different producers who each mean
something different by it — and a single checker over a population like that is a mistake we made
once already and wrote up. So I am not doing it, and anyone who wants to should argue against a
specific recorded reason rather than starting fresh.

**Where I was wrong, twice, and caught both myself.**

The first is logged in the fleet-wide wrong-calls file. I claimed the cost of my fix was cheap
because a refused job gets tidied up later by a mechanism that closes tickets when the audit stops
reporting the problem. That mechanism is built, approved, deployed — and has never once run,
because the audit that feeds it has been switched off since the 11th. So the real cost is three
wasted rebuilds and a ticket a human has to look at. That is the same price you knowingly accepted
for the equivalent rule on the other question back on the 8th, which is the honest way to put it.
What makes this worth writing down is that I had *read* the note saying "this is not exercised" and
still didn't apply it forty lines later: a mechanism existing and a mechanism running are separate
facts, and reading code only ever tells you the first.

The second I caught before it left my notes. Chasing the above, I found that two thirds of the
platform's job history lives in a second, archived table — so nearly every "has this ever happened"
count in this estate is really "in the last seven days". I was about to write that a checker had
wrongly closed 468 jobs instead of the 11 on record. Splitting it by date first showed the real
number is **15**: the other 453 closed before that checker existed at all. Same direction as the
record, four higher, and nothing like thirty times worse. I also re-ran the safety test that
governs that checker across the full 564-job history rather than the 21 it was signed off on, and
it comes out perfectly clean — so that mechanism is better evidenced than its own comment claims.
(The archive trap itself turned out to be already written up by another thread this morning, so I
cited theirs rather than filing a second copy.)

Fable is designing the fix now, against all of the above and with an explicit instruction to tell
me if my preferred approach is the wrong one.

---

## 2026-08-18, evening — built, submitted, and what it does and doesn't buy you

The fix is in (`743bc1945`) and it is with the council (`edfef8cc`). Here is what it actually does,
in plain terms.

When a repair job finishes, the platform asks "did the agent change anything?" — but only for job
types whose owners have explicitly signed them up, and signing up means asserting, with a
measurement, that for this type "nothing changed" cannot mean "repaired". The bug was that if the
agent's report came back unreadable, that assertion was quietly waived and the job passed. From now
on, a signed-up type has to **say** what an unreadable report means for it, and the one type on the
list now says "refuse". Two things make that stick rather than being another good intention: leaving
the choice blank is now a build failure, so the next person to sign a job type up cannot inherit a
waiver they never chose; and if somebody does leave it blank anyway, the code refuses to block —
because the dangerous direction for a *blank* is blocking things nobody meant to block.

The reason I could argue for it at all is that you already ruled on this exact question on the 8th of
August, for the platform's *other* completion check: if the check cannot run, the job does not pass.
That check's own code even refuses a loophole for unreadable input, saying in as many words that
allowing one "would leave a second silent completion path". This gate — written five days after your
ruling — was that second path. So this is your rule applied to its sibling, not a new policy of mine,
which is also why I judged it a council matter rather than something needing your architecture review.
I've written the scope argument out against all four of the relevant rulings so a reviewer can
disagree with it specifically.

**What it does not buy you, and I want this said plainly rather than buried.**

It will not do anything when it ships. The job type it protects has had no traffic since the 17th,
and both of the schedulers that feed it are switched off. So after the next release the honest
status is "installed, not yet seen working" — carried by a test I added that drives the real
completion path, not by a live job. Proving it for real means manufacturing a job, which is a cost
decision and yours, not something I should do for the convenience of a tick.

And the cost, when it does fire, is real: a refused job retries three times and then sits waiting
for a human. It is *not* tidied up afterwards — that is the thing I got wrong earlier today and
logged. The mechanism that would tidy it up is built and has never once run.

**Two other things worth your attention, neither of which I did.** The 15-minute timeout sweep can
complete one of these jobs without either check running, because its exclusion list is wired to the
*other* check's list and structurally cannot see this one. It has never actually happened — I checked
all 594 jobs of both relevant types across the full history — so I left it alone and wrote it up
rather than widening a shared scheduler rule on a hunch. And the four original examples this bug was
raised on are actually three different job types; only one is covered. The other two each need a
decision from someone before a rule could honestly be written: for one of them, handing back an
analysis may well *be* the deliverable, in which case "changed nothing" is a success.

**One good side effect.** Another session's commit had turned a guard red on the shared branch about
40 minutes before I ran the tests. I proved it wasn't mine by testing a clean copy of the branch,
told them, and they had it green again quickly — and in the process they corrected something I'd told
them, which I've logged against myself. That exchange found the same underlying trap from two
directions: two thirds of the platform's job history sits in an archive table, so almost every "has
this ever happened" count in this estate is really "in the last week".

---

## 2026-08-19 — proved it, closed it, and here is what I did to prove it

You asked me to prove it, so I did, and I want to be straight about how — because "proved" can mean
very little if the test could only ever have come out one way.

**First, the fix is in the new build.** `v1.0.1314`. I checked the same way as before: the image's
own record of which commit built it, an ancestry check that says *yes* to my fix and *no* to a later
commit (a check that says yes to everything proves nothing), and the running pods' fingerprint
matching the image I inspected.

**Natural demand still never came.** Fifteen hours after the first release, not one job of the
protected kind had been touched, against 252 jobs completed across the fleet. Both schedulers that
feed it are still off. So there was nothing to wait for.

**How I induced it, and the thing I had to avoid.** The obvious move — set the design audit running
so it files a job — would have proved the *wrong half*. Since yesterday's other fix, the repair agent
hands back a report the check can read, which trips a different rule. The situation my change is
about is a job completed with **no report at all** — and that is a real, existing path: it's exactly
what one of our own orchestrators does, and it has already happened once to this job type. So I
reproduced that path with a tiny single-step probe that calls the completion check directly. **It
spawns no repair agent and reads or writes nothing belonging to any site.**

**Four cases, not one.** This is the part that matters:

| what I supplied | job type | what happened |
|---|---|---|
| **nothing** | the protected one | **refused** — "the handler's result was unreadable… refuses to certify what it cannot read" |
| a readable report, all zeros | the protected one | refused, but with the **other** message — so the two reasons stayed distinct |
| a readable report with real work in it | the protected one | **completed** — the check isn't just refusing everything |
| **nothing** | a job type not signed up | **completed** — the rest of the fleet is untouched |

The first row on its own would have proved almost nothing: a check that refused *everything* would
have produced it, so would one that had started blocking the whole fleet. The other three are what
rule those out, and each came out the opposite way.

Then I deleted the probe and the four test jobs, and checked the deletion at the data rather than
assuming it: nothing left behind, and the real success-rate figure that a live scheduler reads still
reads exactly what it did before I started.

**So the lane is closed.** The bug is fixed, live, and proven.

**What I did NOT fix, and deliberately.** The original complaint was that this family of design
repairs has no proper before-and-after checker. The answer is that it shouldn't get one by that
route, and that answer was already written down in our own code before I arrived — checking would
need a real browser on the completion path, and for several of the types "fixed" is an aesthetic
opinion with nothing to re-run. I've made that a recorded decision rather than an open gap.

**Three things outlive this lane, and none is mine to decide.** I filed the first as its own ticket
(`317`) so closing this one loses nothing: a fifteen-minute cleanup sweep can still complete one of
these jobs with neither check running, because its exemption list is wired to the *other* check's
list. It has never once happened — but only because the schedulers are off, so re-enabling one
re-arms it. The other two need a decision from you or another lane: for one job type, handing back an
analysis may legitimately *be* the deliverable, and for two others nobody has measured what a
successful report even looks like.

**And the honest footnote.** I made three wrong calls over the two days and all three were caught —
one by me, one by a peer session, one by a reviewer. They were the same mistake wearing different
clothes: *a thing that exists, or was written down, is not a thing that operates, or was done.* The
fourth near-miss was in this very proof — my first "must be absent" control turned out to be a commit
that couldn't have been absent. A control is only a control once you've checked it can fail.
