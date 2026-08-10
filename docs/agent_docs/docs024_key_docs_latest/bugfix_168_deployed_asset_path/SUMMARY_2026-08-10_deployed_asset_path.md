# SUMMARY — 2026-08-10 — the retraction sweep is closing real findings on its own, and an owner ruling changed what "closed" is allowed to mean

*Written to be read aloud. Previous: `SUMMARY_2026-08-06b_deployed_asset_path.md`.*

## What we're trying to do

Every customer site is scanned automatically, and problems get filed as items in a queue for a
person to look at. The flaw this lane exists to fix is that **nothing ever re-checked them**. A page
gets flagged in July, somebody fixes the page in August, and the warning sits there indefinitely
saying something that stopped being true. Findings accumulate, the queue stops meaning anything, and
the real problems are buried among the stale ones.

So we are building the other half: a daily job that re-reads parked findings and withdraws the ones
that no longer apply. We call it retraction. The rule throughout has been that it may only ever
close on **positive evidence** — it re-runs the original check and closes only when the check now
comes back clean. Anything it cannot answer stays open.

## Where we've come from

The sweep existed but was starving: it selected the oldest parked rows of any type, and most types
had no way to be judged, so the same unjudgeable rows filled every batch and the ones that *could*
be closed never got looked at. That was fixed at the start of this month, and the sweep has been
draining steadily since.

But it could only judge four kinds of finding. Everything else it loaded, failed to judge, and left.
The work since has been widening that — one category at a time, each one requiring proof that
nothing else already closes it, because this lane once shipped a duplicate closer whose central
claim was false and which fourteen reviewers accepted.

Two days ago we added the fifth category: pages whose copy reads machine-written. Yesterday we added
the sixth: pages making factual claims the site cannot support.

## What we've done

**The fifth category is proven, behaviourally.** On 9 August at 08:38 the sweep ran on its schedule
with nobody watching, re-read all 32 parked "reads machine-written" findings, and closed one — a page
on the leopardess site that had genuinely been rewritten since July. That was the first time anything
had ever closed one of those.

**And the check we had written for ourselves would have hidden it.** The handoff instructed the next
session to confirm success by watching a number called the uncovered backlog fall by about 32. It did
not move — 625 before, 625 after. That number is a total across roughly forty categories; ours
dropped out of it while others grew by exactly the same amount. Following our own instruction
literally would have recorded "the change did nothing" on the morning it worked perfectly. The
correct check is the per-category breakdown. This is written down now so nobody inherits it.

**The sixth category is live too, and it is closing things.** This morning's sweep scanned 30 of
those findings and closed 8. It is the fastest any category has drained.

**And an owner ruling changed the terms it closes on.** The review council objected twice, in
successive rounds, from three independent seats, to the same thing: this category was made
human-only *on purpose*, because it is about factual claims, and letting a scheduled job close
those unattended is a policy change rather than a bug fix. The sharpest form of it —
**the register proves provenance, not correctness**. Our machine can confirm that a number now
appears in a site's approved-facts register. It cannot confirm the number is true. So a careless
register entry would have quietly retracted a live warning about a possibly-false claim.

The owner was given four costed options and chose the strictest workable one: **close only when the
page's own copy has been edited since the finding was raised**. A register edit alone can no longer
close anything. It is about a dozen lines, and it converts a policy argument into something the code
simply will not do. That is now shipped, live, and proven at the running binary.

**Three corrections went onto the record**, because a lane's value is as much in what it retracts as
in what it builds. We had been telling the owner the duplicate-rows backlog was worsening at about
two clashes a day — it fell by eight in fourteen hours, and the trend we had drawn through four noisy
measurements did not survive. We had recorded the wrong chassis version. And we had claimed this
change covered a "two producers" case, invoking an owner ruling that only applies to two producers,
when there is only one — a false claim that had already spread into the concept register, which the
next review round reads as ground truth.

## Where we are now

Both new categories are live and working. The sweep now judges six kinds of finding, scanned 243
items this morning, closed 37, and is not hitting its budget cap.

**The owner's gate held on every one of the eight closures** — all eight were pages whose copy had
genuinely been edited since the finding was raised, spread across ten days and several sites, so
these are real fixes rather than one bulk rebuild sweeping the board.

**But the gate has never actually refused anything, and we should not pretend otherwise.** Zero
findings have been stopped by it. On today's population, every page that scanned clean also happened
to have changed. The gate cost nothing and blocked nothing; it is proven present and proven correct
where it applied, not proven to bite. Its arms are pinned by tests rather than by observation.

**Two seats named a real limitation of it, and they are right.** The timestamp we compare is
component-level, not claim-level. An unrelated edit to the same section satisfies the gate even if
the disputed sentence was never touched. We checked the obvious way this could go wrong — a bulk
rerender bumping everything at once — and it is not what happened here. The limitation stands
regardless, and it is the natural next thing to tighten.

**The same hole is still open on the fifth category**, deliberately. Loosening a site's tone settings
would retract findings whose copy never changed. It is live and already reviewed, and the reviewers'
own argument was that tone is a lower-stakes surface than truth — so extending the gate there is a
separate decision, written down rather than quietly done. One reviewer has noted, fairly, that we now
have two different closing standards inside one shared mechanism.

**The council has said revise three times.** Each time it was right, and twice it caught something
that mattered: a false factual claim in round one, and the policy question in round two, which it
insisted was the owner's rather than ours — and insisted again when we treated it as a courtesy
notification instead of a gate. Round three's objection is that our *submission* under-described a
change we had actually made correctly. That is three rounds of being caught on precision, which is
worth saying plainly.

**A structural finding is still unresolved.** The sweep can only ever see two of the eight states a
finding can be in, and — worse — the report we use to measure how much work remains uncovered
filters the same way, so it cannot see them either. It has been telling us 625 when the real figure
is closer to 1,100, and it does not only omit whole categories: one category is *listed* at 50 while
46 more of its rows sit invisible. We put this through the diagnosis loop twice rather than assert it.
Both runs confirmed the mechanism and neither could read the specific list, because the code index
does not index that kind of code at all — a gap that affects every status list and registry in the
codebase. We stopped at two runs and recorded the pre-flight check.

## Where we're going

Nothing is half-applied and nothing is blocked on a decision.

The immediate work is round three's answer: file the under-described change as its own edit, and give
the compliance seat something better than our own word for the owner's sign-off — that seat cannot
see the ruling from the data available to it, which is a real structural gap and not merely a
formality.

Then, in rough order of value: tighten the gate from component-level to claim-level, which is what
two seats asked for and what would make it bite rather than merely hold; decide whether the tone
category gets the same gate, and resolve the two-standards asymmetry before a seventh category
arrives; and pick up the diagnosis on the invisible backlog, which is the largest single thing
standing between this sweep and the queue actually meaning something.

One idea was captured rather than built: **dating pages when they were last checked**. It is the
question none of this machinery answers. Our gate proves the page moved; nothing anywhere records
whether we ever looked. An empty review queue today means either "everything is fine" or "nobody
checked", and we cannot tell which — the same ambiguity, one level up, that we spent this week
eliminating inside individual page scans.
