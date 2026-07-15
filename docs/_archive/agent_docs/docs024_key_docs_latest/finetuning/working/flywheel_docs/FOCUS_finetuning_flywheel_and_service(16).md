# FOCUS — Finetuning Flywheel & finetuning.uk Service

Living document. Consolidates what we have planned for the AI training flywheel
(internal: our own models get better over time as we build sites) and opens the
separate question of whether finetuning.uk becomes a **self-service fine-tuning
product** for external users.

Last touched: 2026-04-23

---

## 1. Two separate concerns, one doc

| Concern | What it is | Owner | State |
|---|---|---|---|
| **A. Internal flywheel** | Our pipeline produces training data as a byproduct of building sites. We periodically fine-tune local models on that data, swap them in for Claude calls where quality holds, and drop API costs. | Agent orchestration system | Flywheel A (data export) + B (RAG) done. Flywheel C (training) scripted, awaiting first run on GPU VM. Flywheel D (eval) paused. |
| **B. finetuning.uk product** | External users upload their own data and fine-tune a model through a UI on finetuning.uk. Likely monetised. | New product surface | Not started. Questions to answer before scoping. |

A and B share plumbing (Ollama, GPU scheduling, training data export, Unsloth)
but have very different surfaces, failure modes, and economics. Keep them
distinct in our heads.

---

## 2. The internal flywheel — what exists

### 2.1 Data capture

Every LLM call in the system writes to `llm_call_log` (migration 081, flywheel
columns added in 085):

| Column | Purpose |
|---|---|
| `agent_type`, `step_name` | Which agent, which prompt — the join key for "what to train" |
| `model`, `model_resolved`, `provider` | What was actually called |
| `prompt_template`, `prompt_rendered`, `response_text` | Raw training pair |
| `input_tokens`, `output_tokens`, `latency_ms` | Cost/perf signal |
| `success`, `error_message`, `retry_count` | Quality filter |
| `work_item_id` | Link call → outcome (did the fix work?) |
| `prompt_variant` | A/B on prompt versions |
| `vertical` | Per-industry slicing |
| `rag_context_used` | Was RAG injected into this prompt? |

Write path lives in `ai_actions.go` → `LogLLMCall` (fire-and-forget goroutine,
5s timeout, never blocks the workflow). Flywheel context is extracted from
`CollectedData` inside `ExecuteLLMPromptAction`.

Retention: 90 days for successful calls, 180 days for errors. Cleanup function
exists (`cleanup_old_llm_logs`) but nothing schedules it yet.

### 2.2 Knowledge base (RAG — the other half of the flywheel)

`knowledge_base` table (migration 082) — pgvector(768) for nomic-embed-text.
Shared resource, any agent can read via `rag_lookup`, any can write via
`rag_index`. Trigram fallback when Ollama is down.

Best-practices doc says: **filter first by metadata, then rank by similarity.**
Metadata fields that should be on every row: `vertical`, `component_type`,
`content_type`, `source`, `source_quality`. Without these, a vet example gets
retrieved when writing gas-wholesale copy.

RAG knowledge is the **short-term** lever (useful immediately, no training
needed). Fine-tuning is the **long-term** lever (needs 200+ examples, a GPU,
and evaluation).

### 2.3 Inference routing

`ai_endpoint_health` table (migration 085) tracks three endpoints:

| Name | URL | Role |
|---|---|---|
| claude | `https://api.anthropic.com/v1/messages` | Default for high-quality, low-volume |
| cpu-ollama | `http://ollama-adapter.ai-persona-system.svc.cluster.local:11434` | Embeddings + small models (mistral-small3.1, nomic-embed-text) |
| gpu-ollama | `http://ollama-gpu.ai-persona-system.svc.cluster.local:11434` | Llama 3.3 70B, future LoRAs — currently DOWN, not always-on |

Healthy endpoint → claim flows. Unhealthy → items wait (back-to-triage). GPU is
either healthy (work flows) or not — no separate batch scheduler needed.

### 2.4 Model swap / revert

`snapshot_agent()`, `swap_agent_model()`, `revert_agent()` (migration 083) —
per-agent per-step, safely snapshot before swapping an `ai_service` block in
`agent_definitions.default_config`. The nuclear option (full-table backup)
still exists alongside.

### 2.4b Empirical findings from flywheel B steps 1-2 (2026-04-21)

**Operational note:** Always read `/mnt/project/some_schemas` before writing SQL
against any table. The file contains current schemas for `orchestration_states`,
`knowledge_base`, `site_work_items`, `agent_definitions`, `llm_call_log`, and
others. Guessing column names wastes cycles and erodes trust. Orchestration step
outputs live in `collected_data` (jsonb, default `{}`), not `final_result`.

Verified by hand, bypassing the chassis:

- **pgvector + ivfflat + cosine ops** — works. Synthetic orthogonal vectors
  ranked correctly (cat 1.0, dog/piano 0.09 equidistant). 768-dim storage
  and retrieval round-trip clean.
- **Ollama `nomic-embed-text` on cpu-ollama adapter** — serving. Embeddings
  return in JSON `{"embedding":[...]}` form, 768 floats.
- **End-to-end retrieval on real content** — works. Four test chunks
  (French Bulldog BOAS, Labrador temperament, grand piano, EV batteries)
  embedded, inserted, queried with "dog breed airway problems breathing
  difficulty". Results distinguish dog from non-dog with a gap of ~0.17.
- **Task prefixes matter and are load-bearing.** Without `search_document:`
  and `search_query:` prefixes, Labrador (0.6156) narrowly beat French
  Bulldog (0.6139) on a BOAS-specific query — wrong ranking. With prefixes,
  French Bulldog (0.6256) beat Labrador (0.6173) with a 5× wider gap.
  Prefixes are mandatory in production `rag_index` and `rag_lookup`.
- **Nomic v1 is good enough for proof.** On an artificially hard test
  (30-word chunks, ambiguous query), ranking is competent but not
  outstanding. Real RAG has longer chunks, bigger corpora, more specific
  queries, and passes top-5 to the LLM — nomic v1 is likely fine for
  initial use. Upgrade to `nomic-embed-text-v2-moe` is a drop-in when/if
  retrieval quality becomes a limiting factor (same 768 dims, no schema
  change).

Action items from these findings:

- [x] Prefix patch written 2026-04-21: `PATCH_rag_actions_nomic_prefixes.md` + `rag_actions_prefix_patch.go.txt`. Three surgical edits to `platform/orchestration/actions/rag_actions.go` — new `applyNomicPrefix` helper, one-line change in `RAGLookupAction`, one-line change in `RAGIndexAction`. Model-scoped (only triggers for `nomic-embed-*`), guards against double-prefix, logs `prefix_applied` for observability.
- [x] **Prefix patch deployed and verified live 2026-04-21.** Chassis running with prefix code. Verification via `rag-test-agent` orchestration `ffb462ba-67f9-4493-92ef-5c7bad66d493` — log line `"prefix_applied":true` observed on `rag_lookup` step. Vector search returned indexed content with 0.7929 similarity. Indexed content and query both went through nomic task prefixes as designed.
- [x] **Step 3 (chassis integration) passed on 2026-04-21.** Created `rag-test-agent` (deterministic 3-step workflow: rag_index → rag_lookup → complete_workflow), triggered via Kafka on `system.agent.generic.requests`, chassis v1.0.979 executed cleanly to COMPLETED status. `collected_data` shows `index_result.stored=1`, `lookup_result.search_method=vector`, `lookup_result.result_count=1`, `top_similarity=0.6514`, correct content returned. Registry patches (PATCH 03 from 005_PATCHES.md) are confirmed live in v1.0.979. Flywheel B is done.
- Note: `ON CONFLICT` on `agent_definitions` must target `(type, version)`, not `(type)` alone. Project convention.
- Note: Step outputs live in `orchestration_states.collected_data`, not `final_result`. The `complete_workflow` action's `output_fields` config does not imply a write to `final_result`.
- Minor: `execution_path` column was `[]` in our test run despite three steps executing. Not blocking, worth noting for future orchestration-state debugging work.

### 2.4c Known issues to come back to

- **`orchestration_states.execution_path` not populated** — was `[]` after our step-3 run even though three steps executed and their outputs landed in `collected_data`. Either orchestrator isn't writing the path, or it's only written in certain modes. Not affecting the flywheel; relevant when we want visibility into how an orchestration progressed. Worth investigating when we have a reason to use this column (e.g. building a debug / replay UI).
- **RWO PVC rolling-restart deadlock on ollama-adapter** — resolved 2026-04-22. The `strategy.type: Recreate` pattern is now in the kustomize base (`deployments/kustomize/services/ollama-adapter/base/deployment.yaml`) so this cannot recur. Root cause: default `RollingUpdate` strategy plus ReadWriteOnce PVC means new pod can't mount until old pod releases, old pod won't release until new pod is Ready — classic deadlock. Also added `OLLAMA_LOAD_TIMEOUT=10m` and `OLLAMA_KEEP_ALIVE=30m` as env vars; default 60s load timeout was killing first inference after cold pod start because 14.4GB model load takes ~45s.
- **`agent_type` column in `llm_call_log` partially empty** — some calls land with empty string or NULL. Filters (and therefore training-data exports) need to tolerate this. Worth fixing at the log-write site when we're in that code next.

### 2.4d Flywheel D — Claude vs Ollama quality comparison (2026-04-22 — in progress)

**Goal:** Can we swap any production agent's LLM from Claude to a local Ollama-hosted model without unacceptable quality loss? The first real test of the "swap agent_model and compare" loop.

**Target chosen:** `page-content-writer` / `process_sections_loop_iter_0_generate_content`.

| Property | Value | Why we picked it |
|---|---|---|
| Successful production calls | 1,945 | Biggest logged volume of any agent/step |
| Avg output | 150 tokens | Shortest real outputs — fast to eval |
| Structured JSON | 97% | Automatic schema scoring trivially possible |
| Task | Hero section (headline + subheadline) | Bounded, creative, objectively auditable |
| Current Claude model | `claude-sonnet-4-6` | Known good baseline |

Briefing-agent was considered first but had zero logged calls (ran pre-logging era). Webdesign-agent was the next candidate but its task is known to score 3/10 on Mistral Small 3, so the test would confirm what we already know. Content-gap-planner (671 calls, all-JSON, 230 tok) was a viable fallback.

**Approach:** Replay, not re-run. We have 1,945 stored Claude outputs in `llm_call_log` — we don't need to re-invoke the agent. We pull 20 diverse production prompts, POST each to `mistral-small3.1` on the cpu-ollama adapter, compare the fresh Ollama response against the stored Claude response. No orchestration state touched, no `swap_agent_model` called.

**Test set construction:**

```sql
SELECT row_to_json(t) FROM (
  SELECT id::text, orchestration_id::text, prompt_rendered, response_text
  FROM (
    SELECT DISTINCT ON (orchestration_id) id, orchestration_id,
           prompt_rendered, response_text, created_at
    FROM llm_call_log
    WHERE agent_type = 'page-content-writer'
      AND step_name = 'process_sections_loop_iter_0_generate_content'
      AND success = true
      AND orchestration_id IS NOT NULL
    ORDER BY orchestration_id, created_at DESC
  ) one_per_orchestration
  ORDER BY created_at DESC LIMIT 20
) t;
```

One prompt per distinct orchestration → 20 diverse inputs across 20 different sites/runs. Output is NDJSON (one JSON object per line), safe for multi-line fields.

**Infrastructure notes earned the hard way:**

1. The chassis calls Ollama via `/api/chat`, not `/api/generate`. Our eval uses the same endpoint to stay representative.
2. `mistral-small3.1` is a 24B Q4_K_M model. On CPU, prompt processing ~150 tok/s, generation ~2.5 tok/s. Real production prompts (~500-3300 tokens) take 3-6 min per case; worst-case 10-15 min.
3. Request needs a very generous curl timeout (`-m 1800` = 30 min per case) even though typical will be 3-6 min. Ollama's default 60s server-side load timeout also needed raising to 10m.
4. `num_predict: 250` chosen — iter_0 real outputs average 150 tokens, so 250 is generous headroom. Setting 400+ allowed the model to waffle and slowed each call by ~100s.
5. Port-forward to `ollama-adapter` on a non-default port (21434) rather than spawning per-call curl pods — 10-20× faster and avoids pod-exec race conditions we hit early.

**Comparison approach:**

Automatic first. For each (prompt, claude_response, ollama_response) triple, compute:
- `claude_valid_json`, `ollama_valid_json` — baseline JSON compliance rate
- `claude_keys`, `ollama_keys`, `keys_match` — structural schema match
- Length and token counts
- Latency

If the automatic signal is ambiguous, follow up with Claude-as-judge scoring on disputed pairs.

**Current status:** Test set generated (20 diverse cases). Runner has been deployed twice and hit timeouts both times. Final runner with `-m 1800` timeout and `num_predict: 250` is queued to run. Eval may take 30 min - 2 hours depending on real inference speed.

**Known hypothesis, documented before we see the data:**

Based on the project's earlier testing (020e doc), Mistral Small 3 scored 6/10 on content tasks vs Claude's 9/10. So we expect:

- Ollama JSON compliance likely 60-85% (Claude is ~100%) — some responses will wrap JSON in markdown or add preambles
- Key-schema match rate 60-80%
- Subjective quality gap that a Claude judge will flag but not annihilate
- Latency 3-6 min per call vs Claude's ~5s

If those hold roughly, the takeaway is: mistral-small3.1 on CPU is not production-viable for page-content-writer without either (a) prompt-engineering adjustments on the replay path, or (b) fine-tuning a smaller / tuned model, which is exactly flywheel C. That result would be the ROI justification for the fine-tuning work, with a concrete failure-mode list to target.

If Ollama surprises us (90%+ JSON compliance, keys match, quality tolerable), we have a viable swap candidate now. Unlikely but worth measuring.

**Actual results (partial, 2026-04-22):**

| Case | Prompt chars | Prompt toks | Output toks | Wall time |
|---|---|---|---|---|
| 1 | 9,619 | 2,634 | 119 | **27 min** |
| 2 | 12,812 | ? | ? | Still running at last check (27m+) |

Case 1 real throughput: 119 tokens / 1636s = **~4 seconds per token**. Twice as slow as the preflight "hi" call implied. Root cause: **shared adapter contention**. Ollama logs show 3-6-second calls from other sources interleaved with our 27-min case, which means production workflows are competing for CPU during our eval.

Realistic eval time: 20 cases × 25-30 min = **10+ hours**. Beyond "absurd but fine" into "unlikely to finish today".

**Takeaway (even from partial data):** mistral-small3.1 on a shared cpu-ollama adapter is not a practical substrate for production-scale replay-eval. The adapter serves production traffic too and can't be monopolised. Two real paths forward:

1. **Dedicated eval adapter** — spin up a second ollama pod with its own PVC, model, and no production routing. Eliminates contention. One-time setup.
2. **GPU Ollama for evals and production** — the bigger structural move. mistral-small3.1 on a GPU would run case 1 in ~10-20 seconds, and Llama 70B becomes viable. This is what `020e` recommends anyway.

Current action: pause the eval after case 2 completes (or kill it), capture what we have (n=1 or n=2), document the throughput reality, move on. Option 1 or 2 is a separate decision.

**Dedicated eval pod (2026-04-22, refined 2026-04-23):**

Spun up `ollama-eval` as a sibling deployment to `ollama-adapter`, using the same kustomize pattern but with its own PVC (`ollama-eval-models-pvc`), service (`ollama-eval`), and deployment name. Only `mistral-small3.1` pulled — no nomic — to simplify init. Four manifests committed: `ollama-eval-pvc.yaml`, `ollama-eval-deployment.yaml`, `ollama-eval-service.yaml`, `kustomization.yaml`.

Kafka-scheduler does not auto-discover new Ollama services (it probes only endpoints listed in `ai_endpoint_health`), so this pod is invisible to production routing. Exactly what we want — it's ours.

**Memory issue (2026-04-23):** first re-warm attempt on the pod returned `"model requires more system memory (15.8 GiB) than is available (15.5 GiB)"` after the model had been idle long enough to evict. 20Gi pod limit was too tight against a 15.8GiB-resident-footprint model once Ollama's own overhead was counted. Bumped to `requests: 24Gi / limits: 28Gi`. Fix persisted into kustomize base. Lesson: **pod memory limit must be ≥ model file size + ~8-12 GiB headroom** for ollama CPU inference.

Measured after the bump: cold load 22.8s, first "say hello" call 2m52s total wall, warm call 6.9s — matches our earlier uncontended numbers.

**Resumed eval (2026-04-23):**

Runner pointed at the dedicated pod's tunnel. Skip-logic keeps case 1 (already in `results.jsonl` from earlier run) and processes cases 2-20. Per-case estimate now realistic: 2-3 min each, ~45-75 min total — much faster without production contention. `curl -m 1800` (30-min ceiling) left in place for safety but should not trigger.

### 2.4d-comparison Flywheel D — analysis methodology (drafted 2026-04-23, run after eval completes)

The 20 paired `(prompt, claude_response, ollama_response)` triples need a comparison framework. Three levels:

| Level | What | Cost | Information |
|---|---|---|---|
| 1. Structural | jq over results.jsonl — JSON validity, schema-key match, forbidden patterns, length distribution | seconds | Is Ollama format-compliant? |
| 2. Claude-as-judge | Pass pairs to Claude, ask for quality scoring (relevance, voice, fake-specifics) | ~20 API calls, <$1 | Is Ollama's content any good when it is format-compliant? |
| 3. Manual pair review | Eyeball specific pairs flagged as disputed or surprising | minutes | Sanity-check the judge's scoring |

**Level 1 automatic checks run per case:**

- `claude_valid_json`, `ollama_valid_json` — does `.message.content` parse as JSON
- `keys_identical` — do top-level JSON keys match
- `has_code_fence`, `has_placeholder_bracket`, `has_invented_metric`, `starts_with_preamble` — forbidden patterns per prompt rules
- `claude_headline_len`, `ollama_headline_len` (and subheadline) — length imitation

**Prediction baseline (from 020e doc, Mistral Small 3 scored 6/10 on content):**

- Ollama JSON valid: 70-90%
- Schema keys identical: 40-70%
- Forbidden patterns: occasional preambles and invented metrics expected
- Length roughly in the right ballpark

**The decision the analysis answers:** is the gap between Claude and local-model output on this specific agent's task (a) close enough to swap with prompt tweaks, (b) close enough to swap after fine-tuning, or (c) so wide that a different substrate (Llama 70B on GPU) is needed? Level 1 alone answers (a) vs not-(a). Levels 2-3 distinguish (b) from (c).

### 2.4e Flywheel A — training data export (design decided 2026-04-22)

**Format: ChatML messages with metadata sidecar.** Decided 2026-04-22 after weighing Alpaca, ChatML, and raw completion formats. Reasons:

1. We will fine-tune chat-tuned base models (Llama 3.x Instruct, Qwen Instruct, Mistral Instruct) which are trained on this format and learn faster with matching format.
2. Unsloth, Axolotl, LLaMA-Factory all default to ChatML messages.
3. Matches the `/api/chat` endpoint our Ollama adapter uses — training-inference parity.
4. Metadata sidecar keys (source_log_id, agent_type, step_name, orchestration_id, model, created_at) ignored by trainers but give us traceability from any training row back to its source.

**Shape:**

```json
{
  "messages": [
    {"role": "user", "content": "<prompt_rendered>"},
    {"role": "assistant", "content": "<cleaned response_text>"}
  ],
  "metadata": {
    "source_log_id": "<uuid>",
    "agent_type": "page-content-writer",
    "step_name": "process_sections_loop_iter_0_generate_content",
    "orchestration_id": "<uuid>",
    "model": "claude-sonnet-4-6",
    "created_at": "2026-04-20T10:15:06Z",
    "export_version": "1"
  }
}
```

Currently we treat the whole `prompt_rendered` as the user message — no system/user split because the stored prompt is one combined string. If the prompt template later gets a natural system section, we refactor. No data loss.

**Response cleaning** — we discovered (2026-04-22) that 96% of page-content-writer responses start with `{` (clean JSON) but 3.7% (~320 of 8,700 rows) are wrapped in markdown code fences. Exporting those as-is would teach the fine-tuned model to wrap JSON in code fences, exactly the behaviour the production prompt tells the model NOT to do. The export must strip code fences when present. 4 oddball responses (~0.05%) need manual inspection before the first export run.

**`agent_type` column status (checked 2026-04-22):** historically empty rows exist but recent writes are clean — 100% of last-2-day writes have agent_type populated, matching `orchestration_states.owner_agent_type`. No code fix needed. For the export, we either filter to rows where `agent_type IS NOT NULL AND agent_type != ''` or join to `orchestration_states.owner_agent_type` to backfill historical rows. Choice depends on whether we want the biggest possible dataset or just recent-clean.

**Action items:**

- [x] Inspect oddball responses. iter_0 has none; iter_1-8 has 4 that are intelligent responses to edge cases (testimonial skips, schema clarifications). These are not training material for SFT — see "Negative examples" below.
- [x] Decide dataset scope on export — filter to `model = 'claude-sonnet-4-6'` only (drops 34 haiku legacy rows). No join to orchestration_states needed for iter_0 (all recent rows have agent_type populated).
- [x] Choose architectural approach — **agent + action, not ad-hoc SQL**. See §2.4f.
- [ ] Deploy `training-data-exporter` agent + `training_data_export` action to production chassis.
- [ ] Run first export for `page-content-writer` iter_0. Expected ~1,889 rows after invalid-JSON filtering.

**Negative examples / edge cases (2026-04-22):**

Plain supervised fine-tuning (SFT, what Unsloth/Axolotl do by default) treats every
training example as something to imitate. There is no "don't do this" signal. Edge-case
rows where Claude produced prose-explanations instead of JSON (respecting
`on_missing: skip_section` etc.) would teach the fine-tuned model to sometimes produce
prose too. For SFT, these are positive examples of the *wrong* shape and must be
excluded.

Where edge-case outputs *do* become training value:

- **DPO (Direct Preference Optimization)** — pairs of (preferred, rejected) outputs,
  model learns to prefer one. This is where "negative examples" formally live. Requires
  pairs, done after SFT.
- **Constitutional AI / RLHF** — reward signal over behaviour.

For our first training run: plain SFT, edge cases excluded. When/if we reach DPO territory
(after SFT works and we want to polish edge-case behaviour), these same rows become
valuable as the "rejected" side of preference pairs. Keep them in `llm_call_log`,
just exclude from this export.

**iter_0 dataset audit (2026-04-22):**

| Category | Count | % | Action |
|---|---|---|---|
| Clean JSON (starts with `{`) | 1,919 | 97.4% | Include as-is |
| Fenced JSON (starts with ` ``` `) | 51 | 2.6% | Strip fences, include |
| Oddball (prose) | 0 | 0% | N/A in iter_0 |
| **Total** | **1,970** | | |

Export produces ~1,970 training rows. Well above the 200-example threshold for a
first SFT experiment. All rows either start with `{` or are code-fenced — the export
regex handles both; no prose-exclusion clause needed for iter_0 specifically.

### 2.4f Flywheel A — architectural decision: agent + action (2026-04-22)

**Decision: training export lives as a chassis action + agent, not ad-hoc SQL.** The
flywheel is a recurring operation — every agent/step we ever want to fine-tune will
need exporting, every few weeks/months. Building this as a proper pipeline component
now pays off every subsequent time. Also tried SQL regex fence-stripping and hit
Postgres regex quirks that Go string handling sidesteps cleanly.

**Why agent, not standalone Go CLI:**

Reuse existing Go helpers (`stripMarkdownFromResponse` in `ai_actions.go`), operate
through existing infrastructure (Kafka triggers, orchestration state, `llm_call_log`),
and stay consistent with "every agent is an orchestrator". Standalone CLI would need
its own connection management, auth, deployment, and would duplicate the cleaning
helper.

**Design:**

| Component | What it does |
|---|---|
| **`training_data_export` action** | New action in `platform/orchestration/actions/`. Reads config (agent_type, step_name, output_path, filters). Queries `llm_call_log`, cleans responses via `stripMarkdownFromResponse`, validates JSON per-row if `strict_json: true`, writes NDJSON to `output_path`. Returns summary stats. |
| **`training-data-exporter` agent definition** | Single-workflow agent with `export` → `complete` steps. Input parameters (agent_type, step_name, model_filter, output_path) come from `input_data` via `{{.input_data.X}}` templating in the step config. Same pattern as other parameter-taking agents. |
| **Kafka trigger script** | Mirrors the rag-test-agent pattern. `kcat` publishes to `system.agent.generic.requests`. |

The action's config accepts:

```
{
  "agent_type":           "page-content-writer",        // required
  "step_name":            "iter_0_generate_content",    // required
  "output_path":          "/tmp/training_exports/...",  // required
  "model_filter":         "claude-sonnet-4-6",          // optional, default empty = no filter
  "include_fenced":       true,                         // optional, default true
  "strict_json":          true,                         // optional, default true
  "min_response_length":  10,                           // optional
  "max_rows":             100000                        // optional safety cap
}
```

Action returns:

```
{
  "rows_seen":                  1970,
  "rows_exported":              1889,      // after strict_json filter drops ~9 malformed
  "rows_skipped_invalid_json":  9,
  "rows_skipped_scan_error":    0,
  "rows_skipped_marshal_error": 0,
  "output_path":                "/tmp/training_exports/...",
  "file_size_bytes":            23456789,
  "agent_type":                 "page-content-writer",
  "step_name":                  "iter_0_generate_content",
  "model_filter":               "claude-sonnet-4-6",
  "format":                     "chatml",
  "export_version":             "1"
}
```

**Why synchronous rather than spawning a sub-agent:**

Export of 2,000 rows takes a few seconds. Even 100,000 rows under a minute. No benefit
to async orchestration at this scale. If a future export needs streaming into S3 over
hours, we'd reconsider.

**Output location for now:**

Writes to a path inside the agent-chassis pod that happens to pick up the message.
Operator retrieves via `kubectl cp <pod>:/tmp/training_exports/<file> ./`. Ephemeral
— if the pod restarts before the operator retrieves, the file is gone but easily
regenerable. Proper S3/Backblaze upload is a follow-up (add `output_destination: "s3"`
in the action, reuse `params.StorageClient` same way med-price screenshots do).

**Why export_version in metadata:**

Dataset format might evolve — new metadata fields, different message structure,
system/user split later. Versioning each row now means downstream training tooling
can check compatibility before consuming. Cheap, forward-compatible.

**Files produced (v1 → v2):**

- `flywheel_A/training_data_export.go` — v1, reads from static config (superseded)
- `flywheel_A_v2/training_data_export_v2.go` — v2, reads from `params.CollectedData["input_data"]` via `datahelpers.ExtractNestedFieldString`
- `flywheel_A/registry_patch_training_export.go.txt` — 5-line addition to registry.go
- `flywheel_A_v2/create_training_data_exporter_agent_v2.sql` — agent definition with static `config: {}` (v1's `{{.input_data.X}}` templating did not render for deterministic actions)
- `flywheel_A/trigger_training_data_exporter.sh` — Kafka trigger helper (unchanged between v1/v2)

**Smoke test on v1 (2026-04-23) — outcome and lesson:**

Triggered exporter with `max_rows: 5`. Orchestration completed successfully, `rows_exported: 0`. Summary contained literal strings like `"agent_type": "{{.input_data.agent_type}}"`, meaning the chassis did not render template variables in step config for this deterministic action. The action pipeline itself worked end-to-end (registry, Kafka dispatch, execution, result write-back, orchestration complete) — only the parameter-passing was wrong.

**The fix (v2):** action reads parameters from `params.CollectedData["input_data"]` directly using `datahelpers.ExtractNestedFieldString`. This matches the convention used elsewhere in the codebase (confirmed via grep on existing actions). Import path is `github.com/gqls/agentchassis/platform/orchestration/datahelpers` — datahelpers is a sibling package to actions under `platform/orchestration/`, stdlib-only (no circular risk). Agent definition simplified: step config is `{}`, parameters flow from Kafka message `input_data` into `CollectedData`, action reads them there.

**Deployment status (2026-04-23):** agent definition v2 applied to DB. v2 Go file committed/pushed, chassis v1.0.983 deployed.

**v2 smoke retest (2026-04-23):** kcat heredoc was silently mis-routing messages to a "No-op scheduled task" handler — `__raw_message__.input_data: null`, `agent_config.processing_mode: "task"`, target agent never invoked. Root cause: multi-line `<<JSON ... JSON` heredoc in the trigger script mangles JSON before kcat sends. Fix: use shell here-string `<<<'...'` with flat single-quoted JSON. Documented as permanent ops pattern in `016_debugging_guide_v2.md` §9. Once fixed, v2 smoke test returned `rows_seen: 5, rows_exported: 5, file_size_bytes: 44567, agent_type: "page-content-writer"` — real values, not template strings. v2 works.

**v2 full export (2026-04-23):**

| Field | Value |
|---|---|
| rows_seen | 1,951 |
| rows_exported | 1,949 |
| rows_skipped_invalid_json | 2 |
| rows_skipped_scan_error | 0 |
| rows_skipped_marshal_error | 0 |
| file_size_bytes | 21,339,893 |
| output_path | `/tmp/training_exports/page_content_writer_iter0.jsonl` |
| model_filter | `claude-sonnet-4-6` |

Two rows skipped as invalid-after-cleaning, 1,949 clean training records produced. This is a real dataset, actionable for flywheel C.

**Operational gotcha discovered (2026-04-23):** the export action ran inside one of the three permanent `agent-chassis-*` pods — not a dedicated spawned pod as we assumed. File landed at `/tmp/training_exports/` on a specific chassis pod (randomly one of the three, whichever picked up the Kafka message). Other two chassis pods had no directory. This is fine for 21MB but creates tight coupling: file retrieval requires polling all three pods to find it, and pod restarts would lose the data. `site-adoption-agent` has identical `processing_mode: "orchestrator"` but DOES spawn a dedicated pod — the placement of `processing_mode` at the top level of `default_config` vs inside `workflow` may be the differentiator. To be verified before v3 code.

### 2.4g Flywheel A — v3 design with Postgres storage (decided 2026-04-23)

**Decision: training exports write to Postgres, not to a file.** Rationale:

- Per-iter_0 is 21MB. Full page-content-writer export (iter_0 through iter_8) estimated 250MB. Full sweep across all agents 1-2GB. Over 6 months accumulating exports for retraining cycles: 5-15GB. Well within Postgres TOAST-handled JSONB range.
- We already pay for Postgres; adding an S3 integration adds a second storage system to operate, credential-manage, and back up.
- Direct SQL for "what was exported, when, by whom, with what filter" — no joining two systems.
- Training tools (Unsloth/Axolotl) don't need random access into the dataset — they consume NDJSON sequentially. We write out to a temp file at training time with `\copy` or a small exporter action.
- Own schema (`training_exports`) keeps the data cleanly separated from operational tables. Easy to drop/reset independently.

**Schema:**

```sql
CREATE SCHEMA training_exports;

CREATE TABLE training_exports.runs (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_type       TEXT NOT NULL,
    step_name        TEXT NOT NULL,
    model_filter     TEXT,
    rows_seen        INTEGER NOT NULL,
    rows_exported    INTEGER NOT NULL,
    rows_skipped     JSONB NOT NULL DEFAULT '{}'::jsonb,  -- {invalid_json: N, scan_error: N, ...}
    format           TEXT NOT NULL DEFAULT 'chatml',
    export_version   TEXT NOT NULL DEFAULT '1',
    size_bytes       BIGINT NOT NULL,
    triggered_by     UUID,                                 -- optional caller reference
    orchestration_id UUID,                                 -- the orchestration that produced this
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE training_exports.rows (
    id         BIGSERIAL PRIMARY KEY,
    export_id  UUID NOT NULL REFERENCES training_exports.runs(id) ON DELETE CASCADE,
    row_index  INTEGER NOT NULL,
    messages   JSONB NOT NULL,     -- ChatML messages array: [{role, content}, {role, content}]
    metadata   JSONB NOT NULL      -- source_log_id, step_name, created_at, model, export_version
);

CREATE INDEX idx_training_exports_rows_export
    ON training_exports.rows(export_id, row_index);

CREATE INDEX idx_training_exports_runs_agent_step
    ON training_exports.runs(agent_type, step_name, created_at DESC);

-- Protect against accidental duplicates within a single export
CREATE UNIQUE INDEX ux_training_exports_rows_export_source
    ON training_exports.rows(export_id, (metadata->>'source_log_id'));
```

**Additional design choices (locked in 2026-04-23):**

- Unique index `(export_id, metadata->>'source_log_id')` prevents accidental duplicates if an export is re-run or resumed. Different exports can contain the same source row — that's normal for snapshots.
- `completed_at` is nullable. Failed/incomplete exports leave a row with `completed_at IS NULL` for diagnostic purposes rather than being rolled back silently.
- `source_notes` TEXT field for free-form provenance (e.g. `"backfilled from JSONL 2026-04-23"`).
- `recent_runs` view gives an at-a-glance "show me the last 100 exports" query.
- Indexes: `(agent_type, step_name, created_at DESC)`, `(model_filter, created_at DESC) WHERE model_filter IS NOT NULL`, `(export_id, row_index)`, plus the unique constraint above.

At training time, stream out the NDJSON when needed:

```sql
\copy (
    SELECT jsonb_build_object('messages', messages, 'metadata', metadata)
    FROM training_exports.rows
    WHERE export_id = '<UUID>'
    ORDER BY row_index
) TO '/tmp/training.jsonl' WITH (FORMAT text)
```

**v3 components (in order):**

1. **Schema migration** — `flywheel_A_v3/001_training_exports_schema.sql`. Creates `training_exports` schema, `runs` + `rows` tables, indexes, view. Idempotent (`CREATE ... IF NOT EXISTS`). **Status: written, ready to apply.**
2. **Action rewrite** — `training_data_export_v3.go`:
   - Same input contract as v2 (parameters via `input_data`)
   - Opens transaction via `params.DB.BeginTx(ctx, nil)` — pattern confirmed from 4+ existing actions
   - Inserts `training_exports.runs` with initial placeholder counts → gets `export_id UUID`
   - Streams rows from `llm_call_log`, cleans via `stripMarkdownFromResponse`, validates JSON
   - Batched multi-VALUE INSERTs into `training_exports.rows` (500 rows/batch)
   - Updates `runs` row with final counts + `completed_at`
   - Commits
   - Returns `export_id` in result map (no `output_path`, no `file_size_bytes`)
   - **Status: pending.**
3. **Agent definitions:**
   - **Worker `training-data-exporter`** — class=`specialist`, processing_mode=`task`, runs in a spawned pod, does the DB work
   - **Orchestrator `training-data-export-orchestrator`** — class=`orchestrator`, processing_mode=`orchestrator` at top level of default_config, minimal workflow: `spawn_agent` → `call_agent` → `complete_workflow`
   - Trigger script points at the orchestrator (not the worker directly)
   - **Status: pending.**
4. **One-off seeder** — loads the 1,957-row JSONL we already retrieved into `training_exports.rows` as a named seed export with `source_notes = 'backfilled from JSONL 2026-04-23'`. Preserves the dataset we have so we don't re-export just to test the v3 schema.
   - **Status: pending.**

**Scheduling: pure manual-trigger for now.** Decided 2026-04-23. Considered automating as a nightly scheduled task, but chose simplest: one agent, manually triggered when needed. Easy to add a thin `scheduled_tasks` row later that fires the same Kafka trigger if we find ourselves running exports often.

**Considered and rejected: real-time streaming into `training_exports.rows`.** Would couple observability (`llm_call_log`) to training-pipeline logic on the hot path, lose snapshot boundaries (hard to reference "the dataset we trained on"), and duplicate filtering between ingest-time and training-time. Batch exports give us named versioned snapshots we can A/B across retraining cycles — better match for the "occasional LoRA fine-tunes with progressively cleaner data" use case. See changelog entry for the full analysis.

**Dataset profile observed in first real export (2026-04-23, page-content-writer iter_0, claude-sonnet-4-6, n=1,957):**

| Prompt size (chars) | Assistant size (chars) |
|---|---|
| min 5,575 / p25 7,089 / **p50 8,250** / p75 9,825 / mean 9,430 / max 22,152 | min 171 / p25 398 / **p50 448** / p75 518 / mean 512 / max 6,847 |

**Schema heterogeneity observed** — one `(agent_type, step_name)` combo covers three component types:

| Schema | Rows | % | Component |
|---|---|---|---|
| `headline, primary_cta, primary_cta_url, secondary_cta, secondary_cta_url, subheadline` | 1,330 | 68% | Hero with CTAs |
| `headline, subheadline` | 349 | 18% | Minimal hero |
| `accent_color, cta_text, cta_url, logo_text, logo_url, nav_items, primary_color` | 181 | 9% | Header/nav |
| Others (mixed/tool-specific with 30+ bespoke keys) | 97 | 5% | Tool pages and edge cases |

Training implication: the model must learn "Component: hero → hero JSON; Component: header → header JSON" conditional on prompt content. The prompts already include "Section Requirements / Component: X" upfront. For a first SFT run, may be worth filtering to the top-2 schemas (1,679 rows = 86% of dataset) for clean-signal, then widening to the full set in a second pass.

**Time distribution:** 98 rows March / 1,859 April. Recent-heavy, which matches the intuition that April prompt templates are the ones we want to imitate.

### 2.4h Flywheel A — v3 implementation notes (2026-04-23)

**doc 001 review findings — canonical patterns I'd initially missed:**

1. `ExtractActionInputs` + `ActionInputSpec`. The canonical three-layer input handling pattern (config → input_fields → flat lookup, with deprecated-field mapping and defaults). My initial v3 used direct `ExtractNestedFieldString` calls which work but bypass default handling and the input_fields layer. Rewritten to use the canonical pattern — `inputs.Get("agent_type")`, `inputs.GetBool("strict_json", true)`, etc. matches every other action in the codebase.
2. Three confused columns on `agent_definitions`:
    - `category` — free-text functional role (`orchestrator`, `specialist`, `analyst`, etc.)
    - `agent_category` — CHECK-constrained to one of 6 values (`strategist`, `executor`, `analyst`, `integrator`, `coordinator`, `specialist`) — **NOT** `orchestrator`
    - `status` — lifecycle (`active`, `experimental`, `deprecated`, `demo`, `template`)
    Naïve writes put `experimental` in `category` (wrong slot). `improvement-loop` is the reference row: `category=orchestrator`, `agent_category=coordinator`, `status=experimental`.
3. Orchestrator wrapper pattern — `spawn_agent → call_agent → complete`. Not one step, three. `target_role` (not `agent_type`) for `call_agent` lookup. Optional input_mapping fields suffixed with `?` to avoid hard failures when caller omits them.
4. `GetBoolField` / `GetIntField` in datahelpers already exist — my `extractBoolWithDefault` / `extractIntWithDefault` duplicated them. Replaced with canonical.

**Spawning confirmed working (2026-04-23 trigger test):**

First v3 trigger failed with `required input_data.output_path is missing or empty` — but that error comes from v2 code, which was still running (agent definitions had been updated to v3 via SQL, but the Go binary hadn't been rolled out yet). The orchestrator wrapper DID spawn the worker correctly:

- Worker pod appeared: `agent-training-data-exporter-4c4fe86e-tlg6s`
- `__execution_context__.sender.role = "exporter"` — spawn+call linkage worked
- parent_orchestration_id wiring correct (child orch returned failure to parent cleanly)
- `processing_mode: "orchestrator"` at top level + `agent_category = coordinator` IS the combination that produces a dedicated spawned pod

This confirms the architectural decision. When the v3 binary rolls out, the same trigger will succeed.

**Five final files in `flywheel_A_v3/`:**

| File | Status |
|---|---|
| `001_training_exports_schema.sql` | applied and verified |
| `002_create_training_export_agents_v3.sql` | applied and verified |
| `training_data_export_v3.go` | written, awaiting chassis build |
| `trigger_training_export_v3.sh` | tested (confirmed spawning works, failed on stale binary) |
| `003_backfill_jsonl_to_postgres.sh` | ready to load the 1,957-row JSONL we have on disk |

**Reference patterns followed:**

- Worker agent definition: `specialist` category, `processing_mode: task` at top level
- Orchestrator agent definition: `orchestrator` category (free text), `coordinator` agent_category (constrained), `processing_mode: orchestrator` at top level
- Action input extraction via `datahelpers.RegisterActionInputSpec` in `init()` + `ExtractActionInputs` at call time
- `target_role` not `agent_type` for call_agent lookup
- `?`-suffixed optional input mapping fields
- Orchestrator timeout > worker timeout (1800 > 1200)

### 2.4i Flywheel A — v3.1 → v3.2 runtime learnings (2026-04-23)

**v3.0 first trigger failed: "bulk insert 500 rows: driver: bad connection"**

After the orchestrator wrapper landed, the worker spawned correctly in a dedicated pod, started executing, then died during the first bulk-insert flush. Error was at the Go SQL driver level — pgbouncer-backed connection went bad before the 4.5MB INSERT could land.

Diagnosis session covered:
- Postgres `statement_timeout`, `idle_in_transaction_session_timeout`, `lock_timeout`, `tcp_keepalives_*` — all `0` (disabled). Not postgres-side.
- pgbouncer log entries: `closing because: server lifetime over (age=3600s)`, `client_idle_timeout (age=604s)` — connection management happening, but not at the 33-second failure point.
- Chassis connects via `pgbouncer.ai-persona-system.svc.cluster.local:6432` — every statement passes through pgbouncer. Pool mode likely `transaction`.

Root cause hypothesis: the combination of a long-held transaction (single tx around the whole 2000-row export) and pgbouncer's pool management produced the bad-connection state. Specifically: holding a 4.5MB multi-VALUE INSERT open on a pooled connection for ~30 seconds before the first flush commit.

**v3.1: split into per-batch transactions, batch size 500 → 100.**

Architecture changed from:
```
BeginTx → INSERT runs → query → loop(flush batch within tx) → UPDATE runs → Commit
```
to:
```
INSERT runs (no tx) → query (no tx) → loop(BeginTx → INSERT batch → Commit per-batch) → UPDATE runs (no tx)
```

Each batch tx completes in well under a second. Runs row insert and final update are single-statement non-tx operations. No long-held connections.

**v3.1 result — partial success:**

- 1,958 rows landed in `training_exports.rows` with sequential `row_index` 0-1957. Per-batch commits worked.
- Worker action returned success. Child orchestration COMPLETED cleanly.
- Parent orchestration COMPLETED cleanly.
- Action result map had correct counts: `rows_exported:1958, rows_seen:1960` (2 invalid-JSON rows skipped as expected).
- **BUT**: `training_exports.runs` row had `rows_exported=0, size_bytes=0, completed_at=NULL`.

The UPDATE step silently didn't take effect. Possibilities:
1. UPDATE threw an error that was swallowed by the v3.1 "Warn and continue" handling. Log grep for "final counts update failed" returned nothing in the abbreviated log file (but full logs could have it — the upload was explicitly abbreviated).
2. UPDATE returned success with `RowsAffected=0`. Unlikely but possible if there's some pgbouncer transaction visibility issue or UUID-casting edge case.

Manually reconciled the runs row with the correct counts; dataset is usable.

**v3.2: strict error handling on final UPDATE.**

Changes:
- Log line before UPDATE (shows export_id, rows_exported, size_bytes)
- `err` from `ExecContext` now returns action failure (was Warn+continue)
- `RowsAffected` checked — if != 1, action fails with diagnostic message
- Log line after UPDATE confirms rows_affected count

Next run will be loud either way; we'll learn definitively what the UPDATE is doing.

**Spawning architecture — fully confirmed working:**

- Parent orchestrator (`training-data-export-orchestrator`, class=`orchestrator`, agent_category=`coordinator`, `processing_mode: orchestrator` at top level) runs in the main agent-chassis pool.
- Worker (`training-data-exporter`, class=`specialist`, agent_category=`specialist`, `processing_mode: task`) gets its own spawned pod (`agent-training-data-exporter-<id>-xxxxx`) every run.
- `spawn_agent → call_agent(target_role=exporter)` pattern works end-to-end.
- Input mapping with `?` suffix for optionals passes parameters correctly.
- File-output concern from v2 is resolved — data lives in Postgres, pod lifecycle no longer matters.

**Current state (end of 2026-04-23 session):**

- Working dataset in Postgres: export_id `fef7be6b-887f-4bc9-b118-a5a9992c4179`, 1,958 rows, 21.2MB JSONB storage, reconciled with correct counts.
- v3.2 action code written, awaiting chassis rebuild/deploy.
- Backfill script ready to load the JSONL copy we retrieved earlier (for preservation as a second named export).

### 2.5 Fine-tuning path — flywheel C scripted (2026-04-23)

Pipeline shape now concrete, scripts written, awaiting first training run on
a GPU VM.

```
training_exports.runs + .rows  (1,958 rows landed, export_id 146a9a12-...)
    │
    ▼  01_pull_dataset_from_postgres.sh (kubectl exec + COPY TO STDOUT)
/workspace/training_iter0.jsonl on GPU VM
    │
    ▼  02_train_llama_3_3_70b.py (Unsloth QLoRA)
LoRA adapter (~150MB) at /workspace/lora_iter0_full
    │
    ▼  03_inference_test.py (sanity — 5 samples, JSON validity check)
Confident the adapter works
    │
    ▼  [future] merge + export GGUF, or serve via vLLM, or upload to Together/Fireworks
Deployed model, A/B against Claude
```

**Base model: Llama 3.3 70B Instruct.** Decision taken 2026-04-23. Not 3.1 —
3.3 at 70B matches 3.1 405B quality per Meta benchmarks, released Dec 2024,
same VRAM profile as 3.1 70B. Fits on single H100 or A100 80GB with QLoRA
(~50-70GB peak VRAM). Batch size 1, grad accum 8, 3 epochs, lr 2e-4, LoRA
rank 16, targets all linear layers. Use
`unsloth/Llama-3.3-70B-Instruct-bnb-4bit` for 4× faster download.

**Why 70B when 8B would be cheaper:** a narrow structured-JSON task like
page-content-writer iter_0 doesn't need 70B's capacity. 8B likely delivers
95% of the quality at 10% of inference cost. The 70B choice was made
because the hardware was already available for it, and having a strong
baseline is useful even if we later rebuild on 8B. Future run plan: train
both on the same dataset, compare. The infrastructure scripts work for
either size (change one MODEL_NAME constant).

**Files in `flywheel_C/`:**

| File | Purpose |
|---|---|
| `00_vm_setup.sh` | Python 3.12 venv + CUDA-matched torch + Unsloth + deps, VRAM verification |
| `01_pull_dataset_from_postgres.sh` | Stream training data direct from Postgres to local JSONL via kubectl exec |
| `02_train_llama_3_3_70b.py` | Main training script. CLI-configurable, writes adapter + manifest |
| `03_inference_test.py` | Sanity check — generate N outputs from training prompts, check JSON validity |
| `README.md` | Prereqs, sequence, expected timings, what-to-tune table |

**Run sequence (~30-90 min from scratch):**

1. `./00_vm_setup.sh` (one-off, 5-10 min)
2. `source ~/unsloth_env/bin/activate`
3. `./01_pull_dataset_from_postgres.sh <export_id> /workspace/training_iter0.jsonl`
4. Smoke train: `python 02_train_llama_3_3_70b.py --limit 20 --epochs 1 ...`
5. Smoke inference: `python 03_inference_test.py --n 3 ...`
6. Full train (30-90 min): `python 02_train_llama_3_3_70b.py` with defaults
7. Final inference: `python 03_inference_test.py --n 5 --skip 1900` (roughly held-out)

**Expected signals for "first run worked":**

- `train_loss` drops from ~2.5 → ~0.5-1.0 during training
- VRAM peak 50-70GB, well under 80GB ceiling
- Inference outputs parse as JSON without code fences
- Keys match trained schemas (hero-with-CTAs 68%, minimal hero 18%, header/nav 9%)
- Content is topical to the brief, not generic placeholder text

**After first success — remaining flywheel C work (not scripted yet):**

- Proper evaluation via flywheel D's 20-case test harness
- Claude-as-judge quality scoring vs Claude vs ollama-mistral-small3.1 vs trained Llama-70B
- Deployment decision: GGUF+Ollama (CPU, slow) vs vLLM on GPU VM vs hosted (Together/Fireworks)
- Chassis integration as alternative `ai_service` provider
- Production cost-per-page comparison

### 2.5.1 Phase 2 — automating flywheel C (design locked, not built)

Design decided 2026-04-23. Chassis drives, GPU VM serves. The GPU VM becomes
a compute service the chassis calls, rather than a host that needs to reach
into the cluster. This keeps credentials out of the VM, makes the VM easy
to rebuild or replace, and fits the existing orchestrator/agent pattern.

**VM side — HTTP job server (Option B chosen, ~200 lines of Python):**

- `POST /jobs` — body contains dataset (base64 or URL) + hyperparameters, returns `job_id`, spawns training subprocess via `02_train_llama_3_3_70b.py`
- `GET /jobs/{id}` — status, progress, final loss, timestamps
- `GET /jobs/{id}/adapter` — downloads the resulting LoRA adapter (tarball)
- Bearer-token auth
- SQLite or in-memory jobs dict
- Systemd unit to keep it running
- TLS via Caddy auto-TLS or cloud LB

Rejected alternatives:
- **SSH + remote exec** — simpler but synchronous, chassis must hold a 30-90min connection, key management awkward
- **Kafka consumer on VM** — cleanest messaging fit but needs VM → Kafka connectivity (public ingress or VPN), biggest setup cost, overkill for one consumer

**Chassis side — three new components:**

1. `model-trainer` (specialist agent) — takes `export_id` input, fetches rows from `training_exports`, builds JSONL, POSTs to VM's `/jobs`, polls until done, downloads adapter, records in `model_training_runs`
2. `model-evaluator` (specialist agent) — runs the flywheel D test harness against a newly trained adapter, compares to Claude baseline, writes scores
3. `training-flywheel-orchestrator` (orchestrator wrapper) — chains: spawn+call training-data-exporter → spawn+call model-trainer → spawn+call model-evaluator → conditional `swap_agent_model` if score ≥ baseline threshold, else record without deploying

**Schema additions:**

```
model_training_runs    (id, export_id FK, adapter_path, final_loss,
                        train_runtime_s, lora_r, lora_alpha, base_model,
                        hyperparameters JSONB, created_at, completed_at,
                        error TEXT NULLABLE)

model_artefacts        (id, training_run_id FK, storage_path, size_bytes,
                        sha256, created_at)

model_evaluations      (id, training_run_id FK, test_suite, baseline_model,
                        scores JSONB, deployment_decision TEXT, created_at)
```

**Trigger:** either a `scheduled_tasks` row running daily/weekly, or an event
trigger when `training_exports` gains N+ new rows since the last training run.

**Agent count delta for phase 2:** +3 new agent definitions (one orchestrator
wrapper, two specialists). Pattern mirrors training-data-export-orchestrator +
training-data-exporter already built today.

**Preconditions before building phase 2:**

1. Phase 1 complete — manual training run proves the scripts work
2. Phase 1 produces a baseline LoRA + eval output we trust as the "working" reference point
3. Flywheel D eval harness is either working or we know why it's paused
4. VM has stable public endpoint (DNS, TLS)

### 2.6 Fine-tuning candidates (priority order)

| Agent | Output type | Why local wins |
|---|---|---|
| knowledge-extractor | Structured JSON from prose | Runs repeatedly, well-defined schema, data accumulates during research |
| site-classifier | JSON classification | Every build hits it, output small |
| vet-practice-verifier | JSON boolean + evidence | High volume, yes/no + short extract |
| briefing-agent | JSON questionnaire from raw input | Lowest risk — swap to Mistral Small 3 already queued |
| domain-analyst | JSON metadata | Structured extraction |
| content-researcher | Search queries | Short query strings |

**Not good candidates**: page-content-writer (long creative output, quality
matters), visual-design-auditor (judgement), chief-strategist (structural
reasoning, one call per domain — worth the Claude cost).

---

## 3. Three improvement channels — they compound

From `009_model_infrastructure.md`, decision 10:

1. **RAG** — inject verified knowledge into prompts. Immediate effect. No training.
2. **LoRA** — train a local model to replicate an agent's task. Medium-term. Reduces per-call cost.
3. **Prompt evolution** — `prompt_variant` column lets us A/B prompts and pick winners. Ongoing.

Each is useful alone. Together they compound: good prompts with good RAG give
you the best training data, which produces the best fine-tuned model, which
needs a good prompt and good RAG to perform well.

---

## 4. Current state — what to do next on the internal flywheel

### 4.1 Ready-to-go
- [x] llm_call_log populating (flywheel columns wired through `ai_actions.go`)
- [x] Swap / revert functions deployed
- [x] Endpoint health routing deployed
- [x] Ollama CPU adapter up (nomic-embed-text + mistral-small3.1)

### 4.2 Open tasks (small, near-term)
- [ ] `agent_type` column empty in `llm_call_log` — `params.AgentType` not
      reaching `LogLLMCall`. Low priority but blocks clean training-data slicing.
- [ ] Schedule `cleanup_old_llm_logs()` (pg_cron or maintenance agent) —
      table will grow to ~1GB/month at current rates.
- [ ] Ensure `work_item_id` flows through orchestration context so "did the fix
      work" joins become possible.
- [ ] Periodic REINDEX of the ivfflat vector index on `knowledge_base`.

### 4.3 Open tasks (larger, flywheel completion)
- [ ] First real export: pick the agent with the most successful rows in
      `llm_call_log` — probably `page-content-writer` or `site-classifier` —
      and produce a JSONL file. Just the export, no training yet.
- [ ] Dry-run Unsloth end-to-end on a borrowed GPU to validate the flow
      (we've never actually run it).
- [ ] Write an A/B scoring script — produce outputs from Claude and the local
      model for N examples, let a reviewer (human or Claude judge) rank pairs.
- [ ] Bring the GPU Ollama endpoint up reliably (`ollama-gpu` currently DOWN).
      ThunderCompute H100 works but is stopped and manual to start.

### 4.4 Decisions to make
- **Which model as the fine-tune base?** Llama 3.1 8B is what the canine-biology doc assumes. Qwen/DeepSeek/Mistral are alternatives. Pick one to standardise on so we accumulate ops experience with a single stack.
- **Where does training happen?** ThunderCompute single-H100 worked in testing. Other options: RunPod, Lambda, local 3090/4090. Trade-off: spot cost vs queue time vs ops complexity.
- **On-cluster training pod or manual?** Short-term: manual on rented GPU. Long-term: a `training-job` agent that does export → upload → train → pull GGUF → install on ollama-gpu automatically. That's after A/B confirms the path works.

---

## 5. Product pivot: finetuning.uk as a self-service platform

**This is a product decision, not a technical one.** The technical bits are
shared with the internal flywheel, but building an external product is a
different company's worth of surface area.

### 5.1 What would "users can fine-tune from finetuning.uk" require?

| Layer | What's needed |
|---|---|
| Identity | Multi-tenant auth, org/team model, API keys, billing identity |
| Data | Upload (CSV/JSONL), preview, validation, PII scanning, quota, retention, deletion (GDPR) |
| Training | Base model picker, hyperparameter picker (or opinionated defaults), job queue, GPU scheduling across tenants, eval metric display, checkpointing |
| Inference | Per-tenant endpoints, cold-start handling, rate limits, usage metering |
| Billing | Stripe, VAT, invoices, usage reconciliation |
| Ops | Support, docs, status page, abuse prevention, content moderation |
| Legal | ToS, privacy policy, DPA, UK/EU data residency commitments, base model licence compliance |
| UI | The actual finetuning.uk frontend — dashboards, job monitoring, chat playground |

### 5.2 Competitive context (April 2026)

| Player | Angle |
|---|---|
| Fireworks AI | Cheap (~$3 minimum training fee), fast LoRA serving, strong API DX |
| Together AI | 100B+ models, LoRA serving, checkpoint resume, enterprise trust |
| Nscale | SFT-first, speed, UK-presence |
| OpenAI / Vertex / Azure OpenAI | Managed on proprietary models, enterprise default |
| H2O LLM Studio | No-code UI, self-hostable, includes distillation |
| Predibase / Modal / Replicate / RunPod | Various tiers |
| Unsloth Studio (+ Rafay) | Self-host pattern — this is the engine most serverless services wrap |
| Azumo / Accenture / Deloitte | Consulting model — done-for-you |

### 5.3 Where differentiation could come from

Honest read: it is **not** a technically-novel product. Anyone can wrap Unsloth.
Differentiation has to come from positioning, not engineering.

Plausible angles:
- **UK / EU data residency** — easy story, legally meaningful, hard for US competitors to match without setting up local infra.
- **Opinionated simplicity** — 3-click fine-tuning for non-ML users. One base model choice, one dataset format. "If you know how to use Excel, you can fine-tune."
- **Vertical templates** — pre-built datasets and prompts for legal, medical, accounting, customer support. Buy a fine-tune recipe, add your data, go.
- **Self-improvement loop as a feature** — fine-tuned models whose quality keeps improving as users interact with them. This is **exactly** the flywheel we've built internally. It's our edge.
- **Free tier funded by data contribution** — users get free training in exchange for (opted-in) contribution to a shared knowledge base. Only works ethically with genuine opt-in and clear terms.

### 5.4 Concrete risks / things that will bite

1. **GPU supply is lumpy.** At the scale of "dozens of concurrent training jobs", ThunderCompute-on-demand is not reliable enough. Reserved capacity commits you to cost whether or not users arrive.
2. **Multi-tenant isolation is not optional.** A training job reading another tenant's data is an extinction-level bug.
3. **Users upload bad data → bad models → blame you.** Needs strong pre-flight validation and clear expectation-setting.
4. **Base model licences are hairy.** Llama has MAU thresholds; some bases forbid commercial fine-tuning outputs; Mistral varies by model.
5. **Inference hosting is its own product.** Selling fine-tuning without hosting means "download a GGUF and good luck". Selling hosting means running a GPU fleet 24/7 with cold starts solved.
6. **Billing against GPU seconds is error-prone.** Users will dispute. Need clear usage meters and generous grace.
7. **Support load.** People break things in creative ways. A support tier must exist from day one.
8. **Opportunity cost.** Every engineering week on finetuning.uk is a week not on the agent orchestration system (the primary product). Unless the sites-as-a-service business funds it, the pivot should be justified on its own economics.
9. **Claude/OpenAI get cheaper every quarter.** The "fine-tune your own to save money" pitch weakens as frontier API prices drop.

### 5.5 Staging — plausible phases if we do pursue this

Not committing; sketch only.

**Phase 0 — the site tells a story.**
Make finetuning.uk a credible industry site: clear positioning, guides,
comparisons, opinionated advice. No product yet. Works whether we ship a
product or not. Fits the existing agent-site-builder pipeline.

**Phase 1 — managed service, concierge.**
Offer fine-tuning as a bespoke service. Users talk to us, we run the job on
our infra, hand them a GGUF or host it. No self-serve UI. Validates demand
and pricing without building the product.

**Phase 2 — narrow self-serve.**
Pick one vertical (e.g., "fine-tune a customer-service model from your Zendesk
exports"). Build the minimum UI that does **just that** well. Prove
acquisition and retention economics on a narrow wedge before going broad.

**Phase 3 — general self-serve.**
Only after phase 2 has paying users. By this point we know what the UI
actually needs to be.

---

## 6. Overlap with the internal flywheel — what's reusable

| Piece | Internal flywheel | finetuning.uk product |
|---|---|---|
| Ollama GPU adapter | Serves local models | Serves tenant models (with isolation) |
| Unsloth script | Trains our LoRAs | Runs tenant training jobs |
| GPU provisioning | On-demand, manual | Automated, queued, multi-tenant |
| Training data export | `llm_call_log` → JSONL | Tenant upload → JSONL |
| Evaluation | Manual A/B | Per-tenant dashboards |
| llm_call_log schema | Our ops + training | Irrelevant to tenants (they have their own logs) |
| knowledge_base | Our RAG | Optional add-on: tenant RAG |

Most of the infra maps. What's entirely new is: **multi-tenancy, billing,
UI, support, legal.** That's where the product work lives.

---

## 7. Decisions taken (2026-04-21)

1. **Goal = both layers.** finetuning.uk is a generous, credible knowledge
   site **and** a revenue-generating service. Site links outward to credible
   third-party sources without gatekeeping. Generosity is positioning.
2. **Target user = technical-adjacent SMEs / agencies / ops leads.**
   Refined from earlier "non-technical business owners" — that audience has
   the worst economics for a solo operator (highest support load, lowest
   willingness to pay). One level up has budget and understands value.
3. **Hosting = full managed service.** We host the resulting model.
   No "here's a GGUF, good luck" tier.
4. **Infrastructure approach = separate cluster, reuse the framework.**
   Shared GPU pool with LoRA stacking, not per-tenant clusters. Training
   jobs queue onto shared training GPU(s). Tenant isolation via namespace
   + data encryption, not physical separation.
5. **Operator shape = solo, £9-12k/month target, full runway, appetite
   for immediate shipping.** Interim contract gigs acceptable at <50%
   of time to generate market signal and cash flow.
6. **Acquisition = cold only, so content-led.** No warm network to
   activate. finetuning.uk itself is the acquisition engine. The framework
   becomes the content engine that outpaces competitors who have to write
   manually.
7. **Positioning = "AI tailored to your business."** "Fine-tuning" read as
   the verb, not strictly LoRA. The site says what we actually do; the word
   redefines itself through the content.
8. **Posture on multi-agent = capability, not headline.** "18 agents" is
   marketing theatre. The product is the outcome; multi-agent is how we
   deliver it.
9. **Partner directory = strategic asset.** Positions us as honest broker.
   Protects against accepting work we can't deliver. Built from real
   relationships, not a scraped list.
10. **UI-first product build (revised).** Earlier hypothesis was "concierge
    first, UI later". Revised: build the UI as our own operational tooling
    for delivering the work. We're not building a SaaS speculatively —
    we're building our own cockpit that happens to also be a sellable
    product once it's good enough. If an SME joins early to shape it, great;
    if not, we use it ourselves on test data and first customers come to a
    working tool.
11. **Data curation is a first-class product feature, not a service
    add-on.** Most competitors treat bad data as the customer's problem.
    We use the framework to filter, deduplicate, quality-score, classify,
    redact PII, and flag inconsistencies automatically. This is the
    differentiator — most SMEs have tried and failed with RAG *because* of
    their data, not despite good data.
12. **Product-first flagship = RAG automation with built-in data
    curation.** RAG is chosen ahead of text LoRA and image LoRA because:
    (a) the user shows up with a folder of docs, not curated training
    pairs; (b) the infrastructure (knowledge_base, rag_index, rag_lookup,
    Ollama, embeddings) is all built; (c) the data-curation pipeline
    aligns naturally with RAG ingestion; (d) it's the broadest market.
    Text and image LoRA become later tiers once RAG is live and proven.

## 8. Offer structure (revised — product at the centre)

| Tier | Offer | Price | Notes |
|---|---|---|---|
| **Flagship product** | finetuning.uk RAG platform — upload docs, auto-curate, chat over them, API access | £199-1,499/mo subscription + £2-5k setup for concierge onboarding | The main product. Self-serve tier after month 3, concierge-onboarded until then. |
| Mid-range | Custom AI assistant over business knowledge — goes beyond platform defaults into bespoke integration | £5-10k setup + £800-1.5k/mo | For clients whose needs outgrow the platform defaults |
| High | Bespoke multi-agent workflow / fine-tuning / image LoRA | £15-30k project + £500-1.5k/mo | Enters after RAG platform is proven; includes text LoRA and image LoRA offerings as they come online |
| Legacy (existing) | AI-built industry sites + ongoing content ops | £3-5k setup + £1.5-3k/mo | Already demonstrated. Keep as offer but not the lead. |

Ratio at steady state: RAG platform carries ~60% of revenue once there are
5-10 subscribers plus occasional setup fees. Bespoke work tops up.

Plus entry-level productised items (diagnostic call, audit, ebook, tools)
that generate market signal without locking in direction.

### 8a. Data curation as a named feature

Data curation lives inside the RAG product as a visible feature, not a hidden
detail. The ingestion pipeline runs through framework-driven agents:

| Stage | What it does | Framework pieces reused |
|---|---|---|
| Parse | PDF/DOCX/HTML/CSV/URL → text | web-scrape adapter + doc parsers |
| Classify | Tag each chunk by topic, doc type, likely usefulness | agent with content-quality LLM call |
| Deduplicate | SHA256 + near-duplicate detection | existing dedup in rag_index |
| Quality-score | Flag low-value content (boilerplate, navigation, stale) | auditor-pattern agent |
| PII scan | Detect + optionally redact names, emails, IDs | new adapter (pattern-match + LLM review) |
| Inconsistency flag | Surface contradictions across docs | multi-agent comparison |
| Structure extract | Build outlines + topic maps | classification-extraction agent |

Each stage produces a visible report the user can review before the final
index commits. That transparency is part of the sell: they see what
they're getting, including what we chose to exclude and why.

## 9. Content pillars

Five pillars, one voice. Each earns SEO independently and funnels into
one of the offer tiers.

| Pillar | Funnel to |
|---|---|
| AI advisory (fine-tune vs RAG, vendor comparisons, honest guides) | Diagnostic calls + lead offer |
| Sites + content operations | Lead offer |
| Custom AI assistants / RAG over business data | Middle offer |
| Multi-agent workflows | High offer |
| Teaching / newsletter / courses | Products + general inbound |

Compound effect: the framework produces content, content grows the
`knowledge_base`, better knowledge produces better content. This loop is
our structural edge against competitors who produce content manually.

## 10. Shipping ladder (revised — UI-first for RAG platform)

Aspirational dates, not promises. The UI is being built as our own
operational cockpit — every feature has to justify itself by making
concierge delivery faster or better.

**Week 1**
- Reposition finetuning.uk via `site_specs` rewrite; regenerate core pages.
- Publish "Should you fine-tune?" decision guide on the repositioned site.
- "Book a diagnostic call" page live at £250/hr.
- **New:** decide auth stack (Clerk vs Supabase vs Kinde) and scaffold.
- **New:** add `tenant_id` to `knowledge_base` + enforce in `rag_lookup`/`rag_index`.

**Month 1 — MVP RAG platform, used internally only**
- Upload pipeline (PDF/DOCX/HTML/CSV/URL → `rag_index`).
- Chat UI with citations (calls `rag_lookup` → LLM).
- Source management (list, preview, remove).
- Deploy behind auth at `finetuning.uk/app`.
- Use it ourselves on test data (our own docs, public data).
- Still no paying customers yet. Framework first.

**Month 2 — Data curation pipeline becomes visible**
- Parse + classify + deduplicate + quality-score agents wired in.
- Curation report UI — user sees what was kept, dropped, flagged.
- PII detection (pattern-match first, LLM review later).
- Stripe integration + billing meter.
- First concierge customer onboarded (manually, we do the work, they use the UI).
- Free tools shipped (decision tools, cost estimators) for acquisition.

**Month 3 — Semi-self-serve**
- Integrations: Google Drive pull, Slack connector.
- Inconsistency flagging agent.
- API access for customers who want programmatic query.
- Second concierge customer.
- Case study from first customer.

**Month 4-6**
- Self-serve signup open (with generous trial).
- Text LoRA feature added as paid upgrade (for customers whose RAG plateau suggests LoRA would help).
- 3-5 paying customers.
- Product-page content engine running continuously.

**Month 6-12**
- Image LoRA feature added.
- Multi-agent workflow offering for customers outgrowing the defaults.
- Steady retainer book of 5-8 customers.
- Revenue approaching £9-12k target.

## 11. What NOT to ship immediately

Things that look like products but lock us into directions we'll regret:

- Self-serve multi-tenant fine-tuning SaaS (6+ month project; wrong audience)
- Public fine-tuning API
- Subscription "AI assistant for £99" product
- Any infrastructure without a paying customer asking for it

Things to ship immediately produce **market signal** without committing
direction: diagnostic calls, audits, tools, articles, ebooks, case studies.
Rule of thumb: if it doesn't teach us what the market wants, don't ship it
yet.

## 12. Interim gig discipline

Two rules on interim work:

1. Every interim gig should teach us something about the main thing.
   "Can I build an AI assistant for this law firm in 6 weeks?" is a gig
   that pays *and* validates the middle offer.
2. Cap interim work at 50% of available time. Rest goes to finetuning.uk.
   If interim expands to fill time, the main thing never compounds.

## 13. Still-open questions

1. Which base model(s) do we commit to when fine-tuning use cases arrive?
   Llama 3.x default; Qwen and Mistral as plausible alternatives.
2. Pricing model for the middle tier — setup + monthly, or pure monthly
   with setup amortised? Affects cash flow shape.
3. Data residency promise — "UK/EU only"? Legal promises become infra
   constraints.
4. How much of finetuning.uk itself can our agent pipeline build and
   maintain end-to-end — including content updates, new tools, case studies?
5. Which interim gig channels to target first (Upwork, Contra, direct
   agency outreach, AI-specific communities)?
6. Which free tool ships first — the one with best intent-qualification
   for our lead offer?
7. When does the FOCUS doc split — site content plan vs service delivery
   vs internal flywheel? Likely after we've shipped the first iteration.

## 14. Reusable operational patterns (accumulated 2026-04-21/22)

Habits and patterns earned during flywheel B and D. Keep referring back.

### Schema discipline

- **Always read `/mnt/project/some_schemas` before writing SQL against any table.** Don't guess column names. Don't assume `deploy/postgres-clients` — it's `pod/postgres-clients-0` (StatefulSet).
- Orchestration step outputs live in `orchestration_states.collected_data`, not `final_result`. `complete_workflow` doesn't copy to `final_result` automatically.
- `site_work_items` uses `item_type`, not `work_type` (confusing if coming from a generic "work_items" naming).
- `agent_definitions` `ON CONFLICT` must target `(type, version)`, not `(type)` alone.

### Running things against the cluster

- Port-forward once (`kubectl port-forward svc/... 21434:11434 &`), run many curls against localhost. Don't spawn a `kubectl run curl` pod per request — 10-20× slower and prone to CONTAINER_EXITED races on quick succession.
- Set a 30-minute (`-m 1800`) curl timeout when calling CPU inference. "Absurd" timeouts are correct when the realistic upper bound is minutes.
- When using heredocs in the terminal: multi-line content via `-At` TSV is unsafe. Use `row_to_json(...)` → NDJSON for anything that might contain newlines inside fields.

### Ollama specifics (cpu-ollama adapter)

- Adapter uses `/api/chat`, NOT `/api/generate`. Our eval must match to be representative.
- `OLLAMA_LOAD_TIMEOUT=10m` and `OLLAMA_KEEP_ALIVE=30m` (or 60m for eval) are env defaults in the kustomize base.
- Deployment strategy is `Recreate`, not `RollingUpdate` — the PVC is RWO.
- First inference after cold start = ~45s model load, then inference. Plan for this when any eval begins.
- Measured throughput: prompt eval ~150 tok/s, generation ~2.5 tok/s on mistral-small3.1 Q4_K_M in an 8-core CPU pod.
- **Pod memory limit must be ≥ model file size + ~8-12 GiB headroom.** mistral-small3.1 at 15.5GB model file needs at least 24GiB pod limit; 20Gi was 300MB short after Ollama's own overhead. Check this before picking pod resources for any new model.
- Ollama reads free memory from `/proc/meminfo` (host view) but constrains against the cgroup limit. This produces misleading "not enough memory" errors even when the node has plenty free — the cgroup is what matters.
- Nomic task prefixes (`search_document:` / `search_query:`) are load-bearing for retrieval quality; already patched into production `rag_index`/`rag_lookup` (2026-04-21).

### Thunder Compute API specifics (thunder-adapter)

- **Base URL:** `https://api.thundercompute.com:8443/v1`. Note the `:8443` port and the `/v1` path prefix — both required.
- **Authentication:** `Authorization: Bearer <token>`. Token lives in the `personae-default-secrets` k8s Secret under `THUNDER_COMPUTE_API_KEY`. Adapter reads it via env var at startup.
- **Canonical OpenAPI spec:** `https://api.thundercompute.com:8443/openapi.json`. Markdown index at `https://www.thundercompute.com/docs/api-reference/instances/create-instance.md` (and similar paths for delete/list/modify). Mintlify-built; the page-suffix `.md` returns the raw schema YAML embedded in the page, which is more useful than the rendered HTML.
- **Endpoints used by the adapter:** `POST /instances/create`, `GET /instances/list`, `GET /instances/{id}`, `POST /instances/{id}/delete`. Snapshots and ssh-key endpoints exist but are not currently wired.
- **`POST /instances/create` required body fields:** `gpu_type` (string, case-sensitive, e.g. `"A100"` not `"a100"`), `num_gpus` (int), `cpu_cores` (int — NOT `vcpus`), `disk_size_gb` (int), `mode` (enum: `"prototyping"` or `"production"`), `template` (string, e.g. `"ubuntu-22.04"`). `public_key` is optional. Missing any required field returns `400 invalid_request` with `"Invalid request body"` — no detail about which field. Our defaults: cpu_cores=4, disk_size_gb=100, template="ubuntu-22.04", mode="prototyping".
- **Field naming traps.** Our internal vocabulary historically used `gpu` and `vcpus`; Thunder uses `gpu_type` and `cpu_cores`. The chassis-side dispatch action accepts either spelling and normalises before passing to the adapter; the API client struct uses Thunder's names.
- **`POST /instances/create` response shape:** `{identifier: int, key: string, uuid: string}`. `identifier` is the numeric ID used for delete/get/modify. `uuid` is a parallel string ID. `key`'s purpose is undocumented — TODO confirm on first real call whether it's a server-generated SSH access key or something else.
- **Status casing on polling:** unverified. The polling loop in `WaitForRunning` compares against `"running"` (lowercase). If Thunder returns `"RUNNING"` the loop never terminates — switch to `strings.EqualFold` in `IsReadyStatus` if observed.
- **Spend gating lives in DB, not API.** The adapter calls `thunder_provision_check` (a Postgres view) before every create. The view sums `decommissioned_spend_24h + running_estimated_spend + estimated_new_run_cost_usd` and compares to `thunder_config.daily_cap_usd`. Defaults: cap=$100, per-run estimate=$25. For testing, drop the per-run estimate to ~$2 — otherwise even a fresh cap rejects a single test run.
- **Canonical struct definitions:** `internal/adapters/thunder/api/types.go`. Field tags there are the source of truth; if Thunder changes the schema, update that file and the rest follows by compile error.

### Chassis action design patterns

- **Parameters come via `CollectedData["input_data"]`, not templated into step config.** `{{.input_data.X}}` templating does not render for deterministic-action step config. The convention is `datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.field_name")`.
- Import path for the helpers: `github.com/gqls/agentchassis/platform/orchestration/datahelpers`. Sibling package to `actions`, stdlib-only, safe to import.
- New actions need three changes: add the `.go` file, register the handler in `registry.go`, create an agent definition that references the action in its workflow.
- Agent definitions must use `ON CONFLICT (type, version) WHERE deleted_at IS NULL`.
- Action handlers follow the signature `func X(ctx context.Context, params ActionParams) (interface{}, error)` and should handle `params.ExecutionContext.Action == "initialize"` as an early return.
- Action result maps are stored in `orchestration_states.collected_data` under the step's `output_field`, NOT in `final_result`.
- **File writes from actions land on whichever pod handled the message.** By default an orchestration runs inside one of the three main `agent-chassis-*` pods (not a dedicated per-agent pod). Files written to `/tmp` are only present on that one pod, and lost on pod restart. For persistent artefacts, either spawn a dedicated pod (orchestrator wrapper pattern) or write to storage (Postgres / S3). Check which pod has the file: `for pod in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[*].metadata.name}'); do kubectl -n ai-persona-system exec $pod -- ls /tmp/<dir>/; done`
- **To force dedicated pod spawning: orchestrator wrapper pattern.** Three steps: `spawn_agent → call_agent → complete_workflow`. Use `target_role` (not `agent_type`) in `call_agent`. Wrapper's `processing_mode: "orchestrator"` at top level of `default_config` (not inside workflow); `category='orchestrator'`, `agent_category='coordinator'`. Worker's `processing_mode: "task"`, `category='specialist'`, `agent_category='specialist'`.
- **`agent_definitions` three-column trap.** `category` is free text functional role (`orchestrator`, `specialist`). `agent_category` is CHECK-constrained to one of `strategist, executor, analyst, integrator, coordinator, specialist` — NO `orchestrator`. `status` is lifecycle (`active`, `experimental`, etc.). Reference row: `improvement-loop` has `category=orchestrator, agent_category=coordinator, status=experimental`.
- **Input extraction uses `ExtractActionInputs` + `ActionInputSpec`.** Register the spec via `init()` with `datahelpers.RegisterActionInputSpec(...)`. Call `datahelpers.ExtractActionInputs(params.CollectedData, params.StepConfig.Config, Spec, logger)` to get an `*ActionInputs` object with `.Get(key) string`, `.GetBool(key, def)`, `.GetInt(key, def)`, `.GetMap(key)`, `.Has(key)`, `.GetRaw(key)`. Handles three layers (explicit config paths, input_fields, flat lookup) plus deprecated mappings and defaults. Don't duplicate `GetBoolField`/`GetIntField` from `datahelpers`.
- **call_agent input_mapping: explicit fields, not blob.** `"input_mapping": {"input_data": "input_data"}` produces `input_data.input_data.X` which breaks. Map fields individually: `{"agent_type": "input_data.agent_type", "model_filter?": "input_data.model_filter"}`. Suffix `?` for optional — otherwise missing field fails the whole call.
- **Long-held transactions through pgbouncer are fragile.** Pgbouncer in transaction pool mode handles short, self-contained transactions well; long transactions that hold connections for 30+ seconds (large SELECTs streaming + large INSERTs) can trip connection-level failures ("driver: bad connection" at Go SQL layer). Default to per-batch commits for bulk work; don't wrap an entire streaming job in one transaction. Keep each transaction's work under a second where possible.
- **When action stats and DB state disagree, check the final write step explicitly.** An action can return a complete result map with correct counts while the final UPDATE never landed (silent 0-rows-affected or swallowed error). Always check `RowsAffected()` on UPDATEs that are supposed to hit exactly one row, and return errors instead of Warn+continue unless the UPDATE is truly optional.
- **Kafka trigger JSON must be flat single-line and single-quoted.** Multi-line heredocs mangle payloads silently (see `016_debugging_guide_v2.md` §9). Use `<<<'{...flat json...}'` here-strings, or build with `jq -nc` and pass as `<<<"$PAYLOAD"`.

### Eval and replay methodology

- **Replay, don't re-run.** If we have stored outputs in `llm_call_log` from a previous model, we don't need to re-invoke the production agent. Pull the rendered prompts, POST them to the new model, compare responses. Saves orchestration state pollution and is much faster.
- **Fail fast on empty responses.** A per-call loop that silently records empty/error responses and continues produces no signal. Abort on first failure, capture Ollama logs at the moment of failure, diagnose, then resume.
- **Diverse 20 > random 20.** For eval test sets, `DISTINCT ON (orchestration_id)` gives us breadth across different runs/sites; random sampling clusters on whatever recently ran.
- Monitor eval in background with `watch -n 30 "wc -l results.jsonl; tail -1 results.jsonl | jq"` — quick visibility without tailing every line.

### Progress framing

- **Work the smallest useful step.** Steps 1-3 of flywheel B were each single-focus: "prove pgvector works", "prove ollama serves", "prove chassis wires them". Each closed before the next started. No compound failures.
- **When numbers look suspicious, trust the instinct.** 4,307 results in seconds was clearly wrong; we shouldn't have pressed through.

## 15. Changelog

- 2026-04-23 (flywheel C phase 2 design locked, build pending) — Decided
  automation architecture. Chassis drives, GPU VM serves. Rejected SSH
  (synchronous, fragile) and Kafka consumer (overkill for one client).
  Chose HTTP job server on VM (FastAPI-style, POST /jobs → poll GET → fetch
  adapter). Three new chassis components: model-trainer specialist,
  model-evaluator specialist, training-flywheel-orchestrator wrapper. Three
  new tables: model_training_runs, model_artefacts, model_evaluations.
  Gated on phase 1 (manual training run) to confirm the scripts actually
  work before automating them. Design captured in §2.5.1.
- 2026-04-23 (flywheel C scripted, 70B training path chosen) — With flywheel
  A done and a clean 1,958-row dataset in Postgres, wrote `flywheel_C/` —
  five files covering VM setup, dataset pull, training, inference test,
  plus README. Target: Llama 3.3 70B Instruct via Unsloth QLoRA on single
  H100 or A100 80GB (user confirmed VM config). Training script defaults
  to 3 epochs, batch 1, grad_accum 8, lr 2e-4, lora_r 16, max_seq 4096.
  Expected: 30-90 min full train, ~150MB LoRA adapter output, ~50-70GB
  VRAM peak. 70B chosen because hardware is already available; flagged
  that 8B would likely deliver 95% of quality at 10% of inference cost
  for this narrow structured-JSON task, so a follow-up 8B run on the same
  dataset is on the table for direct comparison. Next: user runs the scripts
  on the GPU VM — smoke train (20 rows) first, then full train if smoke
  looks sensible.
- 2026-04-23 (flywheel A v3.1 → v3.2, first dataset landed) — v3 with batch
  size 500 + single big transaction failed on first bulk insert with
  "driver: bad connection" (pgbouncer interaction with long-held tx).
  v3.1 split into per-batch transactions (size 100, each commits quickly).
  v3.1 successfully streamed 1,958 rows into `training_exports.rows` but
  the final UPDATE silently didn't take — runs row stayed at zeros despite
  action returning correct counts. v3.2 adds strict error handling on
  UPDATE (RowsAffected check, immediate error return instead of Warn+continue).
  **First real training dataset now in Postgres:** export_id
  `fef7be6b-887f-4bc9-b118-a5a9992c4179`, 1,958 rows, 21.2MB, reconciled
  manually. Spawning architecture fully validated — worker runs in dedicated
  spawned pod, orchestrator wrapper pattern works end-to-end.
- 2026-04-23 (flywheel A v3 polished, awaiting chassis binary) — Reviewed v3
  against doc 001 development guide. Reworked to use canonical
  `ExtractActionInputs` + `ActionInputSpec` pattern (registers via `init()`,
  uses `inputs.Get/GetBool/GetInt`) instead of direct `ExtractNestedFieldString`.
  Caught three agent_definitions column semantics confusions: `category`
  (free text), `agent_category` (CHECK-constrained to 6 values, NOT
  `orchestrator`), and `status` (lifecycle). Fixed: orchestrator has
  `category='orchestrator', agent_category='coordinator'` matching
  improvement-loop reference row. Agent definitions applied. First trigger
  failed with v2 error message (`output_path missing`) because binary is
  still v2 — agent definitions updated in DB but Go rollout pending. BUT
  spawning verified working: worker pod
  `agent-training-data-exporter-4c4fe86e-tlg6s` spawned, role=exporter,
  parent-child wiring clean. processing_mode:orchestrator at top level +
  agent_category:coordinator is the confirmed combination.
- 2026-04-23 (flywheel A v3 design finalised, schema written) — Considered
  real-time streaming (every new `llm_call_log` row flows into
  `training_exports.rows` automatically) vs manual batch export. Chose
  **pure manual batch** for now. Real-time streaming would couple
  observability to training-pipeline logic, lose snapshot boundaries (can't
  easily reference "dataset we trained on"), and duplicate filter decisions
  between ingest and training time. Batch gives named versioned artefacts
  via `runs.id` UUIDs that support A/B comparison across retraining cycles.
  Scheduling decision: manual trigger only. Adding a `scheduled_tasks` row
  later is cheap if we find ourselves running exports often. Schema SQL
  written (`001_training_exports_schema.sql`) with two tables (`runs`,
  `rows`), three indexes, one unique constraint to prevent duplicate source
  rows within a single export, one convenience view. Dataset profile from
  first real export: 1,957 rows, p50 prompt 8,250 chars, p50 assistant 448
  chars, three dominant schemas (hero-with-CTAs 68%, minimal hero 18%,
  header/nav 9%). Time distribution: 5% March / 95% April (recent-heavy).
- 2026-04-23 (flywheel A v2 operational, v3 designed) — v2 smoke test initially
  failed silently because kcat heredoc mangled the JSON payload; routing fell
  through to a "No-op scheduled task" handler, target agent never invoked.
  Root cause documented in 016_debugging_guide_v2.md §9. Fix: here-string
  `<<<'...'` with flat single-quoted JSON. Once applied, v2 smoke test
  returned real values; full export produced 1,949 training rows (21,339,893
  bytes) for page-content-writer/iter_0. File landed in one of three chassis
  pods' `/tmp` — not a dedicated spawned pod, revealing a spawning-behaviour
  question to resolve for v3. **v3 design decision: move from file output
  to Postgres with own schema (`training_exports.runs` + `training_exports.rows`).**
  Rationale: 21MB-2GB sizes fit Postgres TOAST, no new storage system,
  queryable metadata, streamable via `\copy` at training time. Also plan
  an `training-data-export-orchestrator` wrapper to spawn the worker in
  a dedicated pod. Pending: verify `processing_mode` placement controls
  spawning behaviour before coding v3.
- 2026-04-23 (flywheel A v2 + D infrastructure) — Smoke test on v1 exporter
  revealed `{{.input_data.X}}` templating does not render in deterministic
  action step config. v2 rewrites to read from `params.CollectedData["input_data"]`
  via `datahelpers.ExtractNestedFieldString`, matching codebase convention.
  Import path `github.com/gqls/agentchassis/platform/orchestration/datahelpers`
  confirmed from project source. Agent definition simplified to static
  empty config. v2 applied to DB, chassis v2 image building.
  Flywheel D: dedicated `ollama-eval` pod memory bumped to 24Gi/28Gi
  (20Gi was 300MB short of model load requirement). Resumed eval
  targeting dedicated pod, case 1 preserved from previous run, cases 2-20
  processing. Comparison methodology drafted (3 levels: structural /
  judge / manual).
- 2026-04-22 (flywheel A action implemented) — Architectural decision: export
  lives as chassis action + agent, not ad-hoc SQL. Tried Postgres regex fence-
  stripping, hit quirks; reused existing `stripMarkdownFromResponse` helper
  from `ai_actions.go` instead. Files produced: `training_data_export.go`
  action (~280 lines), registry.go patch, `training-data-exporter` agent
  definition, Kafka trigger script. Config supports per-call parameters via
  `{{.input_data.X}}` templating so one agent handles any agent/step target.
  Awaiting chassis build + deploy.
- 2026-04-22 (flywheel A design + D partial results) — Format decided:
  ChatML messages with metadata sidecar. Response shape audited: 96%
  clean JSON, 3.7% code-fenced (need stripping), 4 oddballs to inspect.
  `agent_type` issue resolved — recent writes are 100% populated,
  historical rows can be joined to `orchestration_states.owner_agent_type`.
  D eval: case 1 completed in 27 min (119 output tokens, ~4 sec/tok).
  Shared cpu-ollama contention makes full 20-case eval impractical
  (10+ hours estimate). Need dedicated eval adapter or GPU endpoint.
- 2026-04-22 (flywheel D in progress + Ollama infra fixes) — Target agent
  chosen: `page-content-writer / iter_0_generate_content` (1,945 calls,
  150 avg tokens, 97% JSON). Briefing-agent ruled out (no logged calls).
  Test set of 20 diverse prompts generated as NDJSON. Eval approach
  settled: replay stored Claude prompts against `mistral-small3.1` via
  `/api/chat`, compare outputs automatically first.
  Infrastructure: fixed RWO PVC rolling-restart deadlock on ollama-adapter
  with `strategy.type: Recreate` in kustomize base (permanent). Added
  `OLLAMA_LOAD_TIMEOUT=10m` and `OLLAMA_KEEP_ALIVE=30m` env vars.
  First-inference cold start on 14.4GB model is ~45s, well under default
  60s load timeout was tripping it.
  Measured CPU throughput on mistral-small3.1: prompt eval ~150 tok/s,
  generation ~2.5 tok/s. Real case takes 3-6 min, worst 10-15 min.
  Runner using `curl -m 1800` and `num_predict: 250` is queued.
- 2026-04-21 (prefix patch live) — Nomic task prefixes now applied in
  production. Verified by log line `"prefix_applied":true` on
  `rag_lookup` in orchestration `ffb462ba-...`. Production RAG path
  ready for real use. Next lane: D (briefing-agent eval).
- 2026-04-21 (flywheel B step 3 complete) — Chassis integration verified.
  `rag-test-agent` orchestration ran to COMPLETED on v1.0.979. Registry
  patches confirmed live. RAG is now runnable from workflows, not just
  manual SQL. Flywheel B is functionally done; next lane is D (eval).
- 2026-04-21 (flywheel B steps 1-2 complete) — Verified end-to-end retrieval
  on real content bypassing chassis. pgvector + nomic-embed-text + cpu-ollama
  all working. Task prefixes confirmed load-bearing. Three production action
  items captured. Ready to proceed to step 3 (chassis integration test).
- 2026-04-21 (fourth pass) — UI-first pivot. Revised from "concierge first,
  UI later" to "build the UI as our own operational cockpit, use it
  ourselves, then bring customers onto it". Data curation becomes a named
  product feature (not hidden in delivery). RAG platform chosen as flagship
  ahead of text LoRA and image LoRA — those become later tiers once RAG is
  live. Offer structure and shipping ladder revised accordingly.
  Business plan spun out to separate doc.
- 2026-04-21 (third pass) — Operator shape nailed: solo, £9-12k target,
  full runway, cold acquisition only, appetite for immediate ship, tolerant
  of interim gigs. Target user refined up from "non-technical owners" to
  "technical-adjacent SMEs / agencies / ops leads". Three-tier offer
  structure locked. Five content pillars defined. Shipping ladder laid
  out through month 12. Clear list of what not to ship. Interim-gig
  discipline captured.
- 2026-04-21 (second pass) — Direction set: both layers (credible site
  + product), hosted service, separate cluster reusing the framework.
- 2026-04-21 — Initial consolidation. Pulled internal flywheel material
  from 018_canine_biology, 009_model_infrastructure, 004_improvement_loop,
  022_ai_endpoint_health.sql, 021_model_swap_and_rollback.sql,
  012b_rag_best_practices_v2. Framed the finetuning.uk product question.
