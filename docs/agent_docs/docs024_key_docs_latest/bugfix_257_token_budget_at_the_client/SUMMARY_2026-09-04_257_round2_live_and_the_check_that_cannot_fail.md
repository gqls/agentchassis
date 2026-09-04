# 257 — round two is live, and the check that was supposed to prove it cannot fail

*2026-09-04. Fourth in the series, after `SUMMARY_2026-09-03_257_round2_the_class_returned_and_is_now_fenced.md`,
which said the work was committed, the verdict was not in, and nothing was live. All three have changed.
Written to be read aloud.*

## What we are trying to do

Unchanged. When we ask a language model to write something we must tell it how long the answer may be.
That number should be a setting an operator controls, not a number typed into the program — and,
equally, we should be able to *tell* which of those two situations we are in, because the failure is
silent.

## Where we have come from

August: only one function in the estate actually read the setting, so any code that talked to a provider
directly ignored it. Fixed at the provider clients, approved, shipped, verified.

Three weeks later the same fault came back in code written *after* that fix — two new pieces of code
each re-implemented the rule by hand and each ended in a typed-in `2000`. Worse than the original,
because a limit sent with the request beats the one the client would work out for itself, so those
callers could never inherit the configured number. Worse again in one case, where the typed-in number
happened to equal the configured one, making our own records unable to tell obedience from neglect.

## What we have done

Deleted both hand-written copies, plus three more found while writing the enforcement — one handing the
model an empty settings object, one that would pass a limit of zero straight through to a provider that
rejects it, and one nobody had ever seen because it talks to a different supplier over plain web
requests and every survey we run looks for calls to *our* client. All five now read the setting through
one shared piece of code, and a build-breaking check refuses a typed-in limit anywhere in that part of
the system.

The review council **approved** it, all reviewers, sixteen minutes. Two objections were worth acting on
rather than noting, and were: one asked us to state why a narrower older check is kept beside the new
one, the other asked us to verify a premise we had asserted. Both answered in the code.

The owner's chassis build went out at 22:06 last night and **the change is live**. Two hundred and eight
calls have gone through the affected places since, every one sending exactly the limit its configuration
states, nothing cut off.

## Where we are now

Live, approved, and with one honest hole that is worth naming out loud, because it is the same shape as
the bug itself.

**This change was designed to alter nothing, and it succeeded — so nothing measured after the fact can
tell the new code from the old.** Every step involved already stated its limit properly, so both
versions send the same numbers. The check the bug file prescribes for exactly this moment — watch a step
whose limit exceeds the typed-in number send the larger number — cannot help either, because that step
was sending the larger number *before* the roll. A check that returns the same answer whichever way the
world is, is not evidence. That is precisely what this round was about, and it turned up inside our own
verification plan one level further out than we had aimed it.

What settled it instead was another team member's work. A different fix, committed twelve minutes after
ours, was proven present in last night's build by a careful test with four controls. Ours sits
underneath theirs in the history, so any build containing theirs contains ours. That is solid, it is one
command, and it is not first-hand — the documents say so.

Our own attempt at a first-hand check failed in a way worth remembering: we asked the running program to
look for 474 possible answers at once, and the check was killed part-way rather than answering. Taken at
face value it would have said the change was *not* there.

## Where we are going

Three decisions wait on the owner rather than on work, and they are set out in today's handoff.

The first is whether to run the one test that would settle this by behaviour instead of inference: put a
limit where the old code never looked and the new code looks first, on a single step, and read what the
next call sends. One setting, one token of difference, reversible in seconds — but it writes to a live
production agent, so it is not a session's call. It would also be the first time anything on the fleet
has exercised that path at all.

The second and third are unchanged from August: whether to merge the last two near-duplicate copies of
the "which limit wins" rule, and whether making direct model calls visible to our truncation monitoring
belongs to this bug or to its own piece of work.

And one that arrived with this round: four settings on the site-adoption agent that ask for a limit
nothing reads — one of them asking for thirty-two thousand and receiving sixteen.
