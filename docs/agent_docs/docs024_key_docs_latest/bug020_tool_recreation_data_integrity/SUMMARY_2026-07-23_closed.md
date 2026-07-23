# SUMMARY — bug 020 CLOSED (2026-07-23)

*Final milestone read-out. The two-part fix is live, verified, and the case is closed.*

## What we were trying to do

Stop the platform inventing data. When we adopt a site whose interactive tool depends
on a real dataset (a searchable directory, a data-backed chart), the recreation must
load from the *same* source — never fabricate a realistic-looking dataset to make the
widget appear to work. Bug 020 was the case where it did exactly that and shipped fake
vet practices and postcodes to a live public site.

## Where we came from

The vetcomparison thread found and contained it and filed a precise diagnosis. We built
a two-part fix: a prompt contract (live since 21 July — never invent records, preserve
the source, honest empty state) and a mechanical gate in code that does not rely on the
model obeying. The gate went through the reviewer council, which made it materially
better — most importantly, it caught that the gate was "failing open" (silently passing
when it couldn't read its input), which we changed to "fail safe" (hold for review).

## What we did

We wired the gate on when its build shipped — and then did the one thing that mattered:
we tested it with a deliberately-fabricated tool instead of trusting the wiring. That
test caught a real, serious bug — the gate was being handed the tool's *contents* where
it expected the *location* of the contents, so it was inspecting nothing and would have
waved everything through (a silent no-op that looked perfectly wired). We caught it only
because of the earlier "fail safe" change and because we insisted on inducing the fault.
We switched the gate off, fixed the bug, added a regression test, and waited for the next
build. On this build we switched it back on and re-ran the same test — and the *real*
detection caught the fabrication by name and held it for review, never publishing it.

## Where we are now

Bug 020 is **closed**. Both halves are live and verified: the prompt contract, and the
mechanical gate (fixed detector on chassis v1.0.1150, wired into tool-recreation-handler,
proven end-to-end by the induced-fault test). A fabricated tool driven through the live
system is now stopped before it can go out. The case file has moved to `/bugs_closed/`,
and the tool-imagery hold that was "waiting for 020" has met its condition.

## Where we're going

Nothing further on 020 itself. Two adjacent items were deliberately left as follow-ups
and are noted in the plan: a machine-readable no-fabrication site flag (candidate 4), and
the fact that tool-recreation's `validate_tool` swallows all validation blockers and
deploys anyway (a separate, pre-existing concern). The transferable lesson — don't trust
a detector until you've made it catch something real; test the layer that actually runs,
not just the pure logic — is logged in WRONG_CALLS and the debugging guide's spirit.
