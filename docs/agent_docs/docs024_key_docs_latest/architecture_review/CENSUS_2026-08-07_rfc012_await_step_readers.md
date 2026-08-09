# CENSUS — readers of an awaited step's key (RFC_012 §3(a) / (a′))

**Commissioned** by the owner ruling of 2026-08-06 (RFC_012, second sitting, point 2).
**Run and written** 2026-08-07 by the `rfc012_await_findings` lane. The handoff called for
this file under yesterday's date; it carries the date it was actually produced.

**This census does not decide (a) or (a′).** It is the artefact those questions are gated
behind. Every query below is given verbatim so it can be re-run; every Go claim names the
file and line it was read at, not inferred from.

---

## 0. The answer in one paragraph

Under the design the owner's commission names — **the reply moved under a `response`
sub-key** — the fleet does **not** break on the config side (every dotted-path reader is
either auto-unwrapped by `ExtractNestedField` or is dead config that nothing reads), but it
**does break in three places in Go**, silently, all in the hero/logo image path. Under the
*other* design RFC_012 §4 names — **an additive merge with the reply's keys left at the top
level** — nothing in this census breaks at all, and the residual risk reduces to key
collision. **62% of awaited steps (138 of 221) already produce the wrapped shape today**,
because the `call_agent`/`spawn_agent` branch has always merged under `.response`.

---

## 1. Scope: what counts as an "awaited step"

A step is awaited when **its action's result carries `await_response: true`**
(`coordinator.go:1910`, `processAwaitResponse`). This is a **runtime** property of the
action, not a config flag — so the set cannot be read off config alone; it has to come from
the Go source and then be joined to live config.

### 1.1 The actions that can await — 40 of 312 registered

Derived mechanically: parse `registry.go` for `"name": {Handler: X}`, locate each handler's
function body, and look inside it (and one level of local helper calls) for either
signalling form.

> **⚠ THE TRAP THAT COST ME THE FIRST PASS, and it halved the answer.** There are **two**
> ways an action signals await, and grepping for the obvious one finds 24 actions instead of
> 40. The map-literal form is `"await_response": true`. The **typed-struct** form is a result
> struct with a `json:"await_response"` tag set by field —
> `web_search_action.go:221` returns `&WebSearchResult{… AwaitResponse: true …}`. I had
> excluded lines containing `json:"await_response"` as "struct tags, not results", which
> silently dropped **every adapter dispatch** — `web_search`, all five `firecrawl_*`,
> `scrape_web`, `git_commit`, `generate_image`, `batch_webscrape`, the browser-run and
> render-audit and repo-analysis requests. Those are precisely the steps most likely to be
> awaited. **A census of a behaviour must enumerate every way the behaviour is expressed.**

| always awaits | conditionally awaits |
|---|---|
| `await_approval`, `batch_webscrape`, `deploy_image_asset`, `derive_brand_head_assets`, `derive_card_asset`, `directory_export_json`, `dispatch_agent`, `dispatch_thunder_decommission`, `dispatch_thunder_prepare_object_url`, `dispatch_thunder_prepare_object_urls`, `dispatch_thunder_prepare_resume_url`, `dispatch_thunder_provision`, `dispatch_thunder_ssh_exec`, `dispatch_thunder_ssh_get_status`, `emit_sprite_css`, `fetch_news_search`, `fetch_scrape`, `firecrawl_crawl`, `firecrawl_extract`, `firecrawl_map`, `firecrawl_scrape`, `generate_image`, `git_adapter_request`, `git_commit`, `git_commit_action`, `render_provocation_feed`, `request_browser_run`, `request_component_browser_run`, `request_human_input`, `request_render_audit`, `request_repo_analysis`, `scrape_web`, `spawn_agent`, `spawn_agent_k8s`, `start_orchestration`, `wait_for_approval_response`, `web_search` | `await_response`, `call_agent` (config `await_response: false` opts out), `retract_page_deployment` (false on two early-return paths), `spawn_group` (`waitForCompletion`) |

**[BOUNDED]** The helper-following is one level deep. An action that awaits via a
**two-level** helper chain would be missed. I checked the shape that actually occurs —
`SpawnAgentAction → buildSpawnResult`, `DeployImageAssetAction → sendGitCommitRequest`,
`WebscrapeAction` reached from five `firecrawl_*` wrappers — and all resolve at one level.

### 1.2 The live awaited steps — 221

```sql
WITH RECURSIVE all_steps AS (
  SELECT ad.type AS agent_type, st.key AS step_name, st.value AS step, 0 AS depth
  FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') st
  WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
    AND ad.default_config ? 'workflow'
  UNION ALL
  SELECT a.agent_type, sub.key, sub.value, a.depth+1
  FROM all_steps a, jsonb_each(a.step->'config'->'sub_workflow'->'steps') sub
  WHERE a.step->'config' ? 'sub_workflow' AND a.depth < 5)
SELECT ... WHERE step->>'action' IN (<the 40 above>);
```

**The recursion is load-bearing:** loops nest their bodies under
`config.sub_workflow.steps` (18 live occurrences), and a top-level-only query — the shape
the lane RUNBOOK warns about for the (d) census — misses every step inside a loop.

---

## 2. The split that decides everything: which branch of `applyResponseToState` a step takes

`coordinator.go:2635`. Four branches, and **one of them already does what (a) proposes**.

| branch | line | behaviour | live steps |
|---|---|---|---|
| **A. agent response** — `call_agent` / `spawn_agent` / dynamic step with a `TargetAgentType` | 2704 | **MERGES**: reads the existing map, sets `["response"] = reply`, keeps everything else | **138** (126 with an `output_field`, 21 inside loops) |
| **B. `output_mapping`** | 2647 | **REPLACES** with a mapped projection of the reply | **2** |
| **C. default** | 2754 | **REPLACES**: `CollectedData[stepName] = normalisedData` | **81** (68 with an `output_field`, 3 inside loops) |
| D. `spawn_agent` with no existing data | 2744 | replaces with `extractSpawnData` | (subset of A's actions) |

**The at-risk set is B + C = 83 steps.** Branch A's 138 steps are already wrapped, so their
readers already tolerate the shape (a) would introduce — they are the empirical evidence
that the wrapped shape is survivable, and they are excluded from the reader search below.

---

## 3. Config-side readers — 83 references, and **zero of them break**

Every string scalar in every active agent's workflow (recursively, arrays normalised) was
matched against the step names and output fields of the at-risk set, excluding routing keys
(`next_step`, `then_step`, `else_step`, `error_step`, `start_step`, …) so that a
`"next_step": "deploy_page"` is not miscounted as a data read. Query:
`scratchpad/census_readers.sql` (kept with this lane's runbook material).

**83 references, in two shapes.**

### 3.1 The 19 dotted-path reads — the apparent breaks

These name a **field inside** the reply (`page_deployed.commit_sha`), so a `.response`
wrapper would apparently invalidate the path. **It does not.** Resolved by resolver family:

| family | refs | resolver | verdict |
|---|---|---|---|
| `training-launcher` thunder presigns (`resume_url/_key/_index`, `checkpoint_urls`, `dataset_url`, `final_url`, `scripts_url`) | 7 | `ExtractActionInputs` Strategy 0 → `ExtractNestedField` (`action_inputs.go:430`) | **SURVIVES** |
| `site-work-orchestrator` `commit_sha: page_deployed.commit_sha` | 1 | `inputs.Get("commit_sha")` (`load_work_item_actions.go:877`) → same resolver | **SURVIVES** |
| `vet-practice-verifier` `fallback_url_field: search_results.results.0.url` | 1 | `ExtractNestedFieldString` (`webscrape_actions.go:236`) | **SURVIVES** — and see §6.1, it is already inert |
| **`commit_from: <x>.commit_sha`** (pageflow-builder, page-rebuild, page-rerender, report-builder, section-editor, site-work-orchestrator) | 6 | **none** | **DEAD CONFIG** |
| **HITL `output_format` templates** `{{.await_human_approval.approved}}` etc. (simple-content-writer-with-approval) | 4 | **none** | **DEAD CONFIG** |

**Why they survive — the single most important mechanism in this census.**
`datahelpers.ExtractNestedField` (`data_helpers.go:1199`) walks a dotted path and, **for
every segment that is not found directly, retries through `["response"]`**:

```go
// Try direct access first
if val, exists := currentMap[part]; exists { current = val; continue }
// Auto-unwrap: try through .response (call_agent/spawn_agent wrapper)
if response, hasResponse := currentMap["response"].(map[string]interface{}); hasResponse {
    if val, exists := response[part]; exists { current = val; continue }
}
```

So a `.response` wrapper is **transparent to every reader that resolves through this
function** — which is every config-side dotted path in the fleet. This was written for
branch A's wrapper and it generalises to (a) for free.

> Read the code, not the comment. I found this via a comment at
> `classify_training_probe_action.go:17` claiming the auto-unwrap exists, and then read the
> function — the comment happened to be right, but a comment asserting a platform-wide
> guarantee is exactly the class this estate has been bitten by.

**`commit_from` is dead, and that is a finding in its own right.** Six agents configure
`commit_from: <x>.commit_sha`. The only occurrence of the literal `commit_from` anywhere in
Go is a **loop-prefixing list** at `coordinator.go:4265`, whose own comment says
"Used by `update_page_status`" — and `update_page_status`'s handler never asks for it. No
input spec names it; no dynamic construction of the key exists (checked). **Six live config
references that nothing reads**, carried forward by copy. They cannot break under (a)
because they do nothing now.

**The HITL templates are dead the same way.** `simple-content-writer-with-approval`'s
`process_data` step declares an `output_format` map of four `{{.await_human_approval.*}}`
templates. `process_data` resolves to `ProcessApprovalDecisionAction`
(`registry.go:1801`, a deprecated alias), which reads the approval via
`extractApprovalResponse(params.CollectedData, …)` and **builds its own result map** — it
never looks at `config.output_format`. The templates are never rendered.

### 3.2 The 64 whole-key handoffs — genuinely UNCLEAR

`input_fields: ["crawl_result"]`, `scrape_field: scrape_results`, `output_fields:
deploy_result`. These hand the **whole object** to a consumer, so the key still exists under
a wrapper and nothing fails at resolution time — the consumer simply receives
`{response: {…}, response_received_at, response_status}` where it expected the bare reply.

Whether that breaks depends on what the consumer does next, and it varies:
`ExtractFields` (`unified_extractor.go:20`) routes a single-segment name to
`extractSingleField` (`unified_extractor.go:400`), whose Strategy 1 is `FindByPath`
("handles unwrapping") and whose Strategy 4 is an aggressive recursive search — either of
which would find a field nested under `.response`. But the post-processing step,
`UnwrapDeep` (`content_search.go:195`), unwraps only `_result`-suffixed keys and a `result`
key — **not `response`**.

**The decisive evidence that whole-key handoffs do NOT transparently tolerate the wrapper**
is inside `ExtractFields` itself, at `unified_extractor.go:200` — a **hardcoded special
case** for one field —

```go
// reviewed_brief often comes as {"response": {...actual data...}, "response_status": "complete"}
if response, ok := rbMap["response"].(map[string]interface{}); ok { … }
```

Someone hit this exact problem on one branch-A field and patched that one field by hand
rather than fixing the general path. **That is what a (a) rollout would generalise to 64
call sites.** Marked UNCLEAR rather than BREAKS because each needs its consumer read; marked
UNCLEAR rather than SURVIVES because the fleet already contains a hand-written workaround
proving the generic path is insufficient.

---

## 4. Go-side readers — **3 confirmed BREAKS, all silent**

Sweep: every one of the 96 at-risk keys (step names + output fields of branches B/C) against
`CollectedData["<key>"]` across `platform/` and `internal/`, non-test.

| file:line | key | owner step | verdict |
|---|---|---|---|
| `v3_site_actions.go:1010` | `hero_deployed` | `site-work-orchestrator.deploy_hero_image` (`deploy_image_asset`) | **BREAKS** |
| `v3_site_actions.go:1021` | `logo_deployed` | `site-work-orchestrator.deploy_logo_image` (`deploy_image_asset`) | **BREAKS** |
| `assemble_from_library.go:452` | `logo_deployed` | as above | **BREAKS** |

All three are the same two-level direct access with no unwrap anywhere in the path:

```go
if heroDeployed, ok := params.CollectedData["hero_deployed"].(map[string]interface{}); ok {
    if imageURL, ok := heroDeployed["image_url"].(string); ok && imageURL != "" { … }
}
```

Under a `.response` wrapper the outer assertion still succeeds (it is still a map), and
`heroDeployed["image_url"]` is simply **absent**. The `ok` guard then skips the block.

**The failure mode is the dangerous one: it is silent and it is a no-op.** There is no
error, no warning — the only log line in the block is on the success path. The page renders
with no hero image and no logo, and nothing anywhere records that a URL was expected. Any
rollout of (a) that does not fix these three sites first will produce exactly the kind of
defect this estate keeps filing: a page that is wrong in a way no status field reports.

**[BOUNDED]** This sweep finds readers that name an at-risk key as a **literal**. A reader
that composes the key at runtime is invisible to it. The known instance of runtime
composition is §5, and it is clean.

---

## 5. Loop-derived keys — checked, and clean

`applyResponseToState` derives an `output_field` at runtime for steps inside a loop, because
those steps do not exist in the `WorkflowPlan`:
`deriveOutputFieldFromLoopStepName` (`coordinator.go:2800`) turns
`build_pages_loop_iter_2_deploy_page` into `page_deployed_2`. **A config-only census cannot
see those keys** — the handoff flagged this as a thing a naive census would miss.

Only **3** at-risk steps sit inside a loop, and they are all the same one:

| agent | loop | step | action | output_field | runtime key |
|---|---|---|---|---|---|
| `pageflow-builder` | `build_pages_loop` | `deploy_page` | `git_commit` | `page_deployed` | `page_deployed_<N>` |
| `page-rebuild` | `build_pages_loop` | `deploy_page` | `git_commit` | `page_deployed` | `page_deployed_<N>` |
| `site-work-orchestrator` | `build_items_loop` | `deploy_page` | `git_commit` | `page_deployed` | `page_deployed_<N>` |

Their only reader is `commit_from: page_deployed.commit_sha`, which the coordinator rewrites
per iteration via `prefixDataReference` (`coordinator.go:4258-4276`) — and which §3.1 shows
is **dead**. So the loop-derived surface adds **no** break.

---

## 6. Two findings this census turned up that are not about (a)

### 6.1 `search_results.results.0.url` can never resolve — an inert fallback

`vet-practice-verifier` configures `fallback_url_field: search_results.results.0.url`, read
via `ExtractNestedFieldString`. That function splits on `.` and does **map access only** —
~~at the `results` segment `current` is a JSON **array**, the
`current.(map[string]interface{})` assertion fails, and the function returns `nil`~~
> **CORRECTED 2026-08-09 (owner-directed fix thread):** off by one segment — `results`
> RESOLVES, via the function's `["response"]` auto-unwrap (the websearch adapter nests its
> payload under `response`), and yields the array; it is the **`0`** segment where the
> `current.(map[string]interface{})` assertion fails and the function returns `nil`. Same
> conclusion, and the distinction matters to the fix: array-index support is needed at the
> segment AFTER the unwrap, not instead of it. Caught by reading the walk against the
> adapter's actual reply shape (`internal/adapters/websearch/adapter.go:532`).

The fallback has therefore **never** produced a URL, and it fails silently by design ("primary
was empty, try fallback, still empty"). Not caused by (a); found while tracing (a)'s
resolvers. ~~Worth a bug file by whoever owns the vet lane.~~ **Being fixed 2026-08-09 at the
owner's direction** (array-index support in `ExtractNestedField` + failure-path logging at the
`scrape_web` caller), so no separate bug file — the fix thread's trail is in
`rfc012_await_findings/` (PLAN 2026-08-09 later section).

### 6.2 Dead config keys survive indefinitely because nothing looks for them

`commit_from` (6 agents) and the four HITL `output_format` templates are configured,
copied between agents, and read by nothing. There is no drift check that would notice. This
is the config-side twin of the `agent_error_log` copy class RSH-008 just retired: the
mechanism that makes a copy cheap also makes a dead copy invisible.

---

## 7. What this means for (a) and (a′) — stated as inputs to the decision, not as the decision

**Design 1 — the reply under a `response` sub-key** (the form the commission names, and the
form branch A already uses):
- config side: **0 breaks** — 9 survive via `ExtractNestedField`'s auto-unwrap, 10 are dead;
- Go side: **3 breaks**, silent, all in the hero/logo image path, all fixable in one commit
  by routing them through `ExtractNestedField`;
- **64 whole-key handoffs UNCLEAR**, and `ExtractFields`' hardcoded `reviewed_brief` unwrap
  is direct evidence that at least some of them need per-consumer work.
- Cost is therefore *not* "every awaited step's readers" as §4 of the RFC feared — it is
  3 known sites plus a 64-item review.

**Design 2 — additive merge, reply keys left at the top level, prior keys kept where they
do not collide:**
- config side: **0 breaks** (every path resolves exactly as it does today);
- Go side: **0 breaks** (the three direct accesses still find `image_url` at the top level);
- residual risk is **key collision** between an action's pre-dispatch findings and the
  reply's own keys — a bounded, checkable problem, and the same class the (d) detector
  already handles for `output_field`.

**The evidence favours design 2 sharply**, and the reason is worth stating plainly: design 1
was proposed because branch A already does it, but branch A gets away with it *only*
because `ExtractNestedField` was taught to compensate — and that compensation covers dotted
paths, not whole-key handoffs or direct map access. Design 2 needs no compensation.

**(a′) `storeActionResult`** is **NOT covered by this census.** The commission's §3(a) scope
is the awaited-response path; `storeActionResult` is the synchronous write and has its own
reader set. Stating the gap rather than implying coverage.

---

## 8. Re-running this census

Everything is re-runnable. The two SQL files are in this lane's scratchpad and reproduced
inline above (§1.2, §3). The Go-side pass is:

```bash
# 1) the awaited action set — BOTH signalling forms
grep -rn '"await_response"\s*:\s*true\|AwaitResponse:\s*true' --include=*.go platform/ | grep -v _test.go
# 2) direct map access on an at-risk key (the silent-break shape)
grep -rn --include=*.go 'CollectedData\["<key>"\]' platform/ internal/ | grep -v _test.go
```

**Figures are dated 2026-08-07 and will move.** `agent_definitions` changes daily; **176**
active non-snapshot agents with a `workflow` [MEASURED 2026-08-07, this census's own run —
not carried from the lane RUNBOOK's (d) baseline, which happens to give the same number]
carried 221 awaited steps on this date. Re-run rather than quoting these
numbers — the ratio of branch A to branch C is the number that matters, and it moves
whenever an agent is seeded.
