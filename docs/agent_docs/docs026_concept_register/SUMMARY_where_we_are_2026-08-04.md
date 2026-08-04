# SUMMARY — where we are, 2026-08-04

*Concept register. Written to be read aloud. The previous milestone summary is
`SUMMARY_where_we_are_2026-07-20.md`; the series is the record, so this is a new
file rather than an edit of that one.*

## What we're trying to do

Keep one place that answers "does this already exist, and is it real?" for every
mechanism, contract, agent, tool and idea in the platform. Not documentation —
documentation says how a thing works. The register says **that** a thing exists,
who built it, what it is for, what will mislead you about it, and whether it is
actually live or merely written down. Two audiences use it: a session about to
build something, who needs to find out it was built in June; and the council seats
that review our changes, whose prompts are seeded from it. That second audience is
why accuracy is not a matter of tidiness — a wrong entry is false evidence inside a
machine review, which is how one register line once blocked a correct plan at high
severity and cost a full round of credits to disprove.

## Where we've come from

It was built in July by sweeping all ~4,100 files under `docs/`: 2,185 raw concepts
extracted, merged to 1,627, classified into 107 categories. Then every one of those
1,627 was checked against the live code and database, which corrected 124 of them —
about 7.6% — and, just as usefully, overturned 106 *proposed* corrections on an
adversarial second pass. That gave the vocabulary its seventh status, `convention`,
for the ~48 concepts that are doctrines rather than artefacts.

Stage three built council seats out of it. Sixteen reviewers now sit on the fix
loop and the council gate — five always, eleven woken by relevance to the files a
change touches — and the constitution and mission documents got a ledger, a commit
gate and an integrity checker so they cannot drift without the owner's word.

Since 20 July the register has had no dedicated thread. It grows by a different
mechanism: every session that builds something reusable registers it in the same
commit, and a coverage check watches for workstreams the register has never heard
of. It has grown 1,627 → 1,756 that way.

## What we've done

Asked whether that self-maintenance actually works. It didn't, in two places, and
both failures had the same shape: **a check whose result could not have come out
otherwise.**

**The master index was 2% short.** 34 concepts had a full register entry and no row
in the index — the whole first half of the claims-verification layer among them,
plus the single deployed-asset path derivation, the markup-safe rewrite helpers, the
SSRF guard, the experience register's four entries, and the public-API trio. The
index is the file people search. Those 34 were invisible in precisely the lookup
they exist for, and a search would have reported them as not existing — the exact
failure the register was built to prevent. It survived about twenty recorded
re-measurements because every one of them counted index rows against the previous
index-row count, which cannot see a row nobody wrote. Backfilled, and the header now
carries the check that does see it: compare entry ids against row ids, both
directions. Both lists are empty at 1,756.

**The coverage check's accepted-backlog list wasn't accepting anything annotated.**
Sessions had been writing the reasoning beside each line — why this lane is a
one-off and needs no entry — and the annotation stopped the line matching, so those
lanes were re-reported as new on every run. Twelve of the seventeen "new" items were
already-settled decisions. Fixed, and the annotations now survive a rebaseline
instead of being wiped by it.

**Then the seven real ones.** One new entry (a pre-commit detector guarding the
undeliverable-reply rule, shipped precisely because widening the real fix needs an
RFC), one second-consumer note on an existing entry, one lane that had already done
its own paperwork, four out of scope.

**And, separately, the estate lost 1,339 duplicate documents.** 441 documents
existed as 1,973 files — every save of a living document kept as its own numbered
copy, up to 57 for one running-notes file. The newest of each stays; git keeps the
rest. Two traps made this less mechanical than it looks, and both are written down:
the unnumbered copy is sometimes the *newest* member rather than the oldest, and the
numbering does not always run in date order.

## Where we are now

The register holds **1,756 concepts across 109 category files**, and for the first
time in the series the index and the category files agree exactly — no entry without
a row, no row without an entry. The coverage check is quiet: no workstream on disk
is unaccounted for, either registered or ratcheted with a stated reason. The
direction-integrity checker is green on every blessed document, copy and council
seat prompt.

What is *not* true: nothing here re-verified whether the 1,756 entries are still
**accurate**. Stage-2 verification ran once, in July, and this session deliberately
did not touch a single existing status. The register's honest claim today is that it
is **complete and consistent**, not that it is current.

## Where we're going

Three things, in the order they matter.

**Staleness is the open flank.** Coverage answers "is it here?"; nothing answers "is
it still true?" A concept-register status line is already a known landmine — a
snapshot that outlives its truth, read as ground truth by council seats. The
building blocks exist: `covers-through` stamps per file, the `landmine-verifier`
that fact-checks one entry against the code, and the bugs-open staleness sweep that
does the same for citations. Pointing that machinery at register entries is the next
real piece of work, and it is a design question, not a chore.

**The index drift will come back unless something watches it.** It is fixed today
and the check is in the header, but that is a convention, and this session exists
because a convention decayed for three weeks unnoticed. The natural home is the
existing pre-commit coverage check, which already runs on the commit path and
already knows where the register is.

**43 register citations now resolve only through git**, a consequence of the
deletion. That is recorded rather than repaired: recovering them would mean
rewriting 43 `sources:` lines to name a commit, which is worth doing only if someone
actually trips over it. The landmine entry is there so the first person who does
knows the evidence exists rather than concluding it was invented.
