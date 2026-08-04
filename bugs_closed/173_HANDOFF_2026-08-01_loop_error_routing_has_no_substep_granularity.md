# 173 — a loop's error routing is all-or-nothing: one substep cannot be made recoverable without making every substep recoverable

**Filed 2026-08-01** by the `bugs_open/165` sites B+C lane, **at the direction of
the council gate's `architecture` and `constitution` seats** (corr
`c69e935a-7134-45c1-81c3-2f1da7831827`, round 2). Status: **OPEN, UNOWNED.**
Latent — nothing is failing today because of it; it is a missing degree of freedom
that pushed a bug fix into a workaround, and the seats caught the workaround.

> **VERIFICATION STATEMENT, per the owner ruling of 2026-07-31.** That ruling says
> a `bugs_open/` file asserting a cross-cutting or structural cause is not filed
> until it has been through `090_TRIGGER_needs_diagnosis_v1.sh`, or the filing
> session states plainly why it substituted equivalent first-hand verification.
> **I substituted first-hand verification, and here is exactly what it consisted
> of**, so the substitution can be judged rather than taken on trust:
>
> - I read the whole routing path in the running source, not from memory or grep
>   hits: `loop_actions.go:66` (reads `continue_on_error` from the loop step's
>   config), `loop_expansion_handler.go:39` and `:157` (propagates one loop-level
>   value onto EVERY injected iteration step), `loop_error_handler.go:71-87`
>   (`shouldContinueLoopOnError` reads it back off the iteration step),
>   `coordinator.go:907` and `:3350-3363` (`routeToErrorStepOrFail` → `failWorkflow`).
> - I confirmed the live config shape by query, not by assumption: the loop that
>   motivated this (`multipage-website-builder.workflow.steps.generate_pages_loop`)
>   has `continue_on_error` unset, and its `substeps.extract_links` has no
>   `error_step`.
> - **What makes this a safe substitution:** the claim is a *closed* one about a
>   single code path I have read end to end, not a hunt for a cause that might live
>   somewhere else. There is no "the cause is not where the symptom is" risk here
>   because there is no symptom — the finding is that a knob does not exist.
> - **What would still be worth the loop, and I have NOT done:** whether any live
>   workflow is *currently* mis-shaped because of this (i.e. loops that set
>   `continue_on_error: true` and thereby silently swallow substep failures they
>   should not). That is a fleet-wide question and is listed under "What is not
>   established" below rather than asserted.

## The defect

`continue_on_error` is a **loop-level** flag. `LoopAction` reads it once from the
loop step's config (`loop_actions.go:66`), passes it through the expansion
(`loop_expansion_handler.go:39`), and stamps the **same value onto every injected
iteration step** (`:157`). `shouldContinueLoopOnError` then reads it back per
iteration step (`loop_error_handler.go:83`).

So the only two states a loop can be in are:

| `continue_on_error` | effect |
|---|---|
| unset / false | **any** substep error fails the entire orchestration (`routeToErrorStepOrFail` finds no `error_step` and falls through to `failWorkflow`) |
| true | **every** substep error is swallowed and the iteration is skipped |

There is no way to say "a failure in *this* substep should skip the item, but a
failure in *that* one should still fail the build". `error_step` exists at step
level and loop expansion does carry it (`loop_expansion_handler.go:480` prefixes
`error_step` among the step-reference fields), so a per-substep `error_step` is
routable — but it requires naming a real recovery step, which is a different thing
from "tolerate this one and carry on".

## Why it matters — the concrete case that produced this file

`bugs_open/165` site C put a completeness floor on `extract_and_sync_links`. That
action is reachable from exactly one live agent and it is **nested**:
`multipage-website-builder` → `workflow.steps.generate_pages_loop` →
`config.substeps.extract_links`. The loop carries no `continue_on_error`.

Consequence: **a refusal on one page's link extraction fails the entire site
build.** For a table that is regeneratable, feeds no rendered output, and is
currently empty fleet-wide, that is wildly disproportionate.

The fix lane's round-2 response was to make the action never error — return a
`(detail, bool)` and skip the sync. **Four council seats ruled that backwards and
they were right**, which is the useful part of this file:

- `constitution` (HIGH): "the plan's own rationale identifies the real cause as the
  workflow's missing `continue_on_error`/`error_step` … then works around that gap
  by making the action silently skip (never error) instead of fixing the workflow's
  error-routing config. This is a fix whose rationale names the mechanism it steps
  around rather than repairs it."
- `architecture` (medium): "the action launders its own failure into a soft
  `maySync=false` skip. That is a workaround built because the core routing
  couldn't carry the load, not a property of `link_registry` itself." And: it
  "establishes a reusable pattern … with no shared vocabulary — the next nested
  substep that hits the same disproportionate-failure problem will invent its own
  ad hoc bool/sentinel rather than the engine offering per-substep
  `continue_on_error`."
- `bug_historian` (HIGH): the skip turned a refusal into a success-shaped return
  whose only signal is a queue.
- `guardian` (medium): a shared-mechanism contract change on the strength of a
  single-caller claim.

The action was reverted to the uniform error contract. **The gap is real and is
this file.**

## Why the obvious config change is NOT the answer either

Setting `continue_on_error: true` on `generate_pages_loop` would fix the
disproportion — and would change failure semantics for **every** substep in the
page-generation loop, turning a page that genuinely failed to build into a
silently skipped one. That is the same silent-drop class this platform keeps
finding, moved rather than removed. It is a workflow-owner's decision and it
should not be a rider on someone's bug fix, which is precisely why it is filed
here instead of done.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Per-substep `continue_on_error`.** Let a substep declare its own tolerance,
   defaulting to the loop's value. Smallest change that removes the false choice:
   `loop_expansion_handler.go:157` already writes the flag onto each injected step,
   so the plumbing is one lookup — read `substep.Config["continue_on_error"]` first
   and fall back to the loop's. Closes the door because the shape "one tolerant
   substep in a strict loop" becomes expressible rather than requiring an action to
   lie about its own outcome.
2. **A named severity on the substep** (`on_error: fail | skip_item | continue`)
   instead of a bool. More expressive, more vocabulary to agree, and it is a shared
   mechanism change — so it wants an RFC rather than a bug patch.
3. **Leave it, and require a recovery step.** Per-substep `error_step` already
   works; document it as the answer and accept that every tolerant substep needs a
   real step to route to. Cheapest, but it is the option that leaves the next
   thread inventing a bool.

## What is NOT established, marked so nobody inherits it as a finding

- **Whether any live workflow is currently mis-shaped by this.** Loops that DO set
  `continue_on_error: true` are, by this same all-or-nothing property, swallowing
  every substep failure — including ones that should fail loudly. **[UNMEASURED]**
  I did not survey them. The query to start from:
  ```sql
  SELECT a.type, s.key, s.value->'config'->>'continue_on_error' AS coe,
         jsonb_object_keys(s.value->'config'->'substeps') AS substep
  FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
  WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
    AND s.value->'config' ? 'substeps';
  ```
- **Whether per-substep granularity has been raised in a prior council round.** The
  `architecture` seat flagged that it could not confirm this from its tier and
  asked for it to be checked. I have not checked it either. **[UNVERIFIED]**
- Whether candidate 1 interacts with the retry/`fallback_step` fields that
  `loop_expansion_handler.go:480` prefixes alongside `error_step`. Not looked at.

## How to verify a fix

The mechanism is inert until induced, like every guard in this family, so a green
build proves nothing. Induce it: give a loop two substeps, make the *tolerant* one
fail and confirm the iteration is skipped while the orchestration continues; then
make the *strict* one fail in the same loop and confirm the orchestration FAILS.
Both branches, or the flag is untested in the direction that matters.

## Provenance

- Council corr `c69e935a-7134-45c1-81c3-2f1da7831827`, round 2 — `constitution`
  (HIGH, edit 4), `architecture` (medium ×2, edit 4), `bug_historian` (HIGH,
  edit 4), `guardian` (medium, edit 4).
- The case that produced it: `bugs_open/165` site C,
  `platform/orchestration/actions/link_registry_prune_floor.go`.
- Code: `loop_actions.go:66`, `loop_expansion_handler.go:39,157,480`,
  `loop_error_handler.go:71-87`, `coordinator.go:907,3350-3363`.

---

## CONTRIBUTION 2026-08-01, from the `165` site A lane — the fleet-wide question, answered; and the blast radius is four loops, not one

Two things this file listed as *not* established. Both are measurable, so here
they are measured. **This is a contribution into the case, not a competing fix —
nothing in the code was touched.**

### 1. The undone question: does any live loop swallow substep failures it should not?

> *"whether any live workflow is currently mis-shaped because of this (i.e. loops
> that set `continue_on_error: true` and thereby silently swallow substep failures
> they should not). That is a fleet-wide question…"*

**Answer: no live loop is currently mis-shaped in that direction.** Census of every
loop step on every active, non-snapshot agent (2026-08-01):

| `continue_on_error` | loops | agents |
|---|---|---|
| `(unset)` | 10 | 8 |
| `true` | 9 | 9 |
| `false` | 1 | 1 |

All **nine** `true` loops are fan-out / dispatch loops, and what they run is the
justification for the flag:

| agent | loop | sub_workflow actions |
|---|---|---|
| `area-sweep-orchestrator` | `sweep_loop` | `call_agent` |
| `business-intel` | `process_batch` | `companies_house_fetch`, `companies_house_search`, `store_ch_enrichment`, `conditional` |
| `component-quality-auditor` | `create_regen_items` | `create_work_item`, `conditional` |
| `content-feed-trigger` | `process_sites` | `call_agent`, `spawn_agent` |
| `internal-linker` | `create_items_loop` | `create_work_item` |
| `model-directory-trigger` | `process_sites` | `call_agent`, `spawn_agent` |
| `thunder-training-monitor` | `monitor_loop` | `call_agent`, `spawn_agent` |
| `tool-auditor` | `create_items_loop` | `create_work_item`, `conditional` |
| `vet-batch-processor` | `process_batch` | `call_agent` |

**Not one wraps a destructive reconciliation**, and specifically: *zero* of the
nine contain `save_page_sections`, `populate_nav_tables`, `sync_page_links` or
`index_code_symbols` — the four actions that now carry a completeness floor. So no
floor refusal is currently being swallowed by a loop anywhere on the fleet. For
"spawn a job per site, one site's failure must not stop the batch", `true` is the
right setting, which is what these all are.

**Measurement gotcha, recorded because it cost a wrong first answer:** a loop's
body lives in `config.sub_workflow.steps` (18 of 20 loops), **not**
`config.substeps` (2 of 20). Counting substeps off the wrong key returns `0` for
nine loops and reads exactly like "these loops are empty, nothing to worry about".

### 2. The blast radius is FOUR loops across four agents, not just `generate_pages_loop`

This file's motivating case was site C inside
`multipage-website-builder.generate_pages_loop`. **Site A has the identical shape
in three more loops**, all `continue_on_error` UNSET with no `error_step` on the
`save_sections` substep:

| agent | loop | substep that can now refuse |
|---|---|---|
| `pageflow-builder` | `build_pages_loop` | `save_sections` (`save_page_sections`) |
| `page-rebuild` | `build_pages_loop` | `save_sections` |
| `site-work-orchestrator` | `build_items_loop` | `save_sections` |

So a completeness refusal on **one** page fails the **entire multi-page build** in
each of these — the same disproportion, at the same severity, for a build that may
have already produced a dozen good pages. That does not change 173's diagnosis; it
triples the number of live loops that want the missing degree of freedom, and it
means fixing 173 buys site A as well as site C.

Not fixed here for the reason 173 itself gives: the missing knob is the cause, and
working around it per-call-site is what the `architecture` and `constitution` seats
already rejected once.

### What is still not established

Whether the three loops above have ever actually aborted a build this way. The
floors are new (A live on `v1.0.1223` 2026-07-31; B and C not yet live), so the
answer today is almost certainly no — but it is unmeasured, and
`orchestration_states` retention (~2 days) will make it unanswerable in arrears.

---

## FIX 2026-08-04 — candidate 1 BUILT and committed (`2e497e846`). **STILL OPEN: inert until a chassis roll.**

Picked up unowned, per this file's own header. Lane docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_173_substep_error_tolerance/`.

**Do not close this on the commit.** The bar in CLAUDE.md is *fixed AND live*, and the live
chassis is `v1.0.1248`, which predates the fix. Measured at the pod the moment after
committing, with both controls:

```
strings /app/agent-chassis | grep -c 'resolveSubstepContinueOnError'                    -> 0   (this change; not live)
strings /app/agent-chassis | grep -c 'continue_on_error is true for this loop iteration step' -> 1   (positive control; the probe reads the binary)
```

### What was built — candidate 1, the shape this file and the seats pointed at

One production file, `platform/orchestration/loop_expansion_handler.go`. The unconditional
stamp at line 157 becomes a resolution:

| substep's own `config.continue_on_error` | resolves to |
|---|---|
| bool `true` | `true` — tolerant substep inside a strict loop |
| bool `false` | `false` — **strict substep inside a tolerant loop** |
| absent | the loop's value — byte-identical to today |
| present, not a bool | the loop's value, **and a WARN naming the substep** |

Candidates 2 and 3 were **not** taken, for this file's own reasons: 2 is a new shared
vocabulary and by the owner ruling of 2026-07-29 wants an RFC rather than a bug patch; 3 is
the option that "leaves the next thread inventing a bool".

### The finding this file does not contain — the key was already writable, and inert

Line 104 deep-clones the substep's config into the injected step, so a substep's own
`continue_on_error` **was already present** and line 157 then **overwrote** it. And
`continue_on_error` is in `datahelpers.frameworkStepConfigKeys`, so `config-key-audit` called
it legitimate framework vocabulary and never flagged it.

**So an author could write the obvious thing, get no error, no warning and no audit finding —
and no effect.** Any substep config predating this fix that declares the key has been inert
since it was written. Registered as a landmine (footprint `loop_expansion_handler.go`) and as
concept **WFA-008**, in the same commit as the code.

### The direction that is easy to get wrong, recorded for the reviewer

`if v, ok := cfg["continue_on_error"].(bool); ok && v` is **wrong**: folding the type
assertion into the truth test reads a declared `false` as *no declaration*, silently
destroying the strict-substep-in-a-tolerant-loop direction. Presence and truth are tested
separately, and the test suite covers both.

**Mutation-proven rather than merely green** (`platform/orchestration/substep_continue_on_error_test.go`,
5 functions / 7 subtests): reverting the production line to the unconditional stamp fails
exactly the three override tests (tolerant-in-strict, strict-in-tolerant,
per-substep-not-per-loop) while the two inheritance tests correctly still pass.

### Answering this file's two "NOT established" items

1. **"Whether any live workflow is currently mis-shaped by this."** Re-measured today, and
   the answer is unchanged from the 08-01 contribution: no loop is mis-shaped in the swallow
   direction. **But the census has drifted and the figures here were stale** — it recorded
   20 loops (10 unset / 9 true / 1 false); today it is **18** (8 unset / 9 true / 1 false).
2. **The blast-radius question this fix needed:** **0 of 79 substeps** across those 18 loop
   steps declare `config.continue_on_error`. Positive control, same CTE grouped rather than
   filtered: 18 loops / 79 substeps — so the predicate reaches substep configs and a non-zero
   answer was reachable. **The change is therefore inert fleet-wide on the day it rolls.**

Still **[UNVERIFIED]**, as this file listed: whether candidate 1 interacts with the
retry/`fallback_step` fields that `loop_expansion_handler.go:480` prefixes alongside
`error_step`. I did not look either, and it is flagged in the council submission as such.

### CORRECTION — this file's motivating loop no longer exists

§"Why it matters — the concrete case that produced this file" is **historical**.
`multipage-website-builder` is now `is_active=f` and deleted (consistent with `7a15c3a47`
"retire(agents): three unused builders are out"), so `generate_pages_loop → extract_links`
cannot recur. **Quote the three live consumers instead**, all still
`continue_on_error` **(unset)** with a substep that can refuse — measured 2026-08-04:

| agent | loop | substep | action |
|---|---|---|---|
| `pageflow-builder` | `build_pages_loop` | `save_sections` | `save_page_sections` |
| `page-rebuild` | `build_pages_loop` | `save_sections` | `save_page_sections` |
| `site-work-orchestrator` | `build_items_loop` | `save_sections` | `save_page_sections` |

The reasoning in this file is unaffected — it never depended on that agent, only illustrated
itself with it.

### Council

Submitted 2026-08-04, correlation `549e25fb-acc1-4806-a2a7-95bf73cca806`; committed with
`Council-Submitted:` because no verdict had been read. The submission argues the scope
question explicitly (owner ruling 2026-07-29 §1: the RFC trigger is a change to the shared
mechanism's **guarantees**, and this adds an opt-in capability reachable by nothing until a
config names it) and names the three consumers rather than merely counting them (§3).

### WHAT IS OWED before this can be closed

1. **The roll.** `make release` is whole-fleet and owner-run, and a roll kills in-flight
   council runs — including this change's own, which was in flight at commit time. Not done
   unilaterally.
2. **The live induction, which is this file's own bar and which no build can satisfy:** give a
   loop two substeps, make the **tolerant** one fail and confirm the iteration is skipped
   while the orchestration continues; then make the **strict** one fail in the same loop and
   confirm the orchestration FAILS. Both branches. The commands are in the lane's
   `RUNBOOK_substep_error_tolerance.md` §R5–R6.
3. **Read the verdict** and act on a REVISE/REJECTED — the code is already on the shared
   branch, so a resubmission is a forward fix, never an amend.

### Council: APPROVED round 1 (`549e25fb-acc1-4806-a2a7-95bf73cca806`)

8 seats approving, 3 advisory objections, **none high-severity**. All three answered with
checks rather than nods — full working in the lane's NOTES. In brief:

- **`guardian` (medium), the `fallback_step`/retry interaction this file listed as unlooked-at:
  there is no interaction, because there is no mechanism.** `fallback_step` and `retry_step`
  occur exactly twice in the whole Go codebase, both in *name-prefixing* lists
  (`loop_expansion_handler.go:528`, `coordinator.go:4243`); no field on `models.Step` and no
  routing read. **0 live definitions declare either key [MEASURED 2026-08-04].** This file's
  third "not established" item can be struck.
- **`bug_historian` (medium), does a tolerance-skip leave a durable trace?** Yes, three:
  `skipToNextLoopIteration` persists `{loop}_iter_{N}_error` (`skipped:true`, failing step,
  timestamp) plus a `{loop}_error_count`; it logs a `Warn` naming the failed step and iteration;
  and `LoopCompleteAction` stamps `status:"error"` per item in the aggregate, distinct from
  `status:"missing"`. **Residual, stated honestly:** no work item is raised and
  `orchestration_states` is retention-clocked, so the trace expires with the row; whether a
  downstream consumer acts on the aggregate's status is per-workflow and **[UNMEASURED]**.
- **`bug_historian` (medium), file the sibling instead of deferring it in prose → `bugs_open/193`.**
  `loop_actions.go:66`'s loop-level read of this same key silently ignores a non-bool, and after
  today's fix its substep-level twin warns — so the mechanism is now loud on one side and silent
  on the other. Latent: all 10 declaring loops declare `boolean` [MEASURED 2026-08-04].

**For the owner, not resolved here:** `reuse_agent` and `architecture` both noted that owner
ruling 2026-07-29 §1 was **self-applied by the change's author** — I argued my own change does
not cross the RFC threshold. Both agreed with the reading, `architecture` at length, but flagged
that the author is not the ruling's owner.

---

## CLOSED 2026-08-04 — FIXED AND LIVE on `v1.0.1250`, and the two-branch induction has RUN

Both halves of this file's own bar are met, so this moves to `bugs_closed/`.

### Live at the pod, with three controls (not inferred from the roll)

`v1.0.1250`, both replicas (`agent-chassis-88cf8787-4dzzx`, `-5z5sn`):

```
NEW  resolveSubstepContinueOnError                        = 2
NEW  "Substep declares continue_on_error with a non-bool" = 1
CTRL "continue_on_error is true for this loop iteration"  = 1   (pre-existing; probe works)
NEG  resolveSubstepContinueOnErrorXYZZY                   = 0   (probe CAN return 0)
```

Pre-roll the same four probes read `0 / 0 / 1 / —`. The negative control is what makes the
positives mean anything: a `grep -c` that never returns 0 is not measuring.

### The induction — BOTH branches, which is what this file demanded

Two throwaway agents, each with the **loop**-level flag set to the **opposite** of the
substep's, so a stamp-regression could not pass as a success: each run would have produced
**the other run's outcome**. Seed: `docs024_key_docs_latest/bugfix_173_substep_error_tolerance/SEED_2026-08-04_induction_agents.sql`.

| run | loop | substep `boom` | status | current_step | mechanism |
|---|---|---|---|---|---|
| `35acb827-6def-4285-92a7-395e131daa01` | **(unset)** strict | **`true`** | **COMPLETED** | `complete` | `iter_0_error` + `iter_1_error` both present, `skipped=true`, `run_loop_error_count=2` |
| `982bf0ce-b948-4da8-a848-184abb106b41` | **`true`** tolerant | **`false`** | **FAILED** | `run_loop_iter_0_boom` | died AT the failing substep; no skip records |

Chassis log corroboration, both iterations of the tolerant run:

```
Skipping failed loop iteration, advancing to next   step=run_loop_iter_0_boom iter=0 total_errors=1
Skipping failed loop iteration, advancing to next   step=run_loop_iter_1_boom iter=1 total_errors=2
```

**Why this is a proof and not a green light.** The tolerant run would read COMPLETED just as
happily if the induced fault had never fired — so the discriminating evidence is the *error
records*, not the status. They exist, for both iterations, with `skipped=true`. And the strict
run is the direction that matters: a tolerant loop did **not** swallow a substep that declared
`false`. Before this fix that was unexpressible.

The induced fault was `SELECT 1/0` via `query_database`, which returns a real Go error
(`query failed: %w`) rather than a soft result — checked in the source before dispatching,
because a fault that does not fire makes the tolerant branch pass for the wrong reason.

**Cleanup done:** both `test-173-*` agent definitions deleted, `0` remaining.

### Council

**APPROVED round 1**, correlation `549e25fb-acc1-4806-a2a7-95bf73cca806`, 8 seats, 3 advisory
objections, none high — all answered with checks (above and in the lane NOTES). Registered as
**WFA-008**.

### What this bug leaves behind, deliberately

- **`bugs_open/193`** — the loop-level read of this same key silently ignores a non-bool where
  its substep-level twin now warns. Filed at the `bug_historian` seat's direction. Latent.
- **A correction for anyone quoting this file:** §"Why it matters" is historical — its
  motivating agent `multipage-website-builder` is deleted. The live consumers are
  `pageflow-builder`, `page-rebuild` and `site-work-orchestrator`.
- **Struck:** this file's third "not established" item (the `fallback_step`/retry interaction).
  There is no interaction because there is no mechanism — both keys occur twice in the whole Go
  tree, in name-prefixing lists only, with 0 live definitions declaring either.
