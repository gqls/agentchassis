# The concept register — what we set out to achieve, where we are, what's next

*2026-07-18. The calm, read-aloud version. Companions: the 2026-07-16/17
summaries in this directory. Technical detail: `PLAN_concept_register.md`,
`DESIGN_relevance_filter.md`, the pilot docs, and the turn-by-turn
`RUNNING_NOTES_concept_register.md`.*

---

## What we set out to achieve

This platform fixes its own bugs: a diagnosis loop works out what's wrong with
citations a human can audit, a proposer drafts a constrained fix, and a review
panel judges that fix before it can become a real pull request. When we
started, that panel had two members. The goal of this whole workstream was to
grow it into a proper council — and to choose each new member from evidence,
not guesswork. The evidence base is the concept register: sixteen hundred
distilled lessons from the platform's four thousand documents, every one
checked against the real running code.

## Where we are now

The council has grown from two reviewers to **nine**, and — just as
importantly — it's grown *smart* about when each one speaks. Two members
always review every fix: one for craftsmanship, one for safety, and the safety
member now carries your standing preference that battle-tested core machinery —
the orchestrator, the messaging backbone — should be left alone in favour of
fixes at a higher layer. The other seven are specialists who are woken only
when a fix actually enters their territory: a historian for a bug shape that
has struck seven times, a reuse checker, a rules referee, a tooling steward, a
guardian for how new sites are adopted, one for the diagnosis machinery's own
honesty gates, and — new today — one for the self-improvement loop, defending
the safety caps that were bolted on after that loop once ran away with itself,
generating eight hundred findings in ten days.

Each seat was installed the same careful way: grounded in specific documented
lessons, applied to the live system with a backup taken first, verified
afterwards, and — since we discovered another team co-editing the same
machinery — changed with surgical precision so nobody's work overwrites
anyone else's.

## The debugging seat, and your multi-model idea

Next on the list are three more specialists, ending with the **debugging
historian**. Worth being clear about what that seat is: it's a *reviewer*. It
reads a proposed fix, consults the platform's largest body of hard-won
debugging lore — seventy-four documented lessons — and asks one question: "has
anything like this happened before, and does this fix account for what we
learned?" It's cheap, advisory, and wakes only when relevant, like the other
specialists.

Your idea — spinning up several different specialised models, from different
companies, with different strengths, to attack bugs the current loop finds
intractable — is something genuinely different, and I'd recommend treating it
as **its own subproject** rather than folding it into this seat. Here's the
distinction: council seats *judge fixes that already exist*. Your idea
*generates diagnoses that don't exist yet* — it belongs at the other end of
the pipeline, at the moment the diagnosis loop honestly gives up. Today, when
the loop exhausts its attempts or can't verify a theory, it hands the bug to
you. Your idea slots exactly there: before (or instead of) reaching a human,
fan the same evidence bundle out to a diverse panel of models — different
vendors, different reasoning styles — and see whether any of them cracks what
the house model couldn't. The plumbing is closer than it might seem: the
platform already routes different models per workflow step, and the register
records both what exists (an Anthropic client, a local-model client) and what's
missing (clients for other vendors, and a recently-found bug in how truncated
model responses are detected — which that project would want fixed first).
I've written the idea up in the plan as a proposed subproject with that
grounding, so it's ready to be picked up whenever you want to start it.

## What's next

Per your direction, the roster continues in order: the compliance eye — the
seat justified not by frequency but by severity, after two live incidents of
fabricated content — then the rendering guardian, then the debugging historian.
Each is now a small, safe, well-rehearsed addition. After that, the council
side of this workstream is essentially complete, and the interesting open
threads are: watching the council handle its first real case end to end,
the neighbouring team's work to put *every* platform change through this same
council, and — if you choose to green-light it — the multi-model diagnosis
gauntlet as the next new build.
