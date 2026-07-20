# SUMMARY — reasoning dataset, as at 2026-07-20 (end of day)

*Milestone read-out, written to be read aloud. Current state only — the
chronology is in `README_where_we_are.md` (plain prose) and
`NOTES_reasoning_dataset.md` (technical). Rewritten this evening after the
quality report landed; the morning version described the report as pending, and
it is now done with a verdict. Supersedes the 07-18 and 07-19 summaries.*

---

## What we're trying to do

Turn the platform's own decision-making into a dataset — not the prompts and
answers we already store, but the *reasoning*: the theory, the evidence for it,
the decision, and whether that decision turned out to be right.

The eventual goal is training a reasoning model. The near-term goal was a corpus
good enough to **evaluate** one. That question now has an answer.

## Where we've come from

The brief assumed we needed to start capturing reasoning. We didn't — the fix
loop was built to emit cited, structured, outcome-labelled verdicts so it could
be audited, and that design makes it usable as data. The job was always curation.

Three things then reshaped the project.

**Looking beyond the fix loop.** That loop is a few hundred rows; the platform
has logged over 41,000 LLM calls and holds 5,500 work items with real outcomes.
Reasoning is produced roughly 100× faster than we planned to collect it. What is
missing is the link between a decision and what happened next.

**Discovering how often we were wrong.** Eight wrong calls now, every one from
the same habit: concluding a *mechanism* from *data* or *documents* without
opening the code underneath, while feeling confident. All eight were caught
before anything shipped — three by checks this repo already mandates that we had
skipped, one by the review council, one by an invariant we deliberately built
into our own tool, and one by measuring a claim we had made and never checked.

**Building the thing, which settled arguments that talking could not.** Several
confident claims in our own plan did not survive contact with the extractor.

## What we've done

**Built and proved the extractor.** It produces 820 records across 112
trajectories — 689 usable, 131 flagged with the reason they are not. Verification
caught two real defects in our own first attempt, both of which would have
shipped silently. One was exposed by an output the system *cannot* produce: a
verdict the safety logic had supposedly *upgraded*, when that logic only ever
downgrades. The tool now refuses to make that claim when it cannot make it
safely, and says so in the data.

**Curated the gold labels by hand — twelve runs.** Deliberately not parsed: in
the source notes the word "FAILED" means an overloaded API in one place and a
lost spawn in another, and an automated pass would have filed both as reasoning
failures. That is the one error a benchmark cannot absorb. Two of the twelve
produce no records at all, correctly, because those runs never started — and the
corpus agreeing with the notes on that point is a small proof the pipeline works.

**Delivered the go/no-go report**, which is the gate the plan set in advance.

**Found and filed three platform defects we were not looking for** — a verifier
that reads deleted content as a successful fix; that same verifier being switched
on for one item type out of about fifty; and a human-review queue of 294 items
that nobody has any working way to action.

**Handed both proposed changes to the threads that own that code**, with three
rounds of review objections already answered so nobody re-spends them.

## Where we are now

**The verdict is in, and it is a split decision.**

| question | answer |
|---|---|
| Train a reasoning model on this? | **No.** 102 usable diagnosis steps across two model generations. |
| Use it as an evaluation set? | **Yes, modestly.** 24 gold-graded steps over 6 trajectories. |
| Where is the unexploited volume? | **The council lane** — 574 usable steps, not one label. |

**Why the small eval set is still worth having.** Two of its three passes are
*abstentions* — runs graded as correct for refusing to answer, one because a
confident half-answer would have been handed to a fixer. The single failure is a
*confident* wrong answer, sure of itself after five iterations. And one run
carries three different judgements of the same step: what the model said, what
the safety logic did to it, and what a human graded it. A benchmark where
declining to answer is the gold behaviour is genuinely scarce.

**One of our own claims did not survive measurement.** The plan called a
particular signal "the most distinctive thing in this corpus". Measured, it is
seven examples of a single failure mode — an illustration, not a signal. Written
before the tool existed, never checked, now corrected in place rather than
quietly dropped.

**One whole lane is a total loss.** Every revise step in the corpus — all
twenty-three — reasoned against blank objections, because they predate the fix
for that defect. Refillable only by running the loop again.

## Where we're going

1. **Mine the council lane.** It is 5.6× the diagnosis lane, entirely unlabelled,
   and has several reviewers judging the *same* plan — which is a preference set,
   the format this kind of training actually wants. Two ways to label it need no
   platform change: whether an approved plan was actually committed, and where
   reviewers disagreed with each other. Biggest return, no new spend.
2. **Harvest the news-feed triage.** A few hundred calls carrying thousands of
   batched, already-structured judgements. One query.
3. **Then a decision that is the owner's:** whether to run the loop deliberately
   at volume. The cheapest form is replaying the roughly thirty bugs we have
   already solved, where the correct diagnosis is written down and grading is
   nearly free — and which would also refill the dead revise lane.
4. **Do not extract again for its own sake.** The pipeline is proven; more passes
   over the same thirteen diagnosis runs add nothing.

**Two open questions only the owner can answer:** whether that 294-item
human-review queue is meant to be worked or emptied, and whether generating
training data at volume is worth the spend.

**The honest ceiling, restated because it should not get lost:** this platform
makes hundreds of decisions a day, not millions. Six months of all of the above
is *tens of thousands* of outcome-labelled reasoning steps — fine-tuning and
evaluation scale, not pretraining scale. Still worth doing: a few thousand
outcome-*verified* traces with real negatives is a scarce asset. But the goal is
a high-quality specialist corpus, not a big one.

---

*Detail: `NOTES_corpus_quality.md` (the report and its threats to validity) ·
`LABELS_benchmark.json` (the twelve curated runs, each quoting its source) ·
`PLAN_2026-07-18_…` and `PLAN_capture_gaps_and_volume.md` (design and
corrections) · `RUNBOOK_reasoning_dataset.md` (commands) · `bugs_open/032`,
`033`, `021` (the defects found).*
