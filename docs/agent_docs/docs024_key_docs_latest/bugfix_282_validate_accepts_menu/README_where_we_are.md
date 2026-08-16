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

## 2026-08-16, later — the review pushed back, and it was right

I put the change through the reviewer council, as the house rules require for
platform code. It came back **revise**: eight of the twelve reviewers approved,
four raised objections, and one of those objections was worth the round on its
own.

**The point it made.** My fix stopped the calculators being deleted, for the one
planner I switched on. It did nothing about the underlying habit: when this step
throws a section away — for any reason, on any site, for any planner — it writes
a single line to a log that rotates every ninety seconds and keeps no record.
A typo, a renamed component, a deleted one: all still vanish exactly as quietly
as the bug I was fixing. The reviewer's phrase was that I had treated it as a
one-caller problem when the platform's own history says the same shape comes back
somewhere else.

So the change now files a proper, durable record every time it drops a section,
for every planner, whether or not anyone opted in — reusing the machinery this
same step already uses for a neighbouring case, so it is not new plumbing. The
old behaviour is unchanged; what changes is that it is no longer invisible. That
is a better fix than the one I submitted, and it exists because the review
insisted.

**Three of the other objections were checkable, and all three came out clean** —
the migration reads the right part of the configuration (and would have refused
to apply if it hadn't), there is only one live copy of the planner's definition
so nothing ambiguous was edited, and the new setting is read as a plain value,
not as a lookup. The fourth was a fair catch on my evidence: I had cited a
database count using a column this system does not have. I re-ran it properly —
the answer was the same, but that is luck, not method, so I corrected the record
rather than leaving it.

**The most useful thing I found today was in my own work, not the code.** Two of
my tests could not fail. One passed even with the new feature's wiring removed —
it checked the shape of the output and never that anything was connected. The
other was supposed to prove that a clean run writes nothing, but the tool it used
can only detect a missing expected action, never an unexpected one, so it would
have passed against code that wrote a record every single time. Both are fixed,
and both are written up, because a test that cannot fail is worse than no test:
it produces the confidence without the check. That is the same failure the bug
itself was — something that looked like it was working because nothing could tell
you otherwise.

**One more, of the same family.** I wrote a recipe for confirming the fix has
reached the live system, then ran it with a control — and the control failed.
The method could not have told the difference between "not shipped" and "shipped
fine", because the thing it inspects records only one version marker, not every
change contained in it. It would have given me the right answer today by
coincidence. Corrected, with the failed control written into the recipe so the
next person inherits the test and not just my conclusion.

**Where it stands.** The configuration half is live now; the code half needs the
next build. Nothing about the fix is in front of a customer until then, and the
loan-calculator work stays held exactly as its own notes say. Whoever cuts that
build needs to raise the version number — rebuilding under the current one ships
the old binary from cache.
