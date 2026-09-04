# SUMMARY 2026-09-04b — round 3: one ladder, and the four places a limit can be written

*Written to be read aloud. Supersedes nothing — the series is the record.*

## What we are trying to do

Make sure that when someone sets a limit on how long a model's answer may be, the system actually uses
the number they set. That sounds like it should be free. It has not been.

## Where we have come from

The bug is a year of the same shape arriving in different clothes. First, only one function in the whole
platform read the setting, so any code that talked to a model directly ignored it and used a small
hardcoded number instead. We fixed that in August by moving the lookup down into the model clients
themselves, so a caller that says nothing now inherits the configured limit.

Then in September we found the fix had been defeated from above: two pieces of code written *after* it
had each re-implemented the rule by hand and each ended in a hardcoded 2000. A hardcoded number is worse
than no number at all, because an explicitly supplied limit wins at the wire — so those callers could
never inherit anything. We deleted every hand-rolled copy and put a test in place that fails the build if
a sixth one appears.

Both of those were about the code side: who reads the setting. This round is about the other side.

## What we have done

We stopped looking for the setting in the places we expected and walked every agent's entire
configuration looking for it anywhere. There are 171 of these limits live across the fleet, and they are
written in five different places. The program was reading two of them — and it read them in the wrong
order, taking the most general one first.

That gave two faults. Ten agents had a leftover limit at the top of their configuration that overrode the
limit written on the step itself, holding an 8000-token step down to as little as 500. And four settings
on the site-adoption agent sat one line outside the block they belonged in, right beside other settings
that *are* read, so nothing looked at them at all — one of them asking for double what it received.

The fix has two halves that do not wait for each other. The program now looks in every place, in a stated
order from most specific to most general, so wherever an operator writes the limit it counts. And a
database change moved fourteen misplaced settings into the one proper place; that half took effect the
moment it applied, so both faults are already gone in production without waiting for new code to ship.

We also built a report that answers "where is each step's limit actually coming from?" against the live
system, because this is a fault no test can catch — the code is fine, the settings are fine, and they
simply do not meet.

## Where we are now

Both faults are fixed. The report is quiet: one step in the whole fleet genuinely has no limit set and is
running on the smallest default we have, and seven are written in the untidy place but do work.

The most useful thing that happened is that the new report found a mistake in our own change. Its first
run flagged eighteen steps as broken, and all eighteen were perfectly healthy — we had mistaken a normal
pattern, a fleet-wide default that individual steps override, for a fault. The same mistake was already
sitting in a warning we had added to the running code an hour earlier, where it would have fired on every
call those eighteen steps made. Both are fixed. The lesson is not that we made a mistake; it is that a
warning which cries on the healthy majority is one nobody reads twice, and nothing but live data would
have told us.

A second mistake was caught the same way. Running the database change as a rehearsal that deliberately
throws itself away revealed that PostgreSQL reads one of our expressions differently from how it looks —
it binds subtraction more tightly than the key lookup beside it. The migration runner's own safety check
could not have caught this: it had skipped the file because of the word "Rollback" in a comment.

Three settings on the HTML builder are deliberately untouched. They are the one place where the "wrong"
spelling is currently the only one that works, so moving them today would quietly break them. There is a
prepared change for those, applied by hand once the new code is live, with the "is it live yet" test
written into the file as a command.

## Where we are going

Three things, none urgent.

The new code is committed but does not take effect until the next time the platform's images are
rebuilt and rolled out. Until then the database half is doing the work on its own, which is by design.
Once it is live, the held change for the HTML builder gets applied and this bug can close.

There is a live experiment running on two agents to prove, by observation rather than inference, that
last week's fix is really in the pods — each puts a slightly different number in a place only the new
code looks at. Both agents work in bursts and both went quiet, so we are waiting for traffic. Both get
put back the moment we have a reading.

And the question of making direct model calls visible to our monitoring is now its own piece of work,
bug 480, with its own folder. Six places in the code talk to a model without writing the line that all
our truncation monitoring reads, so a missing line there is indistinguishable from a step that never ran.
Nothing is on fire; what it costs is that "is anything being cut off?" can only ever be answered "not in
the part we can see".
