# PLAN — make a structural plan refusal recoverable (bugs_open/099, candidate 2)

**Opened:** 2026-07-30. **Bug:** `bugs_open/099_HANDOFF_2026-07-26_feature_designer_plans_die_on_a_rule_it_is_never_told.md`.
**Scope:** candidate 2 only. Candidate 1 (tell the designer one rule, migration `222`)
is applied and live since 2026-07-26 and is NOT revisited here.

## Why this one, and why now

Picked from `bugs_open` as the next case no other thread is working:

- bug file last touched **07-27 17:54**; owning lane `work_item_completion_integrity`
  last committed **07-27**. Three days quiet.
- no open `site_work_items` naming it; none of 07-30's 60+ commits reference it.
- the file itself says candidate 2 "remains the durable fix and is **not done**".

## The defect, in one line

`diagnose_persist_fix_plan` returns a bare Go `error` when a plan fails structural
validation. That fails the step, the workflow's `error_step` routes to
`complete_refused`, and **a completed, good design is discarded** — with
`orchestration_states.error` NULL and the reason only in
`collected_data->'__step_error'`. Candidate 1 taught the designer *one* of the
validator's rules; the validator has a dozen more, and tripping any other one
still destroys the design silently.

## Why this is a framework fix, not a per-rule fix

Candidate 1's shape (state rule N in the prompt) does not close the class: it has to
be repeated per rule, per agent, and it drifts the moment the validator changes. The
class fix is to make the refusal **recoverable and durable** — hand the problems back
to the producing step for a bounded number of repair rounds, and record the refusal
so nothing is destroyed silently. That is rule-agnostic: it works for the dozen rules
that exist and for every rule added later, with no prompt edit.

## Design

### Go — `platform/orchestration/actions/diagnose_persist_fix_plan_action.go`

Two new config keys on the step:

| key | default | meaning |
|---|---|---|
| `repair_step` | `""` | workflow step to route a refusal to. **Empty ⇒ behaviour is exactly as today.** |
| `max_repair_attempts` | `1` | bound on repair rounds. Only consulted when `repair_step` is set. |

On a **structural validation** failure (the `problems` lists from `validateFixPlan`
and `validateStagedPlan` — and *only* those; see "Deliberately out of scope"):

1. `repair_step == ""` → return the error, unchanged. Existing consumers are untouched.
2. Persist the refusal as a durable artefact (rejected plan verbatim + the problems).
3. Count this correlation's refusals **after** the insert, so the count includes this
   one — the same idiom `diagnose_council_decide` uses for its round counter
   (`diagnose_council_decide_action.go:504-520`), scoped by `correlation_id` **and**
   `orchestration_id` for the reason recorded there: the correlation belongs to the
   *diagnosis* and accumulates across re-runs, so correlation alone starts a fresh run
   mid-budget.
4. `attempt <= max_repair_attempts` → return a **result** (not an error) carrying the
   routing flag and the problems. Otherwise → return the error, naming exhaustion.
5. If the refusal insert itself fails → fall back to returning the validation error.
   A bookkeeping failure must never swallow the refusal.

Fail **closed** on a count error (treat as exhausted), following the precedent in
`diagnose_council_decide_action.go:524-526`: a counting failure must not grant extra
LLM rounds.

### The artefact kind — reuse, no DDL

`diagnosis_artifacts.kind` carries a CHECK constraint
(`kind = ANY (ARRAY['bundle','iteration_note','fix_plan','council_report','escalation'])`),
so a new kind would fail **at runtime**, not at build. Rather than DDL on a shared
table, the refusal is written as the already-allowed `iteration_note` with
`metadata->>'note_kind' = 'plan_validation_refusal'` as the discriminator.
`iteration_note` is in the constraint, has **0 rows**, and has **no Go reader** — a
reserved vocabulary slot never used. The only reader of arbitrary kinds
(`fixloop_digest_action.go:229-245`) aggregates `kind || ':' || count` and is
kind-agnostic, so a new label is additive there.

### Result contract

Success path gains `plan_valid: true`. Refusal path returns:

```
persisted: false · plan_valid: false · should_repair_plan: true · repair_attempt: N
validation_problems: []string · validation_problems_text: string
rejected_plan_json: string
```

**The rejected plan is deliberately NOT returned as `plan_json`.** `repropose`'s
prompt renders `{{.plan_persisted.plan_json}}` as "Your previous plan"; reusing that
key would let a *rejected* plan read downstream as a persisted one. A distinct key
makes that mistake unrepresentable.

### Config — `feature-designer` (the bug's actual subject)

- `persist_plan.next_step`: `review_editquality` → `check_plan_valid`
- `persist_plan.config`: add `repair_step: "repair_plan"`, `max_repair_attempts: 1`
- new `check_plan_valid` — `conditional`, `plan_persisted.plan_valid == true`
  → `review_editquality`, else → `repair_plan`
- new `repair_plan` — `execute_llm_prompt`, prompt states the exact problems and the
  rejected plan, `next_step: persist_plan`, `output_field: proposal`

Condition orientation is deliberate: the **council path requires an explicit `true`**.
`compareValues(nil, "true")` is false, so if the field ever fails to resolve the run
routes to repair (stalls, visibly) rather than to review (ships unvalidated). The safe
failure direction is the one that does not reach a reviewer.

## Why NOT simply route to the existing `repropose`

The bug file's candidate 2 says "route the validation problem back into `repropose`
(which exists)". **Checked, and it does not work as written.** `persist_plan` runs
*before* any council, so on a first-pass refusal `repropose`'s prompt renders
`{{.council_reviews.body}}` and `{{.check_results.results_text}}` with nothing behind
them, and it frames a structural problem as a council objection. A dedicated
`repair_plan` step keeps council semantics clean and names the real problem. This is a
correction to the bug file's candidate, recorded there too.

## Deliberately out of scope

- **Truncated / invalid JSON and the byte cap.** These stay terminal. A truncated
  completion is a `max_tokens` fault, and the truncation family (`bugs_open/012`,
  `138`) is explicit that a cut completion must not be silently retried into a loop.
  Recoverable means *structurally invalid*, not *cut off*.
- **Relaxing the validator.** The bug file forbids it and is right: one edit per file
  per stage is what makes a stage a single reviewable commit.
- **`council-gate` / `fix-proposer` opt-in.** The mechanism is available to both and
  neither is changed here. `council-gate` deliberately stays untouched — it is the
  gate this change itself is reviewed by, and disturbing it in the same change that
  it reviews is a needless coupling.

## Consumers to tell (owner ruling 2026-07-29 §3)

`diagnose_persist_fix_plan` has three live consumers:
`council-gate → persist_submission`, `feature-designer → persist_plan`,
`fix-proposer → persist_plan`. What changes about their guarantee: **nothing, unless
they set `repair_step`.** With it unset the action returns the same error on the same
inputs. Named in the council submission.
