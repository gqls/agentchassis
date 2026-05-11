# FOCUS: CollectedData — Architecture, Lifecycle, Observed Issues

Date: 2026-05-11
Status: Analysis only. No code changes proposed yet — recommendations are at the end and should be triaged before any structural edits.

---

## 1. What CollectedData Is

`CollectedData` is the orchestration-scoped working memory for every agent. It is a `map[string]interface{}` field on `OrchestrationState`, persisted to the `orchestration_states.collected_data` JSONB column and reloaded from there on every step.

```go
// platform/orchestration/state.go (struct snippet)
CollectedData      map[string]interface{} `db:"collected_data"`
InitialRequestData json.RawMessage        `db:"initial_request_data"`
FinalResult        json.RawMessage        `db:"final_result"`
```

It is the **single channel** through which step outputs propagate, routing metadata is carried, loop variables are bound, and parent-reply context is preserved across the asynchronous response cycle. Every action sees it as `params.CollectedData`. Every step result writes into it. Every dot-path in workflow config (`input_data.spec.page_name`, `site_record.content_data`, `current_section.name`) resolves against it.

That makes it the most overloaded data structure in the system.

---

## 2. Lifecycle — Who Writes to CollectedData, and When

There are **two distinct CollectedData pipelines** that interact:

### 2a. Per-message CollectedData (transient)
Built fresh by `BuildCollectedData()` in `platform/orchestration/datahelpers/data_helpers.go` every time the chassis receives a message. It is wrapped into a `MessageContext` and is what stateless / first-touch handlers see.

### 2b. Per-orchestration CollectedData (persistent)
Held in `state.CollectedData`, loaded from JSONB on every step, mutated, and written back. This is the canonical accumulator.

When a step runs, `buildActionParams()` (coordinator.go) sets `params.CollectedData = state.CollectedData` — i.e. the **persistent** map is passed by reference into the action. Any mutation by the action is visible to the next save. This is why the historical bug of calling `NormalizeCollectedData()` inside `ExecuteLLMPromptAction` "destroyed accumulated state" — it was rebuilding the per-message view on top of the per-orchestration view.

### Write entry points (canonical list)

| Site | What it writes | When |
|---|---|---|
| `BuildCollectedData` (data_helpers.go) | `__execution_context__`, `__my_requests_topic__`, `__my_responses_topic__`, `__parent_responses_topic__`, `__reply_to_request_id__`, `input_data`, `action`, `config`, `prompt`, `agent_config`, `__raw_message__`, `__work_request__` | First touch of a message; new orchestration init via `NormalizeCollectedData` → `CreateInitialState` |
| `StateRepository.CreateInitialState` (state.go) | The initial seed map | Orchestration creation |
| `storeActionResult` (coordinator.go) | `state.CollectedData[state.CurrentStep] = result` AND `state.CollectedData[step.OutputField] = result` (if non-empty) | After every local-action step |
| `applyResponseToState` (coordinator.go, handleCompleteResponse) | `stepName` and `outputField` keys, plus `response`, `response_received_at`, `response_status`, `initialized` for agent calls | When a child agent's response arrives |
| `applyOutputMapping` (coordinator.go) | A flat projection of selected fields into the same key (no `.response` wrapper) | When step config has `output_mapping` |
| `setLoopVariable` (coordinator.go) | `loop_var_name`, `__current_loop_item_key__`, and propagation of `{output_field}_N → {output_field}` for the current iteration | Before each loop substep |
| `handleLoopExpansion` (loop_expansion_handler.go) | `loop_metadata`, `{loop_name}_item_{N}` for each item | Once per loop encountered |
| `MessageProcessor.process` (processor.go) | `agent_config`, and `__execution_context__` overwrite for children | Per inbound message |

### Read entry points

- `datahelpers.FindByPath` — dot-notation traversal with `.response` auto-unwrap, recursive `UnwrapDeep`, and `input_data.` prefix add/remove fallbacks.
- `datahelpers.ExtractNestedField` — similar.
- `datahelpers.GetInputData` — explicit `input_data` retrieval.
- `datahelpers.ExtractActionInputs` (with `ActionInputSpec`) — the 5-strategy cascade documented in 001 dev guide (Strategy 0 explicit dot-path config → 1/2 `ExtractFields` → 3 deprecated `*_field` → nested-source loop → 4 config-value-as-name).

---

## 3. Anatomy of the Sample Log

The log block you supplied shows `endpoint-health-checker` running a 2-step workflow (`check_health → complete`). The `before_complete` `CollectedData structure` log line is the key artefact. Re-laid out:

```
state.CollectedData (top level):
├── __execution_context__       ← current ExecutionContext
├── __my_requests_topic__       ← "system.agent.generic.requests"
├── __my_responses_topic__      ← "system.agent.generic.responses"
├── __parent_responses_topic__  ← "system.generic.responses"
├── __work_request__            ← who to reply to
├── __raw_message__             ← FULL inbound message, including:
│   ├── __execution_context__   ← duplicate of top-level
│   ├── __my_requests_topic__   ← duplicate (empty string this time)
│   ├── __my_responses_topic__  ← duplicate
│   ├── __parent_responses_topic__
│   ├── __raw_message__         ← THIRD-level nesting: {action, config, input_data}
│   ├── __work_request__        ← duplicate
│   ├── action, agent_config, agent_definition, agent_group, config, input_data
├── action                      ← "orchestrate"
├── agent_config                ← workflow definition
├── agent_definition            ← Endpoint Health Checker definition
├── agent_group                 ← duplicate of agent_definition
├── config                      ← {agent_type: "endpoint-health-checker"}
├── input_data                  ← {action: "check_endpoint_health"}
├── check_health                ← {changed: 0, checked: 2}   ← step name
└── health_result               ← {changed: 0, checked: 2}   ← output_field

state.InitialRequestData (json.RawMessage, separate column):
├── action, agent_group, agent_config, agent_definition, config, input_data
├── __raw_message__, __work_request__, __execution_context__
└── (overlaps with state.CollectedData["__raw_message__"])
```

That single trivial workflow has stored the same `{changed:0, checked:2}` payload **twice** (step-name + output_field), the same agent definition **twice** (agent_definition + agent_group), and the same inbound message **at least three times** (top-level inputs, `__raw_message__`, `__raw_message__.__raw_message__`, and `InitialRequestData`).

This is not a one-off — the same shape appears in every page-content-writer, content-creator-hero, and pageflow-builder log we have. The duplication is structural.

---

## 4. Issues Found

### 4.1 Recursive `__raw_message__` nesting

`BuildCollectedData` does:

```go
collectedData["__raw_message__"] = body   // the *original* body, untouched
```

But by the time this fires for a child agent or a re-routed message, `body` already contains an `__execution_context__` and `__raw_message__` from the parent. So each hop re-wraps the previous. With deep call chains (e.g. site-work-orchestrator → page-content-writer → research-agent → content-creator-hero) we can easily get 4–5 levels of nested `__raw_message__`, each carrying a complete copy of everything above it.

**Cost:** every UpdateState marshals the entire JSON and writes it. With 15 optimistic-lock retries possible per save, a 50KB blob becomes ~750KB of write traffic for a single step.

**Why it's there:** the comment says `// Store the raw message for debugging`. It is genuinely useful for one thing — the `extractReplyToMetadata` fallback chain that walks `collectedData["__raw_message__"]["body"]["data"]` looking for `operation`/`operands` in the calculator agent. Outside that, no production code reads `__raw_message__` more than one level deep.

### 4.2 Duplicate result storage at `step_name` AND `output_field`

```go
// platform/orchestration/coordinator.go, storeActionResult
state.CollectedData[state.CurrentStep] = result
if step.OutputField != "" {
    state.CollectedData[step.OutputField] = result
}
```

The same is mirrored in `applyResponseToState`. The intent is to support both lookup patterns: workflows that reference a step by name (`check_health.changed`) and workflows that reference it by its declared output_field (`health_result.changed`).

This means every step result that has an `output_field` is stored twice in the same map. For the simple case it's a few bytes. For `loop_complete` aggregations of multi-iteration page content, it's tens of KB doubled.

**Note:** there is no synchronisation between the two copies. If a subsequent step mutates `state.CollectedData[stepName]` (some actions do — `setLoopVariable` propagates fields back to base names), the `output_field` copy drifts.

### 4.3 `state.InitialRequestData` and `state.CollectedData["__raw_message__"]` overlap

`InitialRequestData` is a separate `json.RawMessage` column. `__raw_message__` is a key inside the `collected_data` JSONB. Their content is largely the same — the message that initialised this orchestration.

Both are written. Both are persisted. Neither is read by any code path I can find except for logs (`CreateInitialState initialData look for action, request id etc`).

### 4.4 Conflation of unrelated concerns in one map

CollectedData mixes at least six categories of data with no separation:

| Category | Example keys |
|---|---|
| Routing | `__my_requests_topic__`, `__my_responses_topic__`, `__parent_responses_topic__`, `__work_request__`, `__reply_to_request_id__` |
| Identity | `__execution_context__`, `agent_config`, `agent_definition`, `agent_group` |
| Inputs from caller | `input_data`, `action`, `config`, `prompt` |
| Step results | `<step_name>`, `<output_field>`, indexed loop variants (`page_html_0`..N) |
| Loop control | `loop_metadata`, `{loop_name}_item_{N}`, `current_section`, `current_item`, `__current_loop_item_key__`, base-name propagations |
| Debug artefacts | `__raw_message__`, `__work_request__.timestamp` |

The result is that field-name collisions are a known live risk — the 001 dev guide documents the case where `section-editor` declared `content_data` as optional and the nested-source loop silently lifted `site_record.content_data` (an unrelated site plan) into the action's inputs and overwrote a hero section. The guide marks this as **latent for required fields too** — any action invoked from a context where `site_record` is in scope can silently pick up `site_record.site_id` instead of the caller's `site_id`.

That is a structural consequence of the flat-namespace design.

### 4.5 `BuildCollectedData` runs an aggressive "is this input_data?" heuristic

```go
// hasSystemFields(): treats a map as a wrapper if it contains any of
// action, config, agent_config, __execution_context__, workflow, headers
if !hasSystemFields(unnestedBody) {
    collectedData["input_data"] = unnestedBody
} else {
    collectedData["input_data"] = make(map[string]interface{})  // empty
}
```

For inbound work-orchestrator dispatches, the body has both `input_data` AND system fields (`action`, `agent_config`), so the explicit `input_data` is extracted correctly. But for messages that contain `headers` as a field name in their own data (rare but possible — anything user-supplied), `hasSystemFields` returns true and `input_data` is silently emptied.

The "double `body.body` unwrap" with `success` field detection is similarly heuristic. It catches the two known producers (responses from agents) but is brittle to a third format.

### 4.6 `__execution_context__` is overwritten in-place

Per `processor.go process()`:

```go
if msgCtx.IsChildOrchestration() {
    msgCtx.CollectedData["__execution_context__"] = msgCtx.ExecutionContext
}
```

And per `BuildCollectedData`:
```go
collectedData["__execution_context__"] = execCtx
```

The same key carries either:
- The orchestration's *own* execution context (top-level orchestration), or
- The *parent's* execution context (child orchestration after spawn), or
- The *current message's* execution context (transient, per-message rebuild).

Whether you get the parent or the current depends on which write happened most recently. `extractReplyToMetadata` priority-orders `__work_request__` first to avoid relying on which `__execution_context__` won — which is itself evidence that the original semantics weren't clear.

### 4.7 Loop expansion writes per-iteration items into the flat namespace

For a 5-iteration page-content-writer loop, after expansion you have in CollectedData:

```
loop_metadata
process_sections_loop_item_0..4         ← the input items
generated_content_0..4                  ← per-iteration LLM results
page_html_0..4                          ← per-iteration rendered HTML
section_output_0..4                     ← per-iteration final sections
current_section                         ← rebound before each substep
process_sections_loop_iter_0_write_content
process_sections_loop_iter_0_render_section
... (one entry per iteration × substep)
process_sections_loop_complete          ← aggregated array
```

By design — necessary for indexed reference. But: the persisted JSON blob now contains the full HTML of all 5 sections **at least three times**: once at `page_html_N`, once at `generated_content_N.result`, and once inside the `loop_complete` aggregated array. Plus a fourth copy if the parent workflow stores the aggregated result at its own `output_field`.

For a 7-page multipage build, that's 4× × 5 sections × ~5KB of HTML × 7 pages ≈ 700KB of JSON living in `orchestration_states.collected_data`, re-marshalled on every step.

### 4.8 The handleCompleteResponse normalisation strips more than expected

Sample from the historical generic-agent logs:

```
"original_fields": 7,
"normalized_fields": 2,
"normalised data_keys": ["business_name", "business_type"]
```

A 7-field response from `content-creator-hero` was reduced to two fields after `parseResponseBody → CleanDataMap`. The fields stripped are the system-field set (`action`, `config`, `agent_config`, `__execution_context__`, `__raw_message__`, etc.). When a content-creator agent legitimately wants to return something *named* like one of those keys, it disappears.

This isn't theoretical — the dev guide entry on "site_record.content_data" silently overwriting hero content is one instance of name collision causing data loss. `CleanDataMap` is another, on the receive path.

---

## 5. Where the Code Already Compensates

The code does mitigate the design pressure in several places, which is worth flagging because changes to CollectedData semantics will need to update or remove these compensations:

- **`extractReplyToMetadata`** has a 3-tier priority chain (`__work_request__` → `__execution_context__` → current ExecutionContext) precisely because `__execution_context__` can mean different things.
- **`UnwrapDeep`** in datahelpers does recursive `.response` / `result` / markdown-fence stripping to find content despite the layered wrapper soup.
- **`FindByPath`** auto-tries with and without an `input_data.` prefix because the field could be either flat (set by handler) or wrapped (set by orchestrator).
- **`ExtractActionInputs`** explicitly defines a 5-strategy cascade with explicit-first-then-fallback so that callers can route around the flat-namespace collision risk.
- **`applyResponseToState`** has a separate `output_mapping` path that produces flat output for response data — added specifically to avoid `.response` wrapping for `hero_result.image_uri`-style usage.
- **`deriveOutputFieldFromLoopStepName`** retrofits an output_field name from the step name pattern, because dynamically-expanded loop steps don't have an output_field set by the workflow author.

Each of these is a useful tool. Together they describe how much the framework is working around the conflation problem rather than solving it.

---

## 6. What's Actually Going Wrong Right Now (vs Latent Risk)

| Item | Active failure mode | Latent risk |
|---|---|---|
| Recursive `__raw_message__` | DB write amplification, JSON parsing time on every load | Memory pressure on long-running orchestrations |
| Step-name + output_field duplication | Wasted DB bytes; drift if one is mutated | Subtle bug where reader gets stale value |
| `InitialRequestData` + `__raw_message__` overlap | Wasted bytes | None functional |
| Conflation in one flat map | Silent overwrites of action inputs (documented) | Required-field shadowing (silent), cross-site context bleed |
| `hasSystemFields` heuristic | Edge case: legitimate user `headers` field empties `input_data` | None known active |
| `__execution_context__` overwrites | None known active (`__work_request__` priority covers it) | Future caller relying on `__execution_context__` directly gets parent's instead of own |
| Loop iteration data triplication | Bloat of `orchestration_states.collected_data` | DB row size limits hit on very large sites |
| `CleanDataMap` stripping fields | Response field named like a system field is dropped | Hard to predict; only surfaces when an agent author happens to choose a colliding name |

The most immediate, ongoing, observable cost is **DB write amplification from recursive `__raw_message__` × 15 optimistic-lock retries × every step**.

---

## 7. Recommendations (Structural, in Priority Order)

These are stated as options to discuss before any work. Each has different blast-radius.

### R1. Don't store `__raw_message__` recursively
Before assigning to `__raw_message__`, strip system keys (`__execution_context__`, `__raw_message__`, `__work_request__`, `__my_*_topic__`) from the body copy. The single legitimate consumer of `__raw_message__` (calculator agent's legacy fallback) only needs the data fields, not the wrappers.

Blast radius: small. Affects one function (`BuildCollectedData`). Test: calculator agent + a representative response-handling flow.

### R2. Pick one of {step_name, output_field}, not both
The double-write is a contract problem: workflows that reference a step by name read `state.CollectedData[stepName]`, workflows that reference by output_field read the other. We could:

- **Option A:** Always store at `output_field` if set, fall back to step name otherwise. Update callers that read by step name to read by output_field.
- **Option B:** Store at step name only, treat `output_field` purely as an alias resolved at read time.
- **Option C:** Keep as-is, document the duplication as intentional.

C is what we have. A or B halves the per-step write size. A is forward-compatible with the design that's already winning (output_mapping → flat results).

Blast radius: medium. Every workflow that reads by step name needs auditing. The 001 dev guide's data-path-validation appendix is the right tool to start from.

### R3. Drop `InitialRequestData` OR drop `__raw_message__` (not both)
They overlap. Pick one. `InitialRequestData` is a clean separate column, easier to query (`SELECT initial_request_data FROM orchestration_states WHERE …`) than digging into a JSONB key. `__raw_message__` is in scope for actions reading CollectedData.

Recommendation: keep `InitialRequestData` (it's the persisted record of what started this), drop `__raw_message__` once the calculator-style fallbacks are migrated to read from `input_data`.

Blast radius: small if the fallback in `extract_action_input` for calculator-style legacy paths is the only reader.

### R4. Namespace the flat map
The dev guide already calls out the silent collision risk. Three approaches:

- **Light:** Keep the flat map, but require `ExtractActionInputs` callers to declare explicit `input_fields` paths. Forbid the nested-source loop falling back to `site_record.*` for caller-supplied fields. (The fix would be in `ExtractActionInputs` itself: don't fall through to nested sources for fields not present in `input_fields`.)
- **Medium:** Introduce sub-maps: `state.CollectedData["__routing__"]`, `state.CollectedData["__results__"]`, `state.CollectedData["__loop__"]`. Migrate writers to put data in the right sub-map. Reads stay backward-compatible via fallthrough.
- **Heavy:** Replace `map[string]interface{}` with a struct that has typed sub-maps. Big change. Probably not worth it.

Light is the cheapest. Medium is the architectural fix. Heavy is for a v2.

### R5. Per-iteration garbage collection in loops
After `loop_complete` aggregates `page_html_0..N` into a single array, the indexed copies are no longer needed for forward execution. They are kept because workflow authors may reference them downstream, but in practice the aggregated array is the contract.

Add a config flag on the `loop_complete` step (`drop_iteration_keys: true`) to delete `{output_field}_{N}` keys after aggregation. Default off for safety.

Blast radius: small, opt-in.

### R6. State write reduction: only marshal CollectedData on meaningful change
`UpdateStateWithVersion` always re-marshals the full map. With optimistic-lock retries of up to 15 attempts, the cost of a single step's save can be high. If `CollectedData` is unchanged from the loaded version (only `ProcessingHistory` or `LastActivity` updated), we could write a smaller delta. This is a long-term project, not a quick win.

---

## 8. What the Logs Have Been Telling Us

Going back through the sample and the file traces, the persistent pattern is:

1. CollectedData grows monotonically over the orchestration lifetime — there is no cleanup pass except the explicit removal of `AwaitedRequests` entries.
2. Every layer that adds data also keeps the previous layer's data, partly for backward compatibility (multiple fallback paths in `FindByPath`), partly for debugging (`__raw_message__`).
3. The framework has accreted compensating mechanisms (`output_mapping`, `output_field` alias, `target_role` lookup) faster than it has consolidated the underlying conflation.

The endpoint-health-checker log is a small example showing the structural fingerprint clearly because the workflow is trivial — 2 steps, no loops, no child agents — yet the persisted state already shows triple-nested `__raw_message__`, duplicated agent definition, duplicated step result.

---

## 9. Suggested Next Step

Pick R1 (strip system keys before storing `__raw_message__`). Smallest blast radius, biggest immediate effect on DB write traffic, zero behaviour change for any consumer except the calculator agent's legacy fallback. If that lands cleanly, R3 (collapse `InitialRequestData` and `__raw_message__`) is the next surgical change.

R4 (namespacing) is the structural fix and should be planned separately — it touches every action and every workflow. Probably belongs in its own focus doc once R1–R3 are bedded in.

No changes proposed in code form here. This report is groundwork for that conversation.
