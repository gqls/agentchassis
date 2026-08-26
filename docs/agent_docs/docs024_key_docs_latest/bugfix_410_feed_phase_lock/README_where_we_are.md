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
