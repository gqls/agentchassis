# SUMMARY — the disk-pressure lane, 2026-08-15: closed, with the proof

Written to be read aloud. Yesterday's summary said the fix was "built and awaiting
one release"; that release has now run, the effect is measured, and the case is
closed. This entry exists because "where we are now" has genuinely changed.

## What we're trying to do

Keep the cluster's nodes far enough from "disk full" that they never again refuse
a pod during a deploy, and make that margin maintain itself rather than depend on
anyone watching it.

## Where we've come from

An 11 August deploy had a pod turned away by a node that had just crossed its
disk threshold. The investigation found two compounding defects: the scheduler
had never been told disk exists (nothing declared a disk request), and the image
cleanup trigger sat exactly on the refusal line, so the system could not start
making space until it was already turning work away. Along the way we learned the
disk breathes — climbs, reclaims, climbs — and that our hosting provider silently
ignores edits to the cluster-wide kubelet settings object, which forced the fix
to travel as a small per-node component instead.

## What we've done

Everything is now live and verified on the running cluster. Disk requests are
declared and the two heavy CI runners are held on separate nodes by rule. The
cleanup thresholds are changed on every node — cleanup now starts at 70% full
instead of 85%, and images unused for a week are retired continuously. All of it
ships automatically: with every release, and with a fresh install (the terraform
path now carries the node component, the CI runners and both ollama services —
including one, ollama-eval, that had previously been applied by hand and tracked
by nothing).

## Where we are now

The proof arrived with this week's release. Four of five nodes reclaimed
themselves down to the new target within hours and now cycle six to nine
gigabytes clear of the refusal line — the same nodes that were sitting at
six-tenths of a gigabyte three days ago. Roughly a tenfold improvement in the
worst-case margin, with no disruption: no pod restarts, no refusals, nothing
turned away. The fifth node sits at two gigabytes clear — stable, three times
the old worst case, but bounded by what it is running rather than by cleanup
timing: its disk is mostly in-use images plus the system journal, which no image
cleanup can touch. The bug is closed on this evidence.

## Where we're going

One decision remains open, and the fifth node is the argument for it: capping the
system journal returns about 3.4 GB per node permanently, at the cost of keeping
roughly ten days of operating-system logs instead of seventy. The machinery to
apply it fleet-wide already exists — it is one small addition to the node
component the moment the trade is accepted. Beyond that, the lane asks nothing:
per-pod disk ceilings and bigger disks stay parked, and the margins are now a
routine read rather than a worry.
