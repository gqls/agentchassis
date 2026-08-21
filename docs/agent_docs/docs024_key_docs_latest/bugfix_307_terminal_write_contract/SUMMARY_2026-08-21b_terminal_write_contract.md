# Summary — the work-item failure-write contract (bug 307), 2026-08-21

*(First summary for this lane, written at its completion. The series rule applies from here:
any future summary is a NEW file, never an edit of this one.)*

## What we're trying to do

When a piece of queued work fails on this platform, the failure should be handled the same
way every time: the attempt gets counted, the item goes back in the queue with a wait before
the next try, a deliberate human decision on the item is never overwritten, and a passing
infrastructure blip — the whole fleet failing the same way at once — sends the item back
without spending one of its lives at all. That last part is your own ruling, given on 18
August after a three-hour GitHub outage killed a hundred pieces of work outright: "a
transient blip should return the item to queued."

## Where we've come from

Before this work there was no such contract. Two different pieces of code wrote "this item
failed", and they disagreed: one counted attempts but retried instantly — three tries inside
minutes, all into the same dead dependency — and the other counted nothing at all, so for
five kinds of work one failure was permanent, on a normal day, with nothing wrong. Measured
properly that quiet version was the bigger problem: nearly three-quarters of all failed items
died before using their allowance. Neither writer respected a decision a human had recorded.
The outage was the alarm; the daily bleed was the fire.

## What we've done

One shared piece of code now handles every failure the same way, and it shipped through two
review-council rounds — the first of which caught a real defect (a guard list reused where it
must differ) before it ever ran.

Then the part that earned its keep: rather than trusting the tests, we fed the live system a
single deliberately doomed test item and watched what actually happened. That drill found two
real faults the fifteen tests could never see. First, two seconds after the new machinery
correctly re-queued a failed item, an unrelated bit of housekeeping stamped it "complete" —
recording failed work as done and cancelling the retry; a real page on the mortgage
calculator site hit this within hours. Second, when an item genuinely ran out of lives, the
write that should mark it permanently failed crashed on a one-character database technicality
our test stand-in doesn't check — so nothing could ever run out of lives; doomed items just
cycled for ever. Both were fixed the same day (the second by this session, approved by the
council first time; the first by the sister session, on your ruling), and both fixes rode
this afternoon's deploy.

The drill was then run again, end to end, and passed everything: the failed item waited half
an hour, then an hour, between tries; nothing stamped it complete; a "we are not doing this
one" decision made mid-run was left standing, with the system's own log saying so; and the
third failure was marked permanently failed, correctly. The test row was deleted. Bug 307 is
closed and moved to the closed pile, alongside the other three bugs this check-up covered
(301 and 313 were already properly closed; 317 needed only its final move).

## Where we are now

The contract is live on the whole fleet and proven on both real traffic and the drill. Two
genuine infrastructure blips on the evening of the 20th were handled exactly as you asked —
requeued without losing a life, retried after the wait, finished successfully. Since the
final deploy, not one item has been falsely completed. The one thing that cannot be proven on
demand — a real outage leaving nothing dead behind — is written into the closed bug as a
standing watch with an explicit "reopen if" condition, so it cannot quietly become folklore.

## Where we're going

Three smaller pieces remain, each with a named owner, none blocking anything closed here. The
sister session is finishing the database-side half of the completion fix (bug 344) — the
hourly cleanup job can still, in one narrow case, do what the main path no longer can — and
its own cooldown work for that same job (bug 341), which is deliberately held back until 344
lands. A pre-existing question about how "park this for a human" failures count their
attempts (bug 033) stays with its own earlier ruling. And the next real outage is the
standing watch's moment: if it leaves anything dead, 307 reopens; if it leaves nothing, the
ruling that started all this is fully honoured.

The lesson worth carrying beyond this lane is written into the debugging guide: a fix that
changes what a failure writes must audit every *other* writer of the same record, and no test
stand-in replaces one deliberately doomed item run through the real machinery — ninety
seconds of drill found what fifteen tests missed.

> **CORRECTED 2026-08-21, later the same evening:** "Where we're going" says the sister
> session is finishing "the database-side half of the completion fix". Their measurement,
> hours after this was written, showed that half is **unnecessary by construction** — the
> cleanup job only ever touches items in the "claimed" state, and an item put back in the
> queue by the new machinery is never in that state, so the narrow case this summary
> described cannot occur (zero instances measured). What bug 344 still awaits is its review
> verdict, nothing more. The cooldown work (bug 341) is meanwhile applied and live, not
> pending. The error was this summary's author's, made by analogy and corrected by the
> sister session's check.
