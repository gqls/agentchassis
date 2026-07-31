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

---

**2026-07-30 — the other two fixes, and one of them turned out not to be needed.**

Yesterday's fix is live. The binary in both chassis pods now carries the new
wording, and two council rounds have actually run through it — one of them a real
"revise" that correctly named the seat with the genuine objection rather than a
truncated one. So the thing we were worried about — a round blocked by a seat that
merely ran out of room, looking exactly like a round blocked by a considered
objection — now says which it was, in production. The one part still unproven is the
message you see when it IS a truncation, because no round since the roll has had
one. That is a good problem.

**Candidate 2 was "alert on the rate", and the interesting part was working out
which rate.** The obvious one — count the truncations — would have sat at zero for
ever and told us nothing, because we'd already raised the token limit on every seat
that had been getting cut off. But raising a limit doesn't remove the cliff, it moves
it: this same bug caught the architecture seat truncating against its *new*, doubled
limit within hours, purely because someone gave it a longer brief. So the measurement
is now how CLOSE each seat is getting to its limit, which warns before the damage
instead of counting it afterwards.

There are two halves. One is a report anybody can run. The other runs itself every
six hours, costs nothing (no AI involved — it's a single database query), and leaves
a note only when something changes. It fired within a minute of being switched on and
found five seats worth looking at. It is deliberately not a heartbeat: if the same
five seats stay flagged it stays quiet, and only speaks again when the list changes.
An alert that repeats itself every six hours gets ignored within a day.

**The most useful thing it found was a gap nobody was checking.** Three of the seats
run with the old, lower limit on one of the three councils that use them — including
the very seat whose truncation started all this. The sync tool we have only mirrors
two of the councils; the parity checker deliberately doesn't compare across councils,
because councils are legitimately allowed to differ. So a fix the owner explicitly
ruled on reached two places out of three and nothing noticed. I have not changed
those limits: whether to raise them is the same judgement call the owner already
made once, and on that council's own evidence the trigger hasn't been met. Flagged,
written down, left for a decision.

**Candidate 4 was mostly wrong, and finding that out was cheap.** The idea was
sensible: truncation eats the end of a response, so put the important fields at the
start. I checked all 51 reviewer templates and measured what actually gets lost. The
important fields are already at the start — that's why the recovery works at all. The
one field I was sure would be a problem (the severity grade, written last inside each
objection, where losing it silently escalates the objection) has **never once** been
lost: 0 out of 2,713. And moving the reasoning to the front would have made things
worse, because it would push the objections themselves off the end — and those carry
both what blocks the round and what the author needs to fix it.

What did survive scrutiny is the *other* half of the earlier fix: telling the
reviewer to be brief, and why. On the one seat where we tried it, the responses got
shorter rather than just having more room. I've built that as a small script rather
than five hand edits, aimed at the seats the new report says are actually under
pressure. **It is written and tested but not yet switched on** — the command that
writes to the live configuration was refused by this session's permission check, so
that is the one thing I need from you: a yes to running it.

Worth saying plainly: two of today's four hours went on measurements I got wrong
first. I wrote a warning about a specific trap into my own query and then read the
next result straight through it, because the wrong answer happened to make a better
story than the right one. Both are written up where the next person will hit them.
