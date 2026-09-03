# 427 — where we are, 2026-09-03 (third read-out today)

Written after the lane was resumed from `HANDOFF_2026-09-03d_continue_here.md`. A third
summary on one day is unusual and the cadence rule says only write one if the five headings
would genuinely differ. They do: `SUMMARY_2026-09-03b` said *"proven live, mechanism bug
closed"*. That turned out to be true about the artefact and wrong about the mechanism, and
the thing it named as the closing signal cannot fire.

## What we are trying to do

Make boxingonline.com's fight calendar a real, working calendar — the site's own promised
feature, on a paid deliverable the owner has ruled must be fixed before delivery. Underneath
it, close the general gap: nothing on the estate turns a confirmed real-world event into a
dated, correctable fact.

## Where we have come from

The lane found the empty calendar, built an `event-list` component, got a real fixture onto
the page, and proved the deploy all the way to the served bytes. It then found — correctly —
that its own three migrations were **transient**, and escalated that as needing "a person,
not a session", because correcting the store it had missed appeared to require an owner
ruling about plan immutability.

## What we have done

**Established that the transience was real but its mechanism was misidentified.** A re-plan
is *safe*: the planner snaps a deployed page's realised sections onto the plan proposal
before writing. What reverts is an ordinary **page build**, through a different code path
entirely, needing no re-plan. That distinction is the whole of the remedy — and it turned an
owner escalation into a one-page migration with a precedent already in the tree from July.

**Made the fix durable.** Migration `750` aligns the plan authority to the live page, shaped
as an in-place rename rather than the obvious delete-and-reinsert, because `ordering` is a
positional join key for four things — including section imagery, which binds to the ordinal
and not the component name. Applied after an induced-failure run; the artefact is
byte-identical afterwards, which is the point. Council-approved, round 1.

**Found that this has already destroyed work elsewhere.** Two live sites lost sections to
this mechanism while the detector's own warning sat open and unactioned — one of them losing
the very component the July migration was written to rescue. Filed as `bugs_open/469`.
Migration `753` cleared the five-week backlog of those warnings, recording *which side won*
so a destroyed correction is never closed as a success.

**Asked the real question.** `RFC_064`: may a non-planner action correct the current plan's
section rows for one page? There is no sanctioned way to do this today, so five hand-written
instances have accumulated in seven weeks — two of them on the same day, by different lanes.

**Coordinated rather than measured past people.** The `apis.uk` lane was one build away from
losing its own work; warned, they fixed it within the hour using `750` as a template, and
returned a finding that improved the permanent design.

## Where we are now

The fixture is on the page and is now safe from the revert. The fleet backlog is clear. The
architecture question is filed.

**And the page is still not a tool.** The nightly experience check ran four minutes after the
deploy and still reports *"a page about a tool, not a tool"* — because `event-list` is a
static template by design, and the page is classified as a tool, which is held to needing a
control, inline data or a runtime fetch. The signal the last summary named as "how we will
know this is finished" could never have fired. That is the correction this read-out exists to
make.

## Where we are going

The owner has ruled: build a real calendar mechanism. That is dispatched through the tool
pipeline, and it had to come **after** `750` — the only reason the page had not already
reverted is a build guard that switches itself off the moment a real tool component lands, so
building first would have armed the very revert we just removed.

Then: `RFC_064`'s answer, which decides whether the typed action gets built and retires the
hand-written migration for everyone; and `bugs_open/469`, where a human has to say whether
two destroyed sections should come back.
