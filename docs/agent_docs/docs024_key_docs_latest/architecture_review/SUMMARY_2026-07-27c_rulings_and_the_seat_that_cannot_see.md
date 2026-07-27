# Architecture review — where we are, 2026-07-27 (evening)

Third summary of the day, and the first two are not superseded: `SUMMARY_2026-07-27`
was written before the seat existed, `SUMMARY_2026-07-27b` when it had just been
built. This one exists because the shape of the work changed — the owner ruled, the
design closed, and the question turned from *what should we build* into *can the
thing we built actually see*.

---

## What we're trying to do

The owner asked for a process — possibly a council seat — that keeps the
architecture robust, stops it shifting underneath us, and keeps it sufficient for
what we plan to do next, knowing those three goals conflict.

The conservative half already existed: the guardian seat, sole holder of the hard
veto, whose charter says to prefer a fix at a higher layer over editing long-stable
core infrastructure. Nothing anywhere argued the forward half — whether the design
is *sufficient for what is coming*. That absence was the gap, and this workstream
measured whether it mattered, found it did, and built the counterweight.

## Where we've come from

We measured ossification rather than asserting it. Across 259 past council reports
the guardian had deflected `coordinator.go`/`ProcessResponse` upward across six
distinct submissions in seven days, `spawn_actions.go` four — while that same core
moved nine commits in sixty days against 2,123 repo-wide. Pressure on the core is
high, change to it is near zero, and the difference lands as workarounds above it.
That justified staffing a forward seat rather than merely arguing for one.

Along the way the premise it was meant to test refuted itself: `platform/orchestration`
looks like 366 commits in sixty days, but 348 of those are a plug-in action registry
growing. The core is stable *because* the registry absorbs the growth. So the risk was
never churn — it was ossification, which is the opposite failure.

We also found the council had always been able to read its own minutes and nobody had
told it. Every past verdict sits in a table already named in the reviewers' schema
hint; not one of thirty-two reviewer prompts mentioned it. That, plus an index of our
own case files and wrong calls, became the change that shipped.

## What we've done

**The seat is built and every council is now equipped.** The design-stage council had
been left the worst-off of the three — an accident of plumbing, because the script
that copies reviewer improvements between councils spans only two of them and that one
isn't on the path. It now has the council's own minutes, a check against repeating its
own past deflections, and an index of our case files.

**The first evidence arrived, and it was good.** In the first council to run after the
change, one reviewer rejected a plan's claim that "no sixth case exists" by citing two
of our own logged mistakes *by date* — both prior occasions where we'd asserted an
absence without searching for it. Another cited three case files by number. The
guardian invoked its caution rule and then argued itself out of it, saying the
repetition was evidence of a genuinely scattered defect rather than evidence the fix
belonged at a higher layer. That is precisely the judgement we had measured it getting
wrong six times.

**The owner ruled twice, and the design closed.** Don't narrow the guardian — it keeps
the hard veto *and* the full remit including weighing benefit. And don't add the second
forward reviewer that a historian's comment implied, because the new seat already is
that voice. Together: one conservative reviewer at full powers, one forward reviewer,
no duplicates, and the balance struck by the two arguing rather than by trimming
either.

**Then the seat turned out not to be able to see.** Chasing whether it could look
anything up, we found the code index stores only declarations — function signatures,
never bodies — so a search for a route, a config key or "does anything still reference
this" returns empty, and empty is indistinguishable from "doesn't exist". Worse, on two
of the three councils code questions were **never answered at all**: no step existed to
run them. The reviewer whose entire charter is "are we rebuilding something we already
have?" had its code questions dropped on all three lanes. And the new forward reviewer
sat on one of the two lanes with no such step, asking into a void, while its
instructions promised the answers came back next round.

**Three things shipped in response.** First, every reviewer that can ask a code
question — fifteen prompts across three councils — is now told the truth: an empty
result is *no information*, never absence, and an absence claim goes to a human.
Second, the routing was fixed: the prior-art reviewer is now answered, and the
design-stage council gained the step it lacked, so the forward seat's questions finally
go somewhere. Both are configuration, live immediately, no rebuild. Third, the real
fix — making the index store bodies — went to the council gate for review.

## The council decisions, since they are part of the record

**The gate REVISED our plan, with one HIGH objection and six MEDIUM, and it was right
on every count that mattered.** The high objection: our rationale claimed the plan
"fixes routing and content" while its edits only touched content — and it named the
consequence we had missed, that the council's only forward-fitness voice runs on the
exact lane left unrouted. We answered it by *shipping* the routing half rather than
arguing.

Two more were settled by measurement rather than debate. The plan used `CREATE INDEX
CONCURRENTLY` as generic caution; two seats flagged it cannot run inside a transaction,
and checking showed the migration runner's own dry-run probe wraps files in one — so
the "safe" choice would have broken the safety mechanism, on a 4,535-row table where a
plain index takes milliseconds. And the reuse seat caught the sharpest one: we rejected
reusing an existing body-slicing function on an interface mismatch and then specified a
new one "matching that function's convention" — which concedes the two must stay
identical, and is therefore the argument for extracting a shared primitive. The file we
cited already asked for exactly that in its own header, which we had not read.

Round 2 answers all seven and is queued. **No `Council-Reviewed:` trailer has been
claimed** — that is earned by an APPROVED verdict only.

**The pattern is worth stating plainly: three of our four errors this evening were
caught by the council, not by us**, on a submission already pre-flighted for quote
fidelity, schema and scope. That is this workstream's own thesis arriving as evidence
against its author — a reviewer with the written record in front of it caught what the
person who wrote that record the same day did not.

## Where we are now

The design is settled and nothing is waiting on the owner. Every decision is ruled,
built, or deferred behind a named trigger. One item in the decisions file is a proposal
another session handed us, for the owner to read rather than decide.

The forward seat is live, reachable, correctly wired, and **has still never spoken** —
zero reviews. That is not a fault: it only runs on the design lane, and that lane
refuses anything without an owner-approved capability spec. There are five such specs,
two approved, and both belong to other threads. Its first review will arrive when one
of them runs its next round.

Its instrument is now half-built. "Does this symbol exist?" and "what's under this
path?" work today. "Does anything reference this route?" does not, and won't until the
index stores bodies — which is the piece under council review.

One honest caveat on our own measurements: the adoption report's headline was wrong
until this evening. It counted two things separately and we quoted them as a ratio, so
"6 of 90" was never a subset. The true baseline is **2 of 90**. The correction makes
the case stronger, not weaker — the seat consulted its own history even less than we
claimed — but it was the single figure the workstream is judged by, and it was the
unchecked one.

## Where we're going

Three things, in order, and only the first is ours to finish.

**Get the index answering.** Round 2 is with the council. If approved it is a schema
change, a rebuild and a roll — the first thing on this workstream that isn't
configuration. It also carries the piece that has been missing since the beginning:
the same mechanism makes our written history — the wrong-calls ledger, the bug files,
the concept register — searchable by a reviewer at all. That is the "sufficient for
anticipated plans" half of the owner's original question, and it has never had an
instrument.

**Let the seat speak, and then read it honestly.** One design run is all it takes. When
it comes, section 4 of the adoption report finally reads, and the kill switch built into
that report becomes usable: a seat that objects to everything while emitting no signal
is producing confident noise and should be pulled rather than tolerated. We should be
willing to use it.

**Then the deeper limit, which is its own RFC.** Today a reviewer asks a question and
gets the answer *next round* — so it cannot look while reasoning. It must guess, commit
to a verdict, and be corrected a round later. Making that live changes the shape of a
review rather than its inputs, and it should not be smuggled into a change about what
the index contains.
