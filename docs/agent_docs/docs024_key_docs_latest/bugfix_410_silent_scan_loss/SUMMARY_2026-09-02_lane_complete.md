# SUMMARY 2026-09-02 — the silent-scan-loss lane is complete

**What we're trying to do.** Pages are rebuilt by reading their stored sections out of the
database, re-rendering them, and saving the result back — and the save replaces everything
wholesale. So if the read quietly loses something, the loss becomes permanent: the page ships
with a hole or an emptied section, looks freshly built, and reports success. This lane's job was
to make that impossible at the worked site, and expensive to reintroduce anywhere else.

**Where we've come from.** The bug (filed 2026-08-26 as one of three sibling seams that all fail
toward the quiet default) was that a row which failed to read was simply skipped with a log line
nobody watches. We shipped a guard that counts what the database offered against what survived
and refuses the whole load on any loss; a codebase-wide ratchet that stops new code re-growing
the pattern (207 known sites, counts only ever allowed to fall); and we left one stated loose
end: a section whose *content* failed to decode was kept but emptied — invisible to any count.

**What we've done since.** Five days on, a fresh session re-verified every claim (all held —
including the census, still exactly 207 despite twenty new files), and then closed the loose
end. The open question — "may an undecodable section render as an empty one?" — was answered by
measurement rather than debate: the column type cannot store broken JSON, and zero of 2,751 live
values have the one shape that could still fail, so refusing costs nothing today and protects
against the first future mistake. The 55 pages that legitimately have no content stay loadable,
pinned by a test. The review council approved it unanimously, first round.

**Where we are now.** Everything this lane owns is fixed AND live, verified in the running
binaries after the 2026-09-02 deploy, with real traffic behind the evidence: since the start of
the month roughly 1,400 rerender runs, 176 through the guarded code path, zero refusals — the
guard is exercised and silent because the reads genuinely complete.

**Where we're going.** Nothing further in this lane. The bug file stays open for one item that
was always someone else's to pick up: the fleet-wide "a value nobody understands should refuse,
not quietly proceed" design round (candidate 1), which now overlaps a question the 404 lane has
also posed. The owner has a choice to make: spin that candidate into its own file and close the
pattern file, or keep the file open as its tracker. The lane recommends the first.
