# SUMMARY 2026-09-02 — the phase-lock fix confirmed on live traffic; the bug closes and the leftover goes to the file that owns it

Written by the `bugfix_410_feed_phase_lock` lane. All times DB clock. Previous summary in this
series: none — this is the lane's first, written at the milestone that ends it. The running record
is in `NOTES_410_feed_phase_lock.md`; the plain history is in `README_where_we_are.md`.

## What we're trying to do

Our news sites are meant to fetch fresh material every six hours. Twelve of the fourteen were
getting it every **twelve**, while every automated status in the system reported success. The job
was to find out why, fix it, and — the part that matters here — **watch it work rather than
conclude it works.**

## Where we've come from

The cause was a timing mismatch nobody would see by reading either half on its own. A refresh pass
runs on a fixed timetable, drifting only seconds. But when the pass actually fetches a source, it
writes "next due in six hours" counted **from the fetch**, and the fetch happens anywhere from ten
seconds to nine minutes after the timetable fired, because sites are processed one after another.
So a six-hour source becomes due a few seconds *after* the next six-hourly pass has already looked
at it and moved on. It waits another six hours. The lock re-arms itself every cycle, and every run
truthfully reports COMPLETED.

The fix, approved by the review council first time round, was to let the pass serve anything due
within **half a cadence** — three hours — so a source is served on the *nearest* timetable tick
rather than the first tick strictly after it. It had to go in twice, in the site-level admission
query (configuration, live on apply) and in the source-level dispatcher (Go, needing a release),
because the two layers had already drifted apart once before. The cadence is read live rather than
hardcoded, so the look-ahead follows if the timetable ever changes.

Both halves went live on the evening of 26 August. What was left was a single overnight check —
and then the lane went quiet for six days.

## What we've done

Picked it up cold, and found the written check could no longer be run: the records it needed are
kept for two days. **It was also wrong.** It asked that every site served in one pass be served
again in the next — but the fix's own predicted side effect, written four sections further down
the same document, was that demand would now exceed the ten-site cap and some sites would be
correctly turned away. A working fix would have failed its own test. That is logged as a wrong
call and filed as a landmine, because the shape is general: **relieving one bottleneck loads the
next one, and the test written beside the fix cannot see past it.**

So we replaced it with a test a cap cannot confound. A site cannot have been fetched *before* it
was dispatched, so its next due time is at least six hours after its previous dispatch. If the
next pass fires **before** that point and serves the site anyway, the old rule cannot explain it.

Three sites cleared that bar this morning — served by a pass that fired 2, 5 and 11 minutes before
they could possibly have been due. One of them is idea.uk, the site the bug was filed on, the one
we filmed being skipped by thirty-nine seconds. A fourth site looked like a pass and was **thrown
out**: its margin fell four seconds the wrong way, so it would have been served either way.

We also re-proved both halves are still installed rather than assuming it, and closed the one
remaining loose end in the reasoning (that a lagging source might explain the admissions — it
does not; one of the three sites has only a single source).

## Where we are now

**The bug is fixed, live, watched working, and closed.** The file has moved to `bugs_closed/` with
the evidence attached.

**The sites still are not on six hours — they are on about nine — and that is a different problem
with a different name.** Each pass takes only ten sites; there are now fourteen, and since the fix
all of them are genuinely due every time. Four get turned away per pass. They are not starved
(the longest-waiting go first, a fix we made in August, and it is working — the spread is one
pass), but the gap between designed and actual is now purely a capacity limit. We measured it
exactly, wrote it into `bugs_open/316` — the file already named for this queue being
oversubscribed — and **did not touch the cap**, because raising it roughly doubles how often we go
out and fetch news, and that is the owner's call.

Two open questions from the handoff are now answered rather than pending. The other six-hourly job,
the provocation feed, **does not have this flaw** — it picks the day's item from a pool and keeps
no per-source due time, so there is nothing for the timetable to fall out of step with. And a
pass on the evening of 1 September **hung**, served one site, and was killed after four hours,
costing thirteen sites a whole cycle. That is a known separate fault, but it is worth flagging
loudly because **it produces exactly the symptom we just spent a week eliminating.**

## Where we're going

Nothing is outstanding on this bug. What is left is other people's:

- **The owner's:** whether to lift the cap from ten to at least fourteen. Costed in `bugs_open/316`.
  If it is lifted, note the cap is a plain number in a query while the site count grows — today's
  right answer was twelve six days ago.
- **`bugs_open/316`'s:** it now holds the whole of the remaining shortfall, with the numbers.
- **Unowned and still true:** nothing checks on a schedule that the two layers of the due rule
  still agree. A future migration could quietly drop the look-ahead from the configuration half and
  the only signal would be sites going back to twelve hours. That is worth building; it is not
  worth holding this bug open for.
