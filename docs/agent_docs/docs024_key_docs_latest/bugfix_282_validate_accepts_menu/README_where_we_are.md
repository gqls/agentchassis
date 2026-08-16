# Where we are — bug 282 (the planner's calculators were being deleted)

Append-only, plain prose, newest at the bottom.

## 2026-08-16

**The problem, without jargon.** When we rebuild a site, one agent proposes the
layout of every page, and a later step checks the proposal against the list of
components we actually have. Back in August we taught the proposer that it may
use a site's own calculators — but we never taught the checker that calculators
are a legitimate thing to propose. So on the loan-calculator site the proposer
duly put a calculator on each of the twelve calculator pages, and the checker
deleted all twelve on the way past. Nothing errored. The saved plan simply came
out describing twelve calculator pages with no calculators on them.

Nothing on the live site broke, because a separate rule stops us rebuilding a
tool page unattended, and the twelve calculators are locked. But the twelve
redesign tickets have been held since Friday waiting for this.

**What I checked before touching anything.** Bug reports go stale, so I re-proved
the whole story against the live database rather than trusting the file: the menu
the proposer saw (151 components, calculators included), what it actually wrote
(a calculator on twelve pages), and what the checker passed on (none of them).
The deletion happens inside the checking step itself.

**One thing in the bug report was wrong, and I corrected it.** It said a
different name had *survived* the same check, and concluded there must be a
second, hidden code path handling calculators differently. There isn't. That
name belongs to an ordinary section component someone created three days
earlier — so it passed for the most boring possible reason. Had I taken the
report at face value I'd have spent the afternoon hunting a branch that does not
exist. The check that settles it takes a fraction of a second and is now written
down.

**How I've fixed it, and why this way.** The obvious fix is to copy the "which
components count" rule out of the menu and into the checker. I didn't, because
that rule isn't code — it lives in the database as a piece of configuration, and
it had already been changed once since the bug was filed. Copying it would have
left us with three copies of one rule that must never disagree, which is exactly
how this bug happened in the first place.

Instead the checker now reads *the very list the proposer was shown*. There is
one list, in one place; if we widen what the proposer may use again, the checker
follows automatically, because there is nothing to remember to update.

It is off by default and switched on for one agent, so nothing else on the estate
changes behaviour. I proved that by deliberately breaking the switch and watching
the test that guards it fail.

**Where this goes next.** The change is committed and needs a chassis build to
take effect — configuration is live immediately, code is not. Once it has rolled,
the loan-calculator lane re-runs its recompose and we check that the twelve
calculators appear in the saved plan. Only then does the bug close, and only then
do their twelve held tickets get worked.
