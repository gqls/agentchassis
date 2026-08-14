# SUMMARY — 2026-08-14 — the evidence pack no longer starves its own diagnosis

*Written to be read aloud. Current state only — the chronology is in `README_where_we_are.md`, the
technical log in `NOTES_…`.*

---

## What we're trying to do

Two of our sites were shipping pages with no hero image and no logo, and saying nothing about it.
The job was to make that failure audible, then to find its cause. Along the way the job grew a
second half we did not choose: the automated diagnosis system we use to answer such questions
turned out to be unable to answer them — not because the questions were hard, but because the
machinery that assembles its evidence kept quietly dropping the evidence. So this workstream has
really been two things: the hero-and-logo bug itself, and a repair of the diagnosis tooling that
every future investigation will also lean on.

## Where we've come from

As of the last summary (the 12th), the silent failures had been made audible — three code paths now
record a durable row instead of shrugging — and we had found and fixed the first tooling fault: the
evidence pack could not read any function belonging to a type, about a quarter of the codebase,
because two halves of our own tooling spelled function names differently. That fix was approved but
not yet live.

Since then the chain kept paying out, one fault behind another, each one exposed by fixing the last:

- **The spelling fix went live and held.** Function bodies appear in evidence packs again.
- **The pack was giving impossible advice** — telling the model to re-request a whole file that its
  own arithmetic knew could never fit. Fixed, reviewed, live; wasted iterations stopped.
- **The pack was offering ambiguous names** — a same-file neighbour listed by bare name could fetch
  the *wrong function's* text, labelled as the right one, in a section the model treats as ground
  truth. Fixed, approved first round, rolled on the 13th, verified live with controls. The one
  caveat: the true two-functions-one-name collision has not yet been witnessed on a live run —
  no investigation has touched such a file since — but the tests cover it.
- **And the original question got its answer.** With the tooling repaired, the diagnosis of the
  hero-and-logo bug finally ran to a conclusion instead of abstaining: when a workflow pauses to
  wait for a sub-task, the pause keeps only three bookkeeping fields and discards the rest of the
  working state — including the image address the readers needed. Confirmed on two live workflows
  caught mid-pause. The fix sits on a seam every waiting workflow shares, which is why it is
  parked as a decision for you (RFC_012) rather than a patch from us.

## What we've done

Today's work was the last recorded trap in that chain. When the evidence pack shows a file too big
to display whole, it lists that file's functions so the model can ask for the ones it needs by
name — but the list was capped at roughly ten, and the instruction said "name the functions
individually" while withholding the other eighty names. The model cannot name what it has never
been shown, and the hidden names are precisely the ones the search failed to find. We have a
recorded case of this exact line starving a real investigation: the three functions it needed sat
behind a "+79 more" marker.

The fix makes the advice honest: when the pack says "name them individually", it now lists every
name it was withholding — compactly, so the complete list for our biggest file costs under three
thousand characters against a sixty-thousand-character budget. If even the compact list will not
fit, it says how many are missing and gives the exact query that fetches the rest, rather than
trailing off. Where the file *can* be shown whole, nothing changes at all. One subtlety mattered:
the new list had to be exempted from the section's overall size guard, because counting it would
have caused the guard to throw away the entire section on exactly the case the fix exists for.

It is tested — including a deliberate sabotage run proving the tests catch the old behaviour —
committed, and submitted to the review council, whose verdict is running as this is written. Two
small missteps of our own are on the record in the notes: a test assertion drafted against the
wrong rendering format (caught before it ran), and a submission refused once for using a field
value the gate does not accept.

## Where we are now

The tooling chain is repaired end to end: four faults found, fixed and reviewed, three of them
live in production, the fourth (today's) waiting on the next release. Ten evidence packs have been
assembled since the earlier fixes rolled: not one ambiguous name, not one wasted whole-file
iteration. The hero-and-logo bug has a confirmed cause and stays open only because its proper fix
is a design decision reserved to you. Nothing in this lane is blocked on further investigation.

## Where we're going

Three things, none of them digging:

1. **After the next release rolls**, verify today's fix on a live evidence pack — and only against
   a pack that actually scoped an oversized file, because a clean result on an easy file proves
   nothing. The recipe with that control is written in the bug file (273).
2. **Your two decisions**, unchanged but now better grounded: RFC_012 — the pause that discards
   working state, which is the hero-and-logo fix and a family of silent losses besides; and
   RFC_027 — whether the symbol-naming machinery gets one owning implementation after four bugs
   in that corner.
3. **One cosmetic leftover** (a lookup that misses package-level values and wastes a cheap call),
   queued behind everything above.
