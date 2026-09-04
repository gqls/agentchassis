# SUMMARY — provocation pipeline, 2026-09-04: it publishes on its own

*Previous in the series: `SUMMARY_2026-08-10_provocation_pipeline.md`. Written now because
the read-out has genuinely changed: on 2026-08-10 the generator worked and a person was
still in the publishing path. That is no longer true, and it has been observed rather than
argued.*

## What we are trying to do

Put a short, provocative piece of writing on vonc.com every day, written by the system
rather than by a person, in plain English that an ordinary reader understands on first
pass — without a human having to approve each one before it goes out.

## Where we have come from

The hard part was never generating text; it was trusting it. Three things had to be settled.

First, **whether a person signs off.** That question was answered three different ways in
five weeks: no approval step (31 July), then a required one (9 August), then none again
(2 September). Each reversal was a code change, which a reviewer has fairly noted is the
wrong place for a policy that keeps moving.

Second, **what replaces the person.** The judging model was measured and found genuinely
unreliable on this material — the same text drew different verdicts on different days. So
the floor is now arithmetic: a readability rule that counts sentence length and word length
and cannot drift. It was made *fatal* rather than advisory in the same change that removed
the human, and it was proven to actually refuse — by feeding it a real piece of our older,
denser writing and watching it come back rejected with its reasons.

Third, **what catches a bad piece now that nobody reads it first.** The answer is time: a
new piece is never dated for today, only tomorrow at the earliest. That gap is deliberate.
It is the window in which something can be pulled before a reader ever sees it.

## What we have done

The code is live and the pipeline drives itself on three timers: one writes more pieces
when the shelf runs low, one assigns dates, one publishes. The owner read the eight pieces
that existed on 2 September and approved them all. Two decisions were closed the same day:
the arithmetic rule is accepted as sufficient, and we are not building a configuration
switch for the approval step unless the policy moves a fourth time.

## Where we are now

**It published on its own on 3 and 4 September, and that was verified on the live page**,
by checking which piece is actually being served rather than by trusting a timestamp — a
distinction that matters, because the timestamp misleads by design here and briefly misled
us. Thursday's piece moved to the archive when Friday's took over, which is the archive rule
working without anything having to remember to do it.

There is a fortnight of runway. A fleet-wide rebuild ran on 4 September and was checked
against the one risk that mattered — that it might have been taken from before our change
and quietly restore the old approval rule — and it was not.

**Two things are honestly open.** The database column that records who read a piece could
not be written, because the session was not permitted to change live data; so it now
under-reports, and anybody reading it would wrongly conclude nothing has been reviewed. And
the older calibration exercise has not been re-run and must not be quoted as current — the
owner has ruled that acceptable, since nothing now depends on it.

## Where we are going

The interesting problem is no longer whether it works. It is that **a human review does not
stay done.** The owner read a set of eight; the machine wrote six more the next day and will
keep topping the shelf up. So the pile of unread writing rebuilds itself continuously, and
"nobody has read the prose" is not a task to close but a condition to manage.

That needs a decision rather than more engineering: read in batches when convenient, or ask
for a person only on borderline pieces, or accept publication unread and retire after the
fact. The buffer gives a day and the runway gives nine, so there is room to choose properly.
