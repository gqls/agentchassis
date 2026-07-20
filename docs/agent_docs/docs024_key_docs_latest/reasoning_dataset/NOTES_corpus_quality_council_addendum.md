# ADDENDUM — the council lane, mined (2026-07-20)

*Follows §6 of `reasoning_dataset/NOTES_corpus_quality.md`, which named this lane
as the biggest unexploited volume and proposed two ways to label it. Both were
tried. One works well; one is much weaker than I claimed there.*

---

## What was built

`extract.sql` gained a fourth query emitting **one row per (council report ×
seat)**; `cmd/reasoningset` gained `buildCouncil`. Corpus goes from 820 records
to **1,950 across 175 trajectories** — 836 of them seat reviews, the lane that
was previously extracted only as undifferentiated `council_review` LLM calls.

Every council label is **derived, not hand-curated.** That is what makes this
lane usable at scale: no labelling campaign is needed.

## Proposal 1 — terminal outcomes via the trailer: **much weaker than claimed**

§6 proposed joining `Council-Reviewed:` trailers to see whether an approved plan
was actually committed. Measured: **4 trailers exist in the entire repo history**,
against 44 submissions. That is not a label source; it is an anecdote.

> Corrects §6, which presented this as one of two equal unlocks. It is not.

## Proposal 2 — inter-seat disagreement: **works, and it is the signal**

**But the first version of the metric was wrong.** I defined dissent as the seat
disagreeing with the *round's decision*. The council returns `revise` if **any**
seat objects — so that definition labels every approver in a revise round a
dissenter, and measured **520 of 836 (62%)**, which is absurd on its face: those
seats are the majority. Redefined against the seats in the **same round**, it
measures **201 (24%)**, and the per-seat spread is real:

| seat | votes | dissents from its own round |
|---|---|---|
| `bug_historian` | 53 | **62.3%** |
| `honesty` | 44 | 52.3% |
| `guardian` | 82 | 42.7% |
| `editquality` | 80 | 30.0% |
| `feasibility` | 44 | 22.7% |
| `improvement_guardian` | 24 | 16.7% |
| …six seats | — | **0%** |

## Two findings about the council itself

**1. Unanimity essentially never happens.** 822 of 836 votes sit in a round where
seats split. So `contested` is a near-useless filter — it is almost always true —
and the *per-seat* dissent rate above is what carries information. Recording this
because the flag is emitted and a consumer would otherwise assume it discriminates.

**2. Six seats have never objected, and they are not rubber stamps.** `guidelines`
(42/42 approve), `mission`, `constitution`, `llm_reliability`, `adoption_guardian`,
`compliance`. Sampled before concluding: they write 250–1,000 character **scope
determinations** — *"this is an infrastructure-level fix … it doesn't touch
classifier logic, revenue-shape decisions"*. That is a legitimate and correct
approve for a relevance-gated seat.

But it means **their `approve` does not mean what `editquality`'s `approve`
means** — "outside my footprint" versus "I checked the merits and it is good".
Pooling the two would teach a model that those are the same judgement. 155
records are flagged `scope_approve_risk` so a consumer can separate them.

## What a consumer gets

```jsonc
{ "task": "council_seat_review", "step_name": "bug_historian",
  "decision": "OBJECT",
  "reasoning": [ /* the objections, with severity and edit index */ ],
  "reasoning_raw": "…the seat's notes…",
  "labels": { "round_decision": "revise", "dissent": true, "contested": true,
              "scope_approve_risk": false } }
```

Several seats judging **one plan**, with each seat's agreement or dissent from
its peers attached. That is a preference set, which is the format reasoning
post-training actually wants — and it exists today at 5.6× the diagnosis lane
without anyone labelling anything.

## Honest limits

- **No model provenance.** Council rows come from `diagnosis_artifacts`, which
  does not record which model produced a seat's review. It is joinable to
  `llm_call_log` on (correlation, `review_<seat>`) but is not joined today, so
  these rows cannot be sliced by model generation the way verdicts can.
- **`input_state` is empty** on council records. The plan under review lives in
  the `fix_plan` artifact on the same `trajectory_id`; the join is available but
  not materialised, so a record is not yet self-contained.
- **Dissent is a peer measure, not a correctness measure.** A seat dissenting
  often is discriminating, not right. Nothing here says which seat was correct —
  that would need the terminal outcomes that proposal 1 just failed to supply.

---

# ADDENDUM 2 — the feed-triage lane (2026-07-20)

*The third and last lane named in `NOTES_corpus_quality.md` §7. It is the
largest, and the only one where a judgement and its outcome sit on the same row.*

## Result

**Corpus: 820 → 7,712 records across 448 trajectories.** The feed lane alone
contributes **5,671 judgements**, of which **5,658 carry a terminal outcome**:

| outcome | judgements |
|---|---|
| `rejected` | 3,472 |
| `expired` | 1,385 |
| `relevant` | 782 |
| `review` | 19 |

Each record carries the model's own reasoning as emitted — `reason`,
`credibility`, `credibility_reason`, `topics`, `source_attribution` — against the
title and summary it was judging.

## The outcome is NOT a restatement of the judgement — checked, not assumed

The obvious worry: if `status` is just a threshold on `relevance_score`, then
using it as a label is circular — you would be training a model to predict its
own threshold. Averages looked exactly like that (`rejected` 24.5, `relevant`
71.4).

It is not. `rejected` spans **0–90**, and the bands overlap heavily:

| score band | rejected | relevant | expired |
|---|---|---|---|
| 60 | 409 | 73 | 82 |
| 70 | 296 | 215 | 321 |
| 80 | 170 | 307 | 584 |
| 90 | 27 | 175 | 346 |

**492 judgements scored ≥60 and were rejected anyway**, and `duplicate_of` and
expiry explain **none** of them (both zero across that whole set). Something
downstream decides and I do not know what. Those rows are flagged
`score_outcome_divergence` rather than explained — they are the most interesting
subset in the lane and the open question attached to it.

```jsonc
{"decision":"score=72","labels":{"terminal":"rejected","score_outcome_divergence":true},
 "reasoning":{"reason":"Brent crude price movements, US-Iran tensions, and Hormuz
  blockade fears are highly relevant to energy market professionals…"}}
```

## A silent truncation class the existing filter misses

**50 of 439 `score_relevance` responses are unparseable JSON**, cut mid-array —
and **zero** of them trip the `output_tokens >= max_tokens` rule that CLAUDE.md
enshrines. They are `success=true` with plausible-looking text. The extractor
counts and warns rather than skipping silently, because a silent skip would
understate the lane by 11%.

Worth surfacing to whoever owns feed-triage: this is the `bugs_open/008`
stop_reason family in a lane nobody has checked.

## Limits

- **No published-page outcome.** `published_page_id` is populated on **0** of
  6,711 rows, and `work_item_id` on 0. The strongest possible label — "did this
  become an article?" — is not recorded, so `relevant` means "passed triage", not
  "was used".
- **`expired` is ambiguous.** 1,385 judgements, averaging a *higher* score than
  `relevant` (72.2 vs 71.4). Judged good, never acted on. That may be a capacity
  signal rather than a quality one; do not train it as a negative.
- **One model.** The whole lane is `claude-sonnet-4-6`, so it cannot be sliced
  by generation.
