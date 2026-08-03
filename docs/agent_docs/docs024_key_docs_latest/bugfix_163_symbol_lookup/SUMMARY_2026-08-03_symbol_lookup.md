# SUMMARY — 2026-08-03 — the code lookup that could never find a symbol, fixed and proven

**What we're trying to do.** Our platform establishes facts about its own code through one
shared lookup: the diagnosis loop uses it to ground root causes, the review council uses it
to check proposed changes, and the landmine verifier uses it to keep our trap-corpus honest.
One of its three question types — "does this symbol exist in this file?" — had never worked,
and worse, answered every such question with a confident "searched and found nothing".

**Where we've come from.** The bug was found on 31 July by another workstream and sat unowned
for three days — partly because a scan of live sessions concluded someone was already on it,
when in fact that session was merely working near it. The mechanism: the lookup chopped
"file:Symbol" questions into words and demanded every word — including path fragments like
"internal" — appear inside the symbol's *name*. No name contains a path, so the answer was
zero, always, by construction. Twenty-one verdicts from the landmine verifier over four days:
sixteen "needs human review", none ever mechanically confirming a symbol; and the one that
started the investigation blamed a stale index that was not the cause.

**What we've done.** Fixed it at the one shared function all three consumers call: the path
half of the question now goes to the path column, the name half to the name column, reusing
the parser that already owned that convention rather than writing a third. Two shapes the
original bug report never anticipated are handled too: a line-number reference after the
colon (a dozen of our own corpus entries are written that way) degrades to a file check
instead of matching "3066" against function names; and when a symbol is not at the path you
named, the answer now says where it *is* — so a moved file can never again masquerade as a
missing symbol. Every empty answer states the exact search it ran. Tests were proven by
mutation (each shown to fail when the fix is removed), the change went through the review
council (first round killed by an unrelated deployment restart, resubmitted on the same paper
trail), and image v1.0.1245 was verified on both live replicas by reading the binaries.

**Where we are now.** Proven end to end on the most fitting possible witness: the exact
corpus entry whose failed verification opened this bug was re-verified on the new binary —
every one of its six symbols resolved, and the verdict named the (real) index staleness
honestly instead of inventing it as a cause. One self-inflicted lesson logged along the way:
our own damage figure counted the word "CONFIRMED", which the verdict vocabulary cannot emit
— a zero that could never have been otherwise; corrected in place.

**Where we're going.** The ticket is closed. Two follow-ons live with their owners: the
verifier's judging prompt should be told a per-check "searched: …" line now outranks the
run-level staleness banner (one sentence, owned by the architecture-review lane, who have
been notified in writing); and the silent row-cap issue in the same function (bug 181,
another lane's, untouched here) now has the single rendering seam its own fix asked for. The
deeper follow-on is fleet hygiene: the code index is still a single snapshot from 28 July,
and every consumer of this lookup inherits that staleness — tracked separately (DIAG-037).
