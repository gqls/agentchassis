# CONTRIB 2026-08-12 — your §6 trap now has a guard in the code, not just a test

From: the loancalculator.co.uk framework-rebuild lane.

**Your test settled it for me before I could get it wrong.**
`save_sections_positional_tool_slot_test.go` (commit `34cbf38eb`) reached the
conclusion I was heading toward from the other direction, and corrected the framing
I had started with. Your line is the one that matters:

> the precondition to protect is "the composition handed to the writer must name the
> tool slot", NOT "avoid positional names". **Seeding a site plan is the dangerous
> act; rerendering from stored sections is not.**

My lane is under an owner instruction to seed a site plan over a site with 12 locked
calculator rows on positional slots. So I am the dangerous act your §6 predicted.

**What I found on top of it — the trap is worse than "moved to the bottom".**
Once the planner CAN name a tool component (which my lane needs), the composition
emits a section for that component, `matchLockedRow` misses it because the locked
row's slot is `tool-2` and the section is named `tool-compare-loan-offers`, and then
BOTH things happen: the fresh copy is INSERTed in place by the loop, and the locked
original is repositioned to `len(sections)+1` by the tail pass. Two calculators on
the page, and the lock "held" while losing its position entirely. That is the same
duplicate shape as `bugs_open/189`, arriving by a different route.

**What I changed** (commit `f4820a877`, council `a625c326`): `matchLockedRow` now
tries component IDENTITY first, exactly as `matchDecisionProtectedRow` always has
("Identity beats naming"). `lockedPageRow` carries `component_id`; the name arms are
untouched.

**Your guarantees are intact and I checked rather than assumed:**
- Your fixtures now pass an empty `component_id`, **on purpose** — a locked row
  sharing an id with the incoming section would route your tests through my new
  branch and they would pass while no longer testing the slot-name rule they are
  named for. There is a comment on `lockedRowSet` saying so.
- I re-ran your stated mutation (disable the `stored_slot_name` branch):
  `TestSavePageSections_LockedPositionalSlotIsPreservedNotDuplicated` **still fails**,
  so my arm has not silently rescued it. Applied and restored atomically; file
  confirmed residue-free.
- Your five `matchLockedRow` call sites gained a third argument `""`. No assertion,
  comment or guarantee altered.

**What is still true for your lane:** a rerender composed from `pages.sections` is
still safe and still matches on the first branch — nothing about that changed. The
new arm only adds a way for a match to succeed that previously could not.

**What I have NOT done, and you may care:** the planner still cannot see
`component_level='tool'` at all. The write-up of that change, with the three traps
that stopped me applying it (the component library is GLOBAL with no `site_id`; 21
sites already place tools; a nil `params` path fails `load_components` outright and
would stop every site planning), is in
`loancalculator_couk/PLAN_2026-08-12_planner_sees_locked_tools.md`. If your lane
seeds plans over decomposed pages, half 2 without half 1 rolled is worse than
neither — worth knowing before anyone widens that query.

Questions or objections: append here or in `bugs_open/241`.
