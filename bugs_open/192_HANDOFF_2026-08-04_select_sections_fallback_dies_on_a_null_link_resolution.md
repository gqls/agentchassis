# 192 — `page-content-writer`'s `select_sections` step is failing broadly since ~2026-08-03 21:00, unrelated to bugs_open/087

> ## STATUS 2026-08-04 — DIAGNOSED · OUTAGE OVER AND PROVEN · **STILL OPEN, pending the roll**
>
> **Owned by the `bugfix_192_select_sections_wrapper` lane.** Root cause found and
> fixed at source; the reason this stays OPEN is stated so nobody has to guess.
>
> | | |
> |---|---|
> | **Live now** | seed `308` only — a self-retiring third fallback path + the `required` opt-in. **Page builds work again**, proven end to end: item `18bc832c` → `complete`, orchestration `0511e4d1` → `COMPLETED` with `sections_for_render ? 'sections_ready'` **true**, against **false** on the three runs immediately before it. |
> | **Committed, INERT until the next chassis roll** | `2b9d84072` — the unwrap at source, `extract_fields`' opt-in `required`, and the loop's keys-present error. Go changes do nothing here until an image is built and rolled. |
> | **Why still OPEN** | the wrapper is **still produced on every build**; the live seed merely tolerates it. Until the roll the defect is reproducible, which is this repo's stated bar (`CLAUDE.md` → Debugging: *"a fix committed but inert until the next roll stays OPEN"*). |
> | **To close** | after the roll: pod-grep `grep -ac 'required field(s)' /app/agent-chassis` → 1 (0 before) with a long-literal positive control from the same function; confirm `collected_data->'section_plan' ? 'applied'` is **false** and `? 'sections_ready'` **true** on a fresh `page-build-handler` run; then apply a cleanup seed removing path 3 of `select_sections`. |
> | **Known degradation until the roll** | `resolve_links`' `input_mapping` is broken by the same wrapper, so internal CTA resolution is degraded on every build. **Deliberately not shimmed** — `input_mapping` has no ordered fallback, so a shim there would work today and silently re-break on the roll. It self-heals. |
> | **Council** | `Council-Submitted: 7afbf531-5ddd-484e-88c8-091994a0f51f` — verdict not yet read; the trailer asserts nothing, per the standing rule. |
> | **Not this bug** | the overnight `process_sections_loop_iter_N_generate_content` failures (21:00–01:00, ~38 runs) that this filing counted as the same defect. Different step, reachable only *after* this one succeeds, still undiagnosed, nobody on it. |
>
> Working docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_192_select_sections_wrapper/`.
> Full diagnosis at the foot of this file. **The title's "since ~2026-08-03 21:00" is
> wrong** — corrected there.

**Filed 2026-08-04**, discovered while live-verifying `bugs_open/178`'s fix
(`bugfix_154_work_item_routing_columns` lane). **Not yet diagnosed** — this is
a filing with the evidence gathered incidentally, not a completed root-cause.
**A `090` diagnosis run is owed**; not run yet, flagged here so it isn't lost.
*(Run since — and it returned no verdict at all; see the foot of this file.)*

## Symptom

`page-content-writer` orchestrations are failing at `process_sections_loop`
with:

```
step process_sections_loop failed: failed to execute action loop: failed to
get collection at 'sections_for_render.sections_ready': key 'sections_ready'
not found at position 1 in path 'sections_for_render.sections_ready'
```

This is the **exact same error string** as `bugs_open/087`, but NOT the same
cause: 087 is specific to the `page-rebuild` agent, which supplies no
`section_plan` at all, and its own text states plainly *"page-build-handler
... its writer children always carry a real section_plan (26 of 26 recent
runs, all COMPLETED)"* — i.e. the build-handler path was the known-good
control. **This instance is ON the build-handler path.**

## Evidence gathered (incidental, while verifying 178 — not a full diagnosis)

Two orchestrations hit it live on 2026-08-04 ~08:26, dispatched via
`build-dispatch-loop → page-build-handler → page-content-writer` for two
`content_rewrite` work items on `vetcomparison.uk`
(`0883b1aa-d5d6-45ad-a596-df0cc06744ec`, page `guide-cma-compliance`;
`df69efd6-19b7-4788-8fe1-668ea769f3fc`, an unrelated tool page
`tool-gripper-payload-calculator-guide` on a different site — confirming this
is not scoped to one site or one item type).

`select_sections`'s live config (`page-content-writer`, confirmed unchanged
from `bugs_open/087`'s own quote of it):

```json
"select_sections": {
  "action": "extract_fields",
  "config": {"fields": {"sections_ready": [
      "resolved_links.response.link_resolution.sections_ready",
      "input_data.section_plan.sections_ready"]}},
  "next_step": "process_sections_loop",
  "output_field": "sections_for_render"
}
```

On `0883b1aa`, `collected_data.resolved_links` at the time of failure:

```json
{"response": {"link_resolution": {"unresolved": null, "sections_ready": null},
              "resolve_links": {"unresolved": null, "sections_ready": null},
              "input_data": {"site_id": "...", "page_name": "guide-cma-compliance", "page_type": "guide"}},
 "response_status": "complete"}
```

i.e. path 1 (`resolved_links.response.link_resolution.sections_ready`) is
present but explicitly `null`. `ExtractFieldsAction`
(`v3_site_actions.go:4232+`) DOES null-check (`value != nil`) before
accepting a candidate, so on its face this should fall through to path 2 —
and `collected_data.input_data.section_plan.sections_ready` genuinely holds a
full, correct 1-element array at that same point in the SAME row (confirmed
directly: `SELECT collected_data->'input_data'->'section_plan'->'sections_ready'`
on this orchestration returns the real ready section, complete with
`existing_content_html` from the 178 fix, proving THAT part of the pipeline
is healthy). **Why the fallback still doesn't populate `sections_for_render`
is the open question** — `[UNVERIFIED]` whether it's an ordering issue
(select_sections running before `input_data.section_plan` is actually merged
into collected_data, vs. only appearing there later), a second code path,
or something about the map-of-arrays extraction loop not behaving as its
source reads.

**Timing, `[MEASURED]` from `orchestration_states`:** hourly counts of
`process_sections_loop` COMPLETED vs FAILED over the last 3 days show every
run COMPLETED through 2026-08-03 20:00, then FAILED spikes at 21:00 (11),
22:00 (14), 23:00 (12 fail / 1 complete), tapering to 00:00-01:00, quiet
overnight (likely just low traffic, not resolved), then my 2 today at 08:00.
**This predates this session's own chassis roll by many hours** — the fix
this session shipped (178) was pushed as an image at ~20:2x on 08-03 and
only actually deployed by the owner's whole-fleet release the following
morning (pods ~11 min old when checked at 08:2x on 08-04) — so **178's code
cannot be the cause**, and the failure was already live before this session's
image ever ran anywhere. The tree had heavy concurrent activity in the
21:00-23:00 window (many sessions, several image rolls) — no single
suspect commit identified; genuinely not run down.

## Why this matters for other lanes

Roughly half of `page-content-writer` invocations failed outright for a ~4
hour window and it is unclear whether the failure rate has actually dropped
or the quiet overnight period is just low traffic. **Anyone whose work
depends on a content build completing should check this before trusting a
`complete` on `page-content-writer`, and should not assume a `page-build-handler`
COMPLETED status without checking its writer child's own status.**

## Effect on bugs_open/178

**No content was lost** — the failure is upstream of `save_page_sections`
entirely (the workflow never reaches `compile_page`/save), so 178's own
guard-rail concern (silent content loss) does not apply here; this is a loud
failure, not a silent one. But it DOES mean 178's live end-to-end
verification (before/after `content_data` length on a real dispatched item)
could not be completed this session. **What WAS verified**: the
`load_current_section_content` step itself worked exactly as designed —
`section_plan.sections_ready[0].existing_content_html` was populated with
the page's exact current rendered content, matched by slot name, byte for
byte the same live prose. The remaining check (does the writer actually
preserve it end to end) is blocked on this bug, not on 178's own code.

Both parked items (`9e9ec430-ff92-4264-83cc-6072840faad8` still `claimed`,
`18bc832c-c937-4608-9a05-718772d44c88` now `failed` attempt_count=1) are
in a safe, non-terminal-for-long state — `failed` with attempt_count=1 is
not `unresolved`/terminal, so they can retry once this is fixed. Do not
re-dispatch them until 192 is understood; a retry would hit the same wall.

## Fix candidates

Not analysed — this filing is evidence, not a diagnosis. First step for
whoever picks this up: run the `090` trigger
(`./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/090_TRIGGER_needs_diagnosis_v1.sh`)
naming the mechanism above, since the cause is non-obvious and clearly
cross-cutting (hit two unrelated sites, two unrelated item types, in the
same short window).

## How to verify a fix

Re-dispatch either parked item (or any `content_rewrite`/`needs_content_page`
item) via `build-dispatch-loop` for its site, and confirm the
`page-content-writer` child reaches `compile_page`/`COMPLETED` rather than
failing at `process_sections_loop`. Then, separately, re-run 178's own
verification: assert `page_components.content_data` length for the target
slot grows only by the inserted link anchor, not a wholesale replacement.

---

## Third instance, from an unrelated lane: a BRAND-NEW site, on `needs_page`, not `content_rewrite`

**Added 2026-08-04 ~08:30 by the `webdesign_uk_build_service` lane.** Contributing
evidence only. Not diagnosed here, not competing with the owning thread, and the
owed `090` run is still owed.

**Why this instance is worth adding: it removes two possible scopings at once.**
The two instances above are both `content_rewrite` items on **existing** sites.
This one is a **`needs_page`** item on a site that was **created from nothing
minutes earlier**, so:

- **not scoped to `content_rewrite`** — the item type differs;
- **not scoped to existing/adopted sites** — this site had no pages, no history and
  no prior components. It was seeded and submitted fresh through
  `082_submit_domain_unified.sh` today;
- **not a stale-data or migration artefact** — every row involved was written
  today by the current pipeline.

So whatever `select_sections` is failing to find, it is failing on input the
pipeline itself produced from scratch in the same run.

```
site:        webdesign.uk  (created 2026-08-04, first build)
work item:   needs_page / page-build-handler / status=failed
spec:        {"reason":"not_built","page_name":"index","page_role":"landing",
              "plan_id":"4ecaa120-1fa6-4de1-9cd0-39d60c64b729"}
error:       step process_sections_loop failed: failed to execute action loop:
             failed to get collection at 'sections_for_render.sections_ready':
             key 'sections_ready' not found at position 1 in path
             'sections_for_render.sections_ready'
chassis:     v1.0.1247
correlation: a4f05bd6-a548-47a5-8bdb-059e8d75e429  (the submission)
```

**The rest of the cascade ran clean**, which narrows where to look: this is not a
site whose upstream specs were missing. At the time of failure the site had
**12 current `site_specs`** including `identity`, `classification`,
`content_direction`, `design_intent`, `vertical_landscape`, `strategy`, `briefing`
and `resolved_composition`; `needs_domain_research`, `needs_vertical_research`,
`needs_strategy`, `needs_briefing`, `needs_site_plan` and `needs_composition` all
reached `complete`, and several `needs_imagery` items completed too. **Only the
page build failed.**

**One difference from the other two instances that may or may not matter, flagged
rather than asserted [UNVERIFIED]:** this site carries a **seeded, pinned
`evidence_base` with a populated `facts[]` (7) and a large `banned_claims[]` (14)**,
which is unusual — most sites have neither at first build. If the diagnosis ends up
anywhere near the writer's constraint assembly, that is a variable worth excluding
early. I have **not** tested whether removing it changes the outcome, and I would
not read the co-occurrence as a cause: the other two instances have no such spec and
fail identically.

**Impact on this lane, for prioritisation:** webdesign.uk is the shopfront for the
build-service product and **cannot produce its landing page** until this is fixed.
No workaround is being applied here. Hand-building the page is specifically what an
owner ruling now forbids (`CLAUDE.md` → Platform conventions, 2026-08-04), so this
lane is genuinely blocked on 192 rather than merely inconvenienced by it.

---

# DIAGNOSED AND CLAIMED 2026-08-04 ~09:40 BST — `bugfix_192_select_sections_wrapper` lane

**Root cause found. It is `bugs_open/178`'s fix — specifically its `output_field`
reuse, not its logic.** Working docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_192_select_sections_wrapper/`.

> **CORRECTED 2026-08-04 — the onset in this file's title and §Timing is wrong, and
> the error it is built on is a real one worth naming: two different failure modes
> were counted as one.** This file says *"failing broadly since ~2026-08-03 21:00"*
> and records hourly FAILED spikes at 21:00 (11), 22:00 (14), 23:00 (12). Those rows
> are **`process_sections_loop_iter_N_generate_content`** failures — a *different*
> step, and one that can only be reached **after `select_sections` succeeded**. The
> failure this file describes has `current_step = 'process_sections_loop'` exactly,
> and re-measured over the full retained window it appears **only from 08:20 today**:
>
> ```sql
> SELECT date_trunc('hour', updated_at) AS hr,
>        CASE WHEN current_step='process_sections_loop' THEN 'select_sections_miss'
>             WHEN current_step LIKE 'process_sections_loop_iter%' THEN 'iter_generate_content'
>             ELSE current_step END AS mode,
>        status, count(*)
> FROM orchestration_states WHERE owner_agent_type='page-content-writer'
> GROUP BY 1,2,3 ORDER BY 1 DESC;
> --  08-04 08:00 | select_sections_miss  | FAILED    |  3
> --  08-04 01:00 | iter_generate_content | FAILED    |  1
> --  08-03 23:00 | iter_generate_content | FAILED    | 12   <- a DIFFERENT bug
> --  08-03 22:00 | iter_generate_content | FAILED    | 14   <- a DIFFERENT bug
> --  08-03 21:00 | iter_generate_content | FAILED    | 11   <- and 12 COMPLETED the same hour
> ```
>
> **The `21:00` window also had 12 COMPLETED runs**, so "roughly half of
> `page-content-writer` failed outright" was never true of *this* defect. The
> overnight `iter_generate_content` failures are real and still undiagnosed — they
> are **not** this bug and want their own file.
>
> This matters beyond bookkeeping: the wrong onset pointed away from the true cause.
> `08-03 21:00` predates the trigger and made "178 cannot be the cause" look sound.
> The real onset is **08:20 on 08-04**, which is `agent_definitions.updated_at` for
> both `page-build-handler` and `page-content-writer` — i.e. the moment seed
> `299_edit_live_channel_for_content_rewrite_writer.sql` was (re-)applied. Logged in
> `WRONG_CALLS.md`.

## The mechanism

One cause; it kills **both** of `select_sections`' fallback paths, which is why the
file above could not find a surviving route.

1. `plan_sections` writes `collected_data.section_plan` = a **flat** plan
   (`{sections_ready, ready_count, …}`).
2. `load_current_section_content` — 178's new step, inserted by seed `299` between
   `check_has_ready_sections` and `spawn_content_writer` — declares
   **`output_field: section_plan`**, deliberately reusing the key. The seed says so
   in its own header: *"Reuses the `section_plan` output_field plan_sections itself
   uses, so call_content_writer's existing input_mapping needs no change."*
3. **But the action returns a wrapper on every one of its return paths**, including
   all eight it calls "pass-through":
   `{"section_plan": <plan>, "applied": false, "reason": …}` and
   `{"section_plan": <plan>, "applied": true, "matched": N}`
   (`load_current_section_content_action.go:91-97, 158, 173, 211-215`). Its doc
   comment at :73-76 states any non-`edit_live` mode "leaves this action a
   pass-through". **It does not pass through — it re-wraps.**
   `coordinator.go:1859-1861` (`storeActionResult`) stores an action's return value
   **wholesale** under `output_field`, so `section_plan` becomes the wrapper on
   **every page build, in every mode** — which is why this reproduces on
   `content_rewrite` *and* on a fresh-site `needs_page`, exactly as the two
   contributing lanes measured.
4. `call_content_writer` maps `"section_plan": "section_plan"`, forwarding the
   wrapper verbatim as the writer's `input_data.section_plan`. Then:
   - **path 2** (`input_data.section_plan.sections_ready`) misses — the plan is now
     at `input_data.section_plan.section_plan.sections_ready`;
   - **path 1** misses *because of the same wrapper*: `resolve_links`' input_mapping
     is `"sections?": "input_data.section_plan.sections_ready"`, so the link-resolver
     child is handed **no sections** and returns `sections_ready: null`. That is the
     "present but explicitly null" value this file reports — it is a **consequence**
     of the wrapper, not an independent second fault.
5. `ExtractFieldsAction` omits the target key and **returns success**, so
   `sections_for_render = {}`; `process_sections_loop` then fails at
   `loop_actions.go:144/751` naming the missing key rather than the failed extraction.

**Measured, both shapes, live** (`orchestration_states`, 2026-08-04):

```
orch     | current_step          | section_plan keys             | direct_len | nested_len
3b692317 | process_sections_loop | applied,reason,section_plan    |     0      |     7
0883b1aa | process_sections_loop | applied,matched,section_plan   |     0      |     1
df69efd6 | process_sections_loop | applied,matched,section_plan   |     0      |     1
4edcdaeb | ..._iter_1_generate.. | deferred_count,…,sections_ready|     3      |     0   <- flat, pre-seed
```

`sections_for_render` on all three failing rows is `{}`.

## Answering this file's own open question

It asks, `[UNVERIFIED]`, whether the fallback fails because of *ordering* — the step
running before `input_data.section_plan` is merged. **It is not ordering.** The data
is present the whole time; it is one level deeper than the configured path. The
`SELECT collected_data->'input_data'->'section_plan'->'sections_ready'` quoted in the
original evidence returned a real array because that query was run against a row
whose shape it did not check — the same query returns `0` rows-worth of array on the
failing rows, and the plan is at `->'section_plan'->'section_plan'->'sections_ready'`.
Recorded in `WRONG_CALLS.md`: **a query that reads the path you expect cannot tell you
the shape changed underneath it — enumerate the keys.**

## `090` diagnosis loop

Filed per the CLAUDE.md default for a cross-cutting claim (intake
`b45144fa-2051-4ff8-9318-e052a9c3a084`, run `aea3cc68-b274-4df4-a1c1-f60ba47bf09e`).
The diagnosis above is **first-hand** — code read plus live rows in both shapes plus
a fleet census — and did not wait on it; the loop's verdict is recorded in this
lane's NOTES when it lands, whichever way it goes.

## Fix

See the lane's `PLAN_2026-08-04_select_sections_wrapper.md`. Four changes, ordered by
what closes the door: the action returns **the plan itself** on every path (Go);
`extract_fields` gains an **opt-in `required`** list so a target that resolves on no
path fails the step naming the cause, default OFF (Go); the loop's path-miss error
lists the keys actually present (Go); and a config seed that adds a third fallback
path **and** opts `select_sections` into `required` — live immediately, unblocking
builds on the binary already running.
