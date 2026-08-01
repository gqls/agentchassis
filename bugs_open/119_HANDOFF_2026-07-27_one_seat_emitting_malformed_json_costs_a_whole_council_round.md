# 119 — one seat emitting structurally malformed JSON voids a whole council round, even when eight seats approved

**Filed** 2026-07-27 by the bug-sweep thread, off its own `bugs_open/112` submission ·
**Status** OPEN, unowned · **Severity** medium — no correctness risk, pure waste:
each occurrence costs one full council round (credits + ~30 min latency) and
returns a verdict that says nothing about the plan ·
**Distinct from `bugs_closed/019`** (truncation), and 019's fix structurally
cannot cover it — see "Why this is not 019".

---

## Symptom

A council round returns `REVISE` with

```
decided_by : unreadable reviewer(s): review_guardian.result
reviewers  : 10
abstained  : 5
unreadable : 1
```

and **not one word of the verdict is about the plan**. In the observed case
(`SUBMISSION_CORR dfa6205e-b10e-440c-b251-5d791fdeb718`, 2026-07-27 18:21 UTC)
the readable seats were: `editquality` approve, `guidelines` approve,
`diagnosis_guardian` approve, `llm_reliability` approve, `debug_historian`
approve, `constitution` approve, `mission` approve, `prior_art_librarian`
approve, `reuse_agent` object (medium), `tooling_provenance` object (medium).

**Eight approvals, two medium objections, no veto** — and the round was still
spent, because one seat's output would not parse.

## Root cause — the seat emits a complete, well-reasoned review as INVALID JSON

The guardian's result is not truncated and not empty. It is a full review with
two low-severity objections and two `missing` items, and it ends mid-sentence
nowhere — the final characters are a closing `"}`. It simply has a structural
slip: **the review object is closed one bracket too early, leaving `checks`
outside it.**

The stored value, with the break marked:

```
{"reviewer": "guardian", "verdict": "object",
 "objections": [ {...}, {...} ],
 "missing":    [ "...", "..." ]}]     <-- the object closes, then a STRAY ]
 , "checks": [ {"sql": "SELECT ..."} ],
 "notes": "..."}
```

Parsed:

```
json.loads(result)  ->  Extra data: line 1 column 2247 (char 2246)
```

`json.Unmarshal` reads a complete, valid object and then rejects the trailing
content. The seat's *reasoning* survives intact in `collected_data`; only its
machine-readability is lost.

### Why the recovery path does not save it, and is right not to

`ParseLLMJSON` has a Tier 3 (`extractCompleteJSONValue` /
`scanTopLevelJSONValues`, `platform/orchestration/actions/json_envelope.go`)
built for exactly "the answer is complete but not alone". It decodes consecutive
top-level values with a `json.Decoder`, which does stop cleanly after the first
object. It then finds the next `{`/`[` — which here is the `[` of `"checks": [`
— and decodes that array as a **second** top-level value.

It then applies its own rule, and refuses:

> *"a multi-value response is only accepted when every value is an object with an
> IDENTICAL key set, which is what a re-emission looks like and what a
> several-answers-in-one-response does not."*

Here the two values are an **object** and an **array**, so the key sets are not
identical and the whole response is rejected. **That rule is correct and is
evidence-backed** (measured over 5,844 stored responses; relaxing it is how a
hero section gets handed a testimonials object — see the comment at
`json_envelope.go:146-163`). This case should not be fixed by loosening it.

### Why the decider then voids the round, and is also right

`platform/orchestration/actions/diagnose_council_test.go:210-222` enforces,
deliberately:

> `approve alongside an unreadable seat must downgrade to revise`

That is the `bugs_closed/019` contract and it is sound in principle: an
unreadable seat might have been the veto, and silently ignoring it would let a
guardian's objection vanish. **The defect is not the downgrade. It is that
nothing retries the seat**, so a recoverable emission slip is converted straight
into a lost round.

In this instance the downgrade was, in hindsight, protecting nothing: the
guardian's actual verdict was `object` with two **low**-severity objections,
which would not have gated approval. The round was spent to guard against a
possibility that had not occurred — and which one retry would have resolved.

## Why this is not `bugs_closed/019`

019 is **truncation** — a response cut at the output ceiling, with no complete
value to salvage. Its fix (`tolerate_truncation`) degrades the step instead of
voiding the run. It cannot apply here, because there is nothing truncated:

```
llm_call_log, the round's 11 calls, 2026-07-27 18:17-18:24 UTC
  model       = claude-sonnet-5
  max_tokens  = 8000
  output_tokens: 317 … 5663   (highest 5663)
  hit_ceiling = false on EVERY row
  success     = true on EVERY row
```

The guardian call succeeded, well inside its budget, and returned a complete
document. **A truncation guard is the wrong instrument for a well-formed-length
response with a misplaced bracket** — which is why this needs its own case.

## Evidence that it is recurring, not a one-off

The last 18 council reports (`diagnosis_artifacts` where `kind='council_report'`):

| unreadable ≥ 1 | rounds | of which DECIDED by the unreadable seat |
|---|---|---|
| yes | 3 | **2** |
| no | 15 | — |

```
2026-07-27 | revise   | 10 seats | unreadable 1 | decided_by: unreadable reviewer(s): review_guardian.result
2026-07-27 | revise   | 11 seats | unreadable 1 | decided_by: unreadable reviewer(s): review_editquality.result
2026-07-26 | revise   | 10 seats | unreadable 1 | decided_by: gating objection from prior_art_librarian
```

**Two of the last eighteen rounds (~11%) were voided by a parse failure rather
than by a judgement.** It has now hit **two different seats** (`guardian`,
`editquality`) on two different days, so this is not one bad seat prompt — it is
general emission fragility across the roster. [MEASURED — the query is above;
the sample is 18 rounds, which is every council report in the table.]

Cost per occurrence: one round of council credits (10-11 seats), ~30 min of
dispatch-plus-run latency, and a resubmission that must re-state the entire
evidence base because seats have no cross-round memory.

## Fix candidates, ordered by what closes the door

**1 — Retry a seat once when its result fails to parse.** The narrowest change
that removes the class. The seat is a single step; re-running only it costs one
seat's credits against the ten-to-eleven a voided round costs. Keeps the 019
downgrade contract fully intact — an unreadable seat still blocks approval, it
just gets one chance to be readable first. Only if the retry also fails does the
round degrade as it does today.

**2 — Constrain the emission instead of catching it.** The seats already have a
strict output contract (only `{approve,object,veto}` parse; only
`{reviewer,verdict,objections,missing,notes,degraded}` persist). If the provider
supports enforced structured output for this call, a bracket slip becomes
unrepresentable rather than recoverable. Larger change, touches every seat, and
is the only candidate that makes the bad state impossible rather than survivable.

**3 — Salvage the leading object when the remainder is not a competing answer.**
Tempting and I recommend AGAINST it as stated: the Tier 3 rule it would weaken is
backed by a 5,844-response measurement, and the failure it prevents (shipping a
fragment as a whole answer) is worse than a lost round. If anything in this
direction is done it must be scoped to the *council seat* consumer only — which
knows its schema and can check the recovered object has the required keys — and
never to the shared parser.

**4 — Report the waste rather than fix it.** Add `unreadable` seat counts to the
098 coverage report so the tax is visible. Does not fix anything; worth doing
anyway, because right now this is only discoverable by reading `decided_by` on an
individual verdict, and a reader who checks `abstained` instead will not see it
at all.

## How to verify a fix

1. **Reproduce the parse failure in a unit test** from the stored artefact, not
   from a hand-written string — the real one is in `collected_data->'review_guardian'`
   on the orchestration for correlation `dfa6205e-b10e-440c-b251-5d791fdeb718`.
   `orchestration_states` retains 13 days, so capture it before ~2026-08-09.
2. **Induce the failing branch**, not the happy one: feed a seat result with
   trailing non-matching content and confirm the seat is retried, and that a
   still-unreadable seat STILL downgrades an approve to revise. A green round
   proves nothing here — the 019 contract must survive the fix.
3. **Measure the rate again** with the query in the evidence section above. The
   figure to move is "rounds decided by an unreadable seat", currently 2 of 18.

## Pointers

- `platform/orchestration/actions/json_envelope.go:108-228` — the tiered parser
  and, in its comments, the measured reasoning for the rule that rejects this.
- `platform/orchestration/actions/ai_actions.go:448-470` — the truncation
  tolerance path and its note that the council decider "degrades an unreadable
  seat to a loud abstention and refuses to let an approve stand alongside one".
- `platform/orchestration/actions/diagnose_council_test.go:202-228` — the
  downgrade contract, with the test that pins it.
- `bugs_closed/019` — the truncation sibling. Read it first to see why its fix
  does not reach this.
- `RUNBOOK_council_gate.md:104` — "Reading a verdict: check `unreadable`, NOT
  `abstained`". That guidance exists because this happens; this case is the
  attempt to stop it happening.

---

## Reproducible instance, 2026-07-28 — three runs, no verdict, and RESUBMIT_CORR refuted

From the oufe workstream (series-facts change to `datahelpers`). Recorded because
it is the first instance here with a **controlled variable**, and because the
failure shape is worse than the one this bug was filed for: not a bad decision, no
decision at all.

| run | correlation | RESUBMIT_CORR | outcome |
|---|---|---|---|
| round 1 | `da40ddf0` | no | **verdict produced** — REVISE, gating objection from `editquality`, a substantively correct one |
| round 2 | `da40ddf0` | yes | `COMPLETED / complete_invalid`, no `council_report`, no verdict note |
| round 2 retry | `da40ddf0` | yes | identical |
| **control** | `fb593614` | **no** | **identical** — so the resubmission path is NOT the cause |

**What is ruled out.** `complete_invalid`'s own description is "the submission
failed structural validation, or a reviewer/decide step errored". Structural
validation passed on every run: a `fix_plan` artifact was persisted each time
(control: `fix_plan @ 12:57`), carrying the round-2 content. So
`persist_submission` succeeded and the failure is downstream, in a review or
`council_decide` step.

The chassis roll to `v1.0.1185` is also ruled out: the first round-2 failure was
at 12:21 and the pod restarted at 12:38.

**What changed between the round that worked and the three that did not** was only
the submission prose — a longer `rationale` (3,040 bytes), a rewritten `sketch` on
edit 0, and a longer `risks`. Same 3 edits, same 3 files, same 5 `grounded_in`
quotes, 8,775 bytes total, well inside the documented 8-edit / 64KB limits.

**Why this matters more than a wasted round.** The 119 headline case is one seat's
unparseable JSON costing a round — expensive but visible, because a verdict still
appears and names `decided_by`. This shape produces **no artefact a submitter can
read at all**: no `council_report`, no `doc_notes` row, and a terminal status that
looks like completion. A submitter who does not query `current_step` sees
`COMPLETED` and reasonably concludes the council approved in silence.

**The trap that nearly caught me.** The most recent `council-gate` note at the time
was a *different* submission's APPROVED. Reading the latest note rather than
querying by my own correlation would have had me commit a `Council-Reviewed:`
trailer for a verdict belonging to someone else's plan. Always:

```sql
SELECT body FROM doc_notes
WHERE categories ? 'council-gate' AND body LIKE '%<YOUR_CORR>%'
ORDER BY created_at DESC LIMIT 1;
```

**Suggested next step for whoever owns this:** capture the failing step's own
error. `complete_invalid` is reached from 18 different steps and records which one
nowhere the submitter can see, so the first fix is diagnosability — put the
originating step and its error into the completion artefact. Until then a
submitter cannot tell "my JSON is bad" from "a seat fell over", which is the
difference between fixing it and re-running it.

---

## 2026-08-01 — TAKEN, FIXED IN CODE, and this file's headline mechanism could not be reproduced

Picked up by the "bugs_open/119 seat-retry" thread. `who-owns.py` named no owning
workstream and no active session transcript mentioned 119, so this was unowned as the
header says. Workstream docs: `docs024_key_docs_latest/bugfix_119_seat_retry/`
(PLAN / NOTES / RUNBOOK / README_where_we_are).

### 1. The damage is real, recurring, and worse than this file measured

This file measured "2 of the last 18 rounds". The all-history figure, from **all 424**
`council_report` artifacts (RUNBOOK R1):

| week beginning | rounds | ≥1 unreadable seat | **voided by an unreadable seat** |
|---|---|---|---|
| 2026-07-06 | 10 | 0 | 0 |
| 2026-07-13 | 85 | 0 | 0 |
| 2026-07-20 | 162 | 17 | **8** |
| 2026-07-27 | 167 | 21 | **15** |

**23 rounds all-time were decided by a parse failure rather than by a judgement.** Seven
distinct seats have done it — `editquality` 25, `guidelines` 4, `prior_art` 3, `guardian`
3, `deferral_honesty` 2, `checkability` 1, `improvement_guardian` 1 — which confirms this
file's "general emission fragility across the roster", now on a sample of 39 rather than 2.

The cost shape is sharper than "one wasted round": correlation `c5219a69` burned **three
consecutive rounds** (2026-07-31, 15:32 / 15:37 / 15:42), same seat, same orchestration,
ten minutes. Nothing re-asked, so every retry was a fresh round of 10-13 seats.

### 2. [MEASURED] The headline mechanism has ZERO live instances — the DEFECT is entirely real

> **Both halves of this matter and they point different ways.** This file's *defect*
> ("nothing retries the seat") is 100% true of the code and is costing rounds weekly. This
> file's *mechanism* (a complete review closed one bracket early, leaving a stray `]`)
> could not be reproduced at all.

Classifying every unreadable seat against its orchestration (RUNBOOK R3): of 39
instances, **36 orchestrations were already pruned** and all **3** survivors are total
truncations with an *empty* partial — `{"type":"text","result":"","__truncated":true,
"__truncated_output_tokens":8000}`. Fleet-wide, across **785** JSON-declared step outputs
in the live orchestration window (RUNBOOK R6): **782 parsed, 2 unparsable (both
truncated), 0 complete-but-malformed.**

So the fix below targets the **defect** (any unusable answer), not the filed mechanism —
which is why its corrective text branches on which failure actually occurred. A fix aimed
only at the bracket slip would have fired zero times and been reported as a success.

> **CORRECTION to this file's §"How to verify a fix" step 1.** It says to reproduce from
> the stored artefact and "capture it before ~2026-08-09" because "`orchestration_states`
> retains 13 days". **That is false.** Rows survive from 07-13, but these specific
> orchestrations were gone within days, and correlation `dfa6205e`'s is long pruned. The
> instruction was already unsatisfiable when I read it, four days after it was written.
> *What caught it:* trying to follow it. Logged in `WRONG_CALLS.md`.

### 3. A second cause, upstream, found while reading — the declared key was DEAD

`getOutputType` (`ai_actions.go`) read **`output_type`**. The fleet writes
**`output_format`**, and nothing in the LLM path read it. Measured across all 134 active
`execute_llm_prompt` steps (RUNBOOK R5):

| key | value | steps | agents |
|---|---|---|---|
| `output_format` — **never read** | json | **90** | **32** |
| `output_format` — **never read** | text | 10 | 5 |
| `output_type` — read | json | 5 | 5 |
| `output_type` — read | text | 1 | 1 |

**So 90 steps across 32 agents — every council seat included — asked for JSON and were
never told to produce it**, receiving `getDefaultOutputInstructions()` (which mentions
JSON nowhere) instead of `getJSONOutputInstructions()`, whose first line is:

> `- Ensure valid JSON syntax (proper quotes, commas, brackets)`

That is, word for word, this file's documented failure. [INFERRED, and deliberately marked
— I cannot prove the instruction would have prevented the observed stray `]`. What is
MEASURED is that the seats never received it.] Same class as `bugs_open/134`.

### 4. What shipped — commit `4c1a97874`

Both edits in `platform/orchestration/actions/ai_actions.go`; register entry **WFA-005**.

- **Prevention (this file's candidate 2, cheaply).** `getOutputType` now also reads
  `output_format`, **allow-listed** to `json|text|html|markdown`. The gate is load-bearing:
  `query_database` reads the same key name with a *different* vocabulary (`array`/`object`,
  `database_actions.go:26`).
- **Recovery (this file's candidate 1, as recommended).** When a step **declared** json and
  the answer will not parse, re-ask **exactly once**, corrective rather than identical —
  an identical re-ask reproduces the failure, which `c5219a69` proved three times running.
  Truncated → same judgement materially shorter; malformed → same content as one valid
  JSON value with bracket balance spelled out. It gets its own `llm_call_log` row marked
  `RETRY (bugs_open/119…)`, and on recovery the `__truncated` markers are adopted from the
  response actually returned, so a consumer cannot degrade on discarded evidence.
- **Not taken: candidate 3** (loosen the Tier 3 parser). This file recommends against it
  and is right — that rule is backed by a 5,844-response measurement. Untouched.
- **019 and 138 are untouched by construction.** If the re-ask also fails, every line
  below it runs byte-identically: an unreadable seat still blocks an approval, it just
  gets one chance to be readable first. `salvageTruncatedReview` and the `Degraded` gating
  rule are not modified.

Restores a symmetry the action lacked: it retried a call that never **arrived** (four
attempts on 500/502/503/529) and accepted without question a call that arrived
**unusable**.

**Cost, measured before submitting rather than left for a reviewer:** the retry would have
fired **twice** across 785 JSON-declared outputs. It can only fire on a path already
returning a result the consumer cannot use.

### 5. Why this file is still OPEN

**Go changes are inert until an image is built and rolled.** The bar is *fixed AND live*.
Owed, in order:

1. The chassis roll, then the pod-grep on **every** replica — RUNBOOK R10 carries the
   discriminating strings **and a negative control** (`bugs_open/153`: a roll is not
   evidence your fix shipped).
2. The behavioural check, which outranks the grep:
   `SELECT count(*) FROM llm_call_log WHERE error_message LIKE 'RETRY (bugs_open/119%';`
   Non-zero is the mechanism firing; it should stay **small**.
3. The figure to move, re-measured with RUNBOOK R1: **rounds decided by an unreadable
   seat — 23 of 424 all-time, 15 in the week to 2026-08-01.**

**Do NOT grade this on a green council round.** The healthy path already works; that is
not what is broken.

### 6. Council

Submitted as `576832f3-0d39-48b7-af18-97e06297900f`, committed pre-verdict with a
`Council-Submitted:` trailer per the 2026-07-30 rule. The submission names all 32 affected
agent types explicitly rather than only counting them (owner ruling 2026-07-29 point 3).

### 7. Council round 1 — REVISE, and it caught a real hole (2026-08-01)

`576832f3` round 1: **REVISE**, gating objection from `debug_historian` (high), 14
reviewers. Four seats converged on one thing and **they were right**: I had asserted that
`output_format` sits at a depth `getOutputType` reads, citing a 90-step blast radius, and
never measured it. The house landmine they cited is real — a step's prompt and its token
cap sit at *different* depths in `default_config`.

**Measured, both depths walked:** all 101 declarations sit at `config.output_format` — the
map `params.StepConfig.Config` is — with **zero** under `config.ai_service` and **zero** at
the step root. The assumption held. But the re-measurement **corrected my own numbers**,
because my census walked only `workflow.steps` and missed steps nested in a loop's
`sub_workflow`:

> **135 LLM steps, 101 declaring `output_format`, 91 of them json across 33 agents** —
> not 90/32. The missed step is `page-content-writer` → `process_sections_loop` →
> `generate_content`. A census over `agent_definitions` that walks only `workflow.steps`
> is blind to every nested step; that is what WFA-003 exported `WalkSteps` for, and my
> SQL did not use it.

**`bug_historian` (medium) was the substantive one** and is answered with code: a re-ask
lowers the *frequency* of the silent success but not its *shape* — the step still returns
`{result: text, type: text}` and still SUCCEEDS, which is `bugs_closed/076`'s title and
the `missingkey=zero` family. Added **`__json_contract_unmet`**, stamped only when json
was declared AND still not delivered after the re-ask. **A marker, not a hard error:**
failing loud would convert ~91 currently-succeeding steps into failing ones over content
they did not author — `bugs_closed/073`'s defect, and the reason this file's own candidate
2 was declined. Committed `085160a6c`; resubmitted on the same correlation.

### 8. A FOURTH specimen, produced by the review of this very fix

Round 1's own council report carried `unreadable: ["review_llm_reliability.result"]`.
Captured before pruning:

```json
{"type":"text","result":"","__truncated":true,"__truncated_output_tokens":8000}
```

**Four in-retention specimens, four truncations, zero bracket slips.** It also shows the
cap treadmill in one picture — the live per-seat caps:

```
16000: review_editquality, review_guidelines, review_prior_art, review_architecture
 8000: the other THIRTEEN, incl. review_llm_reliability
```

The four at 16000 are exactly the four that had previously failed. `bugs_open/138`'s lane
raised each cap *after* its seat broke, and round 1's failure landed on one of the thirteen
left behind — which is `platform/aiservice/truncation.go:26-29`'s warning happening in
real time, and the strongest argument for a re-ask that asks for **brevity** rather than
headroom.

**Note for whoever closes this:** §2's claim that the filed mechanism is unmeasurable now
rests on 4 specimens rather than 3, all agreeing. That is still a small sample, and the
honest reading is "the bracket slip is not what is biting today", **not** "it never
happens".
