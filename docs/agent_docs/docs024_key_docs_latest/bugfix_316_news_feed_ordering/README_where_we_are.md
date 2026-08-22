# Where we are — the news feed that serves the alphabet (bugs_open/316)

Plain-prose running log, append-only, newest at the bottom.

---

## 2026-08-22 — picking this up, and what I found before touching anything

Nine of our sites are set up to refresh their news feeds automatically. A job wakes up every six hours,
asks the database "which of these sites is due a refresh?", and takes the answer. The problem is the last
line of that question: it says *give me the first five, in alphabetical order*.

So when more than five sites are due at once — which is most of the time — the same names win. Not the
ones that have waited longest. The ones that start with an earlier letter.

**The bug was filed three days ago and it has got worse since.** When it was written, the four sites at
the back of the alphabet were running between seven minutes and three and a half hours late. Today
`webdesign.co.uk`, which is last alphabetically, has not been refreshed for **thirty-one hours** on a
six-hour schedule. I can see it sitting in the queue, due, at four consecutive runs, and not being picked
at any of them. Every one of the last five runs came back with exactly five sites — the ceiling — so
there was never any spare room for it.

That is worth saying plainly: this is not a queue that is slightly behind. It is a queue where one site
has been at the back of the line for over a day and nothing in the system will ever move it forward,
because its position is decided by its name.

**Two separate problems, and I want to keep them separate.**

The first is fairness. Who waits is decided by the alphabet. That is straightforwardly a defect and I can
fix it — the fix is to serve whoever has waited longest, which is what a queue is supposed to do.

The second is capacity, and it is not mine to decide. Adding up what the nine sites have each asked for,
they want **42 refreshes a day**. The job supplies **20**. So we are asking for roughly twice what we
deliver, and no amount of clever ordering creates a slot that does not exist. Even if I removed the
ceiling completely we would supply 36 against 42. Closing that gap means either spending more (run the
job more often, or lift the ceiling) or admitting that some of the schedules we have configured were
aspirational — one site is asking for a three-hourly refresh. **That trade is a spending decision and I
am going to put the arithmetic in front of you rather than quietly pick one.** I have re-checked the
numbers against today's data and they come out exactly as the original report said.

What fixing the ordering *does* buy, even with the shortfall unchanged, is that the lateness gets shared
out instead of always landing on the same four names. Everyone would run a bit behind, rather than five
sites being perfectly on time while four are permanently starved.

**One thing I want to flag about how I am approaching this.** I went and looked at every other place in
the system that asks the database a capped question like this one — there are about thirty. Most are
fine. One looked like the same bug and turned out not to be, and the reason why is the interesting part:
that one hands out work that *gets finished and leaves the list*. Ours hands out work that **comes back
round on a clock** — a site is never done, it just becomes due again later. When work leaves the list,
processing it alphabetically is only a delay. When work comes back round, processing it alphabetically
means the back of the alphabet never gets served at all. That is the distinction, and our records
previously blurred the two together. So the fix I want is not just "change this one line" — it is to
teach the estate to tell those two shapes apart, so the next person who writes a capped scheduled query
gets told about it rather than finding out a year later.

I have asked for a full plan before writing anything, and I will put it through the review council
before it goes in.
