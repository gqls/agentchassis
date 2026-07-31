# SUMMARY — 2026-07-30 — when a reviewer runs out of room, the round pays for it

## What we're trying to do

We have a council of automated reviewers. Before a code change goes in, sixteen or so
of them read it and either approve, object, or veto. Each reviewer gets a fixed budget
of words for its reply.

If a reviewer runs past that budget, its reply is cut off mid-sentence. We recover
what we can, mark the reply as damaged, and — because a serious objection might have
been in the part we lost — we treat the damaged reply as blocking. That is the right
call: better to ask the author to resubmit than to wave through a change a reviewer was
halfway through objecting to.

The problem is what happens next. The record of the decision says the round was
blocked by *that reviewer* — as though it had made a judgement. It hadn't. It ran out
of room. And a reviewer with a high blocking rate is exactly what we look for when
deciding a reviewer is too noisy to keep. **So the failure hides inside its own
evidence: the fix is to give the reviewer more room, and the invitation is to sack
it.** That is what we set out to close.

## Where we've come from

We found this by looking at a reviewer we had just introduced. Its first three reviews
blocked two rounds, which looked like a bad hire. It was being cut off. Doubling its
budget and telling it to be brief took its blocking rate from two-in-three to
two-in-twelve.

That fixed one seat and left the mechanism standing, so we wrote down four ways to
close it: say *why* a round was blocked; measure how often it happens; right-size the
budgets; and put the important fields at the start of each reply, since it is the end
that gets cut.

The third of those was already largely done by the time we looked — an owner decision
raised the budget on the reviewers that had actually been getting cut off. That
changed the argument for the rest: raising a budget doesn't remove the cliff, it moves
it. We proved that on ourselves, when the same reviewer started getting cut off again
against its new, doubled budget, purely because someone gave it a longer brief.

## What we've done

**The record now says which it was.** A blocked round distinguishes "a reviewer
objected" from "a reviewer ran out of room", and every round writes down whether a
budget overrun was involved. Both pods are running it and two real rounds have gone
through it, one of them correctly naming the reviewer with the genuine objection rather
than blaming a truncated one. The blocking rule itself is untouched — deliberately. We
changed what gets recorded, not what gets blocked.

**We now measure it, and the measurement is not the obvious one.** Counting the
cut-offs would have reported zero for ever, because the budget raises had already
stopped them. So we measure how *close* each reviewer gets to its budget, which warns
before the damage rather than counting it afterwards. Two halves: a report anyone can
run, and something that runs itself every six hours, costs nothing — it is one database
query, no AI involved — and leaves a note only when the picture changes. It found five
reviewers worth attention within a minute of being switched on.

**Its first real finding was a gap nobody was checking.** Three of those reviewers run
with the old, smaller budget on one of the three councils that use them — including the
very reviewer whose cut-offs started this. Our sync tool only mirrors two of the six
councils, and our parity checker deliberately doesn't compare councils against each
other, because councils are legitimately allowed to differ. So a decision the owner
made reached two places out of three, and neither existing check could see the third.

**The fourth idea turned out to be mostly unnecessary, and finding that out was
cheap.** The reasoning was sound — cut-offs eat the end, so put what matters at the
start — and the important fields are already at the start; that is why our recovery
works at all. The one case I was certain would be a problem has never happened once in
2,713 records. And the change I was about to make would have been actively worse: it
would have pushed the objections themselves off the end, and those carry both what
blocks the round and what the author needs in order to fix it.

## Where we are now

Three of the four are answered. One is answered by refutation, which is a real answer
and the cheapest one available — a measurement that costs an afternoon beat a change
that would have made things worse.

Two things are outstanding and both are somebody else's call rather than more work.
The budgets on that third council are the same judgement the owner already made once,
and on that council's own evidence the trigger hasn't been met — so it is flagged, not
actioned. And the one piece of today's work that isn't switched on is a small script
that adds a brevity instruction to the reviewers under most pressure: it is written and
tested, and the command that writes to live configuration was refused by this session's
permission check. That is a yes/no, not a problem.

Worth saying plainly: two of today's hours went on measurements I got wrong before I
got them right. In one case I wrote a warning about a specific trap into my own query
and then read the next result straight through it, because the wrong answer made a
better story than the right one. Both are written up where the next person will meet
them, which is the only thing that makes the time worth anything.

## Where we're going

Watch for the one thing still unproven in production: the wording we show when a round
*is* blocked by a cut-off. It hasn't happened since the fix shipped, which is good
news, and it means the message itself has only been proven in tests. We wait for a real
one rather than staging an artificial round.

Beyond that, this bug's remaining work is not ours to do: a decision on the third
council's budgets, and a yes to the brevity script. The mechanism the bug was named for
— a reviewer's word count quietly turning into a verdict — is now visible every time it
happens, and counted before it happens.
