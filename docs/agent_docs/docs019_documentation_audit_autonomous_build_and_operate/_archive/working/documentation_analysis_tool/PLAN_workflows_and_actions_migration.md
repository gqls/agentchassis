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

## Proposed structure (to refine, not settled)

**Shared Go packages (reused, not rebuilt).** `analysis` and `candidates` move into the chassis as-is; add a `bundle` package for the assembler's render logic so the assemble action stays thin. The chassis DB layer replaces `dbcontext`'s psql-shelling; the existing ollama-adapter replaces `embed`'s HTTP client (keep the `Embedder` interface); the analyser's walk becomes a library function so the index action is glue.

**Actions (leaf Go units, one responsibility, registered like `registry.go`).** `index_code`, `index_embeddings`; `resolve_targets` (lexical + semantic + fusion collapsed into one, since it is cohesive compute over the index); `gather_code_context`, `gather_schema_context`, `gather_runtime_evidence`, `match_guidelines`; `assemble_bundle`; and the reviews `review_reuse`, `review_liability`, `review_morality`, `review_correctness`.

**Workflows (thin, data).** An **indexing workflow** (`index_code` → `index_embeddings` → mark index version), commit-triggered, keyed by SHA. A per-task **build workflow** (resolve targets → optional HITL scope confirm → gather context slices → assemble → generate → review → emit/apply), with reviews able to revise-or-gate and route contested calls to HITL.

**Agents (orchestrators; sub-agents reply on the parent topic, separate containers).** A top-level **build agent** runs the build workflow. The **context contributors** (code, schema, runtime, guidelines) and **review contributors** (reuse, liability, morality, correctness) are sub-agents with their own workflows and containers. `resolve_targets` and `assemble_bundle` stay as plain actions inside the build workflow unless we later want HITL or a separate message trail on them.

**Eval stays offline.** `eval_targets` + the ground-truth set is a measurement/flywheel tool, never in the live build path, so it can't pollute build logs; it feeds resolution tuning (Thunder) separately.

---

## Open questions and areas to investigate — BEFORE we implement

### A. Live-system schema and APIs to confirm (the gate: no SQL or Go until these are known)

1. **How a workflow is defined and stored.** We have only inferred workflow shape from step names seen in logs (`check_has_ready_sections`, `complete_workflow`, …). Pull `\d agent_definitions` and any `workflows`/`workflow_steps` table, plus a real example row, and answer: how are steps ordered; how does a step name the action it runs; how are branches/conditions encoded; how are variables passed between steps and named; where do per-step inputs/outputs live. We cannot write a workflow until this is concrete.
2. **Orchestration / run-state tables.** Confirm `orchestration_states` (or equivalent) and whatever records the run trace, so `gather_runtime_evidence` can use the real trace rather than just `agent_error_log` + `site_work_items` (which is all the thin-slice tool used because those two schemas were confirmed). Confirm how `orchestration_id` correlates across parent and sub-agents.
3. **The spawn / sub-agent API in Go.** Find the actual call an agent makes to spawn a sub-agent, and the wiring that sets the sub-agent's responses topic to the parent's. Confirm where headers + body are attached to messages and where the spawn and the messages are logged. This determines how contributors are spawned and how they reply.
4. **The action registry contract.** From `registry.go` (`GetAction`, `IsLocalAction`, `ListActions`) confirm exactly what an action must implement and how it registers (input/output types, signature, local vs remote). New actions must match this exactly.
5. **Kafka topic conventions.** Confirm the responses-topic naming and the message envelope (headers/body) against the running cluster, so our logging and parent-reply wiring match what exists rather than inventing a parallel convention.
6. **The Go Postgres access layer.** Find the chassis's existing DB access pattern (repository/queries layer) so `index_code` / `gather_*` / the store actions use it. `dbcontext`'s psql-shelling was a thin-slice expedient and does not move into the chassis.
7. **HITL representation.** Does a workflow already have a pause / human-approval step type? If so, reuse it for the scope-confirm and the contested-review gates. If not, that is both a build item and a larger design question (how a run parks and resumes).
8. **Artifact storage.** Confirm the existing artifact path (GitHub → Actions → Backblaze) and any artifacts table, for storing the assembled bundle and its provenance.
9. **The generation path.** The chassis already generates site content via a model adapter. Confirm that path so the build workflow's "generate" step reuses it (staying model-agnostic) rather than adding a new integration.

### B. Architecture decisions not yet settled

1. **Action-vs-agent boundary, reconciled with "every agent is an orchestrator."** The proposal makes contributors sub-agents and keeps `resolve_targets`/`assemble` as actions. If every agent is an orchestrator, we should state explicitly why those two are *not* agents (pure transforms; a step-level log entry suffices) and confirm that cut is intended. Re-examine which `gather_*` steps deserve to be agents (e.g. a guidelines contributor that spawns a research sub-agent) versus plain actions.
2. **Index granularity and invalidation.** Full re-index per commit versus incremental (re-embed only changed symbols). The analyser is cheap (parse-only); embeddings are the cost. Decide the incremental strategy, how a build verifies the index matches the SHA it is building against, and what happens if code changes mid-build (the freshness race).
3. **On-demand context versus assemble-then-send (the parked tool-use / second-round-trip fork).** This shapes the build workflow directly: do the `gather_*` actions run up-front to assemble one bundle, or does the model request more context mid-build via a second round-trip? Parked earlier; it must be resolved before the build workflow's shape is fixed, because it changes whether gathering is a fixed sequence or a loop.
4. **Embedding storage at scale.** A table plus in-Go cosine is fine for one repo (thousands of symbols); multi-tenant / many repos likely needs pgvector or a vector store. Decide the threshold, and whether to start on pgvector to avoid a later migration.
5. **Multi-tenancy and genericity.** The tool is meant to be generic (per-tenant config; pluggable per-language analysers). Decide whether to migrate single-tenant (chassis-only) first and generalise later, or bake tenant-scoping in from the start, and where each action loads the active-config. The analyser is Go-only (`go/ast`); per-language producers are a later concern, but the `analysis` contract should stay language-agnostic.
6. **Review gating and the revise loop.** How a review contributor "revises or gates": does a raised concern loop back to re-generate with the concern as input; how many cycles before it goes to HITL; reuse `attempt_count`/`max_attempts` (already on work items) for loop termination. Define the loop and its termination before building the reviews.
7. **Contributors versus checkers overlap (already flagged).** Whether the build-time reviews reuse any of the deployed-site `check_*.go` logic. Unconfirmed; it affects whether review actions share code with the improvement-loop checkers. Tied to the separate checker/improvement-loop investigation.
8. **Provenance model.** Where "what was in the bundle / what was fetched" is recorded (the bundle artifact plus a provenance record), and its shape. This matters more if on-demand retrieval (B3) is adopted, since the fetched set is then dynamic.
9. **The morality/liability standard source.** The configured, layered standard lives in the active-config (designed in the earlier contracts). Confirm that schema is actually in place before the review contributors depend on it.

### C. Sequencing and scope

1. **Thin-slice the migration too.** Suggested first slice: the indexing workflow + `resolve_targets` action + `assemble_bundle` action + one build workflow **without** the review contributors (reviews are a later layer). Prove the pattern end-to-end before breadth.
2. **Prove one sub-agent first.** Stand up a single context contributor as a sub-agent end-to-end to validate the spawn / topic / logging pattern before building all of them.
3. **Keep the standalone tools.** `contextkit` stays as the dogfooding and measurement harness during and after migration; `eval_targets` stays standalone. Don't delete a tool when its action exists.
4. **Schema-confirmation is the gate.** No workflow or SQL until A1 and A2 are confirmed; otherwise we build against a guessed shape and rework.

### D. Smaller technical doubts

1. **Analyser as a library function** (so the index action is glue) — and the qualified-names verbosity that prompted considering it. Minor; decide when moving the code.
2. **Workflow variable-name sync** with each action's expectations — a per-workflow checklist item, not a one-off.
3. **Logging levels** — no `logger.Debug`; ensure agent-creation and message logging (headers + body) is wired the way we want from the first agent.
4. **Per-language analyser pluggability** — deferred; keep the `analysis` contract language-agnostic so a non-Go producer can fill it later.

---

## Reuse inventory to verify (confirm each exists, then reuse — do not reimplement)

- Chassis Postgres access layer (replaces `dbcontext` psql-shelling).
- ollama-adapter (replaces `embed`'s HTTP client; keep the `Embedder` interface, point it at the adapter).
- Action registry (`registry.go`) and its registration contract.
- Spawn / orchestration API and the responses-topic wiring.
- Kafka topic helpers and the message envelope (headers/body).
- Artifact upload (Backblaze) and any artifacts table.
- Generation / model adapter (for the build's generate step).
- `check_*.go` family (for possible review reuse — to investigate, B7).
- active-config schema (for guidelines/standards in the reviews — confirm built, B9).

## Suggested first slice (to agree before starting)

Once A1/A2/A3/A4/A6 are confirmed: build the indexing workflow and the `resolve_targets` and `assemble_bundle` actions, wired into one build agent running a minimal build workflow (resolve → assemble), with no reviews and no sub-agent contributors yet. Then add a single context contributor as a sub-agent to validate the spawn/topic/logging pattern. Reviews, the revise loop, HITL, and the remaining contributors follow once that spine is proven.

---

## Changelog

- 2026-06-09 — Created. Proposed structure recorded; open questions and reuse inventory enumerated. Pre-implementation; nothing built. Next: confirm the live-system schema/API items in section A, then agree the first slice.
