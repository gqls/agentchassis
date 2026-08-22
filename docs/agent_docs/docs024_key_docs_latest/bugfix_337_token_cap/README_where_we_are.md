# Where we are — bug 337 (the tool section that always blows its token limit)

Append-only, newest at the bottom. Plain prose for the owner.

## 2026-08-22 — picked up the bug, confirmed it is still real, measured how big the problem actually is

Bug 337 is about one kind of page section — the "credit health check" interactive tool —
that the AI writer can never finish. The writer is allowed to produce at most 16,000
tokens of output per attempt (roughly 48,000 characters), and this particular section
needs more than that, every single time. The system correctly refuses to keep a
half-finished component (that safety rule exists for good reasons), retries twice more,
hits the same wall twice more, and gives up. Result: two live loan-industry pages have
no working calculator on them, and every future loans site that plans this section will
lose the same page the same way.

What I checked today:

- **The bug is still live.** The limit is still 16,000, the failed work items are still
  parked (plus a fourth occurrence the bug file didn't know about), and nobody else is
  working on it — the team that filed it deliberately left it for someone to take.
- **The limit is too tight even when it works.** Looking at the last two weeks of
  successful runs of this writer: a twentieth of them come within 15% of the ceiling.
  So this isn't one greedy section — the ceiling was set too low for the job generally.
- **Other parts of the system have the same disease.** Several other AI steps run within
  10% of their own ceilings, and a handful have hit them outright over the past month.
  Whatever we do here should help all of them, which matches the direction you suggested:
  raise the limit, and put proper management around these limits rather than picking
  numbers by folklore.
- **A bigger limit is safe with this AI model.** Other steps in our own system already
  run the same model with limits of 32,000 and 64,000 without trouble, including another
  step that writes tool HTML and was given 32,000 from the start.

The plan I'm preparing (details in the PLAN file) has three parts: raise this step's
limit to a measured number; teach the framework to notice "the answer was cut off" and
retry once with a higher ceiling instead of failing three times identically; and add a
small daily check that watches every AI step's real output sizes against its ceiling, so
we find out a limit is getting tight before pages start failing, not after.
