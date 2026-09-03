# SUMMARY — dispatch throughput — 2026-09-03

*(Third summary in the series. The first, 2026-08-26, ended with the dispatch trigger
rationalised; the second, 2026-09-02, ended with the spend governor built but unable to act.
This one is written because the thing it was waiting for happened. To be read aloud.)*

## What we're trying to do

Make the machine that hands out website work fast enough, fair enough, and cheap enough to
run hundreds of sites instead of dozens — without ever letting one site, one stuck job, or
one runaway bill stop everything else.

## Where we've come from

The unfairness and the waste were fixed first, and have held: no site starves, the worst wait
for a site with hour-old work fell from six-to-ten hours to about an hour, wasted attempts
fell from roughly 60% to under 4%, and the queue now regularly drains to empty. What remained
was the money. Four times in three weeks the whole fleet stopped dead because the AI account
hit a wall — once for two and a half days — and each time the machinery kept trying, learned
nothing, and waited for a human to notice. That is what the spend governor was built to
replace: an accidental, total, silent stop, swapped for a deliberate, partial, announced one.

## What we've done since yesterday

The owner gave the two things only they could give: a budget of $2,000 a month, and then the
word "enable". The governor went live at 10:14 this morning. Between those two moments the
last code shipped, the held configuration was applied by hand, the check that guards the
dispatch selector was updated in the same sitting so it would not read the change as damage,
and a ten-minute canary confirmed that with the governor switched off the system behaved
byte-for-byte as before. This morning the owner also raised the AI account's own cap to
$3,000 — which matters more than it sounds: it puts the account's hard wall a thousand
dollars above the budget the governor defends, so the gentle brake now always arrives before
the hard one. That was the last outstanding dependency of the whole design.

Then we watched it. Half an hour either side of the switch, dispatch is unchanged: same
wake-ups, more sites served, no failures, nothing refused — which is exactly right at $386
of $2,000. And we proved the staged slow-down works at every level **without pausing a single
real piece of work**, by pretending the budget had been crossed inside a database transaction
that is then discarded. At the first level the governor would hold back 51 routine
maintenance jobs; at the second, 112, the extra 61 being new page builds; work that needs no
AI keeps flowing at every level. That is precisely the order the owner ruled, and it is now
measured rather than argued.

## Where we are now

The governor is live, watching every piece of AI-bearing work, and saying yes to all of it —
because September's spend is well under the first line. At the current burn (~$124/day, hotter
than last week's estimate) the first real slow-down would arrive around the 11th. Two honest
limits are on the record: the worker code that reads the shedding level has only ever been
exercised at "allow everything", so the last link in the chain is proven present but not
proven firing; and a release rewrites all 200-odd agent configurations about a minute before
new pods start, which did not disturb our hand-applied setting this morning but could, with
nothing reporting it — so that check is now a written after-every-release habit.

## Where we're going

Either we wait for the 11th, or we induce a five-minute slow-down deliberately and watch the
last link fire — the owner's choice, and the only thing standing between us and the next
build item, which is quickening the dispatch trigger (deliberately gated on the governor
being proven, so that going faster can never mean spending faster). Behind that: batching
deployments, the paid-work-first lane, the half-price batch interface for slow work, and a
retention policy so our own measurements stop evaporating after a day.
