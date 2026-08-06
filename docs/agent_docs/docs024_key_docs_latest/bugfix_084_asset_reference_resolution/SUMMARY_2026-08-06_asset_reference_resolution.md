# Summary — 2026-08-06 — proven, closed, and the one sentence worth keeping

## What we're trying to do

Close the blind spot in `bugs_open/084`: a live page could tell a browser to load
a JavaScript file or stylesheet, the file could be missing, and nothing we owned
would notice. The page renders, every status says deployed, and it simply doesn't
work when someone clicks it.

## Where we've come from

Yesterday's summary ended with the work built, reviewed, committed — and
deliberately switched off, because switching on a check the running software
doesn't recognise fails the entire scan. Two of the bug's five suggested fixes had
already been closed off by research: one was ruled an architecture-level change
months ago, and another had quietly been done by a different team on 29 July
without anyone updating the bug.

## What we've done

The build went out, we confirmed the running software contained the check (with a
deliberately-wrong control string that it didn't, to prove we were testing the
pipeline rather than our own spelling), and switched it on.

Then we proved it works — which was the whole difficulty, because there is nothing
broken on the fleet for it to find. A clean run would have been worthless: a
correct check and a silently broken one produce identical output when there is
nothing to detect.

Two things made the proof cheap and safe. First, the documented way to run a check
fires the full improvement loop, which dispatches automated fixers at a live
customer site; reading the definitions showed the discovery agent alone is three
steps and dispatches nothing, so we aimed at that instead. No site was edited.
Second, we induced a fault deliberately: quietest page on the quietest site,
control taken first, a checksummed backup, and both the break and the repair
wrapped in database guards that would abort rather than leave the row altered.

It caught it — one alert, correct in every field — and, equally important, raised
**nothing else** across the site's other eight pages. Then we restored the page,
verified by checksum, and re-ran: the alert stayed open, exactly as our own
documentation predicted, because the check only withdraws an alert when it can see
the thing working again. A documented limitation turned out to be true rather than
a guess. The test alert was cancelled with its provenance recorded.

## Where we are now

Closed and moved to `bugs_closed`. The check is live, enabled, council-approved
first time, and has been observed catching a real fault and clearing up after
itself. The transferable lesson is written into the debugging guide, and two
mistakes of our own are in the wrong-calls log — both about trusting a blank
result, which is the same failure the check itself exists to prevent.

## Where we're going

Three things outlive this lane and none of them belong to it: an architecture
proposal is owed for the shared check we deliberately did not touch; one remaining
suggestion already lives on another team's roadmap; and the last overlaps two open
bugs about rewriters losing content. Each now points at where it actually lives
rather than being carried by a closed file.

One reviewer objection stands unanswered on purpose: we are creating alert types
that nobody is assigned to drain faster than we drain them, and whether that
remains acceptable is a judgement for a person, not something to argue away.

**The sentence worth keeping:** this check reporting nothing is not evidence that
it works. There is a two-minute procedure for proving it still bites, and it
should be run before anyone quotes a clean result.
