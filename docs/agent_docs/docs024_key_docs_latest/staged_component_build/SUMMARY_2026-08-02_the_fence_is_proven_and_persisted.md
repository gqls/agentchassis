# SUMMARY — a component's contract now lives in the database, proven two different ways, 2026-08-02

## What we're trying to do

Give a *component* — a piece of a page shared across many sites, like a teaser panel or a
site header — the same machinery a *tool* already has: a written contract (a PLAN) in the
database, a history of verdicts against it (NOTES), and a real test that drives the
component in a browser and reports honestly when it is broken. The wider goal this sits
inside is `features_open/027`, a deliberately small three-gate version of what was
originally an eight-stage build ladder.

## Where we've come from

The database and the Go binary had both been widened to accept a component's travelling
docs, but that had only been proven by a throwaway probe row that was then deleted — real
evidence the *mechanism* worked, but not the same thing as a component actually having one.
No component had a real, persisted contract yet.

## What we've done

Picked a real, live component (`teaser-reveal-panel`, chosen earlier for having its whole
history already written down) and a real, live placement to test it against — the most
recently rendered of its five pages, on a second site from the one it was built for, which
is itself evidence the component is genuinely shared rather than merely shareable.

Read how the component actually behaves before writing a single check: what needs no
JavaScript at all (the core reveal, which is a native browser feature) versus what lives in
a shared script file used across the whole site (closing a sibling card, sharing a link to
one specific card). Wrote twelve checks against what the component actually promises, then
proved two different things about them: that they all pass against the real page, and —
the harder, more valuable proof — that each one can be made to fail when the thing it
checks is deliberately broken. The tool this lane already had for that second kind of proof
turned out to only really work for a different component built earlier, despite looking
general-purpose; rather than trust that, it was tested, found wanting, and a proper version
built for this component specifically. Every deliberate break was caught by exactly the
check meant to catch it.

Once both proofs were clean, the contract and its history were written into the database
for real — not the throwaway version from before — and then read back out again to confirm
what was written is exactly what gets used, not merely what was intended.

A fresh build of the platform's software was rolled out partway through this work, for
reasons unrelated to any of it. Checked rather than assumed: the database is unchanged, and
nothing in this lane's work needed anything from that new build.

## Where we are now

A real component now has a real, tested, persisted contract in the same place a tool's
contract lives, for the first time. That answers "can this happen at all" cleanly.

What it does not yet answer: whether the platform's automatic testing machinery can pick up
that contract and run it in the live cluster the way it already does for tools. The piece
that dispatches a test to the browser was built assuming the thing under test lives on
exactly one page, which is true for a tool and not true for a component that is deliberately
shared across several — this component sits on five. Teaching that piece the new question is
a small, genuine design decision, written down as two reasonable options, not yet chosen.

## Where we're going

A fresh conversation picks up that decision and the implementation it leads to, using a new
handoff written for exactly this handover. It is real platform code, in the area of the
codebase that has an advisory review process attached, so it deserves a session with full
attention rather than the tail end of an already-long one. Behind that, unchanged and
unaffected by any of the above, sit three smaller items: a backlog of components with no
written test at all, a filed-but-unowned question about documents left behind by renamed
things, and a checklist item now unblocked because a different team closed the bug it was
waiting on.
