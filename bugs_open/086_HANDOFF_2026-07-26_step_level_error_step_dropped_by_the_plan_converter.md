# 086 — the plan converter drops step-level `error_step`, so 55 declared error handlers have never once run

**Filed:** 2026-07-26 · **By:** the `bugs_open/068` thread (068's real mechanism turned out to be this)
**Status:** OPEN — **both halves are now LIVE, and the Go half shipped sooner than anyone chose**:
 · **LIVE**: the contained config twin for the diagnosed step (seed `219`, applied 2026-07-26)
 · **LIVE in v1.0.1169**: the fleet-wide Go fix, committed `dca5649b3` at 18:02Z — another thread
   built and rolled the chassis at ~18:22Z, and `make build` takes committed HEAD, so it rode
   their build. Pod-verified: `strings /app/agent-chassis | grep -c "Step declares an error_step
   that is not in the plan"` = 1 (a string only this change creates), positive control also 1.
   **Council VETOED it; the owner overruled that on the measurement (below) believing it was
   inert until a roll — it was already 20 minutes from shipping.** The 55 handlers are armed now.
   Residual: the ten `→ complete` handlers were still unaudited when they went live.
**Severity:** High for anything that declares a non-fatal step and gets a fatal one instead;
the measured instance (`page-content-writer.resolve_links`) has killed **32 of 32** runs that hit it.

> **Numbering note.** This case was written as `085` and committed under that number in
> `dca5649b3`'s message; `085` was claimed minutes earlier by the brochure workstream
> (`085_..._render_data_advertises_current_page...`). The code comments and this file say **086**;
> the commit message of `dca5649b3` still says 085 and cannot be amended (forward-only).
> The council submission (corr `88ef6d08-ca87-4c90-b682-f85f1e6036f1`) also says 085.

## Symptom

A step whose agent definition declares `"error_step": "<handler>"` fails **fatally** instead of
routing to its handler. The handler is never reached, the orchestration goes to `FAILED`, and the
parent (if any) is told the child failed.

Concretely, `page-content-writer` declares its link-resolve step as explicitly non-fatal:

```json
"resolve_links": {
  "action": "call_agent",
  "next_step": "select_sections",
  "error_step": "select_sections",
  ...
}
```

and the seed that wired it says so in prose:
`content_quality_and_internal_linking/page_content_writer_link_resolver_wiring.sql:6-7`
— *"NEW resolve_links — call_agent once per page, AFTER build_render_context, passing the section
plan; error_step falls back to …"*.

Every `resolve_links` failure ever recorded is fatal anyway:

```sql
SELECT severity, count(*), min(occurred_at)::date, max(occurred_at)::date
FROM agent_error_log WHERE step_name='resolve_links' GROUP BY 1;
-- fatal | 32 | 2026-06-26 | 2026-07-24     (no other severity, ever)
```

30 of the 32 are `Request … timed out after 3 retries` — a slow link resolver, which is precisely
the case the author declared survivable. Not one routed.

## Root cause (code + live-data evidenced)

`convertToWorkflowPlan` builds `models.Step` **field by field** and never names `error_step`
(`platform/messaging/processor.go`, pre-fix):

```go
step := models.Step{
    Action:      p.getStringValue(stepMap, "action"),
    Description: p.getStringValue(stepMap, "description"),
    NextStep:    p.getStringValue(stepMap, "next_step"),
    OutputField: p.getStringValue(stepMap, "output_field"),
    Topic:       p.getStringValue(stepMap, "topic"),
}
```

`models.Step` **has** the field (`pkg/models/contracts.go:60`, `json:"error_step,omitempty"`), and
the coordinator reads it — from the **persisted plan**, not from `agent_definitions`
(`platform/orchestration/coordinator.go:3225-3239`, `routeToErrorStepOrFail`): step-level first,
then `step.Config["error_step"]` as a backward-compatibility fallback.

So `config.error_step` works (the whole `config` map is copied verbatim) and step-level never has.

**The discriminating measurement** — persisted plans, not definitions:

```sql
SELECT count(*) AS steps,
       count(*) FILTER (WHERE value ? 'error_step')            AS step_level,
       count(*) FILTER (WHERE value->'config' ? 'error_step')  AS config_level
FROM (SELECT workflow_plan FROM orchestration_states WHERE created_at > NOW() - INTERVAL '3 days') o,
     jsonb_each(o.workflow_plan->'steps');
-- 14209 | 0 | 1828
```

**Zero, fleet-wide, across every agent.** Side-by-side on one agent: `page-build-handler`'s
definition carries 10 step-level `error_step`s; its newest persisted plan carries none, and only the
6 that are *also* in `config` survive.

## Census of the inert set

```sql
SELECT count(*) AS declared,
       count(*) FILTER (WHERE value->'config'->>'error_step' IS NULL) AS step_level_only,
       count(DISTINCT type) AS agents
FROM agent_definitions, jsonb_each(default_config->'workflow'->'steps')
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND value->>'error_step' IS NOT NULL;
-- 63 | 55 | 19
```

**55 handlers across 19 agents are inert** (the other 8 are saved by a config-level twin). Sample:
`image-build-handler` ×13, `content-gap-planner` ×5, `improvement-loop` ×6, `site-adoption-agent` ×5,
`css-patch-agent` ×4, `page-build-handler` ×3, `page-content-writer` ×2 (incl. `resolve_links`).

Loop **substeps** are a second drop site (`parseSubsteps`) and a third (loop expansion). Censused
live: **0 definitions declare a substep-level `error_step` today**, so that half was latent, not biting.

## The fix (committed `dca5649b3`, inert until the roll)

1. `platform/messaging/processor.go` — carry `ErrorStep` in `convertToWorkflowPlan`, plus an
   advisory warning when an `error_step` names a step absent from the plan (a typo currently
   degrades to "step not found", i.e. the pre-fix outcome with a worse message). Warning only.
2. `platform/orchestration/actions/loop_actions.go` — same field in `parseSubsteps`.
3. `platform/orchestration/loop_expansion_handler.go` — carry it onto each injected step, prefixed
   `<loop>_iter_<N>_<substep>` **only** when it names a substep of that loop; external targets pass
   through. Deliberately **not** `resolveIterationNextStep`: an error handler is not a chain link and
   must not roll over to the loop's `_complete` step on the last iteration.
4. Three tests, each **falsified** by reverting its own line before being trusted
   (`error_step_plan_test.go`, `error_step_loop_expansion_test.go`, `error_step_substep_test.go`).
   They pin the negative controls too: a step with no `error_step` stays empty, and a config-level
   one is **not** promoted into the field.

Run them against a clean tree — `platform/orchestration/orchestration_test.go` does not compile at
HEAD (a stale `NewSagaCoordinator` call, not ours), so `go test ./platform/orchestration/` fails for
unrelated reasons:

```bash
T=$(mktemp -d); git archive HEAD | tar -x -C "$T"
cp platform/messaging/{processor.go,error_step_plan_test.go} "$T/platform/messaging/"
cp platform/orchestration/{loop_expansion_handler.go,error_step_loop_expansion_test.go} "$T/platform/orchestration/"
cp platform/orchestration/actions/{loop_actions.go,error_step_substep_test.go} "$T/platform/orchestration/actions/"
rm -f "$T/platform/orchestration/orchestration_test.go"
(cd "$T" && go test ./platform/orchestration/... ./platform/messaging/)
```

## Blast radius — stated, not hidden

The fix makes **55 handlers live at once**. Ten of them route to `complete`, so a failure that today
FAILS an orchestration will instead COMPLETE it:

| agent | step | action |
|---|---|---|
| blog-content-planner | create_post_pages | create_blog_posts |
| content-gap-planner | plan_gaps | execute_llm_prompt |
| content-gap-planner | apply_plan | apply_gap_plan |
| image-build-handler | mark_work_item_complete | update_work_item_status |
| image-build-handler | flag_rebuild | flag_page_image_rebuild |
| site-adoption-agent | write_design_intent | write_site_spec |
| site-adoption-agent | generate_design_intent | execute_llm_prompt |
| spec-updater | apply_update | update_site_spec_from_item |
| tool-improver | note_refusal | append_doc_note |
| webdesign-agent | fork_theme | fork_theme_from_site |

That is what those definitions **say**, and the converter is the wrong place to overrule them — but
it touches the "complete means complete" lane (`work_item_completion_integrity`, `bugs_closed/017`,
`bugs_closed/028`). It is not silent: `routeToErrorStep` writes an `agent_error_log` row at
`severity='error'` and sets `collected_data.__step_error`. **Before the image rolls, someone should
decide whether any of those ten wants its definition changed rather than its behaviour honoured** —
that is a config edit, live immediately, and does not block this fix.

## How long it has been broken

`error_step` has been in seeds since **2026-02-24** — at `config` level, in the dispatch loop, where
it has always worked. The step-level field arrived on **2026-05-18** (`7d6420c46`, *"adoption error
steps - carry on after error - coordinator patch"*), which added `ErrorStep` to `models.Step` **and**
the coordinator's preference for it, but not to the converter. Every step-level declaration written
since — ten weeks of them — has been discarded on the way in. It hid because the symptom is the
*absence of a rescue*: the run just fails, which looks like an ordinary failure.

## Verification of the LIVE half (seed 219)

- [x] Definition carries both levels, identical: `step_level = config_level = select_sections`,
      `next_step` and callee untouched (the seed's own post-check).
- [x] The config path is demonstrably live in plans built **after** the seed: orchestrations
      created 18:22–18:26Z carry `config.error_step` (e.g. `build-pipeline-trigger`, 2 per plan)
      while step-level remains 0 everywhere — the Go half is still inert, as expected.
- [ ] **Outstanding, needs the next writer run** (none since 17:55Z; they run every few hours,
      driven by page builds): the twin must appear in a fresh `page-content-writer` plan.

  ```sql
  SELECT created_at, status,
         workflow_plan->'steps'->'resolve_links'->'config'->>'error_step' AS cfg_error_step
  FROM orchestration_states WHERE owner_agent_type='page-content-writer'
  ORDER BY created_at DESC LIMIT 1;   -- want: select_sections
  ```

  Then the branch itself: the next `resolve_links` timeout must produce a run that continues to
  `select_sections` with an `agent_error_log` row at `severity='error'` (routed) instead of
  `fatal`. Until that is seen, this is a config change proven at the definition and at the plan
  builder, not yet at the failure it is meant to catch.

## Verification (post-roll — the Go half is not proven live yet)

The obvious pod-grep is vacuous here (the string `error_step` is all over the binary already). The
discriminating marker is the **data**:

```sql
-- today: 0. After the roll, a fresh page-content-writer run must show >0.
SELECT count(*) FILTER (WHERE value ? 'error_step') AS step_level
FROM (SELECT workflow_plan FROM orchestration_states
      WHERE owner_agent_type='page-content-writer' ORDER BY created_at DESC LIMIT 1) o,
     jsonb_each(o.workflow_plan->'steps');
```

Then induce the failing branch: a `page-content-writer` run whose `resolve_links` times out must
continue to `select_sections` and the run must not be `FAILED`, with an `agent_error_log` row at
`severity='error'` (routed) rather than `fatal`.

## The contained fix that IS live (seed 219)

`page-content-writer.resolve_links` now carries `config.error_step: select_sections` alongside its
step-level declaration — the same target, on the path that already works. Applied 2026-07-26,
recorded in `schema_migrations` (`record-only`), REVERT is an exact jsonb key removal in the seed
header. Effect: a slow or failing link resolver no longer kills the writer; it continues at
`select_sections`. That covers 32 of the 109 failures tabulated below and needs no image roll.

The seed mirrors the step-level value rather than hardcoding it, so the two cannot drift; when the
Go fix ships, the coordinator prefers the step-level field and both name the same step, so they
compose. **Convention miss, recorded not papered over:** the seed has no `snapshot_agent()` call
(the house rule for anything touching `agent_definitions`). Recovery does not depend on it here —
the REVERT removes exactly the one key that was added.

## Council: REJECTED (guardian hard veto) — and the owner's ruling

Submitted 2026-07-26, corr `88ef6d08-ca87-4c90-b682-f85f1e6036f1`. 12 seats fired,
`unreadable: 0` (so this is a real objection, not a dead seat — check that first, per
`bugs_closed/019`). Five approved: **editquality** (two low/medium objections: the dangling-target
warning is scope beyond "carry the field", and the loop prefix scheme is asserted by analogy to
`prefixConfigStepReferences` rather than quoted), **debug_historian**, **constitution** ("a
root-cause fix, not a workaround"), **mission**. **prior_art_librarian objected** that the
load-bearing counts were paraphrased rather than shown as query output — fair, and the queries are
now inline in this file.

**guardian: VETO.** Its case, which is the same one this file's blast-radius section makes, pushed
harder: the change lands in the orchestrator's plan-build and loop-expansion core; it activates 55
dormant handlers across 19 agents in one shot; and *"acknowledging the risk is not the same as
containing it."* Its named safest contained alternative is exactly seed 219 above — config twins
for the reviewed steps, one agent at a time — with the converter fix reserved for
"a follow-up architecture review".

**There is no architecture-review agent.** The council gate is the review system; the fleet's six
`*-architect` agents design websites, not code. "Architecture review" means the owner. So the
question went to him with the blast radius re-measured, and the answer (2026-07-26) was
**keep `dca5649b3`** — that conversation was the review the guardian asked for. What decided it:

| step whose handler is inert | would route to | fatal failures, 30d |
|---|---|---|
| image-build-handler · call_imagery_gen | mark_work_item_failed | 68 |
| page-content-writer · resolve_links | select_sections | 32 |
| image-build-handler · call_asset_deployer | mark_work_item_failed | 5 |
| design-audit-agent · call_content_auditor | triage | 2 |
| improvement-loop · call_site_review | triage_findings | 1 |
| tool-auditor · llm_audit | complete_error | 1 |

**109 real failures in 30 days would have been handled instead of fatal, and not one of the ten
`→ complete` handlers appears** — the dangerous subset is precisely the dormant subset (29 of the
55 belong to agents that have not run in three days). The veto's shape was right; the measurement
says the exposure is small and the benefit concrete.

**The ten `→ complete` handlers: audited and DISABLED (seed `220`, applied 2026-07-26).** The
window to do this before the roll did not exist — the roll had already happened. Read in full:

| agent · step | what `complete` means on failure | read |
|---|---|---|
| spec-updater · apply_update | the update it exists to apply is skipped, run reports success | swallows the job |
| content-gap-planner · apply_plan | plan never applied | swallows the job |
| site-adoption-agent · write_design_intent | intent never written | swallows the job |
| blog-content-planner · create_post_pages | no pages created | swallows the job |
| webdesign-agent · fork_theme | no theme forked | swallows the job |
| content-gap-planner · plan_gaps | nothing planned → apply skipped | borderline |
| site-adoption-agent · generate_design_intent | nothing generated → write skipped | borderline |
| image-build-handler · mark_work_item_complete | status write failed, build finishes | defensible |
| image-build-handler · flag_rebuild | rebuild flag unset, build finishes | defensible |
| tool-improver · note_refusal | refusal note not appended | defensible |

All ten end at a `complete` step whose action is `complete_workflow`, so the orchestration would
finish GREEN. **Owner ruling: disable all ten**, including the defensible three. Seed `220` renames
the key `error_step` → `error_step_disabled_086` on exactly those steps — the converter reads only
`error_step`, so the declaration is inert but stays visible and greppable rather than deleted, and
the REVERT is a rename back. `snapshot_agent()` taken for all 8 agents first. Post-checks: **0**
still routing to `complete`, **10** recorded, **45** other handlers intact.

That restores the loud-failure behaviour these ten steps have actually had for the last ten weeks.
**Per-handler review is still owed** — each author declared `complete` for a reason, and a couple
(the bookkeeping three) are probably right. Re-enable individually by renaming the key back.

**No `Council-Reviewed:` trailer**, and there cannot be one: the verdict is REJECTED, and it
post-dates `dca5649b3` regardless. `dca5649b3` will list as un-reviewed in the 098 report —
correctly. This section is the review record.

## Transferable pattern (016b §9)

**A field in an agent definition is only live if the plan converter copies it. Check
`orchestration_states.workflow_plan`, not `agent_definitions`.** Config-under-`config` survives
because the map is copied wholesale; anything named at step level survives only if some Go line
names it too. The general shape: *a hand-maintained field list next to a struct that grows*.

**Related:** `bugs_open/068` (the case this came out of — its stated mechanism was wrong and is
corrected in `bugs_closed/068`), `bugs_open/087` (the next defect on the same rebuild path),
`bugs_closed/042` (numeric step config never reaching actions — same family: config that looks
declared and is not delivered).
