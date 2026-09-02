# Where we are — the news-feed half-speed bug (bugs_open/410, the phase-lock one)

*(Append-only, newest at the bottom. Plain prose for the owner.)*

## 2026-08-26 evening — picked up, cause confirmed, fix written

The idea.uk thread found this afternoon that every news site whose sources refresh
"every 6 hours" is actually refreshing every 12. Nothing errors — every run reports
success — the sites just quietly get half the news updates you designed for, and have
done since the feature went live.

The cause is a timing near-miss, not a failure. After each fetch the system writes down
"come back in 6 hours from NOW". But "now" is a few seconds to a few minutes AFTER the
6-hourly timer fired, because sites are processed one at a time. So when the timer fires
again 6 hours later, each source's come-back time is still a few seconds in the future —
it misses the bus by seconds, and the next bus is 6 hours away. Every site whose sources
are all 6-hourly misses every other bus, for ever, in lockstep.

The fix I've built: when the timer fires, treat anything due within the next 3 hours
(half the timer's cycle) as due now — catch the nearest bus, not the one after. That
window is read live from the scheduler's own setting, so if you ever change the cycle the
window follows automatically. It cannot make anything fetch more often than its label;
it only stops the seconds-scale miss.

Two things you should know before this goes fully live:

1. **It restores the spend you originally designed.** Fixing this roughly doubles feed
   fetching (sites go from 2 to 4 refreshes a day — what "every 6 hours" always meant).
   Flagged to the review council too; if that's not wanted, the honest alternative is to
   relabel the cadence, not to keep the silent half-speed.
2. **Two-step deployment.** The code half rides the next fleet build; the config half
   (migration 653) is deliberately held back and applied by hand AFTER that build is
   confirmed live — applying it early would just create harmless-but-noisy empty runs.

The council review is being submitted alongside the commit; the independent diagnosis
loop was also fired at the claim before I built anything. Tonight around 21:47 UK time
the bug's own prediction gets its decisive test: the idea.uk site should be SKIPPED by
the 8:47pm pass (its sources come due 24–42 seconds after the timer fires). I'll record
the outcome either way.

## 2026-08-26 late evening — it's live, and there's one check left overnight

The fleet build went out and the fix is now live on both halves: the code is running in the
new chassis (I proved that by asking the running binary whether it contains the new rule, with
two control checks either side so a false "yes" couldn't slip through), and the configuration
change went in by hand about ten minutes later, in the order we planned. I read the live
configuration back afterwards rather than trusting the migration's own say-so; the new rule is
there, and the two earlier fixes it sits alongside are untouched.

Before that happened, the evening refresh gave us the bug one last time, on camera. The timer
fired at 20:46:45. The idea.uk site's sources became due at 20:47:24 — thirty-nine seconds
later — so it was skipped, exactly as predicted this afternoon. Three other sites *were*
refreshed in that pass, and each of them had been sitting due since quarter to three: they had
been skipped by the previous run for the same reason. That is the twelve-hour cadence we set
out to fix, doing its thing for the last time.

**What is still outstanding is one overnight check, at about 03:47 UK time.** The three sites
refreshed tonight come due again a few seconds after that pass fires. Under the old behaviour
they would miss it and wait until the morning; under the fix they should be refreshed. If they
are, the fix is doing exactly what it was built to do on real traffic. If they aren't,
something isn't reaching the decision and the handoff says where to look first. I have written
that check into the handoff with the exact query, because it is the difference between "we
believe this works" and "we watched it work".

One piece of honesty worth recording: I also put the original diagnosis through the automated
second-opinion loop, and it came back "not confirmed — stopped, scope not narrowing". That is
*not* the loop disagreeing with us; it reached no conclusion at all, most likely because I
handed it a finished answer rather than a symptom to narrow down. The evidence the fix actually
rests on is the live measurements, the predictions made in advance and then confirmed, and the
review council's approval first time. I'd rather write that down than let a blank verdict read
as a tick.

---

## 2026-09-02 — the overnight check finally got done, and it passed

Nobody came back to this after the 26th, so the check I left for "about 03:47 that night" sat
unrun for six days. The records it needed are only kept for two days, so that exact test was
gone by the time I looked.

The good news is that six days of running turned out to be better evidence than one night would
have been, and the answer is clear: **the fix works.**

Here is the test I could actually run. When the news refresh visits a site, it writes down "come
back in six hours" — counting from the moment it visited, not from the timetable. So if it
visited a site at 3:03am, that site is not officially due again until 9:03am. This morning's
refresh started at **8:58am** — five minutes *early*. Under the old behaviour it would have
looked at that site, decided it was not due yet, and skipped it until the afternoon. That skip
is the entire bug.

It did not skip it. It refreshed it. Same story for two other sites, one of them by nearly
eleven minutes. And one of the three is idea.uk — the very site that started all this, the one
we caught being skipped by thirty-nine seconds. **That is the fix doing exactly what it was
built to do, on real traffic, and there is no other way those three sites could have been
refreshed this morning.**

I threw out a fourth site that also looked like a pass. Its margin was four seconds on the wrong
side, which means it would have been refreshed either way — it proves nothing, and this lane has
already made the mistake once of counting a result that could only come out one way.

### There is a catch, and it is a different problem with a different name

The bug is fixed, but the sites still aren't getting refreshed every six hours. They're getting
refreshed roughly every nine.

The reason is simple arithmetic and it is not this bug. Each refresh pass will only take **ten**
sites. We now have **fourteen** news sites, and since the fix they are *all* genuinely due at
every pass — which is the point. So four sites get turned away each time. They aren't lost, and
they aren't being treated unfairly (the ones waiting longest go first, which is a fix we made
back in August), but four out of fourteen wait an extra six hours.

Before the fix: twelve hours between refreshes. Now: about nine. Designed: six. **We recovered
most of the gap and the rest is a capacity limit, not a defect.**

**This is your decision, not mine, and it costs money.** Lifting the limit from ten to fourteen
would get every site onto the intended six hours, and would increase how often we go out and
fetch news — which is the spend I flagged on the 26th. I have not touched it. I have written the
numbers into the other bug file that already owns this question (`bugs_open/316`, which is
literally titled "the queue is 2× oversubscribed"), so it is recorded and costed wherever you
next look.

One thing to watch if you do change it: the limit of ten is written into the configuration as a
plain number, and the number of news sites grows every time we add one. Six days ago the right
answer was twelve; today it is fourteen. A fixed number here will quietly go out of date again.

### Two other things I checked, so they aren't left as open questions

I said on the 26th that I didn't know whether the *other* six-hourly job — the one that publishes
the provocation feed — had the same flaw. **It doesn't.** It doesn't work the same way at all:
it picks the day's item from a pool rather than tracking "when is each source next due", so there
is no timestamp for the timetable to fall out of step with. That question is closed.

And I re-checked that the fix is still actually installed, rather than assuming it. Both halves
are — the code half survived a later software release, and the configuration half is intact and
untouched.

### One unrelated thing you may want to know about

On the evening of the 1st, a refresh pass **hung**. It got through exactly one site, sat there
for four hours, and was eventually killed automatically. Thirteen sites lost that pass entirely.

That is a known separate fault with its own owner — nothing to do with this fix — but it is worth
saying out loud because **it produces the same symptom we just spent a week eliminating**: a site
that goes twelve hours without a refresh. If someone looks at the feed timings in a month and
sees a twelve-hour gap, that will probably be this, not the bug we just closed.

### Where this leaves the bug

Closed. It was fixed on the 26th, it went live the same evening, and as of this morning we have
watched it work on real traffic rather than inferred it. The file has moved to the closed pile
with the evidence attached.

---

## 2026-09-02, later — a new build went out, the fix survived it, and it passed the test a second time

A fresh chassis build (`v1.0.1354`) was deployed this afternoon. I checked both running copies
rather than trusting the version number, and **both carry the fix.**

While I was there, another refresh pass had happened, so I ran the same test again on a completely
different set of sites — and it passed again. **Three more sites were refreshed by a pass that
fired two, four and seven minutes before those sites could possibly have been due.** That is six
sites now, across two separate occasions, doing something that was flatly impossible before the
fix. I am as confident as I can reasonably be.

Two small things the second run showed that the first could not. There is always **exactly one
site per pass that proves nothing** — it is whichever site got refreshed first last time, because
its clock started closest to the timetable, so it lands on the borderline. I threw that one out
both times. And the four sites that get turned away **rotate properly**: the four turned away this
afternoon are precisely the four that got served this morning. Nobody is being starved; they are
taking turns being late.

One honest gap: the new build went out *after* this afternoon's pass, so everything I watched was
produced by the previous build. The new one demonstrably contains the fix, but it hasn't had a
turn yet. Tonight's pass, around ten to nine, will give it one. **That is a ten-minute check, not
a blocker** — I've written it into the handoff with the exact query.

### The decisions I need from you

**First: do you want to lift the ten-site limit?**

Each refresh pass will only take ten sites. We have fourteen, and since the fix all fourteen are
genuinely due every time — which is the point of the fix. So four get turned away each pass and
wait for the next one.

The effect is that sites refresh about every **nine** hours instead of the intended six. Before
the fix it was twelve, so we have recovered most of the gap, and nobody is stuck at the back of
the queue. Lifting the limit from ten to fourteen would close the rest of it.

**The cost is the reason this is yours and not mine: it roughly doubles how often we go out and
fetch news.** That is the spend I flagged last week. I have not touched it.

If you do lift it, one warning. The limit is a plain number sitting in the configuration, and the
number of news sites grows every time we add one. Last week the right answer was twelve; today it
is fourteen. Whatever number we set will quietly go out of date, so it is worth either working it
out automatically or deliberately leaving room.

**Second, and only if the answer to the first is "leave it": is nine hours simply what we do now?**

If we accept nine, then what we say the system does (six-hourly news) and what it does (nine)
differ permanently. That is fine as long as it is *written down as a decision* rather than sitting
around as an unclosed job someone will keep picking up. I have put the full numbers into the other
bug file that already owns this question (`bugs_open/316`), so it is recorded either way — but
whether that file describes an accepted limit or an outstanding obligation is your call, not mine.

### What is left on this lane

Almost nothing. The bug is closed and needs no further work. Beyond your two decisions above and
tonight's ten-minute confirmation, there is **one thing nobody owns**: nothing checks, on a
schedule, that the two halves of the fix still agree with each other. If a future change quietly
dropped the fix from the configuration half, the only symptom would be sites drifting back to
twelve hours, and we would find out the slow way. Building that check is a small piece of new
platform work that needs its own review round. **I did not build it, and I did not hold the bug
open for it** — but it is the one real gap and I would rather name it than let it disappear.
