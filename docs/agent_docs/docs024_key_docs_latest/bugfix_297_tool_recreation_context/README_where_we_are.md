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

## 2026-08-17 (later) — the fix is live, and the council is looking at it

The migration is applied. The system now shows the analysing model every page on the site rather
than the first ten by menu position — on our biggest site that is 107 pages instead of 10 — and
each page appears exactly once, which closes the duplicate I found earlier today. A backup of the
old configuration was taken automatically first, and a tested one-command revert sits beside the
migration if anything looks wrong.

One small thing worth recording, because it is the kind of judgement I would want to be asked
about: while writing up the risks I noticed the "newest research record" rule could misbehave if a
record ever arrived without a timestamp. None of the twenty-one existing records has that problem,
so I could have left it as a note for the reviewers. I closed it instead — it cost one word in the
query and it means the problem cannot happen rather than merely being unlikely today.

The council round is running (they take about half an hour to reach the front of the queue). I have
committed everything with a marker that says "submitted, verdict not yet read", which is the house
rule here — nobody holds work waiting for a review, because the review is designed to come after.
When the verdict lands I will read it and act on it; if the reviewers find something, that is
cheaper than finding it later, and on the sibling bug two of the rounds found real defects.
