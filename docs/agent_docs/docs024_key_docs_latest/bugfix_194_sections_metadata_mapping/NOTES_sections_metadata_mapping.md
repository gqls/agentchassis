# NOTES — `bugs_open/194`, `save_page_sections` and the caller-supplied metadata path

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-08-04 ~19:35Z — picking the bug, and how I established it was free

Concurrency check before claiming (CLAUDE.md: "check other threads are not already looking
at this bug"), in three layers because each is blind in a different way:

1. `scripts/who-owns.py 194` — **advisory and LAGGING**: it reads commits, so a session
   mid-fix with nothing committed is invisible to it. Said `likely OWNING workstream(s): (none
   identified)`.
2. **Live `.jsonl` transcripts** — the layer that sees uncommitted sessions. 12 sessions
   modified in the last 4 hours mention 194; I read the CONTEXT of each mention rather than
   counting them, because every session that runs `ls bugs_open/` gets a hit per file and a
   raw tally is mostly listing noise:
   - `444c951d` (the `087`/`193`/`195` lane) cites 194 as the case it *spun out* — 41 mentions
     of `save_page_sections`, but as the closing note of 087, not as work in flight;
   - `96678671` was *offered* 194 by its operator ("possibly bugs_open/194?") and picked
     `189` instead — it is mid-fix on `SiblingSignatures` with a lane dir already created;
   - `cf156d11`, `86c1f378` — 194 appears only inside a `bugs_open/` listing or a `who-owns`
     dump. `86c1f378` is a sibling bug-clearing thread still surveying (candidate list
     includes 194), which is why I claimed in a commit rather than after finishing the work.
3. **A claim commit before any research** (`1dc0ddcb7`), so the sibling survey sees it.

**The measurement that could have come out otherwise:** if the transcript grep had shown a
session with `sections_metadata_field` in its recent tool calls, I would have dropped 194.
It showed 2 mentions in `444c951d` — both inside the 087 close-out text.

## Validity re-check at claim time (the bug is 12 hours old, but the tree moves hourly)

Census re-run live, not read off the bug file. Six `save_page_sections` steps fleet-wide,
all with step key `save_sections`:

```sql
SELECT ad.type, s.key, s.value->'config'->>'sections_metadata_field', s.value->'config'->>'html_field'
FROM agent_definitions ad,
LATERAL jsonb_path_query(ad.default_config, '$.**.steps') AS steps,
LATERAL jsonb_each(steps) AS s(key,value)
WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
  AND s.value->>'action' = 'save_page_sections';
```

| agent | `sections_metadata_field` | `html_field` |
|---|---|---|
| page-build-handler | `page_content.response.sections_metadata` | `validation_result.clean_html` |
| page-rebuild | `page_content.response.sections_metadata` (seed 310) | `assembled_page.html` |
| page-rerender | `rerender_sections.sections_metadata` | *(none)* |
| pageflow-builder | **absent** | `assembled_page.html` |
| site-work-orchestrator | **absent** | `assembled_page.html` |
| tool-recreation-handler | **absent** | `validation_result.clean_html` |

Still valid: three instances unfixed. `$.**.steps` matters — the step is nested inside a loop
`sub_workflow` in four of the six and a top-level `jsonb_each` misses them.

## The bug file's one [UNMEASURED] item, now MEASURED

`tool-recreation-handler`'s writer flow. Its step graph has **no writer call at all**:
`recreate_tool` (`execute_llm_prompt` → `tool_recreation`), `validate_tool`
(`validate_page_content` → `validation_result`), saving from `validation_result.clean_html`.
There is no `sections_metadata` anywhere on that path — it recreates a whole-page tool HTML
blob, not a section set. **So a NULL `content_data` there is not this defect**; copying the
key onto it, as the bug file warned, would have been an unmeasured claim. The other two
(`pageflow-builder`, `site-work-orchestrator`) both have `write_page_content` = `call_agent`
with `output_field: page_content` in the same loop sub-workflow as `save_sections`, so
`page_content.response.sections_metadata` resolves for both — the bug file's prediction holds.

## Dormancy — and why `orchestration_states` cannot answer it

`orchestration_states` retains ~1 day of `COMPLETED` rows (`min(created_at)` per status:
COMPLETED 2026-08-03, i.e. yesterday), so "0 runs in the window" there means "none since
yesterday", not "dormant". Sourced instead from `agent_run_stats`, which has no reaper and
spans 2026-07-26 → 2026-08-04 across 95 agent types:

| agent | run_count | last_ran_at |
|---|---|---|
| page-rerender | 2878 | 2026-08-04 |
| page-build-handler | 283 | 2026-08-04 |
| page-rebuild | 4 | 2026-08-04 |
| tool-recreation-handler | 3 | 2026-08-02 |
| **pageflow-builder** | **absent** | — |
| **site-work-orchestrator** | **absent** | — |

The disconfirming shape: if `agent_run_stats` only tracked leaf agents, an orchestrator's
absence would be an artefact of the table rather than a fact about the fleet. It does not —
`build-dispatch-loop` (2352), `content-feed-orchestrator` (161),
`med-price-scrape-orchestrator` (36) are all in it. So the absence is meaningful:
**the two callers a Go-side default would newly re-route are both dormant over 9 days.**
That is good for blast radius and bad for provability — a live run cannot prove them, so
whatever ships needs offline proof that can fail.

## 2026-08-04 ~20:05Z — seed 312 applied: the two callers that HAVE the data now map it

`docs/agent_docs/sql_for_agents/312_pageflow_and_site_work_orchestrator_preserve_content_data.sql`,
snapshot taken first, both `UPDATE 1`, verify block `NOTICE ... OK`. Post-apply census: five of
six callers name the field; `tool-recreation-handler` is the sixth and is correct absent.

**Both branches of the verify block were INDUCED before the seed was trusted**, run alone
against the unmodified rows:

```
ERROR:  194/312: pageflow-builder sections_metadata_field is <NULL>, expected page_content.response.sections_metadata
ERROR:  194/312: site-work-orchestrator sections_metadata_field is <NULL>, expected page_content.response.sections_metadata
```

The second one needed its own run: the two checks are in series inside one `DO`, so the first
`RAISE` aborts the block and the site-work half would never have been exercised. **A verify
block with two guards in series has only been proven for the guard that fired** — inducing the
first one is not evidence about the second, which is the same "a mutation that passes may have
hit a guard in series" shape recorded in MEMORY, seen here in the checking code rather than in
the code under test.

## Why this seed is NOT the fix, only half of it

Seed 312 is the third and fourth hand-written copy of one key (055/065 page-build-handler, 034
page-rerender, 310 page-rebuild, now 312 ×2). The defect is not that four callers forgot a
line; it is that **a saver depends on being told where its own input lives**, so forgetting is
always available. The Go half follows.

## 2026-08-04 ~21:12Z — council APPROVED round 1 (`b6023fc1`), and what I did about the objections

`approved with 4 advisory objection(s) — none high-severity`. 12 seats ran, 5 abstained.
The **architecture** seat approved with one low objection and read the change as
*"RFC_010 being correctly exercised, not evaded"* — which was the ruling I most needed
tested, since a shared-seam change arriving inside a bug patch is exactly what drew the
REJECTED verdict on `bugs_closed/124`.

Objections are advisory. I acted on the four that were cheap and made the change or its
evidence better, and recorded the rest as open questions rather than pretending they were
answered.

**1. `prior_art_librarian` (medium) — "seed 312 is 'already applied, live' … attach the
check before accepting the zero-blast-radius claim."** Right to insist, and note the shape:
the *file existing* proves nothing about the *row*. Final census, live, all six callers:

```
page-build-handler      | page_content.response.sections_metadata | -    | -
pageflow-builder        | page_content.response.sections_metadata | -    | -
page-rebuild            | page_content.response.sections_metadata | -    | -
page-rerender           | rerender_sections.sections_metadata     | -    | -
site-work-orchestrator  | page_content.response.sections_metadata | -    | -
tool-recreation-handler | (none)                                  | true | -
                          ^ sections_metadata_field                 ^ expects_no  ^ requires
```

Five name the field, the sixth declares it has none, `require_sections_metadata` is on
**nobody** — so the new default is consulted by zero live steps and the new refusal is
reachable by nothing, both now readable off the schema instead of off my prose.

**2. `bug_historian`'s "missing" item — should `tool-recreation-handler` be declared, or
"left to drift the way `sections_metadata_field` itself drifted since 2026-02-18?"**
Seeded: `313`, applied, `expects_no_sections_metadata: true`. It changes no behaviour (the
default structurally cannot resolve on that caller either way) — it changes what a reader
of the config can tell without reading two Go files. **Both branches of its verify block
induced**, the second on a synthetic config rather than the live row: `expects_no… is
<NULL>, expected true`, and `declares BOTH expects_no_sections_metadata and
require_sections_metadata`.

**3. `guardian` (low) — "confirm the new COUNT(*) is indexed/bounded rather than taking it
on the author's word."** Measured: `idx_page_components_page` on `(page_id)` exists, and
the per-page row count is **max 8, mean 2.3, across 520 pages**. So the added query is an
index scan over at most 8 rows.

**4. `debug_historian` (medium) — the post-deploy check tests BEHAVIOUR, not
DEPLOYMENT.** Correct, and it is the fleet's most-repeated mistake. Named explicitly now,
in the bug file and RUNBOOK R6: positive control `strings /app/agent-chassis | grep -c
'loses the only thing the'` (a long literal from the new message — short ones compile to
immediates and read 0 on a binary that fully supports them), **plus a discriminating
negative** — `grep -c 'CONTENT_DATA_REGRESSION_V2'`, which must return 0. My change removes
no string, so there is no natural negative control; a near-miss literal is the substitute,
and it is what proves the grep discriminates rather than proving the pipeline.

### The objections I did NOT close, and why they stay open rather than answered

- **`guardian` (medium): seed 312 shipped before the gate ran.** True. The owner ruling of
  2026-07-29 §2 is the answer and I will not dress it up as more than it is: *"review here
  is after the fact, by design … do not claim an ordering constraint you do not have; do
  not pretend you could have waited."* HEAD is shared and DB config is live on apply, so
  there was no version of this that waited. The seat is right that it should be *named*,
  which is what this paragraph is. Its labelling point is fair too — seed 312/313 are
  `config_change` edits filed under `operation: add` because they are new files; the
  operation vocabulary makes a file-shaped and a config-shaped edit indistinguishable.
- **`bug_historian` (medium): for every live caller today the behaviour is still "log a
  warning, still lose the data, still report success" — an improvement on silence, not the
  fail-loud guard.** This is the sharpest objection in the round and it is correct. The
  answer is deliberate (RFC_010: the refusal ships OFF, and the record is what makes the
  opt-in a measurement instead of a guess) but it is a *deferral*, not a fix, and the seat
  is right that a human should decide whether that closes 194. Surfaced to the owner.
- **`reuse_agent` (low): a FOURTH `content_data`-adjacent error code** in an area a landmine
  already calls fragmented. Justified in the file header, not re-litigated here.
- **`render_guardian` / `bug_historian` (low/medium): PARTIAL content_data loss is
  unreported.** Disclosed in the submission as risk 5, still true, still out of scope.

---

## 2026-08-05 ~10:20–10:45Z — the two owed checks, and a correction that stops a false pass

### 3a, interim (the 24h read is NOT due until 2026-08-06 09:10Z)

Roll time confirmed at the artefact, not the tag: both `agent-chassis` replicas report
`startTime` **2026-08-05T09:10:19Z / 09:10:44Z** on `v1.0.1252`.

`CONTENT_DATA_REGRESSION`: **0 rows** [MEASURED 10:16Z]. Positive control in the same run —
since 09:10Z the same table carries `PROCESSING_FAILED` 138, `LLM_API_ERROR` 116, `TIMEOUT`
48, `UNKNOWN` 6, `VALIDATION_ERROR_DROPPED` 4. So the table is being written and the zero is
not an empty-table artefact.

> **But the zero is very nearly VACUOUS, and I am recording it as such rather than as an
> early pass.** Post-roll runs of the callers under test, counted from `orchestration_states`
> (24h retention easily covers a 1h window; column is **`owner_agent_type`**, not
> `agent_type`): `page-rerender` **1**, `page-build-handler` **0**. Against a pre-roll rate of
> ~320 `page-rerender` runs/day, one run is not a sample. `agent_run_stats.last_ran_at` for
> `page-build-handler` is **04:24Z — before the roll**, so that caller has not executed the
> new binary at all. §3a's own warning ("a zero is only meaningful if the callers actually
> RAN") is still the operative fact. Re-read tomorrow ≥09:10Z **and check the run counts in
> the same breath**.

One row needed ruling out: `CONTENT_DATA_ENVELOPE` ×1 at 09:59:20Z, one second after
`page-rerender`'s last run. **Not ours** — it is `bugs_open/190`'s guard
(`content_data_envelope_guard.go:142`), and it reports a *successful* lossless decode. Name
similarity only.

### 3b — the acceptance target list was wrong, and the site the owner offered cannot serve

Owner offered `ai-agent-orchestration.com` (unprompted: "it needs a rebuild anyway, there is
a lot missing on it"), which resolves §3b's "ask before dispatching" blocker on the *spend*
question. It does **not** resolve the *route* question, and the route is where this failed.

Traced the live `site-work-orchestrator` graph before dispatching (R2) rather than trusting
the handoff's trace. The build branch is
`load_work_items → check_has_items → build_items_loop{ write_page_content(output_field:
page_content) → review → assemble_page → deploy_page → save_sections → update_page_status }`,
and the loop's `save_sections` config is `page_content.response.sections_metadata` — seed 312
exactly. So the route is right in shape.

**Then read the loader's predicate instead of assuming it.** `load_work_items` requires
`status IN ('triaged','approved')`; the step adds `pipeline='build'` and
`handler_agent='page-content-writer'`. Measured:

| site | handoff said | real predicate |
|---|---|---|
| `ai-agent-orchestration.com` | 6 | **0** |
| `mortgagecalculator.co.uk` | 13 | **1** |
| the other five listed | 7–17 | **0** |

`ai-agent-orchestration.com` fails on **two independent clauses**: not one of its items
carries `handler_agent='page-content-writer'` (they are `page-rerender` ×42,
`page-build-handler`, `rerender-pages`, `tool-auditor`, `webdesign-agent`, …), and none is in
`triaged`/`approved`. A maintenance run there is a **guaranteed** vacuous pass. Full write-up
in `WRONG_CALLS.md`; the runnable predicate is now RUNBOOK **R7 item 3**.

Scale of the constraint: `page-content-writer` has held **14 work items fleet-wide in all of
history** — 12 `failed`, 1 `complete`, 1 `triaged`. That last one is the only qualifying item
that exists.

**Third route found, and costed:** `pageflow-builder` — the *other* seed-312 caller — is
spawned by **nothing in `agent_definitions`** (censused `agent_type='pageflow-builder'` /
`target_role LIKE '%pageflow%'`: no rows). It is reached only by the new-build intake flow.
⚠ That flow is `hitl_mode: interactive`, i.e. the **Layer-2 carry-forward** path the PASS
criteria call out as the reason a bare `content_data IS NOT NULL` check is a false pass — so
on that route only `sections_source: 'metadata'` discriminates.

### Read-only state of `ai-agent-orchestration.com` (for the owner's rebuild question)

- **31 of 106 `page_components` have `content_data` NULL**, across 10 pages — and on **9 of
  those 10 it is every component on the page** (3/3, 3/3, 5/5 …), not a partial loss. `blog`
  is the only partial (2/3). This site is a live instance of 194's damage class.
- **5 pages have no components at all**: `agent-complexity-estimator-guide`,
  `ai-readiness-quiz-guide`, `tool-llm-cost-calculator-guide` (all `build_status=needs_rebuild`),
  plus `llm-cost-calculator` and `roi-estimator` — the latter two marked **`deployed`** while
  holding nothing.
- **42 queued `page_rerender` items** (21 `detected`, 21 `unresolved`) that are not moving.
  **[MEASURED, and deliberately NOT asserted as the cause]** the overlap with the NULL pages
  is **partial**: 10 of 21 detected and 9 of 21 unresolved sit on a page with a NULL
  component. Half of the stuck queue is explained by the NULLs; half is not, and I have not
  established why.
- **Direct evidence the class bites this site**, not inference: two `needs_page` escalations
  whose summary is literally *"a section had no stored content_data"* —
  `containment-first-architecture` (07-31, still `needs_human_review`) and `services` (complete).
- 32 pages `generic` / 6 `owned`, so the `owned` guard blocks little here.
- 6 items sit claimed by `build-dispatch-loop` but parked at `needs_human_review`, claimed
  08-04 08:37 and earlier — **not in flight**; last site orchestration was 08:36Z.
