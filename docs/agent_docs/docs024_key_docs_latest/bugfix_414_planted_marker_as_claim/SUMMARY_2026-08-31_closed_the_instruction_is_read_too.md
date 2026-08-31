# SUMMARY — 2026-08-31 — closed: the instruction gets read too

Second and final summary for this lane. The first (`SUMMARY_2026-08-27_…`) was written when the
framework fix was approved but the copy was still wrong on the site; the five headings genuinely
differ now, which is the test for writing another.

## What we were trying to do

Get a false statement off a live finance site, and close the hole that put it there.

An experiment in early August hid an instruction in one site's brief — *include the exact phrase:
checked against the FCA handbook, rule by rule* — to test whether the machinery obeys its brief. It
did, and nobody removed the instruction. So lendzy.co.uk told readers, in its own voice, that its
content had been checked against the regulator's handbook rule by rule. Nobody had done that. Then
our own quality-audit machinery read the sentence off the live page, decided it was the site's main
selling point, and filed a job asking a writer to substantiate it.

## Where we came from

The lane that found it stripped the instruction from the brief and recorded the source as fixed. It
wasn't: ten days after the original plant, one of our own agents had read that instruction and
rewritten it, in its own words, into a different part of the site's records — a part the writer never
reads and the site planner does. That copy was still live. **Deleting a planted instruction from the
place you found it does not retract it, because our agents copy instructions to each other.**

Underneath it, a structural gap. Every honesty check we had read what a site *says*. Nothing read
what its brief *tells it to say* — so a brief could lawfully order a page to state something the page
checks would refuse. The one existing mechanism that reads that text reads it in order to *exempt*
it, on the principle that a site's own voice specification outranks the fleet rules.

## What we did

Cleaned both sources under guards that refused to run unless the exact text was where expected.
Rejected the audit job, which was one click from regenerating the page around the false claim. Had
the framework rewrite the copy — and it wrote something better than it replaced: the guide now says
every figure *is given* with the named rule and a link, then adds "that does not make the checker
infallible… rather than take our word for it." We removed a claim asking readers to trust us and got
back one inviting them to check.

Then the class fix, three narrow changes into machinery that already existed: two patterns the
existing honesty check was missing by thirty characters of sentence length and one subject word (that
rule had been firing *zero* times fleet-wide — it was asleep); a new rule for claiming we have
verified something against a rulebook, shipped as a warning; and the claim rules applied to the brief
text our generators are given, across every agent's prompt rather than one. The review council
refused it at the first round — correctly, because I had asserted that nothing else reads brief text
without checking — and approved it at the second.

## Where we are now

**Closed.** The claim is gone from the site and from every brief in the fleet; the patterns are live
in the deployed system and cost nothing — across 2,715 live components they find nothing they were
not written for; the brief-side detector runs daily and reports zero from seven thousand scanned
fields.

The last defect found was **our own detector's**. Its first live run convicted a site for a phrase
that site's brief explicitly *bans* — a "would never say" list. Neither the tests nor a nine-seat
review round caught it; running it did. That is the honest headline of this lane: the checks that
found real problems here were, in order, running the thing against live data, an adversarial review
before building, and other teams re-measuring my claims. The unit tests and the review round caught
nothing that mattered — though the review round did stop me shipping on an unchecked assertion, which
is a different and real kind of catch.

## Where we are going

Nothing is outstanding on this bug. Four things are carried forward, none of them part of it:

- **A poisoned register passes every layer we have.** The register of verified facts is deliberately
  not scanned, because it stores the banned phrases themselves — so a fabricated source written into
  it would sail through everything, since every layer treats it as the thing they check *against*.
  Nobody has found an instance. If one appears it wants its own file.
- **An open decision for the owner.** One review seat dissented on the new rule shipping as a warning
  rather than a refusal, on a finance site where the false version ran 24 days. The argument for
  warning holds; the dissent is recorded so it can be revisited deliberately.
- Two narrower gaps: a *completeness* claim mandated in a brief is still not covered on the brief
  side (the page check refuses that shape on the output side), and the new warning-level rule has no
  post-deployment sweep, so it fires when a page rebuilds and never over pages already published.
- A duplication debt: the brief-surface census re-implements logic another tool already houses.
