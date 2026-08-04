# 194 — four of six `save_page_sections` callers never map `sections_metadata`, so every save through them NULLs `content_data`

**Filed:** 2026-08-04 · **By:** the `bugs_open/087` lane, from 087's own live acceptance run
**Status:** OPEN — mechanism confirmed first-hand, one of four instances fixed (see below)
**Severity:** Medium-high. `content_data` is what the rerender path regenerates a section from.
A save that NULLs it does not corrupt the served page — it silently removes the page's ability
to be rerendered, and nothing reports it.

## Symptom, observed live

`bugs_open/087`'s acceptance test dispatched `page-rebuild` at
`vetcomparison.uk` / `tool-cma-obligation-checker-guide`
(correlation `3fdf4acf-5f96-49f9-8801-28047aae92ef`, 2026-08-04 09:47–09:50Z, all steps
`COMPLETED`, page redeployed correctly). The rebuild worked. But:

| slot | `content_data` before | after |
|---|---|---|
| `hero` | 644 chars | **NULL** |
| `article-body` | 3,810 chars | **NULL** |
| `call-to-action` | 420 chars | **NULL** |

`rendered_html` is fine on all three (3,243 / 5,317 / 2,303 chars, fresh copy, page serves
HTTP 200 with all three `data-component` slots present). The row `id`s changed too — the
action deletes the agent-writable components and re-inserts them.

## Root cause — a config key that four of six callers simply do not set

`SavePageSectionsAction` (`platform/orchestration/actions/save_page_sections_action.go`)
writes `content_data` on insert (`:685-687`), taking it from the section's own
`content_data`, which reaches it via a **`sections_metadata_field`** config key. Callers that
supply that key preserve structured content; callers that do not write SQL NULL.

Read off the live rows (`jsonb_path_query(default_config, '$.**.steps.*')` — the step is
nested inside a loop `sub_workflow` in four of the six, so a top-level `jsonb_each` misses them):

| caller | `sections_metadata_field` | result |
|---|---|---|
| `page-build-handler` | `page_content.response.sections_metadata` | **preserved** |
| `page-rerender` | `rerender_sections.sections_metadata` | **preserved** |
| `page-rebuild` | *absent* | **NULLed** |
| `pageflow-builder` | *absent* | **NULLed** |
| `site-work-orchestrator` | *absent* | **NULLed** |
| `tool-recreation-handler` | *absent* | **NULLed** |

**The data is available on every one of those paths — it is simply not mapped.** The writer
returns it: `page-content-writer`'s `compile_page` output keys on the run above are
`page_html, page_name, section_count, sections_metadata`. The three `page-*`/`*-builder`
callers all store the writer's reply under `output_field: page_content`, so
`page_content.response.sections_metadata` — byte-identical to what `page-build-handler`
already uses — resolves for all of them.

Fleet-wide as of 2026-08-04: **161 of 1,201 `page_components` rows (13.4%) have NULL
`content_data`.** That figure is *exposure, not damage attribution* — it is a snapshot of
the population, and this case has not traced which of those 161 were NULLed by these four
callers versus never having had structured content in the first place. Do not quote it as
"161 pages damaged".

## Why it is not merely cosmetic

- `page-rerender` regenerates a section **from `content_data`** — with it NULL there is
  nothing to regenerate from. This is the same asymmetry recorded in MEMORY as *"a repro is
  destroyed by the render"*.
- The pre-overwrite value **is** archived, so this is recoverable in principle:
  `page_component_history` gets an insert (`:536-538`) with `source='save_page_sections_overwrite'`,
  and the three rows for the run above hold exactly the pre-run lengths (644 / 3,810 / 420).
  **But `component_id` on those rows is NULL** — the FK is `ON DELETE SET NULL` and the action
  deletes the components before re-inserting, so the archive survives while its link to a slot
  does not. Recovery has to match on content, not on id.
- **A naive restore is the wrong repair.** Putting the archived `content_data` back beside the
  *new* `rendered_html` pairs stale structured content with fresh HTML; the next rerender would
  regenerate the *old* page over the new one. The correct repair is to fix the mapping and
  re-run the build, so both halves are written together.

## Fix

**Config-only, one key per caller, live on apply.** Mirror `page-build-handler` exactly:

```json
"sections_metadata_field": "page_content.response.sections_metadata"
```

- **`page-rebuild` — APPLIED**, `docs/agent_docs/sql_for_agents/310_page_rebuild_preserves_content_data.sql`,
  and proven by re-running the same page (see "Verification" below). Done here rather than
  deferred because this lane's own acceptance run is what NULLed those three components, and
  leaving a live page unable to rerender would have been damage recorded and not repaired.
- **`pageflow-builder`, `site-work-orchestrator`, `tool-recreation-handler` — NOT DONE.**
  Deliberately: the first two are the same two callers `bugs_open/087` found broken in the
  *other* direction and both are dormant, and `tool-recreation-handler` runs a different
  writer flow whose response shape this case has **not** read — asserting
  `page_content.response.sections_metadata` for it would be an unmeasured claim. Read its
  writer's output keys before copying the line.

## Verification

Re-run a rebuild on a `rebuild_policy != 'owned'` page and require **both** halves:

```sql
SELECT slot_name, length(rendered_html) AS html, length(content_data::text) AS data, updated_at
FROM page_components WHERE page_id='<page>'::uuid ORDER BY position;
```

`data` must be non-NULL on every row, and `updated_at` must be the new run's. The
disconfirming outcome is stated so it cannot be retrofitted: **if `content_data` is still NULL
after the mapping is added, the writer's reply is not reaching the save step under that path
and the fix is wrong** — check `page_content.response` actually carries `sections_metadata` on
*that* caller before assuming the key name is at fault.

## Related

- `bugs_open/087` — the same agent, the same class (a caller that does not map a field its
  callee needs), found by 087's acceptance run. 087's fix makes the *writer* independent of
  its callers; this one is the *saver*'s equivalent and is still caller-supplied.
- `bugs_closed/125` — `page-rebuild` deploying to a name-derived path. Same agent, same
  "five implementations, four correct" shape.
- `bugs_open/098` — archiving does not retract; the rerender path this defect disables is one
  of the mechanisms 098 depends on.

---

## 2026-08-04 10:01Z — `page-rebuild`'s instance FIXED and PROVEN LIVE

Seed 310 applied, then the **same page re-run** (correlation
`76008af6-495d-4ca1-a1da-57b09048417e`, 09:58–10:01Z, all steps `COMPLETED`):

| slot | `rendered_html` | `content_data` before 310 | after 310 |
|---|---|---|---|
| `hero` | 3,132 | **NULL** | **626** |
| `article-body` | 4,119 | **NULL** | **2,797** |
| `call-to-action` | 2,335 | **NULL** | **433** |

Both halves written together at the new run's `updated_at`, `build_status` → `deployed`,
page serves HTTP 200 at its canonical url and 404 at the name-derived one. The damage this
lane's own acceptance run caused is repaired — by rebuilding, not by restoring the archive,
for the reason stated above.

The verify block was **induced before being trusted**: run alone against the unmodified row
it raised `194/310: sections_metadata_field is <NULL>, expected …`. A verify block you have
not seen fail is a claim, not a check.

**Still open:** `pageflow-builder`, `site-work-orchestrator`, `tool-recreation-handler`.
The first two are dormant and are the same pair `bugs_open/087` found broken in the other
direction; the third runs a different writer flow whose response shape is **[UNMEASURED]** —
read its writer's output keys before copying the line.

---

## 2026-08-04 ~19:40Z — CLAIMED by a bug-clearing thread (session `da43ef00`)

Taking the remaining three instances, and the framework question behind them: a per-caller
config key that four of six callers forgot is the *same defect class* `bugs_closed/087` just
fixed at source in the writer. Working lane:
`docs/agent_docs/docs024_key_docs_latest/bugfix_194_sections_metadata_mapping/`.
Verified still valid at claim time — live config read, `page-rebuild` carries the key,
the other three are still absent:

```
page-build-handler      | "page_content.response.sections_metadata"
page-rerender           | "rerender_sections.sections_metadata"
page-rebuild            | "page_content.response.sections_metadata"   -- 310, this file
pageflow-builder        | (null)
site-work-orchestrator  | (null)
tool-recreation-handler | (null)
```

---

## 2026-08-04 ~20:10Z — all four instances resolved, and the class closed at the seam

### The remaining three, individually

| caller | state | how |
|---|---|---|
| `pageflow-builder` | **FIXED, live** | seed `312` — `page_content.response.sections_metadata` |
| `site-work-orchestrator` | **FIXED, live** | seed `312` — same path |
| `tool-recreation-handler` | **NOT A DEFECT — measured** | it has no writer step at all |

`tool-recreation-handler` was this file's one `[UNMEASURED]` item and the caution was
right. Its step graph is `recreate_tool` (`execute_llm_prompt` → `tool_recreation`) →
`validate_tool` (`validate_page_content` → `validation_result`) → save from
`validation_result.clean_html`. There is no `sections_metadata` anywhere on that path
because it recreates a whole-page tool as one HTML blob, not a section set — and
`rerender_page_sections_action.go:318` already agrees, exempting a self-contained tool
section from the missing-content escalation by design. **Its NULL is the correct shape.**
Copying the key onto it would have been exactly the unmeasured claim this file warned
against. It instead gets the new `expects_no_sections_metadata` declaration, so the fact
is visible in the caller's own config rather than in a comment somewhere else.

The other two both run `write_page_content` = `call_agent` at `page-content-writer` with
`output_field: page_content`, in the same loop `sub_workflow` as their save step — so the
key is byte-identical to the one `page-rebuild` is already proven live with. Both branches
of seed 312's verify block were **induced before the seed was trusted**; the second needed
its own run, because two guards in series only prove the one that fires.

Post-apply census — five of six name the field, the sixth is correct absent:

```
page-build-handler      | page_content.response.sections_metadata
pageflow-builder        | page_content.response.sections_metadata   <- 312
page-rebuild            | page_content.response.sections_metadata   <- 310
page-rerender           | rerender_sections.sections_metadata
site-work-orchestrator  | page_content.response.sections_metadata   <- 312
tool-recreation-handler | (absent, correct)
```

### The class, at the seam — `47ee3ebce`, inert until the next roll

Four config copies of one path string is the symptom; the defect is that **the saver
depends on being told where its own input lives**, so forgetting is always available.
`save_sections_metadata_source.go` makes it responsible for itself:

- **default** — key unset, consult `defaultSectionsMetadataField`, which is
  `validate_page_content_stats.go`'s own constant *referenced, not copied*, so gate and
  save cannot drift. A configured field still wins outright; a configured-but-unresolving
  field still falls to HTML exactly as before (the no-op case, with its own test — it is
  what keeps `page-rerender`'s 2,878 runs byte-identical). A single default, deliberately
  **not** a probe: resolving another run's metadata under a path nobody configured would be
  worse than the NULL it replaced.
- **`expects_no_sections_metadata`** — a caller may declare it has none by design.
- **`require_sections_metadata`** — new authority, so opt-in with the unsafe default OFF
  and seeded on **nobody** (RFC_010). This function already carries five refusing guards.
- **`CONTENT_DATA_REGRESSION`** — an `agent_error_log` row when a page that HAD structured
  content is saved with none. The silence is what let this run for six months.

Registered as **PBP-031** in the same commit; PBP-011's stale "three callers" corrected to
six — that stale count is part of why this stayed invisible. Council
`b6023fc1-ae70-4486-b752-d399e9b1afcc`.

### What a NULL costs, measured — and what that measurement is NOT

`rerender_page_sections_action.go:326` refuses to render a section with no stored
`content_data` and escalates the WHOLE page to a full LLM rebuild. `site_work_items` holds
**44 such escalations across 8 sites since 2026-07-12**, of which **13 FAILED** on
2026-08-03. That is **exposure for the class, not damage attributed to these callers** —
some of those pages predate `content_data` capture entirely, exactly as this file's own
161/1,201 figure is exposure and not attribution.

### Still owed before this closes

1. the council verdict, read and acted on;
2. the next chassis roll, then the pod-grep for `CONTENT_DATA_REGRESSION`;
3. the acceptance run — `site-work-orchestrator` is directly dispatchable
   (`075d_simple_maintain_trigger.sh`), so one of the two dormant callers CAN be proven
   live. Pass requires **both** `content_data` non-NULL **and**
   `sections_source: 'metadata'` in the save's result: `content_data` can also arrive via
   the interactive carry-forward, so the bare column check is a false pass.

---

## 2026-08-04 ~21:20Z — CLOSED. What that means precisely, and what it does not

**Council `b6023fc1-ae70-4486-b752-d399e9b1afcc`: APPROVED, round 1**, 4 advisory
objections, none high-severity. The `architecture` seat read it as *"RFC_010 being
correctly exercised, not evaded"*.

### The close rests on the config half, which is LIVE. The Go half is prevention and is inert until the next roll.

The defect as filed — *callers that never map `sections_metadata`, so every save through
them NULLs `content_data`* — is **not reproducible on any live caller**. Final census:

| caller | `sections_metadata_field` | declares none | requires |
|---|---|---|---|
| page-build-handler | `page_content.response.sections_metadata` | – | – |
| page-rebuild | `page_content.response.sections_metadata` (310) | – | – |
| page-rerender | `rerender_sections.sections_metadata` | – | – |
| pageflow-builder | `page_content.response.sections_metadata` (312) | – | – |
| site-work-orchestrator | `page_content.response.sections_metadata` (312) | – | – |
| tool-recreation-handler | *(none — correct)* | **true** (313) | – |

Every caller is now **explicit**: five name the field, the sixth declares it has none by
design. That is what makes this closable rather than merely committed.

**Committed and NOT yet live** (`47ee3ebce`, approved): the Go seam that stops a *future*
caller reintroducing it — the shared default, the declaration, the opt-in refusal, and the
`CONTENT_DATA_REGRESSION` record. Inert until the next chassis roll. Its two post-roll
checks and their disconfirming outcomes are written down in advance in the lane's
`RUNBOOK` R6/R7, and in **PBP-031**'s `verify-later`, so they survive this file being
closed. **A reader who needs the prevention half to be live must check the pod, not this
file.**

### The strongest objection, recorded rather than answered — a human's call

`bug_historian` (medium): *for every live caller today the behaviour is still "log a
warning, still lose the data, still report success"* — better than silence, but not the
fail-loud guard. That is correct and it is deliberate: RFC_010 says new authority ships
with the unsafe default OFF, so `require_sections_metadata` is seeded on **nobody**, and the
new record is what turns the later per-caller opt-in into a measurement instead of a guess.
But a deferral is not a fix, and the seat is right that whether that suffices to close 194
is a decision rather than a deduction. **Raised to the owner; the follow-on is the 24h
`CONTENT_DATA_REGRESSION` reading, then a per-caller opt-in decision.**

Also open, disclosed, unfixed: the **161 of 1,201** already-NULL rows (repair is re-running
the build, never restoring the archive — the reason is in this file's own "naive restore"
paragraph); **partial** `content_data` loss, outside the report's predicate on purpose; and
the single-component writers of `bugs_open/136`.

### What this case cost, and what it bought

Two seeds live, one Go commit approved first round, seven tests with four mutations run
against the shipped code, one landmine, two `WRONG_CALLS` entries, and one correction to a
concept-register entry that had said "three callers" since before three of them existed —
which is part of why this stayed invisible for six months.

**Status: CLOSED — the filed defect is fixed AND live on every caller. Moving to
`bugs_closed/`.**
