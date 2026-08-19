# README — where we are (bugfix 313/298, the internal linker)

Owner's plain-prose log. Append-only, newest at the bottom.

## 2026-08-19 — picked up, still real, plan formed

You asked this session to fix bugs 313 and 298 — the internal linker. The short version of the
two bugs: the agent that is supposed to add internal links between a site's pages has a broken
"are there any candidate pages?" test. The step that fetches candidates returns a plain list, but
the test asks the list for a named property (`count`) that plain lists don't have. The test can
never pass, so for four months every run has ended with "no candidates" — including a run that
was holding fifteen candidates at the time. The agent has completed 57 link jobs and never once
produced a link. Bug 298 is about the same fetch: it only takes the first 15 pages alphabetically,
which never mattered because of 313, but starts mattering the moment 313 is fixed (8 of 26 sites
have more than 15 linkable pages today).

What I checked before doing anything: nobody else is working on this (the lane that found it
closed yesterday and explicitly left these tickets unowned), and the bug is still there right now
— I re-read the live agent configuration and the code this morning rather than trusting the
two-day-old bug file. It is unchanged, and the "has it ever made a link" counter is still zero.

The plan, in one breath: one config change (live immediately, no build needed) fixes both bugs —
the fetch starts returning its result in the shape the test expects, the 15-page cap comes off
(the per-page text is already capped, so the payload stays modest), and the prompt is updated to
match the new shape (miss that and you trade a dead branch for a garbled prompt). Then two small
platform pieces so this class of bug can't quietly happen again: an offline checker that scans
every live agent for "a test asking a list for a named property" (this bug's exact shape — today
the fleet has exactly one, this one), and an opt-in switch on the test step itself so that a
comparison that cannot evaluate fails loudly instead of silently picking the "no" branch. The
switch is off everywhere by default — only the fixed step opts in, as a tripwire against this
ever drifting back.

Next: write the migration and the Go pieces, put the whole thing through the council, commit,
apply, and then watch for the first real link plan in the logs — the agent runs a couple of times
a day naturally, so the proof should arrive on its own.
