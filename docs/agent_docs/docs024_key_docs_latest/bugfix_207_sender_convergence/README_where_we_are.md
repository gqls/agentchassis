# Where we are — sender convergence (bugs 207, 216, 217)

Plain-prose running log. Append below; never rewrite or reorder.

## 2026-08-08 (night)

The whole series is done and proven in production. In plain terms: when something
fails mid-workflow, the platform used to have three different voices reporting the
failure, and the loudest, fastest one always said "give up" no matter what went
wrong. Tonight all three voices consult the same judgement, and that judgement was
tested live: we broke a workflow on purpose with a timeout, watched the "worth
retrying" verdict go out, and watched the actual retry arrive on the queue. About
half of these failures — thousands a fortnight — were being abandoned when a retry
would plausibly have saved them; those now get up to three attempts before the
platform gives up for real.

One thing to keep an eye on, written into the bug file as a weekly check: a retry
that keeps failing at one level can now prompt the level above to have its own go.
The maths bounds this well and we saw no sign of trouble tonight, but it is the one
genuinely new behaviour, so it gets watched rather than assumed.
