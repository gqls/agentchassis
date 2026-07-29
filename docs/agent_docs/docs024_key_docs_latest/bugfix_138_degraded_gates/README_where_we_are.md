# Where we are — the council's "truncated review becomes a blocking review" bug

Plain prose, append-only, newest at the bottom. The owner keeps this one too.

---

## 2026-07-29 (thread "bugsearch 5")

We put a new reviewer on the council yesterday — the architecture seat, the one
that had never fired in its whole existence. It fired, and immediately started
objecting to things. Two of its first three reviews came back as objections, which
sent the whole round back for a rewrite.

That looks like a seat that is too fussy. We even have a written rule for retiring
a seat that objects too much, and I was the person watching for exactly that,
because I had just put it there and wanted to know whether it was earning its
keep.

It was not being fussy. It was being cut off mid-sentence.

Every reviewer gets a budget for how much it can write. The architecture seat had
a longer brief than the others — four dimensions to think about — and it kept
running past the end of its budget. When that happens the council does something
sensible: it keeps whatever the reviewer managed to say, marks the opinion as
damaged, and refuses to wave it through, on the grounds that the missing bit might
have been the serious objection. That is the right call and I would not change it.

The problem is what the round then reports. It says "blocked by an objection from
architecture". It does not say "blocked because architecture ran out of room". So
from the outside, a reviewer that ran out of paper and a reviewer that found a real
problem look identical — and the fix for those two is opposite. One needs more
room; the other needs the change rewritten.

We fixed the seat itself in the morning: more room, and we moved the important
part of its answer to the front so that if it ever does get cut off, the useful bit
survives. That worked, and the numbers say so plainly. Its objection rate went from
two-in-three to two-in-twelve the moment it stopped being cut off. So it was never
a fussy seat. Had we gone by the objection rate alone we would have removed a
reviewer that was doing its job, and we would have felt entirely justified doing it.

Then I went looking for how often this had happened to the *other* reviewers, and
the answer is: seventeen times in a fortnight. Seventeen rounds sent back for a
rewrite because a reviewer ran out of room, each one costing credits and about
half an hour, and none of them saying so.

**What I have just built** is the small, boring half of the fix: the round now says
which it was. "Blocked by an objection from X" when a reviewer genuinely objected,
and "blocked by a TRUNCATED objection from X — this is a budget problem, not a
judgement" when the only thing standing in the way is a sentence that got cut off.
And it records a yes/no on every round so we can count this from now on instead of
reconstructing it after the fact.

I deliberately did **not** change the blocking rule itself. A cut-off review still
blocks. Letting it through would risk missing a real objection, and a missed real
objection is far more expensive than a wasted round.

Two things I want to flag honestly:

**This does not stop the waste.** It makes it visible and correctly labelled. Those
are different things, and the second is what lets us fix the first properly rather
than by guesswork.

**The obvious fix — just give everyone more room — is not enough on its own**, and
the architecture seat is the proof. It got more room this morning and it had been
running out with the *standard* room only because someone (me) gave it a longer
brief. Any new reviewer with a lot to think about will do the same. Raising the
budget moves the cliff; it does not remove it. Another thread has already raised
the budget for the two worst offenders on an owner call, and that was worth doing —
but every seat that has actually hit this is now on the larger budget, and I still
expect it to recur, which is precisely why I wanted the label first.

One small correction to something I nearly wrote in this file an hour ago: I was
about to report that none of the reviewers had ever had their budget raised, based
on a query of mine that came back clean and uniform for all seventeen of them. The
query was looking in the wrong place in the config and was quietly answering a
different question. Three of them had been raised already. Nothing was lost, but
it is a good example of the way a bad check does not fail — it just tells you
something plausible.
