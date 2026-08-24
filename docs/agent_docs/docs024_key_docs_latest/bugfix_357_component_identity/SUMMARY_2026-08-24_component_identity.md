# SUMMARY — 2026-08-24, component identity (`bugs_open/357`) — ARMED

## What we are trying to do

Stop the platform lying about what it has stored. Every page section is kept as a
row saying which component built it, next to the markup it actually serves. On
twenty-two live pages those two things disagree completely: the row says "I am the
shared hero banner" and the markup is an entire interactive calculator. Nothing
errors, because nothing compares them — so a checker complains that a page is
missing its headline while the page serves its own headline, and no repair dares
touch the row, because the only repair available would regenerate a banner over the
tool.

The owner ruled for the durable answer rather than a patch: **record which component
actually produced a section's bytes, at the moment it is produced**, then fix the
mislabelled pages once that record exists.

## Where we have come from

A record was built and approved on 22 August. It ran for a day writing nothing —
its value was being dropped at a hop between the part that produces it and the part
that stores it, and both ends had been verified while the middle was assumed. That
was found on the 23rd and fixed as a *contract* rather than a key, because the same
hop had swallowed a different field before and the per-field remedy taken then was
precisely why it swallowed the next one.

The fix itself had been stopped twice by review, both times correctly: once as
unsafe (it renamed a slot, which makes the next rebuild duplicate the tool) and once
as useless (it recorded the problem and stopped nothing). That deadlock — the safe
plan and the effective plan could not be the same plan — is what sent the question
to an architecture ruling in the first place.

## What we have done

**The record now works, and it is measured rather than believed.** It went live
overnight. Sections written before about nine this morning carry nothing; every hour
after is at or near complete — 46 of 48, 117 of 119, 24 of 24, 58 of 58. The control
holds in the other direction: of the ~987 sections written *before* the release,
**none** has a record, because nothing backfills. And it proved its worth within two
hours: six sections looked wrong at first glance, and turned out to name the template
that genuinely produced them against a component somebody edited twenty minutes
later. Before this, those six were indistinguishable from sections built with the new
version.

**The deadlock was broken by separating two things the platform had been treating as
one.** A section has a *slot* — where it sits on its page — and a *component* — what
made it. All the damage is on the component; all the danger is on the slot. So the
fix corrects what a row IS and never touches what it is CALLED. A page fragment that
nothing can identify is now attached to a component that provably reproduces it —
proved by rendering and comparing byte for byte — or left honestly unlabelled. Never
labelled by position.

**And today it was armed.** Both halves are through review. The component was seeded,
then the switch was turned on across every live pipeline that can produce these rows.

## Where we are now

**The mislabelling has been stopped at its source.** That is the sentence that could
not be written yesterday, and it is the whole point of the lane.

Arming went wider than this lane's own runbook proposed, and the reason is worth
recording because it was a measurement rather than a judgement: the runbook said arm
one pipeline as a canary — the obvious producer of these tool pages. Enumerating the
surface first showed that **five** of the six pipelines can produce them, not one,
because the path that mints these rows is reached by any builder whose structured
data comes back empty. Arming the obvious one would have left four still minting,
which is the exact shape review keeps catching in this estate: one call site gets the
rigorous fix while the mechanism stays generic everywhere else.

The twenty-two existing rows are untouched and still wrong. Their repair is written
and deliberately unapplied — it may only run once the record is demonstrably working
on a live page of the shape it creates, and it enforces that precondition itself
rather than trusting a runbook.

Nothing about what any page SERVES has changed, in either direction. The bytes were
never touched.

## Where we are going

1. **Watch the first adoption.** Not "did the config apply" — that is verified — but a
   real page coming through and landing correctly typed, regenerable and stamped, with
   its slot name unchanged. Three things would say stop, each invisible to the obvious
   "is the tool still there?" check: a page gaining a row, a mislabelled row acquiring
   the *wrong* record, or a new fragment landing with no component at all.
2. **Then the twenty-two.** Re-counted on the day rather than trusting the number,
   because the population minted twelve rows in a single day last week.
3. ~~**A decision for the owner on six of them** — pages marked as claimed by a human,
   stable since June for exactly that reason. The repair names them and skips them.~~

   > **CORRECTED 2026-08-24, same day — and the correction inverts the point.**
   > `rebuild_policy='owned'` does **not** mean a person claimed the page. The guard's
   > own words: such a page *"belongs to a tool/widget or is a runtime-fill shell"*,
   > and the flag is written in code (`create_report_page_action.go:176`); **172 of 704**
   > pages estate-wide carry it. I inferred a meaning from the column's name and never
   > read its definition, then escalated that inference into a decision for the owner.
   >
   > The consequence is not a smaller point but the opposite one: these six are the
   > **only** rows the producer fix can never heal, because the owned-page guard
   > returns at `save_page_sections_action.go:186` and adoption runs at `:397` — the
   > save is refused two hundred lines before adoption is reached. They were the rows
   > I was least willing to touch and they are the only ones that cannot fix
   > themselves. On the owner's instruction they are now **included** in the repair,
   > which targets all 22.
4. **Still open, and named so it is not mistaken for covered:** the rerender path
   re-derives a component from the slot name when none is recorded, so a fragment that
   fails adoption can still be re-labelled by it. And five other pieces of code write
   these rows with no watch on them at all.
