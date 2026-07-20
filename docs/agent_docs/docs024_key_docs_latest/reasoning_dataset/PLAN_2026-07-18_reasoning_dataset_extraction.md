# PLAN — reasoning-dataset extraction (read-only ETL)

*Planning response to `HANDOFF_2026-07-18_reasoning_training_dataset.md`, written
2026-07-18 after grounding every volume claim against the live DB. The handoff's
core reframe is correct — the reasoning is already persisted, so this is curation
not capture. Three of its factual premises need correcting, and one live bug
found during planning removes an entire lane from the corpus.*

---

## Context

The owner wants the fix-loop's reasoning captured as training data for a
reasoning model. The handoff established that the reasoning already exists in
structured, cited, outcome-labelled form in `diagnosis_artifacts`,
`orchestration_states.collected_data`, and `llm_call_log` — so the job is a
read-only ETL that assembles (state → reasoning → decision → outcome) tuples,
not new instrumentation.

Planning confirmed the reframe and the join topology. It also found that the
corpus is smaller than a training run needs, that one lane of it is invalid, and
that the benchmark labels cannot be parsed automatically. Those three facts
change what this project should aim at — see **Recommendation** below.

---

## 1. Corrections to the handoff (verified against the live DB, 2026-07-18)

### 1a. The `repropose` lane is invalid, 100% of it — cause already fixed, corpus still poisoned

> **CORRECTED 2026-07-18 ~14:00Z, before any code was written.** As first drafted
> this section claimed the defect was live and unfiled, and recommended filing it.
> Both claims were wrong. It was **already filed** as
> `bugs_open/016_HANDOFF_2026-07-18_council_revise_prompts_drop_reviewer_output.md`
> — found the same day by the experience-loop thread — and the `fix-proposer` row
> was **fixed at 13:15:11Z** by the council-gate thread. Grepping the bugs_open
> index before filing (CLAUDE.md's rule) is what caught this; I had drafted a
> duplicate first. The evidence below stands and the corpus consequence is
> unchanged — what changes is that this is a *historical* corruption of the data,
> not an open bug for this thread to raise. Quantification and the pre/post-fix
> boundary are recorded in 016.

The handoff lists the `<no value>` trap (`bugs_open/016`) as landmine #5: *"verify
the input_state a step actually SAW was complete."* It is not an edge case to
check for. It is every row:

| step | rows | rendered prompt contains `<no value>` |
|---|---|---|
| `repropose` | 19 | **19 (100%)** |
| `review_debug_historian` | 13 | **13 (100%)** |
| `reframe` | 2 | **2 (100%)** |
| `propose` | 16 | 1 |
| `verdict` | 89 | 1 |

Each affected `repropose` prompt blanks **3–6 reviewer sections** — the template
emits the heading (`## Edit-quality reviewer said`) followed by `<no value>`.

The data was present. For run `53da3a30` (the BUG A run recorded as proven
end-to-end), `collected_data->'review_editquality'->'result'` is a well-formed
object with all seven keys (`notes`, `checks`, `missing`, `verdict`, `reviewer`,
`objections`, `code_checks`), 1561 chars. The template path
`{{.review_editquality.result}}` (`0NN_fix_proposer.sql:365` and every later
version) matches that path exactly. So the value exists and the reference is
right, but the template context handed to the `repropose` step does not carry it
— an input_fields wiring defect, not missing data.

**Consequence for this project:** the "objection → resolution" pairs the handoff
names as premium signal in §1 are not that. The reviser was revising a plan while
being shown empty objections; its reasoning is not grounded in what it appears to
be grounded in. All 19 `repropose` rows are **quarantined, not trained on**.

**Consequence beyond this project:** the council's revise loop was decorative for
the entire life of the corpus. The render cause is fixed but **unexercised** — no
repropose run has started since 13:15:11Z, so the fix is unproven in the wild.
And 016's second finding is still open and live: the `repropose` prompt references
**6 of 13 seeded seats**, so a revise round still cannot see 7 seats' objections.
Both belong to the fixloop thread; this thread must not fix them (see §6). The
open item to raise is the 6-of-13 coverage gap, not the render bug.

### 1b. Benchmark labels are not machine-parseable

The handoff's §4 says to *"parse the RUBRIC/NOTES docs, or add a small
`benchmark_grades` table."* Parsing will not work:

- There is exactly **one** `RUBRIC_*.md` in the whole repo
  (`RUBRIC_2026-07-16_two_config_bugs.md`), not a set.
- `NOTES_running_fixloop(10).md` is 141KB and contains only ~21 lines mentioning
  PASS/PARTIAL/FAIL, in at least four incompatible shapes (a prospective grading
  key; two markdown tables with *different* columns; inline bolded prose; a
  narrative run-ledger line carrying several runs at once).
- The value vocabulary is open: `partial`, `**PASS, CITED**`,
  `**FAIL — and actively dismissed**`, `not reached`, `abstained (neutral)`.
- Several uses of FAIL/FAILED mean *infrastructure failure*, not a grade
  (e.g. `NOTES:1965`, an API 529). A regex would mislabel these.
- `NOTES_running_fixloop(9).md` predates the discipline — zero grades.

**Take the handoff's second option.** Hand-curate a small `LABELS_benchmark.json`
keyed on the 8-char correlation prefixes used throughout the notes. There are
~10 graded outcomes; enumerating them by hand is faster and far more accurate
than a parser, and the file becomes the durable artifact the notes are not.

### 1c. Truncation is a near-non-issue; the model split is the real confound

The handoff calls truncated reasoning "poison." Correct in principle, negligible
in practice: **6 rows out of 445** trip the filter, all old-regime
(`output_tokens >= max_tokens`); zero new-regime
(`error_message LIKE 'response truncated%'`). Keep the filter — it is cheap and
the rule is enshrined — but it is not a project risk.

The confound that *is* material is the one the handoff mentions only in passing.
Verdicts split **claude-sonnet-4-6: 70 / claude-sonnet-5: 19**. Slicing by model
to avoid mixing generations leaves two slices, neither large enough to train on.
`provenance.model` is mandatory, and every downstream count must be reported
per-model.

### 1d. Volumes (live, replacing the handoff's figures)

| | handoff | live 2026-07-18 |
|---|---|---|
| bundles | 38 | 38 (13 correlations) |
| fix_plans | 39 | 43 |
| council_reports | 34 | 43 |
| escalations | 5 | 5 |
| verdict+review rows in `llm_call_log` | 296 | **445** |
| orchestrations carrying a verdict | 26 | 26 |
| **diagnosis trail steps** | — | **79** |

Trail depth per run: 5 runs × 1 step, 5 × 2, 8 × 3, 8 × 5.

Verdict outcome distribution is heavily skewed: **57 UNVERIFIABLE / 18 CONFIRMED
/ 13 REFUTED** across 89 verdict calls. Honest abstention dominates the corpus.
That is defensible behaviour by the loop but it is a lopsided label distribution.

---

## 2. Recommendation — build the pipeline, aim it at eval, not training

**79 diagnosis reasoning steps and ~350 council review steps across 13
trajectories will not train a reasoning model.** Nothing in the extraction design
changes that; it is a volume fact. Stating it plainly up front is more useful
than a pipeline that quietly produces a 79-line JSONL.

What the corpus *is* good for, today:

1. **An eval set / benchmark.** 13 trajectories, ~10 with pre-registered gold
   grades, with cite-or-abstain structure and honest-abstention positives, is a
   respectable evaluation harness for judging whether a candidate model
   diagnoses as well as the current loop. This is the highest-value near-term use
   and needs no more data than exists.
2. **A guard-trip corpus** (novel — see §3c). The coercion logic labels exactly
   *what made a piece of reasoning invalid*, in natural language, automatically,
   on every degraded verdict. That is a supervision signal for "reasoning that
   looks right and isn't" which is rare and hard to construct by hand.
3. **A ready pipeline for when volume arrives.** Build it now while the corpus is
   small enough to inspect by hand and verify end-to-end. Scaling is then a
   parameter change, not a project.

The real strategic question — worth deciding before Phase 3, and it is the
owner's call — is whether to **run the loop deliberately at volume** to generate
training data, which is a different and much larger undertaking than this ETL.

---

## 3. Design

### 3a. Architecture: psql extracts, Go transforms, runs outside the cluster

Follow the `cmd/claimscan` idiom (`cmd/claimscan/main.go:14-22`): a local CLI
that reads psql output, with the extraction query documented in the file header.
It needs no DB code, no config, no image, no makefile change.

**Do not run this as a pod.** `platform/orchestration/actions/training_data_export.go:3-8`
records the lesson directly: v3 replaced *"earlier file-writing behaviour that
landed on ephemeral chassis pods"* and had to be `kubectl cp`'d off before the
pod died. Running outside the cluster makes that problem not apply. There is also
no read replica — only `postgres-clients-0` behind pgbouncer — so a heavy scan
competes with production; streaming one `SELECT` through `kubectl exec` avoids
holding a pool connection.

New files:

```
cmd/reasoningset/main.go          # the transformer: psql JSON in → JSONL out
cmd/reasoningset/extract.sql      # the single extraction query, documented
docs/.../reasoning_dataset/LABELS_benchmark.json   # hand-curated gold grades
```

Gotcha to carry over from `098_REPORT_unreviewed_commits_v1.sh:63`: `kubectl exec -i`
consumes the enclosing loop's stdin — any loop over correlation ids needs
`< /dev/null` or the run stops dead.

### 3b. Join contract

Three ids, easily conflated:

- `correlation_id` — one agent run's identity. **`diagnosis_artifacts.correlation_id`
  deliberately collapses a diagnosis run and its fix-proposer run onto one key**
  (the proposer is fired with a fresh envelope and `input_data.fix_correlation_id`
  = the diagnosis correlation). This is the `trajectory_id`.
- `orchestration_id` — one *workflow* run. Since a correlation accumulates across
  re-runs (13 correlations, 43 council_reports), this is the disambiguator and
  becomes `run_id`.
- Council-gate submissions mint their own `SUBMISSION_CORR` with **no diagnosis
  behind them** — they land in `diagnosis_artifacts` under the same column with no
  bundle. Detect and tag these; they are a different task type, not a broken
  trajectory.

Cast warnings: `orchestration_states.correlation_id` is `uuid`,
`diagnosis_artifacts.correlation_id` is `text`, `llm_call_log.correlation_id` is
`varchar(255)`. Every cross-table join needs an explicit cast. Also
`orchestration_states.last_activity` is `timestamp WITHOUT time zone` while
`created_at` is `timestamptz` on a BST host — do not do interval arithmetic
across them.

`LogLLMCall` is fire-and-forget with a 5s timeout
(`platform/orchestration/actions/llm_call_logger.go:34`), so `llm_call_log` rows
can be silently missing. Never assume 1:1 with workflow steps; treat a missing
row as absent evidence, not as zero.

### 3c. Record shape

Refines the handoff's §3. One record = one reasoning step.

```jsonc
{
  "trajectory_id": "<correlation_id>",       // diagnosis + its proposer runs
  "run_id":        "<orchestration_id>",     // disambiguates re-runs
  "step_index":    2,                        // iteration
  "task": "diagnosis_verdict|council_review|council_decide|repropose|escalation",

  "input_state":   "...",   // bundle body / plan+reviews as the step SAW it
  "reasoning":     {...},   // verdict JSON (citations/scope/symptom_check) or review notes
  "decision":      "CONFIRMED|REFUTED|UNVERIFIABLE|approve|object|veto",

  "input_complete": true,   // false if <no value> in prompt_rendered, or bundle truncated
  "exclude_reason": null,   // "no_value_injection" | "truncated" | "premise_shift" | null

  "guard": {                                  // NEW — see below
    "raw_decision":     "CONFIRMED",
    "coerced_decision": "UNVERIFIABLE",
    "tripped":          true,
    "diagnostic":       "confirmed without both evidence families ..."
  },

  "labels": {
    "self_outcome":    "UNVERIFIABLE",
    "benchmark_grade": "PASS|PARTIAL|FAIL|null",   // hand-curated, subset
    "terminal":        "merged|approved|escalated|rejected|null"
  },
  "provenance": {"agent_type","model","model_resolved","max_tokens",
                 "orchestration_id","created_at"}
}
```

Two deliberate departures from the handoff:

**`input_complete` / `exclude_reason` instead of dropping rows.** Bad rows stay in
the file, flagged. A filtered corpus that silently omits 19 `repropose` rows reads
as "we have no repropose data"; a flagged one records *why*, which is itself the
finding. Consumers filter on `input_complete`.

> **CORRECTED 2026-07-20, after measuring it.** This section calls the guard
> block "the most distinctive thing in this corpus". Measured: **7 trips, all
> carrying the SAME diagnostic** ("confirmed on one evidence family only"), plus
> 14 records too ambiguous to assert on. Seven examples of one failure mode is an
> illustration, not a training signal. The claim was written before the extractor
> existed and was never checked. The mechanism is still worth having — it is free
> and grows if the loop runs more — but it is not a reason to invest. Full
> measurement: `NOTES_corpus_quality.md` §5.

**The `guard` block — the novel signal, not in the handoff.** The trail stores the
**coerced** verdict (`pkg/diagnose/advance.go:93-97`) while
`collected_data.verdict.result` holds the **raw** model output. Pairing them
yields, for free, a before/after label on every guard trip. And
`pkg/diagnose/step.go:67-101` *prepends a diagnostic sentence to `NeededEvidence`*
naming exactly what the verdict was missing — zero citations; CONFIRMED without
both evidence families; empty `symptom_check`; an unexplained observation; an
explained-but-uncited one. That is a natural-language label for *why this
reasoning is invalid*, generated automatically, on real model output. It is the
most distinctive thing in this corpus and it should be a first-class field.

Note when parsing the trail: `Step`, `Verdict`, `Citation` and `Scope` carry **no
json tags** (`pkg/diagnose/loop.go:79-153`), so trail entries serialise with Go
field names and **integer enums** — `Outcome` 0=Unverifiable, 1=Confirmed,
2=Refuted; `Tier` 0=static, 1=state, 2=runtime. Only `symptom_check` and the inner
`sql`/`why` are snake_case. `verdict.result` by contrast is snake_case wire
format. The two shapes must both be handled.

### 3d. Blinding

The fixloop docs (`fixloop_eg_dartsonline/`) are deliberately excluded from the
loop's own input so benchmarks stay honest. They must not enter any set the loop
is evaluated on. Since §2 recommends **eval** as the primary use, this constraint
binds harder than the handoff implies, not softer: the extractor emits only DB
content, never doc text, and `LABELS_benchmark.json` ships as a separate file that
is never concatenated into `input_state`.

---

## 4. Phasing

| phase | work | output | est. |
|---|---|---|---|
| **0** | ~~File the `<no value>` bug~~ — already filed as `bugs_open/016`; evidence appended instead. Raise the 6-of-13 seat gap with the fixloop thread; agree the read contract (artifact kinds + `collected_data` shape, versioned or frozen) | 016 updated; contract note | ½ day |
| **1** | `extract.sql` + `cmd/reasoningset` → JSONL for all 13 trajectories, with `guard`, `input_complete`, provenance | `reasoning_v1.jsonl` | 1 day |
| **2** | Hand-curate `LABELS_benchmark.json` from the one rubric + the ~10 graded runs in NOTES(10); join it in | gold labels attached | ½ day |
| **3** | Quality report: surviving steps per model, per outcome, per task; hand-verify 2 trajectories end-to-end | `NOTES_corpus_quality.md` — **the go/no-go** | ½ day |
| **4** | Owner decision: eval harness / deliberate volume generation / shelve | — | owner call |

Phases 1–3 are ~2 days and produce a corpus you can judge. **Phase 3 is a real
gate** — if the surviving-step count per model slice is as small as §1d predicts,
the honest output is "this is an eval set," and Phase 4 should not default to a
training run.

---

## 5. Verification

- **Counts reconcile.** JSONL record counts per task/model must equal the SQL
  aggregates in §1d. A mismatch means a join dropped rows — most likely a
  uuid/text cast.
- **Hand-verify two trajectories end-to-end**, one CONFIRMED and one
  guard-degraded, against the documented artifact fetch
  (`090_TRIGGER_needs_diagnosis_v1.sh:364-371`): the emitted `input_state` must
  byte-match the `kind='bundle'` body for that correlation and iteration.
- **Guard block spot-check.** Find a verdict where raw ≠ coerced; confirm
  `guard.diagnostic` matches the prepended sentence in the trail's
  `NeededEvidence`, and that `guard.tripped` is false where raw == coerced.
- **Blinding check.** `grep -c 'fixloop_eg_dartsonline\|RUBRIC_\|NOTES_running'`
  over the JSONL must be 0.
- **Exclusion audit.** Every `input_complete: false` row must carry a non-null
  `exclude_reason`, and the count of `no_value_injection` rows must equal the
  §1a table (19 + 13 + 2 + 1 + 1 = 36).

---

## 6. Scope boundary

This thread does **not** fix the `<no value>` bug or the seat-coverage gap, change
the loop, or touch `platform/`, `internal/`, `pkg/`. It reads. The one coordination
point is the read contract: if the fixloop thread changes the verdict or council
JSON shape, this ETL breaks — so that shape needs to be either frozen or versioned
before Phase 1.

Note the corpus is about to become **bimodal**: pre-13:15:11Z repropose rows have
blank objections, post-fix ones will not. That is a provenance boundary, not a
reason to wait — extract now, flag by timestamp, and the split is itself a useful
natural experiment (same task, same models, with and without objections visible).
