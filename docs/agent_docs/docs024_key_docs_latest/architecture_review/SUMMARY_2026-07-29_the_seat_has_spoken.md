# SUMMARY — 2026-07-29: the seat has spoken, and one of its objections became an RFC

*New file, not an edit of `SUMMARY_2026-07-28_the_seat_can_see.md`. That one is
still an accurate record of what we believed yesterday, and the distance between
the two is the point: yesterday the seat could **see**; today it has **spoken**,
and something happened because it did.*

---

## What we're trying to do

Make sure the platform's architecture stays sound as it grows — that it doesn't
drift under us, and that it is sufficient not just for what we're building this
week but for what we'll want next. The owner's framing from the start: there is a
real tension between rewriting things to be best and keeping things that are
battle-tested, and somebody in the review process should be arguing the forward
side of that, because everyone else in the room is arguing the safe side.

Concretely: a council seat, `review_architecture`, whose remit is not "is this code
correct" but "is this the right shape, and what does it commit us to".

## Where we've come from

The design was settled in a week of rulings (D1–D11). One conservative seat at full
remit, one forward seat, no duplicates — which meant the entire forward half rested
on a single seat. That seat was built, seated on `feature-designer`, and verified
reachable.

**And then it said nothing at all, for days.** Not because it was broken: it sat on
a lane that almost never runs, behind a gate requiring an owner-approved spec, and
the few tickets that qualified belonged to other threads. We diagnosed that
correctly as a rate limit rather than a fault, and then — this is the honest part —
we mostly waited.

Alongside it ran D11: the observation that a seat which cannot look things up is
being asked to reason from memory. Layer 1 put the actual source of every Go symbol
into a searchable index. Layer 2 routed seats' questions to it. **Layer 1b — putting
our own written documents in there too — has been in council for eight rounds and
is the one thing still unfinished.**

## What we've done

**Since yesterday evening, three things, and only one of them was mine.**

**The seat was moved, and it started talking.** Another session seated
`review_architecture` on `fix-proposer` and `council-gate` — the lanes that actually
carry traffic — after the owner reversed the earlier ruling against putting a
forward seat on the fix lane. **It has now reviewed 9 rounds in about ten hours:
five approve, four object.**

**Its objections are the right kind, which was the real question.** A forward seat
that just re-checks correctness is a waste of a seat. This one doesn't. It flagged a
fleet-wide lookup table with *"no owner or update trigger"*; it said of a change
documented in a comment that *"a doc comment is not an enforcement mechanism"*; of a
carefully-built change going through as a bug patch, that *"care does not relocate
the review track"*; and of a clean measurement across 15 sites, that it was *"a
snapshot, not a standing guarantee"*. It cited this workstream's own process
document by name, enforced this workstream's own ruling about shared seams, and on
one occasion **objected to a round that everyone else approved.**

**And one of those objections became an RFC.** It objected that two new keys added
to a shared vocabulary were an architecture-scope change even though they were
small, additive and measured at zero collision. The owner instructed that it be
routed to a real architecture review. `RFC_002` was filed this morning, naming the
seat's ruling and the correlation id it came from.

**Separately, the layer 1b work reached a decision point and stopped.** Round 8 came
back 8 approve / 1 object, and the last objection was correct: putting our documents
into the code index would let a paragraph from a bug write-up be cited as *"the code
says so"* in a confirmed diagnosis. The owner chose the fix (option 1b). It is built
and in council now, deliberately **ahead of** the documents it guards and separate
from them, because it changes how every diagnosis run labels its evidence.

## Where we are now

**The workstream's oldest open question is answered.** For a week the honest state
was "we have built a forward-looking seat and we do not know whether one seat can
carry that remit, because it has never spoken." It has now spoken nine times, and
the answer looks like yes: it argues the forward side, it does not duplicate the
conservative seats, it is calibrated enough to say *"flagging for the record, not to
block"*, and it produced a piece of durable prior art on its first real disagreement.

**Two honest limits on that.** First, the RFC needed the owner to instruct it — we
have shown the path works when a human walks it, not that it runs on its own. One
RFC is not a rate. Second, **I published a handoff last night stating the seat had
zero reviews; it had fired ten minutes after I wrote it, and I found out this
morning by accident** while doing unrelated work. The claim was the headline of my
own workstream and I never re-ran the one query that checks it.

**Layer 1b is still not finished** — eight rounds in, now waiting on its guard to
clear council. The guard wedged on a known-flaky seat and has been resubmitted.

## Where we're going

Three things, in order.

**Finish layer 1b.** Get the guard approved, built and rolled — verified by a string
the change *deletes*, which is stronger evidence than one it adds. Then the markdown
plan returns for its ninth round with the hazard already closed.

**Find out whether the RFC trigger works without a human.** The seat is designed so
that its verdict is a trigger, not a veto. We have one instance and an owner in the
loop. The question worth measuring next is what happens on the next objection of
that shape when nobody instructs anything.

**Watch the object rate.** The design has a deliberate kill switch: if the seat
objects a lot and none of it carries signal, it is noise and should be pulled. At
four in nine, with every objection carrying its reasoning, it passes today. It is
worth re-reading rather than re-counting in a week — the metric undercounts correct
behaviour, which we already know from the guardian's stability-preference numbers.
