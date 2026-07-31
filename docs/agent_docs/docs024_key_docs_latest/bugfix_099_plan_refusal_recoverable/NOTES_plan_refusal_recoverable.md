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

### Post-submission: the two residual risks I named, now MEASURED

I listed both in the council submission as things a reviewer should check. Better to
answer them myself than ask the council to — measured while the round ran.

**Risk 3 — could `plan_persisted.plan_valid` resolve to something else?**
`resolveFieldValue` has five fallback strategies including a recursive key search
(`conditional_branch_action.go:397-412`), so a same-named key elsewhere could in
principle be found. Nothing else uses the name:

```bash
grep -rn "plan_valid" platform/ internal/ pkg/ --include=*.go \
  | grep -v diagnose_persist_fix_plan_action.go | grep -v diagnose_plan_refusal_test.go
# (no output)
```
```sql
SELECT type FROM agent_definitions WHERE default_config::text LIKE '%plan_valid%'
  AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
-- (0 rows)
```
Clean **as of 2026-07-30**. This is a name-uniqueness claim, so it expires the moment
someone else picks the name — it is not a structural guarantee.

**Risk 2 — does any consumer key off `persisted`?** No. Every reference to a
`plan_persisted` field, fleet-wide:

```sql
SELECT DISTINCT a.type || ' :: ' || m[1]
  FROM agent_definitions a,
       regexp_matches(a.default_config::text, 'plan_persisted\.[a-z_]+', 'g') AS m
 WHERE a.deleted_at IS NULL AND COALESCE(a.is_snapshot,false)=false AND a.is_active
 ORDER BY 1;
-- council-gate     :: plan_persisted.plan_json
-- council-gate     :: plan_persisted.summary
-- feature-designer :: plan_persisted.plan_json
-- fix-proposer     :: plan_persisted.plan_json
```

(Note the earlier attempt at this query failed with *"set-returning functions must
appear at top level of FROM"* — `regexp_matches(...,'g')` returns a set, so it goes
in `FROM` directly, not wrapped in `unnest()` in the select list.)

**This is the strongest evidence for the design decision I nearly got wrong.** All
three agents read `plan_persisted.plan_json`. Had the refusal returned the rejected
plan under that key — which was my first cut — **all three would have read a rejected
plan as a persisted one.** The rename to `rejected_plan_json` is now backed by a
measurement rather than by my reasoning about `repropose`'s prompt.

**One thing this surfaced that I had not considered:** `council-gate` also reads
`plan_persisted.summary`, which the refusal path does **not** return. That costs
nothing today because `council-gate` is not opted in and never takes the refusal
path. But it is a live consideration for the open question "should the other two
consumers be opted in" — opting in `council-gate` without adding `summary` to the
refusal result would render that field empty rather than error. Recorded here so the
next thread does not have to rediscover it.

### Caught after committing: I copied a DEPRECATED action alias

The `check_plan_valid` router was written with `"action": "conditional"`, copied from
the sibling `check_revise` step in the same agent — the "match the surrounding code"
instinct. The registry says otherwise:

```go
// registry.go:65-78
"conditional_branch": {Handler: ConditionalBranchAction, Category: "core", ...},
"conditional":        {Handler: ConditionalBranchAction, ..., Deprecated: true,
                       DeprecatedBy: "conditional_branch"},
```

Same handler, so behaviour is identical and the copied version would have worked —
which is exactly why nothing would ever have flagged it. Switched to
`conditional_branch`. **Surrounding code is a good default and not evidence**: every
existing `check_*` step in this agent uses the deprecated name, so consistency with
them meant inheriting their debt. Re-dry-ran, still clean.

Caught only because I went to confirm that a `conditional` step with no `next_step`
really does route via `next_step_override` — which it does (`check_revise` has no
`next_step` either, verified against the live row). Looking for one answer turned up
a different one.

### Roll state while waiting

Live chassis is `v1.0.1208`; another session has `IMAGE_TAG` staged at `v1.0.1209`
in the working tree but has not built it (no such local image). My commit
`8eebba0cf` IS an ancestor of HEAD, so whoever builds next carries the Go half
whether or not they know about it — the "your commit is a deploy" property.

Incidental confirmation of `bugs_open/153` sitting in plain sight in `docker images`:
`v1.0.1206` and `v1.0.1207` have the **same image id** (`64fe88a6a564`). A tag bump
did not rebuild. This is why the post-roll check has to be a pod-grep with a positive
control, not a version string.

## 2026-07-30 — council round 1: REVISE, and it was worth the round

Corr `f4a4628f-3b90-4054-a875-f2cf72b83e72`. **13 seats, 8 approve, 5 object,
0 unreadable, `gated_by_truncation: false`** — so a genuine judgement, not a token
overrun (FIX-055's flag earning its keep on its first look at my own submission).
Gated by `llm_reliability`.

**A trap I walked into reading the verdict, and caught by pattern-recognition rather
than by care.** My first query rendered `o->>'detail'` for each objection and printed
`(no objections)` for **all five** objecting seats — a clean, uniform, plausible
answer. The field is `problem`, not `detail`. This is the *identical* trap already in
`WRONG_CALLS.md` from 2026-07-29 (`ad791d6db`: "a wrong-depth JSON path returns a
clean uniform answer for all 17 seats rather than erroring"). I recognised the shape
only because a uniform "nothing here" across five independent seats is not what
disagreement looks like. **A repeat of an already-logged trap is the argument for
automating the check, not for trying harder** — logged again as a repeat.

### What each objection was, and what I did

**`llm_reliability` (HIGH, gating) — root `ai_service` shadows the step block.**
Cited MDL-039: "a step-level block is completely dead when a root block exists".
**REFUTED, twice over, by measurement:**
- `feature-designer` has **no root `ai_service`** — its only root key is `workflow`.
- That behaviour *was* `bugs_open/009` and is **fixed**. `resolveAIServiceConfig`
  (`ai_actions.go:40-96`) is now a per-key overlay, root → step → runtime, "later
  wins PER KEY", and its own comment says it "replaces first-found-wins, under which
  the ENTIRE step block was dead config whenever a root block existed".

So MDL-039 describes pre-fix behaviour. **But the seat's `missing` item was still
right** — I had asserted nothing either way. Added a migration assertion that
`repair_plan`'s `ai_service` carries all four keys, which stays useful under an
overlay (the real residual is a key *nobody* supplies, not a shadowed one) and keeps
holding if a root block is ever added.

**`llm_reliability` (MEDIUM) — set a `thinking` key; adaptive thinking eats
`max_tokens`.** The *concern* is real and live here. The *remedy* is not: **zero live
steps fleet-wide set `ai_service.thinking`, and `execute_llm_prompt` does not read
that key** — the extended-thinking knob it reads is `budget_tokens`
(`ai_actions.go:359-360`). Setting `thinking` would be dead config, exactly
`bugs_open/134`'s class. Answered on sizing instead: `repair_plan` uses the same
model and the same `max_tokens` (32000) as `design`, which demonstrably produces
staged plans today, and emits an artefact of the same shape — plus a migration
WARNING if the two ever diverge.

**`reuse_agent` (MEDIUM) — a second implementation of "bounded retry counted from
`diagnosis_artifacts`".** **CONFIRMED, and acted on.** Measured: no shared helper
existed — three hand-written copies of the query (two in
`diagnose_council_decide_action.go`, mine about to be the third). Extracted
`countRunArtifacts` (`diagnose_artifact_count.go`) and switched **all three** call
sites to it. Pure refactor for council-decide: same SQL, same `nullIfEmpty` argument,
same behaviour, its own tests unchanged and passing. The seat's stated risk was the
right one — "if the council-decide counter's scoping is later fixed, this new copy
will silently not get the fix" — and one function makes that impossible.

**`debug_historian` (MEDIUM) — `snapshot_agent` double-overload trap.**
**CONFIRMED, and my rollback block was WRONG.** The 2-arg form writes to
`agent_definitions_backup`, **not** an `is_snapshot` row in `agent_definitions`; and
that table copies `created_at` verbatim from the source, so "the newest snapshot"
ordered by `created_at` is a **tie** — measured, all three `feature-designer` backups
share source `created_at 2026-07-17 18:06:05`. Only `snapshot_taken_at` discriminates.
Rewrote the block with restore SQL verified to return the right row. **Migration 222
(candidate 1) carries the identical wrong instruction** — not edited, it is another
lane's applied file, but the trap is now in `LANDMINES.md` so either file's reader is
warned. This was the objection with real teeth: it would have bitten at the worst
possible moment, mid-rollback.

**`bug_historian` (MEDIUM) — `fix-proposer` left exposed.** Fair. Its `missing` named
the actual gap: not that the sibling was left, but that *nothing would ever remind
anyone*. Wrote `273_fix_proposer_plan_repair_loop.sql` — **dry-run clean, so proven
applicable, and deliberately NOT applied**, because `fix-proposer` belongs to another
lane and changing another lane's live agent from outside is the collision CLAUDE.md
warns about. The work is one command away instead of forgotten. Recorded in the bug
file too.

**`bug_historian` (LOW) — recoverable is not visible.** Right, and this was the best
observation of the round: 099's *original* complaint was that the loss was **silent**,
and my first cut made it survivable without making it findable — you had to know to
look at `metadata->>'note_kind'`. Added an `agent_error_log` row with a distinct code
`FIX_PLAN_VALIDATION_REFUSED`, the surface this platform already uses for exactly
this class. Written on **both** outcomes (routed-to-repair at `warning`, exhausted at
`error`), so the row's absence means "no refusal happened", never "it happened
quietly".

**`guardian` (MEDIUM) — shared-code blast radius.** No change needed; the seat
explicitly credited the mitigation and asked it be flagged as blast-radius rather
than framed as consumer-local. It is right that the `repairStep == ""` check is now
load-bearing for two agents' *unchanged* guarantee. Stated plainly in the resubmission
and re-verified `iteration_note` still has no Go reader outside my file.

**Round 1 cost me one resubmission and bought: a corrected rollback that would
otherwise have failed silently, a deduplicated counter, an observability surface, and
a sibling migration. Three of the five objections were things I could have measured
and hadn't.**

### My WRONG_CALLS append was swept into another lane's commit

Committed my round-1 record with three pathspecs; only two files changed. The
`WRONG_CALLS.md` entry was already in HEAD — carried in by `15c6cf317`
("prove(S6 in the cluster)"), an unrelated lane's commit, between my append and my
commit.

This is the hazard CLAUDE.md states outright: *"Committing per task stops YOU
sweeping up OTHERS' WIP; it cannot stop a session that still runs `git add -A` from
sweeping up YOURS."* Nothing is lost and forward-only still holds, so there is
nothing to do about it — recorded because the documented risk actually firing is
worth a data point, and because anyone auditing who wrote that entry will find the
wrong session on the commit.

Not logged in `WRONG_CALLS.md` itself: I did not make an error, and that file is for
claims of mine that turned out false.

## 2026-07-31 — the fix went live, and then another session's roll killed both my runs

### The Go half shipped without me rolling anything

The fleet rolled to **v1.0.1214** while I was writing docs. My commit `8eebba0cf` was
already an ancestor of HEAD, and `make build-*` builds from committed HEAD, so it
carried. This is the "your commit is a deploy" property working in my favour for once:
I never rolled, and I deliberately did not, because a roll kills in-flight councils.

Verified on **both replicas**, not inferred from the tag:

```
plan_validation_refusal (r1):        2
FIX_PLAN_VALIDATION_REFUSED (r2):    1
"staged plan failed validation":     0   <-- MY CONTROL, and it is ZERO
diagnose_persist_fix_plan (control): 11
```

### MISSTEP — I picked a positive control that my own refactor destroyed

The zero is not a failed deploy. `"staged plan failed validation"` no longer exists as
a contiguous literal: my refactor replaced it with `"%s failed validation: %s"` where
`%s` is `"staged plan"`, assembled at runtime. **With only that control I would have
read a successful deploy as a failed one** and gone hunting a build problem that did
not exist. The second control settled it in the same exec.

The rule is now in `LANDMINES.md` and `WRONG_CALLS.md`: **a positive control must be
invariant under your own diff** — take it from a file you did not touch, or use two.
It is a sibling of the already-recorded trap that a short literal never reaches rodata.

### 272 applied, and the corrected rollback proven

Applied after the pod-grep, in the documented order. Verified path-qualified:
`persist_plan -> check_plan_valid -> (review_editquality | repair_plan -> persist_plan)`,
cap 1. And the **corrected** restore query returns the right row:

```sql
SELECT snapshot_taken_at, snapshot_reason FROM agent_definitions_backup
 WHERE type='feature-designer' ORDER BY snapshot_taken_at DESC LIMIT 1;
-- 2026-07-31 14:57:03 | 272_feature_designer_plan_repair_loop.sql: pre-update
```

So the `debug_historian` objection's fix is demonstrated, not merely written.

### Choosing the induced test by MEASUREMENT changed it

I was about to arm `max_edits=1` (per-stage). Then I read the plan this work item
actually produces — correlation `c91bb061`, the one 099 itself cites:

```
s1: 1 edit(s)   s2: 1 edit(s)   s3: 1 edit(s)
```

**One edit per stage — that cap would never have tripped**, and the run would have
sailed through the success path looking like a pass. It has **3 stages**, so
`max_stages=2` trips deterministically *and* is repairable by merging two stages with
no scope loss (s1 colour machinery, s2 tests, s3 handler — different files, so merging
cannot trip the one-edit-per-file rule either). It is also a *different* rule from the
one candidate 1 taught, which is the rule-agnostic claim under test.

**That is the third time on this bug that an assumed verification would have returned a
false pass.** First 099's own stated procedure; then `max_edits=1`; and the deploy
control above. The common shape: a check that passes for a reason unrelated to the
thing being checked.

### MISSTEP — I armed a shared live agent before checking who else was running

Armed `max_stages=2` on the live `feature-designer` row, *then* checked for other
in-flight runs. Only mine was running, so no harm — but the order was wrong, and DB
config is live immediately. Same family as CLAUDE.md's "checking the pod does not
check the queue". The check is now the first step of the RUNBOOK's induction section,
with its own gotcha: `orchestration_states` has **no `agent_type` column**, so match on
a step name unique to the agent's `workflow_plan`.

### Then another session rolled to v1.0.1215 and killed both my runs

Pods restarted **15:00:20** and **15:00:42**. My council round 2 had reached
`review_tooling_provenance` at **14:59:29**; my designer run entered `design` at
**14:58:55**. Both died about a minute later, and both sat at `EXECUTING_STEP` with a
frozen `updated_at` — the documented mask (`64ae96cd2`, "EXECUTING_STEP hid it for an
hour"). Nothing errored. Nothing said anything.

**I did the one thing available to me — I did not roll — and it was not enough.** On a
shared tree the roll that kills your council need not be yours, and there is no
mechanism that reserves the fleet for the duration of a review. Worth stating plainly
because the existing landmine reads as advice to the roller ("check for in-flight
councils before you build"); the corollary for the *reviewee* is that a council round
is simply not durable across a roll, and a long one is a coin flip on a tree that
rolled four times in an evening on 07-27 and twice in an hour today.

Practical consequence adopted: the stall detector. My re-run monitor now flags
`updated_at` older than 420s as "likely killed by a roll" rather than waiting out the
full timeout, because the frozen-EXECUTING_STEP signature is indistinguishable from
slow progress until you look at the clock.

Re-fired both: designer `c5219a69` (orch `46ecdb25`), council on the same trail corr
`f4a4628f` (run orch `5d635e84`). DB config survived the roll, as expected — the
armed cap and migration 272 were both still in place, so only the runs were lost.

### The re-run PROVED the router live — on the valid branch

Trail for `46ecdb25` / corr `c5219a69`:

```
load_spec > check_spec_approved > load_schema_hint > design > persist_plan
  > check_plan_valid > review_editquality > review_bug_historian > …
```

`check_plan_valid` — the step migration 272 created — **fired, evaluated
`plan_persisted.plan_valid == true`, and routed on to `review_editquality`.** So the
new step graph is live and the happy path is intact: the router resolves the field,
the conditional_branch action works, and nothing regressed. That is worth having, but
it is not the failing branch.

**The induced refusal did not trip**, because the designer produced a **2**-stage plan
this time where the earlier one had 3 (`stage_count: 2, edit_count: 3`), and my cap was
`max_stages=2`. LLM output shape is not stable between runs — which is itself the
answer to why "re-fire and see" is a poor verification strategy for this bug.

### A real finding about my own repair design: cap violations are not repairable

Reading the validator's exact messages before arming a tighter cap, they split into two
classes:

| rule | message | repairable without losing scope? |
|---|---|---|
| duplicate file in a stage, modify-before-add, create-then-delete, forward `depends_on`, empty goal, bad stage id, contradictory `artifact_role`, missing checklist entry | states the structural defect | **yes** — rearrange |
| `max_edits` (per stage) | `"stage N: X edits exceeds the per-stage cap Y"` | **yes** — split into more stages |
| `max_stages` | `"N stages exceeds cap M — a build this broad needs splitting into more than one feature"` | **no** — it explicitly asks for less scope |
| `max_total_edits` | `"N edits in total exceeds cap M — a build this broad needs splitting"` | **no** — same |

My repair prompt says *"Do not drop scope"*. Against the bottom two rows that is a
**contradiction**: the validator asks for a smaller plan and the repair prompt forbids
one. Such a refusal will burn its one repair round and then go terminal.

**That is not a correctness bug** — the loop is bounded, and it ends exactly where it
ends today, so the worst case is one extra LLM call before the same terminal outcome
the platform already had. And arguably a plan that is genuinely too big *should* reach
a human rather than be silently shrunk by a model. But it means the repair loop helps
the **structural** class and not the **size** class, which is a narrower claim than
"any validator rule becomes recoverable". Recorded here, in the bug file, and as
FIX-057's open question rather than quietly left as an unstated limit.

Consequence for the live test: `max_stages=1` would fire deterministically but proves
only the loop and its bound, never a successful repair. `max_edits=1` is the cap whose
message the designer can actually satisfy — split one 2-edit stage into two stages,
all edits kept, and `max_stages` 6 leaves room. That is the induction to use, and it
fires whenever any stage carries ≥2 edits (two of the last three plans did; this run's
s1 had 2 edits across two DISTINCT files, so splitting cannot trip the
one-edit-per-file rule either).

## 2026-07-31 — THE REPAIR BRANCH FIRED LIVE, and then taught me something

Run 3, corr `54e57db8`, orch `2a1a4e84`, with `max_edits=1` armed (`max_stages`
restored to 6 first). Trail:

```
load_spec > check_spec_approved > load_schema_hint > design > persist_plan
  > check_plan_valid > repair_plan
```

**Every stage of the mechanism worked, live:**

- the refusal note was written, naming exactly the induced problem:
  `["stage s1: 2 edits exceeds the per-stage cap 1"]`, `shape: staged plan`,
  `problem_count: 1`;
- its `body` holds the **full 10,488-byte, 2-stage design** — the plan that, before
  this fix, would have been destroyed with no trace;
- the operator record was written:
  `Fix plan refused (staged plan): 1 structural problem(s), attempt 1 of 1 — routed to
  repair`, severity `warning`, `repair_attempt 1`, `exhausted false`;
- `check_plan_valid` evaluated false and routed to `repair_plan`.

### …and then the run died anyway, for a reason I had ruled out of scope

Final state `complete_refused`, **0 `fix_plan` artifacts**, and only **one** refusal
note — so `persist_plan` was never reached a second time as a *structural* refusal.
`error` was NULL, as this bug's own landmine says it would be. The reason was in
`__step_error`:

```
step persist_plan failed: … plan does not match the staged-plan schema:
json: cannot unmarshal array into Go struct field fixPlanEdit.stages.edits.file
of type string
```

The repair **ran and produced a plan**. Told to split a 2-edit stage, the model
emitted `"file": [...]` — an array where the schema wants a string. That hit the
schema-mismatch branch, which I had deliberately made terminal.

**So the loop preserved the design, recorded it, handed it back — and then lost the
run anyway, having spent the repair budget.** The mechanism was right; my scope
boundary was wrong.

### The scope boundary was wrong, and the evidence is the first live round

I had lumped *schema mismatch* in with *truncation*, on the reasoning that both are
"malformed output". They are not the same failure:

| | JSON valid? | cause | retry sane? |
|---|---|---|---|
| truncation (`!json.Valid`) | **no** | hit `max_tokens` mid-token | **no** — a token-budget fault; retrying is the `bugs_open/012`/`138` trap |
| schema mismatch | **yes** | model used the wrong shape for a field | **yes** — mechanical, and the unmarshal error names the offending field |

Widened accordingly: well-formed-but-wrong-shape now routes to repair with the
unmarshal error as the problem text (so the prompt is told *which field*), while
truncation stays terminal, unchanged. The two tests are deliberately a pair —
`TruncatedJSONStaysTerminal` and `SchemaMismatchIsRepairable`, same action, same
config, opposite directions — plus one asserting the widening is invisible when
`repair_step` is unset.

**This is the value of insisting on a live failing-branch test.** All 15 unit tests
passed before this run; the boundary they encoded was my own assumption, so no test
could have caught it. It took one real repair round to show that the single most
likely repair failure was the one class I had excluded.

### Also caught: the operator row had an empty `orchestration_id`

The `agent_error_log` row wrote with `orchestration_id` blank — the correlation was in
`context`, but the column every other query joins on was empty. Now populated.
(Schema note for the runbook: that table's timestamp is **`occurred_at`**, not
`created_at`; querying `created_at` errors rather than returning nothing, which is at
least honest.)

### Status

The widening and the `orchestration_id` fix are **committed but NOT yet live** — Go,
so inert until the next roll. The *core* of candidate 2 IS live and proven on
`v1.0.1215`: a structural refusal now preserves the design, records it twice, and
routes it to repair instead of discarding it silently. What has **not** yet been
observed live is a repair round that *completes successfully*, and that is what the
next roll plus one more induced run should show. **The bug does not close until then**
— "the routing works" is not the same claim as "a design was saved".
