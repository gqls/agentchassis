# README — where we are (bugs_open/297)

Plain prose, append-only, newest at the bottom.

## 2026-08-17 — picked up, checked, and the fix turned out simpler than its sibling

This bug is the second of three "silent row cap" bugs the 275 work uncovered. When the system
rebuilds an interactive tool (a calculator, a game) on one of our sites, it first asks the
database "what other pages does this site have?" so the AI understands how the tool fits the
site. That question was capped at 10 answers — and which 10 was decided by menu position, not
relevance. On the biggest site the AI saw 10 pages out of 107.

I checked nobody else was fixing it (two ways — the ownership script, and reading the other live
sessions' transcripts), confirmed it is still real against the live database, and then measured
before choosing a remedy. The measurement was good news: unlike the sibling bug, where each row
carried a long description that had to be trimmed before the cap could go, these rows are one
short line each — page name, type, title. Showing ALL pages on even the biggest site costs about
two thousand words of context in a prompt that already contains the entire original page. So the
cap can simply go, with nothing trimmed and nothing hidden.

While measuring I found a second, smaller defect in the same query: one page on one site has two
research records attached, and the query lists that page twice — today, on the live system. The
fix closes that door too: each page now contributes exactly one line, using its newest research
record.

The change is one database migration (453), written to the same hardened pattern the sibling's
council review settled: it takes a backup snapshot first, refuses to run if another session has
already changed the row, and verifies its own result before committing. Next: council review,
commit, apply, verify live.
