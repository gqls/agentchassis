# SUMMARY — reasoning dataset: what it is, where we are, where we're going

*2026-07-18. Written to be read aloud. Figures are live-DB, not estimates.*

---

## What we're doing

Capturing the fix-loop's **reasoning** as a dataset — not just the prompts and
answers we already save, but the chain of *how it got there*: theory, evidence,
verdict, and where to look next.

The useful correction, which the handoff got right and which changes the size of
the job: **the reasoning is already written down.** The loop was built to emit
cited, structured, outcome-labelled verdicts, because that's what made it
auditable. That same design makes it training data. So this is curation and
extraction, not new instrumentation — we are assembling what the live database
already holds, not building capture.

One thing to be clear about, because it sounds like it should be true: we cannot
harvest the model's raw private chain-of-thought. The API doesn't return it. But
we don't want it — the structured verdict trail is better signal, because it comes
with ground-truth labels that raw thinking-tokens lack.

## What we've done

Planned it, and grounded every number against the live database rather than
carrying the brief's figures forward. That was worth doing: the corpus is larger
than the handoff said on some axes (445 reasoning rows, not 296) and the brief
had three premises that don't survive contact with the data.

We also found — independently, coming at it from the data side — a defect that two
other threads had found the same day from the debugging side.

## What we found

**The corpus is an evaluation set, not a training set.** 79 diagnosis reasoning
steps across 13 trajectories, split across two model generations (70 verdicts on
sonnet-4-6, 19 on sonnet-5). That will not train a reasoning model, and no clever
extraction changes it. Said plainly now rather than discovered after building a
pipeline.

**One whole lane of the data is invalid.** Every `repropose` call in the corpus —
19 of 19 — was made while the reviewer objections it was supposed to be addressing
rendered as blank. The revise loop was revising against nothing. Those are exactly
the "objection → resolution" pairs the brief named as premium signal, so that lane
is quarantined.

**That defect was already found and already fixed** — filed as `bugs_open/016` by
the experience-loop thread, fixed in the live config by the council-gate thread at
13:15Z, about forty-five minutes before we looked. We had drafted a duplicate bug
report; grepping the bugs_open index first, which the repo's own rules require, is
what stopped it going out. What we added instead was the blast radius: it was 100%
of the lane, not a sample, and the data was demonstrably present while the prompt
showed blanks.

**But it isn't finished.** Two things remain, both now recorded in 016:

- The fix is **unexercised**. No revise round has run since it landed, so it is
  correct-looking but unproven. (We briefly believed it had failed — two calls
  timestamped after the fix still showed blanks. They belonged to a run that
  *started* before it. The log timestamp is the step's, not the run's. Worth
  knowing, because it flips a verdict either way.)
- The reviser is still **half-blind by a different mechanism**. Thirteen council
  seats are seeded; the revise prompt references six. Seven seats' objections —
  54% of the council — never reach the reviser at all. That gap arrived through
  seat growth rather than a code defect, which means it will recur the next time a
  seat is added unless the shape changes.

**The benchmark labels can't be automated.** One rubric file, about ten graded
runs, four incompatible prose formats, and several uses of "FAILED" that mean an
API error rather than a grade. We'll hand-curate ten labels rather than write a
parser that would confidently mislabel them.

**One unexpected asset.** The loop's own guard logic already writes down, in plain
language, *why* a piece of reasoning was invalid — zero citations, confirmed
without both evidence families, an unexplained observation. Pairing the raw model
verdict against the guard-corrected one gives labelled examples of reasoning that
looks right and isn't. That's rare, hard to construct by hand, and it comes free.

## Where we're going

Build the extraction — a read-only job, running outside the cluster, emitting one
record per reasoning step with provenance and three families of labels. About two
days of work across three phases, each ending in something inspectable.

Aim it at **evaluation first**: 13 trajectories with gold grades makes a credible
benchmark for judging whether a candidate model diagnoses as well as the current
loop. That is achievable with the data we have today.

Then a real decision point, which is yours: the corpus is too small to train on,
so the question is whether to **run the loop deliberately at volume** to generate
training data. That is a considerably larger undertaking than this ETL, and it
should be chosen on purpose rather than drifted into.

The immediate next step is coordination, not code — telling the fixloop thread
about the seven invisible seats, and agreeing that the verdict and council data
shapes won't change under us mid-build.

---

*Detail: `PLAN_2026-07-18_reasoning_dataset_extraction.md` (design, corrections,
phasing) · `RUNBOOK_reasoning_dataset.md` (commands) ·
`NOTES_reasoning_dataset.md` (running record) · `bugs_open/016` (the defect).*
