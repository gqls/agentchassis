# HANDOFF — the truncation contract (bugs_closed/076), continue here

**Written 2026-07-26 ~21:15 UTC.** Cold-start entry point for anyone picking up
what is left of `076`. Every figure below was run live against `clients_db` or
the running pod on that date — **re-ground them before you rely on them**, they
go stale within days.

> **UPDATED 2026-07-26 evening — R1 IS BUILT.** This file is still the cold-start
> entry, but read it knowing §R1 below is **done**, not pending. Start with:
> `PLAN_2026-07-26_r1_static_truncation_consumer_check.md` (design + the
> correction to this file's own R1 brief), `RUNBOOK_…` (every command, with its
> gotcha), `NOTES_…` (what shipped, and three defects found *in the checks*),
> `README_where_we_are.md`. The remaining items are R3/R4/R5 plus the reduced
> R2 — all small, none keeping anything open.
>
> **Why this directory used to hold one file and now holds five.** The case was
> CLOSED, so five docs about finished work would have been padding. R1 turned it
> back into a workstream, and the five were started at the beginning of it — as
> the paragraph this replaces asked.

**Read first:** `bugs_closed/076_HANDOFF_2026-07-25_truncated_llm_responses_tolerated_at_113_unguarded_call_sites.md`
(the case file — the filename keeps its refuted headline so old pointers still
resolve; the corrected statement is the H1) and `016b` §9 *"A contract with two
halves fails closed on the half you wrote and is enforced by nothing on the half
you rely on"*.

---

## 1. The one-paragraph version

`ExecuteLLMPromptAction` may keep an LLM response the model cut at `max_tokens`,
stamping `__truncated` beside the result instead of failing the step. That is
only sound while something downstream **reads** the marker. Tolerance was already
opt-in and defaulted false (the producer half was fine); the **consumer half was
enforced by nothing**. `076` added a runtime guard — a producing step now refuses
to tolerate unless its workflow contains a step that reads the marker — plus a
durable record of what the consumer then did about it. **The case is CLOSED:
fixed, live, and proven by inducing the failing branch.** The one real follow-on
— a static check over the config, so a bad seed is caught before any run exists —
**was built on 2026-07-26 evening** (§R1). What remains is four small carried
items, none of which reopens anything.

## 2. State as of 2026-07-26 21:10 UTC — verified, not assumed

| thing | state | evidence |
|---|---|---|
| case file | `bugs_closed/076_…` | moved in `8a80549d4` |
| live | **yes**, chassis **v1.0.1171** | pod-grep: `guard:1 REFUSED:1 degrade:1` |
| council | **APPROVED ×2** — `470678f4…` (guard), `1535e2ac…` (residuals) | round 1 each, 12 reviewers, `unreadable:0` |
| trailer | **impossible on both** — verdicts post-date their commits, forward-only forbids the amend | permanent `098` false negative, recorded so nobody re-investigates |
| exposure | 37 tolerating steps: `council-gate` 16, `fix-proposer` 16, `feature-designer` 5 | all consumed by `diagnose_council_decide`, which is guarded |
| degradation rows | 3, between 18:08 and 18:33 | `error_code='TRUNCATION_DEGRADED_REVIEW'` |

Commits: `4208e7f41` + `37a4e7b9a` (guard), `0657143b5` (durable record +
inverse lockstep test), `c640a8642` (the `REFUSED` log prefix), `8a80549d4` /
`8d8bb64ad` / `2f4a899c7` / `c2dd72ae1` (the case file).

Code: `platform/orchestration/actions/truncation_guard.go` (registry +
`findTruncationAwareConsumer`), `ai_actions.go:~395–500` (the call site and the
log prefix), `diagnose_council_decide_action.go` (`recordTruncationDegradation`),
`truncation_guard_test.go` + `truncation_degradation_record_test.go`.

**The important number, because it re-frames everything:** the case was filed as
*"113 unguarded call sites"* and that was **wrong**. Those 113 steps never
tolerated anything — they fail closed by default. The exposure was 37 steps, all
already guarded, i.e. **latent, not active**. The measurement that produced the
original figure counted definitions that *mention* `__truncated` in config text,
but one guarded agent's guard is in **Go**. Read the producer before believing a
count about consumers.

---

## 3. THE RESIDUALS

### R1 — the static check at seed time *(**BUILT 2026-07-26 evening**)*

> **DONE. What follows is the original brief, kept because its reasoning is still
> the reasoning — with two corrections it earned.** What shipped:
>
> - `scripts/truncation_registry.py` — parses `truncationAwareActions` and
>   `acceptsTruncatedConfigKey` out of `truncation_guard.go`. **This is the answer
>   to the landmine below**: no checker holds a copy, and the parser raises rather
>   than falling back to a remembered list.
> - `…/fixloop_eg_dartsonline/103_LINT_truncation_consumer.py` — the live-DB lint
>   (`--verbose`, `--self-test`, `--strict`). Also lists every `accepts_truncated`
>   declaration as an unverified claim, which is as much of **R2** as is reachable
>   without a Go change.
> - `check_truncation_without_reader` in `scripts/pattern-check.py` — commit time,
>   via `.githooks/pre-commit`. **0 findings over 849 tracked `.sql` files.**
> - a pointer at the end of `scripts/migration/run-migrations.sh --apply`.
>
> **CORRECTION 1 to the brief below.** It proposed the pre-commit layer as
> catching "a **seed file** introducing the bad config… most config arrives as a
> committed seed". Measured, all three files in the repo that arm the flag are
> `jsonb_set` **patches** whose target workflow lives in the DB — a textual check
> over them can only guess, and on all three the guess is wrong (their targets are
> guarded). L2 is therefore scoped to files that **embed** a workflow. Building
> what was written would have flagged three correct files on day one.
>
> **CORRECTION 2.** The brief's "realistic shapes" list said a report "blocks
> nothing" as if that were its weakness. It is the design: `pattern-check.py`'s
> header records that a blocking check on a shared tree becomes a fleet-wide
> outage and then gets disabled permanently.
>
> **Validation, since the fleet is clean and a clean report proves nothing:** three
> probes seeded live (no reader / hatch reader / string-valued flag) were flagged,
> cleared and reported inert respectively, then deleted — exactly the induction the
> brief below asks for. Full record: `NOTES_…`, `RUNBOOK_…`, and §R1 of the case
> file.

<details><summary>Original brief (kept for its reasoning)</summary>

**What.** A check over `agent_definitions.default_config` that flags any step
with `tolerate_truncation: true` whose workflow contains no truncation-aware
consumer — at **seed/registration time**, rather than at the moment a truncation
actually happens.

**Why it matters.** The shipped guard is a **floor that fires late**. It catches
the bad config only when a response is actually cut, in production, by failing
that run. A static check catches the same config **before any run exists**. It is
also the only layer at which **R2** can be validated.

**Raised by.** Council seat `guardian`, medium severity, corr `470678f4`:
*"has a higher-layer alternative been ruled out? E.g. validating
'tolerate_truncation:true with no reader' as a static check over
agent_definitions.default_config at workflow registration/deploy time."* The
submission never discussed that axis — a fair hit.

**The complication you must design around, and it is the whole difficulty.**
There is **no registration or deploy step to hook**. CLAUDE.md's own invariant is
that *DB config is live immediately* — a seed can add `tolerate_truncation: true`
to a live workflow with no build, no deploy and no restart. So a genuine
"registration-time" gate does not exist to attach to. Realistic shapes:

- a **report** in the `097/098` family, run on demand and/or scheduled, listing
  offending steps (cheapest, catches drift within a day, blocks nothing);
- a check inside `scripts/pattern-check.py` or the migration lint, so a **seed
  file** introducing the bad config is caught at commit time (catches the common
  path — most config arrives as a committed seed — but not a hand-run `UPDATE`);
- a startup scan in the chassis that logs/records offenders on boot (catches
  everything eventually, but only at the next roll).

Probably **the first two together**. Do not build a blocking gate on the third
without an owner ruling — it would make a bad seed take the fleet down.

**The query is already written** — it is the live mirror of the guard's own
predicate, used to measure blast radius before shipping:

```sql
WITH steps AS (
  SELECT d.type, e.k AS step, e.v AS step_cfg, d.default_config->'workflow'->'steps' AS all_steps
  FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') e(k,v)
  WHERE d.is_active AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
    AND e.v->>'action'='execute_llm_prompt'
    AND COALESCE(e.v->'config'->>'tolerate_truncation','false')='true'
)
SELECT type, step,
       EXISTS (
         SELECT 1 FROM jsonb_each(all_steps) o(ok,ov)
         WHERE o.ok <> steps.step
           AND (ov->>'action' IN ('diagnose_council_decide','verify_report_prose')
                OR COALESCE(ov->'config'->>'accepts_truncated','false')='true')
       ) AS has_guarded_consumer
FROM steps ORDER BY type, step;
```

**That query was run on 2026-07-26 21:12 and works** — it returns exactly **37
rows, every one `has_guarded_consumer = t`** (council-gate 16, fix-proposer 16,
feature-designer 5). So the report R1 would produce is currently **empty**, which
is the expected and desired state: R1 is about catching the *next* bad seed, not
an existing offender. That also means you cannot validate R1 against live data
alone — **seed a deliberate offender and confirm the check flags it**, then
remove it.

**Landmine in that query:** the action list is a **hand-copy of the Go registry**
`truncationAwareActions`. Two hand-maintained lists that must agree is exactly
the drift class this platform keeps paying for (see the council gate's `099`
roster mirror). If you build R1, **generate the list from the Go source or make a
test assert the two agree** — do not leave a second copy.

**Done looks like:** an offending seed is caught before it can be exercised, and
the report shows zero offenders today (it should — blast radius was measured zero
on 2026-07-26 and all 37 tolerating steps sit in guarded workflows).

</details>

---

### R2 — `accepts_truncated` is trusted, never checked *(REDUCED, not closed)*

> **2026-07-26:** R1's lint now **lists** every step declaring the hatch, labelled
> as an unverified claim, so a wrong or copy-pasted flag is visible to anyone who
> runs it — proven against a probe that used it. Still **zero** live users. What
> remains unbuilt is *verification* that a declaring action really reads the
> marker, which needs a Go-side change and is the one part that would flip the
> "no council submission" decision. The paragraph below still stands.

`accepts_truncated: true` in a step's config declares "my action handles a
partial". **Nothing verifies that.** A wrong or copy-pasted flag re-opens the
exposure for that workflow and no test, seed check or report would say so.

Raised by `bug_historian` (medium, corr `470678f4`), who put it well: it
*"reopens exactly the pattern this fix is meant to close"*.

**It is demonstrated, not theoretical** — my own positive-control probe used
`accepts_truncated: true` on a `complete_workflow` step that reads nothing at
all, and the guard duly let the truncation through. That was the point of the
control, but it is also the hole.

**Why it was shipped anyway:** the Go registry is falsifiable by test; a config
flag lives in the DB, so no Go test can reach it. The hatch exists for actions
too generic for the registry to speak for. Today **zero** live steps use it —
verify before assuming:

```sql
SELECT d.type, e.k FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') e(k,v)
WHERE d.is_active AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
  AND COALESCE(v->'config'->>'accepts_truncated','false')='true';
```

**Do not fix this separately** — R1 is where it gets validated. If R1 is never
built, consider whether the hatch earns its keep at all: a hatch with no users
and no validation is a liability, and deleting it is a one-line change plus the
test that pins it.

---

### R3 — `truncationMarkerExemptions` is a trust boundary the test cannot police *(accepted, small)*

`TestEveryMarkerReaderIsRegisteredOrExempt` requires every non-test file reading
`__truncated` to implement a registered action **or** sit in an exemption map
with a stated reason. The test rejects an **empty** reason. It cannot reject a
**false** one — a future contributor could exempt a real consumer with a
plausible-sounding rationale and it would pass.

Raised by `guardian` and `editquality` (both low, corr `1535e2ac`), and both are
right: this pushes the maintenance surface down a level rather than removing it.

**Why it is nonetheless much smaller than what it replaced:** three entries, each
with a stated reason, living in the same package as the thing they exempt; and a
**rename** shows up as an unexplained new reader rather than silently passing.
Current entries — `ai_actions.go` (the producer: it stamps the marker, never
reads one back), `types.go` (documentation only), `truncation_guard.go` (the
registry itself).

**Probably leave it.** Recorded so nobody rediscovers it as news.

---

### R4 — the guard's floor may be too low, deliberately *(UNVERIFIED, by design)*

`findTruncationAwareConsumer` asks whether the workflow contains **a** reader —
not whether the truncated value ever **reaches** it. A workflow whose guarded
consumer sits on a branch the fragment never flows down would still pass.

**This was chosen on evidence, not laziness.** The stricter gate (candidate 2:
"does *this* step's value get referenced by a guarded consumer") needs
reference-detection across prompt templates and config paths, and a textual scan
against live config matched **28** "consumers" in the council agents alone,
nearly all spurious. A gate built on it would fail closed on sound workflows.

The code says so in its own doc comment, which is the point — *"a FLOOR, not a
proof"*. Only revisit if a real case shows the floor being cleared by a workflow
that then ships a fragment. **No such case has been seen.**

---

### R5 — `platform/orchestration/orchestration_test.go:171` does not compile *(NOT OURS)*

```
vet: platform/orchestration/orchestration_test.go:171:74:
     not enough arguments in call to orchestration.NewSagaCoordinator
```

Still broken at HEAD, confirmed 2026-07-26 21:10. Pre-existing, another lane's
in-flight signature change, untouched by this work — **but it means that
package's tests do not run at all**, so anyone relying on `platform/orchestration`
test coverage is relying on nothing. `platform/orchestration/actions` (where all
of `076` lives) compiles and passes cleanly.

See the `shared-tree-wont-compile` practice: test against `git archive HEAD` plus
your own files overlaid; never edit another session's test to make yours run.

---

### Settled, so you do not re-open them

- ~~the consumer's degradation is invisible~~ — **fixed**, `0657143b5`; three real
  rows captured on day one.
- ~~the registry can silently underclaim~~ — **fixed**, inverse lockstep test.
- ~~`diagnose_council_decide`'s degrade path never exercised end to end~~ —
  **exercised**, twice, by natural production traffic.
- ~~downstream might key on `severity`+`error_code`~~ — **checked 21:10 and
  clear.** `reconcile_superseded_reviews_action.go:215` filters a different
  `error_code` *and* `site_id` (these rows leave it NULL);
  `diagnose_load_runtime_action.go:267` reads generically into a diagnosis bundle.
  Three DB definitions match both words coincidentally (schema hint + their own
  objection severity). One real effect: a `diagnose_load_runtime` bundle gathered
  with a NULL `site_id` now includes these rows as context.

---

## 4. Loose ends from my session — check before you fire anything

- **Two stranded council orchestrations on corr `1535e2ac`**, both
  `EXECUTING_STEP`, neither progressing: one created `17:57:34` (stale ~3h), one
  `18:44:58` (stale ~2h). Both are casualties of image rolls, not defects in the
  submission — **the verdict I acted on is the `18:35:57` run, which reached
  `complete_approved`.** They are inert clutter; if a reaper claims them, expect
  it. Related: `bugs_open/070` (stale reaper keys on row age).
- **A submission fired on a fresh correlation `75faf4d4-…` at 18:33 never
  appeared at all.** If it ever lands it will produce a duplicate council round
  on the same plan. Harmless, but do not read it as a new verdict.
- **The probe agents were deleted** (`truncation-guard-probe`,
  `truncation-guard-probe-ok`) — confirmed 0 remaining. If you re-run the
  verification, delete them again afterwards: they are `is_active` definitions in
  the live fleet.

## 5. Operational knowledge you will need

- **A chassis roll strands every in-flight orchestration** at whatever step it
  was on, leaving `status = 'EXECUTING_STEP'` — which looks identical to healthy
  progress. My own roll killed my own council run. The only way to see it is
  staleness, not status:
  `EXTRACT(EPOCH FROM (now()-updated_at))`.
- **Roll the chassis ALONE.** All ~183 agent definitions run the single
  `docker.io/aqls/agent-chassis` image, so bump `IMAGE_TAG`, build, push, edit
  **only** `deployments/kustomize/services/agent-chassis/overlays/production/uk_001/kustomization.yaml`,
  `kubectl apply -k` that one overlay, then
  `UPDATE agent_definitions SET image_tag=…`. **Never `make deploy-agents`** — it
  seds all 14 services to `IMAGE_TAG` and the other 13 have no image at it.
- **Budget ~6 minutes of dispatch latency before concluding anything.** I
  produced three tidy, confident, mutually exclusive explanations for a
  "vanished" dispatch inside half an hour — a rejected `client_id`, the ~300s
  post-restart drop, and "the exit test proves it was dropped". All three were
  refuted; it was ordinary queue latency (5m37s). Each guess cost a duplicate
  dispatch; waiting would have cost nothing. If you do use `bugs_open/052`'s exit
  test (*has anything newer drained past me?*), the comparator must be **on your
  lane** (`endpoint-health-checker` / `build-pipeline-trigger` are on their own
  cron topic per `bugs_closed/030`) **and published after your message** (FIFO:
  an older job progressing is compatible with yours still waiting).
- **Two markers, one underscore apart, same package.** `__truncated` is the LLM
  cut; `__truncated__` is `workflow_actions.go`'s stub for a Kafka response over
  `max_response_bytes`. Unrelated mechanisms; `strings.Contains` conflates them.
  `readsLLMTruncationMarker` exists solely to keep them apart.
- **`agent_error_log`'s timestamp column is `occurred_at`, not `created_at`.**
- **`check_ad_category`** allows only strategist/executor/analyst/integrator/
  coordinator/specialist — `'system'` is rejected when seeding a probe agent.

## 6. How to re-verify the whole thing

**Deployment** (discriminating — the strings are new, so they cannot pass on an
old binary; the last two are controls):

```bash
POD=$(kubectl get pods -n ai-persona-system -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl exec -n ai-persona-system $POD -- sh -c '
  strings /app/agent-chassis | grep -c "no step in this workflow consumes the __truncated marker"  # 1
  strings /app/agent-chassis | grep -c "REFUSED (bugs_open/076"                                    # 1
  strings /app/agent-chassis | grep -c "TRUNCATION_DEGRADED_REVIEW"                                # 1
  strings /app/agent-chassis | grep -c "tolerate_truncation"                                       # 3  (positive control)
  strings /app/agent-chassis | grep -c "this_string_should_not_exist_076"'                         # 0  (negative control)
```

**Correctness — induce it, and READ WHAT THE FAILURE WROTE.** Seed two
throwaway `diagnose`/`coordinator` agents, identical but for `accepts_truncated`
on the downstream step: `workflow.start_step`, `processing_mode: "orchestrator"`,
one `execute_llm_prompt` step with `max_tokens: 16` and
`tolerate_truncation: true`, then a `complete_workflow` step. Fire each with the
`095_TRIGGER` envelope on `system.agent.generic.requests`, `action=orchestrate`,
`config.agent_type=<probe>`. Expect:

| probe | expected |
|---|---|
| no reader | `orchestration_states.status = FAILED` at the LLM step; `.error` names bugs_open/076; **no output key in `collected_data`**; `llm_call_log` prefixed `REFUSED (bugs_open/076…` |
| `accepts_truncated: true` | `COMPLETED`; result carries `"__truncated": true`; `error` empty; `llm_call_log` prefixed `TOLERATED (step continued…` |

The **control is what makes it a proof** — a guard that blanket-failed every
truncation looks identical on the first row alone. Then delete both agents.

**Regression** — the 37 guarded steps: any live `council-gate` or `fix-proposer`
round reaching a verdict on the new binary. Other threads' rounds are better
evidence than one you control.

**Unit** — `go test ./platform/orchestration/actions/ -run 'Truncation'`. And
**falsify anything you add before believing it.** This case's first lockstep test
was vacuous: `truncation_guard.go` names `__truncated` in its own doc comment, so
every registry entry validated itself and a deliberately bogus entry passed.
Five probes were run against the current tests and all five failed with the right
message before passing on restore.

## 7. Pointers

- `bugs_closed/076_…` — the case file; §"LIVE AND VERIFIED", §"RESIDUALS CLOSED",
  §"Council on the residuals", §"Residual (carried past the close)".
- `bugs_closed/019` — one truncated reviewer voided a whole council round; the
  case that introduced tolerate-and-mark. Its workstream dir is
  `docs024_key_docs_latest/bugfix_019_council_truncation/`.
- `bugs_closed/046` — truncation casualties swept from live tools
  (`truncation_casualties_046/`). Different truncation, adjacent family.
- `bugs_open/012` — a rewrite persisted a fragment and reported success; the
  `output_tokens == max_tokens` rule in CLAUDE.md comes from it.
- `bugs_open/021` — durable write guard covers one path only. **Same shape as
  076**, owned by the durable-write-guard workstream — contribute there, do not
  start a competing fix.
- `016b` §9 — the transferable pattern, appended 2026-07-26.
- `WRONG_CALLS.md` — my three refuted dispatch theories, and why the tally
  matters more than any one entry.
