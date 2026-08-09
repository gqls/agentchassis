# loancalculator.co.uk — the critics were holding another site's rules

*Written 2026-08-09, from the work of the previous evening. The last read-out
(`SUMMARY_2026-08-08c`) closed the voice rollout: twenty-six pages rewritten, every
calculator proven untouched. This one is about a different thing entirely — the platform
bug that exercise exposed on its way out, and what looking at it properly turned up.
Series, not replacement: the three 08-08 summaries stand.*

## What we are trying to do

The site itself is finished, and nothing in what follows changes that. What is left is a
fault this lane found in the machinery, not in the site: the agent that plans an
"experience" — a whole journey across a site rather than a single page — cannot currently
plan for any site except one. We own that bug because we found it, and the job is to fix
it properly rather than to work around it.

There is a second thing being tested here, quieter but more useful in the long run:
whether a fault of this shape can be found, understood and corrected without anyone
having to notice the symptom first. Nobody complained. Nothing failed. The bug had been
sitting in production for three weeks producing confident, well-formed, completely wrong
work.

## Where we have come from

The site was adopted at the end of July, decomposed into framework components in early
August, and rewritten page by page into a house voice. That finished on 8 August: all
twenty-six pages in the new voice, every one serving, all eleven calculators giving
identical answers to before.

The last question on the site was an editorial one — whether a page written for someone
who has just missed a loan payment should open with free charity advice rather than with
negotiation tactics. To answer it, we pointed the experience planner at the site. It came
back with a confident, detailed plan about a completely different site: a game, its daily
provocation, its timed round. You made the editorial call yourself in the end, and the
page was reordered through the framework the same evening. The planner's failure was
written up and left filed.

## What we have done

**Picked the bug up, and found it is considerably worse than filed.**

The cause was never subtle. Somebody had written one particular site's situation directly
into the planning agent's instructions — its broken pages, the decisions you had taken
about it, the exact file its widgets read. Every run, for any site, is told that *that* is
the problem it is fixing, immediately after being told which site it is actually planning
for. It went unnoticed for three weeks for a simple reason: of the sixty-one plans that
exist, fifty-nine are for the site it is hardcoded to. Ours was the first time it had ever
been asked about anywhere else, and it was wrong both times it tried.

**Three findings, each of which made the fix bigger.**

*First, it is in five places, not one.* The original write-up named a single set of
instructions. I have to record how I established that, because I got it wrong myself,
twice: I searched for the offending words in lower case, and in two of the five files the
name is capitalised. So two whole sections came back clean when they were not — and I had
already told you "three places" and started building the fix around three. Re-running the
same search case-blind changed the answer. It is a recurrence of a rule we already have
written down, and it is logged where we log wrong calls.

*Second — and this is the one that matters — **the critics are contaminated too.*** The
agent writes a plan and a small council of critics then judges it. In the original write-up
I said the council was the part that worked, because it correctly refused the nonsense
plan. **That was too generous and I have corrected it.** Three of the four critics are
themselves told, in their own instructions, what the *other* site's data file is called,
what its core loop is, and what counts as a fabricated number *there*. So is the step that
rewrites a plan after it has been rejected — which is the step that produced the second
wrong plan in the run we filed from.

Two things follow, and neither is academic. The refusal we praised proves less than we
thought: a plan that wrong would have been rejected by a fair critic *and* by a
contaminated one, so the episode tells us nothing about whether the critics are sound. And
the danger now runs the other way. Once the planning instructions are fixed, a *correct*
plan for this site could still be objected to by a critic looking for a data feed and a
timed round this site was never going to have — and that would look exactly like the fix
having failed. Which is why the fix repairs the critics as well, and why the person
verifying it needs to be told this before they read the result.

*Third, the hardcoded facts have gone stale as well as being about the wrong site.* One
critic — the one that can veto a plan on its own, for dishonesty — is told the other site
has no verified facts at all, so any number in a plan must be invented. That was true when
it was written in July. It stopped being true at nine o'clock on the morning of 8 August,
when that site gained four verified facts. So a plan for that site is now told by its own
anti-fabrication critic that four real facts do not exist. Nothing goes back and updates a
fact pinned inside an instruction when the world moves underneath it. That is an argument
for the change we are making even if the original bug had never happened.

**Written the fix, and stopped short of switching it on.** It takes the site-specific
brief out of the shared instructions and puts it where it belongs — attached to the site,
as data. Each experience can have its own brief; the agent reads whichever one applies. If
a site has no brief, which will be the normal case, the agent is told so explicitly and
told not to borrow anyone else's. The other site keeps its brief word for word: it moves
house in the same single operation that removes it from the shared instructions, so there
is no moment at which it is lost.

Then I ran the whole thing against the real database with the final "save" turned into
"discard" — so every safety check, every edit and the verification actually executed, and
were then thrown away. It passed, and I confirmed afterwards that it had left no trace.
Two errors in the change were caught by that dry run and would have been caught by nothing
else.

## Where we are now

- **The site is finished and untouched by any of this.** Twenty-six pages in the new
  voice, all serving, all eleven calculators identical to the reference recording.
- **The fix is written, tested as far as it can be without spending a real run, and not
  applied.** It changes configuration only, so it goes live the moment it is run — no
  rebuild, no release.
- **It has never been run against a live planning job.** The dry run proves the change is
  mechanically sound. It does not prove the agent then writes a better plan.
- **The two wrong plans are still on file and still switched off**, demoted by hand when
  the bug was found. They are the evidence a fix should be proven against.
- **A second, separate fault is written up and untouched:** when the council rejects a
  plan, the rejected plan still becomes the official one. Nothing is building from the one
  it happened to.

## Where we are going

The next session applies the fix and spends a real planning run proving it — and the
proof has to be run in both directions. A check that only confirms our site no longer
mentions the other one would pass perfectly on a fix that quietly loads nothing at all.
So the other site has to be run too, and has to come out the *opposite* way. That
asymmetry is the whole of the test.

Then the second fault needs a decision rather than a preference. One route is a small
rewiring of the order things happen in; the other changes a piece of shared machinery and
should therefore go through review before it ships. Both are costed and neither has been
chosen, because choosing quietly is how shared machinery gets changed by whoever happened
to be passing.

What this stretch demonstrated, beyond a bug found and a fix drafted: the most expensive
error in the whole episode was not the hardcoded instructions — it was believing the
reviewing layer was clean because it had produced the right answer once. A correct verdict
from a compromised judge is still a compromised judge, and it costs nothing to check and
a great deal to assume.
