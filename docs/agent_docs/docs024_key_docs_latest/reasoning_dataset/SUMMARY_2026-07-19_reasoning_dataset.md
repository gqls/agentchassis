# SUMMARY — reasoning dataset, as at 2026-07-19

*Milestone read-out. Current state, no chronology — the history is in
`README_where_we_are.md` (plain prose) and `NOTES_reasoning_dataset.md`
(technical). Supersedes `SUMMARY_reasoning_dataset_2026-07-18.md`.*

---

## What we're trying to do

Turn the platform's own decision-making into a dataset — not the prompts and
answers we already store, but the reasoning: the theory, the evidence for it, the
decision, and whether that decision turned out to be right.

The eventual goal is training a reasoning model. The near-term goal is a corpus
good enough to *evaluate* one, because the reasoning we have today is high
quality but there is not much of it.

## Where we've come from

The brief assumed we needed to start capturing reasoning. We don't — the fix loop
was built to emit cited, structured, outcome-labelled verdicts so it could be
audited, and that same design makes it training data. So the job was always
curation, not instrumentation.

What changed the shape of the project was looking beyond the fix loop. It is 445
rows. The platform as a whole has logged **40,785 LLM calls** and holds **5,292
work items with real terminal outcomes**. The reasoning is being produced roughly
100× faster than we were planning to collect it. What is missing is not reasoning
— it is the link between a decision and what happened next.

## What we've done

- **Sized the corpus honestly against the live database** rather than carrying
  the brief's figures forward. The fix loop yields ~79 reasoning steps across 13
  trajectories, split across two model generations.
- **Written the standing documents** — plan, runbook, technical log, this
  read-out, and the running plain-prose history.
- **Found and quantified a defect** that invalidates one lane of the corpus: every
  revise step in the fix loop's history reasoned against blank reviewer
  objections. Already filed by another thread as `bugs_open/016`; we added the
  blast radius (100% of the lane, not a sample) and a verification trap worth
  knowing — a config fix must be graded on when the *run* started, not when the
  step logged, or a long run straddling the fix reads as a failure.
- **Mapped the six capture gaps** that stand between us and volume, and put the
  two most valuable through the council gate as change proposals.
- **Been wrong four times and corrected it in place** — see below.

## Where we are now

**The corpus is an evaluation set, not a training set.** That is a volume fact,
not a limitation of the extraction design, and it will not change without a
deliberate decision to generate data.

**Two changes are with the council, both in their second round.** Neither was
rejected; both drew specific objections that were correct on checking.

- **A — record where a work item came from.** One column linking an item to the
  run whose reasoning raised it. Makes ~15,000 auditor judgements a month
  checkable against their outcomes, and would give the first per-agent accuracy
  signal the platform has had.
- **B — actually verify that fixes worked.** The mechanism to re-check a defect
  before marking it done already exists and works; it has been enabled for
  **one** item type out of about fifty. That is why 9 items in the platform's
  entire history have been independently verified. In review, a council seat
  found a real bug in that existing mechanism — it treats "the component has
  vanished" as "fixed", when a vanished component is equally the signature of a
  rebuild silently deleting content. The revision fixes that at source.

**The extraction job itself has not been built.** It is not blocked by either
submission.

**Four wrong calls, all caught before anything shipped**, all from the same
habit: reasoning from data or documents without reading the code underneath.
Three were caught by checks this repo already mandates and we had skipped; the
fourth was caught by the council. They are recorded rather than tidied away
because the pattern is the useful part.

## Where we're going

1. **Finish the council rounds** and hand both approved plans to the threads that
   own that code. Implementation is not ours.
2. **Build the extractor** — a read-only job, running outside the cluster,
   emitting one record per reasoning step with provenance and three families of
   labels. Roughly two days, ending in a corpus that can be inspected.
3. **Harvest the sleeper source.** One agent — the news-feed triage — already
   emits well-structured judgements *and* batches many per call, so its 423 calls
   carry thousands of individual decisions. It needs a query, not a change.
4. **Then a decision that is the owner's.** The corpus is too small to train on.
   The question is whether to run the loop deliberately at volume to generate
   data, which is a much larger undertaking than any of the above. The cheapest
   version is replaying the known-answer bugs in `/bugs_open/` and
   `/bugs_closed/`, where the correct diagnosis is already written down.

**The honest ceiling, restated because it should not get lost:** this platform
makes hundreds of decisions a day, not millions. Six months of everything above
is *tens of thousands* of outcome-labelled reasoning steps. That is fine-tuning
and evaluation scale, not pretraining scale. It is still worth doing — a few
thousand outcome-*verified* traces with real negatives is a genuinely scarce
asset — but the goal is a high-quality specialist corpus, not a big one.
