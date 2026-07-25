# Summary — bugfix 003 spawn loss — 2026-07-25

**What we're trying to do.** Stop the platform silently losing the messages
that connect a parent job to the child agents it spawns. When one is lost, the
parent sits idle for thirty to ninety minutes until a cleanup sweep kills it —
work vanishes, and nobody is told.

**Where we've come from.** The bug was filed on the 15th of July when a site
rebuild kept failing on it. Investigation found three separate causes: a flaky
network path (split into its own case), a message-handling loop that told the
queue "done" before the work was actually done, and a retry system that lived
only in the memory of a single process — restart it and every pending retry
evaporated. The first protective fix (a cleanup sweep) went live on the 20th
along with an honest health check; the two big fixes — hold the "done"
acknowledgement until the work truly finishes, and move the retry list into
the database where any server can pick it up — were designed then but not
built.

**What we've done.** Built both big fixes in full, with the database schema
applied first so old and new code could coexist safely. Along the way we found
and fixed three subtle defects that would have made retries useless even after
shipping. The advisory review council rejected the change twice — not on
correctness (it conceded that in round two) but on shape: a rewrite of core
delivery plumbing, it said, deserves an architecture review, which didn't
exist. The owner agreed to follow that cautious path — but before we could,
another session's routine deployment carried our committed code to production.
Four and a half hours of live running showed it working: nineteen automatic
retries, seven jobs saved that would have been lost, nothing stuck, nothing
crashed. Shown the changed facts, the owner ruled: keep it. We also built the
architecture-review track the council asked for, and wrote this redesign up as
its first entry.

**Where we are now.** The new delivery guarantees are live across the whole
fleet and measurably healing losses. The database schema, the config
improvements and the graceful-shutdown settings are all live. The review
council's verdict trail, the wrong calls we made and caught, and the honest
after-the-fact review record are all written down. No review stamp was claimed
— the council never approved.

**Where we're going.** Three pieces of proof remain before this closes: a
deliberate kill-the-server-mid-job test under controlled conditions, a re-run
of the health-check restart test from the 20th, and the weekly numbers around
the 1st of August confirming the ninety-minute strandings have stopped. After
that: make sure every message carries the identifier the de-duplication needs
(absences are now loud but not fixed), and the separate question of running
more than one chassis server, which is blocked on a known race.
