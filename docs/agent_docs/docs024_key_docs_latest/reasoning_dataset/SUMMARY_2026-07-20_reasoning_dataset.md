# SUMMARY — reasoning dataset, as at 2026-07-20

*Milestone read-out, written to be read aloud. Current state only — the
chronology is in `README_where_we_are.md` (plain prose) and
`NOTES_reasoning_dataset.md` (technical). Supersedes the 07-18 and 07-19
summaries. All figures re-queried against the live database today.*

---

## What we're trying to do

Turn the platform's own decision-making into a dataset — not the prompts and
answers we already store, but the *reasoning*: the theory, the evidence for it,
the decision, and whether that decision turned out to be right.

The eventual goal is training a reasoning model. The near-term goal is a corpus
good enough to **evaluate** one, because what we have is high quality and there
is not much of it.

## Where we've come from

The brief assumed we needed to start capturing reasoning. We didn't — the fix
loop was built to emit cited, structured, outcome-labelled verdicts so it could
be audited, and that same design makes it usable as data. The job was always
curation, not instrumentation.

Two things then changed the shape of the project.

**The first was looking beyond the fix loop.** That loop is a few hundred rows.
The platform as a whole has now logged **41,378 LLM calls** and holds **5,514
work items** with real terminal outcomes. Reasoning is being produced roughly
100× faster than we were planning to collect it. What is missing is not
reasoning — it is the link between a decision and what happened next.

**The second was discovering how often we were wrong.** This workstream has made
seven wrong calls, every one from the same habit: concluding a *mechanism* from
*data* or *documents* without opening the code underneath, and being confident
while doing it. All seven were caught before anything shipped. Three were caught
by checks this repo already mandates and we had skipped; one by the review
council; one by a deliberate invariant we built into our own tool. That pattern
is now the most useful thing this thread has learned, and it is written down
where the next person will hit it.

## What we've done

**Built the extractor, and verified it rather than trusting it.** It produces
**820 records across 112 trajectories** — 689 usable, 131 flagged with the reason
they are not. Verification caught two real defects in our own first attempt, both
of which would have shipped silently:

- We were pairing each verdict against the wrong record of what the loop did with
  it. The tell was an output the system *cannot* produce — a verdict the safety
  logic had supposedly *upgraded*, when that logic only ever downgrades. The tool
  now refuses to make the claim when it cannot make it safely, and says so.
- Documentation we deliberately keep out of the loop's sight was leaking into the
  corpus — not from the loop, but from review submissions that legitimately
  discussed those same files. Those rows are now flagged, so an evaluation
  consumer drops them and a training consumer can choose.

**Found and filed three real platform defects**, none of which we went looking
for:

- **A verifier that reads deleted content as a successful fix.** The mechanism
  that double-checks whether a defect is really gone reports success when the
  thing it is checking has *vanished* — and a vanished component is exactly what
  happens when a rebuild silently deletes content. So a content-loss incident is
  currently recorded as a verified fix, by the mechanism built to stop us
  trusting self-reported successes. (`bugs_open/032`)
- **That same mechanism is switched on for one item type out of about fifty** —
  which is why, of 4,644 completed work items, exactly **9** have ever been
  independently verified. Filed as an instance under an existing bug rather than
  as a rival account of the same pattern.
- **A human-review queue nobody can work.** 294 items, the oldest from March,
  arriving faster than ever. The routes that would let a person action one have
  never been called, and a fourth was written, finished, and never wired up.
  (`bugs_open/033`)

**Handed both of our proposed changes to the threads that own that code**, with
three rounds of review objections already answered so nobody re-spends them.

## Where we are now

**The corpus is an evaluation set, not a training set.** 93 diagnosis reasoning
steps across 33 runs, split across two model generations. That is a volume fact,
not a limitation of the tooling, and it will not change without a deliberate
decision to generate data.

**The extractor works and is proven** — its output byte-matches the database for
the records we hand-checked, every excluded row carries its reason, and it fails
loudly rather than quietly when it cannot align two sources.

**Both changes are with their owners.** One is two small edits from likely
approval; the other needs implementation work we are not permitted to do.

**Three bugs are open and none is ours to fix.** The verifier defect is the one
that matters — it masks content loss in production whether or not anyone ever
builds a dataset.

## Where we're going

1. **Finish the gold labels.** Three of about ten are curated. Deliberately by
   hand: in the source notes, several uses of the word "FAILED" mean the API was
   overloaded rather than that the reasoning was wrong, and an automated parser
   would file those as negatives — the one error a benchmark cannot absorb.
2. **The quality report, which is the real go/no-go.** How many usable steps
   survive per model generation. Early read says usable for evaluation, well
   short of training scale.
3. **Harvest the sleeper source.** One agent — the news-feed triage — already
   emits well-structured judgements *and* batches many per call, so a few hundred
   calls carry thousands of individual decisions. It needs a query, not a change.
4. **Then a decision that is the owner's.** The corpus is too small to train on.
   The question is whether to run the loop deliberately at volume to generate
   data. The cheapest version is replaying the bugs we have already solved, where
   the correct answer is written down and grading is nearly free.

**Two open questions only the owner can answer:** whether that human-review queue
is meant to be worked or emptied, and whether generating training data at volume
is worth the spend.

**The honest ceiling, restated because it should not get lost:** this platform
makes hundreds of decisions a day, not millions. Six months of all of the above
is *tens of thousands* of outcome-labelled reasoning steps — fine-tuning and
evaluation scale, not pretraining scale. Still worth doing: a few thousand
outcome-*verified* traces with real negatives is a genuinely scarce asset. But
the goal is a high-quality specialist corpus, not a big one.
