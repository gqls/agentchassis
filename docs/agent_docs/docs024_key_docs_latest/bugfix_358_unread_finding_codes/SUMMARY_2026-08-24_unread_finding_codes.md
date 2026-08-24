# SUMMARY — 2026-08-24 — unread finding codes (bugs_open/358)

*The milestone read-out, written to be read aloud. Previous: `SUMMARY_2026-08-22_...`. Current
state only; the chronology is in NOTES and README_where_we_are.*

## What we're trying to do

Stop the platform writing down things it notices and then never reading them. Dozens of detectors
record findings in one table; most of those records had no reader, and a cleanup job deletes them
after thirty days — so the system's memory of its own detections was a sliding window nobody looked
through. The fix is not a reader for everything: it is that **no finding code can exist without a
recorded decision about who reads it, or an honest statement that nobody does.**

## Where we've come from

Two days ago we had a registry of every code with a declared disposition, a checker that grades the
registry against the live table, and a council-approved design — but it all ran only when a person
remembered to run it, which is the very failure it exists to catch.

## What we've done

Put it on a clock, and watched the clock work. The check now runs daily in the cluster at 07:30 UTC,
writes one record per run including clean ones (so silence means "did not run", never "nothing
wrong"), and states in every record which of its questions it asked and which it deliberately left
to commit time. The owner ratified the first batch of rulings — seven codes now honestly recorded as
"written deliberately, read by nobody, accepted" — and the undecided backlog is capped so it can
shrink but never grow.

On its first live day it caught two brand-new codes arriving unread, one of them recording a real
degradation: two pages published with no internal links after a database timeout, noted by the
system, read by nothing. Chasing how they got past our own safeguard exposed that the safeguard was
a claim, not a thing — a comment naming a test file that had never been written. That test now
exists and works by reading the code rather than by being told a list, and the false claim is
retracted where it was made.

## Where we are now

The loop is closed and proven end to end: a new unread code turns the check red within a day; a
declaration turns it green; the next release carries the declaration into the cluster. Live figures
today: 43 codes observed, 55 declared, 0 findings, 25 undecided (at the cap), plus 13 codes that
exist in source but have never fired, held on their own shrinking list. Everything is
council-approved, and both approvals came only after the reviewers made this work show — rather than
assert — what the tree contained, which cost four rounds and taught a discipline now written into
the runbook.

## Where we're going

The remaining work is judgement, not construction: the owner rules the 25 undecided codes batch by
batch, the same for the 13 unfired ones, and each ruling lowers the cap. Separately, the degradation
the check surfaced on day one belongs to another workstream and is the strongest candidate in the
registry for a real automated reader. This lane's machinery is done; from here its output is the
backlog number going down.
