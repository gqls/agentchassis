# HANDOFF — capture the fix-loop's REASONING as a training dataset for a reasoning model

*Filed 2026-07-18 from the diagnosis-fixloop thread, at the owner's request. A
standalone workstream: this is data engineering + ML dataset curation, NOT
diagnosis-loop development — start it as its own chat. It COORDINATES with the
fixloop thread (that thread owns the artifact schema this reads) but does not
change the loop. Read top to bottom; it is self-sufficient.*

---

## 1. The reframe that changes the whole job: the reasoning is ALREADY persisted

The owner's framing was "we save LLM prompts + responses (for a base model); now
also save the REASONING STEPS (for a reasoning model)." The important correction
up front: **on this platform the reasoning is already written down** — and in a
form that is *better* training signal than raw chain-of-thought. So this
workstream is **curation + labelling + extraction**, not new capture. Most of the
value is already sitting in the live DB; the work is to assemble it into
(state → reasoning → decision → outcome) tuples.

### Where the reasoning already lives (verified 2026-07-18)
- **The diagnosis verdict** (`llm_call_log.response_text` for `step_name='verdict'`,
  and `orchestration_states.collected_data->'verdict'->'result'`) is a JSON object
  that IS the reasoning: `outcome` (CONFIRMED/REFUTED/UNVERIFIABLE) + **`citations`**
  (evidence-grounded justification, each with tier/where/quote) + `revised_hypothesis`
  + `next_scope` + `symptom_check` + `data_requests`. This is "here is my theory,
  here is the evidence, here is why it is confirmed/refuted, here is where to look
  next" — explicit, grounded reasoning.
- **The multi-iteration TRAJECTORY** (`collected_data->'route'->'diagnose_state'->'trail'`,
  and one `diagnosis_artifacts` bundle per iteration): the scope narrowing from
  symptom → cause across iterations. A reasoning *trajectory*, not a single step —
  the model following evidence to a cause named nothing like the symptom.
- **The council reasoning** (`collected_data->'review_*'->'result'->'notes'`, and
  `diagnosis_artifacts` kind=`council_report`): each reviewer's argument about the
  plan, its `checks`/`code_checks` (what facts it needed), and the `repropose`
  reasoning (how the plan changed in response). Objection → resolution pairs.
- **Volume today** (small but growing, and much is benchmark-graded): 38 bundles,
  39 fix_plans, 34 council_reports, 5 escalations; 296 verdict/review rows in
  `llm_call_log`; 26 orchestrations carrying a verdict.

### The raw-chain-of-thought caveat (read before promising "reasoning tokens")
The Anthropic API does **not** return the model's raw internal chain of thought —
on Sonnet/Opus it is summarised or omitted (`thinking.display`), never the raw
tokens. So you **cannot** harvest the model's literal private reasoning from the
responses. Two consequences:
1. If "reasoning steps" means raw CoT tokens — that is not available from the API
   for these runs, and retrofitting `thinking:{display:"summarized"}` only yields
   summaries, not raw traces.
2. **But you do not want the raw CoT anyway.** The loop is *designed* to emit its
   reasoning as structured, cited, outcome-labelled output (cite-or-abstain
   verdicts; evidence-following scope). That is the exact format you want a
   reasoning model to *produce*, and it comes with ground-truth labels raw CoT
   lacks. Treat the structured verdict trail as the reasoning signal; treat the
   summarised thinking blocks (if you enable them) as an optional auxiliary field.

## 2. Why this is premium training data: it is OUTCOME-LABELLED

Reasoning datasets are only as good as their labels. This corpus has unusually
strong ones, already attached or cheaply joinable:
- **Verdict outcome** (CONFIRMED/REFUTED/UNVERIFIABLE) — the loop's own
  self-label per step.
- **Benchmark grades** — the fixloop workstream pre-registers a rubric BEFORE a
  run and grades the verdict against it (PASS/PARTIAL/FAIL). These are gold
  human-audited labels on a subset. (`fixloop_eg_dartsonline/RUBRIC_*.md`,
  `NOTES_running_fixloop(10).md`.)
- **Terminal outcome** — did a CONFIRMED diagnosis's fix get council-APPROVED?
  Did the PR get merged by a human? A confirmed→approved→merged chain is a
  positive trajectory; a REFUTED-that-was-correct is a positive *abstention*; a
  CONFIRMED-that-a-human-rejected is a negative. These join by `correlation_id` /
  `fix_correlation_id` across `diagnosis_artifacts`, `orchestration_states`, and
  (for merges) the PR record.
- **The honesty terminals are signal, not noise.** This session produced several
  REFUTED/UNVERIFIABLE/escalate outcomes that were *correct* refusals of flawed
  inputs. A reasoning model should learn to abstain and to refuse a partial
  fix — these are the hard positive examples, and they are labelled as such in
  the notes.

## 3. Proposed dataset shape

One training example = one reasoning STEP, plus a trajectory id so multi-step
chains can be reconstructed:

```
{
  "trajectory_id": "<correlation_id>",
  "step_index":    <iteration>,
  "task":          "diagnosis" | "council_review" | "repropose",
  "input_state":   <the bundle / plan+reviews the model saw this step>,
  "reasoning":     <the verdict JSON / review notes — citations, scope, symptom_check>,
  "decision":      "CONFIRMED|REFUTED|UNVERIFIABLE|approve|object|veto|...",
  "labels": {
    "self_outcome":   "...",              // the loop's own verdict
    "benchmark_grade":"PASS|PARTIAL|FAIL|null",   // gold, subset
    "terminal":       "merged|approved|escalated|rejected|null"  // trajectory end
  },
  "provenance": { "agent_type", "model", "commit_sha", "created_at" }
}
```

The trajectory (all steps sharing a `trajectory_id`, ordered by `step_index`) is
the reasoning *chain* — the thing a reasoning model learns to walk.

## 4. Extraction approach (a read-only ETL job, not a platform change)
- Join `diagnosis_artifacts` (bundle per iteration = input_state; council_report;
  escalation) + `orchestration_states.collected_data` (verdict, route trail,
  review_*) + `llm_call_log` (prompt_rendered, response_text, model, tokens) on
  `correlation_id`, ordered by iteration.
- Attach labels: verdict outcome (in the artifact); benchmark grade (parse the
  RUBRIC/NOTES docs, or add a small `benchmark_grades` table the fixloop thread
  populates); terminal outcome (approved/merged, from the fix correlation).
- Emit JSONL. Read-only; it touches no live workflow. Runs as a script or a small
  Go job against the read replica.

## 5. Landmines (from the fixloop thread's hard-won experience)
- **Blind the corpus.** The fixloop DOCS (`fixloop_eg_dartsonline/`) are
  deliberately EXCLUDED from the loop's own input so benchmarks stay honest —
  keep them out of any training set that will be evaluated on the benchmark, or
  you leak the answers.
- **Premise-shift / stale runs.** Several 2026-07-16..18 runs diagnosed a premise
  that had already changed (another thread shipped a fix mid-run) → a REFUTED that
  is an artefact of timing, not reasoning quality. The NOTES flag these; filter or
  label them, don't treat every REFUTED as a clean negative.
- **Config churn provenance.** The proposer/reviewer MODEL changed mid-week
  (sonnet-4-6 → sonnet-5, 2026-07-18) and `max_tokens` was raised; roster grew
  2→13 seats. Record `model` + `commit_sha` per example so you can slice by
  model — mixing model generations in one dataset without a provenance field will
  confound training.
- **Truncated reasoning is poison.** A verdict/review whose `output_tokens ==
  max_tokens` was CUT (CLAUDE.md's enshrined rule) — its reasoning is a fragment.
  Filter `output_tokens >= max_tokens` rows OUT of the training set; they look
  like reasoning and are not.
- **The `<no value>` trap (bugs_open/016).** A council reviser whose objection
  injection rendered `<no value>` reasoned without seeing the objections. Verify
  the input_state a step actually SAW was complete before treating its reasoning
  as a valid (state→reasoning) pair.

## 6. Coordination + recommendation
- **Separate thread: YES.** This is ETL + dataset curation + (eventually) a
  training pipeline — a different skill and cadence from loop development. Starting
  it here would derail the loop work and vice versa.
- **Coordinate with the fixloop thread on ONE thing:** the artifact schema
  (`diagnosis_artifacts` kinds, `collected_data` shape) is the source of truth for
  this dataset; if that thread changes the verdict/council JSON shape, this ETL
  breaks. Agree a stable read contract (or version the artifacts).
- **First deliverable for the new thread:** the read-only ETL emitting JSONL for
  the ~26 graded trajectories that exist today, with the three label families
  attached — enough to inspect quality and decide whether the signal is worth a
  training run before investing in scale.

## 7. Pointers
- Verdict prompt + wire format: `fix-proposer`/`diagnose-agent` `default_config`
  (verdict step), `pkg/diagnose/verdict_wire.go`, `pkg/diagnose/step.go`
  (the coercion/guard logic that shapes what a valid verdict is).
- Reasoning already stored: `llm_call_log` (prompt_rendered/response_text),
  `diagnosis_artifacts` (kind ∈ bundle|fix_plan|council_report|escalation by
  correlation_id/iteration), `orchestration_states.collected_data`.
- Benchmark labels: `fixloop_eg_dartsonline/RUBRIC_*.md`,
  `SUMMARY_the_immune_system_2026-07-18.md`, `NOTES_running_fixloop(10).md`.
