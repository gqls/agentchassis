# SUMMARY 2026-08-04 — the missing knob exists, and is one roll from working

## What we're trying to do

Give the platform a way to say that *one* step inside a loop may fail without killing the
whole job, while its neighbours stay strict. That sounds small. It is the difference between
"one page failed to save, so we shipped the other eleven and flagged that one" and "one page
failed to save, so the entire site build died".

## Where we've come from

We had exactly one dial, and it covered the whole loop. Either every failure inside the loop
was fatal, or every failure was shrugged off and the item skipped. Nothing in between.

Three days ago that forced a bad decision. Another thread needed one small step — writing
links into a table that regenerates itself and is currently empty — to be allowed to fail
without destroying a site build. With no way to say "just this step", they made the step
report success when it had not succeeded. Our review council rejected that, four seats
independently, and the objection worth remembering is the constitution seat's: it was *"a fix
whose rationale names the mechanism it steps around rather than repairs it"*. The workaround
was reverted and the gap was written up as a bug for somebody else to fix. Nobody had.

## What we've done

Built the knob. A step inside a loop can now declare its own tolerance; a step that says
nothing inherits the loop's setting exactly as before. It works in both directions, and the
second one matters as much as the first: as well as making a single step tolerant inside a
strict loop, you can now make a single step *strict* inside a tolerant one — several of our
dispatch loops currently swallow everything, and this lets a step that must not be swallowed
opt back out.

Three things are worth reporting beyond "it works".

**We found a trap nobody had recorded.** You could already write this setting on an individual
step. It was accepted, it passed the configuration audit as legitimate, and it was then
silently overwritten a few lines later. So anyone who tried the obvious thing got no error, no
warning, and no effect. Any configuration written that way has been inert since the day it was
written, and its author probably believed otherwise. That is now recorded as a landmine.

**We proved the tests can fail.** A test that passes whether or not the fix is present proves
nothing, and this estate has shipped that mistake before. So we deliberately broke the fix and
confirmed that exactly the right three tests failed, then restored it. That is the evidence
the fix is actually being tested, rather than the suite merely being green.

**The review council approved it first time**, with three advisory objections, none serious.
We answered all three by going and checking rather than by agreeing. One asked whether a
skipped step leaves any trace or just vanishes — it leaves three, and we can now say exactly
which. One asked about an interaction with two other settings; it turns out those two settings
are read by nothing at all anywhere in the codebase, so there is no interaction to have. The
third asked us to file a sibling defect as a proper ticket instead of burying it in our own
risk section, on the grounds that a deferral recorded only in the document that defers it never
reaches whoever audits the mechanism later. That was right, and it is now bug 193.

## Where we are now

The fix is committed, reviewed, approved and tested, and it is **not yet live**. The running
chassis predates it — we measured that at the pod rather than assuming it, with a control to
prove the measurement itself worked.

Our own rule is that a bug stays open until the fix is genuinely running, not merely committed,
so **bug 173 remains open**. That is deliberate. Making it live means rolling a new build to
the whole fleet, which is an owner operation and which interrupts around thirty other sessions'
work — not something one thread does on its own initiative for a change that is, by
measurement, inert until somebody opts into it.

The change affects nothing on the day it ships: of the 79 steps sitting inside loops across the
fleet, exactly zero currently use the new setting.

## Where we're going

Three things, in order.

1. **The roll**, whenever the next fleet build happens. Nothing special is needed — the fix is
   in the committed code and any build picks it up.
2. **The live test**, which is the only thing a build cannot give us: make the tolerant step
   fail for real and confirm the job carries on, then make the strict step fail and confirm the
   job stops. Both directions, or we have only tested the half we hoped for. Then 173 closes.
3. **Bug 193**, the sibling: the same setting read at the loop level still ignores a mistyped
   value in silence, where the new step-level code warns. Nothing is broken today — we checked,
   and all ten loops using it have it written correctly — but a mechanism that is loud on one
   side and silent on the other is a worse place to leave things than uniform silence, because
   the first thing anyone infers from "no warning" is "no problem".

**One question for the owner.** Two seats pointed out that the rule exempting this change from
a heavier architecture review was applied *by us, to our own change*. Both agreed with our
reading, one of them at length. But they are right that the author of a change is not the right
person to rule on whether it needs reviewing, and that is a judgement worth confirming rather
than inheriting.
