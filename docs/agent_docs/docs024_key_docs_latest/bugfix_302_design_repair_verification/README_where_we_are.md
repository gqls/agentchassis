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
