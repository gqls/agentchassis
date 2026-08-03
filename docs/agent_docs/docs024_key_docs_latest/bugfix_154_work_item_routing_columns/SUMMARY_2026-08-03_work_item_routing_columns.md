# SUMMARY — 2026-08-03 — routing columns, fair dispatch, and the content guard

First summary for this workstream. Written the night the content-protection
phase closed: guard live, proven by a deliberate trap, and approved by the
review panel twice over.

## What we're trying to do

Two things, which turned out to be one lane. First: make the fleet's work
queue trustworthy — jobs routed with the information their handlers actually
need, claimed in a fair order, with no site able to silently starve the
others. Second, discovered along the way: stop the platform quietly
destroying page content while every status light shows green. The thread that
ties them together is that in both cases the system *reported success* — the
failures were invisible precisely because nothing looked wrong.

## Where we've come from

This began as one narrow bug (154): work items carried a routing value in a
database column that the orchestration never read, so a whole class of jobs
failed on arrival. Fixing and witnessing that led straight into a second
defect (176): the dispatcher's chooser and its loader disagreed about what
was claimable, so the queue had "quiet spells" that were actually
starvation — one stuck job could stall every site behind it. Both are fixed
and closed; the fleet now claims work in strict oldest-first order across
sites, verified live.

Then the routine follow-through caught something worse. A job whose task was
"add one link to a page" completed successfully — and a second look showed it
had regenerated the entire section and thrown away 57% of the prose: an
engineer's reference cut from seven paragraphs to four, on a live customer
page. Nothing detected it. The page-level safety check couldn't see it (the
loss was diluted across healthy sections), and the job reported complete. A
fleet-wide sweep found the same shape had already cost another site 70% of
its FAQ and a third a 32KB interactive tool. That became bug 178. The owner's
direction: restore every damaged page, and stop the class.

## What we've done

The restores went in first, every byte sourced from the history table's own
pre-overwrite snapshots — including one merge that kept both truths (the old
reference prose *and* the new link the job was legitimately asked to add).
Two pages were deliberately *not* restored after verification showed their
shrinks were genuine improvements, and one (relojistas' glossary) is held
back waiting on a decision because a blind restore could not be done safely.

Then the prevention: a new per-section guard on the save path. If a rewrite
would cut a substantial prose section below half its current text, the whole
save is refused, nothing is written, and a "needs a human" item appears in
the queue. The review panel's first verdict was REVISE, with one serious
worry: might the refusal itself be silently swallowed by the layer above, so
the protection existed but nobody ever heard it fire?

We answered that with evidence rather than argument. On our own
demonstration site, with the pass criteria written down *before* the outcome,
we set a trap the guard could not tell from the real thing. It refused —
twice, since the system retries — reported the failure honestly, raised the
human-review item within a fraction of a second, and changed not one byte of
the page. Everything was then restored and checked character-for-character.
Alongside that we measured what the reviewers asked to see measured (six
pipelines route through this save path; the guard's default covers all six)
and fixed a genuine catch of theirs (human-locked sections are now excluded
from the comparison, since the save cannot overwrite them anyway).
Resubmitted: **approved**.

The trap also caught something nine reviewer seats had not: the refusal's
one-line queue summary was borrowed from a sibling check — it described the
wrong problem and pointed operators at the wrong setting. That got the same
cure the panel had approved days earlier for an identical defect one layer
up: each refusal path must now supply its own sentence, and the compiler
refuses a new caller that doesn't. Submitted, **approved the same night**,
and its one substantive advisory (the "couldn't measure" path deserved its
own sentence too) was implemented within the hour rather than filed.

## Where we are now

Dispatch is fair and healthy. The damaged pages are restored, with the named
exceptions above. The guard is in the running binary on both replicas,
proven live, and council-approved. The wording corrections are committed and
approved but inert until the next image roll — one routine check is owed
after that roll (the grep lines are in the handoff). One structural note is
now tracked in the bug file: this save path carries *three* separate
content-loss checks, each with its own threshold and blind spot; the panel's
position, which we've made a standing rule, is that anyone reaching for a
*fourth* must instead design the unified detector.

## Where we're going

The guard stops the bleeding; it doesn't explain the wound. The open half of
bug 178 is the root cause — *why* does a job asked to insert one link
regenerate the whole section? That needs a diagnosis run, and its fix
candidates (edit-don't-regenerate, or emit only the delta) are unstarted.
The sibling save paths that don't route through this chokepoint remain
unguarded and are named in the bug file. Beyond 178, this lane's watch list
is unchanged: the spurious tool-content items (177), the spawn hang (169
part A), and the scheduler's selector asymmetry. The relojistas glossary
slot still needs an owner decision before anything restores it.
