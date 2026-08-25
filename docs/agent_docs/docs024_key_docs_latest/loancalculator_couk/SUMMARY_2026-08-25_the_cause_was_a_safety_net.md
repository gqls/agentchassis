# Where the loancalculator.co.uk work stands — 2026-08-25, evening

**What we're trying to do.** Run loancalculator.co.uk entirely through the framework —
twelve working calculators, each protected by a lock so automation cannot rewrite them —
and, since Saturday, answer one specific question: why did an owner-released rebuild
duplicate the calculator on one page out of ten, when the other nine came through
cleanly?

**Where we've come from.** Saturday's rebuild wave left `/tools/loan-vs-savings.html`
with two copies of its calculator — the protected one and a byte-identical, unlinked
twin — and the twin then blocked every automatic repair of that page. On Sunday the
damage was repaired by hand and verified at the served page, and the acceptance harness
was re-baselined. Monday established what the fault was *not*: it had not recurred
anywhere, the everyday re-render route was proven clean on all ten protected pages, and
a promising-looking lead in the history table was measured and dismissed. But the cause
was still unknown, and the code that produced it was still running.

**What we've done.** Today the cause was found and pinned, with the actual lines of code
and a measurement at every step. The culprit is a safety net inside the save machinery
whose job is to stop a rebuild losing an interactive tool: it checks whether the new
content still contains the stored tool, and if not, it adds the stored copy back. The
check compares names — and the two sides name the same calculator differently. The
stored copy is filed under its shelf position ("tool-2"); the rebuild names it by what
it is ("tool-loan-vs-savings"). No match, so the safety net "rescued" a calculator that
was never in danger, onto a page that already had it. Why only this page? The safety net
only examines stored rows carrying one particular status label, and exactly one
protected calculator in the entire fleet carries it — this one. The nine clean siblings
were never examined at all, which is why they proved nothing. The finding, the evidence,
a fleet-wide census of what is still exposed, and a ranked fix are all written into the
bug file; the trap is recorded where future sessions will trip over it, and the
transferable lesson — the same matching question was fixed in two places and forgotten
in a third — is in the debugging guide.

**Where we are now.** The site is healthy: 28 pages serving, every calculator proving
its own arithmetic against the golden baseline, locks intact. The bug is understood but
not yet fixed: the flawed matcher is live in the current build, and the one exposed row
is still exposed. It cannot fire by accident on the everyday route — only a full
rebuild of this one page can trip it, and nothing queues one automatically today.

**Where we're going.** Two steps close it. The code fix is small and copies a pattern
already used twice in the same file — teach the safety net to recognise a tool by its
identity, not just its name — and goes through the usual council review. Separately
there is a one-line data change that would disarm the single exposed row today, but it
touches a row the owner locked, so it is his call, and it has been put to him in the
running log.
