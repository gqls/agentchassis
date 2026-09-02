# SUMMARY — dispatch throughput — 2026-09-02

*(Second summary in the series; the first, 2026-08-26, told the story up to ruling B going
live. This one closes that whole arc and opens the next. Written to be read aloud.)*

## What we're trying to do

Make the machine that hands out website work fast enough, fair enough, and cheap enough to
run hundreds of sites instead of dozens — without ever letting one site, one stuck job, or
one runaway bill stop everything else.

## Where we've come from

The 26 August summary ended with ruling B just applied: one dispatch trigger firing every
thirty seconds instead of two stepping on each other, and a plan to widen the work batches
and fix a newly-found unfairness — the "queue-jumping" bug (413), where one old low-priority
item could freeze a site's place in the queue and starve every younger site. Batch widening
was prepared and approved but not applied; the starvation fix belonged to another session
and was queued behind our measurements.

## What we've done since

Everything on that list shipped, was measured, and held. The 24-hour reading passed its gate
and the batch widening went live on the 27th. The starvation fix went live the same day at
lunchtime, and its very first picks were the sites that had been stuck longest. We measured
it at two hours and again at three days: the worst wait for a site with hour-old work fell
from six-to-ten HOURS to about sixty-eight MINUTES, wasted claim attempts fell from roughly
60% before this workstream to 3.9%, and the entire backlog drained to nothing. The bug was
closed yesterday by its owning session with our measurements as the closing evidence.

Along the way the coordination itself became a story: four different sessions corrected each
other's claims — in both directions, usually within the hour — over a stuck-claim mechanism
that turned out to be bounded by a safety sweeper nobody had catalogued, over a meter that
went blind when claims were released, and over migration bookkeeping that seven of our own
hand-applied changes had quietly skipped. Every correction is written down where the error
was made.

The month's four AI-account blackouts (the last one two and a half days long) turned the
"spend governor" from a queued idea into the top build item. The owner ruled the shedding
order — routine maintenance pauses first, new site builds second, research last — and the
governor's foundations are now live and review-approved: a meter that prices every AI call
(August ran about $2,113, suppressed by the blackouts), the work classification, the
thresholds, and a heartbeat that recomputes every two minutes. It deliberately does nothing
yet.

## Where we are now

Dispatch is in the best state it has ever been measured: fair (no site starves), efficient
(3.9% wasted attempts), and fast enough that the queue runs dry. The governor watches but
cannot act: it needs one number from the owner — the monthly budget it should defend — and
one small piece of code (stage B) that makes workers respect the shedding level, which ships
switched off and goes through review before anything flips on.

## Where we're going

Stage B is the next build. After it: the option to quicken the trigger further (gated, by
design, on the governor being live), then the standing queue — batching deployments, the
paid-work priority lane, the half-price batch API for slow work, and a retention policy so
our own measurements stop evaporating after a day. The one open policy question for the
owner besides the budget number: whether individual low-priority items should get an age
ceiling of their own (the queue-jumping fix stops them starving other sites, but nothing yet
bounds their own wait).
