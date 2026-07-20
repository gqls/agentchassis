# CORPUS QUALITY REPORT — Phase 3, the go/no-go

*2026-07-20. The gate the plan set: extract, label, then decide whether the
signal is worth a training run before investing in scale. Every figure is from
`reasoning_v1.jsonl` (820 records, extracted 2026-07-20 11:18Z) cross-checked
against the live DB. Method and commands: `RUNBOOK_reasoning_dataset.md`.*

---

## Verdict, up front

| question | answer |
|---|---|
| Train a reasoning model on this? | **No.** 102 usable diagnosis steps across two model generations. |
| Use it as an evaluation set? | **Yes, modestly.** 24 gold-graded steps over 6 trajectories, with a genuinely unusual property — see §4. |
| Was the guard block worth building? | **Yes, but it is smaller than we claimed.** See §5 — a correction. |
| Where is the unexploited volume? | The **council lane**: 574 usable steps, zero gold labels. §6. |

**Recommendation: stop extracting, start labelling — or generate volume
deliberately.** The pipeline is built and proven; more passes over the same
13 diagnosis runs will not produce more signal.

---

## 1. What survives the filters

820 records → **689 usable** (`input_complete: true`), 131 flagged.

| task | total | usable | sonnet-4-6 | sonnet-5 |
|---|---|---|---|---|
| `council_review` | 676 | **574** | 91 | 483 |
| `diagnosis_verdict` | 105 | **102** | 70 | 32 |
| `propose` | 16 | 13 | 12 | 1 |
| `repropose` | 21 | **0** | — | — |
| `reframe` | 2 | **0** | — | — |

Exclusions, all flagged with a reason rather than dropped:

| reason | rows | |
|---|---|---|
| `no_value_injection` | 69 | reasoned against blank inputs (`bugs_open/016`) |
| `blinded_docs` | 43 | cites fixloop docs — must not be evaluated against |
| `truncated` | 16 | `output_tokens >= max_tokens`; the completion was cut |
| `call_failed` | 3 | the call errored |

**The `repropose` and `reframe` lanes are a total loss — 23 of 23 rows.** Every
one predates the 016 render fix, so every reviser in the corpus was revising
against blank objections. The handoff named these as the premium
"objection → resolution" signal. There is none. That lane can only be refilled by
running the loop again post-fix.

## 2. The bimodal split is real and lopsided

Records whose **run started** before the 016 fix (13:15:11Z, 2026-07-18):

| task | pre-fix | post-fix |
|---|---|---|
| `diagnosis_verdict` | 89 | 16 |
| `council_review` | 160 | 516 |
| `repropose` | 17 | 4 |

Do not pool these. Note the inversion: diagnosis is overwhelmingly *pre*-fix
while council is overwhelmingly *post*-fix, because the council has kept running
(this thread's own six submissions are in there) while no new diagnosis has been
dispatched. The 4 post-fix repropose rows are still flagged — they belong to runs
that started before the fix.

## 3. Gold labels

**24 graded records across 6 trajectories** — 9 PASS, 10 PARTIAL, 5 FAIL. All are
`diagnosis_verdict`; `LABELS_benchmark.json` carries 12 runs, but two produced no
records at all (correctly — one never started, one lost its spawn), and three
grade a *subsystem* rather than a diagnosis.

| trajectory | grade | terminal | steps |
|---|---|---|---|
| `e505f70f` | PASS | approved | 5 |
| `5120c0dc` | PASS | **abstained** | 3 |
| `b606dbf6` | PASS | **abstained** | 1 |
| `dd1186b9` | PARTIAL | confirmed_partial | 5 |
| `960b554d` | PARTIAL | escalated | 5 |
| `4d43d002` | **FAIL** | confirmed_wrong | 5 |

## 4. What makes this set unusual, and worth keeping

Three properties that are hard to buy or construct:

**Two of the three PASSes are abstentions.** `5120c0dc` and `b606dbf6` both
refused to answer and were graded PASS for it — one because a confident
half-answer would have shipped to a fixer, one because the symptom it was given
embedded unevidenced claims and cite-or-abstain correctly refused. A benchmark
where *declining to answer* is the gold behaviour is rare, and it is exactly the
behaviour a reasoning model is hardest to train for.

**The FAIL is a confident wrong answer.** `4d43d002` returned CONFIRMED,
`stopped_by=confirmed`, after 5 iterations. The notes head that turn *"the loop
FAILED the rubric — instructively."* A negative where the model was sure is worth
more than one where it hedged.

**`960b554d` is a labelled raw-vs-gated pair.** Its raw verdict was CONFIRMED and
"rubric-perfect: 5 static citations"; the coercion gated it to UNVERIFIABLE at
iteration-cap; the human grade is PARTIAL on the gated output. Three different
judgements of one step, all recorded.

## 5. CORRECTION — the guard block is real but much smaller than we claimed

The plan called the guard corpus *"probably the most distinctive signal in the
corpus"* and *"the most distinctive thing in this corpus"*. Measured:

**7 guard trips. All seven carry the same diagnostic** — *"confirmed on one
evidence family only; a confirm needs BOTH a static (code/schema) citation
showing the mechanism AND a state/runtime citation…"*. A further **14** records
are `guard.alignment: ambiguous` and assert nothing.

So it is not a rich taxonomy of invalid reasoning. It is **seven examples of one
failure mode**, which is a nice illustration and not a training signal. The claim
was written before the extractor existed and was never checked; it is corrected
here and in the plan rather than quietly dropped. The mechanism is still worth
having — it is free, and it will grow if the loop runs more — but it should not
be cited as a reason to invest.

## 6. Where the unexploited volume actually is

**The council lane: 574 usable steps, and not one gold label.** It is 5.6× the
diagnosis lane and entirely unlabelled. Each record is a reviewer's argument
about a plan, with `approve`/`object`/`veto` as a natural decision — and multiple
seats judging the *same* plan, which is a preference set, the format reasoning
post-training actually wants.

Two things would unlock it, neither requiring a platform change:

1. **Terminal outcomes as weak labels.** Did the plan a council approved get
   committed with a `Council-Reviewed` trailer? Did an objection get answered in
   the next round? Joinable from `diagnosis_artifacts` + git.
2. **Inter-seat disagreement as the signal.** Where seats split on one plan, the
   reasoning is load-bearing. That is computable from the corpus we already have,
   today, with no new labelling.

This thread's own six council submissions are in this lane — with the objections,
our answers, and the outcomes all recorded.

## 7. What to do next, in order

1. **Mine the council lane** (§6). Biggest return, no new spend, no platform change.
2. **Harvest `feed-triage`** — a few hundred calls carrying thousands of batched,
   already-structured judgements. One query.
3. **Decide on deliberate volume.** The cheapest form is replaying the ~30 cases
   in `/bugs_open/` + `/bugs_closed/` where the correct diagnosis is already
   written down, so grading is nearly free. N bugs × M models generates graded
   trajectories on demand — and would refill the dead `repropose` lane post-fix.
4. **Do not extract again for its own sake.** The pipeline is proven; re-running
   it over the same 13 diagnosis runs adds nothing.

## 8. Threats to validity

- **Six trajectories is a small benchmark.** It can rank two obviously-different
  models; it cannot resolve close ones.
- **The graded set is nearly all pre-fix and mostly sonnet-4-6** (70 of 102
  diagnosis steps). A model evaluated on it is being judged against a corpus
  produced by an older generation under a since-fixed defect.
- **Labels are one curator's reading** of one notes file. They are traceable —
  every entry quotes its source line — but they are not independently graded.
- **`blinded_docs` is a coarse filter.** It flags any row mentioning the fixloop
  docs, which over-excludes (a council submission *about* those scripts is not
  the same as leaking rubric answers). Better to over-exclude than leak, but the
  43 rows are worth a second look before anyone treats them as lost.
