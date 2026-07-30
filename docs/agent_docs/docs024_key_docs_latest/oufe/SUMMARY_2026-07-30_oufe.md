# SUMMARY — 2026-07-30 — the claims floor lands, and the checker queue starts moving

*Written because C1 landed. The previous summary (`SUMMARY_2026-07-29_oufe.md`) was
written before the orphan discovery and predates the whole checker-queue thread, so
"where we are now" genuinely differs.*

---

## What we're trying to do

Two things at once, and they have drifted apart, which is worth saying plainly.

**oufe.com itself** is a grounded site about a narrow area of insolvency law. Its
whole point is that it can be checked: every figure carries a source, the tools say
out loud that they can be wrong, and the site is not permitted to publish a number it
cannot ground. That constraint is the product, not a limitation of it.

**The platform underneath it** has to be able to keep that promise on its own. A site
that is honest because a careful thread was watching is not honest; it is lucky. So
the second job — which has taken most of the last two sessions — is making the
machinery enforce what oufe asserts about itself, everywhere, without anyone
remembering to.

## Where we've come from

oufe went live with eight pages, a working tool, and a footer note admitting the site
can be wrong. Then the honest question — *where would a reader actually find the
second tool?* — turned up that nothing linked the **first** one. That opened a fleet
census (eleven unreachable tool pages across five sites), then the reason they were
unreachable at birth, then the discovery-checker layer that was supposed to catch it.
The owner asked for that layer to be written up as a workable queue. That is
`bugs_open/149`: twelve measured items, ordered, each carrying the query that found it.

The owner also overturned the headline of one of those items within hours, correctly,
and the rule from it is the one this lane now runs on: **a lack of evidence that
something works is not evidence that it doesn't — it may simply not have run.**

## What we've done

Today: the top of that queue.

**The claims gate the owner asked for is built.** Copy that the page pipeline writes
is now checked for claims at the point it is **saved**, rather than by asking each
agent's configuration to remember a validation step. Six agents save page content;
two of them were checking anything. All six are covered now, and so is the seventh
that nobody has written yet, because the check lives in the saving itself. A claim we
have ruled no site of ours may make — the "everything here is independently verified"
family — **refuses the save**. A number we cannot source is written down for a human
and allowed through, because that particular check is wrong often enough that blocking
on it would stall builds for no good reason.

**Two of the queue's own items were wrong, and re-measuring is what found both.**

- The bug file put the missing gate on `page-content-writer`. That agent **saves
  nothing** — it writes the words and hands them back to whichever of four parents
  asked. A check there would have inspected the copy in transit. We had made and
  corrected this same mistake once before, and the note of it was still in the code.
- The bug file said a claims detector had **never found anything, ever**, and argued
  from that zero that it needed re-siting. It had found two things — **two and a half
  minutes before we wrote that it never had.** The count was taken earlier in the
  session and written up at the end of it without re-running.

**And a smaller item alongside:** a discovery checker given a name that doesn't exist
used to shrug and carry on, reporting success with a quietly smaller set of checks.
From outside, that is indistinguishable from a clean site. It now stops and says so,
and reports which checks actually ran rather than which it was asked to run.

**The cost of the new gate was measured before it was proposed**, not left for the
reviewer to ask about: every piece of content we have — 949 across 14 sites — scanned
with the gate's own engine. **Three** would be refused. All three assert something
untrue; two of them are a bug we had already filed.

## Where we are now

The code is committed (`f61dce806`) and the image is built and verified to contain it
(`v1.0.1208`). It is **not live**: the fix is inert until the fleet is rolled, and the
roll is deliberately held because our own council review was still running and a roll
kills a review in flight. That is the single outstanding action.

`bugs_open/149` **stays open**, at three items of twelve. That is not incompleteness
for its own sake: several of the remaining items are decisions rather than code —
whether to delete six checks nobody runs or find them a home, and two changes to
shared defaults that warrant their own review round. Those are the owner's to call.

oufe's own site state is unchanged since the 07-29 measurements and has **not** been
re-verified today, so nothing from that list should be quoted without re-running it.

The honest summary of the lane's balance: **oufe's second tool is still not built.**
That has now been true across three sessions. Each deferral was individually right —
building a second unreachable tool would have doubled an invisible surface, and the
platform gap underneath was real — but the pattern is worth naming rather than
explaining again.

## Where we're going

1. **Roll the fix** once the council verdict lands, and prove it at the pod on both
   replicas rather than at the tag — an image tag is not evidence of a rebuild.
2. **The owner's decisions on the rest of 149**: seat-or-delete the six checks that
   are configured nowhere, and the two shared-default changes.
3. **oufe's second tool** — the relevant-alternative tool. Design is settled, the
   grounding citations already exist, there is now a footer slot to link it from, and
   the blocker that justified deferring it twice is cleared.

The thing this session actually bought, stated once: **the next time an agent writes a
false claim into a page, the platform stops it — rather than a person noticing.**
