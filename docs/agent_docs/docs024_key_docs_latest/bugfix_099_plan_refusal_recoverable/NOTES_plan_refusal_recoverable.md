# NOTES — bugs_open/099 candidate 2 (append-only, newest at the bottom)

## 2026-07-30 — session 1

### Picking the bug

Asked to find the next `bugs_open` case no other thread is working. `who-owns.py`
is too conservative to answer this on its own: it printed **"VERDICT: OWNED or
recently active"** for every candidate I tried, including `115`, for which it also
printed "likely OWNING workstream(s): (none identified)". It counts any *mention*
inside an active workstream directory, and on a tree with 40 active lanes almost
every bug is mentioned somewhere.

What actually discriminated: **last commit touching the bug FILE**, per file:

```bash
for f in bugs_open/*.md; do
  echo "$(git log -1 --format='%ad' --date=format:'%m-%d %H:%M' -- "$f") | $(basename $f)"
done | sort
```

That gave a clean split — a cluster last touched 07-27/28 and everything from 07-29
onward. Cross-checked the quiet ones against `site_work_items` and against the day's
60+ commits. Picked `099`, whose own file says candidate 2 "remains the durable fix
and is **not done**".

### The bug file's candidate 2 is not implementable as written — corrected

099 says: *"Route the validation problem back into `repropose` (which exists) with
the problem text."* I started to do exactly that, then read `repropose`'s live
prompt_template. It renders:

- `{{.council_reviews.body}}` — "The council's reviews — EVERY seat that voted this round"
- `{{.check_results.results_text}}`
- `{{.code_lookup_results.results_text}}`

`persist_plan` runs **before any council**. So on a first-pass validation refusal
those three render against nothing, and the prompt tells the model a council asked
for revision when no council has seen it. Reusing `repropose` would have produced a
malformed prompt on the exact path the fix exists to serve.

Built a dedicated `repair_plan` step instead. Recorded as a correction in the bug
file, not silently designed around.

### `plan_json` was the trap I nearly walked into

First cut of the refusal result returned the rejected plan as `plan_json` — the same
key the success path uses. That would have been wrong in a way tests would not have
caught: `repropose` renders `{{.plan_persisted.plan_json}}` as *"Your previous
plan"*, so a **rejected** plan would read downstream as a persisted one. Renamed to
`rejected_plan_json`, and there is now an explicit test asserting `plan_json` is
**absent** on the refusal path.

### MISSTEP — I wrote a template field that does not exist

The `repair_plan` prompt's context section was written as `{{.spec_row.body}}`.
There is no `body` field. The design step's own prompt uses
`{{.spec_row.work_item_id}}`, `{{.spec_row.summary}}`, `{{.spec_row.spec_text}}`:

```sql
SELECT unnest(regexp_matches(
  default_config->'workflow'->'steps'->'design'->'config'->>'prompt_template',
  '\{\{\.spec_row[^}]*\}\}', 'g'))
FROM agent_definitions WHERE type='feature-designer'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- {{.spec_row.work_item_id}} / {{.spec_row.summary}} / {{.spec_row.spec_text}}
```

**A wrong template path renders an empty string and says nothing.** The repair step
would have run, produced a plan with no spec context, and looked fine. Caught only
because I went to check the field name before applying rather than after. Logged in
`WRONG_CALLS.md`.

### Two silent traps found in the process, both now landmines

1. **`diagnosis_artifacts.kind` is CHECK-constrained** to
   `bundle|iteration_note|fix_plan|council_report|escalation`. My first design gave
   the refusal note its own kind. That compiles and fails at runtime, on the failing
   branch — the branch nobody exercises before shipping. Reused the allowed-but-unused
   `iteration_note` slot with a `metadata->>'note_kind'` discriminator instead, so no
   DDL on a shared table.
2. **A NULL `orchestration_id` never satisfies `= $2`.** The run-scoped count that
   bounds the repair loop would return 0 every time, and 0 reads as "first attempt"
   — an unbounded loop with no error and no warning. Guarded by refusing terminally
   when the run id is absent.

   `diagnose_council_decide_action.go:514-517` has the same shape and no guard.
   **[UNMEASURED]** whether its `OrchestrationID` is ever actually empty — so this is
   a shape recorded in `LANDMINES.md`, **not** a bug I have filed. I have not
   measured it and am not asserting it.

### Tests

11, all passing. The four end-to-end ones matter most: the six helper tests hand
`problems` in, so they cannot prove the action **reaches** the refusal path. The
invalid fixture is built by mutating the shared `goodStagedPlan()` and then asserts
`validateStagedPlan` returns **exactly one** problem, the duplicate-file one — a
hand-rolled "invalid" plan can be invalid for reasons the test never intended, and
then the e2e tests would pass because the plan was valid.

Also asserted: the terminal path **does not touch the DB** (sqlmock with no
registered expectations), which is the real contract for the two consumers that are
not opted in.

### Verification of the config half

Dry-ran migration `272` by piping the file with `COMMIT` → `ROLLBACK`. All four
UPDATEs bit, the verification `DO` block's NOTICE fired, transaction rolled back.
The guard that refuses if `persist_plan.next_step` is not `review_editquality` also
proved the precondition still held at that moment.

**Not applied yet, deliberately.** DB config is live immediately and Go is inert
until a roll, so `272` goes on after the image is pod-verified. Applied early it is
inert-but-harmless (`check_plan_valid` is simply never reached, because the step
still fails), but "harmless" is not "useful" and the ordering is the documented one.
