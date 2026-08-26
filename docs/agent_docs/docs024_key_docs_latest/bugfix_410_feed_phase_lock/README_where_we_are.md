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
