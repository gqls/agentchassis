# SUMMARY — P1 and P2 are done and proven live; P3 turned out stale and is now a measured backlog, not a build

## What we're trying to do

Give a *component* — a piece of a page shared across many sites, like a teaser panel or a
site header — the same machinery a *tool* already has: a written contract (a PLAN) in the
database, a mutation-proven test of that contract, and an automatic check that drives the
real page in a real browser in the cluster and reports honestly when it's broken. The wider
goal is `features_open/027`, a deliberately small three-gate version of what started as an
eight-stage build ladder. The gate count was cut on purpose (owner ruling, PLAN D8,
2026-07-31): only three things get machinery — the claim written before the build, mutation
discipline on every check, and verification through a real visitor gesture. Everything else
in the original eight-stage sketch is unfunded, not disproved, and stays that way without
evidence.

## Where we've come from

By the start of this stretch of work, the database and the Go binary had both been widened
to accept a component's travelling docs, but only a throwaway probe had exercised that —
real evidence the mechanism worked, not evidence any component actually had a contract. And
even with a contract, nothing could get it in front of a real browser: the piece that
dispatches an automatic check assumed the thing under test lives on exactly one page, true
for a tool and false for a component deliberately shared across several. `teaser-reveal-panel`
sits on five pages across two sites for exactly that reason.

## What we've done

**Gave `teaser-reveal-panel` a real contract.** Read how the component actually behaves
before writing a single check against it, wrote twelve checks against what it actually
promises, and proved two things about them: they all pass against the live page, and — the
harder proof — each one can be made to fail on purpose. The tool already in the lane for that
second proof turned out to only really work for a different component built earlier, despite
looking general-purpose; tested rather than trusted, found wanting, and a proper version was
built for this component specifically. Wrote the contract and its history into the database
for real, then read it back out to confirm what's stored is exactly what gets used.

**Closed the dispatch gap.** The piece that sends an automatic check to the browser could
only ask "which page" the way a tool asks it — one name, one page. Rather than teach that
existing piece a second way to answer (which every tool's own testing already depends on),
built it a twin that resolves a component's page from an explicit, database-checked
placement instead of a name. Reused the old piece's own machinery everywhere the two
genuinely do the same work, so they can't drift apart. Proved it twice: once as code (it
compiles, the existing tests still pass, and the old path behaves exactly as before), and
once for real — sent a live request into the cluster asking the panel's actual page to be
checked. All fifteen checks that could run passed. Alongside it, in the same request,
deliberately pointed the same machinery at the wrong page — a real page, just not the one
with this panel on it — to prove it refuses rather than silently testing the wrong thing. It
refused, with exactly the message written for that situation, which is what makes the refusal
mean something.

**Then, asked what's next, checked before building it.** The plan's own next phase said
"build the remaining gates" — but that line was written the day *before* the owner's ruling
that retired exactly that approach, and nobody had gone back to reconcile the two. Reading the
dates side by side settled it: the plan's next phase was stale, not just old. Rather than
build the wrong thing, re-labelled it as what it actually collapses into — applying the three
gates that *are* funded to more components and tools — and measured the size of that instead
of starting on it. Of 49 active tools, 36 have no written contract at all; of 112 active
section components, 111 don't, including two smaller open questions surfaced along the way and
deliberately not chased further (a handful of contracts that don't match any known component,
and whether some component types are even meant to be covered by this yet).

**Checked two side items before touching either, rather than assuming this lane owns them.**
One — a filed question about renamed things losing their paperwork — was already filed by
this same lane and is still nobody's job; left it as it stood. The other — a safety check that
most existing contracts don't yet use, now that the bug preventing its use is long fixed —
turned out to affect thirty contracts belonging mostly to *other* work, not this lane's own.
Measured the size of that gap too, and filed it as its own tracked item rather than editing
other people's contracts without being asked.

## Where we are now

The two pieces of machinery this phase set out to build — a real, tested contract for a
component, and a real way to check it live — both exist and have been proven working, not
merely written. Nothing is broken, nothing is blocking. One thing is still genuinely
in-flight: the review this kind of platform change goes through was submitted and hasn't
come back yet. Everything else that's open is now honestly sized rather than vaguely "owed" —
a measured backlog of contracts still to write, and one small filed item for a safety check
still to spread — with the scope decision on both left where it belongs, which is not with
one thread unilaterally deciding to take on a hundred-plus subjects or edit other lanes' work.

## Where we're going

Read the review's verdict when it lands and act on it if it asks for changes — that's the one
thing actually pending. Past that, the path forward is a scope decision rather than a
technical one: how much of the measured backlog to close, in what order, and whether the
filed safety-check item gets picked up by this lane, by the lanes that own the affected
contracts, or left for later. Nothing about the machinery itself is in question any more —
what's left is deciding how much of the fleet to point it at.
