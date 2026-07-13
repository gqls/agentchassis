# PLAN — Context-Assembly Tool ("Documentation is Code"): Build and Service

**Status:** implementation plan. Moves from design discussion to building. Dogfooded on our own chassis repo first, then offered as a paid service behind the gateway being built separately.

**Builds on:** `FOCUS_context_authored_derived_change` (authored/derived/change, the bundle idea), `FOCUS_best_practice_doc_tree` + `FOCUS_doc_tree_adoption` (standards, constitution, tagging, retrieval split), `FOCUS_salience_and_multi_author_mediation` (the optimal machinery), `MASTER_autonomous_build_and_operate` (cascade, verification, trust). This plan is the concrete first build of the front half of that system.

---

## 1. Objective and thesis

**The tool.** For a development task — a new feature, a bug, a maintenance change — the tool assembles a task-scoped context bundle from the documentation and the codebase, and feeds generation against ground-truth verification, so the result is **more likely to be correct** than the current workflow of pasting code and a doc into a chat. It directly removes two failures of that workflow: a fresh chat starts with no context, and a pasted snapshot goes stale as the code moves and the window fills.

**Thesis — documentation is code.** In an AI-driven workflow the documentation (standards, intent, trajectory) is an operational input that shapes generated code as much as the source does. So it should be versioned, kept true (drift detection), and composed into generation deterministically — treated as a first-class artifact, not passive reference. The tool operationalises documentation. Whether or not this is the first tool to do so does not matter; the aim is for it to work at least as well as what exists, validated against the landscape (§7), not assumed.

**Two audiences.** Ourselves first (dogfood on the chassis repo, which also improves the current workflow immediately), then external tenants as a paid service behind the existing gateway.

---

## 2. Problems it solves (from the discussions)

- **No context in a fresh chat / paste-and-rot.** The tool assembles current context on demand from the repo and docs, so nothing is pasted and nothing goes stale.
- **Salience loss.** The bundle is composed at the right altitude per step (full intent at framing/decision points, a thin tether during implementation).
- **Completeness for rules vs recall for exploration.** Rules are loaded by tag (complete — no governing rule missed); the broad corpus is searched by embedding (recall).
- **Reuse not enforced.** A reuse search ("what already does something like this") runs before generation, making reuse-before-recreate mechanical rather than a remembered habit.
- **No ground truth in the loop.** A verification harness checks generated changes against build/test/validation rather than trusting the model's confidence.

---

## 3. Design principles

- **Engine/config split.** A tenant-agnostic engine (assembly logic, bundle format, retrieval, verification orchestration) plus per-stack adapters (language analysers, verification runners, doc-structure mapping). This is what makes it a service and, as a side effect, keeps it clean for our own use. Building it tenant-agnostic from the start is a forcing function, not overhead.
- **Design for the optimal, deliver incrementally.** Early phases expose the interfaces the cascade router, decision-point checkers, and mediator will plug into, so reaching the optimal solution is adding components, not rebuilding. This is how we get to optimal sooner: the seams are there from the first version.
- **Reuse what exists.** pgvector + nomic + ollama for retrieval; the chassis adapter pattern for the new analysers and runners; the existing `check_*.go` validators as the verification seed; the orchestrator/spawn for the agentic steps in later phases; the gateway for the service front.
- **Dogfood first.** Everything is proven on our own Go+Postgres chassis repo before a tenant sees it.

---

## 4. Architecture

### 4.1 Bundle builder (the core)
- **Code analysis.** Go: signatures and types, a call-graph neighbourhood (callers/callees by signature for the target, without their bodies), and a **reuse search** (existing functions/structs similar to the task). SQL/workflows: current schema extraction (information_schema) and workflow variable mapping. Tooling: `go/packages`, `go/ast`, gopls.
- **Doc retrieval.** Tag-based selection of the governing standards (complete, from the manifest) + embedding search over the corpus (recall), reusing the existing vector store.
- **Authored layer.** Constitution (always), the why-chain for the task's area, and direction-of-travel (freshness-stamped).
- **Assembler.** Composes these into a structured bundle: authored intent + standards + in-scope code (full text) wrapped in a signature-level neighbourhood + fresh derived state + pointers to everything else. Altitude-aware per step type.

### 4.2 Verification harness
- Build/vet/test runner + the existing `check_*.go` validators + schema/migration dry-run. Returns structured ground truth (pass/fail + errors) and drives iterate-on-failure. The `check_*.go` suite is the existing seed; the new part is the build/test runner and a clean structured-result interface.

### 4.3 Delivery
- **Phase 1:** a local CLI that emits the bundle (to paste into a chat or feed a model). Fastest path, immediate workflow improvement, no connectivity needed.
- **Service:** an API behind the gateway, same engine.

### 4.4 Seams for the optimal machinery (defined now, built later)
- A **cascade router** interface (reuse → generate+verify → compete → HITL) the generation step calls.
- **Decision-point checker** hooks (single-axis salience checks).
- A **mediator** interface for contested/cross-area changes.
Defining these as interfaces in the early phases is what lets §Phase 4 slot in.

### 4.5 Tenant isolation (service)
- Per-tenant doc store, embeddings, and config. Sandboxed verification (running tenant code is a security surface — see §6).

---

## 5. Route (phases)

Each phase stands alone and is dogfooded before it ships to tenants.

> **Status (as of this build).** A **thin slice of Phase 1 is built and exercised on real code** (the chassis adoption files), ahead of the full Phase 0 contracts: `analyser.go` (signatures, call-graph neighbourhood, per-function calls), `assembler.go` (constitution + task + reference docs + in-scope full bodies + call-graph neighbourhood + `-include` for wiring files the call graph can't reach + schema + runtime sections), `dbcontext.go` (schema via `\d`, row data with multipass sizing, and runtime evidence via `-runtime-site`), and `resolve_targets.go` (a **lexical** target-resolution baseline). First real run: a 30 KB bundle vs the script's 1.7 MB, with the call graph surfacing the right neighbourhood and `-include` closing the `registry.go` gap. **Still hand-supplied:** the framing/task text and the matched guidelines. **Next build:** the **embedding layer** — semantic recall for target resolution (and the same retrieval for matched guidelines), added once a ground-truth task→files set exists to prove it beats the lexical baseline. The Phase 0 contracts and the constitution remain the larger unbuilt foundation; the thin slice is deliberately ahead of them to test the core thesis first. Note also the terminology reconciliation flagged elsewhere: build-time **review** steps (liability, morality) are *contributors* to the build, distinct from improvement-loop **checkers** (deployed-vs-plan/spec) — to be settled when the checker/improvement-loop part is investigated.
>
> **Packaging.** The tools are now a single Go module, `contextkit/`, with the two contracts they share defined once — `internal/analysis` (the analyser's output shape) and `internal/candidates` (the ranked-candidate shape) — and each tool a command under `cmd/` referencing those by qualified name. This is the seam these graduate along: when they become chassis actions, the two `internal/` packages move under the chassis module path and the command mains become action entry points; the contracts don't change. Update to the embedding layer since: `embed.go` (Ollama-backed semantic index + offline stand-in), `fuse.go` (RRF merge), and `eval_targets.go` + `groundtruth_targets.json` (recall@N/MRR) are built and run end-to-end; the lexical baseline gets both decisive symbols within top-12 on the one seeded task, and the real semantic recall is pending a run against the Ollama model.

**Phase 0 — Foundations.**
Three things. (a) **Settle and implement the shared contracts** — the active-config schema, the work-items, decision-log, trust-ledger, capabilities-catalog, and change-layer contracts, and the bundle shape. These are the tables and shapes everything else reads and writes; verify the `site_work_items` / `agent_definitions` reuse assumptions hold at DDL time (schema-before-SQL). (b) Define the engine/config split. (c) Write the **constitution** for the chassis repo and tag the core docs by concern + `applies_to` (adoption-doc Phases 1–2), reusing the existing vector store. *Value:* the authored layer exists, the contracts are in place, and the current workflow already benefits from a tight baseline. *Entry:* none.

**Phase 1 — Bundle builder (MVP, read-only).**
Build the Go analyser (signatures, call-graph neighbourhood, reuse search), the schema extractor, tag-based rule selection + embedding search, and the assembler that produces the bundle (per `PLAN_bundle_shape_contract`). Emit the bundle via CLI for our own use. *Value:* removes paste-and-rot and the fresh-chat-no-context problem; immediately better, fresher prompting. *Entry:* Phase 0.

> **Onboarding split (which onboarding work belongs where).** Onboarding *ourselves* — producing our own config so the bundle builder has something to read — is a **Phase 0/1 prerequisite**, done with the docs-authoritative path on our own repo. The five onboarding agent specs describe the general case, but only the parts needed to onboard our own repo (run stack-discovery on our repo, extract conventions from 001/003/naming, capture our intent via the constitution and why-chain) are needed now. Onboarding *arbitrary tenants* — sandboxed probing, the code-inference convention mode, config derivation for repos we don't know — is **Phase 3 service** work and is not needed to get value for ourselves.

**Phase 2 — Verification loop.**
Add the build/test/validate harness (extending `check_*.go`) and iterate-on-failure. *Value:* "more likely to be correct" gets ground truth — the bundle improves the input, verification catches the output. *Entry:* Phase 1 in real use.

**Phase 3 — Service.**
Behind the gateway: multi-tenant, per-tenant isolation, sandboxed verification, and the **full onboarding flow for arbitrary tenants** (point at a repo + docs → derive config → tag/embed → ready), including the code-inference convention mode and the sandboxing gate. *Value:* the paid service. *Entry:* Phases 1–2 stable on our own repo.

**Phase 4 — Toward optimal.**
Implement the seams: reuse-first cascade, decision-point salience checkers, and multi-author/mediator for contested generation. Evidence-driven, on top of the proven bundle+verify base. *Value:* the correctness machinery that differentiates it. *Entry:* Phase 2 proven; seams in use.

---

## 6. Service considerations

- **Gateway integration.** Auth, billing, and routing are handled by the gateway built separately; this tool is a service behind it. Do not build billing here.
- **Multi-tenancy and isolation.** Tenant code and docs are sensitive; per-tenant stores and strict isolation are required from Phase 3.
- **Sandboxed verification.** Running tenant code to verify it is a real security surface — the Tier-C security concern made concrete. Verification must run sandboxed, and this is a gating requirement before any tenant code is executed.
- **Onboarding is the hard service problem.** Deriving a working config (languages, doc structure, verification commands, conventions) for an arbitrary repo is the adoption bottleneck — the engine is generic, but a tenant only gets value once its config is right. Worth treating onboarding quality as a first-class deliverable, not a setup script.
- **Per-stack adapters.** Go + Postgres first (us). Additional language analysers and verification runners are added as tenants need them, as adapters against the stable engine.

---

## 7. Differentiation (with the honest caveat)

The distinctive design, stated as a belief to validate rather than a fact: this is not a better code-search. It is a correctness-oriented context-plus-verification system that (a) operationalises the **authored layer** — project-specific standards and intent, loaded with completeness for rules — alongside code, (b) puts **verification in the loop** as ground truth, and (c) is built around **bounding uncertainty** (reuse-first cascade, salience checks, mediator for contested changes) rather than generate-and-hope.

**Caveat:** before "better than what's out there" is a claim, run a competitive scan of current tools (context retrieval, repo-aware assistants, verification-in-loop offerings) and locate the genuine gaps. Listed as a near-term task; novelty is not assumed.

---

## 8. Risks and open questions

- **Onboarding/config derivation** for arbitrary repos (§6) — the hardest service problem.
- **Sandboxed verification security** — gating before tenant code runs.
- **Authored-layer upkeep** — the constitution, why-chain, and tags must stay true, which is the curation/confirm-not-initiate machinery and the drift detector; without upkeep the authored layer misleads.
- **Optimal-end open questions still apply** — scope of priority changes (§9.7 of the salience doc), rival-base availability, the cascade router's tier choice (the least-bounded step). These land in Phase 4, not before.

---

## 9. First concrete step

Start Phase 0 on the chassis repo, in this order:
1. **Stand up the contracts as real tables** — active config, work-items, decision log, trust ledger, capabilities catalog, change layer, and the bundle shape. Verify the `site_work_items` / `agent_definitions` reuse assumptions against the live schema at DDL time.
2. **Write the constitution** from the standing rules across 001/003 and the working preferences, and tag the core docs.
3. **Build the Go analyser** (signatures + reuse search) and the schema extractor, and wire tag-based doc selection over the existing vector store.
4. **Assemble and emit a first bundle** as a CLI for our own use (Phase 1).

Steps 2 and 3 are independent of each other and can run in parallel once the contracts (step 1) exist. The contracts come first because everything else reads and writes them.
