# SUMMARY 2026-08-06 — the token-cap watchdog

## What we're trying to do

Stop a whole class of failure that has bitten this platform repeatedly and always
looks like something else. Every step where we ask an AI to write something carries a
maximum output length. When a step's output grows past that limit the step fails — and
because the limit is a setting rather than a property of the site being built, the
failure presents as *a problem with that site*. Someone then investigates the site.

## Where we've come from

Bug 183 is one instance: the step that works out what a newly adopted domain is about
had a limit set long ago and never revisited. It sat at 92.5% of that limit for
months, then the tail crossed and it failed on several unrelated sites in one
afternoon, using up all three retries on each and leaving them with no strategic
profile at all. The limit has since been raised twice, so the incident is over.

But this has happened at least six times before under different bug numbers, and the
platform's own source code carries a warning about it: raising a limit *moves* the
cliff, it does not remove it. A week ago another lane built a monitor for exactly this
— but only for the sixteen council review seats. The other hundred-odd steps in the
fleet had nothing watching them.

## What we've done

Widened that monitor to the whole fleet. `fleet-step-token-pressure` runs every six
hours, costs nothing (one database query — no AI is involved, no orchestration, no
credits), and reports any step running close to its ceiling or already hitting it. It
writes only when the situation *changes*, so it will not become an alert people learn
to ignore. The two monitors now partition the fleet exactly and share their
thresholds.

It is verified in the way that actually discriminates: the query was executed verbatim
out of the live scheduled row, the platform's scheduler then ran it on its own, and
the deduplication was proven by re-running it and getting nothing. Most importantly it
was tested with the clock **pinned to the day before bug 183's failure** — where it
raises the alarm a full day early. The first version I wrote would have been silent
until after the damage; that was only visible because I asked what it would have said
*then* rather than what it says now.

## Where we are now

Live, running, and it earned its place on the first run by finding a bug nobody had
filed — now `bugs_open/205`. A step in the vet-practice verification pipeline has been
failing **100% of its calls for 34 hours**. Two things are wrong: it has **no length
limit configured at all**, so it silently inherits a built-in fallback of 2048, the
smallest number in the estate; and the failures are the *same two records* being
re-picked by a batch job every five minutes, failing, and therefore never being marked
done — a loop that would have run indefinitely.

I have not changed either. Raising a limit on another lane's pipeline has always been
an owner call here, and the loop's mechanism has one step I have not read, which is
marked as inferred rather than asserted. The census that bounds the first problem *is*
done: 8 of 126 steps configure no limit anywhere, and six of those are dormant — traps
waiting for their first real workload.

## Where we're going

Bug 183 is one decision from closing, and the decision is the owner's, not an
investigation: its remaining idea is to split that step into four smaller pieces so it
cannot outgrow its limit at all. That is a structural change to a pipeline other lanes
are actively using; with five times the headroom and growth now monitored, waiting is
safe. If the structural version is still wanted, 183 stays open and should say so.

Bug 205 wants two things from someone: a decision on the vet step's limit, and a read
of the batch job's selection logic to confirm the loop. The loop is the urgent half —
it is spending credits every five minutes to fail identically.

The wider question 205 raises is whether a workflow step should be *allowed* to
inherit an unstated default at all, or whether reaching the fallback should be loud.
That is the fix that generalises, and it is written up as 205's first candidate.
