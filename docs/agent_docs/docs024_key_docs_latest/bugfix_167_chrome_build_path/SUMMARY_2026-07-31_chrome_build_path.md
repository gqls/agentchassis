# SUMMARY — 2026-07-31 — the chrome build path: the last of 118's three questions

## What we're trying to do

Make the platform give **one** answer to one question — *which component from the
library may serve a site's header, footer or head?* — at **every** place that asks
it, and make a wrong answer impossible rather than merely unlikely.

## Where we've come from

`bugs_closed/118` established the predicate and fixed the two places that *assign*
chrome to a site. It deliberately left three more — the places that *build a page* —
and wrote them up as `bugs_open/167` with an honest reason: fixing them looked like
it would change the header and footer on every page in the fleet, and a
fleet-visible change should not ride inside a fix measured to have none. That was a
correct call and it parked the bug as an owner decision.

## What we've done

**We re-ran the measurement the parking decision rested on, and it had expired.**
Both predicates already return the same component for header and for footer. The
reason is that 118's own fleet repoint had switched on `header-theme-chrome` at
12:39 that afternoon — hours *before* it wrote 167 — and once it is active,
alphabetical order puts it ahead of the section-level impostor. So the fix costs
nothing visible, and the owner call it was waiting for no longer applied.

The three renderers now use the one predicate and, critically, use its answer
**only when it reports the component is eligible**. That gate is the whole fix. The
resolver always returns *something*, by design, and for `<head>` — which has no
eligible component in the library at all — that something is an 8,500-character
page section. The "one-line each" fix the bug file describes would have put a page
section into every page's `<head>`, creating the bug on the one slot that never had
it.

We also added the **door**: a scan that fails if a chrome function name is ever
handed to the section-shaped lookup again. 118 already had a scan for this class,
and it could not see any of these three — it matches hand-typed SQL and skips the
very file they live in. Each of its exemptions is correct, and each is exactly
where the next instance hid.

**The council returned REVISE, and both objections were right.** Five seats
independently asked whether the fix can reach served output at all, given a landmine
naming these three functions next to a caching layer. The answer is that the
landmine's *footprint* matched while its *mechanism* did not — this file never
touches the cached table and has no idempotence exit — but the submission had not
said so, and silence about a landmine that names your symbols reads as not having
looked. The gating objection was sharper: we had found a **fourth** unguarded path
and filed it honestly instead of guarding it, while three deployed sites are
serving a switched-off header through it *right now*. Honest filing is not a guard.
So that path now fails loud — it logs an error naming the site and component — while
still rendering what it renders, because repointing three live sites is a visible
change and remains a decision, not a bug fix.

## Where we are now

The fix and the guard are **committed** (`8b29404d6`, `11f8b9e08`), the full
package test suite is green against a clean `HEAD`, and round 2 is with the
council on the same trail.

**It is not live.** Go changes need an image rolled; we pod-grepped both replicas
and confirmed the fix is absent from the running `v1.0.1223`, with a positive
control in the same command so the zero means something. Another session was
mid-build at the time and builds take committed code, so it should ride along —
but that is an expectation and the runbook carries the command to verify it.

`bugs_open/167` has been closed and moved to `bugs_closed/` as instructed. Strictly
that is ahead of this repo's own bar, which is *fixed **and** live*; the file says
so at the top, in bold, with the reopen condition.

**One thing needs a decision, and it is the only thing we are asking for.** Three
deployed sites — ai-agent-orchestration.com, finetuning.uk and gaswholesalers.com —
are pinned to a header component the library says is switched off, and have been
rendering it on every page build. Correcting that moves them onto a different,
smaller header: visibly different markup on live pages. It is written up as
`bugs_open/170` with the query, the four affected rows, and one trap for whoever
takes it — the fourth site pins its *own* fork, which is correct and must not be
"fixed".

## Where we're going

Nothing further is owed in this lane once the verdict lands and is acted on. Three
things sit downstream, all filed and none of them ours:

- the roll, after which the pod-grep in the runbook proves it shipped;
- `bugs_open/170`, the pin path, waiting on the decision above;
- `bugs_open/166` and `bugs_open/149` — the repair that could not repair, and the
  page-rerender queue — both live lanes with their own sessions, and both the
  reason a fix in the binary and a page on the internet can disagree for days.
