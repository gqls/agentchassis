# SUMMARY 2026-08-02 — the contact component that invented business facts, closed

## What we're trying to do

Stop the platform asserting things about a business that nobody ever told it.
The specific case: a shared `contact-info` component that drew Email, Phone and
Hours cards on a contact page, and — when a site had not supplied one of those
details — **made one up** and printed it in exactly the same style as the real
ones. Eight live commercial sites were publishing invented opening hours; one was
publishing `+1234567890` as a phone number, as a clickable `tel:` link.

## Where we've come from

Filed as `bugs_open/140` on 29 July by the oufe lane, against six sites, and left
unowned for four days. By this morning it was **eight** — two more sites had
walked into it while the ticket sat there, which is the argument for repairing the
mechanism rather than the pages.

The filing framed the fix as a policy change needing an owner's ruling, because it
would visibly change eight sites belonging to other lanes.

## What we've done

**Found that it was not a policy question at all.** The component's own
`input_schema` already declared `"on_missing": "skip_field"` for phone, hours and
address. The template disobeyed the contract it published. The schema even has a
first-class way to express a *legitimate* default — the section heading carries
`"fallback": "Contact Us"` — so the platform already distinguished inventing a
**fact** from supplying a **label**; only the template had lost track of it. That
turned an owner's judgement call into a defect with an unarguable fix.

**Fixed it at source** (migration 287): every card gated on its own datum, the
four invented literals deleted, so a detail nobody supplied cannot render. In
passing this repaired a second defect nobody had filed — the template read
`.title`/`.intro` while the schema declares `section_title`/`intro_text`, so
**eight bespoke headings and every intro paragraph were being silently
discarded** in favour of a hardcoded "Contact Information". Our quality checker
had recorded that desync on 18 May and nothing ever read the finding.

**Closed the two ways it comes back.** The detector actually named for this
defect — it raises alerts titled "Fabricated contact info" — contained not one
literal our own library ships; across the fleet its nine patterns matched **1**
row while missing **9** live fabrications. It was blind to its own platform's
output. It now knows them. And because any such list is only as good as somebody's
memory, a new standing lint (`check_placeholder_fallbacks.py`, **CGV-029**) reads
the **live component library** instead of a list, separating a fact default from a
label default.

**Council APPROVED at round 1**, seven advisory objections, none high-severity.
The checkable ones were answered with checks rather than argument: whether any
rows were locked (none), and two precedent claims I had asserted without evidence
(both verified). One objection I had to refute — it inherited a sentence of mine
that had gone stale between writing the submission and running the tool.

**Proved it on the artefact, then found the proof was not the mechanism I
thought.** This is the substantial thing that happened today, and it is a
correction to my own reasoning rather than a discovery about the bug.

## Where we are now

**Closed.** All eight sites verified clean, stored *and* served on the live web:
no fabricated hours, no fabricated phone, no `example.com`. **No Hours card exists
anywhere on the fleet**, because no site supplies opening hours — which is exactly
right. Each page shows precisely what its own data supports.

Two outcomes better than predicted, both checked rather than assumed:
`finetuning.uk` now renders a postal Address card — a field that was declared and
sourced but drawn nowhere, with data the site had all along — and `idea.uk`'s
phone is now backed by its own records, which resolves a stale-artefact concern
(`bugs_open/117`'s family) the ticket had flagged separately for that site. No
site lost anything real.

**The correction that matters more than the fix.** This morning I said the eight
pages would repair themselves as they rebuilt. They would not have. The backlog
drained — 294 queued jobs → 0 — **six contact pages were rebuilt and came back
with the invented hours intact**. A "rebuild" is two different operations: one
regenerates each section from its template, the other re-staples the page from
section HTML already stored, and **the second is the default**. Our own code says
so in a test message — *"deploys stale HTML"* — in a file I had already opened
that day for another reason. I had measured the queue rigorously and let that
stand in for knowing what draining it would do, which is a different question.
Seven jobs of the *first* kind then repaired all seven pages in twenty minutes,
the fabrication count falling in step with each completion.

That trap is now a fleet-wide landmine, because it applies to every template or
render-config fix anyone ships, not just this one.

## Where we're going

Nothing further is owed on this bug. Two threads leave it:

**`RFC_009`, open and deliberately undecided.** Two council seats independently
made the round's sharpest point: this is the **second** component caught
disobeying its own `input_schema` (the fallback footer was the first), and both
times the repair was a hand-edited template. Nothing between the declared contract
and the renderer enforces it, so a third component will do the same. The RFC
states the two hard questions rather than picking an answer — enforce at render
time (general, but can break live pages) or refuse at the write path (safe, but
only helps future components) — because that is a human call.

**A pod-grep owed at the next chassis roll.** The two new detector patterns are Go
and inert until then, and then only as far as `quality-discovery-agent` is driven
(`bugs_open/149`). They are a backstop against recurrence, not the fix, which is
why the bug closed without them.
