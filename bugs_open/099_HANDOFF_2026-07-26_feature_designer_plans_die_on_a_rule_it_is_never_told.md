# HANDOFF — the feature designer's plans die on a validator rule its prompt never states

**Filed:** 2026-07-26, by the `bugfix_077` thread, on the first real use of the
feature builder since it was proven end-to-end on 07-25.
**Severity:** MEDIUM. Not an outage. It silently destroys a completed, good design
and spends the full designer cost to do it — and the failure is *systematic*, not
flaky, so retrying without changing anything buys the same result.
**Status:** OPEN — **for candidate 2 only.** Candidate 1 (the one-line config
fix) IS applied, live, and re-verified after the `v1.0.1174` fleet roll on
2026-07-27; see §"Candidate 1 IS APPLIED" and §"Re-verified 2026-07-27". Anyone
arriving here from a summary that says "a one-line fix is identified but
deliberately not applied" is reading a stale pointer — that was true only on
2026-07-26 morning and the file records the reversal.

---

## The mechanism

`diagnose_persist_fix_plan` refuses a staged plan in which the same file appears
twice in one stage:

```go
// platform/orchestration/actions/diagnose_persist_fix_plan_action.go:526-532
f := strings.TrimSpace(e.File)
if f == "" { continue }
if seenPath[f] {
    problems = append(problems, fmt.Sprintf("%s: %s appears in more than one edit of this stage", tag, f))
}
seenPath[f] = true
```

The `feature-designer` prompt never states that constraint. Its plan rules cap
*quantities* — rule 4, verbatim: *"CAPS: at most 6 stages, 8 edits per stage, 24
edits total"* — and say nothing about uniqueness of `file` within a stage.
Measured 2026-07-26 against the live row:

```sql
SELECT CASE WHEN default_config::text ILIKE '%more than one edit%' THEN 'STATED'
            WHEN default_config::text ILIKE '%one edit per file%'  THEN 'stated (variant)'
            ELSE 'NOT STATED' END
FROM agent_definitions WHERE type='feature-designer'
  AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
-- NOT STATED
```

So a designer that splits one file's work into two readable edits — *"add the
helper"* and *"add the entry point"* — produces a plan that is coherent, within
every stated cap, and unpersistable. It cannot know.

## Live evidence

Run `fix_correlation_id 42665066-b492-4ed8-9ea3-d95fac57bc5a`, work item
`7b89fb35-f42c-45d1-b64d-214aff56d918` (`bugs_closed/077`'s colour-fixer
capability gap).

- `check_spec_approved` → `{"condition_met": true, ...}` — **the owner-approval
  gate passed**; this is not a rejected submission.
- The designer ran to completion and produced a design AND a proposal.
- `__step_error`:
  *"step persist_plan failed: failed to execute action diagnose_persist_fix_plan:
  staged plan failed validation: stage s1 edit 2:
  platform/orchestration/actions/discovery_checks/check_hardcoded_section_colors.go
  appears in more than one edit of this stage"*
- Workflow routed to `complete_refused`. **No `fix_plan` artifact was written**, so
  the design is only recoverable from `orchestration_states.collected_data->'design'`.

**The design that was lost was a good one**, which is what makes this worth
filing rather than shrugging at. Its own risk note:

> *"The plan deliberately answers none of the spec's three open design questions
> (bare #fff, semantic tints without existing palette entries, inline style=""
> attributes) — it only claims the two cases the evidence explicitly supports
> (exact matches to palette.background and palette.background_alt); a reviewer
> should reject any stage that widens further than that without an owner decision
> on those questions."*

Its s1 was: *add `derivePaletteColorMap`, add `replaceCSSColorsWithPalette`, leave
`replaceCSSColors` untouched so the tree still builds* — two edits, one file, one
stage. Exactly the shape the validator forbids and the prompt permits.

## This is the same class as `bugs_closed/077`

077 was a DETECTOR whose predicate was wider than its HANDLER's remit, filing work
the handler could never do. This is a PRODUCER whose output space is wider than its
CONSUMER accepts, emitting plans the persister can never store. Same shape, one
level up: **two ends of one contract, written at different times, with nothing
holding them in lockstep.** 016b §9's rule generalises directly — *"ask whether the
thing you are checking is the thing the other end promised to do"* — and the
remedy is the same: state the constraint where the producer can see it, or derive
one end from the other.

## Fix candidates, cheapest first

1. **Tell the designer.** Extend rule 4 in the `design` step's `prompt_template`:
   *"…and AT MOST ONE EDIT PER FILE PER STAGE — combine multiple changes to one
   file into a single edit whose sketch describes them all; a repeated path fails
   validation and the whole plan is discarded."* Config-only, live immediately, no
   image. Also worth adding to `reframe`/`repropose` if they carry their own copy
   of the rules — **check, do not assume**.
2. **Make the refusal recoverable.** `persist_plan` failing currently discards a
   completed design. Route the validation problem back into `repropose` (which
   exists) with the problem text, so the designer can merge the duplicate edits
   instead of the run dying. This is the durable fix; (1) only lowers the rate.
3. **Validate at the source.** Have the designer's own output schema check
   uniqueness before `persist_plan` sees it, so the failure names itself at the
   step that produced it.

Do **not** "fix" this by relaxing the validator: one edit per file per stage is
what makes a stage a single reviewable commit, and the caged implementer applies
edits per file.

## Candidate 1 IS APPLIED — migration `222`, 2026-07-26

> **This section previously said "WHY NOT APPLIED HERE", on the reasoning that the
> `feature-designer` row belongs to the feature-builder workstream and editing
> another workstream's live agent mid-flight is the collision CLAUDE.md warns
> about. That reasoning was sound and I changed the decision, so I am recording
> the change rather than quietly rewriting it.** What changed: the defect blocks
> the owner's standing instruction to drive a capability gap through the designer's
> gate, and it is systematic rather than flaky — the natural decomposition of that
> feature (add a helper, add an entry point, one file) reproduces it, so retrying
> without a change just spends another designer run to reach the same refusal.

`docs/agent_docs/sql_for_agents/222_feature_designer_one_edit_per_file_per_stage.sql`
— applied and verified. It takes a `snapshot_agent` first, extends rule 4 **in
place** with `replace()` on the exact sentence rather than rewriting the template
(so a concurrent edit elsewhere in the prompt survives), and **RAISEs if the
replace did not bite** — a silent no-op would otherwise leave the bug in place
while reporting success. `reframe` and `repropose` were checked and carry no copy
of the caps rule, so there is no second edit to keep in lockstep.

**VERIFIED on the failing branch, same day.** The identical submission (work item
`7b89fb35`) was re-fired after the migration as
`FEATURE_CORR c91bb061-250e-4a1a-819f-78c625733956`. It cleared `persist_plan` and
**wrote a `fix_plan` artifact** — the exact discriminating test this file specifies
below, and the thing that did not exist on the first run. The plan it persisted is
three stages:

```
s1  Derive a palette-based color map inside the existing color-matching machinery
s2  Flip the pinned remit test cases the widening is supposed to flip, and prove
    non-matches stay unmapped
s3  Wire the handler to fetch the site's palette into the same shared function
```

Note s2: the designer read the code_pointer warning that
`check_hardcoded_section_colors_test.go`'s `withinRemit: false` fixtures pin the
CURRENT remit and are the ones a widening must deliberately flip — and made that a
gated stage of its own rather than quietly editing the test to pass. Which is also
evidence for the other half of the lesson: the plan is only as good as the spec's
`code_pointers`, and those come from whoever files the gap.

**This is still only candidate 1, and it only lowers the rate.** Candidate 2 —
routing a `persist_plan` validation failure back into `repropose` instead of
discarding a completed design — remains the durable fix and is **not done**. The
bug stays OPEN for it. A designer that hits any *other* validator rule still loses
its design silently, and there are a dozen more rules in that validator.

**For the feature-builder workstream:** this is a live config change to your agent,
made from outside your lane. It is snapshotted and reversible (rollback in the
migration header). If you re-seed `feature-designer` from a `.sql`, fold the rule
into the seed or this fix is clobbered — the "config re-seed clobber" landmine.

## Re-verified 2026-07-27 — candidate 1 SURVIVED a fleet-wide agent re-seed

Checked by the bug-backlog triage sweep. This is the "config re-seed clobber"
landmine the section above warns the feature-builder workstream about, and it
came within a minute of firing, so it is worth recording that the check was run
rather than assumed.

**All 180 live `agent_definitions` rows were rewritten at ~15:10 UTC** — one
minute before the `v1.0.1174` pods came up at 15:11:15Z, i.e. a `deploy-agents`
seed sync riding with the roll:

```sql
SELECT count(*) FILTER (WHERE updated_at > NOW() - INTERVAL '1 hour') AS updated_last_hour,
       count(*) AS total
FROM agent_definitions WHERE deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
-- 180 | 180
```

**The rule survived it, and it survived in the right place** — not merely
somewhere in the row's JSON, but in the `design` step's own `prompt_template`,
which is what migration 222 edited:

```sql
SELECT (default_config->'workflow'->'steps'->'design'->'config'->>'prompt_template')
       ILIKE '%ONE EDIT PER FILE PER STAGE%' AS in_design_step
FROM agent_definitions WHERE type='feature-designer'
  AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
-- t
```

Note the difference between that query and the one in §"The mechanism", which
tests `default_config::text`. **A whole-row `::text ILIKE` cannot tell you the
rule is on the path the designer reads** — it would pass just as happily if the
string had landed in `reframe`, or in a comment, or in a step that never runs.
Use the path-qualified form when re-checking this after any re-seed.

`schema_migrations` still records
`222_feature_designer_one_edit_per_file_per_stage.sql`, applied 2026-07-26
21:36:46, `applied_by='record-only'`. **Beware the number:** `222` is used twice
in `sql_for_agents/` (this one and `222_news_components_carry_their_own_css.sql`,
applied 18:40 the same day) — resolve by filename, never by number.

## Why this stays OPEN

Candidate 2 — routing a `persist_plan` validation failure back into `repropose`
rather than discarding a completed design — is **not done**, and it is the
durable fix. Candidate 1 only lowers the rate for one rule; the validator
(`diagnose_persist_fix_plan_action.go`) has a dozen more, and a designer that
trips any of the others still loses its whole design silently, with
`orchestration_states.error` NULL. That is a Go change, so it is inert until a
roll — it did **not** ship in `v1.0.1174`.

## How to verify a fix

Re-fire the designer on the same work item and require BOTH:

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/0NN_TRIGGER_feature_designer_v1.sh \
  7b89fb35-f42c-45d1-b64d-214aff56d918
```

1. a `fix_plan` artifact EXISTS for the new correlation
   (`SELECT kind FROM diagnosis_artifacts WHERE correlation_id='<FEATURE_CORR>'`), and
2. the run does NOT end at `complete_refused`.

**Both, not either.** The run "completing" is not evidence — this one COMPLETED,
at `complete_refused`, with no artifact and no error column set. `error` was NULL;
the reason lived only in `collected_data->'__step_error'`.

## Landmine for whoever picks this up

`orchestration_states.error` was **empty** on a run that failed a step. The real
message is in `collected_data->>'__step_error'`, and `final_result` was NULL too.
A dashboard keyed on `error` reports this run as clean.

## References

- `platform/orchestration/actions/diagnose_persist_fix_plan_action.go:526-532` — the rule.
- `bugs_closed/077` — same class, one level down; its capability gap is the work item above.
- `016b` §9 — *"A 'did the fix work?' check must assert the HANDLER's remit…"* and
  *"A detector must PARTITION its population by the handler's remit…"*.

---

## CANDIDATE 2 BUILT — 2026-07-30

Committed to HEAD; **Go, so inert until the next chassis roll.** Config half is
`docs/agent_docs/sql_for_agents/272_feature_designer_plan_repair_loop.sql`,
dry-run clean and deliberately **not applied until the image is pod-verified**
(DB config is live immediately, Go is not).

Workstream docs: `docs024_key_docs_latest/bugfix_099_plan_refusal_recoverable/`
(PLAN, RUNBOOK, NOTES, README_where_we_are). Concept register: **FIX-057**.

### What it does

`diagnose_persist_fix_plan` gains two config keys, `repair_step` and
`max_repair_attempts` (default 1). **With `repair_step` unset the action behaves
exactly as before** — which is why the other two consumers of it
(`council-gate → persist_submission`, `fix-proposer → persist_plan`) need no change
and get no behaviour change.

With `repair_step` set, a **structural** validation failure:

1. persists the rejected plan verbatim + the exact problems as a durable artefact;
2. returns a RESULT (not an error) carrying `plan_valid: false`,
   `should_repair_plan: true`, `validation_problems_text`, `rejected_plan_json`;
3. is routed by a `conditional` step back to a repair prompt, up to
   `max_repair_attempts` times, then fails exactly as it does today.

**The gate is not lowered.** An invalid plan is still never persisted as a
`fix_plan` on either path — the original justification ("persisting a malformed plan
would hand F1.1b garbage to turn into a branch") is preserved in full. The change is
that *failing the step* and *not persisting* turned out to be two separate claims,
and only the second was load-bearing.

**Rule-agnostic, which is the point.** This file's own "Why this stays OPEN" section
says candidate 1 "only lowers the rate for one rule; the validator has a dozen more".
Candidate 2 needs no prompt edit for any of them, present or future.

### CORRECTION to candidate 2 as written above

> This file's candidate 2 says: *"Route the validation problem back into `repropose`
> (which exists) with the problem text."* **That is not implementable as written, and
> I started doing it before checking.** `persist_plan` runs **before any council**, so
> on a first-pass refusal `repropose`'s live prompt_template renders
> `{{.council_reviews.body}}`, `{{.check_results.results_text}}` and
> `{{.code_lookup_results.results_text}}` against nothing, and it frames a structural
> problem as a council objection ("A council reviewed your previous plan and asked
> for revision"). The refusal path the fix exists to serve is exactly the path where
> that prompt is malformed.
>
> Built a dedicated `repair_plan` step instead, whose prompt states plainly that the
> plan was *not* reviewed and *not* rejected on its merits, gives the validator's
> literal complaints, and asks for the same design with only those problems fixed.
> **What caught it:** reading `repropose`'s live `prompt_template` before wiring to
> it, rather than trusting the step's name. Whether the two loops should later
> converge is left open in FIX-057.

### Deliberately NOT done

- **Truncated / invalid JSON and the byte cap stay terminal.** A cut completion is a
  `max_tokens` fault, not a repairable plan (`bugs_open/012`, `138`). There is a test
  asserting truncated JSON does not enter the repair loop even with `repair_step` set.
- **`fix-proposer` is not opted in.** It has the same defect and the fix is one
  migration, but it belongs to another lane; routing that from outside is the
  collision this repo's rules warn about. Named in FIX-057's open question.
- **`council-gate` is not opted in**, and that is a positive choice: it is the gate
  this change is reviewed by.
- **The validator is not relaxed.** This file forbids it and is right.

### Two silent traps found building it (both now in `LANDMINES.md`)

1. **`diagnosis_artifacts.kind` is CHECK-constrained** to
   `bundle|iteration_note|fix_plan|council_report|escalation`. Giving the refusal
   note its own kind compiles and fails at **runtime**, on the failing branch. It
   reuses the allowed-but-unused `iteration_note` slot with
   `metadata->>'note_kind' = 'plan_validation_refusal'` — so **any reader of
   `iteration_note` must filter on that key.** No DDL on a shared table.
2. **A NULL `orchestration_id` never satisfies `= $2`**, so the run-scoped refusal
   count would return 0 for ever and 0 reads as "first attempt" — an unbounded loop
   with no error. Guarded by refusing terminally when the run id is absent.
   `diagnose_council_decide_action.go:514-517` has the same shape and no guard;
   **[UNMEASURED]** whether its run id is ever empty, so that is recorded as a shape
   to check, not filed as a defect.

### How to verify (unchanged from this file's own section, which is right)

Require **BOTH** a `fix_plan` artifact for the new correlation **AND** that the run
does not end at `complete_refused`. A completing run is not evidence — the original
failing run COMPLETED, at `complete_refused`, with `error` NULL. Commands in the
workstream RUNBOOK §"Verify the fix on the failing branch".

Read the refusals the loop recorded:

```sql
SELECT created_at, metadata->>'shape', metadata->>'problem_count', metadata->'problems'
  FROM diagnosis_artifacts
 WHERE kind='iteration_note' AND metadata->>'note_kind'='plan_validation_refusal'
 ORDER BY created_at DESC;
```

## COUNCIL ROUND 1 — REVISE, and three objections were right

Corr `f4a4628f-3b90-4054-a875-f2cf72b83e72`. 13 seats, 8 approve, 5 object,
`gated_by_truncation: false` (a genuine judgement, not a token overrun).

**CONFIRMED and fixed:**

- **`debug_historian`** — the rollback block in `272` was **wrong**. `snapshot_agent`
  has two overloads; the 2-arg form writes to `agent_definitions_backup`, **not** an
  `is_snapshot` row, and that table copies `created_at` verbatim from the source, so
  "the newest snapshot" ordered by `created_at` is a tie (measured: all three
  `feature-designer` backups share `2026-07-17 18:06:05`). Only `snapshot_taken_at`
  discriminates. **Migration `222` — candidate 1 of this same bug — carries the
  identical wrong instruction** and has not been edited (another lane's applied
  file); the trap is in `LANDMINES.md`.
- **`reuse_agent`** — the bounded-retry counter was a third hand-written copy.
  Extracted `countRunArtifacts` (`diagnose_artifact_count.go`) and switched all three
  call sites, `diagnose_council_decide`'s two included (pure refactor, its tests
  unchanged and passing).
- **`bug_historian`** — recoverable was not **visible**, and silence was this bug's
  original complaint. Refusals now also write `agent_error_log` with
  `error_code='FIX_PLAN_VALIDATION_REFUSED'`, on **both** outcomes, so the row's
  absence means no refusal rather than a quiet one.
- **`bug_historian`** — `fix-proposer` had no visible trigger to ever revisit.
  `273_fix_proposer_plan_repair_loop.sql` is written and **dry-run clean**, so proven
  applicable, and deliberately **NOT applied** — that agent belongs to another lane.

**Refuted by measurement (checks added anyway):** a root `ai_service` cannot shadow
the step block — `feature-designer` has none, *and* that behaviour was
`bugs_open/009` and is fixed (`resolveAIServiceConfig` is a per-key overlay where the
step wins; MDL-039 describes pre-fix behaviour). And an `ai_service.thinking` key
would be dead config — 0 live steps set it and `execute_llm_prompt` reads
`budget_tokens`.

## ⚠ CORRECTION — this file's "How to verify a fix" CANNOT prove candidate 2

The section above says to re-fire work item `7b89fb35` and require a `fix_plan`
artifact plus no `complete_refused`. **That was written for candidate 1, and
candidate 1 worked.** Its rule is still live in the design step:

```sql
SELECT (default_config->'workflow'->'steps'->'design'->'config'->>'prompt_template')
       ILIKE '%ONE EDIT PER FILE PER STAGE%' FROM agent_definitions
 WHERE type='feature-designer' AND deleted_at IS NULL
   AND COALESCE(is_snapshot,false)=false;   -- t, checked 2026-07-30
```

So that run now takes the **success** path. It satisfies both stated conditions while
the repair loop **never fires** — a false PASS, and exactly the "a PASS from a blind
check" shape. Candidate 2 must be proved by **inducing** a refusal, deliberately via a
*different* rule from the one candidate 1 taught, since being rule-agnostic is the
whole claim. Procedure (arm `max_edits=1`, which is repairable without losing scope;
require four conditions including that a refusal note exists and that the run reached
`repair_plan`) is in the workstream RUNBOOK §"Verify the fix on the failing branch".

## CORRECTION 2026-07-31 — "any validator rule becomes recoverable" was TOO STRONG

The candidate-2 write-up above says the fix is rule-agnostic and that "a designer that
hits any *other* validator rule" is covered. Reading the validator's own messages
before running a live induction, that is not quite true, and the overstatement is
mine:

| class | rules | repairable without losing scope? |
|---|---|---|
| **structural** | duplicate file in a stage, modify-before-add, create-then-delete, forward `depends_on`, empty goal, bad stage id, contradictory `artifact_role`, missing checklist entry, and the **per-stage** `max_edits` cap | **yes** — rearrange or split stages |
| **size caps** | `max_stages` — *"a build this broad needs splitting into more than one feature"*; `max_total_edits` — *"a build this broad needs splitting"* | **no** — they ask for less scope |

The repair prompt says *"do not drop scope"*, so against the size caps it is a
contradiction: the refusal burns its one repair round and goes terminal.

**This is not a correctness bug.** The loop is bounded, so the worst case is one extra
LLM call before the same terminal outcome the platform already had — and a plan that is
genuinely too big arguably *should* reach a human rather than be silently shrunk by a
model. But the accurate claim is **"the structural class becomes recoverable"**, not
"every rule does". Recorded as FIX-057's open question: whether the repair step should
detect a size-cap problem and escalate explicitly rather than attempt a repair it is
forbidden to make.

**What caught it:** reading each validator message before choosing which cap to arm for
the live test, instead of assuming any cap would do. The first cap I picked
(`max_edits=1`) turned out not to fire on this work item's plan at all; the second
(`max_stages=2`) fired on one run's plan shape and not the next one's. LLM plan shape
is not stable between runs, which is on its own a reason "re-fire and see whether it
passes" is a poor verification strategy for this bug.
