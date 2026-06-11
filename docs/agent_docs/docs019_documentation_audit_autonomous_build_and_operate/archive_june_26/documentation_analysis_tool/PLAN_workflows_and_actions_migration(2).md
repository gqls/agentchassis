# Plan — migrating contextkit to chassis workflows, actions, and agents

**Status: pre-implementation, discussion-first.** Nothing here is built. The structure below is a proposal to refine together; the open-questions section is the point of this document. Update the changelog at the bottom as decisions land, and move items from "open" to "decided" rather than deleting them.

---

## What this migrates

The standalone `contextkit` module (analyser, assembler, embed, dbcontext, resolve_targets, fuse, eval_targets, plus the `analysis` and `candidates` contracts) into the chassis as workflows, actions, and agents — so context assembly and the build-time reviews run as first-class, logged, orchestrated work instead of CLI tools. This is additive: the standalone tools stay as the dogfooding and measurement harness, and `eval_targets` in particular stays offline.

## Constraints that shape the design (our standing rules, restated so the plan is judged against them)

- Workflows stay thin (SQL/data); complexity lives in Go actions.
- No subworkflows in SQL — separable concerns become spawned sub-agents with their own workflows (clearer logs, separate responsibility).
- Every agent is an orchestrator.
- Sub-agents reply on the caller's (parent's) responses topic, not their own.
- Reuse and adapt existing functions/architecture before creating new ones.
- Confirm DB schema before writing SQL; parameterised SQL only.
- text + CHECK over enums; `version` + `previous_version_id`; `deleted_at` soft-delete.
- No `logger.Debug` (won't show); log agent creation and inter-agent messages with headers and body.
- Workflow variable names must match what the actions expect.

## Confirmed model (from the guideline docs — 001 development guide, 002 system architecture)

These were the "gate" unknowns; the guides settle most of them. (The empty `00x_*.sql` files in the project were placeholders; the model lives in the guide docs.)

- **A workflow is declarative JSON in `agent_definitions.default_config` (`templates_db`).** A step is keyed by name: `{ "action": ..., "config": {...}, "next_step": ..., "output_field": ... }`. Logic stays in Go actions; the JSON expresses intent.
- **Every agent is an orchestrator, and a workflow can be a single step.** The canonical small form is the "wrapper orchestrator" (spawn → call → complete) that gives a worker its own pod, logs, and topic — so wrapping one action in its own agent is normal, not heavyweight.
- **There is a library of generic, config-driven actions to reuse before writing Go:** `query_database`, `spawn_agent`, `call_agent`, `loop`, `conditional`, `rag_lookup`, and the work-item lifecycle (`claim_work_item`, `complete_work_item`, `complete_workflow`, `fail_work_item`), among others. New Go actions are for genuinely novel compute only.
- **Group by shared context; don't make every check its own agent.** The architecture explicitly rejects "a registry of mini-actions" for LLM-assisted work: checks that share a context load are steps inside one *group agent* (one context load, one LLM call per group). Pure structural/binary checks use the existing `DiscoveryCheck` registry, not agents.
- **The promotion pattern is the evolution path.** A check starts as an action step in a group agent's workflow; when it needs independence (vision AI, a research sub-agent, its own workflow) it is promoted to a spawned sub-agent — the workflow changes one line (`action: spawn_agent`).
- **Spawning** is `spawnAgentKubernetesJobFromDefinition` (+ `setupAgentTopics`); a spawned Job pod gets a per-spawn topic `job.<corr[:8]>-<orch[:8]>-<type>-<parent_step>.requests`, its own logs/resources/idle-timeout, and terminates on completion. `agent_definitions.topics` is a declaration; the Deployment manifest actually subscribes.
- **Reuse discipline is encoded:** doc 001 says to first search `agent_definitions` for a similar agent, and `default_config::text ILIKE '%<action>%'` for an existing use of an action, before creating anything.

## Proposed structure (now aligned to that model)

**Shared Go packages.** `analysis` and `candidates` move in as-is; a `bundle` package for the assembler's render. Reuse the chassis DB layer (not `dbcontext` psql), the ollama-adapter (keep the `Embedder` interface), the generation/model adapter, the spawn API, and artifact upload. The analyser's walk becomes a library function.

**New Go actions — only the novel compute.** The analyser walk (build the code-structure index), the embedding build, and target resolution (lexical + semantic + fusion folded into one). Possibly a bundle-assemble action, though parts may be templating + `query_database`. Everything else likely maps to generics: `gather_schema_context` and `gather_runtime_evidence` are `query_database` with config; `match_guidelines` may be `rag_lookup` (investigate the overlap — A5); generation is the existing model-adapter path via `call_agent`.

**Indexing agent** (commit-triggered, SHA-keyed): a small workflow over the analyse + embed actions, marking an index version.

**Build agent** (a top-level orchestrator like `pageflow-builder` / `site-work-orchestrator`): resolve targets → optional HITL scope confirm → gather context (a few `query_database` steps + the code-context action) → assemble → generate → review → emit.

**Reviews and any LLM-assisted contributors are group agents, not individual agents.** Following `content-quality-auditor` / `design-audit-agent`: one context load, a step per check, promoted to a sub-agent only when one needs independence. This replaces the earlier proposal of separate `review_reuse` / `review_liability` / … agents. The build-time review group is the same *shape* as the deployed-site auditors — which is also where the checker / improvement-loop concept lives, so confirm reuse rather than duplicate.

**Eval stays offline** — measurement/flywheel only, never in the build path.

---

## Open questions and areas to investigate — BEFORE we implement

### A. Live-system schema and APIs — mostly answered by the guides; confirm the remainder against live + code

The Confirmed-model section settles how workflows are stored and shaped, the generic action library, the spawn mechanics and topics, the grouping principle, and the promotion pattern. What still needs confirming directly:

1. **The run-trace / orchestration tables.** Confirm `orchestration_states` (or equivalent) and what records the per-step run trace, so `gather_runtime_evidence` uses the real trace, not just `agent_error_log` + `site_work_items`. Confirm how `orchestration_id` correlates across parent and sub-agents.
2. **HITL representation.** No human-pause/approval step type appeared in the guides. Confirm whether a workflow can park for human approval and resume (for the scope-confirm and contested-review gates). If not, that is a design item, not just a lookup.
3. **The generation / model-adapter path.** Confirm the exact action/adapter the build's generate step reuses (model-agnostic), so generation is not a new integration.
4. **Artifact storage.** Confirm the artifacts table / Backblaze path for the assembled bundle and its provenance.
5. **`rag_lookup` and any existing retrieval (the most important reuse check).** A `rag_lookup` action already exists — confirm what it indexes and how, because it may overlap with `embed` / `resolve_targets`. Reuse or extend it rather than adding a parallel retrieval path.
6. **Confirm-in-code the spawn call** (`spawnAgentKubernetesJobFromDefinition`, `setupAgentTopics`) and the action-registration contract, before writing the new actions/agents.

### B. Architecture decisions not yet settled

1. **Action-vs-agent boundary, reconciled with "every agent is an orchestrator."** The proposal makes contributors sub-agents and keeps `resolve_targets`/`assemble` as actions. If every agent is an orchestrator, we should state explicitly why those two are *not* agents (pure transforms; a step-level log entry suffices) and confirm that cut is intended. Re-examine which `gather_*` steps deserve to be agents (e.g. a guidelines contributor that spawns a research sub-agent) versus plain actions.
2. **Index granularity and invalidation.** Full re-index per commit versus incremental (re-embed only changed symbols). The analyser is cheap (parse-only); embeddings are the cost. Decide the incremental strategy, how a build verifies the index matches the SHA it is building against, and what happens if code changes mid-build (the freshness race).
3. **On-demand context versus assemble-then-send — leaning decided: the gather phase is a loop.** Because tasks can get long, a single up-front assemble is likely to under- or over-shoot, so the working decision is that gathering is iterative: assemble a seed, let the build request more context, gather again, rather than one fixed sequence. This has direct precedent in the system — the `loop` action exists and `call_agent`-substep loops already run (`vet-batch process_batch`, `content-feed process_sites`), and `rag_lookup` shows retrieval-as-a-step is normal. Still to settle within that decision: how the build expresses "I need more" in a model-agnostic way (the parked second-round-trip — the model states needs in its output, the loop fulfils them and re-gathers), and the loop's termination/budget. Provenance (B8) must then record the dynamic fetched set, not a static bundle.
4. **Embedding storage at scale.** A table plus in-Go cosine is fine for one repo (thousands of symbols); multi-tenant / many repos likely needs pgvector or a vector store. Decide the threshold, and whether to start on pgvector to avoid a later migration.
4a. **CPU Ollama feasibility — gating, do this early.** Before relying on the small CPU Ollama model for the embedding index (and any in-build retrieval), check whether it is adequate at all on two axes: **speed** (time to embed the whole chassis — thousands of symbols — and per-query latency inside a build) and **quality** (does it actually beat the lexical baseline on the ground-truth set, via `eval_targets`). If it is too slow, fall back to Thunder GPU for the bulk index; if quality is poor, try a different embedding model. This also informs the `rag_lookup` reuse check (A5) — if existing RAG already uses a workable embedder, reuse it.
5. **Multi-tenancy and genericity.** The tool is meant to be generic (per-tenant config; pluggable per-language analysers). Decide whether to migrate single-tenant (chassis-only) first and generalise later, or bake tenant-scoping in from the start, and where each action loads the active-config. The analyser is Go-only (`go/ast`); per-language producers are a later concern, but the `analysis` contract should stay language-agnostic.
6. **Review gating and the revise loop.** How a review contributor "revises or gates": does a raised concern loop back to re-generate with the concern as input; how many cycles before it goes to HITL; reuse `attempt_count`/`max_attempts` (already on work items) for loop termination. Define the loop and its termination before building the reviews.
7. **Contributors versus checkers overlap (already flagged).** Whether the build-time reviews reuse any of the deployed-site `check_*.go` logic. Unconfirmed; it affects whether review actions share code with the improvement-loop checkers. Tied to the separate checker/improvement-loop investigation.
8. **Provenance model.** Where "what was in the bundle / what was fetched" is recorded (the bundle artifact plus a provenance record), and its shape. This matters more if on-demand retrieval (B3) is adopted, since the fetched set is then dynamic.
9. **The morality/liability standard source.** The configured, layered standard lives in the active-config (designed in the earlier contracts). Confirm that schema is actually in place before the review contributors depend on it.

### C. Sequencing and scope

1. **Thin-slice the migration too.** Suggested first slice: the indexing workflow + `resolve_targets` action + `assemble_bundle` action + one build workflow **without** the review contributors (reviews are a later layer). Prove the pattern end-to-end before breadth.
2. **Prove one sub-agent first.** Stand up a single context contributor as a sub-agent end-to-end to validate the spawn / topic / logging pattern before building all of them.
3. **Keep the standalone tools.** `contextkit` stays as the dogfooding and measurement harness during and after migration; `eval_targets` stays standalone. Don't delete a tool when its action exists.
4. **Schema-confirmation is the gate.** The workflow shape is now known; the remaining gate before writing the build path is A1 (run-trace tables) and A5 (`rag_lookup` overlap) — confirm those rather than building a parallel retrieval or guessing the run trace.

### D. Smaller technical doubts

1. **Analyser as a library function** (so the index action is glue) — and the qualified-names verbosity that prompted considering it. Minor; decide when moving the code.
2. **Workflow variable-name sync** with each action's expectations — a per-workflow checklist item, not a one-off.
3. **Logging levels** — no `logger.Debug`; ensure agent-creation and message logging (headers + body) is wired the way we want from the first agent.
4. **Per-language analyser pluggability** — deferred; keep the `analysis` contract language-agnostic so a non-Go producer can fill it later.

---

## Reuse inventory to verify (confirm each, then reuse — do not reimplement)

Before creating anything, doc 001's discipline: search `agent_definitions` for a similar agent, and `default_config::text ILIKE '%<action>%'` for an existing use of an action.

- **`rag_lookup` / existing retrieval** — the key one. Confirm what it indexes and how; it may already do part of what `embed`/`resolve_targets` do. Reuse or extend, don't parallel.
- Generic actions: `query_database`, `spawn_agent`, `call_agent`, `loop`, `conditional`, the work-item lifecycle — map gather/generate/control-flow onto these before writing Go.
- Chassis Postgres access layer (replaces `dbcontext` psql-shelling).
- ollama-adapter (replaces `embed`'s HTTP client; keep the `Embedder` interface).
- Spawn API (`spawnAgentKubernetesJobFromDefinition`, `setupAgentTopics`) and the action-registration contract (`registry.go`).
- Artifact upload (Backblaze) and any artifacts table.
- Generation / model adapter (for the build's generate step).
- The deployed-site auditor agents (`design-audit-agent`, `content-quality-auditor`, `site-review-agent`) as the *template* for the review group — and the locus of the checker/improvement-loop concept (possible reuse — B7).
- active-config schema (for guidelines/standards in the reviews — confirm built, B9).

## Suggested first slice (to agree before starting)

Once A1 (run-trace) and A5 (`rag_lookup`) are confirmed: build the indexing agent (analyse + embed actions) and the `resolve_targets` action, wired into a minimal build agent (resolve → assemble), with no reviews and no sub-agent contributors yet. Then add a single context contributor as a sub-agent — using the wrapper-orchestrator shape — to validate the spawn/topic/logging pattern. Reviews (as a group agent), the revise loop, and HITL follow once that spine is proven.

---

## Changelog

- 2026-06-09 — Created. Proposed structure recorded; open questions and reuse inventory enumerated. Pre-implementation; nothing built.
- 2026-06-09 — Grounded the model in docs 001/002 (the `00x_*.sql` examples were empty placeholders). Confirmed: workflows are JSON in `agent_definitions.default_config` (templates_db); step shape; the generic action library (incl. `rag_lookup`); group-agents-by-shared-context over individual mini-action agents; the promotion pattern; spawn mechanics/topics. Corrected the proposed structure accordingly (group agents, generic-action reuse, fewer new Go actions). Section A reduced to the remaining live/code confirmations; flagged `rag_lookup` as the key reuse check. Still pre-implementation.
- 2026-06-09 — Two real `agent_definitions` rows (`page-build-handler`, `build-dispatch-loop`) and the action code (`spawn_actions.go`, `call_agent.go`, `ai_actions.go`) confirmed the model concretely; findings folded into the debug guide (016 §6.0: agent = DB row not Go type; step shape; spawn/call as a pair; the description-vs-config trap). Decisions: gather phase is a **loop** (B3), reasoning = long tasks + existing loop/`rag_lookup` precedent; added the **CPU Ollama feasibility check** (B4a, speed + quality) as an early gate on the embedding layer. Still pre-implementation.
