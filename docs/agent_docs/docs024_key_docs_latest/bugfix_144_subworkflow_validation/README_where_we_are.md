# Where we are — sub-workflow validation (bug 144)

Plain prose, append-only, newest at the bottom.

---

**2026-07-29, afternoon.** Picked up bug 144 because nobody owned it and it is the
kind of bug that costs a bit of everything: not dramatic, but sitting underneath a lot
of other work.

The short version of what was wrong. When one of our agents runs a workflow, the very
first thing the system does is check that workflow makes sense — every step names a
real action, nothing points at a step that doesn't exist, there are no loops that
would run forever. That check has always run once, over the top level. But a lot of
our workflows contain a loop, and a loop contains its own little workflow inside it —
"for each page, do these five things". Those five things were never checked by
anything at all. Eighty-five of them, live, across eighteen agents.

The part that made it survive so long is worth saying out loud, because it will happen
again in some other corner. We have a second, offline report that answers the same
sort of question — which settings do our live workflows actually use? — and it was
written to look in exactly the same place: the top level. So the two halves were blind
in the same direction, and whenever anyone cross-checked one against the other they
agreed. **Two things that are wrong in the same way look exactly like two things that
are right.** It was eventually found sideways: a reviewer objected to a completely
different claim, on the grounds that it had been measured with a query that only
looked at the top level.

What I did about it. The validator now goes into the loops. The offline report now
uses the *same* piece of code to walk the workflow, rather than its own copy — so the
two cannot quietly drift apart again; if one goes blind, the other stops compiling.

The interesting bit was the risk. The bug file was blunt about it: if you just switch
the checks on for eighty-five steps that have never been checked, you may start
refusing to run workflows that work perfectly well today. So before writing the fix, I
pulled every live definition out of the database and ran the proposed new validator
over all 178 of them. Nothing new is refused. To make sure that stays true for the
next person, the harness that does this ships with the change: it validates each
workflow twice, once as it is and once with the loops removed, and only complains
about the *difference* — so somebody else's pre-existing problem can never be charged
to your patch.

Two things I only got right because I read the code that actually runs the loops
rather than trusting the bug report's suggested fix. First, a step inside a loop is
allowed to point at a step outside the loop — that's how a loop hands back to the
main workflow — so treating that as an error would have broken things. Second, the
loop runner quietly ignores several fields you might write on a nested step; so
rather than pretend to enforce those, the validator now tells you they are being
thrown away. Both of those would have looked fine in testing and bitten later.

One side-effect worth flagging: the new harness turned up three agent definitions that
are refused by the validator *today*, before any of my changes — they name actions
that no longer exist anywhere in the code. They're from a retired generation of the
pipeline that was never switched off, which a closed bug (044) already noted. What's
new is that three live builders still try to call two of them. I've filed that
separately rather than fixing it here.

And a housekeeping note: the repository as committed doesn't currently compile in one
corner (`cmd/reasoningset`, a reporting tool somebody is mid-way through). It doesn't
affect any of the images we deploy — those build only the service they need — but it
does break the quick "did I just break the build?" check that all of us rely on. I've
left it alone rather than rewriting someone else's half-finished work.

The change is committed and has gone to the review council. It will start doing
anything the next time a chassis image is built and rolled.
