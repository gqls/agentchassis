# 086 — the plan converter drops step-level `error_step`, so 55 declared error handlers have never once run

**Filed:** 2026-07-26 · **By:** the `bugs_open/068` thread (068's real mechanism turned out to be this)

> **STATUS 2026-07-27: CLOSED.** The outstanding item was never the binary — it was
> **data**: no agent declaring a step-level `error_step` had run since the fix, so the
> converter was proven in `strings` and nowhere else. It has now run, and the persisted
> plans show a clean 0 → 10 step across the roll boundary. The ten `→ complete` handlers
> were audited and disabled on 2026-07-26 (seed `220`) and are verified still disabled.
> See **"CLOSURE 2026-07-27"** at the foot.

**Status (as filed 2026-07-26):** OPEN — **both halves are now LIVE, and the Go half shipped sooner than anyone chose**:
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

## Verification of the Go half (shipped in v1.0.1169 — binary proven, data pending)

- [x] **Binary**: the running chassis carries a string only this change creates —
      `strings /app/agent-chassis | grep -c "Step declares an error_step that is not in the plan"`
      = 1, with a pre-existing string as positive control = 1. (The obvious marker, `error_step`
      itself, is vacuous — it was all over the binary before.)
- [ ] **Data — still outstanding.** No agent that declares a step-level `error_step` has run since
      the roll (only `endpoint-health-checker` and `build-pipeline-trigger`, neither of which
      declares one, so their `0` proves nothing). The discriminating query is below; it needs one
      run of `page-content-writer`, `image-build-handler`, `page-build-handler`,
      `content-gap-planner`, `tool-auditor` or `webdesign-agent`.

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

---

# CLOSURE 2026-07-27 — the data verification, with a roll-boundary control

The file's own discriminating measurement was *"a fresh `page-content-writer` run must
show >0"* — i.e. **read `orchestration_states.workflow_plan`, never `agent_definitions`**.
That box is now ticked, on `page-build-handler` (which declares exactly 10 step-level
`error_step`s) rather than the writer, because it is what actually ran.

**The control is the point — this is a before/after across the v1.0.1169 roll (~18:22Z
2026-07-26), on one agent, in the persisted artefact:**

```
page-build-handler, step-level error_step count in its own workflow_plan
  2026-07-26 15:30:55Z   0     <- pre-roll
  2026-07-26 15:36:06Z   0
  2026-07-26 15:46:33Z   0
  2026-07-26 17:48:13Z   0
  2026-07-26 17:54:29Z   0
  ---- v1.0.1169 rolls ~18:22Z ----
  2026-07-26 18:44:14Z  10     <- post-roll
  2026-07-26 19:06:46Z  10
  ... every plan since ...
  2026-07-27 15:46:06Z  10     (orch 5ad72a12, COMPLETED, on v1.0.1174)
```

Fleet-wide since the v1.0.1174 roll (15:11:15Z): **208 steps, 10 step-level, 33
config-level.** The pre-fix figure in this file was **0 step-level out of 14,209**. The
converter carries the field.

**Seeds re-verified live, not assumed:**

- `219_writer_resolve_links_config_error_step.sql` — applied `2026-07-26 18:23:50Z`.
- `220_disable_error_step_to_complete_handlers.sql` — applied `2026-07-26 18:32:46Z`.
- Current definition census: **10** steps carry `error_step_disabled_086`, **53** still
  carry a live `error_step`. Matches the seed's own post-check (10 disabled, 45 other
  handlers intact, plus the 8 that were already config-twinned).

**What is NOT claimed here, and why it does not hold the case open.** No step-level
`error_step` has been *observed routing* in the wild yet. The nearest live routing —
`build-pipeline-trigger`'s `spawn_dispatch → complete_idle`, firing repeatedly on
2026-07-27 — is **config-level**, checked against the definition rather than assumed, so
it is **not** evidence for this fix. That check is worth recording because the row looks
identical either way: `COMPLETED / complete_idle`, `error` NULL, the failure only in
`collected_data.__step_error`.

The bar for `/bugs_closed/` is *fixed AND live*. The defect was **"the converter discards
the field"**, and the discarding has stopped — proven in the persisted plan, at the roll
boundary, on the agent this file names. Observing a routed step-level handler is a
**standing watch on the blast radius**, not a remaining instance of the defect. Watch for
a rise in `severity='error'` (routed) rows displacing `fatal` ones, and in orchestrations
carrying `collected_data.__step_error` — per the guardian's surviving advisory objection,
that rise **is the fix working**.

**Per-handler review of the ten disabled handlers is still owed** (unchanged from above) —
each author declared `complete` for a reason and the three bookkeeping ones are probably
right. Re-enable individually by renaming `error_step_disabled_086` back. That is a config
decision, not a defect.

---

# PER-HANDLER REVIEW 2026-07-28 — the owed audit, discharged

The review the owner's ruling made a condition of disabling all ten. Verdicts below;
the short answer is **7 stay disabled, 2 want re-pointing rather than a yes/no, 1 is
another workstream's to decide.** Nothing was changed by this audit — it is a reading.

## Containment re-verified live, first

Seed 220's own three post-checks, re-run 2026-07-28 ~17:0xZ:

```
still_routing_to_complete          0    (want 0)   OK
error_step_disabled_086 recorded  10    (want 10)  OK
other step-level handlers intact  44    (want 45)  DRIFT -1, see loose ends
```

The containment **survived a 181-agent bulk re-seed at 14:25:02.999304Z today** — all
19 handler-carrying agents were rewritten and the ten renames are still in place. That
is worth stating because "config re-seed clobber" is a recorded landmine on this fleet;
here it did not bite.

## The premise moved: 2 of the 10 now carry live traffic

The ruling rested on *"none has fired in 30 days"*. As of today that is no longer true
of two of them. From `orchestration_states.processing_history` (**note:
`execution_path` is dead — 0 of 2225 rows populated — so it is the wrong column to ask**):

| step | runs | first seen | last seen |
|---|---|---|---|
| `image-build-handler.mark_work_item_complete` | 4 | 2026-07-28 12:16:01Z | 2026-07-28 14:42:42Z |
| `image-build-handler.flag_rebuild` | 4 | 2026-07-28 12:16:01Z | 2026-07-28 14:42:42Z |

All four orchestrations `COMPLETED`, `__step_error` NULL, `error` NULL — so **the happy
path ran and the disabled handler was never exercised.** No evidence either way on
whether disabling was right; only that the exposure is now real rather than theoretical.
The other eight have **zero** entries.

> `[UNVERIFIED — flagged, not resolved]` `orchestration_states` retains back only to
> **2026-07-13**, i.e. ~13 days. A "none in 30 days" claim is not answerable from this
> table. The 07-26 figure may have come from a source with longer retention; I did not
> establish which. Anyone repeating the 30-day figure should name its source first.

## The finding the three-way split missed

Reading all 19 agents' step maps rather than the ten steps in isolation: **only ONE of
the seven affected agents has a real error terminal, and it is the one whose handlers
bypassed it.**

`image-build-handler` declares `mark_work_item_failed → complete_error`, and every one
of its five `call_*` steps routes errors there. `mark_work_item_complete` and
`flag_rebuild` alone routed to `complete`. That asymmetry reads as **carelessness, not
intent** — the author had a loud lane and these two did not use it.

The other six agents (`blog-content-planner`, `content-gap-planner`, `site-adoption-agent`,
`spec-updater`, `tool-improver`, `webdesign-agent`) have **no error terminal at all** —
every terminal they own is a `complete_workflow`. For them `error_step: complete` was
the author reaching for the only terminal that existed. There is nothing better to point
at without designing a new lane, which is a change, not a review.

## Verdicts

**Stay disabled — no alternative exists, failing loudly is correct (7)**

| agent :: step | on failure, routing to `complete` would have meant |
|---|---|
| `blog-content-planner :: create_post_pages` | no blog pages created, run reports green |
| `content-gap-planner :: apply_plan` | the gap plan never applied |
| `content-gap-planner :: plan_gaps` | skips `apply_plan`; a green run that planned nothing is indistinguishable from one that found no gaps |
| `site-adoption-agent :: generate_design_intent` | skips `write_design_intent`; adoption completes with no design intent |
| `site-adoption-agent :: write_design_intent` | the spec is never written |
| `spec-updater :: apply_update` | the whole purpose of the agent — merge a field into `site_specs` — silently skipped |
| `webdesign-agent :: fork_theme` | theme never forked |

**Want re-pointing, not a yes/no — owner's call (2)**

`image-build-handler.mark_work_item_complete` and `.flag_rebuild`. Neither should be
re-enabled as they were, and leaving them disabled is second-best. Point them at
`mark_work_item_failed` (which already routes to `complete_error`): the work item gets
marked failed so the immune system can see it, instead of either a green lie or a
generic loud failure that leaves the row claiming success.

`flag_rebuild` is the sharper case. Its action is `flag_page_image_rebuild`; if that
fails silently the imagery is generated and deployed and **the page is never flagged, so
nothing ever references it — which is precisely `bugs_open/114`'s symptom.** Re-enabling
this one as-authored would manufacture 114 on demand. These are also the only two with
live traffic, so this is the one item here with a clock on it.

**Not ours to decide (1)**

`tool-improver.note_refusal`. It sits on the refusal branch (`refuse_mangled_write →
note_refusal`), and its `next_step` and `error_step` were **both `complete`** — the
handler drew no distinction at all. Its success-path twin `append_note` has never had an
`error_step`, so today the two branches are consistent (both fail loudly) and
re-enabling would make them inconsistent. Against that, the note *is* the point of the
refusal branch: "record the refusal on the tool's travelling NOTES so the next agent…".
`scripts/who-owns.py 126` puts `tool-improver`'s refusal path with the **oufe
workstream** (ACTIVE, 50 commits/14d, `bugs_open/126`). Contributed there; not touched
here.

## Loose ends, stated rather than tidied away

- **The 45 → 44 drift is only partly resolved.** On the seven agents seed 220
  snapshotted, the diff pre-update vs live is **nothing lost, one added** —
  `tool-improver.update_component` gained a step-level `error_step → refuse_mangled_write`
  (the oufe thread's 126 work, `6e29d6d19`). So those seven went **+1**, which makes the
  expected total 46, not 45. `[UNRESOLVED]` **two handlers are therefore unaccounted for
  among the other 12 agents**, and there is no snapshot of them from 07-26 to diff, so
  the DB cannot say which. Not chased further.
- **The cheap check that would have settled it, and would settle the next one:** there is
  **no baseline for those 12 agents**. `snapshot_agent(<type>, 'baseline')` on each costs
  one call apiece and makes the next drift a two-table diff instead of an open question.
- **Revert remains available for all ten** — seed 220's snapshots are real and verified
  present: 7 rows in `agent_definitions_backup` with `snapshot_reason LIKE '220_%'`,
  taken `2026-07-26 18:32:26.229Z`, one per affected agent.

## A misstep in this audit, recorded because it nearly became the finding

I first looked for those snapshots in `agent_definitions WHERE is_snapshot` — found
**one row in the entire table**, for none of these agents, and was one step from writing
up "seed 220's safety net does not exist". It does exist. The two-arg
`snapshot_agent(type, reason)` overload that seed 220 actually calls writes to
**`agent_definitions_backup`**, a different table from the one-arg overload, which writes
to `agent_definitions`. Two same-named functions with different destinations.

What caught it was reading the function before believing the query — and before that, the
earlier reflex of demanding a positive control: the first diff query returned "0 rows =
nothing lost", which was **vacuous**, because its snapshot CTE was empty. A query whose
input set is empty answers every question with reassurance. `[Pattern: an EXCEPT/NOT
EXISTS diff needs its baseline COUNTED first, or the null result means nothing.]`

---

# OWNER RULING 2026-07-28 (evening) — the two are re-pointed, and it is LIVE

**Ruling: option 1 — re-point both `image-build-handler` handlers at
`mark_work_item_failed`.** Applied the same evening as seed
`docs/agent_docs/sql_for_agents/259_image_build_handler_error_steps_repointed.sql`.
DB config, so live immediately — no image roll involved.

This discharges the last item the per-handler audit left open. The other eight stay
disabled (seven by the audit's verdict, one — `tool-improver.note_refusal` — still oufe's
to decide via `bugs_open/126`).

## What was checked BEFORE applying, not after

- **The converter really carries step-level `error_step` on the running binary.**
  Pod `agent-chassis-f757fcf65-bg9t7`, `v1.0.1192`, started `18:23:03Z`:
  fix marker `Step declares an error_step that is not in the plan` → **1**;
  positive controls `routeToErrorStepOrFail` → **3**, `error_step` → **7**;
  negative control `zzz_not_a_real_marker_zzz` → **0**. Without this the change would
  have been inert and looked applied.
- **Persisted plans carry it in data, not just in the binary:** 66 step-level
  `error_step` entries across `image-build-handler` `workflow_plan`s created in the
  previous 2 days. (This is also the data-level proof the original write-up listed as
  outstanding.)
- **The target lane works and terminates.** Two orchestrations reached
  `mark_work_item_failed` on 07-28 (`37939c31…`, `b6b45178…`) — both `COMPLETED` at
  `current_step = complete_error`. No cycle: that step's own `error_step` is
  `complete_error`, a terminal.
- **The target step exists** — the seed refuses to run if it does not, because a
  dangling `error_step` is only a `Warn` in the converter and would be silently inert.

## Post-checks, all as predicted

| check | want | got |
|---|---|---|
| the two steps live, pointing at `mark_work_item_failed` | 2 | 2 |
| step-level `error_step → complete` anywhere on the fleet | 0 | 0 |
| still carrying `error_step_disabled_086` | 8 (was 10) | 8 |
| live step-level handlers fleet-wide | 46 (was 44) | 46 |
| pre-change snapshot in `agent_definitions_backup` | ≥1 | 1, `18:50:33.117Z` |
| steps pointing at a step that does not exist | 0 | 0 |

## Two consequences recorded rather than glossed

1. **`flag_rebuild` runs AFTER `mark_work_item_complete`, so its failure path flips the
   work item `complete` → `failed`.** `UpdateWorkItemStatusAction`'s UPDATE is
   unconditional on the current status (`platform/orchestration/actions/v3_site_actions.go:4644-4661`)
   and only the `complete` branch sets `completed_at`, so such a row ends up
   `status='failed'` with `completed_at` populated and `attempt_count` incremented twice.
   That combination is odd-looking on purpose — it is a truer record than a green
   orchestration sitting over unreferenced imagery.
2. **That flip does not regenerate the image.** The claim query takes only
   `triaged`/`approved` (`claim_work_item_action.go:102`), so a failed item parks for
   triage instead of retrying. The reason is not lost either: with no `error_message`
   literal and a non-`complete` status, the action records `__step_error.message` into the
   row's `error` column (`v3_site_actions.go:4599-4617`).

## Still not witnessed

The failing branch has **never** been exercised — all 8 runs on 07-28 completed with
`__step_error` NULL. This is deployed-and-correct-by-construction, not proven in flight.
See `[[verify-the-failing-branch]]`. The watch query:

```sql
SELECT os.orchestration_id, os.status, os.current_step,
       os.collected_data->'__step_error'->>'failed_step' AS failed_step,
       os.updated_at
FROM orchestration_states os
WHERE os.owner_agent_type='image-build-handler'
  AND os.collected_data ? '__step_error'
  AND os.updated_at > '2026-07-28 18:50:33+00'
ORDER BY os.updated_at DESC;
```

A row where `failed_step` is `mark_work_item_complete` or `flag_rebuild`, ending at
`complete_error` with the `site_work_items` row marked `failed` and its `error` populated,
is what closes this.

## A clobber risk found while applying, and blocked

`107_image_build_handler.sql` — the seed that created `flag_rebuild` — still contains
`"error_step": "complete"`, and its section-0 precondition ("expect
`flag_rebuild_exists = f`") is a **printed expectation, not a guard**: the UPDATE runs
regardless. A replay would silently restore the routing that seeds 220 and 259 both
removed. A dated `SUPERSEDED IN PART — DO NOT REPLAY AS WRITTEN` block now sits above
that section. Nothing else in the repo defines these steps: no Go source, no deployment
config, no other seed — checked, so the DB row and the numbered seeds are the whole story.
