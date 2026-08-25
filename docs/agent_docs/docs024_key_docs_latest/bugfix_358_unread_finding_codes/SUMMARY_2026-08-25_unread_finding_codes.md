# SUMMARY — 2026-08-25 — unread finding codes (bugs_open/358)

*The milestone read-out, written to be read aloud. Previous: `SUMMARY_2026-08-24_...`. Current
state only; the chronology is in NOTES and README_where_we_are. This is likely the lane's LAST
summary — see the final section for why.*

## What we're trying to do

Stop the platform writing down things it notices and then never reading them. Dozens of detectors
record findings in one table; most of those records had no reader, and a cleanup job deletes them
after thirty days — so the system's memory of its own detections was a sliding window nobody looked
through. The fix was never a reader for everything: it is that **no finding code can exist without
a recorded decision about who reads it, or an honest statement that nobody does.**

## Where we've come from

Yesterday the machinery was done and proven — a registry of dispositions, a daily check on the
cluster's own clock, a commit-time scan that catches new codes before they ever fire — and what
remained was judgement: twenty-five codes still undecided, thirteen more written but never fired,
and a handful of design questions only the owner could settle.

## What we've done

The owner ruled on all seven open decisions in one sitting, and every ruling was applied the same
day. Twenty-four codes are now honestly recorded as "written deliberately, read by nobody,
accepted", each with its reason and the retention it accepts. One code turned out to have no writer
at all — its only writer was a database change never applied — and was retired rather than kept as
furniture. The four codes genuinely worth an automated reader were commissioned as work: three new
bug files with the evidence attached, each routed at the lane that owns the territory, and one
contribution into an existing open bug rather than a competing file. A naming clash that would have
turned the daily check red on some future stranger was renamed away, and a guard now fails any
commit that names a new code as an extension of an existing one — proven by the very clash it
retires. Checking for prior work before filing made every commissioned ask smaller: one "drift"
problem collapsed to a single item type; one "build rotation" ask turned out to be the unread half
of an already-closed bug.

The machinery also survived its first contact with strangers. Another workstream added three new
codes — the commit-time scan caught all three before any had fired, the one case the daily check
structurally cannot see. That same workstream held its own commit to avoid splitting ours
mid-flight; in return, our commit swept their registry entries along as the one kind of passenger
no discipline prevents — found by a reviewer, measured, attributed on both sides within the hour,
and left visible in the review record rather than tidied away.

## Where we are now

**The undecided backlog is zero** — thirty-two to zero in twelve days — and the cap sits at its
terminal zero, meaning any brand-new code is a finding the next morning unless declared in the same
commit. The daily check runs at 07:30 on its own schedule and writes its row, clean or not. Every
review this lane ever opened — six council correlations across three days — ended approved, and it
is worth saying plainly that across all of them the reviewers never found a defect in the code:
every rejection was the write-up failing to show what the tree contained, which is its own lesson
and is logged, five times, in the shared record of wrong calls.

## Where we're going

Nowhere, by design — this lane's work is complete, and that is the success condition. What remains
lives elsewhere and runs on its own: the commissioned readers get built by the lanes that own them,
and each code flips from "nobody reads this" to "consumed" the day its reader ships; the thirteen
never-fired codes get ruled if and when one first fires, which the morning check will announce; and
the daily row keeps the whole registry honest. An incoming session has work only if the check goes
red — and the worked example for that took two hours, start to green.
