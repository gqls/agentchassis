# SUMMARY — bug 046 CLOSED: all nine casualties repaired, permanent guard live (2026-07-24, evening)

## What we're trying to do
Nine interactive components on our sites were built from half-finished AI
generations — JavaScript cut off mid-stream — and the platform had no way to see
the damage. The job: repair all of it, and leave the platform able to catch any
recurrence by itself.

## Where we've come from
The cause was fixed long ago, but nobody swept for the wreckage; the census that
found the nine only happened as a side-effect of another bug's calibration run.
By this morning we had a proven detector and one repaired tool; by this
afternoon all six components sitting on live pages were repaired and verified
against live bytes; three orphaned components (on no page, serving nobody)
remained, pending a decision.

## What we've done
The owner chose rebuild over deletion for the orphans, so all three were
regenerated through the same guarded pipeline (adapted to skip the page lookup
they don't have), checked for completeness and invented data, and version-
snapshotted. That was the last of the nine.

## Where we are now
**Bug 046 is closed, by its own written test:** the fleet-wide census query
returns zero damaged components, and the two live URLs named in the original
filing both serve fully balanced JavaScript, verified against tonight's actual
bytes. Every one of the nine casualties is repaired. The `truncated_component`
detector and its completion-time verifier remain live in production as the
permanent guard, and the transferable lesson — fixing a corruption's cause does
not repair its casualties; a corruption-class close needs a census — is recorded
in the shared debugging guide.

## Where we're going
Nothing further on this bug. Two reusable assets outlive it: the detector (runs
on every completeness discovery pass) and the scripted repair recipes
(restore-and-deliver, rebuild-and-deliver, and the orphan variant) in the
workstream's scripts directory, ready for the next corruption class that needs
them.
