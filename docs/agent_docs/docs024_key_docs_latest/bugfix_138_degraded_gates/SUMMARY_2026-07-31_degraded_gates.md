# SUMMARY — 2026-07-31 — the fix that was refuted, and the one that replaced it

## What we're trying to do

Before a code change goes in, a council of automated reviewers reads it. Each reviewer
gets a fixed budget of words for its reply.

If a reviewer runs past that budget its reply is cut off mid-sentence. We keep what we
can, mark it damaged, and — because a serious objection might have been in the part we
lost — treat the damaged reply as blocking. That is the right call.

The problem was what the record then said. It named the reviewer, as though it had made
a judgement. It hadn't; it ran out of room. And a reviewer that blocks a lot is exactly
what we look for when deciding a reviewer is too noisy to keep. **The failure hid inside
its own evidence: the fix is to give the reviewer more room, and the invitation is to
sack it.**

## Where we've come from

We found it by looking at a reviewer we had just introduced, which blocked two of its
first three rounds and looked like a bad hire. It was being cut off. More room plus an
instruction to be brief took its blocking rate from two-in-three to two-in-twelve.

That fixed one seat and left the mechanism standing, so we wrote down four ways to close
it: say *why* a round was blocked; measure how often it happens; give the reviewers that
need it more room; and put the important fields at the start of each reply, since it is
the end that gets cut.

The third was mostly done already. The first shipped two days ago. This week was the
other two — and the fourth did not survive contact with the evidence.

## What we've done

**The record now says which it was**, and both live copies of the service have been
carrying that since yesterday. It survived today's rebuild too, which is worth checking
and rarely is: a rebuild from an older starting point would have quietly removed a fix
that was working an hour earlier, and nothing else would have said so.

**We now measure the problem before it bites.** Counting the cut-offs would have
reported zero for ever, because we had already given more room to every reviewer that
had been getting cut. So instead we measure how *close* each reviewer is getting to its
limit. There is a report anyone can run and something that checks itself every six hours,
costs nothing — one database query, no AI — and leaves a note only when the picture
changes.

**It has since found two problems on its own**, which is the point of building it rather
than promising to look. One was a reviewer whose typical reply uses under two-thirds of
its budget but whose longest came within about sixteen words of being cut — invisible to
any measure of the typical case. The other was more important: **a reviewer we had given
double the room had already grown back into it**, in three days, with no other change.
That is the clearest evidence we have that more room buys time rather than safety.

**The fourth idea was mostly wrong, and finding that out was cheap.** The reasoning was
sound — cut-offs eat the end, so put what matters at the start — and it turned out the
important fields are already at the start, which is why our recovery works at all. The
one case I was certain would be a problem has never happened once in 2,713 records. And
the change I was about to make would have been actively worse: it would have pushed the
objections themselves off the end, and those carry both what blocks a round and what the
author needs in order to fix it.

**What did survive is the other half: telling reviewers to be brief, and why.** We
measured it properly this time — comparing rounds started before and after the change,
on the same reviewer, on the same afternoon, with its room limit unchanged. Its longest
reply went from ninety-eight per cent of the limit to fifty-five. That is the first time
we have been able to separate "be brief" from "have more room", because the first time we
tried it the two shipped together.

On the strength of that, you decided today to give the reviewers more room where it was
still missing, and to roll the brevity instruction out to all of them. Both are done: it
now covers forty-eight of the fifty-one reviewers.

## Where we are now

All four are answered. One of them is answered by being disproved, which is a real answer
and much cheaper than the change it replaced.

Three reviewers are deliberately not covered, and the reason matters more than the
number. Two already have a hand-written version of the same instruction and the tool
refuses to overwrite prose someone wrote on purpose. The third belongs to a different
kind of council where the thing the instruction *says* is not actually true — and putting
a confident false statement into a prompt that a reviewer will act on is worse than
leaving it out.

The rollout also created a risk the careful version did not have: forty-seven reviewers
changed behaviour on the strength of one measured case. Most were nowhere near their
limit, so I expect no visible difference. **The thing to watch is not whether replies get
shorter — it is whether they start raising fewer objections**, which would mean we bought
brevity with coverage. The instruction says explicitly to cut words and never findings,
but that is a hope until it is measured.

One item remains genuinely unfinished: the message we show when a round *is* blocked by a
cut-off has still never appeared in production, because it hasn't happened since the fix
shipped. Good news, and it means that wording is proven only in tests. I would rather
wait for a real one than stage a fake.

I should also say that a third of today went on measurements I got wrong before I got
them right — three of them, one of which put a duplicate and a false sentence into the
record before I caught it. The pattern was not carelessness: **each wrong answer confirmed
something I already had good reason to believe.** They are all written up where the next
person will meet them, which is the only thing that makes the time worth anything.

## Where we're going

Watch the objection counts either side of today, and catch the blocked-by-cut-off message
the first time it fires — at which point this can close. Give the other reviewers a couple
of days to accumulate enough history to say whether brevity helped them too; right now
only one of them has a fair before-and-after.

Beyond that, the useful thing this leaves behind is not the fix. It is that we can now see
a reviewer approaching its limit before it costs anybody a round, and that we found out —
by measuring rather than arguing — that raising a limit is a delay and telling a reviewer
to be brief is a remedy.
