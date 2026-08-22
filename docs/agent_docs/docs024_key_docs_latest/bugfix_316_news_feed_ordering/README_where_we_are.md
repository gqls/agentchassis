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

## 2026-08-22, later — what I built, and the one thing I need from you

The fix itself turned out to be small, and two of the three things the original report recommended were
wrong. Both corrections came from reading code rather than from being clever, and I want to set them out
because the second one is the sort of thing that wastes a day if nobody notices it.

**Correction one.** The report said to order the queue "oldest due first, nulls first". Sensible on its
face. But the query has a second job nobody had spotted: it is also how a *brand new* news site gets its
sources set up in the first place. Such a site has no schedule yet, so it sorts as "null" — and under
"nulls first" it would go to the front of the queue every single run, for ever, because nothing about
fetching can give it a timestamp it does not have. If its setup ever quietly failed, it would sit at the
head of the queue jamming all eight other sites indefinitely, and nothing would say so. I have sent those
sites to the *back* instead. It is not a free choice — a brand-new site now waits longer to be set up —
but the failure it avoids is silent and unbounded, and the one it accepts is obvious (a new site visibly
has no news). Nothing changes today: I checked, and no site is currently in that state.

**Correction two, and this is the useful one.** The report's second recommendation was "raise the cap from
5 to 10 or more". That would have done **nothing at all**. There is a *second* limit of 5 immediately
after the first one, in the step that actually processes the sites. Raise the query's limit alone and it
hands over ten sites while the next step still processes five and stops. Worse, the way we currently
measure this problem is by watching whether the query comes back full — so it would have flipped from
"5 of 5, at the limit" to "10 of 10, not at the limit", and **the instrument would have reported success
while nothing whatsoever had improved**. Both numbers have to move together, or neither does.

**What I have actually built.** Two things.

The first is the fix: a one-query change so the feed picks whoever has waited longest instead of whoever
is earliest in the alphabet. I tested it against the live database before changing anything, and it
returns *the same five sites* as the current query at this moment — same set, different order, with
`webdesign.co.uk` moving from last place to first. That is exactly what I wanted to see: it changes who
goes first and nothing else.

The second is a check that looks for this *kind* of mistake anywhere in the system, every day, and says
so. I want to be straight about the argument for it, because there is a decent argument against. Today
there is exactly **one** place in the whole platform with this defect — the one we are fixing. Building a
detector for a problem with one instance can be over-engineering. What decided it for me is *how this was
found*: by a person reading a query three days after a different investigation happened to list it, by
which point one site had been stale for over a day. We already have a check watching that exact query,
and it cannot see this — it counts how many rows come back, and the count is the same whether the order
is fair or not. It had been truthfully reporting "this query is at its limit" every six hours for days
while the actual problem went unnoticed.

I also made sure the check works before the fix hides the evidence. I ran it against the live system
first: 194 agents examined, exactly one problem found, and I have saved that output. Then the fix goes in
and the same check should find nothing — and that "nothing" means something, because I have proof the
check was capable of finding something an hour earlier.

Two smaller things worth knowing. While building it, the check reported *its own fix* as a problem — I had
been too crude about recognising a renamed column — and its own tests caught that before anything shipped.
And I found that this same starvation problem was solved once before in another part of the system, which
gave me a specific warning to check against: that ordering by "who is most overdue" only works if an item
that *fails* still gets pushed back in the queue, or it jams the front for ever. I checked all three
places that update the timestamp, and all three do push it back. So this is safe for the right reason,
not by luck.

**What I need from you.** The fairness fix is ready and I have put it through the review council. The
capacity question is genuinely yours: the nine sites are asking for **42 refreshes a day** and we supply
**20**. Ordering fixes who suffers, not the shortfall. Your options are to spend more (run the job more
often, or raise *both* limits together), or to accept that the schedules we configured were more
ambitious than we meant — one site is asking to be refreshed every three hours. I have not touched either
number and I am not going to without you saying so.
